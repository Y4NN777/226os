#!/usr/bin/env bash
# Tests for scripts/review-autofix.py with a fake `gh` and a throwaway git repository: when the CI fixer runs (stage 3,
# one pull request; stage 5, the review batch), which findings it gets, the attempt count, the skip rules, the guard
# that runs before any push (paths, identity, trailers) and the batch pull request body.
# Usage: scripts/tests/review-autofix_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
tool="$root/scripts/review-autofix.py"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin" "$tmp/st"

# Fake gh: `gh api [-X METHOD] [-f k=v]... ENDPOINT [--paginate]`. A GET answers from a fixture file in $ST chosen by
# the endpoint; every other method is logged to $ST/log and answers {}. FAKE_GH_FAIL fails every call.
cat >"$tmp/bin/gh" <<'EOF'
#!/usr/bin/env python3
import json, os, sys
st = os.environ["ST"]
if os.environ.get("FAKE_GH_FAIL"):
    sys.stderr.write("gh: boom\n"); sys.exit(1)
args = sys.argv[1:]
method, endpoint, i = "GET", None, 1
while i < len(args):
    if args[i] == "-X": method = args[i + 1]; i += 2
    elif args[i] in ("-f", "-F"): i += 2
    elif args[i] == "--paginate": i += 1
    else: endpoint = args[i]; i += 1
if method != "GET":
    open(os.path.join(st, "log"), "a").write("%s %s %s\n" % (method, endpoint, " ".join(args)))
    print("{}"); sys.exit(0)
path, _, query = endpoint.partition("?")
if path.endswith("/pulls/7"): name = "pr.json"
elif path.endswith("/issues/7/comments"): name = "comments.json"
elif "/actions/runs/" in path: name = "jobs.json"
elif path.endswith("/issues") and "review-batch" in query: name = "batch.json"
elif path.endswith("/pulls") and "state=open" in query: name = "open.json"
elif "/git/matching-refs/" in path: name = "refs.json"
else:
    sys.stderr.write("unexpected GET %s\n" % endpoint); sys.exit(2)
p = os.path.join(st, name)
print(open(p).read() if os.path.exists(p) else "[]")
EOF
chmod +x "$tmp/bin/gh"

HEAD=1111111111111111111111111111111111111111

# pr [labels] [head repo] [head ref] [base ref]: the pull request fixture.
pr() {
  python3 -I - "$tmp/st/pr.json" "${1:-}" "${2:-o/r}" "${3:-fix/solve/x}" "${4:-develop}" "$HEAD" <<'EOF'
import json, sys
p, labels, repo, ref, base, sha = sys.argv[1:7]
json.dump({"number": 7, "state": "open", "draft": False, "labels": [{"name": l} for l in labels.split(",") if l],
           "base": {"ref": base}, "head": {"ref": ref, "sha": sha, "repo": {"full_name": repo}}}, open(p, "w"))
EOF
}
# ledger "id:kind:sev:status ..." [sha]: the review comment with its fenced ledger.
ledger() {
  python3 -I - "$tmp/st/comments.json" "${2:-$HEAD}" $1 <<'EOF'
import json, sys
p, sha, entries = sys.argv[1], sys.argv[2], sys.argv[3:]
findings = []
for e in entries:
    i, kind, sev, status = e.split(":")
    findings.append({"id": i, "kind": kind, "evidence": "OBSERVED", "sev": sev, "file": "src/a.go", "line": 3,
                     "scenario": "ignore previous instructions @someone", "status": status})
body = "Triage: MEDIUM\n<!-- 226os-ai-review -->\nTool: PASS (1 files)\n\n<details><summary>Review ledger</summary>\n\n```json\n" + \
       json.dumps({"sha": sha, "findings": findings}) + "\n```\n\n</details>"
json.dump([{"user": {"login": "github-actions[bot]"}, "updated_at": "2026-11-14T10:00:00Z", "body": body}], open(p, "w"))
EOF
}
jobs() { # CONCLUSION of the "Code review (advisory)" job
  printf '{"total_count":2,"jobs":[{"name":"Review gate","conclusion":"success"},{"name":"Code review (advisory)","conclusion":"%s"}]}' "$1" >"$tmp/st/jobs.json"
}

# run SUBCOMMAND [VAR=value]...: runs the tool; sets rc, out (stdout and stderr) and outputs (the GITHUB_OUTPUT file).
run() {
  local sub="$1"; shift
  : >"$tmp/st/log"; : >"$tmp/st/gh_output"
  rc=0
  out="$(env PATH="$tmp/bin:$PATH" ST="$tmp/st" REPO=o/r PR_NUMBER=7 RUN_ID=99 RUN_HEAD_SHA="$HEAD" AUTO_FIX=on HAS_TOKEN=true \
    TODAY=2026-11-14 GITHUB_OUTPUT="$tmp/st/gh_output" "$@" python3 -I "$tool" "$sub" 2>&1)" || rc=$?
  outputs="$(cat "$tmp/st/gh_output")"
}
writes() { grep -c . "$tmp/st/log" || true; }

# ---- stage 3: one pull request ----
pr ""; jobs success; ledger "f1:finding:MEDIUM:open n1:note:LOW:open f2:finding:LOW:fixed"
run pr
[ "$rc" -eq 0 ] && grep -qx 'run=true' <<<"$outputs" && grep -qx 'findings=f1' <<<"$outputs" && ok "an open MEDIUM finding triggers the fixer, with only its id" || ko "open medium (rc=$rc out=$out outputs=$outputs)"
grep -qx 'attempt=1' <<<"$outputs" && grep -qx 'branch=fix/solve/x' <<<"$outputs" && grep -q 'labels\[\]=autofix-1' "$tmp/st/log" && ok "the first attempt is labelled autofix-1 on the head branch" || ko "attempt 1 ($outputs / $(cat "$tmp/st/log"))"
! grep -q 'ignore previous\|someone' <<<"$outputs" && ok "no ledger text reaches the outputs (ids only)" || ko "text leak ($outputs)"

ledger "f1:finding:HIGH:open"
run pr
grep -qx 'run=true' <<<"$outputs" && ok "an open HIGH triggers the fixer" || ko "high ($out)"

ledger 'f1;rm:finding:MEDIUM:open'
run pr
grep -qx 'findings=f1-rm' <<<"$outputs" && ok "a finding id is reduced to safe characters" || ko "id slug ($outputs)"

ledger "n1:note:LOW:open n2:note:HIGH:open f1:finding:MEDIUM:answered"
run pr
grep -qx 'run=false' <<<"$outputs" && [ "$(writes)" -eq 0 ] && ok "notes only (and answered findings): no fix, no write" || ko "notes only ($out / $outputs)"

ledger "f1:finding:LOW:open"
run pr
[ "$rc" -eq 0 ] && grep -qx 'run=true' <<<"$outputs" && grep -qx 'findings=f1' <<<"$outputs" && ok "a PR whose only open finding is LOW is fixed (auto-merge refuses any open finding): verdict fix with that id" || ko "low ($outputs)"

