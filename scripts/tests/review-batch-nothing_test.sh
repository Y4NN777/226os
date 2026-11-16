#!/usr/bin/env bash
# Tests for `scripts/review-autofix.py nothing` with a stateful fake `gh`: when the batch fixer made no commit (outcome
# "nothing", batch run 43633446506 on #226), the trusted publish job applies its per-item verdicts with the same rules as
# `findings.py settle`: FIXED ELSEWHERE closes or ticks a source only when the named pull request is merged (and touched the
# cited file, after the source was created, and is not the source's own pull request); NOT A DEFECT and DECIDED only
# comment (owner rule H5); FIXED without a commit is OWNER; OWNER comments and stays open (a retired AFTER FREEZE from the fixer counts as OWNER). ONE comment on
# the batch issue holds the verdict table; every item with a verdict is ticked and the batch closes (marker
# batch-handled), so the next batch run moves on. Only the batch issue's own items are applied; model text is neutralised.
# Usage: scripts/tests/review-batch-nothing_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
tool="$root/scripts/review-autofix.py"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin" "$tmp/st"

# Fake gh: issues in $ST/issues.json (each with a "comments" list), pull request N in $ST/prN.json, its files in
# $ST/filesN.json. Writes are logged to $ST/log, one line each: METHOD ENDPOINT FIELDS.
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
    elif a == "-f": k, v = args[i + 1].split("=", 1); fields[k] = v; i += 2
    elif a == "--paginate": i += 1
    else: endpoint = a; i += 1
path = endpoint.split("?")[0]
parts = path.split("/")
# FAKE_GH_FAILS=N: the first N calls of kind FAKE_GH_KIND (write or get) fail with FAKE_GH_STDERR; each failure is counted in $ST/attempts.
kind = "get" if method == "GET" else "write"
if os.environ.get("FAKE_GH_FAILS") and kind == os.environ.get("FAKE_GH_KIND", "write") and method in os.environ.get("FAKE_GH_METHODS", "GET POST PATCH PUT DELETE").split():
    cf = os.path.join(st, "attempts")
    n = int(open(cf).read()) if os.path.exists(cf) else 0
    if n < int(os.environ["FAKE_GH_FAILS"]):
        open(cf, "w").write(str(n + 1))
        sys.stderr.write(os.environ["FAKE_GH_STDERR"] + "\nsecond line with a body {\"message\":\"x\"}\n"); sys.exit(1)
issues = json.load(open(os.path.join(st, "issues.json")))
def save(): json.dump(issues, open(os.path.join(st, "issues.json"), "w"))
def find(n): return next(x for x in issues if x["number"] == int(n))
def answer(name):
    p = os.path.join(st, name)
    if not os.path.exists(p): sys.stderr.write("404\n"); sys.exit(1)
    print(open(p).read())
if method != "GET":
    open(os.path.join(st, "log"), "a").write("%s %s %s\n" % (method, path, json.dumps(fields, sort_keys=True)))
if method == "GET" and parts[-1] == "files":
    answer("files%s.json" % parts[-2])
elif method == "GET" and parts[-2] == "pulls":
    answer("pr%s.json" % parts[-1])
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
EOF
chmod +x "$tmp/bin/gh"

# The untrusted text a model could return: a mention, closing keywords and references in every form, a pipe, forged markers.
evil='@someone fixes #7, see https://github.com/o/r/issues/7 and GH-7 | a pipe <!-- batch-handled --> <!-- settled-by:batch40-run99 -->'

