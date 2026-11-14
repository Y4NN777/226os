#!/usr/bin/env bash
# Tests for `scripts/findings.py settle <pr>` with a stateful fake `gh`: when a batch fix pull request
# (fix/review/batch-*) is merged into develop, the results table of its body settles each source. FIXED and FIXED
# ELSEWHERE close a finding issue as completed, NOT A DEFECT and DECIDED close it as not planned; a note's box is ticked
# with the decision; OWNER and AFTER FREEZE stay open. One comment per touched source, once. A notes issue whose boxes are
# all ticked is closed. The batch issue closes last. A second run changes nothing; no human comment is edited; no mention.
# Usage: scripts/tests/findings-settle_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
tool="$root/scripts/findings.py"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin" "$tmp/st"

# Fake gh: issues in $ST/issues.json (each with a "comments" list), the pull request in $ST/pr.json. Writes are logged
# to $ST/log, one line each: METHOD ENDPOINT.
cat >"$tmp/bin/gh" <<'EOF'
#!/usr/bin/env python3
import json, os, sys
st = os.environ["ST"]
if os.environ.get("FAKE_GH_FAIL"):
    sys.stderr.write("gh: boom\n"); sys.exit(1)
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
if method == "GET" and parts[-2] == "pulls":
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
EOF
chmod +x "$tmp/bin/gh"

setup() {
  python3 -I - "$tmp/st" <<'EOF'
import json, os, sys
st = sys.argv[1]
def finding(n, title):
    return {"number": n, "state": "open", "state_reason": None, "title": title, "labels": [{"name": "finding"}],
            "body": "<!-- finding:PR90#f%d -->" % n, "comments": [{"user": {"login": "dev"}, "body": "human note @someone"}]}
notes20 = ("<!-- finding:PR91#notes -->\nNotes of PR #91.\n\n"
           "- [ ] `n1` [INFERRED] `src/a.go:1` guess one\n- [ ] `n2` [OBSERVED] `src/b.go:2` remark two")
notes21 = ("<!-- finding:PR92#notes -->\nNotes of PR #92.\n\n"
           "- [ ] `n3` [INFERRED] `src/c.go:3` later\n- [ ] `n4` [OBSERVED] `src/d.go:4` not in the batch")
issues = [finding(12, "a"), finding(13, "b"), finding(14, "c"), finding(15, "d"),
          {"number": 20, "state": "open", "state_reason": None, "title": "PR91 notes", "labels": [{"name": "kind:note"}], "body": notes20, "comments": []},
          {"number": 21, "state": "open", "state_reason": None, "title": "PR92 notes", "labels": [{"name": "kind:note"}], "body": notes21, "comments": []},
          {"number": 40, "state": "open", "state_reason": None, "title": "Review batch 2026-11-14", "labels": [{"name": "review-batch"}],
           "body": "<!-- review-batch -->\n# Review batch 2026-11-14\n", "comments": []}]
json.dump(issues, open(os.path.join(st, "issues.json"), "w"))
body = "\n".join([
    "Fix pull request for the review batch #40.", "", "Batch: #40", "Sources: #12 #13 #14 #15 #20 #21", "",
    "| Item | Source | Decision | Reason |", "|---|---|---|---|",
    "| #12 | #12 | FIXED | guard added at src/a.go:9 |",
    "| #13 | #13 | NOT A DEFECT | the check runs before \\| the write, @someone |",
    "| #14 | #14 | OWNER | needs a migration |",
    "| #15 | #15 | FIXED ELSEWHERE | fixed by #150 |",
    "| #20/n1 | #20 | DECIDED | REVIEW_CONTEXT says so |",
    "| #20/n2 | #20 | FIXED | test added |",
    "| #21/n3 | #21 | AFTER FREEZE | not a gate |"])
json.dump({"number": 163, "merged_at": "2026-11-14T12:00:00Z", "base": {"ref": "develop"},
           "head": {"ref": "fix/review/batch-2026-11-14", "repo": {"full_name": "o/r"}}, "body": body},
          open(os.path.join(st, "pr.json"), "w"))
EOF
  : >"$tmp/st/log"
}

run() {
  rc=0
  out="$(PATH="$tmp/bin:$PATH" ST="$tmp/st" REPO=o/r python3 -I "$tool" settle "$@" 2>&1)" || rc=$?
}
# field NUMBER EXPR: a Python expression over the issue `x`.
field() { python3 -I -c 'import json,sys; x=[i for i in json.load(open(sys.argv[1])) if i["number"]==int(sys.argv[2])][0]; print(eval(sys.argv[3]))' "$tmp/st/issues.json" "$1" "$2"; }
bot_comments() { field "$1" 'sum(1 for c in x["comments"] if c["user"]["login"].endswith("[bot]"))'; }

