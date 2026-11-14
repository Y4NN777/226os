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
  out="$(PATH="$tmp/bin:$PATH" FX="$tmp" REPO=o/r PR_NUMBER=7 PR_HEAD_SHA="${HEAD_SHA:-}" REVIEW_OUTCOME="${OUTCOME:-success}" TOOL_LINE="${TOOL:-Tool: PASS (3 files)}" REVIEW_JSON="${JSON:-$REVIEW}" DEEP_PENDING="${DEEP_PENDING:-}" REVIEW_RANGE="${RANGE-0123456789abcdef0123456789abcdef01234567..fedcba9876543210fedcba9876543210fedcba98}" REVIEW_NOW=14:05 GITHUB_OUTPUT="$tmp/gh-output" python3 -I "$tool" 2>&1)" || rc=$?
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

[ "$(sed -n 4p <<<"$body")" = "Reviewed: unknown (0123456..fedcba9) at 14:05 UTC" ] && ok "line 4 says what this run reviewed: head, range, time" || ko "reviewed line ($body)"
grep -q '^comment_id=999$' "$tmp/gh-output" && ok "the new comment's id is written to GITHUB_OUTPUT (the deep job of the same run edits only it)" || ko "comment id output ($(cat "$tmp/gh-output" 2>/dev/null))"

# One comment PER RUN: a second run creates a new comment and collapses every older bot review comment.
comments "[{\"id\":1,\"node_id\":\"N1\",\"user\":{\"login\":\"$BOT\"},\"body\":\"Triage: MEDIUM | blast radius: x\\nold\",\"created_at\":\"2026-11-14T08:00:00Z\",\"updated_at\":\"2026-11-14T11:00:00Z\"},{\"id\":2,\"node_id\":\"N2\",\"user\":{\"login\":\"$BOT\"},\"body\":\"Triage: NONE | blast radius: x\\n<!-- 226os-ai-review -->\\nTool: PASS (1 files)\",\"created_at\":\"2026-11-14T09:00:00Z\",\"updated_at\":\"2026-11-14T09:00:00Z\"},{\"id\":3,\"node_id\":\"N3\",\"user\":{\"login\":\"$BOT\"},\"body\":\"Owner review requested.\",\"updated_at\":\"2026-11-14T12:00:00Z\"},{\"id\":4,\"node_id\":\"N4\",\"user\":{\"login\":\"Y4NN777\"},\"body\":\"Triage: NONE | blast radius: me\\n<!-- 226os-ai-review -->\",\"updated_at\":\"2026-11-14T12:00:00Z\"}]"
run
grep -q '^api -X POST repos/o/r/issues/7/comments$' <<<"$calls" && ! grep -q 'PATCH' <<<"$calls" && ok "a second run creates a new comment, it never edits an older one" || ko "create per run ($calls)"
grep -q 'graphql.*id=N2' <<<"$calls" && grep -q 'graphql.*id=N1' <<<"$calls" && ok "every older bot review comment (marker or Triage) is minimized as outdated" || ko "minimize ($calls)"
! grep -q 'N3\|comments/3' <<<"$calls" && ! grep -q 'N4\|comments/4' <<<"$calls" && ok "another bot comment and a human comment (even with the marker) are never touched" || ko "untouched ($calls)"
! grep -qi 'DELETE' <<<"$calls" && ok "nothing is ever deleted" || ko "delete ($calls)"

# The ledger carries forward: the previous comment's open entries (ids and statuses) the model left out are kept.
python3 -I - "$tmp/comments.json" <<'PY2'
import json, sys
led = {"sha": "a" * 40, "findings": [{"id": "old-open", "kind": "finding", "sev": "MEDIUM", "status": "answered"}, {"id": "old-fixed", "kind": "finding", "sev": "MEDIUM", "status": "fixed"}, {"id": "dup", "kind": "finding", "sev": "MEDIUM", "status": "open"}, {"id": "d9", "kind": "finding", "sev": "HIGH", "status": "open", "source": "deep"}], "deep": {"status": "PASS", "sha": "a" * 40}}
body = "Triage: MEDIUM | blast radius: x\n<!-- 226os-ai-review -->\nTool: PASS (1 files)\n\n<details><summary>Review ledger</summary>\n\n```json\n" + json.dumps(led) + "\n```\n\n</details>"
json.dump([{"id": 2, "node_id": "N2", "user": {"login": "github-actions[bot]"}, "body": body, "created_at": "2026-11-14T09:00:00Z"}], open(sys.argv[1], "w"))
PY2
HEAD_SHA=0123456789abcdef0123456789abcdef01234567 JSON='{"body":"Triage: MEDIUM | blast radius: x\n\n<details><summary>Review ledger</summary>\n\n```json\n{\"sha\":\"x\",\"findings\":[{\"id\":\"dup\",\"kind\":\"finding\",\"sev\":\"MEDIUM\",\"status\":\"fixed\",\"detail\":\"d\",\"fix\":\"f\"},{\"id\":\"new\",\"kind\":\"finding\",\"sev\":\"MEDIUM\",\"status\":\"open\",\"detail\":\"d\",\"fix\":\"f\"}]}\n```\n\n</details>"}' run
python3 -I -c '
import json, re, sys
body = json.loads(sys.stdin.read().split("BODY ", 1)[1].splitlines()[0])["body"]
d = json.loads(re.search(r"```json\s*(\{.*?\})\s*```", body, re.S).group(1))
f = {x["id"]: x for x in d["findings"]}
assert f["dup"]["status"] == "fixed" and f["new"]["status"] == "open"
assert f["old-open"]["status"] == "answered" and "old-fixed" not in f
assert f["d9"].get("source") == "deep" and "deep" not in d' <<<"$calls" && ok "the ledger carries the previous comment's entries forward (ids and statuses kept, fixed ones dropped)" || ko "carry forward ($calls)"

# posted_body: the body of the last write in the call log.
posted_body() { sed -n 's/^BODY //p' <<<"$calls" | tail -1 | python3 -I -c 'import json,sys; print(json.load(sys.stdin)["body"])'; }
# ledger_sha: the sha of the fenced ledger of the posted body (empty when there is none).
ledger_sha() { posted_body | python3 -I -c 'import json,re,sys; m=re.search(r"```json\s*(\{.*?\})\s*```", sys.stdin.read(), re.S); print(json.loads(m.group(1))["sha"] if m else "")'; }
head_sha=0123456789abcdef0123456789abcdef01234567

# The ledger sha is never the model's: it is always the validated PR head.
comments '[]'
HEAD_SHA="$head_sha" JSON='{"body":"Triage: NONE | blast radius: x\n\nok\n\n<details><summary>Review ledger</summary>\n\n```json\n{\"sha\":\"fedcba9876543210fedcba9876543210fedcba98\",\"findings\":[{\"id\":\"f1\",\"status\":\"open\",\"detail\":\"d\",\"fix\":\"f\"}]}\n```\n\n</details>"}' run
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
[ "$rc" -eq 0 ] && grep -q '^api -X POST repos/o/r/issues/7/comments$' <<<"$calls" && grep -q 'graphql.*id=N2' <<<"$calls" && [ "$(sed -n 1p <<<"$pb")" = "Triage: NOT RUN | the review step ended with failure and returned no text" ] && [ "$(sed -n 2p <<<"$pb")" = "<!-- 226os-ai-review -->" ] && ok "no review text (timeout or failure): a new marker comment says Triage: NOT RUN with the reason, the older one collapsed" || ko "not run comment (rc=$rc $calls $pb)"
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

posted() { sed -n 's/^BODY //p' <<<"$calls" | tail -1 | python3 -I -c 'import json,sys; print(json.load(sys.stdin)["body"])' 2>/dev/null || true; }
ledger_of() { python3 -I -c 'import json,re,sys; m=re.search(r"```json\s*(\{.*?\})\s*```", sys.stdin.read(), re.S); print(json.dumps(json.loads(m.group(1)), sort_keys=True) if m else "NONE")' 2>/dev/null || true; }

# DEEP_PENDING: line 4 says the deep review runs next (auto-merge refuses until the deep job replaces it).
comments '[]'
DEEP_PENDING=true run
[ "$(sed -n 's/^BODY //p' <<<"$calls" | python3 -I -c 'import json,sys; print(json.load(sys.stdin)["body"])' | sed -n 5p)" = "Deep: PENDING (a P0 path is touched: the deep review runs next)" ] && ok "DEEP_PENDING=true writes the Deep: PENDING line 4" || ko "deep pending ($calls)"

