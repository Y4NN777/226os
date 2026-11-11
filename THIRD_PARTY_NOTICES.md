# 226OS Third-Party Notices

**Status:** DRAFT inventory policy; dependency inventory is not yet complete

## Policy

Every direct and transitive runtime/build dependency must be pinned in a lockfile, identified by package and version, mapped to an SPDX license, and reviewed against CADEV rules before release. Do not copy application source from reference repositories. References in `docs/research/REFERENCES.md` are research inputs, not declared dependencies.

The final inventory must include notices, source/license links, modification status, attribution requirements, and any reciprocal-license or trademark obligations. Vulnerability and license reports are evidence-bearing artifacts tied to the target SHA.

## Planned inventory categories

- Frontend: React, TypeScript, Vite, TanStack packages, Meta Astryx, styling and test tooling.
- Backend: Go runtime libraries, Chi, pgx, sqlc tooling, migration and test tooling.
- Infrastructure: Docker/Traefik images and action dependencies.
- Reference material: Apache Answer, Flarum, Nuxt HackerNews, 225 Open Source, and other research sources; no code integration is asserted by this file.

## Current status

No implementation dependency set, lockfile, SPDX report, or transitive notice bundle exists yet. Release compliance is `BLOCKED` until the inventory is generated from the actual pinned dependency graph and reviewed by a human.
