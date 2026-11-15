#!/usr/bin/env bash
# Tests for `scripts/findings.py settle <pr>`, notes closed after triage (owner, 2026-11-15): for an item of a NOTES
# issue, a NOT A DEFECT or DECIDED verdict that cites an ADR (docs/decisions/NNNN-*.md at the tip of develop, first Status line ACCEPTED, text naming the note's file as a
# whole path AND the note's own ledger id as a whole token; checked by the script, never by the model) ticks the box; a path:line citation must exist but never counts, and the issue closes as not planned once every box is settled. Without a verifiable reference it stays a comment
# (rule H5 of #187). A FINDING is never closed by those verdicts: a human does.
# Usage: scripts/tests/findings-settle-notes_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
tool="$root/scripts/findings.py"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin" "$tmp/st/tree/src" "$tmp/st/tree/docs/decisions" "$tmp/st/tree/.github" "$tmp/st/tree/src/api/internal/solve"
printf 'l1\nl2\nl3\nl4\nl5\n' >"$tmp/st/tree/src/a.go"
# ADRs: 0014 is ACCEPTED and names the files of the notes under test; 0015 is not ACCEPTED; 0016 is ACCEPTED but names none.
printf '# ADR\n\n- **status:** accepted (owner)\n\nCovers src/a.go and Makefile for the notes `n1` and `n2`.\n' >"$tmp/st/tree/docs/decisions/0014-review.md"
printf '# ADR\n\n- **Status:** PROPOSED\n\nCovers src/a.go.\n' >"$tmp/st/tree/docs/decisions/0015-proposed.md"
printf '# ADR\n\n- **Status:** ACCEPTED\n\nCovers nothing in particular.\n' >"$tmp/st/tree/docs/decisions/0016-unrelated.md"
printf '# ADR\n\n- **Status:** PROPOSED\n\n```\nStatus: accepted\n```\nCovers src/a.go.\n' >"$tmp/st/tree/docs/decisions/0017-late-status.md"
printf '# ADR\n\n- **Status:** ACCEPTED, superseded by ADR-0020\n\nCovers src/a.go.\n' >"$tmp/st/tree/docs/decisions/0018-superseded.md"
printf '# ADR\n\n- **Status:** ACCEPTED\n\nSee old/src/a.go, src/a.go.bak and https://x.example/y/src/a.go?q=1 only.\n' >"$tmp/st/tree/docs/decisions/0019-longer-paths.md"
printf '# ADR\n\n- **Status:** ACCEPTED (owner, 2026-11-15)\n\nThe rule for `src/a.go`, note `n1`. Also see src/a.go.\n' >"$tmp/st/tree/docs/decisions/0020-upper.md"
printf '# ADR\n\n- **Status:** ACCEPTED\n\nMentions src/a.go in passing, names no note.\n' >"$tmp/st/tree/docs/decisions/0021-file-no-id.md"
printf '# ADR\n\n- **Status:** ACCEPTED\n\nCovers src/a.go for notes n1-2 and n10 only, and a-n1.\n' >"$tmp/st/tree/docs/decisions/0022-longer-id.md"
printf '# ADR\n\n- **Status:** ACCEPTED\n\nCovers src/a.go; the note (`n1`), decided.\n' >"$tmp/st/tree/docs/decisions/0023-exact-id.md"
printf '# ADR\n\n- **Status:** ACCEPTED\n\nCovers src/a.go for the note N1 only.\n' >"$tmp/st/tree/docs/decisions/0024-other-case.md"
printf '# ADR\n\n- **Status:** ACCEPTED\n\nCovers src/a.go for the note `a.b` only.\n' >"$tmp/st/tree/docs/decisions/0025-dotted-id.md"
printf '# ADR\n\n- **Status:** ACCEPTED\n\nCovers src/a.go for the note a-b only.\n' >"$tmp/st/tree/docs/decisions/0026-slug-form.md"
printf 'r1\nr2\n' >"$tmp/st/tree/README.md"
printf 'm1\nm2\nm3\n' >"$tmp/st/tree/Makefile"
printf '# sensitive\n.github/*\n' >"$tmp/st/tree/.github/sensitive-paths.txt"
printf '# P0\nsrc/api/internal/solve/*\n*auth*\n' >"$tmp/st/tree/.github/deep-review-paths.txt"
printf 'v1\nv2\nv3\n' >"$tmp/st/tree/src/api/internal/solve/votes.go"
mkdir -p "$tmp/st/tree/.github/workflows"
printf 'w1\nw2\nw3\n' >"$tmp/st/tree/.github/workflows/a.yml"

