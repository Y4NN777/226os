#!/usr/bin/env python3
"""AI-use records, one per file (stdlib only, fails closed).

Every pull request used to append its record at the end of AI_USAGE.md, so every merge conflicted with every other open
pull request. Now a record is ONE FILE, docs/ai-usage/YYYY-MM-DD-slug.md, and two pull requests never touch the same
lines. AI_USAGE.md keeps its policy and its historical records and ends with a generated block, between

  <!-- ai-usage:generated:start ... -->   and   <!-- ai-usage:generated:end -->

that `build` rewrites from the record files (sorted by file name, so by date). Nobody edits the block by hand.

  build   rewrite the generated block (the owner runs it before the submission and commits the result)
  check   exit 1 when AI_USAGE.md is not exactly what `build` would write (stale)

A record file must start with a `## ` heading and contain a ```yaml block with a `date: "YYYY-MM-DD"` and a
`contributor:` line; anything else is refused (exit 2) and the file is named. Files in docs/ai-usage/ that are not named
YYYY-MM-DD-slug.md (README, notes, drafts) are not records.
Usage: ai-usage.py build|check [--root DIR]. Exit codes: 0 ok, 1 stale (check), 2 usage error or an invalid record.
"""
import argparse
import os
import re
import sys

START = "<!-- ai-usage:generated:start (do not edit by hand: `make ai-usage` rewrites everything between the two markers from docs/ai-usage/*.md) -->"
START_PREFIX = "<!-- ai-usage:generated:start"
END = "<!-- ai-usage:generated:end -->"
NAME = re.compile(r"^\d{4}-\d{2}-\d{2}-[a-z0-9-]+\.md$")
DATE = re.compile(r'(?m)^date:\s*"\d{4}-\d{2}-\d{2}"\s*$')
CONTRIBUTOR = re.compile(r"(?m)^contributor:\s*\S")
YAML_BLOCK = re.compile(r"```yaml\s*\n(.*?)```", re.S)


class Invalid(Exception):
    """A record file that cannot go into AI_USAGE.md."""


def records(root):
    folder = os.path.join(root, "docs", "ai-usage")
    if not os.path.isdir(folder):
        return []
    out = []
    for name in sorted(os.listdir(folder)):
        if not NAME.match(name):
            continue
        with open(os.path.join(folder, name), encoding="utf-8") as handle:
            text = handle.read().strip("\n")
        if not text.startswith("## "):
            raise Invalid("docs/ai-usage/%s: a record must start with a `## ` heading" % name)
        block = YAML_BLOCK.search(text)
        if not block:
            raise Invalid("docs/ai-usage/%s: a record needs a ```yaml block" % name)
        if not DATE.search(block.group(1)):
            raise Invalid('docs/ai-usage/%s: the yaml block needs a `date: "YYYY-MM-DD"` line' % name)
        if not CONTRIBUTOR.search(block.group(1)):
            raise Invalid("docs/ai-usage/%s: the yaml block needs a `contributor:` line" % name)
        out.append(text)
    return out


def generated_block(recs):
    body = "\n\n".join(recs)
    return START + "\n\n" + (body + "\n\n" if body else "") + END + "\n"


def rebuilt(current, recs):
    """AI_USAGE.md with its generated block replaced (or appended when there is none yet)."""
    block = generated_block(recs)
    start = current.find(START_PREFIX)
    end = current.find(END)
    if start != -1 and end != -1 and end > start:
        return current[:start] + block + current[end + len(END):].lstrip("\n")
    return current.rstrip("\n") + "\n\n" + block


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("command", choices=["build", "check"])
    parser.add_argument("--root", default=".")
    parser.add_argument("--only-on-promotion", action="store_true",
                        help="check only when GITHUB_BASE_REF is main (a pull request into main); otherwise NOT RUN")
    args = parser.parse_args()
    if args.only_on_promotion and args.command == "check" and os.environ.get("GITHUB_BASE_REF") != "main":
        print("ai-usage: NOT RUN (the stale check applies to a pull request into main; the file is stale by design before)")
        return 0
    root = os.path.abspath(args.root)
    target = os.path.join(root, "AI_USAGE.md")
    if not os.path.isfile(target):
        print("ai-usage: BLOCKED: %s does not exist" % target, file=sys.stderr)
        return 2
    try:
        recs = records(root)
    except Invalid as exc:
        print("ai-usage: FAIL: %s" % exc, file=sys.stderr)
        return 2
    with open(target, encoding="utf-8") as handle:
        current = handle.read()
    wanted = rebuilt(current, recs)
    if args.command == "check":
        if wanted != current:
            print("ai-usage: FAIL: AI_USAGE.md is stale: run `make ai-usage` and commit it", file=sys.stderr)
            return 1
        print("ai-usage: PASS (%d record file(s), AI_USAGE.md up to date)" % len(recs))
        return 0
    if wanted != current:
        with open(target, "w", encoding="utf-8") as handle:
            handle.write(wanted)
    print("ai-usage: AI_USAGE.md rebuilt from %d record file(s)" % len(recs))
    return 0


if __name__ == "__main__":
    sys.exit(main())
