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

# --- the E2E suite runs in its OWN job and environment: npm and Playwright code never meets the deploy secrets ------
# (finding deploy-secrets-reach-e2e). Read with PyYAML (the workflow parsed as data, not grepped).
facts="$(python3 -I - "$wf" <<'PY'
import json, sys, yaml
w = yaml.safe_load(open(sys.argv[1]))
jobs = w["jobs"]
out = {}
for name, job in jobs.items():
    text = yaml.safe_dump(job)
    out[name] = {
        "environment": job.get("environment"),
        "needs": job.get("needs"),
        "permissions": job.get("permissions"),
        "if": job.get("if"),
        "timeout": job.get("timeout-minutes"),
        "outputs": sorted((job.get("outputs") or {}).keys()),
        "text": text,
        "runs": "\n".join(s.get("run", "") for s in job.get("steps", [])),
        "uses": [s.get("uses", "") for s in job.get("steps", [])],
        "checkout": [s.get("with") for s in job.get("steps", []) if s.get("uses", "").startswith("actions/checkout")],
    }
print(json.dumps(out))
PY
)"
q() { jq -r "$1" <<<"$facts"; }
[ "$(q '.e2e.environment')" = staging-e2e ] && ok "the e2e job uses the environment staging-e2e" || ko "e2e environment: $(q '.e2e.environment')"
[ "$(q '.deploy.environment')" != staging-e2e ] && [ "$(q '.deploy.environment | type')" = string ] && ok "the deploy job keeps its own environment" || ko "deploy environment"
for s in DEPLOY_SSH_KEY SOPS_AGE_KEY; do
  [ "$(q '.e2e.text' | grep -c "$s" || true)" -eq 0 ] && ok "the e2e job never mentions $s" || ko "e2e job references $s"
done
[ "$(q '.e2e.text' | grep -c 'secrets\.E2E_USER' || true)" -ge 1 ] && [ "$(q '.e2e.text' | grep -c 'secrets\.E2E_PASSWORD' || true)" -ge 1 ] && ok "the e2e job reads E2E_USER and E2E_PASSWORD from its environment" || ko "e2e secrets"
[ "$(q '.e2e.text' | grep -o 'secrets\.[A-Z0-9_]*' | sort -u | tr '\n' ' ')" = 'secrets.E2E_PASSWORD secrets.E2E_USER ' ] && ok "the e2e job reads exactly those two secrets" || ko "e2e secrets set"
[ "$(q '.e2e.permissions | tojson')" = '{"contents":"read"}' ] && ok "the e2e job has contents: read only" || ko "e2e permissions: $(q '.e2e.permissions | tojson')"
[ "$(q '.e2e.needs | sort | join(",")')" = deploy,plan ] && ok "the e2e job needs plan and deploy" || ko "e2e needs"
q '.e2e.if' | grep -q "needs.plan.outputs.target == 'staging'" && ok "the e2e job runs for staging only" || ko "e2e condition"
[ "$(q '.e2e.checkout | length')" -ge 1 ] && [ "$(q '.e2e.checkout[0]["persist-credentials"]')" = false ] && [ "$(q '.e2e.checkout[0].ref')" = '${{ needs.plan.outputs.sha }}' ] && ok "e2e checks out the deployed SHA with persist-credentials false" || ko "e2e checkout"
[ "$(q '.e2e.outputs | join(",")')" = e2e_ran ] && ok "the e2e job outputs e2e_ran" || ko "e2e outputs"
q '.e2e.runs' | grep -q 'GITHUB_ENV' && ko "the e2e job writes to GITHUB_ENV" || ok "no E2E value goes through GITHUB_ENV"
q '.e2e.runs' | grep -q 'scripts/e2e-env.sh' && ok "the e2e job masks the values (scripts/e2e-env.sh)" || ko "e2e masking"
q '.e2e.text' | grep -q 'NOT RUN' && ok "the e2e job keeps the NOT RUN notice" || ko "NOT RUN notice"
q '.e2e.text' | grep -q 'junit.xml' && ! q '.e2e.text' | grep -q 'playwright-report' && ok "the e2e job uploads only the junit XML" || ko "e2e upload"

# The deploy job holds the SSH key and the sops age key: it runs no package manager and no browser.
if q '.deploy.runs' | grep -qiE '\b(pnpm|npm|npx|node|playwright|yarn)\b' || q '.deploy.uses | join(" ")' | grep -qE 'pnpm/|setup-node|upload-artifact'; then
  ko "the deploy job runs npm, pnpm or Playwright code"
else ok "the deploy job runs no pnpm, npm, npx or Playwright"; fi
q '.deploy.text' | grep -q 'test:e2e' && ko "the deploy job still mentions the E2E script" || ok "the deploy job no longer runs the E2E suite"
[ "$(q '.deploy.timeout')" -le 45 ] && ok "the deploy timeout is back to the deploy budget (45 minutes)" || ko "deploy timeout $(q '.deploy.timeout')"
q '.deploy.outputs | index("base_url")' | grep -q '^[0-9]' && ok "the deploy job still outputs base_url" || ko "base_url output"
q '.deploy.outputs | index("e2e_ran")' | grep -q '^[0-9]' && ko "e2e_ran is still an output of deploy" || ok "e2e_ran is no longer an output of deploy"

# Order and clean-up (review): the account is checked and masked before any npm code; the deploy job has ONE always()
# clean-up of the deploy secrets; the junit report is checked for the account before it can be uploaded.
order="$(python3 -I - "$wf" <<'PY'
import json, sys, yaml
w = yaml.safe_load(open(sys.argv[1]))
def names(j): return [s.get("name", "") for s in w["jobs"][j]["steps"]]
e = names("e2e")
d = w["jobs"]["deploy"]["steps"]
up = [s for s in w["jobs"]["e2e"]["steps"] if s.get("name") == "Upload the junit report"]
print(json.dumps({
  "mask": e.index("Check and mask the E2E account"), "install": e.index("Install from the lockfile"),
  "junit": e.index("Check the junit report for the E2E account"), "upload": e.index("Upload the junit report"),
  "cleanups": sum(1 for s in d if "rm -rf" in s.get("run", "") and "DEPLOY_TMP" in s.get("run", "")),
  "upload_if": up[0].get("if", ""),
}))
PY
)"
oq() { jq -r "$1" <<<"$order"; }
[ "$(oq '.mask < .install')" = true ] && ok "the account is checked and masked before the dependencies are installed" || ko "mask/install order: $order"
[ "$(oq '.cleanups')" -eq 1 ] && ok "the deploy job has a single clean-up of the deploy secrets" || ko "cleanups: $(oq '.cleanups')"
[ "$(oq '.junit < .upload')" = true ] && oq '.upload_if' | grep -q "steps.junit.outcome == 'success'" && ok "the junit report is uploaded only after the account check passed" || ko "junit check/upload: $order"
q '.e2e.runs' | grep -q 'scripts/e2e-junit-check.sh' && ok "the e2e job runs scripts/e2e-junit-check.sh" || ko "junit check not run"

# promote waits for the E2E suite itself.
[ "$(q '.promote.needs | sort | join(",")')" = deploy,e2e,plan,resolve ] && ok "promote needs plan, resolve, deploy and e2e" || ko "promote needs: $(q '.promote.needs | tojson')"
q '.promote.if' | grep -q "needs.e2e.outputs.e2e_ran == 'true'" && ! q '.promote.if' | grep -q 'needs.deploy.outputs.e2e_ran' && ok "promote tags only when the e2e job really ran" || ko "promote condition: $(q '.promote.if')"
finish
