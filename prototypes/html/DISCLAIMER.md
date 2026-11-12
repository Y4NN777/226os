# READ THIS FIRST: these prototypes are a design preview, not the product

**For the frontend owner (`PANK4SS`) and anyone who works on `src/web`.** Last updated 2026-11-12.

## What this folder is

A static HTML mock-up made before the product decisions of 2026-11-12. It shows layout, states and tone. It is **not** the specification, **not** production code, and it is **not deployed**. The real application is `src/web` (React 19, Astryx, TanStack) and it replaces these pages.

**All people, questions, projects, resources, articles and numbers in the prototypes are fictional.** Never copy them into the app, the fixtures or any document as if they were real. No real names or emails anywhere.

## What is outdated here

| In the prototype | Now |
|---|---|
| Navigation puts Explorer, Résoudre, Découvrir, Construire, Communauté, Mentorat on the same level | **Learn & Grow** and **Connect** are the two prominent pillars in navigation, home and wording. Every other pillar stays present. Nothing is removed or hidden. |
| `connect.html` is a "coming soon" page | It becomes a **people directory** (country, language and skill filters, caller-visible fields only, email never shown). The "spaces" block stays labelled **bientôt**. Mentoring, CV review, interviews and opportunities are not features yet. |
| Six Learn topics, seven resource categories | **14 learn topics and 18 resource categories** after the taxonomy migration (see `TODO.md`, section E2). Use the labels from that section, not the prototype's. |
| No assistant | A **grounded AI assistant**: answers only from published content, shows its sources, says when it has none, ends with "Comment vérifier", is labelled AI and cannot be voted or accepted as an answer. |
| No profile visibility setting | A **profile visibility** setting (public, members, private) on the settings page. A hidden profile looks exactly like an unknown one. |
| Examples are mostly payments and webhooks | Payments and Mobile Money stay in Discover (a differentiator), but they are no longer the face of the product. Learning, understanding and using LLMs and agents, open source and people are. |
| Fictional demo data | Staging and production hold real content only. Demo fixtures with non-real accounts exist in **local-dev only**. |

## What to follow instead

In this order:
1. `api/openapi.yaml` and the generated types in `src/web/src/api/schema.d.ts` for every data shape and every status code.
2. `docs/SPEC.md` for behaviour and acceptance scenarios, and `docs/PRD.md` (amendment 3.2) for product intent.
3. `DESIGN.md` for tokens, Graphite & Lime, accessibility and the Astryx rules.
4. `TODO.md`, section **S7 (Frontend)**, for your tasks, including the **Prototype alignment** checklist.

## How to use the prototypes

- **Reuse:** layout ideas, loading / empty / error states, copy tone, `data-testid` names and the FR/EN strings, after checking French typography (non-breaking space before `: ; ? !`) and natural developer French.
- **Do not:** treat them as acceptance criteria, port their mock data, or spend time keeping them in sync with the app. Do not add features to the prototypes. If a page needs to change, change the app.
- **If a prototype contradicts the SPEC, the OpenAPI or `TODO.md`, the prototype is wrong.**

The old screenshots were removed on purpose, because they showed the outdated direction.