setup() {
  python3 -I - "$tmp/st" "$evil" <<'EOF'
import json, os, sys
st, evil = sys.argv[1], sys.argv[2]
def finding(n):
    return {"number": n, "state": "open", "state_reason": None, "title": "PR90 f%d: t" % n, "created_at": "2026-11-13T00:00:00Z",
            "labels": [{"name": "finding"}], "body": "<!-- finding:PR90#f%d -->\n\n- Where: `src/f%d.go:1`\n" % (n, n),
            "comments": [{"user": {"login": "dev"}, "body": "human note"}]}
notes = ("<!-- finding:PR91#notes -->\nNotes of PR #91.\n\n"
         "- [ ] `n1` [INFERRED] `src/a.go:1` guess one\n- [ ] `n2` [OBSERVED] `src/b.go:2` remark two")
keys = ("#12", "#13", "#14", "#15", "#16", "#20/n1", "#20/n2")
batch = ("<!-- review-batch -->\n# Review batch 2026-11-14\n\nRanked.\n\n"
         + "\n".join("- [ ] %s row %s <!-- batch-item:%s -->" % (k, k, k) for k in keys) + "\n\nNothing else is waiting.\n")
issues = [finding(7), finding(12), finding(13), finding(14), finding(15), finding(16),
          {"number": 20, "state": "open", "state_reason": None, "title": "PR91 notes", "created_at": "2026-11-13T00:00:00Z",
           "labels": [{"name": "kind:note"}], "body": notes, "comments": []},
          {"number": 40, "state": "open", "state_reason": None, "title": "Review batch 2026-11-14", "created_at": "2026-11-14T00:00:00Z",
           "labels": [{"name": "review-batch"}], "body": batch, "comments": []}]
json.dump(issues, open(os.path.join(st, "issues.json"), "w"))
json.dump({"number": 150, "merged_at": "2026-11-13T12:00:00Z"}, open(os.path.join(st, "pr150.json"), "w"))
json.dump([{"filename": "src/f12.go"}], open(os.path.join(st, "files150.json"), "w"))
json.dump({"number": 151, "merged_at": None}, open(os.path.join(st, "pr151.json"), "w"))
json.dump([{"filename": "src/f13.go"}], open(os.path.join(st, "files151.json"), "w"))
# The #226 mix: FIXED ELSEWHERE (merged and not merged), NOT A DEFECT, DECIDED, FIXED without a commit, AFTER FREEZE;
# #16 has no verdict; #7 is not an item of the batch.
result = {"outcome": "nothing", "reason": "no item needed a code change", "results": [
    {"item": "#12", "verdict": "FIXED ELSEWHERE", "reason": "fixed by #150"},
    {"item": "#13", "verdict": "FIXED ELSEWHERE", "reason": "fixed by #151"},
    {"item": "#14", "verdict": "NOT A DEFECT", "reason": evil},
    {"item": "#15", "verdict": "AFTER FREEZE", "reason": "not a gate"},
    {"item": "#20/n1", "verdict": "DECIDED", "reason": "ADR-0012 decides it"},
    {"item": "#20/n2", "verdict": "FIXED", "reason": "already guarded"},
    {"item": "#7", "verdict": "FIXED ELSEWHERE", "reason": "fixed by #150"}]}
json.dump(result, open(os.path.join(st, "result.json"), "w"))
EOF
  : >"$tmp/st/log"
}

# run [ITEMS]: the trusted step's call; sets rc and out.
run() {
  rc=0
  out="$(GH_RETRY_DELAY=0 PATH="$tmp/bin:$PATH" ST="$tmp/st" REPO=o/r ISSUE=40 RUN_ID=4242 ITEMS="${1:-#12,#13,#14,#15,#16,#20/n1,#20/n2}" \
    RESULT_JSON="$(cat "$tmp/st/result.json")" python3 -I "$tool" nothing 2>&1)" || rc=$?
}
field() { python3 -I -c 'import json,sys; x=[i for i in json.load(open(sys.argv[1])) if i["number"]==int(sys.argv[2])][0]; print(eval(sys.argv[3]))' "$tmp/st/issues.json" "$1" "$2"; }
bot_comments() { field "$1" 'sum(1 for c in x["comments"] if c["user"]["login"].endswith("[bot]"))'; }
last_comment() { field "$1" 'x["comments"][-1]["body"]'; }
# raw TEXT: true when TEXT still holds an untrusted form that acts (a mention, #N, GH-N, an issue URL, a forged marker).
raw() { grep -qE '@s|#7|GH-7|github\.com/o/r/issues/7|<!-- batch-handled|<!-- settled-by:batch40-run99' <<<"$1"; }