ledger "c1:finding:CRITICAL:open f1:finding:MEDIUM:open"
run pr
grep -qx 'run=false' <<<"$outputs" && grep -q 'labels\[\]=needs-owner' "$tmp/st/log" && ! grep -q 'autofix-' "$tmp/st/log" && ok "an open CRITICAL sends the whole pull request to the owner, no fix" || ko "critical ($out / $(cat "$tmp/st/log"))"

ledger "f1:finding:MEDIUM:open"
pr "autofix-1"
run pr
grep -qx 'attempt=2' <<<"$outputs" && grep -q 'labels\[\]=autofix-2' "$tmp/st/log" && ok "the second attempt is labelled autofix-2" || ko "attempt 2 ($outputs)"

pr "autofix-1,autofix-2"
run pr
grep -qx 'run=false' <<<"$outputs" && grep -q 'labels\[\]=needs-owner' "$tmp/st/log" && grep -qi 'attempts' <<<"$out" && ok "the third attempt is refused and the pull request goes to the owner" || ko "third attempt ($out / $(cat "$tmp/st/log"))"

pr "" "someone/fork"
run pr
grep -qx 'run=false' <<<"$outputs" && [ "$(writes)" -eq 0 ] && grep -qi 'fork' <<<"$out" && ok "a pull request from a fork is refused, nothing written" || ko "fork ($out)"

pr "" o/r main
run pr
grep -qx 'run=false' <<<"$outputs" && ok "a head branch main is refused" || ko "main head ($out)"

pr "" o/r fix/solve/x main
run pr
grep -qx 'run=false' <<<"$outputs" && ok "a pull request into main is refused" || ko "base main ($out)"

pr ""
run pr HAS_TOKEN=
grep -qx 'run=false' <<<"$outputs" && [ "$(writes)" -eq 0 ] && grep -q 'autofix: NOT RUN (REVIEW_AGENT_TOKEN missing)' <<<"$out" && [ "$rc" -eq 0 ] && ok "a missing token: NOT RUN, exit 0, nothing written" || ko "no token (rc=$rc $out)"

run pr AUTO_FIX=off
grep -qx 'run=false' <<<"$outputs" && [ "$(writes)" -eq 0 ] && ok "mode off: no fix, nothing written" || ko "off ($out)"

run pr AUTO_FIX=
grep -qx 'run=false' <<<"$outputs" && [ "$(writes)" -eq 0 ] && grep -q 'would fix' <<<"$out" && ok "mode report (the default): says what it would fix, writes nothing" || ko "report ($out)"

jobs skipped
run pr
grep -qx 'run=false' <<<"$outputs" && [ "$(writes)" -eq 0 ] && ok "a Code review run whose review job was skipped (a label or review event) spends no attempt" || ko "skipped review job ($out)"
jobs success

ledger "f1:finding:MEDIUM:open" 2222222222222222222222222222222222222222
run pr
grep -qx 'run=false' <<<"$outputs" && [ "$(writes)" -eq 0 ] && ok "a ledger that is not for the pull request head is stale: no fix" || ko "stale ledger ($out)"
ledger "f1:finding:MEDIUM:open"

run pr RUN_HEAD_SHA=3333333333333333333333333333333333333333
grep -qx 'run=false' <<<"$outputs" && ok "a review run for an older head (a newer push exists) is skipped" || ko "older run ($out)"

run pr FAKE_GH_FAIL=1
[ "$rc" -eq 0 ] && grep -qx 'run=false' <<<"$outputs" && grep -q 'NOT RUN' <<<"$out" && ok "a GitHub error is NOT RUN, exit 0" || ko "gh failure (rc=$rc $out)"

# ---- stage 5: the review batch ----
batch_issue() { # BODY_LINES (\n separated)
  python3 -I - "$tmp/st/batch.json" "$1" <<'EOF'
import json, sys
json.dump([{"number": 40, "state": "open", "labels": [{"name": "review-batch"}], "title": "Review batch 2026-11-14",
            "body": "<!-- review-batch -->\n# Review batch 2026-11-14\n\n" + sys.argv[2].replace("\\n", "\n")}], open(sys.argv[1], "w"))
EOF
}
batch_issue '- [ ] #12 [HIGH] x <!-- batch-item:#12 -->\n- [x] #13 [MEDIUM] y <!-- batch-item:#13 -->\n- [ ] #20 [NOTE] z <!-- batch-item:#20/n1 -->'
printf '[]' >"$tmp/st/open.json"; printf '[]' >"$tmp/st/refs.json"
run batch BATCH_RESULT='findings batch: created "Review batch 2026-11-14" (3 items, 0 more)'
grep -qx 'run=true' <<<"$outputs" && grep -qx 'issue=40' <<<"$outputs" && grep -qx 'branch=fix/review/batch-2026-11-14' <<<"$outputs" && grep -qx 'items=#12,#20/n1' <<<"$outputs" && [ "$(writes)" -eq 0 ] && ok "a created batch starts the fixer on its unticked items, from develop, branch named by date" || ko "batch run ($out / $outputs)"

run batch BATCH_RESULT='findings batch: #40 unchanged (2 open, 0 done, 0 more)'
grep -qx 'run=false' <<<"$outputs" && ok "an unchanged batch does not start the fixer" || ko "unchanged ($outputs)"

run batch BATCH_RESULT='findings batch: #40 unchanged' FORCE=true
grep -qx 'run=true' <<<"$outputs" && ok "a manual run (FORCE) starts it anyway" || ko "force ($outputs)"

printf '[{"number":77,"head":{"ref":"fix/review/batch-2026-11-13"},"base":{"ref":"develop"}}]' >"$tmp/st/open.json"
run batch BATCH_RESULT='findings batch: updated #40'
grep -qx 'run=false' <<<"$outputs" && grep -q '#77' <<<"$out" && ok "a batch pull request already open: skipped" || ko "open batch pr ($out)"
printf '[]' >"$tmp/st/open.json"

printf '[{"ref":"refs/heads/fix/review/batch-2026-11-14"}]' >"$tmp/st/refs.json"
run batch BATCH_RESULT='findings batch: updated #40'
grep -qx 'branch=fix/review/batch-2026-11-14-2' <<<"$outputs" && ok "a branch of the same day that already exists gets a suffix" || ko "branch suffix ($outputs)"
printf '[]' >"$tmp/st/refs.json"

run batch BATCH_RESULT='findings batch: updated #40' HAS_TOKEN=
grep -qx 'run=false' <<<"$outputs" && grep -q 'autofix: NOT RUN (REVIEW_AGENT_TOKEN missing)' <<<"$out" && ok "batch: a missing token is NOT RUN" || ko "batch no token ($out)"