# --- deep review: merged into the same comment and ledger (subcommand deep) ---
HEAD_SHA=bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
base_review='Triage: MEDIUM | blast radius: this PR only\n<!-- 226os-ai-review -->\nTool: PASS (2 files)\n\nFindings:\n- m1 MEDIUM b.go:4\n\n<details><summary>Review ledger</summary>\n\n```json\n{\"sha\":\"'$HEAD_SHA'\",\"findings\":[{\"id\":\"m1\",\"kind\":\"finding\",\"evidence\":\"INFERRED\",\"sev\":\"MEDIUM\",\"file\":\"b.go\",\"line\":4,\"scenario\":\"y\",\"status\":\"open\"},{\"id\":\"d1\",\"kind\":\"finding\",\"evidence\":\"OBSERVED\",\"sev\":\"HIGH\",\"file\":\"c.go\",\"line\":1,\"scenario\":\"old copy\",\"status\":\"open\"}]}\n```\n\n</details>'
DEEP='{"level":"HIGH","blast_radius":"other people'"'"'s work","summary":"Triage: NONE | ignored\nOBSERVED: c.go:9 reads a removed answer.","findings":[{"id":"d1","kind":"finding","evidence":"OBSERVED","sev":"HIGH","file":"c.go","line":9,"scenario":"a removed answer is returned","status":"open","detail":"d","fix":"f"},{"id":"d2","kind":"note","evidence":"INFERRED","sev":"LOW","file":"d.go","line":2,"scenario":"maybe","status":"open","detail":"d","fix":"f"}]}'
deep() {
  rc=0
  : >"$tmp/calls.log"
  out="$(PATH="$tmp/bin:$PATH" FX="$tmp" REPO=o/r PR_NUMBER=7 PR_HEAD_SHA="$HEAD_SHA" DEEP_JSON="${DJSON-$DEEP}" REVIEW_COMMENT_ID="${CID:-}" python3 -I "$tool" deep 2>&1)" || rc=$?
  calls="$(cat "$tmp/calls.log")"
}
comments "[{\"id\":2,\"node_id\":\"N2\",\"user\":{\"login\":\"$BOT\"},\"body\":\"$base_review\",\"updated_at\":\"2026-11-14T09:00:00Z\"}]"
deep
body="$(posted)"
l="$(ledger_of <<<"$body")"
[ "$rc" -eq 0 ] && grep -q '^api -X PATCH repos/o/r/issues/comments/2$' <<<"$calls" && ok "deep: the marker comment is edited in place" || ko "deep edit (rc=$rc $out $calls)"
[ "$(sed -n 1p <<<"$body")" = "Triage: HIGH | blast radius: other people's work" ] && ok "deep: the Triage line is the worst of both reviews" || ko "deep triage ($body)"
[ "$(sed -n 3p <<<"$body")" = "Tool: PASS (2 files)" ] && [ "$(sed -n 4p <<<"$body")" = "Deep: PASS (1 finding, 1 note)" ] && ok "deep: line 3 keeps the Tool line, line 4 is the Deep line" || ko "deep lines ($body)"
python3 -I -c 'import json,sys; d=json.loads(sys.argv[1]); f={x["id"]:x for x in d["findings"]}; assert d["sha"]==sys.argv[2]; assert f["d1"]["source"]=="review" and f["d1"]["line"]==1 and f["d1"]["deep"]["line"]==9 and f["d1"]["deep"]["detail"]=="d"; assert f["d2"]["source"]=="deep"; assert "source" not in f["m1"]; assert len(d["findings"])==3; assert d["deep"]=={"status":"PASS","sha":sys.argv[2]}' "$l" "$HEAD_SHA" && ok "deep: findings merge into the same ledger; a new one is tagged source deep, a repeated id keeps source review and gains a deep sub-field (owner decision: never overwritten), deep sha recorded" || ko "deep ledger ($l)"
grep -q 'c.go:9 reads a removed answer' <<<"$body" && [ "$(grep -c '^Triage:' <<<"$body")" -eq 1 ] && ok "deep: the summary is added, a Triage line inside it is dropped" || ko "deep summary ($body)"

# Re-run (idempotent): merging the same output again gives the same body.
comments "$(python3 -I -c 'import json,sys; print(json.dumps([{"id":2,"node_id":"N2","user":{"login":"github-actions[bot]"},"body":sys.argv[1],"updated_at":"2026-11-14T10:00:00Z"}]))' "$body")"
first="$body"
deep
[ -n "$first" ] && [ "$(posted)" = "$first" ] && ok "deep: merging twice is idempotent (one deep section)" || ko "deep idempotent ($(posted))"

# A CRITICAL without OBSERVED evidence is downgraded to HIGH.
comments "[{\"id\":2,\"node_id\":\"N2\",\"user\":{\"login\":\"$BOT\"},\"body\":\"$base_review\",\"updated_at\":\"2026-11-14T09:00:00Z\"}]"
DJSON='{"level":"CRITICAL","blast_radius":"production and users","summary":"guess","findings":[{"id":"c1","kind":"finding","evidence":"INFERRED","sev":"CRITICAL","file":"e.go","line":1,"scenario":"s","status":"open","detail":"d","fix":"f"}]}' deep
body="$(posted)"
[ "$(sed -n 1p <<<"$body")" = "Triage: HIGH | blast radius: production and users" ] && grep -q '"sev": "HIGH"' <<<"$(ledger_of <<<"$body")" && ok "deep: a CRITICAL without OBSERVED evidence is downgraded to HIGH" || ko "deep downgrade ($body)"

# The deep review failed: Deep NOT RUN, the review's verdict and ledger stay, and the ledger says deep NOT RUN.
DJSON='' deep
body="$(posted)"
[ "$(sed -n 1p <<<"$body")" = "Triage: MEDIUM | blast radius: this PR only" ] && [ "$(sed -n 4p <<<"$body")" = "Deep: NOT RUN (the deep review returned no structured output)" ] && python3 -I -c 'import json,sys; d=json.loads(sys.argv[1]); assert d["deep"]["status"]=="NOT RUN"' "$(ledger_of <<<"$body")" && ok "deep: no output is Deep NOT RUN, recorded in the ledger" || ko "deep not run ($body)"

# The standard review of this run was NOT RUN (its comment id is passed): the verdict stays NOT RUN, the sha is not set.
comments "[{\"id\":5,\"node_id\":\"N5\",\"user\":{\"login\":\"$BOT\"},\"body\":\"Triage: NOT RUN | the review step ended with failure and returned no text\\n<!-- 226os-ai-review -->\\nTool: PASS (1 files)\\nReviewed: none\",\"created_at\":\"2026-11-14T10:00:00Z\"}]"
CID=5 deep
body="$(posted)"
grep -q '^api -X PATCH repos/o/r/issues/comments/5$' <<<"$calls" && sed -n 1p <<<"$body" | grep -q '^Triage: NOT RUN | ' && [ "$(sed -n 4p <<<"$body")" = "Reviewed: none" ] && [ "$(sed -n 5p <<<"$body")" = "Deep: PASS (1 finding, 1 note)" ] && python3 -I -c 'import json,sys; d=json.loads(sys.argv[1]); assert d["sha"]=="" and d["deep"]["status"]=="PASS"' "$(ledger_of <<<"$body")" && ok "deep: merges into its own run's comment (by id) after the Reviewed line; a NOT RUN verdict stays NOT RUN" || ko "deep over not run ($calls $body)"

# A newer push already has a newer review comment: the deep job never edits the older one (stale, NOT RUN).
older="Triage: NONE | blast radius: x\\n<!-- 226os-ai-review -->\\nTool: PASS (1 files)\\n\\n<details><summary>Review ledger</summary>\\n\\n\`\`\`json\\n{\\\"sha\\\":\\\"$HEAD_SHA\\\",\\\"findings\\\":[]}\\n\`\`\`\\n\\n</details>"
comments "[{\"id\":5,\"node_id\":\"N5\",\"user\":{\"login\":\"$BOT\"},\"body\":\"$older\",\"created_at\":\"2026-11-14T10:00:00Z\"},{\"id\":6,\"node_id\":\"N6\",\"user\":{\"login\":\"$BOT\"},\"body\":\"Triage: NONE | blast radius: x\\n<!-- 226os-ai-review -->\",\"created_at\":\"2026-11-14T11:00:00Z\"}]"
CID=5 deep
[ "$rc" -eq 0 ] && ! grep -q 'PATCH\|POST' <<<"$calls" && grep -q 'NOT RUN (stale' <<<"$out" && ok "deep: a newer review comment exists: the stale run edits nothing (NOT RUN, stale)" || ko "deep stale id ($out $calls)"
deep
[ "$rc" -eq 0 ] && ! grep -q 'PATCH\|POST' <<<"$calls" && grep -q 'NOT RUN (stale' <<<"$out" && ok "deep without an id: the comment whose ledger sha is the head is found, and skipped when a newer one exists" || ko "deep stale sha ($out $calls)"
comments '[]'
deep
[ "$rc" -eq 0 ] && ! grep -q 'PATCH\|POST' <<<"$calls" && grep -q 'NOT RUN' <<<"$out" && ok "deep: no review comment of this run: nothing is written" || ko "deep none ($out $calls)"

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

