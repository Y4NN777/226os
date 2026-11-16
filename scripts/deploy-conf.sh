#!/usr/bin/env bash
# The non-secret deploy settings of one environment, read from a versioned file (deploy/<target>.conf) and validated.
#
# Usage: scripts/deploy-conf.sh <staging | production> [file]      (file defaults to deploy/<target>.conf of this checkout)
# Output (stdout, only when everything is valid): the four lines KEY=value, safe to append to $GITHUB_ENV.
# Exit 2 (REFUSED, nothing on stdout) on anything else.
#
# The file is data, never code: it is read line by line and never evaluated or sourced. Allowed lines: blank, a
# `# comment`, and exactly one `KEY=value` for each of DEPLOY_HOST, DEPLOY_USER, PUBLIC_HOST, DEPLOY_KNOWN_HOSTS (the
# PUBLIC host key line of the server, `<DEPLOY_HOST> ssh-ed25519 <base64>`). No quotes, no spaces around `=`, no unknown
# key, no duplicate key. Every value is checked with a strict regex (the same ones deploy.yml used for the GitHub
# variables), so no shell metacharacter, newline or quote can get through. A value holding FILL-ME is a placeholder:
# the deploy stops with "fill deploy/<target>.conf". Errors name the key, never the value.
#
# Trust: the workflow checks out the commit it deploys (develop for staging, main for production) and reads the file
# there, never from a pull request.
set -euo pipefail

die() { printf 'deploy-conf: REFUSED: %s\n' "$*" >&2; exit 2; }

[ $# -ge 1 ] && [ $# -le 2 ] || { printf 'usage: %s <staging | production> [file]\n' "$0" >&2; exit 64; }
target="$1"
case "$target" in staging | production) ;; *) printf 'usage: %s <staging | production> [file]\n' "$0" >&2; exit 64 ;; esac
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
file="${2:-$root/deploy/$target.conf}"
rel="deploy/$target.conf"
[ -f "$file" ] || die "$rel does not exist"

keys=(DEPLOY_HOST DEPLOY_USER PUBLIC_HOST DEPLOY_KNOWN_HOSTS)
declare -A val=()
lineno=0
line_re='^([A-Z][A-Z_]*)=(.*)$'
while IFS= read -r line || [ -n "$line" ]; do
  lineno=$((lineno + 1))
  case "$line" in '' | '#'*) continue ;; esac
  [[ "$line" =~ $line_re ]] || die "$rel:$lineno is not KEY=value (comments start at the first column with #)"
  k="${BASH_REMATCH[1]}"
  v="${BASH_REMATCH[2]}"
  case " ${keys[*]} " in *" $k "*) ;; *) die "$rel:$lineno unknown key $k (allowed: ${keys[*]})" ;; esac
  [ -z "${val[$k]+set}" ] || die "$rel:$lineno $k is set twice"
  val[$k]="$v"
done <"$file"

for k in "${keys[@]}"; do
  [ -n "${val[$k]+set}" ] || die "$rel has no $k"
  [ -n "${val[$k]}" ] || die "$k is empty in $rel"
done
for k in "${keys[@]}"; do
  case "${val[$k]}" in *FILL-ME*) die "$k is still a placeholder: fill $rel (see docs/runbooks/deploy-setup.md)" ;; esac
done

host_re='^[A-Za-z0-9.:-]+$'
user_re='^[a-z_][a-z0-9_-]*$'
public_re='^([a-z0-9]([a-z0-9-]*[a-z0-9])?\.)+[a-z]+$'
known_re='^([A-Za-z0-9.:-]+) ssh-ed25519 [A-Za-z0-9+/]+=*$'
[[ "${val[DEPLOY_HOST]}" =~ $host_re ]] || die "DEPLOY_HOST is not a host name or an IP address ($rel)"
[[ "${val[DEPLOY_USER]}" =~ $user_re ]] || die "DEPLOY_USER is not a user name ($rel)"
[[ "${val[PUBLIC_HOST]}" =~ $public_re ]] || die "PUBLIC_HOST must be a lower-case DNS name ($rel)"
[[ "${val[DEPLOY_KNOWN_HOSTS]}" =~ $known_re ]] || die "DEPLOY_KNOWN_HOSTS must be one line '<DEPLOY_HOST> ssh-ed25519 <public key>' ($rel)"
[ "${BASH_REMATCH[1]}" = "${val[DEPLOY_HOST]}" ] || die "DEPLOY_KNOWN_HOSTS must name the same host as DEPLOY_HOST ($rel)"

for k in "${keys[@]}"; do
  printf '%s=%s\n' "$k" "${val[$k]}"
done