# Fake gh: issues in $ST/issues.json, the pull request in $ST/pr.json, the tip of develop in $ST/tree.
cat >"$tmp/bin/gh" <<'PY'
#!/usr/bin/env python3
import base64, json, os, sys
st = os.environ["ST"]
args = sys.argv[1:]
method, fields, endpoint, i = "GET", {}, None, 1
while i < len(args):
    a = args[i]
    if a == "-X": method = args[i + 1]; i += 2
    elif a == "-f": k, v = args[i + 1].split("=", 1); fields[k] = v; i += 2
    elif a == "--paginate": i += 1
    else: endpoint = a; i += 1
path = endpoint.split("?")[0]
parts = path.split("/")
issues = json.load(open(os.path.join(st, "issues.json")))
def save(): json.dump(issues, open(os.path.join(st, "issues.json"), "w"))
def find(n): return next(x for x in issues if x["number"] == int(n))
if method != "GET":
    open(os.path.join(st, "log"), "a").write("%s %s %s\n" % (method, path, json.dumps(fields, sort_keys=True)))
if method == "GET" and "contents" in parts:
    assert endpoint.endswith("?ref=develop"), endpoint
    f = os.path.join(st, "tree", *parts[parts.index("contents") + 1:])
    if not os.path.isfile(f): sys.stderr.write("404\n"); sys.exit(1)
    print(json.dumps({"type": "file", "encoding": "base64", "content": base64.b64encode(open(f, "rb").read()).decode()}))
elif method == "GET" and parts[-1] == "files":
    print(open(os.path.join(st, "files.json")).read())
elif method == "GET" and parts[-2] == "pulls":
    print(open(os.path.join(st, "pr.json")).read())
elif method == "GET" and parts[-2] == "issues":
    x = dict(find(parts[-1])); x.pop("comments", None); print(json.dumps(x))
elif method == "GET" and parts[-1] == "comments":
    print(json.dumps(find(parts[-2])["comments"]))
elif method == "POST" and parts[-1] == "comments":
    find(parts[-2])["comments"].append({"user": {"login": "github-actions[bot]"}, "body": fields["body"]}); save(); print("{}")
elif method == "PATCH" and parts[-2] == "issues":
    x = find(parts[-1])
    for k in ("body", "state", "state_reason"):
        if k in fields: x[k] = fields[k]
    save(); print("{}")
else:
    sys.stderr.write("unexpected %s %s\n" % (method, endpoint)); sys.exit(2)
PY
chmod +x "$tmp/bin/gh"

