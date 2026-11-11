/* Solve lane: FICTIONAL extra data + small shared helpers (Markdown renderer, preview binding).
   Question/answer bodies are USER CONTENT in their original language (never translated by the UI switch).
   Sources point to example.org only. Vote/answer counts come from this fixture, not from real metrics. */
window.MOCK = window.MOCK || {};
(function (M) {
  var esc = function (s) { return String(s).replace(/[&<>"']/g, function (c) { return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]; }); };
  var L = function (a) { return a.join('\n'); };

  /* More list rows so sorting, filtering and pagination are demonstrable (the shared file has 8). */
  var extra = [
    { id: 'q-money-decimal', lang: 'fr', title: 'Stocker des montants en XOF/XAF sans centimes : entier ou NUMERIC ?', excerpt: 'Le franc CFA n’a pas de sous-unité en pratique. Un bigint suffit-il, ou faut-il garder NUMERIC pour les conversions ?',
      tags: ['postgresql', 'payments'], country: 'CI', author: 'p-salif', votes: 7, answers: 2, accepted: true, date: '2026-11-01', env: 'PostgreSQL 16', resource: null, project: null },
    { id: 'q-sms-dlr', lang: 'fr', title: 'Accusés de livraison SMS (DLR) perdus lors des pics de trafic', excerpt: 'Environ un message sur dix reste « envoyé » sans accusé. Faut-il interroger le fournisseur ou fiabiliser le webhook ?',
      tags: ['sms', 'webhook', 'node'], country: 'SN', author: 'p-fatou', votes: 4, answers: 1, accepted: false, date: '2026-10-31', env: 'Node 22', resource: 'r-sahel-sms', project: null },
    { id: 'q-rate-limit', lang: 'en', title: 'Handling 429 responses when batch-sending SMS through a gateway', excerpt: 'We queue 50k messages and the gateway starts returning 429 after a few seconds. What backoff strategy works in practice?',
      tags: ['sms', 'node', 'rate-limit'], country: 'KE', author: 'p-wanjiru', votes: 10, answers: 3, accepted: true, date: '2026-10-30', env: 'Node 22', resource: 'r-sahel-sms', project: null },
    { id: 'q-flutter-lowend', lang: 'fr', title: 'Flutter sur téléphones d’entrée de gamme : réduire le temps de démarrage', excerpt: 'Sur 2 Go de RAM, le premier écran met plus de six secondes. Quelles pistes avant de réécrire en natif ?',
      tags: ['flutter', 'performance'], country: 'SN', author: 'p-fatou', votes: 3, answers: 0, accepted: false, date: '2026-10-29', env: 'Flutter 3.24', resource: null, project: null },
    { id: 'q-kyc-liveness', lang: 'en', title: 'Selfie liveness check fails on older Android browsers', excerpt: 'getUserMedia works but the liveness widget never receives frames on Android 8 WebView. Any known workaround?',
      tags: ['kyc', 'javascript'], country: 'NG', author: 'p-chinedu', votes: 2, answers: 0, accepted: false, date: '2026-10-28', env: 'Android 8 WebView', resource: 'r-baobab-kyc', project: null },
    { id: 'q-pwa-cache', lang: 'fr', title: 'Service worker : mettre en cache des appels API pour des réseaux 2G', excerpt: 'Quelle stratégie (stale-while-revalidate, cache-first) pour garder une liste utilisable hors connexion sans servir des données périmées ?',
      tags: ['javascript', 'offline', 'pwa'], country: 'BF', author: 'p-aminata', votes: 9, answers: 2, accepted: false, date: '2026-10-26', env: 'Chrome 130', resource: null, project: 'pr-offline-drafts' }
  ];

  /* Full threads exist for a few questions; the others fall back to the excerpt (marked as fixture gap in the UI). */
  var T = {};
  T['q-webhook-sig'] = {
    body: L([
      'Je reçois les notifications de paiement d’un fournisseur Mobile Money (fiche fictive **Kalimba Pay**). Chaque requête contient un en-tête `X-Signature` en hexadécimal, mais mon HMAC-SHA256 ne correspond jamais.',
      '', 'Voici ce que je fais actuellement :', '',
      '```go',
      'func verify(w http.ResponseWriter, r *http.Request) {',
      '\tvar payload map[string]any',
      '\tjson.NewDecoder(r.Body).Decode(&payload)',
      '\tbody, _ := json.Marshal(payload) // corps re-sérialisé',
      '\tmac := hmac.New(sha256.New, []byte(secret))',
      '\tmac.Write(body)',
      '\texpected := hex.EncodeToString(mac.Sum(nil))',
      '\tif expected != r.Header.Get("X-Signature") {',
      '\t\thttp.Error(w, "bad signature", http.StatusUnauthorized)',
      '\t}',
      '}',
      '```', '',
      'Ce que j’ai essayé :', '',
      '- comparer en minuscules et en majuscules ;',
      '- tester avec le secret du bac à sable ;',
      '- vérifier l’horloge du serveur.', '',
      'Faut-il signer le **corps brut** ou le corps décodé ? Et comment comparer proprement les deux signatures ?'
    ]),
    answers: [
      { id: 'a-ws-1', author: 'p-wanjiru', lang: 'en', votes: 11, date: '2026-11-09', accepted: true,
        md: L([
          'Sign the **raw request body**, byte for byte. Decoding and re-encoding the JSON changes key order and whitespace, so the digest can never match.', '',
          'Read the body once, verify it, then decode from the same bytes:', '',
          '```go',
          'body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))',
          'if err != nil { http.Error(w, "bad request", 400); return }',
          'mac := hmac.New(sha256.New, []byte(secret))',
          'mac.Write(body)',
          'got, _ := hex.DecodeString(r.Header.Get("X-Signature"))',
          'if !hmac.Equal(mac.Sum(nil), got) { http.Error(w, "invalid signature", 401); return }',
          '// only now: json.Unmarshal(body, &payload)',
          '```', '',
          '`hmac.Equal` compares in constant time, which avoids timing leaks.'
        ]),
        sources: [{ label: 'Kalimba Pay (fictive) : signature des webhooks', url: 'https://example.org/docs/kalimba#webhooks' }, { label: 'Go : package crypto/hmac', url: 'https://example.org/go/crypto-hmac' }] },
      { id: 'a-ws-2', author: 'p-salif', lang: 'fr', votes: 4, date: '2026-11-09', accepted: false,
        md: L([
          'Pensez aussi à l’**horodatage**. Si votre fournisseur envoie un en-tête de date, incluez-le dans le message signé et refusez les requêtes trop anciennes (quelques minutes) pour limiter le rejeu.', '',
          'Gardez le secret hors du dépôt : variable d’environnement, jamais dans le code.'
        ]), sources: [] },
      { id: 'a-ws-3', author: 'p-imane', lang: 'fr', votes: 2, date: '2026-11-10', accepted: false,
        md: L([
          'Vérifiez qu’un proxy (Traefik, nginx) ne modifie pas le corps : compression, normalisation des fins de ligne.', '',
          'Journalisez la **longueur** du corps reçu (pas son contenu) et comparez-la à un exemple du bac à sable : une différence d’octets explique presque toujours une signature invalide.'
        ]), sources: [] }
    ]
  };
  T['q-idempotency'] = {
    body: L([
      'Our payment provider retries a callback up to five times if it does not get a `200` quickly. We sometimes process the same event twice and credit a wallet two times.', '',
      'Current table:', '',
      '```sql',
      'CREATE TABLE payment_events (',
      '  id bigserial PRIMARY KEY,',
      '  provider_ref text NOT NULL,',
      '  status text NOT NULL,',
      '  received_at timestamptz NOT NULL DEFAULT now()',
      ');',
      '```', '',
      'What is a safe way to deduplicate **without** dropping a legitimate status update (for example `pending` then `success` for the same reference)?'
    ]),
    answers: [
      { id: 'a-id-1', author: 'p-chinedu', lang: 'en', votes: 9, date: '2026-11-07', accepted: true,
        md: L([
          'Add a unique constraint on what actually identifies an event, not only the payment reference:', '',
          '```sql',
          'ALTER TABLE payment_events',
          '  ADD CONSTRAINT uq_event UNIQUE (provider_ref, status);', '',
          'INSERT INTO payment_events (provider_ref, status)',
          'VALUES ($1, $2)',
          'ON CONFLICT DO NOTHING;',
          '```', '',
          'Credit the wallet in the **same transaction**, and only when a row was inserted (check the affected row count).'
        ]),
        sources: [{ label: 'PostgreSQL : INSERT ... ON CONFLICT (page fictive)', url: 'https://example.org/postgres/on-conflict' }] },
      { id: 'a-id-2', author: 'p-aminata', lang: 'fr', votes: 5, date: '2026-11-08', accepted: false,
        md: 'Si le fournisseur envoie un **identifiant d’événement**, stockez-le : `UNIQUE (event_id)` est plus fiable que le couple référence/statut.', sources: [] },
      { id: 'a-id-3', author: 'p-yves', lang: 'fr', votes: 3, date: '2026-11-08', accepted: false,
        md: 'Répondez `200` rapidement et traitez le message de façon asynchrone : beaucoup de relances viennent de réponses trop lentes.', sources: [] },
      { id: 'a-id-4', author: 'p-wanjiru', lang: 'en', votes: 2, date: '2026-11-09', accepted: false,
        md: 'Keep the raw payload in a `jsonb` column for audit. It makes replays and debugging much easier.', sources: [] }
    ]
  };
  T['q-ussd-timeout'] = {
    body: L([
      'Notre parcours USSD de paiement ouvre une session côté opérateur. Après environ 30 secondes sans réponse, la session est coupée et l’utilisateur doit recommencer, alors que le paiement est parfois déjà initié.', '',
      '```php',
      '$state = $cache->get(\'ussd:\' . $msisdn);',
      'if ($state === null) { $state = [\'step\' => \'amount\']; } // état perdu à chaque coupure',
      '$cache->set(\'ussd:\' . $msisdn, $state, 30);',
      '```', '',
      'Questions :', '',
      '- où conserver l’état de la transaction ?',
      '- comment reprendre sans doublon si l’utilisateur recompose le code ?'
    ]),
    answers: [
      { id: 'a-us-1', author: 'p-imane', lang: 'fr', votes: 6, date: '2026-11-08', accepted: false,
        md: L([
          'Dissociez l’état de la **transaction** de celui de la **session**. La session opérateur peut mourir, la transaction non :', '',
          '```php',
          '$cache->set(\'ussd:tx:\' . $txId, $state, 600); // TTL indépendant de la session',
          '```', '',
          'Retrouvez ensuite la transaction par numéro et montant à la reconnexion.'
        ]), sources: [{ label: 'Sahel SMS Gateway (fictive) : sessions USSD', url: 'https://example.org/docs/sahel-sms#ussd' }] },
      { id: 'a-us-2', author: 'p-salif', lang: 'fr', votes: 3, date: '2026-11-08', accepted: false,
        md: 'À la reconnexion, détectez une transaction **en attente** et proposez « Reprendre » ou « Annuler » plutôt que de créer un nouveau paiement.', sources: [] }
    ]
  };
  T['q-geocode'] = {
    body: L([
      'Livraison à Douala : les adresses sont décrites par des repères (« en face de la pharmacie, derrière le marché »), sans numéro de rue. Je pense stocker un point et un rayon d’incertitude :', '',
      '```sql',
      'CREATE TABLE delivery_points (',
      '  id bigserial PRIMARY KEY,',
      '  description text NOT NULL,',
      '  landmark_id bigint REFERENCES landmarks(id),',
      '  geom geography(Point, 4326),',
      '  precision_m integer',
      ');',
      '```', '',
      'Est-ce la bonne structure ? Comment gérer les points confirmés par le livreur ?'
    ]),
    answers: [
      { id: 'a-geo-1', author: 'p-chinedu', lang: 'en', votes: 8, date: '2026-11-03', accepted: true,
        md: L([
          'Make landmarks a first-class table and index the geometry:', '',
          '```sql',
          'CREATE INDEX delivery_points_geom_idx',
          '  ON delivery_points USING GIST (geom);', '',
          '-- points within 300 m of a landmark',
          'SELECT id FROM delivery_points',
          'WHERE ST_DWithin(geom, $1::geography, 300);',
          '```'
        ]), sources: [{ label: 'PostGIS : ST_DWithin (page fictive)', url: 'https://example.org/postgis/st-dwithin' }] },
      { id: 'a-geo-2', author: 'p-aminata', lang: 'fr', votes: 4, date: '2026-11-04', accepted: false,
        md: 'Gardez l’historique des corrections des livreurs dans une table séparée et utilisez le point confirmé le plus récent, sans écraser l’original.', sources: [] },
      { id: 'a-geo-3', author: 'p-imane', lang: 'fr', votes: 1, date: '2026-11-04', accepted: false,
        md: 'Pour retrouver un repère écrit de plusieurs façons, l’extension `pg_trgm` aide à chercher par similarité.', sources: [] }
    ]
  };
  T['q-offline-sync'] = {
    body: L([
      'Je construis un éditeur de brouillons utilisé sur des connexions qui coupent plusieurs fois par heure. Je veux une file d’attente locale simple, avec reprise, sans service tiers.', '',
      '```ts',
      'const queue = JSON.parse(localStorage.getItem(\'drafts\') ?? \'[]\');',
      'queue.push({ id: crypto.randomUUID(), body, createdAt: Date.now() });',
      'localStorage.setItem(\'drafts\', JSON.stringify(queue));',
      '```', '',
      'Comment rejouer la file dans l’ordre et éviter les doublons quand la connexion revient à moitié ?'
    ]),
    answers: []
  };

  /* Seeded so the demo user (Aminata, p-aminata) owns an ACCEPTED answer: the question page then shows the "openToCollab" prompt. */
  T['q-money-decimal'] = {
    body: L([
      'Le franc CFA (XOF/XAF) n’a pas de sous-unité en pratique. Je stocke aujourd’hui les montants en `NUMERIC(12,2)` et je me demande si c’est utile.', '',
      '```sql',
      'CREATE TABLE payments (',
      '  id bigserial PRIMARY KEY,',
      '  amount numeric(12,2) NOT NULL,',
      '  currency char(3) NOT NULL',
      ');',
      '```', '',
      'Un `bigint` suffit-il, ou faut-il garder `NUMERIC` pour les conversions entre devises ?'
    ]),
    answers: [
      { id: 'a-md-1', author: 'p-aminata', lang: 'fr', votes: 5, date: '2026-11-02', accepted: true,
        md: L([
          'Stockez l’**unité mineure** dans un `bigint` et gardez la devise à côté. Pour le XOF, l’unité mineure est le franc lui-même : aucun arrondi à gérer.', '',
          '```sql',
          'ALTER TABLE payments',
          '  ALTER COLUMN amount TYPE bigint USING amount::bigint,',
          '  ADD CONSTRAINT amount_positive CHECK (amount > 0);',
          '```', '',
          'Gardez `NUMERIC` seulement pour les taux de change, jamais pour les montants.'
        ]), sources: [{ label: 'PostgreSQL : types numériques (page fictive)', url: 'https://example.org/postgres/numeric-types' }] },
      { id: 'a-md-2', author: 'p-wanjiru', lang: 'en', votes: 2, date: '2026-11-03', accepted: false,
        md: 'Store the ISO 4217 minor-unit exponent in a currency table, so adding a currency with decimals later does not need a migration.', sources: [] }
    ]
  };

  M.solve = {
    extraQuestions: extra,
    all: function () { return M.questions.concat(extra); },
    get: function (id) { return M.byId(M.solve.all(), id); },
    /* thread = {body (markdown), answers[], full}; full=false means the fixture has no answer bodies for this question. */
    thread: function (q) {
      var t = T[q.id];
      if (t) return { body: t.body, answers: t.answers.map(function (a) { return Object.assign({}, a); }), full: true };
      return { body: q.excerpt, answers: [], full: false };
    },

    /* ---- search helpers (mock only; deterministic, no AI) ---- */
    tokens: function (str) {
      var STOP = 'le la les un une des du de et ou en au aux pour par sur dans avec sans est sont que qui quoi comment faut il elle ce cette ces mon ma mes ton son sa ses plus pas ne se ya the a an of to in on for with without is are how do does what when why can i we you it its my this that and or be by from as at not'.split(' ');
      return String(str || '').normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase().split(/[^a-z0-9+#.]+/).filter(function (w) { return w.length > 2 && STOP.indexOf(w) < 0; });
    },
    /* similar(title, tags, excludeId, n): questions whose title/excerpt/tags overlap with the typed title. Tags weigh more. */
    similar: function (title, tags, excludeId, n) {
      var tk = M.solve.tokens(title), tg = (tags || []).map(function (x) { return x.toLowerCase(); });
      if (!tk.length && !tg.length) return [];
      return M.solve.all().filter(function (q) { return q.id !== excludeId; }).map(function (q) {
        var tt = M.solve.tokens(q.title), ex = M.solve.tokens(q.excerpt), sc = 0;
        tk.forEach(function (w) {
          if (tt.some(function (x) { return x === w || (w.length > 4 && x.indexOf(w) === 0) || (x.length > 4 && w.indexOf(x) === 0); })) sc += 3;
          else if (ex.indexOf(w) > -1) sc += 1;
          if (q.tags.indexOf(w) > -1) sc += 3;
        });
        tg.forEach(function (x) { if (q.tags.indexOf(x) > -1) sc += 2; });
        return { q: q, sc: sc };
      }).filter(function (r) { return r.sc >= 3; }).sort(function (a, b) { return b.sc - a.sc || (a.q.date < b.q.date ? 1 : -1); }).slice(0, n || 3).map(function (r) { return r.q; });
    },
    /* unansweredByTag(tags, excludeId, n): unanswered questions sharing at least one tag (newest first); tagMatch = the first shared tag. */
    unansweredByTag: function (tags, excludeId, n) {
      tags = tags || [];
      return M.solve.all().filter(function (q) { return q.id !== excludeId && q.answers === 0; }).map(function (q) {
        var shared = q.tags.filter(function (x) { return tags.indexOf(x) > -1; });
        return { q: q, shared: shared };
      }).filter(function (r) { return r.shared.length; }).sort(function (a, b) { return b.shared.length - a.shared.length || (a.q.date < b.q.date ? 1 : -1); }).slice(0, n || 3)
        .map(function (r) { return Object.assign({}, r.q, { tagMatch: r.shared[0] }); });
    },
    /* resourceFor(tags): the API/tool most linked to questions that share these tags (falls back to an SDK match). */
    resourceFor: function (tags) {
      var count = {}; tags = tags || [];
      M.solve.all().forEach(function (q) { if (!q.resource) return; var k = q.tags.filter(function (x) { return tags.indexOf(x) > -1; }).length; if (k) count[q.resource] = (count[q.resource] || 0) + k; });
      var best = Object.keys(count).sort(function (a, b) { return count[b] - count[a]; })[0];
      if (!best) { var r = (M.resources || []).filter(function (x) { return (x.sdk || []).some(function (s) { return tags.indexOf(s) > -1; }); })[0]; return r || null; }
      return M.byId(M.resources, best);
    },
    esc: esc,

    /* Fallback test-id assigner now lives in app.js (App.TestIds, runs on every page); these stay as no-op aliases. */
    tid: function (root) { window.App.TestIds.assign(root); },
    autoTid: function () { window.App.TestIds.auto(); },

    /* Markdown lives in assets/js/markdown.js (shared); kept as thin aliases for the Solve pages. */
    renderMd: function (md, o) { return window.Markdown.render(md, o); },
    unbalancedFence: function (md) { return window.Markdown.unbalancedFence(md); }
  };
})(window.MOCK);
