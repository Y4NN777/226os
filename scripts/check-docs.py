#!/usr/bin/env python3
"""Documentation checks (stdlib only, fail closed). Keeps the documents honest about what the code says.

Checks, over the Markdown files under --root (default: the current directory):
  links    every relative link target in a .md file exists (anchors are not checked; http(s) and mailto are skipped)
  privacy  no absolute home path or agent scratch path (/home/<user>/, /Users/<user>/, /tmp/claude-) in a .md file
  ops      the operation count quoted in docs/quality/COVERAGE_SCOPE.md ("all N operations") equals the number of
           operationId entries in api/openapi.yaml
  adr      every docs/decisions/NNNN-*.md file is listed in docs/decisions/README.md

Exits 0 when every check passes, 1 on any failure, 2 on a usage error or when there is nothing to check (BLOCKED).
Excluded directories: .git, node_modules, reports, coverage, dist, .claude/worktrees, docs/archive (history is kept as is,
but its links are still checked: only privacy and ops ignore it).
"""
import argparse
import os
import re
import sys

SKIP_DIRS = {".git", "node_modules", "reports", "coverage", "dist"}
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


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--root", default=".")
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
    for error in errors:
        print("check-docs: FAIL: " + error, file=sys.stderr)
    if errors:
        return 1
    print("check-docs: PASS (%d Markdown files: links, privacy, operation count, ADR index)" % len(files))
    return 0


if __name__ == "__main__":
    sys.exit(main())
