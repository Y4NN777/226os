# 226os Third-Party Notices

**Status:** Policy in force (ADR-0001, 2026-11-11). Direct-dependency inventory recorded 2026-11-12 (below); legal and CADEV review PENDING. Every licence marked `to verify` is unverified.

## Licence of original code

Original 226os code is licensed under Apache-2.0 (`LICENSE`, `NOTICE`). Third-party components keep their own licences, listed below.

## Policy

Every direct and transitive runtime/build dependency must be pinned in a lockfile, identified by package and version, mapped to an SPDX license, and reviewed against CADEV rules before release. Do not copy application source from reference repositories. References in `docs/research/REFERENCES.md` are research inputs, not declared dependencies.

The final inventory must include notices, source/license links, modification status, attribution requirements, and any reciprocal-license or trademark obligations. Vulnerability and license reports are evidence-bearing artifacts tied to the target SHA.

## Planned inventory categories

- Frontend: React, TypeScript, Vite, TanStack packages, Meta Astryx, styling and test tooling.
- Backend: Go runtime libraries, Chi, pgx, sqlc tooling, migration and test tooling.
- Infrastructure: Docker/Traefik images and action dependencies.
- Reference material: Apache Answer, Flarum, Nuxt HackerNews, 225 Open Source, and other research sources; no code integration is asserted by this file.

## Current status

The Go module file, the npm manifests and `pnpm-lock.yaml` now exist, so the "no lockfile" statement of earlier revisions no longer holds. The inventory below covers **direct** dependencies and base images only. The complete, regenerable inventory is the output of `make licenses` (`scripts/licenses.sh`), which writes `reports/licenses/` (git-ignored, tied to the target SHA); it covers Go modules and npm production dependencies, not container-image OS packages. Release compliance stays `BLOCKED` until the entries marked `to verify` below are verified (G4/G7 evidence). `make licenses` result for the whole tree: NOT RUN by the author of this record; the figures below were read from an existing local `reports/licenses/` output (allowed list Apache-2.0, BSD-2-Clause, BSD-3-Clause, ISC, MIT, Unlicense; 47 dependencies checked; result PASS), not re-run.

## Go direct modules (`go.mod`, runtime `require` block)

| Module | Version | Licence | Basis |
|---|---|---|---|
| `github.com/getkin/kin-openapi` | v0.149.0 | MIT | `reports/licenses/go-licenses.csv` |
| `github.com/go-chi/chi/v5` | v5.5.0 | MIT | `reports/licenses/go-licenses.csv` |
| `github.com/jackc/pgx/v5` | v5.13.0 | MIT | `reports/licenses/go-licenses.csv` |
| `github.com/oapi-codegen/runtime` | v1.9.0 | Apache-2.0 | `reports/licenses/go-licenses.csv` |
| `github.com/pressly/goose/v3` | v3.30.0 | MIT | LICENSE header in the local Go module cache (not in the generated report) |
| `go.yaml.in/yaml/v3` | v3.0.6 | to verify | not in the generated report; v3.0.6 not in the local module cache |

Transitive Go modules are listed in `reports/licenses/` (`make licenses`) and are not itemised here.

## npm direct dependencies (`package.json`, `src/web/package.json`)

Exact versions are pinned in the manifests and `pnpm-lock.yaml`.

- Runtime (`src/web`): `@astryxdesign/core` (Meta Astryx), `@stylexjs/stylex`, `@tanstack/react-form`, `@tanstack/react-query`, `@tanstack/react-router`, `openapi-fetch`, `react`, `react-dom`.
- Tooling (root): `@redocly/cli`, `openapi-typescript`, `typescript`. Tooling (`src/web`): `@astryxdesign/cli`, ESLint and plugins, Vite and its React plugin, Vitest and coverage, Testing Library, jsdom, Prettier, `typescript-eslint`, `globals`, `@types/*`. These are build and test tools, not shipped in the web image.
- Licence summary recorded in `docs/quality/ASTRYX_REPORT.md` (`pnpm licenses list --filter web --prod`, 30 production packages): MIT 28; BSD-3-Clause 1 (`intl-messageformat`, pulled by Astryx core); Unlicense 1 (`isbot`, pulled by TanStack Router). Licences of the development tools are `to verify` (not covered by that summary).
- **Owner decision 2026-11-12:** `isbot` (Unlicense, a public-domain dedication) is accepted. Unlicense is on the allowed list of `scripts/check-licenses.py`, so `make licenses` passes it. Residual risk: the CADEV organiser has not ruled on it.

