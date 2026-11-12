#!/usr/bin/env bash
# Reproducible OpenAPI code generation and drift check (ADR-0004 point 10).
#
# Usage: scripts/check-generated.sh [check|generate]      (default: check)
#   generate  regenerate the committed outputs from api/openapi.yaml, nothing else.
#   check     generate, then fail when a generated file differs from the index
#             (git diff) or is not tracked (untracked or missing from the index).
#             In CI the index equals HEAD, so this is the "generate && git diff
#             --exit-code" gate. Locally, `git add` the generated files first.
#
# Generators (both pinned): oapi-codegen through `go tool` (go.mod `tool` line) and
# openapi-typescript through `pnpm exec` (pnpm-lock.yaml). No npx, no floating
# versions. Exit 0 = in sync, 1 = drift or generator failure.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

spec=api/openapi.yaml
go_config=api/oapi-codegen.yaml
generated=(
  apps/api/internal/platform/apigen/apigen.gen.go
  apps/web/src/api/schema.d.ts
)

generate() {
  go tool oapi-codegen -config "$go_config" "$spec"
  mkdir -p apps/web/src/api
  pnpm exec openapi-typescript "$spec" -o apps/web/src/api/schema.d.ts >/dev/null
}

mode="${1:-check}"
case "$mode" in
  generate) generate; exit 0 ;;
  check) ;;
  *) printf 'usage: %s [check|generate]\n' "$0" >&2; exit 2 ;;
esac

generate

fail=0
for f in "${generated[@]}"; do
  if [ ! -f "$f" ]; then
    printf 'check-generated: missing generated file: %s\n' "$f" >&2
    fail=1
  elif ! git ls-files --error-unmatch -- "$f" >/dev/null 2>&1; then
    printf 'check-generated: generated file is not tracked (git add it): %s\n' "$f" >&2
    fail=1
  fi
done

if ! git diff --exit-code --stat -- "${generated[@]}"; then
  printf 'check-generated: generated files differ from the index; commit the regenerated output\n' >&2
  fail=1
fi

untracked="$(git ls-files --others --exclude-standard -- "${generated[@]}")"
if [ -n "$untracked" ]; then
  printf 'check-generated: untracked generated files:\n%s\n' "$untracked" >&2
  fail=1
fi

[ "$fail" -eq 0 ] && printf 'check-generated: OK (generated files in sync with %s)\n' "$spec"
exit "$fail"
