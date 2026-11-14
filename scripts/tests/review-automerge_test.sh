#!/usr/bin/env bash
# Tests for scripts/review-automerge.py with a fake `gh`: it merges (a merge commit, the evaluated head only) only when
# every condition holds, refuses with a reason otherwise, updates the other open pull requests into develop after a
# merge, writes only with the bot token, and writes nothing when the mode is off or report or the token is missing.
# Usage: scripts/tests/review-automerge_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
tool="$root/scripts/review-automerge.py"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin" "$tmp/fx"
export H=1111111111111111111111111111111111111111
OLD=2222222222222222222222222222222222222222
export BS=3333333333333333333333333333333333333333

# Fake gh: reads answer from fixtures by endpoint; every call is logged with the token it used.
cat >"$tmp/bin/gh" <<'GH'
#!/usr/bin/env bash
printf 'TOKEN=%s %s\n' "${GH_TOKEN:-}" "$*" >>"$FX/calls.log"
[ -z "${FAKE_GH_FAIL:-}" ] || { echo "gh: boom" >&2; exit 1; }
case "$*" in
  "api repos/o/r/commits/$H/pulls --paginate") cat "$FX/assoc.json" ;;
  "api repos/o/r/pulls/7 --paginate") cat "$FX/pull.json" ;;
  "api repos/o/r/pulls/7/files --paginate") cat "$FX/files.json" ;;
  "api repos/o/r/commits/$H/check-runs?per_page=100 --paginate") cat "$FX/checks.json" ;;
  "api repos/o/r/issues/7/comments --paginate") cat "$FX/comments.json" ;;
  "api repos/o/r/pulls/7/reviews --paginate") cat "$FX/reviews.json" ;;
  "api repos/o/r/pulls/7/commits --paginate") cat "$FX/commits.json" ;;
  "api repos/o/r/contents/.github/sensitive-paths.txt?ref=$BS --paginate") cat "$FX/contents.json" ;;
  "api repos/o/r/contents/.github/deep-review-paths.txt?ref=$BS --paginate") cat "$FX/deep-paths.json" ;;
  "api repos/o/r/contents/.github/pipeline-paths.txt?ref=$BS --paginate") cat "$FX/pipeline-paths.json" ;;
  "api -X POST repos/o/r/issues/7/labels "*|"api -X POST repos/o/r/issues/7/comments "*) echo '{}' ;;
  "api repos/o/r/pulls?state=open&base=develop&per_page=100 --paginate") cat "$FX/open.json" ;;
  "pr merge 7 --repo o/r --merge --match-head-commit $H") [ -z "${FAKE_MERGE_FAIL:-}" ] || { echo "405 not mergeable" >&2; exit 1; } ;;
  "api -X PUT repos/o/r/pulls/"*"/update-branch") echo '{}' ;;
  *) echo "unexpected: $*" >&2; exit 2 ;;
esac
GH
chmod +x "$tmp/bin/gh"

printf '{"content":"%s"}' "$(printf '.github/*\nscripts/review-*\nscripts/findings.py\n.claude/*\n' | base64 -w0)" >"$tmp/fx/pipeline-paths.json"
printf '{"content":"%s"}' "$(printf 'src/api/internal/identity/*\ndb/queries/*\n' | base64 -w0)" >"$tmp/fx/deep-paths.json"
printf '{"content":"%s"}' "$(printf 'scripts/*\n.github/*\n' | base64 -w0)" >"$tmp/fx/contents.json"

