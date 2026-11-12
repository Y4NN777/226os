#!/usr/bin/env bash
# Everything under secrets/ must be SOPS-encrypted (ADR-0009). Fails closed.
#
# This is a FORMAT check, not a proof of integrity (only `sops decrypt` proves that): it makes it hard to
# commit a plaintext secret under secrets/ and it ties each encrypted file to its recipients list. The
# secret scanner (gitleaks, default rules) is a second, partly overlapping defense: it does not flag real
# ciphertext and does flag keyword-style plaintext, but it misses plaintext without a keyword, so neither
# defense alone is a guarantee.
#
# Rules for a secrets/ directory (it must exist; an empty one passes):
#   - only regular files (no symlinks): *.enc.env, *.recipients and README.md;
#   - every file at most 256 KiB;
#   - *.enc.env (dotenv, written by sops):
#       * every line is KEY=VALUE, no comments, no blank-key lines;
#       * a normal KEY is UPPER_SNAKE_CASE ([A-Z][A-Z0-9_]*, at most 64) like every variable of this project; its value is EMPTY (sops leaves empty values
#         empty) or a full envelope ENC[AES256_GCM,data:<base64>,iv:<44 chars>,tag:<24 chars>,type:<t>];
#       * the only sops_ keys allowed are the exact metadata keys sops writes, each with its value format:
#         sops_mac (envelope), sops_version (x.y.z), sops_lastmodified (UTC timestamp), sops_unencrypted_suffix
#         (_unencrypted), sops_age__list_N__map_enc (an age envelope), sops_age__list_N__map_recipient
#         (a public age key); any other sops_ key is refused (it could carry plaintext);
#       * sops_mac, sops_version and at least one age recipient are present;
#       * the file's recipients equal the entries of its <profile>.recipients file (a removed person must
#         be followed by `scripts/secrets.sh rotate`, which this check makes visible);
#       * no private age key anywhere;
#   - *.recipients: comments (at most 80 characters), blank lines and PUBLIC age keys (age1...) only;
#   - README.md: at most 8 KiB and no KEY=VALUE line.
# Usage: scripts/check-secrets-encrypted.sh [DIR]    (default: ./secrets)
set -euo pipefail
# Bracket ranges in [[ =~ ]] depend on the locale: force the same verdict everywhere.
export LC_ALL=C

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
dir="${1:-$root/secrets}"
[ -d "$dir" ] || { printf 'check-secrets-encrypted: FAIL: secrets directory not found: %s\n' "$dir" >&2; exit 1; }

fails=0
bad() { printf 'check-secrets-encrypted: FAIL: %s\n' "$*" >&2; fails=$((fails + 1)); }

envelope='^ENC\[AES256_GCM,data:[A-Za-z0-9+/]+=*,iv:[A-Za-z0-9+/]{43}=,tag:[A-Za-z0-9+/]{22}==,type:(str|int|float|bytes|bool)\]$'
recipient='^age1[a-z0-9]{58}$'
age_enc='^-----BEGIN AGE ENCRYPTED FILE-----(\\n[A-Za-z0-9+/=]+)+\\n-----END AGE ENCRYPTED FILE-----\\n$'
timestamp='^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$'
version='^[0-9]+\.[0-9]+\.[0-9]+$'
keyname='^[A-Z][A-Z0-9_]{0,63}$'
privkey='AGE-SECRET-KEY-'
max_size=262144

