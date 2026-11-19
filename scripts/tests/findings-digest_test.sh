#!/usr/bin/env bash
# Tests for the owner-decision loop of the review batch (`scripts/findings.py digest` and `digest-apply`, workflow
# .github/workflows/review-digest.yml). Items a no-commit batch run gave a verdict (batch-handled) stay open for the owner
# and out of new batches; `digest` lists them in ONE "Décisions en attente (revue IA)" issue, grouped by verdict, and
# `digest-apply` acts on the sources only for the rows the OWNER ticked in the `issues: edited` event (sender and the
# body before the edit, both set by GitHub). A model verdict never closes anything by itself (ADR-0014).
# Stateful fake `gh`; the batch bodies are rendered with findings.py itself so the held rows match the ranking.
# Usage: scripts/tests/findings-digest_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
tool="$root/scripts/findings.py"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin" "$tmp/st"

# Fake gh: issues in $ST/issues.json, each with a "comments" list. GET issues filters by `labels=`; writes are logged to
# $ST/log (METHOD PATH FIELDS). FAKE_GH_FAIL fails every call; FAKE_GH_FAIL_WRITE fails every write.
cat >"$tmp/bin/gh" <<'EOF'
#!/usr/bin/env python3
import json, os, sys
st = os.environ["ST"]
if os.environ.get("FAKE_GH_FAIL"):
    sys.stderr.write("gh: boom\n"); sys.exit(1)
args = sys.argv[1:]
method, fields, endpoint, i = "GET", {}, None, 1
while i < len(args):
    a = args[i]
    if a == "-X": method = args[i + 1]; i += 2
    elif a == "-f": k, v = args[i + 1].split("=", 1); fields.setdefault(k, []).append(v); i += 2
    elif a == "--paginate": i += 1
    else: endpoint = a; i += 1
path, _, query = endpoint.partition("?")
parts = path.split("/")
if method != "GET":
    if os.environ.get("FAKE_GH_FAIL_WRITE"):
        sys.stderr.write("gh: write refused\n"); sys.exit(1)
    open(os.path.join(st, "log"), "a").write("%s %s %s\n" % (method, path, json.dumps(fields, sort_keys=True)))
issues = json.load(open(os.path.join(st, "issues.json")))
def save(): json.dump(issues, open(os.path.join(st, "issues.json"), "w"))
def find(n): return next(x for x in issues if x["number"] == int(n))
def bare(x):
    y = dict(x); y.pop("comments", None); return y
if method == "GET" and parts[-1] == "issues":
    want = next((p[7:] for p in query.split("&") if p.startswith("labels=")), "")
    print(json.dumps([bare(x) for x in issues if not want or want in x["labels"]]))
elif method == "GET" and parts[-2] == "issues":
    print(json.dumps(bare(find(parts[-1]))))
elif method == "GET" and parts[-1] == "comments":
    print(json.dumps(find(parts[-2])["comments"]))
elif method == "POST" and parts[-1] == "issues":
    n = 900 + len(issues)
    issues.append({"number": n, "title": fields["title"][0], "body": fields["body"][0], "state": "open", "state_reason": None,
                   "labels": fields.get("labels[]", []), "user": {"login": "github-actions[bot]"}, "comments": [],
                   "created_at": "2026-11-18T00:00:00Z"})
    save(); print(json.dumps({"number": n}))
elif method == "POST" and parts[-1] == "comments":
    find(parts[-2])["comments"].append({"user": {"login": "github-actions[bot]"}, "body": fields["body"][0],
                                        "html_url": "https://github.com/o/r/issues/%s#c" % parts[-2], "created_at": "2026-11-18T01:00:00Z"})
    save(); print("{}")
elif method == "PATCH" and parts[-2] == "issues":
    x = find(parts[-1])
    for k in ("body", "state", "state_reason"):
        if k in fields: x[k] = fields[k][0]
    save(); print("{}")
