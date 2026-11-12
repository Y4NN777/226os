> Prototype documents. The prototype is frozen; `DESIGN.md` governs `src/web`.

# 226os prototype: Direction v2 (flows and premium look)

**Status:** Approved (see CONVENTIONS.md §12).
**Inputs:**
- the owner's feedback of 2026-11-11: no vertical bars; sign-in should not sit inside the app; no landing page; the prototype lacks a premium feel; flow problems affect the whole prototype, not just auth;
- four web research reports (sources listed below);
- the PRD, which decides scope. Every item is tagged **P0**, **P1 preview** or **Out**.

---

## 1. Principles

1. **Public content, no wall.** Anyone can read questions, the catalogue and projects without an account. Sign-in is asked for only at the moment of an action: answering, voting, applying or reviewing. Then the user returns to that action. [Stack Overflow, Hugging Face]
2. **One page, one task.** Sign-in, sign-up and every onboarding step each get their own page. [GOV.UK "one thing per page", GOV.UK create-accounts]
3. **Every step ends with a next action.** A success screen always leads somewhere: a similar question, the API used, the project that needs it.
4. **Progressive disclosure, at most two levels.** Secondary filters, metadata and the prototype panel stay hidden until asked for. [NN/g]
5. **State is shown with an icon and a word, never with a coloured bar.** [Primer StateLabel]

## 2. Flow map

### Logged out
```
Landing (/) ─┬─> Browse questions ─> Question ─> [Answer / Vote] ─> Sign in? ─> back to the action
             ├─> Browse catalogue ─> API page ─> [Review] ─────────> Sign in? ─> back
             ├─> Browse projects ─> Project ─> [Apply] ────────────> Sign in? ─> back
             └─> Create an account ─> Check your email ─> Onboarding ─> Home
```

### Accounts (standalone layout, no app chrome) — P0
- `connexion.html`: email and password (autocomplete `username` / `current-password`), a show/hide icon, "Mot de passe oublié ?" next to the field, and a link to sign-up.
- `inscription.html`: name, email and password (`new-password`), with the rules shown before typing, and a link to sign-in.
- `verifier-email.html`: "Vérifiez votre boîte mail", a resend button, neutral wording.
- `mot-de-passe-oublie.html`: one email field. The confirmation message is the same whether or not the account exists. [OWASP]
- `reinitialiser.html`: new password, plus states for an invalid or expired link.
- States: generic error, "Trop de tentatives, réessayez plus tard", and session expired with a return to the original page.

### Gentle onboarding — P0 for the minimal profile
**Rule:** no profile field on the account forms. Sign-up asks only for name, email and password. The profile is built gradually.

**Owner principle (revision 2026-11-11): every profile item is tied to a value in the product and is asked for when that value becomes concrete. Never ask for something blindly.**

1. **Onboarding = one screen after verifying the email:** "Qu'est-ce qui vous amène ?" with four choices: ask a question, help others, integrate an API, contribute to a project. Then "Sur quoi pouvez-vous aider ?" with 1–3 skills. The effect is shown immediately ("12 questions attendent quelqu'un qui connaît Go"). The username is generated from the name and can be changed in settings.
2. **Just-in-time requests** (dismissible inline prompt that states the value, remembered once dismissed):
   - **Country:** in Découvrir, "Voir les API disponibles au Burkina ? Mémoriser ce pays".
   - **GitHub:** in the application dialog, "Les mainteneurs regardent votre GitHub en premier".
   - **Ouvert aux collaborations, plus bio:** after a first accepted answer, or when the user wants to be found by maintainers.
   - **Languages:** never asked. Inferred from the interface; offered once when content in another language is read.
   - **Avatar:** never asked. The generated monogram is the default; changeable in settings.
