/* Explore / Home page logic + its UI dictionary (domain: home). Mock data from assets/data/*.js. */
I18N.register({
  fr: {
    'home.title': 'Explorer — 226os', 'home.h1': 'Résolvez, découvrez et construisez avec des développeurs africains',
    'home.lede': 'Questions techniques, API et outils locaux, projets open source, membres. Tout est relié.',
    'home.search.label': 'Rechercher sur 226os', 'home.search.placeholder': 'Ex. webhook, USSD, PostGIS', 'home.search.submit': 'Rechercher',
    'home.kind.all': 'Tout', 'home.kind.question': 'Questions', 'home.kind.resource': 'API et outils', 'home.kind.project': 'Projets', 'home.kind.person': 'Membres', 'home.kind.label': 'Type de contenu',
    'home.try': 'Essayez :', 'home.results.title': 'Résultats pour « {q} »', 'home.results.clear': 'Effacer la recherche',
    'home.empty.title': 'Aucun résultat pour « {q} »', 'home.empty.body': 'Vérifiez l’orthographe, essayez un terme plus général ou changez de type de contenu. Vous pouvez aussi poser la question à la communauté.', 'home.empty.cta': 'Poser une question',
    'home.q.title': 'Questions récentes', 'home.q.more': 'Toutes les questions', 'home.q.votes': 'votes', 'home.q.answers': '{n} réponse(s)', 'home.q.accepted': 'Réponse acceptée',
    'home.r.title': 'API et outils par pays', 'home.r.more': 'Parcourir le catalogue', 'home.r.filter': 'Filtrer par pays', 'home.r.verified': 'Source vérifiée', 'home.r.unverified': 'Non vérifié',
    'home.r.countries': 'Pays pris en charge :', 'home.r.questions': '{n} question(s) liée(s)',
    'home.p.title': 'Projets qui cherchent des contributeurs', 'home.p.more': 'Tous les projets', 'home.p.needs': 'Besoins :', 'home.p.status.active': 'Actif', 'home.p.status.idea': 'Idée', 'home.p.status.seeking': 'Cherche des contributeurs',
    'home.m.title': 'Membres', 'home.m.more': 'Voir les profils', 'home.m.open.mentorat': 'Ouvert au mentorat', 'home.m.open.collaboration': 'Ouvert à la collaboration',
    'home.link.title': 'Du problème au projet', 'home.link.lede': 'Exemple fictif : une question mène à l’API utilisée, puis au projet qui en a besoin.',
    'home.link.question': 'Question', 'home.link.resource': 'API / outil', 'home.link.project': 'Projet',
    'home.dom.title': 'Les huit domaines', 'home.dom.lede': 'Ce que couvre le prototype (P0) et ce qui vient ensuite (P1/P2), selon le PRD.',
    'home.dom.col.domain': 'Domaine', 'home.dom.col.now': 'Dans le prototype (P0)', 'home.dom.col.later': 'Prévu ensuite (P1/P2)',
    'home.dom.solve.now': 'Q&R, extraits de code, réponse acceptée, votes, recherche', 'home.dom.solve.later': 'Réputation, articles, discussions, revue de code',
    'home.dom.discover.now': 'Catalogue API/SDK, fiches sourcées, pays, catégories', 'home.dom.discover.later': 'Comparaison, retours, vérifications automatisées',
    'home.dom.build.now': 'Projets GitHub, besoins, candidatures, liens', 'home.dom.build.later': 'Matching, synchronisation GitHub avancée',
    'home.dom.identity.now': 'Connexion et profils minimaux', 'home.dom.identity.later': 'Réputation, fil d’activité, communautés',
    'home.dom.connect.now': 'Maquette seule (P1/P2)', 'home.dom.connect.later': 'Mentorat, carrière, opportunités, groupes d’étude',
    'home.dom.language.now': 'Interface FR/EN', 'home.dom.language.later': 'Traduction des contenus, recherche translingue',
    'home.dom.access.now': 'Responsive, états de chargement, navigation clavier', 'home.dom.access.later': 'Mode léger, cache avancé, brouillons hors ligne',
    'home.dom.intelligence.now': 'Recherche textuelle, liens contextuels', 'home.dom.intelligence.later': 'Suggestions, recherche sémantique, assistant sourcé'
  },
  en: {
    'home.title': 'Explore — 226os', 'home.h1': 'Solve, discover and build with African developers',
    'home.lede': 'Technical questions, local APIs and tools, open-source projects and people: content that links together.',
    'home.search.label': 'Search 226os', 'home.search.placeholder': 'e.g. webhook, USSD, PostGIS', 'home.search.submit': 'Search',
    'home.kind.all': 'All', 'home.kind.question': 'Questions', 'home.kind.resource': 'APIs and tools', 'home.kind.project': 'Projects', 'home.kind.person': 'People', 'home.kind.label': 'Content type',
    'home.try': 'Try:', 'home.results.title': 'Results for “{q}”', 'home.results.clear': 'Clear search',
    'home.empty.title': 'No results for “{q}”', 'home.empty.body': 'Check the spelling, try a more general term or change the content type. You can also ask the community.', 'home.empty.cta': 'Ask a question',
    'home.q.title': 'Recent questions', 'home.q.more': 'All questions', 'home.q.votes': 'votes', 'home.q.answers': '{n} answer(s)', 'home.q.accepted': 'Accepted answer',
    'home.r.title': 'APIs and tools by country', 'home.r.more': 'Browse the catalogue', 'home.r.filter': 'Filter by country', 'home.r.verified': 'Source verified', 'home.r.unverified': 'Verification is dated',
    'home.r.countries': 'Supported countries:', 'home.r.questions': '{n} linked question(s)',
    'home.p.title': 'Projects looking for contributors', 'home.p.more': 'All projects', 'home.p.needs': 'Needs:', 'home.p.status.active': 'Active', 'home.p.status.idea': 'Idea', 'home.p.status.seeking': 'Seeking contributors',
    'home.m.title': 'People', 'home.m.more': 'See profiles', 'home.m.open.mentorat': 'Open to mentoring', 'home.m.open.collaboration': 'Open to collaboration',
    'home.link.title': 'A question, an API, a project: linked', 'home.link.lede': 'Fictional Problem-to-Project example: each item leads to the others without hunting for URLs.',
    'home.link.question': 'Question', 'home.link.resource': 'API / tool', 'home.link.project': 'Project',
    'home.dom.title': 'The eight domains', 'home.dom.lede': 'What is available in this prototype (P0) and what comes later (P1/P2), per the PRD.',
    'home.dom.col.domain': 'Domain', 'home.dom.col.now': 'In the prototype (P0)', 'home.dom.col.later': 'Planned later (P1/P2)',
    'home.dom.solve.now': 'Q&A, code snippets, accepted answer, votes, search', 'home.dom.solve.later': 'Reputation, articles, discussions, code review',
    'home.dom.discover.now': 'API/SDK catalogue, sourced listings, countries, categories', 'home.dom.discover.later': 'Comparison, feedback, automated checks',
    'home.dom.build.now': 'GitHub projects, needs, applications, links', 'home.dom.build.later': 'Matching, advanced GitHub sync',
    'home.dom.identity.now': 'Sign-in and minimal profiles', 'home.dom.identity.later': 'Reputation, activity feed, communities',
    'home.dom.connect.now': 'Mockup only (P1/P2)', 'home.dom.connect.later': 'Mentoring, careers, opportunities, study groups',
    'home.dom.language.now': 'FR/EN interface', 'home.dom.language.later': 'Content translation, cross-language search',
    'home.dom.access.now': 'Responsive, loading states, keyboard navigation', 'home.dom.access.later': 'Light mode, advanced cache, offline drafts',
    'home.dom.intelligence.now': 'Text search, contextual links', 'home.dom.intelligence.later': 'Suggestions, semantic search, sourced assistant'
  }
});

