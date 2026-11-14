#!/usr/bin/env python3
"""Write the AI review comment of ONE RUN on a pull request (stdlib only), deterministically.

The review returns its text; this script, run by the workflow from the BASE branch, posts it:
  - line 1 stays the `Triage:` verdict (the Review gate reads it), line 2 is the hidden marker MARKER, line 3 the fixed
    `Tool:` line of the preview step (a `Tool:` line written by the model is dropped), line 4
    `Reviewed: <head sha7> (<range>) at <HH:MM> UTC` (REVIEW_RANGE, the time it is written; REVIEW_NOW for tests);
  - every run CREATES a new comment (the timeline stays visible) and writes its id to GITHUB_OUTPUT (comment_id); then
    every OLDER bot review comment (marker, or a `Triage:` first line) is minimized as OUTDATED, never deleted, never a
    human's comment. Every reader takes the NEWEST CREATED marker comment;
  - the new ledger starts from the previous comment's: entries the model left out keep their id and status (fixed
    ones are dropped);
  - the ledger `sha` is always PR_HEAD_SHA (validated), never what the model wrote. A missing, bare (no <details>),
    hidden-comment or broken ledger is dropped, never replaced by an empty one at the head: the next run is `first`.
No usable review text (empty, invalid, timed out, no `Triage:` first line): the marker comment becomes
`Triage: NOT RUN | <reason>` and carries the PREVIOUS ledger unchanged (or none), so nothing looks reviewed and the
commits not reviewed are reviewed on the next run. The Review gate reads no verdict from it, except after a CRITICAL:
then line 1 stays the CRITICAL verdict and its evidence is kept, so a push that exhausts the review never lifts a block.
Line 4 is `Deep: PENDING (...)` when DEEP_PENDING is true (a P0 path is touched; the deep job replaces it).

Ledger entries (the register detail): every entry the model writes carries `detail` (the proof and reasoning, at most
600 characters) and `fix` (one line, at most 300; for a note it may be "none: accept" with a reason). A missing or
blank one is filled with "not recorded by the review" and the comment names the ids on line 2 of the text: the Triage
level and every entry stay (a missing field never lowers or hides a verdict; the deep review does the same and keeps
its level, an invalid deep entry is dropped and named, the level stays). The script, never the model, sets `source` (review or deep),
`first_sha` (the head the entry was first raised at, kept when the entry is listed again with the same file and line or
carried forward), `location` (verified, mismatch or unverified: the file exists at that commit, the line is inside it and an identifier the text names in backticks is within 40 lines of it; `mismatch` when none is but one is elsewhere, `probable` then gives that file:line; read as
git objects from LOCATION_REPO, else the current directory, else `pr/`) and `redacted`. `detail`, `fix` and the
scenario are cut, neutralised (no @mention, `#N` disarmed with a zero-width space, no HTML, one line) and replaced by
"[redacted: looked like a secret]" when they match a secret shape.

Subcommand `deep` (stage 2): merges the deep review's structured output (DEEP_JSON: level, blast_radius, summary,
findings) into the comment and ledger of the SAME run (REVIEW_COMMENT_ID, else the marker comment whose ledger sha is
PR_HEAD_SHA); when a newer review comment exists it edits nothing (NOT RUN, stale). A new entry is tagged `"source":"deep"`; when the deep review repeats an id already in the stage 1 ledger, the stage 1
entry keeps `"source":"review"`, gains a `deep` sub-field with the deep detail, fix and location, and takes the worse
severity of the two (never overwritten); line 4 becomes `Deep: PASS (...)` or `Deep: NOT RUN (<reason>)`, the Triage line becomes the worse of both levels
(a NOT RUN verdict stays NOT RUN), the ledger records `"deep":{"status":"PASS"|"NOT RUN","sha":PR_HEAD_SHA}` and keeps
its own sha, and the merge is idempotent.
Subcommand `deep-needed`: prints `true` when the pull request touches a P0 path (.github/sensitive-paths.txt plus
.github/deep-review-paths.txt under ROOT, default the repository of this script), else `false` (also on any error).
Fails OPEN: a GitHub error prints NOT RUN and exits 0, so it never turns the pull request red.

Environment: REPO (owner/name), PR_NUMBER, REVIEW_JSON (the review's structured output, {"body": "..."}) or
REVIEW_BODY_FILE (a file with the text, for a local run), TOOL_LINE (the preview step's `Tool:` line), PR_HEAD_SHA (the
pull request head, 40 hex characters), REVIEW_OUTCOME (the review step's outcome, for the NOT RUN reason), DEEP_PENDING,
DEEP_JSON (subcommand deep), ROOT (subcommand deep-needed), GH_BIN.
"""
import datetime
import fnmatch
import json
import os
import re
import subprocess
import sys
import tempfile

