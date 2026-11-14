#!/usr/bin/env bash
# Tests for scripts/findings.py with a stateful fake `gh`: on merge, the open findings of the review's ledger become
# issues (CRITICAL and HIGH one each, up to 5; everything else one follow-ups checklist per pull request), once
# (idempotent by marker), and the dashboard ranks what to fix next from issue state only.
# Usage: scripts/tests/findings_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
tool="$root/scripts/findings.py"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin" "$tmp/st"

# Fake gh (python): `gh api [-X METHOD] [-f k=v]... ENDPOINT [--paginate]`. State in $ST: comments.json, pull.json,
# issues.json. POST repos/o/r/issues appends an issue; PATCH repos/o/r/issues/N updates it; every write is logged.
cat >"$tmp/bin/gh" <<'EOF'
#!/usr/bin/env python3
import json, os, sys
st = os.environ["ST"]
args = sys.argv[1:]
if os.environ.get("FAKE_GH_FAIL"):
    sys.stderr.write("gh: boom\n"); sys.exit(1)
assert args[0] == "api", args
method, fields, endpoint = "GET", {}, None
i = 1
while i < len(args):
    a = args[i]
    if a == "-X": method = args[i + 1]; i += 2
    elif a == "-f":
        k, v = args[i + 1].split("=", 1); fields.setdefault(k, []).append(v); i += 2
    elif a == "--paginate": i += 1
    else: endpoint = a; i += 1
def load(name):
    return json.load(open(os.path.join(st, name)))
def save(name, data):
    json.dump(data, open(os.path.join(st, name), "w"))
def log(entry):
    open(os.path.join(st, "log.jsonl"), "a").write(json.dumps(entry) + "\n")
path = endpoint.split("?")[0]
if method == "GET":
    if path.endswith("/comments"): print(json.dumps(load("comments.json")))
    elif "/compare/" in path: print(json.dumps(load("compare.json")))
    elif "/pulls/" in path: print(json.dumps(load("pull.json")))
    elif path.endswith("/issues"):
        query = endpoint.split("?", 1)[1] if "?" in endpoint else ""
        want = next((p[7:] for p in query.split("&") if p.startswith("labels=")), "finding")
        print(json.dumps([x for x in load("issues.json") if want in [l if isinstance(l, str) else l["name"] for l in x["labels"]]]))
    else: sys.stderr.write("unexpected GET " + endpoint + "\n"); sys.exit(2)
elif method == "POST" and path.endswith("/issues"):
    if fields.get("assignees[]") and os.environ.get("FAKE_ASSIGN_FAIL"):
        sys.stderr.write("assignee is not a collaborator\n"); sys.exit(1)
    issues = load("issues.json")
    issue = {"number": 100 + len(issues), "title": fields["title"][0], "body": fields["body"][0], "state": "open",
             "state_reason": None, "labels": fields.get("labels[]", []), "assignees": fields.get("assignees[]", []),
             "created_at": "2026-11-13T12:00:00Z", "html_url": "https://example.invalid/issues/%d" % (100 + len(issues))}
    issues.append(issue); save("issues.json", issues)
    log({"op": "create", "title": issue["title"], "labels": issue["labels"], "assignees": issue["assignees"]})
    print(json.dumps(issue))
elif method == "PATCH" and "/issues/" in path:
    n = int(path.rsplit("/", 1)[1])
    issues = load("issues.json")
    for x in issues:
        if x["number"] == n:
            for k in ("title", "body", "state"):
                if k in fields: x[k] = fields[k][0]
    save("issues.json", issues); log({"op": "patch", "number": n}); print("{}")
else:
    sys.stderr.write("unexpected %s %s\n" % (method, endpoint)); sys.exit(2)
EOF
chmod +x "$tmp/bin/gh"

# ledger_comments FINDINGS_JSON: a comment list with one bot comment carrying the fenced JSON ledger.
ledger_comments() {
  python3 -I - "$1" <<'EOF'
import json, sys
body = "Triage: HIGH | blast radius: x\n\nsummary\n\n<details><summary>Review ledger</summary>\n\n```json\n" + \
       json.dumps({"sha": "abcdef1234567", "findings": json.loads(sys.argv[1])}) + "\n```\n\n</details>"
print(json.dumps([{"user": {"login": "github-actions[bot]"}, "created_at": "2026-11-13T10:00:00Z", "body": body}]))
EOF
}

reset() {
  printf '[]' >"$tmp/st/comments.json"
  printf '[]' >"$tmp/st/issues.json"
  printf '{"number":94,"title":"S1 profiles","merged":true,"user":{"login":"NICE-DEV226"},"head":{"ref":"feature/identity/profiles","sha":"fedcba9876543"}}' >"$tmp/st/pull.json"
  printf '{"files":[]}' >"$tmp/st/compare.json"
  : >"$tmp/st/log.jsonl"
}