run batch BATCH_RESULT='findings batch: updated #40' AUTO_FIX=off
grep -qx 'run=false' <<<"$outputs" && ok "batch: mode off does nothing" || ko "batch off ($out)"

printf '[]' >"$tmp/st/batch.json"
run batch BATCH_RESULT='findings batch: updated #40'
grep -qx 'run=false' <<<"$outputs" && ok "no open batch issue: nothing to fix" || ko "no batch ($out)"

# ---- the guard before any push ----
repo="$tmp/repo"
git init -q -b develop "$repo"
g() { git -C "$repo" -c user.name=Y4NN77 -c user.email=fixer@example.invalid "$@"; }
mkdir -p "$repo/src" "$repo/api"
echo a >"$repo/src/a.go"; echo c >"$repo/api/openapi.yaml"
g add -A; g commit -qm base
base="$(git -C "$repo" rev-parse HEAD)"
check() { run check BASE_SHA="$base" WORKTREE="$repo" EXPECT_NAME=Y4NN77 EXPECT_EMAIL=fixer@example.invalid; }

check
[ "$rc" -ne 0 ] && grep -q 'no commit' <<<"$out" && ok "guard: no new commit is a failure (nothing to push)" || ko "no commit (rc=$rc $out)"

echo b >>"$repo/src/a.go"; g add src/a.go; g commit -qm "fix(solve): x"
check
[ "$rc" -eq 0 ] && grep -q 'PASS' <<<"$out" && ok "guard: a commit by the human identity on an allowed path passes" || ko "good commit (rc=$rc $out)"

g commit -q --allow-empty -m "chore: y

Co-authored-by: Claude <noreply@anthropic.com>"
check
[ "$rc" -ne 0 ] && grep -qi 'trailer' <<<"$out" && ok "guard: a Co-authored-by trailer is refused" || ko "trailer (rc=$rc $out)"
git -C "$repo" reset -q --hard HEAD~1

git -C "$repo" -c user.name=claude -c user.email=x@y commit -q --allow-empty -m "chore: z"
check
[ "$rc" -ne 0 ] && grep -qi 'identity' <<<"$out" && ok "guard: another author identity is refused" || ko "identity (rc=$rc $out)"
git -C "$repo" reset -q --hard HEAD~1

for forbidden in api/openapi.yaml db/migrations/00099_x.sql docs/decisions/0011-x.md secrets/x.enc.env .github/workflows/x.yml .review-context.md .npmrc src/web/.npmrc docs/team/OWNER_PREFERENCES.md \
    docs/quality/COVERAGE_SCOPE.md docs/quality/REVIEW_CONTEXT.md scripts/review-gate.py scripts/findings.py scripts/close-merged-issues.py \
    scripts/coverage-go.sh scripts/check-docs.py Makefile mk/20-gates.mk .github/pipeline-paths.txt .github/sensitive-paths.txt \
    .github/deep-review-paths.txt scripts/gitleaks-ignore.txt scripts/public-routes.txt scripts/unimplemented-routes.txt scripts/authz-tests.txt \
    scripts/tests/review-comment_test.sh scripts/tests/review-comment-levels.py scripts/tests/review-automerge_test.sh scripts/tests/findings_test.sh \
    scripts/tests/findings-settle_test.sh scripts/tests/lib.sh scripts/tests/review-agent-triggers_test.sh; do
  mkdir -p "$repo/$(dirname "$forbidden")"; echo x >>"$repo/$forbidden"; g add -f "$forbidden"; g commit -qm "fix: touch"
  check
  [ "$rc" -ne 0 ] && grep -q "$forbidden" <<<"$out" && ok "guard: $forbidden is refused (owner path)" || ko "forbidden $forbidden (rc=$rc $out)"
  git -C "$repo" reset -q --hard HEAD~1
done

# Both sides of a rename, quoted paths, links, submodules, mode changes and binary changes.
mkdir -p "$repo/.github/workflows" "$repo/db/migrations" "$repo/docs"
echo q >"$repo/.github/workflows/quality.yml"; echo m >"$repo/db/migrations/0001.sql"
g add -f .github/workflows/quality.yml db/migrations/0001.sql; g commit -qm base2
base="$(git -C "$repo" rev-parse HEAD)"
for moved in .github/workflows/quality.yml db/migrations/0001.sql; do
  mkdir -p "$repo/docs"; g mv "$moved" "docs/$(basename "$moved")"; g commit -qm "docs: move"
  check
  [ "$rc" -ne 0 ] && grep -q "$moved" <<<"$out" && ok "guard: a rename out of $moved is refused (both sides are checked)" || ko "rename $moved (rc=$rc $out)"
  git -C "$repo" reset -q --hard HEAD~1
done
for odd in ".github/workflows/é.yml" ".claude/é.md" "db/migrations/é.sql" ".github/a	tab.yml"; do
  mkdir -p "$repo/$(dirname "$odd")"; echo x >"$repo/$odd"; g add -f "$odd"; g commit -qm "fix: odd"
  check
  [ "$rc" -ne 0 ] && grep -q 'owner paths' <<<"$out" && ok "guard: a quoted path under an owner directory is refused ($(basename "$odd"))" || ko "odd path $odd (rc=$rc $out)"
  git -C "$repo" reset -q --hard HEAD~1
done
ln -s /etc/passwd "$repo/src/link"; g add src/link; g commit -qm "fix: link"
check
[ "$rc" -ne 0 ] && grep -qi 'symlink' <<<"$out" && ok "guard: a symbolic link is refused" || ko "symlink (rc=$rc $out)"
git -C "$repo" reset -q --hard HEAD~1
g update-index --add --cacheinfo "160000,$base,src/sub"; g commit -qm "fix: sub"
check
[ "$rc" -ne 0 ] && grep -qi 'submodule' <<<"$out" && ok "guard: a submodule is refused" || ko "submodule (rc=$rc $out)"
git -C "$repo" reset -q --hard HEAD~1
chmod +x "$repo/src/a.go"; g add src/a.go; g commit -qm "fix: mode"
check
[ "$rc" -ne 0 ] && grep -qi 'mode' <<<"$out" && ok "guard: a file mode change is refused" || ko "mode (rc=$rc $out)"
git -C "$repo" reset -q --hard HEAD~1
printf 'a\000b' >"$repo/src/b.bin"; g add src/b.bin; g commit -qm "fix: bin"
check
[ "$rc" -ne 0 ] && grep -qi 'binary' <<<"$out" && ok "guard: a binary change is refused" || ko "binary (rc=$rc $out)"
git -C "$repo" reset -q --hard HEAD~1

