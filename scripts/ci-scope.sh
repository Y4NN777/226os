#!/usr/bin/env bash
# CI scope (ADR-0010): reads the changed paths on stdin, NUL-separated (`git diff --name-only --no-renames -z
# base...head`: --no-renames lists BOTH sides of a move, -z keeps odd file names intact) and prints `docs` or `full`.
#   docs  every path is documentation that no gate reads: a Markdown file at the repository root, a document or image
#         under docs/ (except docs/quality/, read by coverage-go.sh) or prototypes/, LICENSE, NOTICE.
#         If a gate starts reading a path that matches, add that path to the `full` cases below.
#   full  anything else, an empty list, a blank line, a path with `..`, or any doubt. Fail safe: only a positive
#         match skips work. CI also forces `full` for pull requests into main, pushes and manual runs.
# Even for `docs`, CI still runs the repository group (secret scan, encrypted-file check, licences).
#
# Usage: git diff --name-only --no-renames -z <base>...<head> | scripts/ci-scope.sh
set -euo pipefail

if [ $# -ne 0 ]; then
  printf 'usage: git diff --name-only --no-renames -z <base>...<head> | %s\n' "$0" >&2
  exit 64
fi

seen=0
while IFS= read -r -d '' path; do
  seen=1
  case "$path" in
    '' | *..* | /*) echo full; exit 0 ;;
    docs/quality/*) echo full; exit 0 ;;
    docs/*.md | docs/*.txt | docs/*.png | docs/*.jpg | docs/*.jpeg | docs/*.gif | docs/*.svg | docs/*.webp | docs/*.pdf) ;;
    prototypes/*.html | prototypes/*.css | prototypes/*.js | prototypes/*.md | prototypes/*.txt | prototypes/*.png | prototypes/*.jpg | prototypes/*.jpeg | prototypes/*.gif | prototypes/*.svg | prototypes/*.webp | prototypes/*.json) ;;
    LICENSE | NOTICE) ;;
    */*) echo full; exit 0 ;;
    *.md) ;;
    *) echo full; exit 0 ;;
  esac
done

if [ "$seen" -eq 0 ]; then echo full; else echo docs; fi
