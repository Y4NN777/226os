#!/usr/bin/env bash
# Create the next sequential, forward-only migration file with the mandatory header
# (ADR-0005, checked by scripts/check-migrations.sh). No Down section is ever written.
# Usage: scripts/migrate-new.sh <slug> [migrations-dir]      (default dir: db/migrations)
#   slug  lower-case letters, digits and underscores, starting with a letter.
# The header fields are placeholders (TODO) that the author must replace before review.
set -euo pipefail

slug="${1-}"
dir="${2:-db/migrations}"
die() { printf 'migrate-new: %s\n' "$*" >&2; exit 1; }

[[ "$slug" =~ ^[a-z][a-z0-9_]{0,62}$ ]] || die "slug must match ^[a-z][a-z0-9_]*\$ (got '${slug}'); usage: migrate-new.sh <slug> [dir]"
[ -d "$dir" ] || die "not a directory: $dir"

last=0
shopt -s nullglob
for path in "$dir"/*; do
  name="${path##*/}"
  [[ "$name" =~ ^([0-9]{5})_[a-z0-9_]+\.sql$ ]] || die "unexpected file '$name' in $dir (run scripts/check-migrations.sh)"
  n=$((10#${BASH_REMATCH[1]}))
  [ "$n" -gt "$last" ] && last=$n
  [[ "$name" == *"_${slug}.sql" ]] && die "a migration with slug '${slug}' already exists: $name"
done
next="$(printf '%05d' $((last + 1)))"
file="$dir/${next}_${slug}.sql"
[ ! -e "$file" ] || die "refusing to overwrite $file"

umask 022
cat >"$file" <<TEMPLATE
-- Migration: ${next}_${slug}
-- Purpose: TODO one sentence: what changes and why.
-- Tables/indexes: TODO tables, columns, indexes, constraints touched.
-- Lock risk: TODO lock level per statement; use CREATE INDEX CONCURRENTLY with NO TRANSACTION when needed.
-- Data volume: TODO rows affected and expected duration.
-- Transaction: TODO single transaction (goose default) or NO TRANSACTION and why.
-- Compatibility window: TODO how the previous app version behaves against the new schema.
-- Restore plan: TODO forward-only: restore the pre-migration pg_dump -Fc backup or redeploy the previous image by digest.
-- Owner: TODO GitHub username
-- +goose Up
TODO write forward-only SQL here and delete this line.
TEMPLATE
printf 'migrate-new: created %s (fill in the header fields, no Down section)\n' "$file"