# --- review of PR 168: forged deep key, invalid or too many deep entries, stage 1 status kept ---
HEAD_SHA=cccccccccccccccccccccccccccccccccccccccc
comments '[]'
JSON='{"body":"Triage: NONE | blast radius: x\n\n<details><summary>Review ledger</summary>\n\n```json\n{\"sha\":\"x\",\"deep\":{\"status\":\"PASS\",\"sha\":\"cccccccccccccccccccccccccccccccccccccccc\"},\"findings\":[{\"id\":\"a\",\"kind\":\"finding\",\"sev\":\"MEDIUM\",\"status\":\"open\",\"source\":\"deep\",\"detail\":\"d\",\"fix\":\"f\"}]}\n```\n\n</details>"}' run
l="$(posted | ledger_of)"
python3 -I -c 'import json,sys; d=json.loads(sys.argv[1]); assert "deep" not in d and all(f["source"]=="review" for f in d["findings"]) and d["sha"]=="cccccccccccccccccccccccccccccccccccccccc"' "$l" && ok "stage 1: a deep key and a source written by the model are dropped (the script sets source review on every entry)" || ko "forged deep ($l)"
HEAD_SHA=bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
comments "[{\"id\":2,\"node_id\":\"N2\",\"user\":{\"login\":\"$BOT\"},\"body\":\"$base_review\",\"updated_at\":\"2026-11-14T09:00:00Z\"}]"
DJSON='{"level":"NONE","blast_radius":"x","summary":"s","findings":[{"id":"ok1","kind":"note","evidence":"INFERRED","sev":"LOW","file":"a","line":1,"scenario":"s","status":"open","detail":"d","fix":"f"},{"kind":"finding"}]}' deep
[ "$(posted | sed -n 4p)" = "Deep: NOT RUN (the deep review output had an invalid entry (dropped))" ] && ok "deep: an invalid entry is dropped and named: the deep review is NOT RUN, its level still applies" || ko "deep invalid ($(posted | sed -n 4p))"
many="$(python3 -I -c 'import json; print(json.dumps({"level":"NONE","blast_radius":"x","summary":"s","findings":[{"id":"n%d"%i,"kind":"note","evidence":"INFERRED","sev":"LOW","file":"a","line":1,"scenario":"s","status":"open","detail":"d","fix":"f"} for i in range(9)]}))')"
DJSON="$many" deep
[ "$(posted | sed -n 4p)" = "Deep: NOT RUN (too many findings)" ] && ok "deep: more than 8 entries is NOT RUN (too many findings)" || ko "deep too many ($(posted | sed -n 4p))"
many="$(python3 -I -c 'import json; e=[{"id":"n%d"%i,"kind":"note","evidence":"INFERRED","sev":"LOW","file":"a","line":1,"scenario":"s","status":"open","detail":"d","fix":"f"} for i in range(8)]+[{"id":"h1","kind":"finding","evidence":"OBSERVED","sev":"HIGH","file":"a","line":2,"scenario":"s","status":"open","detail":"d","fix":"f"}]; print(json.dumps({"level":"HIGH","blast_radius":"x","summary":"s","findings":e}))')"
DJSON="$many" deep
python3 -I -c 'import json,sys; d=json.loads(sys.argv[1]); deep=[x for x in d["findings"] if x.get("source")=="deep"]; assert len(deep)==8 and any(x["id"]=="h1" for x in deep) and d["deep"]["status"]=="NOT RUN"' "$(posted | ledger_of)" && ok "deep: with more than 8 entries the 8 most severe are kept in the ledger (the HIGH first), the result stays NOT RUN" || ko "deep keep severe ($(posted | ledger_of))"
DJSON='{"level":"NONE","blast_radius":"x","summary":"s","findings":[{"id":"m1","kind":"finding","evidence":"INFERRED","sev":"MEDIUM","file":"b.go","line":4,"scenario":"y","status":"fixed","detail":"d","fix":"f"}]}' deep
python3 -I -c 'import json,sys; d=json.loads(sys.argv[1]); f=[x for x in d["findings"] if x["id"]=="m1"]; assert len(f)==1 and f[0]["status"]=="open" and "source" not in f[0]' "$(posted | ledger_of)" && ok "deep: a deep entry never marks a stage 1 finding fixed (the stage 1 entry is kept)" || ko "deep status ($(posted | ledger_of))"
DJSON='{"level":"HIGH","blast_radius":"x","summary":"s","findings":[{"id":"m1","kind":"finding","evidence":"OBSERVED","sev":"HIGH","file":"b.go","line":4,"scenario":"worse","status":"open","detail":"d","fix":"f"}]}' deep
python3 -I -c 'import json,sys; d=json.loads(sys.argv[1]); f=[x for x in d["findings"] if x["id"]=="m1"]; assert len(f)==1 and f[0]["sev"]=="HIGH" and f[0]["source"]=="review" and f[0]["deep"]["sev"]=="HIGH" and f[0]["scenario"]=="y"' "$(posted | ledger_of)" && ok "deep: a deep entry that is not better raises the severity of the stage 1 entry, which keeps its source review" || ko "deep override ($(posted | ledger_of))"

need_renamed() { printf '[{"filename":"src/web/x.ts","previous_filename":"%s","status":"renamed"}]' "$1" >"$tmp/files.json"; PATH="$tmp/bin:$PATH" FX="$tmp" REPO=o/r PR_NUMBER=7 ROOT="$tmp/root" python3 -I "$tool" deep-needed 2>/dev/null; }
[ "$(need_renamed src/api/internal/identity/a.go)" = "true" ] && ok "deep-needed: a rename out of an identity path is true" || ko "deep-needed rename"


# --- the register detail: every entry carries detail and fix, the script owns first_sha, source, location ---
# A tiny git repository stands for the checkout of the reviewed commit: a.go has 10 lines.
repo="$tmp/locrepo"
mkdir -p "$repo"
git -C "$repo" init -q
seq 1 10 | sed 's/^/line /' >"$repo/a.go"
git -C "$repo" add a.go
git -C "$repo" -c user.name=t -c user.email=t@example.invalid commit -q -m init
locsha="$(git -C "$repo" rev-parse HEAD)"
old_sha=1111111111111111111111111111111111111111

