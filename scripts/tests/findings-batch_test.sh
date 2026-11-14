#!/usr/bin/env bash
# Tests for `scripts/findings.py batch` with a stateful fake `gh`: the open findings and the undecided notes are ranked
# (severity, area, age; stale INFERRED notes above the other notes) into ONE open "Review batch YYYY-MM-DD" checklist
# issue, updated in place (idempotent by marker, ticked boxes kept), closed as completed when every source is closed or
# decided. It never writes to a source issue, never mentions anyone, and fails CLOSED when gh cannot be read.
# Usage: scripts/tests/findings-batch_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
tool="$root/scripts/findings.py"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin" "$tmp/st"

# Fake gh (python): `gh api [-X METHOD] [-f k=v]... ENDPOINT [--paginate]`. State in $ST/issues.json. GET issues filters
# by the `labels=` query; POST appends an issue; PATCH updates title, body, state and state_reason. Every write is
# logged to $ST/log.jsonl. FAKE_GH_FAIL fails every call; FAKE_WRITE_FAIL fails only the writes.
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
def load():
    return json.load(open(os.path.join(st, "issues.json")))
def save(data):
    json.dump(data, open(os.path.join(st, "issues.json"), "w"))
def log(entry):
    open(os.path.join(st, "log.jsonl"), "a").write(json.dumps(entry) + "\n")
path = endpoint.split("?")[0]
if method != "GET" and os.environ.get("FAKE_WRITE_FAIL"):
    sys.stderr.write("gh: write refused\n"); sys.exit(1)
if method == "GET" and path.endswith("/pulls"):
    assert "base=develop" in endpoint, endpoint
    p = os.path.join(st, "pulls.json")
    print(json.dumps(json.load(open(p)) if os.path.exists(p) else []))
elif method == "GET" and path.endswith("/issues"):
    query = endpoint.split("?", 1)[1] if "?" in endpoint else ""
    want = next((p[7:] for p in query.split("&") if p.startswith("labels=")), "")
    print(json.dumps([x for x in load() if not want or want in [l if isinstance(l, str) else l["name"] for l in x["labels"]]]))
elif method == "POST" and path.endswith("/issues"):
    issues = load()
    issue = {"number": 900 + len(issues), "title": fields["title"][0], "body": fields["body"][0], "state": "open",
             "state_reason": None, "labels": fields.get("labels[]", []), "assignees": fields.get("assignees[]", []),
             "created_at": "2026-11-14T00:00:00Z"}
    issues.append(issue); save(issues)
    log({"op": "create", "number": issue["number"], "title": issue["title"], "labels": issue["labels"], "assignees": issue["assignees"]})
    print(json.dumps(issue))
elif method == "PATCH" and "/issues/" in path:
    n = int(path.rsplit("/", 1)[1])
    issues = load()
    for x in issues:
        if x["number"] == n:
            for k in ("title", "body", "state", "state_reason"):
                if k in fields: x[k] = fields[k][0]
    save(issues); log({"op": "patch", "number": n, "fields": sorted(fields)}); print("{}")
else:
    sys.stderr.write("unexpected %s %s\n" % (method, endpoint)); sys.exit(2)
EOF
chmod +x "$tmp/bin/gh"

reset() {
  printf '[]' >"$tmp/st/issues.json"
  printf '[]' >"$tmp/st/pulls.json"
  : >"$tmp/st/log.jsonl"
}

# run ARGS...: runs the tool with a pinned clock; sets rc and out.
run() {
  rc=0
  out="$(PATH="$tmp/bin:$PATH" ST="$tmp/st" REPO=o/r FINDINGS_NOW=2026-11-14T00:00:00Z python3 -I "$tool" "$@" 2>&1)" || rc=$?
}

