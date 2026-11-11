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
    esc: esc,

    /* Fallback test-id assigner: names every text/icon/interactive element that has no explicit data-testid after
       its nearest testid ancestor (`<ancestor>-<icon|tag>-N`). Runs at boot and on DOM changes inside <main>. */
    tid: function (root) {
      var skip = { SCRIPT: 1, STYLE: 1, BR: 1 }, inter = /^(A|BUTTON|INPUT|SELECT|TEXTAREA|LABEL|IMG|PRE|TIME|OPTION|P|LI|H1|H2|H3|H4|DT|DD|CODE|STRONG|EM|LEGEND)$/;
      [].slice.call(root.querySelectorAll('*')).forEach(function (e) {
        if (e.hasAttribute('data-testid') || skip[e.tagName]) return;
        var own = [].some.call(e.childNodes, function (n) { return n.nodeType === 3 && n.textContent.trim(); }), icon = e.classList.contains('icon');
        if (!(own || icon || inter.test(e.tagName))) return;
        var anc = e.parentElement && e.parentElement.closest('[data-testid]'), a = anc ? anc.getAttribute('data-testid') : 'solve';
        var m = icon && /\bi-([a-z-]+)/.exec(e.className), kind = icon ? 'icon-' + (m ? m[1] : 'x') : /^tok-/.test(e.className) ? 'tok' : e.tagName.toLowerCase(), n = 1;
        while (document.querySelector('[data-testid="' + a + '-' + kind + '-' + n + '"]')) n++;
        e.setAttribute('data-testid', a + '-' + kind + '-' + n);
      });
    },
    autoTid: function () {
      var main = document.querySelector('main'), q = false; if (!main) return;
      var run = function () { q = false; M.solve.tid(main); };
      new MutationObserver(function () { if (!q) { q = true; setTimeout(run, 0); } }).observe(main, { childList: true, subtree: true });
      run();
    },

    /* ---- tiny Markdown renderer (escape first; supports headings, lists, quotes, fences, inline code/bold/italic/links) ---- */
    hl: function (code, lang) {
      var kw = /^(func|package|import|return|if|else|for|range|var|const|let|type|struct|defer|nil|true|false|async|await|function|new|select|from|where|create|table|insert|into|values|on|conflict|do|nothing|primary|key|not|null|unique|index|using|alter|add|constraint|references)$/i;
      var com = { go: '//', js: '//', ts: '//', php: '//', sql: '--', python: '#', sh: '#', bash: '#', yaml: '#' }[lang] || '//';
      var re = new RegExp('(' + com.replace(/[/#-]/g, '\\$&') + '[^\\n]*)|("(?:\\\\.|[^"\\\\\\n])*"|\'(?:\\\\.|[^\'\\\\\\n])*\')|([A-Za-z_][A-Za-z0-9_]*)(\\()?', 'g');
      var out = '', last = 0, m;
      while ((m = re.exec(code))) {
        out += esc(code.slice(last, m.index)); last = re.lastIndex;
        if (m[1]) out += '<span class="tok-com">' + esc(m[1]) + '</span>';
        else if (m[2]) out += '<span class="tok-str">' + esc(m[2]) + '</span>';
        else if (kw.test(m[3]) && !m[4]) out += '<span class="tok-kw">' + esc(m[3]) + '</span>';
        else if (m[4]) out += '<span class="tok-fn">' + esc(m[3]) + '</span>' + esc(m[4]);
        else out += esc(m[3]);
      }
      return out + esc(code.slice(last));
    },
    inline: function (s) {
      s = esc(s);
      s = s.replace(/`([^`]+)`/g, '<code>$1</code>');
      s = s.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>').replace(/(^|[^*])\*([^*\s][^*]*)\*/g, '$1<em>$2</em>');
      s = s.replace(/\[([^\]]+)\]\(((?:https?:\/\/|\/|\.\.?\/)[^)\s]+)\)/g, function (m, t, u) {
        return /^https?:/.test(u) ? '<a class="link-external" href="' + u + '" rel="noopener noreferrer">' + t + '</a>' : '<a href="' + u + '">' + t + '</a>';
      });
      return s;
    },
    /* opts: {id: unique prefix for code ids/testids, copyLabel(lang), codeLabel(lang), headStart: first heading level} */
    renderMd: function (md, o) {
      o = o || {}; var lines = String(md || '').replace(/\r/g, '').split('\n'), html = '', i = 0, n = 0, para = [], list = null;
      var inl = M.solve.inline;
      function flushP() { if (para.length) { html += '<p>' + inl(para.join(' ')) + '</p>'; para = []; } }
      function flushL() { if (list) { html += '<' + list.t + '>' + list.items.map(function (x) { return '<li>' + inl(x) + '</li>'; }).join('') + '</' + list.t + '>'; list = null; } }
      while (i < lines.length) {
        var ln = lines[i], f = ln.match(/^(```|~~~)\s*([A-Za-z0-9+#-]*)\s*$/);
        if (f) {
          flushP(); flushL(); var buf = []; i++;
          while (i < lines.length && !/^(```|~~~)\s*$/.test(lines[i])) { buf.push(lines[i]); i++; }
          i++; n++; var lang = (f[2] || 'text').toLowerCase(), cid = (o.id || 'md') + '-code-' + n, name = { go: 'Go', sql: 'SQL', php: 'PHP', ts: 'TypeScript', js: 'JavaScript', python: 'Python', sh: 'Shell', bash: 'Shell' }[lang] || lang;
          html += '<div class="code-block" data-testid="' + cid + '-block"><div class="code-head"><span data-testid="' + cid + '-lang">' + esc(name) + '</span>' +
            '<button type="button" class="btn btn-sm" data-copy="#' + cid + '" aria-label="' + esc(o.copyLabel ? o.copyLabel(name) : 'Copy') + '" data-testid="' + cid + '-copy"><span class="icon i-copy" aria-hidden="true"></span><span>' + esc(o.copyText || 'Copy') + '</span></button></div>' +
            '<pre tabindex="0" aria-label="' + esc(o.codeLabel ? o.codeLabel(name) : name) + '" data-testid="' + cid + '-pre"><code id="' + cid + '">' + M.solve.hl(buf.join('\n'), lang) + '</code></pre></div>';
          continue;
        }
        var h = ln.match(/^(#{1,4})\s+(.*)$/), li = ln.match(/^\s*[-*]\s+(.*)$/), ol = ln.match(/^\s*\d+[.)]\s+(.*)$/), bq = ln.match(/^>\s?(.*)$/);
        if (!ln.trim()) { flushP(); flushL(); }
        else if (h) { flushP(); flushL(); var lv = Math.min(4, (o.headStart || 3) + h[1].length - 1); html += '<h' + lv + '>' + inl(h[2]) + '</h' + lv + '>'; }
        else if (li) { flushP(); if (!list || list.t !== 'ul') { flushL(); list = { t: 'ul', items: [] }; } list.items.push(li[1]); }
        else if (ol) { flushP(); if (!list || list.t !== 'ol') { flushL(); list = { t: 'ol', items: [] }; } list.items.push(ol[1]); }
        else if (bq) { flushP(); flushL(); html += '<blockquote><p>' + inl(bq[1]) + '</p></blockquote>'; }
        else { flushL(); para.push(ln.trim()); }
        i++;
      }
      flushP(); flushL(); return html;
    },
    /* true when a ``` fence is opened but never closed (client-side "snippet valide" check) */
    unbalancedFence: function (md) { var c = (String(md).match(/^(```|~~~)/gm) || []).length; return c % 2 === 1; }
  };
})(window.MOCK);