# mkreview ENTRIES_JSON: the model's REVIEW_JSON (verdict, text, fenced ledger with a wrong sha).
mkreview() {
  python3 -I - "$1" <<'EOF2'
import json, sys
body = "Triage: MEDIUM | blast radius: x\n\ntext\n\n<details><summary>Review ledger</summary>\n\n```json\n" + \
       json.dumps({"sha": "x", "findings": json.loads(sys.argv[1])}) + "\n```\n\n</details>"
print(json.dumps({"body": body}))
EOF2
}
# entry_of ID: one entry of the posted ledger as JSON (or "NONE").
entry_of() { posted_body | python3 -I -c '
import json, re, sys
m = re.search(r"```json\s*(\{.*?\})\s*```", sys.stdin.read(), re.S)
d = json.loads(m.group(1)) if m else {"findings": []}
print(json.dumps(next((f for f in d["findings"] if f["id"] == sys.argv[1]), "NONE")))' "$1"; }
pyassert() { python3 -I -c "$1" "$2"; }
E_GOOD='{"id":"g1","kind":"finding","evidence":"OBSERVED","sev":"MEDIUM","file":"a.go","line":3,"scenario":"s","status":"open","detail":"proof of it","fix":"do this"}'

comments '[]'
HEAD_SHA="$locsha" LOCATION_REPO="$repo" JSON="$(mkreview "[$E_GOOD]")" run
[ "$rc" -eq 0 ] && pyassert 'import json,sys; f=json.loads(sys.argv[1]); assert f["detail"]=="proof of it" and f["fix"]=="do this" and f["source"]=="review" and f["first_sha"]=="'"$locsha"'" and f["location"]=="verified"' "$(entry_of g1)" 2>/dev/null && ok "an entry with detail and fix is kept; the script sets source review, first_sha (the head) and location verified" || ko "good entry ($(entry_of g1))"

# A missing or empty detail or fix NEVER lowers or hides the verdict: it is filled with a marker and the comment names the id.
comments '[]'
E_NOFIX='{"id":"nofix","kind":"finding","evidence":"OBSERVED","sev":"HIGH","file":"a.go","line":3,"scenario":"s","status":"open","detail":"proof"}'
HEAD_SHA="$locsha" JSON="$(mkreview "[$E_GOOD,$E_NOFIX]")" run
pb="$(posted_body)"
[ "$(sed -n 1p <<<"$pb")" = "Triage: HIGH | blast radius: x" ] && [ "$(sed -n 2p <<<"$pb")" = "<!-- 226os-ai-review -->" ] && grep -q 'Detail or fix not recorded by the review (the verdict is unchanged): `nofix`\.' <<<"$pb" && ok "a missing fix keeps the verdict (raised to the open HIGH entry's level) and every entry, and the comment names the id" || ko "missing fix ($pb)"
pyassert 'import json,sys; f=json.loads(sys.argv[1]); assert f["detail"]=="proof" and f["fix"]=="not recorded by the review" and f["sev"]=="HIGH"' "$(entry_of nofix)" 2>/dev/null && pyassert 'import json,sys; f=json.loads(sys.argv[1]); assert f["fix"]=="do this"' "$(entry_of g1)" 2>/dev/null && ok "the missing field is filled with the not-recorded marker, the complete entry is untouched" || ko "marker ($(entry_of nofix))"
HEAD_SHA="$locsha" JSON="$(mkreview '[{"id":"nodetail","kind":"note","evidence":"OBSERVED","sev":"LOW","file":"a.go","line":3,"scenario":"s","status":"open","detail":"   ","fix":"none: accept"}]')" run
pyassert 'import json,sys; f=json.loads(sys.argv[1]); assert f["detail"]=="not recorded by the review" and f["fix"]=="none: accept"' "$(entry_of nodetail)" 2>/dev/null && ok "a blank detail is a missing detail, filled" || ko "blank detail ($(posted_body))"
HEAD_SHA="$locsha" JSON="$(mkreview '[{"id":"bad|id","kind":"note","evidence":"OBSERVED","sev":"LOW","scenario":"s","status":"open"}]')" run
pb="$(posted_body)"
[ "$(sed -n 1p <<<"$pb")" = "Triage: MEDIUM | blast radius: x" ] && grep -q '`bad-id`' <<<"$pb" && ok "the id in the notice is a slug: no pipe, newline or marker can ride on it" || ko "id slug ($pb)"
# RED of the review of PR 187: a first CRITICAL with one entry lacking fix was posted NOT RUN and the gate failed open.
mkcrit() {
  python3 -I - <<'EOF2'
import json
body = "Triage: CRITICAL | blast radius: x\n\nOBSERVED: a.go:3 literal key.\n\n<details><summary>Review ledger</summary>\n\n```json\n" + \
       json.dumps({"sha": "x", "findings": [{"id": "key", "kind": "finding", "evidence": "OBSERVED", "sev": "CRITICAL", "file": "a.go", "line": 3, "scenario": "s", "status": "open", "detail": "proof"}]}) + "\n```\n\n</details>"
print(json.dumps({"body": body}))
EOF2
}
comments '[]'
HEAD_SHA="$locsha" JSON="$(mkcrit)" run
pb="$(posted_body)"
[ "$(sed -n 1p <<<"$pb")" = "Triage: CRITICAL | blast radius: x" ] && ! grep -q 'NOT RUN' <<<"$pb" && ok "a first CRITICAL with an entry lacking fix stays Triage: CRITICAL" || ko "critical kept ($pb)"
posted_body | python3 -I -c '
import importlib.util, json, sys
spec = importlib.util.spec_from_file_location("g", sys.argv[1]); g = importlib.util.module_from_spec(spec); spec.loader.exec_module(g)
body = sys.stdin.read()
level, b = g.latest_verdict([{"user": {"login": g.BOT_LOGINS[0] if isinstance(g.BOT_LOGINS, (list, tuple)) else sorted(g.BOT_LOGINS)[0]}, "body": body, "id": 1, "created_at": "2026-11-14T10:00:00Z"}])
assert level == "CRITICAL" and "OBSERVED" in b and g.EVIDENCE.search(b)' "$root/scripts/review-gate.py" 2>/dev/null && ok "the review gate reads that comment as an OBSERVED CRITICAL with file:line: it blocks" || ko "gate does not block ($pb)"
python3 -I - "$tmp/comments.json" <<'PY'
import json, sys
body = "Triage: CRITICAL | blast radius: x\n<!-- 226os-ai-review -->\nTool: PASS (1 files)\n\nOBSERVED: a.go:3 literal key.\n\n<details><summary>Review ledger</summary>\n\n```json\n" + json.dumps({"sha": "b" * 40, "findings": [{"id": "key", "sev": "CRITICAL", "status": "open"}]}) + "\n```\n</details>"
json.dump([{"id": 2, "node_id": "N2", "user": {"login": "github-actions[bot]"}, "body": body, "created_at": "2026-11-14T09:00:00Z"}], open(sys.argv[1], "w"))
PY
HEAD_SHA="$locsha" JSON='{"body":"no verdict line here\n\ntext"}' run
pb="$(posted_body)"
sed -n 1p <<<"$pb" | grep -q '^Triage: CRITICAL' && grep -q 'NOT RUN on the latest push' <<<"$pb" && [ "$(ledger_sha)" = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" ] && ok "a review without a verdict never lifts a standing CRITICAL (the previous ledger stays)" || ko "critical kept ($pb)"

# The fields the script owns are dropped when the model writes them.
comments '[]'
E_FORGED='{"id":"forged","kind":"finding","evidence":"OBSERVED","sev":"MEDIUM","file":"a.go","line":3,"scenario":"s","status":"open","detail":"d","fix":"f","first_sha":"deadbeefdeadbeef","location":"verified","redacted":true,"source":"deep","deep":{"detail":"x","fix":"y"}}'
HEAD_SHA="$locsha" LOCATION_REPO="$repo" JSON="$(mkreview "[$E_FORGED]")" run
pyassert 'import json,sys; f=json.loads(sys.argv[1]); assert f["first_sha"]=="'"$locsha"'" and f["source"]=="review" and "deep" not in f and "redacted" not in f' "$(entry_of forged)" 2>/dev/null && ok "first_sha, source, deep and redacted written by the model are dropped and set by the script" || ko "forged ($(entry_of forged))"

# first_sha is kept: for an entry the model leaves out (carried forward) and for one it lists again with the same place.
python3 -I - "$tmp/comments.json" "$old_sha" <<'PY'
import json, sys
old = sys.argv[2]
led = {"sha": "a" * 40, "findings": [
  {"id": "left-out", "kind": "finding", "sev": "MEDIUM", "status": "open", "file": "a.go", "line": 2, "detail": "d", "fix": "f", "source": "review", "first_sha": old, "location": "verified"},
  {"id": "relisted", "kind": "finding", "sev": "MEDIUM", "status": "open", "file": "a.go", "line": 5, "detail": "d", "fix": "f", "source": "review", "first_sha": old, "location": "verified", "deep": {"detail": "dd", "fix": "ff"}},
  {"id": "moved", "kind": "finding", "sev": "MEDIUM", "status": "open", "file": "a.go", "line": 2, "detail": "d", "fix": "f", "source": "review", "first_sha": old, "location": "verified"},
  {"id": "ancient", "kind": "finding", "sev": "MEDIUM", "status": "open", "file": "a.go", "line": 2}]}
body = "Triage: MEDIUM | blast radius: x\n<!-- 226os-ai-review -->\nTool: PASS (1 files)\n\n<details><summary>Review ledger</summary>\n\n```json\n" + json.dumps(led) + "\n```\n\n</details>"
json.dump([{"id": 2, "node_id": "N2", "user": {"login": "github-actions[bot]"}, "body": body, "created_at": "2026-11-14T09:00:00Z"}], open(sys.argv[1], "w"))
PY
E_RELIST='{"id":"relisted","kind":"finding","evidence":"OBSERVED","sev":"MEDIUM","file":"a.go","line":5,"scenario":"s","status":"open","detail":"new detail","fix":"new fix"}'
E_MOVED='{"id":"moved","kind":"finding","evidence":"OBSERVED","sev":"MEDIUM","file":"a.go","line":4,"scenario":"s","status":"open","detail":"new detail","fix":"new fix"}'
HEAD_SHA="$locsha" LOCATION_REPO="$repo" JSON="$(mkreview "[$E_RELIST,$E_MOVED]")" run
pyassert 'import json,sys; f=json.loads(sys.argv[1]); assert f["first_sha"]=="'"$old_sha"'" and f["detail"]=="new detail" and f["deep"]["detail"]=="dd"' "$(entry_of relisted)" 2>/dev/null && ok "an entry listed again with the same place keeps its first_sha and its deep sub-field (the new detail replaces the old)" || ko "relisted ($(entry_of relisted))"
pyassert 'import json,sys; f=json.loads(sys.argv[1]); assert f["first_sha"]=="'"$locsha"'" and f["location"]=="verified"' "$(entry_of moved)" 2>/dev/null && ok "an entry whose file or line changed gets the head as first_sha (the permalink stays right)" || ko "moved ($(entry_of moved))"
pyassert 'import json,sys; f=json.loads(sys.argv[1]); assert f["first_sha"]=="'"$old_sha"'" and f["line"]==2' "$(entry_of left-out)" 2>/dev/null && ok "an entry the model left out is carried forward with its first_sha" || ko "left-out ($(entry_of left-out))"
pyassert 'import json,sys; f=json.loads(sys.argv[1]); assert "first_sha" not in f and "detail" not in f' "$(entry_of ancient)" 2>/dev/null && ok "an older entry (no detail, no first_sha) is carried as it is: nothing is invented for it" || ko "ancient ($(entry_of ancient))"

# Location: file exists at the reviewed commit and the line is inside it, else unverified.
comments '[]'
mk() { python3 -I -c 'import json,sys; print(json.dumps({"id":sys.argv[1],"kind":"finding","evidence":"OBSERVED","sev":"MEDIUM","file":sys.argv[2],"line":int(sys.argv[3]),"scenario":"s","status":"open","detail":"d","fix":"f"}))' "$@"; }
HEAD_SHA="$locsha" LOCATION_REPO="$repo" JSON="$(mkreview "[$(mk in a.go 10),$(mk past a.go 11),$(mk gone nope.go 1),$(mk nofile '' 1),$(mk dots ../x 1),$(mk fileonly a.go 0)]")" run
loc() { entry_of "$1" | python3 -I -c 'import json,sys; print(json.load(sys.stdin)["location"])'; }
[ "$(loc in)" = verified ] && [ "$(loc past)" = unverified ] && [ "$(loc gone)" = unverified ] && [ "$(loc nofile)" = unverified ] && [ "$(loc dots)" = unverified ] && [ "$(loc fileonly)" = verified ] && ok "location: last line verified; a line past the end, a missing file, no file and a ../ path are unverified; a file without a line is verified" || ko "location ($(loc in) $(loc past) $(loc gone) $(loc nofile) $(loc dots) $(loc fileonly))"
HEAD_SHA="$locsha" LOCATION_REPO="$tmp/not-a-repo" JSON="$(mkreview "[$(mk in a.go 3)]")" run
[ "$(loc in)" = unverified ] && ok "location: no repository to read the commit from is unverified (never a guess)" || ko "no repo ($(loc in))"

# Safety: the same neutraliser as titles, plus HTML; a secret shape is replaced and flagged. The secret-shaped strings
# are built from pieces here so that this file itself holds none (the secret scan reads it).
comments '[]'
python3 -I - "$tmp/safe.json" <<'PY'
import json, sys
secrets = {
  "pem": "-----BEGIN " + "RSA PRIVATE KEY-----",
  "aws": "AKI" + "A" + "ABCDEFGHIJKLMNOP",
  "ghp": "gh" + "p_" + "a1B2" * 9,
  "slack": "xox" + "b-" + "1234567890-abcdefghij",
  "sk": "s" + "k-" + "abcdefghijklmnopqrstuvwx",
  "jwt": "ey" + "JhbGciOiJIUzI1NiJ9." + "ey" + "JzdWIiOiIxMjM0NTY3ODkwIn0." + "abcdEFGH12345678",
  "auth": "Authoriz" + "ation: Bearer abcdef1234567890xyz",
}
def entry(i, kind="finding", sev="MEDIUM", **kw):
    e = {"id": i, "kind": kind, "evidence": "OBSERVED", "sev": sev, "file": "a.go", "line": 1, "scenario": "s", "status": "open"}
    e.update(kw)
    return e
entries = [entry("mention", detail="@Y4NN777 fixes #172 and closes #5 <b>bold</b> <!-- finding:PR1#z --> <script>x</script>\nsecond line", fix="line one\nline two with #9"),
           entry("long", "note", "LOW", detail="x" * 900, fix="y" * 500),
           entry("benign", "note", "LOW", detail="the Authorization header is logged and the token sk is mentioned", fix="none: accept")]
for name, val in secrets.items():
    entries.append(entry("sec-" + name, detail="the value is " + val + " in the log", fix="rotate it"))
entries.append(entry("sec-fix", detail="plain", fix="use " + secrets["ghp"]))
body = "Triage: MEDIUM | blast radius: x\n\ntext\n\n<details><summary>Review ledger</summary>\n\n```json\n" + json.dumps({"sha": "x", "findings": entries}) + "\n```\n\n</details>"
json.dump({"body": body}, open(sys.argv[1], "w"))
PY
HEAD_SHA="$locsha" LOCATION_REPO="$repo" JSON="$(cat "$tmp/safe.json")" run
pyassert 'import json,sys; f=json.loads(sys.argv[1]); d=f["detail"]; x=f["fix"]; assert "@Y4NN777" not in d and "@​Y4NN777" in d and "#172" not in d and "#​172" in d and "#​5" in d; assert "<" not in d and "-->" not in d; assert "\n" not in d and "\n" not in x and "#​9" in x' "$(entry_of mention)" 2>/dev/null && ok "detail and fix: @mention and #N disarmed with a zero-width space, no HTML or marker, one line" || ko "neutralise ($(entry_of mention))"
pyassert 'import json,sys; f=json.loads(sys.argv[1]); assert len(f["detail"])<=600 and len(f["fix"])<=300' "$(entry_of long)" 2>/dev/null && ok "detail is cut to 600 characters and fix to 300" || ko "length ($(entry_of long))"
pyassert 'import json,sys; f=json.loads(sys.argv[1]); assert "Authorization header" in f["detail"] and "redacted" not in f' "$(entry_of benign)" 2>/dev/null && ok "no false positive: the word Authorization or sk without a value is kept" || ko "benign ($(entry_of benign))"
allred=1
for k in pem aws ghp slack sk jwt auth; do
  pyassert 'import json,sys; f=json.loads(sys.argv[1]); assert f["detail"]=="[redacted: looked like a secret]" and f["redacted"] is True and f["fix"]=="rotate it"' "$(entry_of "sec-$k")" 2>/dev/null || { allred=0; ko "secret shape $k not redacted ($(entry_of "sec-$k"))"; }
done
[ "$allred" -eq 1 ] && ok "private key block, AWS key, GitHub token, Slack token, sk- key, JWT and Authorization header values are replaced and flagged" || true
pyassert 'import json,sys; f=json.loads(sys.argv[1]); assert f["fix"]=="[redacted: looked like a secret]" and f["detail"]=="plain" and f["redacted"] is True' "$(entry_of sec-fix)" 2>/dev/null && ok "a secret shape in the fix is replaced too" || ko "fix secret ($(entry_of sec-fix))"

# --- deep: detail and fix are required; a repeated id keeps its source (the #180 case) ---
HEAD_SHA="$locsha"
note_base='Triage: NONE | blast radius: x\n<!-- 226os-ai-review -->\nTool: PASS (1 files)\n\n<details><summary>Review ledger</summary>\n\n```json\n{\"sha\":\"'$locsha'\",\"findings\":[{\"id\":\"n180\",\"kind\":\"note\",\"evidence\":\"INFERRED\",\"sev\":\"LOW\",\"file\":\"a.go\",\"line\":3,\"scenario\":\"stage 1 remark\",\"status\":\"open\",\"detail\":\"stage 1 detail\",\"fix\":\"none: accept\",\"source\":\"review\",\"first_sha\":\"'$old_sha'\",\"location\":\"verified\"},{\"id\":\"h1\",\"kind\":\"finding\",\"evidence\":\"OBSERVED\",\"sev\":\"HIGH\",\"file\":\"a.go\",\"line\":4,\"scenario\":\"hh\",\"status\":\"open\",\"detail\":\"stage 1 high\",\"fix\":\"fix h1\",\"source\":\"review\",\"first_sha\":\"'$old_sha'\",\"location\":\"verified\"}]}\n```\n\n</details>'
DEEP180='{"level":"MEDIUM","blast_radius":"x","summary":"s","findings":[{"id":"n180","kind":"finding","evidence":"OBSERVED","sev":"MEDIUM","file":"a.go","line":7,"scenario":"deep scenario","status":"open","detail":"deep detail","fix":"deep fix"},{"id":"h1","kind":"note","evidence":"INFERRED","sev":"LOW","file":"a.go","line":4,"scenario":"s","status":"open","detail":"deep says less","fix":"none"}]}'
comments "[{\"id\":2,\"node_id\":\"N2\",\"user\":{\"login\":\"$BOT\"},\"body\":\"$note_base\",\"created_at\":\"2026-11-14T09:00:00Z\"}]"
LOCATION_REPO="$repo" DJSON="$DEEP180" deep
posted_entry() { posted | python3 -I -c 'import json,re,sys; m=re.search(r"```json\s*(\{.*?\})\s*```", sys.stdin.read(), re.S); d=json.loads(m.group(1)); print(json.dumps(next(f for f in d["findings"] if f["id"]==sys.argv[1])))' "$1"; }
pyassert 'import json,sys; f=json.loads(sys.argv[1]); assert f["source"]=="review" and f["sev"]=="MEDIUM" and f["kind"]=="finding" and f["detail"]=="stage 1 detail" and f["first_sha"]=="'"$old_sha"'" and f["line"]==3; d=f["deep"]; assert d["detail"]=="deep detail" and d["fix"]=="deep fix" and d["sev"]=="MEDIUM" and d["line"]==7 and d["location"]=="verified"' "$(posted_entry n180)" 2>/dev/null && ok "#180: a stage 1 note repeated by the deep review keeps source review, gains a deep sub-field and takes the worse severity" || ko "180 case ($(posted_entry n180))"
pyassert 'import json,sys; f=json.loads(sys.argv[1]); assert f["sev"]=="HIGH" and f["source"]=="review" and f["deep"]["fix"]=="none"' "$(posted_entry h1)" 2>/dev/null && ok "the worse severity wins in both directions (a deep LOW never lowers a stage 1 HIGH)" || ko "worst sev ($(posted_entry h1))"
first="$(posted)"
comments "$(python3 -I -c 'import json,sys; print(json.dumps([{"id":2,"node_id":"N2","user":{"login":"github-actions[bot]"},"body":sys.argv[1],"created_at":"2026-11-14T09:00:00Z"}]))' "$first")"
LOCATION_REPO="$repo" DJSON="$DEEP180" deep
[ "$(posted)" = "$first" ] && ok "the merge with a repeated id is idempotent" || ko "idempotent sub-field ($(posted | ledger_of))"
comments "[{\"id\":2,\"node_id\":\"N2\",\"user\":{\"login\":\"$BOT\"},\"body\":\"$note_base\",\"created_at\":\"2026-11-14T09:00:00Z\"}]"
DJSON='{"level":"NONE","blast_radius":"x","summary":"s","findings":[{"id":"nofix","kind":"note","evidence":"INFERRED","sev":"LOW","file":"a.go","line":1,"scenario":"s","status":"open","detail":"only a detail"}]}' deep
[ "$(posted | sed -n 4p)" = "Deep: PASS (0 findings, 1 note)" ] && posted | grep -q 'Detail or fix not recorded by the review (the verdict is unchanged): `nofix`\.' && pyassert 'import json,sys; f=json.loads(sys.argv[1]); assert f["fix"]=="not recorded by the review" and f["detail"]=="only a detail"' "$(posted_entry nofix)" 2>/dev/null && ok "a deep entry without fix is filled and named, the Deep line is not NOT RUN" || ko "deep missing ($(posted | sed -n 1,6p))"
# Review of 8f28cca: the level is worked out again from the FINAL merged ledger (down as well as up).
carried_base() { # LEVEL STATUS: a stage 1 comment carrying one deep HIGH entry
  python3 -I -c '
import json, sys
led = {"sha": sys.argv[3], "findings": [{"id": "dh", "kind": "finding", "evidence": "OBSERVED", "sev": "HIGH", "file": "a.go", "line": 3, "scenario": "s", "status": sys.argv[2], "detail": "d", "fix": "f", "source": "deep", "first_sha": sys.argv[3], "location": "verified"}]}
body = "Triage: %s | blast radius: x\n<!-- 226os-ai-review -->\nTool: PASS (1 files)\n\n<details><summary>Review ledger</summary>\n\n```json\n%s\n```\n\n</details>" % (sys.argv[1], json.dumps(led))
print(json.dumps([{"id": 2, "node_id": "N2", "user": {"login": "github-actions[bot]"}, "body": body, "created_at": "2026-11-14T09:00:00Z"}]))' "$1" "$2" "$locsha"
}
# stage1_then_deep PENDING: stage 1 (a model that declares NONE, no entry) over a previous comment carrying an open deep HIGH,
# then that posted comment becomes the base the deep stage merges into (same head).
stage1_posted_as_comment() {
  posted_body | python3 -I -c 'import json,sys; print(json.dumps([{"id":2,"node_id":"N2","user":{"login":"github-actions[bot]"},"body":sys.stdin.read(),"created_at":"2026-11-14T10:00:00Z"}]))'
}
stage1_then_deep() { # PENDING DJSON
  comments "$(carried_base HIGH open)"
  HEAD_SHA="$locsha" DEEP_PENDING="$1" JSON="$(mkreview '[]' | sed 's/Triage: MEDIUM/Triage: NONE/')" run
  s1_line="$(posted_body | sed -n 1p)"
  comments "$(stage1_posted_as_comment)"
  HEAD_SHA="$locsha" DJSON="$2" deep
}
# Replacement design (review of bb0abb0): deep never lowers; stage 1 leaves carried deep entries out while the deep stage runs.
FIXED_DH='{"level":"NONE","blast_radius":"x","summary":"s","findings":[{"id":"dh","kind":"finding","evidence":"OBSERVED","sev":"HIGH","file":"a.go","line":3,"scenario":"s","status":"fixed","detail":"d","fix":"f"}]}'
stage1_then_deep "" "$FIXED_DH"
[ "$s1_line" = "Triage: HIGH | blast radius: x" ] && ok "no deep stage in this run: a carried open deep HIGH keeps stage 1 at HIGH" || ko "deep not pending ($s1_line)"
# Exhaustive table of the level invariants I1-I4 (scripts/review-comment.py, LEVEL SPEC): every declared level x stage 1
# ledger x carried deep entry x deep run, with the gate's reading of each comment; it replaces the one-off level cases.
tbl="$(python3 -I "$here/review-comment-levels.py" 2>&1)" || true
tail_line="$(tail -1 <<<"$tbl")"
[ "$tail_line" = "cases=504 failures=0" ] && ok "level table: $tail_line (declared x stage 1 ledger x carried x deep run: I1-I4 and the gate's reading)" || ko "level table ($(head -5 <<<"$tbl"))"
# The carried counter rises by exactly 1 per review run: a deep entry is counted by the deep merge when the deep stage runs
# in that run, by stage 1 only when none runs.
carried_n() { python3 -I -c 'import json,sys; print(json.loads(sys.argv[1]).get("carried"))' "$(posted_entry dh)"; }
as_previous() { comments "$(python3 -I -c 'import json,sys; print(json.dumps([{"id":2,"node_id":"N2","user":{"login":"github-actions[bot]"},"body":sys.argv[1],"created_at":"2026-11-14T11:00:00Z"}]))' "$(posted)")"; }
OMIT='{"level":"NONE","blast_radius":"x","summary":"s","findings":[]}'
stage1_then_deep true "$OMIT"; n1="$(carried_n)"
as_previous
HEAD_SHA="$locsha" DEEP_PENDING=true JSON="$(mkreview '[]' | sed 's/Triage: MEDIUM/Triage: NONE/')" run
comments "$(stage1_posted_as_comment)"
HEAD_SHA="$locsha" DJSON="$OMIT" deep; n2="$(carried_n)"
[ "$n1" = "1" ] && [ "$n2" = "2" ] && ok "carried counter, deep stage runs: 1 after one run, 2 after two (the deep merge counts it, stage 1 does not)" || ko "carried pending ($n1, $n2)"
comments "$(carried_base HIGH open)"
HEAD_SHA="$locsha" JSON="$(mkreview '[]' | sed 's/Triage: MEDIUM/Triage: NONE/')" run; m1="$(carried_n)"
as_previous
HEAD_SHA="$locsha" JSON="$(mkreview '[]' | sed 's/Triage: MEDIUM/Triage: NONE/')" run; m2="$(carried_n)"
[ "$m1" = "1" ] && [ "$m2" = "2" ] && ok "carried counter, no deep stage: 1 after one run, 2 after two (stage 1 counts it)" || ko "carried no deep ($m1, $m2)"
# Review of 21a8a5e: the level goes DOWN only when a review explicitly marks an entry fixed.
mkbase() { # LEVEL FINDINGS_JSON: a stage 1 comment
  python3 -I -c '
import json, sys
led = {"sha": sys.argv[3], "findings": json.loads(sys.argv[2])}
body = "Triage: %s | blast radius: x\n<!-- 226os-ai-review -->\nTool: PASS (1 files)\n\n<details><summary>Review ledger</summary>\n\n```json\n%s\n```\n\n</details>" % (sys.argv[1], json.dumps(led))
print(json.dumps([{"id": 2, "node_id": "N2", "user": {"login": "github-actions[bot]"}, "body": body, "created_at": "2026-11-14T09:00:00Z"}]))' "$1" "$2" "$locsha"
}
CLEAN='{"level":"NONE","blast_radius":"x","summary":"s","findings":[]}'
comments "$(mkbase MEDIUM '[]')"
DJSON="$CLEAN" deep
[ "$(posted | sed -n 1p)" = "Triage: MEDIUM | blast radius: x" ] && ok "deep: a stage 1 MEDIUM with an empty ledger stays MEDIUM after a clean deep review (unexplained level kept)" || ko "empty ledger lowered ($(posted | sed -n 1,3p))"
comments "$(mkbase HIGH '[{"id":"nt","kind":"note","evidence":"INFERRED","sev":"LOW","file":"a.go","line":1,"scenario":"s","status":"open","detail":"d","fix":"f","source":"review"}]')"
DJSON="$CLEAN" deep
[ "$(posted | sed -n 1p)" = "Triage: HIGH | blast radius: x" ] && ok "deep: a stage 1 HIGH backed only by a note stays HIGH after a clean deep review" || ko "note-only lowered ($(posted | sed -n 1,3p))"
comments '[{"id":2,"node_id":"N2","user":{"login":"github-actions[bot]"},"body":"Triage: HIGH | blast radius: x\n<!-- 226os-ai-review -->\nTool: PASS (1 files)\n\nno ledger at all","created_at":"2026-11-14T09:00:00Z"}]'
CID=2 DJSON="$CLEAN" deep
[ "$(posted | sed -n 1p)" = "Triage: HIGH | blast radius: x" ] && ok "deep: a stage 1 HIGH with no ledger at all stays HIGH after a clean deep review" || ko "no ledger lowered ($(posted | sed -n 1,3p))"
NOTRUN_FIXED='{"level":"NONE","blast_radius":"x","summary":"s","findings":[{"id":"dh","kind":"finding","evidence":"OBSERVED","sev":"HIGH","file":"a.go","line":3,"scenario":"s","status":"fixed","detail":"d","fix":"f"},{"kind":"note","detail":"no id"}]}'
stage1_then_deep true "$NOTRUN_FIXED"
[ "$(posted | sed -n 1p)" = "Triage: HIGH | blast radius: x" ] && posted | grep -q '^Deep: NOT RUN' && pyassert 'import json,sys; assert json.loads(sys.argv[1])["status"]=="open"' "$(posted_entry dh)" 2>/dev/null && ok "deep NOT RUN with a fixed entry: the ledger keeps it open and the line goes up to HIGH, never down" || ko "not run fixed ($(posted | sed -n 1,5p))"
# next run (no deep stage): the open entry is still carried and stage 1 stays HIGH
comments "$(python3 -I -c 'import json,sys; print(json.dumps([{"id":2,"node_id":"N2","user":{"login":"github-actions[bot]"},"body":sys.argv[1],"created_at":"2026-11-14T11:00:00Z"}]))' "$(posted)")"
HEAD_SHA=bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb JSON="$(mkreview '[]' | sed 's/Triage: MEDIUM/Triage: NONE/')" run
[ "$(posted_body | sed -n 1p)" = "Triage: HIGH | blast radius: x" ] && ok "the next run still carries it open: HIGH (the NOT RUN did not lower the level one run late)" || ko "next run ($(posted_body | sed -n 1,3p))"
# a stage 1 level that nothing explains survives a deep fixed entry (no floor arithmetic to fool)
comments "$(carried_base HIGH open)"
HEAD_SHA="$locsha" DEEP_PENDING=true JSON="$(mkreview '[]' | sed 's/Triage: MEDIUM/Triage: HIGH/')" run
comments "$(stage1_posted_as_comment)"
HEAD_SHA="$locsha" DJSON="$FIXED_DH" deep
[ "$(posted | sed -n 1p)" = "Triage: HIGH | blast radius: x" ] && ok "an unexplained stage 1 HIGH written this run stays HIGH after the deep run fixes a carried entry" || ko "unexplained high ($(posted | sed -n 1,3p))"
comments "$(carried_base NONE fixed)"
DJSON='{"level":"NONE","blast_radius":"deep blast","summary":"s","findings":[{"id":"dn","kind":"finding","evidence":"OBSERVED","sev":"HIGH","file":"a.go","line":2,"scenario":"s","status":"open","detail":"d","fix":"f"}]}' deep
[ "$(posted | sed -n 1p)" = "Triage: HIGH | blast radius: deep blast" ] && ok "deep: a new open OBSERVED HIGH still raises the Triage line" || ko "deep raise again ($(posted | sed -n 1,4p))"
# Status is case-insensitive wherever entries are read.
comments '[]'
HEAD_SHA="$locsha" JSON="$(mkreview '[{"id":"cs","kind":"finding","evidence":"OBSERVED","sev":"HIGH","file":"a.go","line":3,"scenario":"s","status":" Open ","detail":"d","fix":"f"}]' | sed 's/Triage: MEDIUM/Triage: NONE/')" run
[ "$(posted_body | sed -n 1p)" = "Triage: HIGH | blast radius: x" ] && pyassert 'import json,sys; assert json.loads(sys.argv[1])["status"]=="open"' "$(entry_of cs)" 2>/dev/null && ok "stage 1: a status written ' Open ' counts as open and is stored as open" || ko "status case ($(posted_body | sed -n 1,3p))"
comments "$(carried_base HIGH OPEN)"
HEAD_SHA="$locsha" JSON="$(mkreview '[]' | sed 's/Triage: MEDIUM/Triage: NONE/')" run
posted_body | grep -q 'Carried forward without a new decision.*`dh`' && ok "carry_forward: an OPEN (upper case) entry is carried and named as undecided" || ko "carry case ($(posted_body | sed -n 1,8p))"
comments "[{\"id\":2,\"node_id\":\"N2\",\"user\":{\"login\":\"$BOT\"},\"body\":\"$note_base\",\"created_at\":\"2026-11-14T09:00:00Z\"}]"

# Review of PR 187 (deep-critical-kept-without-evidence): the level is derived from the kept entries, never the declared one.
DJSON='{"level":"CRITICAL","blast_radius":"x","summary":"s","findings":[{"id":"inf","kind":"finding","evidence":"INFERRED","sev":"CRITICAL","file":"a.go","line":1,"scenario":"s","status":"open","detail":"d","fix":"f"},{"id":"bad","kind":"finding","sev":"CRITICAL"}]}' deep
[ "$(posted | sed -n 1p)" = "Triage: HIGH | blast radius: x" ] && ! posted | sed -n 1p | grep -q CRITICAL && posted | grep -q 'it claimed CRITICAL: a human must read the deep review' && ok "deep: a declared CRITICAL whose only CRITICAL entry is INFERRED (lowered to HIGH) plus a dropped entry is HIGH, and the dropped claim is named" || ko "deep critical no evidence ($(posted | sed -n 1,6p))"
DJSON='{"level":"NONE","blast_radius":"x","summary":"s","findings":[{"id":"real","kind":"finding","evidence":"OBSERVED","sev":"HIGH","file":"a.go","line":1,"scenario":"s","status":"open","detail":"d","fix":"f"}]}' deep
[ "$(posted | sed -n 1p)" = "Triage: HIGH | blast radius: x" ] && ok "deep: a declared NONE with a valid open OBSERVED HIGH entry is HIGH" || ko "deep raise ($(posted | sed -n 1,3p))"
comments '[]'
HEAD_SHA="$locsha" JSON="$(python3 -I - <<'EOF2'
import json
led = {"sha": "x", "findings": [{"id": "k", "kind": "finding", "evidence": "INFERRED", "sev": "CRITICAL", "file": "a.go", "line": 3, "scenario": "s", "status": "open", "detail": "d", "fix": "f"}]}
print(json.dumps({"body": "Triage: CRITICAL | blast radius: x\n\ntext\n\n<details><summary>Review ledger</summary>\n\n```json\n" + json.dumps(led) + "\n```\n\n</details>"}))
EOF2
)" run
[ "$(posted_body | sed -n 1p)" = "Triage: HIGH | blast radius: x" ] && ok "stage 1: a declared CRITICAL without an OBSERVED CRITICAL entry with file:line is HIGH" || ko "stage 1 critical cap ($(posted_body | sed -n 1,3p))"
HEAD_SHA="$locsha" JSON="$(mkreview "[$E_GOOD]" | sed 's/Triage: MEDIUM/Triage: NONE/')" run
[ "$(posted_body | sed -n 1p)" = "Triage: MEDIUM | blast radius: x" ] && ok "stage 1: a declared NONE with an open MEDIUM finding is MEDIUM" || ko "stage 1 raise ($(posted_body | sed -n 1,3p))"
comments "[{\"id\":2,\"node_id\":\"N2\",\"user\":{\"login\":\"$BOT\"},\"body\":\"$note_base\",\"created_at\":\"2026-11-14T09:00:00Z\"}]"
DJSON='{"level":"CRITICAL","blast_radius":"x","summary":"s","findings":[{"id":"nofix","kind":"finding","evidence":"OBSERVED","sev":"CRITICAL","file":"a.go","line":1,"scenario":"s","status":"open","detail":"d"},{"kind":"note","detail":"no id"}]}' deep
[ "$(posted | sed -n 1p)" = "Triage: CRITICAL | blast radius: x" ] && posted | grep -q '^Deep: NOT RUN (the deep review output had an invalid entry (dropped))$' && ok "a deep CRITICAL with a missing field and an invalid entry still raises the Triage line to CRITICAL" || ko "deep critical ($(posted | sed -n 1,6p))"
LOCATION_REPO="$repo" DJSON='{"level":"NONE","blast_radius":"x","summary":"s","findings":[{"id":"new-deep","kind":"note","evidence":"INFERRED","sev":"LOW","file":"a.go","line":1,"scenario":"s","status":"open","detail":"mail me at @Y4NN777 #44","fix":"none: accept"}]}' deep
pyassert 'import json,sys; f=json.loads(sys.argv[1]); assert f["source"]=="deep" and f["first_sha"]=="'"$locsha"'" and "@​Y4NN777" in f["detail"] and "#​44" in f["detail"]' "$(posted_entry new-deep)" 2>/dev/null && ok "a new deep entry is tagged source deep with first_sha, and its detail goes through the same neutraliser" || ko "deep new ($(posted_entry new-deep))"


# The deep job forces its output with --json-schema: the schema must let detail and fix through (and require them),
# or every deep entry would lack them and the deep review would always be NOT RUN.
python3 -I - "$root/.github/workflows/code-review.yml" <<'EOF' && ok "code-review.yml: the deep --json-schema has detail and fix in the finding properties and in required" || ko "deep schema lacks detail or fix"
import json, re, sys
text = open(sys.argv[1], encoding="utf-8").read()
schemas = [json.loads(m) for m in re.findall(r"--json-schema '(\{.*?\})'\s*$", text, re.M)]
deep = [s for s in schemas if "findings" in s.get("properties", {})]
assert len(deep) == 1, "one deep schema expected"
item = deep[0]["properties"]["findings"]["items"]
assert "detail" in item["properties"] and "fix" in item["properties"], "properties"
assert item["properties"]["detail"].get("maxLength") == 600
assert "detail" in item["required"] and "fix" in item["required"], "required"
EOF
grep -q 'LOCATION_REPO' "$root/.github/workflows/code-review.yml" && ok "code-review.yml sets LOCATION_REPO for the location check" || ko "LOCATION_REPO not set in code-review.yml"


# --- location: the file and line exist, but is it what the text talks about? (the #172 case) ---
# #172 cited .github/workflows/code-review.yml:423 for `deep_main`, which lives in scripts/review-comment.py.
mm="$tmp/mmrepo"
mkdir -p "$mm/.github/workflows" "$mm/scripts"
python3 -I - "$mm" <<'PY'
import sys
root = sys.argv[1]
wf = ["step %d" % i for i in range(1, 601)]
wf[422] = "        run: python3 -I review-comment.py deep   # the merge step"
open(root + "/.github/workflows/code-review.yml", "w").write("\n".join(wf) + "\n")
py = ["x%d = %d" % (i, i) for i in range(1, 601)]
py[467] = "def deep_main():"
py[290] = "    return 0  # create_and_collapse"
open(root + "/scripts/review-comment.py", "w").write("\n".join(py) + "\n")
PY
git -C "$mm" init -q
git -C "$mm" add -A
git -C "$mm" -c user.name=t -c user.email=t@example.invalid commit -q -m mm
mmsha="$(git -C "$mm" rev-parse HEAD)"
mkt() { python3 -I -c 'import json,sys; print(json.dumps({"id":sys.argv[1],"kind":"note","evidence":"INFERRED","sev":"LOW","file":sys.argv[2],"line":int(sys.argv[3]),"scenario":"s","status":"open","detail":sys.argv[4],"fix":"none: accept"}))' "$@"; }
comments '[]'
HEAD_SHA="$mmsha" LOCATION_REPO="$mm" JSON="$(mkreview "[$(mkt n172 .github/workflows/code-review.yml 423 'the stale check in `deep_main` skips the edit'),$(mkt near scripts/review-comment.py 291 'the loop after `create_and_collapse` returns'),$(mkt nowhere scripts/review-comment.py 100 'the helper `no_such_function` is wrong'),$(mkt plain scripts/review-comment.py 100 'a plain sentence without identifiers'),$(mkt farwindow scripts/review-comment.py 100 'see `deep_main` below')]")" run
[ "$(loc n172)" = mismatch ] && pyassert 'import json,sys; assert json.loads(sys.argv[1])["probable"]=="scripts/review-comment.py:468"' "$(entry_of n172)" 2>/dev/null && ok "the #172 case: code-review.yml:423 cited for deep_main is a mismatch, the probable location is scripts/review-comment.py:468" || ko "172 ($(entry_of n172))"
[ "$(loc near)" = verified ] && [ "$(loc plain)" = verified ] && ok "an identifier within 40 lines of the cited line, or none named, stays verified" || ko "near/plain ($(loc near) $(loc plain))"
[ "$(loc nowhere)" = unverified ] && ok "an identifier found nowhere is unverified" || ko "nowhere ($(loc nowhere))"
[ "$(loc farwindow)" = mismatch ] && ok "an identifier 368 lines from the cited line is outside the window: mismatch" || ko "window ($(loc farwindow))"

# --- review notes of #172 (PR168) ---
# unreachable-code-after-return: main() ends with its return (no dead loop after it that refers to a name that does not exist).
python3 -I - "$tool" <<'EOF2' && ok "main() has no statement after its last return (the old minimize loop is gone)" || ko "dead code after return in main()"
import ast, sys
tree = ast.parse(open(sys.argv[1], encoding="utf-8").read())
main = next(n for n in tree.body if isinstance(n, ast.FunctionDef) and n.name == "main")
first = next(i for i, n in enumerate(main.body) if isinstance(n, ast.Return))
assert first == len(main.body) - 1, "%d statement(s) after the first top-level return" % (len(main.body) - 1 - first)
EOF2

# carry-forward-keeps-omitted-open: an entry the model left out is carried, marked `carried` and named in the comment.
python3 -I - "$tmp/comments.json" <<'PY'
import json, sys
led = {"sha": "a" * 40, "findings": [{"id": "gone-quiet", "kind": "finding", "sev": "HIGH", "status": "open", "file": "a.go", "line": 2, "detail": "d", "fix": "f", "source": "review", "first_sha": "1" * 40, "location": "verified", "carried": 1}]}
body = "Triage: HIGH | blast radius: x\n<!-- 226os-ai-review -->\nTool: PASS (1 files)\n\n<details><summary>Review ledger</summary>\n\n```json\n" + json.dumps(led) + "\n```\n\n</details>"
json.dump([{"id": 2, "node_id": "N2", "user": {"login": "github-actions[bot]"}, "body": body, "created_at": "2026-11-14T09:00:00Z"}], open(sys.argv[1], "w"))
PY
HEAD_SHA="$locsha" LOCATION_REPO="$repo" JSON="$(mkreview "[$E_GOOD]")" run
pyassert 'import json,sys; f=json.loads(sys.argv[1]); assert f["status"]=="open" and f["carried"]==2' "$(entry_of gone-quiet)" 2>/dev/null && grep -q 'Carried forward without a new decision.*gone-quiet' <<<"$(posted_body)" && ok "an open entry the model omits is carried with a count and named in the comment (fix, answer or accept it)" || ko "carried ($(entry_of gone-quiet) $(posted_body | head -12))"
pyassert 'import json,sys; assert "carried" not in json.loads(sys.argv[1])' "$(entry_of g1)" 2>/dev/null && ok "an entry the model lists itself carries no count (the model cannot write it)" || ko "carried on listed"

finish
