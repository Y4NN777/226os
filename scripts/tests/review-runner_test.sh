#!/usr/bin/env bash
# Static test of the review workflows' runner needs (first live auto-fix run, PR #199, run 43625357930):
#  1. every job that runs anthropics/claude-code-action installs bubblewrap and socat (apt, no other package) in a step
#     BEFORE the action: Claude Code's Linux Bash sandbox needs them, the fixer cannot test or commit without them;
#  2. every job that writes a label on a pull request (labels API, or a trusted script that does) has
#     `pull-requests: write` (`issues: write` alone answers HTTP 403 on a pull request), and no job that checks out
#     pull request code (a ref other than the default branch) has it.
# Usage: scripts/tests/review-runner_test.sh [workflows-dir]
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
dir="${1:-$root/.github/workflows}"

# One line per finding: "OK <text>" or "KO <text>".
report="$(python3 -I - "$dir" <<'EOF'
import glob, os, re, sys, yaml

CLAUDE = "anthropics/claude-code-action"
# Jobs that label through a trusted script (no label call in the YAML): the script writes with the job's GITHUB_TOKEN.
SCRIPT_LABEL = re.compile(r"review-autofix\.py[\"']?\s+pr\b")
LABEL_API = re.compile(r"/labels\b|gh\s+(issue|pr)\s+edit\b.*--(add|remove)-label|gh\s+label\b")

def steps_of(job):
    return job.get("steps") or []

def runs_text(step):
    return str(step.get("run") or "")

def checks_out_pr_code(job):
    # A checkout with a `ref` is the pull request (or planned) commit; one without is the default branch.
    for step in steps_of(job):
        if str(step.get("uses") or "").startswith("actions/checkout@") and (step.get("with") or {}).get("ref"):
            return True
    return False

for path in sorted(glob.glob(os.path.join(sys.argv[1], "review-*.yml"))):
    wf = os.path.basename(path)
    doc = yaml.safe_load(open(path))
    for name, job in (doc.get("jobs") or {}).items():
        label = "%s job %s" % (wf, name)
        steps = steps_of(job)
        perms = job.get("permissions") or {}
        # 1. the sandbox dependencies before the fixer
        idx = [i for i, s in enumerate(steps) if str(s.get("uses") or "").startswith(CLAUDE + "@")]
        if idx:
            before = "\n".join(runs_text(s) for s in steps[: idx[0]])
            installs = re.search(r"apt-get\s+install\b[^\n]*\bbubblewrap\b[^\n]*\bsocat\b", before) or \
                re.search(r"apt-get\s+install\b[^\n]*\bsocat\b[^\n]*\bbubblewrap\b", before)
            print(("OK " if installs else "KO ") + "%s installs bubblewrap and socat before the fixer" % label)
        # 2. label writers have pull-requests: write
        writes_labels = any(LABEL_API.search(runs_text(s)) or SCRIPT_LABEL.search(runs_text(s)) for s in steps)
        has_write = perms.get("pull-requests") == "write"
        if writes_labels:
            print(("OK " if has_write else "KO ") + "%s labels pull requests and has pull-requests: write" % label)
        if has_write:
            print(("KO " if checks_out_pr_code(job) else "OK ") + "%s has pull-requests: write and checks out no pull request code" % label)
            print(("OK " if writes_labels else "KO ") + "%s has pull-requests: write only because it labels" % label)
EOF
)"

n=0
while IFS= read -r line; do
  [ -n "$line" ] || continue
  n=$((n + 1))
  case "$line" in
    "OK "*) ok "${line#OK }" ;;
    *) ko "${line#KO }" ;;
  esac
done <<<"$report"

[ "$n" -ge 4 ] && ok "the checks found their jobs ($n assertions)" || ko "the checks found too few jobs ($n assertions)"
finish