3. **Completion checklist on your own profile** (the LinkedIn "Profile strength", Upwork, Stripe and Vercel setup-checklist pattern). Visible **only to the profile owner**, on their profile page. Never on the home page.
   - A progress meter (percentage and "x/6").
   - Items: skills (done at onboarding), country, GitHub, bio, open to collaboration, languages (confirm). Each has **one line on why it matters in 226os** ("Le pays affiche d'abord les API qui couvrent votre marché").
   - Each item is filled in **in place** (inline editor or a small panel) and ticks with a subtle animation.
   - At 100%, a short "Profil complet" confirmation, then the card disappears for good. It can be collapsed before that.
   - In-context prompts (country in Découvrir, GitHub when applying, and so on) tick the same items.
4. The first home is chosen by the answer to "Qu'est-ce qui vous amène ?". No separate "Par où commencer ?" chooser.

### Profile and settings space — P0
- `profil.html`: the public view of the profile.
- `parametres.html` (standalone, simple sections, no tabs that scroll):
  - **Profil:** name, username, avatar, bio, skills, country, languages, GitHub, availability.
  - **Compte:** email, password, active sessions with sign-out (G0-AUTH).
  - **Préférences:** interface language (overrides automatic detection), theme.
  - **Confidentialité:** visibility of each field. Email is never public (AUTHORIZATION_MATRIX).
- Opened from the account menu ("Mon profil", "Paramètres") and from every checklist item.

### Logged-in home — P0
- A "Pour vous" view filtered by the declared skills and tags. This is deterministic, with no AI: PRD 5.8 makes Intelligence P0 deterministic.
- Blocks: unanswered questions in my tags, projects looking for my skills, recently verified APIs, and my activity (my questions, my applications).
- **Out:** inbox, notifications, email digest. They are not in the PRD; they need an ADR or a PRD change first.

### Core loops
| Loop | Steps | Next action after success | Priority |
|---|---|---|---|
| Ask | Title → similar questions while typing → details, code, tags → publish → "what happens next" screen | Explore the related API | P0 |
| Answer / accept | Question → answer → the author accepts | "Another unanswered question in this tag" | P0 |
| Integrate an API | Catalogue → API page (sources, pitfalls, snippet) → leave a review | Related open-source projects | Catalogue P0; reviews **P1 preview** |
| Contribute | Project → need → apply (short form) → the owner accepts | First contribution steps | P0 |
| Publish a project or need | GitHub URL → need, skills, level → publish | Share and see similar profiles | P0 |
| Mentoring | Designed only | — | **P1 preview** |

### Empty states that guide
- Empty feed: "Choisissez 3 sujets pour commencer", with suggested tags.
- Question with no answer: "Soyez le premier à répondre", with a similar question.
- No applications: open projects that need your skills.
- First-use empty and no-results empty look different.

## 3. Premium look

### Removed
- Every vertical bar or left-border stripe: active menu item, accepted answer, callouts.
- Tabs on the sign-in page (they also caused the scrollbar bug).
- The lime band on the hero. The app keeps sober surfaces; the brand moment lives on the landing page.

### Surfaces and depth [Linear]
- Separation comes from layers of tinted surfaces (background, then panel, then raised, then dialog), not from edges.
- At most one hairline separator inside a list. No border on cards.
- Shadows only on overlays (menus, dialogs, toasts).

### Colour restraint [Supabase, Linear]
- Primary green is reserved for actions and state: primary button, link, focus, selection.
- Everything else uses the Graphite neutral ramp.
- Lime is a rare brand moment: logo, landing page, an accepted state on dark backgrounds.
- Copper only for warning or attention.

### State
- Accepted, verified and to-be-reverified are a pill with an icon and a word on a neutral row.
- An accepted answer adds a discreet "Réponse acceptée" header.

### Typography [Linear, Vercel]
- Stronger contrast between heading and body: headings at weight 650–700 with slightly tight tracking (−0.01em to −0.02em), body text at a comfortable 1.55 line height.
- Hierarchy by weight and colour more than by size.
- `tabular-nums` on counts, votes and dates.
- Prose limited to 68ch.

