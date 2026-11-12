#!/usr/bin/env bash
# The one way to create, edit, decrypt and rotate the SOPS + age secrets files (ADR-0009).
#
#   secrets/<profile>.enc.env      variable names in clear, every value encrypted; COMMITTED
#   secrets/<profile>.recipients   PUBLIC age keys allowed to decrypt that profile; COMMITTED
#   .env.<profile>                 the decrypted file the Compose wrapper reads; git-ignored, mode 600
#   ~/.config/sops/age/keys.txt    each person's PRIVATE age key; never in Git, never printed here
#
# Commands:
#   keygen                  create your age key (refuses to overwrite one) and print the PUBLIC key
#   init <profile>          encrypt the env template (.env.example, comments stripped) for the profile's recipients
#   edit <profile>          edit the encrypted file in $EDITOR (sops decrypts and re-encrypts)
#   decrypt <profile> [--force]   write .env.<profile> (mode 600); refuses to overwrite without --force
#   rotate <profile>        re-encrypt for the CURRENT recipients list and a fresh data key
#   check                   verify nothing under secrets/ is plaintext (scripts/check-secrets-encrypted.sh)
# Profiles: local-dev, staging-test, production-live. Needs sops and age (keygen only needs age-keygen).
#
# Test and CI overrides: SECRETS_DIR, SECRETS_OUT_DIR, SECRETS_ENV_TEMPLATE, SOPS_AGE_KEY_FILE.
set -euo pipefail
umask 077

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
secrets_dir="${SECRETS_DIR:-$root/secrets}"
out_dir="${SECRETS_OUT_DIR:-$root}"
template="${SECRETS_ENV_TEMPLATE:-$root/.env.example}"
export SOPS_AGE_KEY_FILE="${SOPS_AGE_KEY_FILE:-${XDG_CONFIG_HOME:-$HOME/.config}/sops/age/keys.txt}"

die() { printf 'secrets: %s\n' "$*" >&2; exit 1; }
usage() {
  sed -n '2,19p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//' >&2
  printf 'usage: %s keygen | init|edit|decrypt|rotate <profile> | check\n' "$0" >&2
  exit 64
}

need_sops() { command -v sops >/dev/null 2>&1 || die "sops is not installed (https://github.com/getsops/sops); install it, then retry"; }

profile_arg() {
  case "${1:-}" in
    local-dev | staging-test | production-live) printf '%s' "$1" ;;
    *) die "profile must be local-dev, staging-test or production-live (got '${1:-}')" ;;
  esac
}

# recipients PROFILE: the comma-separated public age keys of secrets/<profile>.recipients.
recipients() {
  local f="$secrets_dir/$1.recipients" line list=""
  [ -f "$f" ] || die "$f is missing: each person runs 'scripts/secrets.sh keygen', then adds the PUBLIC key line to that file in a pull request"
  while IFS= read -r line || [ -n "$line" ]; do
    case "$line" in '' | '#'*) continue ;; esac
    [[ "$line" =~ ^age1[a-z0-9]{58}$ ]] || die "$f: '$line' is not a public age key (age1...)"
    list="${list:+$list,}$line"
  done <"$f"
  [ -n "$list" ] || die "$f lists no recipient: add at least one public age key"
  printf '%s' "$list"
}

# Plaintext is only ever held in a private directory (RAM-backed when available), removed on exit.
private_tmp() {
  local base="${TMPDIR:-/tmp}"
  [ -d /dev/shm ] && [ -w /dev/shm ] && base=/dev/shm
  mktemp -d "$base/secrets.XXXXXX"
}

cmd="${1:-}"
[ -n "$cmd" ] || usage
shift || true