mkissue() { # NUMBER STATE "label,label" CREATED TITLE
  python3 -I - "$tmp/st/issues.json" "$@" <<'EOF'
import json, sys
p, n, state, labels, created, title = sys.argv[1:7]
d = json.load(open(p))
d.append({"number": int(n), "state": state, "state_reason": "completed" if state == "closed" else None,
          "labels": labels.split(","), "created_at": created, "title": title, "body": "<!-- finding:PR%s#x -->" % n})
json.dump(d, open(p, "w"))
EOF
}
mknote() { # NUMBER STATE CREATED BODY (a literal \n in BODY is a newline)
  python3 -I - "$tmp/st/issues.json" "$@" <<'EOF'
import json, sys
p, n, state, created, body = sys.argv[1:6]
d = json.load(open(p))
d.append({"number": int(n), "state": state, "state_reason": "not_planned" if state == "closed" else None,
          "labels": ["kind:note", "src:ai", "area:web", "due:after-freeze"], "created_at": created, "title": "PR%s notes" % n,
          "body": body.replace("\\n", "\n")})
json.dump(d, open(p, "w"))
EOF
}
# set_issue NUMBER KEY VALUE: change one field of a fake issue (a human closing or editing it).
set_issue() {
  python3 -I - "$tmp/st/issues.json" "$@" <<'EOF'
import json, sys
p, n, k, v = sys.argv[1:5]
d = json.load(open(p))
for x in d:
    if x["number"] == int(n): x[k] = v.replace("\\n", "\n")
json.dump(d, open(p, "w"))
EOF
}
get() { # NUMBER FIELD
  python3 -I -c 'import json,sys; d=([x for x in json.load(open(sys.argv[1])) if x["number"]==int(sys.argv[2])] or [{}])[0]; v=d.get(sys.argv[3], ""); print(v if isinstance(v,str) else json.dumps(v))' "$tmp/st/issues.json" "$1" "$2"
}
batches() { # the numbers of the batch issues (label review-batch)
  python3 -I -c 'import json,sys; print(" ".join(str(x["number"]) for x in json.load(open(sys.argv[1])) if "review-batch" in x["labels"]))' "$tmp/st/issues.json"
}
writes() { wc -l <"$tmp/st/log.jsonl" | tr -d ' '; }

seed() {
  mkissue 201 open "finding,sev:medium,due:after-freeze,area:web" 2026-11-10T10:00:00Z "PR90 old-web: an old web nit"
  mkissue 202 open "finding,sev:high,due:before-freeze,area:web" 2026-11-12T10:00:00Z "PR91 high-web: web high"
  mkissue 203 open "finding,sev:high,due:before-freeze,area:identity" 2026-11-13T10:00:00Z "PR92 high-id: identity high"
  mkissue 204 open "finding,sev:critical,due:before-freeze,area:web" 2026-11-13T11:00:00Z "PR93 crit: critical web"
  mkissue 205 open "finding,sev:low,due:after-freeze,area:identity" 2026-11-01T10:00:00Z "PR94 low: low identity"
  mkissue 206 open "finding,sev:medium,due:after-freeze,area:identity" 2026-11-11T10:00:00Z "PR95 med-id: medium identity"
  mkissue 207 closed "finding,sev:critical,due:before-freeze,area:identity" 2026-11-01T10:00:00Z "PR96 done: closed critical"
  mknote 301 open 2026-11-01T10:00:00Z '<!-- finding:PR301#notes -->\n- [ ] `a` [INFERRED] `a.go:1` old guess\n- [x] `b` [INFERRED] `b.go:2` decided\n- [ ] `c` [OBSERVED] `c.go:3` seen long ago'
  mknote 302 open 2026-11-13T10:00:00Z '<!-- finding:PR302#notes -->\n- [ ] `d` [INFERRED] `d.go:1` fresh guess'
  mknote 303 closed 2026-10-30T10:00:00Z '<!-- finding:PR303#notes -->\n- [ ] `e` [INFERRED] `e.go:1` dismissed'
}

# ---------- ranking ----------
reset
seed
run batch --print
order="$(grep -o 'batch-item:#[0-9/a-z]*' <<<"$out" | sed 's/batch-item://' | tr '\n' ' ' || true)"
[ "$rc" -eq 0 ] && [ "$order" = "#204 #203 #202 #206 #201 #205 #301/a #301/c #302/d " ] \
  && ok "ranked by severity, then area (identity before web), then age; stale INFERRED notes above the other notes: $order" \
  || ko "ranking (rc=$rc order=$order out=$out)"
! grep -q '#207\|#301/b\|#303' <<<"$out" && ok "closed findings, ticked notes and closed notes issues are not in the batch" || ko "done items listed ($out)"
grep -q -- '- \[ \] #204 \[CRITICAL\].*critical web' <<<"$out" && grep -q -- '- \[ \] #301 \[NOTE INFERRED' <<<"$out" \
  && ok "each line links its source issue, with the severity and a one-line scenario" || ko "line format ($out)"
[ ! -s "$tmp/st/log.jsonl" ] && ok "batch --print writes nothing to GitHub" || ko "print wrote ($(cat "$tmp/st/log.jsonl"))"

