/* Recherche globale: deterministic text search grouped by pillar. Namespace: srch.* */
I18N.register({
  fr: {
    'srch.title': 'Recherche · 226os', 'srch.h1': 'Rechercher sur 226os', 'srch.h1.q': 'Résultats pour « {q} »',
    'srch.label': 'Votre recherche', 'srch.ph': 'Un mot, une API, un projet, un tutoriel', 'srch.go': 'Rechercher',
    'srch.total.one': '1 résultat', 'srch.total.many': '{n} résultats',
    'srch.g.q': 'Questions', 'srch.g.r': 'API & outils', 'srch.g.p': 'Projets open source', 'srch.g.l': 'Apprendre',
    'srch.all': 'Voir tout', 'srch.jump': 'Aller à un groupe',
    'srch.start.title': 'Que cherchez-vous ?', 'srch.start.body': 'La recherche parcourt les questions, les API et outils, les projets et les contenus à apprendre. Essayez l’un de ces sujets.',
    'srch.none.title': 'Aucun résultat pour « {q} »', 'srch.none.body': 'Vérifiez l’orthographe, essayez un mot plus court ou un sujet plus large. Les accents et les majuscules sont ignorés.',
    'srch.none.ask': 'Poser la question', 'srch.none.write': 'Écrire un article', 'srch.none.try': 'Essayez plutôt',
    'srch.group.none': 'Aucun résultat dans ce groupe.',
    'srch.answers': '{n} réponse(s)', 'srch.noanswer': 'Sans réponse',
    'srch.p.needs': '{n} besoin(s) ouvert(s)'
  },
  en: {
    'srch.title': 'Search · 226os', 'srch.h1': 'Search 226os', 'srch.h1.q': 'Results for “{q}”',
    'srch.label': 'Your search', 'srch.ph': 'A word, an API, a project, a tutorial', 'srch.go': 'Search',
    'srch.total.one': '1 result', 'srch.total.many': '{n} results',
    'srch.g.q': 'Questions', 'srch.g.r': 'APIs & tools', 'srch.g.p': 'Open-source projects', 'srch.g.l': 'Learn',
    'srch.all': 'See all', 'srch.jump': 'Jump to a group',
    'srch.start.title': 'What are you looking for?', 'srch.start.body': 'Search covers questions, APIs and tools, projects and learning content. Try one of these topics.',
    'srch.none.title': 'No result for “{q}”', 'srch.none.body': 'Check the spelling, try a shorter word or a broader topic. Accents and capitals are ignored.',
    'srch.none.ask': 'Ask the question', 'srch.none.write': 'Write an article', 'srch.none.try': 'Try instead',
    'srch.group.none': 'No result in this group.',
    'srch.answers': '{n} answer(s)', 'srch.noanswer': 'No answer',
    'srch.p.needs': '{n} open need(s)'
  }
});
(function () {
  'use strict';
  var root = document.getElementById('srch-root'); if (!root) return;
  var A = window.App, M = window.MOCK, L = M.learn, t = function (k, p) { return I18N.t(k, p); }, esc = A.esc, norm = L.norm;
  var q = (A.params.get('q') || '').trim(), LIM = 3;
  function ic(n, i) { return '<span class="icon i-' + n + '" aria-hidden="true" data-testid="' + i + '"></span>'; }
  function words() { return norm(q).split(/\s+/).filter(Boolean); }
  function match(fields) { var h = norm(fields.join(' ')), w = words(); return w.every(function (x) { return h.indexOf(x) > -1; }); }
  function score(title) { var h = norm(title); return words().reduce(function (a, w) { return a + (h.indexOf(w) > -1 ? 2 : 0); }, 0); }
  function mark(s) {
    var out = esc(s), ws = words().filter(function (w) { return w.length > 1; }); if (!ws.length) return out;
    var n = norm(s); if (n.length !== s.length) return out;   /* only when normalisation keeps offsets */
    var spans = []; ws.forEach(function (w) { var i = n.indexOf(w); if (i > -1) spans.push([i, i + w.length]); });
    spans.sort(function (a, b) { return a[0] - b[0]; }); var res = '', last = 0;
    spans.forEach(function (sp) { if (sp[0] >= last) { res += esc(s.slice(last, sp[0])) + '<mark class="ln-mark">' + esc(s.slice(sp[0], sp[1])) + '</mark>'; last = sp[1]; } });
    return spans.length ? res + esc(s.slice(last)) : out;
  }
  function cc(c, id) { return '<span class="country" data-testid="' + id + '"><span class="country-code" data-testid="' + id + '-code">' + c + '</span><span class="visually-hidden">' + esc(M.countryName(c)) + '</span></span>'; }
  function tags(a, id) { return '<ul class="tag-list list-reset" data-testid="' + id + '">' + a.slice(0, 3).map(function (x) { return '<li><span class="tag" data-testid="' + id + '-' + norm(x).replace(/[^a-z0-9]+/g, '-') + '">' + esc(x) + '</span></li>'; }).join('') + '</ul>'; }

  function groups() {
    var qs = (M.solve ? M.solve.all() : M.questions).filter(function (x) { return match([x.title, x.excerpt, x.tags.join(' ')]); }).sort(function (a, b) { return score(b.title) - score(a.title) || b.votes - a.votes; });
    var rs = M.resources.filter(function (x) { var c = M.byId(M.categories, x.category); return match([x.name, x.provider, x.summary, c ? c.fr + ' ' + c.en : '', x.sdk.join(' ')]); }).sort(function (a, b) { return score(b.name) - score(a.name); });
    var ps = (M.build && M.build.projects ? M.build.projects() : M.projects).filter(function (x) { return match([x.name, x.summary, x.stack.join(' '), (x.needs || []).map(function (n) { return n.title || n; }).join(' ')]); }).sort(function (a, b) { return score(b.name) - score(a.name); });
    var ls = L.listable().filter(function (x) { return L.matches(x, q); }).sort(function (a, b) { return score(L.txt(b.title)) - score(L.txt(a.title)); });
    var enc = encodeURIComponent(q);
    return [
      { k: 'q', icon: 'help', items: qs, href: 'solve.html?q=' + enc, row: rowQ },
      { k: 'r', icon: 'compass', items: rs, href: 'discover.html?q=' + enc, row: rowR },
      { k: 'p', icon: 'hammer', items: ps, href: 'build.html?q=' + enc, row: rowP },
      { k: 'l', icon: 'book', items: ls, href: 'apprendre.html?q=' + enc, row: function (x) { return L.row(x, 'srch-l', ''); } }
    ];
  }
  function rowQ(x) {
    var p = M.person(x.author), n = 'srch-q-' + x.id;
    return '<li class="row" data-testid="' + n + '-row"><div class="row-lead">' + A.avatar(p || '?', { testid: n + '-avatar' }) + '</div><div class="row-main"><h3 class="row-title" data-testid="' + n + '-title"><a href="question.html?id=' + x.id + '" lang="' + x.lang + '" data-user-content data-testid="' + n + '-link">' + mark(x.title) + '</a></h3><div class="meta">' + tags(x.tags, n + '-tags') + '<span class="meta-item">' + cc(x.country, n + '-country') + '</span><time class="meta-item" datetime="' + x.date + '" data-testid="' + n + '-date">' + esc(M.fmtDate(x.date)) + '</time></div></div>' +
      '<div class="row-aside">' + (x.accepted ? A.statePill('accepted', { testid: n + '-accepted' }) : '') + '<span class="tabular-nums" data-testid="' + n + '-answers">' + esc(x.answers ? t('srch.answers', { n: x.answers }) : t('srch.noanswer')) + '</span></div></li>';
  }
  function rowR(x) {
    var n = 'srch-r-' + x.id, cat = M.byId(M.categories, x.category);
    return '<li class="row" data-testid="' + n + '-row"><div class="row-lead">' + A.thumb(x.name, { icon: 'compass', testid: n + '-thumb' }) + '</div><div class="row-main"><h3 class="row-title" data-testid="' + n + '-title"><a href="resource.html?id=' + x.id + '" data-testid="' + n + '-link">' + mark(x.name) + '</a></h3><p class="row-excerpt" data-testid="' + n + '-summary">' + esc(x.summary) + '</p><div class="meta"><span class="meta-item" data-testid="' + n + '-category">' + esc(M.label(cat)) + '</span></div></div>' +
      '<div class="row-aside">' + A.statePill(x.verified ? 'verified' : 'unverified', { testid: n + '-pill' }) + '</div></li>';
  }
  function rowP(x) {
    var n = 'srch-p-' + x.id;
    return '<li class="row" data-testid="' + n + '-row"><div class="row-lead">' + A.thumb(x.name, { icon: 'hammer', testid: n + '-thumb' }) + '</div><div class="row-main"><h3 class="row-title" data-testid="' + n + '-title"><a href="project.html?id=' + x.id + '" data-testid="' + n + '-link">' + mark(x.name) + '</a></h3><p class="row-excerpt" lang="' + x.lang + '" data-user-content data-testid="' + n + '-summary">' + esc(x.summary) + '</p><div class="meta">' + tags(x.stack, n + '-tags') + '<span class="meta-item">' + cc(x.country, n + '-country') + '</span></div></div>' +
      '<div class="row-aside">' + A.statePill('open', { testid: n + '-pill' }) + '</div></li>';
  }
  function chips(id) {
    return '<ul class="ln-chips list-reset" data-testid="' + id + '">' + L.suggest.map(function (s) { return '<li><a class="chip" href="recherche.html?q=' + encodeURIComponent(s) + '" data-testid="' + id + '-' + s + '">' + esc(s) + '</a></li>'; }).join('') + '</ul>';
  }
  function render() {
    document.title = (q ? t('srch.h1.q', { q: q }) : t('srch.h1')) + ' · 226os';
    var gs = q ? groups() : [], total = gs.reduce(function (a, g) { return a + g.items.length; }, 0), body = '';
    if (!q) body = '<section class="ln-suggest" data-testid="srch-start"><h2 data-testid="srch-start-title">' + esc(t('srch.start.title')) + '</h2><p class="hint" data-testid="srch-start-body">' + esc(t('srch.start.body')) + '</p>' + chips('srch-suggest') + '</section>';
    else if (!total) body = '<div class="ln-none" role="status" data-testid="srch-none">' + ic('search', 'srch-none-icon') + '<h2 class="h3" data-testid="srch-none-title">' + esc(t('srch.none.title', { q: q })) + '</h2><p data-testid="srch-none-body">' + esc(t('srch.none.body')) + '</p><div class="ln-invite-actions"><a class="btn btn-primary btn-sm" href="solve.html#ask" data-testid="srch-none-ask">' + ic('plus', 'srch-none-ask-icon') + esc(t('srch.none.ask')) + '</a><a class="btn btn-sm" href="ecrire.html" data-write data-testid="srch-none-write">' + ic('edit', 'srch-none-write-icon') + esc(t('srch.none.write')) + '</a></div></div>' +
      '<section class="ln-suggest" data-testid="srch-try"><h2 data-testid="srch-try-title">' + esc(t('srch.none.try')) + '</h2>' + chips('srch-try-list') + '</section>';
    else body = '<nav class="ln-jump" aria-label="' + esc(t('srch.jump')) + '" data-testid="srch-jump">' + gs.map(function (g) { return '<a class="chip" href="#srch-g-' + g.k + '" data-testid="srch-jump-' + g.k + '">' + esc(t('srch.g.' + g.k)) + ' <span class="chip-count" data-testid="srch-jump-' + g.k + '-count">' + g.items.length + '</span></a>'; }).join('') + '</nav>' +
      gs.map(function (g) {
        return '<section class="ln-group" id="srch-g-' + g.k + '" aria-labelledby="srch-g-' + g.k + '-h" data-testid="srch-group-' + g.k + '"><div class="ln-group-head">' + ic(g.icon, 'srch-group-' + g.k + '-icon') + '<h2 id="srch-g-' + g.k + '-h" data-testid="srch-group-' + g.k + '-title">' + esc(t('srch.g.' + g.k)) + '</h2><span class="ln-group-count" data-testid="srch-group-' + g.k + '-count">' + g.items.length + '</span>' +
          (g.items.length ? '<a href="' + g.href + '" data-testid="srch-group-' + g.k + '-all"><span>' + esc(t('srch.all')) + '</span>' + ic('arrow-right', 'srch-group-' + g.k + '-all-icon') + '</a>' : '') + '</div>' +
          (g.items.length ? '<ul class="list list-reset" data-testid="srch-group-' + g.k + '-list">' + g.items.slice(0, LIM).map(g.row).join('') + '</ul>' : '<p class="hint" data-testid="srch-group-' + g.k + '-none">' + esc(t('srch.group.none')) + '</p>') + '</section>';
      }).join('');
    root.innerHTML = '<header class="page-header" data-testid="srch-header"><h1 data-testid="srch-h1">' + esc(q ? t('srch.h1.q', { q: q }) : t('srch.h1')) + '</h1>' + (q ? '<p class="results-meta" role="status" aria-live="polite" data-testid="srch-total">' + esc(t(total === 1 ? 'srch.total.one' : 'srch.total.many', { n: total })) + '</p>' : '') + '</header>' +
      '<div class="ln-search" data-testid="srch-box"><form role="search" action="recherche.html" method="get" data-testid="srch-form"><div class="search-field search-field-lg">' + ic('search', 'srch-input-icon') + '<label class="visually-hidden" for="srch-q" data-testid="srch-label">' + esc(t('srch.label')) + '</label><input id="srch-q" class="input" type="search" name="q" value="' + esc(q) + '" autocomplete="off" placeholder="' + esc(t('srch.ph')) + '" data-testid="srch-input"></div><button type="submit" class="btn btn-primary" data-testid="srch-submit"><span data-testid="srch-submit-label">' + esc(t('srch.go')) + '</span></button></form></div>' + body;
    I18N.apply(root);
  }
  root.addEventListener('click', function (e) { var w = e.target.closest('[data-write]'); if (w) { e.preventDefault(); var tg = new URL(w.getAttribute('href'), location.href).href; if (A.Session.requireAuth(tg)) location.href = tg; } });
  document.addEventListener('i18n:change', render);
  render();
  var inp = document.getElementById('srch-q'); if (inp) { inp.focus(); inp.setSelectionRange(inp.value.length, inp.value.length); }
})();
