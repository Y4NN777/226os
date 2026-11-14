#!/usr/bin/env python3
"""Review gate (stdlib only): the one required check that lets the AI review stop a CRITICAL change and lets the
owner review only the sensitive paths, without making everyone wait for a review.

It checks two things on a pull request:
  1. Triage verdict. The review comment of the Actions bot (the one with the marker `<!-- 226os-ai-review -->`, else
     the most recently updated one) whose first line is exactly
     `Triage: <CRITICAL|HIGH|MEDIUM|NONE> | blast radius: <phrase>` is the verdict. Only a CRITICAL that cites OBSERVED
     evidence (`file:line`) can fail the gate; a CRITICAL without it is downgraded. The label `owner-override`
     cancels a CRITICAL only when the owner added it (the last labeled/unlabeled event decides). Comments by humans
     are ignored.
  2. Sensitive paths. If the pull request touches a path listed in `.github/sensitive-paths.txt` (read from the BASE
     branch, so a pull request cannot edit the list to escape it), the owner must have approved it (the owner's
     latest review state is APPROVED) unless the owner wrote the pull request.

Fail OPEN: a missing verdict, an unreadable GitHub answer or a missing list is NOT RUN, never a failure. An outage of
the review is not a verdict.

Mode (GATE_MODE): `off` does nothing, `report` (default) prints WOULD FAIL and exits 0, `enforce` exits 1 on a failure.
The verdict is decided from the LEDGER of the comment (scripts/review-comment.py effective_level), not from its prose, and
only when the ledger sha is PR_HEAD_SHA: a stale ledger is NOT RUN, except that a stale CRITICAL keeps blocking. A file list
that does not match the pull request's own count (3000 files) or a pull request with more than 250 commits fails CLOSED;
renames count with their old path.

Environment: REPO (owner/name), PR_NUMBER, PR_HEAD_SHA, PR_AUTHOR, OWNER (default Y4NN777), BASE_REF (default develop), GH_BIN (default gh,
authenticated through GH_TOKEN). Exit codes: 0 pass, 1 failure in enforce mode.
"""
import base64
import fnmatch
import importlib.util
import json
import os
import re
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
BOT_LOGINS = {"github-actions[bot]", "github-actions"}
MARKER = "<!-- 226os-ai-review -->"
VERDICT = re.compile(r"^Triage:\s*(CRITICAL|HIGH|MEDIUM|NONE)\s*\|\s*blast radius:\s*\S.*$")
EVIDENCE = re.compile(r"[\w./-]+\.\w+:\d+")
OVERRIDE_LABEL = "owner-override"
LIST_PATH = ".github/sensitive-paths.txt"
STRAY_LIST = ".github/stray-paths.txt"
STRAY_OK_LABEL = "stray-ok"
NO_TEST_LABEL = "no-test"
FINDING_LABEL = "finding"
# GitHub's closing keywords, with an optional colon ("Fixes: #12"); a negation before the keyword ("does not close #12",
# "doesn't fix #12", "never closes #12") is not one.
CLOSES = re.compile(r"(?<!not )(?<!n't )(?<!never )\b(?:close[sd]?|fix(?:e[sd])?|resolve[sd]?):?\s+#(\d+)", re.I)
TEST_FILE = re.compile(r"(_test\.go$|\.test\.tsx?$|^scripts/tests/)")


def load(name, filename):
    spec = importlib.util.spec_from_file_location(name, os.path.join(HERE, filename))
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


# The ledger rules (blocking, effective_level, changed_paths) are the ones the review comment is written with: one source.
comment = load("review_comment", "review-comment.py")
FULL_SHA = re.compile(r"^[0-9a-f]{40}$")
MAX_COMMITS = 250  # GitHub lists at most 250 commits of a pull request
MAX_FILES = 3000   # and at most 3000 files


class GitHubError(Exception):
    """GitHub could not be read: the gate fails open."""