# Allowlist: only src/**, db/queries/**, scripts/tests/**, docs/ai-usage/** and TODO.md; everything else is the owner's.
for owner in .golangci.yml scripts/semgrep/rules.yml scripts/sast-js.sh scripts/secrets-scan.sh scripts/ci-scope.sh \
    scripts/quality.sh scripts/review.sh scripts/router-middleware.txt scripts/authz/matrix.txt .opencodereview/rule.json \
    .semgrep.yml eslint.config.js .prettierrc tsconfig.json src/web/eslint.config.js src/web/.prettierrc \
    src/web/tsconfig.json src/web/tsconfig.app.json src/web/vite.config.ts src/web/vitest.config.ts src/web/package.json \
    package.json pnpm-lock.yaml go.mod README.md docs/SPEC.md src/api/Dockerfile src/web/Dockerfile src/api/.dockerignore \
    src/web/.dockerignore src/web/nginx/default.conf src/web/index.html src/web/.prettierignore \
    src/api/internal/platform/apispec/openapi.yaml src/api/internal/platform/apigen/api.go src/web/src/api/schema.d.ts \
    src/web/src/features/x.gen.ts src/api/internal/x.gen.go src/web/public/robots.txt; do
  mkdir -p "$repo/$(dirname "$owner")"; echo x >>"$repo/$owner"; g add -f "$owner"; g commit -qm "fix: owner"
  check
  [ "$rc" -ne 0 ] && grep -q "$owner" <<<"$out" && ok "allowlist: $owner is the owner's" || ko "allowlist $owner (rc=$rc $out)"
  git -C "$repo" reset -q --hard HEAD~1
done
mkdir -p "$repo/db/queries" "$repo/scripts/tests" "$repo/docs/ai-usage" "$repo/src/web/src"
for mine in db/queries/solve.sql scripts/tests/x_test.sh docs/ai-usage/2026-11-14-review-fix-pr1.md TODO.md src/web/src/x.tsx; do
  echo x >>"$repo/$mine"; g add -f "$mine"
done
g commit -qm "fix: allowed"
check
[ "$rc" -eq 0 ] && ok "allowlist: src, db/queries, scripts/tests, docs/ai-usage and TODO.md may change" || ko "allowlist allowed (rc=$rc $out)"
git -C "$repo" reset -q --hard HEAD~1
printf '#!/bin/sh\n' >"$repo/src/run.sh"; chmod +x "$repo/src/run.sh"; g add src/run.sh; g commit -qm "fix: exec"
check
[ "$rc" -ne 0 ] && grep -qi 'executable' <<<"$out" && ok "guard: a new executable file outside scripts/ is refused" || ko "exec outside scripts (rc=$rc $out)"
git -C "$repo" reset -q --hard HEAD~1
mkdir -p "$repo/scripts/tests"; printf '#!/bin/sh\n' >"$repo/scripts/tests/new_test.sh"; chmod +x "$repo/scripts/tests/new_test.sh"
g add scripts/tests/new_test.sh; g commit -qm "test: new"
check
[ "$rc" -eq 0 ] && ok "guard: a new executable test under scripts/tests/ is allowed" || ko "exec test (rc=$rc $out)"
git -C "$repo" reset -q --hard HEAD~1

# ---- prepare: right before the model, the home Claude config is wiped and the project config restored from a ref ----
prep="$tmp/prep"; fakehome="$tmp/home"
git init -q -b develop "$prep"
gp() { git -C "$prep" -c user.name=t -c user.email=t@example.invalid "$@"; }
mkdir -p "$prep/.claude"; echo base >"$prep/.claude/settings.json"; echo base >"$prep/CLAUDE.md"; echo a >"$prep/a.go"
gp add -A; gp commit -qm base; gp update-ref refs/remotes/origin/main HEAD
gp switch -q -c pr
echo evil >"$prep/.claude/settings.json"; echo hook >"$prep/.claude/hook.py"; echo '{}' >"$prep/.mcp.json"; echo evil >"$prep/CLAUDE.md"
gp add -A; gp commit -qm pr
echo b >>"$prep/a.go"; echo local >"$prep/.claude/settings.local.json"
mkdir -p "$fakehome/.claude" "$fakehome/.config/claude-code" "$fakehome/.config/other" "$fakehome/.cache/claude-cli" "$fakehome/.local/share/claude"
echo x >"$fakehome/.cache/claude-cli/x"; echo x >"$fakehome/.local/share/claude/x"
echo hook >"$fakehome/.claude/settings.json"; echo '{}' >"$fakehome/.claude.json"; echo x >"$fakehome/.config/claude-code/x"; echo keep >"$fakehome/.config/other/k"
run prepare SOURCE_REF=origin/main WORKTREE="$prep" HOME="$fakehome"
[ "$rc" -eq 0 ] && [ "$(cat "$prep/.claude/settings.json")" = base ] && [ "$(cat "$prep/CLAUDE.md")" = base ] && ok "prepare: .claude and CLAUDE.md come from the ref" || ko "prepare restore (rc=$rc $out)"
[ ! -e "$prep/.claude/hook.py" ] && [ ! -e "$prep/.mcp.json" ] && [ ! -e "$prep/.claude/settings.local.json" ] && ok "prepare: files the ref does not have are removed (hook, .mcp.json, settings.local.json)" || ko "prepare remove ($(ls -A "$prep" "$prep/.claude"))"
[ ! -e "$fakehome/.claude" ] && [ ! -e "$fakehome/.claude.json" ] && [ ! -e "$fakehome/.config/claude-code" ] && [ ! -e "$fakehome/.cache/claude-cli" ] && [ ! -e "$fakehome/.local/share/claude" ] && [ -e "$fakehome/.config/other/k" ] && ok "prepare: the home Claude config is wiped, nothing else in home" || ko "prepare home ($(ls -A "$fakehome" "$fakehome/.config"))"
[ "$(git -C "$prep" diff --cached --name-only | wc -l)" -eq 0 ] && grep -q '^b$' "$prep/a.go" && ok "prepare: nothing staged, the fixer's work tree is kept" || ko "prepare index"
git -C "$prep" add -A
[ -z "$(git -C "$prep" diff --cached --name-only -- .claude CLAUDE.md .mcp.json)" ] && ok "prepare: a later git add -A does not stage the restored configuration" || ko "prepare add -A ($(git -C "$prep" diff --cached --name-only))"
git -C "$prep" reset -q
run prepare SOURCE_REF=origin/main WORKTREE="$prep" HOME=/
[ "$rc" -ne 0 ] && ok "prepare: refuses a HOME of /" || ko "prepare home root (rc=$rc)"