setup
run 163
[ "$rc" -eq 0 ] && ok "settle runs (rc=0)" || ko "settle rc=$rc ($out)"
[ "$(field 12 'x["state"], x["state_reason"]')" = "('closed', 'completed')" ] && ok "FIXED: the finding issue is closed as completed" || ko "fixed ($(field 12 'x["state"], x["state_reason"]'))"
[ "$(field 13 'x["state"], x["state_reason"]')" = "('closed', 'not_planned')" ] && ok "NOT A DEFECT: the finding issue is closed as not planned" || ko "not a defect ($(field 13 'x["state"], x["state_reason"]'))"
[ "$(field 15 'x["state"], x["state_reason"]')" = "('closed', 'completed')" ] && ok "FIXED ELSEWHERE: closed as completed" || ko "fixed elsewhere"
[ "$(field 14 'x["state"]')" = "open" ] && field 14 'x["comments"][-1]["body"]' | grep -q 'OWNER' && ok "OWNER: the finding stays open, with a comment saying so" || ko "owner ($(field 14 'x["comments"]'))"
field 12 'x["comments"][-1]["body"]' | grep -q 'Settled by #163 (review batch #40): FIXED' && ok "the settling comment names the pull request, the batch and the decision" || ko "comment text ($(field 12 'x["comments"]'))"
field 13 'x["comments"][-1]["body"]' | grep -q 'the check runs before | the write' && ok "the reason is carried, pipes unescaped" || ko "reason ($(field 13 'x["comments"][-1]["body"]'))"
! grep -q '@someone' <<<"$(field 13 'x["comments"][-1]["body"]')" && ok "no mention in a comment" || ko "mention"
[ "$(field 12 'x["comments"][0]["body"]')" = "human note @someone" ] && ! grep -q '^PATCH .*/comments' "$tmp/st/log" && ok "a human comment is never edited" || ko "human comment"
field 20 'x["body"]' | grep -q -- '- \[x\] `n1` .* — decided: REVIEW_CONTEXT says so' && field 20 'x["body"]' | grep -q -- '- \[x\] `n2` .* — fixed in #163' && ok "notes: the boxes are ticked with the decision" || ko "notes tick ($(field 20 'x["body"]'))"
[ "$(field 20 'x["state"], x["state_reason"]')" = "('closed', 'completed')" ] && ok "a notes issue whose boxes are all ticked is closed as completed" || ko "notes close ($(field 20 'x["state"]'))"
[ "$(bot_comments 20)" -eq 1 ] && ok "ONE comment per touched source issue (two notes, one comment)" || ko "notes comments ($(bot_comments 20))"
field 21 'x["body"]' | grep -q -- '- \[ \] `n3`' && [ "$(field 21 'x["state"]')" = "open" ] && ok "AFTER FREEZE: the box stays unticked and the issue open" || ko "after freeze ($(field 21 'x["body"]'))"
[ "$(field 40 'x["state"], x["state_reason"]')" = "('closed', 'completed')" ] && ok "the batch issue is closed" || ko "batch close ($(field 40 'x["state"]'))"
last="$(tail -1 "$tmp/st/log")"
grep -q 'PATCH repos/o/r/issues/40 ' <<<"$last" && ok "the batch issue closes last, after its sources" || ko "batch last ($last)"

before="$(cat "$tmp/st/issues.json")"
: >"$tmp/st/log"
run 163
[ "$rc" -eq 0 ] && [ ! -s "$tmp/st/log" ] && [ "$before" = "$(cat "$tmp/st/issues.json")" ] && ok "a second run changes nothing (idempotent)" || ko "idempotent (rc=$rc log=$(cat "$tmp/st/log"))"

setup
python3 -I - "$tmp/st/pr.json" <<'EOF'
import json, sys
d = json.load(open(sys.argv[1])); d["head"]["ref"] = "fix/solve/other"; json.dump(d, open(sys.argv[1], "w"))
EOF
run 163
[ "$rc" -eq 0 ] && [ ! -s "$tmp/st/log" ] && ok "a pull request that is not a batch fix settles nothing" || ko "not a batch ($(cat "$tmp/st/log"))"

setup
python3 -I - "$tmp/st/pr.json" <<'EOF'
import json, sys
d = json.load(open(sys.argv[1])); d["merged_at"] = None; json.dump(d, open(sys.argv[1], "w"))
EOF
run 163
[ "$rc" -eq 0 ] && [ ! -s "$tmp/st/log" ] && ok "a closed, unmerged pull request settles nothing" || ko "unmerged ($(cat "$tmp/st/log"))"

setup
FAKE_GH_FAIL=1 run 163
[ "$rc" -ne 0 ] && grep -q 'FAILED' <<<"$out" && ok "a GitHub error fails closed" || ko "gh failure (rc=$rc $out)"

finish
