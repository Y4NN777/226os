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

test-e2e: ## G5: Playwright P0 flows, FR/EN, mobile, axe (NOT IMPLEMENTED: later slice)
	$(call NOT_IMPLEMENTED,test-e2e,Playwright end-to-end tests are a later slice (G5))

perf: ## G6: reproducible dataset, EXPLAIN plans, API latency, page weight (NOT IMPLEMENTED: later slice)
	$(call NOT_IMPLEMENTED,perf,performance baseline is a later slice (G6))