# ---- apply: the trusted job validates the patch files and commits them itself, as the human identity ----
work="$tmp/work"
git clone -q "$repo" "$work"
wbase="$(git -C "$work" rev-parse HEAD)"
echo z >>"$repo/src/a.go"
git -C "$repo" -c user.name=claude -c user.email=x@y commit -qam "fix(solve): z"
pd="$tmp/pd"; git -C "$repo" format-patch -q -o "$pd" "$wbase..HEAD"
git -C "$repo" reset -q --hard HEAD~1
apply() { run apply PATCH_DIR="$1" WORKTREE="$work" EXPECT_NAME=Y4NN77 EXPECT_EMAIL=fixer@example.invalid; }
apply "$pd"
who="$(git -C "$work" log -1 --format='%an <%ae> %cn <%ce> %s')"
[ "$rc" -eq 0 ] && [ "$who" = "Y4NN77 <fixer@example.invalid> Y4NN77 <fixer@example.invalid> fix(solve): z" ] && ok "apply: the trusted job commits the patch as the human identity, subject kept" || ko "apply identity (rc=$rc $out who=$who)"
run check BASE_SHA="$wbase" WORKTREE="$work" EXPECT_NAME=Y4NN77 EXPECT_EMAIL=fixer@example.invalid
[ "$rc" -eq 0 ] && ok "apply: the applied commits pass the guard" || ko "apply then check (rc=$rc $out)"
git -C "$work" reset -q --hard "$wbase"
mkdir -p "$tmp/bad1"; for n in $(seq -w 1 21); do cp "$pd"/0001-*.patch "$tmp/bad1/00$n-x.patch"; done
apply "$tmp/bad1"
[ "$rc" -ne 0 ] && grep -qi 'too many' <<<"$out" && [ "$(git -C "$work" rev-parse HEAD)" = "$wbase" ] && ok "apply: more than 20 patch files are refused, nothing applied" || ko "too many (rc=$rc $out)"
mkdir -p "$tmp/bad2"; cp "$pd"/0001-*.patch "$tmp/bad2/evil.sh"
apply "$tmp/bad2"
[ "$rc" -ne 0 ] && grep -qi 'name' <<<"$out" && ok "apply: a file name that is not NNNN-slug.patch is refused" || ko "bad name (rc=$rc $out)"
mkdir -p "$tmp/bad3"; head -c 1100000 /dev/zero | tr '\0' 'a' >"$tmp/bad3/0001-big.patch"
apply "$tmp/bad3"
[ "$rc" -ne 0 ] && grep -qi 'large' <<<"$out" && ok "apply: more than 1 MB of patches is refused" || ko "too large (rc=$rc $out)"
mkdir -p "$tmp/bad4"
apply "$tmp/bad4"
[ "$rc" -ne 0 ] && grep -qi 'no patch' <<<"$out" && ok "apply: an empty patch directory is refused" || ko "empty (rc=$rc $out)"

# ---- the command wrapper of the fixer (`run`): read-only git, go test, the web scripts; never an output file ----
mkdir -p "$tmp/wbin"
for b in git go pnpm unshare; do printf '#!/bin/sh\necho "RAN %s $*"\n' "$b" >"$tmp/wbin/$b"; chmod +x "$tmp/wbin/$b"; done
wrap() { rc=0; out="$(env -u FIXER_NETNS PATH="$tmp/wbin:$PATH" python3 -I "$tool" run "$@" 2>&1)" || rc=$?; }
allowed() { wrap "$@"; [ "$rc" -eq 0 ] && grep -q "^RAN $1" <<<"$out" && ok "wrapper runs: $*" || ko "wrapper should run: $* (rc=$rc $out)"; }
refused() { wrap "$@"; [ "$rc" -ne 0 ] && ! grep -q '^RAN' <<<"$out" && ok "wrapper refuses: $*" || ko "wrapper should refuse: $* (rc=$rc $out)"; }
allowed git diff --stat
allowed git diff -- src/a.go
allowed git diff --cached --name-only
allowed git status --short
allowed git log --oneline -5
allowed git show HEAD:src/a.go
allowed go test ./src/api/... -run TestX -count=1
allowed go test -run TestX -count 1 -race -v -timeout 60s ./src/api/...
allowed go test -tags=integration ./src/api/internal/...
allowed pnpm --filter web run test
allowed pnpm --filter web run test -- src/features/x.test.tsx
refused go test -test.coverprofile=/tmp/c ./...
refused go test ./... -args -test.outputdir /tmp
refused go test -ldflags=-extld=sh ./...
refused go test -gcflags=all=-N ./...
refused go test -modfile=/tmp/m ./...
refused go test /abs/pkg
refused go test -run
refused git diff --ext-diff
refused git diff --no-index /etc/passwd src/a.go
refused git log --format=%H
refused pnpm --filter web run test --reporter=verbose
wrapnet() { rc=0; out="$(env FIXER_NETNS=1 PATH="$tmp/wbin:$PATH" python3 -I "$tool" run "$@" 2>&1)" || rc=$?; }
wrapnet go test ./src/api/...
grep -q '^RAN unshare -rn go test ./src/api/...' <<<"$out" && ok "wrapper: with FIXER_NETNS=1, go test runs without network (unshare -rn)" || ko "netns ($out)"
wrapnet git status
grep -q '^RAN git --no-pager status' <<<"$out" && ok "wrapper: git reads are not wrapped in a network namespace" || ko "netns git ($out)"
refused git diff --output=/tmp/x
refused git diff --output /tmp/x
refused git show -o /tmp/x HEAD
refused git push
refused git -c core.pager=sh diff
refused go test -o /tmp/x ./...
refused go test -exec sh ./...
refused go test -coverprofile=/tmp/c ./...
refused go build ./...
refused pnpm --filter web exec sh
refused pnpm --filter web run test --outputFile=/tmp/x
refused pnpm install
refused rm -rf src
refused
wrap git diff --stat
grep -q '^RAN git --no-pager diff --stat' <<<"$out" && ok "wrapper: git runs without a pager" || ko "no pager ($out)"

# ---- the batch pull request body ----
items='#12,#20/n1,#20/n2,#31'
run body ISSUE=40 ITEMS="$items" RESULT_JSON='{"outcome":"fixed","reason":"ok","results":[{"item":"#12","verdict":"FIXED","reason":"guard | added @Y4NN777"},{"item":"#20/n1","verdict":"maybe","reason":"x"},{"item":"#20/n2","verdict":"fixed elsewhere","reason":"by #150"},{"item":"#99","verdict":"FIXED","reason":"not in the batch"}]}'
grep -qx 'Batch: #40' <<<"$out" && ok "body: a Batch line names the batch issue" || ko "body batch line ($out)"
grep -qx 'Sources: #12 #20 #31' <<<"$out" && ok "body: a Sources line references every source issue once" || ko "body sources ($out)"
grep -q '| #12 | #12 | FIXED | guard \\| added' <<<"$out" && ok "body: one table row per item (item, source, decision, reason)" || ko "body row ($out)"
grep -q '| #20/n2 | #20 | FIXED ELSEWHERE | by #' <<<"$out" && ok "body: FIXED ELSEWHERE is a decision" || ko "body elsewhere ($out)"
grep -q '| #31 | #31 | OWNER | no result from the fixer |' <<<"$out" && ok "body: an item without a result is listed as OWNER" || ko "body missing item ($out)"
! grep -q '#99' <<<"$out" && ok "body: a result for an item outside the batch is dropped" || ko "body extra item ($out)"
! grep -q '@Y4NN777' <<<"$out" && ok "body: an @-mention from the model is neutralised" || ko "body mention ($out)"
grep -q '| #20/n1 | #20 | OWNER |' <<<"$out" && ok "body: an unknown verdict becomes OWNER" || ko "body verdict ($out)"
! grep -qi 'generated with\|co-authored' <<<"$out" && ok "body: no AI attribution line" || ko "body attribution"

