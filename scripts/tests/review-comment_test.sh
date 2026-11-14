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
  out="$(PATH="$tmp/bin:$PATH" FX="$tmp" REPO=o/r PR_NUMBER=7 PR_HEAD_SHA="${HEAD_SHA:-}" REVIEW_OUTCOME="${OUTCOME:-success}" TOOL_LINE="${TOOL:-Tool: PASS (3 files)}" REVIEW_JSON="${JSON:-$REVIEW}" python3 -I "$tool" 2>&1)" || rc=$?
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

# posted_body: the body of the last write in the call log.
posted_body() { sed -n 's/^BODY //p' <<<"$calls" | tail -1 | python3 -I -c 'import json,sys; print(json.load(sys.stdin)["body"])'; }
# ledger_sha: the sha of the fenced ledger of the posted body (empty when there is none).
ledger_sha() { posted_body | python3 -I -c 'import json,re,sys; m=re.search(r"```json\s*(\{.*?\})\s*```", sys.stdin.read(), re.S); print(json.loads(m.group(1))["sha"] if m else "")'; }
head_sha=0123456789abcdef0123456789abcdef01234567

# The ledger sha is never the model's: it is always the validated PR head.
comments '[]'
HEAD_SHA="$head_sha" JSON='{"body":"Triage: NONE | blast radius: x\n\nok\n\n<details><summary>Review ledger</summary>\n\n```json\n{\"sha\":\"fedcba9876543210fedcba9876543210fedcba98\",\"findings\":[{\"id\":\"f1\",\"status\":\"open\"}]}\n```\n\n</details>"}' run
[ "$(ledger_sha)" = "$head_sha" ] && grep -q '"f1"' <<<"$(posted_body)" && ok "the model's ledger sha is overwritten with PR_HEAD_SHA, its findings kept" || ko "ledger sha ($(posted_body))"
# read_ledger: the ledger the next run reads from the posted body (scripts/review-context.py), or "none".
read_ledger() {
  posted_body | python3 -I -c '
import importlib.util, json, sys
spec = importlib.util.spec_from_file_location("rc", sys.argv[1]); m = importlib.util.module_from_spec(spec); spec.loader.exec_module(m)
led = m.parse_ledger([{"user": {"login": "github-actions[bot]"}, "body": sys.stdin.read()}])
print(led["sha"] if led else "none")' "$root/scripts/review-context.py"
}
# A missing, bare (no <details>) or broken ledger is never replaced by an empty one at the head: the next run is first.
HEAD_SHA="$head_sha" JSON='{"body":"Triage: NONE | blast radius: x\n\nno ledger written"}' run
[ "$(read_ledger)" = "none" ] && ok "a review without a ledger gets none (no empty ledger at the head): the next run is first" || ko "no ledger ($(posted_body))"
HEAD_SHA="$head_sha" JSON='{"body":"Triage: NONE | blast radius: x\n\n```json\n{\"sha\":\"fedcba9876543210fedcba9876543210fedcba98\",\"findings\":[]}\n```"}' run
[ "$(read_ledger)" = "none" ] && ok "a bare ledger (no <details>) is not usable: no ledger at the head" || ko "bare ledger ($(posted_body))"
HEAD_SHA="$head_sha" JSON='{"body":"Triage: NONE | blast radius: x\n\n<!-- review-ledger {\"sha\":\"fedcba9876543210fedcba9876543210fedcba98\",\"findings\":[]} -->"}' run
[ "$(read_ledger)" = "none" ] && ok "a hidden-comment ledger with the model's sha is dropped, not trusted" || ko "legacy ledger ($(posted_body))"
HEAD_SHA="$head_sha" JSON='{"body":"Triage: NONE | blast radius: x\n\n<details><summary>Review ledger</summary>\n\n```json\n{not json\n```\n\n</details>"}' run
[ "$(read_ledger)" = "none" ] && ok "a broken ledger JSON is not replaced by an empty ledger at the head" || ko "broken ledger ($(posted_body))"

