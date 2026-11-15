#!/usr/bin/env bash
# Tests for `scripts/findings.py settle <pr>`, notes closed after triage (owner, 2026-11-15): for an item of a NOTES
# issue, a NOT A DEFECT or DECIDED verdict that cites a reference which exists at the tip of develop (`path:line` or
# `path:a-b` inside the file, or a docs/decisions/NNNN-*.md file, checked by the script, never by the model) ticks the
# box, and the issue closes as not planned once every box is settled. Without a verifiable reference it stays a comment
# (rule H5 of #187). A FINDING is never closed by those verdicts: a human does.
# Usage: scripts/tests/findings-settle-notes_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
tool="$root/scripts/findings.py"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin" "$tmp/st/tree/src" "$tmp/st/tree/docs/decisions"
printf 'l1\nl2\nl3\nl4\nl5\n' >"$tmp/st/tree/src/a.go"
printf '# ADR\n' >"$tmp/st/tree/docs/decisions/0014-review.md"

# Fake gh: issues in $ST/issues.json, the pull request in $ST/pr.json, the tip of develop in $ST/tree.
cat >"$tmp/bin/gh" <<'PY'
#!/usr/bin/env python3
import base64, json, os, sys
st = os.environ["ST"]
args = sys.argv[1:]
method, fields, endpoint, i = "GET", {}, None, 1
while i < len(args):
    a = args[i]
    if a == "-X": method = args[i + 1]; i += 2
    elif a == "-f": k, v = args[i + 1].split("=", 1); fields[k] = v; i += 2
    elif a == "--paginate": i += 1
    else: endpoint = a; i += 1
path = endpoint.split("?")[0]
parts = path.split("/")
issues = json.load(open(os.path.join(st, "issues.json")))
def save(): json.dump(issues, open(os.path.join(st, "issues.json"), "w"))
def find(n): return next(x for x in issues if x["number"] == int(n))
if method != "GET":
    open(os.path.join(st, "log"), "a").write("%s %s %s\n" % (method, path, json.dumps(fields, sort_keys=True)))
if method == "GET" and "contents" in parts:
    assert endpoint.endswith("?ref=develop"), endpoint
    f = os.path.join(st, "tree", *parts[parts.index("contents") + 1:])
    if not os.path.isfile(f): sys.stderr.write("404\n"); sys.exit(1)
    print(json.dumps({"type": "file", "encoding": "base64", "content": base64.b64encode(open(f, "rb").read()).decode()}))
elif method == "GET" and parts[-1] == "files":
    print(open(os.path.join(st, "files.json")).read())
elif method == "GET" and parts[-2] == "pulls":
    print(open(os.path.join(st, "pr.json")).read())
elif method == "GET" and parts[-2] == "issues":
    x = dict(find(parts[-1])); x.pop("comments", None); print(json.dumps(x))
elif method == "GET" and parts[-1] == "comments":
    print(json.dumps(find(parts[-2])["comments"]))
elif method == "POST" and parts[-1] == "comments":
    find(parts[-2])["comments"].append({"user": {"login": "github-actions[bot]"}, "body": fields["body"]}); save(); print("{}")
elif method == "PATCH" and parts[-2] == "issues":
    x = find(parts[-1])
    for k in ("body", "state", "state_reason"):
        if k in fields: x[k] = fields[k]
    save(); print("{}")
else:
    sys.stderr.write("unexpected %s %s\n" % (method, endpoint)); sys.exit(2)
PY
chmod +x "$tmp/bin/gh"