case "$cmd" in
  keygen)
    command -v age-keygen >/dev/null 2>&1 || die "age-keygen is not installed (https://github.com/FiloSottile/age)"
    [ ! -e "$SOPS_AGE_KEY_FILE" ] || die "$SOPS_AGE_KEY_FILE already exists: not overwriting it. Your public key: $(age-keygen -y "$SOPS_AGE_KEY_FILE")"
    mkdir -p "$(dirname "$SOPS_AGE_KEY_FILE")"
    age-keygen -o "$SOPS_AGE_KEY_FILE" >/dev/null 2>&1
    chmod 600 "$SOPS_AGE_KEY_FILE"
    printf 'Created %s (private, mode 600: back it up somewhere safe, never commit it).\n' "$SOPS_AGE_KEY_FILE"
    printf 'Your PUBLIC key, to add to secrets/<profile>.recipients in a pull request:\n%s\n' "$(age-keygen -y "$SOPS_AGE_KEY_FILE")"
    ;;

  init)
    p="$(profile_arg "${1:-}")"
    need_sops
    target="$secrets_dir/$p.enc.env"
    [ ! -e "$target" ] || die "$target exists: use 'edit' (or 'rotate' after a recipients change)"
    [ -f "$template" ] || die "env template $template not found"
    list="$(recipients "$p")"
    tmpd="$(private_tmp)"
    trap 'rm -rf "$tmpd"' EXIT
    # Comments and blank lines are dropped: a comment would be plaintext inside the encrypted file.
    grep -vE '^[[:space:]]*(#|$)' "$template" >"$tmpd/in.env" || true
    mkdir -p "$secrets_dir"
    sops encrypt --age "$list" --input-type dotenv --output-type dotenv "$tmpd/in.env" >"$tmpd/out.enc.env"
    mv "$tmpd/out.enc.env" "$target"
    printf 'Created %s from the template (placeholders). Run: scripts/secrets.sh edit %s\n' "$target" "$p"
    ;;

  edit)
    p="$(profile_arg "${1:-}")"
    need_sops
    [ -f "$secrets_dir/$p.enc.env" ] || die "$secrets_dir/$p.enc.env does not exist: run 'init $p' first"
    exec sops edit --input-type dotenv --output-type dotenv "$secrets_dir/$p.enc.env"
    ;;

  decrypt)
    p="$(profile_arg "${1:-}")"
    force=0
    [ "${2:-}" = "--force" ] && force=1
    need_sops
    src="$secrets_dir/$p.enc.env"
    dest="$out_dir/.env.$p"
    [ -f "$src" ] || die "$src does not exist"
    [ "$force" -eq 1 ] || [ ! -e "$dest" ] || die "$dest exists: refusing to overwrite it (use --force)"
    mkdir -p "$out_dir"
    partial="$dest.partial.$$"
    trap 'rm -f "$partial"' EXIT
    sops decrypt --input-type dotenv --output-type dotenv "$src" >"$partial" 2>/dev/null || die "could not decrypt $src: is your private key ($SOPS_AGE_KEY_FILE) one of its recipients?"
    chmod 600 "$partial"
    mv "$partial" "$dest"
    printf 'Wrote %s (mode 600, git-ignored). It holds plaintext secrets: do not copy it anywhere.\n' "$dest"
    ;;

  rotate)
    p="$(profile_arg "${1:-}")"
    need_sops
    src="$secrets_dir/$p.enc.env"
    [ -f "$src" ] || die "$src does not exist"
    list="$(recipients "$p")"
    tmpd="$(private_tmp)"
    trap 'rm -rf "$tmpd"' EXIT
    sops decrypt --input-type dotenv --output-type dotenv "$src" >"$tmpd/plain.env" 2>/dev/null || die "could not decrypt $src: is your private key ($SOPS_AGE_KEY_FILE) one of its recipients?"
    sops encrypt --age "$list" --input-type dotenv --output-type dotenv "$tmpd/plain.env" >"$tmpd/new.enc.env"
    mv "$tmpd/new.enc.env" "$src"
    printf 'Re-encrypted %s for the current recipients with a fresh data key.\n' "$src"
    ;;

  check)
    exec "$root/scripts/check-secrets-encrypted.sh" "$secrets_dir"
    ;;

  *) usage ;;
esac
