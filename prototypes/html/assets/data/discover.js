/* Discover lane mock data (FICTIONAL). Every provider, endpoint, snippet and source is invented for the prototype.
   No real company, API or integration is implied. Free text (summary, auth/limits detail, notes, pitfalls) is USER CONTENT stored once in its original language: { lang, text } (shown with the original-language badge, never swapped to the UI language).
   Taxonomy labels ({fr,en}) are UI strings. Extends window.MOCK.resources (shared) with catalogue-only fields. */
window.MOCK = window.MOCK || {};
(function (M) {
  var D = M.discover = {};
  D.today = '2026-11-11';          /* fixed reference date so "stale" is deterministic in the prototype */
  D.staleDays = 60;
  D.categories = [
    { id: 'payment', fr: 'Paiement / Mobile Money', en: 'Payments / Mobile Money' },
    { id: 'kyc', fr: 'Identité / KYC', en: 'Identity / KYC' },
    { id: 'sms', fr: 'SMS / USSD', en: 'SMS / USSD' },
    { id: 'maps', fr: 'Cartes / géo', en: 'Maps / geo' },
    { id: 'gov', fr: 'Données publiques', en: 'Government open data' },
    { id: 'logistics', fr: 'Logistique', en: 'Logistics' },
    { id: 'fintech', fr: 'Fintech', en: 'Fintech' }
  ];
  D.types = [{ id: 'api', fr: 'API', en: 'API' }, { id: 'sdk', fr: 'SDK', en: 'SDK' }, { id: 'tool', fr: 'Outil', en: 'Tool' }, { id: 'project', fr: 'Projet open source', en: 'Open-source project' }];
  D.kinds = { official: { fr: 'Documentation officielle', en: 'Official documentation', cls: 'badge-verified', icon: 'book' },
    changelog: { fr: 'Journal des changements', en: 'Changelog', cls: 'badge-verified', icon: 'git-branch' },
    community: { fr: 'Confirmation communautaire', en: 'Community confirmation', cls: 'badge-accepted', icon: 'users' },
    claim: { fr: 'Déclaré par le fournisseur, non vérifié', en: 'Provider claim, not verified', cls: 'badge-warn', icon: 'alert' } };

  function base(id) { return M.byId(M.resources, id); }
  function mk(id, o) { var b = base(id) || {}; var r = {}; Object.keys(b).forEach(function (k) { r[k] = b[k]; }); Object.keys(o).forEach(function (k) { r[k] = o[k]; }); r.summaryLang = o.summaryLang || 'fr'; return r; }

  D.items = [
    mk('r-kalimba', { type: 'api', docLang: ['fr', 'en'], auth: { model: 'hmac', lang: 'fr', text: 'Clé secrète + signature HMAC-SHA256 des webhooks (en-tête X-Signature).' },
      sources: [{ kind: 'official', title: 'Guide Collecte et webhooks', url: 'https://example.org/docs/kalimba/collect', published: '2026-10-20', checked: '2026-10-28', lang: 'fr' },
        { kind: 'changelog', title: 'Changelog v2.3', url: 'https://example.org/docs/kalimba/changelog', published: '2026-10-14', checked: '2026-10-28', lang: 'en' },
        { kind: 'community', title: 'Confirmation : corps brut requis pour la signature', url: 'pages/question.html?id=q-webhook-sig', published: '2026-11-09', checked: '2026-11-10', lang: 'fr', internal: true }],
      limits: { lang: 'fr', text: 'Callbacks rejoués jusqu’à 5 fois ; pas de garantie d’ordre.' },
      pitfalls: [{ lang: 'fr', text: 'Calculer le HMAC sur le corps brut, pas sur le JSON re-sérialisé.' }, { lang: 'en', text: 'Callbacks can arrive twice: store an idempotency key per transaction.' }, { lang: 'fr', text: 'Le pays CI exige un identifiant de marchand distinct (selon la source officielle).' }],
      notes: '## Collecter un paiement\nCréez une intention de paiement côté serveur, puis redirigez l’utilisateur vers la page hébergée.\n\n- Utilisez la clé **test** en développement.\n- Vérifiez la signature avant tout traitement.\n- Répondez `200` rapidement ; le traitement lourd se fait en tâche de fond.\n\n## Webhooks\nChaque callback contient `event`, `reference` et `amount`. Voir les questions liées pour la déduplication.',
      notesLang: 'fr', snippet: { lang: 'go', code: '// Fictional example: verify a Kalimba Pay webhook\nfunc verify(secret, body []byte, sig string) bool {\n  mac := hmac.New(sha256.New, secret)\n  mac.Write(body) // raw body, not re-encoded JSON\n  want := hex.EncodeToString(mac.Sum(nil))\n  return hmac.Equal([]byte(want), []byte(sig))\n}' },
      questions: ['q-webhook-sig', 'q-idempotency'], projects: ['pr-webhook-kit'] }),
    mk('r-sahel-sms', { type: 'api', docLang: ['fr'], auth: { model: 'apikey', lang: 'fr', text: 'Clé d’API en en-tête Authorization (Bearer).' },
      sources: [{ kind: 'official', title: 'Référence SMS et USSD', url: 'https://example.org/docs/sahel-sms/ref', published: '2026-09-30', checked: '2026-10-12', lang: 'fr' },
        { kind: 'claim', title: 'Page de couverture opérateurs', url: 'https://example.org/docs/sahel-sms/coverage', published: '2026-08-01', checked: '2026-10-12', lang: 'fr' }],
      limits: { lang: 'fr', text: 'Sessions USSD limitées à 30 secondes côté opérateur.' },
      pitfalls: [{ lang: 'en', text: 'Accented French characters switch the message to UCS-2 and multiply segments.' }, { lang: 'fr', text: 'La couverture opérateurs est déclarée par le fournisseur, non confirmée par des tiers.' }],
      notes: '## Envoyer un SMS\nAppelez `POST /v1/messages` avec `to`, `text` et `sender`.\n\n- Restez en GSM-7 pour limiter les segments.\n- Conservez l’état de session USSD côté serveur.',
      notesLang: 'fr', snippet: { lang: 'js', code: '// Fictional example: send one SMS\nconst res = await fetch("https://api.example.org/sahel-sms/v1/messages", {\n  method: "POST",\n  headers: { Authorization: `Bearer ${process.env.SAHEL_KEY}` },\n  body: JSON.stringify({ to: "+226XXXXXXXX", text: "Code 4821" })\n});' },
      questions: ['q-ussd-timeout', 'q-sms-encoding'], projects: ['pr-ussd-flow'] }),
    mk('r-baobab-kyc', { type: 'api', docLang: ['en'], summaryLang: 'fr', auth: { model: 'oauth2', lang: 'fr', text: 'OAuth 2.0 client credentials, jetons de 1 heure.' },
      sources: [{ kind: 'official', title: 'Identity Check quickstart (beta)', url: 'https://example.org/docs/baobab/quickstart', published: '2026-08-20', checked: '2026-08-30', lang: 'en' }],
      limits: { lang: 'fr', text: 'Quota bêta : 100 vérifications par jour.' },
      pitfalls: [{ lang: 'en', text: 'Photos with glare fail the quality check; validate client-side first.' }],
      notes: '## Verify a document\nUpload the front and back image, then poll the check until it reaches `complete`.\n\n- Beta: fields may change without notice.',
      notesLang: 'en', snippet: { lang: 'py', code: '# Fictional example: create a check\nr = requests.post(\n    "https://api.example.org/baobab/v0/checks",\n    headers={"Authorization": f"Bearer {token}"},\n    files={"front": open("front.jpg", "rb")},\n)\nprint(r.json()["status"])' },
      questions: ['q-kyc-ocr'], projects: [] }),
    mk('r-harmattan', { type: 'api', docLang: ['fr', 'en'], auth: { model: 'apikey', lang: 'fr', text: 'Clé d’API par environnement (test / production).' },
      sources: [{ kind: 'official', title: 'Zones de livraison par repères', url: 'https://example.org/docs/harmattan/zones', published: '2026-09-10', checked: '2026-09-18', lang: 'fr' },
        { kind: 'community', title: 'Retour d’intégration à Douala', url: 'pages/question.html?id=q-geocode', published: '2026-11-03', checked: '2026-11-09', lang: 'fr', internal: true }],
      limits: { lang: 'fr', text: 'Pas de numéro de rue : la zone est décrite par un repère.' },
      pitfalls: [{ lang: 'fr', text: 'Normaliser les repères (casse, accents) avant de rechercher une zone.' }],
      notes: '## Suivre un colis\nInterrogez `GET /v1/parcels/{id}` ; les statuts sont `created`, `in_transit`, `delivered`.',
      notesLang: 'fr', snippet: { lang: 'sh', code: '# Fictional example\ncurl -H "X-Api-Key: $HARMATTAN_KEY" \\\n  https://api.example.org/harmattan/v1/parcels/PCL-0001' },
      questions: ['q-geocode'], projects: ['pr-geo-informal'] }),
    mk('r-savane-ledger', { type: 'sdk', docLang: ['en'], summaryLang: 'fr', auth: { model: 'none', lang: 'fr', text: 'Bibliothèque locale : pas d’authentification réseau.' },
      sources: [{ kind: 'official', title: 'README du dépôt (licence indiquée)', url: 'https://example.org/docs/savane-ledger/readme', published: '2026-09-28', checked: '2026-10-02', lang: 'en' }],
      limits: { lang: 'fr', text: 'Version 0.x : l’API publique peut changer.' },
      pitfalls: [{ lang: 'en', text: 'Amounts are integers in minor units; never pass floats.' }],
      notes: '## Post an entry\nEach transaction needs balanced debit and credit lines.\n\n- Use integer minor units.\n- Wrap postings in a database transaction.',
      notesLang: 'en', snippet: { lang: 'go', code: '// Fictional example\ntx := ledger.NewTx("pay-001")\ntx.Debit("cash", 1500)\ntx.Credit("revenue", 1500)\nerr := book.Post(ctx, tx)' },
      questions: [], projects: [] }),
    mk('r-acacia-otp', { type: 'api', docLang: ['fr', 'en'], auth: { model: 'apikey', lang: 'fr', text: 'Clé d’API serveur ; ne jamais l’exposer côté client.' },
      sources: [{ kind: 'claim', title: 'Page produit : délivrabilité annoncée', url: 'https://example.org/docs/acacia-otp/product', published: '2026-06-30', checked: '2026-07-21', lang: 'fr' }],
      limits: { lang: 'fr', text: 'Aucune garantie de délivrabilité n’est affirmée ici.' },
      pitfalls: [{ lang: 'fr', text: 'Source unique et ancienne : vérifier avant de dépendre de cette fiche.' }],
      notes: '## Envoyer un code\nDemandez un code avec `POST /v3/otp`, puis vérifiez-le avec `POST /v3/otp/verify`.',
      notesLang: 'fr', snippet: { lang: 'php', code: '<?php\n// Fictional example\n$code = $acacia->otp()->send("+212XXXXXXXXX");\n$ok = $acacia->otp()->verify($code->id, $_POST["code"]);' },
      questions: [], projects: [] }),
    { id: 'r-nile-maps', name: 'Nile Maps Tiles', provider: 'Nile Cartographie (fictif)', type: 'api', category: 'maps', countries: ['KE', 'RW', 'GH'], docLang: ['en'],
      version: 'v1.1', summary: 'Tuiles vectorielles et géocodage inverse pour les villes d’Afrique de l’Est. Fiche fictive.', summaryLang: 'fr', verified: true, verifiedAt: '2026-10-22', sources: 2,
      auth: { model: 'apikey', lang: 'fr', text: 'Clé d’API en paramètre de requête.' } },
    { id: 'r-sahel-opendata', name: 'Sahel Open Data Portal', provider: 'Portail fictif des données publiques', type: 'api', category: 'gov', countries: ['BF', 'SN', 'CI'], docLang: ['fr'],
      version: 'v1', summary: 'Jeux de données publics fictifs (budgets, écoles) en JSON et CSV.', summaryLang: 'fr', verified: false, verifiedAt: '2026-06-02', sources: 1,
      auth: { model: 'none', lang: 'fr', text: 'Accès ouvert, limité à 60 requêtes par minute.' } },
    { id: 'r-griot-ussd', name: 'Griot USSD Kit', provider: 'Griot Labs (fictif)', type: 'sdk', category: 'sms', countries: ['SN', 'CI', 'BF'], docLang: ['fr', 'en'],
      version: 'v0.7', summary: 'SDK pour modéliser des menus USSD avec états persistants. Exemple fictif.', summaryLang: 'fr', verified: true, verifiedAt: '2026-11-01', sources: 2,
      auth: { model: 'none', lang: 'fr', text: 'Bibliothèque locale ; l’agrégateur USSD gère l’authentification.' } },
    { id: 'r-ubuntu-id', name: 'Ubuntu ID Lookup', provider: 'Ubuntu Registre (fictif)', type: 'api', category: 'kyc', countries: ['KE', 'RW'], docLang: ['en'],
      version: 'v2.0', summary: 'Consultation de numéros d’identité de test. Fiche fictive, aucun registre réel.', summaryLang: 'fr', verified: true, verifiedAt: '2026-09-05', sources: 1,
      auth: { model: 'oauth2', lang: 'fr', text: 'OAuth 2.0 avec portée limitée par pays.' } },
    { id: 'r-geojson-cli', name: 'geojson-lint', provider: 'Communauté fictive', type: 'tool', category: 'maps', countries: ['CM', 'MA'], docLang: ['en'],
      version: 'v0.3', summary: 'Outil en ligne de commande qui valide des fichiers GeoJSON. Exemple fictif.', summaryLang: 'fr', verified: false, verifiedAt: '2026-05-14', sources: 1,
      auth: { model: 'none', lang: 'fr', text: 'Outil local, aucune authentification.' } }
  ];

  /* Fill the lighter items with complete evidence fields (kept short; still fictional). */
  var extra = {
    'r-nile-maps': { sourcesList: [{ kind: 'official', title: 'Tiles quickstart', url: 'https://example.org/docs/nile-maps/tiles', published: '2026-10-15', checked: '2026-10-22', lang: 'en' }, { kind: 'changelog', title: 'Changelog v1.1', url: 'https://example.org/docs/nile-maps/changelog', published: '2026-10-10', checked: '2026-10-22', lang: 'en' }],
      limits: { lang: 'fr', text: 'Limite de 5 000 tuiles par jour sur l’offre gratuite.' }, pitfalls: [{ lang: 'en', text: 'Zoom levels above 16 are not available outside capital cities.' }],
      notes: '## Display a map\nRequest tiles from `/v1/tiles/{z}/{x}/{y}` and add attribution.\n\n- Cache tiles client-side.', notesLang: 'en', snippet: { lang: 'js', code: '// Fictional example\nconst url = "https://api.example.org/nile/v1/tiles/{z}/{x}/{y}?key=" + KEY;\nmap.addSource("nile", { type: "vector", tiles: [url] });' }, questions: ['q-geocode'], projects: ['pr-geo-informal'] },
    'r-sahel-opendata': { sourcesList: [{ kind: 'claim', title: 'Description du portail', url: 'https://example.org/docs/sahel-opendata/about', published: '2026-05-20', checked: '2026-06-02', lang: 'fr' }],
      limits: { lang: 'fr', text: 'Les jeux de données n’ont pas de calendrier de mise à jour garanti.' }, pitfalls: [{ lang: 'fr', text: 'Vérifier la date de publication de chaque jeu avant de l’utiliser.' }],
      notes: '## Lire un jeu de données\nAppelez `GET /v1/datasets/{id}.json` ; le CSV est disponible avec l’extension `.csv`.', notesLang: 'fr', snippet: { lang: 'sh', code: '# Fictional example\ncurl https://api.example.org/sahel-opendata/v1/datasets/schools.json' }, questions: [], projects: [] },
    'r-griot-ussd': { sourcesList: [{ kind: 'official', title: 'Guide des menus USSD', url: 'https://example.org/docs/griot/menus', published: '2026-10-28', checked: '2026-11-01', lang: 'fr' }, { kind: 'community', title: 'Retour : état de session et expiration', url: 'pages/question.html?id=q-ussd-timeout', published: '2026-11-08', checked: '2026-11-10', lang: 'fr', internal: true }],
      limits: { lang: 'fr', text: 'Dépend de l’agrégateur USSD choisi.' }, pitfalls: [{ lang: 'fr', text: 'Sauvegarder l’état avant chaque réponse pour survivre aux coupures de session.' }],
      notes: '## Définir un menu\nChaque écran est un état ; les transitions sont déclarées avec `on(input)`.', notesLang: 'fr', snippet: { lang: 'js', code: '// Fictional example\nconst menu = ussd.menu("home")\n  .on("1", "balance")\n  .on("2", "send");' }, questions: ['q-ussd-timeout'], projects: ['pr-ussd-flow'] },
    'r-ubuntu-id': { sourcesList: [{ kind: 'official', title: 'Lookup API reference', url: 'https://example.org/docs/ubuntu-id/ref', published: '2026-08-28', checked: '2026-09-05', lang: 'en' }],
      limits: { lang: 'fr', text: 'Environnement de test uniquement dans cet exemple fictif.' }, pitfalls: [{ lang: 'en', text: 'Scopes are per country; a Kenyan token is rejected for RW lookups.' }],
      notes: '## Look up a test ID\nSend `GET /v2/ids/{number}` with a country-scoped token.', notesLang: 'en', snippet: { lang: 'py', code: '# Fictional example\nr = requests.get(f"https://api.example.org/ubuntu/v2/ids/{n}",\n                 headers={"Authorization": f"Bearer {tok}"})' }, questions: ['q-kyc-ocr'], projects: [] },
    'r-geojson-cli': { sourcesList: [{ kind: 'claim', title: 'README du dépôt', url: 'https://example.org/docs/geojson-lint/readme', published: '2026-05-10', checked: '2026-05-14', lang: 'en' }],
      limits: { lang: 'fr', text: 'Valide la structure, pas l’exactitude des coordonnées.' }, pitfalls: [{ lang: 'en', text: 'Large files are read fully into memory.' }],
      notes: '## Validate a file\nRun `geojson-lint file.geojson`; exit code 1 means invalid.', notesLang: 'en', snippet: { lang: 'sh', code: '# Fictional example\ngeojson-lint districts.geojson' }, questions: [], projects: [] }
  };
  D.items.forEach(function (r) {
    var e = extra[r.id]; if (!e) return;
    r.sources = e.sourcesList; r.limits = e.limits; r.pitfalls = e.pitfalls; r.notes = e.notes; r.notesLang = e.notesLang; r.snippet = e.snippet; r.questions = e.questions; r.projects = e.projects;
    delete r.sourcesList;
  });
  D.items.forEach(function (r) {
    r.sdk = r.sdk || [];
    var off = r.sources.filter(function (s) { return !s.internal && s.kind !== 'community'; });
    r.verifiedAt = r.verifiedAt || off.reduce(function (m, s) { return s.checked > m ? s.checked : m; }, '0000');
    r.hasOfficial = r.sources.some(function (s) { return s.kind === 'official' || s.kind === 'changelog'; });
  });
  /* status: verified = official/changelog source checked within 60 days; stale = older; unverified = provider claim only */
  D.status = function (r) {
    if (!r.hasOfficial) return 'unverified';
    var days = (new Date(D.today) - new Date(r.verifiedAt)) / 864e5; return days > D.staleDays ? 'stale' : 'verified';
  };

  /* Community reviews (FICTIONAL sample, PRD 5.2 community feedback pulled forward from P1). Reviewers come from MOCK.people.
     Text is USER CONTENT in its original language. Aggregates are computed from this sample only. */
  var RV = {
    'r-kalimba': [
      ['p-aminata', 5, '2026-10-30', 'fr', 'Intégration en une demi-journée. La signature HMAC des webhooks est bien documentée et le bac à sable se comporte comme la production.', 1, 9],
      ['p-kofi', 4, '2026-10-12', 'en', 'Solid API and the Go SDK is clean. Wish idempotency keys were documented on the refund endpoint, I found out by trial and error.', 1, 6],
      ['p-salif', 2, '2026-09-21', 'fr', 'Le SDK PHP est en retard sur l’API : deux champs de la v2.3 manquent. Le support a répondu en deux jours, mais j’ai dû écrire mes propres appels.', 0, 11],
      ['p-fatou', 4, '2026-09-02', 'fr', 'Très bonne couverture au Sénégal et en Côte d’Ivoire. Les limites de débit sont strictes en bac à sable, prévoyez de la mise en file.', 0, 3],
      ['p-chinedu', 3, '2026-08-14', 'en', 'Works well, but the dashboard logs expire after 7 days, which made debugging a late webhook harder than it should be.', 0, 2]],
    'r-sahel-sms': [
      ['p-fatou', 5, '2026-10-18', 'fr', 'Livraison fiable vers les opérateurs locaux, et le statut de livraison arrive bien par webhook. Parfait pour nos codes de confirmation.', 1, 7],
      ['p-yves', 3, '2026-10-02', 'fr', 'Bon service, mais les délais augmentent le soir aux heures de pointe. Prévoyez un second canal pour les messages urgents.', 0, 5],
      ['p-wanjiru', 4, '2026-09-10', 'en', 'Simple REST interface and honest documentation about sender-ID rules per country. Pricing page is hard to find though.', 0, 2]],
    'r-baobab-kyc': [
      ['p-chinedu', 4, '2026-10-25', 'en', 'Document checks are accurate for Nigerian and Ghanaian IDs. Selfie matching occasionally fails in low light, so keep a manual review path.', 1, 8],
      ['p-aminata', 2, '2026-10-01', 'fr', 'Les réponses d’erreur sont peu explicites : un « rejet » sans motif exploitable. Difficile d’expliquer à l’utilisateur ce qu’il doit corriger.', 1, 10],
      ['p-imane', 4, '2026-08-29', 'fr', 'Conformité claire et hébergement documenté. Bon compromis pour une équipe sans juriste dédié.', 0, 4]],
    'r-harmattan': [
      ['p-yves', 5, '2026-10-20', 'fr', 'Le suivi de colis par point de repère fonctionne vraiment à Douala, là où les adresses n’ont pas de numéro de rue.', 1, 6],
      ['p-kofi', 3, '2026-09-15', 'en', 'Tracking events are useful, but the webhook retries stop after two attempts. We lost a delivery update during an outage on our side.', 0, 4],
      ['p-salif', 4, '2026-08-05', 'fr', 'API propre et exemples utiles. Il manque une collection Postman officielle.', 0, 1]],
    'r-savane-ledger': [
      ['p-wanjiru', 5, '2026-10-27', 'en', 'Great test suite and readable code. We use it as the reference for double-entry examples in our workshops.', 1, 7],
      ['p-aminata', 4, '2026-09-30', 'fr', 'SDK Go très propre, une vraie référence. La documentation des migrations pourrait être plus détaillée.', 1, 5],
      ['p-imane', 3, '2026-09-04', 'en', 'Does what it says, but no official support channel: issues sometimes wait weeks for a reply.', 0, 3]],
    'r-acacia-otp': [
      ['p-fatou', 4, '2026-10-22', 'fr', 'Envoi de codes par SMS et USSD stable. L’expiration configurable est un vrai plus pour les réseaux lents.', 1, 5],
      ['p-chinedu', 2, '2026-10-05', 'en', 'Rate limiting is aggressive and the error does not tell you how long to wait. Users got locked out during our launch.', 0, 9],
      ['p-yves', 4, '2026-08-19', 'fr', 'Intégration simple, bon rapport qualité prix pour un MVP.', 0, 2]],
    'r-nile-maps': [
      ['p-yves', 4, '2026-10-16', 'fr', 'Tuiles rapides et détaillées sur Douala et Lagos. Le style sombre manque encore de contraste pour les petites routes.', 1, 6],
      ['p-wanjiru', 5, '2026-09-18', 'en', 'Vector tiles with proper attribution rules, and the Rust tile client worked on the first try.', 0, 4]],
    'r-sahel-opendata': [
      ['p-aminata', 3, '2026-10-08', 'fr', 'Données utiles mais jeux de données parfois non mis à jour depuis des mois. Vérifiez la date de chaque fichier.', 0, 8],
      ['p-imane', 4, '2026-09-12', 'en', 'Good CSV and JSON exports with clear licences. Some columns change names between releases.', 0, 3]],
    'r-griot-ussd': [
      ['p-fatou', 5, '2026-10-29', 'fr', 'Le moteur d’états gère bien les coupures de session des opérateurs. Nous l’utilisons en production pour un service de paiement.', 1, 12],
      ['p-salif', 4, '2026-09-25', 'fr', 'Backend PHP encore jeune mais utilisable. À relire avant la mise en production.', 0, 3],
      ['p-kofi', 2, '2026-09-01', 'en', 'Docs are thin outside the quick start. I needed to read the source to understand session storage.', 0, 5]],
    'r-geojson-cli': [
      ['p-yves', 4, '2026-10-10', 'fr', 'Petit outil pratique pour valider des fichiers GeoJSON avant import. Les messages d’erreur indiquent la ligne exacte.', 0, 2]]
  };
  D.reviews = {};
  Object.keys(RV).forEach(function (k) {
    D.reviews[k] = RV[k].map(function (a, i) { return { id: k + '-' + (i + 1), person: M.byId(M.people, a[0]), country: (M.byId(M.people, a[0]) || {}).country, rating: a[1], date: a[2], lang: a[3], text: a[4], verified: !!a[5], helpful: a[6] }; });
  });
  D.stats = function (id, extra) {
    var l = (D.reviews[id] || []).concat(extra || []); if (!l.length) return null;
    return { avg: l.reduce(function (s, x) { return s + x.rating; }, 0) / l.length, n: l.length };
  };
  D.items.forEach(function (r) { var st = D.stats(r.id); r._rating = st ? st.avg : 0; r._nrev = st ? st.n : 0; });

  /* Open-source projects surfaced in the catalogue for visibility. They stay canonical in Build (pages/project.html); nothing is copied:
     these entries are derived read-only from MOCK.build (data/build.js must be loaded first) and link to the project page. */
  var PROJ_MAP = { 'pr-webhook-kit': { category: 'payment', sdk: ['go'] }, 'pr-ussd-flow': { category: 'sms', sdk: ['php'] }, 'pr-geo-informal': { category: 'maps', sdk: ['python'] },
    'pr-gsm7-check': { category: 'sms', sdk: ['node'] }, 'pr-ledger-lite': { category: 'fintech', sdk: ['go'] } };
  D.projectItems = function () {
    if (!M.build || !M.build.projects) return [];
    return M.build.projects().filter(function (p) { return PROJ_MAP[p.id]; }).map(function (p) {
      var o = M.byId(M.people, p.owner), m = PROJ_MAP[p.id];
      return { id: p.id, kind: 'project', type: 'project', name: p.name, provider: o ? o.name : '', category: m.category, countries: [p.country], sdk: m.sdk, docLang: [p.lang], summary: p.summary, summaryLang: p.lang,
        license: p.license, stage: p.stage, version: '', verifiedAt: p.updated, updated: p.updated, href: 'project.html?id=' + p.id, sources: [], _rating: 0, _nrev: 0 };
    });
  };
  D.get = function (id) { return M.byId(D.items, id); };
  D.cat = function (id) { return M.byId(D.categories, id); };
})(window.MOCK);