setup() {
  python3 -I - "$tmp/st" <<'PY'
import json, os, sys
st = sys.argv[1]
def notes(n, pr, items):
    body = "<!-- finding:PR%d#notes -->\nNotes.\n\n" % pr + "\n".join(
        "- [ ] `%s` [OBSERVED] `src/a.go:1` remark %s" % (k, k) for k in items)
    return {"number": n, "state": "open", "state_reason": None, "title": "PR%d notes" % pr, "created_at": "2026-11-13T00:00:00Z",
            "labels": [{"name": "kind:note"}], "body": body, "comments": []}
def at(n, pr, where):
    x = notes(n, pr, ["n1"])
    x["body"] = x["body"].replace("`src/a.go:1`", where)
    return x
def p0(n, pr):
    x = notes(n, pr, ["n1"])
    x["body"] = x["body"].replace("`src/a.go:1`", "`src/api/internal/solve/votes.go:1`")
    return x
def mk(n, pr):
    x = notes(n, pr, ["n1"])
    x["body"] = x["body"].replace("`src/a.go:1`", "`Makefile:1`")
    return x
finding = {"number": 39, "state": "open", "state_reason": None, "title": "a finding", "created_at": "2026-11-13T00:00:00Z",
           "labels": [{"name": "finding"}], "body": "<!-- finding:PR90#f39 -->\n\n- Where: `src/a.go:1`\n", "comments": []}
issues = [notes(30, 91, ["n1", "n2"]), notes(31, 92, ["n1"]), notes(32, 93, ["n1"]), notes(33, 94, ["n1"]), notes(34, 95, ["n1", "n2"]),
          notes(35, 96, ["n1", "n2"]), notes(36, 97, ["n1", "n2"]), notes(37, 98, ["n1"]), notes(41, 99, ["n1"]), notes(42, 100, ["n1"]), notes(43, 101, ["n1", "n2"]), mk(44, 102), p0(45, 103), at(53, 111, "`src/a.go:1`"), at(54, 112, "`src/a.go:abc`"), at(55, 113, "`src/a.go:1`"), at(56, 114, "`src/a.go:1`"), at(57, 115, "`src/a.go:1`"), at(58, 116, "`src/a.go:1`"), at(59, 117, "`?:1`"), at(60, 118, "`src/a.go:1`"), at(61, 119, "`src/a.go:1`"), at(62, 120, "`src/a.go:1`"), at(63, 121, "`src/a.go:1`"), at(64, 122, "`src/a.go:1`"), at(65, 123, "`src/a.go:1`"), at(66, 124, "`src/a.go:1`"), at(67, 125, "`src/a.go:1`"), at(68, 126, "`src/a.go:1`"), notes(69, 127, ["a.b"]), notes(70, 128, ["a.b"]), at(46, 104, "`./src/api/internal/solve/votes.go:1`"), at(47, 105, "`pr/src/api/internal/solve/votes.go:1`"), at(48, 106, "`src/gone/x.go:1`"), at(49, 107, "`.github/workflows/a.yml:1`"), notes(50, 108, ["n1"]), notes(51, 109, ["n1"]), at(52, 110, "remark without a location"), finding]
rows = [("#30/n1", "#30", "NOT A DEFECT", "the guard sits at src/a.go:3 and docs/decisions/0014-review.md settles it, the review missed it"),
        ("#30/n2", "#30", "DECIDED", "docs/decisions/0014-review.md settles it"),
        ("#31/n1", "#31", "NOT A DEFECT", "see src/gone.go:1"),
        ("#32/n1", "#32", "DECIDED", "see src/a.go:99 (past the end)"),
        ("#33/n1", "#33", "NOT A DEFECT", "it is fine, trust me"),
        ("#34/n1", "#34", "NOT A DEFECT", "src/a.go:2-4 and docs/decisions/0014-review.md show it"),
        ("#35/n1", "#35", "NOT A DEFECT", "src/a.go:4-9 shows it"),
        ("#36/n1", "#36", "DECIDED", "docs/decisions/9999-none.md"),
        ("#36/n2", "#36", "NOT A DEFECT", "../../etc/passwd:1 and src/a.go:0"),
        ("#37/n1", "#37", "NOT A DEFECT", "README.md:1 says so"),
        ("#41/n1", "#41", "NOT A DEFECT", "src/a.go:1 is fine"),
        ("#42/n1", "#42", "DECIDED", "docs/decisions/0014-review.md, see src/a.go:99"),
        ("#43/n1", "#43", "NOT A DEFECT", "src/a.go:4 shows it"),
        ("#44/n1", "#44", "NOT A DEFECT", "Makefile:3 guards it (docs/decisions/0014-review.md); Retry:3 is a word, not a file"),
        ("#45/n1", "#45", "NOT A DEFECT", "src/api/internal/solve/votes.go:3 handles it"),
        ("#46/n1", "#46", "DECIDED", "docs/decisions/0014-review.md"),
        ("#47/n1", "#47", "DECIDED", "docs/decisions/0014-review.md"),
        ("#48/n1", "#48", "DECIDED", "docs/decisions/0014-review.md"),
        ("#49/n1", "#49", "NOT A DEFECT", ".github/workflows/a.yml:3 handles it"),
        ("#50/n1", "#50", "NOT A DEFECT", "src/a.go:1-2 is fine"),
        ("#51/n1", "#51", "NOT A DEFECT", "src/a.go:01 is fine"),
        ("#52/n1", "#52", "DECIDED", "docs/decisions/0014-review.md"),
        ("#43/n2", "#43", "NOT A DEFECT", "src/a.go:2-3 and docs/decisions/0014-review.md handle it"),
        ("#53/n1", "#53", "NOT A DEFECT", "src/a.go:3 proves it"),
        ("#54/n1", "#54", "NOT A DEFECT", "src/a.go:3 proves it"),
        ("#55/n1", "#55", "DECIDED", "docs/decisions/0015-proposed.md settles it"),
        ("#56/n1", "#56", "DECIDED", "docs/decisions/0016-unrelated.md settles it"),
        ("#57/n1", "#57", "DECIDED", "docs/decisions/0014-review.md settles it"),
        ("#58/n1", "#58", "DECIDED", "docs/decisions/0014-review.md, but src/a.go:99 is past the end"),
        ("#59/n1", "#59", "DECIDED", "docs/decisions/0014-review.md settles it"),
        ("#60/n1", "#60", "DECIDED", "docs/decisions/0014-review.md and docs/decisions/0016-unrelated.md"),
        ("#61/n1", "#61", "DECIDED", "docs/decisions/0017-late-status.md settles it"),
        ("#62/n1", "#62", "DECIDED", "docs/decisions/0018-superseded.md settles it"),
        ("#63/n1", "#63", "DECIDED", "docs/decisions/0019-longer-paths.md settles it"),
        ("#64/n1", "#64", "DECIDED", "docs/decisions/0020-upper.md settles it"),
        ("#65/n1", "#65", "DECIDED", "docs/decisions/0021-file-no-id.md settles it"),
        ("#66/n1", "#66", "DECIDED", "docs/decisions/0022-longer-id.md settles it"),
        ("#67/n1", "#67", "DECIDED", "docs/decisions/0023-exact-id.md settles it"),
        ("#68/n1", "#68", "DECIDED", "docs/decisions/0024-other-case.md settles it"),
        ("#69/a-b", "#69", "DECIDED", "docs/decisions/0025-dotted-id.md settles it"),
        ("#70/a-b", "#70", "DECIDED", "docs/decisions/0026-slug-form.md settles it"),
        ("#39", "#39", "NOT A DEFECT", "src/a.go:3 proves it")]
issues.append({"number": 40, "state": "open", "state_reason": None, "title": "Review batch 2026-11-15", "labels": [{"name": "review-batch"}],
               "body": "<!-- review-batch -->\n# Review batch 2026-11-15\n\n" + "\n".join(
                   "- [ ] %s x <!-- batch-item:%s -->" % (k, k) for k, *_ in rows), "comments": []})
x43 = [i for i in issues if i["number"] == 43][0]
x43["body"] = x43["body"].replace("- [ ] `n1`", "- [x] `n1`", 1)
json.dump(issues, open(os.path.join(st, "issues.json"), "w"))
json.dump([{"filename": "src/other.go"}], open(os.path.join(st, "files.json"), "w"))
body = "\n".join(["Batch fix.", "", "Batch: #40", "", "| Item | Source | Decision | Reason |", "|---|---|---|---|"]
                 + ["| %s | %s | %s | %s |" % r for r in rows])
json.dump({"number": 170, "merged_at": "2026-11-15T12:00:00Z", "base": {"ref": "develop"},
           "head": {"ref": "fix/review/batch-2026-11-15", "repo": {"full_name": "o/r"}}, "body": body}, open(os.path.join(st, "pr.json"), "w"))
PY
  : >"$tmp/st/log"
}
run() {
  rc=0
  out="$(PATH="$tmp/bin:$PATH" ST="$tmp/st" REPO=o/r python3 -I "$tool" settle "$@" 2>&1)" || rc=$?
}
field() { python3 -I -c 'import json,sys; x=[i for i in json.load(open(sys.argv[1])) if i["number"]==int(sys.argv[2])][0]; print(eval(sys.argv[3]))' "$tmp/st/issues.json" "$1" "$2"; }
state() { field "$1" '(x["state"], x["state_reason"])'; }
unticked() { field "$1" 'x["body"]' | grep -c -- '- \[ \]' || true; }