# pull [state base headrepo draft labels-json author]
pull() {
  python3 -I -c 'import json,sys; a=sys.argv; json.dump({"number":7,"state":a[2],"draft":a[5]=="true","base":{"ref":a[3],"sha":a[9]},"head":{"sha":a[7],"repo":{"full_name":a[4]}},"labels":json.loads(a[6]),"user":{"login":a[8]}}, open(a[1],"w"))' \
    "$tmp/fx/pull.json" "${1:-open}" "${2:-develop}" "${3:-o/r}" "${4:-false}" "${5:-[]}" "$H" "${6:-dev}" "$BS"
}
files() { python3 -I -c 'import json,sys; json.dump([{"filename": f} for f in sys.argv[2:]], open(sys.argv[1], "w"))' "$tmp/fx/files.json" "$@"; }
# checks NAME:STATUS:CONCLUSION ...
checks() { python3 -I -c 'import json,sys; json.dump({"total_count":len(sys.argv)-2,"check_runs":[dict(zip(("name","status","conclusion"),c.split(":")), id=i) for i,c in enumerate(sys.argv[2:])]}, open(sys.argv[1],"w"))' "$tmp/fx/checks.json" "$@"; }
# review_comment TRIAGE LEDGER-JSON [DEEP-LINE]
review_comment() {
  python3 -I -c 'import json,sys; body="%s\n<!-- 226os-ai-review -->\nTool: PASS (2 files)\n%s\n<details><summary>Review ledger</summary>\n\n```json\n%s\n```\n\n</details>" % (sys.argv[2], sys.argv[4], sys.argv[3]); json.dump([{"id":1,"user":{"login":"github-actions[bot]"},"body":body,"updated_at":"2026-11-14T10:00:00Z"}], open(sys.argv[1],"w"))' "$tmp/fx/comments.json" "$1" "$2" "${3:-}"
}
GOOD_CHECKS=("Validate branch promotion:completed:success" "quality:completed:success" "Review gate:completed:success" "Code review (advisory):completed:failure")
NOTE='{"id":"n1","kind":"note","evidence":"INFERRED","sev":"LOW","file":"a.go","line":1,"scenario":"s","status":"open"}'

reset() {
  pull
  files src/web/src/main.tsx
  checks "${GOOD_CHECKS[@]}"
  review_comment "Triage: NONE | blast radius: this PR only" "{\"sha\":\"$H\",\"findings\":[$NOTE]}"
  printf '[]' >"$tmp/fx/reviews.json"
  printf '[{"sha":"%s","author":{"login":"Y4NN777"}}]' "$H" >"$tmp/fx/commits.json"
  printf '[{"number":7,"state":"open","base":{"ref":"develop"}}]' >"$tmp/fx/assoc.json"
  printf '[{"number":7,"head":{"sha":"%s","repo":{"full_name":"o/r"}}},{"number":8,"head":{"sha":"%s","repo":{"full_name":"o/r"}}},{"number":9,"head":{"sha":"%s","repo":{"full_name":"fork/r"}}}]' "$H" "$OLD" "$OLD" >"$tmp/fx/open.json"
}

# run [MODE]: sets rc, out, calls, writes (the calls that write).
run() {
  rc=0
  : >"$tmp/fx/calls.log"
  out="$(PATH="$tmp/bin:$PATH" FX="$tmp/fx" REPO=o/r HEAD_SHA="$H" PR_NUMBER="${PRN:-}" AUTO_MERGE="${1-on}" BOT_TOKEN="${TOKEN-bot-secret}" GH_TOKEN=read-token OWNER=Y4NN777 python3 -I "$tool" 2>&1)" || rc=$?
  calls="$(cat "$tmp/fx/calls.log")"
  writes="$(grep -E 'pr merge|-X PUT|-X POST|-X PATCH|-X DELETE' <<<"$calls" || true)"
}

refused() { # REASON-REGEX DESCRIPTION
  run on
  if [ "$rc" -eq 0 ] && grep -q 'REFUSE #7' <<<"$out" && grep -Eq "$1" <<<"$out" && [ -z "$writes" ]; then ok "refused: $2"; else ko "refused: $2 (rc=$rc out=$out writes=$writes)"; fi
}