# ---------- create, then idempotent update ----------
reset
seed
run batch
b="$(batches)"
[ "$rc" -eq 0 ] && [ "$b" = "910" ] && [ "$(get 910 title)" = "Review batch 2026-11-14" ] && grep -q '<!-- review-batch -->' <<<"$(get 910 body)" \
  && ok "batch creates one 'Review batch 2026-11-14' issue with its hidden marker" || ko "create (rc=$rc batches=$b out=$out)"
l="$(get 910 labels)"
grep -q '"review-batch"' <<<"$l" && ! grep -q '"finding"\|"kind:note"' <<<"$l" && [ "$(get 910 assignees)" = "[]" ] \
  && ok "the batch issue is labelled review-batch only (out of the findings and notes counts) and assigned to nobody" || ko "labels ($l)"
before="$(writes)"
run batch
[ "$rc" -eq 0 ] && [ "$(batches)" = "910" ] && [ "$(writes)" = "$before" ] \
  && ok "a second run with nothing changed creates no duplicate and writes nothing" || ko "idempotent (rc=$rc batches=$(batches) writes=$(writes) out=$out)"

# ---------- ticked boxes are kept; a new HIGH joins the open batch ----------
set_issue 910 body "$(get 910 body | sed 's/^- \[ \] \(#206 .*\)$/- [x] \1/')"
mkissue 208 open "finding,sev:high,due:before-freeze,area:db" 2026-11-13T12:00:00Z "PR97 new-high: new db high"
run batch
body="$(get 910 body)"
[ "$rc" -eq 0 ] && grep -q -- '- \[x\] #206 ' <<<"$body" && ok "a box a human ticked stays ticked on update (its source is still open)" || ko "tick kept ($body)"
grep -q -- '- \[ \] #208 \[HIGH\]' <<<"$body" && [ "$(batches)" = "910" ] && ok "a new HIGH joins the open batch, still one batch issue" || ko "new high ($body)"
set_issue 201 state closed
run batch
body="$(get 910 body)"
grep -q -- '- \[x\] #201 .*(done)' <<<"$body" && [ "$(get 910 state)" = "open" ] && ok "a closed source is shown ticked and done; the batch stays open while others are open" || ko "done line ($body)"

# ---------- no write to a source issue ----------
python3 -I - "$tmp/st/log.jsonl" <<'EOF' && ok "every write targets the batch issue only: no source finding or notes issue is edited or closed" || ko "source write ($(cat "$tmp/st/log.jsonl"))"
import json, sys
bad = [e for e in map(json.loads, open(sys.argv[1])) if e.get("number") != 910]
sys.exit(1 if bad else 0)
EOF

# ---------- close when everything is closed or decided; the next run starts a new batch ----------
for n in 202 203 204 205 206 208; do set_issue "$n" state closed; done
set_issue 301 body '<!-- finding:PR301#notes -->\n- [x] `a` [INFERRED] `a.go:1` old guess\n- [x] `b` [INFERRED] `b.go:2` decided\n- [x] `c` [OBSERVED] `c.go:3` seen long ago'
set_issue 302 state closed
run batch
[ "$rc" -eq 0 ] && [ "$(get 910 state)" = "closed" ] && [ "$(get 910 state_reason)" = "completed" ] \
  && ok "the batch closes as completed when every source is closed or decided" || ko "close (rc=$rc state=$(get 910 state) reason=$(get 910 state_reason) out=$out)"
[ "$(batches)" = "910" ] && ok "the closing run does not open the next batch itself" || ko "closing run created ($(batches))"
mkissue 209 open "finding,sev:medium,due:after-freeze,area:web" 2026-11-13T13:00:00Z "PR98 later: a later medium"
run batch
read -r -a bs <<<"$(batches)"
[ "${#bs[@]}" -eq 2 ] && [ "${bs[0]}" = "910" ] && [ "$(get "${bs[1]}" state)" = "open" ] && grep -q 'batch-item:#209' <<<"$(get "${bs[1]}" body)" && ok "the next run starts a new batch with what is open" || ko "next batch ($(batches))"

# ---------- a merged fix pull request closes the batch ----------
reset
seed
run batch
grep -q 'ONE fix pull' <<<"$(get 910 body)" && grep -q 'fix/review/batch-2026-11-14' <<<"$(get 910 body)" \
  && ok "the batch body says it is the scope of one fix pull request and names the branch" || ko "one-PR text ($(get 910 body))"
