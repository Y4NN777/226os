#!/usr/bin/env python3
"""Close the issues a merged pull request names (stdlib only). GitHub honours `Closes #N` only on a merge into the
default branch (main); work merges into develop, so this does it there.

Reads the BODY of pull request PR_NUMBER (never its commits or comments). When it is merged into develop from this
repository, every `close|closes|closed|fix|fixes|fixed|resolve|resolves|resolved #N` (any case, an optional colon) names
an issue of this repository. For each one that is open: one comment "Fixed in #<pr> (merged into develop, <short sha>)."
then close it as completed. Left alone: pull requests, issues already closed, missing issues, another repository's
`owner/repo#N`, and issues labelled `review-batch` (`findings.py settle` closes a batch after its sources). An issue
labelled `needs-owner` gets a comment and stays open. Idempotent: the comment carries a marker and is written once; a
closed issue is never touched again. No mention.

Fails CLOSED: a GitHub error is "close-merged: FAILED" with exit 1 (the job is not a required check).
Environment: REPO (owner/name), PR_NUMBER, GH_BIN (default gh, authenticated through GH_TOKEN).
"""
import json
import os
import re
import subprocess
import sys

KEYWORD = re.compile(r"(?<![\w/])(?:close[sd]?|fix(?:e[sd])?|resolve[sd]?)\s*:?\s*#(\d+)(?!\d)", re.I)
MARK = "<!-- closed-by-merge:PR%s -->"
SKIP_LABEL = "review-batch"
OWNER_LABEL = "needs-owner"


class Missing(Exception):
    """The issue does not exist (404)."""


def gh(binary, *args):
    done = subprocess.run([binary, "api", *args], capture_output=True, text=True, timeout=60, check=False)
    if done.returncode != 0:
        if "404" in done.stderr:
            raise Missing()
        raise RuntimeError("gh api %s failed" % args[-1].split("?")[0])
    return done.stdout


def get(binary, endpoint):
    text = gh(binary, endpoint).strip()
    return json.loads(text) if text else None


def named_issues(body):
    seen = []
    for m in KEYWORD.finditer(body or ""):
        if m.group(1) not in seen:
            seen.append(m.group(1))
    return seen


def close_one(binary, repo, pr, sha, number):
    try:
        issue = get(binary, "repos/%s/issues/%s" % (repo, number))
    except Missing:
        print("close-merged: #%s does not exist, skipped" % number)
        return
    labels = {x if isinstance(x, str) else x.get("name", "") for x in issue.get("labels") or []}
    if "pull_request" in issue or issue.get("state") != "open" or SKIP_LABEL in labels:
        return
    mark = MARK % pr
    comments = get(binary, "repos/%s/issues/%s/comments?per_page=100" % (repo, number)) or []
    if not any(mark in (c.get("body") or "") for c in comments):
        if OWNER_LABEL in labels:
            text = "%s\nNamed as fixed by #%s (merged into develop, %s); left open because it is labelled needs-owner." % (mark, pr, sha)
        else:
            text = "%s\nFixed in #%s (merged into develop, %s)." % (mark, pr, sha)
        gh(binary, "-X", "POST", "repos/%s/issues/%s/comments" % (repo, number), "-f", "body=%s" % text)
    if OWNER_LABEL in labels:
        print("close-merged: #%s is labelled needs-owner: commented, left open" % number)
        return
    gh(binary, "-X", "PATCH", "repos/%s/issues/%s" % (repo, number), "-f", "state=closed", "-f", "state_reason=completed")
    print("close-merged: closed #%s" % number)


def main():
    repo, pr = os.environ.get("REPO", ""), os.environ.get("PR_NUMBER", "")
    binary = os.environ.get("GH_BIN", "gh")
    if not repo or not re.match(r"^\d+$", pr):
        print("close-merged: FAILED (REPO and PR_NUMBER are required)")
        return 2
    try:
        info = get(binary, "repos/%s/pulls/%s" % (repo, pr)) or {}
        head_repo = ((info.get("head") or {}).get("repo") or {}).get("full_name")
        if not info.get("merged_at") or (info.get("base") or {}).get("ref") != "develop" or head_repo != repo:
            print("close-merged: #%s is not a pull request of this repository merged into develop: nothing to close" % pr)
            return 0
        sha = (info.get("merge_commit_sha") or "")[:7] or "?"
        numbers = named_issues(info.get("body"))
        for number in numbers:
            close_one(binary, repo, pr, sha, number)
        print("close-merged: #%s names %d issue(s)" % (pr, len(numbers)))
        return 0
    except (RuntimeError, Missing, ValueError, OSError, subprocess.SubprocessError) as exc:
        print("close-merged: FAILED (%s); the next merge or a re-run closes what is left" % (str(exc) or type(exc).__name__))
        return 1


if __name__ == "__main__":
    sys.exit(main())