def parse_pages(text):
    """Concatenated JSON documents (gh --paginate) to one list."""
    decoder = json.JSONDecoder()
    items, i = [], 0
    text = text.strip()
    while i < len(text):
        obj, end = decoder.raw_decode(text, i)
        items.extend(obj if isinstance(obj, list) else [obj])
        i = end
        while i < len(text) and text[i].isspace():
            i += 1
    return items


def gh_api(gh, endpoint):
    try:
        done = subprocess.run([gh, "api", endpoint, "--paginate"], capture_output=True, text=True, timeout=60, check=False)
    except (OSError, subprocess.SubprocessError) as exc:
        raise GitHubError(str(exc)) from exc
    if done.returncode != 0:
        raise GitHubError("gh api %s failed" % endpoint.split("?")[0])
    try:
        return parse_pages(done.stdout)
    except ValueError as exc:
        raise GitHubError("unreadable answer for %s" % endpoint) from exc


def latest_verdict(comments):
    """(level, body) of the review comment's verdict, or None. The bot comment with the marker MARKER (the one review
    comment, scripts/review-comment.py) wins; without one (an older pull request), the most recently UPDATED bot
    verdict, so a stale duplicate created later never wins. When marker comments exist (one per run), the newest
    created one is the only candidate: its
    `Triage: NOT RUN` (the review gave no result) is no verdict, and an older duplicate is never used instead."""
    bots = [c for c in comments if (c.get("user") or {}).get("login") in BOT_LOGINS]
    marked = [c for c in bots if MARKER in (c.get("body") or "")]
    if marked:
        # One comment per run: the newest CREATED marker comment is the only candidate (older runs are minimized).
        newest = max(marked, key=lambda c: (c.get("created_at") or c.get("updated_at") or "", c.get("id") or 0))
        m = VERDICT.match((newest.get("body") or "").strip().splitlines()[0] if (newest.get("body") or "").strip() else "")
        return (m.group(1), newest.get("body") or "") if m else None
    best = None
    for c in bots:
        body = c.get("body") or ""
        first = body.strip().splitlines()[0] if body.strip() else ""
        m = VERDICT.match(first)
        if m:
            key = (MARKER in body, c.get("updated_at") or c.get("created_at") or "")
            if best is None or key >= best[0]:
                best = (key, m.group(1), body)
    return None if best is None else (best[1], best[2])


def owner_override(events, owner):
    last = None
    for e in events:
        if e.get("event") in ("labeled", "unlabeled") and (e.get("label") or {}).get("name") == OVERRIDE_LABEL:
            if last is None or e.get("created_at", "") >= last.get("created_at", ""):
                last = e
    return last is not None and last.get("event") == "labeled" and (last.get("actor") or {}).get("login") == owner


def check_verdict(gh, repo, pr, owner, head=""):
    """(failure message or None, note). The level is the one the head's LEDGER supports (comment.effective_level: a
    CRITICAL needs a blocking OBSERVED CRITICAL entry with a file and a line), never the prose of the comment. A ledger
    that does not cover the head is stale: NOT RUN, except that a stale CRITICAL keeps blocking (a push never lifts it)."""
    comments = gh_api(gh, "repos/%s/issues/%s/comments" % (repo, pr))
    found = latest_verdict(comments)
    if found is None:
        return None, "triage verdict: NOT RUN (no review verdict on this pull request)"
    declared, body = found
    ledger = comment.ledger_of(body)
    level = comment.effective_level(declared, ledger["findings"] if ledger else [])
    stale = ledger is not None and (not FULL_SHA.match(head or "") or ledger.get("sha") != head)
    if level != "CRITICAL":
        if stale:
            return None, "triage verdict: NOT RUN (stale: the ledger covers %s, the head is %s)" % (str(ledger.get("sha"))[:12] or "(none)", (head or "unknown")[:12])
        if declared == "CRITICAL":
            return None, "triage verdict CRITICAL downgraded: no OBSERVED CRITICAL entry with file:line in the ledger, treated as HIGH"
        return None, "triage verdict %s: never blocks" % level
    if owner_override(gh_api(gh, "repos/%s/issues/%s/events" % (repo, pr)), owner):
        return None, "triage verdict CRITICAL cancelled by an owner override (label %s added by %s)" % (OVERRIDE_LABEL, owner)
    why = " (stale ledger: it still blocks until a review of the head lifts it)" if stale else ""
    return "CRITICAL review verdict with OBSERVED evidence%s (add the label %s to override, owner only)" % (why, OVERRIDE_LABEL), ""


