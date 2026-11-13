#!/usr/bin/env bash
# Tests for scripts/review-gate.py with a fake `gh`: the gate fails only on an OBSERVED CRITICAL from the Actions bot
# (unless the owner overrides it) or on a sensitive path without the owner's approval, only in `enforce` mode, and it
# fails OPEN when the review did not run or GitHub cannot be read.
# Usage: scripts/tests/review-gate_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
gate="$root/scripts/review-gate.py"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin" "$tmp/fx"

# Fake gh: `gh api <endpoint> --paginate` answers from fixture files by endpoint; FAKE_GH_FAIL makes it fail.
cat >"$tmp/bin/gh" <<'EOF'
#!/usr/bin/env bash
[ -z "${FAKE_GH_FAIL:-}" ] || { echo "gh: boom" >&2; exit 1; }
ep="$2"
case "$ep" in
  *"/contents/.github/sensitive-paths.txt"*) [ -z "${FAKE_NO_LIST:-}" ] || { echo "404" >&2; exit 1; }; cat "$FX/contents.json" ;;
  *"/contents/.github/stray-paths.txt"*) [ -z "${FAKE_NO_STRAY:-}" ] || { echo "404" >&2; exit 1; }; cat "$FX/stray.json" ;;
  *"/pulls/7") cat "$FX/pull.json" ;;
  *"/issues/12") cat "$FX/issue12.json" ;;
  *"/comments"*) cat "$FX/comments.json" ;;
  *"/events"*) cat "$FX/events.json" ;;
  *"/reviews"*) cat "$FX/reviews.json" ;;
  *"/files"*) cat "$FX/files.json" ;;
  *) echo "unexpected endpoint $ep" >&2; exit 2 ;;
esac
EOF
chmod +x "$tmp/bin/gh"

list_b64="$(printf 'db/migrations/*\n.github/*\nscripts/*\nmk/*\nsecrets/*\napi/openapi.yaml\n# a comment\n\n' | base64 -w0)"
printf '{"content":"%s","encoding":"base64"}' "$list_b64" >"$tmp/fx/contents.json"

stray_b64="$(printf '# one pattern per line, prefixed by when it applies\nalways: .devin/*\nalways: .cursor/*\nalways: .idea/*\nalways: .vscode/*\nfeature: .gitignore\n' | base64 -w0)"
printf '{"content":"%s","encoding":"base64"}' "$stray_b64" >"$tmp/fx/stray.json"

reset() {
  printf '{"body":"","labels":[],"head":{"ref":"feature/identity/profiles"},"user":{"login":"dev"}}' >"$tmp/fx/pull.json"
  printf '{"number":12,"labels":[{"name":"finding"}]}' >"$tmp/fx/issue12.json"
  printf '[]' >"$tmp/fx/comments.json"
  printf '[]' >"$tmp/fx/events.json"
  printf '[]' >"$tmp/fx/reviews.json"
  printf '[{"filename":"src/api/internal/app/app.go"}]' >"$tmp/fx/files.json"
}

# comment LOGIN CREATED BODY: appends one issue comment.
comment() {
  python3 -I - "$tmp/fx/comments.json" "$1" "$2" "$3" <<'EOF'
import json, sys
p, login, created, body = sys.argv[1:5]
d = json.load(open(p))
d.append({"user": {"login": login}, "created_at": created, "body": body})
json.dump(d, open(p, "w"))
EOF
}

# event ACTION LOGIN CREATED LABEL: appends one issue event.
event() {
  python3 -I - "$tmp/fx/events.json" "$1" "$2" "$3" "$4" <<'EOF'
import json, sys
p, action, login, created, label = sys.argv[1:6]
d = json.load(open(p))
d.append({"event": action, "actor": {"login": login}, "created_at": created, "label": {"name": label}})
json.dump(d, open(p, "w"))
EOF
}

# review LOGIN STATE SUBMITTED: appends one pull request review.
review() {
  python3 -I - "$tmp/fx/reviews.json" "$1" "$2" "$3" <<'EOF'
import json, sys
p, login, state, when = sys.argv[1:5]
d = json.load(open(p))
d.append({"user": {"login": login}, "state": state, "submitted_at": when})
json.dump(d, open(p, "w"))
EOF
}

files() { python3 -I -c 'import json,sys; json.dump([{"filename": f} for f in sys.argv[2:]], open(sys.argv[1], "w"))' "$tmp/fx/files.json" "$@"; }

# run MODE [author]: runs the gate; sets rc and out.
run() {
  rc=0
  out="$(PATH="$tmp/bin:$PATH" FX="$tmp/fx" GATE_MODE="$1" REPO=o/r PR_NUMBER=7 PR_AUTHOR="${2:-dev}" OWNER=Y4NN777 BASE_REF=develop python3 -I "$gate" 2>&1)" || rc=$?
}