# No usable review (empty, invalid, timed out): a NOT RUN marker comment, so nothing looks reviewed. The previous
# ledger (last reviewed commit and findings) is carried unchanged, so the commits not reviewed are reviewed next time.
python3 -I - "$tmp/comments.json" <<'PY'
import json, sys
body = "Triage: HIGH | blast radius: x\n<!-- 226os-ai-review -->\nTool: PASS (1 files)\n\n<details><summary>Review ledger</summary>\n\n```json\n" + json.dumps({"sha": "a" * 40, "findings": [{"id": "keep", "status": "open"}]}) + "\n```\n</details>"
json.dump([{"id": 2, "node_id": "N2", "user": {"login": "github-actions[bot]"}, "body": body, "updated_at": "2026-11-14T09:00:00Z"}], open(sys.argv[1], "w"))
PY
HEAD_SHA="$head_sha" OUTCOME=failure JSON='{}' run
pb="$(posted_body)"
[ "$rc" -eq 0 ] && grep -q '^api -X PATCH repos/o/r/issues/comments/2$' <<<"$calls" && [ "$(sed -n 1p <<<"$pb")" = "Triage: NOT RUN | the review step ended with failure and returned no text" ] && [ "$(sed -n 2p <<<"$pb")" = "<!-- 226os-ai-review -->" ] && ok "no review text (timeout or failure): the marker comment says Triage: NOT RUN with the reason" || ko "not run comment (rc=$rc $calls $pb)"
[ "$(ledger_sha)" = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" ] && grep -q '"keep"' <<<"$pb" && ok "a NOT RUN comment carries the previous ledger, never the head" || ko "not run ledger ($pb)"
# A CRITICAL stays a CRITICAL when the next push gives no review (max turns, timeout): the gate still blocks.
python3 -I - "$tmp/comments.json" <<'PY'
import json, sys
body = ("Triage: CRITICAL | blast radius: the whole team\n<!-- 226os-ai-review -->\nTool: PASS (1 files)\n\n"
        "OBSERVED: src/api/config.go:12 holds a literal API key.\n\n<details><summary>Review ledger</summary>\n\n```json\n"
        + json.dumps({"sha": "b" * 40, "findings": [{"id": "key", "sev": "CRITICAL", "status": "open"}]}) + "\n```\n</details>")
json.dump([{"id": 2, "node_id": "N2", "user": {"login": "github-actions[bot]"}, "body": body, "updated_at": "2026-11-14T09:00:00Z"}], open(sys.argv[1], "w"))
PY
HEAD_SHA="$head_sha" OUTCOME=failure JSON='{}' run
pb="$(posted_body)"
verdict="$(python3 -I -c '
import importlib.util, sys
spec = importlib.util.spec_from_file_location("gate", sys.argv[1]); m = importlib.util.module_from_spec(spec); spec.loader.exec_module(m)
found = m.latest_verdict([{"user": {"login": "github-actions[bot]"}, "body": sys.stdin.read()}])
print(found[0] if found and "OBSERVED" in found[1] and m.EVIDENCE.search(found[1]) else "none")' "$root/scripts/review-gate.py" <<<"$pb")"
[ "$verdict" = "CRITICAL" ] && grep -q 'NOT RUN' <<<"$pb" && [ "$(ledger_sha)" = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" ] && ok "CRITICAL, then a push with no review: the CRITICAL and its evidence stay, the gate still blocks" || ko "critical kept (verdict=$verdict $pb)"

comments '[]'
HEAD_SHA="$head_sha" JSON='{"body":"no verdict here"}' run
pb="$(posted_body)"
[ "$rc" -eq 0 ] && grep -q '^Triage: NOT RUN | the review text has no Triage first line$' <<<"$pb" && [ -z "$(ledger_sha)" ] && ok "a text without a Triage line: NOT RUN comment, no ledger when there was none" || ko "no verdict (rc=$rc $pb)"
FAKE_GH_FAIL=1 run
[ "$rc" -eq 0 ] && grep -q 'NOT RUN' <<<"$out" && ok "a GitHub error is NOT RUN, exit 0" || ko "gh fail (rc=$rc $out)"
TOOL='Tool: anything the model wants' run
body="$(sed -n 's/^BODY //p' <<<"$calls" | python3 -I -c 'import json,sys; print(json.load(sys.stdin)["body"])')"
[ "$(sed -n 3p <<<"$body")" = "Tool: NOT RUN (the preview step gave no result)" ] && ok "a malformed Tool line is replaced by NOT RUN" || ko "bad tool ($body)"

finish