# run ARGS...: runs the tool; sets rc and out.
run() {
  rc=0
  out="$(PATH="$tmp/bin:$PATH" ST="$tmp/st" REPO=o/r PR_NUMBER="${PRN:-94}" PR_AUTHOR=NICE-DEV226 python3 -I "$tool" "$@" 2>&1)" || rc=$?
}

# issues_json: the fake issues as a compact JSON for assertions.
count_issues() { python3 -I -c 'import json,sys; print(len(json.load(open(sys.argv[1]))))' "$tmp/st/issues.json"; }
field() { python3 -I -c 'import json,sys; d=json.load(open(sys.argv[1]))[int(sys.argv[2])]; v=d[sys.argv[3]]; print(v if isinstance(v,str) else json.dumps(v))' "$tmp/st/issues.json" "$1" "$2"; }

F_HIGH='{"id":"skills-upsert","sev":"HIGH","file":"db/queries/identity.sql","line":154,"scenario":"double submit returns 500","status":"open"}'
F_MED1='{"id":"comment-wrong","sev":"MEDIUM","file":"a.go","line":3,"scenario":"stale comment","status":"open"}'
F_MED2='{"id":"checklist","sev":"MEDIUM","file":"b.go","line":9,"scenario":"false never leaves the checklist","status":"open"}'
F_LOWFIX='{"id":"typo","sev":"LOW","file":"c.go","line":1,"scenario":"typo","status":"fixed"}'
F_ANSWERED='{"id":"ans","sev":"HIGH","file":"d.go","line":1,"scenario":"answered","status":"answered"}'

reset
ledger_comments "[$F_HIGH,$F_MED1,$F_MED2,$F_LOWFIX,$F_ANSWERED]" >"$tmp/st/comments.json"
run harvest
[ "$rc" -eq 0 ] && [ "$(count_issues)" -eq 2 ] && ok "one HIGH and two MEDIUM make 2 issues: one finding, one follow-ups (fixed and answered are not filed)" || ko "count (rc=$rc n=$(count_issues) out=$out)"
grep -q 'finding:PR94#skills-upsert' <<<"$(field 0 body)" && grep -q 'db/queries/identity.sql:154' <<<"$(field 0 body)" && grep -q 'double submit returns 500' <<<"$(field 0 body)" && ok "the finding issue carries the marker, the place and the failing scenario" || ko "body 0 ($(field 0 body))"
l="$(field 0 labels)"
grep -q '"finding"' <<<"$l" && grep -q '"sev:high"' <<<"$l" && grep -q '"src:ai"' <<<"$l" && grep -q '"area:identity"' <<<"$l" && grep -q '"due:before-freeze"' <<<"$l" && ok "labels: finding, sev, src, area from the branch, due:before-freeze for a HIGH" || ko "labels 0 ($l)"
grep -q 'NICE-DEV226' <<<"$(field 0 assignees)" && ok "the issue is assigned to the pull request author" || ko "assignee ($(field 0 assignees))"
grep -q 'PR94 follow-ups' <<<"$(field 1 title)" && grep -q 'finding:PR94#followups' <<<"$(field 1 body)" && grep -q -- '- \[ \].*a.go:3' <<<"$(field 1 body)" && grep -q -- '- \[ \].*b.go:9' <<<"$(field 1 body)" && ok "the follow-ups issue is a checklist of the MEDIUM findings" || ko "followups ($(field 1 body))"
! grep -q 'typo' <<<"$(field 1 body)" && ok "a fixed finding is not in the follow-ups" || ko "fixed in followups"
l="$(field 1 labels)"
grep -q '"sev:medium"' <<<"$l" && grep -q '"due:after-freeze"' <<<"$l" && ok "the follow-ups issue takes its highest severity and due:after-freeze" || ko "labels 1 ($l)"

run harvest
[ "$rc" -eq 0 ] && [ "$(count_issues)" -eq 2 ] && ok "harvest is idempotent: a second run creates nothing" || ko "idempotent (n=$(count_issues) out=$out)"

reset
many="[$(python3 -I -c 'import json; print(",".join(json.dumps({"id":"h%d"%i,"sev":"HIGH","file":"x.go","line":i,"scenario":"s%d"%i,"status":"open"}) for i in range(1,8)))')]"
ledger_comments "$many" >"$tmp/st/comments.json"
run harvest
[ "$(count_issues)" -eq 6 ] && ok "7 HIGH findings: at most 5 single issues, the overflow goes to the follow-ups issue" || ko "cap (n=$(count_issues))"
grep -q 'h6' <<<"$(field 5 body)" && grep -q 'h7' <<<"$(field 5 body)" && ok "the overflow findings are in the follow-ups checklist" || ko "overflow ($(field 5 body))"

