#!/usr/bin/env bash
# Tests for `scripts/review-autofix.py record` (ADR-0015 point 8): an attempt is recorded for (pull request, finding, head SHA)
# when the run ENDS, in the hidden state of the bot's status comment, and only when it happened. A run whose head moved,
# that was cancelled or that the infrastructure aborted is SUPERSEDED and consumes nothing. Also the shape of the
# workflow that calls it (review-autofix.yml: the end job, no write before the fixer runs, no pull request code).
# Usage: scripts/tests/review-autofix-record_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
tool="$root/scripts/review-autofix.py"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin" "$tmp/st"

# Fake gh: GET answers from fixtures in $ST (pull request, comments); every write is logged, and a body is kept.
cat >"$tmp/bin/gh" <<'EOF'
#!/usr/bin/env python3
import os, sys
st = os.environ["ST"]
if os.environ.get("FAKE_GH_FAIL"):
    sys.stderr.write("gh: boom\n"); sys.exit(1)
args = sys.argv[1:]
method, endpoint, i = "GET", None, 1
while i < len(args):
    if args[i] == "-X": method = args[i + 1]; i += 2
    elif args[i] in ("-f", "-F"):
        if args[i + 1].startswith("body="): open(os.path.join(st, "body-last"), "w").write(args[i + 1][5:])
        i += 2
    elif args[i] == "--paginate": i += 1
    else: endpoint = args[i]; i += 1
if method != "GET":
    if os.environ.get("FAKE_GH_FAIL_WRITE") and os.environ["FAKE_GH_FAIL_WRITE"] in endpoint:
        sys.stderr.write("gh: refused\n"); sys.exit(1)
    open(os.path.join(st, "log"), "a").write("%s %s %s\n" % (method, endpoint, " ".join(args).replace("\n", " / ")))
    print("{}"); sys.exit(0)
path = endpoint.partition("?")[0]
name = "pr.json" if path.endswith("/pulls/7") else "comments.json" if path.endswith("/issues/7/comments") else None
if name is None:
    sys.stderr.write("unexpected GET %s\n" % endpoint); sys.exit(2)
p = os.path.join(st, name)
print(open(p).read() if os.path.exists(p) else "[]")
EOF
chmod +x "$tmp/bin/gh"

HEAD=1111111111111111111111111111111111111111
MOVED=2222222222222222222222222222222222222222
OLD=4444444444444444444444444444444444444444

pr_head() { printf '{"number":7,"state":"open","head":{"ref":"fix/solve/x","sha":"%s"},"base":{"ref":"develop"}}' "$1" >"$tmp/st/pr.json"; }
# status STATE_JSON [login]: the comments of the pull request: one status comment (id 55).
status() {
  python3 -I - "$tmp/st/comments.json" "${2:-github-actions[bot]}" "$1" <<'EOF2'
import json, sys
p, login, state = sys.argv[1:4]
json.dump([{"id": 55, "user": {"login": login, "type": "Bot" if login.endswith("[bot]") else "User"}, "created_at": "2026-11-14T11:00:00Z",
            "body": "<!-- 226os-autofix-status -->\nstatus\n<!-- autofix-state " + state + " -->"}], open(p, "w"))
EOF2
}
ROWS='[{"id":"f1","where":"src/a.go:3","class":"prose","state":"READY","fact":"no attempt yet","next":"the fixer runs","who":"pipeline"},{"id":"f2","where":"src/b.go:3","class":"prose","state":"READY","fact":"no attempt yet","next":"x","who":"pipeline"}]'
fresh() { status '{"v":1,"sha":"'$HEAD'","run":"90","attempts":{},"rows":'"$ROWS"'}'; pr_head "$HEAD"; }

# record [VAR=value]...: sets rc, out, and attempts (the attempts of f1 in the written state, "" when nothing was written).
record() {
  : >"$tmp/st/log"; rm -f "$tmp/st/body-last"
  rc=0
  out="$(env PATH="$tmp/bin:$PATH" ST="$tmp/st" REPO=o/r PR_NUMBER=7 RUN_ID=99 HEAD_SHA="$HEAD" FINDINGS=f1 AUTO_FIX=on \
    FIX_RESULT=success OUTCOME=fixed PUSH=pushed "$@" python3 -I "$tool" record 2>&1)" || rc=$?
  attempts="$(python3 -I -c '
import json, re, sys
try: body = open(sys.argv[1]).read()
except OSError: sys.exit(0)
print(json.dumps(json.loads(re.search(r"<!-- autofix-state (\{.*?\}) -->", body, re.S).group(1))["attempts"].get("f1", {}), sort_keys=True))' "$tmp/st/body-last" 2>/dev/null || true)"
  body="$(cat "$tmp/st/body-last" 2>/dev/null || true)"
}
wrote() { [ -s "$tmp/st/log" ]; }

