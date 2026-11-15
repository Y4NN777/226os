#!/usr/bin/env python3
"""Findings register (stdlib only): the AI review's findings become GitHub issues when a pull request merges, and one
dashboard says what to fix next. Nobody files a finding by hand: classifying one is closing its issue.

  harvest    (run on a merged pull request) read the ledger of the review's comment; each finding still `open`
             becomes an issue: CRITICAL and HIGH one issue each (at most 5 per pull request), everything else (and the
             overflow) one "PR<N> follow-ups" checklist issue. Idempotent: every issue carries a marker
             `<!-- finding:PR<N>#<id> -->` and a second run files nothing twice. A malformed ledger is skipped and
             logged, never guessed.
             Entries of kind `note` (remarks, evidence OBSERVED or INFERRED) are not findings: the open ones become one
             "PR<N> notes" checklist issue (marker `PR<N>#notes`, labels kind:note, due:after-freeze, no `finding` label).
  dashboard  rank the open `finding` issues (due, severity, area, age) and print the result (`--print`) or write it to
             the one "Findings dashboard" issue (created if missing). Closed issues feed the rates:
             "completed" = fixed, "not planned" = accepted or not a defect; the rates appear after 20 closed findings.
             A separate Notes section counts the undecided notes and flags INFERRED ones undecided for 7 days or more.
  batch      triage in batches: the open `finding` issues and the undecided boxes of the open notes issues, ranked by
             severity (notes after LOW, stale INFERRED notes first), then area, then age, become ONE open
             "Review batch YYYY-MM-DD" checklist issue (marker `<!-- review-batch -->`, label review-batch): the top 15,
             frozen when the batch is created (a new CRITICAL or HIGH joins it), with a count of the rest. A batch is the
             scope of ONE fix pull request (branch fix/review/batch-YYYY-MM-DD, its body names the batch issue). An
             update never unticks a box; when that pull request is merged into develop, or every source is closed or
             decided, the batch is closed as completed and the next run starts a new one. It never writes to a source issue and never mentions anyone. `--print` writes nothing.
             A batch closed by the no-commit path (review-autofix.py `nothing`, marker batch-handled) holds its ticked
             items: they got a verdict and stay open for the owner, so they stay out of new batches (and never join one as
             urgent) while their row text is unchanged; the next batch moves on to the waiting items.
  settle PR  when the batch fix pull request PR (fix/review/batch-*) is merged into develop, its results table settles
             each source. Only FIXED and FIXED ELSEWHERE act: they close a finding issue as completed or tick a note's
             box. NOT A DEFECT and DECIDED never close a finding (a model verdict is not a decision): their reason is
             posted as a comment and a human closes it. For a NOTE they tick its box when the reason cites a reference
             that exists at the tip of develop (`path:line` or `path:a-b` inside the note's own file, or a
             docs/decisions/NNNN-*.md file; checked by the script); without such a reference they only comment. OWNER and AFTER FREEZE stay open. One comment per touched source ("Settled by #PR (review batch #N): ..."), once (marker); a
             notes issue whose boxes are all ticked is closed (as not planned when a verdict ticked one, else as completed);
             the batch issue closes last. Fails CLOSED like
             `batch`.

Every issue is self-contained: the ledger entry's `detail` (proof and reasoning), `fix`, `source` (review or deep), a
permalink at `first_sha` (only when the location was verified) and a link to the exact review comment are written in it,
after the same neutraliser and secret redaction as the titles. An older ledger without them shows "not recorded (older
review)" instead of failing. `batch` puts the items that share an id and a file (the same defect on several pull
requests) in ONE row that lists every source (`<!-- batch-also:... -->`); `settle` settles every source of the row.

Labels (created on first use by the issue API): finding, sev:*, src:ai, area:*, due:before-freeze|due:after-freeze,
review-batch.
Fail OPEN (harvest, dashboard): a GitHub error is printed as NOT RUN and the exit code stays 0, so the register never
fails a workflow. Fail CLOSED (batch): a GitHub error is printed as FAILED and the exit code is 1, because a batch built
from a partial read would lie about what is left.
Environment: REPO (owner/name), PR_NUMBER (harvest), PR_AUTHOR, GH_BIN (default gh, GH_TOKEN auth).
"""
import base64
import datetime
import importlib.util
import os
import re
import subprocess
import sys
import urllib.parse

SERIOUS = ("CRITICAL", "HIGH")
SEV_RANK = {"CRITICAL": 0, "HIGH": 1, "MEDIUM": 2, "LOW": 3}
AREA_FIRST = ("identity", "db", "data", "contract", "platform")
MAX_SINGLE = 5
STALE_NOTE_DAYS = 7
NOTE_LINE = re.compile(r"^\s*- \[( |x|X)\]\s*`[^`]*`\s*\[(OBSERVED|INFERRED)\]")
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


def gh_write_result(binary, method, endpoint, fields):
    """(exit code, stdout, stderr) of `gh api -X METHOD endpoint -f k=v ...`, retried once when that is safe."""
    argv = [binary, "api", "-X", method, endpoint]
    for key, value in fields:
        argv += ["-f", "%s=%s" % (key, value)]
    return rc_module().gh_call(argv, method, endpoint)


def gh_write(binary, method, endpoint, fields):
    """(ok, stdout) of `gh api -X METHOD endpoint -f k=v ...`."""
    code, out, _ = gh_write_result(binary, method, endpoint, fields)
    return code == 0, out


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


def list_issues(ctx, binary, repo, label="finding"):
    """Every issue with the label (open and closed), without pull requests: `finding` (default) or `kind:note`."""
    items = ctx.gh_api(binary, "repos/%s/issues?labels=%s&state=all&per_page=100" % (repo, label))
    return [i for i in items if "pull_request" not in i]


def finding_labels(sev, area):
    due = "due:before-freeze" if sev in SERIOUS else "due:after-freeze"
    return ["finding", "sev:%s" % sev.lower(), "src:ai", "area:%s" % area, due]


def note_labels(area):
    """Notes are remarks, not defects: no `finding` label (they stay out of the findings counts), always after the freeze."""
    return ["kind:note", "src:ai", "area:%s" % area, "due:after-freeze"]


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


_RC = []


def rc_module():
    """scripts/review-context.py (loaded once): the shared text helpers (defuse, redact) live there."""
    if not _RC:
        _RC.append(load_context_module())
    return _RC[0]