reset
printf '[{"user":{"login":"github-actions[bot]"},"created_at":"2026-11-13T10:00:00Z","body":"<details><summary>Review ledger</summary>\\n```json\\n{not json\\n```\\n</details>"}]' >"$tmp/st/comments.json"
run harvest
[ "$rc" -eq 0 ] && [ "$(count_issues)" -eq 0 ] && grep -q 'skipped' <<<"$out" && ok "a malformed ledger is skipped and logged, never guessed" || ko "malformed (rc=$rc n=$(count_issues) out=$out)"

reset
run harvest
[ "$rc" -eq 0 ] && [ "$(count_issues)" -eq 0 ] && ok "no ledger: nothing is filed" || ko "no ledger"

reset
ledger_comments "[$F_HIGH]" >"$tmp/st/comments.json"
printf '{"number":94,"title":"S1","merged":false,"user":{"login":"NICE-DEV226"},"head":{"ref":"feature/identity/profiles"}}' >"$tmp/st/pull.json"
run harvest
[ "$(count_issues)" -eq 0 ] && ok "a pull request closed without a merge files nothing" || ko "not merged (n=$(count_issues))"

reset
ledger_comments "[$F_HIGH]" >"$tmp/st/comments.json"
FAKE_ASSIGN_FAIL=1 run harvest
[ "$(count_issues)" -eq 1 ] && [ "$(field 0 assignees)" = "[]" ] && ok "an author who cannot be assigned: the issue is still created, without an assignee" || ko "assign retry (n=$(count_issues) a=$(field 0 assignees) out=$out)"

reset
ledger_comments "[$F_HIGH]" >"$tmp/st/comments.json"
FAKE_GH_FAIL=1 run harvest
[ "$rc" -eq 0 ] && grep -q 'NOT RUN' <<<"$out" && ok "a GitHub error fails open (NOT RUN, exit 0)" || ko "gh failure (rc=$rc out=$out)"

# A stray @name in the ledger text must not notify anyone (review note on #115): the mention is neutralised.
reset
F_MENTION='{"id":"mention","sev":"HIGH","file":"a.go","line":1,"scenario":"@Y4NN777 should look: a bad scenario","status":"open"}'
ledger_comments "[$F_MENTION]" >"$tmp/st/comments.json"
run harvest
[ "$(count_issues)" -eq 1 ] && ! grep -q '@Y4NN777' <<<"$(field 0 title) $(field 0 body)" && grep -q 'Y4NN777' <<<"$(field 0 body)" && ok "an @mention in the ledger text is neutralised in the issue (the name stays readable)" || ko "mention (n=$(count_issues) t=$(field 0 title) b=$(field 0 body))"

# A human who removed the marker from an issue body must not cause a duplicate on a re-run (review note on #115): the
# title of the issue is checked too.
reset
ledger_comments "[$F_HIGH,$F_MED1]" >"$tmp/st/comments.json"
printf '[{"number":100,"title":"PR94 skills-upsert: double submit returns 500","body":"edited by a human, marker gone","state":"open","state_reason":null,"labels":["finding"],"created_at":"2026-11-13T09:00:00Z"},{"number":101,"title":"PR94 follow-ups","body":"edited by a human","state":"open","state_reason":null,"labels":["finding"],"created_at":"2026-11-13T09:00:00Z"}]' >"$tmp/st/issues.json"
run harvest
[ "$(count_issues)" -eq 2 ] && ok "an issue whose marker was edited away is recognised by its title: no duplicate" || ko "title dedupe (n=$(count_issues) out=$out)"

# The title carries the pull request number, so the title check is per pull request: the same finding id and wording on
# another pull request is a different issue (review note on #115).
reset
ledger_comments "[$F_HIGH]" >"$tmp/st/comments.json"
printf '[{"number":100,"title":"PR90 skills-upsert: double submit returns 500","body":"<!-- finding:PR90#skills-upsert -->","state":"open","state_reason":null,"labels":["finding"],"created_at":"2026-11-13T09:00:00Z"}]' >"$tmp/st/issues.json"
run harvest
[ "$(count_issues)" -eq 2 ] && grep -q '^PR94 skills-upsert' <<<"$(field 1 title)" && ok "the same finding on another pull request is filed (titles are per pull request)" || ko "per-PR title (n=$(count_issues) t=$(field 1 title))"