BOT_LOGINS = {"github-actions[bot]", "github-actions"}
MARKER = "<!-- 226os-ai-review -->"
VERDICT = re.compile(r"^Triage:\s*(CRITICAL|HIGH|MEDIUM|NONE)\s*\|\s*blast radius:\s*\S.*$")
TOOL = re.compile(r"^Tool: (PASS \(\d+ files?\)|NOT RUN \([^()\n]{1,200}\))$")
FULL_SHA = re.compile(r"^[0-9a-f]{40}$")
FENCED = re.compile(r"(<details>\s*<summary>[^<]*ledger[^<]*</summary>.*?```json\s*)(\{.*?\})(\s*```)", re.S | re.I)
HIDDEN = re.compile(r"<!--\s*review-ledger\s+\{.*?\}\s*-->", re.S)
LEDGER_BLOCK = re.compile(r"<details>\s*<summary>[^<]*ledger[^<]*</summary>.*?```json\s*\{.*?\}\s*```\s*</details>", re.S | re.I)
NOT_RUN = re.compile(r"^Triage:\s*NOT RUN\b")
DEEP_START = "<!-- 226os-deep-review -->"
DEEP_END = "<!-- /226os-deep-review -->"
ANY_LEDGER = re.compile(r"<details>\s*<summary>[^<]*ledger[^<]*</summary>.*?</details>", re.S | re.I)
LEDGER_JSON = re.compile(r"<details>\s*<summary>[^<]*ledger[^<]*</summary>.*?```json\s*(\{.*?\})\s*```", re.S | re.I)
DEEP_SECTION = re.compile(re.escape(DEEP_START) + r".*?" + re.escape(DEEP_END), re.S)
DEEP_PENDING_LINE = "Deep: PENDING (a P0 path is touched: the deep review runs next)"
LEVELS = ["NONE", "MEDIUM", "HIGH", "CRITICAL"]
SEVERITIES = {"CRITICAL", "HIGH", "MEDIUM", "LOW"}
STATUSES = {"open", "answered", "accepted", "fixed"}
P0_LISTS = (".github/sensitive-paths.txt", ".github/deep-review-paths.txt")
DEEP_MAX_ENTRIES = 8
SEV_RANK = {"LOW": 0, "MEDIUM": 1, "HIGH": 2, "CRITICAL": 3}
STATUS_RANK = {"open": 0, "answered": 1, "accepted": 1, "fixed": 2}
REDACTED = "[redacted: looked like a secret]"
# Secret shapes (private key blocks, AWS, GitHub and Slack tokens, sk- keys, JWT, an Authorization header WITH a value).
# The same list is in scripts/review-context.py: this script must stay one standalone file (the workflow copies it alone).
SECRET_SHAPES = [re.compile(p) for p in (
    r"-----BEGIN [A-Z ]*PRIVATE KEY-----",
    r"\b(?:AKIA|ASIA|AGPA|AIDA|AROA|ANPA)[0-9A-Z]{16}\b",
    r"\bgh[pousr]_[A-Za-z0-9]{30,}",
    r"\bgithub_pat_[A-Za-z0-9_]{20,}",
    r"\bxox[abposr]-[A-Za-z0-9-]{10,}",
    r"\bsk-[A-Za-z0-9_-]{20,}",
    r"\beyJ[A-Za-z0-9_-]{8,}\.eyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}",
    r"(?i)\bauthorization\s*[:=]\s*[\"']?(?:(?:bearer|basic|token)\s+)?[A-Za-z0-9._~+/=-]{16,}",
)]
ZWSP = chr(0x200B)
DETAIL_MAX = 600
FIX_MAX = 300
LOCATE_MAX_BYTES = 5000000
SCRIPT_OWNED = ("source", "first_sha", "location", "probable", "redacted", "deep", "carried")
MINIMIZE = "mutation($id: ID!) { minimizeComment(input: {subjectId: $id, classifier: OUTDATED}) { minimizedComment { isMinimized } } }"


class GitHubError(Exception):
    """GitHub could not be read or written: the step says NOT RUN and goes on."""


def gh(binary, args, payload=None):
    path = None
    try:
        if payload is not None:
            fd, path = tempfile.mkstemp(suffix=".json")
            with os.fdopen(fd, "w", encoding="utf-8") as fh:
                json.dump(payload, fh)
            args = args + ["--input", path]
        done = subprocess.run([binary] + args, capture_output=True, text=True, timeout=60, check=False)
    except (OSError, subprocess.SubprocessError) as exc:
        raise GitHubError(str(exc)) from exc
    finally:
        if path:
            os.unlink(path)
    if done.returncode != 0:
        raise GitHubError("gh %s failed" % " ".join(a for a in args[:3] if not a.startswith("{")))
    return done.stdout


def parse_pages(text):
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


def first_line(body):
    body = (body or "").strip()
    return body.splitlines()[0] if body else ""


def is_bot(c):
    return (c.get("user") or {}).get("login") in BOT_LOGINS


def created_key(c):
    """Order of creation: created_at, then the id (ids grow)."""
    try:
        cid = int(c.get("id") or 0)
    except (TypeError, ValueError):
        cid = 0
    return (c.get("created_at") or c.get("updated_at") or "", cid)


def review_comment(comments):
    """The current review comment: the NEWEST CREATED bot comment with the marker (one comment per run, older ones are
    minimized), else the newest updated bot comment whose first line is a `Triage:` verdict (a pull request reviewed
    before the marker existed), else None."""
    stamp = lambda c: c.get("updated_at") or c.get("created_at") or ""
    bots = [c for c in comments if is_bot(c)]
    marked = [c for c in bots if MARKER in (c.get("body") or "")]
    if marked:
        return max(marked, key=created_key)
    triage = [c for c in bots if VERDICT.match(first_line(c.get("body")))]
    return max(triage, key=stamp) if triage else None


def looks_secret(text):
    return any(p.search(text or "") for p in SECRET_SHAPES)


def defuse(text, limit=None):
    """One line that cannot act when written to an issue or a prompt: no HTML tag or comment (a forged marker must not
    parse), `<` escaped, an @mention and a #N reference broken with a zero-width space, whitespace collapsed."""
    text = " ".join(str(text or "").split())
    text = re.sub(r"<!--.*?-->|</?[A-Za-z!][^>]*>", "", text)
    text = text.replace("<!--", "").replace("<", "&lt;")
    text = re.sub(r"@(?=[A-Za-z0-9_-])", "@" + ZWSP, text)
    text = re.sub(r"#(?=\d)", "#" + ZWSP, text)
    text = " ".join(text.split())
    return text if limit is None or len(text) <= limit else text[:limit - 3] + "..."