else:
    sys.stderr.write("unexpected %s %s\n" % (method, endpoint)); sys.exit(2)
EOF
chmod +x "$tmp/bin/gh"

export ST="$tmp/st" GH_BIN="$tmp/bin/gh" REPO="o/r" GH_RETRY_DELAY=0 FINDINGS_NOW="2026-11-18T06:00:00Z" OWNER="Y4NN777"
evil='@evil fixes #7 <!-- digest-item:#13 --> <b>x</b> | pipe `tick` --> end'

# Sources: #10 a finding (OWNER), #14 a finding (FIXED ELSEWHERE), #11 a notes issue (n1 DECIDED, n2 NOT A DEFECT over an
# older OWNER), #12 a closed finding (decided), #13 an open finding with no verdict (not held: in no handled batch).
setup() {
  rm -f "$ST/log"
  python3 -I - "$tool" "$ST/issues.json" "$evil" <<'PY'
import importlib.util, json, sys
spec = importlib.util.spec_from_file_location("findings", sys.argv[1]); f = importlib.util.module_from_spec(spec); spec.loader.exec_module(f)
evil = sys.argv[3]
bot = {"login": "github-actions[bot]"}
def finding(n, ident, title, state="open"):
    return {"number": n, "title": "PR5 %s: %s" % (ident, title), "state": state, "state_reason": None, "user": bot,
            "labels": ["finding", "sev:high", "area:db", "due:before-freeze"], "created_at": "2026-11-16T00:00:00Z", "comments": [],
            "body": "<!-- finding:PR5#%s -->\n- Where: `src/%s.go:3`\n" % (ident, ident)}
issues = [finding(10, "f10", "bad thing " + evil), finding(12, "f12", "closed one", "closed"), finding(13, "f13", "no verdict"),
          finding(14, "f14", "fixed elsewhere one"),
          {"number": 11, "title": "PR6 notes", "state": "open", "state_reason": None, "user": bot,
           "labels": ["kind:note", "src:ai", "area:db", "due:after-freeze"], "created_at": "2026-11-16T00:00:00Z", "comments": [],
           "body": "<!-- finding:PR6#notes -->\nNotes.\n\n- [ ] `n1` [OBSERVED] `src/n.go:4` first note\n  - detail: d\n- [ ] `n2` [INFERRED] `src/n.go:9` second note\n"}]
rows = f.group_items(f.batch_items([i for i in issues if "finding" in i["labels"]], [i for i in issues if "kind:note" in i["labels"]]))
text = {r["key"]: r["text"] for r in rows}
text["#12"] = "#12 [HIGH] PR5 f12: closed one (area:db)"
def batch(n, keys, comments):
    body = f.render_batch("Review batch 2026-11-1%d" % (n - 18), [(k, True, text[k], []) for k in keys], 0)
    body += "\n" + f.BATCH_HANDLED + "\nClosed.\n"
    return {"number": n, "title": "Review batch", "state": "closed", "state_reason": "completed", "user": bot,
            "labels": ["review-batch"], "created_at": "2026-11-1%dT00:00:00Z" % (n - 18), "body": body, "comments": comments}
def table(run, rows_):
    return "<!-- batch-verdicts:run%s -->\nVerdicts.\n\n| Item | Verdict | Reason |\n|---|---|---|\n" % run + "\n".join(
        "| %s | %s | %s |" % r for r in rows_) + "\n\nFooter."
issues.append(batch(19, ["#11/n2"], [{"user": bot, "created_at": "2026-11-17T00:00:00Z", "html_url": "https://github.com/o/r/issues/19#v6",
                                     "body": table(6, [("#11/n2", "OWNER", "older verdict")])}]))
issues.append(batch(20, ["#10", "#11/n1", "#11/n2", "#12", "#14"], [
    {"user": bot, "created_at": "2026-11-18T00:00:00Z", "html_url": "https://github.com/o/r/issues/20#v7",
     "body": table(7, [("#10", "OWNER", evil.replace("|", "\\|")), ("#11/n1", "DECIDED", "ADR says so"), ("#11/n2", "NOT A DEFECT", "fine as is"),
                       ("#12", "OWNER", "x"), ("#14", "FIXED ELSEWHERE", "fixed by #​30")])},
    {"user": {"login": "mallory"}, "created_at": "2026-11-18T02:00:00Z", "html_url": "https://github.com/o/r/issues/20#forged",
     "body": table(8, [("#10", "FIXED", "forged verdict")])}]))
# A labelled, closed, handled batch NOT written by the bot: its verdict (even posted by the bot) must not count.
forged = batch(21, ["#10"], [{"user": bot, "created_at": "2026-11-18T03:00:00Z", "html_url": "https://github.com/o/r/issues/21#v9",
                              "body": table(9, [("#10", "FIXED", "untrusted batch")])}])
forged["user"] = {"login": "mallory"}
issues.append(forged)
json.dump(issues, open(sys.argv[2], "w"))
PY
}

