#!/usr/bin/env python3
"""Runtime dependency licence gate (ADR-0006 `licenses`, G4/G7).

Reads the Go inventory (go-licenses `report` CSV: module,url,licence) and the JS inventory
(`pnpm licenses list --prod --json`: {licence: [package,...]}), checks every licence against the
owner-approved allow list and writes a machine-readable inventory plus a summary.

Allowed: MIT, BSD-2-Clause, BSD-3-Clause, ISC, Apache-2.0, Unlicense.
SPDX expressions: `A OR B` passes when ANY alternative is allowed; `A AND B` needs every part
allowed; parentheses are honoured. Anything else (off-list, UNKNOWN, empty, non-SPDX such as
`MIT*`) fails. An empty inventory fails: nothing proved is not a PASS.
Exit 0 = every dependency allowed, 1 = at least one violation or unreadable input.
"""
import argparse
import csv
import json
import os
import re
import sys

ALLOWED = {"MIT", "BSD-2-Clause", "BSD-3-Clause", "ISC", "Apache-2.0", "Unlicense"}
# Usual spellings of the allowed ids (compared case-insensitively). Deliberately short:
# an ambiguous name such as plain "BSD" is NOT mapped and therefore fails.
SYNONYMS = {
    "apache 2.0": "Apache-2.0",
    "apache-2": "Apache-2.0",
    "apache2": "Apache-2.0",
    "apache license 2.0": "Apache-2.0",
    "apache license, version 2.0": "Apache-2.0",
    "bsd 2-clause": "BSD-2-Clause",
    "bsd-2": "BSD-2-Clause",
    "bsd 3-clause": "BSD-3-Clause",
    "bsd-3": "BSD-3-Clause",
    "the unlicense": "Unlicense",
}
TOKEN = re.compile(r"\(|\)|\bOR\b|\bAND\b|[^\s()]+(?:\s+(?!OR\b|AND\b)[^\s()]+)*", re.IGNORECASE)


def normalise(name):
    key = name.strip()
    low = key.lower()
    if low in SYNONYMS:
        return SYNONYMS[low]
    for allowed in ALLOWED:
        if low == allowed.lower():
            return allowed
    return key


def evaluate(expr):
    """True when the SPDX expression is satisfied by the allow list."""
    tokens = [t for t in TOKEN.findall(expr or "")]
    if not tokens:
        return False
    pos = 0

    def parse_or():
        nonlocal pos
        value = parse_and()
        while pos < len(tokens) and tokens[pos].upper() == "OR":
            pos += 1
            rhs = parse_and()
            value = value or rhs  # both sides are parsed: a malformed tail still raises
        return value

    def parse_and():
        nonlocal pos
        value = parse_atom()
        while pos < len(tokens) and tokens[pos].upper() == "AND":
            pos += 1
            rhs = parse_atom()
            value = value and rhs
        return value

    def parse_atom():
        nonlocal pos
        if pos >= len(tokens):
            raise ValueError("unexpected end")
        tok = tokens[pos]
        if tok == "(":
            pos += 1
            value = parse_or()
            if pos >= len(tokens) or tokens[pos] != ")":
                raise ValueError("missing )")
            pos += 1
            return value
        if tok == ")" or tok.upper() in ("OR", "AND"):
            raise ValueError("unexpected " + tok)
        pos += 1
        return normalise(tok) in ALLOWED

    try:
        value = parse_or()
    except ValueError:
        return False
    return value and pos == len(tokens)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--go-csv", required=True)
    ap.add_argument("--js-json", required=True)
    ap.add_argument("--out", required=True)
    ap.add_argument("--first-party", required=True, help="Go module path prefix of this repository (not checked)")
    args = ap.parse_args()

    problems = []
    rows = []  # (ecosystem, name, version, licence, verdict)

    try:
        with open(args.go_csv, newline="") as fh:
            for rec in csv.reader(fh):
                if len(rec) < 3:
                    problems.append(f"go: malformed line {rec!r}")
                    continue
                name, _url, lic = rec[0], rec[1], rec[2]
                if name == args.first_party or name.startswith(args.first_party + "/"):
                    rows.append(("go", name, "", lic, "first-party (not checked)"))
                    continue
                verdict = "ALLOWED" if evaluate(lic) else "REJECTED"
                rows.append(("go", name, "", lic, verdict))
                if verdict != "ALLOWED":
                    problems.append(f"go: {name}: licence {lic!r} is not on the allow list")
    except OSError as exc:
        problems.append(f"go: cannot read inventory: {exc}")

    try:
        with open(args.js_json) as fh:
            data = json.load(fh)
        if not isinstance(data, dict):
            raise ValueError("top level is not an object")
        for lic, pkgs in data.items():
            for pkg in pkgs:
                versions = ",".join(pkg.get("versions", []))
                verdict = "ALLOWED" if evaluate(lic) else "REJECTED"
                rows.append(("js", pkg.get("name", "?"), versions, lic, verdict))
                if verdict != "ALLOWED":
                    problems.append(f"js: {pkg.get('name', '?')}@{versions}: licence {lic!r} is not on the allow list")
    except (OSError, ValueError) as exc:
        problems.append(f"js: cannot read inventory: {exc}")

    for eco in ("go", "js"):
        if not any(r[0] == eco and not r[4].startswith("first-party") for r in rows):
            problems.append(f"{eco}: empty inventory (nothing was proved)")

    os.makedirs(args.out, exist_ok=True)
    with open(os.path.join(args.out, "licenses-inventory.csv"), "w", newline="") as fh:
        w = csv.writer(fh)
        w.writerow(["ecosystem", "name", "version", "licence", "verdict"])
        w.writerows(sorted(rows))
    with open(os.path.join(args.out, "licenses-summary.txt"), "w") as fh:
        fh.write("allowed: " + ", ".join(sorted(ALLOWED)) + "\n")
        fh.write(f"checked: {sum(1 for r in rows if r[4] in ('ALLOWED', 'REJECTED'))} dependencies\n")
        fh.write("result: " + ("PASS" if not problems else "FAIL") + "\n")
        for p in problems:
            fh.write("violation: " + p + "\n")

    for p in problems:
        print("check-licenses: " + p, file=sys.stderr)
    if problems:
        print(f"check-licenses: FAIL ({len(problems)} problem(s))", file=sys.stderr)
        return 1
    print(f"check-licenses: PASS ({sum(1 for r in rows if r[4] == 'ALLOWED')} runtime dependencies, all on the allow list)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