def clean_field(text, limit):
    """(text, redacted): a secret shape replaces the whole text; otherwise it is defused and cut."""
    raw = " ".join(str(text or "").split())
    if looks_secret(raw):
        return REDACTED, True
    return defuse(raw, limit), False


def slug_id(identifier):
    """An id as it may appear in a verdict line: letters, digits, underscore and dash only."""
    return re.sub(r"[^\w-]+", "-", str(identifier or "?"))[:40] or "?"


def is_blank(value):
    return not isinstance(value, str) or not value.strip()


MISSING_FIELD = "not recorded by the review"


def fill_missing(text):
    """(text, ids): the review text with every blank `detail` or `fix` of its fenced ledger entries set to MISSING_FIELD,
    and the slug ids that were filled. A missing field never lowers or hides the verdict: the Triage line and every entry
    stay, the comment says which ids lack text. No usable ledger: the text is returned as it is."""
    m = FENCED.search(text or "")
    if not m:
        return text, []
    try:
        data = json.loads(m.group(2))
    except ValueError:
        return text, []
    ids = []
    for f in data.get("findings") if isinstance(data, dict) and isinstance(data.get("findings"), list) else []:
        if isinstance(f, dict) and (is_blank(f.get("detail")) or is_blank(f.get("fix"))):
            ids.append(slug_id(f.get("id")))
            for key in ("detail", "fix"):
                if is_blank(f.get(key)):
                    f[key] = MISSING_FIELD
    if not ids:
        return text, []
    return text[:m.start(2)] + json.dumps(data) + text[m.end(2):], ids


def missing_notice(ids):
    return "Detail or fix not recorded by the review (the verdict is unchanged): %s." % ", ".join("`%s`" % i for i in ids)


def location_dirs():
    env = os.environ.get("LOCATION_REPO")
    return [env] if env else [".", "pr"]


BACKTICKED = re.compile(r"`([^`\n]{3,80})`")
WINDOW = 40  # lines each side of the cited line: a function or step the text names is within about a screen of it


def named_identifiers(text, path):
    """The identifiers a text names in backticks (function, variable, step, file names): no spaces (a command is not an
    identifier), a trailing `()` or `:line` removed, never the cited file itself."""
    out = []
    for token in BACKTICKED.findall(text or ""):
        token = re.sub(r"(\(\)|:\d+(-\d+)?)$", "", token.strip())
        if len(token) >= 3 and not re.search(r"\s", token) and token != path and token != os.path.basename(path) and token not in out:
            out.append(token)
    return out[:8]


def probable_location(head, path, idents):
    """`file:line` of the first identifier found in the commit (outside the window, since none was inside it) `head` (git grep, fixed strings), or ''."""
    for folder in location_dirs():
        for ident in idents:
            try:
                done = subprocess.run(["git", "-C", folder, "grep", "-n", "-F", "-I", "-m", "1", "-e", ident, head, "--"],
                                      capture_output=True, text=True, timeout=20, check=False)
            except (OSError, subprocess.SubprocessError):
                return ""
            for hit in done.stdout.splitlines():
                m = re.match(r"^[0-9a-f]{40}:(.+?):(\d+):", hit)
                if m and not re.search(r"[\x00-\x1f]", m.group(1)):
                    return "%s:%s" % (m.group(1), m.group(2))
    return ""


def locate(head, path, line, text=""):
    """(location, probable). `verified`: `path` exists at the commit `head`, `line` is inside it (no line: the file only)
    and, when `text` names identifiers in backticks, at least one appears within WINDOW lines of the cited line (anywhere
    in the file when there is no line). `mismatch`: the file and line exist but none of the named identifiers is near
    them and one is elsewhere (`probable` is that file:line). `unverified`: no file, no line, no repository to
    read, or the identifiers are nowhere. The commit is read as git objects (cat-file, an argument list, a size cap):
    nothing of the pull request is executed or checked out."""
    if (not FULL_SHA.match(head or "") or not isinstance(path, str) or not path or len(path) > 300 or path.startswith(("/", "-"))
            or ".." in path.split("/") or re.search(r"[\x00-\x1f]", path)):
        return "unverified", ""
    line = line if isinstance(line, int) and not isinstance(line, bool) else 0
    spec = "%s:%s" % (head, path)
    for folder in location_dirs():
        try:
            size = subprocess.run(["git", "-C", folder, "cat-file", "-s", spec], capture_output=True, text=True, timeout=20, check=False)
            if size.returncode != 0:
                continue
            if int(size.stdout.strip() or "0") > LOCATE_MAX_BYTES:
                return "unverified", ""
            blob = subprocess.run(["git", "-C", folder, "cat-file", "blob", spec], capture_output=True, timeout=20, check=False)
        except (OSError, subprocess.SubprocessError, ValueError):
            return "unverified", ""
        if blob.returncode != 0:
            continue
        lines = blob.stdout.decode("utf-8", "replace").split("\n")
        if lines and lines[-1] == "":
            lines.pop()
        if not 0 <= line <= len(lines):
            return "unverified", ""
        idents = named_identifiers(text, path)
        if not idents:
            return "verified", ""
        lo, hi = (max(0, line - 1 - WINDOW), line + WINDOW) if line else (0, len(lines))
        near = "\n".join(lines[lo:hi])
        if any(i in near for i in idents):
            return "verified", ""
        elsewhere = probable_location(head, path, idents)
        return ("mismatch", elsewhere) if elsewhere else ("unverified", "")
    return "unverified", ""