(function () {
  var A = window.App, M = window.MOCK, t = I18N.t, esc = A.esc, $ = A.$, F = A.Filter;
  var resCtl, kind = 'all', q = (A.params.get('q') || '').trim(), countries = [];
  var LANGNAME = { fr: { fr: 'français', en: 'French' }, en: { fr: 'anglais', en: 'English' } };

  function slug(x) { return String(x).toLowerCase().normalize('NFD').replace(/[\u0300-\u036f]/g, '').replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, ''); }
  /* Row ids use the stable data id (`question-row-q-webhook-sig`); rows of the search-results list get a `result-` infix so both lists can coexist. */
  function cc(code, i) { var id = 'country-' + code.toLowerCase() + '-' + i; return '<span class="country" data-testid="' + id + '"><span class="country-code" data-testid="' + id + '-code">' + code + '</span><span class="visually-hidden" data-testid="' + id + '-name">' + esc(M.countryName(code)) + '</span></span>'; }
  function langBadge(l, id) { return '<span class="badge" lang="' + I18N.lang + '" title="' + esc(t('common.originalLang') + (I18N.lang === 'fr' ? '\u00a0: ' : ': ') + LANGNAME[l][I18N.lang]) + '" data-testid="' + id + '"><span aria-hidden="true" data-testid="' + id + '-code">' + l.toUpperCase() + '</span><span class="visually-hidden" data-testid="' + id + '-name">' + esc(t('common.originalLang') + (I18N.lang === 'fr' ? '\u00a0: ' : ': ') + LANGNAME[l][I18N.lang]) + '</span></span>'; }
  function tags(a, id) { return '<ul class="tag-list list-reset">' + a.map(function (x, i) { return '<li><span class="tag" data-testid="' + id + '-tag-' + slug(x) + '">' + esc(x) + '</span></li>'; }).join('') + '</ul>'; }
  function kindBadge(k, i) { var o = M.kinds[k]; return '<span class="badge" data-testid="kind-' + k + '-' + i + '"><span class="icon i-' + o.icon + '" aria-hidden="true" data-testid="kind-' + k + '-' + i + '-icon"></span>' + esc(o[I18N.lang]) + '</span>'; }

  function rowQ(x, i, withKind) {
    var p = M.person(x.author), n = (withKind ? 'result-' : '') + x.id;
    return '<li class="row" data-testid="question-row-' + n + '"><div class="row-lead"><div class="vote" role="img" data-testid="question-vote-' + n + '" aria-label="' + x.votes + ' ' + esc(t('home.q.votes')) + '"><span class="vote-count" data-testid="question-votes-' + n + '">' + x.votes + '</span><span class="text-xs text-muted" aria-hidden="true" data-testid="question-votes-label-' + n + '">' + esc(t('home.q.votes')) + '</span></div></div>' +
      '<div class="row-main"><h3 class="row-title"><a href="pages/question.html?id=' + x.id + '" lang="' + x.lang + '" data-user-content data-testid="question-link-' + n + '">' + esc(x.title) + '</a></h3>' +
      '<p class="row-excerpt" lang="' + x.lang + '" data-user-content data-testid="question-excerpt-' + n + '">' + esc(x.excerpt) + '</p>' + tags(x.tags, 'question-' + n) +
      '<div class="meta"><span class="meta-item">' + cc(x.country, 'q' + n) + '</span><span class="meta-item" data-testid="question-author-' + n + '">' + esc(p.name) + '</span><time class="meta-item" datetime="' + x.date + '" data-testid="question-date-' + n + '">' + M.fmtDate(x.date) + '</time></div></div>' +
      '<div class="row-aside">' + (withKind ? kindBadge('question', n) : '') + (x.accepted ? '<span class="badge badge-accepted" data-testid="question-accepted-' + n + '"><span class="icon i-check" aria-hidden="true" data-testid="question-accepted-icon-' + n + '"></span>' + esc(t('common.accepted')) + '</span>' : '') +
      '<span data-testid="question-answers-' + n + '">' + esc(t('home.q.answers', { n: x.answers })) + '</span>' + langBadge(x.lang, 'question-lang-' + n) + '</div></li>';
  }
  function rowR(x, i, withKind) {
    var n = (withKind ? 'result-' : '') + x.id, cat = M.byId(M.categories, x.category);
    return '<li class="row" data-testid="resource-row-' + n + '"><div class="row-lead"><span class="icon icon-lg i-compass text-muted" aria-hidden="true" data-testid="resource-icon-' + n + '"></span></div>' +
      '<div class="row-main"><h3 class="row-title"><a href="pages/resource.html?id=' + x.id + '" data-testid="resource-link-' + n + '">' + esc(x.name) + '</a></h3>' +
      '<p class="row-excerpt" lang="fr" data-user-content data-testid="resource-summary-' + n + '">' + esc(x.summary) + '</p>' +
      '<div class="meta"><span class="meta-item" data-testid="resource-provider-' + n + '">' + esc(x.provider) + '</span><span class="meta-item" data-testid="resource-category-' + n + '">' + esc(M.label(cat)) + '</span><span class="meta-item mono" data-testid="resource-version-' + n + '">' + esc(x.version) + '</span></div>' +
      '<div class="cluster"><span class="text-sm text-muted" data-testid="resource-countries-label-' + n + '">' + esc(t('home.r.countries')) + '</span>' + x.countries.map(function (c, j) { return cc(c, 'r' + n); }).join('') + '</div></div>' +
      '<div class="row-aside">' + (withKind ? kindBadge('resource', n) : '') + (x.verified ? '<span class="badge badge-verified" data-testid="resource-verified-' + n + '"><span class="icon i-shield-check" aria-hidden="true" data-testid="resource-verified-icon-' + n + '"></span>' + esc(t('common.verified')) + '</span>' : '<span class="badge badge-warn" data-testid="resource-unverified-' + n + '">' + esc(t('home.r.unverified')) + '</span>') +
      '<span data-testid="resource-verified-on-' + n + '">' + esc(t('common.verifiedOn')) + ' <time datetime="' + x.verifiedAt + '" data-testid="resource-verified-date-' + n + '">' + M.fmtDate(x.verifiedAt) + '</time></span></div></li>';
  }
  function rowP(x, i, withKind) {
    var n = (withKind ? 'result-' : '') + x.id;
    return '<li class="row" data-testid="project-row-' + n + '"><div class="row-lead"><span class="icon icon-lg i-hammer text-muted" aria-hidden="true" data-testid="project-icon-' + n + '"></span></div>' +
      '<div class="row-main"><h3 class="row-title"><a href="pages/project.html?id=' + x.id + '" class="mono" data-testid="project-link-' + n + '">' + esc(x.name) + '</a></h3>' +
      '<p class="row-excerpt" lang="' + x.lang + '" data-user-content data-testid="project-summary-' + n + '">' + esc(x.summary) + '</p>' + tags(x.stack, 'project-' + n) +
      '<div class="meta"><span class="meta-item">' + cc(x.country, 'p' + n) + '</span><span class="meta-item" data-testid="project-needs-' + n + '"><span class="text-muted" data-testid="project-needs-label-' + n + '">' + esc(t('home.p.needs')) + '</span> ' + esc(x.needs.join(', ')) + '</span></div></div>' +
      '<div class="row-aside">' + (withKind ? kindBadge('project', n) : '') + '<span class="badge ' + (x.status === 'seeking' ? 'badge-accent' : '') + '" data-testid="project-status-' + n + '">' + esc(t('home.p.status.' + x.status)) + '</span>' + langBadge(x.lang, 'project-lang-' + n) + '</div></li>';
  }
  function rowM(x, i, withKind) {
    var n = (withKind ? 'result-' : '') + x.id;
    return '<li class="row" data-testid="person-row-' + n + '"><div class="row-lead"><span class="avatar" aria-hidden="true" data-testid="person-avatar-' + n + '">' + x.initials + '</span></div>' +
      '<div class="row-main"><h3 class="row-title"><a href="pages/profile.html?id=' + x.id + '" data-testid="person-link-' + n + '">' + esc(x.name) + '</a></h3>' +
      '<p class="row-excerpt" lang="' + x.lang + '" data-user-content data-testid="person-headline-' + n + '">' + esc(x.headline) + '</p>' + tags(x.skills, 'person-' + n) + '</div>' +
      '<div class="row-aside">' + (withKind ? kindBadge('person', n) : '') + '<span class="meta-item">' + cc(x.country, 'm' + n) + ' <span data-testid="person-city-' + n + '">' + esc(x.city) + '</span></span>' +
      x.open.map(function (o, j) { return '<span class="badge" data-testid="person-open-' + n + '-' + o + '">' + esc(t('home.m.open.' + o)) + '</span>'; }).join('') + '</div></li>';
  }
  var ROW = { question: rowQ, resource: rowR, project: rowP, person: rowM };
  var SRC = { question: ['questions', ['title', 'excerpt', 'tags', 'author']], resource: ['resources', ['name', 'provider', 'summary', 'countries', 'category', 'sdk']], project: ['projects', ['name', 'summary', 'stack', 'needs']], person: ['people', ['name', 'headline', 'skills', 'city']] };
  function list(items, fn, id) { return '<ul class="list list-reset" data-testid="' + id + '">' + items.map(function (x, i) { return fn(x, i); }).join('') + '</ul>'; }
  function searchAll(term) {
    var out = {}; Object.keys(SRC).forEach(function (k) { var s = SRC[k]; out[k] = F.run(M[s[0]], { q: term, fields: s[1] }); }); return out;
  }

  /* ---------- search ---------- */
  function renderResults() {
    var box = $('#results'), disc = $('#discover'); if (!box) return;
    box.hidden = !q; disc.hidden = !!q; if (!q) return;
    var all = searchAll(q), ks = kind === 'all' ? Object.keys(SRC) : [kind], rows = [];
    ks.forEach(function (k) { all[k].forEach(function (x) { rows.push({ k: k, x: x }); }); });
    var head = '<div class="section-head"><h2 id="results-h" data-testid="results-title">' + esc(t('home.results.title', { q: q })) + '</h2><a href="index.html" class="btn btn-sm" data-clear-search data-testid="results-clear"><span class="icon i-x" aria-hidden="true" data-testid="results-clear-icon"></span>' + esc(t('home.results.clear')) + '</a></div>';
    var body = rows.length ? '<p class="results-meta" role="status" data-testid="results-count">' + esc(t('common.results', { n: rows.length })) + '</p><ul class="list list-reset" data-testid="results-list">' + rows.map(function (r, i) { return ROW[r.k](r.x, i, true); }).join('') + '</ul>'
      : '<div class="state" data-testid="results-empty" role="status"><span class="icon i-search" aria-hidden="true" data-testid="results-empty-icon"></span><h3 data-testid="results-empty-title">' + esc(t('home.empty.title', { q: q })) + '</h3><p data-testid="results-empty-body">' + esc(t('home.empty.body')) + '</p><a class="btn btn-primary btn-sm" href="pages/solve.html#ask" data-testid="results-empty-ask">' + esc(t('home.empty.cta')) + '</a></div>';
    box.innerHTML = head + body;
    ['question', 'resource', 'project', 'person'].forEach(function (k) { var c = $('[data-kind-count="' + k + '"]'); if (c) c.textContent = '(' + all[k].length + ')'; });
    var ca = $('[data-kind-count="all"]'); if (ca) ca.textContent = '(' + Object.keys(all).reduce(function (s, k) { return s + all[k].length; }, 0) + ')';
  }
  function setQ(v) {
    q = v.trim(); $('#home-q').value = q;
    var u = new URL(location.href); if (q) u.searchParams.set('q', q); else u.searchParams.delete('q'); history.replaceState(null, '', u);
    $$c('[data-kind-count]', function (c) { if (!q) c.textContent = ''; }); renderResults();
  }
  function $$c(s, f) { A.$$(s).forEach(f); }

  /* ---------- discovery sections (simulated async) ---------- */
  function mountSections() {
    A.Sim.load($('#sec-questions'), { load: function () { return F.run(M.questions, { sort: { key: 'date', dir: 'desc' } }).slice(0, 5); }, render: function (it, box) { box.innerHTML = list(it, rowQ, 'questions-list'); } });
    resCtl = A.Sim.load($('#sec-resources'), { delay: 800, skeleton: 3, load: function () { return F.run(M.resources, { where: { countries: countries }, sort: { key: 'name' } }).slice(0, 4); },
      render: function (it, box) { box.innerHTML = list(it, rowR, 'resources-list'); } });
    A.Sim.load($('#sec-projects'), { delay: 700, skeleton: 3, load: function () { return M.projects.filter(function (p) { return p.needs.length; }).slice(0, 3); }, render: function (it, box) { box.innerHTML = list(it, rowP, 'projects-list'); } });
    A.Sim.load($('#sec-people'), { delay: 500, skeleton: 3, load: function () { return M.people.slice(0, 4); }, render: function (it, box) { box.innerHTML = list(it, rowM, 'people-list'); } });
  }
  function renderLinked() {
    var x = M.byId(M.questions, 'q-webhook-sig'), r = M.byId(M.resources, x.resource), p = M.byId(M.projects, x.project);
    function it(k, href, ttl, id) { return '<a class="linked-item" href="' + href + '" data-testid="' + id + '"><span class="kind" data-testid="' + id + '-kind"><span class="icon icon-sm i-' + M.kinds[k].icon + '" aria-hidden="true" data-testid="' + id + '-icon"></span>' + esc(t('home.link.' + k)) + '</span><strong data-testid="' + id + '-title"' + (k === 'question' ? ' lang="' + x.lang + '" data-user-content' : '') + '>' + esc(ttl) + '</strong></a>'; }
    $('#linked').innerHTML = it('question', 'pages/question.html?id=' + x.id, x.title, 'linked-question') + it('resource', 'pages/resource.html?id=' + r.id, r.name, 'linked-resource') + it('project', 'pages/project.html?id=' + p.id, p.name, 'linked-project');
  }
  function renderCountries() {
    var codes = []; M.resources.forEach(function (r) { r.countries.forEach(function (c) { if (codes.indexOf(c) < 0) codes.push(c); }); }); codes.sort();
    $('#country-chips').innerHTML = codes.map(function (c) { return '<button type="button" class="chip" aria-pressed="' + (countries.indexOf(c) > -1) + '" data-country="' + c + '" data-testid="country-chip-' + c.toLowerCase() + '"><span class="country-code" data-testid="country-chip-' + c.toLowerCase() + '-code">' + c + '</span><span class="visually-hidden" data-testid="country-chip-' + c.toLowerCase() + '-name">' + esc(M.countryName(c)) + '</span></button>'; }).join('');
  }
  function renderKinds() {
    var ks = [['all', 'home.kind.all'], ['question', 'home.kind.question'], ['resource', 'home.kind.resource'], ['project', 'home.kind.project'], ['person', 'home.kind.person']];
    $('#kind-chips').innerHTML = ks.map(function (k) { return '<button type="button" class="chip" aria-pressed="' + (kind === k[0]) + '" data-kind="' + k[0] + '" data-testid="kind-chip-' + k[0] + '">' + esc(t(k[1])) + ' <span class="chip-count" data-kind-count="' + k[0] + '" data-testid="kind-chip-' + k[0] + '-count"></span></button>'; }).join('');
    if (q) renderResults();
  }
  function renderDomains() {
    var ids = ['solve', 'discover', 'build', 'identity', 'connect', 'language', 'access', 'intelligence'], body = $('#domains-body'); if (!body) return;
    body.innerHTML = ids.map(function (d) { return '<tr data-testid="domain-row-' + d + '"><th scope="row" data-testid="domain-name-' + d + '">' + esc(t('nav.' + d)) + '</th><td data-testid="domain-now-' + d + '">' + esc(t('home.dom.' + d + '.now')) + '</td><td class="text-muted" data-testid="domain-later-' + d + '">' + esc(t('home.dom.' + d + '.later')) + '</td></tr>'; }).join('');
  }
  function renderTry() {
    $('#try-list').innerHTML = ['webhook', 'Mobile Money', 'USSD', 'PostGIS', 'KYC'].map(function (s, i) { return '<li><a class="chip" href="index.html?q=' + encodeURIComponent(s) + '" data-try="' + esc(s) + '" data-testid="try-' + slug(s) + '">' + esc(s) + '</a></li>'; }).join('');
  }

  function init() {
    document.title = t('home.title'); $('#home-q').value = q;
    renderKinds(); renderCountries(); renderLinked(); renderDomains(); renderTry(); mountSections(); renderResults();
    $('#home-search').addEventListener('submit', function (e) { e.preventDefault(); setQ($('#home-q').value); });
    document.addEventListener('click', function (e) {
      var k = e.target.closest('[data-kind]'), c = e.target.closest('[data-country]'), tr = e.target.closest('[data-try]'), cl = e.target.closest('[data-clear-search]');
      if (k) { kind = k.getAttribute('data-kind'); renderKinds(); }
      if (c) { var v = c.getAttribute('data-country'), i = countries.indexOf(v); i > -1 ? countries.splice(i, 1) : countries.push(v); renderCountries(); resCtl.refresh(); }
      if (tr) { e.preventDefault(); setQ(tr.getAttribute('data-try')); }
      if (cl) { e.preventDefault(); setQ(''); }
    });
    document.addEventListener('i18n:change', function () { document.title = t('home.title'); renderKinds(); renderCountries(); renderLinked(); renderDomains(); renderTry(); renderResults(); });
  }
  init();
})();
