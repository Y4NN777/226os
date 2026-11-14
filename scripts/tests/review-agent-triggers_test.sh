#!/usr/bin/env bash
# Static guard (ADR-0013): no job that declares `environment: review-agent` (string, mapping, quoted or flow form) may be
# startable by pull_request, pull_request_review or pull_request_target (check_dir says what counts as guarded).
# Those events run code or text a pull request controls; the environment holds REVIEW_AGENT_TOKEN. The checks below run on temporary fixtures first
# (each trigger and each environment form must be caught), then on the real .github/workflows/.
# Usage: scripts/tests/review-agent-triggers_test.sh   (WORKFLOWS_DIR=<dir> checks another directory instead)
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"

# check_dir <dir>: prints one line per violation, exits 1 when there is one (or when no workflow is found).
# Stdlib only (no YAML library is a project dependency), so it reads the text and errs on the side of refusing.
# A job names review-agent when its non-comment text does (environment as string, mapping, quoted or flow form). In a
# workflow whose top-level `on:` block names a forbidden event, such a job is allowed only when it, or a job it needs
# (transitively), has a one-line `if:` that selects a non-pull-request event (`github.event_name == 'schedule'`,
# 'workflow_dispatch', 'workflow_run' or 'push'): then a pull request event never starts it. No guard, or a guard
# the reader cannot see (a multi-line `if:`), is a violation.
check_dir() {
  python3 -I - "$1" <<'PY'
import glob, os, re, sys

EVENTS = re.compile(r"(?<![\w-])(pull_request|pull_request_review|pull_request_target)(?![\w-])")
GUARD = re.compile(r"github\.event_name\s*==\s*'(schedule|workflow_dispatch|workflow_run|push)'")
bad = 0
files = sorted(glob.glob(os.path.join(sys.argv[1], "*.yml")) + glob.glob(os.path.join(sys.argv[1], "*.yaml")))
if not files:
    print("no workflow found in %s" % sys.argv[1])
    sys.exit(1)


def jobs_of(lines):
    jobs, name, inside = {}, None, False
    for ln in lines:
        if re.match(r"^jobs\s*:", ln):
            inside = True
            continue
        if inside and re.match(r"^\S", ln):
            inside = False
        if not inside:
            continue
        m = re.match(r"^  ([\w-]+)\s*:\s*$", ln)
        if m:
            name = m.group(1)
            jobs[name] = []
        elif name:
            jobs[name].append(ln)
    return jobs


def needs_of(block):
    out, listing = [], False
    for ln in block:
        m = re.match(r"^    needs\s*:\s*(.*)$", ln)
        if m:
            out += re.findall(r"[\w-]+", m.group(1))
            listing = not m.group(1).strip()
            continue
        if listing:
            m = re.match(r"^      -\s*([\w-]+)", ln)
            if m:
                out.append(m.group(1))
            elif ln.strip():
                listing = False
    return out


def guarded(name, jobs, seen=()):
    if name in seen or name not in jobs:
        return False
    for ln in jobs[name]:
        m = re.match(r"^    if\s*:\s*(.*)$", ln)
        if m and GUARD.search(m.group(1)):
            return True
    return any(guarded(n, jobs, seen + (name,)) for n in needs_of(jobs[name]))


for path in files:
    with open(path, encoding="utf-8") as fh:
        lines = [re.sub(r"(^|\s)#.*$", "", ln.rstrip("\n")) for ln in fh]
    on_lines, inside = [], False
    for ln in lines:
        if re.match(r"""^(on|"on"|'on')\s*:""", ln):
            inside = True
            on_lines.append(ln.split(":", 1)[1])
            continue
        if inside and re.match(r"^\S", ln):
            inside = False
        if inside:
            on_lines.append(ln)
    hit = sorted(set(EVENTS.findall("\n".join(on_lines))))
    if not hit:
        continue
    jobs = jobs_of(lines)
    for name, block in jobs.items():
        if any("review-agent" in ln for ln in block) and not guarded(name, jobs):
            print("%s: job %s declares environment review-agent and trigger %s can start it" % (os.path.basename(path), name, ",".join(hit)))
            bad += 1
sys.exit(1 if bad else 0)
PY
}

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