def refine_entry(f, previous, head):
    """The model's entry with the fields the script owns: cleaned detail, fix and scenario; source, first_sha, location
    and the deep sub-field from the previous ledger when it listed the same id (first_sha only while the file and line are
    the same: the permalink must keep pointing at the right place), else the head."""
    for key in SCRIPT_OWNED:
        f.pop(key, None)
    f["detail"], hit1 = clean_field(f.get("detail"), DETAIL_MAX)
    f["fix"], hit2 = clean_field(f.get("fix"), FIX_MAX)
    hit3 = isinstance(f.get("scenario"), str) and looks_secret(f["scenario"])
    if hit3:
        f["scenario"] = REDACTED
    if hit1 or hit2 or hit3:
        f["redacted"] = True
    previous = previous if isinstance(previous, dict) else {}
    f["source"] = previous.get("source") if previous.get("source") in ("review", "deep") else "review"
    if isinstance(previous.get("deep"), dict):
        f["deep"] = previous["deep"]
    text = "%s %s" % (f.get("scenario") or "", f.get("detail") or "")
    if FULL_SHA.match(str(previous.get("first_sha") or "")) and previous.get("file") == f.get("file") and previous.get("line") == f.get("line"):
        f["first_sha"] = previous["first_sha"]
    else:
        f["first_sha"] = head
        previous = {}
    f["location"], probable = locate(f["first_sha"], f.get("file"), f.get("line"), text)
    if f["location"] == "unverified" and previous.get("location") in ("verified", "mismatch"):
        # The reviewed commit may no longer be in this checkout (a force push): keep what was established before.
        f["location"], probable = previous["location"], previous.get("probable", "")
    if probable:
        f["probable"] = probable
    return f


def pin_ledger(text, head, previous_ledger=None):
    """The text with its fenced ledger's sha set to head (findings kept). Anything else is no usable ledger and nothing
    is invented: a hidden-comment ledger (it would carry the model's sha) is removed, a bare or broken one is left as
    text the parser ignores, and none is added. Unchanged when head is not a full sha (a local run)."""
    if not FULL_SHA.match(head or ""):
        return text
    text = HIDDEN.sub("", text)
    m = FENCED.search(text)
    if m:
        try:
            data = json.loads(m.group(2))
        except ValueError:
            data = None
        if isinstance(data, dict):
            data.pop("last_reviewed_commit", None)
            data.pop("commit", None)
            # Only the script writes `deep`, `source`, `first_sha`, `location` and `redacted`: a model-written one would
            # forge a deep review PASS or a permalink.
            data.pop("deep", None)
            before = {x.get("id"): x for x in (previous_ledger or {}).get("findings", []) if isinstance(x, dict)}
            for f in data.get("findings") if isinstance(data.get("findings"), list) else []:
                if isinstance(f, dict):
                    refine_entry(f, before.get(f.get("id")), head)
            data["sha"] = head
            return text[:m.start(2)] + json.dumps(data) + text[m.end(2):]
    return text


def tool_or_default(tool_line):
    return tool_line if TOOL.match(tool_line or "") else "Tool: NOT RUN (the preview step gave no result)"


def reviewed_line(head, rng, done=True):
    """Line 4: what this run reviewed (head, range from the pack, both shortened) and when the comment was written."""
    sha7 = head[:7] if FULL_SHA.match(head or "") else "unknown"
    rng = re.sub(r"\b([0-9a-f]{7})[0-9a-f]{33}\b", r"\1", one_line(rng, 120)) or "range unknown"
    now = os.environ.get("REVIEW_NOW") or datetime.datetime.now(datetime.timezone.utc).strftime("%H:%M")
    if done:
        return "Reviewed: %s (%s) at %s UTC" % (sha7, rng, now)
    return "Reviewed: none, NOT RUN on %s (%s) at %s UTC" % (sha7, rng, now)


def carry_forward(text, previous_ledger):
    """The previous comment's ledger entries the model left out are carried into the new ledger with their id and
    status (fixed ones are dropped: they were reported fixed already). Nothing is added when the text has no usable
    ledger (no ledger is ever invented)."""
    if not isinstance(previous_ledger, dict):
        return text
    m = FENCED.search(text)
    if not m:
        return text
    try:
        data = json.loads(m.group(2))
    except ValueError:
        return text
    if not isinstance(data, dict) or not isinstance(data.get("findings"), list):
        return text
    ids = {f.get("id") for f in data["findings"] if isinstance(f, dict)}
    carried = []
    for f in previous_ledger.get("findings", []):
        if isinstance(f, dict) and f.get("id") and f.get("id") not in ids and f.get("status") != "fixed":
            f = dict(f)
            if f.get("status") == "open":
                # Still open and not mentioned by this review: the model made no new decision. Counted and named in the
                # comment, so a stale entry that keeps blocking auto-merge is visible and a person can settle it.
                f["carried"] = int(f.get("carried") or 0) + 1 if isinstance(f.get("carried"), int) else 1
                carried.append(slug_id(f["id"]))
            data["findings"].append(f)
    out = text[:m.start(2)] + json.dumps(data) + text[m.end(2):]
    if carried:
        notice = "Carried forward without a new decision (this review did not mention them: fix, answer or accept them): %s.\n\n" % ", ".join("`%s`" % c for c in carried)
        out = out[:m.start(1)] + notice + out[m.start(1):]
    return out


