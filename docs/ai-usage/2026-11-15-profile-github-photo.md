## Profile photo from the member's GitHub avatar

```yaml
date: "2026-11-15"
contributor: "PANK4SS"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (web-implementer subagent) led by Claude Opus 5.5"
purpose: "A profile photo taken from the member's GitHub avatar: the ProfilePhoto component (validated login, lazy image, no referrer, initials fallback), its use in the profile hero and the Connect rows, the help line on the GitHub field, the CSP img-src entry and their tests."
repository_or_source_context: "branch feature/profile/web-github-photo; TODO.md S7 block; api/openapi.yaml read only (github field of MyProfile, PublicProfile, ProfileSummary)"
files_or_sections_affected: ["src/web/src/components/ProfilePhoto.tsx", "src/web/src/components/ProfilePhoto.test.tsx", "src/web/src/features/profile/**", "src/web/src/features/connect/**", "src/web/nginx/security-headers.conf", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency; pictures are loaded at run time from GitHub, not stored or redistributed"
closure: "open: waits for the owner's decision and review"
human_reviewers: []
tests_or_evidence: ["RED: ProfilePhoto.test.tsx failed on the missing module; the ProfilePage and ConnectPage photo and hint tests failed before the wiring", "GREEN: see the commit report (vitest, typecheck, lint, prettier)"]
open_risks: ["privacy: GitHub sees the viewer's IP address when a photo loads (no referrer sent); the owner decides whether to accept it", "the shell avatar stays initials: the session user carries no github field and no request was added", "CSP img-src conflicts with the other branch that adds opengraph.githubassets.com: the lead resolves it at integration"]
```