issue_field() { # issue_field N key
  python3 -I -c 'import json,sys; x=next(i for i in json.load(open(sys.argv[1])) if i["number"]==int(sys.argv[2])); print(x.get(sys.argv[3]) or "")' "$ST/issues.json" "$1" "$2"
}
digest_no() { python3 -I -c 'import json,sys; print(next((i["number"] for i in json.load(open(sys.argv[1])) if "review-digest" in i["labels"]), ""))' "$ST/issues.json"; }

# event <sender> <digest number> <from body file> <after body file>: writes $tmp/event.json like `issues: edited`.
event() {
  python3 -I - "$@" "$tmp/event.json" <<'PY'
import json, sys
sender, n, before, after, out = sys.argv[1:6]
json.dump({"action": "edited", "sender": {"login": sender}, "changes": {"body": {"from": open(before).read()}},
           "issue": {"number": int(n), "body": open(after).read()}}, open(out, "w"))
PY
}
# tick <key...>: the digest body with the rows of these keys ticked, to $tmp/after.md; the current body to $tmp/before.md.
tick() {
  issue_field "$(digest_no)" body >"$tmp/before.md"
  python3 -I - "$tmp/before.md" "$tmp/after.md" "$@" <<'PY'
import sys
src, out, keys = sys.argv[1], sys.argv[2], sys.argv[3:]
lines = []
for line in open(src).read().split("\n"):
    if any(line.endswith("<!-- digest-item:%s -->" % k) for k in keys):
        line = line.replace("- [ ]", "- [x]", 1)
    lines.append(line)
open(out, "w").write("\n".join(lines))
PY
}
# apply_owner_edit: the digest issue now holds the edited body (the owner saved it), then digest-apply runs on the event.
set_digest_body() {
  python3 -I - "$ST/issues.json" "$(digest_no)" "$1" <<'PY'
import json, sys
p, n, f = sys.argv[1], int(sys.argv[2]), sys.argv[3]
d = json.load(open(p)); next(i for i in d if i["number"] == n)["body"] = open(f).read(); json.dump(d, open(p, "w"))
PY
}
apply() { GITHUB_EVENT_PATH="$tmp/event.json" python3 -I "$tool" digest-apply >"$tmp/out" 2>&1; }