def effective_level(declared, entries):
    """The verdict level the ledger supports. Never the model's declared level alone: a CRITICAL needs an open OBSERVED
    CRITICAL finding with file and line (else it is at most HIGH, the rule the gate and review-ping apply), and an open
    finding raises the level to its own severity (a CRITICAL without that evidence counts as HIGH). Notes never count."""
    top = 0
    qualified = False
    for f in entries:
        if not isinstance(f, dict) or str(f.get("kind", "finding")).lower() == "note" or f.get("status") != "open":
            continue
        sev = str(f.get("sev", "MEDIUM")).upper()
        line = f.get("line")
        if sev == "CRITICAL":
            if str(f.get("evidence", "")).upper() == "OBSERVED" and f.get("file") and isinstance(line, int) and not isinstance(line, bool) and line > 0:
                qualified = True
            else:
                sev = "HIGH"
        top = max(top, {"MEDIUM": 1, "HIGH": 2, "CRITICAL": 3}.get(sev, 0))
    level = declared
    if level == "CRITICAL" and not qualified:
        level = "HIGH"
    return max(level, LEVELS[top], key=LEVELS.index)


def pin_level(text):
    """The text with its first line's level set to the one its own ledger supports (blast radius text kept)."""
    m = FENCED.search(text or "")
    first = (text or "").split("\n", 1)[0]
    v = VERDICT.match(first.strip())
    if not m or not v:
        return text
    try:
        data = json.loads(m.group(2))
    except ValueError:
        return text
    if not isinstance(data, dict) or not isinstance(data.get("findings"), list):
        return text
    level = effective_level(v.group(1), data["findings"])
    if level == v.group(1):
        return text
    return first.replace(v.group(1), level, 1) + text[len(first):]


def compose(text, tool_line, head="", deep_line=None, reviewed=None, previous_ledger=None):
    """The comment body: verdict, marker, Tool line, then the review without any marker or Tool line of its own, its
    ledger pinned to head. None when the text does not start with a valid verdict (the gate could not read it)."""
    lines = (text or "").strip().splitlines()
    if not lines or not VERDICT.match(lines[0].strip()):
        return None
    rest = [l for l in lines[1:] if MARKER not in l and not l.startswith(("Tool:", "Deep:", "Reviewed:"))]
    while rest and not rest[0].strip():
        rest.pop(0)
    top = [lines[0].strip(), MARKER, tool_or_default(tool_line)] + ([reviewed] if reviewed else []) + ([deep_line] if deep_line else [])
    text = pin_ledger("\n".join(top + [""] + rest), head, previous_ledger)
    if FULL_SHA.match(head or ""):
        text = carry_forward(text, previous_ledger)
    return pin_level(text) + "\n"


def not_run(reason, tool_line, previous, reviewed=None):
    """The NOT RUN marker comment: no verdict the gate reads, and the previous review comment's ledger block unchanged
    (none when it had none), so the unreviewed commits stay unreviewed in the ledger."""
    note = ("The AI review produced no result on this run (%s): nothing in the new commits was reviewed. Push again or "
            "add the label `ai-review` to retry." % reason)
    old = ((previous or {}).get("body") or "").strip().splitlines()
    m = VERDICT.match(old[0].strip()) if old else None
    if m and m.group(1) == "CRITICAL":
        # A CRITICAL is never lifted by a review that did not run: keep the verdict and its evidence, add the reason.
        rest = [l for l in old[1:] if MARKER not in l and not l.startswith(("Tool:", "Reviewed:", "Deep:"))]
        while rest and not rest[0].strip():
            rest.pop(0)
        return "\n".join([old[0].strip(), MARKER, tool_or_default(tool_line)] + ([reviewed] if reviewed else [])
                         + ["", "NOT RUN on the latest push. " + note, ""] + rest) + "\n"
    m = LEDGER_BLOCK.search((previous or {}).get("body") or "")
    body = ["Triage: NOT RUN | %s" % reason, MARKER, tool_or_default(tool_line)] + ([reviewed] if reviewed else []) + ["", note]
    if m:
        body += ["", m.group(0)]
    return "\n".join(body) + "\n"


def review_text():
    path = os.environ.get("REVIEW_BODY_FILE")
    if path:
        try:
            with open(path, encoding="utf-8") as fh:
                return fh.read()
        except OSError:
            return None
    try:
        data = json.loads(os.environ.get("REVIEW_JSON") or "")
    except ValueError:
        return None
    body = data.get("body") if isinstance(data, dict) else None
    return body if isinstance(body, str) else None