printf '[{"number":150,"merged_at":null,"body":"Fixes the batch #910"},{"number":151,"merged_at":"2026-11-14T03:00:00Z","body":"unrelated, see #9100 and #91"}]' >"$tmp/st/pulls.json"
run batch
[ "$rc" -eq 0 ] && [ "$(get 910 state)" = "open" ] && ok "an unmerged pull request naming the batch, or a merged unrelated one, does not close it" || ko "unrelated pr (state=$(get 910 state) out=$out)"
printf '[{"number":152,"merged_at":"2026-11-14T04:00:00Z","body":"Review batch #910: fixes the top items"}]' >"$tmp/st/pulls.json"
before="$(writes)"
run batch --print
[ "$rc" -eq 0 ] && [ "$(writes)" = "$before" ] && grep -q 'would close #910' <<<"$out" && [ "$(get 910 state)" = "open" ] \
  && ok "a promotion-pull-request run (--print) only reports what it would do and makes no gh write" || ko "print with merged pr (writes=$(writes) out=$out)"
run batch
[ "$rc" -eq 0 ] && [ "$(get 910 state)" = "closed" ] && [ "$(get 910 state_reason)" = "completed" ] && grep -q 'pull request #152 is merged' <<<"$(get 910 body)" \
  && ok "a merged fix pull request whose body names the batch closes it as completed, sources still open" || ko "merged pr close (state=$(get 910 state) out=$out)"
[ "$(get 204 state)" = "open" ] && ok "closing the batch leaves its source issues open (humans classify them)" || ko "source closed"
run batch
read -r -a bs <<<"$(batches)"
[ "${#bs[@]}" -eq 2 ] && [ "$(get "${bs[1]}" state)" = "open" ] && grep -q 'batch-item:#204' <<<"$(get "${bs[1]}" body)" \
  && ok "the next run opens a new batch from what is still open" || ko "new batch after merge ($(batches))"

# ---------- nothing open: nothing created ----------
reset
mkissue 207 closed "finding,sev:critical,due:before-freeze,area:identity" 2026-11-01T10:00:00Z "PR96 done: closed critical"
run batch
[ "$rc" -eq 0 ] && [ -z "$(batches)" ] && ok "nothing open: no batch issue is created" || ko "empty (rc=$rc batches=$(batches) out=$out)"

# ---------- cap ----------
reset
for n in $(seq 401 420); do mkissue "$n" open "finding,sev:medium,due:after-freeze,area:web" "2026-11-1${n: -1}T10:00:00Z" "PR$n m$n: medium $n"; done
run batch
body="$(get 920 body)"
[ "$(grep -c 'batch-item:' <<<"$body")" -eq 15 ] && grep -q '5 more open items' <<<"$body" && ok "the batch is capped at 15 lines with a count of the rest" || ko "cap ($body)"

# ---------- no mention ----------
reset
mkissue 501 open "finding,sev:high,due:before-freeze,area:web" 2026-11-13T10:00:00Z "PR99 m: @Y4NN777 should look"
mknote 502 open 2026-11-01T10:00:00Z '<!-- finding:PR502#notes -->\n- [ ] `m` [INFERRED] `m.go:1` ask @NICE-DEV226 first'
run batch
body="$(get 902 body) $(get 902 title)"
! grep -qE '@[A-Za-z0-9_-]' <<<"$body" && grep -q 'Y4NN777' <<<"$body" && grep -q 'NICE-DEV226' <<<"$body" \
  && ok "an @name in a source title or note is neutralised: the batch mentions nobody (the name stays readable)" || ko "mention ($body)"

# ---------- fail closed ----------
reset
seed
FAKE_GH_FAIL=1 run batch
[ "$rc" -ne 0 ] && grep -q 'findings batch: FAILED' <<<"$out" && ok "gh unreadable: batch fails closed with a clear message (rc=$rc)" || ko "gh failure (rc=$rc out=$out)"
FAKE_WRITE_FAIL=1 run batch
[ "$rc" -ne 0 ] && grep -q 'findings batch: FAILED' <<<"$out" && ok "a write GitHub refuses: batch fails closed (rc=$rc)" || ko "write failure (rc=$rc out=$out)"
rc=0
out="$(PATH="$tmp/bin:$PATH" ST="$tmp/st" python3 -I "$tool" batch 2>&1)" || rc=$?
[ "$rc" -ne 0 ] && grep -q 'REPO' <<<"$out" && ok "batch without REPO fails closed" || ko "no repo (rc=$rc out=$out)"

finish
