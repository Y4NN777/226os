#!/usr/bin/env bash
# Forward-only migration policy check (ADR-0005 point 3, DATABASE_DESIGN.md 5).
#
# Usage: scripts/check-migrations.sh [migrations-dir]      (default: db/migrations)
# Env:   MIGRATIONS_BASE_REF  optional git ref (e.g. origin/main). When set, any
#                             modified, deleted or renamed migration relative to
#                             that ref fails (files are immutable after merge).
#
# Exit 0 when every rule holds, 1 otherwise. Read-only.
set -euo pipefail

dir="${1:-db/migrations}"
fail=0
err() { printf 'check-migrations: %s\n' "$*" >&2; fail=1; }

[ -d "$dir" ] || { err "not a directory: $dir"; exit 1; }

required_header=(
  'Purpose' 'Tables/indexes' 'Lock risk' 'Data volume' 'Transaction'
  'Compatibility window' 'Restore plan' 'Owner'
)

expected=1
count=0
shopt -s nullglob dotglob
for path in "$dir"/*; do
  name="${path##*/}"
  count=$((count + 1))

  if [ ! -f "$path" ] || ! [[ "$name" =~ ^([0-9]{5})_[a-z0-9_]+\.sql$ ]]; then
    err "$name: file name must match NNNNN_slug.sql (lower-case slug)"
    continue
  fi
  number=$((10#${BASH_REMATCH[1]}))
  if [ "$number" -ne "$expected" ]; then
    err "$name: expected migration number $(printf '%05d' "$expected") (sequential, no gap, no duplicate)"
  fi
  expected=$((number + 1))

  if grep -Eiq '^[[:space:]]*--[[:space:]]*\+goose[[:space:]]+Down' "$path"; then
    err "$name: contains a goose Down section (migrations are forward-only)"
  fi
  # goose parses ANY comment containing the plus-goose token as an annotation, so
  # every such line must be a known annotation at the start of the line
  # (a Down marker is reported by the rule above).
  bad="$(grep -nF '+goose' "$path" \
    | grep -Ev ':[[:space:]]*--[[:space:]]*\+goose[[:space:]]+(Up|Down|StatementBegin|StatementEnd|NO TRANSACTION)[[:space:]]*$' || true)"
  if [ -n "$bad" ]; then
    err "$name: unsupported goose annotation or prose containing the goose token: $(printf '%s' "$bad" | head -n 1)"
  fi
  if ! grep -Eq '^--[[:space:]]*\+goose[[:space:]]+Up[[:space:]]*$' "$path"; then
    err "$name: missing '-- +goose Up'"
    continue
  fi

  # Header = everything before the Up marker.
  header="$(sed -n '/^--[[:space:]]*+goose[[:space:]]\+Up/q;p' "$path")"
  for field in "${required_header[@]}"; do
    if ! printf '%s\n' "$header" | grep -Eq "^-- ${field}:[[:space:]]*[^[:space:]]"; then
      err "$name: header field '${field}:' missing or empty"
    fi
  done
done

[ "$count" -gt 0 ] || err "no migration files in $dir"

if [ -n "${MIGRATIONS_BASE_REF:-}" ]; then
  # Fail closed: an unresolvable ref (typo, shallow clone) or a failing diff must
  # never look like "nothing changed".
  if ! git -C "$dir" rev-parse --verify --quiet "${MIGRATIONS_BASE_REF}^{commit}" >/dev/null; then
    err "MIGRATIONS_BASE_REF '${MIGRATIONS_BASE_REF}' does not resolve to a commit (shallow clone or typo?); cannot verify immutability"
  elif ! changed="$(git -C "$dir" diff --name-status --diff-filter=MDRTU "${MIGRATIONS_BASE_REF}...HEAD" -- .)"; then
    err "git diff against ${MIGRATIONS_BASE_REF} failed; cannot verify immutability"
  elif [ -n "$changed" ]; then
    err "applied migrations must not change since ${MIGRATIONS_BASE_REF}:"
    printf '%s\n' "$changed" >&2
  fi
fi

if [ "$fail" -ne 0 ]; then
  exit 1
fi
printf 'check-migrations: OK (%d files in %s)\n' "$count" "$dir"
