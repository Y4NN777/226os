/* Accueil (signed-in home), v2. A deterministic "Pour vous" feed built from the declared skills, country and intent of the simulated session.
   No AI, no inbox, no profile checklist here (the checklist lives on the own profile page only). Mock data only. Namespace: feed.* */
I18N.register({
  fr: {
    'feed.title': 'Accueil — 226os',
    'feed.hello': 'Bonjour, {name}',
    'feed.topics.label': 'Vos sujets', 'feed.topics.edit': 'Modifier', 'feed.topics.none': 'Aucun sujet choisi',
    'feed.lead.ask.title': 'Posez votre question à la communauté',
    'feed.lead.ask.body': 'Décrivez votre blocage : quelqu’un qui connaît votre stack vous répondra.',
    'feed.lead.ask.cta': 'Poser une question',
    'feed.lead.help.title.one': '1 question attend une réponse',
    'feed.lead.help.title.many': '{n} questions attendent une réponse',
    'feed.lead.help.body': 'Choisies d’après vos sujets, les moins servies d’abord.',
    'feed.lead.help.cta': 'Voir les questions',
    'feed.lead.integrate.title.one': '1 API vérifiée couvre votre pays',
    'feed.lead.integrate.title.many': '{n} API vérifiées couvrent votre pays',
    'feed.lead.integrate.title.any': 'Intégrez une API sourcée',
    'feed.lead.integrate.body': 'Sources datées, pièges connus et exemples de code.',
    'feed.lead.integrate.cta': 'Ouvrir le catalogue',
    'feed.lead.contribute.title.one': '1 projet cherche vos compétences',
    'feed.lead.contribute.title.many': '{n} projets cherchent vos compétences',
    'feed.lead.contribute.body': 'Des besoins précis, avec le niveau attendu.',
    'feed.lead.contribute.cta': 'Voir les projets',
    'feed.q.title': 'Questions à résoudre',
    'feed.q.sub': 'Dans vos sujets, sans réponse acceptée.',
    'feed.q.fallback': 'Rien de non résolu dans vos sujets pour l’instant. Voici d’autres questions ouvertes.',
    'feed.q.more': 'Toutes les questions',
    'feed.q.answers': '{n} réponse(s)', 'feed.q.noanswer': 'Sans réponse',
    'feed.q.empty.title': 'Toutes les questions ont une réponse acceptée', 'feed.q.empty.body': 'Revenez plus tard, ou posez la vôtre.', 'feed.q.empty.cta': 'Poser une question',
    'feed.p.title': 'Projets qui ont besoin de vous',
    'feed.p.sub': 'Besoins ouverts qui correspondent à vos compétences.',
    'feed.p.fallback': 'Aucun besoin ne correspond exactement. Voici des projets ouverts.',
    'feed.p.more': 'Tous les projets', 'feed.p.needs': '{n} besoin(s) ouvert(s)',
    'feed.p.empty.title': 'Aucun projet ouvert pour le moment', 'feed.p.empty.body': 'Publiez le vôtre pour trouver des contributeurs.', 'feed.p.empty.cta': 'Publier un projet',
    'feed.r.title': 'API récemment vérifiées',
    'feed.r.sub': 'Qui couvrent {country}.',
    'feed.r.subAny': 'Sources vérifiées, les plus récentes d’abord.',
    'feed.r.more': 'Parcourir le catalogue', 'feed.r.on': 'le {date}',
    'feed.r.empty.title': 'Aucune API vérifiée', 'feed.r.empty.body': 'Le catalogue s’enrichit au fil des sources.', 'feed.r.empty.cta': 'Parcourir le catalogue',
    'feed.a.title': 'Votre activité',
    'feed.a.q': 'Vos questions', 'feed.a.apps': 'Vos candidatures',
    'feed.a.q.empty.title': 'Pas encore de question', 'feed.a.q.empty.body': 'Décrivez un blocage, la communauté vous répond.', 'feed.a.q.empty.cta': 'Poser une question',
    'feed.a.apps.empty.title': 'Aucune candidature', 'feed.a.apps.empty.body': 'Ces projets cherchent vos compétences.', 'feed.a.apps.empty.cta': 'Trouver un projet',
    'feed.app.pending': 'En attente', 'feed.app.accepted': 'Acceptée', 'feed.app.declined': 'Refusée',
    'feed.first.title': 'Choisissez 3 sujets pour commencer',
    'feed.first.body': 'Votre accueil se compose d’après vos sujets : questions à résoudre, projets et API pour vous. Vous pourrez les modifier à tout moment.',
    'feed.first.group': 'Sujets proposés',
    'feed.first.count': '{n}/3 choisis',
    'feed.first.cta': 'Composer mon accueil',
    'feed.first.hint': 'Choisissez au moins un sujet.',
    'feed.l.title': 'Apprendre', 'feed.l.more': 'Tout apprendre',
    'feed.l.sub.cont': 'Reprenez là où vous vous êtes arrêté.', 'feed.l.sub.rec': 'Formation gratuite et articles choisis d’après vos sujets.',
    'feed.l.empty.title': 'Rien à lire pour l’instant', 'feed.l.empty.body': 'Les contenus arrivent. Écrivez le premier !', 'feed.l.empty.cta': 'Écrire un article',
    'feed.l.share': 'Partagez ce que vous savez'
  },
  en: {
    'feed.title': 'Home — 226os',
    'feed.hello': 'Hello, {name}',
    'feed.topics.label': 'Your topics', 'feed.topics.edit': 'Edit', 'feed.topics.none': 'No topic chosen',
    'feed.lead.ask.title': 'Ask the community',
    'feed.lead.ask.body': 'Describe where you are stuck: someone who knows your stack will answer.',
    'feed.lead.ask.cta': 'Ask a question',
    'feed.lead.help.title.one': '1 question is waiting for an answer',
    'feed.lead.help.title.many': '{n} questions are waiting for an answer',
    'feed.lead.help.body': 'Picked from your topics, least served first.',
    'feed.lead.help.cta': 'See the questions',
    'feed.lead.integrate.title.one': '1 verified API covers your country',
    'feed.lead.integrate.title.many': '{n} verified APIs cover your country',
    'feed.lead.integrate.title.any': 'Integrate a sourced API',
    'feed.lead.integrate.body': 'Dated sources, known pitfalls and code samples.',
    'feed.lead.integrate.cta': 'Open the catalogue',
    'feed.lead.contribute.title.one': '1 project is looking for your skills',
    'feed.lead.contribute.title.many': '{n} projects are looking for your skills',
    'feed.lead.contribute.body': 'Specific needs, with the expected level.',
    'feed.lead.contribute.cta': 'See the projects',
    'feed.q.title': 'Questions to solve',
    'feed.q.sub': 'In your topics, without an accepted answer.',
    'feed.q.fallback': 'Nothing unresolved in your topics right now. Here are other open questions.',
    'feed.q.more': 'All questions',
    'feed.q.answers': '{n} answer(s)', 'feed.q.noanswer': 'No answer',
    'feed.q.empty.title': 'Every question has an accepted answer', 'feed.q.empty.body': 'Come back later, or ask your own.', 'feed.q.empty.cta': 'Ask a question',
    'feed.p.title': 'Projects that need you',
    'feed.p.sub': 'Open needs that match your skills.',
    'feed.p.fallback': 'No need matches exactly. Here are open projects.',
    'feed.p.more': 'All projects', 'feed.p.needs': '{n} open need(s)',
    'feed.p.empty.title': 'No open project right now', 'feed.p.empty.body': 'Publish yours to find contributors.', 'feed.p.empty.cta': 'Publish a project',
    'feed.r.title': 'Recently verified APIs',
    'feed.r.sub': 'Covering {country}.',
    'feed.r.subAny': 'Verified sources, most recent first.',
    'feed.r.more': 'Browse the catalogue', 'feed.r.on': 'on {date}',
    'feed.r.empty.title': 'No verified API', 'feed.r.empty.body': 'The catalogue grows as sources are checked.', 'feed.r.empty.cta': 'Browse the catalogue',
    'feed.a.title': 'Your activity',
    'feed.a.q': 'Your questions', 'feed.a.apps': 'Your applications',
    'feed.a.q.empty.title': 'No question yet', 'feed.a.q.empty.body': 'Describe where you are stuck; the community answers.', 'feed.a.q.empty.cta': 'Ask a question',
    'feed.a.apps.empty.title': 'No application', 'feed.a.apps.empty.body': 'These projects look for your skills.', 'feed.a.apps.empty.cta': 'Find a project',
    'feed.app.pending': 'Pending', 'feed.app.accepted': 'Accepted', 'feed.app.declined': 'Declined',
    'feed.first.title': 'Pick 3 topics to get started',
    'feed.first.body': 'Your home is built from your topics: questions to solve, projects and APIs for you. You can change them at any time.',
    'feed.first.group': 'Suggested topics',
    'feed.first.count': '{n}/3 picked',
    'feed.first.cta': 'Build my home',
    'feed.first.hint': 'Pick at least one topic.',
    'feed.l.title': 'Learn', 'feed.l.more': 'Everything to learn',
    'feed.l.sub.cont': 'Pick up where you left off.', 'feed.l.sub.rec': 'A free training and articles picked from your topics.',
    'feed.l.empty.title': 'Nothing to read yet', 'feed.l.empty.body': 'Content is on its way. Write the first piece!', 'feed.l.empty.cta': 'Write an article',
    'feed.l.share': 'Share what you know'
  }
});

