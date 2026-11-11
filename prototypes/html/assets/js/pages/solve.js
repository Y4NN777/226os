/* Solve list page + ask flow (#ask). Dictionary: assets/i18n/solve.js. Data: MOCK.questions + MOCK.solve.
   Views: list (default) and ask (#ask). Ask = title first, live "Questions similaires", details, tags, publish, then "Ce qui se passe maintenant".
   Publishing is the gated action: signed out, the draft is kept in sessionStorage and App.Session.requireAuth sends the user to sign-in, then back to #ask. Nothing is persisted otherwise. */
(function () {
  'use strict';
  var M = window.MOCK, S = M.solve, esc = App.esc, $ = App.$, $$ = App.$$, t = function (k, p) { return I18N.t(k, p); };
  var PER = 8, all = S.all();
  function pl(base, n) { return t(base + (n === 1 ? '.one' : '.other'), { n: n }); }
  function langName(l) { return t('solve.lang.' + l); }
  /* resume URL: current page without the QA override (?session=) so signing in is not undone, plus an anchor to resume */
  function resumeUrl(hash) { var u = new URL(location.href); u.searchParams.delete('session'); u.hash = hash || ''; return u.href; }
  function pill(kind, label, tid) { return '<span class="state-pill state-pill-' + kind + '" data-testid="' + tid + '"><span class="icon i-' + (kind === 'accepted' ? 'check-circle' : 'circle-dot') + '" aria-hidden="true" data-testid="' + tid + '-icon"></span>' + esc(label) + '</span>'; }

  /* ---------- list state ---------- */
  var st = { q: '', status: '', tags: [], country: '', lang: '', sort: 'recent', page: 1 };
  var form = $('#solve-form');
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
  var deepTag = App.params.get('tag'); if (deepTag && topTags.indexOf(deepTag) < 0 && tagCounts[deepTag]) topTags.push(deepTag);
  $('#solve-tag-filters').innerHTML = topTags.map(function (x) {
    return '<label class="solve-tagchoice" data-testid="solve-tag-filter-' + x + '-choice"><input type="checkbox" name="tag" value="' + esc(x) + '"' + (x === deepTag ? ' checked' : '') + ' data-testid="solve-tag-filter-' + x + '-input"><span class="mono" data-testid="solve-tag-filter-' + x + '-label">' + esc(x) + '</span></label>';
  }).join('');
  function fillCountries() {
    var sel = $('#solve-country'), cur = sel.value, used = {};
    all.forEach(function (q) { used[q.country] = 1; });
    sel.innerHTML = '<option value="" data-testid="solve-country-all">' + esc(t('common.all')) + '</option>' +
      M.countries.filter(function (c) { return used[c.code]; }).map(function (c) { return '<option value="' + c.code + '" data-testid="solve-country-' + c.code.toLowerCase() + '">' + esc(c.code + ' — ' + M.label(c)) + '</option>'; }).join('');
    sel.value = cur;
  }
  fillCountries();
  var dstatus = App.params.get('status'); if (dstatus) { var rr = form.querySelector('input[name="status"][value="' + dstatus + '"]'); if (rr) rr.checked = true; }

  function readForm() {
    st.q = form.elements.q.value;
    st.status = (form.querySelector('input[name="status"]:checked') || {}).value || '';
    st.tags = [].slice.call(form.querySelectorAll('input[name="tag"]:checked')).map(function (i) { return i.value; });
    st.country = form.elements.country.value; st.lang = form.elements.lang.value;
  }
  function activeCount() { return st.tags.length + (st.country ? 1 : 0) + (st.lang ? 1 : 0); }   /* status is visible in the toolbar, so it is not counted on the filter button */
  function syncChrome() {
    var n = activeCount(), b = $('#solve-filter-count');
    b.hidden = !n; b.textContent = '';
    $('#solve-reset').disabled = !(n || st.q || st.status);
  }

  /* ---------- results ---------- */
  var simOpts = {
    delay: 500, skeleton: 4, emptyKey: 'solve.list.empty',
    empty: { icon: 'inbox', action: { labelKey: 'solve.list.empty.cta', icon: 'plus', href: '#ask', primary: true, testid: 'solve-empty-ask' } },
    load: function () { var r = filtered(); $('#solve-count').textContent = pl('solve.count', r.length); return r.length ? r : [{ __empty: true }]; },
    render: renderList
  };
  var sim = App.Sim.load($('#solve-list'), simOpts);
  function cc(code, id) { return '<span class="country" data-testid="solve-country-code-' + id + '"><span class="country-code" data-testid="solve-country-code-' + id + '-code">' + code + '</span><span class="visually-hidden" data-testid="solve-country-code-' + id + '-name">' + esc(M.countryName(code)) + '</span></span>'; }
  function row(q) {
    var n = q.id, p = M.person(q.author) || { name: q.author }, v = q.votes, a = q.answers;
    var state = q.accepted ? pill('accepted', t('solve.row.accepted'), 'solve-pill-' + n) : a === 0 ? pill('open', t('solve.row.unanswered'), 'solve-pill-' + n) : '';
    return '<li class="row" data-testid="solve-row-' + n + '">' +
      '<div class="row-lead"><div class="solve-counts" data-testid="solve-stats-' + n + '">' +
      '<span class="solve-stat" data-testid="solve-votes-' + n + '"><span class="icon i-arrow-up" aria-hidden="true" data-testid="solve-votes-' + n + '-icon"></span><strong class="tabular-nums" data-testid="solve-votes-' + n + '-count">' + v + '</strong><span class="visually-hidden" data-testid="solve-votes-' + n + '-label">' + esc(t('solve.row.votes.' + (v === 1 ? 'one' : 'other'))) + '</span></span>' +
      '<span class="solve-stat' + (q.accepted ? ' solve-stat-accepted' : a === 0 ? ' solve-stat-none' : '') + '" data-testid="solve-answers-' + n + '"><span class="icon i-' + (q.accepted ? 'check-circle' : 'message') + '" aria-hidden="true" data-testid="solve-answers-' + n + '-icon"></span><strong class="tabular-nums" data-testid="solve-answers-' + n + '-count">' + a + '</strong><span class="visually-hidden" data-testid="solve-answers-' + n + '-label">' + esc(t('solve.row.answers.' + (a === 1 ? 'one' : 'other'))) + '</span></span>' +
      '</div></div>' +
      '<div class="row-main">' +
      '<h3 class="row-title"><a href="question.html?id=' + q.id + '" lang="' + q.lang + '" data-user-content data-testid="solve-link-' + n + '">' + esc(q.title) + '</a></h3>' +
      '<p class="row-excerpt" lang="' + q.lang + '" data-user-content data-testid="solve-excerpt-' + n + '">' + esc(q.excerpt) + '</p>' +
      '<div class="solve-row-meta">' + state +
      '<ul class="solve-tags list-reset" data-testid="solve-tags-' + n + '">' + q.tags.map(function (x) { return '<li><span class="mono" data-testid="solve-tag-' + n + '-' + x + '">' + esc(x) + '</span></li>'; }).join('') + '</ul></div>' +
      '<div class="meta">' +
      '<span class="meta-item" data-testid="solve-by-' + n + '"><span class="person">' + App.avatar(p, { size: 'sm', testid: 'solve-avatar-' + n }) + '<a class="row-above" href="profile.html?id=' + q.author + '" data-testid="solve-author-' + n + '">' + esc(p.name) + '</a></span></span>' +
      '<span class="meta-item" data-testid="solve-country-' + n + '">' + cc(q.country, n) + '</span>' +
      '<time class="meta-item" datetime="' + q.date + '" data-testid="solve-date-' + n + '">' + M.fmtDate(q.date) + '</time>' +
      '<span class="meta-item mono" data-testid="solve-env-' + n + '"><span class="visually-hidden" data-testid="solve-env-label-' + n + '">' + esc(t('solve.row.env')) + ' </span>' + esc(q.env) + '</span></div></div>' +
      '<div class="row-aside">' +
      '<span class="solve-lang" data-testid="solve-lang-badge-' + n + '"><span aria-hidden="true" data-testid="solve-lang-badge-' + n + '-code">' + q.lang.toUpperCase() + '</span><span class="visually-hidden" data-testid="solve-lang-badge-' + n + '-name">' + esc(t('solve.row.lang', { l: langName(q.lang) })) + '</span></span>' +
      '</div></li>';
  }
  function noResult() {
    var q = st.q.trim();
    return '<div class="state solve-noresult" data-testid="solve-noresult"><span class="icon i-search" aria-hidden="true" data-testid="solve-noresult-icon"></span>' +
      '<h3 data-testid="solve-noresult-title">' + esc(q ? t('solve.noresult.title.q', { q: q }) : t('solve.noresult.title.f')) + '</h3>' +
      '<p data-testid="solve-noresult-body">' + esc(t('solve.noresult.body')) + '</p>' +
      '<div class="state-actions" data-testid="solve-noresult-actions"><button type="button" class="btn btn-sm" id="solve-noresult-clear" data-testid="solve-noresult-clear"><span class="icon i-refresh" aria-hidden="true" data-testid="solve-noresult-clear-icon"></span>' + esc(t('solve.noresult.clear')) + '</button>' +
      '<a class="btn btn-sm btn-primary" href="#ask" id="solve-noresult-ask" data-testid="solve-noresult-ask"><span class="icon i-plus" aria-hidden="true" data-testid="solve-noresult-ask-icon"></span>' + esc(t('solve.noresult.ask')) + '</a></div></div>';
  }
  function renderList(items, box) {
    if (items[0] && items[0].__empty) { $('#solve-count').textContent = pl('solve.count', 0); box.innerHTML = noResult(); return; }
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
    App.View.apply('solve');
  }
  $('#solve-list').addEventListener('click', function (e) {
    var b = e.target.closest('[data-page]');
    if (b && !b.disabled) {
      st.page = +b.getAttribute('data-page'); sim.refresh();
      var nb = $('[data-testid="solve-page-' + st.page + '"]'); if (nb) nb.focus();
      $('#solve-count').scrollIntoView({ block: 'start' }); return;
    }
    if (e.target.closest('#solve-noresult-clear')) { resetAll(); return; }
    if (e.target.closest('#solve-noresult-ask')) { saveDraft({ title: st.q.trim() }); }
  });

  /* ---------- wiring ---------- */
  function change() { readForm(); st.page = 1; syncChrome(); sim.refresh(); }
  function resetAll() {
    form.reset(); st.sort = 'recent'; $$('[data-sort]').forEach(function (x) { x.setAttribute('aria-checked', x.getAttribute('data-sort') === 'recent'); }); change(); form.elements.q.focus();
  }
  form.addEventListener('submit', function (e) { e.preventDefault(); change(); });
  form.addEventListener('input', function () { change(); });
  form.addEventListener('change', function () { change(); });
  $('#solve-reset').addEventListener('click', resetAll);
  $$('[data-sort]').forEach(function (b) { b.addEventListener('click', function () {
    st.sort = b.getAttribute('data-sort'); $$('[data-sort]').forEach(function (x) { x.setAttribute('aria-checked', x === b); }); change();
  }); });
  var tog = $('#solve-filter-toggle'), panel = $('#solve-filters');
  tog.addEventListener('click', function () { var open = tog.getAttribute('aria-expanded') !== 'true'; tog.setAttribute('aria-expanded', open); panel.hidden = !open; });
  var initQ = App.params.get('q'); if (initQ) form.elements.q.value = initQ;
  readForm(); if (deepTag) { tog.setAttribute('aria-expanded', 'true'); panel.hidden = false; }
  syncChrome(); if (initQ || dstatus || deepTag) sim.refresh();

  /* the page action is primary only for signed-out visitors: signed in, the header (+) is the dominant action */
  function syncCta() { $('#solve-cta').classList.toggle('btn-primary', !App.Session.signedIn); $('#ask-signedout-note').hidden = App.Session.signedIn; }
  syncCta(); document.addEventListener('session:change', syncCta);

  /* ================= Ask flow ================= */
  var ask = $('#ask-form'), errs = {}, ORDER = ['title', 'body', 'tags'], lastTitle = '', lastTags = [], step = 1, DK = '226os.askDraft';
  ask.elements.lang.value = I18N.lang;
  function tagList() { return ask.elements.tags.value.split(',').map(function (x) { return x.trim().toLowerCase(); }).filter(Boolean); }
  function saveDraft(patch) {
    try { var d = JSON.parse(sessionStorage.getItem(DK) || '{}'); sessionStorage.setItem(DK, JSON.stringify(Object.assign(d, patch))); } catch (e) {}
  }
  function readDraft() { try { return JSON.parse(sessionStorage.getItem(DK) || 'null'); } catch (e) { return null; } }
  function clearDraft() { try { sessionStorage.removeItem(DK); } catch (e) {} }
  function snapshot() { saveDraft({ title: ask.elements.title.value, body: ask.elements.body.value, tags: ask.elements.tags.value, lang: ask.elements.lang.value }); }

  function validate() {
    var e = {}, title = ask.elements.title.value.trim(), body = ask.elements.body.value.trim(), tags = tagList();
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
  $('#ask-summary').addEventListener('click', function (e) { var a = e.target.closest('[data-focus]'); if (a) { e.preventDefault(); document.getElementById(a.getAttribute('data-focus')).focus(); } });

  /* steps: 1 = title (+ similar), 2 = details */
  function setStep(n, focus) {
    step = n; $('#ask-details').hidden = n < 2; $('#ask-continue-row').hidden = n >= 2;
    $('#ask-step-1').toggleAttribute('aria-current', n === 1); $('#ask-step-2').toggleAttribute('aria-current', n === 2);
    if (n === 1) { $('#ask-step-1').setAttribute('aria-current', 'step'); $('#ask-step-2').removeAttribute('aria-current'); } else { $('#ask-step-2').setAttribute('aria-current', 'step'); $('#ask-step-1').removeAttribute('aria-current'); }
    $('#ask-step-1').classList.toggle('is-done', n > 1);
    if (n === 2 && focus) ask.elements.body.focus();
  }
  function toDetails() {
    var v = ask.elements.title.value.trim();
    if (!v || v.length < 15) { errs = {}; errs.title = !v ? 'solve.ask.err.title.req' : 'solve.ask.err.title.short'; setErrors(); ask.elements.title.focus(); return; }
    delete errs.title; setErrors(); setStep(2, true);
  }
  $('#ask-continue').addEventListener('click', toDetails);

  /* similar questions (mock search over the same dataset) */
  var simTimer = null;
  function renderSimilar() {
    var box = $('#ask-sim-body'), v = ask.elements.title.value.trim(), tg = tagList(), items = v.length >= 3 ? S.similar(v, tg, null, 3) : null;
    if (!items) { box.setAttribute('data-state', 'idle'); box.innerHTML = '<p class="solve-sim-msg" data-testid="ask-sim-idle"><span class="icon i-search" aria-hidden="true" data-testid="ask-sim-idle-icon"></span><span data-testid="ask-sim-idle-text">' + esc(t('solve.similar.idle')) + '</span></p>'; return; }
    if (!items.length) { box.setAttribute('data-state', 'none'); box.innerHTML = '<p class="solve-sim-msg is-new" data-testid="ask-sim-none"><span class="icon i-sparkles" aria-hidden="true" data-testid="ask-sim-none-icon"></span><span data-testid="ask-sim-none-text">' + esc(t('solve.similar.none')) + '</span></p>'; return; }
    box.setAttribute('data-state', 'results');
    box.innerHTML = '<p class="solve-sim-lead" data-testid="ask-sim-lead">' + esc(t('solve.similar.lead')) + '</p><ul class="solve-sim-list list-reset" data-testid="ask-sim-list">' + items.map(function (q) {
      var n = q.id, a = q.answers;
      return '<li data-testid="ask-sim-item-' + n + '"><a href="question.html?id=' + n + '" target="_blank" rel="noopener" lang="' + q.lang + '" data-user-content data-testid="ask-sim-link-' + n + '">' + esc(q.title) + '<span class="visually-hidden" data-testid="ask-sim-link-' + n + '-newtab"> ' + esc(t('solve.similar.newtab')) + '</span></a>' +
        '<span class="solve-sim-meta" data-testid="ask-sim-meta-' + n + '">' + (q.accepted ? pill('accepted', t('solve.row.accepted'), 'ask-sim-pill-' + n) : '<span class="tabular-nums" data-testid="ask-sim-count-' + n + '">' + esc(pl('solve.q.answers.n', a)) + '</span>') + '<span class="icon i-external" aria-hidden="true" data-testid="ask-sim-ext-' + n + '"></span></span></li>';
    }).join('') + '</ul>';
  }
  function scheduleSimilar() { clearTimeout(simTimer); simTimer = setTimeout(renderSimilar, 120); }

  /* tag suggestions: top tags, one tap adds it */
  function renderTagSuggest() {
    var cur = tagList();
    $('#ask-tag-suggest').innerHTML = topTags.slice(0, 6).filter(function (x) { return cur.indexOf(x) < 0; }).map(function (x) {
      return '<button type="button" class="solve-suggest mono" data-add-tag="' + esc(x) + '" aria-label="' + esc(t('solve.ask.tags.add', { t: x })) + '" data-testid="ask-tag-suggest-' + x + '"><span class="icon i-plus" aria-hidden="true" data-testid="ask-tag-suggest-' + x + '-icon"></span>' + esc(x) + '</button>';
    }).join('');
  }
  $('#ask-tag-suggest').addEventListener('click', function (e) {
    var b = e.target.closest('[data-add-tag]'); if (!b) return;
    var cur = tagList(); if (cur.length >= 5) return; cur.push(b.getAttribute('data-add-tag')); ask.elements.tags.value = cur.join(', ');
    delete errs.tags; setErrors(); renderTagSuggest(); scheduleSimilar(); snapshot(); ask.elements.tags.focus();
  });

  ask.addEventListener('input', function (e) {
    if (errs[e.target.name]) { delete errs[e.target.name]; setErrors(); }
    if (e.target.name === 'title' || e.target.name === 'tags') scheduleSimilar();
    if (e.target.name === 'tags') renderTagSuggest();
    snapshot();
  });
  ask.elements.lang.addEventListener('change', snapshot);
  ask.elements.title.addEventListener('keydown', function (e) { if (e.key === 'Enter' && step === 1) { e.preventDefault(); toDetails(); } });

  ask.addEventListener('submit', function (e) {
    e.preventDefault(); if (step === 1) { toDetails(); return; }
    errs = validate(); setErrors();
    var first = ORDER.filter(function (f) { return errs[f]; })[0];
    if (first) { $('#ask-summary').focus(); return; }
    snapshot();
    if (!App.Session.requireAuth(resumeUrl('#ask'))) return;   /* the gated action: signed out, go to sign-in and come back to the draft */
    var btn = $('#ask-submit'); btn.setAttribute('aria-busy', 'true'); btn.disabled = true;
    setTimeout(function () {   /* simulated network round trip */
      btn.removeAttribute('aria-busy'); btn.disabled = false;
      lastTitle = ask.elements.title.value.trim(); lastTags = tagList();
      App.Toast.show({ kind: 'success', msg: t('solve.ask.toast') });
      showDone(); clearDraft(); ask.reset(); ask.elements.lang.value = I18N.lang; $('#ask-preview').innerHTML = ''; errs = {}; setErrors(); setStep(1); renderSimilar(); renderTagSuggest();
    }, 500);
  });
  ask.addEventListener('reset', function () { errs = {}; setTimeout(function () { setErrors(); $('#ask-preview').innerHTML = ''; }, 0); });

  /* Markdown preview (same renderer as question bodies) */
  function preview() {
    var v = ask.elements.body.value, box = $('#ask-preview');
    box.setAttribute('lang', ask.elements.lang.value); box.setAttribute('data-user-content', '');
    box.innerHTML = v.trim() ? S.renderMd(v, { id: 'ask-preview', copyText: t('solve.code.copyshort'), copyLabel: function (l) { return t('solve.code.copy', { l: l }); }, codeLabel: function (l) { return t('solve.code.label', { l: l }); } })
      : '<p class="text-muted" data-testid="ask-preview-empty">' + esc(t('solve.preview.empty')) + '</p>';
  }
  $('#ask-tab-preview').addEventListener('click', preview);   /* arrow keys also fire click (App.Tabs) */
  ask.elements.body.addEventListener('input', preview);

  /* success: what happens now + next actions */
  function showDone() {
    $('#ask-compose').hidden = true; $('#ask-done').hidden = false; renderDone();
    var h = $('#ask-done-h'); h.focus(); window.scrollTo(0, 0);
  }
  function renderDone() {
    if (!lastTitle) return;
    $('#ask-done-title').textContent = '\u00ab\u00a0' + lastTitle + '\u00a0\u00bb';
    $('#ask-timeline-2-text').textContent = t('solve.done.s2', { tags: lastTags.join(', ') });
    var r = S.resourceFor(lastTags), un = S.unansweredByTag(lastTags, null, 3), h = '';
    if (!un.length) un = S.all().filter(function (x) { return x.answers === 0; }).slice(0, 3).map(function (x) { return Object.assign({}, x, { tagMatch: x.tags[0] }); });
    if (r) h += '<div data-testid="ask-after-api"><h3 class="solve-mini-h" data-testid="ask-after-api-h">' + esc(t('solve.done.api.h')) + '</h3><a class="linked-item solve-linked" href="resource.html?id=' + r.id + '" data-testid="ask-after-api-link">' + App.thumb(r.name, { size: 'sm', icon: 'compass', testid: 'ask-after-api-thumb' }) +
      '<span class="solve-linked-text"><strong data-testid="ask-after-api-name">' + esc(r.name) + '</strong><span class="text-xs text-muted" data-testid="ask-after-api-sub">' + esc(r.provider + ' · ' + r.version) + '</span></span><span class="solve-linked-go" data-testid="ask-after-api-go"><span data-testid="ask-after-api-cta">' + esc(t('solve.done.api.cta')) + '</span><span class="icon i-arrow-right" aria-hidden="true" data-testid="ask-after-api-go-icon"></span></span></a></div>';
    h += '<div data-testid="ask-after-unans"><h3 class="solve-mini-h" data-testid="ask-after-unans-h">' + esc(t('solve.done.unans.h')) + '</h3>' +
      (un.length ? '<ul class="solve-mini-list list-reset" data-testid="ask-after-unans-list">' + un.map(function (q) {
        return '<li data-testid="ask-after-unans-item-' + q.id + '"><a href="question.html?id=' + q.id + '" lang="' + q.lang + '" data-user-content data-testid="ask-after-unans-link-' + q.id + '">' + esc(q.title) + '</a><span class="mono text-muted" data-testid="ask-after-unans-tag-' + q.id + '">' + esc(q.tagMatch) + '</span></li>';
      }).join('') + '</ul>' : '<p class="text-sm text-muted" data-testid="ask-after-unans-none">' + esc(t('solve.done.unans.none')) + '</p>') +
      '<p><a class="link-plain text-sm" href="solve.html?status=unanswered" data-testid="ask-after-unans-all">' + esc(t('solve.done.unans.all')) + '</a></p></div>';
    $('#ask-after-body').innerHTML = h;
  }
  $('#ask-again').addEventListener('click', function () { $('#ask-done').hidden = true; $('#ask-compose').hidden = false; ask.elements.title.focus(); });

  /* restore a draft (after sign-in, or a title carried over from the "no result" state) */
  function restore() {
    var d = readDraft(); if (!d) return;
    if (d.title != null) ask.elements.title.value = d.title;
    if (d.body != null) ask.elements.body.value = d.body;
    if (d.tags != null) ask.elements.tags.value = d.tags;
    if (d.lang) ask.elements.lang.value = d.lang;
    if (d.body || d.tags) setStep(2); else setStep(1);
    preview(); renderSimilar(); renderTagSuggest();
  }

  /* ---------- view routing: #ask opens the ask view, anything else the list ---------- */
  function route() {
    var isAsk = location.hash === '#ask';
    $('#solve-page-list').hidden = isAsk; $('#ask').hidden = !isAsk;
    document.title = t(isAsk ? 'solve.ask.h' : 'solve.title');
    if (isAsk) {
      if (!$('#ask-done').hidden && !lastTitle) { $('#ask-done').hidden = true; $('#ask-compose').hidden = false; }
      if ($('#ask-done').hidden) restore();
      var h = $(!$('#ask-done').hidden ? '#ask-done-h' : '#ask-h'); window.scrollTo(0, 0); h.focus({ preventScroll: true });
    }
  }
  window.addEventListener('hashchange', route);
  renderSimilar(); renderTagSuggest(); setStep(1); route();

  document.addEventListener('i18n:change', function () { fillCountries(); syncChrome(); readForm(); setErrors(); sim.refresh(); renderSimilar(); renderTagSuggest(); renderDone(); document.title = t(location.hash === '#ask' ? 'solve.ask.h' : 'solve.title'); });
})();