# ---- an attempt that happened is recorded, with the head it ran on ----
fresh
record
[ "$rc" -eq 0 ] && [ "$attempts" = '{"'$HEAD'": "99"}' ] && grep -q '^PATCH .*issues/comments/55' "$tmp/st/log" && ok "a pushed fix is recorded for (finding, head SHA, run), in the status comment, edited in place" || ko "pushed (rc=$rc attempts=$attempts $out)"
grep -q '| f1 | .*PUSHED' <<<"$body" && grep -q '| f2 | .*READY' <<<"$body" && ok "only the finding the run was for changes row" || ko "rows ($body)"
grep -q 'labels\[\]=autofix-1' "$tmp/st/log" && ok "a pushed fix leaves the provenance label autofix-1 (review-gate reads it; it is never counted)" || ko "label ($(cat "$tmp/st/log"))"
record
[ "$attempts" = '{"'$HEAD'": "99"}' ] && ok "recording the same run twice changes nothing (one entry per finding and head)" || ko "idempotent ($attempts)"

fresh
record PUSH=owner
[ "$attempts" = '{"'$HEAD'": "99"}' ] && grep -q '| f1 | .*NEEDS_OWNER' <<<"$body" && ! grep -q 'labels\[\]=autofix-' "$tmp/st/log" && ok "the guard refused the patch: the attempt happened, the finding goes to the owner" || ko "guard refusal ($attempts $body)"

fresh
record OUTCOME=owner PUSH=
[ "$attempts" = '{"'$HEAD'": "99"}' ] && grep -q '| f1 | .*NEEDS_OWNER' <<<"$body" && ok "the fixer left it to the owner: recorded, NEEDS_OWNER" || ko "fixer owner ($attempts $body)"

fresh
record OUTCOME=nothing PUSH=
[ "$attempts" = '{"'$HEAD'": "99"}' ] && grep -q '| f1 | .*FAILED (attempt 1)' <<<"$body" && ok "the fixer ran on the current head and made no commit: attempt 1 used, FAILED" || ko "nothing ($attempts $body)"

# ---- a run that did not happen consumes nothing ----
for case in "FIX_RESULT=cancelled OUTCOME= PUSH=" "FIX_RESULT=failure OUTCOME= PUSH=" "FIX_RESULT=skipped OUTCOME= PUSH=" "PUSH=moved" "PUSH=notoken" "PUSH=" "OUTCOME="; do
  fresh
  # shellcheck disable=SC2086
  record $case
  [ "$rc" -eq 0 ] && [ "$attempts" = '{}' ] && grep -q '| f1 | .*SUPERSEDED' <<<"$body" && ok "SUPERSEDED, no attempt used: $case" || ko "superseded $case (rc=$rc attempts=$attempts $out)"
done
fresh; pr_head "$MOVED"
record OUTCOME=nothing PUSH=
[ "$attempts" = '{}' ] && grep -qi 'head moved' <<<"$body" && ok "the head moved while the fixer ran (no push): SUPERSEDED, nothing used" || ko "moved head ($attempts $body)"
fresh; pr_head "$MOVED"
record
[ "$attempts" = '{"'$HEAD'": "99"}' ] && ok "after OUR push the head has moved by design: the attempt is still recorded" || ko "own push ($attempts)"

# ---- a second attempt on a new head, then the cap by head ----
status '{"v":1,"sha":"'$HEAD'","run":"90","attempts":{"f1":{"'$OLD'":"11"}},"rows":'"$ROWS"'}'; pr_head "$HEAD"
record
[ "$attempts" = '{"'$HEAD'": "99", "'$OLD'": "11"}' ] && grep -q 'labels\[\]=autofix-2' "$tmp/st/log" && ok "a second head is a second attempt (label autofix-2 as provenance)" || ko "second ($attempts)"

