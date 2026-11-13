#!/usr/bin/env bash
# Fail-closed tests for the Makefile (ADR-0006): unimplemented targets exit non-zero with
# NOT IMPLEMENTED / NOT RUN, destructive and live targets refuse before touching Docker,
# `make quality` cannot mask a failing step, and the make sources contain no `|| true`
# and no ignored-error (`-` prefixed) recipe lines.
# Nothing here edits a real file or starts a container: the quality driver is exercised on a
# scratch copy with injected steps, and Docker is replaced by a fake that records its argv.
# Usage: scripts/tests/make_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# A fake docker that answers the version probes and records every other call.
mkdir -p "$tmp/bin"
cat >"$tmp/bin/docker" <<FAKE
#!/usr/bin/env bash
case "\$*" in
  "compose version"* | "--version"* | "version"*) echo "Docker fake 1.0"; exit 0 ;;
esac
printf '%s\n' "\$*" >>"$tmp/docker.log"
exit 99
FAKE
chmod +x "$tmp/bin/docker"

# mk DIR ENV... -- make-args : run make from DIR without inherited make state.
mk() {
  local dir="$1"
  shift
  (cd "$dir" && env -u MAKEFLAGS -u MFLAGS -u MAKELEVEL "$@")
}
# run_make NAME... sets rc and out.
run_make() {
  : >"$tmp/docker.log"
  rc=0
  out="$(mk "$root" env -i PATH="$tmp/bin:$PATH" HOME="$HOME" "$@" 2>&1)" || rc=$?
}