shopt -s nullglob dotglob
for f in "$dir"/*; do
  name="$(basename "$f")"
  if [ -L "$f" ]; then bad "$name is a symlink: not allowed under secrets/"; continue; fi
  if [ ! -f "$f" ]; then bad "$name is not a regular file"; continue; fi
  size="$(stat -c %s "$f")"
  if [ "$size" -gt "$max_size" ]; then bad "$name is too large ($size bytes, at most $max_size)"; continue; fi
  if grep -qi -- "$privkey" "$f"; then
    bad "$name contains a private age key: rotate it, never commit it"
    continue
  fi
  case "$name" in
    README.md)
      [ "$size" -le 8192 ] || bad "README.md is larger than 8 KiB"
      if grep -Eq '^[A-Za-z_][A-Za-z0-9_]*=.+' "$f"; then bad "README.md has a KEY=VALUE line: documentation must not carry values"; fi
      ;;
    *.recipients)
      n=0 lineno=0
      while IFS= read -r line || [ -n "$line" ]; do
        lineno=$((lineno + 1))
        [ -z "$line" ] && continue
        case "$line" in
          '#'*)
            [ "${#line}" -le 80 ] || bad "$name:$lineno: a comment longer than 80 characters (comments are plaintext)"
            continue
            ;;
        esac
        if [[ "$line" =~ $recipient ]]; then n=$((n + 1)); else bad "$name:$lineno: not a public age key (age1...): recipients files hold public keys only"; fi
      done <"$f"
      [ "$n" -gt 0 ] || bad "$name lists no recipient"
      ;;
    *.enc.env)
      have_mac=0 have_version=0 lineno=0
      file_recipients=()
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
            if [[ "$value" =~ $envelope ]]; then have_mac=1; else bad "$name:$lineno: sops_mac is not an encrypted envelope"; fi
            ;;
          sops_version)
            if [[ "$value" =~ $version ]]; then have_version=1; else bad "$name:$lineno: sops_version is not a version (x.y.z)"; fi
            ;;
          sops_lastmodified)
            [[ "$value" =~ $timestamp ]] || bad "$name:$lineno: sops_lastmodified is not a UTC timestamp"
            ;;
          sops_unencrypted_suffix)
            [ "$value" = "_unencrypted" ] || bad "$name:$lineno: sops_unencrypted_suffix has an unexpected value"
            ;;
          sops_age__list_[0-9]*__map_enc)
            [[ "$key" =~ ^sops_age__list_[0-9]+__map_enc$ ]] || bad "$name:$lineno: malformed sops key name"
            [[ "$value" =~ $age_enc ]] || bad "$name:$lineno: $key is not an age envelope"
            ;;
          sops_age__list_[0-9]*__map_recipient)
            [[ "$key" =~ ^sops_age__list_[0-9]+__map_recipient$ ]] || bad "$name:$lineno: malformed sops key name"
            if [[ "$value" =~ $recipient ]]; then file_recipients+=("$value"); else bad "$name:$lineno: age recipient is not a public key"; fi
            ;;
          sops_*)
            bad "$name:$lineno: unknown sops_ key '$key' (only the exact metadata keys sops writes are allowed; any other could carry plaintext)"
            ;;
          *)
            if [[ ! "$key" =~ $keyname ]]; then
              bad "$name:$lineno: invalid key name (valid names are UPPER_SNAKE_CASE, [A-Z][A-Z0-9_]*): a key name must not carry data"
            elif [ -n "$value" ] && [[ ! "$value" =~ $envelope ]]; then
              bad "$name: $key is not an encrypted value (plaintext or malformed envelope)"
            fi
            ;;
        esac
      done <"$f"
      [ "$have_mac" -eq 1 ] || bad "$name has no sops_mac: not a complete SOPS file"
      [ "$have_version" -eq 1 ] || bad "$name has no sops_version: not a complete SOPS file"
      [ "${#file_recipients[@]}" -gt 0 ] || bad "$name has no age recipient (sops_age__list_N__map_recipient)"
      # The encrypted file and its recipients list must agree, so that removing a person is followed by a rotate.
      rfile="${f%.enc.env}.recipients"
      if [ ! -f "$rfile" ]; then
        bad "$name needs its recipients file $(basename "$rfile")"
      elif [ "${#file_recipients[@]}" -gt 0 ]; then
        want="$(grep -E "$recipient" "$rfile" | sort -u || true)"
        have="$(printf '%s\n' "${file_recipients[@]}" | sort -u)"
        [ "$want" = "$have" ] || bad "$name: its recipients differ from $(basename "$rfile"): run 'scripts/secrets.sh rotate' after changing the list (a removed person can still decrypt until then)"
      fi
      ;;
    *) bad "$name does not belong in secrets/ (only *.enc.env, *.recipients and README.md)" ;;
  esac
done

if [ "$fails" -ne 0 ]; then
  printf 'check-secrets-encrypted: FAIL (%s problem(s)); see docs/decisions/0009-secrets-with-sops-and-age.md\n' "$fails" >&2
  exit 1
fi
echo "check-secrets-encrypted: PASS"
