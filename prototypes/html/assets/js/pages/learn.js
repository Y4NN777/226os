/* Apprendre (hub): featured free training, filters (type, topic, level, language, text), list/grid toggle, invitation to write. Namespace: lrn.* (shared ln.* lives in assets/data/learn.js). */
I18N.register({
  fr: {
    'lrn.title': 'Apprendre : articles, tutoriels et formations · 226os',
    'lrn.h1': 'Apprendre',
    'lrn.lede': 'Articles, tutoriels et formations gratuites, écrits par des développeurs africains pour des problèmes concrets d’ici.',
    'lrn.write': 'Écrire un article',
    'lrn.feature.label': 'Formation à la une',
    'lrn.feature.free': 'Formation gratuite',
    'lrn.feature.program': 'Voir le programme',
    'lrn.feature.steps': 'Au programme',
    'lrn.feature.signin': 'Connectez-vous pour suivre votre progression.',
    'lrn.feature.done': 'Formation terminée. Bravo !',
    'lrn.filters.type': 'Type', 'lrn.filters.all': 'Tout',
    'lrn.filters.group': 'Filtrer par type',
    'lrn.filters.q': 'Rechercher', 'lrn.filters.q.ph': 'Un mot, un sujet, un auteur',
    'lrn.filters.topic': 'Sujet', 'lrn.filters.level': 'Niveau', 'lrn.filters.lang': 'Langue',
    'lrn.filters.any': 'Tous',
    'lrn.count.one': '1 contenu', 'lrn.count.many': '{n} contenus',
    'lrn.empty.title': 'Aucun contenu ne correspond',
    'lrn.empty.body': 'Essayez d’autres mots, retirez un filtre, ou écrivez vous-même ce qui manque.',
    'lrn.empty.reset': 'Réinitialiser les filtres',
    'lrn.invite.title': 'Partagez ce que vous savez',
    'lrn.invite.body': 'Un tutoriel sur l’intégration qui vous a coûté deux jours en fera gagner un à la personne suivante. Articles et tutoriels sont gratuits, lisibles sans compte, en français ou en anglais.',
    'lrn.invite.cta': 'Écrire un article'
  },
  en: {
    'lrn.title': 'Learn: articles, tutorials and trainings · 226os',
    'lrn.h1': 'Learn',
    'lrn.lede': 'Free articles, tutorials and trainings written by African developers about concrete problems from here.',
    'lrn.write': 'Write an article',
    'lrn.feature.label': 'Featured training',
    'lrn.feature.free': 'Free training',
    'lrn.feature.program': 'See the programme',
    'lrn.feature.steps': 'In this training',
    'lrn.feature.signin': 'Sign in to track your progress.',
    'lrn.feature.done': 'Training completed. Well done!',
    'lrn.filters.type': 'Type', 'lrn.filters.all': 'All',
    'lrn.filters.group': 'Filter by type',
    'lrn.filters.q': 'Search', 'lrn.filters.q.ph': 'A word, a topic, an author',
    'lrn.filters.topic': 'Topic', 'lrn.filters.level': 'Level', 'lrn.filters.lang': 'Language',
    'lrn.filters.any': 'All',
    'lrn.count.one': '1 item', 'lrn.count.many': '{n} items',
    'lrn.empty.title': 'Nothing matches',
    'lrn.empty.body': 'Try other words, remove a filter, or write what is missing yourself.',
    'lrn.empty.reset': 'Reset filters',
    'lrn.invite.title': 'Share what you know',
    'lrn.invite.body': 'A tutorial about the integration that cost you two days saves one for the next person. Articles and tutorials are free, readable without an account, in French or English.',
    'lrn.invite.cta': 'Write an article'
  }
});