def main():
    repo, pr = os.environ.get("REPO", ""), os.environ.get("PR_NUMBER", "")
    binary = os.environ.get("GH_BIN", "gh")
    if not repo or not re.match(r"^\d+$", pr):
        print("review-comment: NOT RUN (REPO and PR_NUMBER are required)")
        return 0
    tool_line = os.environ.get("TOOL_LINE", "")
    head = os.environ.get("PR_HEAD_SHA", "")
    rng = os.environ.get("REVIEW_RANGE", "")
    text = review_text()
    deep_line = DEEP_PENDING_LINE if os.environ.get("DEEP_PENDING") == "true" else None
    try:
        comments = parse_pages(gh(binary, ["api", "repos/%s/issues/%s/comments" % (repo, pr), "--paginate"]))
    except (GitHubError, ValueError) as exc:
        print("review-comment: NOT RUN (%s)" % exc)
        return 0
    previous = review_comment(comments)
    first = (text or "").strip().splitlines()[:1]
    filled = []
    if first and VERDICT.match(first[0].strip()):
        text, filled = fill_missing(text)
        if filled:
            lines = text.strip().splitlines()
            text = "\n".join(lines[:1] + [missing_notice(filled), ""] + lines[1:])
    body = compose(text, tool_line, head, deep_line, reviewed_line(head, rng), ledger_of((previous or {}).get("body")))
    if body is None:
        outcome = re.sub(r"[^a-z_]", "", os.environ.get("REVIEW_OUTCOME", "").lower())[:20]
        if text is None or not text.strip():
            reason = "the review step ended with %s and returned no text" % (outcome or "an unknown outcome")
        else:
            reason = "the review text has no Triage first line"
        body = not_run(reason, tool_line, previous, reviewed_line(head, rng, done=False))
        print("review-comment: the review gave no usable text (%s): writing a NOT RUN comment" % reason)
    # One comment PER RUN: always a new comment, then every older bot review comment is collapsed.
    try:
        created = create_and_collapse(binary, repo, pr, comments, body)
    except (GitHubError, ValueError, KeyError, TypeError) as exc:
        print("review-comment: NOT RUN (%s)" % exc)
        return 0
    out = os.environ.get("GITHUB_OUTPUT")
    if out and created.get("id"):
        try:
            with open(out, "a", encoding="utf-8") as fh:
                fh.write("comment_id=%s\n" % int(created["id"]))
        except (OSError, ValueError, TypeError):
            pass
    return 0


def ledger_section(ledger):
    return "<details><summary>Review ledger</summary>\n\n```json\n%s\n```\n\n</details>" % json.dumps(ledger, ensure_ascii=False)


def one_line(text, limit=160):
    text = " ".join(str(text or "").split()).replace("|", "/")
    return text[:limit]


def ledger_of(body):
    """The ledger dict of a comment body (the fenced JSON inside <details>), or None."""
    m = LEDGER_JSON.search(body or "")
    if not m:
        return None
    try:
        data = json.loads(m.group(1))
    except ValueError:
        return None
    if not isinstance(data, dict):
        return None
    if not isinstance(data.get("findings"), list):
        data["findings"] = []
    return data


def is_review_comment(c):
    """A bot review comment of any run: with the marker, or (older pull requests) a `Triage:` first line."""
    return is_bot(c) and (MARKER in (c.get("body") or "") or bool(VERDICT.match(first_line(c.get("body")))))


def create_and_collapse(binary, repo, pr, comments, body):
    """Create the comment of this run, then minimize every OLDER bot review comment as OUTDATED (never delete, never a
    human's comment). Returns the created comment. Raises GitHubError."""
    created = json.loads(gh(binary, ["api", "-X", "POST", "repos/%s/issues/%s/comments" % (repo, pr)], {"body": body}) or "{}")
    created = created if isinstance(created, dict) else {}
    print("review-comment: created comment %s" % created.get("id", "?"))
    for c in comments:
        if c.get("id") == created.get("id") or not c.get("node_id") or not is_review_comment(c):
            continue
        try:
            gh(binary, ["api", "graphql", "-f", "query=%s" % MINIMIZE, "-f", "id=%s" % c["node_id"]])
            print("review-comment: minimized the older review comment %s as outdated" % c.get("id"))
        except GitHubError as exc:
            print("review-comment: could not minimize comment %s (%s)" % (c.get("id"), exc))
    return created


def edit(binary, repo, target, body):
    gh(binary, ["api", "-X", "PATCH", "repos/%s/issues/comments/%s" % (repo, int(target["id"]))], {"body": body})
    print("review-comment: edited comment %s" % target["id"])


def fetch_comments(binary, repo, pr):
    return parse_pages(gh(binary, ["api", "repos/%s/issues/%s/comments" % (repo, pr), "--paginate"]))


def target_env():
    repo, pr = os.environ.get("REPO", ""), os.environ.get("PR_NUMBER", "")
    if not repo or not re.match(r"^\d+$", pr):
        return None
    return repo, pr, os.environ.get("GH_BIN", "gh")


# --- stage 2: the deep review merges into the same comment and ledger ---

def deep_entry(f, head=""):
    """One deep ledger entry, normalised and tagged source deep, or None when it is not valid."""
    if not isinstance(f, dict) or not isinstance(f.get("id"), str) or not f["id"].strip() or f.get("status") not in STATUSES:
        return None
    kind = "note" if str(f.get("kind", "finding")).lower() == "note" else "finding"
    evidence = str(f.get("evidence", "")).upper()
    evidence = evidence if evidence in ("OBSERVED", "INFERRED") else "INFERRED"
    sev = str(f.get("sev", "MEDIUM")).upper()
    sev = "LOW" if kind == "note" else (sev if sev in SEVERITIES else "MEDIUM")
    line = f.get("line") if isinstance(f.get("line"), int) and not isinstance(f.get("line"), bool) else 0
    path = str(f.get("file") or "")[:300]
    if sev == "CRITICAL" and not (evidence == "OBSERVED" and path and line > 0):
        sev = "HIGH"  # a CRITICAL must be OBSERVED with file:line
    detail, hit1 = clean_field(f.get("detail"), DETAIL_MAX)
    fix, hit2 = clean_field(f.get("fix"), FIX_MAX)
    scenario = one_line(f.get("scenario"), 400)
    hit3 = looks_secret(scenario)
    entry = {"id": f["id"].strip()[:40], "kind": kind, "evidence": evidence, "sev": sev, "file": path, "line": line,
             "scenario": REDACTED if hit3 else scenario, "status": f["status"], "detail": detail, "fix": fix, "source": "deep"}
    if hit1 or hit2 or hit3:
        entry["redacted"] = True
    if FULL_SHA.match(head or ""):
        entry["first_sha"] = head
        entry["location"], probable = locate(head, path, line, "%s %s" % (scenario, detail))
        if probable:
            entry["probable"] = probable
    return entry


