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
# FAIL CLOSED. Stdlib only (no YAML library is a project dependency), so this is a line reader for the shapes this
# repository writes, and every shape it cannot read with certainty takes the dangerous reading:
#   - a job's `environment` that is not a plain name (expression, anchor, alias, flow mapping, no readable name) holds review-agent;
#   - an `on:` that cannot be read (missing, the YAML 1.1 key `true`, alias, flow mapping, multi-line flow list, anything
#     but a block mapping, a block list or a one-line flow list of plain names) has every pull_request* trigger;
#   - an `if:` is a guard only when its top level is `&&` of terms (no top-level `||`) and one term is
#     `github.event_name == 'schedule'|'workflow_dispatch'|'workflow_run'|'push'` or a parenthesised `||` of only such tests;
#   - a guard coming only from `needs:` (a plain name, one-line list or block list) is accepted when the job has no `if:`, or
#     an `if:` without a status function, or one that also requires a non-empty output of that guarded need (always() alone
#     would run the job after a skipped need);
#   - a workflow that mentions an environment must have readable jobs (2-space keys, block mappings), no anchor, alias or
#     merge key, and no `workflow_call` trigger (callers cannot be proven).
# So the test can refuse a safe file; it cannot accept an unsafe one (what it cannot read, it refuses).
check_dir() {
  python3 -I - "$1" <<'PY'
import glob, os, re, sys

ALL = ["pull_request", "pull_request_review", "pull_request_target"]
EVENTS = re.compile(r"(?<![\w-])(pull_request|pull_request_review|pull_request_target)(?![\w-])")
EQ = r"github\.event_name\s*==\s*'(?:schedule|workflow_dispatch|workflow_run|push)'"
GUARD_TERM = re.compile(r"^(?:%s|\(\s*%s(?:\s*\|\|\s*%s)*\s*\))$" % (EQ, EQ, EQ))
STATUS = re.compile(r"\b(always|cancelled|failure|success)\s*\(")
PLAIN = re.compile(r"""^["']?([A-Za-z0-9_.-]+)["']?$""")
NAME = r"[\w-]+"
bad = 0
files = sorted(glob.glob(os.path.join(sys.argv[1], "*.yml")) + glob.glob(os.path.join(sys.argv[1], "*.yaml")))
if not files:
    print("no workflow found in %s" % sys.argv[1])
    sys.exit(1)


def indent(ln):
    return len(ln) - len(ln.lstrip(" "))


def structural(lines):
    """Lines outside block scalars (run: | ...), where shell text may hold & and *."""
    out, skip = [], None
    for ln in lines:
        if skip is not None:
            if not ln.strip() or indent(ln) > skip:
                continue
            skip = None
        if re.search(r"[:\-]\s*[|>][+-]?\d*\s*$", ln):
            skip = indent(ln)
        out.append(ln)
    return out


def split_top(expr, op):
    """Split on a top-level operator outside parentheses and quotes."""
    parts, depth, cur, quote, i = [], 0, "", None, 0
    while i < len(expr):
        c = expr[i]
        if quote:
            quote = None if c == quote else quote
        elif c in "'\"":
            quote = c
        elif c == "(":
            depth += 1
        elif c == ")":
            depth -= 1
        elif depth == 0 and expr.startswith(op, i):
            parts.append(cur)
            cur, i = "", i + len(op)
            continue
        cur += c
        i += 1
    parts.append(cur)
    return [p.strip() for p in parts]


def conjuncts(value):
    """The top-level && terms of an if: value, or None when it cannot be read (a top-level ||, ${{ }} unbalanced...)."""
    v = value.strip()
    m = re.match(r"^\$\{\{\s*(.*?)\s*\}\}$", v)
    if m:
        v = m.group(1)
    if not v or v[0] in "|>&*!\"'" or "${{" in v:
        return None
    if len(split_top(v, "||")) > 1:
        return None
    return split_top(v, "&&")


def read_on(lines):
    """Returns (events, certain). Anything not readable with certainty is every pull_request* event."""
    start = None
    for i, ln in enumerate(lines):
        if re.match(r"""^(on|"on"|'on'|true|"true"|'true')\s*:""", ln):
            start = i
            break
    if start is None:
        return list(ALL), False
    key = lines[start]
    if re.match(r"^(true|\"true\"|'true')", key):
        return list(ALL), False
    inline = key.split(":", 1)[1].strip()
    if inline:
        if re.match(r"^[\w-]+$", inline) or re.match(r"^\[\s*[\w-]+(\s*,\s*[\w-]+)*\s*\]$", inline):
            return sorted(set(EVENTS.findall(inline))), True
        return list(ALL), False
    body = []
    for ln in lines[start + 1:]:
        if re.match(r"^\S", ln):
            break
        body.append(ln)
    for ln in body:
        if ln.strip() and indent(ln) == 2 and not re.match(r"^  (?:-\s+)?[A-Za-z_][\w-]*\s*:?(?:\s.*)?$", ln):
            return list(ALL), False
        if re.search(r"(?:^|[\s:\[{,-])[&*][\w-]", ln):
            return list(ALL), False
    if not any(ln.strip() for ln in body):
        return list(ALL), False
    return sorted(set(EVENTS.findall("\n".join(body)))), True


def parse_jobs(lines):
    """Returns (jobs {name: lines}, ok). ok is False when the jobs block is not plain block mappings at two spaces."""
    jobs, name, inside, found, ok = {}, None, False, False, True
    for ln in lines:
        if re.match(r"^jobs\s*:", ln):
            found, inside = True, True
            if ln.split(":", 1)[1].strip():
                ok = False
            continue
        if inside and re.match(r"^\S", ln):
            inside = False
        if not inside or not ln.strip():
            continue
        if indent(ln) == 2:
            m = re.match(r"^  (%s)\s*:\s*$" % NAME, ln)
            if m:
                name = m.group(1)
                jobs[name] = []
            else:
                ok = False
        elif name is None or indent(ln) < 2:
            ok = False
        else:
            jobs[name].append(ln)
    for blk in jobs.values():
        first = next((l for l in blk if l.strip()), None)
        if first is not None and indent(first) != 4:
            ok = False
    return jobs, found and ok


def env_holds_review_agent(block):
    """True when the job's environment is, or might be, review-agent (anything but a plain other name)."""
    for i, ln in enumerate(block):
        m = re.match(r"^    environment\s*:\s*(.*)$", ln)
        if not m:
            continue
        v = m.group(1).strip()
        if not v:
            v = None  # the mapping ends at the first line indented less than six spaces
            for x in block[i + 1:]:
                if x.strip() and indent(x) < 6:
                    break
                n = re.match(r"^      name\s*:\s*(.*)$", x)
                if n:
                    v = n.group(1).strip()
                    break
            if v is None:
                return True
        p = PLAIN.match(v)
        return not p or p.group(1) == "review-agent"
    return False


def needs_of(block):
    """Plain needs names, or None when the form is not a plain name, a one-line list or a block list."""
    out, listing = [], False
    for ln in block:
        m = re.match(r"^    needs\s*:\s*(.*)$", ln)
        if m:
            v = m.group(1).strip()
            if not v:
                listing = True
            elif re.match(r"^%s$" % NAME, v):
                out.append(v)
            elif re.match(r"^\[\s*%s(\s*,\s*%s)*\s*\]$" % (NAME, NAME), v):
                out += re.findall(NAME, v)
            else:
                return None
            continue
        if listing:
            m = re.match(r"^      -\s*(%s)\s*$" % NAME, ln)
            if m:
                out.append(m.group(1))
            elif ln.strip() and indent(ln) <= 4:
                listing = False
            elif ln.strip():
                return None
    return out


def own_if(block):
    for ln in block:
        m = re.match(r"^    if\s*:\s*(.*)$", ln)
        if m:
            return m.group(1)
    return None


def guarded(name, jobs, seen=()):
    if name in seen or name not in jobs:
        return False
    cond = own_if(jobs[name])
    terms = conjuncts(cond) if cond is not None else []
    if terms and any(GUARD_TERM.match(t) for t in terms):
        return True
    needs = needs_of(jobs[name]) or []
    for n in needs:
        if not guarded(n, jobs, seen + (name,)):
            continue
        if cond is None:
            return True
        if terms is None:
            continue
        if not STATUS.search(cond):
            return True
        # a status function runs the job after a skipped need: require a non-empty output of that need as well
        if any(re.match(r"^needs\.%s\.outputs\.[\w-]+\s*==\s*'[^']+'$" % re.escape(n), t) for t in terms):
            return True
    return False


for path in files:
    base = os.path.basename(path)
    with open(path, encoding="utf-8") as fh:
        raw = [ln.rstrip("\n") for ln in fh]
    lines = [re.sub(r"(^|\s)#.*$", "", ln) for ln in raw]
    mentions = any(re.search(r"\benvironment\b|review-agent", ln) for ln in lines)
    if not mentions:
        continue
    events, certain = read_on(lines)
    jobs, readable = parse_jobs(lines)
    holders = [n for n, b in jobs.items() if env_holds_review_agent(b) or any("review-agent" in ln for ln in b)]
    if not readable:
        print("%s: the jobs block cannot be read with certainty (flow mapping, odd indentation or no jobs) and the file mentions an environment" % base)
        bad += 1
        continue
    if any(re.search(r"(?:^|[\s\[{,:-])[&*][A-Za-z0-9_][\w-]*|<<\s*:", ln) for ln in structural(lines)):
        print("%s: a YAML anchor, alias or merge key in a workflow that has an environment" % base)
        bad += 1
        continue
    if not holders:
        continue
    if "workflow_call" in "\n".join(lines):
        print("%s: reusable workflow (workflow_call) holds review-agent; its callers cannot be proven" % base)
        bad += 1
        continue
    hit = events
    if not hit:
        continue
    for name in holders:
        if not guarded(name, jobs):
            print("%s: job %s declares or may declare environment review-agent and trigger %s%s can start it" % (base, name, ",".join(hit), "" if certain else " (on: unreadable, assumed)"))
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

# --- Fail-closed fixtures: forms the reader cannot read with certainty count as the dangerous reading. ----------------
# put <name>: writes stdin as <name>/w.yml. The "bad" ones are safe-looking or exotic, never accepted; the "good" ones
# are the narrow shapes the reader can prove.
put() { mkdir -p "$tmp/$1"; cat >"$tmp/$1/w.yml"; }
put env_expr <<'Y'
on:
  pull_request:
jobs:
  j:
    runs-on: ubuntu-latest
    environment: ${{ vars.ENV_NAME }}
    steps:
      - run: echo ok
Y
expect_bad env_expr "an environment name built from an expression counts as review-agent"
put env_map_expr <<'Y'
on:
  pull_request:
jobs:
  j:
    runs-on: ubuntu-latest
    environment:
      name: ${{ inputs.target }}
    steps:
      - run: echo ok
Y
expect_bad env_map_expr "a mapping environment whose name is an expression is refused"
put env_map_noname <<'Y'
on:
  pull_request:
jobs:
  j:
    runs-on: ubuntu-latest
    environment:
      url: https://example.invalid
    steps:
      - run: echo ok
Y
expect_bad env_map_noname "a mapping environment without a readable name is refused"
put env_alias <<'Y'
on:
  pull_request:
jobs:
  a:
    runs-on: ubuntu-latest
    environment: &e staging
    steps:
      - run: echo ok
  b:
    runs-on: ubuntu-latest
    environment: *e
    steps:
      - run: echo ok
Y
expect_bad env_alias "an environment given by anchor and alias is refused"
put merge_key <<'Y'
on:
  pull_request:
x-defaults: &defaults
  runs-on: ubuntu-latest
jobs:
  j:
    <<: *defaults
    environment: staging
    steps:
      - run: echo ok
Y
expect_bad merge_key "a YAML anchor or merge key in a workflow that has an environment is refused"
put flow_job <<'Y'
on:
  pull_request:
jobs:
  j: {runs-on: ubuntu-latest, environment: {name: review-agent}, steps: [{run: echo ok}]}
Y
expect_bad flow_job "a job written as a flow mapping is refused"
put flow_job_hidden <<'Y'
on:
  pull_request:
jobs:
  j: {runs-on: ubuntu-latest, environment: "${{ vars.E }}", steps: [{run: echo ok}]}
Y
expect_bad flow_job_hidden "a flow-mapping job with an expression environment is refused"
put odd_indent <<'Y'
on:
  pull_request:
jobs:
    j:
        runs-on: ubuntu-latest
        environment: review-agent
Y
expect_bad odd_indent "jobs indented other than two spaces cannot be read and are refused"
put if_or_pr <<'Y'
on:
  pull_request:
jobs:
  j:
    if: github.event_name == 'push' || github.event_name == 'pull_request'
    environment: review-agent
    runs-on: ubuntu-latest
    steps:
      - run: echo ok
Y
expect_bad if_or_pr "an if: that also lets pull_request through is unguarded"
put if_or_true <<'Y'
on:
  pull_request:
jobs:
  j:
    if: github.event_name == 'schedule' || vars.ANY == 'x'
    environment: review-agent
    runs-on: ubuntu-latest
    steps:
      - run: echo ok
Y
expect_bad if_or_true "an if: with a branch that is not an event_name test is unguarded"
put if_andor <<'Y'
on:
  pull_request:
jobs:
  j:
    if: github.event_name == 'schedule' && vars.A == 'x' || github.event.pull_request
    environment: review-agent
    runs-on: ubuntu-latest
    steps:
      - run: echo ok
Y
expect_bad if_andor "a && b || c is unguarded (precedence cannot be trusted)"
put if_neg <<'Y'
on:
  pull_request:
jobs:
  j:
    if: "!(github.event_name == 'schedule')"
    environment: review-agent
    runs-on: ubuntu-latest
    steps:
      - run: echo ok
Y
expect_bad if_neg "a negated event_name test is unguarded"
put needs_always <<'Y'
on:
  pull_request:
jobs:
  plan:
    if: github.event_name == 'schedule'
    runs-on: ubuntu-latest
    steps:
      - run: echo ok
  j:
    needs: [plan]
    if: always()
    environment: review-agent
    runs-on: ubuntu-latest
    steps:
      - run: echo ok
Y
expect_bad needs_always "always() on the job defeats a guard that only comes from needs"
put needs_expr <<'Y'
on:
  pull_request:
jobs:
  plan:
    if: github.event_name == 'schedule'
    runs-on: ubuntu-latest
    steps:
      - run: echo ok
  j:
    needs: ${{ fromJSON('["plan"]') }}
    environment: review-agent
    runs-on: ubuntu-latest
    steps:
      - run: echo ok
Y
expect_bad needs_expr "a needs: given as an expression is not a proven guard"
put on_true_key <<'Y'
true:
  pull_request:
jobs:
  j:
    environment: review-agent
    runs-on: ubuntu-latest
    steps:
      - run: echo ok
Y
expect_bad on_true_key "on: read by YAML 1.1 as the key true is read as every trigger"
put on_alias <<'Y'
x-triggers: &triggers
  pull_request:
on: *triggers
jobs:
  j:
    environment: review-agent
    runs-on: ubuntu-latest
    steps:
      - run: echo ok
Y
expect_bad on_alias "on: given by alias is read as every trigger"
put on_flow_map <<'Y'
on: { pull_request: {}, push: {} }
jobs:
  j:
    environment: review-agent
    runs-on: ubuntu-latest
    steps:
      - run: echo ok
Y
expect_bad on_flow_map "on: as a flow mapping is refused"
put on_missing <<'Y'
jobs:
  j:
    environment: review-agent
    runs-on: ubuntu-latest
    steps:
      - run: echo ok
Y
expect_bad on_missing "a workflow whose on: cannot be found is read as every trigger"
put on_multiline_flow <<'Y'
on: [
  push,
  pull_request
]
jobs:
  j:
    environment: review-agent
    runs-on: ubuntu-latest
    steps:
      - run: echo ok
Y
expect_bad on_multiline_flow "a multi-line flow list of triggers is refused"
put call <<'Y'
on:
  workflow_call:
jobs:
  j:
    environment: review-agent
    runs-on: ubuntu-latest
    steps:
      - run: echo ok
Y
expect_bad call "a reusable workflow holding review-agent is refused (its callers cannot be proven)"
put good_if_and <<'Y'
on:
  pull_request:
  schedule:
    - cron: "0 * * * *"
jobs:
  j:
    if: (github.event_name == 'schedule' || github.event_name == 'workflow_dispatch') && vars.AUTO_FIX != 'off'
    environment: review-agent
    runs-on: ubuntu-latest
    steps:
      - run: echo ok
Y
expect_good good_if_and "a job whose own if: is (schedule || workflow_dispatch) && other is allowed"
put good_needs_out <<'Y'
on:
  pull_request:
  schedule:
    - cron: "0 * * * *"
jobs:
  plan:
    if: github.event_name == 'schedule'
    runs-on: ubuntu-latest
    steps:
      - run: echo ok
  j:
    needs: [plan]
    if: always() && needs.plan.outputs.run == 'true' && (needs.plan.outputs.k == 'a' || needs.plan.outputs.k == 'b')
    environment: review-agent
    runs-on: ubuntu-latest
    steps:
      - run: echo ok
Y
expect_good good_needs_out "always() is accepted when the job also requires a non-empty output of its guarded need"
put good_quoted_env <<'Y'
on:
  pull_request:
jobs:
  a:
    runs-on: ubuntu-latest
    environment: 'staging'
    steps:
      - run: echo ok
  b:
    runs-on: ubuntu-latest
    environment:
      name: "staging"
    steps:
      - run: echo ok
Y
expect_good good_quoted_env "plain quoted environment names other than review-agent are allowed"

dir="${WORKFLOWS_DIR:-$root/.github/workflows}"
rc=0; out="$(check_dir "$dir" 2>&1)" || rc=$?
[ "$rc" -eq 0 ] && ok "no review-agent job is reachable from a pull request event in $dir" || ko "violation in $dir: $out"
n="$(grep -l 'environment:' "$dir"/*.y*ml 2>/dev/null | wc -l)"
[ "$n" -gt 0 ] && ok "the scan saw workflows with an environment ($n)" || ko "no workflow with an environment found in $dir"

finish
