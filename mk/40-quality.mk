# Aggregates (ADR-0006). A sequential driver (scripts/quality.sh) runs every step through make,
# records each exit code, prints a final PASS/FAIL summary and exits non-zero if any step failed.
# Nothing is masked: no `make -k`, no ignored exit codes. QUALITY_FAIL_FAST=1 stops at the first failure and
# lists the remaining steps as NOT RUN. Logs and the summary go to $(REPORTS_DIR)/.
#
# QUALITY_STEPS is a plain variable so a test can inject steps on the command line in a scratch copy;
# CI never overrides it. The default list is G1 to G4: doctor generate lint typecheck test-unit coverage
# test-integration security licenses.

QUALITY_STEPS := doctor generate lint typecheck test-unit coverage test-integration security licenses

quality: ## G1 to G4 aggregate: every step runs, per-step PASS/FAIL summary, non-zero if any step failed
	@scripts/quality.sh $(MAKE_BIN) quality $(QUALITY_STEPS)

quality-full: ## quality plus test-e2e and perf (G5, G6); fails by design while those are NOT IMPLEMENTED
	@scripts/quality.sh $(MAKE_BIN) quality-full $(QUALITY_STEPS) test-e2e perf

# Parallel CI groups (ADR-0010). GitHub Actions runs one job per group (`make ci-<group>`), each through the same
# fail-closed driver, and a final job checks that every step below has a PASS for the tested commit. The groups
# cover EXACTLY the leaf steps of `quality` (lint, test-integration and security expanded into their parts):
# scripts/tests/make_test.sh compares `make ci-steps` with `make quality-leaf-steps`, so a step added to one and
# forgotten in the other fails the test. `make quality` stays the sequential local definition of done.
# The `repo` group is the only one that also runs for a documentation-only change (scripts/ci-scope.sh); a gate that
# starts reading a documentation path must be added to the `full` cases there.
CI_GROUPS := static lint-go lint-contract lint-web lint-scripts unit integration-go integration-containers security images repo
CI_STEPS_static := doctor generate typecheck
CI_STEPS_lint-go := doctor lint-go
CI_STEPS_lint-contract := doctor lint-contract
CI_STEPS_lint-web := doctor lint-web
CI_STEPS_lint-scripts := doctor lint-scripts
CI_STEPS_unit := doctor test-unit coverage
CI_STEPS_integration-go := doctor test-integration-go
CI_STEPS_integration-containers := doctor test-integration-containers
CI_STEPS_security := doctor security-authz security-govulncheck security-sast security-js-audit security-sast-js
CI_STEPS_images := doctor security-images
CI_STEPS_repo := doctor security-secrets-encrypted security-secrets licenses

QUALITY_LEAF_STEPS = $(foreach s,$(QUALITY_STEPS),$(if $(filter $(s),lint),$(LINT_STEPS),$(if $(filter $(s),test-integration),$(TEST_INTEGRATION_STEPS),$(if $(filter $(s),security),$(SECURITY_STEPS),$(s)))))

ci-groups: ## The CI groups, one parallel job each (ADR-0010)
	@echo $(CI_GROUPS)

# SCOPE_GROUPS narrows the list to some groups (CI passes `repo` for a documentation-only change).
ci-steps: ## Every step the CI groups run, de-duplicated (the aggregator expects a PASS for each); SCOPE_GROUPS=<groups> narrows it
	@echo $(sort $(foreach g,$(or $(SCOPE_GROUPS),$(CI_GROUPS)),$(CI_STEPS_$(g))))

quality-leaf-steps: ## The leaf steps of `quality` (lint, test-integration and security expanded)
	@echo $(sort $(QUALITY_LEAF_STEPS))

ci-%: FORCE ## One CI group, e.g. `make ci-lint-go`: its steps through the fail-closed driver, summary in reports/
	@scripts/quality.sh $(MAKE_BIN) ci-$* $(CI_STEPS_$*)

FORCE:

test-e2e: ## G5: Playwright P0 flows, FR/EN, mobile, axe (NOT IMPLEMENTED: later slice)
	$(call NOT_IMPLEMENTED,test-e2e,Playwright end-to-end tests are a later slice (G5))

perf: ## G6: reproducible dataset, EXPLAIN plans, API latency, page weight (NOT IMPLEMENTED: later slice)
	$(call NOT_IMPLEMENTED,perf,performance baseline is a later slice (G6))