# --- digest lists the held items grouped by verdict ------------------------------------------------------------------
setup
if python3 -I "$tool" digest >"$tmp/out" 2>&1; then ok "digest exits 0"; else ko "digest exits 0: $(cat "$tmp/out")"; fi
n="$(digest_no)"
if [ -n "$n" ] && [ "$(issue_field "$n" title)" = "Décisions en attente (revue IA)" ]; then ok "digest creates the issue with its title"; else ko "digest issue: $(cat "$tmp/out")"; fi
body="$(issue_field "$n" body 2>/dev/null || true)"
printf '%s\n' "$body" >"$tmp/digest.md"
grep -q '^<!-- review-digest -->$' "$tmp/digest.md" && ok "stable marker" || ko "stable marker"
row() { grep -E "<!-- digest-item:$1 -->\$" "$tmp/digest.md" || true; }
section_of() { awk -v k="<!-- digest-item:$1 -->" '/^## /{s=$0} index($0,k){print s}' "$tmp/digest.md"; }
[ "$(section_of '#10')" = "## OWNER" ] && ok "#10 under OWNER" || ko "#10 section: $(section_of '#10')"
[ "$(section_of '#11/n1')" = "## DECIDED" ] && ok "note n1 under DECIDED" || ko "n1 section: $(section_of '#11/n1')"
[ "$(section_of '#11/n2')" = "## NOT A DEFECT" ] && ok "newest verdict wins (n2 NOT A DEFECT over OWNER)" || ko "n2 section: $(section_of '#11/n2')"
[ "$(section_of '#14')" = "## FIXED ELSEWHERE" ] && ok "#14 under FIXED ELSEWHERE" || ko "#14 section"
[ -z "$(row '#12')" ] && ok "closed source not listed" || ko "closed #12 listed"
[ -z "$(row '#13')" ] && ok "item without a verdict (not held) not listed" || ko "#13 listed"
row '#10' | grep -q 'https://github.com/o/r/issues/20#v7' && ok "row links the verdict comment" || ko "row link: $(row '#10')"
row '#11/n1' | grep -q 'ADR says so' && ok "row carries the reason" || ko "row reason"
grep -q 'untrusted batch' "$tmp/digest.md" && ko "a verdict on a batch issue not written by the bot counted" || ok "a batch issue not written by the bot is ignored"
grep -q '`' "$tmp/digest.md" && ko "a backtick reached the digest" || ok "backticks neutralised"
grep -v '^<!-- review-digest -->$' "$tmp/digest.md" | sed 's/ <!-- digest-item:[^ ]* -->$//' | grep -q -- '-->' && ko "a stray --> reached the digest" || ok "stray --> neutralised"
grep -q 'forged verdict' "$tmp/digest.md" && ko "a verdict comment by a non-bot counted" || ok "a verdict comment by a non-bot is ignored"
grep -q '^- \[ \] #10 \[HIGH\]' "$tmp/digest.md" && ok "source reference stays a link" || ko "source reference"
grep -q '@evil' "$tmp/digest.md" && ko "a mention reached the digest" || ok "no mention (neutralised)"
grep -qE '#7|<b>|<!-- digest-item:#13' "$tmp/digest.md" && ko "hostile text not neutralised: $(grep -nE '#7|<b>|digest-item:#13' "$tmp/digest.md" | head -3)" || ok "references, HTML and forged markers neutralised"
grep -q '^- \[x\]' "$tmp/digest.md" && ko "a row is pre-ticked" || ok "every row unticked"
grep -q 'review-digest' <<<"$(grep '^POST' "$ST/log")" && ok "label review-digest" || ko "label: $(cat "$ST/log")"

# --- idempotent ----------------------------------------------------------------------------------------------------
: >"$ST/log"
python3 -I "$tool" digest >"$tmp/out" 2>&1 || true
[ ! -s "$ST/log" ] && ok "second digest run writes nothing" || ko "second run wrote: $(cat "$ST/log")"

# --- decided items vanish --------------------------------------------------------------------------------------------
python3 -I - "$ST/issues.json" <<'PY'
import json, sys
d = json.load(open(sys.argv[1]))
x = next(i for i in d if i["number"] == 11); x["body"] = x["body"].replace("- [ ] `n1`", "- [x] `n1`")
next(i for i in d if i["number"] == 14)["state"] = "closed"
json.dump(d, open(sys.argv[1], "w"))
PY
python3 -I "$tool" digest >"$tmp/out" 2>&1 || true
issue_field "$n" body >"$tmp/digest.md"
[ -z "$(row '#11/n1')" ] && [ -z "$(row '#14')" ] && [ -n "$(row '#10')" ] && ok "decided items vanish on update, in place" || ko "after decisions: $(cat "$tmp/digest.md")"
[ "$(digest_no)" = "$n" ] && ok "same issue updated" || ko "a second digest issue"