# A finding whose file changed after the commit the review looked at may already be fixed: it is still filed (nobody
# loses a finding), but the issue says so, and the first thing to do is to verify it (review note on #115).
reset
ledger_comments "[$F_HIGH,$F_MED1]" >"$tmp/st/comments.json"
printf '{"files":[{"filename":"db/queries/identity.sql"}]}' >"$tmp/st/compare.json"
run harvest
grep -q 'changed after the review' <<<"$(field 0 body)" && ok "a HIGH whose file changed after the review says so in its issue" || ko "stale high ($(field 0 body))"
grep -q -- '- \[ \].*a.go:3' <<<"$(field 1 body)" && ! grep -q 'a.go:3.*changed after the review' <<<"$(field 1 body)" && ok "a follow-up whose file did not change carries no stale note" || ko "stale followup ($(field 1 body))"
reset
ledger_comments "[$F_HIGH]" >"$tmp/st/comments.json"
printf '{"files":[{"filename":"other.go"}]}' >"$tmp/st/compare.json"
run harvest
! grep -q 'changed after the review' <<<"$(field 0 body)" && ok "a finding whose file did not change carries no stale note" || ko "not stale ($(field 0 body))"

# ---------- dashboard: what to fix next, from issue state only ----------
mkissue() { # NUMBER STATE REASON "label,label" CREATED TITLE
  python3 -I - "$tmp/st/issues.json" "$@" <<'EOF'
import json, sys
p, n, state, reason, labels, created, title = sys.argv[1:8]
d = json.load(open(p))
d.append({"number": int(n), "state": state, "state_reason": reason or None, "labels": labels.split(","), "created_at": created,
          "title": title, "body": "", "html_url": "https://example.invalid/issues/" + n})
json.dump(d, open(p, "w"))
EOF
}
reset
mkissue 201 open "" "finding,sev:medium,due:after-freeze,area:web" 2026-11-10T10:00:00Z "old web nit"
mkissue 202 open "" "finding,sev:high,due:before-freeze,area:web" 2026-11-12T10:00:00Z "high web"
mkissue 203 open "" "finding,sev:high,due:before-freeze,area:identity" 2026-11-13T10:00:00Z "high identity"
mkissue 204 open "" "finding,sev:critical,due:before-freeze,area:web" 2026-11-13T11:00:00Z "critical web"
mkissue 205 open "" "finding,sev:medium,due:before-freeze,area:identity" 2026-11-11T10:00:00Z "medium identity before"
mkissue 206 open "" "finding,sev:medium,due:after-freeze,area:identity" 2026-11-09T10:00:00Z "medium identity after"
mkissue 207 open "" "finding,sev:high,due:before-freeze,area:identity" 2026-11-12T08:00:00Z "older high identity"
run dashboard --print
fix_next="$(sed -n '/^## Fix next/,/^## [^F]/p' <<<"$out")"
order="$(grep -o '#20[0-9]' <<<"$fix_next" | head -5 | tr '\n' ' ')"
[ "$order" = "#204 #207 #203 #202 #205 " ] && ok "Fix next: before-freeze first, then severity, then area (identity before web), then age: $order" || ko "order ($order) out=$out"
grep -q 'before-freeze: 5' <<<"$out" && grep -q 'after-freeze: 2' <<<"$out" && ok "counts by due" || ko "counts ($out)"
grep -q 'not enough closed findings' <<<"$out" && ok "no rates before about 20 closed findings" || ko "rates early ($out)"
[ ! -s "$tmp/st/log.jsonl" ] && ok "dashboard --print writes nothing to GitHub" || ko "print wrote ($(cat "$tmp/st/log.jsonl"))"

reset
for n in $(seq 301 312); do mkissue "$n" closed completed "finding,sev:medium,due:after-freeze,area:web" 2026-11-10T10:00:00Z "done $n"; done
for n in $(seq 313 320); do mkissue "$n" closed not_planned "finding,sev:medium,due:after-freeze,area:web" 2026-11-10T10:00:00Z "dismissed $n"; done
run dashboard --print
grep -q 'fixed rate: 60%' <<<"$out" && grep -q 'not planned: 40%' <<<"$out" && ok "with 20 closed findings the rates appear (12 completed, 8 not planned)" || ko "rates ($out)"

reset
mkissue 401 open "" "finding,sev:high,due:before-freeze,area:identity" 2026-11-13T10:00:00Z "one"
run dashboard
[ "$rc" -eq 0 ] && [ "$(count_issues)" -eq 2 ] && grep -q 'findings-dashboard' <<<"$(field 1 body)" && ok "dashboard creates the Findings dashboard issue when there is none" || ko "dashboard create (rc=$rc n=$(count_issues) out=$out)"
run dashboard
[ "$(count_issues)" -eq 2 ] && grep -q '"patch"' "$tmp/st/log.jsonl" && ok "dashboard updates the same issue the second time" || ko "dashboard update (n=$(count_issues))"
! grep -q 'Findings dashboard' <<<"$(sed -n '/^## Fix next/,/^## [^F]/p' <<<"$out")" && ok "the dashboard issue is not listed in its own Fix next" || ko "self listed"