(function () {
  'use strict';
  var root = document.getElementById('home-root'); if (!root) return;      /* the old landing page also loaded this file; do nothing there */
  var A = window.App, M = window.MOCK, S = A.Session, t = function (k, p) { return I18N.t(k, p); }, esc = A.esc, CFG = A.config;
  if (!S.requireAuth()) return;
  var B = M.build, SUGGEST = ['go', 'postgresql', 'javascript', 'react', 'php', 'python', 'webhook', 'mobile-money', 'ussd', 'flutter', 'docker', 'kyc'];
  var ALIAS = { postgresql: ['postgres', 'postgis'], postgres: ['postgresql', 'postgis'], javascript: ['node', 'typescript'], typescript: ['node', 'javascript'], node: ['javascript', 'typescript'], 'mobile-money': ['payments', 'payment'], docker: ['devops'], webhook: ['webhooks'] };
  var picked = [], mounted = false;
  var LANGNAME = { fr: 'FR', en: 'EN' };

  function topicSet(u) { var s = {}; (u.skills || []).forEach(function (k) { k = String(k).toLowerCase(); s[k] = 1; (ALIAS[k] || []).forEach(function (a) { s[a] = 1; }); }); return s; }
  function any(arr, set) { return (arr || []).some(function (x) { return set[String(x).toLowerCase()]; }); }
  function cc(code, id) { return '<span class="country" data-testid="' + id + '"><span class="country-code" data-testid="' + id + '-code">' + code + '</span><span class="visually-hidden" data-testid="' + id + '-name">' + esc(M.countryName(code)) + '</span></span>'; }
  function tags(a, id, max) { return '<ul class="tag-list list-reset" data-testid="' + id + '">' + a.slice(0, max || 3).map(function (x) { return '<li><span class="tag" data-testid="' + id + '-' + String(x).replace(/[^a-z0-9]+/gi, '-').toLowerCase() + '">' + esc(x) + '</span></li>'; }).join('') + '</ul>'; }
  function first(name) { return String(name || '').split(/\s+/)[0] || ''; }

  /* ---------- data (deterministic) ---------- */
  function questions(u, set) {
    var all = (M.solve ? M.solve.all() : M.questions).filter(function (q) { return !q.accepted; }), self = M.connect.selfPerson(u);
    all = all.filter(function (q) { return !self || q.author !== self.id; });
    var mine = all.filter(function (q) { return any(q.tags, set); }), fb = false, list = mine;
    if (!mine.length) { list = all; fb = true; }
    list = list.slice().sort(function (a, b) { return a.answers - b.answers || (a.date < b.date ? 1 : -1); });
    return { items: list, fallback: fb };
  }
  function projects(u, set) {
    var all = B.projects().filter(function (p) { return p.openCount > 0; }), self = M.connect.selfPerson(u);
    all = all.filter(function (p) { return !self || p.owner !== self.id; });
    var mine = all.filter(function (p) { return any(p.stack, set) || any(p.skills, set); }), fb = !mine.length;
    return { items: (fb ? all : mine).slice().sort(function (a, b) { return a.updated < b.updated ? 1 : -1; }), fallback: fb };
  }
  function resources(u) {
    var all = M.resources.filter(function (r) { return r.verified; }), cov = u.country ? all.filter(function (r) { return r.countries.indexOf(u.country) > -1; }) : [];
    var list = (cov.length ? cov : all).slice().sort(function (a, b) { return a.verifiedAt < b.verifiedAt ? 1 : -1; });
    return { items: list, covered: !!cov.length };
  }
  function myQuestions(u) { var self = M.connect.selfPerson(u); return self ? (M.solve ? M.solve.all() : M.questions).filter(function (q) { return q.author === self.id; }).sort(function (a, b) { return a.date < b.date ? 1 : -1; }) : []; }
  function myApps(u) {
    var self = M.connect.selfPerson(u), out = []; if (!self) return out;
    B.projects().forEach(function (p) { p.applications.forEach(function (a) { if (a.person === self.id) { var n = p.needs.filter(function (x) { return x.id === a.need; })[0]; out.push({ app: a, project: p, need: n }); } }); });
    return out.sort(function (a, b) { return a.app.date < b.app.date ? 1 : -1; });
  }

  /* ---------- rows ---------- */
  function rowQ(q) {
    var p = M.person(q.author), n = q.id;
    return '<li class="row" data-testid="feed-q-row-' + n + '"><div class="row-lead">' + A.avatar(p || '?', { testid: 'feed-q-avatar-' + n }) + '</div>' +
      '<div class="row-main"><h3 class="row-title"><a href="question.html?id=' + n + '" lang="' + q.lang + '" data-user-content data-testid="feed-q-link-' + n + '">' + esc(q.title) + '</a></h3>' +
      '<div class="meta">' + tags(q.tags, 'feed-q-tags-' + n, 3) + '<span class="meta-item">' + cc(q.country, 'feed-q-country-' + n) + '</span><time class="meta-item" datetime="' + q.date + '" data-testid="feed-q-date-' + n + '">' + M.fmtDate(q.date) + '</time></div></div>' +
      '<div class="row-aside"><span class="id-state-text tabular-nums" data-testid="feed-q-answers-' + n + '"><span class="icon i-message" aria-hidden="true" data-testid="feed-q-answers-icon-' + n + '"></span>' + esc(q.answers ? t('feed.q.answers', { n: q.answers }) : t('feed.q.noanswer')) + '</span></div></li>';
  }
  function rowP(p) {
    var n = p.id, d = (p.needs.filter(function (x) { return x.status === 'open'; })[0] || {}).title;
    return '<li class="row" data-testid="feed-p-row-' + n + '"><div class="row-lead">' + A.thumb(p.name, { icon: 'hammer', testid: 'feed-p-thumb-' + n }) + '</div>' +
      '<div class="row-main"><h3 class="row-title"><a href="project.html?id=' + n + '" class="id-mono" data-testid="feed-p-link-' + n + '">' + esc(p.name) + '</a></h3>' +
      (d ? '<p class="row-excerpt" lang="' + p.lang + '" data-user-content data-testid="feed-p-need-' + n + '">' + esc(d) + '</p>' : '') +
      '<div class="meta">' + tags(p.stack, 'feed-p-tags-' + n, 3) + '<span class="meta-item">' + cc(p.country, 'feed-p-country-' + n) + '</span></div></div>' +
      '<div class="row-aside">' + A.statePill('open', { testid: 'feed-p-pill-' + n }) + '<span class="tabular-nums" data-testid="feed-p-needs-' + n + '">' + esc(t('feed.p.needs', { n: p.openCount })) + '</span></div></li>';
  }
  function rowR(r) {
    var n = r.id, cat = M.byId(M.categories, r.category), stale = false;
    return '<li class="row" data-testid="feed-r-row-' + n + '"><div class="row-lead">' + A.thumb(r.name, { icon: 'compass', testid: 'feed-r-thumb-' + n }) + '</div>' +
      '<div class="row-main"><h3 class="row-title"><a href="resource.html?id=' + n + '" data-testid="feed-r-link-' + n + '">' + esc(r.name) + '</a></h3>' +
      '<div class="meta"><span class="meta-item" data-testid="feed-r-category-' + n + '">' + esc(M.label(cat)) + '</span>' + r.countries.slice(0, 4).map(function (c, i) { return '<span class="meta-item">' + cc(c, 'feed-r-country-' + n + '-' + (i + 1)) + '</span>'; }).join('') + '</div></div>' +
      '<div class="row-aside">' + A.statePill(stale ? 'stale' : 'verified', { testid: 'feed-r-pill-' + n }) + '<time class="tabular-nums" datetime="' + r.verifiedAt + '" data-testid="feed-r-date-' + n + '">' + esc(t('feed.r.on', { date: M.fmtDate(r.verifiedAt) })) + '</time></div></li>';
  }
  function rowApp(x) {
    var n = x.app.id, st = x.app.status;
    return '<li class="row" data-testid="feed-app-row-' + n + '"><div class="row-lead">' + A.thumb(x.project.name, { icon: 'hammer', size: 'sm', testid: 'feed-app-thumb-' + n }) + '</div>' +
      '<div class="row-main"><h3 class="row-title"><a href="project.html?id=' + x.project.id + '" class="id-mono" data-testid="feed-app-link-' + n + '">' + esc(x.project.name) + '</a></h3>' +
      '<p class="row-excerpt" lang="' + x.project.lang + '" data-user-content data-testid="feed-app-need-' + n + '">' + esc(x.need ? x.need.title : '') + '</p></div>' +
      '<div class="row-aside">' + A.statePill(st === 'accepted' ? 'accepted' : st === 'declined' ? 'closed' : 'draft', { testid: 'feed-app-pill-' + n, labelKey: 'feed.app.' + st }) + '</div></li>';
  }
  function rowMyQ(q) {
    var n = q.id;
    return '<li class="row" data-testid="feed-my-row-' + n + '"><div class="row-main"><h3 class="row-title"><a href="question.html?id=' + n + '" lang="' + q.lang + '" data-user-content data-testid="feed-my-link-' + n + '">' + esc(q.title) + '</a></h3>' +
      '<div class="meta"><time class="meta-item" datetime="' + q.date + '" data-testid="feed-my-date-' + n + '">' + M.fmtDate(q.date) + '</time><span class="meta-item" data-testid="feed-my-answers-' + n + '">' + esc(q.answers ? t('feed.q.answers', { n: q.answers }) : t('feed.q.noanswer')) + '</span></div></div>' +
      '<div class="row-aside">' + (q.accepted ? A.statePill('accepted', { testid: 'feed-my-pill-' + n }) : '') + '</div></li>';
  }
  function learnItems(set) {
    var Ln = M.learn; if (!Ln) return [];
    var arts = Ln.items().filter(function (x) { return any(x.tags, set); }), rest = Ln.items().filter(function (x) { return arts.indexOf(x) < 0; });
    return arts.concat(rest).slice(0, 2);
  }
  function learnBox(b, items) {
    var Ln = M.learn, s = Ln.seriesList[0], ls = Ln.lessons(s.id), done = Ln.doneCount(s.id), next = Ln.nextLesson(s.id), n = ls.length;
    var cta = done && next ? t('ln.continue', { n: next.n }) : done ? t('ln.review') : t('ln.start'), go = done && next ? next : ls[0];
    var prog = done ? '<div class="ln-progress" data-testid="feed-l-progress"><div class="ln-progress-text"><strong data-testid="feed-l-progress-text">' + esc(t('ln.progress', { done: done, total: n })) + '</strong><span class="tabular-nums" data-testid="feed-l-progress-pct">' + Math.round(100 * done / n) + '\u00a0%</span></div><div class="ln-bar" role="progressbar" aria-valuemin="0" aria-valuemax="' + n + '" aria-valuenow="' + done + '" aria-label="' + esc(t('ln.progress', { done: done, total: n })) + '" data-testid="feed-l-bar"><span style="width:' + Math.round(100 * done / n) + '%"></span></div></div>' : '';
    var card = '<div class="ln-mini" data-testid="feed-l-course"><div class="ln-mini-top"><span class="ln-icon-badge" aria-hidden="true" data-testid="feed-l-course-badge"><span class="icon i-sparkles" data-testid="feed-l-course-icon"></span></span><span class="hint" data-testid="feed-l-course-team">' + esc(t('ln.team')) + '</span></div>' +
      '<h3 data-testid="feed-l-course-title">' + esc(Ln.txt(s.title)) + '</h3><p data-testid="feed-l-course-meta">' + esc(t('ln.lessons', { n: n }) + ' · ' + Ln.fmtMin(Ln.minutesOf(s.id))) + '</p>' + prog +
      '<div class="ln-actions"><a class="btn btn-primary btn-sm" href="article.html?id=' + go.id + '" data-testid="feed-l-course-cta"><span data-testid="feed-l-course-cta-label">' + esc(cta) + '</span><span class="icon i-arrow-right" aria-hidden="true" data-testid="feed-l-course-cta-icon"></span></a></div></div>';
    b.innerHTML = card + (items.length ? '<ul class="list list-reset" style="margin-top:var(--space-3)" data-testid="feed-l-list">' + items.map(function (x) { return Ln.row(x, 'feed-l', ''); }).join('') + '</ul>' : '');
  }
  function list(items, fn, id) { return '<ul class="list list-reset" data-testid="' + id + '">' + items.map(fn).join('') + '</ul>'; }
  function emptyBox(kind, icon, id, href) {
    return '<div class="id-empty" data-testid="' + id + '"><span class="icon i-' + icon + '" aria-hidden="true" data-testid="' + id + '-icon"></span><h3 data-i18n="feed.' + kind + '.empty.title" data-testid="' + id + '-title"></h3><p data-i18n="feed.' + kind + '.empty.body" data-testid="' + id + '-body"></p>' +
      '<a class="btn btn-sm" href="' + href + '" data-testid="' + id + '-cta"><span data-i18n="feed.' + kind + '.empty.cta" data-testid="' + id + '-cta-label"></span></a></div>';
  }
  function block(id, icon, titleKey, moreKey, moreHref, subHtml) {
    return '<section class="id-block" aria-labelledby="feed-' + id + '-h" data-testid="feed-' + id + '"><div class="id-head"><span class="icon i-' + icon + '" aria-hidden="true" data-testid="feed-' + id + '-icon"></span><h2 id="feed-' + id + '-h" data-i18n="' + titleKey + '" data-testid="feed-' + id + '-title"></h2>' +
      (moreKey ? '<a href="' + moreHref + '" data-i18n="' + moreKey + '" data-testid="feed-' + id + '-more"></a>' : '') + '</div>' + (subHtml ? '<p class="id-sub" id="feed-' + id + '-sub" data-testid="feed-' + id + '-sub">' + subHtml + '</p>' : '') + '<div id="feed-' + id + '-box" data-testid="feed-' + id + '-box"></div></section>';
  }

  /* ---------- first-use (no skills) ---------- */
  function renderFirst() {
    var u = S.user;
    root.innerHTML = '<header class="home-hero" data-testid="feed-hero"><h1 data-testid="feed-h1">' + esc(t('feed.hello', { name: first(u.name) })) + '</h1></header>' +
      '<section class="home-first" aria-labelledby="feed-first-h" data-testid="feed-first"><div class="stack-sm"><h2 id="feed-first-h" data-i18n="feed.first.title" data-testid="feed-first-title"></h2><p data-i18n="feed.first.body" data-testid="feed-first-body"></p></div>' +
      '<div class="home-pick" role="group" data-i18n-attr="aria-label:feed.first.group" data-testid="feed-first-topics">' + SUGGEST.map(function (s) { return '<button type="button" class="chip" aria-pressed="' + (picked.indexOf(s) > -1) + '" data-pick="' + s + '" data-testid="feed-first-chip-' + s + '">' + esc(s) + '</button>'; }).join('') + '</div>' +
      '<div class="home-first-foot"><button type="button" class="btn btn-primary" id="feed-first-save"' + (picked.length ? '' : ' disabled') + ' data-testid="feed-first-save"><span data-i18n="feed.first.cta" data-testid="feed-first-save-label"></span><span class="icon i-arrow-right" aria-hidden="true" data-testid="feed-first-save-icon"></span></button>' +
      '<span class="home-first-count" role="status" data-testid="feed-first-count">' + esc(picked.length ? t('feed.first.count', { n: picked.length }) : t('feed.first.hint')) + '</span></div></section>';
    I18N.apply(root);
  }

  /* ---------- feed ---------- */
  function lead(u, d) {
    var it = u.intent || 'help', ic = { ask: 'help', help: 'message', integrate: 'compass', contribute: 'hammer' }[it] || 'message', cfg = { ask: [CFG.ask, 'ask'], help: [CFG.questions, 'help'], integrate: [CFG.catalogue, 'integrate'], contribute: [CFG.projects, 'contribute'] }[it] || [CFG.questions, 'help'];
    var title, body = t('feed.lead.' + it + '.body');
    if (it === 'help') { var n = d.q.items.length; title = t('feed.lead.help.title.' + (n === 1 ? 'one' : 'many'), { n: n }); }
    else if (it === 'contribute') { var m = d.p.items.length; title = t('feed.lead.contribute.title.' + (m === 1 ? 'one' : 'many'), { n: m }); }
    else if (it === 'integrate') { var k = d.r.covered ? d.r.items.length : 0; title = k ? t('feed.lead.integrate.title.' + (k === 1 ? 'one' : 'many'), { n: k }) : t('feed.lead.integrate.title.any'); }
    else title = t('feed.lead.ask.title');
    return '<section class="home-lead" aria-labelledby="feed-lead-h" data-testid="feed-lead"><span class="home-lead-icon" aria-hidden="true" data-testid="feed-lead-icon-wrap"><span class="icon i-' + ic + '" data-testid="feed-lead-icon"></span></span>' +
      '<div class="home-lead-text"><strong id="feed-lead-h" class="tabular-nums" data-testid="feed-lead-title">' + esc(title) + '</strong><span data-testid="feed-lead-body">' + esc(body) + '</span></div>' +
      '<a class="btn btn-primary" href="' + R(cfg[0]) + '" data-testid="feed-lead-cta"><span data-testid="feed-lead-cta-label">' + esc(t('feed.lead.' + cfg[1] + '.cta')) + '</span><span class="icon i-arrow-right" aria-hidden="true" data-testid="feed-lead-cta-icon"></span></a></section>';
  }
  function R(href) { return href.replace(/^pages\//, ''); }       /* AppConfig routes are relative to the site root; this page lives in pages/ */
  function fill(box, spec, instant) {
    var e = spec.empty, rd = spec.render; spec = Object.assign({}, spec, { render: function (it, b) { rd(it, b); if (A.View) A.View.apply('home'); } });   /* {kind, icon, id, href}: same guided empty state in the simulated and the instant path */
    if (instant) { var items = spec.load(); if (items.length) spec.render(items, box); else box.innerHTML = emptyBox(e.kind, e.icon, e.id, e.href); I18N.apply(box); return; }
    A.Sim.load(box, { delay: spec.delay, skeleton: spec.skeleton || 3, load: spec.load, render: spec.render, empty: { icon: e.icon, titleKey: 'feed.' + e.kind + '.empty.title', bodyKey: 'feed.' + e.kind + '.empty.body', action: { labelKey: 'feed.' + e.kind + '.empty.cta', href: e.href, testid: e.id + '-cta' }, retry: false } });
  }

  function renderFeed(instant) {
    var u = S.user, set = topicSet(u), q = questions(u, set), p = projects(u, set), r = resources(u), mq = myQuestions(u), ma = myApps(u), d = { q: q, p: p, r: r };
    var order = { ask: ['q', 'r', 'p'], help: ['q', 'p', 'r'], integrate: ['r', 'q', 'p'], contribute: ['p', 'q', 'r'] }[u.intent || 'help'] || ['q', 'p', 'r'];
    var skillsHtml = (u.skills || []).length ? '<ul class="tag-list list-reset" data-testid="feed-topics-list">' + u.skills.map(function (s) { return '<li><span class="tag" data-testid="feed-topic-' + esc(s) + '">' + esc(s) + '</span></li>'; }).join('') + '</ul>' : '<span data-i18n="feed.topics.none" data-testid="feed-topics-none"></span>';
    var subs = {
      q: t(q.fallback ? 'feed.q.fallback' : 'feed.q.sub'),
      p: t(p.fallback ? 'feed.p.fallback' : 'feed.p.sub'),
      r: r.covered ? t('feed.r.sub', { country: M.countryName(u.country) }) : t('feed.r.subAny')
    };
    var specs = {
      q: { icon: 'help', title: 'feed.q.title', more: 'feed.q.more', href: R(CFG.questions), lim: 4, spec: { delay: 350, load: function () { return q.items.slice(0, order[0] === 'q' ? 4 : 3); }, render: function (it, b) { b.innerHTML = list(it, rowQ, 'feed-q-list'); }, empty: { kind: 'q', icon: 'help', id: 'feed-q-empty', href: R(CFG.ask) } } },
      p: { icon: 'hammer', title: 'feed.p.title', more: 'feed.p.more', href: R(CFG.projects), spec: { delay: 450, load: function () { return p.items.slice(0, order[0] === 'p' ? 4 : 3); }, render: function (it, b) { b.innerHTML = list(it, rowP, 'feed-p-list'); }, empty: { kind: 'p', icon: 'hammer', id: 'feed-p-empty', href: R(CFG.projects) } } },
      r: { icon: 'compass', title: 'feed.r.title', more: 'feed.r.more', href: R(CFG.catalogue), spec: { delay: 550, load: function () { return r.items.slice(0, order[0] === 'r' ? 4 : 3); }, render: function (it, b) { b.innerHTML = list(it, rowR, 'feed-r-list'); }, empty: { kind: 'r', icon: 'compass', id: 'feed-r-empty', href: R(CFG.catalogue) } } }
    };
    var lrn = M.learn ? { icon: 'book', title: 'feed.l.title', more: 'feed.l.more', href: R(CFG.learn) } : null;
    var main = order.map(function (k) { var s = specs[k]; return block(k, s.icon, s.title, s.more, s.href, esc(subs[k])); }).join('') +
      (lrn ? block('l', lrn.icon, lrn.title, lrn.more, lrn.href, esc(t(M.learn.doneCount('s-ia') ? 'feed.l.sub.cont' : 'feed.l.sub.rec'))) : '');
    var activity = '<section class="id-block" aria-labelledby="feed-a-h" data-testid="feed-a"><div class="id-head"><span class="icon i-clock" aria-hidden="true" data-testid="feed-a-icon"></span><h2 id="feed-a-h" data-i18n="feed.a.title" data-testid="feed-a-title"></h2></div>' +
      '<h3 class="id-sub id-sub-flush" data-i18n="feed.a.q" data-testid="feed-a-q-title" ></h3><div id="feed-a-q-box" data-testid="feed-a-q-box"></div>' +
      '<h3 class="id-sub id-sub-gap" data-i18n="feed.a.apps" data-testid="feed-a-apps-title"></h3><div id="feed-a-apps-box" data-testid="feed-a-apps-box"></div></section>';
    root.innerHTML = '<header class="home-hero" data-testid="feed-hero"><h1 data-testid="feed-h1">' + esc(t('feed.hello', { name: first(u.name) })) + '</h1>' +
      '<div class="home-topics" data-testid="feed-topics"><span data-i18n="feed.topics.label" data-testid="feed-topics-label"></span>' + skillsHtml + '<a href="parametres.html#skills" data-i18n="feed.topics.edit" data-testid="feed-topics-edit"></a><div class="home-view" id="feed-view" data-testid="feed-view"></div></div></header>' +
      lead(u, d) + '<div class="home-layout" data-testid="feed-layout"><div class="home-main" data-testid="feed-main">' + main + '</div><aside class="home-aside" data-testid="feed-aside">' + activity + '</aside></div>';
    I18N.apply(root);
    if (A.View) A.View.init(document.getElementById('feed-view'), '.home-main .list', 'home');
    order.forEach(function (k) { fill(document.getElementById('feed-' + k + '-box'), specs[k].spec, instant); });
    if (lrn) { var lb = document.getElementById('feed-l-box'); A.Sim.load(lb, { delay: instant ? 0 : 600, skeleton: 2, load: function () { return [1]; }, render: function (it, b) { learnBox(b, learnItems(set)); } }); }
    fill(document.getElementById('feed-a-q-box'), { delay: 400, skeleton: 2, load: function () { return mq.slice(0, 3); }, render: function (it, b) { b.innerHTML = list(it, rowMyQ, 'feed-my-list'); }, empty: { kind: 'a.q', icon: 'message', id: 'feed-a-q-empty', href: R(CFG.ask) } }, instant);
    fill(document.getElementById('feed-a-apps-box'), { delay: 500, skeleton: 2, load: function () { return ma.slice(0, 3); }, render: function (it, b) { b.innerHTML = list(it, rowApp, 'feed-app-list'); }, empty: { kind: 'a.apps', icon: 'users', id: 'feed-a-apps-empty', href: R(CFG.projects) } }, instant);
  }

  function render(instant) {
    var u = S.user; if (!u) { S.requireAuth(); return; }
    document.title = t('feed.title');
    if (!(u.skills || []).length) renderFirst(); else renderFeed(instant !== false && mounted);
    mounted = true;
  }

  root.addEventListener('click', function (e) {
    var c = e.target.closest('[data-pick]');
    if (c) { var s = c.getAttribute('data-pick'), i = picked.indexOf(s); if (i > -1) picked.splice(i, 1); else picked.push(s); c.setAttribute('aria-pressed', i < 0); var sv = document.getElementById('feed-first-save'); sv.disabled = !picked.length; document.querySelector('[data-testid="feed-first-count"]').textContent = picked.length ? t('feed.first.count', { n: picked.length }) : t('feed.first.hint'); return; }
    if (e.target.closest('#feed-first-save') && picked.length) { S.update({ skills: picked.slice() }); }
  });
  document.addEventListener('session:change', function () { if (S.signedIn) render(true); });
  document.addEventListener('i18n:change', function () { render(true); });
  render(false);
})();