def deep_output(head=""):
    """(data, None) or (None, reason) from DEEP_JSON."""
    raw = os.environ.get("DEEP_JSON") or ""
    if not raw.strip():
        return None, "the deep review returned no structured output"
    try:
        data = json.loads(raw)
    except ValueError:
        return None, "the deep review output is not valid JSON"
    if not isinstance(data, dict) or str(data.get("level", "")).upper() not in LEVELS or not isinstance(data.get("findings"), list):
        return None, "the deep review output has no level or findings"
    filled = []
    for f in data["findings"]:
        if deep_entry(f, "") is not None and (is_blank(f.get("detail")) or is_blank(f.get("fix"))):
            filled.append(slug_id(f.get("id")))
            for key in ("detail", "fix"):
                if is_blank(f.get(key)):
                    f[key] = MISSING_FIELD
    raw_entries = [deep_entry(f, head) for f in data["findings"]]
    entries, why, claimed = [e for e in raw_entries if e], None, []
    if not all(raw_entries):
        # An invalid entry is dropped and named; it never sets the level. When it claimed HIGH or CRITICAL the line says so
        # explicitly, so a human looks at it.
        for f, e in zip(data["findings"], raw_entries):
            if e is None and isinstance(f, dict) and str(f.get("sev", "")).upper() in ("HIGH", "CRITICAL"):
                claimed.append(str(f.get("sev")).upper())
        why = "the deep review output had an invalid entry (dropped)" + (
            "; it claimed %s: a human must read the deep review" % "/".join(sorted(set(claimed))) if claimed else "")
    if len(entries) > DEEP_MAX_ENTRIES:
        # Keep the most severe (findings before notes), list them, and say NOT RUN: the review did not respect its cap.
        entries = sorted(entries, key=lambda e: (e["kind"] != "finding", -SEV_RANK.get(e["sev"], 0)))[:DEEP_MAX_ENTRIES]
        why = "too many findings"
    level = effective_level(str(data["level"]).upper(), entries)
    return {"level": level, "blast": one_line(data.get("blast_radius")) or "unknown",
            "summary": str(data.get("summary") or ""), "entries": entries, "filled": filled}, why


def plural(n, word):
    return "%d %s%s" % (n, word, "" if n == 1 else "s")


def strip_body(body):
    """The free text of a review comment: without the first line, marker, Tool and Deep lines, deep section and ledger."""
    body = ANY_LEDGER.sub("", DEEP_SECTION.sub("", body or ""))
    lines = body.strip().splitlines()[1:]
    lines = [l for l in lines if MARKER not in l and not l.startswith(("Tool:", "Deep:", "Reviewed:"))]
    return "\n".join(lines).strip()


def clean_summary(text):
    text = ANY_LEDGER.sub("", DEEP_SECTION.sub("", text or ""))
    keep = [l for l in text.strip().splitlines()
            if not l.startswith(("Triage:", "Tool:", "Deep:")) and MARKER not in l and DEEP_START not in l and DEEP_END not in l]
    return "\n".join(keep).strip()


def merge_deep(base, data, why, head):
    """The merged comment body from the current review comment body and the deep output (or its NOT RUN reason)."""
    lines = (base or "").strip().splitlines()
    first = lines[0].strip() if lines else ""
    tool = lines[2].strip() if len(lines) > 2 and TOOL.match(lines[2].strip()) else tool_or_default("")
    reviewed = [lines[3].strip()] if len(lines) > 3 and lines[3].startswith("Reviewed:") else []
    ledger = ledger_of(base) or {"sha": "", "findings": []}
    if data is not None:
        # A deep entry never overwrites a stage 1 entry of the same id: the stage 1 entry keeps its source (review), gains
        # a `deep` sub-field and takes the worse severity. The deep review can never mark a stage 1 finding fixed,
        # answered or accepted: an entry with a better status than the stage 1 entry is ignored.
        stage1 = {f.get("id"): f for f in ledger["findings"] if isinstance(f, dict) and f.get("source") != "deep"}
        before = {f.get("id"): f for f in ledger["findings"] if isinstance(f, dict) and f.get("source") == "deep"}
        standalone = []
        for e in data["entries"]:
            s1 = stage1.get(e["id"])
            if s1 is None:
                old = before.get(e["id"]) or {}
                if FULL_SHA.match(str(old.get("first_sha") or "")) and old.get("file") == e["file"] and old.get("line") == e["line"]:
                    e["first_sha"], e["location"] = old["first_sha"], old.get("location", e.get("location", "unverified"))
                standalone.append(e)
                continue
            if STATUS_RANK[e["status"]] > STATUS_RANK.get(s1.get("status"), 0):
                continue
            sub = {k: e.get(k) for k in ("detail", "fix", "sev", "file", "line", "location", "probable") if e.get(k) is not None}
            if e.get("redacted"):
                sub["redacted"] = True
            s1["source"] = s1.get("source") if s1.get("source") in ("review", "deep") else "review"
            s1["deep"] = sub
            s1["kind"] = "finding" if "finding" in (s1.get("kind", "finding"), e["kind"]) else "note"
            worst = max(str(s1.get("sev", "LOW")).upper(), e["sev"], key=lambda v: SEV_RANK.get(v, 0))
            s1["sev"] = "LOW" if s1["kind"] == "note" else worst
            if s1["sev"] == "CRITICAL" and e["sev"] == "CRITICAL":
                s1["evidence"] = "OBSERVED"
            s1["status"] = e["status"]
        ledger["findings"] = [f for f in ledger["findings"] if not (isinstance(f, dict) and f.get("source") == "deep")] + standalone
        ledger["deep"] = {"status": "NOT RUN" if why else "PASS", "sha": head}
        found = sum(1 for e in data["entries"] if e["kind"] == "finding")
        if why:
            deep_line = "Deep: NOT RUN (%s)" % one_line(why)
        else:
            deep_line = "Deep: PASS (%s, %s)" % (plural(found, "finding"), plural(len(data["entries"]) - found, "note"))
        m = VERDICT.match(first)
        if m and LEVELS.index(data["level"]) > LEVELS.index(m.group(1)):
            first = "Triage: %s | blast radius: %s" % (data["level"], data["blast"])
        summary = clean_summary(data["summary"]) or "(no text)"
        if data.get("filled"):
            summary = missing_notice(data["filled"]) + "\n\n" + summary
    else:
        ledger["deep"] = {"status": "NOT RUN", "sha": head}
        deep_line = "Deep: NOT RUN (%s)" % one_line(why)
        summary = "The deep review did not run (%s): this is not a pass." % one_line(why)
    rest = strip_body(base)
    out = [first, MARKER, tool] + reviewed + [deep_line, ""] + ([rest, ""] if rest else [])
    out += [DEEP_START, "### Deep review (P0 paths, adversarial)", "", summary, DEEP_END, "", ledger_section(ledger)]
    return "\n".join(out) + "\n"


