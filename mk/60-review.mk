# Optional AI code review (open-code-review, pinned in scripts/review.sh, run through npx, never installed globally).
# It sends the changed files to the LLM provider you configure: this is advisory help for the human review required by
# ADR-0008, not a gate and not part of `make quality`. Project rules: .opencodereview/rule.json.

review-login: ## One-time: sign in with ChatGPT for the local AI review (a preview of the tool, one person, local only)
	scripts/review.sh login

review: ## Advisory AI review of HEAD against origin/develop (REVIEW_BASE=<ref> to change it); sends the diff to your LLM provider
	scripts/review.sh run
