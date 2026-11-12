#!/usr/bin/env bash
# Print the image tag for the current checkout: the full revision SHA, with a -dirty suffix
# when the working tree has uncommitted changes (a tag must never claim a revision it does not match).
# Shared by scripts/build-images.sh and scripts/image-scan.sh.
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."
sha="$(git rev-parse HEAD)"
if [ -n "$(git status --porcelain --untracked-files=normal)" ]; then
  printf '%s-dirty\n' "$sha"
else
  printf '%s\n' "$sha"
fi
