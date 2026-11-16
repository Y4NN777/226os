#!/usr/bin/env bash
# Tests for scripts/sync-main-into-develop.sh (step 6 of docs/GIT_WORKFLOW.md, automated): a fake `gh` answers from
# files and records argv, so nothing touches GitHub. Usage: scripts/tests/sync-main-into-develop_test.sh
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
sync="$root/scripts/sync-main-into-develop.sh"
wf="$root/.github/workflows/sync-main-into-develop.yml"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir "$tmp/bin"
main_sha="$(printf 'a%.0s' {1..40})"
# Fake gh: compare -> AHEAD; main ref -> main_sha; branch ref -> BRANCH_EXISTS; pr list -> OPEN_PR; FAIL_ON=<word> fails.
cat >"$tmp/bin/gh" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >>"$LOG"
if [ -n "${FAIL_ON:-}" ] && [[ "$*" == *"$FAIL_ON"* ]]; then echo "HTTP 422: refused" >&2; exit 1; fi
case "$*" in
  *"compare/develop...main"*) printf '%s\n' "$AHEAD" ;;
  *"git/ref/heads/main"*) printf '%s\n' "$MAIN_SHA" ;;
  *"git/ref/heads/chore/platform/sync-main-into-develop"*) [ "$BRANCH_EXISTS" = 1 ] && printf '%s\n' "$MAIN_SHA" || { echo "HTTP 404" >&2; exit 1; } ;;
  *"pr list"*) printf '%s\n' "$OPEN_PR" ;;
  *"pr create"*) echo "https://github.com/o/r/pull/900" ;;
  *) : ;;
esac
EOF
chmod +x "$tmp/bin/gh"

# run VAR=value...: the script with only the given variables; sets rc, out and the gh log.
run() {
  : >"$tmp/log"
  rc=0
  out="$(env -i PATH="$tmp/bin:$PATH" LOG="$tmp/log" GITHUB_REPOSITORY=o/r GH_TOKEN=t MAIN_SHA="$main_sha" \
    AHEAD=0 BRANCH_EXISTS=0 OPEN_PR="" "$@" "$sync" 2>&1)" || rc=$?
}

run AHEAD=0
[ "$rc" -eq 0 ] && grep -q 'NO-OP' <<<"$out" && ! grep -qE 'pr create|git/refs' "$tmp/log" \
  && ok "develop already contains main: NO-OP, nothing created" || ko "no-op (rc=$rc $out $(cat "$tmp/log"))"

run AHEAD=1
if [ "$rc" -eq 0 ] && grep -q "repos/o/r/git/refs -f ref=refs/heads/chore/platform/sync-main-into-develop -f sha=$main_sha" "$tmp/log" \
  && grep -q 'pr create --repo o/r --base develop --head chore/platform/sync-main-into-develop' "$tmp/log" \
  && grep -q 'pr merge --repo o/r --auto --merge https://github.com/o/r/pull/900' "$tmp/log"; then
  ok "develop lacks main: the branch is created at main, the PR opened into develop and auto-merge armed (merge commit)"
else ko "create (rc=$rc $out $(cat "$tmp/log"))"; fi
! grep -qE -- '--squash|--rebase' "$tmp/log" && ok "never squash or rebase" || ko "merge method"

run AHEAD=2 BRANCH_EXISTS=1 OPEN_PR=901
if [ "$rc" -eq 0 ] && grep -q "repos/o/r/git/refs/heads/chore/platform/sync-main-into-develop -X PATCH -f sha=$main_sha -F force=true" "$tmp/log" \
  && ! grep -q 'pr create' "$tmp/log" && grep -q 'pr merge --repo o/r --auto --merge 901' "$tmp/log"; then
  ok "an existing branch and open PR: the branch moves to main, the PR is reused and auto-merge armed"
else ko "reuse (rc=$rc $out $(cat "$tmp/log"))"; fi

run AHEAD=1 FAIL_ON="git/refs"
[ "$rc" -ne 0 ] && grep -q 'FAILED' <<<"$out" && ! grep -q 'pr create' "$tmp/log" \
  && ok "a refused branch update (e.g. a workflow file without the Workflows permission) fails visibly" || ko "refused ref (rc=$rc $out)"

run AHEAD=abc
[ "$rc" -ne 0 ] && ok "a non-numeric compare answer fails closed" || ko "bad compare (rc=$rc)"

rc=0; out="$(env -i PATH="$tmp/bin:$PATH" LOG="$tmp/log" "$sync" 2>&1)" || rc=$?
[ "$rc" -ne 0 ] && grep -q 'BLOCKED' <<<"$out" && ok "missing GITHUB_REPOSITORY or GH_TOKEN is BLOCKED" || ko "missing env (rc=$rc $out)"

# Static checks of the workflow.
check() {
  if python3 -I - "$wf" "$1" <<'PY'
import sys, yaml
w = yaml.safe_load(open(sys.argv[1])); on = w.get(True, w.get("on"))
jobs = w["jobs"]; runs = [s.get("run", "") for j in jobs.values() for s in j.get("steps", [])]
sys.exit(0 if eval(sys.argv[2]) else 1)
PY
  then ok "$2"; else ko "$2"; fi
}
[ -f "$wf" ] || { ko "workflow exists"; finish; }
check 'sorted(on) == ["workflow_dispatch", "workflow_run"] and sorted(on["workflow_run"]["workflows"]) == ["Deploy", "Deploy (emergency)"] and on["workflow_run"]["types"] == ["completed"]' "triggers: Deploy and Deploy (emergency) completed, or a manual run"
check 'not any("pull_request" in str(k) for k in on)' "no pull_request trigger (the job holds REVIEW_AGENT_TOKEN)"
check 'all(j.get("environment") == "review-agent" for j in jobs.values()) and w.get("permissions") == {}' "review-agent environment, no default token permission"
check '"head_branch == \x27main\x27" in str(list(jobs.values())[0].get("if", ""))' "only a run on main syncs"
check 'w["concurrency"]["group"] == "sync-main-into-develop" and w["concurrency"]["cancel-in-progress"] is False' "one sync at a time"
check 'not any("${{" in r for r in runs)' "no expression interpolated in a run: (values through env)"
check 'any("scripts/sync-main-into-develop.sh" in r for r in runs)' "the job runs the tested script"
finish
