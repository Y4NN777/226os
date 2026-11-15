#!/usr/bin/env bash
# Tests for scripts/server-bootstrap.sh: arguments are checked BEFORE the root check (so these tests run as a normal
# user), a private key is refused without being echoed, a real run refuses to start as non-root, and --dry-run prints
# the steps and changes nothing. Nothing here touches the machine.
# Usage: scripts/tests/server-bootstrap_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
script="$root/scripts/server-bootstrap.sh"

key='ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFakeFakeFakeFakeFakeFakeFakeFakeFakeFakeFake ci@226os'

run() { rc=0; out="$("$script" "$@" 2>&1)" || rc=$?; }

run
[ "$rc" -eq 64 ] && grep -q 'usage' <<<"$out" && ok "no argument is a usage error" || ko "no argument (rc=$rc)"

run 'not a key'
[ "$rc" -eq 64 ] && grep -q 'public key' <<<"$out" && ok "a value that is not an SSH public key is refused" || ko "garbage key (rc=$rc)"

run "$(printf '%s\n%s' "$key" "$key")"
[ "$rc" -eq 64 ] && ok "a multi-line key is refused" || ko "multi-line key (rc=$rc)"

secret='-----BEGIN OPENSSH PRIVATE KEY-----'
run "$secret"
if [ "$rc" -eq 64 ] && ! grep -q 'BEGIN' <<<"$out"; then ok "a private key is refused and never echoed"; else ko "private key (rc=$rc)"; fi

run --bogus "$key"
[ "$rc" -eq 64 ] && ok "an unknown option is a usage error" || ko "unknown option (rc=$rc)"

run --ssh-fingerprint 'SHA256:short' --dry-run "$key"
[ "$rc" -eq 64 ] && grep -q 'ssh-fingerprint' <<<"$out" && ok "a malformed --ssh-fingerprint is refused" || ko "bad fingerprint (rc=$rc)"

run --ssh-fingerprint
[ "$rc" -eq 64 ] && ok "--ssh-fingerprint without a value is a usage error" || ko "fingerprint value (rc=$rc)"

run --yes "$key"
[ "$rc" -eq 64 ] && ok "the old --yes option is gone (a fingerprint is required instead)" || ko "--yes (rc=$rc)"

fp="SHA256:$(printf 'A%.0s' {1..43})"
run --ssh-fingerprint "$fp" --dry-run "$key"
if [ "$rc" -eq 0 ] && grep -q 'sshd -T' <<<"$out" && grep -q 'ufw limit' <<<"$out" && grep -q 'restrict' <<<"$out" && grep -q 'fingerprint' <<<"$out" && grep -q 'age' <<<"$out"; then
  ok "--dry-run describes the sshd port, ufw limit, restricted key, fingerprint check and age"
else ko "dry run details (rc=$rc out=$out)"; fi

run --dry-run "$key"
if [ "$rc" -eq 0 ] && grep -q 'dry run' <<<"$out" && grep -q 'Docker Engine' <<<"$out" && grep -q 'ufw' <<<"$out" &&
  grep -q 'unattended-upgrades' <<<"$out" && grep -q '/opt/226os' <<<"$out" && grep -q 'deploy' <<<"$out"; then
  ok "--dry-run prints every step and exits 0"
else ko "dry run (rc=$rc out=$out)"; fi
! grep -q 'AAAAC3Nza' <<<"$out" && ok "--dry-run does not print the key material" || ko "dry run printed the key"

if [ "$(id -u)" -eq 0 ]; then
  printf 'skip - the non-root refusal cannot be tested as root\n'
else
  run "$key"
  [ "$rc" -eq 2 ] && grep -q 'REFUSED' <<<"$out" && grep -q 'root' <<<"$out" && ok "a real run refuses to start as non-root" || ko "non-root (rc=$rc out=$out)"
fi

finish