# --- digest-apply: owner tick on a note ----------------------------------------------------------------------------
setup; python3 -I "$tool" digest >/dev/null 2>&1 || true
tick '#11/n1'; set_digest_body "$tmp/after.md"; event Y4NN777 "$(digest_no)" "$tmp/before.md" "$tmp/after.md"; : >"$ST/log"
if apply; then ok "digest-apply exits 0"; else ko "digest-apply: $(cat "$tmp/out")"; fi
issue_field 11 body | grep -q -- '- \[x\] `n1`.*décidé par le propriétaire (digest #'"$(digest_no)"')' && ok "owner tick ticks the note with the suffix" || ko "note not ticked: $(issue_field 11 body) / $(cat "$tmp/out")"
issue_field 11 body | grep -q -- '- \[ \] `n2`' && ok "the other note stays" || ko "n2 changed"
[ "$(issue_field 11 state)" = "open" ] && ok "notes issue with an open box stays open" || ko "notes issue closed early"
grep -q 'POST repos/o/r/issues/11/comments' "$ST/log" && ok "comment on the notes issue" || ko "no comment"
issue_field "$(digest_no)" body >"$tmp/digest.md"
[ -z "$(row '#11/n1')" ] && ok "applied row removed from the digest" || ko "row still there"
: >"$ST/log"; apply || true
grep -q '11/comments' "$ST/log" && ko "re-run commented twice" || ok "re-run of the same event is idempotent"

# --- two quick owner ticks: the second event is based on the body before the first run rewrote it --------------
setup; python3 -I "$tool" digest >/dev/null 2>&1 || true
tick '#11/n1'; cp "$tmp/before.md" "$tmp/b1.md"; cp "$tmp/after.md" "$tmp/a1.md"
python3 -I - "$tmp/a1.md" "$tmp/a2.md" <<'PY'
import sys
out = [l.replace("- [ ]", "- [x]", 1) if l.endswith("<!-- digest-item:#11/n2 -->") else l for l in open(sys.argv[1]).read().split("\n")]
open(sys.argv[2], "w").write("\n".join(out))
PY
set_digest_body "$tmp/a2.md"
event Y4NN777 "$(digest_no)" "$tmp/b1.md" "$tmp/a1.md"; apply || ko "apply 1: $(cat "$tmp/out")"
event Y4NN777 "$(digest_no)" "$tmp/a1.md" "$tmp/a2.md"; apply || ko "apply 2: $(cat "$tmp/out")"
issue_field 11 body | grep -q -- '- \[x\] `n2`' && ok "second quick owner tick is applied after the first run rewrote the digest" || ko "second tick lost: $(cat "$tmp/out")"
[ "$(issue_field 11 state)" = "closed" ] && ok "both notes settled" || ko "#11 state $(issue_field 11 state)"
: >"$ST/log"; apply || true
grep -q '11/comments' "$ST/log" && ko "replay of the second event acted again" || ok "replay of a settled row is a no-op"

# --- the footer counts in the cap ----------------------------------------------------------------------------------
if python3 -I - "$tool" >"$tmp/cap" 2>&1 <<'PY'
import importlib.util, sys
spec = importlib.util.spec_from_file_location("findings", sys.argv[1]); f = importlib.util.module_from_spec(spec); spec.loader.exec_module(f)
bad = []
for width in range(200, 1400, 7):
    rows = [("OWNER", "#%d" % i, ["#%d" % i], "#%d %s" % (i, "y" * width)) for i in range(1000, 1400)]
    body = f.render_digest(rows)
    if len(body) > f.DIGEST_MAX or "de plus" not in body:
        bad.append((width, len(body)))