# --- the happy path ---
reset; run on
[ "$rc" -eq 0 ] && grep -q "^TOKEN=bot-secret pr merge 7 --repo o/r --merge --match-head-commit $H$" <<<"$calls" && ok "every condition holds: a merge commit of exactly the evaluated head, with the bot token" || ko "merge (rc=$rc out=$out calls=$calls)"
! grep -q -- '--squash\|--rebase' <<<"$calls" && ok "never squash or rebase" || ko "merge method ($calls)"
grep -q '^TOKEN=bot-secret api -X PUT repos/o/r/pulls/8/update-branch' <<<"$calls" && ! grep -q 'pulls/7/update-branch\|pulls/9/update-branch' <<<"$calls" && ok "update-branch is called for the other open pull requests into develop (not the merged one, not a fork)" || ko "update-branch ($calls)"
! grep -v 'pr merge\|update-branch' <<<"$calls" | grep -q 'bot-secret' && ok "reads use the workflow token, only writes use the bot token" || ko "token split ($calls)"
! grep -q 'bot-secret' <<<"$out" && ok "the token never appears in the output" || ko "token leak ($out)"
grep -q 'pr merge 7' <<<"$calls" && grep -q 'kind[^,]*note' "$tmp/fx/comments.json" && ok "an open note alone does not block (the happy-path ledger has one)" || ko "note ($out)"

# --- refusals ---
reset; pull closed; refused 'not open' "the pull request is not open"
reset; pull open main; refused 'base is main' "the base is not develop (never main)"
reset; pull open develop fork/r; refused 'head repository' "a fork"
reset; pull open develop o/r true; refused 'draft' "a draft"
for l in needs-owner do-not-merge owner-override-pending; do reset; pull open develop o/r false "[{\"name\":\"$l\"}]"; refused "label $l" "the label $l"; done
reset; files .github/workflows/x.yml; pull open develop o/r false '[]' Y4NN777; run on
grep -q 'REFUSE #7.*workflow' <<<"$out" && ! grep -q 'pr merge' <<<"$calls" && ok "refused: a change to .github/workflows (the bot never merges one)" || ko "workflows ($out)"
reset; checks "Validate branch promotion:completed:success" "quality:completed:failure" "Review gate:completed:success"; refused 'quality' "a required check failed"
reset; checks "Validate branch promotion:completed:success" "quality:in_progress:" "Review gate:completed:success"; refused 'quality' "a required check is still running"
reset; checks "Validate branch promotion:completed:success" "Review gate:completed:success"; refused 'quality' "a required check is missing"
reset; checks "Validate branch promotion:completed:success" "quality:completed:skipped" "Review gate:completed:success"; refused 'quality' "a skipped required check is not SUCCESS"
reset; printf '[]' >"$tmp/fx/comments.json"; refused 'no review comment' "no review comment"
reset; review_comment "Triage: NONE | blast radius: x" "{\"sha\":\"$OLD\",\"findings\":[]}"; refused 'ledger sha' "the ledger sha is not the head (the review did not cover the latest commit)"
reset; review_comment "Triage: NOT RUN | the review returned no structured output" "{\"sha\":\"$H\",\"findings\":[]}"; refused 'NOT RUN' "the review is NOT RUN"
reset; review_comment "Triage: CRITICAL | blast radius: x" "{\"sha\":\"$H\",\"findings\":[]}"; refused 'CRITICAL' "a CRITICAL verdict"
reset; review_comment "Triage: MEDIUM | blast radius: x" "{\"sha\":\"$H\",\"findings\":[{\"id\":\"m1\",\"kind\":\"finding\",\"sev\":\"MEDIUM\",\"status\":\"open\"}]}"; refused 'open finding' "an open finding (any level)"
reset; review_comment "Triage: NONE | blast radius: x" "{\"sha\":\"$H\",\"findings\":[{\"id\":\"m1\",\"kind\":\"finding\",\"sev\":\"MEDIUM\",\"status\":\"fixed\"},$NOTE]}"
run on
grep -q 'pr merge 7' <<<"$calls" && ok "a fixed finding and an open note do not block" || ko "fixed and note ($out)"
reset; files src/api/internal/identity/service/x.go; refused 'deep review' "a P0 path without a deep review"
reset; files src/api/internal/identity/service/x.go; review_comment "Triage: NONE | blast radius: x" "{\"sha\":\"$H\",\"findings\":[],\"deep\":{\"status\":\"PASS\",\"sha\":\"$OLD\"}}"; refused 'deep review' "a deep review of an older commit"
reset; files src/api/internal/identity/service/x.go; review_comment "Triage: NONE | blast radius: x" "{\"sha\":\"$H\",\"findings\":[],\"deep\":{\"status\":\"NOT RUN\",\"sha\":\"$H\"}}"; refused 'deep review' "a deep review NOT RUN"
reset; files src/api/internal/identity/service/x.go; review_comment "Triage: NONE | blast radius: x" "{\"sha\":\"$H\",\"findings\":[],\"deep\":{\"status\":\"PASS\",\"sha\":\"$H\"}}"
run on
grep -q 'pr merge 7' <<<"$calls" && ok "a P0 path with a deep review PASS of the head merges" || ko "deep pass ($out)"
reset; review_comment "Triage: NONE | blast radius: x" "{\"sha\":\"$H\",\"findings\":[]}" "Deep: PENDING (a P0 path is touched: the deep review runs next)"; refused 'deep review is not done' "Deep: PENDING on line 4, even when the lists match nothing"
reset; review_comment "Triage: NONE | blast radius: x" "{\"sha\":\"$H\",\"findings\":[]}" "Deep: NOT RUN (the deep review returned no structured output)"; refused 'deep review is not done' "Deep: NOT RUN on line 4"
reset; files db/queries/solve.sql; refused 'P0 path' "the deep-review list is read from the base branch (db/queries)"
reset; files scripts/x.sh; review_comment "Triage: NONE | blast radius: x" "{\"sha\":\"$H\",\"findings\":[],\"deep\":{\"status\":\"PASS\",\"sha\":\"$H\"}}"; refused 'owner' "a sensitive path without the owner's approval"
reset; files scripts/x.sh; review_comment "Triage: NONE | blast radius: x" "{\"sha\":\"$H\",\"findings\":[],\"deep\":{\"status\":\"PASS\",\"sha\":\"$H\"}}"; printf '[{"user":{"login":"Y4NN777"},"state":"APPROVED","commit_id":"%s","submitted_at":"2026-11-14T10:00:00Z"}]' "$H" >"$tmp/fx/reviews.json"
run on
grep -q 'pr merge 7' <<<"$calls" && ok "a sensitive path approved by the owner merges" || ko "sensitive approved ($out)"
reset; FAKE_GH_FAIL=1 run on
[ "$rc" -eq 0 ] && grep -q 'NOT RUN' <<<"$out" && ! grep -q 'pr merge' <<<"$calls" && ok "a GitHub read error is NOT RUN: no merge, exit 0" || ko "gh fail (rc=$rc $out)"
reset; FAKE_MERGE_FAIL=1 run on
[ "$rc" -eq 0 ] && grep -q 'GitHub refused the merge' <<<"$out" && ! grep -q 'update-branch' <<<"$calls" && ok "a merge GitHub refuses (conflict, unresolved thread) is reported, exit 0, no update-branch" || ko "merge fail (rc=$rc $out)"

