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
  check  the guard before any push, run from a trusted copy: at least one new commit since BASE_SHA; every author and
         committer is EXPECT_NAME <EXPECT_EMAIL>; no Co-authored-by trailer and no "Generated with" line; no changed
         path the owner keeps (api/openapi.yaml R7, db/migrations R6, ADRs, secrets, .env files, .github, the Claude
         harness, the review workspace files). Exit 1 with the reason on a failure.
  body   the batch pull request body from the fixer's structured result (RESULT_JSON): it names the batch issue
         (#ISSUE, which closes the batch on merge) and has one table row per item, verdict FIXED, NOT A DEFECT,
         DECIDED, OWNER or AFTER FREEZE with a reason. Model text is data: @-mentions are neutralised, pipes escaped.

Mode: AUTO_FIX = off | report (default: says what it would do, writes nothing) | on. HAS_TOKEN must be `true` (the
workflow passes `secrets.REVIEW_BOT_TOKEN != ''`): without it, "autofix: NOT RUN (REVIEW_BOT_TOKEN missing)" and
nothing is written. Only `on` with the token writes, and only labels.
Outputs (`pr`, `batch`) are `key=value` lines appended to GITHUB_OUTPUT: run=true|false, then attempt, findings,
branch, head_sha (pr) or issue, branch, items (batch). They carry ids and numbers only, never ledger or issue text.
Fail OPEN for pr and batch (a GitHub error is NOT RUN with run=false, exit 0); fail CLOSED for check.
Environment: REPO, PR_NUMBER, RUN_ID, RUN_HEAD_SHA, BATCH_RESULT, FORCE, TODAY (YYYY-MM-DD), BASE_SHA, WORKTREE,
EXPECT_NAME, EXPECT_EMAIL, ISSUE, RESULT_JSON, GH_BIN (default gh), GIT_BIN (default git).
"""
import datetime
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
VERDICTS = ("FIXED", "NOT A DEFECT", "DECIDED", "OWNER", "AFTER FREEZE")
# Paths the fixer never changes: the contract (R7), migrations (R6), decisions, secrets, CI, the harness itself and the
# files the workflows leave in the workspace.
FORBIDDEN_EXACT = ("api/openapi.yaml", "CLAUDE.md", "AGENTS.md")
FORBIDDEN_PREFIX = ("db/migrations/", "docs/decisions/", "secrets/", ".github/", ".claude/")
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
        return stop("autofix: NOT RUN (REVIEW_BOT_TOKEN missing)")
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
        return stop("autofix: NOT RUN (REVIEW_BOT_TOKEN missing)")
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
    return (path in FORBIDDEN_EXACT or path.startswith(FORBIDDEN_PREFIX) or name.startswith(".env")
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
    bad = [p for p in git_out(git, worktree, "diff", "--name-only", base, "HEAD").split("\n") if p and forbidden(p)]
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


def cell(text, width=300):
    text = re.sub(r"@(?=[A-Za-z0-9_-])", "@" + chr(0x200B), " ".join(str(text or "").split()))
    text = text.replace("|", "\\|")
    return text if len(text) <= width else text[: width - 1] + "..."


def cmd_body():
    issue = re.sub(r"\D", "", os.environ.get("ISSUE", ""))
    out = ["Fix pull request for the review batch #%s (one batch, one pull request; it closes the batch when it is "
           "merged into `develop`)." % issue, ""]
    try:
        data = json.loads(os.environ.get("RESULT_JSON", ""))
        rows = data.get("results") if isinstance(data, dict) else None
        if not isinstance(rows, list):
            raise ValueError("no results")
    except ValueError:
        out += ["No result from the fixer could be read: check every item of #%s by hand." % issue]
        print("\n".join(out))
        return 0
    out += ["| Item | Result | Reason |", "|---|---|---|"]
    for r in rows[:40]:
        if not isinstance(r, dict):
            continue
        verdict = str(r.get("verdict", "")).upper().strip()
        out.append("| %s | %s | %s |" % (cell(r.get("item"), 40), verdict if verdict in VERDICTS else "OWNER", cell(r.get("reason"))))
    out += ["", "Outcome: %s. %s" % (cell(data.get("outcome"), 20), cell(data.get("reason"))), "",
            "Written by the CI fixer (`.claude/commands/review-fix.md`). Humans still close or tick each source with the "
            "reason; this pull request goes through review like any other."]
    print("\n".join(out))
    return 0


def main():
    sub = sys.argv[1] if len(sys.argv) > 1 else ""
    commands = {"pr": cmd_pr, "batch": cmd_batch, "check": cmd_check, "body": cmd_body}
    if sub not in commands:
        print("usage: review-autofix.py pr|batch|check|body")
        return 2
    return commands[sub]()


if __name__ == "__main__":
    sys.exit(main())