print(bad[:3]); sys.exit(1 if bad else 0)
PY
then ok "digest with the footer stays within the cap"; else ko "cap exceeded: $(cat "$tmp/cap")"; fi

# --- a box ticked by someone else ----------------------------------------------------------------------------------
setup; python3 -I "$tool" digest >/dev/null 2>&1 || true
tick '#11/n1'; set_digest_body "$tmp/after.md"; event mallory "$(digest_no)" "$tmp/before.md" "$tmp/after.md"; : >"$ST/log"
apply || true
grep -q 'issues/11' "$ST/log" && ko "a non-owner tick acted" || ok "a box ticked by someone else is ignored"
grep -q 'NO-OP' "$tmp/out" && ok "says NO-OP" || ko "output: $(cat "$tmp/out")"
# Then the owner edits something else: the box was ticked before the owner's edit, so it does not count.
cp "$tmp/after.md" "$tmp/before2.md"; printf '\n' >>"$tmp/after.md"; set_digest_body "$tmp/after.md"
event Y4NN777 "$(digest_no)" "$tmp/before2.md" "$tmp/after.md"; : >"$ST/log"; apply || true
grep -q 'issues/11' "$ST/log" && ko "a box ticked before the owner's edit acted" || ok "a box already ticked before the owner's edit does not count"

# --- a tampered row (key rewritten) is refused ---------------------------------------------------------------------
setup; python3 -I "$tool" digest >/dev/null 2>&1 || true
issue_field "$(digest_no)" body | sed 's/<!-- digest-item:#10 -->$/<!-- digest-item:#13 -->/' >"$tmp/before.md"
sed 's/^- \[ \]\(.*digest-item:#13 -->\)$/- [x]\1/' "$tmp/before.md" >"$tmp/after.md"
set_digest_body "$tmp/after.md"; event Y4NN777 "$(digest_no)" "$tmp/before.md" "$tmp/after.md"; : >"$ST/log"; apply || true
grep -qE 'issues/1[03]' "$ST/log" && ko "a tampered row acted: $(cat "$ST/log")" || ok "a row whose key or text was changed is refused"

# --- fully ticked notes issue closes as not planned; findings close with the right reason --------------------------
setup; python3 -I "$tool" digest >/dev/null 2>&1 || true
tick '#11/n1' '#11/n2' '#10' '#14'; set_digest_body "$tmp/after.md"; event Y4NN777 "$(digest_no)" "$tmp/before.md" "$tmp/after.md"
apply || ko "apply: $(cat "$tmp/out")"
[ "$(issue_field 11 state)/$(issue_field 11 state_reason)" = "closed/not_planned" ] && ok "fully ticked notes issue closes as not planned" || ko "#11: $(issue_field 11 state)/$(issue_field 11 state_reason)"
[ "$(issue_field 10 state)/$(issue_field 10 state_reason)" = "closed/not_planned" ] && ok "OWNER finding closes as not planned" || ko "#10: $(issue_field 10 state)/$(issue_field 10 state_reason)"
[ "$(issue_field 14 state)/$(issue_field 14 state_reason)" = "closed/completed" ] && ok "FIXED ELSEWHERE finding closes as completed" || ko "#14: $(issue_field 14 state)/$(issue_field 14 state_reason)"
grep -q '#10' <<<"$(issue_field "$(digest_no)" body)" && ko "applied rows still in the digest" || ok "digest re-rendered without the applied rows"

