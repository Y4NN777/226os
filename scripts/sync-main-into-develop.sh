#!/usr/bin/env bash
# Step 6 of the promotion flow (docs/GIT_WORKFLOW.md), automated: after main moves (a promotion or an emergency deploy),
# open the back-merge pull request main -> develop so develop never stays behind main. Without it, the required check
# `Validate branch promotion` (scripts/check-develop-contains-main.sh) refuses every other pull request into develop.
#
#   develop already contains main         -> NO-OP
#   otherwise                             -> branch chore/platform/sync-main-into-develop = main (created or moved),
#                                            pull request into develop (reused when one is open), auto-merge armed
#                                            with a MERGE COMMIT (never squash or rebase: the promotion SHA must stay)
# The branch name is the one check-develop-contains-main.sh accepts. Fails closed: a missing variable, a GitHub error
# or a non-numeric answer stops with FAILED (a run that cannot sync must be visible, never a silent green).
#
# Environment: GITHUB_REPOSITORY (owner/name), GH_TOKEN (a token whose pull requests trigger CI: REVIEW_AGENT_TOKEN).
set -euo pipefail

BRANCH="chore/platform/sync-main-into-develop"
need() { [ -n "${!1:-}" ] || { printf 'sync-main-into-develop: BLOCKED: %s is not set\n' "$1" >&2; exit 1; }; }
fail() { printf 'sync-main-into-develop: FAILED: %s\n' "$*" >&2; exit 1; }
need GITHUB_REPOSITORY
need GH_TOKEN
repo="$GITHUB_REPOSITORY"

ahead="$(gh api "repos/$repo/compare/develop...main" --jq .ahead_by 2>/dev/null)" || fail "cannot compare develop and main"
[[ "$ahead" =~ ^[0-9]+$ ]] || fail "unexpected compare answer"
if [ "$ahead" -eq 0 ]; then
  echo "sync-main-into-develop: NO-OP (develop already contains main)"
  exit 0
fi

sha="$(gh api "repos/$repo/git/ref/heads/main" --jq .object.sha 2>/dev/null)" || fail "cannot read main"
[[ "$sha" =~ ^[0-9a-f]{40}$ ]] || fail "unexpected main SHA"

# The branch is a pure mirror of main: create it, or move it (force) when it exists from an earlier promotion.
if gh api "repos/$repo/git/ref/heads/$BRANCH" --jq .object.sha >/dev/null 2>&1; then
  gh api "repos/$repo/git/refs/heads/$BRANCH" -X PATCH -f sha="$sha" -F force=true >/dev/null \
    || fail "cannot move $BRANCH to main (a token without the Workflows permission cannot point a branch at workflow changes)"
else
  gh api "repos/$repo/git/refs" -f ref="refs/heads/$BRANCH" -f sha="$sha" >/dev/null \
    || fail "cannot create $BRANCH at main (a token without the Workflows permission cannot point a branch at workflow changes)"
fi

pr="$(gh pr list --repo "$repo" --base develop --head "$BRANCH" --state open --json number --jq '.[0].number // empty')" \
  || fail "cannot list pull requests"
if [ -z "$pr" ]; then
  pr="$(gh pr create --repo "$repo" --base develop --head "$BRANCH" \
    --title "chore(platform): rétro-fusion de main dans develop" \
    --body "Étape 6 du flux de promotion (docs/GIT_WORKFLOW.md), ouverte automatiquement après la mise à jour de \`main\` (\`${sha:0:7}\`). La branche est \`main\` telle quelle : aucun fichier modifié. Fusion **par merge commit** uniquement, jamais squash ni rebase. Tant qu'elle manque, \`Validate branch promotion\` refuse les autres PR vers \`develop\`.")" \
    || fail "cannot open the pull request"
fi
gh pr merge --repo "$repo" --auto --merge "$pr" >/dev/null || fail "cannot arm auto-merge on $pr"
echo "sync-main-into-develop: back-merge $pr armed (merge commit) for main ${sha:0:7}"