(function () {
  'use strict';
  var root = document.getElementById('lrn-root'); if (!root) return;
  var A = window.App, M = window.MOCK, L = M.learn, S = A.Session, t = function (k, p) { return I18N.t(k, p); }, esc = A.esc;
  var P = new URLSearchParams(location.search);
  var F = { type: P.get('type') || '', topic: P.get('topic') || '', level: P.get('level') || '', lang: P.get('lang_f') || '', q: P.get('q') || '' };
  var TYPES = ['article', 'tutorial', 'course'], sim = null;
  function ic(n, id) { return '<span class="icon i-' + n + '" aria-hidden="true" data-testid="' + id + '"></span>'; }

  function writeBtn(id, cls, labelKey) {
    return '<a class="btn ' + cls + '" href="ecrire.html" data-write data-testid="' + id + '">' + ic('edit', id + '-icon') + '<span data-i18n="' + labelKey + '" data-testid="' + id + '-label"></span></a>';
  }

  /* ---------- featured training ---------- */
  function feature() {
    var s = L.seriesList[0], ls = L.lessons(s.id), done = L.doneCount(s.id), next = L.nextLesson(s.id), n = ls.length, inn = S.signedIn;
    var cta, primaryHref = 'formation.html?id=' + s.id;
    if (inn && done && next) cta = t('ln.continue', { n: next.n }); else if (inn && done && !next) cta = t('ln.review'); else cta = t('ln.start');
    var progress = inn && done ? '<div class="ln-progress" data-testid="learn-feature-progress"><div class="ln-progress-text"><strong data-testid="learn-feature-progress-text">' + esc(t('ln.progress', { done: done, total: n })) + '</strong><span class="tabular-nums" data-testid="learn-feature-progress-pct">' + Math.round(100 * done / n) + ' %</span></div>' +
      '<div class="ln-bar" role="progressbar" aria-valuemin="0" aria-valuemax="' + n + '" aria-valuenow="' + done + '" aria-label="' + esc(t('ln.progress', { done: done, total: n })) + '" data-testid="learn-feature-bar"><span style="width:' + Math.round(100 * done / n) + '%"></span></div></div>' : '';
    var steps = ls.map(function (l) {
      var d = L.isDone(l.id);
      return '<a class="ln-step' + (d ? ' is-done' : '') + '" href="article.html?id=' + l.id + '" data-testid="learn-feature-step-' + l.n + '"><span class="ln-step-n" data-testid="learn-feature-step-' + l.n + '-n">' + (d ? ic('check', 'learn-feature-step-' + l.n + '-tick') + '<span class="visually-hidden" data-testid="learn-feature-step-' + l.n + '-done">' + esc(t('ln.done')) + '</span>' : l.n) + '</span>' +
        '<span class="ln-step-t" data-testid="learn-feature-step-' + l.n + '-title">' + esc(L.txt(l.title)) + '</span><span class="ln-step-m" data-testid="learn-feature-step-' + l.n + '-min">' + esc(L.fmtMin(l.minutes)) + '</span></a>';
    }).join('');
    return '<section class="ln-feature" aria-labelledby="lrn-feature-h" data-testid="learn-feature"><div class="ln-feature-main" data-testid="learn-feature-main">' +
      '<p class="ln-kicker" data-testid="learn-feature-kicker">' + ic('sparkles', 'learn-feature-kicker-icon') + '<span data-i18n="ln.team" data-testid="learn-feature-team"></span></p>' +
      '<h2 id="lrn-feature-h" data-testid="learn-feature-title">' + esc(L.txt(s.title)) + '</h2>' +
      '<p class="ln-feature-sum" data-testid="learn-feature-summary">' + esc(L.txt(s.summary)) + '</p>' +
      '<p class="ln-facts" data-testid="learn-feature-facts"><span class="meta-item" data-testid="learn-feature-lessons">' + ic('book', 'learn-feature-lessons-icon') + esc(t('ln.lessons', { n: n })) + '</span>' +
      '<span class="meta-item" data-testid="learn-feature-duration">' + ic('clock', 'learn-feature-duration-icon') + esc(L.fmtMin(L.minutesOf(s.id))) + '</span>' +
      '<span class="meta-item" data-testid="learn-feature-level">' + esc(t('ln.level.range', { a: t('ln.level.' + s.level).toLowerCase(), b: t('ln.level.' + s.levelTo).toLowerCase() })) + '</span>' +
      '<span class="meta-item" data-testid="learn-feature-free">' + ic('check-circle', 'learn-feature-free-icon') + '<span data-i18n="ln.free" data-testid="learn-feature-free-label"></span></span></p>' + progress +
      '<div class="ln-actions" data-testid="learn-feature-actions"><a class="btn btn-primary" href="' + (done && next ? 'article.html?id=' + next.id : done ? primaryHref : 'article.html?id=' + ls[0].id) + '" data-testid="learn-feature-cta"><span data-testid="learn-feature-cta-label">' + esc(cta) + '</span>' + ic('arrow-right', 'learn-feature-cta-icon') + '</a>' +
      '<a class="ln-link" href="' + primaryHref + '" data-testid="learn-feature-program"><span data-i18n="lrn.feature.program" data-testid="learn-feature-program-label"></span>' + ic('chevron-right', 'learn-feature-program-icon') + '</a></div>' +
      (inn ? '' : '<p class="hint" data-i18n="lrn.feature.signin" data-testid="learn-feature-signin"></p>') +
      '</div><div class="ln-steps ln-steps-fold" role="group" aria-labelledby="lrn-steps-h" data-testid="learn-feature-steps"><p class="ln-steps-title" id="lrn-steps-h" data-i18n="lrn.feature.steps" data-testid="learn-feature-steps-title"></p>' + steps + '</div></section>';
  }

  /* ---------- toolbar ---------- */
  function counts() {
    var all = L.listable(), c = { '': all.length }; TYPES.forEach(function (k) { c[k] = all.filter(function (x) { return x.type === k; }).length; }); return c;
  }
  function opts(list, keyPrefix, cur, allKey) {
    return '<option value="">' + esc(t(allKey)) + '</option>' + list.map(function (k) { return '<option value="' + k + '"' + (cur === k ? ' selected' : '') + '>' + esc(t(keyPrefix + k)) + '</option>'; }).join('');
  }
  function toolbar() {
    var c = counts();
    var chips = [''].concat(TYPES).map(function (k) {
      return '<button type="button" class="chip" data-type="' + k + '" aria-pressed="' + (F.type === k) + '" data-testid="learn-type-' + (k || 'all') + '"><span data-testid="learn-type-' + (k || 'all') + '-label">' + esc(k ? t('ln.type.' + k) + 's' : t('lrn.filters.all')) + '</span> <span class="chip-count" data-testid="learn-type-' + (k || 'all') + '-count">' + c[k] + '</span></button>';
    }).join('');
    return '<div class="ln-toolbar" data-testid="learn-toolbar"><div class="ln-chips" role="group" aria-label="' + esc(t('lrn.filters.group')) + '" data-testid="learn-types">' + chips + '</div>' +
      '<div class="ln-toolbar-main"><div class="field"><label class="label" for="lrn-q" data-i18n="lrn.filters.q" data-testid="learn-q-label"></label><div class="search-field">' + ic('search', 'learn-q-icon') +
      '<input id="lrn-q" class="input" type="search" name="q" value="' + esc(F.q) + '" autocomplete="off" placeholder="' + esc(t('lrn.filters.q.ph')) + '" data-testid="learn-q-input"></div></div>' +
      '<div class="ln-selects" data-testid="learn-selects">' +
      '<div class="field"><label class="label" for="lrn-topic" data-i18n="lrn.filters.topic" data-testid="learn-topic-label"></label><select id="lrn-topic" class="select" data-testid="learn-topic-input">' + opts(L.topics, 'ln.topic.', F.topic, 'lrn.filters.any') + '</select></div>' +
      '<div class="field"><label class="label" for="lrn-level" data-i18n="lrn.filters.level" data-testid="learn-level-label"></label><select id="lrn-level" class="select" data-testid="learn-level-input">' + opts(L.levels, 'ln.level.', F.level, 'lrn.filters.any') + '</select></div>' +
      '<div class="field"><label class="label" for="lrn-lang" data-i18n="lrn.filters.lang" data-testid="learn-lang-label"></label><select id="lrn-lang" class="select" data-testid="learn-lang-input">' + opts(['fr', 'en'], 'ln.lang.', F.lang, 'lrn.filters.any') + '</select></div></div></div></div>';
  }

  function filtered() {
    return L.listable().filter(function (x) {
      return (!F.type || x.type === F.type) && (!F.topic || x.topic === F.topic) && (!F.lang || x.lang === F.lang) && L.matches(x, F.q) &&
        (!F.level || x.level === F.level || (x.type === 'course' && L.levels.indexOf(F.level) >= L.levels.indexOf(x.level) && L.levels.indexOf(F.level) <= L.levels.indexOf(x.levelTo)));
    });
  }
  function syncUrl() {
    var u = new URL(location.href);
    [['type', F.type], ['topic', F.topic], ['level', F.level], ['lang_f', F.lang], ['q', F.q]].forEach(function (p) { if (p[1]) u.searchParams.set(p[0], p[1]); else u.searchParams.delete(p[0]); });
    try { history.replaceState(null, '', u.href); } catch (e) {}
  }
  function reset() { F = { type: '', topic: '', level: '', lang: '', q: '' }; syncUrl(); mount(); }

  function drawList(items, box) {
    box.innerHTML = '<ul class="list list-reset ln-list" data-testid="learn-list">' + items.map(function (x) { return L.row(x, 'learn-item', ''); }).join('') + '</ul>';
    if (A.View) A.View.apply('learn');
    var n = items.length, r = document.getElementById('lrn-count'); if (r) r.textContent = t(n === 1 ? 'lrn.count.one' : 'lrn.count.many', { n: n });
  }
  function results() {
    var box = document.getElementById('lrn-list');
    sim = A.Sim.load(box, {
      delay: 250, skeleton: 3, load: filtered, render: drawList,
      empty: { icon: 'search', titleKey: 'lrn.empty.title', bodyKey: 'lrn.empty.body', action: { labelKey: 'lrn.empty.reset', icon: 'refresh', onClick: reset, testid: 'learn-empty-reset' }, retry: false }
    });
  }
  function refresh() {
    syncUrl(); var items = filtered(), r = document.getElementById('lrn-count'); if (r) r.textContent = t(items.length === 1 ? 'lrn.count.one' : 'lrn.count.many', { n: items.length });
    if (sim) sim.refresh();
  }

  function mount() {
    root.innerHTML = '<header class="page-header ln-head" data-testid="learn-header"><div class="page-header-row"><div class="stack-sm"><h1 data-i18n="lrn.h1" data-testid="learn-h1"></h1><p class="lede" data-i18n="lrn.lede" data-testid="learn-lede"></p></div>' +
      writeBtn('learn-write', 'btn-primary', 'lrn.write') + '</div></header>' +
      feature() +
      '<section class="ln-browse" aria-label="' + esc(t('lrn.h1')) + '" data-testid="learn-browse">' + toolbar() +
      '<div class="ln-resultsbar" data-testid="learn-resultsbar"><p class="results-meta" id="lrn-count" role="status" aria-live="polite" data-testid="learn-count"></p><div id="lrn-view" data-testid="learn-view"></div></div>' +
      '<div id="lrn-list" data-state-id="learn-state" data-testid="learn-list-box"></div></section>' +
      '<section class="ln-invite" aria-labelledby="lrn-invite-h" data-testid="learn-invite"><span class="ln-icon-badge" aria-hidden="true" data-testid="learn-invite-badge">' + ic('edit', 'learn-invite-icon') + '</span>' +
      '<div><h2 id="lrn-invite-h" data-i18n="lrn.invite.title" data-testid="learn-invite-title"></h2><p data-i18n="lrn.invite.body" data-testid="learn-invite-body"></p></div>' +
      '<div class="ln-invite-actions">' + writeBtn('learn-invite-write', 'btn-primary', 'lrn.invite.cta') + '</div></section>';
    document.title = t('lrn.title'); I18N.apply(root);
    if (A.View) A.View.init(document.getElementById('lrn-view'), '.ln-list', 'learn');
    results();
    var items = filtered(); document.getElementById('lrn-count').textContent = t(items.length === 1 ? 'lrn.count.one' : 'lrn.count.many', { n: items.length });
  }

  root.addEventListener('click', function (e) {
    var c = e.target.closest('[data-type]');
    if (c) { F.type = c.getAttribute('data-type'); root.querySelectorAll('[data-type]').forEach(function (b) { b.setAttribute('aria-pressed', String(b === c)); }); refresh(); return; }
    var w = e.target.closest('[data-write]');
    if (w) { e.preventDefault(); var target = new URL(w.getAttribute('href'), location.href).href; if (S.requireAuth(target)) location.href = target; }
  });
  root.addEventListener('change', function (e) {
    var m = { 'lrn-topic': 'topic', 'lrn-level': 'level', 'lrn-lang': 'lang' }[e.target.id]; if (m) { F[m] = e.target.value; refresh(); }
  });
  root.addEventListener('input', function (e) { if (e.target.id === 'lrn-q') { F.q = e.target.value; refresh(); } });
  document.addEventListener('i18n:change', mount);
  document.addEventListener('session:change', mount);
  mount();
})();