# --- fail closed ---------------------------------------------------------------------------------------------------
setup; python3 -I "$tool" digest >/dev/null 2>&1 || true
tick '#10'; set_digest_body "$tmp/after.md"; event Y4NN777 "$(digest_no)" "$tmp/before.md" "$tmp/after.md"; : >"$ST/log"
if FAKE_GH_FAIL=1 GITHUB_EVENT_PATH="$tmp/event.json" python3 -I "$tool" digest-apply >"$tmp/out" 2>&1; then ko "apply on a gh error exits 0"; else ok "apply fails closed on a gh error"; fi
grep -q 'FAILED' "$tmp/out" && ok "says FAILED" || ko "output: $(cat "$tmp/out")"
[ ! -s "$ST/log" ] && [ "$(issue_field 10 state)" = "open" ] && ok "nothing written after the failure" || ko "wrote: $(cat "$ST/log")"
if FAKE_GH_FAIL=1 python3 -I "$tool" digest >"$tmp/out" 2>&1; then ko "digest on a gh error exits 0"; else ok "digest fails closed on a gh error"; fi
if FAKE_GH_FAIL_WRITE=1 GITHUB_EVENT_PATH="$tmp/event.json" python3 -I "$tool" digest-apply >"$tmp/out" 2>&1; then ko "apply on a refused write exits 0"; else ok "apply fails closed on a refused write"; fi
GITHUB_EVENT_PATH="$tmp/missing.json" python3 -I "$tool" digest-apply >"$tmp/out" 2>&1 && rc=0 || rc=$?
[ "$rc" -ne 0 ] && ok "apply without a readable event fails closed" || ko "apply without event exit 0"

# --- workflow static checks -----------------------------------------------------------------------------------------
wf="$root/.github/workflows/review-digest.yml"
if [ -f "$wf" ] && python3 -I - "$wf" >"$tmp/wf" 2>&1 <<'PY'
import re, sys, yaml
path = sys.argv[1]
text = open(path).read()
w = yaml.safe_load(text)
on = w.get("on", w.get(True))
errs = []
if [c.get("cron") for c in on.get("schedule", [])] != ["0 6 * * 1"]: errs.append("weekly schedule")
if (on.get("issues") or {}).get("types") != ["edited"]: errs.append("issues: edited only")
if set(on) - {"schedule", "workflow_dispatch", "issues"}: errs.append("extra triggers %s" % (set(on) - {"schedule", "workflow_dispatch", "issues"}))
if w.get("permissions") != {"contents": "read"}: errs.append("top permissions")
for name, job in w["jobs"].items():
    if "environment" in job: errs.append("%s has an environment" % name)
    perms = job.get("permissions") or {}
    if perms != {"contents": "read", "issues": "write"}: errs.append("%s permissions %s" % (name, perms))
    for step in job.get("steps", []):
        if "uses" in step and not re.search(r"@[0-9a-f]{40}\b", step["uses"]): errs.append("unpinned %s" % step["uses"])
        if step.get("uses", "").startswith("actions/checkout"):
            if (step.get("with") or {}).get("ref") != "develop" or (step.get("with") or {}).get("persist-credentials") is not False:
                errs.append("%s checkout must be develop without credentials" % name)
        if "${{" in step.get("run", ""): errs.append("%s: expression inside run" % name)
        if "run" in step and "findings.py" in step["run"] and "python3 -I" not in step["run"]: errs.append("python without -I")
apply = w["jobs"].get("apply") or {}
if "issues" not in str(apply.get("if")) or "review-digest" not in str(apply.get("if")): errs.append("apply if: issues event and label")
if "concurrency" in apply: errs.append("apply must not share a concurrency group (a pending owner edit would be dropped)")
if "digest-apply" not in str(apply): errs.append("apply runs digest-apply")
if "digest" not in str(w["jobs"].get("digest")): errs.append("digest job")
print("\n".join(errs))
sys.exit(1 if errs else 0)
PY
then ok "review-digest.yml: triggers, least privilege, pins, env values, default-branch scripts"; else ko "workflow: $(cat "$tmp/wf" 2>/dev/null)"; fi

finish