def defuse(text, limit=None):
    return rc_module().defuse(text, limit)


def field_text(f, key):
    """The detail or the fix of an entry as one safe line: secret shapes replaced, HTML, @mentions and #N disarmed; the
    plain mention when an older ledger did not record it."""
    raw = f.get(key) or ""
    if not raw:
        return rc_module().NOT_RECORDED
    return defuse(rc_module().redact(raw)[0])


def flagged(f):
    """True when a secret-looking text was replaced (at write time or now)."""
    rc = rc_module()
    deep = f.get("deep") or {}
    return bool(f.get("redacted") or deep.get("redacted") or any(rc.looks_secret(x) for x in (f.get("detail"), f.get("fix"), deep.get("detail"), deep.get("fix"))))


def permalink(repo, f):
    """A stable link to file:line at the commit the finding was raised on; "" unless the location was verified."""
    if not repo or f.get("location") != "verified" or not f.get("first_sha") or not f.get("file"):
        return ""
    return "https://github.com/%s/blob/%s/%s%s" % (repo, f["first_sha"], urllib.parse.quote(f["file"], safe="/"), "#L%d" % f["line"] if f.get("line") else "")


def source_text(f):
    return "review (the deep review also raised it)" if f.get("source") != "deep" and f.get("deep") else f.get("source") or "review"


def detail_block(f, repo, url, bullet, labels):
    """The detail, fix, source, permalink and review comment of an entry, one line each. `labels` is the style:
    capitalised labels for a field list, lower case for the indented lines under a checklist box."""
    cap = (lambda t: t.capitalize()) if labels == "cap" else (lambda t: t)
    out = [bullet + "%s: %s" % (cap("detail"), field_text(f, "detail")), bullet + "%s: %s" % (cap("fix"), field_text(f, "fix"))]
    deep = f.get("deep")
    if deep and (deep.get("detail") or deep.get("fix")):
        out.append(bullet + "%s: %s | fix: %s" % (cap("deep review"), field_text(deep, "detail"), field_text(deep, "fix")))
    out.append(bullet + "%s: %s" % (cap("source"), source_text(f)))
    link = permalink(repo, f)
    if link:
        out.append(bullet + "%s: %s" % (cap("permalink"), link))
    elif f.get("location") == "mismatch":
        probable = defuse(f.get("probable") or "")
        out.append(bullet + "%s: MISMATCH: what the text names is not near the cited `%s:%s`; probable location: `%s`%s" % (
            cap("location"), neutral(f.get("file")) or "?", f.get("line") or "?", probable or "?",
            " (%s)" % "https://github.com/%s/blob/%s/%s" % (repo, f["first_sha"], urllib.parse.quote(probable.rsplit(":", 1)[0], safe="/")
                                                            + "#L" + probable.rsplit(":", 1)[1]) if repo and f.get("first_sha") and ":" in probable else ""))
    elif f.get("location") == "unverified":
        out.append(bullet + "%s: unverified (the file or the line was not found at the reviewed commit)" % cap("location"))
    if url:
        out.append(bullet + "%s: %s" % (cap("review comment"), url))
    if flagged(f):
        out.append(bullet + "%s: a text that looked like a secret was replaced by the review pipeline" % cap("note"))
    return out


def neutral(text):
    """Text that came from the ledger (model output): an @mention would notify the person, so a zero-width space is
    put after the @. The name stays readable."""
    return re.sub(r"@(?=[A-Za-z0-9_-])", "@" + chr(0x200B), text or "")


def slug(identifier):
    """A finding id as it appears in a marker: letters, digits, underscore and dash only."""
    return re.sub(r"[^\w-]+", "-", identifier)


def single_title(pr, f):
    """The title carries the pull request number: the title check against a human-edited issue is then per pull request."""
    return defuse("PR%s %s: %s" % (pr, f["id"], f["title"][:90]))


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


def single_body(pr, pr_title, f, stale=False, repo="", url=""):
    lines = [
        marker_of(pr, slug(f["id"])),
        "Found by the AI review of PR #%s (%s)." % (pr, neutral(pr_title)),
        "",
        "- Where: `%s:%s`" % (neutral(f["file"]) or "?", f["line"] or "?"),
        "- Severity: %s" % f["sev"],
        "- Scenario: %s" % defuse(f["title"]),
    ]
    lines += detail_block(f, repo, url, "- ", "cap")
    if stale:
        lines.append("- Note: " + STALE + ".")
    lines += ["", "Close as **completed** when it is fixed (with a test), or as **not planned** when it is accepted or not a "
              "defect (say why in the closing comment)."]
    return "\n".join(lines)


def checklist_line(f, evidence, changed):
    note = " (%s)" % STALE if f["file"] in changed else ""
    return "- [ ] `%s` [%s] `%s:%s` %s%s" % (neutral(f["id"]), evidence, neutral(f["file"]) or "?", f["line"] or "?", defuse(f["title"]), note)


def followups_body(pr, pr_title, findings, changed=frozenset(), repo="", url=""):
    lines = [marker_of(pr, "followups"), "Open findings of the AI review of PR #%s (%s) that do not get their own issue. "
             "Tick a box when it is fixed; close the issue when none is left." % (pr, neutral(pr_title)), ""]
    for f in findings:
        lines.append(checklist_line(f, f["sev"], changed))
        lines += detail_block(f, repo, url, "  - ", "low")
    return "\n".join(lines)


