#!/usr/bin/env bash
# Container verification (ADR-0006): static policy checks, `compose config` for every profile,
# image builds, and `docker inspect` of the built images. Nothing is started and no secret is read:
# dummy env files are generated in a temp dir. Exit 0 only when every check passes.
#
# Usage: scripts/verify-containers.sh [--no-build]
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"
build=1
[ "${1-}" = "--no-build" ] && build=0

fails=0
pass() { printf 'PASS  %s\n' "$1"; }
fail() { printf 'FAIL  %s\n' "$1" >&2; fails=$((fails + 1)); }
need() { command -v "$1" >/dev/null 2>&1 || { printf 'BLOCKED: %s is required\n' "$1" >&2; exit 1; }; }
need docker
need python3

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# 1. Digest pins: every external FROM and every non-local image: line.
for f in src/api/Dockerfile src/web/Dockerfile; do
  bad=0
  stages=" "
  while read -r kw ref rest; do
    [ "${kw^^}" = FROM ] || continue
    [[ "$ref" == --platform=* ]] && { ref="${rest%% *}"; }
    if [[ "$stages" == *" $ref "* || "$ref" == scratch ]]; then :; elif [[ ! "$ref" =~ @sha256:[0-9a-f]{64}$ ]]; then
      fail "$f: FROM $ref is not pinned by digest"; bad=1
    fi
    if [[ "${rest,,}" =~ (^|[[:space:]])as[[:space:]]+([A-Za-z0-9_.-]+) ]]; then stages+="${BASH_REMATCH[2]} "; fi
  done <"$f"
  [ "$bad" -eq 0 ] && pass "$f: all FROM lines pinned by digest"
done
bad=0
while IFS= read -r line; do
  ref="${line#*image:}"; ref="${ref# }"
  if [[ "$ref" =~ @sha256:[0-9a-f]{64}$ ]]; then continue; fi
  # Locally built images are variables whose default is a 226os-* tag (deployed profiles override with a digest).
  if [[ "$ref" =~ ^\$\{[A-Z_]+:-226os-[a-z]+:dev\}$ ]]; then continue; fi
  fail "compose.yaml: image '$ref' is not pinned by digest"; bad=1
done < <(grep -E '^[[:space:]]+image:' compose.yaml)
[ "$bad" -eq 0 ] && pass "compose.yaml: external images pinned by digest"

# 2. Dummy env files, one per profile, then `compose config` + policy checks on the resolved model.
mk_env() { # profile
  cat >"$tmp/env.$1" <<ENV
APP_ENV=$1
POSTGRES_USER=app226
POSTGRES_DB=app226
POSTGRES_PASSWORD=dummy-not-a-secret
DATABASE_URL=postgres://app226:dummy-not-a-secret@postgres:5432/app226?sslmode=disable
CURSOR_KEY=ZHVtbXktbm90LWEtc2VjcmV0LWR1bW15LW5vdC1hLXNlY3JldA
PUBLIC_HOST=verify.example.test
ACME_EMAIL=ops@example.test
ENV
}
for p in local-dev staging-test production-live; do
  mk_env "$p"
  if COMPOSE_ENV_FILE="$tmp/env.$p" scripts/compose.sh "$p" config --format json >"$tmp/config.$p.json" 2>"$tmp/config.$p.err"; then
    pass "compose config valid: $p"
  else
    fail "compose config invalid: $p ($(tail -n1 "$tmp/config.$p.err"))"; continue
  fi
  if python3 -I - "$p" "$tmp/config.$p.json" <<'PY'
import json, sys
profile, path = sys.argv[1], sys.argv[2]
cfg = json.load(open(path))
errs = []
svcs = cfg["services"]
if not svcs:
    errs.append("no services selected")
for name, s in svcs.items():
    if name == "postgres" and s.get("ports"):
        errs.append("postgres publishes ports")
    if s.get("ports") and not name.startswith("traefik"):
        errs.append(f"{name} publishes ports (only traefik may)")
    for p in s.get("ports", []):
        if profile != "local-dev" and name == "traefik" and p.get("published") not in ("80", "443"):
            errs.append(f"traefik unexpected port {p}")
        if profile == "local-dev" and p.get("host_ip") != "127.0.0.1":
            errs.append(f"{name} local-dev port not bound to 127.0.0.1: {p}")
    if s.get("privileged"):
        errs.append(f"{name} is privileged")
    for v in s.get("volumes", []):
        if "docker.sock" in str(v.get("source", "")):
            errs.append(f"{name} mounts the Docker socket")
    if s.get("cap_drop") != ["ALL"]:
        errs.append(f"{name} cap_drop is not [ALL]")
    if "no-new-privileges:true" not in s.get("security_opt", []):
        errs.append(f"{name} lacks no-new-privileges")
    if not s.get("read_only"):
        errs.append(f"{name} is not read_only")
    user = str(s.get("user", ""))
    if name != "acme-init" and (not user or user.split(":")[0] in ("0", "root")):
        errs.append(f"{name} does not run as an explicit non-root user ({user!r})")
    if not s.get("logging"):
        errs.append(f"{name} has no logging config")
    lim = (s.get("deploy") or {}).get("resources", {}).get("limits", {})
    if not lim.get("memory") or not lim.get("cpus"):
        errs.append(f"{name} lacks cpu/memory limits")
    if name != "acme-init" and name != "migrate" and not s.get("healthcheck") and name != "migrate-gated":
        errs.append(f"{name} lacks a healthcheck")
    if name in ("postgres", "api", "web") and "backend" in (s.get("networks") or {}) and name == "web":
        errs.append("web must not join the backend network")
    if name == "postgres" and set(s.get("networks") or {}) != {"backend"}:
        errs.append("postgres must be on the backend network only")
    env = s.get("environment") or {}
    if "APP_ENV" in env and env["APP_ENV"] != profile:
        errs.append(f"{name} APP_ENV={env['APP_ENV']} != profile {profile}")
