#!/usr/bin/env python3
"""Exhaustive table test of the level invariants I1-I4 of scripts/review-comment.py (called by review-comment_test.sh).

Every combination of: the level stage 1 declares x the ledger stage 1 writes x an open entry carried from the previous
comment x what the deep run does. The expectation (oracle) is written here, independently of the script. Prints one
`FAIL ...` line per violated invariant and `cases=N failures=M`; exit 0 only when M is 0.
"""
import importlib.util
import itertools
import json
import os
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
SCRIPTS = os.path.join(HERE, "..")
HEAD = "a" * 40
LEVELS = ["NONE", "MEDIUM", "HIGH", "CRITICAL"]


def load(name, filename):
    spec = importlib.util.spec_from_file_location(name, os.path.join(SCRIPTS, filename))
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


rc = load("review_comment", "review-comment.py")
gate = load("review_gate", "review-gate.py")


def entry(eid, kind, evidence, sev, status="open", file="a.go", line=3, **extra):
    e = {"id": eid, "kind": kind, "evidence": evidence, "sev": sev, "file": file, "line": line, "scenario": "s",
         "status": status, "detail": "d", "fix": "f"}
    e.update(extra)
    return e


STAGE1_LEDGERS = {
    "missing": None,
    "empty": [],
    "notes only": [entry("s-note", "note", "INFERRED", "LOW")],
    "open MEDIUM": [entry("s-med", "finding", "INFERRED", "MEDIUM")],
    "open HIGH OBSERVED": [entry("s-high", "finding", "OBSERVED", "HIGH")],
    "open CRITICAL INFERRED": [entry("s-crit-inf", "finding", "INFERRED", "CRITICAL")],
    "open CRITICAL OBSERVED": [entry("s-crit-obs", "finding", "OBSERVED", "CRITICAL")],
}
CARRIED = {
    "none": None,
    "open HIGH": entry("dh", "finding", "OBSERVED", "HIGH", source="deep", first_sha=HEAD, location="unverified"),
    "open CRITICAL OBSERVED": entry("dh", "finding", "OBSERVED", "CRITICAL", source="deep", first_sha=HEAD, location="unverified"),
}
FIXED_DH = entry("dh", "finding", "OBSERVED", "HIGH", status="fixed")
NEW_HIGH = entry("newh", "finding", "OBSERVED", "HIGH", line=5)
DEEP_RUNS = {
    "not run": None,
    "PASS marks carried fixed": {"level": "NONE", "blast_radius": "x", "summary": "s", "findings": [FIXED_DH]},
    "PASS omits carried": {"level": "NONE", "blast_radius": "x", "summary": "s", "findings": []},
    "PASS adds new HIGH": {"level": "NONE", "blast_radius": "x", "summary": "s", "findings": [NEW_HIGH]},
    "NOT RUN with a fixed entry": {"level": "NONE", "blast_radius": "x", "summary": "s", "findings": [FIXED_DH, {"kind": "note"}]},
    "crash (no output)": "",
}


def worth(e):
    """Oracle: the level one entry supports (I2: a CRITICAL needs OBSERVED, a file and a line > 0, else HIGH)."""
    if e.get("kind") == "note" or str(e.get("status")).lower() != "open":
        return "NONE"
    sev = e.get("sev")
    if sev == "CRITICAL":
        return "CRITICAL" if e.get("evidence") == "OBSERVED" and e.get("file") and e.get("line", 0) > 0 else "HIGH"
    return sev if sev in LEVELS else "MEDIUM"


def level_of(declared, entries):
    """Oracle for I1 + I2: max(declared, every open entry); a declared CRITICAL without an OBSERVED CRITICAL entry is HIGH."""
    if declared == "CRITICAL" and not any(worth(e) == "CRITICAL" for e in entries):
        declared = "HIGH"
    return max([declared] + [worth(e) for e in entries], key=LEVELS.index)


def stage1_text(declared, ledger):
    body = "Triage: %s | blast radius: x\n\ntext\n" % declared
    if ledger is not None:
        body += "\n<details><summary>Review ledger</summary>\n\n```json\n%s\n```\n\n</details>" % json.dumps({"sha": "x", "findings": ledger})
    return body


def previous_comment(carried):
    entries = [dict(carried)] if carried else []
    return "Triage: NONE | blast radius: x\n%s\nTool: PASS (1 files)\n\n<details><summary>Review ledger</summary>\n\n```json\n%s\n```\n\n</details>\n" % (
        rc.MARKER, json.dumps({"sha": "b" * 40, "findings": entries}))


def posted_level(body):
    return rc.VERDICT.match(body.splitlines()[0]).group(1)


def gate_level(body):
    found = gate.latest_verdict([{"user": {"login": "github-actions[bot]"}, "body": body, "id": 1, "created_at": "2026-11-14T10:00:00Z"}])
    return found[0] if found else None


