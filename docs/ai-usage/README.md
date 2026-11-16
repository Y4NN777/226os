# AI-use records, one file per record

Add your record here as a new file, never by editing `AI_USAGE.md`. Two pull requests then never touch the same lines, so they never conflict on it (every pull request used to append at the end of `AI_USAGE.md`, and every merge broke the other open ones).

**File:** `docs/ai-usage/YYYY-MM-DD-short-slug.md`, lower case, digits and dashes only (for example `2026-11-14-profile-visibility-ui.md`). Files with another name (this README, notes, drafts) are not records.

**Content:** a `## Title` heading, then the yaml block of the template in `AI_USAGE.md` ("Required record for each material AI-assisted contribution"), at least `date: "YYYY-MM-DD"` and `contributor:` (your GitHub login). The record is closed by the owner, never by its own contributor.

**`AI_USAGE.md`** keeps the policy, the historical records and a block at the end that is generated from these files. Do not edit that block: before the submission the owner runs `make ai-usage`, commits the result, and `make ai-usage-check` must then pass.

```
make ai-usage         # rewrite the generated block of AI_USAGE.md from docs/ai-usage/
make ai-usage-check   # fail when AI_USAGE.md is stale
```