def notes_body(pr, pr_title, notes, changed=frozenset(), repo="", url=""):
    lines = [marker_of(pr, "notes"), "Notes of the AI review of PR #%s (%s): remarks that are not defects. Tick a box when you decided "
             "(fine as is, or fixed) and say why in a comment; close the issue as **not planned** when none needs action. "
             "OBSERVED means the reviewer read it in the code; INFERRED means it is a guess to check." % (pr, neutral(pr_title)), ""]
    for n in notes:
        lines.append(checklist_line(n, n["evidence"], changed))
        lines += detail_block(n, repo, url, "  - ", "low")
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
    url = ledger.get("url", "")
    open_entries = [f for f in ledger["findings"] if f["status"] == "open"]
    open_findings = [f for f in open_entries if f["kind"] == "finding"]
    open_notes = [f for f in open_entries if f["kind"] == "note"]
    if not open_entries:
        print("findings: no open finding on pull request #%s" % pr)
        return 0
    author = author_env or (info.get("user") or {}).get("login", "")
    area = area_from_branch((info.get("head") or {}).get("ref", ""))
    title = info.get("title", "")
    existing, existing_titles = set(), set()
    for issue in list_issues(ctx, binary, repo) + list_issues(ctx, binary, repo, "kind:note"):
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
        body = single_body(pr, title, f, stale=f["file"] in changed, repo=repo, url=url)
        if create_issue(binary, repo, issue_title, body, finding_labels(f["sev"], area), author):
            filed += 1
            print("findings: filed %s" % key)
        else:
            print("findings: NOT RUN for %s (the issue could not be created)" % key)
    if rest and "PR%s#followups" % pr not in existing and "PR%s follow-ups" % pr not in existing_titles:
        worst = min((f["sev"] for f in rest), key=lambda s: SEV_RANK[s])
        labels = finding_labels(worst, area)
        labels[-1] = "due:after-freeze"
        if create_issue(binary, repo, "PR%s follow-ups" % pr, followups_body(pr, title, rest, changed, repo, url), labels, author):
            filed += 1
            print("findings: filed PR%s#followups (%d findings)" % (pr, len(rest)))
        else:
            print("findings: NOT RUN for PR%s#followups (the issue could not be created)" % pr)
    if open_notes and "PR%s#notes" % pr not in existing and "PR%s notes" % pr not in existing_titles:
        if create_issue(binary, repo, "PR%s notes" % pr, notes_body(pr, title, open_notes, changed, repo, url), note_labels(area), author):
            filed += 1
            print("findings: filed PR%s#notes (%d notes)" % (pr, len(open_notes)))
        else:
            print("findings: NOT RUN for PR%s#notes (the issue could not be created)" % pr)
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


def parse_time(text):
    try:
        return datetime.datetime.strptime(text, "%Y-%m-%dT%H:%M:%SZ").replace(tzinfo=datetime.timezone.utc)
    except (TypeError, ValueError):
        return None


def now():
    """FINDINGS_NOW (an ISO UTC time) pins the clock in tests."""
    return parse_time(os.environ.get("FINDINGS_NOW", "")) or datetime.datetime.now(datetime.timezone.utc)


def render_notes(notes_issues):
    """The Notes section: open notes issues, the undecided lines (unticked boxes), and the INFERRED ones that stayed
    undecided for STALE_NOTE_DAYS days or more (a guess nobody checked is the one to look at first)."""
    open_issues = [i for i in notes_issues if i.get("state") == "open" and DASHBOARD_MARK not in (i.get("body") or "")]
    undecided, stale, stale_issues = 0, 0, []
    for i in open_issues:
        created = parse_time(i.get("created_at"))
        old = created is not None and (now() - created).days >= STALE_NOTE_DAYS
        for line in (i.get("body") or "").splitlines():
            m = NOTE_LINE.match(line)
            if m and m.group(1) == " ":
                undecided += 1
                if old and m.group(2) == "INFERRED":
                    stale += 1
                    if "#%s" % i.get("number") not in stale_issues:
                        stale_issues.append("#%s" % i.get("number"))
    out = ["", "## Notes", "", "Remarks of the AI review that are not defects (one checklist issue per pull request, label `kind:note`).", "",
           "- open notes issues: %d" % len(open_issues), "- undecided: %d" % undecided,
           "- INFERRED undecided for %d days or more: %d%s" % (STALE_NOTE_DAYS, stale, " (%s)" % ", ".join(stale_issues) if stale_issues else "")]
    return out


def render_dashboard(issues, notes_issues=()):
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
    out += render_notes(notes_issues)
    out += ["", "## Self-measurement", ""]
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
    body = render_dashboard(issues, list_issues(ctx, binary, repo, "kind:note"))
    for i in issues:
        if DASHBOARD_MARK in (i.get("body") or ""):
            gh_write(binary, "PATCH", "repos/%s/issues/%s" % (repo, i["number"]), [("body", body)])
            print("findings: dashboard updated (#%s)" % i["number"])
            return
    if create:
        ok, _ = gh_write(binary, "POST", "repos/%s/issues" % repo,
                         [("title", "Findings dashboard"), ("body", body), ("labels[]", "finding"), ("labels[]", "dashboard")])
        print("findings: dashboard created" if ok else "findings: NOT RUN (the dashboard issue could not be created)")


BATCH_MARK = "<!-- review-batch -->"
BATCH_LABEL = "review-batch"
BATCH_SIZE = 15
NOTE_ITEM = re.compile(r"^\s*- \[( |x|X)\]\s*`([^`]*)`\s*\[(OBSERVED|INFERRED)\]\s*(.*)$")
BATCH_LINE = re.compile(r"^- \[( |x|X)\] (.*?) <!-- batch-item:(#\d+(?:/[\w-]+)?) -->$")
ALSO = re.compile(r"\s*<!-- batch-also:([^>]*?) -->")
ITEM_KEY_RE = re.compile(r"^#\d+(?:/[\w-]+)?$")
WHERE = re.compile(r"^- Where: `([^`]*):([^:`]*)`", re.M)
NOTE_WHERE = re.compile(r"^`([^`]*):([^:`]*)`")
NOTES_TITLE = re.compile(r"^PR(\d+) notes")
NOTE_RANK = len(SEV_RANK)  # notes come after every finding severity; a stale INFERRED note first


class BatchError(Exception):
    """A write GitHub refused: the batch fails closed."""


def one_line(text, width=120):
    text = " ".join(neutral(text).split())
    return text if len(text) <= width else text[:width - 1] + "..."


def defect_id(ident, where):
    """The key under which items are one defect: the id, the file AND the line, all three. A generic id on the same file at
    another line, or with no known line, is another defect and stays its own row (settling one must never close the other).
    The title is deliberately NOT part of the key (#212 L8, accepted trade-off): a note's text has no `<id>: ` prefix and a
    finding's title is cut, so a title key would split the same defect raised on several pull requests."""
    if not ident or not where or where.group(1) in ("", "?") or not re.fullmatch(r"\d+", where.group(2) or ""):
        return None
    return (ident, where.group(1), int(where.group(2)))


