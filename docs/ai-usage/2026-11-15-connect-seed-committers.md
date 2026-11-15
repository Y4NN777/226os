## Local-dev Connect seed from committers.top

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5"
purpose: "The AI wrote a stdlib generator that turns committers.top country files (the `users` list) into a local-dev seed of real public GitHub developers for the Connect directory, the make seed change that loads db/seeds/local-dev/ for APP_ENV=local-dev only, and the tests."
repository_or_source_context: "branch feature/connect/seed-committers, from develop 213bf3d"
files_or_sections_affected: ["scripts/content/committers-seed.py", "db/seeds/local-dev/00001_connect_committers.sql", "mk/50-ops.mk", "scripts/tests/seed_test.sh", "scripts/tests/committers-seed_test.sh", "scripts/tests/fixtures/committers-bf.yml", "src/api/internal/platform/pgtest/local_dev_seed_integration_test.go", "db/seeds/README.md", "TODO.md"]
third_party_sources_consulted: ["committers.top data (github.com/ashkulz/committers.top, MIT, gh-pages, _data/locations/*.yml)"]
license_review: "data from an MIT repository; real public GitHub profile data, local-dev only by owner decision; no new dependency"
closure: "open"
human_reviewers: []
tests_or_evidence: ["seed_test.sh 20/20 (RED first for the local-dev directory)", "committers-seed_test.sh 17/17 (RED first)", "TestLocalDevSeed* pass against PostgreSQL; the taken-name test fails when the skip clause is removed", "existing TestSeed* unchanged and passing"]
open_risks: ["real people's names in a local-dev database: never load db/seeds/local-dev/ elsewhere", "the data ages: regenerate with the generator", "make seed end to end against the local-dev stack NOT RUN in this session"]
```
