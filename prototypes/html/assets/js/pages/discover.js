/* Discover catalogue page logic + UI dictionary (domain: discover). Data: MOCK.discover (fictional). */
I18N.register({
  fr: {
    'discover.title': 'Discover : API et outils — 226os', 'discover.h1': 'API, SDK et outils pour développeurs',
    'discover.lede': 'Un annuaire sourcé d’API, de SDK et d’outils utiles en Afrique : filtrez par pays, catégorie, langage de SDK, fournisseur ou langue de documentation, puis ouvrez une fiche avec ses sources datées.',
    'discover.compare': 'Comparer', 'discover.fictional.title': 'Exemples fictifs. ', 'discover.fictional.body': 'Les fournisseurs, API et sources de ce prototype sont inventés : aucune entreprise ni intégration réelle n’est suggérée.',
    'discover.filters.h': 'Recherche et filtres', 'discover.results.h': 'Résultats du catalogue',
    'discover.search.label': 'Rechercher une API, un SDK ou un outil', 'discover.search.placeholder': 'Ex. webhook, USSD, géocodage',
    'discover.f.country': 'Pays pris en charge', 'discover.f.category': 'Catégorie', 'discover.f.type': 'Type', 'discover.f.doclang': 'Langue de la documentation', 'discover.f.sdk': 'SDK / langage', 'discover.f.provider': 'Fournisseur',
    'discover.all.country': 'Tous les pays', 'discover.all.category': 'Toutes les catégories', 'discover.all.type': 'Tous les types', 'discover.all.doclang': 'Toutes les langues', 'discover.all.sdk': 'Tous les langages', 'discover.all.provider': 'Tous les fournisseurs',
    'discover.sort.verified': 'Vérification la plus récente', 'discover.sort.name': 'Nom (A–Z)', 'discover.sort.countries': 'Couverture (pays)',
    'discover.reset': 'Réinitialiser', 'discover.count': '{n} fiche(s)', 'discover.count.filtered': '{n} fiche(s) pour les filtres actifs',
    'discover.docs': 'Doc.', 'discover.docs.label': 'Langue de la documentation :', 'discover.countries.label': 'Pays :',
    'discover.verifiedOn': 'Vérifié le', 'discover.sources': '{n} source(s)', 'discover.questions': '{n} question(s) liée(s)',
    'discover.status.verified': 'Source vérifiée', 'discover.status.stale': 'Vérification ancienne', 'discover.status.unverified': 'Non vérifié',
    'discover.fictional': 'exemple fictif', 'discover.type.api': 'API', 'discover.type.sdk': 'SDK', 'discover.type.tool': 'Outil',
    'discover.empty.title': 'Aucune fiche ne correspond', 'discover.empty.body': 'Retirez un filtre ou essayez un autre pays ou une autre catégorie. Vous pouvez aussi poser la question à la communauté.',
    'discover.error.title': 'Impossible de charger le catalogue', 'discover.error.body': 'La connexion simulée a échoué. Vos filtres sont conservés ; réessayez.',
    'discover.empty.ask': 'Poser une question',
    'discover.note.langs': 'Résumé en {l}'
  },
  en: {
    'discover.title': 'Discover: APIs and tools — 226os', 'discover.h1': 'APIs, SDKs and developer tools',
    'discover.lede': 'A sourced directory of APIs, SDKs and tools useful in Africa: filter by country, category, SDK language, provider or documentation language, then open a listing with its dated sources.',
    'discover.compare': 'Compare', 'discover.fictional.title': 'Fictional samples. ', 'discover.fictional.body': 'Providers, APIs and sources in this prototype are invented: no real company or integration is implied.',
    'discover.filters.h': 'Search and filters', 'discover.results.h': 'Catalogue results',
    'discover.search.label': 'Search for an API, SDK or tool', 'discover.search.placeholder': 'e.g. webhook, USSD, geocoding',
    'discover.f.country': 'Supported country', 'discover.f.category': 'Category', 'discover.f.type': 'Type', 'discover.f.doclang': 'Documentation language', 'discover.f.sdk': 'SDK / language', 'discover.f.provider': 'Provider',
    'discover.all.country': 'All countries', 'discover.all.category': 'All categories', 'discover.all.type': 'All types', 'discover.all.doclang': 'All languages', 'discover.all.sdk': 'All languages', 'discover.all.provider': 'All providers',
    'discover.sort.verified': 'Most recently verified', 'discover.sort.name': 'Name (A–Z)', 'discover.sort.countries': 'Coverage (countries)',
    'discover.reset': 'Reset', 'discover.count': '{n} listing(s)', 'discover.count.filtered': '{n} listing(s) for the active filters',
    'discover.docs': 'Docs', 'discover.docs.label': 'Documentation language:', 'discover.countries.label': 'Countries:',
    'discover.verifiedOn': 'Verified on', 'discover.sources': '{n} source(s)', 'discover.questions': '{n} linked question(s)',
    'discover.status.verified': 'Source verified', 'discover.status.stale': 'Verification is dated', 'discover.status.unverified': 'Not verified',
    'discover.fictional': 'fictional sample', 'discover.type.api': 'API', 'discover.type.sdk': 'SDK', 'discover.type.tool': 'Tool',
    'discover.empty.title': 'No listing matches', 'discover.empty.body': 'Remove a filter or try another country or category. You can also ask the community.',
    'discover.error.title': 'Could not load the catalogue', 'discover.error.body': 'The simulated connection failed. Your filters are kept; please retry.',
    'discover.empty.ask': 'Ask a question',
    'discover.note.langs': 'Summary in {l}'
  }
});

