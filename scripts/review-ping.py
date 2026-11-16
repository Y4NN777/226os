#!/usr/bin/env python3
"""Owner ping (stdlib only): after the AI review, label the pull request `needs-owner` and ask the owner for a review
when a CRITICAL or HIGH finding is still open in the review's ledger (not marked `fixed`: an answered or accepted HIGH
or CRITICAL still pings, the owner settles it; an unknown status is open). It replaces an @-mention in the model's text: the
decision is read from the ledger, not guessed, and it works the same whatever the model wrote.

- Reads the newest bot comment that carries a valid ledger (the parser of scripts/review-context.py).
- A CRITICAL or HIGH finding not marked fixed: POST the label `needs-owner`; request the owner's review unless the owner wrote the pull
  request (GitHub refuses a review request to the author).
- Nothing serious open: remove the label only when the Actions bot added it (the last labeled event); one the owner or
  anyone else added stays (#212 L1).
- Uses `gh api` only: `gh pr edit` fails on the Projects classic GraphQL API.

Fail OPEN: any GitHub error is printed and the exit code stays 0, so an outage never fails a pull request.
Environment: REPO (owner/name), PR_NUMBER, PR_AUTHOR, OWNER (default Y4NN777), GH_BIN (default gh, GH_TOKEN auth).
"""
import importlib.util
import os
import subprocess
import sys

LABEL = "needs-owner"
SERIOUS = ("CRITICAL", "HIGH")
BOT_LOGINS = {"github-actions[bot]", "github-actions"}


def label_added_by_bot(ctx, binary, repo, pr):
    """True only when the LAST needs-owner label event on the pull request is a `labeled` one by the Actions bot (this
    script's own label). A label the owner or anyone else added is theirs to remove; an unreadable or missing event means
    keep the label."""
    try:
        events = ctx.gh_api(binary, "repos/%s/issues/%s/events" % (repo, pr))
    except Exception:  # fail safe: keep the label
        return False
    last = None
    for e in events or []:
        if isinstance(e, dict) and e.get("event") in ("labeled", "unlabeled") and (e.get("label") or {}).get("name") == LABEL:
            if last is None or (e.get("created_at") or "") >= (last.get("created_at") or ""):
                last = e
    return last is not None and last.get("event") == "labeled" and (last.get("actor") or {}).get("login") in BOT_LOGINS


def load_context_module():
    """scripts/review-context.py has a dash in its name: load it by path (same directory as this file)."""
    path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "review-context.py")
    spec = importlib.util.spec_from_file_location("review_context", path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def gh(binary, *args):
    done = subprocess.run([binary, *args], capture_output=True, text=True, timeout=60, check=False)
    return done.returncode == 0, done.stdout


def main():
    repo, pr = os.environ.get("REPO", ""), os.environ.get("PR_NUMBER", "")
    if not repo or not pr:
        print("review-ping: NOT RUN (REPO and PR_NUMBER are required)")
        return 0
    binary = os.environ.get("GH_BIN", "gh")
    owner = os.environ.get("OWNER", "Y4NN777")
    author = os.environ.get("PR_AUTHOR", "")
    try:
        ctx = load_context_module()
        comments = ctx.gh_api(binary, "repos/%s/issues/%s/comments" % (repo, pr))
        ledger = ctx.parse_ledger(comments)
    except Exception as exc:  # fail open: a broken read must never fail the pull request
        print("review-ping: NOT RUN (%s)" % type(exc).__name__)
        return 0
    # Only an explicit `fixed` clears a CRITICAL or HIGH (parse_finding normalises the status: unknown is open).
    open_serious = [f for f in (ledger or {}).get("findings", []) if f["kind"] == "finding" and f["status"] != "fixed" and f["sev"] in SERIOUS]
    try:
        if open_serious:
            gh(binary, "api", "-X", "POST", "repos/%s/issues/%s/labels" % (repo, pr), "-f", "labels[]=%s" % LABEL)
            if author and author != owner:
                gh(binary, "api", "-X", "POST", "repos/%s/pulls/%s/requested_reviewers" % (repo, pr), "-f", "reviewers[]=%s" % owner)
            print("review-ping: %d open CRITICAL/HIGH finding(s): labelled %s%s" % (
                len(open_serious), LABEL, "" if author == owner else ", review requested from the owner"))
        elif label_added_by_bot(ctx, binary, repo, pr):
            gh(binary, "api", "-X", "DELETE", "repos/%s/issues/%s/labels/%s" % (repo, pr, LABEL))
            print("review-ping: no open CRITICAL or HIGH finding: removed the label this script added")
        else:
            print("review-ping: no open CRITICAL or HIGH finding (a needs-owner label not added by the bot stays)")
    except (OSError, subprocess.SubprocessError) as exc:
        print("review-ping: NOT RUN (%s)" % type(exc).__name__)
    return 0


if __name__ == "__main__":
    sys.exit(main())
