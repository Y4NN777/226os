#!/usr/bin/env python3
"""Documentation checks (stdlib only, fail closed). Keeps the documents honest about what the code says.

Checks, over the Markdown files under --root (default: the current directory):
  links    every relative link target in a .md file exists (anchors are not checked; http(s) and mailto are skipped)
  privacy  no absolute home path or agent scratch path (/home/<user>/, /Users/<user>/, /tmp/claude-) in a .md file
  ops      the operation count quoted in docs/quality/COVERAGE_SCOPE.md ("all N operations") equals the number of
           operationId entries in api/openapi.yaml
  adr      every docs/decisions/NNNN-*.md file is listed in docs/decisions/README.md

  dates    (report mode) an AI_USAGE.md record whose `date:` is after the HEAD commit date (UTC, one day of tolerance) is
           a WARN; `--strict-dates` makes it a failure. NOT RUN outside a git repository.

Exits 0 when every check passes, 1 on any failure, 2 on a usage error or when there is nothing to check (BLOCKED).
Excluded directories: .git, node_modules, reports, coverage, dist, .claude/worktrees, docs/archive (history is kept as is,
but its links are still checked: only privacy and ops ignore it).
"""
import argparse
import datetime
import os
import re
import subprocess
import sys

SKIP_DIRS = {".git", "node_modules", "reports", "coverage", "dist"}
RECORD_DATE = re.compile(r'^date:\s*"(\d{4}-\d{2}-\d{2})"')
RECORD_FILE = re.compile(r"^\d{4}-\d{2}-\d{2}-[a-z0-9-]+\.md$")
LINK = re.compile(r"\]\(([^)\s]+)\)")
FENCE = re.compile(r"```.*?```", re.S)
INLINE = re.compile(r"`[^`\n]*`")
PRIVATE = re.compile(r"(/home/[A-Za-z0-9._-]+/|/Users/[A-Za-z0-9._-]+/|/tmp/claude-)")


def markdown_files(root):
    found = []
    for dirpath, dirnames, filenames in os.walk(root):
        rel = os.path.relpath(dirpath, root)
        dirnames[:] = [d for d in dirnames if d not in SKIP_DIRS and os.path.join(rel, d) != os.path.join(".claude", "worktrees")]
        for name in filenames:
            if name.endswith(".md"):
                found.append(os.path.normpath(os.path.join(rel, name)))
    return sorted(found)


def read(root, path):
    with open(os.path.join(root, path), encoding="utf-8") as handle:
        return handle.read()


def check_links(root, files):
    errors = []
    for path in files:
        text = INLINE.sub("", FENCE.sub("", read(root, path)))
        for match in LINK.finditer(text):
            target = match.group(1)
            if re.match(r"(https?:|mailto:|#)", target):
                continue
            target = target.split("#", 1)[0]
            if not target:
                continue
            resolved = os.path.normpath(os.path.join(root, os.path.dirname(path), target))
            if not os.path.exists(resolved):
                errors.append("links: %s links to a missing file: %s" % (path, match.group(1)))
    return errors


def check_privacy(root, files):
    errors = []
    for path in files:
        if path.startswith("docs/archive/"):
            continue
        for number, line in enumerate(read(root, path).splitlines(), 1):
            if PRIVATE.search(line):
                errors.append("privacy: %s:%d contains an absolute home or agent scratch path" % (path, number))
    return errors


def check_ops(root):
    spec = os.path.join(root, "api", "openapi.yaml")
    scope = os.path.join(root, "docs", "quality", "COVERAGE_SCOPE.md")
    if not os.path.exists(spec) and not os.path.exists(scope):
        return []
    if not (os.path.exists(spec) and os.path.exists(scope)):
        return ["ops: api/openapi.yaml and docs/quality/COVERAGE_SCOPE.md must both exist"]
    with open(spec, encoding="utf-8") as handle:
        actual = len(re.findall(r"^\s*operationId:", handle.read(), re.M))
    quoted = re.findall(r"all (\d+) operations", read(root, "docs/quality/COVERAGE_SCOPE.md"))
    if len(quoted) != 1:
        return ["ops: docs/quality/COVERAGE_SCOPE.md must state the count once as 'all N operations' (found %d)" % len(quoted)]
    if int(quoted[0]) != actual:
        return ["ops: COVERAGE_SCOPE.md says %s operations, api/openapi.yaml has %d" % (quoted[0], actual)]
    return []


