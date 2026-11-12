---
name: ci-cd-delivery
description: 226os delivery pipeline (G1-G7 wiring). Covers fail-closed Make targets, least-privilege SHA-pinned GitHub Actions, SAST/SCA/secrets/licences, non-root multi-stage images, Compose + Traefik TLS with a private DB, backup/restore, rollback, and develop->main promotion of the tested build. Use when touching Makefile, .github/workflows/**, infra/**, docker/**, **/Dockerfile, compose*.yml.
---
# CI/CD and Delivery

## Scope and triggers
- `Makefile`, `mk/**`, `.github/workflows/**`, `compose.yaml`, `deploy/**`, `**/Dockerfile`, `.env.example`, release/rollback runbooks.
- Ops and platform slices (S8 and platform work) are read from their block in `TODO.md` with owner/reviewer per ADR-0008. G0-OPS-002 (Datacloud hosting, DNS, registry, backup, approver) is ACCEPTED (ADR-0006), but nothing is deployed: no deploy workflow exists yet. `.github/workflows/` holds `quality.yml` and `branch-policy.yml`. G5 and G6 are NOT IMPLEMENTED as targets, G7 is NOT RUN. Any deployment job stays `NOT IMPLEMENTED` until the owner opens it.

## Project rules (binding)
- The approved interface is the real Make targets (`make help` lists them): `doctor|generate|lint|typecheck|test-unit|coverage|test-integration|security|licenses|quality|build`, with the Make files under `mk/`. Read each recipe before running it, and prefer a single `make quality` that proves all gates ran (`AGENTS.md`). GitHub Actions and Make are the only success path (`docs/archive/spec-kit/CONTRACTS.md:59`). CI runs the gates as parallel groups (ADR-0010).
- G5 (`make test-e2e`) and G6 (`make perf`) exist as targets in `mk/40-quality.mk` but are NOT IMPLEMENTED and fail by design; never report them as PASS or skip them tacitly.
- Gates fail closed: no step may be skipped tacitly, and there is no exception for an exposed secret or a critical vulnerability (`docs/TEST.md:52`). `NOT RUN` != `PASS` (`AGENTS.md:31`, `AGENTS.md:126`).
- Use minimal `permissions`, OIDC where possible, and never expose SSH keys or tokens to fork PRs. Production requires environment approval, and nothing auto-deploys from a PR. `docs/SHIP.md:21-22`. The in-repo pattern is at `.github/workflows/branch-policy.yml:9-10` (`contents: read`) and `:18-22` (untrusted refs passed via `env`, not inlined).
- Scan secrets and dependencies on every PR, plus a runtime licence report. A high or critical exploitable vulnerability blocks release. `docs/TEST.md:38`, `docs/RULES.md:26`. The G4 steps are the `security-*` targets in `mk/30-security.mk` and `make licenses`.
- Linter versions are pinned in the toolchain files (ADR-0006); run sqlc drift. `docs/DEV.md:46-47`. G1 covers format/lint/types/reproducible OpenAPI+sqlc generation. `AGENTS.md:47`.
- Images: multi-stage, non-root where possible, pinned versions and digests, healthchecks, resource limits, structured logs, secrets via secure config. `docs/SHIP.md:9`, `docs/RULES.md:31`, `docs/archive/spec-kit/CONTRACTS.md:56-57`.
- Topology: separate dev and prod Compose files; Traefik is the only public edge with Let's Encrypt TLS; the DB is never public. `docs/SHIP.md:6-7`, `docs/architecture/ARCHITECTURE.md:24`. Expose `/healthz` (liveness) and `/readyz` (readiness). `docs/SHIP.md:32`.
- Flow: work branch -> PR develop -> staging smoke -> PR main -> approval -> backup -> migrations -> deploy -> smoke -> monitor. `docs/SHIP.md:13-15`, `docs/GIT_WORKFLOW.md:51-56`.
- `branch-policy.yml:27-41` enforces that work branches target develop and that only `develop` or `hotfix/*` may target main.
- Keep the tested SHA unchanged during promotion and do not rebuild a different revision. `AGENTS.md:22`, `docs/SHIP.md:20`.
- Hotfix: branch from `main`, open a PR into `main` with checks and approval, then back-merge to `develop` by PR. `docs/GIT_WORKFLOW.md:60`.
- Rollback: redeploy the previous image by digest. An app rollback does not undo a migration, so use the tested restore plan. Encrypted backups and a verified restore come first. `docs/SHIP.md:38-40`, `docs/SHIP.md:27-28`.
- Branch protection: required checks, dismiss stale approvals, no force-push/delete, one human approval on `main`, `production` environment approval. `docs/GIT_WORKFLOW.md:77-83`.
- No deploy, cloud mutation, `docker compose down -v`, `sudo`, `curl | sh` or secret access without explicit approval. `AGENTS.md:85`, `AGENTS.md:125`.
- TENSION (resolve by ADR before the first release): GitHub's PR merge methods are merge commit, squash and rebase. Squash and rebase create new SHAs, and there is no fast-forward-only PR option. [GitHub merge methods](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/about-merge-methods-on-github)
- TENSION: `docs/GIT_WORKFLOW.md:70` allows "squash or merge", but squash cannot preserve the tested SHA. Never squash or rebase develop->main; use a merge commit.
- TENSION: `docs/GIT_WORKFLOW.md:82` (require linear history) blocks merge commits. Turn it off on `main`, or no PR method can satisfy both rules.
- TENSION: `docs/SHIP.md:20` says "exact develop SHA", while `docs/SHIP.md:15` and `docs/GIT_WORKFLOW.md:56` say "exact main SHA". With a merge commit these are different SHAs. Proposed: deploy the image digest built from the tested develop SHA, and check that `main^{tree}` equals that commit's tree.