def batch_items(finding_issues, note_issues):
    """The open work: each open finding issue is one item (a follow-ups issue too), each unticked box of an open notes
    issue is one item. rank = (due, severity, area, age); notes rank after LOW, INFERRED undecided for 7 days first."""
    items = []
    for i in finding_issues:
        if i.get("state") != "open" or DASHBOARD_MARK in (i.get("body") or ""):
            continue
        labels = names(i.get("labels"))
        sev = label_value(labels, "sev:").upper() or "?"
        mark, where = MARK.search(i.get("body") or ""), WHERE.search(i.get("body") or "")
        ident = mark.group(1).partition("#")[2] if mark else ""
        gid = defect_id(ident, where) if ident and ident not in ("followups", "notes") else None
        items.append({"key": "#%s" % i.get("number"), "rank": sort_key(i), "number": i.get("number") or 0, "gid": gid,
                      "pr": mark.group(1).partition("#")[0][2:] if mark else "?",
                      "text": "#%s [%s] %s (area:%s)" % (i.get("number"), sev, one_line(i.get("title", "")),
                                                         label_value(labels, "area:") or "?")})
    for i in note_issues:
        if i.get("state") != "open" or DASHBOARD_MARK in (i.get("body") or ""):
            continue
        created = parse_time(i.get("created_at"))
        old = created is not None and (now() - created).days >= STALE_NOTE_DAYS
        area = label_value(names(i.get("labels")), "area:")
        for line in (i.get("body") or "").splitlines():
            m = NOTE_ITEM.match(line)
            if not m or m.group(1) != " ":
                continue
            stale = old and m.group(3) == "INFERRED"
            where = NOTE_WHERE.match(m.group(4))
            title = MARK.search(i.get("body") or "") or NOTES_TITLE.match(i.get("title") or "")
            items.append({"key": "#%s/%s" % (i.get("number"), slug(m.group(2))), "number": i.get("number") or 0,
                          "gid": defect_id(slug(m.group(2)), where),
                          "pr": (title.group(1).partition("#")[0][2:] if title.group(1).startswith("PR") else title.group(1)) if title else "?",
                          "rank": (1, NOTE_RANK + (0 if stale else 1), 0 if area in AREA_FIRST else 1, i.get("created_at", "")),
                          "text": "#%s [NOTE %s%s] `%s` %s" % (i.get("number"), m.group(3), ", %d+ days" % STALE_NOTE_DAYS if stale else "",
                                                             one_line(m.group(2), 40), one_line(m.group(4)))})
    return sorted(items, key=lambda it: (it["rank"], it["key"]))


def group_items(items):
    """One row per defect: the items that share an id, a file and a line (the same finding raised on several pull requests, or a
    finding and a note) are ONE row, keyed by the oldest source, that lists every source. Rows are ranked by their worst
    member. An item without an id, a file and a line stays alone."""
    groups = {}
    for it in items:
        groups.setdefault(it["gid"] or ("alone", it["key"]), []).append(it)
    rows = []
    for members in groups.values():
        members.sort(key=lambda it: (it["number"], it["key"]))
        head = members[0]
        text = head["text"]
        if len(members) > 1:
            text += " [sources: %s]" % ", ".join("%s (PR%s)" % (m["key"], m["pr"]) for m in members)
        rows.append({"key": head["key"], "also": [m["key"] for m in members[1:]], "sources": [m["key"] for m in members],
                     "rank": min(m["rank"] for m in members), "text": text})
    return sorted(rows, key=lambda r: (r["rank"], r["key"]))


BATCH_HANDLED = "<!-- batch-handled -->"
HANDLED_LINE = re.compile(r"^<!-- batch-handled -->$", re.M)


def held_items(batch_issues):
    """{source key: row text} of the items that got a verdict and stay open for the owner: the ticked rows of the CLOSED
    batch issues that `review-autofix.py nothing` closed (the marker on a line of its own, which no row can forge). A later
    batch records over an earlier one. Such an item stays out of new batches while its row text is the same (title,
    severity, area, sources); when it changes, or when it is closed, it is no longer held."""
    held = {}
    for b in sorted(batch_issues, key=lambda i: i.get("number", 0)):
        body = b.get("body") or ""
        if b.get("state") != "closed" or BATCH_MARK not in body or not HANDLED_LINE.search(body):
            continue
        for line in body.splitlines():
            m = BATCH_LINE.match(line)
            if m and m.group(1) != " ":
                also = ALSO.search(line)
                for src in [m.group(3)] + [k for k in (also.group(1).split(",") if also else []) if ITEM_KEY_RE.match(k)]:
                    held[src] = ALSO.sub("", m.group(2))
    return held


def render_batch(title, lines, rest, held=0):
    day = title.rsplit(" ", 1)[-1]
    out = [BATCH_MARK, "# " + title, "",
           "Ranked by severity, then area, then age: work from the top. This batch is the scope of exactly ONE fix pull "
           "request into `develop`: branch `fix/review/batch-%s`, and its pull request body names this issue (#number). "
           "A human ticks or closes each source finding or note with the reason (a finding: close its issue as "
           "**completed** or **not planned**; a note: tick its box in its notes issue and say why). Ticking a box here "
           "only tracks progress: the sources decide. This issue closes itself when its fix pull request is merged into "
           "`develop` or when every source is closed or decided; the next run then starts a new batch from what is "
           "still open." % day, ""]
    out += ["- [%s] %s%s <!-- batch-item:%s -->" % ("x" if ticked else " ", text, " <!-- batch-also:%s -->" % ",".join(also) if also else "", key)
            for key, ticked, text, also in lines]
    out += ["", "%d more open items wait for the next batch." % rest if rest else "Nothing else is waiting."]
    if held:
        out.append("%d items with a verdict wait for the owner (ticked in a closed batch); each comes back when it changes." % held)
    return "\n".join(out) + "\n"


def batch_write(binary, method, endpoint, fields):
    code, _, err = gh_write_result(binary, method, endpoint, fields)
    if code != 0:
        raise BatchError("GitHub refused %s" % rc_module().error_detail(method, endpoint, code, err))