# ---- modes, forged state, failures ----
fresh
record AUTO_FIX=report
! wrote && grep -q 'NOT RUN' <<<"$out" && ok "report mode writes nothing" || ko "report ($out)"
status '{"v":1,"attempts":{}}' someone; pr_head "$HEAD"
record
[ "$rc" -eq 0 ] && grep -q '^POST .*issues/7/comments' "$tmp/st/log" && ! grep -q 'comments/55' "$tmp/st/log" && ok "a human's status comment is not edited: the bot writes its own" || ko "human comment ($(cat "$tmp/st/log"))"
fresh
record FAKE_GH_FAIL=1 OUTCOME=nothing PUSH=
[ "$rc" -ne 0 ] && grep -q 'FAILED' <<<"$out" && ok "a GitHub error is never exit 0 once a run must be recorded: FAILED, exit 1, nothing recorded" || ko "gh failure (rc=$rc $out)"
record HEAD_SHA=nothex
[ "$rc" -eq 0 ] && ! wrote && grep -q 'NOT RUN' <<<"$out" && ok "an unreadable head is NOT RUN" || ko "bad head ($out)"
record 'FINDINGS=f1;rm,f2'
[ "$rc" -eq 0 ] && grep -q '"f1-rm"' <<<"$body" && ok "finding ids are reduced to safe characters" || ko "ids ($body)"

