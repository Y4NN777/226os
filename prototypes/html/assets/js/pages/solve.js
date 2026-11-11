/* Solve list page: search, filters, sort, pagination, simulated async states, and the "Ask a question" form (#ask).
   Dictionary: assets/i18n/solve.js. Data: MOCK.questions + MOCK.solve.extraQuestions. Nothing is persisted. */
(function () {
  'use strict';
  var M = window.MOCK, S = M.solve, esc = App.esc, $ = App.$, $$ = App.$$, t = function (k, p) { return I18N.t(k, p); };
  var PER = 6, all = S.all();
  function pl(base, n) { return t(base + (n === 1 ? '.one' : '.other'), { n: n }); }
  function langName(l) { return t('solve.lang.' + l); }

  /* ---------- state ---------- */
  var st = { q: '', status: '', tags: [], country: '', lang: '', sort: 'recent', page: 1 };
  var form = $('#solve-form');

  /* ---------- derived items ---------- */
  var prepared = all.map(function (q) {
    var p = M.person(q.author) || { name: q.author };
    return Object.assign({}, q, { authorName: p.name, status: [q.answers === 0 ? 'unanswered' : 'answered'].concat(q.accepted ? ['accepted'] : []) });
  });
  function filtered() {
    var out = App.Filter.run(prepared, { q: st.q, fields: ['title', 'excerpt', 'tags', 'env', 'authorName'], where: { country: st.country, lang: st.lang, status: st.status } });
    if (st.tags.length) out = out.filter(function (q) { return st.tags.every(function (tg) { return q.tags.indexOf(tg) > -1; }); });
    var byDate = function (a, b) { return a.date < b.date ? 1 : a.date > b.date ? -1 : 0; };
    out.sort(st.sort === 'votes' ? function (a, b) { return b.votes - a.votes || byDate(a, b); }
      : st.sort === 'unanswered' ? function (a, b) { return a.answers - b.answers || byDate(a, b); } : byDate);
    return out;
  }

  /* ---------- filter controls ---------- */
  var tagCounts = {}; all.forEach(function (q) { q.tags.forEach(function (x) { tagCounts[x] = (tagCounts[x] || 0) + 1; }); });
  var topTags = Object.keys(tagCounts).sort(function (a, b) { return tagCounts[b] - tagCounts[a] || (a < b ? -1 : 1); }).slice(0, 10);
  $('#solve-tag-filters').innerHTML = topTags.map(function (x) {
    return '<label class="choice" data-testid="solve-tag-filter-' + x + '-choice"><input type="checkbox" name="tag" value="' + esc(x) + '" data-testid="solve-tag-filter-' + x + '-input"><span class="mono" data-testid="solve-tag-filter-' + x + '-label">' + esc(x) + '</span></label>';
  }).join('');
  function fillCountries() {
    var sel = $('#solve-country'), cur = sel.value, used = {};
    all.forEach(function (q) { used[q.country] = 1; });
    sel.innerHTML = '<option value="" data-testid="solve-country-all">' + esc(t('common.all')) + '</option>' +
      M.countries.filter(function (c) { return used[c.code]; }).map(function (c) { return '<option value="' + c.code + '" data-testid="solve-country-' + c.code.toLowerCase() + '">' + esc(c.code + ' — ' + M.label(c)) + '</option>'; }).join('');
    sel.value = cur;
  }
  fillCountries();

  function readForm() {
    st.q = form.elements.q.value;
    st.status = (form.querySelector('input[name="status"]:checked') || {}).value || '';
    st.tags = [].slice.call(form.querySelectorAll('input[name="tag"]:checked')).map(function (i) { return i.value; });
    st.country = form.elements.country.value; st.lang = form.elements.lang.value;
  }
  function activeCount() { return (st.status ? 1 : 0) + st.tags.length + (st.country ? 1 : 0) + (st.lang ? 1 : 0); }
  function syncChrome() {
    var n = activeCount(), b = $('#solve-filter-count');
    b.hidden = !n; b.textContent = '';
    $('#solve-reset').disabled = !(n || st.q);
  }

  /* ---------- results ---------- */
  var sim = App.Sim.load($('#solve-list'), {
    delay: 600, skeleton: 4, emptyKey: 'solve.list.empty',
    load: function () { var r = filtered(); $('#solve-count').textContent = pl('solve.count', r.length); return r; },
    render: renderList
  });
  function cc(code, id) { return '<span class="country" data-testid="solve-country-code-' + id + '"><span class="country-code" data-testid="solve-country-code-' + id + '-code">' + code + '</span><span class="visually-hidden" data-testid="solve-country-code-' + id + '-name">' + esc(M.countryName(code)) + '</span></span>'; }
  function stats(q, n, kind) {
    var id = kind + '-' + n;
    return '<span class="solve-stat" data-testid="solve-votes-' + id + '"><strong data-testid="solve-votes-' + id + '-count">' + q.votes + '</strong> <span class="solve-stat-label" data-testid="solve-votes-' + id + '-label">' + esc(t('solve.row.votes.' + (q.votes === 1 ? 'one' : 'other'), { n: '' }).trim()) + '</span></span>' +
      '<span class="solve-stat' + (q.accepted ? ' solve-stat-accepted' : q.answers === 0 ? ' solve-stat-none' : '') + '" data-testid="solve-answers-' + id + '">' +
      (q.accepted ? '<span class="icon i-check" aria-hidden="true" data-testid="solve-answers-' + id + '-icon"></span>' : '') + '<strong data-testid="solve-answers-' + id + '-count">' + q.answers + '</strong> <span class="solve-stat-label" data-testid="solve-answers-' + id + '-label">' + esc(t('solve.row.answers.' + (q.answers === 1 ? 'one' : 'other'), { n: '' }).trim()) + '</span></span>';
  }
  function row(q, i) {
    var n = q.id, p = M.person(q.author) || { name: q.author };
    return '<li class="row" data-testid="solve-row-' + n + '">' +
      '<div class="row-lead"><div class="solve-stats" data-testid="solve-stats-' + n + '">' + stats(q, n, 'lead') + '</div></div>' +
      '<div class="row-main">' +
      '<h3 class="row-title"><a href="question.html?id=' + q.id + '" lang="' + q.lang + '" data-user-content data-testid="solve-link-' + n + '">' + esc(q.title) + '</a></h3>' +
      '<p class="row-excerpt" lang="' + q.lang + '" data-user-content data-testid="solve-excerpt-' + n + '">' + esc(q.excerpt) + '</p>' +
      '<ul class="tag-list list-reset" data-testid="solve-tags-' + n + '">' + q.tags.map(function (x) { return '<li><span class="tag" data-testid="solve-tag-' + n + '-' + x + '">' + esc(x) + '</span></li>'; }).join('') + '</ul>' +
      '<div class="meta">' +
      '<span class="meta-item solve-inline-stats" data-testid="solve-inline-stats-' + n + '">' + stats(q, n, 'inline') + '</span>' +
      '<span class="meta-item" data-testid="solve-by-' + n + '"><span class="person">' + App.avatar(p, { size: 'sm', testid: 'solve-avatar-' + n }) + '<a class="row-above" href="profile.html?id=' + q.author + '" data-testid="solve-author-' + n + '">' + esc(p.name) + '</a></span></span>' +
      '<span class="meta-item" data-testid="solve-country-' + n + '">' + cc(q.country, n) + '</span>' +
      '<time class="meta-item" datetime="' + q.date + '" data-testid="solve-date-' + n + '">' + M.fmtDate(q.date) + '</time>' +
      '<span class="meta-item mono" data-testid="solve-env-' + n + '"><span class="visually-hidden" data-testid="solve-env-label-' + n + '">' + esc(t('solve.row.env')) + ' </span>' + esc(q.env) + '</span></div></div>' +
      '<div class="row-aside">' +
      '<span class="badge" data-testid="solve-lang-badge-' + n + '"><span aria-hidden="true" data-testid="solve-lang-badge-' + n + '-code">' + q.lang.toUpperCase() + '</span><span class="visually-hidden" data-testid="solve-lang-badge-' + n + '-name">' + esc(t('solve.row.lang', { l: langName(q.lang) })) + '</span></span>' +
      '</div></li>';
  }
  function renderList(items, box) {
    var pages = Math.max(1, Math.ceil(items.length / PER)); if (st.page > pages) st.page = pages;
    var slice = items.slice((st.page - 1) * PER, st.page * PER);
    $('#solve-count').textContent = pl('solve.count', items.length);
    var pg = '';
    if (pages > 1) {
      pg = '<nav class="pagination" aria-label="' + esc(t('common.pagination')) + '" data-testid="solve-pagination">' +
        '<button type="button" class="btn" data-page="' + (st.page - 1) + '"' + (st.page === 1 ? ' disabled' : '') + ' data-testid="solve-page-prev"><span class="icon i-chevron-left" aria-hidden="true" data-testid="solve-page-prev-icon"></span><span data-testid="solve-page-prev-label">' + esc(t('common.prev')) + '</span></button>';
      for (var i = 1; i <= pages; i++) pg += '<button type="button" class="btn" data-page="' + i + '" aria-label="' + esc(t('solve.page.label', { n: i })) + '"' + (i === st.page ? ' aria-current="page"' : '') + ' data-testid="solve-page-' + i + '">' + i + '</button>';
      pg += '<button type="button" class="btn" data-page="' + (st.page + 1) + '"' + (st.page === pages ? ' disabled' : '') + ' data-testid="solve-page-next"><span data-testid="solve-page-next-label">' + esc(t('common.next')) + '</span><span class="icon i-chevron-right" aria-hidden="true" data-testid="solve-page-next-icon"></span></button></nav>';
    }
    box.innerHTML = '<ul class="list list-reset" data-testid="solve-rows">' + slice.map(row).join('') + '</ul>' + pg;
  }
  $('#solve-list').addEventListener('click', function (e) {
    var b = e.target.closest('[data-page]'); if (!b || b.disabled) return;
    st.page = +b.getAttribute('data-page'); sim.refresh();
    var nb = $('[data-testid="solve-page-' + st.page + '"]'); if (nb) nb.focus();
    $('#solve-count').scrollIntoView({ block: 'start' });
  });

  /* ---------- wiring ---------- */
  function change() { readForm(); st.page = 1; syncChrome(); sim.refresh(); }
  form.addEventListener('submit', function (e) { e.preventDefault(); change(); });
  form.addEventListener('input', function (e) { if (e.target.id !== 'ask-form') change(); });
  form.addEventListener('change', function () { change(); });
  $('#solve-reset').addEventListener('click', function () {
    form.reset(); st.sort = 'recent'; $$('[data-sort]').forEach(function (x) { x.setAttribute('aria-checked', x.getAttribute('data-sort') === 'recent'); }); change(); form.elements.q.focus();
  });
  $$('[data-sort]').forEach(function (b) { b.addEventListener('click', function () {
    st.sort = b.getAttribute('data-sort'); $$('[data-sort]').forEach(function (x) { x.setAttribute('aria-checked', x === b); }); change();
  }); });
  var tog = $('#solve-filter-toggle'), panel = $('#solve-filters');
  tog.addEventListener('click', function () { var open = tog.getAttribute('aria-expanded') !== 'true'; tog.setAttribute('aria-expanded', open); panel.classList.toggle('solve-filters-collapsed', !open); });
  var initQ = App.params.get('q'); if (initQ) { form.elements.q.value = initQ; readForm(); sim.refresh(); }
  syncChrome();
  document.addEventListener('i18n:change', function () { fillCountries(); syncChrome(); readForm(); setErrors(); sim.refresh(); updateSuccess(); });

  /* ---------- Ask form ---------- */
  var ask = $('#ask-form'), errs = {}, ORDER = ['title', 'body', 'tags'], lastTitle = '';
  ask.elements.lang.value = I18N.lang;
  function validate() {
    var e = {}, title = ask.elements.title.value.trim(), body = ask.elements.body.value.trim(), tags = ask.elements.tags.value.split(',').map(function (x) { return x.trim().toLowerCase(); }).filter(Boolean);
    if (!title) e.title = 'solve.ask.err.title.req'; else if (title.length < 15) e.title = 'solve.ask.err.title.short';
    if (!body) e.body = 'solve.ask.err.body.req'; else if (body.length < 30) e.body = 'solve.ask.err.body.short'; else if (S.unbalancedFence(body)) e.body = 'solve.ask.err.body.fence';
    if (!tags.length) e.tags = 'solve.ask.err.tags.req'; else if (tags.length > 5) e.tags = 'solve.ask.err.tags.many'; else if (tags.some(function (x) { return !/^[a-z0-9.+#-]{1,24}$/.test(x); })) e.tags = 'solve.ask.err.tags.bad';
    return e;
  }
  function setErrors() {
    var list = [], sm = $('#ask-summary');
    ORDER.forEach(function (f) {
      var field = $('#ask-' + f + '-field'), inp = $('#ask-' + f), msg = $('#ask-' + f + '-error'), hint = $('#ask-' + f + '-hint'), key = errs[f];
      field.classList.toggle('has-error', !!key); msg.hidden = !key;
      if (key) { inp.setAttribute('aria-invalid', 'true'); inp.setAttribute('aria-describedby', hint.id + ' ' + msg.id);
        msg.innerHTML = '<span class="icon i-alert-circle" aria-hidden="true" data-testid="ask-' + f + '-error-icon"></span><span data-testid="ask-' + f + '-error-text">' + esc(t(key)) + '</span>';
        list.push('<li><a href="#ask-' + f + '" data-focus="ask-' + f + '" data-testid="ask-summary-link-' + f + '">' + esc(t(key)) + '</a></li>'); }
      else { inp.removeAttribute('aria-invalid'); inp.setAttribute('aria-describedby', hint.id); }
    });
    sm.hidden = !list.length; $('#ask-summary-list').innerHTML = list.join('');
  }
  function updateSuccess() { if (lastTitle) $('#ask-success-body').textContent = t('solve.ask.success.body', { t: lastTitle }); }
  $('#ask-summary').addEventListener('click', function (e) { var a = e.target.closest('[data-focus]'); if (a) { e.preventDefault(); var f = document.getElementById(a.getAttribute('data-focus')); f.focus(); } });
  ask.addEventListener('submit', function (e) {
    e.preventDefault(); errs = validate(); setErrors();
    var first = ORDER.filter(function (f) { return errs[f]; })[0];
    if (first) { $('#ask-summary').focus(); return; }
    var btn = $('#ask-submit'); btn.setAttribute('aria-busy', 'true'); btn.disabled = true;
    setTimeout(function () {   /* simulated network round trip */
      btn.removeAttribute('aria-busy'); btn.disabled = false;
      lastTitle = ask.elements.title.value.trim(); updateSuccess(); $('#ask-success').hidden = false;
      App.Toast.show({ kind: 'success', msg: t('solve.ask.toast') });
      ask.reset(); ask.elements.lang.value = I18N.lang; $('#ask-preview').innerHTML = ''; errs = {}; setErrors();
      $('#ask-success').scrollIntoView({ block: 'nearest' });
    }, 500);
  });
  ask.addEventListener('reset', function () { errs = {}; setTimeout(function () { setErrors(); $('#ask-preview').innerHTML = ''; }, 0); });
  ask.addEventListener('input', function (e) { if (errs[e.target.name]) { delete errs[e.target.name]; setErrors(); } });

  /* Markdown preview (same renderer as question bodies) */
  function preview() {
    var v = ask.elements.body.value, box = $('#ask-preview');
    box.setAttribute('lang', ask.elements.lang.value); box.setAttribute('data-user-content', '');
    box.innerHTML = v.trim() ? S.renderMd(v, { id: 'ask-preview', copyText: t('solve.code.copyshort'), copyLabel: function (l) { return t('solve.code.copy', { l: l }); }, codeLabel: function (l) { return t('solve.code.label', { l: l }); } })
      : '<p class="text-muted" data-testid="ask-preview-empty">' + esc(t('solve.preview.empty')) + '</p>';
  }
  var pt = $('#ask-tab-preview'); pt.addEventListener('click', preview);   /* arrow keys also fire click (App.Tabs) */
  ask.elements.body.addEventListener('input', preview);

  /* Deep link #ask: move focus to the form heading so keyboard/screen-reader users land in the right place */
  function toAsk() { if (location.hash === '#ask') { var h = $('#ask-h'); setTimeout(function () { h.focus({ preventScroll: true }); h.scrollIntoView({ block: 'start' }); }, 0); } }
  window.addEventListener('hashchange', toAsk); toAsk();
 
})();