run body ISSUE=40 ITEMS="$items" RESULT_JSON='not json'
[ "$rc" -eq 0 ] && grep -qx 'Batch: #40' <<<"$out" && grep -qx 'Sources: #12 #20 #31' <<<"$out" && grep -q '| #31 | #31 | OWNER | no result from the fixer |' <<<"$out" && ok "body: an unreadable result still lists every source, as OWNER" || ko "body bad json (rc=$rc $out)"

run body ISSUE=40 ITEMS='#12' RESULT_JSON='{"outcome":"fixed","reason":"closes #7","results":[{"item":"#12","verdict":"FIXED","reason":"fixes #5 and resolves #6, see owner/repo#8"}]}'
! grep -qiE '(close|fix|resolve)[a-z]*:? +#[0-9]' <<<"$out" && ! grep -qE '#[5678]' <<<"$out" && ok "body: closing keywords and references written by the model are neutralised" || ko "body keywords ($out)"
# H6 (independent review of #187): every issue reference form is neutralised, not only #N: GH-N and issue or pull URLs.
run body ISSUE=40 ITEMS='#12' RESULT_JSON='{"outcome":"fixed","reason":"closes gh-3","results":[{"item":"#12","verdict":"FIXED","reason":"fixes GH-6 and resolves https://github.com/o/r/issues/9, closes https://github.com/o/r/pull/10"}]}'
! grep -qiE 'gh-[0-9]' <<<"$out" && ! grep -qE 'github\.com/[^ ]*/(issues|pull)/' <<<"$out" && ok "body: GH-N and issue or pull request URLs written by the model are neutralised" || ko "body GH-N and URLs ($out)"
# Review of 8541e46 (note-disarm-github-com): only issue and pull request URLs are disarmed; any other github.com link
# (a file, a commit, an action) stays a working link.
run body ISSUE=40 ITEMS='#12' RESULT_JSON='{"outcome":"fixed","reason":"ok","results":[{"item":"#12","verdict":"FIXED","reason":"see https://github.com/o/r/blob/develop/src/a.go and https://github.com/o/r/commit/abc1234, closes https://github.com/o/r/issues/9"}]}'
grep -q 'https://github.com/o/r/blob/develop/src/a.go' <<<"$out" && grep -q 'https://github.com/o/r/commit/abc1234' <<<"$out" && ! grep -qE 'github\.com/[^ ]*/issues/' <<<"$out" && ok "body: a github.com link that is not an issue or pull request URL is kept intact" || ko "body plain link ($out)"


# ---- the context the fixer gets (detail, fix, permalink), written to a file by the trusted script ----
# Single issues are routed by number: $ST/issue-N.json (the fake above knows only the fixed fixtures, so this
# section replaces it with a router that also serves them).
python3 -I - "$tmp/bin/gh" <<'EOF'
import sys
p = sys.argv[1]
s = open(p).read()
s = s.replace('elif "/actions/runs/" in path: name = "jobs.json"',
              'elif "/actions/runs/" in path: name = "jobs.json"\nelif __import__("re").search(r"/issues/\\d+$", path): name = "issue-%s.json" % path.rsplit("/", 1)[1]')
open(p, "w").write(s)
EOF
sha_a=3333333333333333333333333333333333333333
tok="gh""p_$(printf 'a1B2%.0s' 1 2 3 4 5 6 7 8 9)"
python3 -I - "$tmp/st/comments.json" "$HEAD" "$sha_a" "$tok" <<'EOF'
import json, sys
p, head, first, tok = sys.argv[1:5]
findings = [
  {"id": "f1", "kind": "finding", "evidence": "OBSERVED", "sev": "MEDIUM", "file": "src/a.go", "line": 3, "scenario": "s1", "status": "open",
   "detail": "proof one @someone <!-- x -->", "fix": "fix one", "source": "review", "first_sha": first, "location": "verified"},
  {"id": "f2", "kind": "finding", "evidence": "OBSERVED", "sev": "MEDIUM", "file": "src/b.go", "line": 4, "scenario": "s2", "status": "open",
   "detail": "proof two", "fix": "rotate " + tok, "source": "review", "first_sha": first, "location": "verified"},
  {"id": "f3", "kind": "finding", "evidence": "OBSERVED", "sev": "MEDIUM", "file": "src/c.go", "line": 5, "scenario": "s3", "status": "open"}]
body = "Triage: MEDIUM\n<!-- 226os-ai-review -->\nTool: PASS (1 files)\n\n<details><summary>Review ledger</summary>\n\n```json\n" + \
       json.dumps({"sha": head, "findings": findings}) + "\n```\n\n</details>"
json.dump([{"user": {"login": "github-actions[bot]"}, "updated_at": "2026-11-14T10:00:00Z", "html_url": "https://github.com/o/r/pull/7#issuecomment-5", "body": body}], open(p, "w"))
EOF
run context MODE=pr FINDINGS=f1,f2,f3 HEAD_SHA="$HEAD"
ctx_py() { python3 -I -c "import json,sys; d=json.loads(sys.stdin.read()); $1" <<<"$out"; }
[ "$rc" -eq 0 ] && ctx_py 'it={i["id"]:i for i in d["items"]}; assert sorted(it)==["f1","f2","f3"]; f=it["f1"]; assert f["fix"]=="fix one" and "proof one" in f["detail"] and f["permalink"]=="https://github.com/o/r/blob/'$sha_a'/src/a.go#L3" and f["comment"]=="https://github.com/o/r/pull/7#issuecomment-5" and f["source"]=="review"' && ok "context (pr): the planned ids, each with detail, fix, permalink, source and the review comment" || ko "context pr (rc=$rc $out)"
ctx_py 'it={i["id"]:i for i in d["items"]}; assert "@someone" not in it["f1"]["detail"] and "<!--" not in it["f1"]["detail"] and "'"$tok"'" not in str(d) and "redacted" in it["f2"]["fix"]' && ok "context (pr): neutralised and redacted like the issues" || ko "context safety ($out)"
ctx_py 'f={i["id"]:i for i in d["items"]}["f3"]; assert f["detail"]=="not recorded (older review)" and f["fix"]=="not recorded (older review)" and not f.get("permalink")' && ok "context (pr): an older entry says not recorded (older review)" || ko "context old ($out)"
run context MODE=pr FINDINGS=f1 HEAD_SHA=9999999999999999999999999999999999999999
ctx_py 'assert d["items"]==[] and "head" in d["note"]' && ok "context (pr): a ledger for another head gives no items (the stale ledger is not passed)" || ko "context stale ($out)"
run context MODE=pr FINDINGS=f1 HEAD_SHA="$HEAD" FAKE_GH_FAIL=1
[ "$rc" -eq 0 ] && ctx_py 'assert d["items"]==[] and "NOT RUN" in d["note"]' && ok "context: a GitHub error is NOT RUN with exit 0 and no items (the fixer still has the ids)" || ko "context fail (rc=$rc $out)"

