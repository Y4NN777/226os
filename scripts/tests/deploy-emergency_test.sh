#!/usr/bin/env bash
# Static checks of .github/workflows/deploy-emergency.yml, the emergency production deploy that skips the staging
# gate: manual dispatch only, main only, explicit confirmation and reason (through env, never interpolated in a
# run:), production environment (approval), concurrency shared with deploy.yml's production group, own build and
# scan, never seeds, never tags or attests tested-*, and the job holding packages: write holds no secret.
# Usage: scripts/tests/deploy-emergency_test.sh
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
wf="$root/.github/workflows/deploy-emergency.yml"
main="$root/.github/workflows/deploy.yml"
shared="$root/.github/workflows/deploy-production.yml"

[ -f "$wf" ] || { ko "deploy-emergency.yml exists"; finish; }
[ -f "$shared" ] || { ko "deploy-production.yml exists"; finish; }

# check <python expression over w (the workflow) and m (deploy.yml)> <description>
check() {
  if python3 -I - "$wf" "$main" "$1" "$shared" <<'EOF'
import sys, yaml
w = yaml.safe_load(open(sys.argv[1])); m = yaml.safe_load(open(sys.argv[2])); r = yaml.safe_load(open(sys.argv[4]))
on = w.get(True, w.get("on"))   # PyYAML reads the key `on` as True
jobs = w["jobs"]
runs = [s.get("run", "") for j in jobs.values() for s in j.get("steps", [])]
rruns = [s.get("run", "") for j in r["jobs"].values() for s in j.get("steps", [])]
allruns = runs + rruns
def secret_refs(job): return "secrets." in yaml.safe_dump(job)
sys.exit(0 if eval(sys.argv[3]) else 1)
EOF
  then ok "$2"; else ko "$2"; fi
}

check 'list(on) == ["workflow_dispatch"]' "the only trigger is workflow_dispatch"
check 'on["workflow_dispatch"]["inputs"]["reason"]["required"] is True' "a reason input is required"
check 'any(i.get("required") is True and k != "reason" for k, i in on["workflow_dispatch"]["inputs"].items())' "a confirmation input is required"
check 'not any("inputs." in r or "github.event.inputs" in r for r in runs)' "no input is interpolated in a run: (env only)"
check 'list(jobs["plan"]["steps"][1]["env"].items()) and jobs["plan"]["steps"][1]["env"]["DISPATCH_REF"] == "${{ github.ref }}" and jobs["plan"]["steps"][1]["env"]["DISPATCH_SHA"] == "${{ github.sha }}" and "scripts/deploy-plan.sh" in jobs["plan"]["steps"][1]["run"] and "grep -qx \x27emergency=true\x27" in jobs["plan"]["steps"][1]["run"]' "the plan step hands the dispatch ref and SHA to deploy-plan.sh (main only) and requires emergency=true"
check 'set(jobs) == {"plan", "build", "deploy", "record"} and "plan" in jobs["build"]["needs"] and "plan" in jobs["deploy"]["needs"] and jobs["deploy"]["with"]["sha"] == "${{ needs.plan.outputs.sha }}"' "every later job hangs off the plan: no other path decides what is deployed"
check 'jobs["deploy"].get("uses") == "./.github/workflows/deploy-production.yml" and jobs["deploy"].get("secrets") == "inherit"' "the deploy job calls the shared production workflow (environment secrets inherited)"
check 'list(r.get(True, r.get("on"))) == ["workflow_call"] and r["jobs"]["deploy"]["environment"] == "production"' "the shared job runs behind the plain production environment (approval)"
check 'w["concurrency"]["group"] == "deploy-production" and w["concurrency"]["cancel-in-progress"] is False' "concurrency deploy-production, never cancelled"
check '"deploy-production" in str(m["concurrency"]["group"]) or "production" in str(m["concurrency"]["group"])' "deploy.yml has a production concurrency group to share"
check 'any("scripts/build-images.sh" in r for r in runs) and any("scripts/image-scan.sh" in r for r in runs)' "it builds and scans its own images"
check 'all(not secret_refs(j) for j in jobs.values() if (j.get("permissions") or {}).get("packages") == "write")' "the job with packages: write holds no secret"
check 'not any("inputs." in x for x in rruns)' "the shared workflow passes its inputs through env only"
check 'not any("seed" in x for x in runs) and "db/seeds" not in open(sys.argv[1]).read()' "the emergency workflow never seeds"
check 'not any("seed" in x for x in rruns) and "db/seeds" not in open(sys.argv[4]).read()' "the shared production workflow never ships or runs seeds"
check 'not any(("deploy-images.sh tag" in x) or ("attest" in x) for x in allruns)' "it never tags or attests tested-*"
check 'not any("deploy-verify-tested" in x for x in allruns)' "it skips the tested-<tree> verification"
check 'any("::warning" in x and "EMERGENCY" in x for x in runs) and any("EMERGENCY: not tested on staging" in x for x in runs) and any("can replace it" in x for x in runs)' "it warns and writes the EMERGENCY summary line"
check 'any("scripts/deploy-remote.sh production-live" in x for x in rruns) and r["jobs"]["deploy"]["env"]["PROFILE"] == "production-live"' "the shared job deploys with deploy-remote.sh production-live (backup, migrate, up, smoke)"
check 'not any("pull_request" in str(k) for k in on)' "no pull_request trigger"
finish
