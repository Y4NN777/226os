/* Apprendre pillar: FICTIONAL learning content (articles, tutorials, one free training series), shared helpers and the shared dictionary (ln.*).
   Authors are fictional members from people.js. The founder series carries the platform label "Formation gratuite · par l'équipe 226os" (no person named).
   Titles/bodies of articles are USER CONTENT in their original `lang`; series and lesson titles are platform content ({fr,en}), lesson bodies are French.
   Everything here is invented for the prototype; no real API, price, quota or benchmark is claimed. Load after people.js and common.js. */
window.MOCK = window.MOCK || {};
(function (M) {
  'use strict';
  var L = M.learn = {};
  var fr = function (s) { return String(s).replace(/ ([:;?!»])/g, ' $1').replace(/« /g, '« '); };
  var join = function (a) {   /* French typography on prose lines only, never inside code fences */
    var code = false;
    return a.map(function (l) { if (/^```/.test(l)) { code = !code; return l; } return code ? l : fr(l); }).join('\n');
  };

  L.types = { article: { icon: 'file-text' }, tutorial: { icon: 'code' }, course: { icon: 'book' } };
  L.levels = ['beginner', 'intermediate', 'advanced'];
  L.topics = ['ai', 'payments', 'backend', 'data', 'mobile', 'devops'];
  L.suggest = ['webhook', 'mobile-money', 'go', 'postgis', 'ussd', 'rag', 'docker'];

  /* ---------------- The free AI training series ---------------- */
  L.seriesList = [{
    id: 's-ia', type: 'course', team: true, lang: 'fr', topic: 'ai', level: 'beginner', levelTo: 'advanced', date: '2026-10-12',
    title: { fr: 'IA appliquée pour développeurs africains', en: 'Applied AI for African developers' },
    summary: {
      fr: 'Six leçons gratuites pour intégrer un modèle de langage dans un vrai produit, avec une connexion instable, des documents en français et un budget serré.',
      en: 'Six free lessons to put a language model into a real product, with an unstable connection, French documents and a tight budget.'
    },
    outcomes: [
      { fr: 'Appeler une API de modèle de langage sans perdre de requêtes sur une connexion lente', en: 'Call a language-model API without losing requests on a slow connection' },
      { fr: 'Interroger vos documents en français avec un RAG simple', en: 'Query your French documents with a simple RAG' },
      { fr: 'Mesurer la qualité d’un modèle sur vos propres données', en: 'Measure a model’s quality on your own data' },
      { fr: 'Maîtriser coûts, quotas et repli en production', en: 'Keep costs, quotas and fallbacks under control in production' }
    ],
    prereqs: [{ fr: 'Savoir lire du code JavaScript ou Python', en: 'Be able to read JavaScript or Python' }, { fr: 'Connaître les requêtes HTTP et le JSON', en: 'Know HTTP requests and JSON' }],
    tags: ['ia', 'llm', 'rag', 'evaluation'],
    lessons: ['l-ia-1', 'l-ia-2', 'l-ia-3', 'l-ia-4', 'l-ia-5', 'l-ia-6']
  }];

  var lessons = [
    { n: 1, level: 'beginner', minutes: 25, date: '2026-10-12', tags: ['llm', 'http', 'reseau'],
      title: { fr: 'Appeler une API de modèle de langage depuis une connexion lente', en: 'Calling a language-model API from a slow connection' },
      summary: { fr: 'Délais, nouvelles tentatives et flux pour qu’une requête n’échoue pas à la première coupure.', en: 'Timeouts, retries and streaming so a request does not fail at the first drop.' },
      prereq: { fr: 'Aucun : les bases de HTTP suffisent.', en: 'None: HTTP basics are enough.' },
      body: join([
        '## Pourquoi votre requête échoue',
        'Une réponse de modèle peut prendre de 5 à 40 secondes. Sur une connexion mobile instable, un délai trop court coupe la requête et vous la payez quand même côté fournisseur.',
        '## Trois réglages qui changent tout',
        '- Un **délai explicite** par requête, plus long que vos appels habituels.',
        '- Des **nouvelles tentatives espacées** (1 s, 2 s, 4 s) avec un maximum.',
        '- Le **flux** (streaming) : vous affichez le début de la réponse tout de suite.',
        '## Exemple minimal',
        '```js',
        'async function ask(prompt, tries = 3) {',
        '  for (let i = 0; i < tries; i++) {',
        '    const ctrl = new AbortController();',
        '    const timer = setTimeout(() => ctrl.abort(), 45000);',
        '    try {',
        '      const res = await fetch(API_URL, { method: "POST", signal: ctrl.signal,',
        '        headers: { "Content-Type": "application/json" }, body: JSON.stringify({ prompt }) });',
        '      if (res.ok) return await res.json();',
        '    } catch (e) { /* coupure : on réessaie */ }',
        '    finally { clearTimeout(timer); }',
        '    await new Promise(r => setTimeout(r, 1000 * 2 ** i));',
        '  }',
        '  throw new Error("indisponible");',
        '}',
        '```',
        '## À retenir',
        'Réduisez aussi la taille de ce que vous envoyez : un prompt de 300 mots passe mieux qu’un historique complet. Dans la leçon suivante, on cadre les sorties du modèle.'
      ]) },
    { n: 2, level: 'beginner', minutes: 30, date: '2026-10-17', tags: ['llm', 'prompt', 'json'],
      title: { fr: 'Des prompts et des sorties structurées fiables', en: 'Reliable prompts and structured outputs' },
      summary: { fr: 'Demander du JSON, le valider, et réessayer proprement quand le modèle se trompe.', en: 'Ask for JSON, validate it and retry cleanly when the model gets it wrong.' },
      prereq: { fr: 'Leçon 1 terminée.', en: 'Lesson 1 done.' },
      body: join([
        '## Un prompt est un contrat',
        'Un bon prompt décrit le rôle, l’entrée et **le format exact de la sortie**. Si votre code attend du JSON, dites-le, donnez un exemple, et validez le résultat.',
        '## Valider avant d’utiliser',
        'Ne faites jamais confiance à la sortie brute. Vérifiez les champs et les types ; sinon, réessayez une fois avec le message d’erreur.',
        '```js',
        'const schema = ["categorie", "urgence"];',
        'function valid(out) {',
        '  try {',
        '    const o = JSON.parse(out);',
        '    return schema.every(k => k in o) ? o : null;',
        '  } catch (e) { return null; }',
        '}',
        '```',
        '## Cas pratique',
        'Classer des messages de support en « paiement », « compte » ou « autre » avec un niveau d’urgence. Gardez 20 exemples réels : ils serviront à l’évaluation dans la leçon 4.',
        '> Une sortie invalide n’est pas une panne : c’est un cas à prévoir.'
      ]) },
    { n: 3, level: 'intermediate', minutes: 45, date: '2026-10-23', tags: ['rag', 'embeddings', 'francais'],
      title: { fr: 'RAG sur de la documentation en français', en: 'RAG over French documentation' },
      summary: { fr: 'Découper, indexer et interroger vos documents, en gérant accents et abréviations locales.', en: 'Split, index and query your documents, handling accents and local abbreviations.' },
      prereq: { fr: 'Leçons 1 et 2, notions de recherche textuelle.', en: 'Lessons 1 and 2, basic text-search notions.' },
      body: join([
        '## Le principe',
        'Le RAG ajoute au prompt les passages de vos documents les plus proches de la question. Le modèle répond à partir de ce contexte, pas de sa mémoire.',
        '## Découper intelligemment',
        'Découpez par section, pas au hasard : 200 à 400 mots avec un petit recouvrement. Conservez le titre de la section dans chaque morceau.',
        '```python',
        'def chunks(text, size=300, overlap=40):',
        '    words = text.split()',
        '    step = size - overlap',
        '    return [" ".join(words[i:i + size]) for i in range(0, len(words), step)]',
        '```',
        '## Pièges propres au français',
        '- Normalisez les accents pour la recherche : « resiliation » doit trouver « résiliation ».',
        '- Gardez les sigles tels quels (« OHADA », « CFA ») et ajoutez leur forme longue à l’index.',
        '- Citez toujours la source dans la réponse, avec le titre du document.',
        'Un index PostgreSQL avec recherche plein texte suffit pour démarrer ; les vecteurs viennent ensuite.'
      ]) },
    { n: 4, level: 'intermediate', minutes: 40, date: '2026-10-30', tags: ['evaluation', 'qualite', 'donnees'],
      title: { fr: 'Évaluer un modèle sur vos données locales', en: 'Evaluating a model on your local data' },
      summary: { fr: 'Un jeu de test de 30 exemples, une grille simple, et des chiffres que vous pouvez défendre.', en: 'A 30-example test set, a simple grid and numbers you can stand behind.' },
      prereq: { fr: 'Leçons 2 et 3.', en: 'Lessons 2 and 3.' },
      body: join([
        '## Les benchmarks publics ne disent pas tout',
        'Un modèle excellent en anglais peut peiner sur vos noms de quartiers, vos montants en francs CFA ou vos messages mélangeant français et langues locales. Il faut mesurer chez vous.',
        '## Construire un jeu de test',
        'Prenez 30 cas réels, anonymisés, avec la réponse attendue. Séparez les cas faciles, les cas limites et les pièges.',
        '```python',
        'def score(cases, predict):',
        '    ok = sum(1 for c in cases if predict(c["input"]) == c["expected"])',
        '    return ok / len(cases)',
        '```',
        '## Lire les résultats',
        '- Comparez **deux modèles** sur le même jeu, pas un modèle à un chiffre abstrait.',
        '- Notez les erreurs par catégorie : elles guident le prompt ou le choix du modèle.',
        '- Refaites la mesure à chaque changement de prompt.'
      ]) },
    { n: 5, level: 'intermediate', minutes: 30, date: '2026-11-05', tags: ['cout', 'quotas', 'production'],
      title: { fr: 'Coûts, quotas et limites de débit', en: 'Costs, quotas and rate limits' },
      summary: { fr: 'Estimer le coût d’une requête, respecter les limites et gérer proprement les erreurs 429.', en: 'Estimate the cost of a request, respect limits and handle 429 errors cleanly.' },
      prereq: { fr: 'Leçons 1 à 4.', en: 'Lessons 1 to 4.' },
      body: join([
        '## Compter en jetons',
        'Le coût dépend du nombre de jetons en entrée et en sortie. Pour estimer : environ 4 caractères par jeton en anglais, un peu moins en français.',
        '```js',
        'const cost = (inTok, outTok, p) => (inTok * p.in + outTok * p.out) / 1e6;',
        '// p = tarif fictif en dollars par million de jetons',
        '```',
        '## Budgéter',
        '- Fixez un **plafond quotidien** et coupez les fonctions non essentielles au-delà.',
        '- Réduisez le contexte envoyé : résumez l’historique plutôt que de le répéter.',
        '- Choisissez le plus petit modèle qui passe votre jeu de test.',
        '## Gérer les 429',
        'Une erreur 429 signifie « trop de requêtes ». Lisez l’en-tête `Retry-After` quand il existe, attendez, puis réessayez avec le même identifiant de requête pour ne pas facturer deux fois.'
      ]) },
    { n: 6, level: 'advanced', minutes: 40, date: '2026-11-10', tags: ['cache', 'hors-ligne', 'production'],
      title: { fr: 'Cache, repli et hors ligne en production', en: 'Cache, fallback and offline in production' },
      summary: { fr: 'Rendre votre fonction IA utile même quand le réseau ou le fournisseur est indisponible.', en: 'Keep your AI feature useful even when the network or the provider is down.' },
      prereq: { fr: 'Leçons 1 à 5.', en: 'Lessons 1 to 5.' },
      body: join([
        '## Une fonction IA qui ne bloque jamais',
        'Votre produit doit continuer à servir l’utilisateur quand le modèle est lent ou injoignable. Trois mécanismes suffisent.',
        '## 1. Un cache par empreinte',
        'Si la même question revient, ne payez pas deux fois. Hachez le prompt normalisé et stockez la réponse.',
        '```js',
        'const key = await sha256(normalize(prompt));',
        'const hit = cache.get(key);',
        'if (hit) return hit;',
        '```',
        '## 2. Un repli',
        'Un modèle plus petit, une réponse prédéfinie ou une recherche plein texte classique : tout vaut mieux qu’un écran d’erreur.',
        '## 3. Une file hors ligne',
        'Mettez la requête en file locale, affichez « en attente », et envoyez-la au retour du réseau. C’est le même principe que pour des brouillons hors ligne.',
        'Vous avez maintenant tous les éléments pour livrer. Racontez votre expérience dans un article : la communauté en profitera.'
      ]), links: { p: 'pr-offline-drafts' } }
  ];
  L.lessonList = lessons.map(function (x) {
    return Object.assign({ id: 'l-ia-' + x.n, type: 'lesson', series: 's-ia', team: true, lang: 'fr', topic: 'ai', links: {} }, x);
  });

  /* ---------------- Articles and tutorials (fictional members) ---------------- */
  L.itemList = [
    { id: 'a-webhook-momo', type: 'tutorial', author: 'p-aminata', lang: 'fr', topic: 'payments', level: 'intermediate', minutes: 12, date: '2026-11-06',
      title: 'Vérifier un webhook Mobile Money pas à pas',
      summary: 'Calculer et comparer la signature HMAC-SHA256 sur le corps brut, sans se faire piéger par l’encodage ni par les rejeux.',
      tags: ['go', 'webhook', 'mobile-money', 'hmac'], links: { q: 'q-webhook-sig', r: 'r-kalimba', p: 'pr-webhook-kit' },
      body: join([
        '## Ce que vous allez faire',
        'Valider qu’un appel entrant vient bien de votre fournisseur de paiement, en Go, avec une signature HMAC-SHA256. Le piège le plus courant : calculer la signature sur un corps déjà décodé.',
        '## 1. Lire le corps brut',
        'Lisez les octets exacts de la requête **avant** tout décodage JSON. Si vous les décodez puis les ré-encodez, l’ordre des clés et les espaces changent, et la signature ne correspond plus.',
        '```go',
        'body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))',
        'if err != nil {',
        '    http.Error(w, "corps illisible", http.StatusBadRequest)',
        '    return',
        '}',
        '```',
        '## 2. Calculer la signature',
        'Le fournisseur signe le corps avec un secret partagé. Recalculez-la de votre côté.',
        '```go',
        'mac := hmac.New(sha256.New, []byte(secret))',
        'mac.Write(body)',
        'expected := hex.EncodeToString(mac.Sum(nil))',
        '```',
        '## 3. Comparer en temps constant',
        'N’utilisez pas `==` : une comparaison classique révèle des informations par son temps d’exécution.',
        '```go',
        'if !hmac.Equal([]byte(expected), []byte(r.Header.Get("X-Signature"))) {',
        '    http.Error(w, "signature invalide", http.StatusUnauthorized)',
        '    return',
        '}',
        '```',
        '## 4. Refuser les rejeux',
        'Stockez l’identifiant de chaque événement traité et ignorez les doublons. Un webhook peut être renvoyé plusieurs fois par le fournisseur.',
        '## Pour aller plus loin',
        'Une bibliothèque open source reprend ces étapes pour vous : consultez les liens associés ci-dessous.'
      ]) },
    { id: 'a-idempotence', type: 'article', author: 'p-salif', lang: 'fr', topic: 'payments', level: 'intermediate', minutes: 8, date: '2026-11-03',
      title: 'Idempotence des paiements : pourquoi et comment',
      summary: 'Un client qui clique deux fois ne doit jamais payer deux fois. Les clés d’idempotence, expliquées avec un exemple simple.',
      tags: ['paiement', 'idempotence', 'api'], links: { q: 'q-idempotency', r: 'r-kalimba', p: 'pr-webhook-kit' },
      body: join([
        '## Le problème',
        'Sur un réseau instable, une requête de paiement peut partir, aboutir côté serveur et ne jamais revenir au client. Le client réessaie : sans protection, il est débité deux fois.',
        '## La clé d’idempotence',
        'Le client génère un identifiant unique par intention de paiement et l’envoie dans chaque tentative. Le serveur exécute le paiement une seule fois et renvoie la même réponse aux tentatives suivantes.',
        '```php',
        '$cle = $_SERVER["HTTP_IDEMPOTENCY_KEY"] ?? null;',
        'if ($cle && $reponse = $cache->get($cle)) {',
        '    return $reponse; // déjà traité',
        '}',
        '```',
        '## Règles à respecter',
        '- Une clé par action, jamais réutilisée pour un autre montant.',
        '- Conservez les clés au moins 24 heures.',
        '- Si la requête est encore en cours, répondez « en cours » plutôt que de relancer.',
        'Ces règles valent aussi pour vos propres webhooks : traitez chaque événement une seule fois.'
      ]) },
    { id: 'a-ussd-sessions', type: 'tutorial', author: 'p-fatou', lang: 'fr', topic: 'mobile', level: 'beginner', minutes: 15, date: '2026-10-28',
      title: 'Gérer les sessions USSD sans perdre l’état',
      summary: 'Un parcours USSD en trois écrans, avec un état persistant côté serveur et un délai expiré géré proprement.',
      tags: ['ussd', 'php', 'session'], links: { q: 'q-ussd-timeout', r: 'r-sahel-sms', p: 'pr-ussd-flow' },
      body: join([
        '## Pourquoi l’état doit vivre côté serveur',
        'Une session USSD est courte et vous ne contrôlez pas le terminal. À chaque appui, l’opérateur envoie le texte cumulé ; c’est à vous de savoir où en est l’utilisateur.',
        '## Le modèle',
        'Une session = un identifiant, une étape, quelques données. Stockez-la avec une durée de vie légèrement supérieure à celle de l’opérateur.',
        '```php',
        '$etat = $sessions->get($idSession) ?? ["etape" => "accueil"];',
        'switch ($etat["etape"]) {',
        '    case "accueil": return "CON 1. Solde\\n2. Envoyer";',
        '    case "montant": return "CON Montant ?";',
        '}',
        '```',
        '## Gérer l’expiration',
        'Si la session a expiré, ne montrez pas une erreur : repartez de l’accueil et gardez le dernier choix utile.'
      ]) },
    { id: 'a-sms-gsm7', type: 'article', author: 'p-kofi', lang: 'en', topic: 'mobile', level: 'beginner', minutes: 6, date: '2026-10-22',
      title: 'Why your SMS splits in two: GSM-7 versus UCS-2',
      summary: 'One accented character can halve the length of a message. How encodings work and how to keep bills predictable.',
      tags: ['sms', 'encoding', 'unicode'], links: { q: 'q-sms-encoding', r: 'r-sahel-sms' },
      body: [
        '## The surprise',
        'A message of 150 characters fits in one SMS, until one customer name contains a character outside the GSM-7 alphabet. The whole message then switches to UCS-2 and the limit drops from 160 to 70 characters.',
        '## What to do',
        '- Measure length **after** choosing the encoding, not before.',
        '- Replace typographic quotes and dashes with plain ones when the text allows it.',
        '- Warn the sender when a message will cost two segments.',
        '```js',
        'const GSM7 = /^[A-Za-z0-9 @$!"#%&\'()*+,\\-./:;<=>?\\n]*$/;',
        'const limit = GSM7.test(text) ? 160 : 70;',
        '```',
        '## Takeaway',
        'Show the segment count in your UI. Users stop writing novels when they see the number.'
      ].join('\n') },
    { id: 'a-geo-informal', type: 'tutorial', author: 'p-yves', lang: 'fr', topic: 'data', level: 'intermediate', minutes: 18, date: '2026-10-30',
      title: 'Modéliser des adresses informelles avec PostGIS',
      summary: 'Repères, quartiers et points GPS approximatifs : un schéma simple pour livrer là où il n’y a pas de numéro de rue.',
      tags: ['postgresql', 'postgis', 'logistique'], links: { q: 'q-geocode', r: 'r-harmattan', p: 'pr-geo-informal' },
      body: join([
        '## Le constat',
        'Beaucoup d’adresses se décrivent par un repère : « derrière le marché, portail bleu ». Votre base doit accepter cette réalité tout en restant interrogeable.',
        '## Un schéma qui fonctionne',
        'Stockez le texte du repère, un point approximatif et un rayon d’incertitude.',
        '```sql',
        'create table adresse (',
        '  id bigint primary key generated always as identity,',
        '  repere text not null,',
        '  quartier text,',
        '  position geography(point, 4326),',
        '  rayon_m int default 150',
        ');',
        'create index on adresse using gist (position);',
        '```',
        '## Chercher à proximité',
        'Une requête par distance retrouve les adresses voisines pour proposer des doublons probables au livreur.',
        '```sql',
        'select id, repere from adresse',
        'where st_dwithin(position, st_makepoint(-4.02, 5.34)::geography, 300);',
        '```'
      ]) },
    { id: 'a-traefik-502', type: 'tutorial', author: 'p-imane', lang: 'en', topic: 'devops', level: 'intermediate', minutes: 10, date: '2026-10-25',
      title: 'Keeping Traefik routes after a Docker Compose restart',
      summary: 'Why a restarted container disappears from Traefik, and the three labels and one network setting that fix it.',
      tags: ['docker', 'traefik', 'compose'], links: { q: 'q-docker-traefik' },
      body: [
        '## Symptom',
        'After `docker compose restart`, the service answers with a 502 until Traefik itself is restarted. The labels look correct.',
        '## Cause',
        'The container came back on a different network address, or it is not on the network Traefik watches.',
        '## Fix',
        '```yaml',
        'services:',
        '  api:',
        '    networks: [edge]',
        '    labels:',
        '      - traefik.enable=true',
        '      - traefik.docker.network=edge',
        '      - traefik.http.routers.api.rule=Host(`api.example.org`)',
        'networks:',
        '  edge:',
        '    external: true',
        '```',
        '## Check',
        'Open the Traefik dashboard and confirm the service shows one server with a healthy address.'
      ].join('\n') },
    { id: 'a-offline-first', type: 'article', author: 'p-yves', lang: 'fr', topic: 'backend', level: 'intermediate', minutes: 9, date: '2026-10-20',
      title: 'Brouillons hors ligne : concevoir pour les réseaux instables',
      summary: 'File d’attente locale, reprise automatique et conflits : les principes qui évitent de perdre le texte d’un utilisateur.',
      tags: ['javascript', 'offline', 'react'], links: { q: 'q-offline-sync', p: 'pr-offline-drafts' },
      body: join([
        '## Le principe',
        'L’utilisateur ne doit jamais perdre ce qu’il a écrit. Enregistrez localement à chaque pause de frappe, puis synchronisez quand le réseau revient.',
        '## Trois règles',
        '- Chaque brouillon a un **identifiant local** créé avant tout envoi.',
        '- La file rejoue les envois dans l’ordre, avec une pause entre deux échecs.',
        '- En cas de conflit, gardez les deux versions et laissez l’utilisateur choisir.',
        '```ts',
        'type Brouillon = { id: string; texte: string; maj: number; envoye: boolean };',
        'const file: Brouillon[] = JSON.parse(localStorage.getItem("file") ?? "[]");',
        '```',
        '## Ce que l’on affiche',
        'Un état simple : « enregistré sur cet appareil » puis « synchronisé ». Les utilisateurs se fient à ce que vous leur montrez.'
      ]) },
    { id: 'a-kyc-ocr', type: 'article', author: 'p-chinedu', lang: 'en', topic: 'ai', level: 'advanced', minutes: 11, date: '2026-10-15',
      title: 'OCR for identity documents: what to expect and how to fail safely',
      summary: 'Realistic accuracy, the fields that fail first and a review flow that never auto-rejects a real customer.',
      tags: ['kyc', 'ocr', 'python'], links: { q: 'q-kyc-ocr', r: 'r-baobab-kyc' },
      body: [
        '## Set expectations',
        'OCR on photos taken with a phone in poor light will misread some characters. Plan for it instead of hoping it will not happen.',
        '## Fields that fail first',
        '- Names with accents or hyphens.',
        '- Dates printed in a small font.',
        '- Document numbers that mix letters and digits (0 versus O).',
        '## Fail safely',
        '```python',
        'def decide(fields):',
        '    if all(f.confidence > 0.92 for f in fields):',
        '        return "auto"',
        '    return "review"  # a human looks, the customer is never auto-rejected',
        '```',
        '## Measure',
        'Keep a small sample of real, consented documents and track the share sent to review each week.'
      ].join('\n') }
  ];

  /* ---------------- Lookup ---------------- */
  var MKEY = 'learn.mine';
  function mineRaw() { try { var a = JSON.parse(localStorage.getItem('226os.' + MKEY)); return Array.isArray(a) ? a : []; } catch (e) { return []; } }
  L.mine = mineRaw;
  L.addMine = function (it) { var a = mineRaw(); a.unshift(it); try { localStorage.setItem('226os.' + MKEY, JSON.stringify(a.slice(0, 20))); } catch (e) {} return it; };
  L.items = function () { return L.mine().concat(L.itemList); };
  L.listable = function () { return L.seriesList.concat(L.items()).sort(function (a, b) { var c = (b.type === 'course') - (a.type === 'course'); return c || (a.date < b.date ? 1 : a.date > b.date ? -1 : 0); }); };
  L.get = function (id) {
    return L.seriesList.filter(function (x) { return x.id === id; })[0] || L.lessonList.filter(function (x) { return x.id === id; })[0] || L.items().filter(function (x) { return x.id === id; })[0] || null;
  };
  L.lessons = function (sid) { return L.lessonList.filter(function (x) { return x.series === sid; }); };
  L.minutesOf = function (sid) { return L.lessons(sid).reduce(function (a, x) { return a + x.minutes; }, 0); };
  L.txt = function (v) { return v == null ? '' : (typeof v === 'string' ? v : (v[I18N.lang] || v.fr || '')); };
  L.person = function (it) { if (!it) return null; var p = it.author ? M.person(it.author) : null; return p || (it.authorName ? { name: it.authorName } : null); };
  L.href = function (it, base) { return (base || '') + (it.type === 'course' ? 'formation.html?id=' : 'article.html?id=') + it.id; };
  L.icon = function (it) { return L.types[it.type === 'lesson' ? 'course' : it.type].icon; };
  L.fmtMin = function (n) { var h = Math.floor(n / 60), m = n % 60; return h ? (m ? h + ' h ' + (m < 10 ? '0' : '') + m : h + ' h') : n + ' min'; };

  /* ---------------- Progress (simulated, per signed-in user). ?done=l-ia-1,l-ia-2 is a QA override, never persisted. ---------------- */
  function pkey() { var u = App.Session.user; return 'learn.done.' + (u && u.username || 'demo'); }
  L.done = function () {
    if (!App.Session.signedIn) return {};
    var q = new URLSearchParams(location.search).get('done'), o = {};
    if (q) { q.split(',').forEach(function (x) { if (x) o[x] = true; }); return o; }
    try { return JSON.parse(localStorage.getItem('226os.' + pkey())) || {}; } catch (e) { return {}; }
  };
  L.isDone = function (id) { return !!L.done()[id]; };
  L.setDone = function (id, on) {
    var o = L.done(); if (on) o[id] = true; else delete o[id];
    try { localStorage.setItem('226os.' + pkey(), JSON.stringify(o)); } catch (e) {}
    document.dispatchEvent(new CustomEvent('learn:progress', { detail: { id: id, done: !!on } }));
  };
  L.doneCount = function (sid) { var d = L.done(); return L.lessons(sid).filter(function (x) { return d[x.id]; }).length; };
  L.nextLesson = function (sid) { var d = L.done(); return L.lessons(sid).filter(function (x) { return !d[x.id]; })[0] || null; };

  /* ---------------- Search (deterministic: accent- and case-insensitive, every word must match) ---------------- */
  var norm = function (s) { return String(s == null ? '' : s).normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase(); };
  L.norm = norm;
  L.hay = function (it) {
    var p = L.person(it), tp = it.topic ? I18N.t('ln.topic.' + it.topic) : '', ls = it.type === 'course' ? L.lessons(it.id).map(function (x) { return L.txt(x.title); }).join(' ') : '';
    return norm([L.txt(it.title), L.txt(it.summary), (it.tags || []).join(' '), p ? p.name : '', tp, ls, it.type === 'lesson' ? '' : (typeof it.body === 'string' ? it.body : '')].join(' '));
  };
  L.matches = function (it, q) {
    var words = norm(q).split(/\s+/).filter(Boolean); if (!words.length) return true;
    var h = L.hay(it); return words.every(function (w) { return h.indexOf(w) > -1; });
  };

  /* ---------------- Shared row (list and grid). pre = test id prefix, base = path prefix to the pages folder. ---------------- */
  L.row = function (it, pre, base) {
    var A = window.App, esc = A.esc, M_ = window.MOCK, n = pre + '-' + it.id, title = L.txt(it.title), sum = L.txt(it.summary), course = it.type === 'course', p = L.person(it), t = I18N.t;
    var typeBadge = '<span class="badge ln-type ln-type-' + it.type + '" data-testid="' + n + '-type"><span class="icon i-' + L.icon(it) + '" aria-hidden="true" data-testid="' + n + '-type-icon"></span><span data-testid="' + n + '-type-label">' + esc(t('ln.type.' + it.type)) + '</span></span>';
    var lvl = course ? t('ln.level.range', { a: t('ln.level.' + it.level).toLowerCase(), b: t('ln.level.' + it.levelTo).toLowerCase() }) : t('ln.level.' + it.level);
    var mins = course ? L.minutesOf(it.id) : it.minutes;
    var by = course ? '<span class="meta-item ln-team" data-testid="' + n + '-author">' + esc(t('ln.team')) + '</span>'
      : '<span class="meta-item" data-testid="' + n + '-author">' + (p ? esc(p.name) : '') + '</span>';
    var done = course ? L.doneCount(it.id) : 0, total = course ? L.lessons(it.id).length : 0;
    var aside = course
      ? (done ? A.statePill(done === total ? 'accepted' : 'open', { testid: n + '-progress', label: t('ln.progress', { done: done, total: total }) }) : '<span class="tabular-nums" data-testid="' + n + '-count">' + esc(t('ln.lessons', { n: total })) + '</span>')
      : '<time class="tabular-nums" datetime="' + it.date + '" data-testid="' + n + '-date">' + esc(M_.fmtDate(it.date)) + '</time>';
    return '<li class="row ln-row" data-testid="' + n + '-row"><div class="row-lead">' + A.thumb(title, { icon: L.icon(it), tone: course ? 0 : undefined, testid: n + '-thumb' }) + '</div>' +
      '<div class="row-main"><div class="ln-row-top">' + typeBadge + '</div><h3 class="row-title" data-testid="' + n + '-title"><a href="' + L.href(it, base) + '"' + (course ? '' : ' lang="' + it.lang + '" data-user-content') + ' data-testid="' + n + '-link">' + esc(title) + '</a></h3>' +
      '<p class="row-excerpt"' + (course ? '' : ' lang="' + it.lang + '" data-user-content') + ' data-testid="' + n + '-summary">' + esc(sum) + '</p>' +
      '<div class="meta">' + by + '<span class="meta-item" data-testid="' + n + '-level">' + esc(lvl) + '</span>' +
      '<span class="meta-item" data-testid="' + n + '-minutes"><span class="icon i-clock" aria-hidden="true" data-testid="' + n + '-minutes-icon"></span>' + esc(L.fmtMin(mins)) + '</span>' +
      '<span class="badge" data-testid="' + n + '-lang"><span aria-hidden="true" data-testid="' + n + '-lang-code">' + it.lang.toUpperCase() + '</span><span class="visually-hidden" data-testid="' + n + '-lang-name">' + esc(t('ln.lang.' + it.lang)) + '</span></span></div></div>' +
      '<div class="row-aside">' + aside + '</div></li>';
  };

  /* ---------------- Shared dictionary ---------------- */
  I18N.register({
    fr: {
      'ln.type.article': 'Article', 'ln.type.tutorial': 'Tutoriel', 'ln.type.course': 'Formation', 'ln.type.lesson': 'Leçon',
      'ln.level.beginner': 'Débutant', 'ln.level.intermediate': 'Intermédiaire', 'ln.level.advanced': 'Avancé', 'ln.level.range': '{a} à {b}',
      'ln.topic.ai': 'IA', 'ln.topic.payments': 'Paiement', 'ln.topic.backend': 'Backend', 'ln.topic.data': 'Données', 'ln.topic.mobile': 'Mobile et SMS', 'ln.topic.devops': 'DevOps',
      'ln.lang.fr': 'Français', 'ln.lang.en': 'Anglais',
      'ln.team': 'Formation gratuite · par l’équipe 226os', 'ln.free': 'Gratuit',
      'ln.lessons': '{n} leçons', 'ln.minutes': '{n} min', 'ln.read': '{n} min de lecture',
      'ln.progress': '{done} sur {total} leçons',
      'ln.by': 'Par {name}', 'ln.done': 'Terminée',
      'ln.start': 'Commencer la formation', 'ln.continue': 'Continuer : leçon {n}', 'ln.review': 'Revoir la formation', 'ln.open': 'Voir la formation'
    },
    en: {
      'ln.type.article': 'Article', 'ln.type.tutorial': 'Tutorial', 'ln.type.course': 'Training', 'ln.type.lesson': 'Lesson',
      'ln.level.beginner': 'Beginner', 'ln.level.intermediate': 'Intermediate', 'ln.level.advanced': 'Advanced', 'ln.level.range': '{a} to {b}',
      'ln.topic.ai': 'AI', 'ln.topic.payments': 'Payments', 'ln.topic.backend': 'Backend', 'ln.topic.data': 'Data', 'ln.topic.mobile': 'Mobile and SMS', 'ln.topic.devops': 'DevOps',
      'ln.lang.fr': 'French', 'ln.lang.en': 'English',
      'ln.team': 'Free training · by the 226os team', 'ln.free': 'Free',
      'ln.lessons': '{n} lessons', 'ln.minutes': '{n} min', 'ln.read': '{n} min read',
      'ln.progress': '{done} of {total} lessons',
      'ln.by': 'By {name}', 'ln.done': 'Done',
      'ln.start': 'Start the training', 'ln.continue': 'Continue: lesson {n}', 'ln.review': 'Review the training', 'ln.open': 'View the training'
    }
  });
})(window.MOCK);