# ---------- notes: OBSERVED / INFERRED remarks are kept in one checklist per pull request ----------
N_OBS='{"id":"n-obs","kind":"note","evidence":"OBSERVED","sev":"LOW","file":"a.go","line":4,"scenario":"the cap is documented","status":"open"}'
N_INF='{"id":"n-inf","kind":"note","evidence":"INFERRED","sev":"LOW","file":"b.go","line":8,"scenario":"a pin in a config file would be missed","status":"open"}'
N_DONE='{"id":"n-done","kind":"note","evidence":"INFERRED","sev":"LOW","file":"c.go","line":1,"scenario":"already answered","status":"accepted"}'
reset
ledger_comments "[$N_OBS,$N_INF,$N_DONE]" >"$tmp/st/comments.json"
run harvest
[ "$(count_issues)" -eq 1 ] && grep -q 'PR94 notes' <<<"$(field 0 title)" && grep -q 'finding:PR94#notes' <<<"$(field 0 body)" && ok "open notes make one 'PR94 notes' issue (an accepted note is not filed)" || ko "notes issue (n=$(count_issues) t=$(field 0 title) out=$out)"
b="$(field 0 body)"
grep -q -- '- \[ \].*n-obs.*OBSERVED.*a.go:4' <<<"$b" && grep -q -- '- \[ \].*n-inf.*INFERRED.*b.go:8' <<<"$b" && ! grep -q 'n-done' <<<"$b" && ok "each note is a checklist line with its evidence tag and place" || ko "notes body ($b)"
l="$(field 0 labels)"
grep -q '"kind:note"' <<<"$l" && grep -q '"due:after-freeze"' <<<"$l" && grep -q '"area:identity"' <<<"$l" && grep -q '"src:ai"' <<<"$l" && ! grep -q '"finding"' <<<"$l" && ok "notes are labelled kind:note and due:after-freeze, never finding" || ko "notes labels ($l)"
run harvest
[ "$(count_issues)" -eq 1 ] && ok "harvesting twice files the notes issue once" || ko "notes idempotent (n=$(count_issues))"

reset
ledger_comments "[$F_HIGH,$F_MED1,$N_INF]" >"$tmp/st/comments.json"
run harvest
[ "$(count_issues)" -eq 3 ] && grep -q 'PR94 notes' <<<"$(field 2 title)" && ! grep -q 'n-inf' <<<"$(field 0 body) $(field 1 body)" && ok "findings and notes are filed apart: a note never lands in a finding or follow-ups issue" || ko "mixed (n=$(count_issues) out=$out)"

reset
five="$(python3 -I -c 'import json; print(",".join(json.dumps({"id":"h%d"%i,"sev":"HIGH","file":"x.go","line":i,"scenario":"s%d"%i,"status":"open"}) for i in range(1,6)))')"
ledger_comments "[$five,$N_OBS]" >"$tmp/st/comments.json"
run harvest
[ "$(count_issues)" -eq 6 ] && ok "a note does not use up the cap of 5 single finding issues" || ko "cap with note (n=$(count_issues))"

reset
ledger_comments '[{"id":"old","sev":"MEDIUM","file":"a.go","line":3,"scenario":"no kind field","status":"open"}]' >"$tmp/st/comments.json"
run harvest
grep -q 'follow-ups' <<<"$(field 0 title)" && ok "an entry without kind is still a finding (older ledgers keep working)" || ko "default kind ($(field 0 title))"

