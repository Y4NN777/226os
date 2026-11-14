#!/usr/bin/env python3
"""Write the ONE AI review comment of a pull request (stdlib only), deterministically.

The review returns its text; this script, run by the workflow from the BASE branch, posts it:
  - line 1 stays the `Triage:` verdict (the Review gate reads it), line 2 is the hidden marker MARKER, line 3 the fixed
    `Tool:` line of the preview step (a `Tool:` line written by the model is dropped);
  - the bot comment that carries the marker is edited by id; without one, the newest updated bot `Triage:` comment of an
    older run is adopted; otherwise a new comment is created. Another bot comment is never edited;
  - every other bot `Triage:` comment (a duplicate of an older run) is minimized as OUTDATED, never deleted;
  - the ledger `sha` is always PR_HEAD_SHA (validated), never what the model wrote. A missing, bare (no <details>),
    hidden-comment or broken ledger is dropped, never replaced by an empty one at the head: the next run is `first`.
No usable review text (empty, invalid, timed out, no `Triage:` first line): the marker comment becomes
`Triage: NOT RUN | <reason>` and carries the PREVIOUS ledger unchanged (or none), so nothing looks reviewed and the
commits not reviewed are reviewed on the next run. The Review gate reads no verdict from it, except after a CRITICAL:
then line 1 stays the CRITICAL verdict and its evidence is kept, so a push that exhausts the review never lifts a block.
Line 4 is `Deep: PENDING (...)` when DEEP_PENDING is true (a P0 path is touched; the deep job replaces it).

Subcommand `deep` (stage 2): merges the deep review's structured output (DEEP_JSON: level, blast_radius, summary,
findings) into the same comment and ledger. Its entries are tagged `"source":"deep"` (the deep entry wins on the same
id), line 4 becomes `Deep: PASS (...)` or `Deep: NOT RUN (<reason>)`, the Triage line becomes the worse of both levels
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


def review_comment(comments):
    """The review comment: the newest updated bot comment with the marker, else the newest updated bot comment whose
    first line is a `Triage:` verdict (a pull request reviewed before the marker existed), else None."""
    stamp = lambda c: c.get("updated_at") or c.get("created_at") or ""
    bots = [c for c in comments if is_bot(c)]
    marked = [c for c in bots if MARKER in (c.get("body") or "")]
    if marked:
        return max(marked, key=stamp)
    triage = [c for c in bots if VERDICT.match(first_line(c.get("body")))]
    return max(triage, key=stamp) if triage else None


def pin_ledger(text, head):
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
            # Only the deep merge writes `deep` and `source`: a model-written one would forge a deep review PASS.
            data.pop("deep", None)
            for f in data.get("findings") if isinstance(data.get("findings"), list) else []:
                if isinstance(f, dict):
                    f.pop("source", None)
            data["sha"] = head
            return text[:m.start(2)] + json.dumps(data) + text[m.end(2):]
    return text


def tool_or_default(tool_line):
    return tool_line if TOOL.match(tool_line or "") else "Tool: NOT RUN (the preview step gave no result)"


def compose(text, tool_line, head="", deep_line=None):
    """The comment body: verdict, marker, Tool line, then the review without any marker or Tool line of its own, its
    ledger pinned to head. None when the text does not start with a valid verdict (the gate could not read it)."""
    lines = (text or "").strip().splitlines()
    if not lines or not VERDICT.match(lines[0].strip()):
        return None
    rest = [l for l in lines[1:] if MARKER not in l and not l.startswith(("Tool:", "Deep:"))]
    while rest and not rest[0].strip():
        rest.pop(0)
    top = [lines[0].strip(), MARKER, tool_or_default(tool_line)] + ([deep_line] if deep_line else [])
    return pin_ledger("\n".join(top + [""] + rest), head) + "\n"


def not_run(reason, tool_line, previous):
    """The NOT RUN marker comment: no verdict the gate reads, and the previous review comment's ledger block unchanged
    (none when it had none), so the unreviewed commits stay unreviewed in the ledger."""
    note = ("The AI review produced no result on this run (%s): nothing in the new commits was reviewed. Push again or "
            "add the label `ai-review` to retry." % reason)
    old = ((previous or {}).get("body") or "").strip().splitlines()
    m = VERDICT.match(old[0].strip()) if old else None
    if m and m.group(1) == "CRITICAL":
        # A CRITICAL is never lifted by a review that did not run: keep the verdict and its evidence, add the reason.
        rest = [l for l in old[1:] if MARKER not in l and not l.startswith("Tool:")]
        while rest and not rest[0].strip():
            rest.pop(0)
        return "\n".join([old[0].strip(), MARKER, tool_or_default(tool_line), "", "NOT RUN on the latest push. " + note,
                          ""] + rest) + "\n"
    m = LEDGER_BLOCK.search((previous or {}).get("body") or "")
    body = ["Triage: NOT RUN | %s" % reason, MARKER, tool_or_default(tool_line), "", note]
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
    text = review_text()
    deep_line = DEEP_PENDING_LINE if os.environ.get("DEEP_PENDING") == "true" else None
    body = compose(text, tool_line, os.environ.get("PR_HEAD_SHA", ""), deep_line)
    try:
        comments = parse_pages(gh(binary, ["api", "repos/%s/issues/%s/comments" % (repo, pr), "--paginate"]))
    except (GitHubError, ValueError) as exc:
        print("review-comment: NOT RUN (%s)" % exc)
        return 0
    target = review_comment(comments)
    if body is None:
        outcome = re.sub(r"[^a-z_]", "", os.environ.get("REVIEW_OUTCOME", "").lower())[:20]
        if text is None or not text.strip():
            reason = "the review step ended with %s and returned no text" % (outcome or "an unknown outcome")
        else:
            reason = "the review text has no Triage first line"
        body = not_run(reason, tool_line, target)
        print("review-comment: the review gave no usable text (%s): writing a NOT RUN comment" % reason)
    try:
        if target is not None:
            gh(binary, ["api", "-X", "PATCH", "repos/%s/issues/comments/%s" % (repo, int(target["id"]))], {"body": body})
            print("review-comment: edited comment %s" % target["id"])
        else:
            created = json.loads(gh(binary, ["api", "-X", "POST", "repos/%s/issues/%s/comments" % (repo, pr)], {"body": body}) or "{}")
            target = created if isinstance(created, dict) else {}
            print("review-comment: created comment %s" % target.get("id", "?"))
    except (GitHubError, ValueError, KeyError, TypeError) as exc:
        print("review-comment: NOT RUN (%s)" % exc)
        return 0
    for c in comments:
        if c is target or c.get("id") == target.get("id") or not is_bot(c) or not c.get("node_id"):
            continue
        if not VERDICT.match(first_line(c.get("body"))):
            continue
        try:
            gh(binary, ["api", "graphql", "-f", "query=%s" % MINIMIZE, "-f", "id=%s" % c["node_id"]])
            print("review-comment: minimized the older review comment %s as outdated" % c.get("id"))
        except GitHubError as exc:
            print("review-comment: could not minimize comment %s (%s)" % (c.get("id"), exc))
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


def write(binary, repo, pr, comments, target, body):
    """Edit the target comment (else create one), then minimize the older bot Triage duplicates. Raises GitHubError."""
    if target is not None:
        gh(binary, ["api", "-X", "PATCH", "repos/%s/issues/comments/%s" % (repo, int(target["id"]))], {"body": body})
        print("review-comment: edited comment %s" % target["id"])
    else:
        try:
            created = json.loads(gh(binary, ["api", "-X", "POST", "repos/%s/issues/%s/comments" % (repo, pr)], {"body": body}) or "{}")
        except ValueError:
            created = {}
        target = created if isinstance(created, dict) else {}
        print("review-comment: created comment %s" % target.get("id", "?"))
    for c in comments:
        if c is target or c.get("id") == target.get("id") or not is_bot(c) or not c.get("node_id"):
            continue
        if not VERDICT.match(first_line(c.get("body"))):
            continue
        try:
            gh(binary, ["api", "graphql", "-f", "query=%s" % MINIMIZE, "-f", "id=%s" % c["node_id"]])
            print("review-comment: minimized the older review comment %s as outdated" % c.get("id"))
        except GitHubError as exc:
            print("review-comment: could not minimize comment %s (%s)" % (c.get("id"), exc))


def fetch_comments(binary, repo, pr):
    return parse_pages(gh(binary, ["api", "repos/%s/issues/%s/comments" % (repo, pr), "--paginate"]))


def target_env():
    repo, pr = os.environ.get("REPO", ""), os.environ.get("PR_NUMBER", "")
    if not repo or not re.match(r"^\d+$", pr):
        return None
    return repo, pr, os.environ.get("GH_BIN", "gh")


# --- stage 2: the deep review merges into the same comment and ledger ---

def deep_entry(f):
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
    return {"id": f["id"].strip()[:40], "kind": kind, "evidence": evidence, "sev": sev, "file": path, "line": line,
            "scenario": one_line(f.get("scenario"), 400), "status": f["status"], "source": "deep"}


def deep_output():
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
    entries = [deep_entry(f) for f in data["findings"]]
    if not all(entries):
        return None, "the deep review output has an invalid entry"
    why = None
    if len(entries) > DEEP_MAX_ENTRIES:
        # Keep the most severe (findings before notes), list them, and say NOT RUN: the review did not respect its cap.
        entries = sorted(entries, key=lambda e: (e["kind"] != "finding", -SEV_RANK.get(e["sev"], 0)))[:DEEP_MAX_ENTRIES]
        why = "too many findings"
    level = str(data["level"]).upper()
    if level == "CRITICAL" and not any(e["sev"] == "CRITICAL" and e["kind"] == "finding" for e in entries):
        level = "HIGH"
    return {"level": level, "blast": one_line(data.get("blast_radius")) or "unknown",
            "summary": str(data.get("summary") or ""), "entries": entries}, why


def plural(n, word):
    return "%d %s%s" % (n, word, "" if n == 1 else "s")


def strip_body(body):
    """The free text of a review comment: without the first line, marker, Tool and Deep lines, deep section and ledger."""
    body = ANY_LEDGER.sub("", DEEP_SECTION.sub("", body or ""))
    lines = body.strip().splitlines()[1:]
    lines = [l for l in lines if MARKER not in l and not l.startswith(("Tool:", "Deep:"))]
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
    ledger = ledger_of(base) or {"sha": "", "findings": []}
    if data is not None:
        # A deep entry may replace a stage 1 entry of the same id only when its status is not better: the deep review
        # can never mark a stage 1 finding fixed, answered or accepted.
        stage1 = {f.get("id"): f for f in ledger["findings"] if isinstance(f, dict) and f.get("source") != "deep"}
        merged = [e for e in data["entries"]
                  if e["id"] not in stage1 or STATUS_RANK[e["status"]] <= STATUS_RANK.get(stage1[e["id"]].get("status"), 0)]
        ids = {e["id"] for e in merged}
        kept = [f for f in ledger["findings"] if not (isinstance(f, dict) and (f.get("source") == "deep" or f.get("id") in ids))]
        ledger["findings"] = kept + merged
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
    else:
        ledger["deep"] = {"status": "NOT RUN", "sha": head}
        deep_line = "Deep: NOT RUN (%s)" % one_line(why)
        summary = "The deep review did not run (%s): this is not a pass." % one_line(why)
    rest = strip_body(base)
    out = [first, MARKER, tool, deep_line, ""] + ([rest, ""] if rest else [])
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
    data, why = deep_output()
    try:
        comments = fetch_comments(binary, repo, pr)
    except (GitHubError, ValueError) as exc:
        print("review-comment deep: NOT RUN (%s)" % exc)
        return 0
    target = review_comment(comments)
    base = (target or {}).get("body") or not_run("the standard review wrote no comment", "", None)
    body = merge_deep(base, data, why, head)
    print("review-comment deep: %s" % body.splitlines()[3])
    try:
        write(binary, repo, pr, comments, target, body)
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