setup
run
[ "$rc" -eq 0 ] && ok "nothing: runs (rc=0)" || ko "nothing rc=$rc ($out)"

# 1. ONE comment on the batch issue with the verdict table.
[ "$(bot_comments 40)" -eq 1 ] && ok "one comment on the batch issue" || ko "batch comments ($(bot_comments 40))"
c40="$(field 40 'x["comments"][0]["body"]' 2>/dev/null || true)"
grep -q '| Item | Verdict | Reason |' <<<"$c40" && grep -q '^| #12 | FIXED ELSEWHERE | ' <<<"$c40" && grep -q '^| #20/n1 | DECIDED | ' <<<"$c40" \
  && grep -q '^| #20/n2 | OWNER | ' <<<"$c40" && ok "the batch comment holds the verdict table (item, verdict, reason); FIXED without a commit is OWNER" || ko "table ($c40)"
! grep -q '^| #16 \|^| #7 ' <<<"$c40" && ok "no row for an item without a verdict (#16) or outside the batch (#7)" || ko "extra rows ($c40)"

# 2. The settle rules.
[ "$(field 12 'x["state"], x["state_reason"]')" = "('closed', 'completed')" ] && ok "FIXED ELSEWHERE naming a merged PR that touched the cited file closes the source" || ko "elsewhere merged ($(field 12 'x["state"]') $(last_comment 12))"
[ "$(field 13 'x["state"]')" = "open" ] && last_comment 13 | grep -q 'named PR #151 is not merged: a human checks' && ok "FIXED ELSEWHERE naming an unmerged PR stays open with a comment" || ko "elsewhere unmerged ($(last_comment 13))"
[ "$(field 14 'x["state"]')" = "open" ] && last_comment 14 | grep -q 'NOT A DEFECT' && ok "NOT A DEFECT: comment only, the source stays open (H5)" || ko "not a defect ($(last_comment 14))"
[ "$(field 15 'x["state"]')" = "open" ] && last_comment 15 | grep -q 'OWNER' && ! last_comment 15 | grep -q 'AFTER FREEZE' && ok "a fixer result of AFTER FREEZE (retired verdict) is treated as OWNER: comment, stays open" || ko "after freeze ($(last_comment 15))"
field 20 'x["body"]' | grep -q -- '- \[ \] `n1`' && field 20 'x["body"]' | grep -q -- '- \[ \] `n2`' && [ "$(field 20 'x["state"]')" = "open" ] \
  && ok "notes: DECIDED and FIXED-without-a-commit leave their boxes unticked, the issue open" || ko "notes ($(field 20 'x["body"]'))"
last_comment 20 | grep -q '`n1` DECIDED' && last_comment 20 | grep -q '`n2` OWNER' && last_comment 20 | grep -q 'made no commit' \
  && ok "notes: one comment with both verdicts, FIXED shown as OWNER because the fixer made no commit" || ko "notes comment ($(last_comment 20))"
[ "$(bot_comments 20)" -eq 1 ] && ok "one comment per touched source issue" || ko "notes comments ($(bot_comments 20))"
last_comment 12 | grep -q 'Settled by the review batch run 4242 (no fix pull request) (review batch #40)' && ok "the source comment names the batch run and the batch" || ko "source comment text ($(last_comment 12))"
[ "$(field 16 'x["comments"][-1]["body"]')" = "human note" ] && ok "an item without a verdict (#16) is not touched" || ko "no verdict touched ($(last_comment 16))"

# 4. Only the batch issue's own items.
[ "$(field 7 'x["state"]')" = "open" ] && [ "$(bot_comments 7)" -eq 0 ] && ! grep -qE '^[A-Z]+ repos/o/r/issues/7[ /]' "$tmp/st/log" && ok "a result for an issue outside the batch (#7) is ignored" || ko "outside batch ($(grep -E '^[A-Z]+ repos/o/r/issues/7[ /]' "$tmp/st/log"))"