setup
run 170
[ "$rc" -eq 0 ] && ok "settle runs (rc=0)" || ko "settle rc=$rc ($out)"
[ "$(state 30)" = "('closed', 'not_planned')" ] && [ "$(unticked 30)" -eq 0 ] && ok "notes: NOT A DEFECT and DECIDED citing an ACCEPTED ADR that names the file (plus an existing own-file line) tick both boxes, the issue closes as not planned" || ko "30 ($(state 30) $(field 30 'x["body"]'))"
field 30 'x["comments"][-1]["body"]' | grep -q 'the guard sits at src/a.go:3 and' && ok "the verdict comment is posted" || ko "30 comment ($(field 30 'x["comments"]'))"
[ "$(state 31)" = "('open', None)" ] && [ "$(unticked 31)" -eq 1 ] && field 31 'x["comments"][-1]["body"]' | grep -q 'stays open: a human closes it as not planned' && ok "a citation to a missing file: box unticked, issue open, today's comment" || ko "31 ($(state 31) $(field 31 'x["comments"]'))"
[ "$(state 32)" = "('open', None)" ] && [ "$(unticked 32)" -eq 1 ] && ok "a line past the end of the file: stays open" || ko "32 ($(state 32))"
[ "$(state 33)" = "('open', None)" ] && [ "$(unticked 33)" -eq 1 ] && ok "no reference at all: stays open" || ko "33 ($(state 33))"
[ "$(state 34)" = "('open', None)" ] && [ "$(unticked 34)" -eq 1 ] && field 34 'x["body"]' | grep -q -- '- \[x\] `n1`' && ok "mixed issue: the verified note is ticked, the other is not settled, the issue stays open" || ko "34 ($(state 34) $(field 34 'x["body"]'))"
[ "$(state 35)" = "('open', None)" ] && [ "$(unticked 35)" -eq 2 ] && ok "a range that runs past the end of the file: stays open" || ko "35 ($(state 35))"
[ "$(state 36)" = "('open', None)" ] && [ "$(unticked 36)" -eq 2 ] && ok "a missing ADR, a path with .. and line 0: stay open" || ko "36 ($(state 36))"
[ "$(state 37)" = "('open', None)" ] && [ "$(unticked 37)" -eq 1 ] && ok "a line in ANOTHER existing file (not the note's own file, not an ADR): stays open" || ko "37 ($(state 37) $(field 37 'x["body"]'))"
[ "$(state 41)" = "('open', None)" ] && [ "$(unticked 41)" -eq 1 ] && ok "a verdict citing only the note's own exact line proves nothing: stays open" || ko "41 ($(state 41))"
[ "$(state 42)" = "('open', None)" ] && [ "$(unticked 42)" -eq 1 ] && ok "a valid ADR next to a line past the end: one bad citation keeps the note open" || ko "42 ($(state 42))"
[ "$(state 43)" = "('closed', 'completed')" ] && ok "a notes issue with a box ticked before this run closes as completed, not not planned" || ko "43 ($(state 43))"
field 43 'x["comments"][-1]["body"]' | grep -q '`n1` NOT A DEFECT ticked' && ko "43: the comment claims n1 was ticked by this run ($(field 43 'x["comments"][-1]["body"]'))" || ok "the comment never says ticked for a note this run did not tick (n1 was ticked before)"
[ "$(state 44)" = "('closed', 'not_planned')" ] && ok "a note on a file without an extension (Makefile) closes on a line of that file; a word like Retry:3 is not read as a file" || ko "44 ($(state 44) $(field 44 'x["comments"]'))"
[ "$(state 45)" = "('open', None)" ] && [ "$(unticked 45)" -eq 1 ] && ok "a note on a P0 path (deep-review-paths) is never ticked by a verdict, even with a valid own-file line" || ko "45 ($(state 45))"
for n in 46 47; do [ "$(state $n)" = "('open', None)" ] && ok "a P0 note recorded as ./ or pr/ path is normalised and stays open ($n)" || ko "$n ($(state $n))"; done
[ "$(state 48)" = "('open', None)" ] && ok "a note whose file is not on develop fails closed (stays open)" || ko "48 ($(state 48))"
[ "$(state 49)" = "('open', None)" ] && ok "a note on a sensitive path (sensitive-paths.txt) stays open" || ko "49 ($(state 49))"
[ "$(state 50)" = "('open', None)" ] && ok "a range that contains the note's own line does not count (1-2 on a note at line 1)" || ko "50 ($(state 50))"
[ "$(state 51)" = "('open', None)" ] && ok "a padded own line (01) does not count" || ko "51 ($(state 51))"
[ "$(state 52)" = "('open', None)" ] && ok "a note with no location is never ticked, even with a valid ADR" || ko "52 ($(state 52))"
grep -q 'contents read failed' <<<"$out" && ok "a failed contents read is logged, not silent" || ko "no log of the failed read ($out)"
[ "$(state 39)" = "('open', None)" ] && field 39 'x["comments"][-1]["body"]' | grep -q 'stays open: a human closes it as not planned' && ok "a FINDING with NOT A DEFECT and a valid reference stays open (a human closes it)" || ko "39 ($(state 39))"
[ "$(state 53)" = "('open', None)" ] && [ "$(unticked 53)" -eq 1 ] && ok "a citation of another existing line of the note's own file proves nothing: stays open" || ko "53 ($(state 53))"
[ "$(state 54)" = "('open', None)" ] && [ "$(unticked 54)" -eq 1 ] && ok "an unparsable own line with an own-file cite: stays open" || ko "54 ($(state 54))"
[ "$(state 55)" = "('open', None)" ] && [ "$(unticked 55)" -eq 1 ] && ok "an ADR whose Status is not ACCEPTED (even naming the file): stays open" || ko "55 ($(state 55))"
[ "$(state 56)" = "('open', None)" ] && [ "$(unticked 56)" -eq 1 ] && ok "an ACCEPTED ADR that does not name the note's file: stays open" || ko "56 ($(state 56))"
[ "$(state 57)" = "('closed', 'not_planned')" ] && [ "$(unticked 57)" -eq 0 ] && ok "an ACCEPTED ADR that names the note's file ticks it and the issue closes as not planned" || ko "57 ($(state 57))"
[ "$(state 58)" = "('open', None)" ] && [ "$(unticked 58)" -eq 1 ] && ok "a good ADR next to a line past the end of the own file: nothing is accepted" || ko "58 ($(state 58))"
[ "$(state 59)" = "('open', None)" ] && [ "$(unticked 59)" -eq 1 ] && ok "a note whose own file is \"?\" is never ticked" || ko "59 ($(state 59))"
[ "$(state 60)" = "('closed', 'not_planned')" ] && ok "an ADR that counts next to an existing ADR that does not name the file: the counting one is enough" || ko "60 ($(state 60))"
[ "$(state 61)" = "('open', None)" ] && [ "$(unticked 61)" -eq 1 ] && ok "a PROPOSED ADR with a later 'Status: accepted' line in a fence: stays open (the first Status line decides)" || ko "61 ($(state 61))"
[ "$(state 62)" = "('open', None)" ] && [ "$(unticked 62)" -eq 1 ] && ok "'ACCEPTED, superseded by ...' is not ACCEPTED: stays open" || ko "62 ($(state 62))"
[ "$(state 63)" = "('open', None)" ] && [ "$(unticked 63)" -eq 1 ] && ok "an ADR that only names the file inside a longer path or URL (old/src/a.go, src/a.go.bak): stays open" || ko "63 ($(state 63))"
[ "$(state 64)" = "('closed', 'not_planned')" ] && [ "$(unticked 64)" -eq 0 ] && ok "an upper-case '**Status:** ACCEPTED (owner, date)' ADR naming the file in backticks and at a sentence end ticks the note" || ko "64 ($(state 64))"
[ "$(state 65)" = "('open', None)" ] && [ "$(unticked 65)" -eq 1 ] && ok "an ACCEPTED ADR that names the file but not the note's id (a passing mention): stays open" || ko "65 ($(state 65))"
[ "$(state 66)" = "('open', None)" ] && [ "$(unticked 66)" -eq 1 ] && ok "an ACCEPTED ADR naming the file and only longer ids (n1-2, n10, a-n1): stays open" || ko "66 ($(state 66))"
[ "$(state 67)" = "('closed', 'not_planned')" ] && [ "$(unticked 67)" -eq 0 ] && ok "an ACCEPTED ADR naming the file and the exact id (backticks, parenthesis, comma) ticks the note" || ko "67 ($(state 67))"
[ "$(state 68)" = "('open', None)" ] && [ "$(unticked 68)" -eq 1 ] && ok "the id match is case-sensitive (N1 is not n1): stays open" || ko "68 ($(state 68))"
[ "$(state 69)" = "('closed', 'not_planned')" ] && [ "$(unticked 69)" -eq 0 ] && ok "an ADR naming the id as the ledger wrote it (a.b, not the slug) ticks the note" || ko "69 ($(state 69))"
[ "$(state 70)" = "('open', None)" ] && [ "$(unticked 70)" -eq 1 ] && ok "an ADR naming only the slug form (a-b) of the ledger id a.b: stays open" || ko "70 ($(state 70))"
[ "$(state 40)" = "('closed', 'completed')" ] && ok "the batch issue still closes last" || ko "40 ($(state 40))"

