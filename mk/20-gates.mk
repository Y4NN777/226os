# G1 to G3 gates (ADR-0006). Every target requires `doctor` first and fails on the first failing command.
# All tools are pinned: Go tools through `go tool` (go.mod `tool` block), JS tools through
# pnpm-lock.yaml, ShellCheck by image digest (scripts/lint-shell.sh).

generate: doctor ## G1: run oapi-codegen, openapi-typescript, sqlc, then fail on drift from the index
	scripts/check-generated.sh check
	scripts/check-sqlc.sh

lint: doctor ## G1: gofmt, vet, staticcheck, golangci-lint, ESLint, Prettier, OpenAPI lint, migration checks, ShellCheck, script tests
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
	@echo "== lint: sqlc vet"
	go tool sqlc vet -f db/sqlc.yaml
	@echo "== lint: ESLint"
	pnpm --filter web run lint
	@echo "== lint: Prettier check"
	pnpm --filter web run format:check
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
	@echo "== lint: ShellCheck (pinned image)"
	scripts/lint-shell.sh
	@echo "== lint: script tests (no Docker flags)"
	@for t in scripts/tests/*_test.sh; do echo "-- $$t"; "$$t"; done

typecheck: doctor ## G1: tsc --noEmit (strict) for web, go build ./...
	pnpm --filter web run typecheck
	go build ./...

test-unit: doctor ## G2: go test -race, Vitest
	go test -race -count=1 ./...
	pnpm --filter web run test

coverage: doctor ## G2: Go first-party statements exactly 100.00%, web above 70% on four metrics; reports under coverage/
	scripts/coverage-go.sh $(COVERAGE_DIR)
	pnpm --filter web run coverage --coverage.reportsDirectory=$(abspath $(COVERAGE_DIR))/web

test-integration: doctor ## G3: real ephemeral PostgreSQL 18 (go test -tags integration), script tests with Docker, container verification (needs Docker)
	go test -race -count=1 -tags integration ./...
	@# lint already ran every script test; only db-reset_test.sh has an extra Docker pass.
	@echo "-- scripts/tests/db-reset_test.sh (DB_RESET_TEST_DOCKER=1)"; DB_RESET_TEST_DOCKER=1 scripts/tests/db-reset_test.sh
	scripts/verify-containers.sh
