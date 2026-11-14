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
  "api repos/o/r/pulls/7/files --paginate") cat "$FX/files.json" ;;
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
  out="$(PATH="$tmp/bin:$PATH" FX="$tmp" REPO=o/r PR_NUMBER=7 TOOL_LINE="${TOOL:-Tool: PASS (3 files)}" REVIEW_JSON="${JSON-$REVIEW}" python3 -I "$tool" 2>&1)" || rc=$?
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

# Fallback: an empty or invalid review still writes the marker comment, with the verdict NOT RUN (never a pass).
posted() { sed -n 's/^BODY //p' <<<"$calls" | tail -1 | python3 -I -c 'import json,sys; print(json.load(sys.stdin)["body"])' 2>/dev/null || true; }
ledger_of() { python3 -I -c 'import json,re,sys; m=re.search(r"```json\s*(\{.*?\})\s*```", sys.stdin.read(), re.S); print(json.dumps(json.loads(m.group(1)), sort_keys=True) if m else "NONE")' 2>/dev/null || true; }
comments '[]'
JSON='{"body":"no verdict here"}' run
body="$(posted)"
[ "$rc" -eq 0 ] && grep -q '^api -X POST repos/o/r/issues/7/comments$' <<<"$calls" && [ "$(sed -n 1p <<<"$body")" = "Triage: NOT RUN | the review text has no Triage first line" ] && [ "$(sed -n 2p <<<"$body")" = "<!-- 226os-ai-review -->" ] && ok "a text without a Triage first line writes the marker comment with Triage: NOT RUN" || ko "no verdict (rc=$rc $out $body)"
[ "$(ledger_of <<<"$body")" = '{"findings": [], "sha": ""}' ] && ok "NOT RUN without a previous ledger writes an empty ledger" || ko "empty ledger ($body)"
JSON='' run
[ "$rc" -eq 0 ] && [ "$(posted | sed -n 1p)" = "Triage: NOT RUN | the review returned no structured output" ] && ok "an empty REVIEW_JSON is NOT RUN with its reason" || ko "empty json (rc=$rc $out)"
JSON='{not json' run
[ "$rc" -eq 0 ] && [ "$(posted | sed -n 1p)" = "Triage: NOT RUN | the review output is not valid JSON" ] && ok "an invalid REVIEW_JSON is NOT RUN with its reason" || ko "bad json (rc=$rc $out)"
JSON='{}' run
[ "$rc" -eq 0 ] && posted | head -1 | grep -q '^Triage: NOT RUN | ' && ok "a structured output without a body is NOT RUN, exit 0" || ko "no json (rc=$rc $out)"

# NOT RUN on a later push keeps the previous ledger (an open HIGH must not vanish); its sha stays the old one.
prev='Triage: HIGH | blast radius: one slice\n<!-- 226os-ai-review -->\nTool: PASS (2 files)\n\n<details><summary>Review ledger</summary>\n\n```json\n{\"sha\":\"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\",\"findings\":[{\"id\":\"h1\",\"kind\":\"finding\",\"evidence\":\"OBSERVED\",\"sev\":\"HIGH\",\"file\":\"a.go\",\"line\":3,\"scenario\":\"x\",\"status\":\"open\"}]}\n```\n\n</details>'
comments "[{\"id\":2,\"node_id\":\"N2\",\"user\":{\"login\":\"$BOT\"},\"body\":\"$prev\",\"updated_at\":\"2026-11-14T09:00:00Z\"}]"
JSON='' run
body="$(posted)"
grep -q '^api -X PATCH repos/o/r/issues/comments/2$' <<<"$calls" && [ "$(sed -n 1p <<<"$body")" = "Triage: NOT RUN | the review returned no structured output" ] && grep -q '"id": "h1"' <<<"$(ledger_of <<<"$body")" && grep -q '"sha": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"' <<<"$(ledger_of <<<"$body")" && ok "NOT RUN carries the previous ledger forward, with its old sha" || ko "carry forward ($body)"

