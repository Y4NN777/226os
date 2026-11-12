#!/usr/bin/env bash
# `make build` (G7 input, ADR-0006): build the three multi-stage images (api, web, migrate)
# tagged with the git SHA and record their digests in $REPORTS_DIR/image-digests.txt.
#
#   226os-api:<sha>  226os-web:<sha>  226os-migrate:<sha>
#
# A dirty working tree gets the tag <sha>-dirty so a tag never claims a revision it does not match.
# The recorded value is the local image ID (content digest of the image config/manifest in the
# local store). The registry digest that G7 deploys by is assigned when the image is pushed to GHCR
# (not done here: no push, no registry credentials).
# Needs the Docker daemon and the pinned base images (pulled by digest on first build).
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

sha="$(git rev-parse HEAD)"
tag="$(scripts/image-tag.sh)"
[ "$tag" = "$sha" ] || printf 'build: working tree is dirty; tagging %s\n' "$tag" >&2

reports="${REPORTS_DIR:-reports}"
mkdir -p "$reports"
digests="$reports/image-digests.txt"
printf '# image digests (local image ID) for revision %s (tag %s)\n' "$sha" "$tag" >"$digests"

# name:dockerfile:target
for spec in "226os-api:src/api/Dockerfile:api" "226os-migrate:src/api/Dockerfile:migrate" "226os-web:src/web/Dockerfile:web"; do
  IFS=: read -r name file target <<<"$spec"
  docker build --file "$file" --target "$target" --build-arg "VCS_REF=$sha" --tag "$name:$tag" .
  id="$(docker image inspect "$name:$tag" --format '{{.Id}}')"
  [[ "$id" =~ ^sha256:[0-9a-f]{64}$ ]] || { printf 'build: unexpected image id for %s: %s\n' "$name" "$id" >&2; exit 1; }
  printf '%s:%s %s\n' "$name" "$tag" "$id" >>"$digests"
done

cat "$digests"
printf 'build: OK (digests recorded in %s)\n' "$digests"