def merged_fix(ctx, binary, repo, issue):
    """The first pull request merged into develop after the batch was created whose body names the batch issue
    (`#N` or its URL), or None. Since ADR-0013 `develop` is the default branch, so `Closes #N` closes an issue natively on a merge into develop (needs-owner ones included); `close-merged-issues.py` stays as the idempotent backstop, and this check settles the batch whichever way the issue closed."""
    number = issue.get("number")
    # Only a pull request that says it IS the batch's fix counts: a `Batch: #N` line (the body written for the batch fix
    # pull request), a closing keyword before the reference, A bare mention does not, and the pull request's head branch must be a `fix/review/batch-` one.
    ref_n = r"(?:#%s(?!\d)|https://github\.com/[\w.-]+/[\w.-]+/issues/%s(?!\d))" % (number, number)
    # (No bare "review batch #N" match: anyone can write those words; and the head must be the batch branch, #212 L7.)
    ref = re.compile(r"^Batch:\s*%s|\b(?:close[sd]?|fix(?:e[sd])?|resolve[sd]?)\b:?\s+(?:the\s+)?(?:review\s+)?(?:batch\s+)?%s" % (ref_n, ref_n), re.I | re.M)
    created = issue.get("created_at", "")
    pulls = ctx.gh_api(binary, "repos/%s/pulls?state=closed&base=develop&sort=updated&direction=desc&per_page=100" % repo)
    for pr in pulls:
        merged = pr.get("merged_at") or ""
        if merged and merged >= created and str((pr.get("head") or {}).get("ref") or "").startswith(BATCH_PREFIX) and ref.search(pr.get("body") or ""):
            return pr
    return None


SETTLE_MARK = "<!-- settled-by:PR%s -->"
DECISIONS = ("FIXED", "FIXED ELSEWHERE", "NOT A DEFECT", "DECIDED", "OWNER", "AFTER FREEZE")
# Only a fix closes or ticks by itself; NOT A DEFECT and DECIDED are the fixer's opinion: a human closes a finding; a note is
# ticked only with a reference the script verified (verified_reference).
CLOSE_AS = {"FIXED": "completed", "FIXED ELSEWHERE": "completed"}
STAYS_OPEN = {"NOT VERIFIED": "stays open: a human checks", "OWNER": "stays open for the owner", "AFTER FREEZE": "stays open until after the freeze",
              "NOT A DEFECT": "stays open: a human closes it as not planned if they agree",
              "DECIDED": "stays open: a human closes it as not planned or ticks it if they agree"}
SETTLE_ROW = re.compile(r"^\|\s*(#\d+(?:/[\w-]+)?)\s*\|")
BATCH_PREFIX = "fix/review/batch-"


def settle_rows(body):
    """[(key, decision, reason)] from the results table of a batch fix pull request: | item | source | decision |
    reason |. Escaped pipes (\\|) belong to the text. Rows with an unknown decision are ignored (left open)."""
    rows = []
    for line in (body or "").splitlines():
        line = line.strip()
        if not SETTLE_ROW.match(line):
            continue
        cells = [c.strip().replace("\\|", "|") for c in re.split(r"(?<!\\)\|", line.strip().strip("|"))]
        if len(cells) < 4 or cells[2].upper() not in DECISIONS:
            continue
        rows.append((cells[0], cells[2].upper(), " | ".join(cells[3:])))
    return rows


def settle_text(decision, reason, pr):
    if decision == "FIXED":
        return "fixed in #%s" % pr
    return "%s: %s" % (decision.lower(), one_line(reason, 200))


PR_REF = re.compile(r"#(\d+)")
# A reference a NOT A DEFECT or DECIDED verdict of a NOTE must cite to tick it: `path:line` or `path:a-b` in the note's own
# file, or an ADR file.
# A path with a directory or an extension; a bare name (Makefile, Dockerfile) counts only when it is the note's own file,
# so a word such as "Retry:3" in the reason is never read as a citation.
CITE_LINE = re.compile(r"(?<![\w./-])((?:[\w.-]+/)*[\w.-]+):(\d+)(?:-(\d+))?(?!\d)")
CITE_ADR = re.compile(r"(?<![\w./-])(docs/decisions/\d{4}-[\w.-]+\.md)(?![\w.-])")
TRUNK = "develop"


def trunk_file(ctx, binary, repo, path):
    """The text of `path` at the tip of develop, or None (missing, a directory, too large, unreadable: never a guess)."""
    if ".." in path.split("/") or path.startswith("/"):
        return None
    try:
        items = ctx.gh_api(binary, "repos/%s/contents/%s?ref=%s" % (repo, urllib.parse.quote(path), TRUNK))
        item = items[0]
        if item.get("type") != "file" or item.get("encoding") != "base64":
            return None
        return base64.b64decode(item.get("content") or "").decode("utf-8")
    except Exception:  # any failure leaves the note open, which is today's behaviour
        return None


def verified_reference(ctx, binary, repo, reason, own_file, own_line=""):
    """The reference that lets a NOTE verdict tick its box, else None. Checked here, never by the model.

    Every reference the verdict cites must exist at the tip of develop: one missing ADR, missing file or line past the
    end of its file and nothing is accepted (a valid ADR cannot carry a wrong line). Then at least one must count: a
    `docs/decisions/NNNN-*.md` file, or `path:line` / `path:a-b` within OWN_FILE, the file the note itself points at, other
    than the note's own location OWN_LINE (citing the place the note already names proves nothing). A line of another
    file proves nothing about this note (owner, 2026-11-15)."""
    text = re.sub("[\u200b\u200c\u200d\u2060\ufeff]", "", reason or "")
    accepted = None
    for m in CITE_ADR.finditer(text):
        if trunk_file(ctx, binary, repo, m.group(1)) is None:
            return None
        accepted = accepted or m.group(1)
    for m in CITE_LINE.finditer(text):
        if "/" not in m.group(1) and "." not in m.group(1) and m.group(1) != own_file:
            continue
        first, last = int(m.group(2)), int(m.group(3) or m.group(2))
        body = trunk_file(ctx, binary, repo, m.group(1))
        if body is None:
            return None
        total = body.count("\n") + (1 if body and not body.endswith("\n") else 0)
        if not 1 <= first <= last <= total:
            return None
        cited = m.group(2) + ("-" + m.group(3) if m.group(3) else "")
        if own_file and m.group(1) == own_file and cited != own_line:
            accepted = accepted or m.group(0)
    return accepted


ORIGIN_TITLE = re.compile(r"^PR(\d+)\b")