def check_adr(root):
    folder = os.path.join(root, "docs", "decisions")
    if not os.path.isdir(folder):
        return []
    index = os.path.join(folder, "README.md")
    if not os.path.exists(index):
        return ["adr: docs/decisions/README.md is missing"]
    with open(index, encoding="utf-8") as handle:
        listed = handle.read()
    return [
        "adr: docs/decisions/%s is not listed in docs/decisions/README.md" % name
        for name in sorted(os.listdir(folder))
        if re.match(r"\d{4}-.+\.md$", name) and name not in listed
    ]


def head_commit_date(root):
    """The date (UTC) of the last NON-merge commit reachable from HEAD, or the HEAD commit when there is none, or None
    outside a git repository. In CI the checked-out HEAD is a merge commit dated "now": the author's last real commit
    is the honest reference for the dates they wrote."""
    for extra in (["--no-merges"], []):
        try:
            done = subprocess.run(["git", "-C", root, "log", "-1", "--format=%cI"] + extra + ["HEAD"], capture_output=True,
                                  text=True, timeout=30, check=False)
        except (OSError, subprocess.SubprocessError):
            return None
        if done.returncode != 0:
            return None
        if done.stdout.strip():
            try:
                return datetime.datetime.fromisoformat(done.stdout.strip()).astimezone(datetime.timezone.utc).date()
            except ValueError:
                return None
    return None


def check_ai_usage_dates(root):
    """(warnings, note): an AI_USAGE.md record dated after the HEAD commit date (one day of tolerance for timezones and
    rebases) was written with a wrong date. Report mode by default; --strict-dates makes the warnings failures."""
    # AI_USAGE.md (policy, historical records and the generated block) and the one-record-per-file folder docs/ai-usage/.
    paths = [p for p in ["AI_USAGE.md"] if os.path.exists(os.path.join(root, p))]
    folder = os.path.join(root, "docs", "ai-usage")
    if os.path.isdir(folder):
        paths += ["docs/ai-usage/%s" % n for n in sorted(os.listdir(folder)) if RECORD_FILE.match(n)]
    if not paths:
        return [], ""
    head = head_commit_date(root)
    if head is None:
        return [], "check-docs: NOT RUN: ai-usage dates (no git history under %s)" % root
    limit = head + datetime.timedelta(days=1)
    warnings = []
    for path in paths:
        inside_generated = False
        for number, line in enumerate(read(root, path).splitlines(), 1):
            # The generated block of AI_USAGE.md repeats the record files, which are scanned on their own: skip it so a
            # bad date warns once (review note on #120).
            if path == "AI_USAGE.md":
                if line.startswith("<!-- ai-usage:generated:start"):
                    inside_generated = True
                elif line.startswith("<!-- ai-usage:generated:end"):
                    inside_generated = False
                    continue
                if inside_generated:
                    continue
            match = RECORD_DATE.match(line)
            if not match:
                continue
            try:
                day = datetime.date.fromisoformat(match.group(1))
            except ValueError:
                warnings.append("ai-usage: %s:%d date %s is not a valid date" % (path, number, match.group(1)))
                continue
            if day > limit:
                warnings.append("ai-usage: %s:%d date %s is after the HEAD commit date %s (one day of tolerance)"
                                % (path, number, day, head))
    return warnings, ""


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--root", default=".")
    parser.add_argument("--strict-dates", action="store_true",
                        help="fail (instead of warn) on an AI_USAGE.md record dated after the HEAD commit date")
    args = parser.parse_args()
    root = os.path.abspath(args.root)
    if not os.path.isdir(root):
        print("check-docs: BLOCKED: %s is not a directory" % root, file=sys.stderr)
        return 2
    files = markdown_files(root)
    if not files:
        print("check-docs: BLOCKED: no Markdown file found under %s" % root, file=sys.stderr)
        return 2
    errors = check_links(root, files) + check_privacy(root, files) + check_ops(root) + check_adr(root)
    date_warnings, date_note = check_ai_usage_dates(root)
    if args.strict_dates:
        errors += date_warnings
        date_warnings = []
    for error in errors:
        print("check-docs: FAIL: " + error, file=sys.stderr)
    for warning in date_warnings:
        print("check-docs: WARN: " + warning, file=sys.stderr)
    if date_note:
        print(date_note, file=sys.stderr)
    if errors:
        return 1
    print("check-docs: PASS (%d Markdown files: links, privacy, operation count, ADR index)" % len(files))
    return 0


if __name__ == "__main__":
    sys.exit(main())