BOT='github-actions[bot]'
crit_ok='Triage: CRITICAL | blast radius: the whole team

| Dimension | Level | Note |
|---|---|---|
| Security | CRITICAL | a key is committed |

OBSERVED: src/api/config.go:12 holds a literal API key.
Fix: remove it and rotate the key.'
crit_inferred='Triage: CRITICAL | blast radius: production and users

INFERRED: this could leak data if the handler is reached.'

reset
run enforce
[ "$rc" -eq 0 ] && grep -q 'NOT RUN' <<<"$out" && ok "no review comment: the gate passes and says NOT RUN" || ko "no comment (rc=$rc out=$out)"

for lvl in NONE HIGH MEDIUM; do
  reset; comment "$BOT" 2026-11-13T01:00:00Z "Triage: $lvl | blast radius: this PR only"
  run enforce
  [ "$rc" -eq 0 ] && grep -q "verdict $lvl" <<<"$out" && ok "$lvl never blocks, even in enforce" || ko "$lvl (rc=$rc out=$out)"
done

reset; comment "$BOT" 2026-11-13T01:00:00Z "$crit_ok"
run enforce
[ "$rc" -eq 1 ] && grep -q 'FAIL' <<<"$out" && ok "an OBSERVED CRITICAL fails in enforce mode" || ko "critical enforce (rc=$rc out=$out)"
run report
[ "$rc" -eq 0 ] && grep -q 'WOULD FAIL' <<<"$out" && ok "report mode shows WOULD FAIL and does not block" || ko "critical report (rc=$rc out=$out)"
run off
[ "$rc" -eq 0 ] && ! grep -q 'WOULD FAIL\|FAIL' <<<"$out" && grep -q 'off' <<<"$out" && ok "off mode does nothing" || ko "critical off (rc=$rc out=$out)"
run ""
[ "$rc" -eq 0 ] && grep -q 'WOULD FAIL' <<<"$out" && ok "the default mode is report" || ko "default mode (rc=$rc out=$out)"

reset; comment "$BOT" 2026-11-13T01:00:00Z "$crit_inferred"
run enforce
[ "$rc" -eq 0 ] && grep -q 'downgraded' <<<"$out" && ok "a CRITICAL without OBSERVED evidence is downgraded and does not block" || ko "inferred critical (rc=$rc out=$out)"

reset; comment dev-user 2026-11-13T01:00:00Z "$crit_ok"
run enforce
[ "$rc" -eq 0 ] && ok "a human comment with the same text is ignored" || ko "human comment (rc=$rc out=$out)"

reset; comment "$BOT" 2026-11-13T01:00:00Z "$crit_ok"; comment "$BOT" 2026-11-13T02:00:00Z "Triage: NONE | blast radius: this PR only"
run enforce
[ "$rc" -eq 0 ] && ok "the latest bot verdict wins" || ko "latest wins (rc=$rc out=$out)"

reset; comment "$BOT" 2026-11-13T01:00:00Z "not a verdict at all"
run enforce
[ "$rc" -eq 0 ] && grep -q 'NOT RUN' <<<"$out" && ok "a bot comment without the verdict line is not a verdict" || ko "no verdict line (rc=$rc out=$out)"

reset; comment "$BOT" 2026-11-13T01:00:00Z "$crit_ok"; event labeled Y4NN777 2026-11-13T02:00:00Z owner-override
run enforce
[ "$rc" -eq 0 ] && grep -q 'owner override' <<<"$out" && ok "the owner's label overrides a CRITICAL" || ko "owner override (rc=$rc out=$out)"

reset; comment "$BOT" 2026-11-13T01:00:00Z "$crit_ok"; event labeled dev-user 2026-11-13T02:00:00Z owner-override
run enforce
[ "$rc" -eq 1 ] && ok "the same label added by someone else does not override" || ko "foreign label (rc=$rc out=$out)"

reset; comment "$BOT" 2026-11-13T01:00:00Z "$crit_ok"; event labeled Y4NN777 2026-11-13T02:00:00Z owner-override; event unlabeled Y4NN777 2026-11-13T03:00:00Z owner-override
run enforce
[ "$rc" -eq 1 ] && ok "a removed override label no longer overrides" || ko "removed label (rc=$rc out=$out)"

reset; comment "$BOT" 2026-11-13T01:00:00Z "$crit_ok"; event labeled Y4NN777 2026-11-13T02:00:00Z other-label
run enforce
[ "$rc" -eq 1 ] && ok "only the owner-override label counts" || ko "other label (rc=$rc out=$out)"