# --- deep review: merged into the same comment and ledger (subcommand deep) ---
HEAD_SHA=bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
base_review='Triage: MEDIUM | blast radius: this PR only\n<!-- 226os-ai-review -->\nTool: PASS (2 files)\n\nFindings:\n- m1 MEDIUM b.go:4\n\n<details><summary>Review ledger</summary>\n\n```json\n{\"sha\":\"'$HEAD_SHA'\",\"findings\":[{\"id\":\"m1\",\"kind\":\"finding\",\"evidence\":\"INFERRED\",\"sev\":\"MEDIUM\",\"file\":\"b.go\",\"line\":4,\"scenario\":\"y\",\"status\":\"open\"},{\"id\":\"d1\",\"kind\":\"finding\",\"evidence\":\"OBSERVED\",\"sev\":\"HIGH\",\"file\":\"c.go\",\"line\":1,\"scenario\":\"old copy\",\"status\":\"open\"}]}\n```\n\n</details>'
DEEP='{"level":"HIGH","blast_radius":"other people'"'"'s work","summary":"Triage: NONE | ignored\nOBSERVED: c.go:9 reads a removed answer.","findings":[{"id":"d1","kind":"finding","evidence":"OBSERVED","sev":"HIGH","file":"c.go","line":9,"scenario":"a removed answer is returned","status":"open"},{"id":"d2","kind":"note","evidence":"INFERRED","sev":"LOW","file":"d.go","line":2,"scenario":"maybe","status":"open"}]}'
deep() {
  rc=0
  : >"$tmp/calls.log"
  out="$(PATH="$tmp/bin:$PATH" FX="$tmp" REPO=o/r PR_NUMBER=7 PR_HEAD_SHA="$HEAD_SHA" DEEP_JSON="${DJSON-$DEEP}" python3 -I "$tool" deep 2>&1)" || rc=$?
  calls="$(cat "$tmp/calls.log")"
}
comments "[{\"id\":2,\"node_id\":\"N2\",\"user\":{\"login\":\"$BOT\"},\"body\":\"$base_review\",\"updated_at\":\"2026-11-14T09:00:00Z\"}]"
deep
body="$(posted)"
l="$(ledger_of <<<"$body")"
[ "$rc" -eq 0 ] && grep -q '^api -X PATCH repos/o/r/issues/comments/2$' <<<"$calls" && ok "deep: the marker comment is edited in place" || ko "deep edit (rc=$rc $out $calls)"
[ "$(sed -n 1p <<<"$body")" = "Triage: HIGH | blast radius: other people's work" ] && ok "deep: the Triage line is the worst of both reviews" || ko "deep triage ($body)"
[ "$(sed -n 3p <<<"$body")" = "Tool: PASS (2 files)" ] && [ "$(sed -n 4p <<<"$body")" = "Deep: PASS (1 finding, 1 note)" ] && ok "deep: line 3 keeps the Tool line, line 4 is the Deep line" || ko "deep lines ($body)"
python3 -I -c 'import json,sys; d=json.loads(sys.argv[1]); f={x["id"]:x for x in d["findings"]}; assert d["sha"]==sys.argv[2]; assert f["d1"]["source"]=="deep" and f["d1"]["line"]==9; assert f["d2"]["source"]=="deep"; assert "source" not in f["m1"]; assert len(d["findings"])==3; assert d["deep"]=={"status":"PASS","sha":sys.argv[2]}' "$l" "$HEAD_SHA" && ok "deep: findings merge into the same ledger, tagged source deep, deep wins on the same id, deep sha recorded" || ko "deep ledger ($l)"
grep -q 'c.go:9 reads a removed answer' <<<"$body" && [ "$(grep -c '^Triage:' <<<"$body")" -eq 1 ] && ok "deep: the summary is added, a Triage line inside it is dropped" || ko "deep summary ($body)"

# Re-run (idempotent): merging the same output again gives the same body.
comments "$(python3 -I -c 'import json,sys; print(json.dumps([{"id":2,"node_id":"N2","user":{"login":"github-actions[bot]"},"body":sys.argv[1],"updated_at":"2026-11-14T10:00:00Z"}]))' "$body")"
first="$body"
deep
[ -n "$first" ] && [ "$(posted)" = "$first" ] && ok "deep: merging twice is idempotent (one deep section)" || ko "deep idempotent ($(posted))"

