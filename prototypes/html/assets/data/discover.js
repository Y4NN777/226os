/* Discover lane mock data (FICTIONAL). Every provider, endpoint, snippet and source is invented for the prototype.
   No real company, API or integration is implied. Free text (summary, notes, pitfalls) is USER CONTENT with its own `lang`.
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
  D.types = [{ id: 'api', fr: 'API', en: 'API' }, { id: 'sdk', fr: 'SDK', en: 'SDK' }, { id: 'tool', fr: 'Outil', en: 'Tool' }];
  D.kinds = { official: { fr: 'Documentation officielle', en: 'Official documentation', cls: 'badge-verified', icon: 'book' },
    changelog: { fr: 'Journal des changements', en: 'Changelog', cls: 'badge-verified', icon: 'git-branch' },
    community: { fr: 'Confirmation communautaire', en: 'Community confirmation', cls: 'badge-accepted', icon: 'users' },
    claim: { fr: 'Affirmation du fournisseur, non vérifiée', en: 'Provider claim, not verified', cls: 'badge-warn', icon: 'alert' } };

  function base(id) { return M.byId(M.resources, id); }
  function mk(id, o) { var b = base(id) || {}; var r = {}; Object.keys(b).forEach(function (k) { r[k] = b[k]; }); Object.keys(o).forEach(function (k) { r[k] = o[k]; }); r.summaryLang = o.summaryLang || 'fr'; return r; }

  D.items = [
    mk('r-kalimba', { type: 'api', docLang: ['fr', 'en'], auth: { model: 'hmac', fr: 'Clé secrète + signature HMAC-SHA256 des webhooks (en-tête X-Signature).', en: 'Secret key + HMAC-SHA256 webhook signature (X-Signature header).' },
      sources: [{ kind: 'official', title: 'Guide Collecte et webhooks', url: 'https://example.org/docs/kalimba/collect', published: '2026-10-20', checked: '2026-10-28', lang: 'fr' },
        { kind: 'changelog', title: 'Changelog v2.3', url: 'https://example.org/docs/kalimba/changelog', published: '2026-10-14', checked: '2026-10-28', lang: 'en' },
        { kind: 'community', title: 'Confirmation : corps brut requis pour la signature', url: 'pages/question.html?id=q-webhook-sig', published: '2026-11-09', checked: '2026-11-10', lang: 'fr', internal: true }],
      limits: { fr: 'Callbacks rejoués jusqu’à 5 fois ; pas de garantie d’ordre.', en: 'Callbacks retried up to 5 times; no ordering guarantee.' },
      pitfalls: [{ lang: 'fr', text: 'Calculer le HMAC sur le corps brut, pas sur le JSON re-sérialisé.' }, { lang: 'en', text: 'Callbacks can arrive twice: store an idempotency key per transaction.' }, { lang: 'fr', text: 'Le pays CI exige un identifiant de marchand distinct (selon la source officielle).' }],
      notes: '## Collecter un paiement\nCréez une intention de paiement côté serveur, puis redirigez l’utilisateur vers la page hébergée.\n\n- Utilisez la clé **test** en développement.\n- Vérifiez la signature avant tout traitement.\n- Répondez `200` rapidement ; le traitement lourd se fait en tâche de fond.\n\n## Webhooks\nChaque callback contient `event`, `reference` et `amount`. Voir les questions liées pour la déduplication.',
      notesLang: 'fr', snippet: { lang: 'go', code: '// Fictional example: verify a Kalimba Pay webhook\nfunc verify(secret, body []byte, sig string) bool {\n  mac := hmac.New(sha256.New, secret)\n  mac.Write(body) // raw body, not re-encoded JSON\n  want := hex.EncodeToString(mac.Sum(nil))\n  return hmac.Equal([]byte(want), []byte(sig))\n}' },
      questions: ['q-webhook-sig', 'q-idempotency'], projects: ['pr-webhook-kit'] }),
    mk('r-sahel-sms', { type: 'api', docLang: ['fr'], auth: { model: 'apikey', fr: 'Clé d’API en en-tête Authorization (Bearer).', en: 'API key in the Authorization header (Bearer).' },
      sources: [{ kind: 'official', title: 'Référence SMS et USSD', url: 'https://example.org/docs/sahel-sms/ref', published: '2026-09-30', checked: '2026-10-12', lang: 'fr' },
        { kind: 'claim', title: 'Page de couverture opérateurs', url: 'https://example.org/docs/sahel-sms/coverage', published: '2026-08-01', checked: '2026-10-12', lang: 'fr' }],
      limits: { fr: 'Sessions USSD limitées à 30 secondes côté opérateur.', en: 'USSD sessions limited to 30 seconds on the operator side.' },
      pitfalls: [{ lang: 'en', text: 'Accented French characters switch the message to UCS-2 and multiply segments.' }, { lang: 'fr', text: 'La couverture opérateurs est déclarée par le fournisseur, non confirmée par des tiers.' }],
      notes: '## Envoyer un SMS\nAppelez `POST /v1/messages` avec `to`, `text` et `sender`.\n\n- Restez en GSM-7 pour limiter les segments.\n- Conservez l’état de session USSD côté serveur.',
      notesLang: 'fr', snippet: { lang: 'js', code: '// Fictional example: send one SMS\nconst res = await fetch("https://api.example.org/sahel-sms/v1/messages", {\n  method: "POST",\n  headers: { Authorization: `Bearer ${process.env.SAHEL_KEY}` },\n  body: JSON.stringify({ to: "+226XXXXXXXX", text: "Code 4821" })\n});' },
      questions: ['q-ussd-timeout', 'q-sms-encoding'], projects: ['pr-ussd-flow'] }),
    mk('r-baobab-kyc', { type: 'api', docLang: ['en'], summaryLang: 'fr', auth: { model: 'oauth2', fr: 'OAuth 2.0 client credentials, jetons de 1 heure.', en: 'OAuth 2.0 client credentials, one-hour tokens.' },
      sources: [{ kind: 'official', title: 'Identity Check quickstart (beta)', url: 'https://example.org/docs/baobab/quickstart', published: '2026-08-20', checked: '2026-08-30', lang: 'en' }],
      limits: { fr: 'Quota bêta : 100 vérifications par jour.', en: 'Beta quota: 100 checks per day.' },
      pitfalls: [{ lang: 'en', text: 'Photos with glare fail the quality check; validate client-side first.' }],
      notes: '## Verify a document\nUpload the front and back image, then poll the check until it reaches `complete`.\n\n- Beta: fields may change without notice.',
      notesLang: 'en', snippet: { lang: 'py', code: '# Fictional example: create a check\nr = requests.post(\n    "https://api.example.org/baobab/v0/checks",\n    headers={"Authorization": f"Bearer {token}"},\n    files={"front": open("front.jpg", "rb")},\n)\nprint(r.json()["status"])' },
      questions: ['q-kyc-ocr'], projects: [] }),
    mk('r-harmattan', { type: 'api', docLang: ['fr', 'en'], auth: { model: 'apikey', fr: 'Clé d’API par environnement (test / production).', en: 'API key per environment (test / production).' },
      sources: [{ kind: 'official', title: 'Zones de livraison par repères', url: 'https://example.org/docs/harmattan/zones', published: '2026-09-10', checked: '2026-09-18', lang: 'fr' },
        { kind: 'community', title: 'Retour d’intégration à Douala', url: 'pages/question.html?id=q-geocode', published: '2026-11-03', checked: '2026-11-09', lang: 'fr', internal: true }],
      limits: { fr: 'Pas de numéro de rue : la zone est décrite par un repère.', en: 'No street numbers: the area is described by a landmark.' },
      pitfalls: [{ lang: 'fr', text: 'Normaliser les repères (casse, accents) avant de rechercher une zone.' }],
      notes: '## Suivre un colis\nInterrogez `GET /v1/parcels/{id}` ; les statuts sont `created`, `in_transit`, `delivered`.',
      notesLang: 'fr', snippet: { lang: 'sh', code: '# Fictional example\ncurl -H "X-Api-Key: $HARMATTAN_KEY" \\\n  https://api.example.org/harmattan/v1/parcels/PCL-0001' },
      questions: ['q-geocode'], projects: ['pr-geo-informal'] }),
    mk('r-savane-ledger', { type: 'sdk', docLang: ['en'], summaryLang: 'fr', auth: { model: 'none', fr: 'Bibliothèque locale : pas d’authentification réseau.', en: 'Local library: no network authentication.' },
      sources: [{ kind: 'official', title: 'README du dépôt (licence indiquée)', url: 'https://example.org/docs/savane-ledger/readme', published: '2026-09-28', checked: '2026-10-02', lang: 'en' }],
      limits: { fr: 'Version 0.x : l’API publique peut changer.', en: 'Version 0.x: public API may change.' },
      pitfalls: [{ lang: 'en', text: 'Amounts are integers in minor units; never pass floats.' }],
      notes: '## Post an entry\nEach transaction needs balanced debit and credit lines.\n\n- Use integer minor units.\n- Wrap postings in a database transaction.',
      notesLang: 'en', snippet: { lang: 'go', code: '// Fictional example\ntx := ledger.NewTx("pay-001")\ntx.Debit("cash", 1500)\ntx.Credit("revenue", 1500)\nerr := book.Post(ctx, tx)' },
      questions: [], projects: [] }),
    mk('r-acacia-otp', { type: 'api', docLang: ['fr', 'en'], auth: { model: 'apikey', fr: 'Clé d’API serveur ; ne jamais l’exposer côté client.', en: 'Server API key; never expose it client-side.' },
      sources: [{ kind: 'claim', title: 'Page produit : délivrabilité annoncée', url: 'https://example.org/docs/acacia-otp/product', published: '2026-06-30', checked: '2026-07-21', lang: 'fr' }],
      limits: { fr: 'Aucune garantie de délivrabilité n’est affirmée ici.', en: 'No deliverability guarantee is stated here.' },
      pitfalls: [{ lang: 'fr', text: 'Source unique et ancienne : vérifier avant de dépendre de cette fiche.' }],
      notes: '## Envoyer un code\nDemandez un code avec `POST /v3/otp`, puis vérifiez-le avec `POST /v3/otp/verify`.',
      notesLang: 'fr', snippet: { lang: 'php', code: '<?php\n// Fictional example\n$code = $acacia->otp()->send("+212XXXXXXXXX");\n$ok = $acacia->otp()->verify($code->id, $_POST["code"]);' },
      questions: [], projects: [] }),
    { id: 'r-nile-maps', name: 'Nile Maps Tiles', provider: 'Nile Cartographie (fictif)', type: 'api', category: 'maps', countries: ['KE', 'RW', 'GH'], docLang: ['en'],
      version: 'v1.1', summary: 'Tuiles vectorielles et géocodage inverse pour les villes d’Afrique de l’Est. Fiche fictive.', summaryLang: 'fr', verified: true, verifiedAt: '2026-10-22', sources: 2,
      auth: { model: 'apikey', fr: 'Clé d’API en paramètre de requête.', en: 'API key as a query parameter.' } },
    { id: 'r-sahel-opendata', name: 'Sahel Open Data Portal', provider: 'Portail fictif des données publiques', type: 'api', category: 'gov', countries: ['BF', 'SN', 'CI'], docLang: ['fr'],
      version: 'v1', summary: 'Jeux de données publics fictifs (budgets, écoles) en JSON et CSV.', summaryLang: 'fr', verified: false, verifiedAt: '2026-06-02', sources: 1,
      auth: { model: 'none', fr: 'Accès ouvert, limité à 60 requêtes par minute.', en: 'Open access, limited to 60 requests per minute.' } },
    { id: 'r-griot-ussd', name: 'Griot USSD Kit', provider: 'Griot Labs (fictif)', type: 'sdk', category: 'sms', countries: ['SN', 'CI', 'BF'], docLang: ['fr', 'en'],
      version: 'v0.7', summary: 'SDK pour modéliser des menus USSD avec états persistants. Exemple fictif.', summaryLang: 'fr', verified: true, verifiedAt: '2026-11-01', sources: 2,
      auth: { model: 'none', fr: 'Bibliothèque locale ; l’agrégateur USSD gère l’authentification.', en: 'Local library; the USSD aggregator handles authentication.' } },
    { id: 'r-ubuntu-id', name: 'Ubuntu ID Lookup', provider: 'Ubuntu Registre (fictif)', type: 'api', category: 'kyc', countries: ['KE', 'RW'], docLang: ['en'],
      version: 'v2.0', summary: 'Consultation de numéros d’identité de test. Fiche fictive, aucun registre réel.', summaryLang: 'fr', verified: true, verifiedAt: '2026-09-05', sources: 1,
      auth: { model: 'oauth2', fr: 'OAuth 2.0 avec portée limitée par pays.', en: 'OAuth 2.0 with country-scoped permissions.' } },
    { id: 'r-geojson-cli', name: 'geojson-lint', provider: 'Communauté fictive', type: 'tool', category: 'maps', countries: ['CM', 'MA'], docLang: ['en'],
      version: 'v0.3', summary: 'Outil en ligne de commande qui valide des fichiers GeoJSON. Exemple fictif.', summaryLang: 'fr', verified: false, verifiedAt: '2026-05-14', sources: 1,
      auth: { model: 'none', fr: 'Outil local, aucune authentification.', en: 'Local tool, no authentication.' } }
  ];

  /* Fill the lighter items with complete evidence fields (kept short; still fictional). */
  var extra = {
    'r-nile-maps': { sourcesList: [{ kind: 'official', title: 'Tiles quickstart', url: 'https://example.org/docs/nile-maps/tiles', published: '2026-10-15', checked: '2026-10-22', lang: 'en' }, { kind: 'changelog', title: 'Changelog v1.1', url: 'https://example.org/docs/nile-maps/changelog', published: '2026-10-10', checked: '2026-10-22', lang: 'en' }],
      limits: { fr: 'Limite de 5 000 tuiles par jour sur l’offre gratuite.', en: '5,000 tiles per day on the free plan.' }, pitfalls: [{ lang: 'en', text: 'Zoom levels above 16 are not available outside capital cities.' }],
      notes: '## Display a map\nRequest tiles from `/v1/tiles/{z}/{x}/{y}` and add attribution.\n\n- Cache tiles client-side.', notesLang: 'en', snippet: { lang: 'js', code: '// Fictional example\nconst url = "https://api.example.org/nile/v1/tiles/{z}/{x}/{y}?key=" + KEY;\nmap.addSource("nile", { type: "vector", tiles: [url] });' }, questions: ['q-geocode'], projects: ['pr-geo-informal'] },
    'r-sahel-opendata': { sourcesList: [{ kind: 'claim', title: 'Description du portail', url: 'https://example.org/docs/sahel-opendata/about', published: '2026-05-20', checked: '2026-06-02', lang: 'fr' }],
      limits: { fr: 'Les jeux de données n’ont pas de calendrier de mise à jour garanti.', en: 'Datasets have no guaranteed update schedule.' }, pitfalls: [{ lang: 'fr', text: 'Vérifier la date de publication de chaque jeu avant de l’utiliser.' }],
      notes: '## Lire un jeu de données\nAppelez `GET /v1/datasets/{id}.json` ; le CSV est disponible avec l’extension `.csv`.', notesLang: 'fr', snippet: { lang: 'sh', code: '# Fictional example\ncurl https://api.example.org/sahel-opendata/v1/datasets/schools.json' }, questions: [], projects: [] },
    'r-griot-ussd': { sourcesList: [{ kind: 'official', title: 'Guide des menus USSD', url: 'https://example.org/docs/griot/menus', published: '2026-10-28', checked: '2026-11-01', lang: 'fr' }, { kind: 'community', title: 'Retour : état de session et expiration', url: 'pages/question.html?id=q-ussd-timeout', published: '2026-11-08', checked: '2026-11-10', lang: 'fr', internal: true }],
      limits: { fr: 'Dépend de l’agrégateur USSD choisi.', en: 'Depends on the chosen USSD aggregator.' }, pitfalls: [{ lang: 'fr', text: 'Sauvegarder l’état avant chaque réponse pour survivre aux coupures de session.' }],
      notes: '## Définir un menu\nChaque écran est un état ; les transitions sont déclarées avec `on(input)`.', notesLang: 'fr', snippet: { lang: 'js', code: '// Fictional example\nconst menu = ussd.menu("home")\n  .on("1", "balance")\n  .on("2", "send");' }, questions: ['q-ussd-timeout'], projects: ['pr-ussd-flow'] },
    'r-ubuntu-id': { sourcesList: [{ kind: 'official', title: 'Lookup API reference', url: 'https://example.org/docs/ubuntu-id/ref', published: '2026-08-28', checked: '2026-09-05', lang: 'en' }],
      limits: { fr: 'Environnement de test uniquement dans cet exemple fictif.', en: 'Sandbox only in this fictional sample.' }, pitfalls: [{ lang: 'en', text: 'Scopes are per country; a Kenyan token is rejected for RW lookups.' }],
      notes: '## Look up a test ID\nSend `GET /v2/ids/{number}` with a country-scoped token.', notesLang: 'en', snippet: { lang: 'py', code: '# Fictional example\nr = requests.get(f"https://api.example.org/ubuntu/v2/ids/{n}",\n                 headers={"Authorization": f"Bearer {tok}"})' }, questions: ['q-kyc-ocr'], projects: [] },
    'r-geojson-cli': { sourcesList: [{ kind: 'claim', title: 'README du dépôt', url: 'https://example.org/docs/geojson-lint/readme', published: '2026-05-10', checked: '2026-05-14', lang: 'en' }],
      limits: { fr: 'Valide la structure, pas l’exactitude des coordonnées.', en: 'Validates structure, not coordinate accuracy.' }, pitfalls: [{ lang: 'en', text: 'Large files are read fully into memory.' }],
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
  D.get = function (id) { return M.byId(D.items, id); };
  D.cat = function (id) { return M.byId(D.categories, id); };
})(window.MOCK);