# --- review of PR 168: stale owner approval, owner identity, level without a finding ---
SENS_OK="{\"sha\":\"$H\",\"findings\":[],\"deep\":{\"status\":\"PASS\",\"sha\":\"$H\"}}"
reset; files scripts/x.sh; review_comment "Triage: NONE | blast radius: x" "$SENS_OK"; printf '[{"user":{"login":"Y4NN777"},"state":"APPROVED","commit_id":"%s","submitted_at":"2026-11-14T10:00:00Z"}]' "$OLD" >"$tmp/fx/reviews.json"
refused 'owner' "an owner approval of an older commit does not cover the head"
reset; files scripts/x.sh; review_comment "Triage: NONE | blast radius: x" "$SENS_OK"; pull open develop o/r false '[]' Y4NN777
run on
grep -q 'pr merge 7' <<<"$calls" && ok "an owner-authored sensitive change whose commits are all the owner's merges" || ko "owner authored ($out)"
reset; files scripts/x.sh; review_comment "Triage: NONE | blast radius: x" "$SENS_OK"; pull open develop o/r false '[]' Y4NN777
printf '[{"sha":"%s","author":{"login":"Y4NN777"}},{"sha":"%s","author":{"login":"someone"}}]' "$OLD" "$H" >"$tmp/fx/commits.json"
refused 'owner' "an owner-authored PR with a commit by someone else is not owner-authored"
reset; review_comment "Triage: HIGH | blast radius: x" "{\"sha\":\"$H\",\"findings\":[]}"; refused 'no open finding explains' "a HIGH verdict with no open finding"
reset; review_comment "Triage: MEDIUM | blast radius: x" "{\"sha\":\"$H\",\"findings\":[$NOTE]}"; refused 'no open finding explains' "a MEDIUM verdict with only a note"
reset; run on
grep -q "contents/.github/pipeline-paths.txt?ref=$BS" <<<"$calls" && grep -q "contents/.github/deep-review-paths.txt?ref=$BS" <<<"$calls" && ! grep -q 'ref=develop' <<<"$calls" && ok "the path lists are read at the pull request's base sha" || ko "lists ref ($calls)"