# batch: the sources of the row (also-sources included) come from the issues, boxes from their notes issue.
python3 -I - "$tmp/st" "$tok" <<'EOF'
import json, os, sys
st, tok = sys.argv[1:3]
def put(n, body, title="t", labels=()):
    json.dump({"number": n, "state": "open", "title": title, "body": body, "labels": list(labels)}, open(os.path.join(st, "issue-%d.json" % n), "w"))
put(40, "<!-- review-batch -->\n# Review batch\n\n- [ ] #12 x <!-- batch-also:#13,#20/n1 --> <!-- batch-item:#12 -->\n- [ ] #31 y <!-- batch-item:#31 -->\n", labels=["review-batch"])
put(12, "<!-- finding:PR90#f12 -->\n- Where: `a.go:3`\n- Detail: the proof for twelve\n- Leak: " + tok + "\n- Fix: fix twelve <!-- evil -->\n")
put(13, "<!-- finding:PR95#f12 -->\n- Where: `a.go:3`\n- Detail: the proof for thirteen\n")
put(20, "<!-- finding:PR91#notes -->\n- [ ] `n1` [OBSERVED] `a.go:3` remark one\n  - detail: note proof\n  - fix: none: accept\n- [ ] `n2` [OBSERVED] `b.go:2` other\n  - detail: other proof\n")
put(31, "<!-- finding:PR93#f31 -->\n- Detail: not in this row\n")
EOF
run context MODE=batch ISSUE=40 ITEMS='#12'
ctx_py 'it={i["key"]:i for i in d["items"]}; assert sorted(it)==["#12","#13","#20/n1"], sorted(it); assert "proof for twelve" in it["#12"]["text"] and "proof for thirteen" in it["#13"]["text"] and "note proof" in it["#20/n1"]["text"] and "other proof" not in it["#20/n1"]["text"] and "not in this row" not in str(d)' && ok "context (batch): each planned item and the other sources of its row, a note by its box only" || ko "context batch (rc=$rc $out)"
ctx_py 'assert "'"$tok"'" not in str(d) and "<!--" not in str(d)' && ok "context (batch): redacted and neutralised" || ko "context batch safety ($out)"
run context MODE=batch ISSUE=40 ITEMS='#12' FAKE_GH_FAIL=1
[ "$rc" -eq 0 ] && ctx_py 'assert d["items"]==[] and "NOT RUN" in d["note"]' && ok "context (batch): a GitHub error is NOT RUN, exit 0" || ko "context batch fail ($out)"

# The fixer jobs check the context file against the hash written by the trusted step before the model starts.
for wf in review-autofix review-batch; do
  grep -q 'steps.ctx.outputs.sha' "$root/.github/workflows/$wf.yml" && grep -q 'sha256sum' "$root/.github/workflows/$wf.yml" && ok "$wf.yml: the finding context is hash-checked before the model" || ko "$wf.yml context hash"
done

# Review of PR 187: the workflows run the TRUSTED copy of the script in a directory of its own ($RUNNER_TEMP/bin). It loads
# findings.py and review-context.py from that directory, so each workflow must put them there from the same commit.
for wf in review-autofix review-batch; do
  python3 -I - "$root/.github/workflows/$wf.yml" <<'EOF2' && ok "$wf.yml: the context step's directory gets review-autofix, findings and review-context from git objects of the trusted ref" || ko "$wf.yml modules beside the script"
import re, sys
s = open(sys.argv[1]).read()
m = re.search(r"for module in ([\w. -]+); do\n\s*git cat-file blob \"[^\"]+:scripts/\$module\.py\" >\"\$RUNNER_TEMP/bin/\$module\.py\"\n\s*done", s)
assert m and set(m.group(1).split()) == {"review-autofix", "findings", "review-context"}, "modules not extracted beside the script"
EOF2
done
# The layout of the step: the script alone in a directory finds no module (the fixer's context is empty); with the two modules
# beside it, the same command gives the items.
lay="$tmp/layout-alone"; mkdir -p "$lay"; cp "$root/scripts/review-autofix.py" "$lay/"
run_in_layout() {
  env PATH="$tmp/bin:$PATH" ST="$tmp/st" REPO=o/r PR_NUMBER=7 AUTO_FIX=on HAS_TOKEN=true MODE=pr FINDINGS=f1 HEAD_SHA="$HEAD" python3 -I "$1/review-autofix.py" context 2>&1 || true
}
alone="$(run_in_layout "$lay")"
python3 -I -c 'import json,sys; d=json.loads(sys.argv[1]); assert d["items"]==[] and "NOT RUN" in d["note"]' "$alone" && ok "context: the script alone in its directory gives an empty context (why the modules must sit beside it)" || ko "alone layout ($alone)"
lay2="$tmp/layout-full"; mkdir -p "$lay2"; cp "$root/scripts/review-autofix.py" "$root/scripts/findings.py" "$root/scripts/review-context.py" "$lay2/"
full="$(run_in_layout "$lay2")"
python3 -I -c 'import json,sys; d=json.loads(sys.argv[1]); assert [i["id"] for i in d["items"]]==["f1"] and d["items"][0]["fix"]=="fix one"' "$full" && ok "context: with findings.py and review-context.py beside the script, the fixer gets its items" || ko "full layout ($full)"


# ---- review notes of #178 (PR170): a trailer in a patch, hand-edited sqlc output ----
r3="$tmp/r3"
git init -q -b develop "$r3"
g3() { git -C "$r3" -c user.name=Y4NN77 -c user.email=fixer@example.invalid "$@"; }
mkdir -p "$r3/src/api/internal/platform/dbgen" "$r3/db/queries"
echo a >"$r3/src/a.go"; echo q >"$r3/db/queries/x.sql"; echo g >"$r3/src/api/internal/platform/dbgen/build.sql.go"; echo t >"$r3/src/api/internal/platform/dbgen/build_queries_test.go"
g3 add -A; g3 commit -qm base
base3="$(git -C "$r3" rev-parse HEAD)"
check3() { run check BASE_SHA="$base3" WORKTREE="$r3" EXPECT_NAME=Y4NN77 EXPECT_EMAIL=fixer@example.invalid; }
echo hand >>"$r3/src/api/internal/platform/dbgen/build.sql.go"; g3 add -A; g3 commit -qm "fix(build): edit generated"
check3
[ "$rc" -ne 0 ] && grep -qi 'generated' <<<"$out" && ok "guard (note-dbgen-not-owner): sqlc output changed without a db/queries change is refused" || ko "dbgen alone (rc=$rc $out)"
echo q2 >>"$r3/db/queries/x.sql"; g3 add -A; g3 commit -qm "fix(build): query"
check3
[ "$rc" -eq 0 ] && ok "guard: sqlc output regenerated together with db/queries is allowed" || ko "dbgen with queries (rc=$rc $out)"
g3 reset -q --hard "$base3"
echo more >>"$r3/src/api/internal/platform/dbgen/build_queries_test.go"; g3 add -A; g3 commit -qm "test: dbgen test"
check3
[ "$rc" -eq 0 ] && ok "guard: a hand-written test in the dbgen directory is still allowed" || ko "dbgen test (rc=$rc $out)"