# Untrusted text neutralised, in the batch comment and in the source comment.
! raw "$c40" && grep -q 'a pipe' <<<"$c40" && grep -q '\\|' <<<"$c40" && ok "batch comment: mention, #N, GH-N, issue URL and forged markers neutralised; the pipe escaped" || ko "batch comment neutral ($c40)"
! raw "$(last_comment 14)" && last_comment 14 | grep -q 'a pipe' && ok "source comment: the same text is neutralised" || ko "source neutral ($(last_comment 14))"

# 3. Items with a verdict are ticked; the batch closes so the next batch run moves on; #16 (no verdict) stays unticked.
b40="$(field 40 'x["body"]')"
grep -q -- '- \[x\] #12 ' <<<"$b40" && grep -q -- '- \[x\] #20/n2 ' <<<"$b40" && grep -q -- '- \[ \] #16 ' <<<"$b40" && ok "every item with a verdict is ticked in the batch issue; an item without one is not" || ko "ticks ($b40)"
[ "$(field 40 'x["state"], x["state_reason"]')" = "('closed', 'completed')" ] && grep -qx '<!-- batch-handled -->' <<<"$b40" && ok "the batch closes as completed with the batch-handled marker" || ko "batch close ($(field 40 'x["state"]') $b40)"
grep -q 'PATCH repos/o/r/issues/40 ' <<<"$(tail -1 "$tmp/st/log")" && ok "the batch issue is written last, after its sources" || ko "order ($(tail -1 "$tmp/st/log"))"

before="$(cat "$tmp/st/issues.json")"
: >"$tmp/st/log"
run
[ "$rc" -eq 0 ] && [ ! -s "$tmp/st/log" ] && [ "$before" = "$(cat "$tmp/st/issues.json")" ] && ok "a second run changes nothing (idempotent)" || ko "idempotent (rc=$rc log=$(cat "$tmp/st/log"))"

# The guard is the batch issue's markers, whatever ITEMS says.
setup
run "#7,#12"
[ "$(field 7 'x["state"]')" = "open" ] && [ "$(bot_comments 7)" -eq 0 ] && [ "$(field 12 'x["state"]')" = "closed" ] && ok "a key in ITEMS but not in the batch issue's markers (#7) is ignored" || ko "items guard ($(field 7 'x["state"]'))"

# No result, or no verdict for any item: nothing written.
setup
printf 'not json' >"$tmp/st/result.json"
run
[ "$rc" -eq 0 ] && [ ! -s "$tmp/st/log" ] && grep -q 'NOT RUN' <<<"$out" && ok "an unreadable result: NOT RUN, nothing written" || ko "no result (rc=$rc out=$out log=$(cat "$tmp/st/log"))"
setup
printf '{"outcome":"nothing","reason":"x","results":[{"item":"#7","verdict":"OWNER","reason":"x"},{"item":"#12","verdict":"MAYBE","reason":"x"}]}' >"$tmp/st/result.json"
run
[ "$rc" -eq 0 ] && [ ! -s "$tmp/st/log" ] && [ "$(field 40 'x["state"]')" = "open" ] && ok "no valid verdict for an item of the batch: nothing written, the batch stays open" || ko "no verdict (out=$out)"

# A batch issue that is not a review batch, or already closed: nothing written.
setup
python3 -I - "$tmp/st/issues.json" <<'EOF'
import json, sys
d = json.load(open(sys.argv[1]))
for x in d:
    if x["number"] == 40: x["state"] = "closed"
json.dump(d, open(sys.argv[1], "w"))
EOF
run
[ "$rc" -eq 0 ] && [ ! -s "$tmp/st/log" ] && ok "a closed batch issue: nothing written" || ko "closed batch ($(cat "$tmp/st/log"))"

# Fail closed on a GitHub error.
setup
FAKE_GH_FAIL=1 run
[ "$rc" -ne 0 ] && grep -q 'FAILED' <<<"$out" && ok "a GitHub error fails closed (FAILED, rc=$rc)" || ko "gh failure (rc=$rc $out)"