def deep_main():
    env = target_env()
    if env is None:
        print("review-comment deep: NOT RUN (REPO and PR_NUMBER are required)")
        return 0
    repo, pr, binary = env
    head = os.environ.get("PR_HEAD_SHA", "")
    head = head if FULL_SHA.match(head) else ""
    data, why = deep_output(head)
    try:
        comments = fetch_comments(binary, repo, pr)
    except (GitHubError, ValueError) as exc:
        print("review-comment deep: NOT RUN (%s)" % exc)
        return 0
    # The comment of THIS run: by the id the review job wrote (REVIEW_COMMENT_ID), else the newest marker comment whose
    # ledger sha is this head. A newer review comment (a newer push) means this run is stale: edit nothing.
    marked = [c for c in comments if is_bot(c) and MARKER in (c.get("body") or "")]
    cid = os.environ.get("REVIEW_COMMENT_ID", "")
    if re.match(r"^\d+$", cid):
        own = [c for c in marked if str(c.get("id")) == cid]
    else:
        own = [c for c in marked if head and (ledger_of(c.get("body")) or {}).get("sha") == head]
    if not own:
        print("review-comment deep: NOT RUN (no review comment of this run)")
        return 0
    target = max(own, key=created_key)
    newest = review_comment(comments)
    if newest is not None and newest.get("id") != target.get("id"):
        print("review-comment deep: NOT RUN (stale: a newer review comment %s exists; comment %s is left as it is)" % (newest.get("id"), target.get("id")))
        return 0
    body = merge_deep(target.get("body") or "", data, why, head)
    print("review-comment deep: %s" % next((l for l in body.splitlines()[:6] if l.startswith("Deep:")), "Deep: ?"))
    try:
        edit(binary, repo, target, body)
    except (GitHubError, ValueError, KeyError, TypeError) as exc:
        print("review-comment deep: NOT RUN (%s)" % exc)
    return 0


def p0_patterns(root):
    """The P0 path patterns: the sensitive paths plus the deep-review paths under root (missing files are empty)."""
    out = []
    for rel in P0_LISTS:
        try:
            with open(os.path.join(root, rel), encoding="utf-8") as fh:
                out += [l.strip() for l in fh if l.strip() and not l.lstrip().startswith("#")]
        except OSError:
            continue
    return out


def changed_paths(items):
    """Every path a pull request touches, from the pulls/N/files answer: the new name and, for a rename, the old one
    (a rename out of a P0, sensitive or pipeline path touches it)."""
    out = []
    for f in items:
        for key in ("filename", "previous_filename"):
            if isinstance(f, dict) and f.get(key):
                out.append(f[key])
    return out


def p0_hits(files, patterns):
    return sorted(f for f in files if any(fnmatch.fnmatchcase(f, p) for p in patterns))


def deep_needed_main():
    env = target_env()
    root = os.environ.get("ROOT") or os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    if env is None:
        print("false")
        return 0
    repo, pr, binary = env
    try:
        files = changed_paths(parse_pages(gh(binary, ["api", "repos/%s/pulls/%s/files" % (repo, pr), "--paginate"])))
    except (GitHubError, ValueError) as exc:
        print("deep-needed: NOT RUN (%s): no deep review" % exc, file=sys.stderr)
        print("false")
        return 0
    hits = p0_hits(files, p0_patterns(root))
    print("deep-needed: %s" % (", ".join(hits[:6]) if hits else "no P0 path touched"), file=sys.stderr)
    print("true" if hits else "false")
    return 0


if __name__ == "__main__":
    if len(sys.argv) > 1 and sys.argv[1] == "deep":
        sys.exit(deep_main())
    if len(sys.argv) > 1 and sys.argv[1] == "deep-needed":
        sys.exit(deep_needed_main())
    sys.exit(main())