# A CRITICAL without OBSERVED evidence is downgraded to HIGH.
comments "[{\"id\":2,\"node_id\":\"N2\",\"user\":{\"login\":\"$BOT\"},\"body\":\"$base_review\",\"updated_at\":\"2026-11-14T09:00:00Z\"}]"
DJSON='{"level":"CRITICAL","blast_radius":"production and users","summary":"guess","findings":[{"id":"c1","kind":"finding","evidence":"INFERRED","sev":"CRITICAL","file":"e.go","line":1,"scenario":"s","status":"open"}]}' deep
body="$(posted)"
[ "$(sed -n 1p <<<"$body")" = "Triage: HIGH | blast radius: production and users" ] && grep -q '"sev": "HIGH"' <<<"$(ledger_of <<<"$body")" && ok "deep: a CRITICAL without OBSERVED evidence is downgraded to HIGH" || ko "deep downgrade ($body)"

# The deep review failed: Deep NOT RUN, the review's verdict and ledger stay, and the ledger says deep NOT RUN.
DJSON='' deep
body="$(posted)"
[ "$(sed -n 1p <<<"$body")" = "Triage: MEDIUM | blast radius: this PR only" ] && [ "$(sed -n 4p <<<"$body")" = "Deep: NOT RUN (the deep review returned no structured output)" ] && python3 -I -c 'import json,sys; d=json.loads(sys.argv[1]); assert d["deep"]["status"]=="NOT RUN"' "$(ledger_of <<<"$body")" && ok "deep: no output is Deep NOT RUN, recorded in the ledger" || ko "deep not run ($body)"

# The standard review was NOT RUN: the verdict stays NOT RUN, the sha is not set to the head.
comments '[]'
deep
body="$(posted)"
sed -n 1p <<<"$body" | grep -q '^Triage: NOT RUN | ' && python3 -I -c 'import json,sys; d=json.loads(sys.argv[1]); assert d["sha"]=="" and d["deep"]["status"]=="PASS"' "$(ledger_of <<<"$body")" && ok "deep: when the standard review left no comment, the verdict stays NOT RUN and the sha is not the head" || ko "deep over not run ($body)"

# deep-needed: P0 paths from the two lists of ROOT.
mkdir -p "$tmp/root/.github"
printf '# c\nscripts/*\n' >"$tmp/root/.github/sensitive-paths.txt"
printf 'src/api/internal/identity/*\ndb/queries/*\n' >"$tmp/root/.github/deep-review-paths.txt"
need() { printf '%s' "$1" >"$tmp/files.json"; PATH="$tmp/bin:$PATH" FX="$tmp" REPO=o/r PR_NUMBER=7 ROOT="$tmp/root" python3 -I "$tool" deep-needed 2>/dev/null; }
[ "$(need '[{"filename":"db/queries/solve.sql"}]')" = "true" ] && [ "$(need '[{"filename":"scripts/x.sh"}]')" = "true" ] && [ "$(need '[{"filename":"docs/a.md"}]')" = "false" ] && ok "deep-needed: a P0 path (either list) is true, other paths false" || ko "deep-needed"
FAKE_GH_FAIL=1 need '[]' | grep -qx false && ok "deep-needed: a GitHub error is false (never blocks)" || ko "deep-needed fail"
FAKE_GH_FAIL=1 run
[ "$rc" -eq 0 ] && grep -q 'NOT RUN' <<<"$out" && ok "a GitHub error is NOT RUN, exit 0" || ko "gh fail (rc=$rc $out)"
TOOL='Tool: anything the model wants' run
body="$(sed -n 's/^BODY //p' <<<"$calls" | python3 -I -c 'import json,sys; print(json.load(sys.stdin)["body"])')"
[ "$(sed -n 3p <<<"$body")" = "Tool: NOT RUN (the preview step gave no result)" ] && ok "a malformed Tool line is replaced by NOT RUN" || ko "bad tool ($body)"

finish
