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
    elif "/pulls/" in path: print(json.dumps(load("pull.json")))
    elif path.endswith("/issues"):
        print(json.dumps([x for x in load("issues.json") if "finding" in [l if isinstance(l, str) else l["name"] for l in x["labels"]]]))
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
  printf '{"number":94,"title":"S1 profiles","merged":true,"user":{"login":"NICE-DEV226"},"head":{"ref":"feature/identity/profiles"}}' >"$tmp/st/pull.json"
  : >"$tmp/st/log.jsonl"
}

# run ARGS...: runs the tool; sets rc and out.
run() {
  rc=0
  out="$(PATH="$tmp/bin:$PATH" ST="$tmp/st" REPO=o/r PR_NUMBER=94 PR_AUTHOR=NICE-DEV226 python3 -I "$tool" "$@" 2>&1)" || rc=$?
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
printf '[{"number":100,"title":"skills-upsert: double submit returns 500","body":"edited by a human, marker gone","state":"open","state_reason":null,"labels":["finding"],"created_at":"2026-11-13T09:00:00Z"},{"number":101,"title":"PR94 follow-ups","body":"edited by a human","state":"open","state_reason":null,"labels":["finding"],"created_at":"2026-11-13T09:00:00Z"}]' >"$tmp/st/issues.json"
run harvest
[ "$(count_issues)" -eq 2 ] && ok "an issue whose marker was edited away is recognised by its title: no duplicate" || ko "title dedupe (n=$(count_issues) out=$out)"

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

finish