### Navigation chrome
- Active item: a soft neutral background and full-contrast text with its icon. No bar.
- Icons 18px, labels and icons aligned on one grid.
- The sidebar gets lighter. Create (+) and search remain the main entry points.

### Motion [Emil Kowalski, Vercel guidelines]
- Buttons scale to 0.97 on `:active`.
- Entries animate from scale ≥ 0.96 plus opacity, eased out with a custom curve (`cubic-bezier(.2,.8,.2,1)`), under 200ms. Tooltips use about 125ms.
- No animation on frequent, repetitive actions.
- `prefers-reduced-motion` removes scale and translation.

### Dark mode
- Elevation means slightly lighter surfaces, not shadows.
- Icons get lighter.
- Lime is used only on dark backgrounds where its contrast is high.

## 4. Landing page (`index.html` becomes the public landing; the app home moves to `accueil.html`)

1. **Header:** logo, links (Questions, Catalogue, Projets), "Se connecter", and a primary "Créer un compte". The language is automatic, with an override in the footer.
2. **Hero:** one sentence saying what the product is and who it is for, a short subtitle, one dominant action, and a secondary link. On the right, a **real preview** of a question card linked to an API and a project, built in HTML with no image. [NN/g]
3. **Real content:** 3 questions, 3 APIs and 3 open projects, curated and clickable.
4. **How it works, in three steps:** ask, then find the sourced API, then build together. This is the PRD's Problem-to-Project differentiator.
5. **Trust:** sourcing rules (dated sources, "non vérifié" shown openly), open source, code of conduct, data handling. Only what exists. No invented numbers.
6. **Final call to action**, then a footer with licence, code of conduct, contribute, privacy and language.
- **Performance:** no video and no hero image. The largest element is text. No pop-up.

## 5. Decisions for you

1. Do you approve this flow map, especially "public content, sign-in at the moment of an action" and the move of the app home to `accueil.html`?
2. Gentle onboarding: is the requirement of username + 1–3 skills, followed by a 6-item checklist, right for you?
3. Out of scope as proposed: inbox, notifications, email digest, GitHub sign-in. Confirm.

## Sources
- NN/g, homepage principles: https://www.nngroup.com/articles/homepage-design-principles/
- NN/g, progressive disclosure: https://www.nngroup.com/articles/progressive-disclosure/
- GOV.UK, create accounts: https://design-system.service.gov.uk/patterns/create-accounts/
- GOV.UK, passwords: https://design-system.service.gov.uk/patterns/passwords/
- GOV.UK, one thing per page: https://designnotes.blog.gov.uk/2015/07/03/one-thing-per-page/
- web.dev, sign-in form best practices: https://web.dev/articles/sign-in-form-best-practices
- OWASP, authentication cheat sheet: https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
- WCAG 2.2, 3.3.8: https://www.w3.org/WAI/WCAG22/Understanding/accessible-authentication-minimum.html
- Stack Overflow, improving the ask experience: https://stackoverflow.blog/2019/09/26/research-update-improving-the-question-asking-experience/
- GitHub, good first issues: https://github.blog/open-source/maintainers/how-we-built-good-first-issues
- Linear, how we redesigned the UI: https://linear.app/now/how-we-redesigned-the-linear-ui
- Supabase, color usage: https://supabase.com/design-system/docs/color-usage
- Primer, StateLabel: https://primer.style/product/components/state-label/
- Emil Kowalski, animation tips: https://emilkowal.ski/ui/7-practical-animation-tips
- Vercel, web interface guidelines: https://vercel.com/design/guidelines
- Chrome, optimise image priority: https://developer.chrome.com/docs/modern-web-guidance/guides/performance/optimize-image-priority

**Not verified by the research:** the exact sign-in proportions of Linear, Stripe and GitHub; the Geist tokens; the Baymard figures. They are not used.
