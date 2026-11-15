#!/usr/bin/env bash
# scripts/secrets.sh must not export SOPS_AGE_KEY_FILE for a missing default file when the key is in SOPS_AGE_KEY
# (CI), and must keep exporting the default path when there is no SOPS_AGE_KEY (the error message names it).
# A fake sops records the SOPS_AGE_KEY_FILE it sees. Usage: scripts/tests/secrets-key-env_test.sh
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin" "$tmp/sec" "$tmp/home"
printf 'APP_ENV=staging-test\n' >"$tmp/sec/staging-test.enc.env"
cat >"$tmp/bin/sops" <<'FAKESOPS'
#!/usr/bin/env bash
printf 'file=%s key_set=%s\n' "${SOPS_AGE_KEY_FILE-UNSET}" "${SOPS_AGE_KEY:+yes}" >"$SEEN"
echo APP_ENV=staging-test
FAKESOPS
chmod +x "$tmp/bin/sops"
run() { rc=0; env -i PATH="$tmp/bin:$PATH" HOME="$tmp/home" SEEN="$tmp/seen" SECRETS_DIR="$tmp/sec" SECRETS_OUT_DIR="$tmp/out$RANDOM" "$@" "$root/scripts/secrets.sh" decrypt staging-test >/dev/null 2>&1 || rc=$?; }

run SOPS_AGE_KEY=AGE-SECRET-KEY-FAKE
[ "$rc" -eq 0 ] && grep -q 'file=UNSET key_set=yes' "$tmp/seen" && ok "SOPS_AGE_KEY set and no key file: SOPS_AGE_KEY_FILE is not exported" || ko "ci case (rc=$rc seen=$(cat "$tmp/seen" 2>/dev/null))"

mkdir -p "$tmp/home/.config/sops/age"
echo k >"$tmp/home/.config/sops/age/keys.txt"
run SOPS_AGE_KEY=AGE-SECRET-KEY-FAKE
grep -q "file=$tmp/home/.config/sops/age/keys.txt" "$tmp/seen" && ok "an existing default key file is still exported" || ko "default file"
rm -rf "$tmp/home/.config"

run
grep -q "file=$tmp/home/.config/sops/age/keys.txt key_set=" "$tmp/seen" && ok "without SOPS_AGE_KEY the default path is exported (error messages name it)" || ko "no key case"

run SOPS_AGE_KEY=AGE-SECRET-KEY-FAKE SOPS_AGE_KEY_FILE="$tmp/explicit.txt"
grep -q "file=$tmp/explicit.txt" "$tmp/seen" && ok "an explicit SOPS_AGE_KEY_FILE is kept" || ko "explicit"
finish
