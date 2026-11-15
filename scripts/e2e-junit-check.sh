#!/usr/bin/env bash
# The junit XML of the E2E job is uploaded as an artifact of a public repository: it must not hold the E2E account.
# Compares the file, raw and XML-entity-decoded, with E2E_USER and E2E_PASSWORD (from the step env) as plain strings, never prints a
# value. A report that holds one of them fails with an ::error:: line and is DELETED, so that no later step uploads it.
# A missing report is not an error (the suite may have stopped before writing it): there is nothing to upload.
# Fails closed on an empty value (an empty string would match every file).
#
# Usage: E2E_USER=... E2E_PASSWORD=... scripts/e2e-junit-check.sh <junit.xml>
set -euo pipefail

[ $# -eq 1 ] || { printf 'usage: %s <junit.xml>\n' "$0" >&2; exit 64; }
file="$1"
for k in E2E_USER E2E_PASSWORD; do
  [ -n "${!k-}" ] || { printf '::error title=E2E account not available::%s is empty, so the report cannot be checked.\n' "$k"; exit 2; }
done
if [ ! -e "$file" ]; then
  printf 'e2e-junit-check: no junit report at %s, nothing to upload\n' "$file"
  exit 0
fi

content="$(cat -- "$file"; printf x)"
content="${content%x}"
# The XML may hold a value escaped (&amp; &lt; &#38; &#x26; ...): compare the entity-decoded text as well.
decoded="$(python3 -I -c 'import html, sys; sys.stdout.write(html.unescape(sys.stdin.read()))' <"$file"; printf x)"
decoded="${decoded%x}"
for k in E2E_USER E2E_PASSWORD; do
  # A test runner may also print the value JS/JSON-escaped (\" and \\) in a failure message.
  escaped="$(python3 -I -c 'import json, os, sys; sys.stdout.write(json.dumps(os.environ[sys.argv[1]])[1:-1])' "$k"; printf x)"
  escaped="${escaped%x}"
  if [[ "$content" == *"${!k}"* || "$decoded" == *"${!k}"* || "$decoded" == *"$escaped"* ]]; then
    rm -f -- "$file"
    printf '::error title=E2E account in the junit report::The report holds the value of %s. It was deleted and is not uploaded.\n' "$k"
    exit 1
  fi
done
printf 'e2e-junit-check: the report holds neither E2E_USER nor E2E_PASSWORD\n'
