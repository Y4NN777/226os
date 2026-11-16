#!/usr/bin/env bash
# Local AI code review with open-code-review (Apache-2.0), pinned to one exact version and run through npx:
# nothing is installed globally. The tool sends the diff (changed files) to the LLM provider you configure:
# only run it on code you are allowed to send there.
#
# Usage:
#   scripts/review.sh login           sign in with ChatGPT (local, one person; a preview of the tool) and select it
#   scripts/review.sh [run] [args]    review HEAD against $REVIEW_BASE (default origin/develop); extra args go to `ocr review`
#   scripts/review.sh ocr <args...>   any other ocr command, e.g. `config provider`
#
# Environment: REVIEW_BASE (default origin/develop), OCR_VERSION (default: the pin below).
# Project rules: .opencodereview/rule.json. Exit codes: 0 ok, 64 usage, otherwise the tool's own code.
# Never put an API key on the command line: use `ocr config provider` or the provider's environment variable.
set -euo pipefail

OCR_VERSION="${OCR_VERSION:-1.13.8}"
REVIEW_BASE="${REVIEW_BASE:-origin/develop}"
pkg="@alibaba-group/open-code-review@${OCR_VERSION}"

ocr() { npx --yes "$pkg" "$@"; }

cd "$(git rev-parse --show-toplevel)"

cmd="${1:-run}"
[ "$#" -gt 0 ] && shift

case "$cmd" in
  login)
    ocr auth login --provider chatgpt
    ocr config set provider chatgpt
    ;;
  run)
    git rev-parse --verify --quiet "${REVIEW_BASE}^{commit}" >/dev/null || {
      printf 'review: base %s not found; run `git fetch origin` or set REVIEW_BASE\n' "$REVIEW_BASE" >&2
      exit 64
    }
    ocr review --from "$REVIEW_BASE" --to HEAD "$@"
    ;;
  ocr)
    [ "$#" -gt 0 ] || { printf 'usage: scripts/review.sh ocr <args...>\n' >&2; exit 64; }
    ocr "$@"
    ;;
  *)
    printf 'usage: scripts/review.sh [login | run [ocr review args] | ocr <args...>]\n' >&2
    exit 64
    ;;
esac
