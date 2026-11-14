# G1 to G3 gates (ADR-0006). Every target requires `doctor` first and fails on the first failing command.
# All tools are pinned: Go tools through `go tool` (go.mod `tool` block), JS tools through
# pnpm-lock.yaml, ShellCheck by image digest (scripts/lint-shell.sh).

generate: doctor ## G1: run oapi-codegen, openapi-typescript, sqlc, then fail on drift from the index
	scripts/check-generated.sh check
	scripts/check-sqlc.sh

# `lint` is the sum of four independent groups; CI runs them as parallel jobs (ADR-0010), `make lint` runs them
# one after the other and stops at the first failure.
LINT_STEPS := lint-go lint-contract lint-web lint-scripts

lint: $(LINT_STEPS) ## G1: gofmt, vet, staticcheck, golangci-lint, sqlc vet, OpenAPI lint, migration checks, ESLint, Prettier, ShellCheck, script tests

lint-go: doctor ## G1: gofmt, go vet, staticcheck, golangci-lint (default and integration build tags)
	@echo "== lint: gofmt"
	@unformatted="$$("$$(go env GOROOT)/bin/gofmt" -l src)"; if [ -n "$$unformatted" ]; then echo "gofmt: unformatted files:" >&2; echo "$$unformatted" >&2; exit 1; fi
	@echo "== lint: go vet (default and integration build tags)"
	go vet ./...
	go vet -tags integration ./...
	@echo "== lint: staticcheck"
	go tool staticcheck ./...
	go tool staticcheck -tags integration ./...
	@echo "== lint: golangci-lint"
	go tool golangci-lint run ./...
	go tool golangci-lint run --build-tags integration ./...

lint-contract: doctor ## G1: sqlc vet, OpenAPI lint (redocly), migration policy
	@echo "== lint: sqlc vet"
	go tool sqlc vet -f db/sqlc.yaml
	@echo "== lint: OpenAPI (redocly)"
	pnpm exec redocly lint api/openapi.yaml
	@echo "== lint: migration policy (forward-only, header, immutability)"
	@base="$(MIGRATIONS_BASE_REF)"; \
	if [ -z "$$base" ]; then \
		for c in origin/develop origin/main; do \
			if git rev-parse --verify --quiet "$$c^{commit}" >/dev/null; then base="$$c"; break; fi; \
		done; \
	fi; \
	if [ -z "$$base" ]; then echo "NOTE: no MIGRATIONS_BASE_REF and no origin/develop or origin/main: immutability against the base branch is NOT RUN (the other migration rules run)"; fi; \
	MIGRATIONS_BASE_REF="$$base" scripts/check-migrations.sh

lint-web: doctor ## G1: ESLint, Prettier check
	@echo "== lint: ESLint"
	pnpm --filter web run lint
	@echo "== lint: Prettier check"
	pnpm --filter web run format:check

lint-scripts: doctor ## G1: ShellCheck (pinned image), script tests (no Docker flags), tracker moves with the slice
	@echo "== lint: ShellCheck (pinned image)"
	scripts/lint-shell.sh
	@echo "== lint: script tests (no Docker flags)"
	@for t in scripts/tests/*_test.sh; do echo "-- $$t"; "$$t"; done
	@echo "== lint: tracker moves with the slice"
	scripts/check-todo-updated.sh

docs-check: doctor ## G1: documentation checks (relative links, no private paths, operation count, ADR index; on a pull request into main also: AI_USAGE.md is not stale)
	python3 -I scripts/check-docs.py
	python3 -I scripts/ai-usage.py check --only-on-promotion

typecheck: doctor ## G1: tsc --noEmit (strict) for web, go build ./...
	pnpm --filter web run typecheck
	go build ./...

test-unit: doctor ## G2: go test -race, Vitest
	go test -race -count=1 ./...
	pnpm --filter web run test

coverage: doctor ## G2: Go first-party statements exactly 100.00%, web above 70% on four metrics; reports under coverage/
	scripts/coverage-go.sh $(COVERAGE_DIR)
	pnpm --filter web run coverage --coverage.reportsDirectory=$(abspath $(COVERAGE_DIR))/web

# `test-integration` is the sum of two independent groups (parallel jobs in CI, ADR-0010).
TEST_INTEGRATION_STEPS := test-integration-go test-integration-containers

test-integration: $(TEST_INTEGRATION_STEPS) ## G3: real ephemeral PostgreSQL 18 (go test -tags integration), script tests with Docker, container verification (needs Docker)

test-integration-go: doctor ## G3: go test -race -tags integration against an ephemeral PostgreSQL 18 (needs Docker)
	go test -race -count=1 -tags integration ./...

test-integration-containers: doctor ## G3: the Docker pass of the script tests, then container verification (needs Docker)
	@# lint already ran every script test; only db-reset_test.sh has an extra Docker pass.
	@echo "-- scripts/tests/db-reset_test.sh (DB_RESET_TEST_DOCKER=1)"; DB_RESET_TEST_DOCKER=1 scripts/tests/db-reset_test.sh
	scripts/verify-containers.sh
