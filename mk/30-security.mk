# G4: security and licences (ADR-0006). Only real checks; each fails closed.
#
# `security` runs every check through the sequential driver (scripts/quality.sh): all checks run,
# each exit code is recorded, a per-check PASS/FAIL summary is printed and ANY failure fails the target.
#
# Gate-time network use (the only ones, all read-only lookups, all fail closed when unreachable):
#   security-govulncheck  govulncheck queries the Go vulnerability database (vuln.go.dev)
#   security-js-audit     pnpm audit queries the npm registry advisory endpoint
#   security-images       Trivy downloads its vulnerability database (ghcr.io); first run also pulls the pinned image
# Pinned tools: govulncheck, golangci-lint (gosec), go-licenses via go.mod `tool`; gitleaks, Trivy and
# ShellCheck and Semgrep by image digest in scripts/*.sh. gitleaks and Semgrep run with --network none.
#
# SAST: Go with gosec through the already pinned golangci-lint (--enable-only=gosec): no new tool, same
# module-pinned version as `lint`. TypeScript/React with Semgrep (image pinned by digest, metrics off,
# --network none) and the repository's own rules in scripts/semgrep/web-rules.yml: no registry download,
# no drift, no third-party rule licence (semgrep-rules is under the non-OSS Semgrep Rules License v1.0 and
# is not vendored). Any finding at ERROR or WARNING severity fails; generated schema.d.ts is excluded.

SECURITY_STEPS := security-authz security-govulncheck security-js-audit security-secrets security-sast security-sast-js security-images

security: doctor ## G4: negative authz route table, govulncheck, JS audit, secret scan, SAST (gosec, Semgrep), image scan
	@scripts/quality.sh $(MAKE_BIN) security $(SECURITY_STEPS)

security-authz: doctor ## G4: every mounted route is public by allowlist or covered by an authz negative test
	scripts/check-authz-routes.sh

security-govulncheck: doctor ## G4: Go vulnerability scan (queries vuln.go.dev)
	go tool govulncheck ./...

security-js-audit: doctor ## G4: JS production dependency audit, high and above (queries the npm registry)
	pnpm audit --prod --audit-level high

security-secrets: doctor ## G4: gitleaks (pinned image) over the working tree and the history of HEAD
	scripts/secrets-scan.sh

security-sast: doctor ## G4: gosec via the pinned golangci-lint
	go tool golangci-lint run --enable-only=gosec --tests=false ./...

security-sast-js: doctor ## G4: Semgrep (pinned image, repository rules) over apps/web/src
	scripts/sast-js.sh

security-images: build ## G4: Trivy (pinned image) on the built api, web, migrate images; fixable HIGH/CRITICAL fails
	scripts/image-scan.sh

licenses: doctor ## G4/G7: runtime dependency licence inventory against the allowed list; artifact under reports/licenses
	scripts/licenses.sh
