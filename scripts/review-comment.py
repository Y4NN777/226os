#!/usr/bin/env python3
"""Write the ONE AI review comment of a pull request (stdlib only), deterministically.

The review returns its text; this script, run by the workflow from the BASE branch, posts it:
  - line 1 stays the `Triage:` verdict (the Review gate reads it), line 2 is the hidden marker MARKER, line 3 the fixed
    `Tool:` line of the preview step (a `Tool:` line written by the model is dropped);
  - the bot comment that carries the marker is edited by id; without one, the newest updated bot `Triage:` comment of an
    older run is adopted; otherwise a new comment is created. Another bot comment is never edited;
  - every other bot `Triage:` comment (a duplicate of an older run) is minimized as OUTDATED, never deleted.
Fails OPEN: a missing text or a GitHub error prints NOT RUN and exits 0, so it never turns the pull request red.

Environment: REPO (owner/name), PR_NUMBER, REVIEW_JSON (the review's structured output, {"body": "..."}) or
REVIEW_BODY_FILE (a file with the text, for a local run), TOOL_LINE (the preview step's `Tool:` line), GH_BIN (default gh).
"""
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


def compose(text, tool_line):
    """The comment body: verdict, marker, Tool line, then the review without any marker or Tool line of its own.
    None when the text does not start with a valid verdict (the gate could not read it)."""
    lines = (text or "").strip().splitlines()
    if not lines or not VERDICT.match(lines[0].strip()):
        return None
    if not TOOL.match(tool_line or ""):
        tool_line = "Tool: NOT RUN (the preview step gave no result)"
    rest = [l for l in lines[1:] if MARKER not in l and not l.startswith("Tool:")]
    while rest and not rest[0].strip():
        rest.pop(0)
    return "\n".join([lines[0].strip(), MARKER, tool_line, ""] + rest) + "\n"


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
    body = compose(review_text(), os.environ.get("TOOL_LINE", ""))
    if body is None:
        print("review-comment: NOT RUN (no review text starting with a Triage line)")
        return 0
    try:
        comments = parse_pages(gh(binary, ["api", "repos/%s/issues/%s/comments" % (repo, pr), "--paginate"]))
    except (GitHubError, ValueError) as exc:
        print("review-comment: NOT RUN (%s)" % exc)
        return 0
    target = review_comment(comments)
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


if __name__ == "__main__":
    sys.exit(main())
