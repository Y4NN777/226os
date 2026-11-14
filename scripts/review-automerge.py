#!/usr/bin/env python3
"""Auto-merge (stdlib only), stage 4 of the AI review pipeline: merge a pull request into develop when CI and the
review agree that nothing is left for a human, then bring the other open pull requests up to date.

It decides MERGE only if ALL of these hold for the pull request (any other case is a REFUSE with its reasons):
  - it is open, not a draft, its base is develop (never main) and its head repository is this repository;
  - it has none of the labels needs-owner, do-not-merge, owner-override-pending;
  - it changes nothing under .github/workflows/ (the bot never merges a workflow change: the owner does);
  - it touches no review pipeline path (.github/pipeline-paths.txt, read at the current tip of develop). With AUTO_MERGE=on
    such a pull request is labelled `needs-owner` with a one-line comment giving the reason (once: the label then
    refuses it before anything else), and the owner merges it;
  - every required check (REQUIRED_CHECKS) has a latest check run on the head that is completed with SUCCESS;
  - the review comment (marker `<!-- 226os-ai-review -->`, scripts/review-comment.py) exists, its Triage is neither
    NOT RUN nor CRITICAL, its ledger sha is the head (the review covered the latest commit) and no `finding` is open
    (notes never block);
  - when a P0 path is touched (.github/sensitive-paths.txt plus .github/deep-review-paths.txt, both read at the
    current tip of develop through the API; renames count with their old path), the ledger records a deep review PASS of the head; and
    line 4 of the comment is never `Deep: PENDING` or `Deep: NOT RUN` (the review job's own signal, whatever the lists);
  - when a sensitive path is touched, the owner's latest review APPROVED exactly the head commit, or the owner opened
    the pull request and authored every commit on it (stricter than the Review gate, applied whatever the gate mode);
  - a HIGH or MEDIUM verdict always has an open finding that explains it (otherwise it is refused, never merged);
  - every path list (sensitive, deep-review, pipeline) is read at the current tip of develop, and a renamed file counts
    with its old path too.
MERGE is a merge commit of exactly the evaluated head (`gh pr merge --merge --match-head-commit`), never squash or
rebase; GitHub's own branch protection still applies (a 405 is reported, exit 0). Then the update-branch API is called
for every other open pull request into develop from this repository (failures are printed, never fatal).

Reads use GH_TOKEN (the workflow's read-only token); only the merge and update-branch use BOT_TOKEN (the secret
REVIEW_AGENT_TOKEN of the GitHub environment review-agent: a merge with GITHUB_TOKEN would trigger no workflow, so the notes harvest would never run).
Mode (AUTO_MERGE): off (nothing), report (default: prints WOULD MERGE / WOULD UPDATE, writes nothing), on.
A missing BOT_TOKEN prints `auto-merge: NOT RUN (REVIEW_AGENT_TOKEN missing)`. Every error is NOT RUN with exit 0.

Environment: REPO, HEAD_SHA (the head commit a workflow_run completed on) or PR_NUMBER (a manual run), AUTO_MERGE,
BOT_TOKEN, GH_TOKEN, BASE (default develop), OWNER (default Y4NN777), REQUIRED_CHECKS (comma separated, default
"Validate branch promotion,quality,Review gate"), GH_BIN (default gh).
"""
import base64
import importlib.util
import os
import re
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
BLOCKING_LABELS = ("needs-owner", "do-not-merge", "owner-override-pending")
DEFAULT_CHECKS = "Validate branch promotion,quality,Review gate"
PIPELINE_LIST = ".github/pipeline-paths.txt"
SENSITIVE_LIST = ".github/sensitive-paths.txt"
OWNER_LABEL = "needs-owner"
FULL_SHA = re.compile(r"^[0-9a-f]{40}$")


def load(name, filename):
    spec = importlib.util.spec_from_file_location(name, os.path.join(HERE, filename))
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