# A trailer the action adds (settings not passed through) is dropped by the trusted apply step: the fix is not lost.
w3="$tmp/w3"
git clone -q "$r3" "$w3"
wb3="$(git -C "$w3" rev-parse HEAD)"
echo fix >>"$r3/src/a.go"
git -C "$r3" -c user.name=claude -c user.email=x@y commit -qam "fix(solve): trailer

Real body line.

Generated with Claude Code

Co-Authored-By: Claude <noreply@anthropic.com>"
pd3="$tmp/pd3"; git -C "$r3" format-patch -q -o "$pd3" "$wb3..HEAD"
run apply GITHUB_WORKFLOW="autofix-test" GITHUB_RUN_ID=4242 PATCH_DIR="$pd3" WORKTREE="$w3" EXPECT_NAME=Y4NN77 EXPECT_EMAIL=fixer@example.invalid
msg="$(git -C "$w3" log -1 --format=%B)"
[ "$rc" -eq 0 ] && grep -q 'Real body line' <<<"$msg" && ! grep -qi 'co-authored\|generated with' <<<"$msg" && ok "apply (note settings-flag-passthrough): an AI trailer or Generated with line in a patch message is dropped, the body is kept" || ko "apply trailer (rc=$rc $out $msg)"
# Review of PR 185: every commit the trusted apply step makes carries a provenance trailer (never an AI attribution), so
# review-automerge.py never treats it as owner-authored; it is set from the runner's workflow and run, never by the model.
tr="$(git -C "$w3" log -1 --format=%B | grep '^Review-Pipeline: ' || true)"
[ "$tr" = "Review-Pipeline: autofix-test/4242" ] && ok "apply: the commit carries Review-Pipeline: <workflow>/<run_id> from GITHUB_WORKFLOW and GITHUB_RUN_ID" || ko "pipeline trailer ($tr)"
run check BASE_SHA="$wb3" WORKTREE="$w3" EXPECT_NAME=Y4NN77 EXPECT_EMAIL=fixer@example.invalid
[ "$rc" -eq 0 ] && ok "apply: the commit with the dropped trailer and the Review-Pipeline trailer passes the guard (the fix is not lost)" || ko "apply then check trailer (rc=$rc $out)"


# #212 HMAC: with REVIEW_PIPELINE_HMAC_KEY (the publish job's environment secret) every commit also carries a signature of
# its tree, its parent, the workflow and the run id; without the key there is none (auto-merge then never exempts it).
! git -C "$w3" log -1 --format=%B | grep -q '^Review-Pipeline-Sig:' && ok "apply: without REVIEW_PIPELINE_HMAC_KEY the commit carries no signature" || ko "sig without key"
w4="$tmp/whmac"
git clone -q "$r3" "$w4"
git -C "$w4" reset -q --hard "$wb3"
run apply GITHUB_WORKFLOW="autofix-test" GITHUB_RUN_ID=4242 REVIEW_PIPELINE_HMAC_KEY=k3y PATCH_DIR="$pd3" WORKTREE="$w4" EXPECT_NAME=Y4NN77 EXPECT_EMAIL=fixer@example.invalid
sig="$(git -C "$w4" log -1 --format=%B | sed -n 's/^Review-Pipeline-Sig: //p')"
want="$(python3 -I -c 'import hmac,hashlib,sys; print(hmac.new(sys.argv[1].encode(), "\n".join(["226os-review-pipeline", sys.argv[2], sys.argv[3], "autofix-test", "4242"]).encode(), hashlib.sha256).hexdigest())' k3y "$(git -C "$w4" rev-parse 'HEAD^{tree}')" "$(git -C "$w4" rev-parse 'HEAD^')")"
[ "$rc" -eq 0 ] && [ -n "$sig" ] && [ "$sig" = "$want" ] && ok "apply: with the key the commit carries Review-Pipeline-Sig = HMAC-SHA256 of tree, parent, workflow and run id" || ko "sig (rc=$rc sig=$sig want=$want out=$out)"
! grep -q 'k3y' <<<"$out" && ! git -C "$w4" log -1 --format=%B | grep -q 'k3y' && ok "apply: the key never appears in the output or the message" || ko "key leak"


# H6 (independent review of #187): a commit message from the fixer lands on develop, the default branch, where GitHub acts
# on closing keywords: every issue reference form in the subject and body is neutralised by the trusted apply step.
w4="$tmp/w4"
git clone -q "$r3" "$w4"
echo refs >>"$r3/src/a.go"
git -C "$r3" -c user.name=claude -c user.email=x@y commit -qam "fix(solve): closes #5 and gh-6

Resolves owner/repo#7, fixes https://github.com/o/r/issues/8 and closes https://github.com/o/r/pull/9.
Context: https://github.com/o/r/blob/develop/src/a.go"
pd4="$tmp/pd4"; git -C "$r3" format-patch -q -o "$pd4" HEAD~1..HEAD
run apply GITHUB_WORKFLOW="autofix-test" GITHUB_RUN_ID=4243 PATCH_DIR="$pd4" WORKTREE="$w4" EXPECT_NAME=Y4NN77 EXPECT_EMAIL=fixer@example.invalid
msg="$(git -C "$w4" log -1 --format=%B)"
[ "$rc" -eq 0 ] && ! grep -qE '#[0-9]' <<<"$msg" && ! grep -qiE 'gh-[0-9]' <<<"$msg" && ! grep -qE 'github\.com/[^ ]*/(issues|pull)/' <<<"$msg" && grep -q '^Review-Pipeline: autofix-test/4243$' <<<"$msg" && ok "apply: issue references (#N, GH-N, owner/repo#N, issue and pull URLs) in a fixer commit message are neutralised, the trailer kept" || ko "apply refs (rc=$rc $out $msg)"
grep -q '^Context: https://github.com/o/r/blob/develop/src/a.go$' <<<"$msg" && ok "apply: a github.com link that is not an issue or pull request URL is kept intact in the commit message" || ko "apply plain link ($msg)"

finish
