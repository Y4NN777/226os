#!/usr/bin/env python3
"""Review auto-fix decisions (stdlib only): when the CI fixer (.claude/commands/review-fix.md) may run, on what, and
whether its commits may be pushed. The workflows only call this script; the model never decides any of it.

  pr     stage 3 (.github/workflows/review-autofix.yml, after the "Code review" workflow). Runs the fixer when the
         ledger of the review comment has an OPEN finding (kind `finding`, not a note) of MEDIUM or HIGH, for the
         pull request head the review looked at. Refused: a fork, a base other than develop, a head branch main or
         develop, a closed or draft pull request, a Code review run whose review job did not run (a label or review
         event), a ledger or run for an older head. An open CRITICAL sends the whole pull request to the owner (label
         needs-owner, no fix). At most 2 attempts, counted by the labels autofix-1 and autofix-2 (written before the
         fixer runs, so a failed run counts); the third is refused and labelled needs-owner.
  batch  stage 5 (.github/workflows/review-batch.yml). Runs the fixer over the unticked items of the open "Review
         batch" issue when the batch step created or updated it (BATCH_RESULT; FORCE=true on a manual run), unless a
         pull request from a fix/review/batch-* branch into develop is already open. Branch fix/review/batch-TODAY,
         with a -2, -3... suffix when that branch exists.
  prepare  right before the model, after `make setup`: wipes ~/.claude, ~/.claude.json and ~/.config/claude* and
         restores .claude, CLAUDE.md, .mcp.json and the other Claude configuration paths from SOURCE_REF (work tree only).
  apply  the trusted job's way in: validates the patch files of PATCH_DIR (at most 20, at most 1 MB in all, names
         NNNN-slug.patch), then applies each one with `git apply --index` and commits it itself as EXPECT_NAME
         <EXPECT_EMAIL> (author and committer), hooks off: the untrusted job's From: line is never trusted.
  check  the guard before any push, run from a trusted copy: at least one new commit since BASE_SHA; every author and
         committer is EXPECT_NAME <EXPECT_EMAIL>; no Co-authored-by trailer and no "Generated with" line; no symbolic
         link, submodule, file mode change, binary change or new executable outside scripts/; and every changed path,
         on BOTH sides of a change (renames split, paths read NUL-separated), is in the ALLOWLIST src/** (minus the web
         build, lint, format, type and threshold configuration and package.json), db/queries/**, scripts/tests/**,
         docs/ai-usage/** and TODO.md. The deny list stays as a second layer: the contract (R7),
         migrations (R6), ADRs, docs/quality, secrets, .env files, any .npmrc, .github, the Claude harness, Make files,
         the review, findings, coverage and check scripts, the route and secret-scan lists, the owner's preferences.
         Exit 1 with the reason; the workflow then labels needs-owner.
  run    `run CMD ARGS...`, installed as `fixer-run` for the model: git status|diff|log|show with an allowlist of
         options, go test with an allowlist of flags (-run, -count, -race, -tags, -v, -timeout) and ./ package paths,
         `pnpm --filter web run test|lint|typecheck|format:check` with paths only. With FIXER_NETNS=1, go test and pnpm
         run in a namespace without network (`unshare -rn`). DEFENCE IN DEPTH, NOT A BOUNDARY: make, go test and pnpm
         run code the pull request controls; the boundary is that the job running them holds no write token.
  body   the batch pull request body from the fixer's structured result (RESULT_JSON): it names the batch issue
         (#ISSUE, which closes the batch on merge) and has one table row per item (item, source, decision, reason): FIXED,
         FIXED ELSEWHERE, NOT A DEFECT, DECIDED, OWNER or AFTER FREEZE. `findings.py settle` reads it on merge. Model text is data: @-mentions and #N references (so closing keywords)
         are neutralised, pipes escaped.

Mode: AUTO_FIX = off | report (default: says what it would do, writes nothing) | on. HAS_TOKEN must be `true` (the
publish jobs alone hold secrets.REVIEW_AGENT_TOKEN, in the environment review-agent, so the decide and plan jobs pass
`true` and the publish job reports a missing token): without it, "autofix: NOT RUN (REVIEW_AGENT_TOKEN missing)" and
nothing is written. Only `on` with the token writes, and only labels.
Outputs (`pr`, `batch`) are `key=value` lines appended to GITHUB_OUTPUT: run=true|false, then attempt, findings,
branch, head_sha (pr) or issue, branch, items (batch). They carry ids and numbers only, never ledger or issue text.
Fail OPEN for pr and batch (a GitHub error is NOT RUN with run=false, exit 0); fail CLOSED for check.
Environment: REPO, PR_NUMBER, RUN_ID, RUN_HEAD_SHA, BATCH_RESULT, FORCE, TODAY (YYYY-MM-DD), BASE_SHA, WORKTREE,
EXPECT_NAME, EXPECT_EMAIL, ISSUE, RESULT_JSON, GH_BIN (default gh), GIT_BIN (default git).
"""
import datetime
import fnmatch
import importlib.util
import json
import os
import re
import subprocess
import sys