## Expert practices
- Set workflow-level `permissions: contents: read` and raise permissions per job only where needed. [GitHub secure use](https://docs.github.com/en/actions/security-for-github-actions/security-guides/security-hardening-for-github-actions)
- Pin every third-party action to a full 40-char commit SHA from the action's own repo, with a `# vX.Y.Z` comment. Update via Dependabot or Renovate. GitHub now mentions workflow dependency lockfiles as superseding SHA pins (UNVERIFIED: availability). [GitHub secure use](https://docs.github.com/en/actions/security-for-github-actions/security-guides/security-hardening-for-github-actions)
- Never interpolate `${{ github.event.* }}` / `head_ref` into `run:`. Pass the value through `env:`. Avoid `pull_request_target`, and never check out PR code under it. [GitHub secure use](https://docs.github.com/en/actions/security-for-github-actions/security-guides/security-hardening-for-github-actions)
- Use OIDC instead of long-lived cloud secrets where the provider supports it. Deploy secrets belong only in the `production`/`staging` environments. [GitHub secure use](https://docs.github.com/en/actions/security-for-github-actions/security-guides/security-hardening-for-github-actions)
- Self-audit against the OpenSSF Scorecard checks: Token-Permissions, Pinned-Dependencies, Dangerous-Workflow, Branch-Protection, SAST, Vulnerabilities, Dependency-Update-Tool, Signed-Releases, SBOM. [Scorecard checks](https://github.com/ossf/scorecard/blob/main/docs/checks.md)
- Cache Go modules/build and the package-manager store keyed on lockfile hashes. Never cache coverage or test results used as gate evidence (`AGENTS.md:126`). Run `concurrency` per ref with cancel-in-progress on PRs.
- Make: start recipes with `SHELL := bash` and `.SHELLFLAGS := -euo pipefail -c`, mark targets `.PHONY`, add `.DELETE_ON_ERROR`, and make `quality` depend on every gate target. A missing tool is a failure, never a skip. (GNU make manual, UNVERIFIED: not fetched)
- Dockerfile: use multi-stage, a minimal or distroless runtime, `FROM image:tag@sha256:<digest>`, `USER` with an explicit non-zero UID, a `.dockerignore`, and dependency layers before source. [Docker best practices](https://docs.docker.com/build/building/best-practices/)
- Build once. Build the image on the tested `develop` SHA, label it with `org.opencontainers.image.revision=<sha>`, push it by digest, and deploy that same digest to staging and production. Never rebuild on `main`.
- Traefik: the HTTP-01 challenge needs :80 reachable, TLS-ALPN-01 needs :443, and DNS-01 is required for wildcards. Persist ACME storage on a protected volume. Redirect HTTP to HTTPS and add security headers via middleware. Expose containers only through labels, with `exposedByDefault=false` (UNVERIFIED option name for the pinned version). [Traefik ACME](https://doc.traefik.io/traefik/reference/install-configuration/tls/certificate-resolvers/acme/)
- Compose prod: put the DB on an `internal: true` network with no `ports:`. Use `healthcheck` + `depends_on: condition: service_healthy`, `read_only`/`cap_drop: [ALL]` where feasible, mem/cpu limits, and secrets as files rather than env literals (UNVERIFIED per pinned Compose spec).
- Backup/restore: take a `pg_dump -Fc` (or a physical backup per ADR), encrypt it, store it off-host, and rotate it. Prove restore on schedule into a scratch DB and run smoke queries before calling the system recoverable. Run migrations as a separate one-shot job after the backup.

## Patterns
```yaml
# .github/workflows/ci.yml (excerpt)
permissions: { contents: read }
concurrency: { group: ci-${{ github.ref }}, cancel-in-progress: ${{ github.event_name == 'pull_request' }} }
jobs:
  quality:
    runs-on: ubuntu-latest   # pin runner image per ADR
    steps:
      - uses: actions/checkout@<40-char-sha> # vX.Y.Z
        with: { persist-credentials: false }
      - run: make quality
      - uses: actions/upload-artifact@<40-char-sha> # vX.Y.Z
        if: always()
        with: { name: evidence-${{ github.sha }}, path: reports/ }
```
```dockerfile
FROM golang:<ver>@sha256:<digest> AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/api ./src/api/cmd/api
FROM <runtime-base>@sha256:<digest>   # minimal non-root base, per ADR
COPY --from=build /out/api /api
USER 10001:10001
ENTRYPOINT ["/api"]
```
```sh
# promotion check (release operator, read-only): main must carry exactly the tested tree
test "$(git rev-parse origin/main^{tree})" = "$(git rev-parse "$TESTED_SHA^{tree}")"
```

## Anti-patterns (reject in review)
- `uses: owner/action@v4` or `@main`. Workflow-level `write-all`. Secrets in PR jobs. `pull_request_target` with checkout.
- `|| true`, `continue-on-error: true`, or `-` prefixed Make lines on gate steps. A Make target that echoes PASS without running.
- Rebuilding images on `main`. `latest` tags. Deploying by tag instead of digest. Auto-deploy to prod on merge without approval.
- Root containers, `privileged`, a published DB port, or Traefik dashboard exposed without auth.
- `.env` with real values committed. Secrets in build args, layers, logs or artifacts.
- "Rollback" that assumes the down-migration restores data (`docs/SHIP.md:38`). Unverified backups.
- Force-push or history rewrite on `develop`/`main` (`docs/GIT_WORKFLOW.md:69`).

## Done checklist
- [ ] Every gate target fails non-zero on failure or a missing tool (show one deliberate RED run).
- [ ] All actions SHA-pinned, `permissions` minimal per job, no untrusted interpolation.
- [ ] SAST, SCA, secret scan and licence report produced per PR, with tools chosen by ADR. Paths recorded.
- [ ] Image: non-root, digest-pinned bases, healthcheck, revision label. Digest recorded.
- [ ] Compose prod validated (`docker compose -f … config` via a reviewed target). DB has no public port. TLS verified on staging.
- [ ] Backup taken, restore drill done, rollback-by-digest rehearsed, with evidence paths.
- [ ] Promotion record: tested SHA, image digest, main SHA plus tree-equality check, approver, smoke results (`docs/SHIP.md:31-35`).
- [ ] Each G1-G7 reported as `PASS/FAIL/NOT RUN/BLOCKED` with command, exit code, report path and SHA. Deploy stays `BLOCKED` until the G7 evidence exists (Datacloud account, DNS, approver, backup and rollback drills).

## References
- `AGENTS.md`, `docs/SHIP.md`, `docs/GIT_WORKFLOW.md`, `docs/RULES.md`, `docs/TEST.md`, `.github/workflows/branch-policy.yml`, `docs/archive/spec-kit/CONTRACTS.md`
- https://docs.github.com/en/actions/security-for-github-actions/security-guides/security-hardening-for-github-actions
- https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/about-merge-methods-on-github
- https://github.com/ossf/scorecard/blob/main/docs/checks.md · https://docs.docker.com/build/building/best-practices/
- https://doc.traefik.io/traefik/reference/install-configuration/tls/certificate-resolvers/acme/
