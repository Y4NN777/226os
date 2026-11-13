#!/usr/bin/env bash
# Tests for scripts/review.sh with a fake npx on PATH: the version is pinned, the base is checked,
# the tool is never installed globally, and a bad subcommand is a usage error.
# Usage: scripts/tests/review_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
script="$root/scripts/review.sh"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin"
log="$tmp/npx.log"
printf '#!/usr/bin/env bash\nprintf "%%s\\n" "$*" >>"%s"\n' "$log" >"$tmp/bin/npx"
chmod +x "$tmp/bin/npx"

repo="$tmp/repo"
git init -q -b main "$repo"
git -C "$repo" config user.name t
git -C "$repo" config user.email t@example.test
git -C "$repo" -c commit.gpgsign=false commit -q --allow-empty -m init
git -C "$repo" branch origin-develop

run() { rc=0; : >"$log"; out="$(cd "$repo" && PATH="$tmp/bin:$PATH" "$script" "$@" 2>&1)" || rc=$?; }

REVIEW_BASE=origin-develop run run
[ "$rc" -eq 0 ] && grep -q '^--yes @alibaba-group/open-code-review@1.13.8 review --from origin-develop --to HEAD$' "$log" \
  && ok "run reviews HEAD against the base with the pinned version through npx" || ko "run (rc=$rc log=$(cat "$log"))"

REVIEW_BASE=origin-develop run
[ "$rc" -eq 0 ] && grep -q ' review --from origin-develop ' "$log" && ok "run is the default command" || ko "default (rc=$rc)"

REVIEW_BASE=origin-develop run run --effort high
grep -q ' --to HEAD --effort high$' "$log" && ok "extra arguments reach ocr review" || ko "extra args ($(cat "$log"))"

REVIEW_BASE=nope run run
[ "$rc" -eq 64 ] && [ ! -s "$log" ] && ok "a missing base is a usage error and the tool is not started" || ko "missing base (rc=$rc)"

OCR_VERSION=9.9.9 REVIEW_BASE=origin-develop run run
grep -q 'open-code-review@9.9.9 ' "$log" && ok "OCR_VERSION overrides the pin" || ko "version override"

run login
[ "$rc" -eq 0 ] && grep -q 'auth login --provider chatgpt' "$log" && grep -q 'config set provider chatgpt' "$log" \
  && ok "login signs in with ChatGPT and selects it" || ko "login ($(cat "$log"))"

run ocr config provider
grep -q 'open-code-review@1.13.8 config provider$' "$log" && ok "ocr passes any command through" || ko "ocr passthrough"

run ocr
[ "$rc" -eq 64 ] && ok "ocr without arguments is a usage error" || ko "ocr no args (rc=$rc)"

run frobnicate
[ "$rc" -eq 64 ] && grep -q usage <<<"$out" && ok "an unknown subcommand is a usage error" || ko "unknown (rc=$rc)"

finish
