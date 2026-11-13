#!/usr/bin/env python3
"""Findings register (stdlib only): the AI review's findings become GitHub issues when a pull request merges, and one
dashboard says what to fix next. Nobody files a finding by hand: classifying one is closing its issue.

  harvest    (run on a merged pull request) read the ledger of the review's comment; each finding still `open`
             becomes an issue: CRITICAL and HIGH one issue each (at most 5 per pull request), everything else (and the
             overflow) one "PR<N> follow-ups" checklist issue. Idempotent: every issue carries a marker
             `<!-- finding:PR<N>#<id> -->` and a second run files nothing twice. A malformed ledger is skipped and
             logged, never guessed.
  dashboard  rank the open `finding` issues (due, severity, area, age) and print the result (`--print`) or write it to
             the one "Findings dashboard" issue (created if missing). Closed issues feed the rates:
             "completed" = fixed, "not planned" = accepted or not a defect; the rates appear after 20 closed findings.

Labels (created on first use by the issue API): finding, sev:*, src:ai, area:*, due:before-freeze|due:after-freeze.
Fail OPEN: a GitHub error is printed as NOT RUN and the exit code stays 0, so the register never fails a workflow.
Environment: REPO (owner/name), PR_NUMBER (harvest), PR_AUTHOR, GH_BIN (default gh, GH_TOKEN auth).
"""
import importlib.util
import os
import re
import subprocess
import sys

SERIOUS = ("CRITICAL", "HIGH")
SEV_RANK = {"CRITICAL": 0, "HIGH": 1, "MEDIUM": 2, "LOW": 3}
AREA_FIRST = ("identity", "db", "data", "contract", "platform")
MAX_SINGLE = 5
MIN_CLOSED_FOR_RATES = 20
DASHBOARD_MARK = "<!-- findings-dashboard -->"
MARK = re.compile(r"<!--\s*finding:(PR\d+#[\w-]+)\s*-->")