comment = load("review_comment", "review-comment.py")
gate = load("review_gate", "review-gate.py")


def latest_runs(items):
    """{check name: its latest check run} from the paginated check-runs answer."""
    runs = []
    for page in items:
        runs += page.get("check_runs", []) if isinstance(page, dict) else []
    latest = {}
    for r in runs:
        name = r.get("name", "")
        if name not in latest or (r.get("id") or 0) >= (latest[name].get("id") or 0):
            latest[name] = r
    return latest


def base_list(gh, repo, path, base):
    """The patterns of a path list at the base commit ref (empty when it is missing)."""
    try:
        items = gate.gh_api(gh, "repos/%s/contents/%s?ref=%s" % (repo, path, base))
        raw = base64.b64decode(items[0]["content"]).decode("utf-8")
    except (gate.GitHubError, KeyError, IndexError, ValueError):
        return []
    return [l.strip() for l in raw.splitlines() if l.strip() and not l.lstrip().startswith("#")]


def label_names(pull):
    return [x if isinstance(x, str) else (x or {}).get("name", "") for x in pull.get("labels") or []]


def owner_authored(gh, repo, number, pull, owner):
    """True only when the owner opened the pull request AND authored every commit on it."""
    if (pull.get("user") or {}).get("login") != owner:
        return False
    commits = gate.gh_api(gh, "repos/%s/pulls/%s/commits" % (repo, number))
    return bool(commits) and all(((c.get("author") or {}).get("login")) == owner for c in commits)


def owner_approved_head(gh, repo, number, owner, head):
    """True when the owner's latest decisive review is APPROVED on exactly the head commit."""
    last = None
    for r in gate.gh_api(gh, "repos/%s/pulls/%s/reviews" % (repo, number)):
        if (r.get("user") or {}).get("login") == owner and r.get("state") in ("APPROVED", "CHANGES_REQUESTED", "DISMISSED"):
            if last is None or r.get("submitted_at", "") >= last.get("submitted_at", ""):
                last = r
    return last is not None and last.get("state") == "APPROVED" and last.get("commit_id") == head


def sensitive_refusal(gh, repo, number, pull, owner, files, patterns, head):
    """The refusal for a sensitive path, or None: the owner authored every commit, or approved exactly the head."""
    hits = comment.p0_hits(files, patterns)
    if not hits:
        return None
    if owner_authored(gh, repo, number, pull, owner) or owner_approved_head(gh, repo, number, owner, head):
        return None
    return "sensitive paths touched (%s) without the owner's approval of the head" % ", ".join(hits[:3])


