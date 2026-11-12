#!/usr/bin/env bash
# RED/GREEN evidence for scripts/check-develop-contains-main.sh: a pull request into develop is refused
# while develop lacks commits that main has (the promotion merge commits), so the back-merge happens
# right after each promotion instead of piling up as "N commits behind main".
# A fake `gh` answers the two compare calls; nothing touches the network.
# Usage: scripts/tests/develop-contains-main_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
guard="$root/scripts/check-develop-contains-main.sh"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin"
cat >"$tmp/bin/gh" <<'EOF'
#!/usr/bin/env bash
# Fake gh: FAKE_AHEAD answers compare/develop...main, FAKE_BEHIND answers compare/main...<sha>.
[ -z "${FAKE_FAIL:-}" ] || { echo "gh: HTTP 500" >&2; exit 1; }
case "$*" in
  *compare/develop...main*) printf '%s\n' "${FAKE_AHEAD:-0}" ;;
  *compare/main...*) printf '%s\n' "${FAKE_BEHIND:-0}" ;;
  *) echo "unexpected gh call: $*" >&2; exit 2 ;;
esac
EOF
chmod +x "$tmp/bin/gh"

# run NAME EXPECT(pass|fail) NEEDLE ENV...: runs the guard with the fake gh and the given environment.
run() {
  local name="$1" expect="$2" needle="$3"
  shift 3
  local rc=0 out
  out="$(env -i PATH="$tmp/bin:$PATH" GITHUB_REPOSITORY=Y4NN777/226os GH_TOKEN=x \
    BASE_REF=develop HEAD_REF=feature/platform/some-work HEAD_SHA=abc123 "$@" "$guard" 2>&1)" || rc=$?
  if [ "$expect" = pass ] && [ "$rc" -ne 0 ]; then ko "$name: should pass (rc=$rc, out=$out)"; return; fi
  if [ "$expect" = fail ] && [ "$rc" -eq 0 ]; then ko "$name: should fail but passed (out=$out)"; return; fi
  if [ -n "$needle" ] && ! grep -qF -- "$needle" <<<"$out"; then ko "$name: output lacks '$needle' (out=$out)"; return; fi
  ok "$name"
}

[ -x "$guard" ] || { ko "scripts/check-develop-contains-main.sh is missing or not executable"; finish; exit 1; }

run "develop contains main: a work PR passes" pass "" FAKE_AHEAD=0
run "develop is behind main: a work PR is refused with the back-merge instruction" fail "back-merge" FAKE_AHEAD=3
run "the message says how far behind develop is" fail "3 commit" FAKE_AHEAD=3
run "a PR that does not target develop is not concerned" pass "" BASE_REF=main FAKE_AHEAD=9
run "the back-merge branch passes while develop is still behind, if its head contains main" pass "" \
  HEAD_REF=chore/platform/sync-main-into-develop FAKE_AHEAD=8 FAKE_BEHIND=0
run "the back-merge branch is refused when its head does not contain main" fail "does not contain main" \
  HEAD_REF=chore/platform/sync-main-into-develop FAKE_AHEAD=8 FAKE_BEHIND=2
run "a GitHub API failure fails closed" fail "could not" FAKE_FAIL=1
run "a non-numeric API answer fails closed" fail "could not" FAKE_AHEAD=oops
run "a missing variable fails closed" fail "GITHUB_REPOSITORY" GITHUB_REPOSITORY=

finish
