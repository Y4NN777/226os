/* Article / tutorial / lesson reading page. Namespace: art.* */
I18N.register({
  fr: {
    'art.title': 'Article · 226os', 'art.crumb.learn': 'Apprendre',
    'art.toc': 'Sur cette page', 'art.related': 'Liens associés',
    'art.rel.q': 'Question liée', 'art.rel.r': 'API ou outil cité', 'art.rel.p': 'Projet associé',
    'art.read': '{n} min de lecture', 'art.copy': 'Copier le lien', 'art.copied': 'Lien copié.',
    'art.lesson.of': 'Leçon {n} sur {total}', 'art.series': 'Formation', 'art.prev': 'Leçon précédente', 'art.next': 'Leçon suivante', 'art.all': 'Toutes les leçons',
    'art.mark': 'Marquer comme terminée', 'art.marked': 'Leçon terminée', 'art.mark.hint': 'Votre progression est enregistrée sur ce compte.',
    'art.tags': 'Sujets', 'art.updated': 'Publié le {date}',
    'art.invite.title': 'Vous aussi, partagez un tutoriel', 'art.invite.body': 'Un problème résolu chez vous aidera quelqu’un d’autre demain. Écrivez-le simplement, avec le code qui fonctionne.', 'art.invite.cta': 'Écrire un tutoriel',
    'art.notfound.title': 'Contenu introuvable', 'art.notfound.body': 'Cet article n’existe pas ou a été retiré.', 'art.notfound.cta': 'Retour à Apprendre'
  },
  en: {
    'art.title': 'Article · 226os', 'art.crumb.learn': 'Learn',
    'art.toc': 'On this page', 'art.related': 'Related links',
    'art.rel.q': 'Related question', 'art.rel.r': 'API or tool cited', 'art.rel.p': 'Related project',
    'art.read': '{n} min read', 'art.copy': 'Copy link', 'art.copied': 'Link copied.',
    'art.lesson.of': 'Lesson {n} of {total}', 'art.series': 'Training', 'art.prev': 'Previous lesson', 'art.next': 'Next lesson', 'art.all': 'All lessons',
    'art.mark': 'Mark as done', 'art.marked': 'Lesson done', 'art.mark.hint': 'Your progress is saved to this account.',
    'art.tags': 'Topics', 'art.updated': 'Published on {date}',
    'art.invite.title': 'You too, share a tutorial', 'art.invite.body': 'A problem solved on your side will help someone else tomorrow. Write it plainly, with the code that works.', 'art.invite.cta': 'Write a tutorial',
    'art.notfound.title': 'Content not found', 'art.notfound.body': 'This article does not exist or was removed.', 'art.notfound.cta': 'Back to Learn'
  }
});
(function () {
  'use strict';
  var root = document.getElementById('art-root'); if (!root) return;
  var A = window.App, M = window.MOCK, L = M.learn, S = A.Session, t = function (k, p) { return I18N.t(k, p); }, esc = A.esc;
  var id = A.params.get('id'), spy = null;
  function ic(n, i) { return '<span class="icon i-' + n + '" aria-hidden="true" data-testid="' + i + '"></span>'; }
  function slug(s) { return L.norm(s).replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '') || 'section'; }
  function questions() { return M.solve ? M.solve.all() : M.questions; }

  function crumbs(parts) {
    return '<nav class="breadcrumbs" aria-label="' + esc(t('common.breadcrumb')) + '" data-testid="art-breadcrumbs"><ol>' + parts.map(function (p, i) {
      return '<li>' + (p.href ? '<a href="' + p.href + '" data-testid="art-crumb-' + i + '">' + esc(p.text) + '</a>' : '<span aria-current="page" data-testid="art-crumb-current">' + esc(p.text) + '</span>') + '</li>';
    }).join('') + '</ol></nav>';
  }
  function related(it) {
    var l = it.links || {}, out = [];
    if (l.q) { var q = M.byId(questions(), l.q); if (q) out.push({ k: 'q', icon: 'help', href: 'question.html?id=' + q.id, title: q.title, lang: q.lang }); }
    if (l.r) { var r = M.byId(M.resources, l.r); if (r) out.push({ k: 'r', icon: 'compass', href: 'resource.html?id=' + r.id, title: r.name }); }
    if (l.p) { var p = M.byId(M.projects, l.p); if (p) out.push({ k: 'p', icon: 'hammer', href: 'project.html?id=' + p.id, title: p.name }); }
    return out;
  }
  function render() {
    var it = L.get(id), forced = A.Sim.forced();
    if (!it || it.type === 'course' || forced) {
      root.innerHTML = crumbs([{ text: t('art.crumb.learn'), href: 'apprendre.html' }, { text: t('art.notfound.title') }]) + '<div id="art-state" data-state-id="art-state" data-testid="art-state"></div>';
      A.Sim.show(document.getElementById('art-state'), forced || 'notfound', { icon: 'file-text', titleKey: 'art.notfound.title', bodyKey: 'art.notfound.body', action: { labelKey: 'art.notfound.cta', href: 'apprendre.html', primary: true, testid: 'art-notfound-cta' } });
      return;
    }
    var lesson = it.type === 'lesson', s = lesson ? L.get(it.series) : null, ls = lesson ? L.lessons(s.id) : [], title = L.txt(it.title), p = L.person(it), mdText = it.body;
    document.title = title + ' · 226os';
    var html = window.Markdown.render(mdText, { id: 'art-md', headStart: 1, copyLabel: function (n) { return t('common.copy') + ' ' + n; }, codeLabel: function (n) { return n; } });
    var heads = [], used = {};
    html = html.replace(/<h2([^>]*)>(.*?)<\/h2>/g, function (m, a, inner) {
      var txt = inner.replace(/<[^>]+>/g, ''), sl = slug(txt); if (used[sl]) sl += '-' + (++used[sl]); else used[sl] = 1;
      heads.push({ id: sl, text: txt.replace(/&amp;/g, '&').replace(/&#39;/g, '’').replace(/&quot;/g, '"').replace(/&lt;/g, '<').replace(/&gt;/g, '>') }); return '<h2 id="' + sl + '"' + a + '>' + inner + '</h2>';
    });
    var tocLinks = heads.map(function (h, i) { return '<li><a href="#' + h.id + '" data-testid="art-toc-link-' + (i + 1) + '">' + esc(h.text) + '</a></li>'; }).join('');
    var tocM = heads.map(function (h, i) { return '<li><a href="#' + h.id + '" data-testid="art-toc-m-link-' + (i + 1) + '">' + esc(h.text) + '</a></li>'; }).join('');
    var who = it.team ? '<span class="ln-by-who" data-testid="art-author">' + A.thumb(t('ln.team'), { icon: 'sparkles', size: 'sm', shape: 'circle', tone: 0, testid: 'art-author-thumb' }) + '<span data-testid="art-author-name">' + esc(t('ln.team')) + '</span></span>'
      : '<span class="ln-by-who" data-testid="art-author">' + A.avatar(p || '?', { size: 'sm', testid: 'art-author-avatar' }) + '<span data-testid="art-author-name">' + esc(p ? p.name : '') + '</span></span>' +
        (p && p.country ? '<span class="country" data-testid="art-author-country"><span class="country-code" data-testid="art-author-country-code">' + p.country + '</span><span class="visually-hidden" data-testid="art-author-country-name">' + esc(M.countryName(p.country)) + '</span></span>' : '');
    var head = '<header class="ln-art-head" data-testid="art-head"><div class="ln-pills" data-testid="art-pills"><span class="badge ln-type ln-type-' + (lesson ? 'course' : it.type) + '" data-testid="art-type">' + ic(L.icon(it), 'art-type-icon') + '<span data-testid="art-type-label">' + esc(t('ln.type.' + it.type)) + '</span></span>' +
      '<span class="badge" data-testid="art-level">' + esc(t('ln.level.' + it.level)) + '</span><span class="badge" data-testid="art-lang"><span aria-hidden="true" data-testid="art-lang-code">' + it.lang.toUpperCase() + '</span><span class="visually-hidden" data-testid="art-lang-name">' + esc(t('ln.lang.' + it.lang)) + '</span></span>' +
      (lesson ? '<a class="badge" href="formation.html?id=' + s.id + '" data-testid="art-series-badge">' + ic('book', 'art-series-badge-icon') + esc(t('art.lesson.of', { n: it.n, total: ls.length })) + '</a>' : '') + '</div>' +
      '<h1' + (lesson ? '' : ' lang="' + it.lang + '" data-user-content') + ' data-testid="art-h1">' + esc(title) + '</h1><p class="ln-art-lede"' + (lesson ? '' : ' lang="' + it.lang + '" data-user-content') + ' data-testid="art-lede">' + esc(L.txt(it.summary)) + '</p>' +
      '<div class="ln-by" data-testid="art-by">' + who + '<time class="meta-item tabular-nums" datetime="' + it.date + '" data-testid="art-date">' + esc(M.fmtDate(it.date)) + '</time><span class="meta-item" data-testid="art-read">' + ic('clock', 'art-read-icon') + esc(t('art.read', { n: it.minutes })) + '</span>' +
      '<button type="button" class="btn btn-icon btn-sm ln-by-copy" data-copylink aria-label="' + esc(t('art.copy')) + '" data-testid="art-copylink">' + ic('link', 'art-copylink-icon') + '</button></div></header>';
    var rel = related(it), relHtml = rel.length ? '<section aria-labelledby="art-rel-h" data-testid="art-related"><h2 class="ln-sec-title" id="art-rel-h" data-testid="art-related-title">' + ic('link', 'art-related-icon') + '<span>' + esc(t('art.related')) + '</span></h2><div class="ln-related" data-testid="art-related-list">' +
      rel.map(function (r) { return '<a class="linked-item" href="' + r.href + '" data-testid="art-rel-' + r.k + '"><span class="kind" data-testid="art-rel-' + r.k + '-kind">' + ic(r.icon, 'art-rel-' + r.k + '-icon') + esc(t('art.rel.' + r.k)) + '</span><strong' + (r.lang ? ' lang="' + r.lang + '" data-user-content' : '') + ' data-testid="art-rel-' + r.k + '-title">' + esc(r.title) + '</strong></a>'; }).join('') + '</div></section>' : '';
    var tags = '<ul class="tag-list list-reset" aria-label="' + esc(t('art.tags')) + '" data-testid="art-tags">' + (it.tags || []).map(function (x) { return '<li><span class="tag" data-testid="art-tag-' + slug(x) + '">' + esc(x) + '</span></li>'; }).join('') + '</ul>';
    var nav = '';
    if (lesson) {
      var i = ls.indexOf(L.get(it.id)), prev = ls[i - 1], next = ls[i + 1], d = L.isDone(it.id);
      nav = '<div class="ln-done" data-testid="art-done"><button type="button" class="btn' + (d ? '' : ' btn-primary') + '" data-markdone aria-pressed="' + d + '" data-testid="art-markdone">' + ic(d ? 'check-circle' : 'check', 'art-markdone-icon') + '<span data-testid="art-markdone-label">' + esc(t(d ? 'art.marked' : 'art.mark')) + '</span></button><p data-testid="art-done-hint">' + esc(t('art.mark.hint')) + '</p></div>' +
        '<nav class="ln-lessonnav" aria-label="' + esc(t('art.series')) + '" data-testid="art-lessonnav">' +
        (prev ? '<a class="ln-nav-card" href="article.html?id=' + prev.id + '" data-testid="art-prev"><span class="kind">' + ic('chevron-left', 'art-prev-icon') + esc(t('art.prev')) + '</span><strong data-testid="art-prev-title">' + esc(L.txt(prev.title)) + '</strong></a>' : '<a class="ln-nav-card" href="formation.html?id=' + s.id + '" data-testid="art-all"><span class="kind">' + ic('book', 'art-all-icon') + esc(t('art.all')) + '</span><strong data-testid="art-all-title">' + esc(L.txt(s.title)) + '</strong></a>') +
        (next ? '<a class="ln-nav-card ln-nav-next' + (d ? ' ln-nav-primary' : '') + '" href="article.html?id=' + next.id + '" data-testid="art-next"><span class="kind">' + esc(t('art.next')) + ic('chevron-right', 'art-next-icon') + '</span><strong data-testid="art-next-title">' + esc(L.txt(next.title)) + '</strong></a>' : '<a class="ln-nav-card ln-nav-next" href="formation.html?id=' + s.id + '" data-testid="art-all-end"><span class="kind">' + esc(t('art.all')) + ic('chevron-right', 'art-all-end-icon') + '</span><strong data-testid="art-all-end-title">' + esc(L.txt(s.title)) + '</strong></a>') + '</nav>';
    }
    var invite = '<section class="ln-invite ln-invite-compact" aria-labelledby="art-inv-h" data-testid="art-invite"><span class="ln-icon-badge" aria-hidden="true" data-testid="art-invite-badge">' + ic('edit', 'art-invite-icon') + '</span><div><h2 id="art-inv-h" data-i18n="art.invite.title" data-testid="art-invite-title"></h2><p data-i18n="art.invite.body" data-testid="art-invite-body"></p></div>' +
      '<a class="btn btn-primary" href="ecrire.html?type=tutorial" data-write data-testid="art-invite-write">' + ic('edit', 'art-invite-write-icon') + '<span data-i18n="art.invite.cta" data-testid="art-invite-write-label"></span></a></section>';
    root.innerHTML = crumbs([{ text: t('art.crumb.learn'), href: 'apprendre.html' }].concat(lesson ? [{ text: L.txt(s.title), href: 'formation.html?id=' + s.id }] : []).concat([{ text: title }])) +
      '<article class="ln-art" data-testid="art">' + head + (heads.length ? '<details class="ln-toc-m" data-testid="art-toc-m"><summary data-testid="art-toc-m-summary"><span data-testid="art-toc-m-label">' + esc(t('art.toc')) + '</span>' + ic('chevron-down', 'art-toc-m-icon') + '</summary><ol data-testid="art-toc-m-list">' + tocM + '</ol></details>' : '') +
      '<div class="ln-art-grid" data-testid="art-grid"><div class="ln-body" data-testid="art-body-col"><div class="prose ln-prose"' + (lesson ? ' lang="fr"' : ' lang="' + it.lang + '"') + ' data-user-content data-testid="art-body">' + html + '</div>' + tags + relHtml + nav + invite + '</div>' +
      (heads.length ? '<aside class="ln-toc" aria-label="' + esc(t('art.toc')) + '" data-testid="art-toc"><div class="ln-toc-box"><p class="ln-toc-title" data-testid="art-toc-title">' + esc(t('art.toc')) + '</p><ol data-testid="art-toc-list">' + tocLinks + '</ol></div></aside>' : '') + '</div></article>';
    I18N.apply(root); if (A.Tabs) A.Tabs.init(root);
    setupSpy();
  }
  function setupSpy() {
    if (spy) spy.disconnect(); var hs = root.querySelectorAll('.ln-prose h2'); if (!hs.length || !window.IntersectionObserver) return;
    spy = new IntersectionObserver(function (es) {
      es.forEach(function (e) { if (e.isIntersecting) { root.querySelectorAll('.ln-toc a').forEach(function (a) { if (a.getAttribute('href') === '#' + e.target.id) a.setAttribute('aria-current', 'location'); else a.removeAttribute('aria-current'); }); } });
    }, { rootMargin: '-72px 0px -70% 0px' });
    hs.forEach(function (h) { spy.observe(h); });
  }
  root.addEventListener('click', function (e) {
    var w = e.target.closest('[data-write]');
    if (w) { e.preventDefault(); var tg = new URL(w.getAttribute('href'), location.href).href; if (S.requireAuth(tg)) location.href = tg; return; }
    if (e.target.closest('[data-markdone]')) { if (!S.requireAuth()) return; var on = !L.isDone(id); L.setDone(id, on); return; }
    if (e.target.closest('[data-copylink]')) {
      var done = function () { A.Toast.show({ kind: 'success', msg: t('art.copied') }); };
      try { navigator.clipboard.writeText(location.href).then(done, done); } catch (x) { done(); }
    }
  });
  document.addEventListener('learn:progress', render); document.addEventListener('i18n:change', render); document.addEventListener('session:change', render);
  render();
})();
