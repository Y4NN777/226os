#!/usr/bin/env bash
# Tests for scripts/review-comment.py with a fake `gh`: exactly one review comment per pull request. The marker comment
# is edited by id, one is created when none exists, a stale bot Triage duplicate is minimized (never deleted), another
# bot comment is never touched, and every failure is NOT RUN with exit 0.
# Usage: scripts/tests/review-comment_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
tool="$root/scripts/review-comment.py"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin"

# Fake gh: the comment list from $FX/comments.json; every call is logged (with the --input body) to $FX/calls.log.
cat >"$tmp/bin/gh" <<'EOF'
#!/usr/bin/env bash
[ -z "${FAKE_GH_FAIL:-}" ] || { echo "gh: boom" >&2; exit 1; }
input=""
args=()
while [ "$#" -gt 0 ]; do
  if [ "$1" = "--input" ]; then input="$(cat "$2")"; shift 2; continue; fi
  args+=("$1"); shift
done
printf '%s\n' "${args[*]}" >>"$FX/calls.log"
[ -z "$input" ] || printf 'BODY %s\n' "$input" >>"$FX/calls.log"
case "${args[*]}" in
  "api repos/o/r/issues/7/comments --paginate") cat "$FX/comments.json" ;;
  "api -X POST repos/o/r/issues/7/comments") echo '{"id":999,"node_id":"N999"}' ;;
  "api -X PATCH repos/o/r/issues/comments/"*) echo '{}' ;;
  "api graphql "*) echo '{}' ;;
  *) echo "unexpected: ${args[*]}" >&2; exit 2 ;;
esac
EOF
chmod +x "$tmp/bin/gh"

BOT='github-actions[bot]'
REVIEW='{"body":"Triage: MEDIUM | blast radius: this PR only\n\nTool: PASS (99 files)\n\nFindings:\n- f1 MEDIUM a.go:3 x"}'

run() {
  rc=0
  : >"$tmp/calls.log"
  out="$(PATH="$tmp/bin:$PATH" FX="$tmp" REPO=o/r PR_NUMBER=7 TOOL_LINE="${TOOL:-Tool: PASS (3 files)}" REVIEW_JSON="${JSON:-$REVIEW}" python3 -I "$tool" 2>&1)" || rc=$?
  calls="$(cat "$tmp/calls.log")"
}

comments() { printf '%s' "$1" >"$tmp/comments.json"; }

# No review comment yet (only an unrelated bot comment): one is created, the other bot comment is untouched.
comments "[{\"id\":5,\"node_id\":\"N5\",\"user\":{\"login\":\"$BOT\"},\"body\":\"Owner review requested.\",\"updated_at\":\"2026-11-14T10:00:00Z\"}]"
run
[ "$rc" -eq 0 ] && grep -q '^api -X POST repos/o/r/issues/7/comments$' <<<"$calls" && ok "no review comment yet: one is created" || ko "create (rc=$rc out=$out calls=$calls)"
! grep -q 'comments/5' <<<"$calls" && ! grep -q 'N5' <<<"$calls" && ok "another bot comment is never edited nor minimized" || ko "other bot ($calls)"
body="$(sed -n 's/^BODY //p' <<<"$calls" | python3 -I -c 'import json,sys; print(json.load(sys.stdin)["body"])')"
[ "$(sed -n 1p <<<"$body")" = "Triage: MEDIUM | blast radius: this PR only" ] && [ "$(sed -n 2p <<<"$body")" = "<!-- 226os-ai-review -->" ] && [ "$(sed -n 3p <<<"$body")" = "Tool: PASS (3 files)" ] && ok "line 1 Triage, line 2 marker, line 3 the Tool line of the workflow" || ko "layout ($body)"
! grep -q 'PASS (99 files)' <<<"$body" && ok "a Tool line written by the model is dropped" || ko "model tool line ($body)"

