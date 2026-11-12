# Discoverability and environment (ADR-0006: doctor, setup).

help: ## List the Make targets
	@echo "226os Make targets (ADR-0006). Read a recipe with 'make -n <target>' before its first execution."
	@grep -hE '^[a-zA-Z0-9_-]+:.*## ' $(sort $(MAKEFILE_LIST)) | awk -F ':.*## ' '{printf "  %-20s %s\n", $$1, $$2}'

doctor: ## Verify pinned toolchain (Go, Node, pnpm), Docker and Compose; prerequisite of the other targets
	@scripts/doctor.sh

setup: doctor ## Install from lockfiles only: go mod download, pnpm install --frozen-lockfile (needs execution approval)
	go mod download
	pnpm install --frozen-lockfile
