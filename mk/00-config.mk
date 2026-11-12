# Shared configuration (ADR-0006). Included first by the Makefile.

# Go toolchain: exactly go.mod's `go` line. `make doctor` fails if go.mod and this value disagree.
export GOTOOLCHAIN := go1.27.4
# Gates never rewrite go.mod/go.sum: a missing or outdated dependency record is a failure.
export GOFLAGS := -mod=readonly

# Evidence directories (git-ignored); CI uploads them as artifacts.
export REPORTS_DIR := reports
COVERAGE_DIR := coverage

# The make binary used by the sequential quality driver. Held in a plain variable so that
# `make -n quality` only PRINTS the driver line instead of executing it.
MAKE_BIN := $(MAKE)

# Immutability base for scripts/check-migrations.sh. Set it in CI to origin/<target branch>; when
# empty, `make lint` falls back to origin/develop then origin/main if they exist and says so.
MIGRATIONS_BASE_REF ?=

# Profile used by `make migrate-up` (scripts/migrate-up.sh gates anything but local-dev).
MIGRATE_PROFILE ?= local-dev

# Fail-closed marker for targets that are not implemented yet. Usage: $(call NOT_IMPLEMENTED,<target>,<reason>)
# Never exits 0, never prints PASS.
define NOT_IMPLEMENTED
	@echo "NOT IMPLEMENTED: make $(1)" >&2
	@echo "NOT RUN: $(2)" >&2
	@exit 1
endef
