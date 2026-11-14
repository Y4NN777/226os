# Build and operations targets (ADR-0006). Anything that starts, stops or alters a runtime goes through
# scripts/compose.sh (APP_ENV must equal the profile, `down`/`rm` never take -v/--volumes) or a guarded script.

build: doctor ## G7 input: multi-stage images api, web, migrate tagged with the git SHA; digests in reports/image-digests.txt
	scripts/build-images.sh

# --- profiles: local-dev (developer machine), staging-test (develop), production-live (main) -------------
up-local-dev: doctor ## Start the local-dev profile (builds the images; the one-shot migrate job exits 0, so no --wait)
	scripts/compose.sh local-dev up -d --build

down-local-dev: doctor ## Stop the local-dev profile (containers and networks; volumes are kept)
	scripts/compose.sh local-dev down

up-staging-test: doctor ## Start the staging-test profile from pinned images (requires .env.staging-test, never committed)
	scripts/compose.sh staging-test up -d --wait

down-staging-test: doctor ## Stop the staging-test profile (volumes are kept)
	scripts/compose.sh staging-test down

smoke: doctor ## Post-deploy smoke (read-only): make smoke PROFILE=local-dev|staging-test|production-live [BASE_URL=https://<host>]
	@[ -n "$(PROFILE)" ] || { echo "smoke: PROFILE is required (local-dev, staging-test or production-live)" >&2; exit 64; }
	scripts/smoke.sh "$(PROFILE)" $(BASE_URL)

# up-production-live is never run from a developer machine (ADR-0006). Guard (_guard-production-up): the
# documented deploy context is an automated release job, i.e. CI=true, no interactive terminal on stdin, and
# an explicit PRODUCTION_UP_CONFIRM=production-live. scripts/compose.sh then still requires APP_ENV=production-live
# in the env file (no CHANGE_ME, real PUBLIC_HOST and ACME_EMAIL). There is no deploy workflow yet (TASK-OPS-001),
# so today no context satisfies the guard. It is a protection against mistakes, not against a determined caller.
up-production-live: _guard-production-up doctor ## Start production-live (refused outside the documented deploy context)
	scripts/compose.sh production-live up -d --wait

_guard-production-up:
	@if [ "$${CI-}" != true ] || [ "$${PRODUCTION_UP_CONFIRM-}" != production-live ] || [ -t 0 ]; then \
		echo "REFUSED: up-production-live runs only from the automated release job (CI=true, no interactive terminal, PRODUCTION_UP_CONFIRM=production-live); never from a developer machine" >&2; \
		exit 2; \
	fi

down-production-live: doctor ## Stop the production-live profile (volumes are kept)
	scripts/compose.sh production-live down

# --- database ------------------------------------------------------------------------------------------
# Forward-only goose (pinned through go.mod). migrate-up delegates to scripts/migrate-up.sh: local-dev runs the
# Compose `migrate` job; any other profile needs MIGRATE_CONFIRM=<profile> and BACKUP_EVIDENCE=<file>.
migrate-up: doctor ## Apply forward-only migrations (MIGRATE_PROFILE=local-dev by default; gated otherwise)
	scripts/migrate-up.sh $(MIGRATE_PROFILE)

migrate-status: doctor ## Show goose migration status for DATABASE_URL (read-only; the URL is passed through the environment)
	@if [ -z "$${DATABASE_URL-}" ]; then echo "REFUSED: migrate-status needs DATABASE_URL in the environment (no env file is read)" >&2; exit 2; fi
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING="$$DATABASE_URL" go tool goose -dir db/migrations status

migrate-new: ## Create the next sequential forward-only migration: make migrate-new NAME=<slug>
	scripts/migrate-new.sh "$(NAME)"

seed: ## Demo seeds, local-dev only (NOT IMPLEMENTED: db/seeds does not exist yet)
	$(call NOT_IMPLEMENTED,seed,db/seeds does not exist yet; when it does it may only run with APP_ENV=local-dev or staging-test)

# db-reset drops and recreates the database. The target checks APP_ENV itself and scripts/db-reset.sh repeats
# that check and additionally requires a local Compose database host; the target must never weaken the script.
db-reset: _guard-db-reset doctor ## Drop and recreate the local database (APP_ENV=local-dev and a local DATABASE_URL only)
	scripts/db-reset.sh

_guard-db-reset:
	@if [ "$${APP_ENV-}" != local-dev ]; then \
		echo "REFUSED: db-reset only runs with APP_ENV=local-dev (got '$${APP_ENV-}')" >&2; \
		exit 2; \
	fi