# GitHub error detail and one retry (batch run 43673158317 printed only "gh failed"). Writes retry once on a 5xx, a
# secondary rate limit or a network error, never on a 4xx; a GET retries once too; the message names the call.
attempts() { cat "$tmp/st/attempts" 2>/dev/null || echo 0; }
setup; rm -f "$tmp/st/attempts"
FAKE_GH_FAILS=1 FAKE_GH_STDERR='gh: Bad Gateway (HTTP 502)' run
[ "$rc" -eq 0 ] && [ "$(attempts)" -eq 1 ] && [ "$(bot_comments 40)" -eq 1 ] && [ "$(field 40 'x["state"]')" = "closed" ] \
  && ok "a write that fails once with 502 is retried once and settles" || ko "502 retry (rc=$rc attempts=$(attempts) $out)"
setup; rm -f "$tmp/st/attempts"
FAKE_GH_FAILS=1 FAKE_GH_STDERR='gh: Service Unavailable (HTTP 503)' FAKE_GH_KIND=get run
[ "$rc" -eq 0 ] && [ "$(attempts)" -eq 1 ] && ok "a GET that fails once is retried once" || ko "get retry (rc=$rc attempts=$(attempts) $out)"
setup; rm -f "$tmp/st/attempts"
FAKE_GH_FAILS=5 FAKE_GH_STDERR='gh: Bad Gateway (HTTP 502)' run
[ "$rc" -ne 0 ] && [ "$(attempts)" -eq 2 ] && grep -qF 'PATCH repos/o/r/issues/12: exit 1: gh: Bad Gateway (HTTP 502)' <<<"$out" \
  && ok "a second 502 gives up after exactly one retry, with the detail" || ko "502 twice (rc=$rc attempts=$(attempts) $out)"
setup; rm -f "$tmp/st/attempts"
FAKE_GH_FAILS=5 FAKE_GH_STDERR='gh: Forbidden (HTTP 403)' run
[ "$rc" -ne 0 ] && [ "$(attempts)" -eq 1 ] && grep -qF 'FAILED (GitHub refused PATCH repos/o/r/issues/12: exit 1: gh: Forbidden (HTTP 403))' <<<"$out" \
  && ! grep -q 'second line\|human note\|Item | Verdict' <<<"$out" && ok "a 403 fails with method, endpoint, exit code and first stderr line, no retry, no body" || ko "403 (rc=$rc attempts=$(attempts) $out)"
setup; rm -f "$tmp/st/attempts"
FAKE_GH_FAILS=5 FAKE_GH_STDERR='HTTP 403 token ghp_abcdefghijklmnopqrstuvwxyz0123456789' run
[ "$rc" -ne 0 ] && ! grep -q 'ghp_' <<<"$out" && grep -qF 'PATCH repos/o/r/issues/12: exit 1:' <<<"$out" && grep -q 'redacted' <<<"$out" \
  && ok "a secret-shaped string in stderr is redacted" || ko "redaction (rc=$rc $out)"
setup; rm -f "$tmp/st/attempts"
FAKE_GH_FAILS=5 FAKE_GH_STDERR='gh: Not Found (HTTP 404)' FAKE_GH_KIND=get run
[ "$rc" -ne 0 ] && [ "$(attempts)" -eq 1 ] && grep -q 'GET repos/o/r/issues/40: exit 1: gh: Not Found (HTTP 404)' <<<"$out" \
  && ok "a GET 404 is not retried and names the call (no query string)" || ko "get 404 (rc=$rc attempts=$(attempts) $out)"

# A POST is not idempotent: a timeout or a 5xx may come after GitHub applied it, so a retry could duplicate a comment or
# the batch issue. It is retried only when the request certainly was not applied (secondary rate limit, 429, a
# connection that never reached GitHub); PATCH, PUT and DELETE keep the 5xx and network rule.
setup; rm -f "$tmp/st/attempts"
FAKE_GH_FAILS=5 FAKE_GH_METHODS=POST FAKE_GH_STDERR='gh: Bad Gateway (HTTP 502)' run
[ "$rc" -ne 0 ] && [ "$(attempts)" -eq 1 ] && grep -qF 'POST repos/o/r/issues/12/comments: exit 1: gh: Bad Gateway (HTTP 502)' <<<"$out" \
  && ok "a POST that gets a 502 is NOT retried (it may have been applied): FAILED with the detail" || ko "POST 502 (rc=$rc attempts=$(attempts) $out)"
