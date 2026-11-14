#!/usr/bin/env bash
# Tests for scripts/close-merged-issues.py with a stateful fake `gh`: a pull request merged into develop closes the
# issues its BODY names with a closing keyword (GitHub only does it on the default branch), with one comment, once.
# Pull requests, other repositories, closed issues and `review-batch` issues are left alone; a `needs-owner` issue gets a
# comment and stays open.
# Usage: scripts/tests/close-merged-issues_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
tool="$root/scripts/close-merged-issues.py"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin" "$tmp/st"

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
parts = endpoint.split("?")[0].split("/")
issues = json.load(open(os.path.join(st, "issues.json")))
def save(): json.dump(issues, open(os.path.join(st, "issues.json"), "w"))
def find(n):
    hit = [x for x in issues if x["number"] == int(n)]
    if not hit:
        sys.stderr.write("HTTP 404\n"); sys.exit(1)
    return hit[0]
if method != "GET":
    open(os.path.join(st, "log"), "a").write("%s %s\n" % (method, "/".join(parts)))
if method == "GET" and parts[-2] == "pulls":
    print(open(os.path.join(st, "pr.json")).read())
elif method == "GET" and parts[-2] == "issues":
    x = dict(find(parts[-1])); x.pop("comments"); print(json.dumps(x))
elif method == "GET" and parts[-1] == "comments":
    print(json.dumps(find(parts[-2])["comments"]))
elif method == "POST" and parts[-1] == "comments":
    find(parts[-2])["comments"].append({"user": {"login": "github-actions[bot]"}, "body": fields["body"]}); save(); print("{}")
elif method == "PATCH" and parts[-2] == "issues":
    x = find(parts[-1])
    for k in ("state", "state_reason"):
        if k in fields: x[k] = fields[k]
    save(); print("{}")
else:
    sys.stderr.write("unexpected %s %s\n" % (method, endpoint)); sys.exit(2)
EOF
chmod +x "$tmp/bin/gh"

# setup BODY [BASE] [HEAD_REPO] [MERGED_AT]
setup() {
  python3 -I - "$tmp/st" "$1" "${2:-develop}" "${3:-o/r}" "${4:-2026-11-14T12:00:00Z}" "${5:-fix/solve/x}" <<'EOF'
import json, os, sys
st, body, base, head_repo, merged, head_ref = sys.argv[1:7]
def issue(n, state="open", labels=(), pr=False):
    x = {"number": n, "state": state, "state_reason": None, "labels": [{"name": l} for l in labels],
         "comments": [{"user": {"login": "dev"}, "body": "a human comment"}]}
    if pr: x["pull_request"] = {"url": "x"}
    return x
issues = [issue(n) for n in (1, 2, 3, 4, 5, 6, 7, 8, 9)] + [issue(10, pr=True), issue(11, "closed"),
          issue(12, labels=("review-batch",)), issue(13, labels=("needs-owner",))]
json.dump(issues, open(os.path.join(st, "issues.json"), "w"))
json.dump({"number": 170, "merged_at": None if merged == "none" else merged, "merge_commit_sha": "abcdef1234567890abcdef1234567890abcdef12",
           "base": {"ref": base}, "head": {"ref": head_ref, "repo": {"full_name": head_repo}}, "body": body},
          open(os.path.join(st, "pr.json"), "w"))
EOF
  : >"$tmp/st/log"
}
run() {
  rc=0
  out="$(PATH="$tmp/bin:$PATH" ST="$tmp/st" REPO=o/r PR_NUMBER=170 python3 -I "$tool" 2>&1)" || rc=$?
}
state() { python3 -I -c 'import json,sys; x=[i for i in json.load(open(sys.argv[1])) if i["number"]==int(sys.argv[2])][0]; print(x["state"], x["state_reason"], sum(1 for c in x["comments"] if c["user"]["login"].endswith("[bot]")))' "$tmp/st/issues.json" "$1"; }
last_comment() { python3 -I -c 'import json,sys; x=[i for i in json.load(open(sys.argv[1])) if i["number"]==int(sys.argv[2])][0]; print(x["comments"][-1]["body"])' "$tmp/st/issues.json" "$1"; }

setup "Closes #1, fixes #2 and Resolved #3.
close #4; FIX #5 / fixed: #6 / resolves #7 / resolve #8 / closed #9
Also fixes #10 (a pull request), closes #11 (closed), closes #12 (batch), fixes #13 (owner).
See #14 and fixes other/repo#15 and fixes #404."
run
[ "$rc" -eq 0 ] && ok "runs (rc=0)" || ko "rc=$rc ($out)"
all=1
for n in 1 2 3 4 5 6 7 8 9; do [ "$(state "$n")" = "closed completed 1" ] || { all=0; ko "keyword variant: #$n is $(state "$n")"; }; done
[ "$all" -eq 1 ] && ok "every keyword variant (close, closes, closed, fix, fixes, fixed, resolve, resolves, resolved; any case; a colon) closes the issue as completed with one comment" || true
last_comment 1 | grep -q 'Fixed in #170 (merged into develop, abcdef1)' && ok "the comment names the pull request and the short merge sha" || ko "comment ($(last_comment 1))"
[ "$(state 10)" = "open None 0" ] && ok "a pull request number is ignored" || ko "pr ($(state 10))"
[ "$(state 11)" = "closed None 0" ] && ok "an already closed issue is untouched" || ko "closed ($(state 11))"
[ "$(state 12)" = "open None 0" ] && ok "a review-batch issue is left to the settle step" || ko "batch ($(state 12))"
[ "$(state 13)" = "open None 1" ] && last_comment 13 | grep -qi 'needs-owner' && ok "a needs-owner issue gets a comment and stays open" || ko "needs-owner ($(state 13))"
! grep -q 'issues/14\|issues/15\|issues/404' "$tmp/st/log" && ok "a plain reference, another repository and a missing issue are not written" || ko "others ($(cat "$tmp/st/log"))"
! grep -q 'PATCH .*/comments' "$tmp/st/log" && ok "no comment is ever edited" || ko "comment edit"

before="$(cat "$tmp/st/issues.json")"
: >"$tmp/st/log"
run
[ "$rc" -eq 0 ] && [ ! -s "$tmp/st/log" ] && [ "$before" = "$(cat "$tmp/st/issues.json")" ] && ok "a second run changes nothing (idempotent)" || ko "idempotent ($(cat "$tmp/st/log"))"

setup "fixes #1" main
run
[ ! -s "$tmp/st/log" ] && ok "a merge into main is left to GitHub" || ko "main ($(cat "$tmp/st/log"))"
setup "fixes #1" develop some/fork
run
[ ! -s "$tmp/st/log" ] && ok "a pull request from another repository closes nothing" || ko "fork ($(cat "$tmp/st/log"))"
setup "fixes #1" develop o/r none
run
[ ! -s "$tmp/st/log" ] && ok "a closed, unmerged pull request closes nothing" || ko "unmerged ($(cat "$tmp/st/log"))"

setup "| #1 | #1 | FIXED | fixes #2 |" develop o/r 2026-11-14T12:00:00Z fix/review/batch-2026-11-14
run
[ "$rc" -eq 0 ] && [ ! -s "$tmp/st/log" ] && ok "a batch fix pull request closes nothing here (its sources are settled by findings.py settle)" || ko "batch pr ($(cat "$tmp/st/log"))"

setup "fixes #1"
FAKE_GH_FAIL=1 run
[ "$rc" -ne 0 ] && grep -q 'FAILED' <<<"$out" && ok "a GitHub error is FAILED (exit 1)" || ko "gh failure (rc=$rc $out)"

finish
