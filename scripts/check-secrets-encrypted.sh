#!/usr/bin/env bash
# Everything under secrets/ must be SOPS-encrypted (ADR-0009). Fails closed.
#
# Why: secrets live in Git ENCRYPTED (SOPS + age): secrets/<profile>.enc.env holds the variable names in
# clear and every value as an ENC[...] envelope. This check and the secret scanner (gitleaks, default rules,
# which does not flag real ciphertext but does flag plaintext) are two independent defenses against a
# plaintext secret landing under secrets/.
# Pure bash: needs neither sops nor age, so it runs in CI and on any machine.
#
# Rules for a secrets/ directory (default ./secrets; absent = nothing to check, pass):
#   - only *.enc.env, *.recipients and README.md may exist;
#   - *.enc.env: every line is KEY=VALUE; a normal key's value is empty (sops leaves empty values empty) or a
#     full SOPS envelope ENC[AES256_GCM,data:..,iv:..,tag:..,type:..]; SOPS metadata keys (sops_*) are allowed; sops_mac (an
#     envelope), sops_version and at least one age recipient (sops_age__list_N__map_recipient=age1...)
#     must be present; no comments (a comment is plaintext); no private age key anywhere;
#   - *.recipients: only comments, blank lines and PUBLIC age keys (age1...); never a private key.
# Usage: scripts/check-secrets-encrypted.sh [DIR]
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
dir="${1:-$root/secrets}"
[ -d "$dir" ] || { echo "check-secrets-encrypted: PASS (no secrets directory yet)"; exit 0; }

fails=0
bad() { printf 'check-secrets-encrypted: FAIL: %s\n' "$*" >&2; fails=$((fails + 1)); }

envelope='^ENC\[AES256_GCM,data:[A-Za-z0-9+/=]*,iv:[A-Za-z0-9+/=]+,tag:[A-Za-z0-9+/=]+,type:(str|int|float|bytes|bool)\]$'
recipient='^age1[a-z0-9]{58}$'
privkey='AGE-SECRET-KEY-'

shopt -s nullglob dotglob
for f in "$dir"/*; do
  [ -f "$f" ] || { bad "$(basename "$f") is not a regular file"; continue; }
  name="$(basename "$f")"
  if grep -q -- "$privkey" "$f"; then
    bad "$name contains a private age key: rotate it, never commit it"
    continue
  fi
  case "$name" in
    README.md) ;;
    *.recipients)
      n=0
      while IFS= read -r line || [ -n "$line" ]; do
        [ -z "$line" ] && continue
        case "$line" in '#'*) continue ;; esac
        if [[ ! "$line" =~ $recipient ]]; then
          bad "$name: line is not a public age key (age1...): recipients files hold public keys only"
        else
          n=$((n + 1))
        fi
      done <"$f"
      [ "$n" -gt 0 ] || bad "$name lists no recipient"
      ;;
    *.enc.env)
      have_mac=0 have_version=0 have_recipient=0 lineno=0
      while IFS= read -r line || [ -n "$line" ]; do
        lineno=$((lineno + 1))
        [ -z "$line" ] && continue
        case "$line" in
          '#'*) bad "$name:$lineno: a comment is plaintext inside an encrypted file"; continue ;;
        esac
        if [[ "$line" != *=* ]]; then bad "$name:$lineno: not a KEY=VALUE line"; continue; fi
        key="${line%%=*}"
        value="${line#*=}"
        case "$key" in
          sops_mac)
            [[ "$value" =~ $envelope ]] && have_mac=1 || bad "$name: sops_mac is not an encrypted envelope"
            ;;
          sops_version) have_version=1 ;;
          sops_age__list_*__map_recipient)
            [[ "$value" =~ $recipient ]] && have_recipient=1 || bad "$name:$lineno: age recipient is not a public key"
            ;;
          sops_*) ;;
          *)
            # sops leaves an EMPTY value empty (there is nothing to hide); anything else must be an envelope.
            [ -z "$value" ] || [[ "$value" =~ $envelope ]] || bad "$name: $key is not an encrypted value (plaintext or malformed envelope)"
            ;;
        esac
      done <"$f"
      [ "$have_mac" -eq 1 ] || bad "$name has no sops_mac: not a complete SOPS file"
      [ "$have_version" -eq 1 ] || bad "$name has no sops_version: not a complete SOPS file"
      [ "$have_recipient" -eq 1 ] || bad "$name has no age recipient (sops_age__list_N__map_recipient)"
      ;;
    *) bad "$name does not belong in secrets/ (only *.enc.env, *.recipients and README.md)" ;;
  esac
done

if [ "$fails" -ne 0 ]; then
  printf 'check-secrets-encrypted: FAIL (%s problem(s)); see docs/decisions/0009-secrets-with-sops-and-age.md\n' "$fails" >&2
  exit 1
fi
echo "check-secrets-encrypted: PASS"