# ---- review follow-ups: aborted runs are recorded too, so a run that consumes nothing is still bounded ----
aborted_of() { python3 -I -c '
import json, re, sys
b = open(sys.argv[1]).read()
print(json.dumps(json.loads(re.search(r"<!-- autofix-state (\{.*?\}) -->", b, re.S).group(1)).get("aborted", {}).get("f1", {}), sort_keys=True))' "$tmp/st/body-last"; }
fresh
record FIX_RESULT=cancelled OUTCOME= PUSH=
[ "$(aborted_of)" = '{"'$HEAD'": ["99"]}' ] && ok "an aborted run is recorded per (finding, head): no attempt, one abort" || ko "aborted ($(aborted_of))"
record FIX_RESULT=cancelled OUTCOME= PUSH=
[ "$(aborted_of)" = '{"'$HEAD'": ["99"]}' ] && ok "the same aborted run recorded twice counts once" || ko "aborted twice ($(aborted_of))"
fresh
record
[ "$(aborted_of)" = '{}' ] && ok "a run that happened is not an abort" || ko "not aborted ($(aborted_of))"

# ---- the record step fails closed after a push, open otherwise ----
fresh
record FAKE_GH_FAIL_WRITE=comments
[ "$rc" -ne 0 ] && grep -q 'labels\[\]=autofix-stopped' "$tmp/st/log" && grep -q 'FAILED' <<<"$out" && ok "a pushed patch whose status cannot be written: FAILED, exit 1, autofix-stopped" || ko "pushed write failure (rc=$rc $out / $(cat "$tmp/st/log"))"
fresh
record FAKE_GH_FAIL_WRITE=comments OUTCOME=nothing PUSH=
[ "$rc" -ne 0 ] && grep -q 'labels\[\]=autofix-stopped' "$tmp/st/log" && grep -q 'FAILED' <<<"$out" && ok "no patch pushed: a write failure still fails closed (an abort or an attempt would be lost): FAILED, exit 1, autofix-stopped" || ko "nothing write failure (rc=$rc $out)"
# A state too large to keep is never truncated: fail closed after a push.
python3 -I - "$tmp/st/comments.json" <<'EOT'
import json, sys
rows = [{"id": "f%d" % i, "where": "w", "class": "prose", "state": "READY", "fact": "x" * 290, "next": "y" * 290, "who": "z" * 290} for i in range(40)]
state = {"v": 1, "sha": "1" * 40, "run": "90", "attempts": {}, "rows": rows}
json.dump([{"id": 55, "user": {"login": "github-actions[bot]", "type": "Bot"}, "created_at": "2026-11-14T11:00:00Z",
            "body": "<!-- 226os-autofix-status -->\nx\n<!-- autofix-state " + json.dumps(state) + " -->"}], open(sys.argv[1], "w"))
EOT
record
[ "$rc" -ne 0 ] && grep -qi 'too large' <<<"$out" && grep -q 'labels\[\]=autofix-stopped' "$tmp/st/log" && ok "a state too large to keep fails closed after a push (autofix-stopped), never truncated" || ko "large state (rc=$rc $out)"

# ---- routing to the owner is per finding: only the findings the guard or the fixer names ----
ROWS2='[{"id":"f1","where":"src/a.go:3","class":"prose","route":"fixer","state":"READY","fact":"","next":"","who":""},{"id":"f2","where":"src/b.go:9","class":"prose","route":"fixer","state":"READY","fact":"","next":"","who":""}]'
two() { status '{"v":1,"sha":"'$HEAD'","run":"90","attempts":{},"rows":'"$ROWS2"'}'; pr_head "$HEAD"; }
routes_of() { python3 -I -c '
import json, re, sys
st = json.loads(re.search(r"<!-- autofix-state (\{.*?\}) -->", open(sys.argv[1]).read(), re.S).group(1))
print(",".join(sorted("%s:%s" % (k, v["code"]) for k, v in st["routes"].items())), "|", ",".join(sorted(st["attempts"])))' "$tmp/st/body-last"; }
two
record FINDINGS=f1,f2 PUSH=owner REFUSED=src/b.go
[ "$(routes_of)" = "f2:guard-refused | f1,f2" ] && grep -q '| f1 | .*FAILED (attempt 1)' <<<"$body" && grep -q '| f2 | .*owner (guard-refused)' <<<"$body" && ok "a guard refusal caused by f2's file: f2 goes to the owner, f1 consumed an attempt and is retried on the next head" || ko "guard attribution ($(routes_of) $body)"
two
record FINDINGS=f1,f2 PUSH=owner REFUSED=docs/elsewhere.go
[ "$(routes_of)" = "f1:guard-refused,f2:guard-refused | f1,f2" ] && grep -qi 'names no finding' <<<"$out$body" && ok "a refusal that names no finding's file: every finding of the run goes to the owner, and it is said" || ko "unattributable ($(routes_of) $out)"
two
record FINDINGS=f1,f2 PUSH=owner
[ "$(routes_of)" = "f1:guard-refused,f2:guard-refused | f1,f2" ] && ok "a refusal with no path (identity, link...): all ids to the owner" || ko "no path ($(routes_of))"
two
record FINDINGS=f1,f2 OWNER_IDS=f1
[ "$(routes_of)" = "f1:fixer-owner | f1,f2" ] && grep -q '| f1 | .*owner (fixer-owner)' <<<"$body" && grep -q '| f2 | .*PUSHED' <<<"$body" && ok "the fixer answered OWNER for f1 and FIXED f2: f1 goes to the owner, f2 is pushed" || ko "fixer owner ids ($(routes_of) $body)"
two
record FINDINGS=f1,f2 OUTCOME=owner PUSH= OWNER_IDS=f2
[ "$(routes_of)" = "f2:fixer-owner | f1,f2" ] && grep -q '| f1 | .*FAILED' <<<"$body" && ok "fixer outcome owner naming f2: only f2 goes to the owner" || ko "outcome owner ids ($(routes_of) $body)"

# ---- pending attempts (written by decide before the fixer) are settled here ----
status '{"v":1,"sha":"'$HEAD'","run":"90","pending":{"f1":{"'$HEAD'":"31"}},"attempts":{},"rows":'"$ROWS"'}'; pr_head "$HEAD"
record
python3 -I -c '
import json, re, sys
st = json.loads(re.search(r"<!-- autofix-state (\{.*?\}) -->", open(sys.argv[1]).read(), re.S).group(1))
assert st["pending"] in ({}, {"f1": {}}) and list(st["attempts"]["f1"]) == ["'$HEAD'"], st' "$tmp/st/body-last" && ok "a completed run settles its pending entry into an attempt" || ko "pending settled ($body)"
status '{"v":1,"sha":"'$HEAD'","run":"90","pending":{"f1":{"'$HEAD'":"31"}},"attempts":{},"rows":'"$ROWS"'}'
record FIX_RESULT=cancelled OUTCOME= PUSH=
python3 -I -c '
import json, re, sys
st = json.loads(re.search(r"<!-- autofix-state (\{.*?\}) -->", open(sys.argv[1]).read(), re.S).group(1))
assert st["pending"] in ({}, {"f1": {}}) and st["attempts"] == {"f1": {}} and st["aborted"]["f1"]["'$HEAD'"] == ["99"], st' "$tmp/st/body-last" && ok "an aborted run clears its pending entry and counts as an abort, not an attempt" || ko "pending aborted ($body)"
# No push token: the row says NOT RUN instead of leaving "fix queued".
status '{"v":1,"sha":"'$HEAD'","run":"90","pending":{"f1":{"'$HEAD'":"31"}},"attempts":{},"rows":'"$ROWS"'}'
record PREFLIGHT_TOKEN=false FIX_RESULT=skipped OUTCOME= PUSH=
grep -q '| f1 | .*NOT RUN' <<<"$body" && grep -qi 'REVIEW_AGENT_TOKEN missing' <<<"$body" && [ "$attempts" = '{}' ] && ok "no push token: the row says NOT RUN (token missing), no attempt and no abort" || ko "no token row ($attempts $body)"
python3 -I -c '
import json, re, sys
st = json.loads(re.search(r"<!-- autofix-state (\{.*?\}) -->", sys.argv[1], re.S).group(1))
assert st["pending"] in ({}, {"f1": {}}) and not st["aborted"].get("f1"), st' "$body" && ok "no push token: the pending entry is cleared" || ko "no token pending"

# ---- the state is read only from a bot, at its expected place ----
python3 -I - "$tmp/st/comments.json" <<'EOT'
import json, sys
st = '{"attempts":{"f1":{"%s":"11"}}}' % ("4" * 40)
def c(i, body, typ="Bot"): return {"id": i, "user": {"login": "github-actions[bot]", "type": typ}, "created_at": "2026-11-14T11:00:00Z", "body": body}
json.dump([c(55, "<!-- 226os-autofix-status -->\nx\n<!-- autofix-state " + st + " -->", "User")], open(sys.argv[1], "w"))
EOT
pr_head "$HEAD"; record
[ "$attempts" = '{"'$HEAD'": "99"}' ] && ok "a comment of a login that is not of type Bot is ignored (only our own attempt is written)" || ko "user type ($attempts)"
python3 -I - "$tmp/st/comments.json" <<'EOT'
import json, sys
st = '{"attempts":{"f1":{"%s":"11"}}}' % ("4" * 40)
body = "<!-- 226os-autofix-status -->\nx\n<!-- autofix-state " + st + " -->\nmore text after it"
json.dump([{"id": 55, "user": {"login": "github-actions[bot]", "type": "Bot"}, "created_at": "2026-11-14T11:00:00Z", "body": body}], open(sys.argv[1], "w"))
EOT
record
[ "$attempts" = '{"'$HEAD'": "99"}' ] && ok "a state marker that is not the last line is not read" || ko "marker position ($attempts)"

# ---- the workflow shape (ADR-0013: the job that writes runs no pull request code) ----
wf="$root/.github/workflows/review-autofix.yml"
python3 -I - "$wf" <<'EOF' && ok "review-autofix.yml: the end job records the attempt (always, default branch only, minimal permissions), nothing is labelled before the fixer runs" || ko "review-autofix.yml shape"
import re, sys
s = open(sys.argv[1]).read()
decide = s[s.index("\n  decide:"):s.index("\n  fix:")]
label = s[s.index("\n  label:"):]
assert "autofix-$" not in decide and "autofix-%" not in decide
assert "if: always() && needs.decide.outputs.run == 'true'" in label, "the end job must run after every outcome of the fixer"
assert "review-autofix.py record" in label
for var in ("FIX_RESULT: ${{ needs.fix.result }}", "OUTCOME: ${{ needs.fix.outputs.outcome }}", "PUSH: ${{ needs.publish.outputs.push }}"):
    assert var in label, var
assert "ref:" not in label, "the end job must not check out the pull request branch"
assert "contents: read" in label and "pull-requests: write" in label and "secrets." not in label
assert "REVIEW_AGENT_TOKEN" not in label
# No token: refuse BEFORE the model (a trusted preflight job in the environment, the untrusted fix job needs it).
pre = s[s.index("\n  preflight:"):s.index("\n  fix:")]
assert "environment: review-agent" in pre and "secrets.REVIEW_AGENT_TOKEN" in pre and "token=" in pre
fix = s[s.index("\n  fix:"):s.index("\n  publish:")]
assert "preflight" in fix.split("steps:")[0] and "needs.preflight.outputs.token == 'true'" in fix.split("steps:")[0]
assert "REVIEW_AGENT_TOKEN" not in fix
assert "REFUSED: ${{ needs.publish.outputs.refused }}" in label and "OWNER_IDS: ${{ needs.fix.outputs.owners }}" in label
assert "refused: ${{ steps.push.outputs.refused }}" in s and "owners: ${{ steps.patch.outputs.owners }}" in s
assert "needs.preflight.outputs.token == 'true'" not in label.split("steps:")[0], "the end job must also run when the token is missing, to say NOT RUN"
assert "PREFLIGHT_TOKEN: ${{ needs.preflight.outputs.token }}" in label
assert "give the attempt back" not in label.lower() and "DELETE" not in label, "no attempt is written before the run, so none is given back"
# values reach the script through env, never interpolated in run:
for run in re.findall(r"run: \|\n((?:\s{10,}.*\n)+)", label):
    assert "${{" not in run, run
EOF

finish
