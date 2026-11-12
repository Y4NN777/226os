#!/usr/bin/env bash
# RED/GREEN evidence for the runtime decryption in scripts/compose.sh (ADR-0009): when secrets/<profile>.enc.env
# exists, the wrapper decrypts it into a private RAM-backed temp file for the compose command only and removes
# it afterwards, so the app starts from the encrypted file in Git with no plaintext .env on disk.
# Needs sops and age; without them the test says NOT RUN (never a silent pass).
# A fake docker records argv and copies the --env-file it receives, but NEVER a file inside the repository
# (a developer's real .env.<profile> lives there and must not be read by a test).
# Usage: scripts/tests/compose-secrets_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"

if ! command -v sops >/dev/null 2>&1 || ! command -v age-keygen >/dev/null 2>&1; then
  printf 'NOT RUN - the runtime decryption test needs sops and age-keygen on PATH (not found)\n'
  exit 0
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir "$tmp/bin" "$tmp/secrets"

cat >"$tmp/bin/docker" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >>"$LOG"
prev=""
for a in "$@"; do
  if [ "$prev" = --env-file ]; then
    printf '%s\n' "$a" >"$PATHLOG"
    case "$a" in
      "$REPO_ROOT"/*) printf 'REPO-ENV-FILE-NOT-COPIED\n' >"$SEEN" ;;
      *) cp "$a" "$SEEN" ;;
    esac
  fi
  prev="$a"
done
EOF
chmod +x "$tmp/bin/docker"

age-keygen -o "$tmp/k1.txt" >/dev/null 2>&1
age-keygen -o "$tmp/k2.txt" >/dev/null 2>&1
printf '%s\n' "$(age-keygen -y "$tmp/k1.txt")" >"$tmp/secrets/local-dev.recipients"
pw="pw-$(head -c 9 /dev/urandom | base64 -w0 | tr -dc 'a-z0-9')"
printf 'APP_ENV=local-dev\nPOSTGRES_PASSWORD=%s\n' "$pw" >"$tmp/template.env"
SECRETS_DIR="$tmp/secrets" SECRETS_ENV_TEMPLATE="$tmp/template.env" SOPS_AGE_KEY_FILE="$tmp/k1.txt" \
  "$root/scripts/secrets.sh" init local-dev >/dev/null
printf 'APP_ENV=local-dev\nFROM_PLAIN_FILE=yes\n' >"$tmp/env.plain"

run() { : >"$tmp/docker.log"; rm -f "$tmp/seen" "$tmp/path"; rc=0
  out="$(env -i PATH="$tmp/bin:$PATH" HOME="$tmp" LOG="$tmp/docker.log" SEEN="$tmp/seen" PATHLOG="$tmp/path" REPO_ROOT="$root" \
    SECRETS_DIR="$tmp/secrets" "$@" 2>&1)" || rc=$?; }

run SOPS_AGE_KEY_FILE="$tmp/k1.txt" "$root/scripts/compose.sh" local-dev up -d
if [ "$rc" -eq 0 ] && grep -q -- '--profile local-dev up -d' "$tmp/docker.log"; then ok "up runs from the encrypted file alone"; else ko "up ($rc $out)"; fi
if grep -qx "POSTGRES_PASSWORD=$pw" "$tmp/seen" 2>/dev/null; then ok "compose received the decrypted values"; else ko "decrypted values missing: $(cat "$tmp/seen" 2>/dev/null)"; fi
envpath="$(cat "$tmp/path" 2>/dev/null || true)"
case "$envpath" in
  "$root"/*) ko "the env file was written inside the repository: $envpath" ;;
  "") ko "docker never received --env-file" ;;
  *) ok "the decrypted file is outside the repository" ;;
esac
if [ -n "$envpath" ] && [ ! -e "$envpath" ]; then ok "the decrypted file is removed after the command"; else ko "decrypted file still exists: $envpath"; fi

run SOPS_AGE_KEY_FILE="$tmp/k2.txt" "$root/scripts/compose.sh" local-dev up -d
if [ "$rc" -ne 0 ] && [ ! -s "$tmp/docker.log" ]; then ok "a key that is not a recipient fails before docker runs"; else ko "wrong key ($rc, docker log: $(cat "$tmp/docker.log"))"; fi
if grep -q "$pw" <<<"$out"; then ko "the failure output contains the secret"; else ok "the failure output holds no secret"; fi

run SOPS_AGE_KEY_FILE="$tmp/k1.txt" COMPOSE_ENV_FILE="$tmp/env.plain" "$root/scripts/compose.sh" local-dev up -d
if [ "$rc" -eq 0 ] && [ "$(cat "$tmp/path")" = "$tmp/env.plain" ] && grep -qx 'FROM_PLAIN_FILE=yes' "$tmp/seen"; then ok "an explicit COMPOSE_ENV_FILE wins over the encrypted file"; else ko "explicit env file ($rc $out)"; fi

run SOPS_AGE_KEY_FILE="$tmp/k1.txt" "$root/scripts/compose.sh" local-dev down -v
if [ "$rc" -ne 0 ] && [ ! -s "$tmp/docker.log" ] && grep -q REFUSED <<<"$out"; then ok "down -v is still refused with an encrypted source"; else ko "down -v ($rc)"; fi

run SOPS_AGE_KEY_FILE="$tmp/k1.txt" "$root/scripts/compose.sh" staging-test up -d
if [ "$rc" -ne 0 ] && [ ! -s "$tmp/docker.log" ]; then ok "a profile without an encrypted file or a plain file stops before docker"; else ko "unknown source ($rc)"; fi

finish
