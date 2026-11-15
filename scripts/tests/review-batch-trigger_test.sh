#!/usr/bin/env bash
# Static checks of the triggers of .github/workflows/review-batch.yml (ADR-0015 phase 0, issue #264): the `0 */2 * * *` cron
# fired once as `schedule` on 2026-11-15 and every other run was a manual dispatch (GitHub schedules are best effort), so the
# pipeline also starts the batch itself, after a merge into develop has been settled: `workflow_run` of "Review merged".
# The cron stays as the fallback. Stdlib only, like the other workflow tests: the file is read as text, job by job.
# Usage: scripts/tests/review-batch-trigger_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
wf="$root/.github/workflows/review-batch.yml"

# job NAME: the text of one job of the workflow (from `  NAME:` to the next job at the same indent).
job() { python3 -I - "$wf" "$1" <<'PY'
import re, sys
s = open(sys.argv[1]).read()
m = re.search(r"^  %s:\n(.*?)(?=^  [\w-]+:\n|\Z)" % re.escape(sys.argv[2]), s, re.S | re.M)
sys.exit("no job %s" % sys.argv[2]) if not m else print(m.group(1))
PY
}
on_block="$(python3 -I -c 'import re,sys; s=open(sys.argv[1]).read(); print(re.search(r"^on:\n(.*?)^permissions:", s, re.S | re.M).group(1))' "$wf")"

grep -qE '^\s+workflows: \["Review merged"\]$' <<<"$on_block" && grep -qE '^\s+types: \[completed\]$' <<<"$on_block" && grep -q '^  workflow_run:' <<<"$on_block" && ok "the batch starts after the \"Review merged\" workflow has settled a merge (workflow_run, completed)" || ko "no workflow_run trigger"
grep -q '^name: Review merged$' "$root/.github/workflows/review-merged.yml" && ok "the workflow named in the trigger exists" || ko "Review merged workflow name"
grep -qE '^\s+- cron: "0 \*/2 \* \* \*"$' <<<"$on_block" && grep -q '^  workflow_dispatch:' <<<"$on_block" && ok "the cron stays as the fallback, and the manual dispatch too" || ko "cron or dispatch lost"

for name in batch plan; do
  j="$(job "$name")"
  head -n 12 <<<"$j" | grep -q "github.event_name == 'workflow_run'" && ok "job $name also runs on workflow_run" || ko "job $name does not run on workflow_run"
done
# "Review merged" does not expose whether the pull request was merged (the workflow_run payload has no such field), so the filter
# is on the conclusion: its settle job is skipped for a pull request closed without a merge, which is not a success or a failure.
for name in batch plan; do
  grep -q "github.event.workflow_run.conclusion == 'success' || github.event.workflow_run.conclusion == 'failure'" <<<"$(job "$name")" && grep -q "github.event.workflow_run.event == 'pull_request'" <<<"$(job "$name")" && grep -q "head_repository.full_name == github.repository" <<<"$(job "$name")" && ok "$name: an explicit workflow_run filter (pull request of this repository, success or failure of Review merged, never skipped or cancelled)" || ko "$name workflow_run filter"
done
! grep -qi 'fired once' "$wf" && ok "no unsourced claim about how often the cron fired (that stays in TODO.md)" || ko "fired once claim"
grep -q "FORCE: \${{ github.event_name == 'workflow_dispatch' }}" "$wf" && ok "only a manual dispatch forces a fixer run: a merge never does" || ko "FORCE"
! grep -qE '^\s+actions: write|^\s+contents: write|^\s+id-token' "$wf" && ok "no new permission: the trigger needs no actions: write" || ko "permissions grew"
grep -qE '^    concurrency:\n?' <<<"$(job batch)" && grep -q 'group: review-batch-write' <<<"$(job batch)" && grep -q 'cancel-in-progress: false' <<<"$(job batch)" && ok "batch: one writer at a time, never cancelled half way (a burst of merges queues one run)" || ko "batch concurrency"
grep -q 'group: review-batch-fix' <<<"$(job fix)" && ok "fix: one fixer at a time" || ko "fix concurrency"
for name in batch plan; do
  python3 -I - "$name" "$(job "$name")" <<'PY' && ok "job $name: no \${{ }} inside a run: block (values go through env)" || ko "job $name interpolates in run:"
import re, sys
for run in re.findall(r"^\s+run: \|\n((?:\s{10,}.*\n|\n)+)", sys.argv[2] + "\n", re.M):
    assert "${{" not in run, run
PY
done
grep -oE 'uses: [^ ]+' "$wf" | grep -vqE '@[0-9a-f]{40}$' && ko "an action is not pinned to a SHA" || ok "every action is pinned to a full commit SHA"

finish