def sensitive_patterns(gh, repo, base):
    try:
        items = gh_api(gh, "repos/%s/contents/%s?ref=%s" % (repo, LIST_PATH, base))
        raw = base64.b64decode(items[0]["content"]).decode("utf-8")
    except (GitHubError, KeyError, IndexError, ValueError):
        return None
    return [ln.strip() for ln in raw.splitlines() if ln.strip() and not ln.lstrip().startswith("#")]


def owner_state(reviews, owner):
    last = None
    for r in reviews:
        if (r.get("user") or {}).get("login") == owner and r.get("state") in ("APPROVED", "CHANGES_REQUESTED", "DISMISSED"):
            if last is None or r.get("submitted_at", "") >= last.get("submitted_at", ""):
                last = r
    return None if last is None else last.get("state")


def pr_files(gh, repo, pr, pull):
    """(items, problem): the pull request's file list, or a reason it cannot be trusted (fail CLOSED: a list that does not
    match the pull request's own count, 3000 files, or more than 250 commits hides paths from every path rule)."""
    items = gh_api(gh, "repos/%s/pulls/%s/files" % (repo, pr))
    try:
        count, commits = int(pull.get("changed_files")), int(pull.get("commits") or 0)
    except (TypeError, ValueError):
        return None, "the pull request's file count is unreadable: its path rules cannot be checked"
    if commits > MAX_COMMITS:
        return None, "the pull request has %d commits (more than %d): its path rules cannot be checked" % (commits, MAX_COMMITS)
    if count != len(items) or count >= MAX_FILES:
        return None, "the file list is incomplete (%d of %d files): its path rules cannot be checked" % (len(items), count)
    return items, None


PIPELINE_BRANCH = "fix/review/"
PIPELINE_TRAILER = re.compile(r"^Review-Pipeline:", re.M | re.I)


def owner_authored(gh, repo, number, pull, owner):
    """True only when the owner opened the pull request AND authored every commit on it, each one VERIFIED (a signature
    GitHub checks: an author email is only a claim, anyone can forge it). The review pipeline commits and pushes with the
    owner's identity and token: a pull request it made (a `fix/review/` branch, an autofix label or any commit with the
    `Review-Pipeline:` trailer) is never owner-authored. More than 250 commits cannot be listed: not owner-authored. A
    forged trailer, branch or label only makes this stricter."""
    if (pull.get("user") or {}).get("login") != owner:
        return False
    if str((pull.get("head") or {}).get("ref") or "").startswith(PIPELINE_BRANCH) or any(n.startswith("autofix-") for n in label_names(pull)):
        return False
    try:
        if int(pull.get("commits") or 0) > MAX_COMMITS:
            return False
    except (TypeError, ValueError):
        return False
    commits = gh_api(gh, "repos/%s/pulls/%s/commits" % (repo, number))
    if not commits or len(commits) > MAX_COMMITS:
        return False
    for c in commits:
        info = c.get("commit") or {}
        if PIPELINE_TRAILER.search(str(info.get("message") or "")):
            return False
        if (info.get("verification") or {}).get("verified") is not True or (c.get("author") or {}).get("login") != owner:
            return False
    return True