def main():
    cases, failures = 0, []

    def bad(label, msg):
        failures.append("FAIL %s: %s" % (label, msg))

    for declared, (lname, ledger), (cname, carried), (dname, deep) in itertools.product(LEVELS, STAGE1_LEDGERS.items(), CARRIED.items(), DEEP_RUNS.items()):
        cases += 1
        label = "declared=%s ledger=%s carried=%s deep=%s" % (declared, lname, cname, dname)
        prev = previous_comment(carried)
        prev_ledger = rc.ledger_of(prev)
        # ---- stage 1 (round 1)
        body = rc.compose(stage1_text(declared, ledger), "Tool: PASS (1 files)", HEAD, rc.DEEP_PENDING_LINE if deep is not None else None,
                          "Reviewed: aaaaaaa (x..y) at 10:00 UTC", prev_ledger)
        if body is None:
            bad(label, "stage 1 comment not written")
            continue
        ledger1 = rc.ledger_of(body)
        s1_entries = [dict(e) for e in (ledger or [])]
        carried_entries = [dict(carried)] if (carried and ledger is not None) else []
        after_stage1 = s1_entries + carried_entries
        line1 = declared if ledger is None else level_of(declared, after_stage1)
        if posted_level(body) != line1:
            bad(label, "stage 1 line is %s, expected %s" % (posted_level(body), line1))
        # ---- the deep run
        final_body, final_entries = body, after_stage1
        if deep is not None:
            os.environ["DEEP_JSON"] = deep if isinstance(deep, str) else json.dumps(deep)
            data, why = rc.deep_output(HEAD)
            final_body = rc.merge_deep(body, data, why, HEAD)
            fixed_now = dname == "PASS marks carried fixed" and carried_entries
            added = [dict(NEW_HIGH)] if dname == "PASS adds new HIGH" else []
            final_entries = [e for e in after_stage1 if not (fixed_now and e["id"] == "dh")] + added
            if fixed_now:
                final_entries.append(dict(FIXED_DH))  # present, status fixed (it leaves the ledger at the next stage 1)
            if dname in ("NOT RUN with a fixed entry", "crash (no output)") and "Deep: NOT RUN" not in final_body:
                bad(label, "a NOT RUN deep run is not announced")
        final1 = posted_level(final_body)
        expected_final = line1 if deep is None else max(line1, level_of("NONE", final_entries), key=LEVELS.index)
        if final1 != expected_final:
            bad(label, "final level %s, expected %s" % (final1, expected_final))
        if gate_level(final_body) != final1:
            bad(label, "the gate reads %s, the comment says %s" % (gate_level(final_body), final1))
        led = rc.ledger_of(final_body)
        # I2: a CRITICAL line needs an open OBSERVED CRITICAL entry with file and line (a missing ledger is the model's own)
        # (the accepted consequence: a carried CRITICAL that the deep run marks fixed leaves the line where stage 1 put it)
        accepted = dname == "PASS marks carried fixed" and cname == "open CRITICAL OBSERVED"
        if final1 == "CRITICAL" and ledger is not None and not accepted:
            if not any(f.get("status") == "open" and f.get("sev") == "CRITICAL" and f.get("evidence") == "OBSERVED" and f.get("file") and f.get("line", 0) > 0
                       for f in (led or {}).get("findings", [])):
                bad(label, "CRITICAL without an open OBSERVED CRITICAL entry")
        # I3: statuses and the counter of the carried entry
        if carried_entries:
            mine = next((f for f in (led or {}).get("findings", []) if f.get("id") == "dh"), None)
            want_status = "fixed" if dname == "PASS marks carried fixed" else "open"
            if mine is None:
                bad(label, "the carried entry vanished")
            else:
                if mine.get("status") != want_status:
                    bad(label, "carried entry status %s, expected %s" % (mine.get("status"), want_status))
                if want_status == "open" and mine.get("carried") != 1:
                    bad(label, "carried counter %s after one run, expected 1" % mine.get("carried"))
        # ---- stage 1 again (round 2, no deep stage), the model omits the carried entries again
        body2 = rc.compose(stage1_text(declared, ledger), "Tool: PASS (1 files)", HEAD, None, "Reviewed: aaaaaaa (x..y) at 11:00 UTC", rc.ledger_of(final_body))
        led2 = rc.ledger_of(body2)
        if carried_entries:
            mine2 = next((f for f in (led2 or {}).get("findings", []) if f.get("id") == "dh"), None)
            if dname == "PASS marks carried fixed":
                if mine2 is not None:
                    bad(label, "a fixed entry is carried again")
            elif mine2 is None or mine2.get("status") != "open" or mine2.get("carried") != 2:
                bad(label, "after two runs the carried entry is %s, expected open with carried 2" % (mine2,))
        open_after = [e for e in (final_entries if led else []) if str(e.get("status")) == "open" and e["id"] not in {x["id"] for x in s1_entries}]
        expected2 = declared if ledger is None else level_of(declared, s1_entries + open_after)
        if posted_level(body2) != expected2:
            bad(label, "round 2 level %s, expected %s" % (posted_level(body2), expected2))
    for f in failures:
        print(f)
    print("cases=%d failures=%d" % (cases, len(failures)))
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main())
