/* Build lane FICTIONAL data: taxonomy, extra projects, contribution needs, applications, activity.
   Extends shared MOCK.projects (read-only) through MOCK.build.meta[id]. Project/need/application texts are USER CONTENT in their own `lang`.
   Nothing here is a real project, person or metric. Repository links point at example.org only. */
window.MOCK = window.MOCK || {};
(function (M) {
  var B = M.build = {};

  B.skills = {
    go: { fr: 'Go / backend', en: 'Go backend' }, postgres: { fr: 'SQL / PostgreSQL', en: 'SQL / PostgreSQL' }, php: { fr: 'PHP / backend', en: 'PHP backend' },
    react: { fr: 'Frontend React', en: 'React frontend' }, node: { fr: 'Node.js / TypeScript', en: 'Node.js / TypeScript' }, python: { fr: 'Python', en: 'Python' },
    docs: { fr: 'Rédaction technique', en: 'Technical writing' }, review: { fr: 'Relecture FR/EN', en: 'FR/EN review' }, ux: { fr: 'UX FR/EN', en: 'UX FR/EN' },
    testing: { fr: 'Tests', en: 'Testing' }, devops: { fr: 'DevOps / CI', en: 'DevOps / CI' }
  };
  B.stages = { idea: { fr: 'Idée', en: 'Idea' }, prototype: { fr: 'Prototype', en: 'Prototype' }, active: { fr: 'Actif', en: 'Active' }, maintained: { fr: 'Maintenance', en: 'Maintenance' } };
  B.licenses = ['MIT', 'Apache-2.0', 'AGPL-3.0', 'CC-BY-4.0'];
  B.levels = { beginner: { fr: 'Débutant accepté', en: 'Beginner friendly' }, intermediate: { fr: 'Intermédiaire', en: 'Intermediate' }, advanced: { fr: 'Avancé', en: 'Advanced' } };
  B.availability = ['lt2', '2to5', '5to10', 'gt10'];

  /* Projects not present in shared MOCK.projects (same shape). Linked question/resource ids exist in shared data. */
  B.extraProjects = [
    { id: 'pr-gsm7-check', name: 'gsm7-check', repo: 'example.org/exemple-226/gsm7-check', summary: 'Small library that detects characters forcing UCS-2 and suggests GSM-7 safe replacements for French text.', lang: 'en',
      stack: ['typescript', 'node'], status: 'active', country: 'CI', owner: 'p-salif', updated: '2026-11-06', resource: 'r-sahel-sms', question: 'q-sms-encoding' },
    { id: 'pr-kyc-photo-qa', name: 'kyc-photo-qa', repo: 'example.org/exemple-226/kyc-photo-qa', summary: 'Client-side checks (blur, glare, framing) that guide users before an ID photo is sent to a KYC API.', lang: 'en',
      stack: ['python', 'typescript'], status: 'idea', country: 'NG', owner: 'p-chinedu', updated: '2026-11-05', resource: 'r-baobab-kyc', question: 'q-kyc-ocr' },
    { id: 'pr-ledger-lite', name: 'ledger-lite', repo: 'example.org/exemple-226/ledger-lite', summary: 'Minimal double-entry ledger examples built on top of an open source SDK. The current scope is complete; looking for reviewers.', lang: 'en',
      stack: ['go', 'postgresql'], status: 'active', country: 'KE', owner: 'p-wanjiru', updated: '2026-10-20', resource: 'r-savane-ledger', question: null }
  ];

  /* Per-project extras. needs[].status: open | closed. License/stage/langs drive filters. */
  B.meta = {
    'pr-webhook-kit': { stage: 'active', license: 'MIT', langs: ['fr', 'en'],
      description: '## Objectif\nValider la signature des webhooks de paiement et **rejouer** les événements sans doublon.\n\n## État actuel\n- Vérification HMAC-SHA256 sur le corps brut\n- Table `webhook_events` avec clé d’idempotence\n- Exemples avec le fournisseur fictif Kalimba Pay\n\n## Comment contribuer\nLes besoins ouverts sont listés ci-dessous. Une première issue « good first issue » est réservée à chaque nouveau contributeur.',
      needs: [
        { id: 'n-wk-1', title: 'Dev Go pour le rejeu idempotent des événements', skills: ['go', 'postgres'], level: 'intermediate', status: 'open' },
        { id: 'n-wk-2', title: 'Documentation du guide d’intégration (FR vers EN)', skills: ['docs', 'review'], level: 'beginner', status: 'open' },
        { id: 'n-wk-3', title: 'Revue de sécurité de la vérification HMAC', skills: ['go'], level: 'advanced', status: 'closed' }
      ],
      activity: [{ d: '2026-11-08', k: 'repo', who: 'p-aminata' }, { d: '2026-11-03', k: 'need', who: 'p-aminata', ref: 'n-wk-2' }, { d: '2026-10-29', k: 'link', who: 'p-aminata', ref: 'q-webhook-sig' }, { d: '2026-10-21', k: 'created', who: 'p-aminata' }],
      applications: [
        { id: 'a-wk-1', need: 'n-wk-1', person: 'p-kofi', status: 'pending', date: '2026-11-09', avail: '2to5', skills: 'Go, PostgreSQL, idempotency keys', lang: 'en',
          motivation: 'I answered the idempotency question and maintain a Node integration with the same provider. I would like to port the replay logic to Go with tests.' },
        { id: 'a-wk-2', need: 'n-wk-2', person: 'p-wanjiru', status: 'accepted', date: '2026-11-05', avail: 'lt2', skills: 'Technical writing, FR/EN review', lang: 'en',
          motivation: 'I can review the English version of the integration guide and keep the glossary consistent.' }
      ] },
    'pr-offline-drafts': { stage: 'idea', license: 'Apache-2.0', langs: ['fr'],
      description: '## Idée\nUne **file d’attente locale** pour brouillons qui reprend automatiquement quand la connexion revient.\n\n## Questions ouvertes\n- Résolution de conflits : dernière écriture ou fusion ?\n- Taille maximale du stockage local sur mobile d’entrée de gamme.',
      needs: [
        { id: 'n-od-1', title: 'Dev React pour le prototype de file d’attente', skills: ['react', 'node'], level: 'intermediate', status: 'open' },
        { id: 'n-od-2', title: 'Tests de coupure réseau (3G instable)', skills: ['testing'], level: 'beginner', status: 'open' }
      ],
      activity: [{ d: '2026-11-04', k: 'link', who: 'p-yves', ref: 'q-offline-sync' }, { d: '2026-11-04', k: 'need', who: 'p-yves', ref: 'n-od-1' }, { d: '2026-11-04', k: 'created', who: 'p-yves' }],
      applications: [{ id: 'a-od-1', need: 'n-od-1', person: 'p-kofi', status: 'pending', date: '2026-11-10', avail: '5to10', skills: 'React, TypeScript', lang: 'en', motivation: 'I have built offline-first forms for field agents and can contribute the queue prototype.' }] },
    'pr-ussd-flow': { stage: 'active', license: 'AGPL-3.0', langs: ['en', 'fr'],
      description: '## Goal\nA small **state machine** for USSD journeys with session persistence, so a 30-second operator timeout does not lose a payment.\n\n## Status\n- Core engine and PHP adapter working\n- Redis-backed session store\n\n## Help wanted\nThe PHP backend needs review before the first tagged release.',
      needs: [
        { id: 'n-uf-1', title: 'PHP backend reviewer for the session store', skills: ['php'], level: 'intermediate', status: 'open' },
        { id: 'n-uf-2', title: 'Integration tests against a simulated operator gateway', skills: ['testing', 'devops'], level: 'intermediate', status: 'open' }
      ],
      activity: [{ d: '2026-11-01', k: 'repo', who: 'p-fatou' }, { d: '2026-10-25', k: 'link', who: 'p-fatou', ref: 'q-ussd-timeout' }, { d: '2026-10-12', k: 'created', who: 'p-fatou' }],
      applications: [{ id: 'a-uf-1', need: 'n-uf-1', person: 'p-kofi', status: 'pending', date: '2026-11-02', avail: 'lt2', skills: 'PHP, code review', lang: 'en', motivation: 'I review PHP payment integrations at work and can give feedback on the session store.' }] },
    'pr-docs-fr': { stage: 'active', license: 'CC-BY-4.0', langs: ['fr', 'en'],
      description: '## But\nUne documentation technique **en français** pour les intégrations locales (paiement, SMS, USSD), relue par des développeurs de la région.\n\n## Format\nMarkdown dans le dépôt, publication automatique par CI.\n\n## Besoin\nRédacteurs et relecteurs : deux heures par semaine suffisent.',
      needs: [
        { id: 'n-df-1', title: 'Rédaction de guides d’intégration Mobile Money', skills: ['docs'], level: 'beginner', status: 'open' },
        { id: 'n-df-2', title: 'Relecture FR des pages techniques', skills: ['review'], level: 'beginner', status: 'open' }
      ],
      activity: [{ d: '2026-10-27', k: 'need', who: 'p-wanjiru', ref: 'n-df-2' }, { d: '2026-10-15', k: 'created', who: 'p-wanjiru' }],
      applications: [{ id: 'a-df-1', need: 'n-df-1', person: 'p-kofi', status: 'pending', date: '2026-10-30', avail: '2to5', skills: 'Technical writing, mobile money', lang: 'en', motivation: 'I write internal integration notes already and would like to publish them in a shared place.' },
        { id: 'a-df-2', need: 'n-df-2', person: 'p-salif', status: 'declined', date: '2026-10-28', avail: 'lt2', skills: 'Relecture, PHP', lang: 'fr', motivation: 'Je peux relire les pages techniques en français.' }] },
    'pr-geo-informal': { stage: 'prototype', license: 'MIT', langs: ['en', 'fr'],
      description: '## Problem\nAddresses are described by **landmarks**, not street numbers. This project models them in PostGIS.\n\n## Approach\n- A `landmark` table with aliases per language\n- Fuzzy matching with `pg_trgm`\n- Confidence stored next to every geocode\n\n## Help wanted\nSQL review of the matching queries and their indexes.',
      needs: [{ id: 'n-gi-1', title: 'SQL / PostGIS reviewer for matching queries and indexes', skills: ['postgres'], level: 'advanced', status: 'open' }],
      activity: [{ d: '2026-10-30', k: 'repo', who: 'p-yves' }, { d: '2026-10-22', k: 'link', who: 'p-yves', ref: 'q-geocode' }, { d: '2026-10-10', k: 'created', who: 'p-yves' }],
      applications: [{ id: 'a-gi-1', need: 'n-gi-1', person: 'p-kofi', status: 'pending', date: '2026-11-01', avail: 'lt2', skills: 'PostgreSQL', lang: 'en', motivation: 'I tuned trigram indexes for fuzzy search before and can review the matching queries.' }] },
    'pr-gsm7-check': { stage: 'prototype', license: 'MIT', langs: ['en', 'fr'],
      description: '## Why\nAccented French characters push a message to UCS-2 and triple its cost. This library flags them and proposes **safe replacements**.\n\n## Planned API\n- `analyze(text)` returns the encoding and segment count\n- `suggest(text)` returns GSM-7 compatible alternatives\n\n## Help wanted\nA Node.js maintainer and UX copy for the warning messages.',
      needs: [
        { id: 'n-gc-1', title: 'Node.js / TypeScript developer for the analyzer', skills: ['node'], level: 'intermediate', status: 'open' },
        { id: 'n-gc-2', title: 'UX writing for warnings, FR and EN', skills: ['ux', 'review'], level: 'beginner', status: 'open' }
      ],
      activity: [{ d: '2026-11-06', k: 'link', who: 'p-salif', ref: 'q-sms-encoding' }, { d: '2026-11-06', k: 'created', who: 'p-salif' }],
      applications: [{ id: 'a-gc-1', need: 'n-gc-1', person: 'p-kofi', status: 'pending', date: '2026-11-10', avail: '2to5', skills: 'Node.js, TypeScript', lang: 'en', motivation: 'I handle SMS encodings in a gateway integration and can build the analyzer.' }] },
    'pr-kyc-photo-qa': { stage: 'idea', license: 'Apache-2.0', langs: ['en'],
      description: '## Idea\nGuide the user while taking an ID photo: detect **blur, glare and framing** on the device before upload.\n\n## Constraints\n- Must run on low-end phones\n- No image leaves the device for the quality check\n\n## Help wanted\nPython prototyping for the checks and UX for the guided capture, in French and English.',
      needs: [
        { id: 'n-kp-1', title: 'Python prototype of blur and glare detection', skills: ['python'], level: 'intermediate', status: 'open' },
        { id: 'n-kp-2', title: 'UX for the guided capture flow, FR/EN', skills: ['ux'], level: 'beginner', status: 'open' }
      ],
      activity: [{ d: '2026-11-05', k: 'link', who: 'p-chinedu', ref: 'q-kyc-ocr' }, { d: '2026-11-05', k: 'created', who: 'p-chinedu' }],
      applications: [{ id: 'a-kp-1', need: 'n-kp-2', person: 'p-kofi', status: 'pending', date: '2026-11-10', avail: 'lt2', skills: 'UX, React', lang: 'en', motivation: 'I design onboarding flows for first-time smartphone users and speak French and English.' }] },
    'pr-ledger-lite': { stage: 'maintained', license: 'MIT', langs: ['en'],
      description: '## Scope\nMinimal **double-entry ledger** examples on top of an open source SDK. The scope is complete and the project is in maintenance.\n\n## Status\nBoth contribution needs are closed. You can still read the code and open an issue on the fictional repository.',
      needs: [
        { id: 'n-ll-1', title: 'Go reviewer for the posting rules', skills: ['go'], level: 'intermediate', status: 'closed' },
        { id: 'n-ll-2', title: 'CI pipeline for example tests', skills: ['devops'], level: 'beginner', status: 'closed' }
      ],
      activity: [{ d: '2026-10-20', k: 'need-closed', who: 'p-wanjiru', ref: 'n-ll-2' }, { d: '2026-10-02', k: 'created', who: 'p-wanjiru' }],
      applications: [{ id: 'a-ll-1', need: 'n-ll-1', person: 'p-kofi', status: 'accepted', date: '2026-10-08', avail: 'lt2', skills: 'Go', lang: 'en', motivation: 'I can review posting rules for rounding and currency edge cases.' }] }
  };

  /* Merged, normalized project list (base mock + extras). Never mutates shared arrays. */
  B.projects = function () {
    return (M.projects || []).concat(B.extraProjects).map(function (p) {
      var m = B.meta[p.id], needs = m.needs;
      var openNeeds = needs.filter(function (n) { return n.status === 'open'; });
      var skills = []; openNeeds.forEach(function (n) { n.skills.forEach(function (s) { if (skills.indexOf(s) < 0) skills.push(s); }); });
      return { id: p.id, name: p.name, repo: p.repo, summary: p.summary, lang: p.lang, stack: p.stack, country: p.country, owner: p.owner, updated: p.updated,
        resource: p.resource, question: p.question, stage: m.stage, license: m.license, langs: m.langs, description: m.description, needs: needs, openCount: openNeeds.length,
        skills: skills, activity: m.activity, applications: m.applications };
    });
  };
  B.byId = function (id) { return B.projects().filter(function (p) { return p.id === id; })[0] || null; };
  B.repoUrl = function (p) { return 'https://example.org/' + p.repo.replace(/^[^/]+\//, ''); };
  B.repoLabel = function (p) { return 'example.org/' + p.repo.replace(/^[^/]+\//, ''); };
  B.question = function (id) { return id ? M.byId(M.questions, id) : null; };
  B.resource = function (id) { return id ? M.byId(M.resources, id) : null; };
})(window.MOCK);