if not cfg["networks"]["backend"].get("internal"):
    errs.append("backend network is not internal")
# The API has database code (S0): it must receive its two secrets, and no other service may get the
# cursor-signing key.
api_env = (svcs.get("api") or {}).get("environment") or {}
for var in ("DATABASE_URL", "CURSOR_KEY"):
    if "api" in svcs and var not in api_env:
        errs.append(f"api does not receive {var}")
for name, s in svcs.items():
    if name != "api" and "CURSOR_KEY" in (s.get("environment") or {}):
        errs.append(f"{name} receives CURSOR_KEY (the API only)")
expected = {"local-dev": {"postgres", "migrate", "api", "web", "traefik-local"},
            "staging-test": {"postgres", "api", "web", "traefik", "acme-init"},
            "production-live": {"postgres", "api", "web", "traefik", "acme-init"}}[profile]
if set(svcs) != expected:
    errs.append(f"services {sorted(svcs)} != expected {sorted(expected)}")
for e in errs:
    print("  " + e, file=sys.stderr)
sys.exit(1 if errs else 0)
PY
  then pass "compose policy (ports, caps, read-only, users, limits, healthchecks, APP_ENV): $p"
  else fail "compose policy violations: $p"; fi
done

# Gated migration job resolves only with its own profile.
if COMPOSE_ENV_FILE="$tmp/env.staging-test" scripts/compose.sh staging-test --profile migrate-gated config --services 2>/dev/null | grep -qx migrate-gated; then
  pass "migrate-gated exists only under the explicit migrate-gated profile"
else fail "migrate-gated profile not resolvable"; fi
if scripts/compose.sh staging-test config --services 2>/dev/null | grep -qx 'migrate\(-gated\)\?'; then
  fail "a migration job is part of the default staging-test profile"
else pass "no migration job in the default staging-test / production-live profiles"; fi

# Mismatched APP_ENV must be refused by the wrapper.
if COMPOSE_ENV_FILE="$tmp/env.local-dev" scripts/compose.sh production-live config >/dev/null 2>&1; then
  fail "APP_ENV/profile mismatch was NOT refused"
else pass "APP_ENV/profile mismatch refused"; fi

# 3. Build and inspect.
if [ "$build" -eq 1 ]; then
  rev="$(git rev-parse HEAD 2>/dev/null || echo unknown)"
  for spec in "api:src/api/Dockerfile:226os-api:dev" "migrate:src/api/Dockerfile:226os-migrate:dev" "web:src/web/Dockerfile:226os-web:dev"; do
    IFS=: read -r target file img tag <<<"$spec"
    if docker build -q -f "$file" --target "$target" --build-arg "VCS_REF=$rev" -t "$img:$tag" . >/dev/null; then
      pass "docker build $img:$tag"
    else fail "docker build $img:$tag"; fi
  done
fi
for img in 226os-api:dev 226os-migrate:dev 226os-web:dev; do
  if ! docker image inspect "$img" >/dev/null 2>&1; then fail "image $img missing (build first)"; continue; fi
  user="$(docker image inspect "$img" --format '{{.Config.User}}')"
  if [[ "$user" =~ ^[1-9][0-9]*(:[1-9][0-9]*)?$ ]]; then pass "$img runs as numeric non-root user $user"; else fail "$img user is '$user'"; fi
  size="$(docker image inspect "$img" --format '{{.Size}}')"
  printf 'INFO  %s size=%s bytes\n' "$img" "$size"
  case "$img" in
    *api* | *web*)
      if [ -n "$(docker image inspect "$img" --format '{{if .Config.Healthcheck}}yes{{end}}')" ]; then pass "$img declares a HEALTHCHECK"; else fail "$img has no HEALTHCHECK"; fi ;;
  esac
done

if [ "$fails" -ne 0 ]; then printf 'verify-containers: %d check(s) FAILED\n' "$fails" >&2; exit 1; fi
echo "verify-containers: ALL PASS"
