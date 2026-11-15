#!/usr/bin/env bash
# Production side of the "tested on staging" proof. The mutable tag `tested-<tree>` is NOT trusted by itself: any
# workflow of the repository with packages: write could re-point it. So every digest that production is about to
# deploy must carry a signed attestation, made by the `promote` job of THIS deploy workflow on develop (Sigstore, via
# actions/attest), whose predicate says that exact tree passed staging.
#
#   deploy-verify-tested.sh <tree> <images.env>
#
# <images.env> is the output of `scripts/deploy-images.sh resolve`: API_IMAGE=, WEB_IMAGE=, MIGRATE_IMAGE= by digest.
# Each image is checked with `gh attestation verify` (needs GH_TOKEN and a registry login, both set by the caller):
# signer workflow, repository, source ref and predicate type are pinned; then at least one verified attestation must
# carry predicate.tree == <tree>. Anything else (no attestation, another tree, a gh error, unreadable output) exits 1
# with an ::error:: line: fail closed, before the production approval is requested.
# Env: ATTEST_REPO (default Y4NN777/226os), only for a fork's own pipeline.
set -euo pipefail

predicate_type='https://226os.tech/attestations/tested-on-staging/v1'
repo="${ATTEST_REPO:-Y4NN777/226os}"
signer="$repo/.github/workflows/deploy.yml"
source_ref=refs/heads/develop

[ $# -eq 2 ] || { printf 'usage: %s <tree> <images.env>\n' "$0" >&2; exit 64; }
tree="$1"
file="$2"
[[ "$tree" =~ ^[0-9a-f]{40,64}$ ]] || { printf 'deploy-verify-tested: invalid tree\n' >&2; exit 64; }
[ -r "$file" ] || { printf 'deploy-verify-tested: cannot read %s\n' "$file" >&2; exit 64; }
[[ "$repo" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]] || { printf 'deploy-verify-tested: invalid ATTEST_REPO\n' >&2; exit 64; }

refuse() { printf '::error title=Tested image not proven::%s\n' "$*"; exit 1; }

ref_re='^[a-z0-9.-]+(:[0-9]+)?(/[a-z0-9._-]+)+@sha256:[0-9a-f]{64}$'
for var in API_IMAGE WEB_IMAGE MIGRATE_IMAGE; do
  ref="$(sed -n "s/^${var}=//p" "$file" | head -n 1)"
  [[ "$ref" =~ $ref_re ]] || refuse "${var} is missing or is not an image reference by digest, so it cannot be verified."
  if ! json="$(gh attestation verify "oci://$ref" --repo "$repo" --signer-workflow "$signer" \
    --source-ref "$source_ref" --predicate-type "$predicate_type" --format json 2>"${RUNNER_TEMP:-${TMPDIR:-/tmp}}/attest.err")"; then
    refuse "${ref} has no valid attestation from the deploy workflow on develop (the tag tested-${tree} alone is not proof: it can be re-pointed). Staging must promote this tree again."
  fi
  if ! jq -e --arg tree "$tree" --arg pt "$predicate_type" \
    'type == "array" and any(.[]; .verificationResult.statement.predicateType == $pt and .verificationResult.statement.predicate.tree == $tree)' \
    >/dev/null 2>&1 <<<"$json"; then
    refuse "${ref} is attested, but not as tested on staging for tree ${tree}. Refusing to deploy a digest proven for another tree."
  fi
  printf 'deploy-verify-tested: %s attested as tested for tree %s\n' "$ref" "$tree"
done
