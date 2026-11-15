#!/usr/bin/env bash
# Static checks of .github/workflows/deploy-emergency.yml, the emergency production deploy that skips the staging
# gate: manual dispatch only, main only, explicit confirmation and reason (through env, never interpolated in a
# run:), production environment (approval), concurrency shared with deploy.yml's production group, own build and
# scan, never seeds, never tags or attests tested-*, and the job holding packages: write holds no secret.
# Usage: scripts/tests/deploy-emergency-workflow_test.sh
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
wf="$root/.github/workflows/deploy-emergency.yml"
main="$root/.github/workflows/deploy.yml"

[ -f "$wf" ] || { ko "deploy-emergency.yml exists"; finish; }

# check <python expression over w (the workflow) and m (deploy.yml)> <description>
check() {
  if python3 -I - "$wf" "$main" "$1" <<'EOF'
import sys, yaml
w = yaml.safe_load(open(sys.argv[1])); m = yaml.safe_load(open(sys.argv[2]))
on = w.get(True, w.get("on"))   # PyYAML reads the key `on` as True
jobs = w["jobs"]
runs = [s.get("run", "") for j in jobs.values() for s in j.get("steps", [])]
def secret_refs(job): return "secrets." in yaml.safe_dump(job)
sys.exit(0 if eval(sys.argv[3]) else 1)
EOF
  then ok "$2"; else ko "$2"; fi
}

check 'list(on) == ["workflow_dispatch"]' "the only trigger is workflow_dispatch"
check 'on["workflow_dispatch"]["inputs"]["reason"]["required"] is True' "a reason input is required"
check 'any(i.get("required") is True and k != "reason" for k, i in on["workflow_dispatch"]["inputs"].items())' "a confirmation input is required"
check 'not any("inputs." in r or "github.event.inputs" in r for r in runs)' "no input is interpolated in a run: (env only)"
check 'any("refs/heads/main" in yaml.safe_dump(j) or "scripts/deploy-plan.sh" in yaml.safe_dump(j) for j in jobs.values())' "the plan refuses any ref other than main (deploy-plan.sh)"
check 'any(j.get("environment") == "production" for j in jobs.values())' "the deploy job uses the production environment (approval)"
check 'w["concurrency"]["group"] == "deploy-production" and w["concurrency"]["cancel-in-progress"] is False' "concurrency deploy-production, never cancelled"
check '"deploy-production" in str(m["concurrency"]["group"]) or "production" in str(m["concurrency"]["group"])' "deploy.yml has a production concurrency group to share"
check 'any("scripts/build-images.sh" in r for r in runs) and any("scripts/image-scan.sh" in r for r in runs)' "it builds and scans its own images"
check 'all(not secret_refs(j) for j in jobs.values() if (j.get("permissions") or {}).get("packages") == "write")' "the job with packages: write holds no secret"
check 'not any("seed" in r for r in runs) and "db/seeds" not in open(sys.argv[1]).read()' "it never seeds"
check 'not any(("deploy-images.sh tag" in r) or ("attest" in r) for r in runs)' "it never tags or attests tested-*"
check 'not any("deploy-verify-tested" in r for r in runs)' "it skips the tested-<tree> verification"
check 'any("::warning" in r and "EMERGENCY" in r for r in runs) and any("EMERGENCY: not tested on staging" in r for r in runs)' "it warns and writes the EMERGENCY summary line"
check 'any("deploy-remote.sh production-live" in r for r in runs)' "it deploys with deploy-remote.sh production-live (backup, migrate, up, smoke)"
check 'not any("pull_request" in str(k) for k in on)' "no pull_request trigger"
finish