def origin_pr(issue):
    """The pull request a source issue was raised on: PR<N> of its finding marker, else of its title ("PR<N> notes",
    "PR<N> follow-ups", "PR<N> <id>: ..."); "" when neither says."""
    m = MARK.search(issue.get("body") or "")
    if m:
        return m.group(1).partition("#")[0][2:]
    t = ORIGIN_TITLE.match(issue.get("title") or "")
    return t.group(1) if t else ""


def unverified(ctx, binary, repo, pr, files, decision, reason, cited, created="", origin=""):
    """None when a FIXED or FIXED ELSEWHERE row is backed by the facts, else the one-line comment that leaves the source open
    (the fixer's verdict is its own opinion): FIXED needs the merged batch pull request's diff (`files`) to touch the file
    the finding cites; FIXED ELSEWHERE needs the pull request named in the reason (#N) to be another one than the source's
    own (`origin`), merged after the source was created (`created`), and to have touched the cited file too."""
    if decision == "FIXED":
        if not cited or cited == "?":
            return "marked FIXED but the finding cites no file: a human checks"
        if files is None or cited not in files:
            return "marked FIXED but #%s did not touch `%s`: a human checks" % (pr, one_line(cited, 120))
        return None
    if decision == "FIXED ELSEWHERE":
        m = PR_REF.search(re.sub("[\u200b\u200c\u200d\u2060\ufeff]", "", reason or ""))  # the batch body was disarmed: #<ZWSP>N
        if not m:
            return "marked FIXED ELSEWHERE but names no pull request: a human checks"
        if origin and m.group(1) == origin:
            return "names its own source PR #%s: a human checks" % origin
        try:
            info = (ctx.gh_api(binary, "repos/%s/pulls/%s" % (repo, m.group(1))) or [{}])[0]
        except Exception:  # not a pull request, or unreadable: never close on it
            info = {}
        if not info.get("merged_at"):
            return "named PR #%s is not merged: a human checks" % m.group(1)
        # The PR the finding was raised on (merged before the source existed) fixes nothing "elsewhere".
        if not created or info["merged_at"] <= created:
            return "named PR #%s was merged before this source was created (or its creation time is unknown): a human checks" % m.group(1)
        try:
            named = set()
            for f in ctx.gh_api(binary, "repos/%s/pulls/%s/files?per_page=100" % (repo, m.group(1))):
                named |= {f.get(k) for k in ("filename", "previous_filename") if f.get(k)}
        except Exception:
            named = set()
        if not cited or cited == "?" or cited not in named:
            return "named PR #%s did not touch `%s`: a human checks" % (m.group(1), one_line(cited or "no file cited", 120))
    return None


def settle_issue(ctx, binary, repo, pr, batch_no, number, entries, files=None, mark=None, by=None):
    """Settle one source issue (a finding issue, or a notes issue with its boxes). Idempotent: an issue that carries
    the settling comment of this pull request (or of `mark`, the batch run of the no-commit path) is not touched again;
    the comment is written last. `by` names who settles it in the comment (default: #PR)."""
    mark = mark or SETTLE_MARK % pr
    by = by or "#%s" % pr
    issue = (ctx.gh_api(binary, "repos/%s/issues/%s" % (repo, number)) or [{}])[0]
    if "pull_request" in issue:
        return
    comments = ctx.gh_api(binary, "repos/%s/issues/%s/comments" % (repo, number))
    if any(mark in (c.get("body") or "") for c in comments):
        return
    endpoint = "repos/%s/issues/%s" % (repo, number)
    held = [l for l in ("needs-owner", BATCH_LABEL) if l in names(issue.get("labels"))]
    if held:
        detail = "; ".join("%s%s" % ("`%s` " % one_line(n, 40) if n else "", d) for n, d, _ in entries)
        text = "%s\nSettled by %s (review batch #%s): %s; left open because it is labelled %s." % (
            mark, by, batch_no or "?", detail, held[0])
        batch_write(binary, "POST", endpoint + "/comments", [("body", text)])
        print("findings settle: #%s is labelled %s: commented, left open" % (number, held[0]))
        return
    notes = {slug(n): (d, r) for n, d, r in entries if n}
    if notes:
        old = issue.get("body") or ""
        cited_by, where = {}, {}
        for line in old.split("\n"):
            mm = NOTE_ITEM.match(line)
            if mm:
                w = NOTE_WHERE.match(mm.group(4))
                cited_by[slug(mm.group(2))] = w.group(1) if w else ""
                where[slug(mm.group(2))] = w.group(2) if w else ""
        for key, (d, r) in list(notes.items()):
            bad = unverified(ctx, binary, repo, pr, files, d, r, cited_by.get(key, ""), issue.get("created_at") or "", origin_pr(issue))
            if bad:
                notes[key] = ("NOT VERIFIED", bad)
        # A NOTE is not a defect: a NOT A DEFECT or DECIDED verdict that cites a reference which exists at the tip of develop
        # ticks its box (and the issue closes as not planned when every box is settled). Findings never do.
        cites = {}
        for key, (d, r) in notes.items():
            if d in ("NOT A DEFECT", "DECIDED"):
                ref = verified_reference(ctx, binary, repo, r, cited_by.get(key, ""), where.get(key, ""))
                if ref:
                    cites[key] = ref
        entries = [(n, *notes.get(slug(n), (d, r))) if n else (n, d, r) for n, d, r in entries]
        # The issue closes as not planned only when this run settled every box by verdict; a box ticked before this run
        # (by a human or a fix) or by a fix in this run keeps the old reason, completed.
        out, ticked_all, seen, by_verdict, by_fix, ticked_now = [], True, 0, False, False, set()
        for line in old.split("\n"):
            m = NOTE_ITEM.match(line)
            if m:
                seen += 1
                hit = notes.get(slug(m.group(2)))
                if m.group(1) != " ":
                    by_fix = True
                elif hit and hit[0] in CLOSE_AS:
                    line = line.replace("- [ ]", "- [x]", 1) + " \u2014 " + settle_text(hit[0], hit[1], pr)
                    by_fix = True
                elif hit and slug(m.group(2)) in cites:
                    line = line.replace("- [ ]", "- [x]", 1) + " \u2014 " + settle_text(hit[0], hit[1], pr)
                    by_verdict = True
                    ticked_now.add(slug(m.group(2)))
                elif m.group(1) == " ":
                    ticked_all = False
            out.append(line)
        new = "\n".join(out)
        if new != old:
            batch_write(binary, "PATCH", endpoint, [("body", new)])
        if seen and ticked_all and issue.get("state") == "open":
            batch_write(binary, "PATCH", endpoint, [("state", "closed"), ("state_reason", "not_planned" if by_verdict and not by_fix else "completed")])
        detail = "; ".join("`%s` %s" % (one_line(n, 40), d if d in CLOSE_AS else "%s ticked (reference `%s` exists)%s" % (
            d, one_line(cites[slug(n)], 80), ": " + one_line(r, 200) if r else "") if slug(n) in ticked_now else "%s (%s)%s" % (
            d, STAYS_OPEN[d], ": " + one_line(r, 200) if r else "")) for n, d, r in entries if n)
    else:
        _, decision, reason = entries[0]
        w = WHERE.search(issue.get("body") or "")
        bad = unverified(ctx, binary, repo, pr, files, decision, reason, w.group(1) if w else "", issue.get("created_at") or "",
                         origin_pr(issue))
        if bad:
            decision, reason = "NOT VERIFIED", bad
        if decision in CLOSE_AS and issue.get("state") == "open":
            batch_write(binary, "PATCH", endpoint, [("state", "closed"), ("state_reason", CLOSE_AS[decision])])
        detail = "%s%s. %s" % (decision, ": " + STAYS_OPEN[decision] if decision in STAYS_OPEN else "", one_line(reason, 200))
    text = "%s\nSettled by %s (review batch #%s): %s" % (mark, by, batch_no or "?", detail)
    batch_write(binary, "POST", endpoint + "/comments", [("body", text)])
    print("findings settle: #%s settled" % number)


