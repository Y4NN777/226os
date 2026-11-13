#!/usr/bin/env python3
"""Review gate (stdlib only): the one required check that lets the AI review stop a CRITICAL change and lets the
owner review only the sensitive paths, without making everyone wait for a review.

It checks two things on a pull request:
  1. Triage verdict. The latest comment written by the Actions bot whose first line is exactly
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
Environment: REPO (owner/name), PR_NUMBER, PR_AUTHOR, OWNER (default Y4NN777), BASE_REF (default develop), GH_BIN (default gh,
authenticated through GH_TOKEN). Exit codes: 0 pass, 1 failure in enforce mode.
"""
import base64
import fnmatch
import json
import os
import re
import subprocess
import sys

BOT_LOGINS = {"github-actions[bot]", "github-actions"}
VERDICT = re.compile(r"^Triage:\s*(CRITICAL|HIGH|MEDIUM|NONE)\s*\|\s*blast radius:\s*\S.*$")
EVIDENCE = re.compile(r"[\w./-]+\.\w+:\d+")
OVERRIDE_LABEL = "owner-override"
LIST_PATH = ".github/sensitive-paths.txt"


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
    """(level, body) of the newest valid verdict by the bot, or None."""
    best = None
    for c in comments:
        if (c.get("user") or {}).get("login") not in BOT_LOGINS:
            continue
        body = c.get("body") or ""
        first = body.strip().splitlines()[0] if body.strip() else ""
        m = VERDICT.match(first)
        if m and (best is None or c.get("created_at", "") >= best[0]):
            best = (c.get("created_at", ""), m.group(1), body)
    return None if best is None else (best[1], best[2])


def owner_override(events, owner):
    last = None
    for e in events:
        if e.get("event") in ("labeled", "unlabeled") and (e.get("label") or {}).get("name") == OVERRIDE_LABEL:
            if last is None or e.get("created_at", "") >= last.get("created_at", ""):
                last = e
    return last is not None and last.get("event") == "labeled" and (last.get("actor") or {}).get("login") == owner


def check_verdict(gh, repo, pr, owner):
    """(failure message or None, note)."""
    comments = gh_api(gh, "repos/%s/issues/%s/comments" % (repo, pr))
    found = latest_verdict(comments)
    if found is None:
        return None, "triage verdict: NOT RUN (no review verdict on this pull request)"
    level, body = found
    if level != "CRITICAL":
        return None, "triage verdict %s: never blocks" % level
    if not ("OBSERVED" in body and EVIDENCE.search(body)):
        return None, "triage verdict CRITICAL downgraded: no OBSERVED evidence with file:line, treated as HIGH"
    if owner_override(gh_api(gh, "repos/%s/issues/%s/events" % (repo, pr)), owner):
        return None, "triage verdict CRITICAL cancelled by an owner override (label %s added by %s)" % (OVERRIDE_LABEL, owner)
    return "CRITICAL review verdict with OBSERVED evidence (add the label %s to override, owner only)" % OVERRIDE_LABEL, ""


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


def check_sensitive(gh, repo, pr, author, owner, base):
    patterns = sensitive_patterns(gh, repo, base)
    if patterns is None:
        return None, "sensitive paths: NOT RUN (no sensitive-paths list on %s)" % base
    files = [f.get("filename", "") for f in gh_api(gh, "repos/%s/pulls/%s/files" % (repo, pr))]
    hits = sorted(f for f in files if any(fnmatch.fnmatchcase(f, p) for p in patterns))
    if not hits:
        return None, "sensitive paths: no sensitive path touched"
    shown = ", ".join(hits[:6]) + (" ..." if len(hits) > 6 else "")
    if author == owner:
        return None, "sensitive paths touched (%s): owner-authored" % shown
    if owner_state(gh_api(gh, "repos/%s/pulls/%s/reviews" % (repo, pr)), owner) == "APPROVED":
        return None, "sensitive paths touched (%s): approved by the owner" % shown
    return "sensitive paths touched (%s) without the owner's approval" % shown, ""


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

    failures = []
    for name, check in (("verdict", lambda: check_verdict(gh, repo, pr, owner)),
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
