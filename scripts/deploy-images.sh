#!/usr/bin/env bash
# Registry side of the deploy (ADR-0006: GHCR, immutable tag by commit SHA, deployment by digest).
#
#   deploy-images.sh push <tag> <local-tag>   tag the locally built 226os-{api,web,migrate}:<local-tag> as
#                                             <namespace>/226os-<name>:<tag>, push them (job needs packages: write) and print
#                                             API_IMAGE=, WEB_IMAGE=, MIGRATE_IMAGE= with the digest of what THIS run pushed
#                                             (the local RepoDigest set by the push, never a re-read of the tag). A tag that
#                                             already exists with other content fails (tags are immutable); the same content is kept.
#   deploy-images.sh resolve <tag>            print the same three lines by registry digest, read from the tag (production only:
#                                             `tested-<tree>`, then verified by scripts/deploy-verify-tested.sh)
#   deploy-images.sh tag <to> <images.env>    add the tag <to> to the three images of the file (API_IMAGE=... by digest); the
#                                             digests are given, no tag is read, nothing is rebuilt
#
# Digests are always read back from the registry, never from local image IDs (scripts/build-images.sh records those).
# Env: IMAGE_NAMESPACE (default ghcr.io/y4nn777; lower case, as registries require). Needs docker with buildx.
set -euo pipefail

usage() { printf 'usage: %s push <tag> <local-tag> | resolve <tag> | tag <to-tag> <images.env>\n' "$0" >&2; exit 64; }
die() { printf 'deploy-images: %s\n' "$*" >&2; exit 1; }

ns="${IMAGE_NAMESPACE:-ghcr.io/y4nn777}"
[[ "$ns" =~ ^[a-z0-9.-]+(:[0-9]+)?(/[a-z0-9._-]+)+$ ]] || { printf 'deploy-images: invalid IMAGE_NAMESPACE (lower case, registry/owner)\n' >&2; exit 64; }
names=(api web migrate)
declare -A refs=()
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
      ref="$ns/226os-$n:$1"
      local_id="$(docker image inspect "226os-$n:$2" --format '{{.Id}}' 2>/dev/null)" || die "no local image 226os-$n:$2"
      [[ "$local_id" =~ ^sha256:[0-9a-f]{64}$ ]] || die "unexpected local image id for 226os-$n:$2"
      if docker buildx imagetools inspect "$ref" >/dev/null 2>&1; then
        # Immutable tags (ADR-0006): a tag that exists is never overwritten and never trusted. It is accepted only when
        # it holds exactly the image built here (same config digest); anything else was pushed by someone else, or by
        # another build of this commit, and fails the build.
        remote_cfg="$(docker buildx imagetools inspect --raw "$ref" 2>/dev/null | jq -r '.config.digest // empty')" || remote_cfg=""
        [ "$remote_cfg" = "$local_id" ] || die "$ref already exists with other content (tags are immutable): refusing to deploy an image this run did not build; push a new commit"
        printf 'deploy-images: %s is already published with this exact content, kept\n' "$ref" >&2
        d="$(digest_of "$1" "$n")"
      else
        docker tag "226os-$n:$2" "$ref"
        docker push "$ref" >&2
        d="$(docker image inspect "$ref" --format '{{range .RepoDigests}}{{println .}}{{end}}' 2>/dev/null | grep -F -- "$ns/226os-$n@" | head -n 1 | sed 's/^.*@//')" || d=""
        [[ "$d" =~ ^sha256:[0-9a-f]{64}$ ]] || die "the push of $ref left no registry digest locally"
      fi
      printf '%s_IMAGE=%s/226os-%s@%s\n' "$(tr '[:lower:]' '[:upper:]' <<<"$n")" "$ns" "$n" "$d"
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
    [ -r "$2" ] || die "cannot read $2"
    # Every image must be given by digest and belong to this namespace under its own name: nothing is read from a tag.
    for n in "${names[@]}"; do
      key="$(tr '[:lower:]' '[:upper:]' <<<"$n")_IMAGE"
      ref="$(sed -n "s/^${key}=//p" "$2" | head -n 1)"
      [[ "$ref" == "$ns/226os-$n@sha256:"* && "${ref#*@}" =~ ^sha256:[0-9a-f]{64}$ ]] || die "$key is missing or is not $ns/226os-$n@sha256:<digest>"
      refs[$n]="$ref"
    done
    for n in "${names[@]}"; do
      docker buildx imagetools create --tag "$ns/226os-$n:$1" "${refs[$n]}"
    done
    ;;
  *) usage ;;
esac