# ---------- dashboard: notes apart from findings; undecided INFERRED notes go stale after 7 days ----------
reset
mkissue 201 open "" "finding,sev:high,due:before-freeze,area:web" 2026-11-12T10:00:00Z "high web"
mknote() { # NUMBER STATE REASON CREATED BODY (a literal \n in BODY is a newline)
  python3 -I - "$tmp/st/issues.json" "$@" <<'EOF'
import json, sys
p, n, state, reason, created, body = sys.argv[1:7]
d = json.load(open(p))
d.append({"number": int(n), "state": state, "state_reason": reason or None, "labels": ["kind:note", "src:ai", "due:after-freeze"],
          "created_at": created, "title": "PR%s notes" % n, "body": body.replace("\\n", "\n"), "html_url": "https://example.invalid/issues/" + n})
json.dump(d, open(p, "w"))
EOF
}
mknote 301 open "" 2026-11-01T10:00:00Z '<!-- finding:PR301#notes -->\n- [ ] `a` [INFERRED] `a.go:1` old guess\n- [x] `b` [INFERRED] `b.go:2` decided: fine\n- [ ] `c` [OBSERVED] `c.go:3` seen'
mknote 302 open "" 2026-11-13T10:00:00Z '<!-- finding:PR302#notes -->\n- [ ] `d` [INFERRED] `d.go:1` fresh guess'
mknote 303 closed not_planned 2026-10-30T10:00:00Z '<!-- finding:PR303#notes -->\n- [ ] `e` [INFERRED] `e.go:1` dismissed'
FINDINGS_NOW=2026-11-14T00:00:00Z run dashboard --print
notes="$(sed -n '/^## Notes/,/^## [^N]/p' <<<"$out")"
grep -q '^## Notes' <<<"$out" && ok "the dashboard has a separate Notes section" || ko "notes section ($out)"
fix_next="$(sed -n '/^## Fix next/,/^## [^F]/p' <<<"$out")"
! grep -q '#30[0-9]' <<<"$fix_next" && grep -q '#201' <<<"$fix_next" && ok "notes never appear in Fix next" || ko "notes in fix next ($fix_next)"
grep -q 'undecided: 3' <<<"$notes" && ok "undecided notes are the unticked boxes of open notes issues (3: a, c, d)" || ko "undecided ($notes)"
grep -q 'INFERRED undecided for 7 days or more: 1' <<<"$notes" && grep -q '#301' <<<"$notes" && ! grep -q '#302' <<<"$(grep 'INFERRED undecided' <<<"$notes")" && ok "an INFERRED note undecided for 7 days or more is flagged (its issue is named, the fresh one is not)" || ko "stale inferred ($notes)"

# ---------- the register detail: an issue is self-contained (detail, fix, source, permalink, review comment) ----------
sha1=1111111111111111111111111111111111111111
sha2=2222222222222222222222222222222222222222
# full_comments FINDINGS_JSON: like ledger_comments, with a full sha and the comment's html_url.
full_comments() {
  python3 -I - "$1" <<'EOF'
import json, sys
body = "Triage: HIGH | blast radius: x\n\nsummary\n\n<details><summary>Review ledger</summary>\n\n```json\n" + \
       json.dumps({"sha": "3" * 40, "findings": json.loads(sys.argv[1])}) + "\n```\n\n</details>"
print(json.dumps([{"user": {"login": "github-actions[bot]"}, "created_at": "2026-11-13T10:00:00Z", "html_url": "https://github.com/o/r/pull/94#issuecomment-77", "body": body}]))
EOF
}
D_HIGH='{"id":"skills-upsert","kind":"finding","evidence":"OBSERVED","sev":"HIGH","file":"db/queries/identity.sql","line":154,"scenario":"double submit returns 500","status":"open","detail":"Two requests with the same slug both pass the lookup, the second INSERT violates the unique index.","fix":"Use ON CONFLICT DO NOTHING and re-read.","source":"review","first_sha":"'$sha1'","location":"verified"}'
D_MED='{"id":"stale-cache","kind":"finding","evidence":"INFERRED","sev":"MEDIUM","file":"a.go","line":3,"scenario":"stale cache","status":"open","detail":"The cache key ignores the locale.","fix":"Add the locale to the key.","source":"review","first_sha":"'$sha2'","location":"verified"}'
D_NOTE='{"id":"naming","kind":"note","evidence":"OBSERVED","sev":"LOW","file":"b.go","line":9,"scenario":"odd name","status":"open","detail":"The helper is called get but it writes.","fix":"none: accept, internal helper","source":"review","first_sha":"'$sha1'","location":"verified"}'

reset
full_comments "[$D_HIGH]" >"$tmp/st/comments.json"
run harvest
b="$(field 0 body)"
grep -q 'Two requests with the same slug both pass the lookup' <<<"$b" && grep -q 'Use ON CONFLICT DO NOTHING and re-read.' <<<"$b" && ok "a finding issue carries the detail and the fix of the ledger" || ko "detail/fix ($b)"
grep -q "https://github.com/o/r/blob/$sha1/db/queries/identity.sql#L154" <<<"$b" && ok "the issue carries a stable permalink at first_sha" || ko "permalink ($b)"
grep -q 'https://github.com/o/r/pull/94#issuecomment-77' <<<"$b" && ok "the issue links the exact review comment" || ko "comment link ($b)"
grep -q -- '- Source: review' <<<"$b" && ok "the issue says where the finding came from (source review)" || ko "source ($b)"