reset; comment "$BOT" 2026-11-13T01:00:00Z "$crit_ok"
FAKE_GH_FAIL=1 run enforce
[ "$rc" -eq 0 ] && grep -q 'NOT RUN' <<<"$out" && ok "a GitHub read error fails open (NOT RUN), never blocks" || ko "gh failure (rc=$rc out=$out)"

# --- sensitive paths: the owner must approve ---
reset; files scripts/check-docs.py src/api/internal/app/app.go
run enforce dev
[ "$rc" -eq 1 ] && grep -q 'scripts/check-docs.py' <<<"$out" && ok "a sensitive path without the owner's approval fails in enforce" || ko "sensitive no approval (rc=$rc out=$out)"
run report dev
[ "$rc" -eq 0 ] && grep -q 'WOULD FAIL' <<<"$out" && ok "the sensitive-path check also stays report-only in report mode" || ko "sensitive report (rc=$rc out=$out)"

reset; files db/migrations/00013_x.sql; review Y4NN777 APPROVED 2026-11-13T05:00:00Z
run enforce dev
[ "$rc" -eq 0 ] && grep -q 'approved by the owner' <<<"$out" && ok "the owner's approval satisfies the sensitive-path check" || ko "owner approved (rc=$rc out=$out)"

reset; files db/migrations/00013_x.sql; review Y4NN777 APPROVED 2026-11-13T05:00:00Z; review Y4NN777 CHANGES_REQUESTED 2026-11-13T06:00:00Z
run enforce dev
[ "$rc" -eq 1 ] && ok "the owner's latest review state counts (changes requested after approval fails)" || ko "latest review (rc=$rc out=$out)"

reset; files db/migrations/00013_x.sql; review someone-else APPROVED 2026-11-13T05:00:00Z
run enforce dev
[ "$rc" -eq 1 ] && ok "an approval by someone else does not satisfy it" || ko "other approver (rc=$rc out=$out)"

reset; files scripts/x.sh
run enforce Y4NN777
[ "$rc" -eq 0 ] && grep -q 'owner-authored' <<<"$out" && ok "an owner-authored PR passes the sensitive-path check" || ko "owner author (rc=$rc out=$out)"

reset; files docs/team/KICKOFF.md src/web/src/main.tsx
run enforce dev
[ "$rc" -eq 0 ] && grep -q 'no sensitive path' <<<"$out" && ok "a PR without sensitive paths passes" || ko "no sensitive (rc=$rc out=$out)"

reset; files scripts/x.sh
FAKE_NO_LIST=1 run enforce dev
[ "$rc" -eq 0 ] && grep -q 'no sensitive-paths list' <<<"$out" && ok "a missing list means no sensitive paths (never blocks)" || ko "no list (rc=$rc out=$out)"

reset; files scripts/x.sh; comment "$BOT" 2026-11-13T01:00:00Z "$crit_ok"
run enforce dev
[ "$rc" -eq 1 ] && [ "$(grep -c 'FAIL' <<<"$out")" -ge 2 ] && ok "both failures are reported together" || ko "both (rc=$rc out=$out)"

# --- stray paths and finding-to-test: report only until GATE_NEW=enforce (the promotion switch) ---
set_pull() { # BODY LABELS_JSON HEAD_REF
  python3 -I -c 'import json,sys; json.dump({"body": sys.argv[2], "labels": json.loads(sys.argv[3]), "head": {"ref": sys.argv[4]}, "user": {"login": "dev"}}, open(sys.argv[1], "w"))' "$tmp/fx/pull.json" "$1" "$2" "$3"
}

reset; files .gitignore src/api/internal/app/app.go
run enforce dev
[ "$rc" -eq 0 ] && grep -q 'stray' <<<"$out" && grep -q '\.gitignore' <<<"$out" && grep -q 'report only' <<<"$out" && ok "a stray .gitignore on a feature branch is reported, and never blocks while the checks are report-only" || ko "stray report (rc=$rc out=$out)"
GATE_NEW=enforce run enforce dev
[ "$rc" -eq 1 ] && grep -q 'stray' <<<"$out" && ok "with GATE_NEW=enforce the stray path fails (the promotion switch)" || ko "stray enforce (rc=$rc out=$out)"

reset; files .gitignore; set_pull "" "[]" "chore/ops/x"
GATE_NEW=enforce run enforce dev
[ "$rc" -eq 0 ] && ok ".gitignore on a chore branch is not stray" || ko "gitignore chore (rc=$rc out=$out)"

reset; files .devin/skills/pr-review/SKILL.md; set_pull "" "[]" "chore/ops/x"
GATE_NEW=enforce run enforce dev
[ "$rc" -eq 1 ] && grep -q '\.devin' <<<"$out" && ok "an editor or agent directory is stray on any branch" || ko "devin (rc=$rc out=$out)"