# A missing or empty note id counts nothing (fail closed), whatever the ADR says.
empty="$(PATH="$tmp/bin:$PATH" ST="$tmp/st" python3 -I - "$root/scripts" <<'PY'
import sys
sys.path.insert(0, sys.argv[1])
import findings
adr = "# ADR\n\n- **Status:** ACCEPTED\n\nCovers src/a.go for `n1`.\n"
print(findings.adr_counts(adr, "src/a.go", "n1"), findings.adr_counts(adr, "src/a.go", ""), findings.adr_counts(adr, "src/a.go"))
PY
)"
[ "$empty" = "True False False" ] && ok "an empty or missing note id makes adr_counts false (fail closed)" || ko "empty id ($empty)"

# The path mention itself never supplies the id: an id equal to a path segment or file stem, or a version or file token.
seg="$(PATH="$tmp/bin:$PATH" ST="$tmp/st" python3 -I - "$root/scripts" <<'PY'
import sys
sys.path.insert(0, sys.argv[1])
import findings
a = "# ADR\n\n- **Status:** ACCEPTED\n\n"
c = findings.adr_counts
print(c(a + "src/api/handler.go", "src/api/handler.go", "handler"), c(a + "src/api/handler.go", "src/api/handler.go", "api"),
      c(a + "TODO.md here", "TODO.md", "TODO"), c(a + "src/a.go and see n1.go", "src/a.go", "n1"), c(a + "src/a.go v n1.2", "src/a.go", "n1"),
      c(a + "src/a.go and the note handler.", "src/a.go", "handler"), c(a + "src/api/handler.go, note `handler`.", "src/api/handler.go", "handler"))