# --- review pipeline paths: never auto-merged; with AUTO_MERGE=on labelled needs-owner with a one-line reason ---
pipeline_ok() { review_comment "Triage: NONE | blast radius: x" "{\"sha\":\"$H\",\"findings\":[],\"deep\":{\"status\":\"PASS\",\"sha\":\"$H\"}}"; pull open develop o/r false '[]' Y4NN777; }
reset; files scripts/review-gate.py; pipeline_ok; run on
[ "$rc" -eq 0 ] && grep -q 'REFUSE #7.*review pipeline (scripts/review-gate.py)' <<<"$out" && ! grep -q 'pr merge' <<<"$calls" && grep -q '^TOKEN=bot-secret api -X POST repos/o/r/issues/7/labels -f labels\[\]=needs-owner$' <<<"$calls" && grep -q 'issues/7/comments -f body=Auto-merge skipped: it touches the review pipeline' <<<"$calls" && ok "a pipeline path is skipped and labelled needs-owner with a one-line reason" || ko "pipeline on (rc=$rc $out $calls)"
reset; files .claude/commands/deep-review.md; pipeline_ok; run report
[ -z "$writes" ] && grep -q 'WOULD LABEL #7 needs-owner' <<<"$out" && ok "report mode: a pipeline path is not labelled, only reported" || ko "pipeline report ($out $writes)"
reset; files src/web/src/features/solve/Question.tsx src/api/internal/build/build.go; run on
grep -q 'pr merge 7' <<<"$calls" && ! grep -q 'labels' <<<"$calls" && ok "an app-only pull request merges, no label" || ko "app only ($out)"

# --- modes and token ---
reset; run off
[ "$rc" -eq 0 ] && [ -z "$calls" ] && grep -q 'off' <<<"$out" && ok "mode off: nothing is read nor written" || ko "off (rc=$rc $out $calls)"
reset; run report
[ "$rc" -eq 0 ] && [ -z "$writes" ] && grep -q 'WOULD MERGE #7' <<<"$out" && grep -q 'WOULD UPDATE #8' <<<"$out" && ok "mode report: prints what it would do, writes nothing" || ko "report (rc=$rc $out $writes)"
reset; run ""
[ -z "$writes" ] && grep -q 'WOULD MERGE' <<<"$out" && ok "the default mode is report" || ko "default ($out $writes)"
reset; TOKEN='' run on
[ "$rc" -eq 0 ] && [ "$out" = "auto-merge: NOT RUN (REVIEW_AGENT_TOKEN missing)" ] && [ -z "$calls" ] && ok "a missing token: NOT RUN, exit 0, no call at all" || ko "no token (rc=$rc $out $calls)"
reset; PRN=7 run report
grep -q 'WOULD MERGE #7' <<<"$out" && ! grep -q "commits/$H/pulls" <<<"$calls" && ok "PR_NUMBER (manual run) evaluates that pull request" || ko "pr number ($out)"
reset; PRN='7;x' run report
grep -q 'NOT RUN' <<<"$out" && [ -z "$calls" ] && ok "a malformed PR_NUMBER is NOT RUN" || ko "bad pr number ($out)"

finish