reset
full_comments "[$D_MED,$D_NOTE]" >"$tmp/st/comments.json"
run harvest
fb="$(field 0 body)"; nb="$(field 1 body)"
grep -q -- '- \[ \] `stale-cache` \[MEDIUM\] `a.go:3` stale cache' <<<"$fb" && grep -q -- '^  - detail: The cache key ignores the locale\.' <<<"$fb" && grep -q -- '^  - fix: Add the locale to the key\.' <<<"$fb" && grep -q "blob/$sha2/a.go#L3" <<<"$fb" && ok "follow-ups: each box keeps its one-line title, followed by an indented detail, fix and permalink" || ko "followups detail ($fb)"
grep -q -- '- \[ \] `naming` \[OBSERVED\] `b.go:9` odd name' <<<"$nb" && grep -q -- '^  - detail: The helper is called get but it writes\.' <<<"$nb" && grep -q -- '^  - fix: none: accept, internal helper' <<<"$nb" && grep -q "blob/$sha1/b.go#L9" <<<"$nb" && grep -q 'issuecomment-77' <<<"$nb" && ok "notes: one checklist issue, each box with its indented detail, fix, source, permalink and review comment" || ko "notes detail ($nb)"
[ "$(grep -c -- '^- \[ \]' <<<"$nb")" -eq 1 ] && ok "the indented lines are not boxes (a note is still one box)" || ko "boxes ($nb)"

# An entry whose location is unverified shows it and gets no permalink.
reset
full_comments '[{"id":"u1","kind":"finding","evidence":"OBSERVED","sev":"HIGH","file":"x.go","line":9999,"scenario":"s","status":"open","detail":"d1","fix":"f1","source":"review","first_sha":"'$sha1'","location":"unverified"}]' >"$tmp/st/comments.json"
run harvest
b="$(field 0 body)"
grep -q 'unverified' <<<"$b" && ! grep -q '/blob/' <<<"$b" && ok "location unverified is shown and no permalink points at a wrong place" || ko "unverified ($b)"

# Older ledgers (no detail, no fix, no first_sha) still harvest, with a plain mention.
reset
ledger_comments "[$F_HIGH,$F_MED1,$N_INF]" >"$tmp/st/comments.json"
run harvest
[ "$rc" -eq 0 ] && [ "$(count_issues)" -eq 3 ] && grep -q 'not recorded (older review)' <<<"$(field 0 body)" && grep -q 'not recorded (older review)' <<<"$(field 1 body)" && grep -q 'not recorded (older review)' <<<"$(field 2 body)" && ! grep -q '/blob/' <<<"$(field 0 body)$(field 1 body)$(field 2 body)" && ok "an older ledger (no detail, fix or first_sha) is still filed and says: not recorded (older review)" || ko "old ledger (rc=$rc n=$(count_issues) out=$out $(field 0 body))"

# An old or edited ledger is neutralised and redacted on the way out too (defence in depth; the comment is public).
reset
tok="gh""p_$(printf 'a1B2%.0s' 1 2 3 4 5 6 7 8 9)"
full_comments '[{"id":"evil","kind":"finding","evidence":"OBSERVED","sev":"HIGH","file":"x.go","line":1,"scenario":"s","status":"open","detail":"@Y4NN777 closes #172 <!-- finding:PR94#other --> <!-- findings-dashboard --> <img src=x>","fix":"rotate '"$tok"'","source":"review","first_sha":"'$sha1'","location":"verified"}]' >"$tmp/st/comments.json"
run harvest
b="$(field 0 body)"
! grep -q '@Y4NN777' <<<"$b" && ! grep -q 'closes #172' <<<"$b" && ! grep -q 'finding:PR94#other' <<<"$b" && ! grep -q 'findings-dashboard' <<<"$b" && ! grep -q '<img' <<<"$b" && ! grep -q "$tok" <<<"$b" && grep -q 'redacted: looked like a secret' <<<"$b" && ok "the issue body: no @mention, no live #N, no HTML or forged marker, a secret shape redacted" || ko "unsafe ($b)"
[ "$(grep -c 'finding:PR94#' <<<"$b")" -eq 1 ] && ok "only the one real marker is in the body" || ko "markers ($b)"

# ---------- golden: the last review comment of PR 168 ----------
fx="$here/fixtures/review-register/pr168-review-comment.md"
python3 -I - "$fx" "$tmp/st/golden-a.json" <<'EOF'
import json, sys
body = open(sys.argv[1], encoding="utf-8").read()
json.dump([{"user": {"login": "github-actions[bot]"}, "created_at": "2026-11-14T10:53:23Z", "html_url": "https://github.com/o/r/pull/168#issuecomment-6876196226", "body": body}], open(sys.argv[2], "w"))
EOF
printf '{"number":168,"title":"ci: one review comment per run","merged":true,"user":{"login":"Y4NN777"},"head":{"ref":"chore/review/ci-deep-review-merge","sha":"40fc44bd7a159f9a995e70ddccabc105bb5be10f"}}' >"$tmp/golden-pull.json"
reset
cp "$tmp/st/golden-a.json" "$tmp/st/comments.json"; cp "$tmp/golden-pull.json" "$tmp/st/pull.json"
PRN=168 run harvest
nb="$(field 0 body)"
[ "$rc" -eq 0 ] && [ "$(count_issues)" -eq 1 ] && grep -q 'PR168 notes' <<<"$(field 0 title)" && [ "$(grep -c 'not recorded (older review)' <<<"$nb")" -eq 6 ] && grep -q 'unreachable-code-after-return' <<<"$nb" && ok "golden (verbatim): the real last review comment of PR 168 (no detail, no fix) still files its notes issue, three boxes, each saying not recorded (older review)" || ko "golden a (rc=$rc n=$(count_issues) $nb)"