## Container base images (`compose.yaml`, `src/api/Dockerfile`, `src/web/Dockerfile`)

Tags and digests are those pinned in the files. Licences of image contents are `to verify` unless stated.

| Image (tag@digest prefix) | Purpose | Licence |
|---|---|---|
| `postgres:18@sha256:e3d24f69...` | database (compose) | to verify |
| `axllent/mailpit@sha256:aef2aacc...` | local-dev mail catcher | MIT (as recorded in `AI_USAGE.md`, S0 mailer record) |
| `traefik:v3.6@sha256:31267173...` | edge and local-dev router | to verify |
| `busybox:stable-musl@sha256:3c6ae800...` | one-shot `acme-init` helper in Compose (staging/production). **No longer in the API image** (removed 2026-11-12; the API image ships its own `/healthcheck` command) | GPL-2.0 per `TODO.md` S8 (observed fact); see the owner decision below |
| `golang:1.27.4-alpine@sha256:0a5f7fe7...` | API build stage (not shipped) | to verify |
| `gcr.io/distroless/static-debian13:nonroot@sha256:25ac8f0c...` | API and migrate runtime base | to verify |
| `node:26-alpine@sha256:6bd9ab43...` | web build stage (not shipped) | to verify |
| `nginxinc/nginx-unprivileged:stable-alpine@sha256:f5116ed4...` | web runtime base | to verify |

**Owner decision 2026-11-12:** runtime base images (including busybox, GPL-2.0, in the `acme-init` helper, and Alpine packages in the web image) are treated as deployed container images, not as bundled project code, so the GPL/LGPL ban of CADEV article 6 is read as applying to the project's own source and its declared dependencies. Residual risk: the organiser has not confirmed this reading. Optional follow-ups in `TODO.md` S8: ask the organiser (cadev@systalink.com), produce a licence report of the built images `api`, `migrate`, `web`. Busybox is already out of the API image. No image licence report exists yet.

## Fonts and icons

- Fonts: system font stack only (`DESIGN.md`); no remote font is loaded and none is bundled.
- Icons: the prototype uses its own CSS sprite; no third-party icon set is declared.

## Brand assets

The 226os logo, symbol and app icons were produced with an AI image generator by the repository owner and resized with tooling (record: "Brand assets (concept D) record" in `AI_USAGE.md`). Copyright and trademark status review with CADEV/legal is PENDING.

## Build and CI tooling (not distributed)

- Go `tool` entries in `go.mod`: golangci-lint, go-licenses, oapi-codegen, goose, sqlc, govulncheck, staticcheck. Used to build, check and generate; not part of the shipped binaries as tools. Their licences are `to verify` (go-licenses reports runtime dependencies only).
- `anthropics/claude-code-action` (pinned by commit SHA to v1.0.275 in `.github/workflows/code-review.yml`, off by default): runs Claude Code in CI for the advisory review; licence `to verify`; not distributed.
- GitHub Actions in `.github/workflows/` are pinned by commit SHA (version in the trailing comment of each `uses:` line). Not distributed.
- open-code-review (`@alibaba-group/open-code-review`, Apache-2.0, pinned to 1.13.8 in `scripts/review.sh`): optional AI code review helper run through `npx`, never installed globally and not distributed. `.claude/commands/delegate-review.md` is adapted from its `delegate-review.md` (commit 4236c871fe8130aa840a2a1c0922f469e87e0416, Apache-2.0, Copyright 2026 Alibaba; changes listed in the file header). With `make review` it sends the diff to the LLM provider the user configures; with `/delegate-review` the reviewer is Claude Code. Licence read from the package metadata (`npm view`) and the repository page, not re-verified from the source tree.
- Other tooling named in Make targets (for example SOPS, age, gitleaks, Trivy, Semgrep): licences `to verify`; not distributed.

## Original-work statement and regenerable inventory

- Original-work statement (FR and EN, owner confirmation pending): `docs/research/REFERENCES.md`, section "Déclaration d'originalité".
- AI use and provenance: `AI_USAGE.md`.
- Regenerable inventory: `make licenses` writes `reports/licenses/` (git-ignored).