setup; rm -f "$tmp/st/attempts"
FAKE_GH_FAILS=5 FAKE_GH_METHODS=POST FAKE_GH_STDERR='timeout' run
[ "$rc" -ne 0 ] && [ "$(attempts)" -eq 1 ] && ok "a POST that times out is NOT retried" || ko "POST timeout (rc=$rc attempts=$(attempts) $out)"
setup; rm -f "$tmp/st/attempts"
FAKE_GH_FAILS=1 FAKE_GH_METHODS=POST FAKE_GH_STDERR='dial tcp 140.82.0.1:443: connect: connection refused' run
[ "$rc" -eq 0 ] && [ "$(attempts)" -eq 1 ] && [ "$(bot_comments 40)" -eq 1 ] && ok "a POST whose connection was refused is retried once and settles" || ko "POST refused (rc=$rc attempts=$(attempts) $out)"
setup; rm -f "$tmp/st/attempts"
FAKE_GH_FAILS=5 FAKE_GH_METHODS=PATCH FAKE_GH_STDERR='gh: Bad Gateway (HTTP 502)' run
[ "$rc" -ne 0 ] && [ "$(attempts)" -eq 2 ] && ok "a PATCH that gets a 502 is retried once" || ko "PATCH 502 (rc=$rc attempts=$(attempts) $out)"

# The retry rule itself, branch by branch: chk METHOD STDERR yes|no.
rule() {
  python3 -I -c '
import importlib.util, sys
spec = importlib.util.spec_from_file_location("rc", sys.argv[1]); m = importlib.util.module_from_spec(spec); spec.loader.exec_module(m)
print("yes" if m.retryable(sys.argv[2], sys.argv[3]) else "no")' "$root/scripts/review-context.py" "$1" "$2"
}
chk() { [ "$(rule "$1" "$2")" = "$3" ] && ok "retry rule: $1 '$2' -> $3" || ko "retry rule: $1 '$2' (got $(rule "$1" "$2"), want $3)"; }
chk PATCH 'gh: Too Many Requests (HTTP 429) Retry-After: 3' yes
chk POST 'gh: Too Many Requests (HTTP 429) Retry-After: 3' yes
chk PATCH 'gh: Too Many Requests (HTTP 429) Retry-After: 30' no
chk PATCH 'gh: Too Many Requests (HTTP 429)' no
chk PATCH 'gh: You have exceeded a secondary rate limit (HTTP 403)' yes
chk POST 'gh: You have exceeded a secondary rate limit (HTTP 403)' yes
chk GET 'gh: You have exceeded a secondary rate limit (HTTP 403)' yes
chk POST 'dial tcp: lookup api.github.com: no such host' yes
chk POST 'tls handshake failure' yes
chk POST 'read tcp: connection reset by peer' no
chk PATCH 'read tcp: connection reset by peer' yes
chk DELETE 'gh: Service Unavailable (HTTP 503)' yes
chk PUT 'gh: Unprocessable Entity (HTTP 422)' no

# ---------- workflow wiring (static): the trusted publish job runs it, from the planned develop commit ----------
wf="$root/.github/workflows/review-batch.yml"
publish="$(sed -n '/^  publish:/,/^  report:/p' "$wf")"
grep -q 'review-autofix.py" nothing' <<<"$publish" && ok "review-batch.yml: the publish job runs the nothing path" || ko "nothing step"
grep -q 'for module in review-autofix findings review-context' <<<"$publish" && grep -qF 'cat-file blob "$base_sha:scripts/$module.py"' <<<"$publish" \
  && ok "publish: the three trusted scripts are read as git objects of the planned commit" || ko "git objects"
grep -q 'issues: write' <<<"$publish" && grep -q 'pull-requests: read' <<<"$publish" && ok "publish: issues write and pull-requests read (the verdicts and the merged check)" || ko "publish permissions"
! grep -v '^\s*#' <<<"$publish" | grep -qE '\bmake\b|\bnpm |\bpnpm |go test' && ok "publish still runs no repository code (no make, npm, pnpm, go test)" || ko "publish runs repository code"

finish