PY
)"
[ "$seg" = "False False False False False False True" ] && ok "the path mention does not supply the id; only a backticked id mention does" || ko "segment ids ($seg)"

# The id counts only when the ADR writes it as code, `id`: a plain word that happens to equal the id (owner, 2026-11-15:
# a passing mention is not a decision) never counts.
plain="$(PATH="$tmp/bin:$PATH" ST="$tmp/st" python3 -I - "$root/scripts" <<'PY'
import sys
sys.path.insert(0, sys.argv[1])
import findings
a = "# ADR\n\n- **Status:** ACCEPTED\n\nCovers src/a.go. "
c = findings.adr_counts
print(c(a + "The handler is fine.", "src/a.go", "handler"), c(a + "Note handler-x is decided.", "src/a.go", "handler-x"),
      c(a + "Note `handler-x` is decided.", "src/a.go", "handler-x"), c(a + "Note `handler-x-2` only.", "src/a.go", "handler-x"))
PY
)"
[ "$plain" = "False False True False" ] && ok "a plain-word id never counts; only \`id\` in backticks does" || ko "plain-word ids ($plain)"

before="$(cat "$tmp/st/issues.json")"
: >"$tmp/st/log"
run 170
[ "$rc" -eq 0 ] && [ ! -s "$tmp/st/log" ] && [ "$before" = "$(cat "$tmp/st/issues.json")" ] && ok "a second run changes nothing (idempotent)" || ko "idempotent (rc=$rc)"

# An unreadable P0 list fails closed: without deep-review-paths.txt, a note that would close stays open.
setup
rm "$tmp/st/tree/.github/deep-review-paths.txt"
run 170
[ "$(state 30)" = "('open', None)" ] && ok "an unreadable P0 list keeps every note open (fails closed)" || ko "missing list ($(state 30))"

finish
