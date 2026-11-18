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

# Every printing of a value a test runner or a log may use is compared, as plain strings (never a pattern): the raw
# file, its XML-entity-decoded text and the URL-decoded form of both; for each value the value itself, its JSON
# escapes (non-ASCII kept literal as JSON.stringify prints it, or written \uXXXX in any hex case), the JSON "\/"
# escape, and a quote escaped while the backslash stays raw. Prints the name of the first value found, never a value.
rc=0
found="$(python3 -I - "$file" <<'PY'
import html, json, os, re, sys
from urllib.parse import unquote
BSL = chr(92)
raw = open(sys.argv[1], "rb").read().decode("utf-8", "replace")
dec = html.unescape(raw)
texts = [raw, dec, unquote(raw), unquote(dec)]
hexu = re.compile(re.escape(BSL) + "u([0-9A-Fa-f]{4})")
texts += [hexu.sub(lambda m: BSL + "u" + m.group(1).lower(), t) for t in texts]
for key in ("E2E_USER", "E2E_PASSWORD"):
    v = os.environ[key]
    forms = {v, v.replace('"', BSL + '"')}
    for ascii_only in (False, True):
        j = json.dumps(v, ensure_ascii=ascii_only)[1:-1]
        forms |= {j, j.replace("/", BSL + "/")}
    if any(f and f in t for f in forms for t in texts):
        print(key)
        sys.exit(1)
PY
)" || rc=$?
if [ "$rc" -ne 0 ]; then
  rm -f -- "$file"
  if [ "$rc" -eq 1 ] && [ -n "$found" ]; then
    printf '::error title=E2E account in the junit report::The report holds the value of %s. It was deleted and is not uploaded.\n' "$found"
  else
    printf '::error title=E2E report check failed::The report could not be checked (exit %s). It was deleted and is not uploaded.\n' "$rc"
  fi
  exit 1
fi
printf 'e2e-junit-check: the report holds neither E2E_USER nor E2E_PASSWORD\n'