def check_sensitive(gh, repo, pr, author, owner, base):
    patterns = sensitive_patterns(gh, repo, base)
    if patterns is None:
        return None, "sensitive paths: NOT RUN (no sensitive-paths list on %s)" % base
    pull = fetch_pull(gh, repo, pr)
    items, problem = pr_files(gh, repo, pr, pull)
    if problem:
        return "sensitive paths: " + problem, ""
    files = comment.changed_paths(items)  # a rename out of a sensitive path touches it
    hits = sorted(f for f in files if any(fnmatch.fnmatchcase(f, p) for p in patterns))
    if not hits:
        return None, "sensitive paths: no sensitive path touched"
    shown = ", ".join(hits[:6]) + (" ..." if len(hits) > 6 else "")
    if owner_authored(gh, repo, pr, pull, owner):
        return None, "sensitive paths touched (%s): owner-authored (every commit verified)" % shown
    if owner_state(gh_api(gh, "repos/%s/pulls/%s/reviews" % (repo, pr)), owner) == "APPROVED":
        return None, "sensitive paths touched (%s): approved by the owner" % shown
    return "sensitive paths touched (%s) without the owner's approval" % shown, ""


def label_names(obj):
    return [x if isinstance(x, str) else (x or {}).get("name", "") for x in obj.get("labels") or []]


def fetch_pull(gh, repo, pr):
    items = gh_api(gh, "repos/%s/pulls/%s" % (repo, pr))
    return items[0] if items else {}


def stray_rules(gh, repo, base):
    """({'always': [...], 'feature': [...]}, or None when the list is missing). Lines are `always: PATTERN` or
    `feature: PATTERN` (the latter only on feature/* and fix/* branches); read from the BASE branch."""
    try:
        items = gh_api(gh, "repos/%s/contents/%s?ref=%s" % (repo, STRAY_LIST, base))
        raw = base64.b64decode(items[0]["content"]).decode("utf-8")
    except (GitHubError, KeyError, IndexError, ValueError):
        return None
    rules = {"always": [], "feature": []}
    for line in raw.splitlines():
        line = line.strip()
        if not line or line.startswith("#") or ":" not in line:
            continue
        when, pattern = (x.strip() for x in line.split(":", 1))
        if when in rules and pattern:
            rules[when].append(pattern)
    return rules


def check_stray(gh, repo, base, pull):
    """(failure or None, note): files that do not belong to a product pull request (agent or editor directories, a
    .gitignore edit on a feature or fix branch). The label stray-ok exempts."""
    rules = stray_rules(gh, repo, base)
    if rules is None:
        return None, "stray paths: NOT RUN (no stray-paths list on %s)" % base
    if STRAY_OK_LABEL in label_names(pull):
        return None, "stray paths: exempted by the label %s" % STRAY_OK_LABEL
    ref = (pull.get("head") or {}).get("ref", "")
    patterns = rules["always"] + (rules["feature"] if ref.startswith(("feature/", "fix/")) else [])
    items, problem = pr_files(gh, repo, pr_number(pull), pull)
    if problem:
        return "stray paths: " + problem, ""
    files = comment.changed_paths(items)
    hits = sorted(f for f in files if any(fnmatch.fnmatchcase(f, p) for p in patterns))
    if not hits:
        return None, "stray paths: none"
    shown = ", ".join(hits[:6]) + (" ..." if len(hits) > 6 else "")
    return "stray paths touched (%s): remove them from this pull request, or add the label %s" % (shown, STRAY_OK_LABEL), ""


def pr_number(pull):
    return pull.get("number") or os.environ.get("PR_NUMBER", "")


def check_finding_tests(gh, repo, pull):
    """(failure or None, note): a pull request that closes a `finding` issue must change a test file, so the finding
    becomes a permanent check. The label no-test exempts."""
    body = pull.get("body") or ""
    numbers = sorted(set(CLOSES.findall(body)))
    closed = []
    for n in numbers:
        issues = gh_api(gh, "repos/%s/issues/%s" % (repo, n))
        if issues and FINDING_LABEL in label_names(issues[0]):
            closed.append(n)
    if not closed:
        return None, "finding to test: no finding issue is closed by this pull request"
    if NO_TEST_LABEL in label_names(pull):
        return None, "finding to test: exempted by the label %s" % NO_TEST_LABEL
    items, problem = pr_files(gh, repo, pr_number(pull), pull)
    if problem:
        return "finding to test: " + problem, ""
    if any(TEST_FILE.search(f.get("filename", "")) for f in items):
        return None, "finding to test: a test file changed"
    return ("closes finding issue(s) %s without changing a test file: add a test that fails without the fix, or add the "
            "label %s" % (", ".join("#" + n for n in closed), NO_TEST_LABEL)), ""