def settle_pr(ctx, binary, repo, info):
    """Settle the sources of a merged batch fix pull request, then close its batch issue (last). False when the pull
    request is not a merged fix/review/batch-* pull request into develop of this repository (nothing is written)."""
    head = info.get("head") or {}
    head_repo = (head.get("repo") or {}).get("full_name")
    if (not info.get("merged_at") or (info.get("base") or {}).get("ref") != "develop"
            or not (head.get("ref") or "").startswith(BATCH_PREFIX) or head_repo != repo):
        print("findings settle: #%s is not a merged batch fix pull request into develop: nothing to settle" % info.get("number"))
        return False
    pr, body = info.get("number"), info.get("body") or ""
    m = re.search(r"^Batch:\s*#(\d+)", body, re.M)
    batch_no = m.group(1) if m else None
    # Only the items of the linked batch issue (its batch-item markers, written by the batch run) are settled: a row
    # naming any other issue, added by the model or by an edit of the body, is refused.
    batch_issue = (ctx.gh_api(binary, "repos/%s/issues/%s" % (repo, batch_no)) or [{}])[0] if batch_no else {}
    rows = {}
    if BATCH_LABEL in names(batch_issue.get("labels")) and BATCH_MARK in (batch_issue.get("body") or ""):
        for line in (batch_issue.get("body") or "").splitlines():
            mm = BATCH_LINE.match(line)
            if mm:
                also = ALSO.search(line)
                rows[mm.group(3)] = [k for k in (also.group(1).split(",") if also else []) if ITEM_KEY_RE.match(k)]
    items = set(rows)
    if not items:
        print("findings settle: #%s names no batch issue with items: nothing to settle" % pr)
        return False
    groups = {}
    for key, decision, reason in settle_rows(body):
        if key not in items:
            print("findings settle: %s is not an item of batch #%s: refused" % (key, batch_no))
            continue
        # A grouped row settles every source it lists (read from the batch issue, never from the table of the pull request).
        for source in [key] + rows[key]:
            number, _, note = source[1:].partition("/")
            groups.setdefault(number, []).append((note or None, decision, reason))
    try:
        files = set()
        for f in ctx.gh_api(binary, "repos/%s/pulls/%s/files?per_page=100" % (repo, pr)):
            files |= {f.get(k) for k in ("filename", "previous_filename") if f.get(k)}
    except Exception:  # unreadable diff: no FIXED row can be verified, so none closes
        files = None
    for number, entries in groups.items():
        settle_issue(ctx, binary, repo, pr, batch_no, number, entries, files)
    if batch_no:
        issue = (ctx.gh_api(binary, "repos/%s/issues/%s" % (repo, batch_no)) or [{}])[0]
        if issue.get("state") == "open" and BATCH_LABEL in names(issue.get("labels")):
            text = (issue.get("body") or "") + "\nClosed: its fix pull request #%s is merged into develop and its sources are settled.\n" % pr
            batch_write(binary, "PATCH", "repos/%s/issues/%s" % (repo, batch_no),
                        [("body", text), ("state", "closed"), ("state_reason", "completed")])
            print("findings settle: batch #%s closed after its sources" % batch_no)
    print("findings settle: #%s done (%d source issue(s))" % (pr, len(groups)))
    return True


def run_settle(binary, repo, pr):
    """Fail CLOSED like the batch: a GitHub error is FAILED with exit 1."""
    if not repo or not re.match(r"^\d+$", pr or ""):
        print("findings settle: FAILED (REPO and a pull request number are required)")
        return 2
    try:
        ctx = load_context_module()
        info = (ctx.gh_api(binary, "repos/%s/pulls/%s" % (repo, pr)) or [{}])[0]
        settle_pr(ctx, binary, repo, info)
        return 0
    except Exception as exc:  # fail closed
        reason = str(exc) if isinstance(exc, BatchError) or type(exc).__name__ == "Unavailable" else type(exc).__name__
        print("findings settle: FAILED (%s); the next run settles what is left" % reason)
        return 1