def refusals(gh, repo, number, base, owner, required):
    """(head sha, [reasons], pipeline reason or None) for one pull request; no reason means MERGE. Raises
    gate.GitHubError on a read error."""
    pull = (gate.gh_api(gh, "repos/%s/pulls/%s" % (repo, number)) or [{}])[0]
    head = (pull.get("head") or {}).get("sha", "")
    why = []
    if pull.get("state") != "open":
        why.append("the pull request is not open")
    if (pull.get("base") or {}).get("ref") != base:
        why.append("its base is %s, not %s" % ((pull.get("base") or {}).get("ref"), base))
    if ((pull.get("head") or {}).get("repo") or {}).get("full_name") != repo:
        why.append("its head repository is not %s" % repo)
    if pull.get("draft"):
        why.append("it is a draft")
    for label in BLOCKING_LABELS:
        if label in label_names(pull):
            why.append("label %s" % label)
    if why or not FULL_SHA.match(head):
        return head, why or ["no head commit"], None

    # Every path list is read at the CURRENT tip of the base branch (what the merge lands on): a list that grew since the
    # review's event demands its deep review or owner approval here, or the pull request is refused with the reason.
    base_ref = base
    files = comment.changed_paths(gate.gh_api(gh, "repos/%s/pulls/%s/files" % (repo, number)))
    if any(f.startswith(".github/workflows/") for f in files):
        why.append("it changes a workflow under .github/workflows (the owner merges it)")
    pipeline = comment.p0_hits(files, base_list(gh, repo, PIPELINE_LIST, base_ref))
    pipeline_why = None
    if pipeline:
        pipeline_why = "it touches the review pipeline (%s): the owner merges it" % ", ".join(pipeline[:3])
        why.append(pipeline_why)

    latest = latest_runs(gate.gh_api(gh, "repos/%s/commits/%s/check-runs?per_page=100" % (repo, head)))
    for name in required:
        run = latest.get(name)
        if run is None:
            why.append("required check %s has not reported on the head" % name)
        elif run.get("status") != "completed" or run.get("conclusion") != "success":
            why.append("required check %s is %s" % (name, run.get("conclusion") or run.get("status")))

    target = comment.review_comment(gate.gh_api(gh, "repos/%s/issues/%s/comments" % (repo, number)))
    body = (target or {}).get("body") or ""
    first = comment.first_line(body)
    ledger = comment.ledger_of(body) or {}
    if target is None:
        why.append("no review comment")
    elif comment.NOT_RUN.match(first):
        why.append("the review is NOT RUN (%s)" % first)
    elif not comment.VERDICT.match(first):
        why.append("the review comment has no Triage verdict")
    elif comment.VERDICT.match(first).group(1) == "CRITICAL":
        why.append("the review verdict is CRITICAL")
    level = comment.VERDICT.match(first).group(1) if target is not None and comment.VERDICT.match(first) else None
    if target is not None and ledger.get("sha") != head:
        why.append("the ledger sha %s is not the head %s (the review did not cover the latest commit)" % (str(ledger.get("sha"))[:12] or "(none)", head[:12]))
    open_ids = [str(f.get("id")) for f in ledger.get("findings", []) if isinstance(f, dict)
                and str(f.get("kind", "finding")).lower() != "note" and f.get("status") == "open"]
    if open_ids:
        why.append("open finding(s): %s" % ", ".join(open_ids[:6]))
    elif level in ("HIGH", "MEDIUM"):
        why.append("the verdict is %s but no open finding explains it" % level)

    lines = body.strip().splitlines()
    deep_line = lines[3].strip() if len(lines) > 3 else ""
    if deep_line.startswith(("Deep: PENDING", "Deep: NOT RUN")):
        why.append("the deep review is not done (%s)" % deep_line)
    patterns = [p for rel in comment.P0_LISTS for p in base_list(gh, repo, rel, base_ref)]
    hits = comment.p0_hits(files, patterns)
    deep = ledger.get("deep") if isinstance(ledger.get("deep"), dict) else {}
    if hits and not (deep.get("status") == "PASS" and deep.get("sha") == head):
        why.append("a P0 path is touched (%s) without a deep review PASS of the head" % ", ".join(hits[:3]))

    failure = sensitive_refusal(gh, repo, number, pull, owner, files, base_list(gh, repo, SENSITIVE_LIST, base_ref), head)
    if failure:
        why.append(failure)
    return head, why, pipeline_why


def write(gh, token, args):
    """A write with the bot token only. (ok, short error)."""
    env = dict(os.environ, GH_TOKEN=token)
    env.pop("BOT_TOKEN", None)
    try:
        done = subprocess.run([gh] + args, capture_output=True, text=True, timeout=120, check=False, env=env)
    except (OSError, subprocess.SubprocessError) as exc:
        return False, type(exc).__name__
    return done.returncode == 0, " ".join((done.stderr or "").split())[:200]