# --- 1. Static policy on the make sources -------------------------------------------------
[ -f "$root/Makefile" ] && ok "Makefile exists" || ko "Makefile exists"
shopt -s nullglob
mk_files=("$root"/Makefile "$root"/mk/*.mk)
shopt -u nullglob
[ "${#mk_files[@]}" -ge 2 ] && ok "mk/*.mk present" || ko "mk/*.mk present"

if grep -nE '\|\|[[:space:]]*(true|:)([[:space:];]|$)' "${mk_files[@]}" >/dev/null 2>&1; then
  ko "no '|| true' / '|| :' in make sources: $(grep -nE '\|\|[[:space:]]*(true|:)([[:space:];]|$)' "${mk_files[@]}" | head -n 3)"
else ok "no '|| true' / '|| :' in make sources"; fi
if grep -nP '^\t[@+]*-' "${mk_files[@]}" >/dev/null 2>&1; then
  ko "no '-' prefixed recipe lines: $(grep -nP '^\t[@+]*-' "${mk_files[@]}" | head -n 3)"
else ok "no '-' prefixed (ignored error) recipe lines"; fi
if grep -nE '^[[:space:]]*-?include[[:space:]]' "$root"/mk/*.mk >/dev/null 2>&1; then
  ko "mk/*.mk must not use include"
else ok "mk files do not include other files"; fi
grep -q '^SHELL := bash$' "$root/Makefile" && ok "SHELL := bash" || ko "SHELL := bash"
grep -q '^\.SHELLFLAGS := -euo pipefail -c$' "$root/Makefile" && ok ".SHELLFLAGS := -euo pipefail -c" || ko ".SHELLFLAGS"
grep -q '^\.DELETE_ON_ERROR:' "$root/Makefile" && ok ".DELETE_ON_ERROR" || ko ".DELETE_ON_ERROR"
grep -q '^\.DEFAULT_GOAL := help$' "$root/Makefile" && ok "default goal is help" || ko "default goal is help"
grep -qE '^include \$\(sort \$\(wildcard mk/\*\.mk\)\)$' "$root/Makefile" && ok "deterministic include of mk/*.mk" || ko "deterministic include"

# --- 2. Every ADR-0006 target exists (dry run: no recipe executed) ----------------------------
targets=(doctor setup generate lint typecheck test-unit coverage test-integration security licenses
  test-e2e perf quality quality-full build up-local-dev down-local-dev up-staging-test down-staging-test
  up-production-live down-production-live migrate-up migrate-status migrate-new seed db-reset help)
missing=""
db="$(mk "$root" env -i PATH="$PATH" HOME="$HOME" make -pn 2>/dev/null)" || db=""
for t in "${targets[@]}"; do
  rc=0
  mk "$root" env -i PATH="$PATH" HOME="$HOME" make -n "$t" >/dev/null 2>&1 || rc=$?
  # up-production-live/db-reset guards run in a dry run too; only "No rule" (rc 2 from make itself) matters
  if [ "$rc" -ne 0 ] && mk "$root" env -i PATH="$PATH" HOME="$HOME" make -n "$t" 2>&1 | grep -q 'No rule to make target'; then
    missing+=" $t"
  fi
  grep -q "^$t:" <<<"$db" || missing+=" $t(not-a-rule)"
done
[ -z "$missing" ] && ok "all ${#targets[@]} ADR-0006 targets are defined" || ko "missing targets:$missing"
rc=0
grep -qE '^\.PHONY:.*\bquality\b' <<<"$db" || rc=$?
[ "$rc" -eq 0 ] && ok "targets are .PHONY" || ko "targets are .PHONY"

# --- 3. Unimplemented targets fail closed ---------------------------------------------------
for t in test-e2e perf seed; do
  run_make make "$t"
  if [ "$rc" -ne 0 ] && grep -q 'NOT IMPLEMENTED' <<<"$out" && grep -q 'NOT RUN' <<<"$out" && ! grep -q 'PASS' <<<"$out"; then
    ok "make $t exits non-zero with NOT IMPLEMENTED / NOT RUN (rc=$rc)"
  else ko "make $t (rc=$rc, out=$out)"; fi
done

# --- 4. db-reset refuses before touching Docker -----------------------------------------------
db_url='postgres://app:x@postgres:5432/app226'
for envs in "APP_ENV=production-live" "APP_ENV=staging-test" "APP_ENV=" "APP_ENV=local-dev-x" "DUMMY=1"; do
  run_make make db-reset "$envs" "DATABASE_URL=$db_url"
  if [ "$rc" -ne 0 ] && grep -q 'REFUSED' <<<"$out" && [ ! -s "$tmp/docker.log" ]; then
    ok "make db-reset refused with [$envs] and no docker call"
  else ko "make db-reset [$envs] (rc=$rc, docker=$(cat "$tmp/docker.log"), out=$out)"; fi
done
# Even a caller-supplied environment cannot reach the script with a non-local APP_ENV.
run_make make db-reset APP_ENV=local-dev "DATABASE_URL=postgres://app:x@db.prod.example.com:5432/app226"
if [ "$rc" -ne 0 ] && grep -q 'REFUSED' <<<"$out" && [ ! -s "$tmp/docker.log" ]; then
  ok "make db-reset still refuses a remote DATABASE_URL (script guard intact)"
else ko "remote host (rc=$rc, out=$out)"; fi
run_make make db-reset APP_ENV=local-dev "DATABASE_URL=$db_url" DRY_RUN=1
if [ "$rc" -eq 0 ] && grep -q 'DRY_RUN' <<<"$out" && [ ! -s "$tmp/docker.log" ]; then
  ok "make db-reset passes the guards in local-dev (dry run, no docker call)"
else ko "db-reset dry run (rc=$rc, out=$out)"; fi

# --- 5. up-production-live refuses outside the documented deploy context ----------------------
run_make make up-production-live
[ "$rc" -ne 0 ] && grep -q 'REFUSED' <<<"$out" && [ ! -s "$tmp/docker.log" ] && ok "up-production-live refused without confirmation" || ko "up-production-live bare (rc=$rc, out=$out)"
run_make make up-production-live PRODUCTION_UP_CONFIRM=production-live
[ "$rc" -ne 0 ] && grep -q 'REFUSED' <<<"$out" && [ ! -s "$tmp/docker.log" ] && ok "up-production-live refused with confirmation but outside CI" || ko "up-production-live no CI (rc=$rc, out=$out)"
run_make make up-production-live CI=true PRODUCTION_UP_CONFIRM=yes
[ "$rc" -ne 0 ] && grep -q 'REFUSED' <<<"$out" && [ ! -s "$tmp/docker.log" ] && ok "up-production-live refused with a wrong confirmation value" || ko "up-production-live wrong confirm (rc=$rc, out=$out)"
run_make make up-production-live CI=true PRODUCTION_UP_CONFIRM=production-live COMPOSE_ENV_FILE=/nonexistent/.env
[ "$rc" -ne 0 ] && grep -q 'REFUSED' <<<"$out" && ! grep -q ' up' "$tmp/docker.log" && ok "up-production-live with full confirmation still stops at the env guard, no compose up" || ko "up-production-live full (rc=$rc, docker=$(cat "$tmp/docker.log"))"

# --- 6. down never removes volumes; db and live recipes are visible in a dry run --------------
for p in local-dev staging-test production-live; do
  dry="$(mk "$root" env -i PATH="$PATH" HOME="$HOME" make -n "down-$p" 2>&1)" || true
  if grep -q "scripts/compose.sh $p down" <<<"$dry" && ! grep -qE '(^|[[:space:]])(-v|--volumes)([[:space:]]|$)' <<<"$dry"; then
    ok "down-$p calls compose.sh down without -v/--volumes"
  else ko "down-$p dry run: $dry"; fi
done

# --- 7. quality aggregation cannot mask a failure (scratch copy, injected steps) ---------------
scratch="$tmp/scratch"
mkdir -p "$scratch/scripts"
cp "$root/Makefile" "$scratch/"
cp -r "$root/mk" "$scratch/mk"
cp "$root/scripts/quality.sh" "$scratch/scripts/" 2>/dev/null || true
cat >"$scratch/mk/99-scratch.mk" <<'SCRATCH'
.PHONY: t-ok t-fail t-after
t-ok:
	@echo step t-ok ran
t-fail:
	@echo step t-fail ran; exit 3
t-after:
	@echo step t-after ran
SCRATCH
srun() { rc=0; out="$(mk "$scratch" env -i PATH="$PATH" HOME="$HOME" "$@" 2>&1)" || rc=$?; }

srun make QUALITY_STEPS="t-ok t-fail t-after" quality
if [ "$rc" -ne 0 ] && grep -qE 'FAIL +t-fail' <<<"$out" && grep -qE 'PASS +t-ok' <<<"$out" \
  && grep -q 'step t-after ran' <<<"$out" && grep -qE 'PASS +t-after' <<<"$out" && grep -q 'quality: FAILED' <<<"$out"; then
  ok "quality exits non-zero when a step fails, runs the remaining steps and lists the failure (rc=$rc)"
else ko "quality with failing step (rc=$rc, out=$out)"; fi

srun make QUALITY_STEPS="t-ok t-fail t-after" QUALITY_FAIL_FAST=1 quality
if [ "$rc" -ne 0 ] && grep -qE 'FAIL +t-fail' <<<"$out" && grep -qE 'NOT RUN +t-after' <<<"$out" && ! grep -q 'step t-after ran' <<<"$out"; then
  ok "QUALITY_FAIL_FAST stops at the first failure and reports the rest as NOT RUN"
else ko "quality fail-fast (rc=$rc, out=$out)"; fi

srun make QUALITY_STEPS="t-ok t-after" quality
if [ "$rc" -eq 0 ] && grep -q 'quality: PASS' <<<"$out"; then ok "quality exits 0 only when every step passed"; else ko "quality control (rc=$rc, out=$out)"; fi

srun make QUALITY_STEPS="t-ok no-such-target" quality
[ "$rc" -ne 0 ] && grep -qE 'FAIL +no-such-target' <<<"$out" && ok "a step that does not exist is a FAIL, not a skip" || ko "quality unknown step (rc=$rc, out=$out)"

srun make QUALITY_STEPS="" quality
[ "$rc" -ne 0 ] && ok "quality with no steps exits non-zero (nothing ran is not PASS)" || ko "quality empty (rc=$rc, out=$out)"

srun make QUALITY_STEPS="t-ok" quality-full
if [ "$rc" -ne 0 ] && grep -qE 'FAIL +test-e2e' <<<"$out" && grep -qE 'FAIL +perf' <<<"$out" && grep -qE 'PASS +t-ok' <<<"$out"; then
  ok "quality-full = quality + test-e2e + perf and fails while those are NOT IMPLEMENTED"
else ko "quality-full (rc=$rc, out=$out)"; fi

# Default step list covers G1-G4 (dry run: the driver is not executed).
dry="$(mk "$root" env -i PATH="$PATH" HOME="$HOME" make -n quality 2>&1)" || true
miss=""
for s in doctor generate lint typecheck test-unit coverage test-integration security licenses docs-check; do
  grep -qE "(^|[[:space:]])$s([[:space:]]|$)" <<<"$dry" || miss+=" $s"
done
[ -z "$miss" ] && ok "default quality steps cover G1-G4 (doctor generate lint typecheck test-unit coverage test-integration security licenses docs-check)" || ko "quality default steps missing:$miss"
dry="$(mk "$root" env -i PATH="$PATH" HOME="$HOME" make -n quality-full 2>&1)" || true
grep -qE '(^|[[:space:]])test-e2e([[:space:]]|$)' <<<"$dry" && grep -qE '(^|[[:space:]])perf([[:space:]]|$)' <<<"$dry" && ok "quality-full adds test-e2e and perf" || ko "quality-full default steps"

# --- 8. migrate-new writes a policy-compliant forward-only file --------------------------------
mig="$tmp/migrations"
mkdir -p "$mig"
cp "$root"/db/migrations/*.sql "$mig/"
last="$(basename "$(printf '%s\n' "$mig"/*.sql | sort | tail -n 1)")"
last="${last%%_*}"
next="$(printf '%05d' "$((10#$last + 1))")"
rc=0
out="$(env -i PATH="$PATH" HOME="$HOME" "$root/scripts/migrate-new.sh" add_widget_table "$mig" 2>&1)" || rc=$?
new="$mig/${next}_add_widget_table.sql"
if [ "$rc" -eq 0 ] && [ -f "$new" ] && ! grep -qi 'goose Down' "$new" && grep -q '^-- +goose Up$' "$new" \
  && env -i PATH="$PATH" HOME="$HOME" "$root/scripts/check-migrations.sh" "$mig" >/dev/null 2>&1; then
  ok "migrate-new creates the next sequential file, header template, no Down, passes check-migrations"
else ko "migrate-new (rc=$rc, out=$out)"; fi
rc=0
env -i PATH="$PATH" HOME="$HOME" "$root/scripts/migrate-new.sh" 'Bad Slug!' "$mig" >/dev/null 2>&1 || rc=$?
[ "$rc" -ne 0 ] && ok "migrate-new refuses an invalid slug" || ko "migrate-new invalid slug accepted"
rc=0
env -i PATH="$PATH" HOME="$HOME" "$root/scripts/migrate-new.sh" add_widget_table "$mig" >/dev/null 2>&1 || rc=$?
[ "$rc" -ne 0 ] && ok "migrate-new refuses a duplicate slug" || ko "migrate-new duplicate slug accepted"

# CI time: lint already runs every script test; test-integration must only add the Docker pass of the scripts
# that read DB_RESET_TEST_DOCKER (running them all twice cost about two minutes).
flagged="$(grep -l 'DB_RESET_TEST_DOCKER'':-' "$root"/scripts/tests/*_test.sh | xargs -n1 basename | tr '\n' ' ')"
integ="$(awk '/^test-integration-containers:/{f=1;next} f&&/^[^\t]/{f=0} f' "$root/mk/20-gates.mk")"
if grep -q 'for t in scripts/tests/\*_test.sh.*DB_RESET_TEST_DOCKER=1' <<<"$integ"; then ko "test-integration re-runs every script test with the Docker flag"; else ok "test-integration does not re-run every script test"; fi
all_flagged_run=1
for s in $flagged; do grep -q "DB_RESET_TEST_DOCKER=1 scripts/tests/$s" <<<"$integ" || all_flagged_run=0; done
[ -n "$flagged" ] && [ "$all_flagged_run" -eq 1 ] && ok "test-integration runs each Docker-aware script test with the flag ($flagged)" || ko "a Docker-aware script test is not run with the flag (flagged: $flagged)"

# --- CI groups (ADR-0010) --------------------------------------------------------------------------
# CI runs the gates as parallel groups. The split must lose nothing: `make lint` and `make test-integration`
# still run every check, and the groups cover exactly the leaf steps of `make quality` (no more, no less).
dry() { mk "$root" env -i PATH="$PATH" HOME="$HOME" make -n "$@" 2>&1; }
steps_of() { { mk "$root" env -i PATH="$PATH" HOME="$HOME" make -s "$1" 2>/dev/null || true; } | tr ' ' '\n' | sed '/^$/d' | sort -u | tr '\n' ' '; }
lint_dry="$(dry lint)"
miss=""
for w in "gofmt" "go vet" "staticcheck" "golangci-lint" "sqlc vet" "ESLint" "Prettier" "redocly" "migration policy" "ShellCheck" "script tests" "check-todo-updated.sh"; do
  grep -q -- "$w" <<<"$lint_dry" || miss+=" [$w]"
done
[ -z "$miss" ] && ok "make lint still runs every lint check after the split" || ko "make lint lost:$miss"
integ_dry="$(dry test-integration)"
miss=""
for w in "-tags integration" "db-reset_test.sh" "verify-containers.sh"; do
  grep -q -- "$w" <<<"$integ_dry" || miss+=" [$w]"
done
[ -z "$miss" ] && ok "make test-integration still runs the Go, script and container checks after the split" || ko "make test-integration lost:$miss"
ci="$(steps_of ci-steps)"
leaf="$(steps_of quality-leaf-steps)"
if [ -n "$ci" ] && [ "$ci" = "$leaf" ]; then ok "the CI groups cover exactly the leaf steps of make quality ($ci)"; else ko "CI groups differ from the quality steps: ci=[$ci] leaf=[$leaf]"; fi
groups="$({ mk "$root" env -i PATH="$PATH" HOME="$HOME" make -s ci-groups 2>/dev/null || true; } | tr ' ' '\n' | sed '/^$/d' | tr '\n' ' ')"
[ -n "$groups" ] && ok "CI groups are listed ($groups)" || ko "make ci-groups prints nothing"
for g in $groups; do
  d="$(dry "ci-$g")"
  if grep -q "scripts/quality.sh" <<<"$d" && grep -q "ci-$g" <<<"$d" && grep -q "doctor" <<<"$d"; then :; else ko "ci-$g does not run through the quality driver with doctor first: $d"; fi
done
ok "every CI group runs through the fail-closed driver, doctor first"
docs_steps="$({ mk "$root" env -i PATH="$PATH" HOME="$HOME" make -s ci-steps SCOPE_GROUPS=repo 2>/dev/null || true; } | tr ' ' '\n' | sed '/^$/d' | sort -u | tr '\n' ' ')"
[ "$docs_steps" = "docs-check doctor licenses security-secrets security-secrets-encrypted " ] && ok "SCOPE_GROUPS=repo narrows the expected steps to the repository group" || ko "docs-scope steps: [$docs_steps]"
run_make make ci-no-such-group
[ "$rc" -ne 0 ] && ok "an unknown CI group fails (no steps is not a pass)" || ko "unknown CI group passed (rc=$rc, out=$out)"

finish
