#!/usr/bin/env bash
# capture-fidelity: screenshots of app pages next to their HTML prototype, for a visual comparison.
# A developer tool, NOT a CI gate. It changes nothing in the app or the prototypes.
#
# Usage:
#   src/web/scripts/capture-fidelity.sh [options] ROUTE=PAGE [ROUTE=PAGE ...]
#     ROUTE  an app path, e.g. /discover or /solve?sort=new
#     PAGE   a file in prototypes/html/pages, e.g. discover.html
#   Example:
#     src/web/scripts/capture-fidelity.sh -t "light dark" /=accueil.html /discover=discover.html
#
# Options (each also has an environment variable):
#   -b URL    app base URL           CAPTURE_BASE_URL  default http://127.0.0.1:8088 (local-dev stack)
#             Use the Vite dev server of your worktree to capture uncommitted work.
#   -o DIR    output folder          CAPTURE_OUT       default ./fidelity-shots
#   -t LIST   themes                 CAPTURE_THEMES    default "light"   (any of: light dark)
#   -s LIST   sizes WxH              CAPTURE_SIZES     default "1280x800 375x812"
#   -f        full-page screenshots  CAPTURE_FULL=1    default: the viewport only
#   -w MS     settle time per page   CAPTURE_SETTLE_MS default 1200 (prototype lists load after a fake delay)
#
# Output: DIR/<slug>-<app|proto>-<width>-<theme>.png, slug = the route made file-safe ("/" is "home").
#
# Browser: Playwright is not a dependency of this repository (only browser binaries may sit in
# ~/.cache/ms-playwright), so this script drives the gstack `browse` binary instead:
#   BROWSE_BIN (default $HOME/.claude/skills/gstack/browse/dist/browse), in its own tab named by
#   BROWSE_SESSION (default "capture-fidelity").
# Language and theme: the prototypes read ?lang=fr&theme=..., the app does not. For the app the
# script writes the app's own preferences in localStorage (226os.locale=fr, 226os.theme=<theme>)
# on the app origin, then loads the route. The prototype is loaded with ?session=out so both
# show the signed-out header (the app is captured without a session).
set -euo pipefail

base_url="${CAPTURE_BASE_URL:-http://127.0.0.1:8088}"
out_dir="${CAPTURE_OUT:-./fidelity-shots}"
themes="${CAPTURE_THEMES:-light}"
sizes="${CAPTURE_SIZES:-1280x800 375x812}"
full="${CAPTURE_FULL:-0}"
settle_ms="${CAPTURE_SETTLE_MS:-1200}"

usage() { sed -n '2,21p' "$0" | sed 's/^# \{0,1\}//'; }

while getopts 'b:o:t:s:fw:h' opt; do
  case "$opt" in
    b) base_url="$OPTARG" ;;
    o) out_dir="$OPTARG" ;;
    t) themes="$OPTARG" ;;
    s) sizes="$OPTARG" ;;
    f) full=1 ;;
    w) settle_ms="$OPTARG" ;;
    h) usage; exit 0 ;;
    *) usage >&2; exit 2 ;;
  esac
done
shift $((OPTIND - 1))
if [ "$#" -eq 0 ]; then
  usage >&2
  exit 2
fi
case "$settle_ms" in '' | *[!0-9]*) echo "capture-fidelity: -w needs a number of milliseconds" >&2; exit 2 ;; esac

browse="${BROWSE_BIN:-$HOME/.claude/skills/gstack/browse/dist/browse}"
if [ ! -x "$browse" ]; then
  echo "capture-fidelity: browse binary not found at $browse (set BROWSE_BIN)" >&2
  exit 1
fi
export BROWSE_SESSION="${BROWSE_SESSION:-capture-fidelity}"

repo_root="$(cd "$(dirname "$0")/../../.." && pwd)"
pages_dir="$repo_root/prototypes/html/pages"
base_url="${base_url%/}"
mkdir -p "$out_dir"

b() { "$browse" "$@" >/dev/null; }
settle() {
  b wait --networkidle || true
  b js "new Promise((done) => setTimeout(done, $settle_ms))"
}
shoot() {
  if [ "$full" = 1 ]; then b screenshot "$1"; else b screenshot --viewport "$1"; fi
  echo "$1"
}

for pair in "$@"; do
  route="${pair%%=*}"
  page="${pair#*=}"
  if [ "$route" = "$pair" ] || [ -z "$page" ] || [ "${route#/}" = "$route" ]; then
    echo "capture-fidelity: expected ROUTE=PAGE with ROUTE starting with /, got '$pair'" >&2
    exit 2
  fi
  if [ ! -f "$pages_dir/$page" ]; then
    echo "capture-fidelity: no prototype page $pages_dir/$page" >&2
    exit 2
  fi
  slug="$(printf '%s' "$route" | sed -e 's#^/##' -e 's#[^A-Za-z0-9]\{1,\}#-#g' -e 's#-$##')"
  slug="${slug:-home}"
  for size in $sizes; do
    width="${size%x*}"
    b viewport "$size"
    for theme in $themes; do
      case "$theme" in light | dark) ;; *) echo "capture-fidelity: unknown theme '$theme'" >&2; exit 2 ;; esac
      # App: set the app's own preferences on its origin, then load the route.
      b goto "$base_url/"
      b js "localStorage.setItem('226os.locale', 'fr'); localStorage.setItem('226os.theme', '$theme'); true"
      b goto "$base_url$route"
      settle
      shoot "$out_dir/$slug-app-$width-$theme.png"
      # Prototype: the same language, theme and signed-out state through its query string.
      b goto "file://$pages_dir/$page?lang=fr&theme=$theme&session=out"
      settle
      shoot "$out_dir/$slug-proto-$width-$theme.png"
    done
  done
done