FIXABLE = ("HIGH", "MEDIUM")
MAX_ATTEMPTS = 2
OWNER_LABEL = "needs-owner"
REVIEW_JOB = "Code review (advisory)"
BATCH_PREFIX = "fix/review/batch-"
PROTECTED_HEADS = ("main", "develop")
VERDICTS = ("FIXED", "FIXED ELSEWHERE", "NOT A DEFECT", "DECIDED", "OWNER", "AFTER FREEZE")
# Paths the fixer never changes: the contract (R7), migrations (R6), decisions, secrets, CI, the harness itself and the
# files the workflows leave in the workspace.
FORBIDDEN_EXACT = ("api/openapi.yaml", "CLAUDE.md", "CLAUDE.local.md", "AGENTS.md", "docs/team/OWNER_PREFERENCES.md",
                   "Makefile", "scripts/findings.py", "scripts/close-merged-issues.py", "scripts/gitleaks-ignore.txt",
                   "scripts/public-routes.txt", "scripts/unimplemented-routes.txt", "scripts/authz-tests.txt",
                   ".mcp.json", ".claude.json", ".gitmodules")
FORBIDDEN_PREFIX = ("db/migrations/", "docs/decisions/", "docs/quality/", "secrets/", ".github/", ".claude/", "mk/",
                    ".husky/")
FORBIDDEN_GLOB = ("scripts/review-*.py", "scripts/coverage-*.sh", "scripts/check-*.py", "scripts/check-*.sh")
FORBIDDEN_NAMES = (".npmrc",)
# The fixer may change ONLY these (an allowlist); everything else is the owner's. Inside src/, the web build, lint,
# format, type and test-threshold configuration and the package manifest stay the owner's too.
ALLOWED_PREFIX = ("src/", "db/queries/", "scripts/tests/", "docs/ai-usage/")
ALLOWED_EXACT = ("TODO.md",)
OWNER_IN_SRC = re.compile(r"^src/web/(package\.json|(vite|vitest|eslint|prettier)\.config\.[cm]?[jt]s|\.eslintrc.*|"
                          r"\.prettierrc.*|tsconfig.*\.json)$")
CLAUDE_CONFIG = (".claude", ".mcp.json", ".claude.json", ".gitmodules", ".ripgreprc", "CLAUDE.md", "CLAUDE.local.md", ".husky")
HOME_CONFIG = (".claude", ".claude.json")
# The wrapper of the fixer's commands (`run`): allowlists of subcommands, options and flags (defence in depth).
WEB_SCRIPTS = ("test", "lint", "typecheck", "format:check")
GIT_OPTIONS = {
    "status": {"--short", "-s", "--porcelain", "-b", "--branch"},
    "diff": {"--stat", "--cached", "--staged", "--name-only", "--name-status", "--no-color", "--"},
    "log": {"--oneline", "--stat", "--name-only", "--no-color", "-p", "--"},
    "show": {"--stat", "--name-only", "--no-color", "--"},
}
GIT_NUMERIC = re.compile(r"^-(\d{1,4}|U\d{1,2})$")
GO_VALUE_FLAGS = {"-run": re.compile(r"^[\w^$|().*+\[\]/-]{1,200}$"), "-count": re.compile(r"^\d{1,3}$"),
                  "-tags": re.compile(r"^[\w,]{1,60}$"), "-timeout": re.compile(r"^\d{1,4}[smh]$")}