(function () {
  'use strict';
  var A = window.App, M = window.MOCK, D = M.discover, t = I18N.t, esc = A.esc, $ = A.$, F = A.Filter, P = A.params;
  var LN = { fr: { fr: 'français', en: 'French' }, en: { fr: 'anglais', en: 'English' } };
  var STATUS = { verified: { cls: 'badge-verified', icon: 'shield-check' }, stale: { cls: 'badge-warn', icon: 'clock' }, unverified: { cls: 'badge-warn', icon: 'alert' } };
  var TYPE_ICON = { api: 'network', sdk: 'code', tool: 'hammer' };
  var form = $('#discover-form'), ctl;

  var SDKN = { go: 'Go', node: 'Node.js', php: 'PHP', python: 'Python', rust: 'Rust' };
  function provClean(r) { return r.provider.replace(/\s*\(fictif\)\s*/, '').replace(/\s*fictif\s*/, ' ').trim(); }
  function slug(s) { return s.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, ''); }
  D.items.forEach(function (r) {
    var c = D.cat(r.category);
    r._prov = provClean(r); r._pslug = slug(r._prov);
    r._hay = [r.name, r.provider, (r.sdk || []).map(function (k) { return k + ' ' + (SDKN[k] || k); }).join(' '), r.summary, r.category, c.fr, c.en, r.type, r.version, r.countries.map(function (x) { return x + ' ' + M.countryName(x); }).join(' ')].join(' ');
    r._n = r.countries.length;
  });

  function lbl(o) { return o[I18N.lang]; }
  function cc(code, id) { return '<span class="country" data-testid="' + id + '"><span class="country-code" data-testid="' + id + '-code">' + code + '</span><span class="visually-hidden" data-testid="' + id + '-name">' + esc(M.countryName(code)) + '</span></span>'; }
  function langBadge(l, id, extra) { var n = LN[l][I18N.lang]; return '<span class="badge" lang="' + I18N.lang + '" data-testid="' + id + '"><span aria-hidden="true" data-testid="' + id + '-code">' + (extra || '') + l.toUpperCase() + '</span><span class="visually-hidden" data-testid="' + id + '-name">' + esc((extra ? t('discover.docs.label') + ' ' : t('common.originalLang') + ' : ') + n) + '</span></span>'; }
  function providerName(r) { return esc(r.provider.replace(/\s*\(fictif\)\s*/, '').replace(/\s*fictif\s*/, ' ').trim()) + ' · ' + esc(t('discover.fictional')); }
  function statusBadge(r, id) { var s = D.status(r), o = STATUS[s]; return '<span class="badge ' + o.cls + '" data-testid="' + id + '"><span class="icon i-' + o.icon + '" aria-hidden="true" data-testid="' + id + '-icon"></span>' + esc(t('discover.status.' + s)) + '</span>'; }

  function opts(sel, all, list, keyFn) {
    var cur = sel.value, pre = sel.id + '-option-';   /* option ids: discover-country-option-all | -<value> */
    sel.innerHTML = '<option value="" data-testid="' + pre + 'all">' + esc(all) + '</option>' + list.map(function (x) { return '<option value="' + x.v + '" data-testid="' + pre + x.v + '">' + esc(x.l) + '</option>'; }).join(''); sel.value = cur;
  }
  function fillSelects() {
    var used = {}; D.items.forEach(function (r) { r.countries.forEach(function (c) { used[c] = 1; }); });
    opts($('#discover-country'), t('discover.all.country'), M.countries.filter(function (c) { return used[c.code]; }).map(function (c) { return { v: c.code, l: c.code + ' — ' + lbl(c) }; }));
    opts($('#discover-category'), t('discover.all.category'), D.categories.map(function (c) { return { v: c.id, l: lbl(c) }; }));
    opts($('#discover-type'), t('discover.all.type'), D.types.map(function (c) { return { v: c.id, l: lbl(c) }; }));
    var sd = {}, pv = {}; D.items.forEach(function (r) { (r.sdk || []).forEach(function (k) { sd[k] = 1; }); pv[r._pslug] = r._prov; });
    opts($('#discover-sdk'), t('discover.all.sdk'), Object.keys(sd).sort().map(function (k) { return { v: k, l: SDKN[k] || k }; }));
    opts($('#discover-provider'), t('discover.all.provider'), Object.keys(pv).sort(function (a, b) { return pv[a].localeCompare(pv[b]); }).map(function (k) { return { v: k, l: pv[k] }; }));
    opts($('#discover-doclang'), t('discover.all.doclang'), [{ v: 'fr', l: 'Français (FR)' }, { v: 'en', l: 'English (EN)' }]);
    var s = $('#discover-sort'), cur = s.value || 'verified';
    s.innerHTML = ['verified', 'name', 'countries'].map(function (k) { return '<option value="' + k + '" data-testid="discover-sort-option-' + k + '">' + esc(t('discover.sort.' + k)) + '</option>'; }).join(''); s.value = cur;
  }
  function vals() {
    return { q: $('#discover-q').value, country: $('#discover-country').value, category: $('#discover-category').value, type: $('#discover-type').value, doclang: $('#discover-doclang').value, sdk: $('#discover-sdk').value, provider: $('#discover-provider').value, sort: $('#discover-sort').value };
  }
  function active(v) { return !!(v.q || v.country || v.category || v.type || v.doclang || v.sdk || v.provider); }
  function load() {
    var v = vals(), sortMap = { verified: { key: 'verifiedAt', dir: 'desc' }, name: { key: 'name', dir: 'asc' }, countries: { key: '_n', dir: 'desc' } };
    $('#discover-reset').disabled = !active(v);
    return F.run(D.items, { q: v.q, fields: ['_hay'], where: { countries: v.country, category: v.category, type: v.type, docLang: v.doclang, sdk: v.sdk, _pslug: v.provider }, sort: sortMap[v.sort] });
  }

  function row(r, i) {
    var c = D.cat(r.category), nq = (r.questions || []).length, p = 'discover-row-' + r.id;
    return '<li class="row" data-testid="' + p + '">' +
      '<div class="row-lead"><span class="badge" data-testid="' + p + '-type"><span class="icon i-' + TYPE_ICON[r.type] + '" aria-hidden="true" data-testid="' + p + '-type-icon"></span>' + esc(t('discover.type.' + r.type)) + '</span></div>' +
      '<div class="row-main"><h3 class="row-title"><a href="resource.html?id=' + r.id + '" data-testid="' + p + '-link">' + esc(r.name) + '</a> <span class="mono text-muted text-sm" data-testid="' + p + '-version">' + esc(r.version) + '</span></h3>' +
      '<p class="text-sm text-muted" data-testid="' + p + '-provider">' + providerName(r) + ' ' + langBadge(r.summaryLang, p + '-summary-lang') + '</p>' +
      '<p class="row-excerpt" lang="' + r.summaryLang + '" data-user-content data-testid="' + p + '-summary">' + esc(r.summary) + '</p>' +
      '<div class="meta discover-meta">' +
        '<span class="meta-item" data-testid="' + p + '-category">' + esc(lbl(c)) + '</span>' +
        '<span class="meta-item" data-testid="' + p + '-countries"><span class="visually-hidden" data-testid="' + p + '-countries-label">' + esc(t('discover.countries.label')) + '</span>' + r.countries.map(function (x, j) { return cc(x, p + '-country-' + x.toLowerCase()); }).join(' ') + '</span>' +
        '<span class="meta-item" data-testid="' + p + '-doclangs"><span class="text-xs" data-testid="' + p + '-doclangs-label">' + esc(t('discover.docs')) + '</span> ' + r.docLang.map(function (l, j) { return langBadge(l, p + '-doclang-' + l, ''); }).join(' ') + '</span>' +
      '</div></div>' +
      '<div class="row-aside">' + statusBadge(r, p + '-status') +
        '<span data-testid="' + p + '-verified"><span class="visually-hidden" data-testid="' + p + '-verified-label">' + esc(t('discover.verifiedOn')) + ' </span><time datetime="' + r.verifiedAt + '" data-testid="' + p + '-verified-date">' + M.fmtDate(r.verifiedAt) + '</time></span>' +
        '<span data-testid="' + p + '-sources"><span class="icon i-book" aria-hidden="true" data-testid="' + p + '-sources-icon"></span> ' + esc(t('discover.sources', { n: r.sources.length })) + '</span>' +
        (nq ? '<span data-testid="' + p + '-questions"><span class="icon i-help" aria-hidden="true" data-testid="' + p + '-questions-icon"></span> ' + esc(t('discover.questions', { n: nq })) + '</span>' : '') +
      '</div></li>';
  }

  function render(items, box) {
    var v = vals();
    box.innerHTML = '<div class="results-meta"><span role="status" data-testid="discover-count">' + esc(t(active(v) ? 'discover.count.filtered' : 'discover.count', { n: items.length })) + '</span></div>' +
      '<ul class="list list-reset discover-rows" data-testid="discover-results">' + items.map(row).join('') + '</ul>';
  }

  function applyParams() {
    var map = { q: 'discover-q', country: 'discover-country', category: 'discover-category', type: 'discover-type', doclang: 'discover-doclang', sdk: 'discover-sdk', provider: 'discover-provider', sort: 'discover-sort' };
    Object.keys(map).forEach(function (k) { var x = P.get(k); if (x != null) $('#' + map[k]).value = x; });
  }

  fillSelects(); applyParams();
  ctl = A.Sim.load($('#discover-list'), { load: load, render: render, delay: 600, skeleton: 4, emptyKey: 'discover.empty' });
  form.addEventListener('input', function () { ctl.refresh(); });
  form.addEventListener('change', function () { ctl.refresh(); });
  form.addEventListener('reset', function () { setTimeout(function () { $('#discover-sort').value = 'verified'; ctl.refresh(); }, 0); });
  form.addEventListener('submit', function (e) { e.preventDefault(); ctl.refresh(); });
  document.addEventListener('i18n:change', function () { fillSelects(); document.title = t('discover.title'); });
  document.title = t('discover.title');
})();