mkfix() { # mkfix <name> <on-block> <environment-block>
  mkdir -p "$tmp/$1"
  printf 'name: x\non:\n%s\njobs:\n  j:\n    runs-on: ubuntu-latest\n%s\n    steps:\n      - run: echo ok\n' "$2" "$3" >"$tmp/$1/w.yml"
}
expect_bad() { local rc=0; check_dir "$tmp/$1" >/dev/null 2>&1 || rc=$?; [ "$rc" -eq 1 ] && ok "$2" || ko "$2 (rc=$rc)"; }
expect_good() { local rc=0; check_dir "$tmp/$1" >/dev/null 2>&1 || rc=$?; [ "$rc" -eq 0 ] && ok "$2" || ko "$2 (rc=$rc)"; }

mkfix pr '  pull_request:' '    environment: review-agent'
expect_bad pr "pull_request with environment: review-agent is refused"
mkfix prt '  pull_request_target:
    types: [opened]' '    environment: review-agent'
expect_bad prt "pull_request_target is refused"
mkfix prr '  pull_request_review:
    types: [submitted]' '    environment: review-agent'
expect_bad prr "pull_request_review is refused"
mkfix lst '  - pull_request' '    environment: review-agent'
printf 'name: x\non: [push, pull_request]\njobs:\n  j:\n    runs-on: ubuntu-latest\n    environment: review-agent\n    steps:\n      - run: echo ok\n' >"$tmp/lst/w.yml"
expect_bad lst "a trigger list is read"
mkfix map '  pull_request:' '    environment:
      name: review-agent
      url: https://example.invalid'
expect_bad map "the mapping form (name: review-agent) is refused"
mkfix quoted '  pull_request:' '    environment: "review-agent"'
expect_bad quoted "the quoted form is refused"
mkfix ok '  workflow_run:
    workflows: [x]
    types: [completed]' '    environment: review-agent'
expect_good ok "workflow_run with review-agent is allowed"
mkfix other '  pull_request:' '    environment: staging'
expect_good other "pull_request without review-agent is allowed"
mkdir -p "$tmp/guard"
printf 'name: x\non:\n  schedule:\n    - cron: "0 * * * *"\n  pull_request:\n    branches: [main]\njobs:\n  plan:\n    if: github.event_name == '"'"'schedule'"'"'\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo ok\n  publish:\n    needs: [plan]\n    environment: review-agent\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo ok\n' >"$tmp/guard/w.yml"
expect_good guard "a pull_request workflow whose review-agent job needs a schedule-only job is allowed"
sed -i 's/^    needs: \[plan\]$/    needs: []/' "$tmp/guard/w.yml"
expect_bad guard "the same job without the guarded need is refused"
mkdir -p "$tmp/empty"
rc=0; check_dir "$tmp/empty" >/dev/null 2>&1 || rc=$?
[ "$rc" -ne 0 ] && ok "a directory without workflow fails (never a silent pass)" || ko "empty dir passed"
mkfix flow '  push:' '    environment: { name: review-agent }'
expect_good flow "review-agent on a push-only workflow is allowed"
printf 'name: x\non: pull_request # not review-agent here\njobs:\n  j:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo ok\n' >"$tmp/flow/w.yml"
expect_good flow "a comment never counts, and a pull_request workflow without review-agent is allowed"

dir="${WORKFLOWS_DIR:-$root/.github/workflows}"
rc=0; out="$(check_dir "$dir" 2>&1)" || rc=$?
[ "$rc" -eq 0 ] && ok "no review-agent job is reachable from a pull request event in $dir" || ko "violation in $dir: $out"
n="$(grep -l 'environment:' "$dir"/*.y*ml 2>/dev/null | wc -l)"
[ "$n" -gt 0 ] && ok "the scan saw workflows with an environment ($n)" || ko "no workflow with an environment found in $dir"

finish