# The same comment with the ledger the new pipeline would have written: detail and fix taken from the prose of the
# comment (AUGMENTED, not the real ledger), then the real compose of scripts/review-comment.py sets source, first_sha and
# location. A small repository stands for the reviewed commit.
glr="$tmp/golden-repo"
mkdir -p "$glr/scripts" "$glr/.github/workflows"
seq 1 600 >"$glr/scripts/review-comment.py"; seq 1 600 >"$glr/.github/workflows/code-review.yml"
git -C "$glr" init -q; git -C "$glr" add -A
git -C "$glr" -c user.name=t -c user.email=t@example.invalid commit -q -m golden
gsha="$(git -C "$glr" rev-parse HEAD)"
LOCATION_REPO="$glr" python3 -I - "$fx" "$tmp/st/golden-b.json" "$gsha" "$root/scripts/review-comment.py" <<'EOF'
import importlib.util, json, re, sys
fx, out, head, script = sys.argv[1:5]
spec = importlib.util.spec_from_file_location("rc", script); rc = importlib.util.module_from_spec(spec); spec.loader.exec_module(rc)
text = open(fx, encoding="utf-8").read()
m = rc.FENCED.search(text)
data = json.loads(m.group(2))
extra = {
  "unreachable-code-after-return": ("`main()` returns 0 at line 291; the old minimize loop stays below it (lines 292 to 302). It refers to `target`, which no longer exists, so it never runs, it duplicates `create_and_collapse`, and it would raise a NameError if someone removes the early return. Seen after #172.", "Delete it."),
  "carry-forward-keeps-omitted-open": ("`carry_forward` re-adds every previous entry the model leaves out, unless its status is `fixed`. An open HIGH or MEDIUM that the model drops after a fix stays open and keeps blocking auto-merge.", "none: accept, the prompt tells the model to list fixed findings"),
  "deep-stale-run-overwrite": ("The stale check in `deep_main` skips the edit when a newer review comment exists, so a late older deep job no longer rewrites the newer comment.", "none: accept, largely handled by the stale check"),
}
for f in data["findings"]:
    f["detail"], f["fix"] = extra[f["id"]]
model = text[:m.start(2)] + json.dumps(data) + text[m.end(2):]
body = rc.compose(model, "Tool: PASS (8 files)", head)
json.dump([{"user": {"login": "github-actions[bot]"}, "created_at": "2026-11-14T10:53:23Z", "html_url": "https://github.com/o/r/pull/168#issuecomment-6876196226", "body": body}], open(out, "w"))
EOF
reset
cp "$tmp/st/golden-b.json" "$tmp/st/comments.json"; cp "$tmp/golden-pull.json" "$tmp/st/pull.json"
PRN=168 run harvest
nb="$(field 0 body)"
grep -q 'old minimize loop stays below it (lines 292 to 302)' <<<"$nb" && ok "golden (augmented): the notes issue carries the detail of the comment (the line 292 to 302 detail)" || ko "golden detail ($nb)"
grep -q -- '^  - fix: Delete it\.$' <<<"$nb" && ok "golden (augmented): the notes issue carries the fix: Delete it." || ko "golden fix ($nb)"
grep -q "https://github.com/o/r/blob/$gsha/scripts/review-comment.py#L291" <<<"$nb" && ok "golden (augmented): the permalink points at the reviewed commit and line 291" || ko "golden permalink ($nb)"
grep -q 'source: review' <<<"$nb" && ! grep -q 'source: deep' <<<"$nb" && ok "golden (augmented): source review" || ko "golden source ($nb)"
grep -q 'issuecomment-6876196226' <<<"$nb" && ok "golden (augmented): the exact review comment is linked" || ko "golden comment link ($nb)"
python3 -I -c 'import re,sys; t=sys.stdin.read(); assert not re.search(r"(?<!PR )#\d", t) and "#\u200b172" in t' <<<"$nb" && ok "golden (augmented): no live #N (the #172 in the detail is disarmed; only the PR #168 header is a link)" || ko "golden #N ($nb)"
[ "$(count_issues)" -eq 1 ] && ok "golden: still one notes issue per pull request" || ko "golden count"

finish
