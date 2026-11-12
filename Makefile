# 226os Makefile: the single entrypoint for the quality gates and local operations (ADR-0006).
#
# This file only sets the shell semantics, the target list and the include order. All recipes live in
# mk/*.mk (included in lexical order):
#   mk/00-config.mk    pinned versions, shared variables, the NOT IMPLEMENTED helper
#   mk/10-env.mk       help, doctor, setup
#   mk/20-gates.mk     generate, lint, typecheck, test-unit, coverage, test-integration
#   mk/30-security.mk  security (G4 checks) and licenses
#   mk/40-quality.mk   quality, quality-full, test-e2e, perf
#   mk/50-ops.mk       build, up-/down-<profile>, migrate-*, seed, db-reset
#
# Fail-closed rules (checked by scripts/tests/make_test.sh): a missing tool or an unimplemented target exits
# non-zero with NOT IMPLEMENTED / NOT RUN; no ignored-error shortcuts; no `-` prefixed recipe lines; nothing prints
# PASS without having run the check; tool versions are pinned; no download at gate time beyond the locked
# dependency fetch and the documented vulnerability-database lookups (see mk/30-security.mk).
# Make targets are not intrinsically safe: read a recipe (make -n <target>) before its first execution.

SHELL := bash
.SHELLFLAGS := -euo pipefail -c
.DELETE_ON_ERROR:
.DEFAULT_GOAL := help

.PHONY: help doctor setup \
	generate lint typecheck test-unit coverage test-integration \
	security security-authz security-secrets-encrypted security-govulncheck security-js-audit security-secrets security-sast security-sast-js security-images \
	licenses test-e2e perf quality quality-full \
	build up-local-dev down-local-dev up-staging-test down-staging-test up-production-live down-production-live \
	_guard-production-up migrate-up migrate-status migrate-new seed db-reset _guard-db-reset

include $(sort $(wildcard mk/*.mk))