def update_others(gh, repo, base, merged, token, mode):
    try:
        pulls = gate.gh_api(gh, "repos/%s/pulls?state=open&base=%s&per_page=100" % (repo, base))
    except gate.GitHubError as exc:
        print("auto-merge: update-branch NOT RUN (%s)" % exc)
        return
    for p in pulls:
        n = p.get("number")
        if n == merged or ((p.get("head") or {}).get("repo") or {}).get("full_name") != repo:
            continue
        if mode != "on":
            print("auto-merge: WOULD UPDATE #%s" % n)
            continue
        ok, err = write(gh, token, ["api", "-X", "PUT", "repos/%s/pulls/%s/update-branch" % (repo, n)])
        print("auto-merge: updated #%s" % n if ok else "auto-merge: could not update #%s (%s)" % (n, err))


def main():
    mode = (os.environ.get("AUTO_MERGE") or "report").strip().lower()
    if mode not in ("off", "report", "on"):
        print("auto-merge: unknown AUTO_MERGE %r, using report" % mode)
        mode = "report"
    if mode == "off":
        print("auto-merge: off (AUTO_MERGE=off): nothing checked")
        return 0
    token = os.environ.get("BOT_TOKEN", "")
    if not token:
        print("auto-merge: NOT RUN (REVIEW_AGENT_TOKEN missing)")
        return 0
    repo = os.environ.get("REPO", "")
    head_sha, number = os.environ.get("HEAD_SHA", ""), os.environ.get("PR_NUMBER", "")
    if not repo or (number and not re.match(r"^\d+$", number)) or (not number and not FULL_SHA.match(head_sha)):
        print("auto-merge: NOT RUN (REPO and a PR_NUMBER or a full HEAD_SHA are required)")
        return 0
    gh = os.environ.get("GH_BIN", "gh")
    base = os.environ.get("BASE", "develop")
    owner = os.environ.get("OWNER", "Y4NN777")
    required = [c.strip() for c in (os.environ.get("REQUIRED_CHECKS") or DEFAULT_CHECKS).split(",") if c.strip()]
    try:
        numbers = [number] if number else sorted({str(p.get("number")) for p in gate.gh_api(gh, "repos/%s/commits/%s/pulls" % (repo, head_sha)) if p.get("number")})
    except gate.GitHubError as exc:
        print("auto-merge: NOT RUN (%s)" % exc)
        return 0
    if not numbers:
        print("auto-merge: no pull request has the head %s" % head_sha[:12])
    for n in numbers:
        try:
            head, why, pipeline_why = refusals(gh, repo, n, base, owner, required)
        except (gate.GitHubError, ValueError, KeyError, TypeError, AttributeError) as exc:
            print("auto-merge: NOT RUN for #%s (%s): no merge" % (n, exc))
            continue
        if why:
            print("auto-merge: REFUSE #%s: %s" % (n, "; ".join(why)))
            if pipeline_why and mode == "on":
                ok, err = write(gh, token, ["api", "-X", "POST", "repos/%s/issues/%s/labels" % (repo, n), "-f", "labels[]=%s" % OWNER_LABEL])
                if ok:
                    write(gh, token, ["api", "-X", "POST", "repos/%s/issues/%s/comments" % (repo, n), "-f",
                                      "body=Auto-merge skipped: %s." % pipeline_why])
                print("auto-merge: labelled #%s %s" % (n, OWNER_LABEL) if ok else "auto-merge: could not label #%s (%s)" % (n, err))
            elif pipeline_why:
                print("auto-merge: WOULD LABEL #%s %s" % (n, OWNER_LABEL))
            continue
        if mode != "on":
            print("auto-merge: WOULD MERGE #%s at %s (merge commit; AUTO_MERGE=report)" % (n, head[:12]))
            update_others(gh, repo, base, int(n), token, mode)
            continue
        ok, err = write(gh, token, ["pr", "merge", n, "--repo", repo, "--merge", "--match-head-commit", head])
        if not ok:
            print("auto-merge: GitHub refused the merge of #%s (%s)" % (n, err))
            continue
        print("auto-merge: MERGED #%s at %s (merge commit)" % (n, head[:12]))
        update_others(gh, repo, base, int(n), token, mode)
    return 0


if __name__ == "__main__":
    sys.exit(main())