setup() {
  python3 -I - "$tmp/st" <<'PY'
import json, os, sys
st = sys.argv[1]
def notes(n, pr, items):
    body = "<!-- finding:PR%d#notes -->\nNotes.\n\n" % pr + "\n".join(
        "- [ ] `%s` [OBSERVED] `src/a.go:1` remark %s" % (k, k) for k in items)
    return {"number": n, "state": "open", "state_reason": None, "title": "PR%d notes" % pr, "created_at": "2026-11-13T00:00:00Z",
            "labels": [{"name": "kind:note"}], "body": body, "comments": []}
finding = {"number": 39, "state": "open", "state_reason": None, "title": "a finding", "created_at": "2026-11-13T00:00:00Z",
           "labels": [{"name": "finding"}], "body": "<!-- finding:PR90#f39 -->\n\n- Where: `src/a.go:1`\n", "comments": []}
issues = [notes(30, 91, ["n1", "n2"]), notes(31, 92, ["n1"]), notes(32, 93, ["n1"]), notes(33, 94, ["n1"]), notes(34, 95, ["n1", "n2"]),
          notes(35, 96, ["n1", "n2"]), notes(36, 97, ["n1", "n2"]), finding]
rows = [("#30/n1", "#30", "NOT A DEFECT", "the guard sits at src/a.go:3, the review missed it"),
        ("#30/n2", "#30", "DECIDED", "docs/decisions/0014-review.md settles it"),
        ("#31/n1", "#31", "NOT A DEFECT", "see src/gone.go:1"),
        ("#32/n1", "#32", "DECIDED", "see src/a.go:99 (past the end)"),
        ("#33/n1", "#33", "NOT A DEFECT", "it is fine, trust me"),
        ("#34/n1", "#34", "NOT A DEFECT", "src/a.go:2-4 shows it"),
        ("#35/n1", "#35", "NOT A DEFECT", "src/a.go:4-9 shows it"),
        ("#36/n1", "#36", "DECIDED", "docs/decisions/9999-none.md"),
        ("#36/n2", "#36", "NOT A DEFECT", "../../etc/passwd:1 and src/a.go:0"),
        ("#39", "#39", "NOT A DEFECT", "src/a.go:3 proves it")]
issues.append({"number": 40, "state": "open", "state_reason": None, "title": "Review batch 2026-11-15", "labels": [{"name": "review-batch"}],
               "body": "<!-- review-batch -->\n# Review batch 2026-11-15\n\n" + "\n".join(
                   "- [ ] %s x <!-- batch-item:%s -->" % (k, k) for k, *_ in rows), "comments": []})
json.dump(issues, open(os.path.join(st, "issues.json"), "w"))
json.dump([{"filename": "src/other.go"}], open(os.path.join(st, "files.json"), "w"))
body = "\n".join(["Batch fix.", "", "Batch: #40", "", "| Item | Source | Decision | Reason |", "|---|---|---|---|"]
                 + ["| %s | %s | %s | %s |" % r for r in rows])
json.dump({"number": 170, "merged_at": "2026-11-15T12:00:00Z", "base": {"ref": "develop"},
           "head": {"ref": "fix/review/batch-2026-11-15", "repo": {"full_name": "o/r"}}, "body": body}, open(os.path.join(st, "pr.json"), "w"))
PY
  : >"$tmp/st/log"
}
run() {
  rc=0
  out="$(PATH="$tmp/bin:$PATH" ST="$tmp/st" REPO=o/r python3 -I "$tool" settle "$@" 2>&1)" || rc=$?
}
field() { python3 -I -c 'import json,sys; x=[i for i in json.load(open(sys.argv[1])) if i["number"]==int(sys.argv[2])][0]; print(eval(sys.argv[3]))' "$tmp/st/issues.json" "$1" "$2"; }
state() { field "$1" '(x["state"], x["state_reason"])'; }
unticked() { field "$1" 'x["body"]' | grep -c -- '- \[ \]' || true; }

setup
run 170
[ "$rc" -eq 0 ] && ok "settle runs (rc=0)" || ko "settle rc=$rc ($out)"
[ "$(state 30)" = "('closed', 'not_planned')" ] && [ "$(unticked 30)" -eq 0 ] && ok "notes: NOT A DEFECT (path:line) and DECIDED (existing ADR) tick both boxes, the issue closes as not planned" || ko "30 ($(state 30) $(field 30 'x["body"]'))"
field 30 'x["comments"][-1]["body"]' | grep -q 'the guard sits at src/a.go:3' && ok "the verdict comment is posted" || ko "30 comment ($(field 30 'x["comments"]'))"
[ "$(state 31)" = "('open', None)" ] && [ "$(unticked 31)" -eq 1 ] && field 31 'x["comments"][-1]["body"]' | grep -q 'stays open: a human closes it as not planned' && ok "a citation to a missing file: box unticked, issue open, today's comment" || ko "31 ($(state 31) $(field 31 'x["comments"]'))"
[ "$(state 32)" = "('open', None)" ] && [ "$(unticked 32)" -eq 1 ] && ok "a line past the end of the file: stays open" || ko "32 ($(state 32))"
[ "$(state 33)" = "('open', None)" ] && [ "$(unticked 33)" -eq 1 ] && ok "no reference at all: stays open" || ko "33 ($(state 33))"
[ "$(state 34)" = "('open', None)" ] && [ "$(unticked 34)" -eq 1 ] && field 34 'x["body"]' | grep -q -- '- \[x\] `n1`' && ok "mixed issue: the verified note is ticked, the other is not settled, the issue stays open" || ko "34 ($(state 34) $(field 34 'x["body"]'))"
[ "$(state 35)" = "('open', None)" ] && [ "$(unticked 35)" -eq 2 ] && ok "a range that runs past the end of the file: stays open" || ko "35 ($(state 35))"
[ "$(state 36)" = "('open', None)" ] && [ "$(unticked 36)" -eq 2 ] && ok "a missing ADR, a path with .. and line 0: stay open" || ko "36 ($(state 36))"
[ "$(state 39)" = "('open', None)" ] && field 39 'x["comments"][-1]["body"]' | grep -q 'stays open: a human closes it as not planned' && ok "a FINDING with NOT A DEFECT and a valid reference stays open (a human closes it)" || ko "39 ($(state 39))"
[ "$(state 40)" = "('closed', 'completed')" ] && ok "the batch issue still closes last" || ko "40 ($(state 40))"

before="$(cat "$tmp/st/issues.json")"
: >"$tmp/st/log"
run 170
[ "$rc" -eq 0 ] && [ ! -s "$tmp/st/log" ] && [ "$before" = "$(cat "$tmp/st/issues.json")" ] && ok "a second run changes nothing (idempotent)" || ko "idempotent (rc=$rc)"

finish
