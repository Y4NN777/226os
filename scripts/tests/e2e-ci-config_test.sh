#!/usr/bin/env bash
# The E2E report is uploaded from a public repository: in CI the Playwright config must keep trace, video and
# screenshot off and use no html reporter (a trace holds the E2E password, the sign-in body and Set-Cookie), and
# the deploy workflow must upload only the junit XML.
# Usage: scripts/tests/e2e-ci-config_test.sh
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
cfg="$root/src/web/playwright.config.ts"
wf="$root/.github/workflows/deploy.yml"

grep -q "trace: isCI ? 'off'" "$cfg" && ok "trace is off in CI" || ko "trace not off in CI"
grep -q "video: 'off'" "$cfg" && grep -q "screenshot: 'off'" "$cfg" && ok "video and screenshot are off" || ko "video/screenshot"
grep -q "isCI$" <(grep -A1 'reporter:' "$cfg") && grep -q "\['junit'" "$cfg" && ok "CI reporters are list and junit" || ko "reporters"
! grep -E "trace: *'(on|retain|on-first-retry)'" "$cfg" | grep -vq "isCI" && ok "no unconditional trace" || ko "unconditional trace"
grep -q 'test-results/junit.xml' "$wf" && ! grep -q 'playwright-report' "$wf" && ok "the workflow uploads only the junit XML" || ko "workflow upload"
! grep -q 'no secrets in it' "$wf" && ok "no 'no secrets' claim" || ko "claim"
finish
