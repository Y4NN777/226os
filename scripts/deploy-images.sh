#!/usr/bin/env bash
# Registry side of the deploy (ADR-0006: GHCR, immutable tag by commit SHA, deployment by digest).
#
#   deploy-images.sh push <tag> <local-tag>   tag the locally built 226os-{api,web,migrate}:<local-tag> as
#                                             <namespace>/226os-<name>:<tag> and push them (job needs packages: write)
#   deploy-images.sh resolve <tag>            print API_IMAGE=, WEB_IMAGE= and MIGRATE_IMAGE= by registry digest
#   deploy-images.sh tag <from> <to>          add the tag <to> to the digests that <from> points to (no rebuild)
#
# Digests are always read back from the registry, never from local image IDs (scripts/build-images.sh records those).
# Env: IMAGE_NAMESPACE (default ghcr.io/y4nn777; lower case, as registries require). Needs docker with buildx.
set -euo pipefail

usage() { printf 'usage: %s push <tag> <local-tag> | resolve <tag> | tag <from-tag> <to-tag>\n' "$0" >&2; exit 64; }
die() { printf 'deploy-images: %s\n' "$*" >&2; exit 1; }

ns="${IMAGE_NAMESPACE:-ghcr.io/y4nn777}"
[[ "$ns" =~ ^[a-z0-9.-]+(:[0-9]+)?(/[a-z0-9._-]+)+$ ]] || { printf 'deploy-images: invalid IMAGE_NAMESPACE (lower case, registry/owner)\n' >&2; exit 64; }
names=(api web migrate)
tag_re='^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$'
check_tag() { [[ "$1" =~ $tag_re ]] || { printf 'deploy-images: invalid tag\n' >&2; exit 64; }; }

# digest_of TAG NAME: the registry digest of <ns>/226os-<name>:<tag>.
digest_of() {
  local ref="$ns/226os-$2:$1" d
  d="$(docker buildx imagetools inspect "$ref" --format '{{.Manifest.Digest}}' 2>/dev/null)" ||
    die "no image $ref in the registry (is the tag pushed, and can this job read the package?)"
  [[ "$d" =~ ^sha256:[0-9a-f]{64}$ ]] || die "the registry answered an unexpected digest for $ref"
  printf '%s' "$d"
}

[ $# -ge 1 ] || usage
cmd="$1"
shift
case "$cmd" in
  push)
    [ $# -eq 2 ] || usage
    check_tag "$1"
    check_tag "$2"
    for n in "${names[@]}"; do
      docker tag "226os-$n:$2" "$ns/226os-$n:$1"
      docker push "$ns/226os-$n:$1"
    done
    ;;
  resolve)
    [ $# -eq 1 ] || usage
    check_tag "$1"
    for n in "${names[@]}"; do
      d="$(digest_of "$1" "$n")"
      printf '%s_IMAGE=%s/226os-%s@%s\n' "$(tr '[:lower:]' '[:upper:]' <<<"$n")" "$ns" "$n" "$d"
    done
    ;;
  tag)
    [ $# -eq 2 ] || usage
    check_tag "$1"
    check_tag "$2"
    for n in "${names[@]}"; do
      d="$(digest_of "$1" "$n")"
      docker buildx imagetools create --tag "$ns/226os-$n:$2" "$ns/226os-$n@$d"
    done
    ;;
  *) usage ;;
esac
