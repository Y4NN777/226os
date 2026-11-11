---
name: tdd-unit
description: RED/GREEN/REFACTOR unit-test discipline for 226os Go (testing, table-driven, -race, fuzz) and React (Vitest + Testing Library) with the G2 coverage gates. Use when touching **/*_test.go, **/*.test.ts, **/*.test.tsx, **/*.spec.ts, **/*.spec.tsx, or when writing code that needs a failing test first.
---
# TDD and Unit Tests (G2)

## Scope and triggers
- Any unit test file: `**/*_test.go`, `**/*.test.ts(x)`, `**/*.spec.ts(x)`; any production change in `apps/api/**` or `apps/web/src/**` (it needs a RED test first).
- G2 evidence: Go statement coverage, frontend 4-metric coverage, race tests.
- Out of scope: real-DB tests (use `integration-testing`), browser journeys (use `bdd-acceptance`).

## Project rules (binding)
- The loop is mandatory. RED means recorded proof of a failing test for absent behaviour or a bug. GREEN means the minimal fix, with results recorded. REFACTOR means tests stay green and invariants stay validated. `docs/TEST.md:29-31`, `docs/RULES.md:18`, `docs/DEV.md:36-37`.
- Never delete or edit a test only to hide a failure. Changing an expected value needs a business justification and a human reviewer. `docs/TEST.md:32`, `AGENTS.md:40`, `AGENTS.md:74`.
- Go G2 means 100.00% of instrumentable first-party statements, measured with `go test -covermode=atomic -coverprofile=coverage.out ./...` plus `go tool cover -func`. The gate fails below 100.00% "sans arrondi trompeur" (no misleading rounding). `docs/TEST.md:6`, `AGENTS.md:48`.
- Frontend G2 means statements, branches, functions and lines are each strictly above 70% under Vitest/V8. Configure 71 as the minimum. One failing metric fails the gate. `docs/TEST.md:7`, `docs/RULES.md:19`.
- Exclusions are allowed only through the reviewed list in `docs/quality/COVERAGE_SCOPE.md`. No exclusion is active yet (`:5`). A coverage report without the reviewed scope file is `BLOCKED`, not `PASS` (`:21`). Never exclude a business branch or a hard endpoint. `docs/TEST.md:8`.
- 100% statement coverage is not branch proof. Permissions, validation, transactions, auth and critical invariants need a positive/negative branch matrix plus integration tests. `docs/TEST.md:9`, `docs/quality/COVERAGE_SCOPE.md:11`.
- Run `go test -race`. Pin `gofmt`/`go vet`/`staticcheck`/`golangci-lint` versions. `docs/DEV.md:46`.
- Frontend uses Vitest + Testing Library. Snapshots are never the only proof. `docs/DEV.md:55`.
- Coverage is reported per commit SHA, with the report, scope and exclusions. No "100%" badge without a real report. `docs/TEST.md:10`.
- Business logic lives outside HTTP handlers, so unit-test the domain/service, not the handler. `docs/DEV.md:10`, `docs/DEV.md:44`.
- Run commands only through reviewed Make targets (`make test-unit`, `make coverage`) after reading their recipes. `AGENTS.md:120-121`.

## Expert practices
- List the test cases first, then take one at a time through red, green, refactor. Do not skip refactor, which is the most common TDD failure. [Fowler](https://martinfowler.com/bliki/TestDrivenDevelopment.html)
- Prove RED for the right reason. The failure message must name the missing behaviour, not a compile error or a nil-pointer panic in setup.
- Write table-driven tests with named cases and `t.Run`. Call `t.Parallel()` only when cases share no mutable state. Capture loop variables explicitly if the module's `go` directive is below 1.22 (UNVERIFIED for the pinned toolchain; check `go.mod`). [Go wiki: TableDrivenTests](https://go.dev/wiki/TableDrivenTests) (UNVERIFIED: not fetched)
- Prefer hand-written fakes that implement the boundary interface over mocking frameworks. Assert on observable outputs and state, not call order.
- Use `t.Cleanup`, `t.TempDir` and `t.Setenv` instead of manual teardown. Never sleep. Inject a clock and an ID generator.
- `-race` finds only races that actually run, so give concurrent code tests that exercise the concurrency. Expect 2-20x slowdown. [go.dev race detector](https://go.dev/doc/articles/race_detector)
- Fuzz parsers and validators (Markdown sanitisation input, GitHub URL normalisation, cursors). The seed corpus (`f.Add`, `testdata/fuzz/`) runs as a normal test under `go test`. Commit minimised crashers as regressions. [go.dev fuzzing](https://go.dev/doc/security/fuzz/)
- `go tool cover -func` prints `total:` with `%.1f`, so 99.96% shows as `100.0%`. Gate on zero uncovered blocks in `coverage.out`, not on the printed percentage. [cmd/cover/func.go](https://raw.githubusercontent.com/golang/go/master/src/cmd/cover/func.go)
- Keep the command fixed at `docs/TEST.md:6`. Since Go 1.22, `go test -cover` counts packages without test files at 0%, so they cannot silently drop out, provided the pinned toolchain is 1.22 or later. Do not add `-coverpkg`: it credits a package with other packages' tests, which deviates from TEST.md:6 and would need an ADR. [Go 1.22 release notes](https://go.dev/doc/go1.22)
- Vitest: set `coverage.provider: 'v8'` and an explicit `coverage.include` over the app source. Without it, only imported files are measured and untested files vanish. Thresholds are minimums, treated as `>=`, hence 71. Never enable `autoUpdate`. [Vitest coverage config](https://vitest.dev/config/coverage)
- Testing Library: query by role, label or text the way a user does. Use `userEvent` over `fireEvent`. Assert on accessible output. [Testing Library guiding principles](https://testing-library.com/docs/guiding-principles) (UNVERIFIED: not fetched)
- Stub network at the boundary of the generated OpenAPI client. Never hand-write DTOs (`docs/DEV.md:51`).

## Patterns
```go
func TestAcceptAnswer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		setup   func(*fakeRepo)
		actor   UserID
		wantErr error
	}{
		{"question author accepts", nil, alice, nil},
		{"non-author denied", nil, bob, ErrForbidden}, // SPEC.md:24-28
		{"anonymous denied", nil, anon, ErrUnauthenticated},
		{"second accept conflicts", preAccepted, alice, ErrConflict}, // one accepted answer, SPEC.md:110
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo := newFakeRepo(t) // in-memory fake, no shared state
			if tc.setup != nil {
				tc.setup(repo)
			}
			svc := NewService(repo, fixedClock)
			err := svc.Accept(context.Background(), tc.actor, qID, aID)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Accept() err = %v, want %v", err, tc.wantErr)
			}
		})
	}
}
```
```sh
# PROPOSED recipe body for `make coverage` (fail closed, no rounding)
set -euo pipefail
EXCL=coverage-excludes.txt            # literal path prefixes mirrored from COVERAGE_SCOPE.md only
[ -r "$EXCL" ] || { echo "missing $EXCL"; exit 1; }   # no scope file = BLOCKED, never pass
go test -race -covermode=atomic -coverprofile=coverage.out ./...
# Sum counts per block (a block may repeat), list blocks never executed.
awk 'NR>1 {c[$1]+=$NF} END {for (b in c) if (c[b]==0) print b}' coverage.out > uncovered.all
rc=0; grep -vF -f "$EXCL" uncovered.all > uncovered.txt || rc=$?
[ "$rc" -le 1 ] || { echo "grep error $rc"; exit 1; }   # 1 = no lines selected; >1 = error
[ ! -s uncovered.txt ] || { cat uncovered.txt; exit 1; } # any uncovered block = FAIL
# Note: an empty $EXCL makes grep -f match nothing, so -v keeps every line (correct).
```
```ts
// vitest.config.ts (excerpt)
coverage: {
  provider: 'v8',
  include: ['src/**/*.{ts,tsx}'],
  exclude: [/* only entries mirrored in docs/quality/COVERAGE_SCOPE.md */],
  thresholds: { statements: 71, branches: 71, functions: 71, lines: 71 },
  reporter: ['text', 'html', 'json-summary'],
}
```
```tsx
it('shows the accept action only to the question author', async () => {
  render(<QuestionView question={q} viewer={bob} />);
  expect(screen.queryByRole('button', { name: /accepter/i })).not.toBeInTheDocument();
});
```

## Anti-patterns (reject in review)
- Code first, test afterwards, or RED evidence missing from the PR.
- Edited assertions, `t.Skip`, `it.skip`/`.only`, or lowered thresholds that turn a failure green (`AGENTS.md:74`).
- Coverage exclusion comments or config entries absent from `COVERAGE_SCOPE.md`.
- Trusting `total: 100.0%` text. Missing `coverage.include`. A coverage gate that passes when the excludes file is missing.
- Tests that execute lines without asserting anything, written just for coverage.
- Mocks asserting internal call sequences. Mocking the unit under test.
- `time.Sleep` / `setTimeout` waits. Real clock, random seed or network in unit tests.
- `t.Parallel()` with shared fakes or package globals. A race-detector failure "fixed" by removing `-race`.
- Snapshot-only component tests (`docs/DEV.md:55`). Querying by CSS class or test ID when a role exists.
- Business logic tested only through HTTP handlers.

## Done checklist
- [ ] Each AC maps to at least one test named after the behaviour; story ID in the PR (`AGENTS.md:57-70`).
- [ ] RED evidence recorded: command, failing test name, exit code ≠ 0, SHA.
- [ ] GREEN evidence recorded: same command, exit code 0, SHA.
- [ ] Refactor done; suite still green; no test or threshold weakened.
- [ ] `make test-unit` (includes `-race`): exit code, log path.
- [ ] `make coverage`: Go zero uncovered first-party blocks; frontend all 4 metrics ≥71. Report paths (`coverage.out`, HTML, `coverage-summary.json`) tied to the SHA.
- [ ] `docs/quality/COVERAGE_SCOPE.md` reviewed version referenced; otherwise G2 = `BLOCKED`.
- [ ] Authz/validation/transaction branches have a positive/negative matrix, with integration coverage linked (`docs/TEST.md:9`).
- [ ] Evidence record filled (`docs/spec-kit/QUALITY_RELEASE.md:28-39`). Anything not executed = `NOT RUN`, never `PASS` (`AGENTS.md:31`).

## References
- `docs/TEST.md`, `docs/quality/COVERAGE_SCOPE.md`, `docs/DEV.md`, `docs/RULES.md`, `AGENTS.md`
- https://martinfowler.com/bliki/TestDrivenDevelopment.html
- https://go.dev/doc/articles/race_detector · https://go.dev/doc/security/fuzz/ · https://go.dev/doc/go1.22
- https://vitest.dev/config/coverage · https://testing-library.com/docs/guiding-principles (UNVERIFIED)