def main():
    mode = os.environ.get("GATE_MODE", "report").strip().lower() or "report"
    if mode not in ("off", "report", "enforce"):
        print("review-gate: unknown GATE_MODE %r, using report" % mode)
        mode = "report"
    if mode == "off":
        print("review-gate: off (CODE_REVIEW_GATE=off): nothing checked")
        return 0
    repo, pr = os.environ.get("REPO", ""), os.environ.get("PR_NUMBER", "")
    if not repo or not pr:
        print("review-gate: NOT RUN (REPO and PR_NUMBER are required)")
        return 0
    owner = os.environ.get("OWNER", "Y4NN777")
    author = os.environ.get("PR_AUTHOR", "")
    base = os.environ.get("BASE_REF", "develop")
    gh = os.environ.get("GH_BIN", "gh")
    head = os.environ.get("PR_HEAD_SHA", "")

    failures = []
    for name, check in (("verdict", lambda: check_verdict(gh, repo, pr, owner, head)),
                        ("sensitive", lambda: check_sensitive(gh, repo, pr, author, owner, base))):
        try:
            failure, note = check()
        except GitHubError as exc:
            print("review-gate: %s check NOT RUN (%s)" % (name, exc))
            continue
        if note:
            print("review-gate: " + note)
        if failure:
            failures.append(failure)

    # The two newer checks have their own switch: report only until the owner sets GATE_NEW=enforce (promotion after 10
    # pull requests without a false positive), whatever GATE_MODE says. They never fail the gate in report mode.
    new_mode = os.environ.get("GATE_NEW", "report").strip().lower() or "report"
    if new_mode not in ("off", "report", "enforce"):
        new_mode = "report"
    new_failures = []
    if new_mode != "off":
        try:
            pull = fetch_pull(gh, repo, pr)
            pull.setdefault("number", pr)
            new_checks = (("stray", lambda: check_stray(gh, repo, base, pull)),
                          ("finding-to-test", lambda: check_finding_tests(gh, repo, pull)))
        except GitHubError as exc:
            print("review-gate: new checks NOT RUN (%s)" % exc)
            new_checks = ()
        for name, check in new_checks:
            try:
                failure, note = check()
            except GitHubError as exc:
                print("review-gate: %s check NOT RUN (%s)" % (name, exc))
                continue
            if note:
                print("review-gate: " + note)
            if failure and new_mode == "enforce":
                new_failures.append(failure)
            elif failure:
                print("review-gate: NOTE (report only, GATE_NEW=report): WOULD FAIL: " + failure)

    # GATE_NEW=enforce blocks on its own: promoting it does not depend on GATE_MODE (review note on #114).
    if new_failures:
        for f in failures:  # what the main checks found is still shown, with the verdict their own mode gives
            print("review-gate: %s: %s" % ("FAIL" if mode == "enforce" else "WOULD FAIL", f))
        for f in new_failures:
            print("review-gate: FAIL (GATE_NEW=enforce): " + f)
        print("review-gate: FAIL")
        return 1

    if not failures:
        print("review-gate: PASS (mode %s)" % mode)
        return 0
    for f in failures:
        print("review-gate: %s: %s" % ("FAIL" if mode == "enforce" else "WOULD FAIL", f))
    if mode == "enforce":
        print("review-gate: FAIL")
        return 1
    print("review-gate: PASS (report mode: nothing is blocked)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