# The marker comment exists, plus an older duplicate Triage comment and another bot comment in between.
comments "[{\"id\":1,\"node_id\":\"N1\",\"user\":{\"login\":\"$BOT\"},\"body\":\"Triage: MEDIUM | blast radius: x\\nold\",\"updated_at\":\"2026-11-14T11:00:00Z\"},{\"id\":2,\"node_id\":\"N2\",\"user\":{\"login\":\"$BOT\"},\"body\":\"Triage: NONE | blast radius: x\\n<!-- 226os-ai-review -->\\nTool: PASS (1 files)\",\"updated_at\":\"2026-11-14T09:00:00Z\"},{\"id\":3,\"node_id\":\"N3\",\"user\":{\"login\":\"$BOT\"},\"body\":\"Owner review requested.\",\"updated_at\":\"2026-11-14T12:00:00Z\"},{\"id\":4,\"node_id\":\"N4\",\"user\":{\"login\":\"Y4NN777\"},\"body\":\"Triage: NONE | blast radius: me\",\"updated_at\":\"2026-11-14T12:00:00Z\"}]"
run
grep -q '^api -X PATCH repos/o/r/issues/comments/2$' <<<"$calls" && ! grep -q 'POST' <<<"$calls" && ok "the marker comment is edited by id, even when a duplicate was updated later" || ko "edit marker ($calls)"
grep -q 'graphql.*id=N1' <<<"$calls" && ok "the older duplicate Triage comment is minimized as outdated" || ko "minimize ($calls)"
! grep -q 'N3\|comments/3' <<<"$calls" && ! grep -q 'N4\|comments/4' <<<"$calls" && ok "another bot comment and a human Triage comment are never touched" || ko "untouched ($calls)"
! grep -qi 'DELETE' <<<"$calls" && ok "nothing is ever deleted" || ko "delete ($calls)"

# No marker (an older pull request): the most recently updated bot Triage comment is adopted and edited.
comments "[{\"id\":1,\"node_id\":\"N1\",\"user\":{\"login\":\"$BOT\"},\"body\":\"Triage: MEDIUM | blast radius: x\",\"created_at\":\"2026-11-13T16:58:00Z\",\"updated_at\":\"2026-11-13T16:58:00Z\"},{\"id\":2,\"node_id\":\"N2\",\"user\":{\"login\":\"$BOT\"},\"body\":\"Triage: NONE | blast radius: x\",\"created_at\":\"2026-11-13T17:06:00Z\",\"updated_at\":\"2026-11-14T08:00:00Z\"}]"
run
grep -q '^api -X PATCH repos/o/r/issues/comments/2$' <<<"$calls" && grep -q 'graphql.*id=N1' <<<"$calls" && ok "no marker: the latest updated Triage comment is edited, the other minimized" || ko "adopt ($calls)"

# Fail open.
comments '[]'
JSON='{"body":"no verdict here"}' run
[ "$rc" -eq 0 ] && grep -q 'NOT RUN' <<<"$out" && [ -z "$calls" ] && ok "a text without a Triage first line is NOT RUN and posts nothing" || ko "no verdict (rc=$rc $out $calls)"
JSON='{}' run
[ "$rc" -eq 0 ] && grep -q 'NOT RUN' <<<"$out" && ok "a structured output without a body is NOT RUN, exit 0" || ko "no json (rc=$rc $out)"
FAKE_GH_FAIL=1 run
[ "$rc" -eq 0 ] && grep -q 'NOT RUN' <<<"$out" && ok "a GitHub error is NOT RUN, exit 0" || ko "gh fail (rc=$rc $out)"
TOOL='Tool: anything the model wants' run
body="$(sed -n 's/^BODY //p' <<<"$calls" | python3 -I -c 'import json,sys; print(json.load(sys.stdin)["body"])')"
[ "$(sed -n 3p <<<"$body")" = "Tool: NOT RUN (the preview step gave no result)" ] && ok "a malformed Tool line is replaced by NOT RUN" || ko "bad tool ($body)"

finish