def batch(ctx, binary, repo, write):
    """Create, update or close the one open review batch. Returns the exit code."""
    findings = list_issues(ctx, binary, repo)
    notes = list_issues(ctx, binary, repo, "kind:note")
    all_batches = list_issues(ctx, binary, repo, BATCH_LABEL)
    batches = sorted((i for i in all_batches if i.get("state") == "open" and BATCH_MARK in (i.get("body") or "")), key=lambda i: i.get("number", 0))
    ranked = group_items(batch_items(findings, notes))
    current = {it["key"]: it for it in ranked}
    by_source = {src: it["key"] for it in ranked for src in it["sources"]}
    # Items with a verdict left open for the owner by a no-commit batch (batch-handled) never block the next batch: they
    # stay out of new batches (and never join one as urgent) until they change.
    record = held_items(all_batches)
    waiting = [it for it in ranked if not any(record.get(src) == it["text"] for src in it["sources"])]
    held = len(ranked) - len(waiting)
    if len(batches) > 1:
        print("findings batch: %d open batch issues, the oldest (#%s) is kept up to date" % (len(batches), batches[0].get("number")))
    if not batches:
        if not waiting:
            print("findings batch: nothing open%s, no batch" % (" but %d item(s) with a verdict held for the owner" % held if held else ""))
            return 0
        title = "Review batch %s" % now().strftime("%Y-%m-%d")
        members = waiting[:BATCH_SIZE]
        body = render_batch(title, [(it["key"], False, it["text"], it["also"]) for it in members], len(waiting) - len(members), held)
        if not write:
            print("findings batch: would create \"%s\" (%d items, %d more)\n\n%s" % (title, len(members), len(waiting) - len(members), body))
            return 0
        batch_write(binary, "POST", "repos/%s/issues" % repo, [("title", title), ("body", body), ("labels[]", BATCH_LABEL)])
        print("findings batch: created \"%s\" (%d items, %d more)" % (title, len(members), len(waiting) - len(members)))
        return 0

    issue = batches[0]
    number, title = issue.get("number"), issue.get("title", "")
    previous = []
    for line in (issue.get("body") or "").splitlines():
        m = BATCH_LINE.match(line)
        if m:
            also = ALSO.search(line)
            previous.append((m.group(3), m.group(1) != " ", ALSO.sub("", m.group(2)),
                             [k for k in (also.group(1).split(",") if also else []) if ITEM_KEY_RE.match(k)]))
    known = {src for key, _, _, also in previous for src in [key] + also}
    # Membership is frozen when the batch is created, so a batch can finish; only an urgent item joins it later. A row
    # follows its sources: when the oldest source of a grouped row closes, the row is keyed by the next one.
    joined = [it for it in waiting if not known.intersection(it["sources"]) and it["rank"][1] <= SEV_RANK["HIGH"]]
    open_map, done_lines = {}, []
    for key, ticked, text, also in previous:
        target = next((by_source[src] for src in [key] + also if src in by_source), None)
        if target is None:
            done_lines.append((key, True, text if text.endswith("(done)") else text + " (done)", also))
        elif target in open_map:
            open_map[target][2] = open_map[target][2] or ticked
        else:
            open_map[target] = [current[target]["rank"], target, ticked, current[target]["text"], current[target]["also"]]
    open_lines = sorted(list(open_map.values()) + [[it["rank"], it["key"], False, it["text"], it["also"]] for it in joined],
                        key=lambda row: (row[0], row[1]))
    lines = [(key, ticked, text, also) for _, key, ticked, text, also in open_lines] + done_lines
    members = {key for key, _, _, _ in lines}
    rest = sum(1 for it in waiting if it["key"] not in members)
    body = render_batch(title, lines, rest, sum(1 for it in ranked if it not in waiting and it["key"] not in members))
    endpoint = "repos/%s/issues/%s" % (repo, number)
    fix_pr = merged_fix(ctx, binary, repo, issue) if open_lines else None
    fix = fix_pr.get("number") if fix_pr else None
    # A merged batch fix pull request settles its sources first and closes the batch itself (the sources before the
    # batch); another pull request that names the batch only closes it, as before.
    if fix_pr and write and settle_pr(ctx, binary, repo, fix_pr):
        return 0
    if not open_lines or fix:
        why = "its fix pull request #%s is merged into develop" % fix if fix else "every source is closed or decided"
        body += "\nClosed: %s.\n" % why
        if not write:
            print("findings batch: would close #%s as completed (%s)\n\n%s" % (number, why, body))
            return 0
        batch_write(binary, "PATCH", endpoint, [("body", body), ("state", "closed"), ("state_reason", "completed")])
        print("findings batch: closed #%s as completed (%s); the next run starts a new batch" % (number, why))
        return 0
    if body == issue.get("body"):
        print("findings batch: #%s unchanged (%d open, %d done, %d more)" % (number, len(open_lines), len(done_lines), rest))
        return 0
    if not write:
        print("findings batch: would update #%s (%d open, %d done, %d more)\n\n%s" % (number, len(open_lines), len(done_lines), rest, body))
        return 0
    batch_write(binary, "PATCH", endpoint, [("body", body)])
    print("findings batch: updated #%s (%d open, %d done, %d more)" % (number, len(open_lines), len(done_lines), rest))
    return 0


def run_batch(binary, repo, write):
    """Fail CLOSED: an unreadable or refused GitHub call is FAILED with exit 1, never a partial batch."""
    if not repo:
        print("findings batch: FAILED (REPO is required: owner/name)")
        return 2
    try:
        ctx = load_context_module()
        return batch(ctx, binary, repo, write)
    except Exception as exc:  # fail closed: any failure is reported, never a silent partial batch
        reason = str(exc) if isinstance(exc, BatchError) or type(exc).__name__ == "Unavailable" else type(exc).__name__
        print("findings batch: FAILED (gh could not be read or written: %s); nothing was changed after the failure" % reason)
        return 1


def main():
    args = sys.argv[1:]
    repo = os.environ.get("REPO", "")
    binary = os.environ.get("GH_BIN", "gh")
    if args and args[0] == "batch":
        return run_batch(binary, repo, write="--print" not in args)
    if args and args[0] == "settle":
        return run_settle(binary, repo, args[1] if len(args) > 1 else "")
    if not args or args[0] not in ("harvest", "dashboard") or not repo:
        print("usage: findings.py harvest|dashboard|batch [--print]|settle PR   (REPO required; PR_NUMBER for harvest)")
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
            print(render_dashboard(list_issues(ctx, binary, repo), list_issues(ctx, binary, repo, "kind:note")))
        else:
            write_dashboard(ctx, binary, repo, create=True)
    except Exception as exc:  # fail open: the register must never fail a workflow
        print("findings: NOT RUN (%s)" % type(exc).__name__)
    return 0


if __name__ == "__main__":
    sys.exit(main())