reset; files .devin/skills/x; set_pull "" '[{"name":"stray-ok"}]' "feature/identity/profiles"
GATE_NEW=enforce run enforce dev
[ "$rc" -eq 0 ] && ok "the label stray-ok exempts a pull request" || ko "stray-ok (rc=$rc out=$out)"

reset; files .devin/skills/x
FAKE_NO_STRAY=1 GATE_NEW=enforce run enforce dev
[ "$rc" -eq 0 ] && grep -q 'stray paths: NOT RUN' <<<"$out" && ok "a missing stray-paths list is NOT RUN, never a failure" || ko "no stray list (rc=$rc out=$out)"

reset; set_pull "Closes #12" "[]" "fix/identity/x"; files src/api/internal/app/app.go
run enforce dev
[ "$rc" -eq 0 ] && grep -q 'finding' <<<"$out" && grep -q 'test' <<<"$out" && grep -q 'report only' <<<"$out" && ok "closing a finding issue without a test change is reported (report only)" || ko "finding no test (rc=$rc out=$out)"
GATE_NEW=enforce run enforce dev
[ "$rc" -eq 1 ] && ok "with GATE_NEW=enforce a finding closed without a test fails" || ko "finding enforce (rc=$rc out=$out)"

reset; set_pull "Fixes #12" "[]" "fix/identity/x"; files src/api/internal/identity/service/profile.go src/api/internal/identity/service/profile_test.go
GATE_NEW=enforce run enforce dev
[ "$rc" -eq 0 ] && ok "a finding closed together with a Go test passes" || ko "finding with test (rc=$rc out=$out)"

reset; set_pull "Resolves #12" "[]" "fix/web/x"; files src/web/src/a.ts scripts/tests/review_test.sh
GATE_NEW=enforce run enforce Y4NN777
[ "$rc" -eq 0 ] && grep -q 'a test file changed' <<<"$out" && ok "a script test under scripts/tests counts as a test (owner-authored: scripts/ is a sensitive path)" || ko "script test (rc=$rc out=$out)"

reset; set_pull "Closes #12" '[{"name":"no-test"}]' "fix/identity/x"; files src/api/internal/app/app.go
GATE_NEW=enforce run enforce dev
[ "$rc" -eq 0 ] && ok "the label no-test exempts a finding closed without a test" || ko "no-test (rc=$rc out=$out)"

reset; set_pull "Closes #12" "[]" "fix/identity/x"; files src/api/internal/app/app.go
printf '{"number":12,"labels":[{"name":"bug"}]}' >"$tmp/fx/issue12.json"
GATE_NEW=enforce run enforce dev
[ "$rc" -eq 0 ] && ok "closing an issue that is not a finding needs no test" || ko "not a finding (rc=$rc out=$out)"

reset; set_pull "Closes #12" "[]" "fix/identity/x"; files src/api/internal/app/app.go
GATE_NEW=off run enforce dev
[ "$rc" -eq 0 ] && ! grep -q 'finding' <<<"$out" && ok "GATE_NEW=off disables the two new checks" || ko "new off (rc=$rc out=$out)"

# The promotion switch is independent: GATE_NEW=enforce blocks even while the main gate is still in report mode
# (review note on #114: promoting GATE_NEW alone used to stay silently non-blocking).
reset; files .devin/skills/x
GATE_NEW=enforce run report dev
[ "$rc" -eq 1 ] && grep -q 'stray' <<<"$out" && ok "GATE_NEW=enforce blocks a stray path even when GATE_MODE=report" || ko "independent switch (rc=$rc out=$out)"
reset; set_pull "Closes #12" "[]" "fix/identity/x"; files src/api/internal/app/app.go
GATE_NEW=enforce run report dev
[ "$rc" -eq 1 ] && ok "GATE_NEW=enforce blocks a finding closed without a test even when GATE_MODE=report" || ko "independent switch finding (rc=$rc out=$out)"
reset; files src/api/internal/app/app.go
GATE_NEW=enforce run report dev
[ "$rc" -eq 0 ] && ok "GATE_NEW=enforce with nothing to report passes in report mode" || ko "independent clean (rc=$rc out=$out)"

# A negation is not a closing keyword (review note on #114).
reset; set_pull "This does not close #12, it only prepares it." "[]" "fix/identity/x"; files src/api/internal/app/app.go
GATE_NEW=enforce run enforce dev
[ "$rc" -eq 0 ] && ok "\"does not close #12\" is not a closing keyword" || ko "negation (rc=$rc out=$out)"
reset; set_pull "Fixes: #12" "[]" "fix/identity/x"; files src/api/internal/app/app.go
GATE_NEW=enforce run enforce dev
[ "$rc" -eq 1 ] && ok "\"Fixes: #12\" (with a colon, as GitHub reads it) is a closing keyword" || ko "colon (rc=$rc out=$out)"

finish