def load_context_module():
    """scripts/review-context.py has a dash in its name: load it by path (same directory as this file)."""
    path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "review-context.py")
    spec = importlib.util.spec_from_file_location("review_context", path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def gh_write(binary, method, endpoint, fields):
    """(ok, stdout) of `gh api -X METHOD endpoint -f k=v ...`."""
    argv = [binary, "api", "-X", method, endpoint]
    for key, value in fields:
        argv += ["-f", "%s=%s" % (key, value)]
    done = subprocess.run(argv, capture_output=True, text=True, timeout=60, check=False)
    return done.returncode == 0, done.stdout


def names(labels):
    return [x if isinstance(x, str) else x.get("name", "") for x in labels or []]


def label_value(labels, prefix):
    for n in labels:
        if n.startswith(prefix):
            return n[len(prefix):]
    return ""


def area_from_branch(ref):
    """feature/identity/profiles -> identity; fix/discover/x -> discover; anything else -> other."""
    parts = (ref or "").split("/")
    return parts[1] if len(parts) >= 3 else "other"


def marker_of(pr, key):
    return "<!-- finding:PR%s#%s -->" % (pr, key)


def list_issues(ctx, binary, repo):
    """Every `finding` issue (open and closed), without pull requests."""
    items = ctx.gh_api(binary, "repos/%s/issues?labels=finding&state=all&per_page=100" % repo)
    return [i for i in items if "pull_request" not in i]


def finding_labels(sev, area):
    due = "due:before-freeze" if sev in SERIOUS else "due:after-freeze"
    return ["finding", "sev:%s" % sev.lower(), "src:ai", "area:%s" % area, due]


def create_issue(binary, repo, title, body, labels, author):
    """Create an issue; if the author cannot be assigned, create it without an assignee."""
    fields = [("title", title), ("body", body)] + [("labels[]", l) for l in labels]
    endpoint = "repos/%s/issues" % repo
    if author and not author.endswith("[bot]"):
        ok, out = gh_write(binary, "POST", endpoint, fields + [("assignees[]", author)])
        if ok:
            return True
    ok, out = gh_write(binary, "POST", endpoint, fields)
    return ok


def neutral(text):
    """Text that came from the ledger (model output): an @mention would notify the person, so a zero-width space is
    put after the @. The name stays readable."""
    return re.sub(r"@(?=[A-Za-z0-9_-])", "@" + chr(0x200B), text or "")


def slug(identifier):
    """A finding id as it appears in a marker: letters, digits, underscore and dash only."""
    return re.sub(r"[^\w-]+", "-", identifier)


def single_title(pr, f):
    """The title carries the pull request number: the title check against a human-edited issue is then per pull request."""
    return neutral("PR%s %s: %s" % (pr, f["id"], f["title"][:90]))


STALE = "its file changed after the review (a later push may have fixed it): verify that it still applies before working on it"


def changed_since(ctx, binary, repo, reviewed_sha, head_sha):
    """The files changed between the commit the review looked at and the head of the merged pull request (the GitHub
    compare API); an empty set when that cannot be read (then no finding is flagged)."""
    if not reviewed_sha or not head_sha or head_sha.startswith(reviewed_sha):
        return set()
    try:
        items = ctx.gh_api(binary, "repos/%s/compare/%s...%s" % (repo, reviewed_sha, head_sha))
    except ctx.Unavailable:
        return set()
    return {f.get("filename", "") for item in items for f in item.get("files", [])}


def single_body(pr, pr_title, f, stale=False):
    lines = [
        marker_of(pr, slug(f["id"])),
        "Found by the AI review of PR #%s (%s)." % (pr, neutral(pr_title)),
        "",
        "- Where: `%s:%s`" % (neutral(f["file"]) or "?", f["line"] or "?"),
        "- Severity: %s" % f["sev"],
        "- Scenario: %s" % neutral(f["title"]),
    ]
    if stale:
        lines.append("- Note: " + STALE + ".")
    lines += ["", "Close as **completed** when it is fixed (with a test), or as **not planned** when it is accepted or not a "
              "defect (say why in the closing comment)."]
    return "\n".join(lines)


def followups_body(pr, pr_title, findings, changed=frozenset()):
    lines = [marker_of(pr, "followups"), "Open findings of the AI review of PR #%s (%s) that do not get their own issue. "
             "Tick a box when it is fixed; close the issue when none is left." % (pr, neutral(pr_title)), ""]
    for f in findings:
        note = " (%s)" % STALE if f["file"] in changed else ""
        lines.append("- [ ] `%s` [%s] `%s:%s` %s%s" % (neutral(f["id"]), f["sev"], neutral(f["file"]) or "?", f["line"] or "?",
                                                        neutral(f["title"]), note))
    return "\n".join(lines)


def harvest(ctx, binary, repo, pr, author_env):
    pulls = ctx.gh_api(binary, "repos/%s/pulls/%s" % (repo, pr))
    info = pulls[0] if pulls else {}
    if not info.get("merged"):
        print("findings: pull request #%s is not merged: nothing to file" % pr)
        return 0
    comments = ctx.gh_api(binary, "repos/%s/issues/%s/comments" % (repo, pr))
    ledger = ctx.parse_ledger(comments)
    if ledger is None:
        problem = any((c.get("user") or {}).get("login") in ctx.BOT_LOGINS
                      and re.search(r"review[ -]ledger", c.get("body") or "", re.I) is not None for c in comments)
        print("findings: ledger skipped (malformed)" if problem else "findings: no ledger on pull request #%s" % pr)
        return 0
    open_findings = [f for f in ledger["findings"] if f["status"] == "open"]
    if not open_findings:
        print("findings: no open finding on pull request #%s" % pr)
        return 0
    author = author_env or (info.get("user") or {}).get("login", "")
    area = area_from_branch((info.get("head") or {}).get("ref", ""))
    title = info.get("title", "")
    existing, existing_titles = set(), set()
    for issue in list_issues(ctx, binary, repo):
        existing.update(MARK.findall(issue.get("body") or ""))
        # A human may edit an issue body and drop the marker: the title is checked too, so a re-run files no duplicate.
        existing_titles.add(issue.get("title", ""))

    # A later push may have fixed a finding without a new review: the finding is still filed (nobody loses one), but
    # its issue says that the file changed after the review, so the first step is to verify it.
    changed = changed_since(ctx, binary, repo, ledger["sha"], (info.get("head") or {}).get("sha", ""))

    serious = [f for f in open_findings if f["sev"] in SERIOUS]
    singles, overflow = serious[:MAX_SINGLE], serious[MAX_SINGLE:]
    rest = overflow + [f for f in open_findings if f["sev"] not in SERIOUS]
    filed = 0
    for f in singles:
        key = "PR%s#%s" % (pr, slug(f["id"]))
        issue_title = single_title(pr, f)
        if key in existing or issue_title in existing_titles:
            continue
        body = single_body(pr, title, f, stale=f["file"] in changed)
        if create_issue(binary, repo, issue_title, body, finding_labels(f["sev"], area), author):
            filed += 1
            print("findings: filed %s" % key)
        else:
            print("findings: NOT RUN for %s (the issue could not be created)" % key)
    if rest and "PR%s#followups" % pr not in existing and "PR%s follow-ups" % pr not in existing_titles:
        worst = min((f["sev"] for f in rest), key=lambda s: SEV_RANK[s])
        labels = finding_labels(worst, area)
        labels[-1] = "due:after-freeze"
        if create_issue(binary, repo, "PR%s follow-ups" % pr, followups_body(pr, title, rest, changed), labels, author):
            filed += 1
            print("findings: filed PR%s#followups (%d findings)" % (pr, len(rest)))
        else:
            print("findings: NOT RUN for PR%s#followups (the issue could not be created)" % pr)
    print("findings: %d issue(s) filed" % filed)
    if filed:
        write_dashboard(ctx, binary, repo, create=False)
    return 0


def sort_key(issue):
    labels = names(issue.get("labels"))
    due = label_value(labels, "due:")
    area = label_value(labels, "area:")
    return (
        {"before-freeze": 0, "after-freeze": 1}.get(due, 2),
        SEV_RANK.get(label_value(labels, "sev:").upper(), 4),
        0 if area in AREA_FIRST else 1,
        issue.get("created_at", ""),
    )


def render_dashboard(issues):
    real = [i for i in issues if DASHBOARD_MARK not in (i.get("body") or "")]
    open_issues = sorted((i for i in real if i.get("state") == "open"), key=sort_key)
    closed = [i for i in real if i.get("state") == "closed"]
    out = [DASHBOARD_MARK, "# Findings dashboard", "", "Computed from the `finding` issues only. Close an issue as **completed** "
           "when it is fixed and as **not planned** when it is accepted or not a defect.", "", "## Fix next", ""]
    if open_issues:
        for i in open_issues[:5]:
            labels = names(i.get("labels"))
            out.append("- #%s [%s] %s (%s, area:%s)" % (i.get("number"), label_value(labels, "sev:").upper() or "?",
                                                         i.get("title", ""), label_value(labels, "due:") or "no due", label_value(labels, "area:") or "?"))
    else:
        out.append("Nothing open.")
    count = lambda p, v: sum(1 for i in open_issues if label_value(names(i.get("labels")), p) == v)
    out += ["", "## Counts", "", "- open: %d" % len(open_issues),
            "- before-freeze: %d | after-freeze: %d" % (count("due:", "before-freeze"), count("due:", "after-freeze")),
            "- critical: %d | high: %d | medium: %d | low: %d" % tuple(count("sev:", s) for s in ("critical", "high", "medium", "low")), ""]
    out += ["## Self-measurement", ""]
    if len(closed) < MIN_CLOSED_FOR_RATES:
        out.append("- not enough closed findings yet (%d/%d): rates would mislead." % (len(closed), MIN_CLOSED_FOR_RATES))
    else:
        done = sum(1 for i in closed if i.get("state_reason") == "completed")
        out.append("- fixed rate: %d%% (completed: %d of %d closed)" % (round(100 * done / len(closed)), done, len(closed)))
        out.append("- not planned: %d%% (accepted or not a defect)" % round(100 * (len(closed) - done) / len(closed)))
    return "\n".join(out) + "\n"


def write_dashboard(ctx, binary, repo, create):
    """Rewrite the dashboard issue; create it only when asked (harvest never creates it)."""
    issues = list_issues(ctx, binary, repo)
    body = render_dashboard(issues)
    for i in issues:
        if DASHBOARD_MARK in (i.get("body") or ""):
            gh_write(binary, "PATCH", "repos/%s/issues/%s" % (repo, i["number"]), [("body", body)])
            print("findings: dashboard updated (#%s)" % i["number"])
            return
    if create:
        ok, _ = gh_write(binary, "POST", "repos/%s/issues" % repo,
                         [("title", "Findings dashboard"), ("body", body), ("labels[]", "finding"), ("labels[]", "dashboard")])
        print("findings: dashboard created" if ok else "findings: NOT RUN (the dashboard issue could not be created)")


def main():
    args = sys.argv[1:]
    repo = os.environ.get("REPO", "")
    binary = os.environ.get("GH_BIN", "gh")
    if not args or args[0] not in ("harvest", "dashboard") or not repo:
        print("usage: findings.py harvest|dashboard [--print]   (REPO required; PR_NUMBER for harvest)")
        return 0 if repo else 0
    try:
        ctx = load_context_module()
        if args[0] == "harvest":
            pr = os.environ.get("PR_NUMBER", "")
            if not pr:
                print("findings: NOT RUN (PR_NUMBER is required)")
                return 0
            return harvest(ctx, binary, repo, pr, os.environ.get("PR_AUTHOR", ""))
        if "--print" in args:
            print(render_dashboard(list_issues(ctx, binary, repo)))
        else:
            write_dashboard(ctx, binary, repo, create=True)
    except Exception as exc:  # fail open: the register must never fail a workflow
        print("findings: NOT RUN (%s)" % type(exc).__name__)
    return 0


if __name__ == "__main__":
    sys.exit(main())
