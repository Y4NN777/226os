#!/usr/bin/env bash
# Tests for scripts/review-ping.py with a fake `gh`: a CRITICAL or HIGH finding that is still open in the review's
# ledger labels the pull request `needs-owner` and asks the owner for a review (unless the owner wrote it); anything
# else only clears the label; a GitHub error never fails the job.
# Usage: scripts/tests/review-ping_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
tool="$root/scripts/review-ping.py"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin"

# Fake gh: logs every call; answers the comments endpoint from a fixture; FAKE_GH_FAIL makes every call fail.
cat >"$tmp/bin/gh" <<'EOF'
#!/usr/bin/env bash
echo "$*" >>"$LOG"
[ -z "${FAKE_GH_FAIL:-}" ] || { echo "gh: boom" >&2; exit 1; }
case "$*" in
  *"/comments"*"--paginate"*) cat "$FX" ;;
  *) echo '{}' ;;
esac
EOF
chmod +x "$tmp/bin/gh"

# ledger SEV STATUS [FILE]: a bot comment carrying a fenced JSON ledger with one finding.
ledger() {
  python3 -I - "$1" "$2" <<'EOF'
import json, sys
sev, status = sys.argv[1], sys.argv[2]
body = "Triage: HIGH | blast radius: x\n\n<details><summary>Review ledger</summary>\n\n```json\n" + json.dumps(
    {"sha": "abcdef1234567", "findings": [{"id": "f1", "sev": sev, "file": "a.go", "line": 1, "scenario": "s", "status": status}]}) + "\n```\n\n</details>"
print(json.dumps([{"user": {"login": "github-actions[bot]"}, "created_at": "2026-11-13T10:00:00Z", "body": body}]))
EOF
}

# run AUTHOR: runs the script; sets rc and out; the calls are in $LOG.
run() {
  : >"$tmp/log"
  rc=0
  out="$(PATH="$tmp/bin:$PATH" LOG="$tmp/log" FX="$tmp/fx.json" REPO=o/r PR_NUMBER=7 PR_AUTHOR="${1:-dev}" OWNER=Y4NN777 python3 -I "$tool" 2>&1)" || rc=$?
}

ledger HIGH open >"$tmp/fx.json"
run dev
[ "$rc" -eq 0 ] && grep -q 'issues/7/labels' "$tmp/log" && grep -q 'needs-owner' "$tmp/log" && ok "an open HIGH labels the pull request needs-owner" || ko "high label (rc=$rc log=$(cat "$tmp/log") out=$out)"
grep -q 'pulls/7/requested_reviewers' "$tmp/log" && grep -q 'Y4NN777' "$tmp/log" && ok "an open HIGH asks the owner for a review" || ko "high review request ($(cat "$tmp/log"))"
! grep -q 'gh pr edit\|^pr edit' "$tmp/log" && ok "gh pr edit is never used (it fails on the Projects classic API)" || ko "pr edit used"

ledger CRITICAL open >"$tmp/fx.json"
run dev
grep -q 'needs-owner' "$tmp/log" && grep -q 'requested_reviewers' "$tmp/log" && ok "an open CRITICAL does the same" || ko "critical ($(cat "$tmp/log"))"

ledger HIGH open >"$tmp/fx.json"
run Y4NN777
grep -q 'needs-owner' "$tmp/log" && ! grep -q 'requested_reviewers' "$tmp/log" && ok "the owner's own pull request is labelled but no review is requested (GitHub refuses it)" || ko "owner author ($(cat "$tmp/log"))"

ledger MEDIUM open >"$tmp/fx.json"
run dev
! grep -q 'POST' "$tmp/log" && ok "only a MEDIUM open: nothing is posted" || ko "medium ($(cat "$tmp/log"))"
grep -q 'DELETE' "$tmp/log" && grep -q 'needs-owner' "$tmp/log" && ok "the needs-owner label is cleared when nothing serious is open" || ko "clear ($(cat "$tmp/log"))"

ledger CRITICAL fixed >"$tmp/fx.json"
run dev
! grep -q 'POST' "$tmp/log" && ok "a fixed CRITICAL does not ping the owner" || ko "fixed ($(cat "$tmp/log"))"

printf '[]' >"$tmp/fx.json"
run dev
[ "$rc" -eq 0 ] && ! grep -q 'POST' "$tmp/log" && ok "no ledger yet: nothing is posted" || ko "no ledger (rc=$rc)"

printf '[{"user":{"login":"dev"},"created_at":"2026-11-13T10:00:00Z","body":"<details><summary>Review ledger</summary>\\n```json\\n{\\"sha\\":\\"abcdef1234567\\",\\"findings\\":[{\\"id\\":\\"f\\",\\"sev\\":\\"HIGH\\",\\"status\\":\\"open\\"}]}\\n```\\n</details>"}]' >"$tmp/fx.json"
run dev
! grep -q 'POST' "$tmp/log" && ok "a ledger written by a human is ignored" || ko "human ledger ($(cat "$tmp/log"))"

ledger HIGH open >"$tmp/fx.json"
FAKE_GH_FAIL=1 run dev
[ "$rc" -eq 0 ] && ok "a GitHub error fails open (exit 0)" || ko "gh failure (rc=$rc out=$out)"

rc=0
out="$(PATH="$tmp/bin:$PATH" LOG="$tmp/log" FX="$tmp/fx.json" python3 -I "$tool" 2>&1)" || rc=$?
[ "$rc" -eq 0 ] && grep -q 'NOT RUN' <<<"$out" && ok "missing REPO and PR_NUMBER is NOT RUN, not a failure" || ko "no env (rc=$rc out=$out)"

# A note is a remark, not a defect: even with a HIGH sev written by the model it does not ping the owner.
python3 -I - >"$tmp/fx.json" <<'EOF'
import json
body = "Triage: NONE\n\n<details><summary>Review ledger</summary>\n\n```json\n" + json.dumps(
    {"sha": "abcdef1234567", "findings": [{"id": "n1", "kind": "note", "evidence": "INFERRED", "sev": "HIGH", "file": "a.go", "line": 1, "scenario": "guess", "status": "open"}]}) + "\n```\n\n</details>"
print(json.dumps([{"user": {"login": "github-actions[bot]"}, "created_at": "2026-11-13T10:00:00Z", "body": body}]))
EOF
run dev
! grep -q 'requested_reviewers' "$tmp/log" && ! grep -q 'POST.*issues/7/labels' "$tmp/log" && ok "an open note never labels the pull request needs-owner nor asks the owner for a review" || ko "note ping ($(cat "$tmp/log"))"


finish