GO_BOOL_FLAGS = ("-race", "-v")
GO_PACKAGE = re.compile(r"^\./[\w./-]*$|^\.$")
PATCH_NAME = re.compile(r"^[0-9]{4}-[a-z0-9-]+\.patch$")
MAX_PATCHES = 20
MAX_PATCH_BYTES = 1000000
ITEM_KEY = re.compile(r"^#\d+(?:/[\w-]+)?$")
TRAILER = re.compile(r"^\s*co-authored-by:|generated with", re.I | re.M)


def load_context_module():
    """scripts/review-context.py has a dash in its name: load it by path (same directory as this file)."""
    path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "review-context.py")
    spec = importlib.util.spec_from_file_location("review_context", path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def emit(**pairs):
    """Append key=value lines to GITHUB_OUTPUT (when set) and echo them."""
    lines = ["%s=%s" % (k, v) for k, v in pairs.items()]
    target = os.environ.get("GITHUB_OUTPUT")
    if target:
        with open(target, "a", encoding="utf-8") as fh:
            fh.write("\n".join(lines) + "\n")
    for line in lines:
        print("output: " + line)


def stop(message):
    print(message)
    emit(run="false")
    return 0


def safe_id(text):
    return re.sub(r"[^\w.-]+", "-", text).strip("-")[:40]


def add_label(gh, repo, number, label):
    done = subprocess.run([gh, "api", "-X", "POST", "repos/%s/issues/%s/labels" % (repo, number), "-f", "labels[]=%s" % label],
                          capture_output=True, text=True, timeout=60, check=False)
    print("autofix: labelled %s" % label if done.returncode == 0 else "autofix: could not label %s" % label)


def mode_and_token():
    mode = (os.environ.get("AUTO_FIX") or "report").strip().lower()
    if mode not in ("off", "report", "on"):
        mode = "report"
    return mode, os.environ.get("HAS_TOKEN", "") == "true"


def label_names(labels):
    return {x if isinstance(x, str) else x.get("name", "") for x in labels or []}


def review_job_ran(ctx, gh, repo, run_id):
    """True when the Code review run had a review job that was not skipped (a label or review event runs only the gate)."""
    if not run_id:
        return True
    pages = ctx.gh_api(gh, "repos/%s/actions/runs/%s/jobs" % (repo, run_id))
    jobs = [j for page in pages for j in (page.get("jobs") or [])]
    return any(j.get("name") == REVIEW_JOB and j.get("conclusion") not in (None, "skipped", "cancelled") for j in jobs)


def decide_pr(ctx, gh, repo, number, mode):
    """(verdict, detail): verdict is refuse | nothing | owner | exhausted | fix (detail: (attempt, ids, info))."""
    info = (ctx.gh_api(gh, "repos/%s/pulls/%s" % (repo, number)) or [{}])[0]
    head, base = info.get("head") or {}, info.get("base") or {}
    if (head.get("repo") or {}).get("full_name") != repo:
        return "refuse", "a pull request from a fork is never fixed (no secrets for forks)"
    if base.get("ref") != "develop":
        return "refuse", "only pull requests into develop are fixed (base %s)" % base.get("ref")
    if head.get("ref") in PROTECTED_HEADS:
        return "refuse", "the head branch %s is never written" % head.get("ref")
    if info.get("state") != "open" or info.get("draft"):
        return "refuse", "the pull request is closed or a draft"
    head_sha = head.get("sha", "")
    run_sha = os.environ.get("RUN_HEAD_SHA", "")
    if run_sha and run_sha != head_sha:
        return "refuse", "this review run is for an older head (a newer push exists)"
    if not review_job_ran(ctx, gh, repo, os.environ.get("RUN_ID", "")):
        return "refuse", "the review job did not run in this Code review run (a label or review event)"
    ledger = ctx.parse_ledger(ctx.gh_api(gh, "repos/%s/issues/%s/comments" % (repo, number)))
    if not ledger:
        return "nothing", "no review ledger on the pull request"
    if not head_sha.startswith(ledger["sha"]):
        return "refuse", "the ledger is for %s, not for the pull request head: the review is stale" % ledger["sha"][:12]
    open_findings = [f for f in ledger["findings"] if f["kind"] == "finding" and f["status"] == "open"]
    if any(f["sev"] == "CRITICAL" for f in open_findings):
        return "owner", "an open CRITICAL goes to the owner, the pull request is not fixed"
    ids = [safe_id(f["id"]) for f in open_findings if f["sev"] in FIXABLE]
    ids = [i for i in ids if i]
    if not ids:
        return "nothing", "no open MEDIUM or HIGH finding (notes are never fixed)"
    labels = label_names(info.get("labels"))
    done = sum(1 for n in range(1, MAX_ATTEMPTS + 1) if "autofix-%d" % n in labels)
    if done >= MAX_ATTEMPTS:
        return "exhausted", "%d attempts already made: the owner decides" % MAX_ATTEMPTS
    return "fix", (done + 1, ids, head)


def cmd_pr():
    repo, number = os.environ.get("REPO", ""), os.environ.get("PR_NUMBER", "")
    gh = os.environ.get("GH_BIN", "gh")
    mode, token = mode_and_token()
    if mode == "off":
        return stop("autofix: off (AUTO_FIX=off)")
    if not repo or not number:
        return stop("autofix: NOT RUN (no pull request for this run)")
    try:
        verdict, detail = decide_pr(load_context_module(), gh, repo, number, mode)
    except Exception as exc:  # fail open: a broken read never fails the workflow
        return stop("autofix: NOT RUN (%s)" % type(exc).__name__)
    writes = mode == "on" and token
    if verdict in ("refuse", "nothing"):
        return stop("autofix: skip (%s)" % detail)
    if mode == "on" and not token:
        return stop("autofix: NOT RUN (REVIEW_AGENT_TOKEN missing)")
    if verdict in ("owner", "exhausted"):
        print("autofix: owner (%s)" % detail)
        if writes:
            add_label(gh, repo, number, OWNER_LABEL)
        else:
            print("autofix: report mode, nothing written")
        emit(run="false")
        return 0
    attempt, ids, head = detail
    if not writes:
        return stop("autofix: report mode: would fix %s on %s (attempt %d of %d)" % (",".join(ids), head.get("ref"), attempt, MAX_ATTEMPTS))
    add_label(gh, repo, number, "autofix-%d" % attempt)
    print("autofix: fixing %d finding(s) on %s (attempt %d of %d)" % (len(ids), head.get("ref"), attempt, MAX_ATTEMPTS))
    emit(run="true", attempt=attempt, findings=",".join(ids), branch=head.get("ref"), head_sha=head.get("sha"))
    return 0


def batch_branch(ctx, gh, repo, day):
    taken = {r.get("ref", "") for page in [ctx.gh_api(gh, "repos/%s/git/matching-refs/heads/%s%s" % (repo, BATCH_PREFIX, day))] for r in page}
    for n in range(1, 10):
        name = "%s%s%s" % (BATCH_PREFIX, day, "" if n == 1 else "-%d" % n)
        if "refs/heads/" + name not in taken:
            return name
    return None


def decide_batch(ctx, gh, repo):
    findings = load_findings_module()
    issues = [i for i in ctx.gh_api(gh, "repos/%s/issues?labels=%s&state=open&per_page=100" % (repo, findings.BATCH_LABEL))
              if "pull_request" not in i and i.get("state") == "open" and findings.BATCH_MARK in (i.get("body") or "")]
    if not issues:
        return "nothing", "no open review batch issue"
    issue = min(issues, key=lambda i: i.get("number", 0))
    for pr in ctx.gh_api(gh, "repos/%s/pulls?state=open&base=develop&per_page=100" % repo):
        if ((pr.get("head") or {}).get("ref") or "").startswith(BATCH_PREFIX):
            return "nothing", "a batch pull request is already open (#%s)" % pr.get("number")
    items = []
    for line in (issue.get("body") or "").splitlines():
        m = findings.BATCH_LINE.match(line)
        if m and m.group(1) == " ":
            items.append(m.group(3))
    if not items:
        return "nothing", "every item of batch #%s is ticked" % issue.get("number")
    day = os.environ.get("TODAY") or datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m-%d")
    if not re.match(r"^\d{4}-\d{2}-\d{2}$", day):
        return "nothing", "TODAY is not a date"
    branch = batch_branch(ctx, gh, repo, day)
    if not branch:
        return "nothing", "every batch branch name of %s is taken" % day
    return "fix", (issue.get("number"), branch, items)


def load_findings_module():
    path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "findings.py")
    spec = importlib.util.spec_from_file_location("findings", path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def cmd_batch():
    repo, gh = os.environ.get("REPO", ""), os.environ.get("GH_BIN", "gh")
    mode, token = mode_and_token()
    if mode == "off":
        return stop("autofix batch: off (AUTO_FIX=off)")
    result = os.environ.get("BATCH_RESULT", "")
    if os.environ.get("FORCE") != "true" and not re.search(r"\b(created|updated)\b", result):
        return stop("autofix batch: skip (the batch was not created or updated in this run)")
    if not repo:
        return stop("autofix batch: NOT RUN (REPO is required)")
    try:
        verdict, detail = decide_batch(load_context_module(), gh, repo)
    except Exception as exc:  # fail open
        return stop("autofix batch: NOT RUN (%s)" % type(exc).__name__)
    if verdict != "fix":
        return stop("autofix batch: skip (%s)" % detail)
    if mode == "on" and not token:
        return stop("autofix: NOT RUN (REVIEW_AGENT_TOKEN missing)")
    number, branch, items = detail
    if mode != "on":
        return stop("autofix batch: report mode: would fix %d item(s) of #%s on %s" % (len(items), number, branch))
    print("autofix batch: fixing %d item(s) of #%s on %s" % (len(items), number, branch))
    emit(run="true", issue=number, branch=branch, items=",".join(items))
    return 0


def git_out(git, worktree, *args):
    done = subprocess.run([git, "-C", worktree, *args], capture_output=True, text=True, timeout=60, check=False)
    if done.returncode != 0:
        raise RuntimeError("git %s failed" % args[0])
    return done.stdout


def forbidden(path):
    name = path.rsplit("/", 1)[-1]
    if not (path in ALLOWED_EXACT or path.startswith(ALLOWED_PREFIX)) or OWNER_IN_SRC.match(path):
        return True
    return (path in FORBIDDEN_EXACT or path.startswith(FORBIDDEN_PREFIX) or name.startswith(".env") or name in FORBIDDEN_NAMES
            or any(fnmatch.fnmatchcase(path, g) for g in FORBIDDEN_GLOB)
            or (("/" not in path) and name.startswith((".review-", ".autofix-"))))


def check_commits(git, worktree, base, name, email):
    """None when the commits after base may be pushed, else the reason."""
    rng = "%s..HEAD" % base
    raw = git_out(git, worktree, "log", "--format=%H%x00%an%x00%ae%x00%cn%x00%ce%x00%B%x1e", rng)
    commits = [c.strip("\n") for c in raw.split("\x1e") if c.strip()]
    if not commits:
        return "no commit after the base: nothing to push"
    for c in commits:
        sha, an, ae, cn, ce, body = (c.split("\x00") + [""] * 6)[:6]
        if an != name or cn != name or (email and (ae != email or ce != email)):
            return "commit %s has another identity than %s (author %s, committer %s)" % (sha[:12], name, an, cn)
        if TRAILER.search(body):
            return "commit %s has an AI trailer or a Generated with line" % sha[:12]
    # NUL-separated and without rename detection: both sides of a move are listed, and no path comes back quoted.
    tokens = git_out(git, worktree, "diff", "--raw", "-z", "--no-renames", "--no-abbrev", base, "HEAD").split("\x00")
    paths, i = [], 0
    while i < len(tokens):
        head = tokens[i]
        if not head.startswith(":"):
            i += 1
            continue
        old_mode, new_mode = head[1:].split()[:2]
        path = tokens[i + 1] if i + 1 < len(tokens) else ""
        i += 2
        paths.append(path)
        if "120000" in (old_mode, new_mode):
            return "a symlink is changed: %s" % path
        if "160000" in (old_mode, new_mode):
            return "a submodule is changed: %s" % path
        if old_mode == "000000" and new_mode == "100755" and not path.startswith("scripts/"):
            return "a new executable file outside scripts/: %s" % path
        if "000000" not in (old_mode, new_mode) and old_mode != new_mode:
            return "a file mode change: %s (%s to %s)" % (path, old_mode, new_mode)
    for entry in git_out(git, worktree, "diff", "--numstat", "-z", "--no-renames", base, "HEAD").split("\x00"):
        if entry.startswith("-\t-\t"):
            return "a binary change: %s" % entry[4:]
    bad = [p for p in paths if p and forbidden(p)]
    if bad:
        return "owner paths changed: %s" % ", ".join(bad[:10])
    return None


def cmd_check():
    git, worktree = os.environ.get("GIT_BIN", "git"), os.environ.get("WORKTREE", ".")
    base, name = os.environ.get("BASE_SHA", ""), os.environ.get("EXPECT_NAME", "")
    if not re.match(r"^[0-9a-f]{40}$", base) or not name:
        print("autofix check: FAILED (BASE_SHA and EXPECT_NAME are required)")
        return 1
    try:
        reason = check_commits(git, worktree, base, name, os.environ.get("EXPECT_EMAIL", ""))
    except Exception as exc:  # fail closed: nothing is pushed on a doubt
        reason = "git could not be read (%s)" % exc
    if reason:
        print("autofix check: FAILED (%s)" % reason)
        return 1
    count = git_out(git, worktree, "rev-list", "--count", "%s..HEAD" % base).strip()
    print("autofix check: PASS (%s commit(s), human identity, no owner path)" % count)
    return 0


def cmd_apply():
    """Validate the patch files, then apply and commit each one here, as the human identity (never the patch's From:)."""
    git, worktree = os.environ.get("GIT_BIN", "git"), os.environ.get("WORKTREE", ".")
    folder, name, email = os.environ.get("PATCH_DIR", ""), os.environ.get("EXPECT_NAME", ""), os.environ.get("EXPECT_EMAIL", "")
    try:
        files = sorted(f for f in os.listdir(folder) if os.path.isfile(os.path.join(folder, f)))
    except OSError:
        files = []
    reason = None
    if not name or not email:
        reason = "EXPECT_NAME and EXPECT_EMAIL are required"
    elif not files:
        reason = "no patch file"
    elif len(files) > MAX_PATCHES:
        reason = "too many patch files (%d, at most %d)" % (len(files), MAX_PATCHES)
    elif any(not PATCH_NAME.match(f) for f in files):
        reason = "a patch file name is not NNNN-slug.patch"
    elif sum(os.path.getsize(os.path.join(folder, f)) for f in files) > MAX_PATCH_BYTES:
        reason = "the patches are too large (more than %d bytes)" % MAX_PATCH_BYTES
    if reason:
        print("autofix apply: FAILED (%s)" % reason)
        return 1
    start = git_out(git, worktree, "rev-parse", "HEAD").strip()
    env = dict(os.environ, GIT_AUTHOR_NAME=name, GIT_AUTHOR_EMAIL=email, GIT_COMMITTER_NAME=name, GIT_COMMITTER_EMAIL=email)
    quiet = ["-c", "core.hooksPath=/dev/null"]
    try:
        for f in files:
            msg, diff = os.path.join(folder, ".msg"), os.path.join(folder, ".diff")
            with open(os.path.join(folder, f), "rb") as fh:
                info = subprocess.run([git, "-C", worktree, "mailinfo", msg, diff], stdin=fh, capture_output=True, timeout=60, check=True)
            subject = next((l[9:] for l in info.stdout.decode("utf-8", "replace").splitlines() if l.startswith("Subject: ")), "")
            with open(msg, encoding="utf-8", errors="replace") as fh:
                body = fh.read().strip()
            if not subject:
                raise RuntimeError("%s has no subject" % f)
            subprocess.run([git, "-C", worktree, *quiet, "apply", "--index", diff], capture_output=True, timeout=60, check=True)
            subprocess.run([git, "-C", worktree, *quiet, "commit", "-q", "--no-verify", "-m", subject + ("\n\n" + body if body else "")],
                           env=env, capture_output=True, timeout=60, check=True)
            for tmp in (msg, diff):
                os.remove(tmp)
    except (subprocess.CalledProcessError, subprocess.TimeoutExpired, OSError, RuntimeError) as exc:
        subprocess.run([git, "-C", worktree, "reset", "-q", "--hard", start], capture_output=True, timeout=60, check=False)
        print("autofix apply: FAILED (%s does not apply: %s); nothing was kept" % (f, type(exc).__name__))
        return 1
    print("autofix apply: PASS (%d patch(es) committed as %s)" % (len(files), name))
    return 0


def cmd_prepare():
    """Right before the model starts (after make setup, which runs pull request code): wipe the Claude configuration
    in HOME (~/.claude, ~/.claude.json, ~/.config/claude*) and restore the project's Claude configuration paths in
    WORKTREE from SOURCE_REF, in the working tree only (nothing staged). A path the ref does not have is removed."""
    import shutil
    git, worktree, ref = os.environ.get("GIT_BIN", "git"), os.environ.get("WORKTREE", "."), os.environ.get("SOURCE_REF", "")
    home = os.path.abspath(os.environ.get("HOME", "") or "/")
    if home in ("/", "") or not ref:
        print("autofix prepare: FAILED (a HOME other than / and SOURCE_REF are required)")
        return 1

    def remove(path):
        if os.path.islink(path) or os.path.isfile(path):
            os.remove(path)
        elif os.path.isdir(path):
            shutil.rmtree(path)

    config = os.path.join(home, ".config")
    targets = [os.path.join(home, n) for n in HOME_CONFIG]
    if os.path.isdir(config):
        targets += [os.path.join(config, n) for n in os.listdir(config) if n.startswith("claude")]
    for path in targets:
        remove(path)
    for path in CLAUDE_CONFIG:
        remove(os.path.join(worktree, path))
        exists = subprocess.run([git, "-C", worktree, "cat-file", "-e", "%s:%s" % (ref, path)], capture_output=True, timeout=60).returncode == 0
        if exists:
            subprocess.run([git, "-C", worktree, "restore", "--source=%s" % ref, "--worktree", "--", path],
                           capture_output=True, timeout=60, check=True)
    print("autofix prepare: home Claude configuration wiped; project configuration restored from %s" % ref)
    return 0


def cell(text, width=300):
    text = re.sub(r"@(?=[A-Za-z0-9_-])", "@" + chr(0x200B), " ".join(str(text or "").split()))
    # A reference written by the model never links or closes anything: `fixes #5` becomes `fixes #\u200b5`.
    text = re.sub(r"#(?=\d)", "#" + chr(0x200B), text)
    text = text.replace("|", "\\|")
    return text if len(text) <= width else text[: width - 1] + "..."


def cmd_body():
    issue = re.sub(r"\D", "", os.environ.get("ISSUE", ""))
    # The items come from the plan (ITEMS, keys read from the batch issue by script), never from the model: every
    # source is listed, and a result for anything else is dropped.
    items = [k for k in os.environ.get("ITEMS", "").split(",") if ITEM_KEY.match(k)]
    sources = []
    for k in items:
        if k.split("/", 1)[0] not in sources:
            sources.append(k.split("/", 1)[0])
    out = ["Fix pull request for the review batch #%s (one batch, one pull request). When it is merged into `develop`, "
           "`scripts/findings.py settle` settles each source below from this table, then closes the batch." % issue, "",
           "Batch: #%s" % issue, "Sources: %s" % " ".join(sources), ""]
    try:
        data = json.loads(os.environ.get("RESULT_JSON", ""))
        rows = data.get("results") if isinstance(data, dict) else None
        if not isinstance(rows, list):
            raise ValueError("no results")
    except ValueError:
        data, rows = {"outcome": "nothing", "reason": "no result from the fixer could be read"}, []
    found = {}
    for r in rows:
        if isinstance(r, dict) and str(r.get("item", "")).strip() in items:
            found.setdefault(str(r["item"]).strip(), r)
    out += ["| Item | Source | Decision | Reason |", "|---|---|---|---|"]
    for key in items:
        r = found.get(key, {"verdict": "OWNER", "reason": "no result from the fixer"})
        verdict = " ".join(str(r.get("verdict", "")).upper().split())
        out.append("| %s | %s | %s | %s |" % (key, key.split("/", 1)[0], verdict if verdict in VERDICTS else "OWNER", cell(r.get("reason"))))
    out += ["", "Outcome: %s. %s" % (cell(data.get("outcome"), 20), cell(data.get("reason"))), "",
            "Written by the CI fixer (`.claude/commands/review-fix.md`). Humans still close or tick each source with the "
            "reason; this pull request goes through review like any other."]
    print("\n".join(out))
    return 0


def git_args_ok(sub, args):
    allowed = GIT_OPTIONS[sub]
    return all((a in allowed or GIT_NUMERIC.match(a)) if a.startswith("-") else True for a in args)


def go_args_ok(args):
    i, packages = 0, 0
    while i < len(args):
        a = args[i]
        name, eq, value = a.partition("=")
        if name in GO_VALUE_FLAGS:
            if not eq:
                if i + 1 >= len(args):
                    return False
                value = args[i + 1]
                i += 1
            if not GO_VALUE_FLAGS[name].match(value):
                return False
        elif a in GO_BOOL_FLAGS:
            pass
        elif GO_PACKAGE.match(a):
            packages += 1
        else:
            return False
        i += 1
    return True


def wrapped(argv):
    """The argv to execute for a fixer command, or None when it is not allowed (allowlists only)."""
    if not argv:
        return None
    if argv[0] == "git" and len(argv) > 1 and argv[1] in GIT_OPTIONS and git_args_ok(argv[1], argv[2:]):
        return ["git", "--no-pager"] + argv[1:]
    if argv[:2] == ["go", "test"] and go_args_ok(argv[2:]):
        return argv
    if (argv[:4] == ["pnpm", "--filter", "web", "run"] and len(argv) > 4 and argv[4] in WEB_SCRIPTS
            and all(a == "--" or not a.startswith("-") for a in argv[5:])):
        return argv
    return None


def cmd_run(argv):
    """`run CMD ARGS...`: the fixer's only way to run git (read-only), go test and the web scripts. Refused: any other
    command, and any argument that writes a file or runs a program (-o, --output, -exec, -coverprofile, ...)."""
    target = wrapped(argv)
    if target is None:
        print("fixer-run: refused (allowed: git status|diff|log|show with a few read options, go test with -run, -count, "
              "-race, -tags, -v, -timeout and ./ packages, pnpm --filter web run test|lint|typecheck|format:check with "
              "paths): %s" % " ".join(argv[:6]))
        return 2
    if os.environ.get("FIXER_NETNS") == "1" and target[0] in ("go", "pnpm"):
        # No network for code the pull request controls; modules and packages were installed before.
        os.environ.update({"GOPROXY": "off", "GOTOOLCHAIN": "local"})
        target = ["unshare", "-rn"] + target
    os.execvp(target[0], target)
    return 0  # not reached


def main():
    sub = sys.argv[1] if len(sys.argv) > 1 else ""
    if sub == "run":
        return cmd_run(sys.argv[2:])
    commands = {"pr": cmd_pr, "batch": cmd_batch, "check": cmd_check, "body": cmd_body, "apply": cmd_apply, "prepare": cmd_prepare}
    if sub not in commands:
        print("usage: review-autofix.py pr|batch|prepare|apply|check|body|run")
        return 2
    return commands[sub]()


if __name__ == "__main__":
    sys.exit(main())
