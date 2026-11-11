/* Shared FICTIONAL mock data: countries, categories, questions, API/tool resources, projects.
   Provider/organisation names are invented; no real company is endorsed; dates/counts are illustrative, not real metrics.
   Titles/excerpts are USER CONTENT in their original `lang` (never auto-translated). Taxonomy labels ({fr,en}) are UI strings.
   Lanes: read these arrays, don't mutate; put lane-specific extras in assets/data/<lane>.js. */
window.MOCK = window.MOCK || {};
(function (M) {
  M.countries = [
    { code: 'BF', fr: 'Burkina Faso', en: 'Burkina Faso' }, { code: 'SN', fr: 'Sénégal', en: 'Senegal' }, { code: 'CI', fr: 'Côte d’Ivoire', en: 'Côte d’Ivoire' },
    { code: 'GH', fr: 'Ghana', en: 'Ghana' }, { code: 'NG', fr: 'Nigéria', en: 'Nigeria' }, { code: 'KE', fr: 'Kenya', en: 'Kenya' },
    { code: 'CM', fr: 'Cameroun', en: 'Cameroon' }, { code: 'MA', fr: 'Maroc', en: 'Morocco' }, { code: 'RW', fr: 'Rwanda', en: 'Rwanda' }
  ];
  M.categories = [
    { id: 'payment', fr: 'Paiement / Mobile Money', en: 'Payments / Mobile Money' }, { id: 'sms', fr: 'SMS', en: 'SMS' },
    { id: 'kyc', fr: 'Identité / KYC', en: 'Identity / KYC' }, { id: 'logistics', fr: 'Logistique', en: 'Logistics' }, { id: 'fintech', fr: 'Fintech', en: 'Fintech' }
  ];
  M.kinds = { question: { fr: 'Question', en: 'Question', icon: 'help' }, resource: { fr: 'API / outil', en: 'API / tool', icon: 'compass' }, project: { fr: 'Projet', en: 'Project', icon: 'hammer' }, person: { fr: 'Membre', en: 'Member', icon: 'user' } };

  M.questions = [
    { id: 'q-webhook-sig', lang: 'fr', title: 'Vérifier la signature d’un webhook Mobile Money en Go (HMAC-SHA256) ?', excerpt: 'Le fournisseur envoie un en-tête X-Signature mais mon calcul HMAC ne correspond jamais. Corps brut ou corps décodé ?',
      tags: ['go', 'webhook', 'mobile-money'], country: 'BF', author: 'p-aminata', votes: 14, answers: 3, accepted: true, date: '2026-11-09', env: 'Go 1.23', resource: 'r-kalimba', project: 'pr-webhook-kit' },
    { id: 'q-ussd-timeout', lang: 'fr', title: 'Session USSD qui expire avant la confirmation du paiement', excerpt: 'Après 30 secondes la session est coupée côté opérateur. Comment garder l’état de la transaction côté serveur ?',
      tags: ['ussd', 'mobile-money'], country: 'SN', author: 'p-fatou', votes: 9, answers: 2, accepted: false, date: '2026-11-08', env: 'PHP 8.3', resource: 'r-sahel-sms', project: null },
    { id: 'q-idempotency', lang: 'en', title: 'Idempotency keys for payment callbacks that arrive twice', excerpt: 'Our provider retries callbacks up to five times. What is a safe schema to deduplicate without losing legitimate updates?',
      tags: ['postgresql', 'webhook', 'payments'], country: 'GH', author: 'p-kofi', votes: 21, answers: 4, accepted: true, date: '2026-11-07', env: 'PostgreSQL 16', resource: 'r-kalimba', project: 'pr-webhook-kit' },
    { id: 'q-sms-encoding', lang: 'en', title: 'SMS with accented French characters is split into too many segments', excerpt: 'Messages with é and ç become UCS-2 and cost three segments. Is there a reliable way to stay on GSM-7?',
      tags: ['sms', 'unicode'], country: 'CI', author: 'p-salif', votes: 6, answers: 1, accepted: false, date: '2026-11-06', env: 'Node 22', resource: 'r-sahel-sms', project: null },
    { id: 'q-kyc-ocr', lang: 'en', title: 'Handling poorly lit ID photos before sending them to a KYC API', excerpt: 'About 20% of uploads fail the quality check. Which client-side checks reduce rejected submissions?',
      tags: ['kyc', 'python'], country: 'NG', author: 'p-chinedu', votes: 11, answers: 2, accepted: false, date: '2026-11-05', env: 'Python 3.12', resource: 'r-baobab-kyc', project: null },
    { id: 'q-offline-sync', lang: 'fr', title: 'Synchroniser des brouillons hors ligne avec une connexion instable', excerpt: 'Je cherche une stratégie simple de file d’attente locale avec reprise, sans dépendre d’un service tiers.',
      tags: ['javascript', 'offline', 'react'], country: 'CM', author: 'p-yves', votes: 8, answers: 0, accepted: false, date: '2026-11-04', env: 'React 19', resource: null, project: 'pr-offline-drafts' },
    { id: 'q-geocode', lang: 'fr', title: 'Géocoder des adresses informelles (sans numéro de rue) pour la livraison', excerpt: 'Les adresses sont décrites par des repères. Comment structurer les données dans PostGIS ?',
      tags: ['postgresql', 'postgis', 'logistique'], country: 'CM', author: 'p-yves', votes: 12, answers: 3, accepted: true, date: '2026-11-03', env: 'PostGIS 3.4', resource: 'r-harmattan', project: null },
    { id: 'q-docker-traefik', lang: 'en', title: 'Traefik returns 502 behind Docker Compose after a container restart', excerpt: 'The router label is correct but the service is not re-discovered until Traefik restarts. What am I missing?',
      tags: ['docker', 'traefik'], country: 'MA', author: 'p-imane', votes: 5, answers: 1, accepted: false, date: '2026-11-02', env: 'Traefik 3', resource: null, project: null }
  ];

  M.resources = [
    { id: 'r-kalimba', name: 'Kalimba Pay API', provider: 'Kalimba Pay (fictif)', category: 'payment', countries: ['BF', 'CI', 'SN', 'GH'], sdk: ['go', 'node', 'php'], version: 'v2.3',
      summary: 'Collecte Mobile Money et webhooks signés. Fiche fictive créée pour le prototype.', verified: true, verifiedAt: '2026-10-28', sources: 2, docs: 'https://example.org/docs/kalimba', questions: 6 },
    { id: 'r-sahel-sms', name: 'Sahel SMS Gateway', provider: 'Sahel Telecom Labs (fictif)', category: 'sms', countries: ['SN', 'CI', 'BF', 'CM'], sdk: ['node', 'python'], version: 'v1.8',
      summary: 'SMS transactionnels et USSD. Couverture opérateurs déclarée par le fournisseur.', verified: true, verifiedAt: '2026-10-12', sources: 1, docs: 'https://example.org/docs/sahel-sms', questions: 4 },
    { id: 'r-baobab-kyc', name: 'Baobab Identity Check', provider: 'Baobab Identité (fictif)', category: 'kyc', countries: ['NG', 'GH', 'KE'], sdk: ['python', 'node'], version: 'v0.9 beta',
      summary: 'Vérification de documents et de numéros d’identité. Limites et quotas documentés dans la source.', verified: false, verifiedAt: '2026-08-30', sources: 1, docs: 'https://example.org/docs/baobab', questions: 2 },
    { id: 'r-harmattan', name: 'Harmattan Delivery API', provider: 'Harmattan Logistique (fictif)', category: 'logistics', countries: ['CM', 'NG', 'KE'], sdk: ['python', 'php'], version: 'v1.2',
      summary: 'Suivi de colis et zones de livraison par repères urbains.', verified: true, verifiedAt: '2026-09-18', sources: 2, docs: 'https://example.org/docs/harmattan', questions: 3 },
    { id: 'r-savane-ledger', name: 'Savane Ledger SDK', provider: 'Savane Open Labs (fictif)', category: 'fintech', countries: ['KE', 'RW', 'GH'], sdk: ['go', 'rust'], version: 'v0.4',
      summary: 'Bibliothèque open source de grand livre double entrée. Licence indiquée dans le dépôt.', verified: true, verifiedAt: '2026-10-02', sources: 1, docs: 'https://example.org/docs/savane-ledger', questions: 1 },
    { id: 'r-acacia-otp', name: 'Acacia OTP', provider: 'Acacia Secure (fictif)', category: 'kyc', countries: ['MA', 'SN', 'CI'], sdk: ['node', 'php', 'python'], version: 'v3.0',
      summary: 'Codes à usage unique par SMS ou appel vocal. Aucune garantie de délivrabilité n’est affirmée ici.', verified: false, verifiedAt: '2026-07-21', sources: 1, docs: 'https://example.org/docs/acacia-otp', questions: 0 }
  ];

  M.projects = [
    { id: 'pr-webhook-kit', name: 'webhook-kit', repo: 'github.com/exemple-226/webhook-kit', summary: 'Bibliothèque Go pour valider et rejouer des webhooks de paiement.', lang: 'fr',
      stack: ['go', 'postgresql'], status: 'active', country: 'BF', owner: 'p-aminata', needs: ['Go / backend', 'Documentation'], applicants: 2, updated: '2026-11-08', resource: 'r-kalimba', question: 'q-webhook-sig' },
    { id: 'pr-offline-drafts', name: 'offline-drafts', repo: 'github.com/exemple-226/offline-drafts', summary: 'File d’attente locale pour brouillons hors ligne, pensée pour réseaux instables.', lang: 'fr',
      stack: ['typescript', 'react'], status: 'idea', country: 'CM', owner: 'p-yves', needs: ['Frontend React', 'Tests'], applicants: 0, updated: '2026-11-04', resource: null, question: 'q-offline-sync' },
    { id: 'pr-ussd-flow', name: 'ussd-flow', repo: 'github.com/exemple-226/ussd-flow', summary: 'Moteur d’états pour parcours USSD avec persistance de session.', lang: 'en',
      stack: ['php', 'redis'], status: 'active', country: 'SN', owner: 'p-fatou', needs: ['PHP / backend'], applicants: 1, updated: '2026-11-01', resource: 'r-sahel-sms', question: 'q-ussd-timeout' },
    { id: 'pr-docs-fr', name: 'docs-afrique-fr', repo: 'github.com/exemple-226/docs-afrique-fr', summary: 'Documentation technique en français pour intégrations locales.', lang: 'fr',
      stack: ['markdown', 'ci'], status: 'seeking', country: 'KE', owner: 'p-wanjiru', needs: ['Rédaction technique', 'Relecture FR'], applicants: 3, updated: '2026-10-27', resource: null, question: null },
    { id: 'pr-geo-informal', name: 'geo-informal', repo: 'github.com/exemple-226/geo-informal', summary: 'Modèle de données PostGIS pour adresses décrites par repères.', lang: 'en',
      stack: ['postgresql', 'postgis', 'python'], status: 'active', country: 'CM', owner: 'p-yves', needs: ['SQL / PostGIS'], applicants: 0, updated: '2026-10-30', resource: 'r-harmattan', question: 'q-geocode' }
  ];

  M.byId = function (arr, id) { return arr.filter(function (x) { return x.id === id; })[0] || null; };
  M.person = function (id) { return M.byId(M.people || [], id); };
  M.countryName = function (code) { var c = M.byId.call(null, M.countries.map(function (x) { return { id: x.code, x: x }; }), code); return c ? c.x[window.I18N ? I18N.lang : 'fr'] : code; };
  M.label = function (o) { return o[window.I18N ? I18N.lang : 'fr']; };
  M.fmtDate = function (iso) { try { return new Date(iso + 'T12:00:00').toLocaleDateString(I18N.lang === 'fr' ? 'fr-FR' : 'en-GB', { day: 'numeric', month: 'short', year: 'numeric' }); } catch (e) { return iso; } };
})(window.MOCK);
