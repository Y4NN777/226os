/* Solve question detail: markdown body + code, votes, answers (accepted first), accept-answer authorization by simulated viewer,
   related API/project (Problem-to-Project), answer form with preview. Dictionary: assets/i18n/solve.js. Nothing is persisted. */
(function () {
  'use strict';
  var M = window.MOCK, S = M.solve, esc = App.esc, $ = App.$, $$ = App.$$, t = function (k, p) { return I18N.t(k, p); };
  function pl(base, n) { return t(base + (n === 1 ? '.one' : '.other'), { n: n }); }
  var id = App.params.get('id') || 'q-webhook-sig', q = S.get(id);
  if (!q) { $('#question-missing').hidden = false; document.title = t('solve.q.missing.title') + ' — 226os'; document.addEventListener('i18n:change', function () { document.title = t('solve.q.missing.title') + ' — 226os'; }); return; }
  $('#question-root').hidden = false;

  /* ---------- state ---------- */
  var th = S.thread(q), answers = th.answers, votes = {}, newN = 0;
  var other = ['p-kofi', 'p-fatou'].filter(function (p) { return p !== q.author; })[0];
  var PERSONAS = { author: q.author, member: other, anon: null };
  var persona = App.params.get('as'); if (!(persona in PERSONAS)) persona = 'author';
  function viewer() { return PERSONAS[persona] ? M.person(PERSONAS[persona]) : null; }
  function pname(pid) { var p = M.person(pid); return p ? p.name : pid; }
  function isAccepted() { return th.full ? answers.some(function (a) { return a.accepted; }) : q.accepted; }
  function answerCount() { return th.full ? answers.length : q.answers; }
  function mdOpts(prefix) { return { id: prefix, headStart: 3, copyText: t('solve.code.copyshort'), copyLabel: function (l) { return t('solve.code.copy', { l: l }); }, codeLabel: function (l) { return t('solve.code.label', { l: l }); } }; }
  function langBadge(l, tid) { return '<span class="badge" data-testid="' + tid + '"><span aria-hidden="true" data-testid="' + tid + '-code">' + l.toUpperCase() + '</span><span class="visually-hidden" data-testid="' + tid + '-name">' + esc(t('solve.row.lang', { l: t('solve.lang.' + l) })) + '</span></span>'; }
  function cc(code, tid) { return '<span class="country" data-testid="' + tid + '"><span class="country-code" data-testid="' + tid + '-code">' + code + '</span><span class="visually-hidden" data-testid="' + tid + '-name">' + esc(M.countryName(code)) + '</span></span>'; }
  function personBox(pid, prefix) {
    var p = M.person(pid); if (!p) return '';
    return '<span class="person"><span class="avatar avatar-sm" aria-hidden="true" data-testid="' + prefix + '-avatar">' + esc(p.initials) + '</span>' +
      '<a href="profile.html?id=' + p.id + '" data-testid="' + prefix + '-author">' + esc(p.name) + '</a></span>' + cc(p.country, prefix + '-country');
  }

  /* ---------- vote control ---------- */
  function voteCtl(key, base, tid) {
    var mine = votes[key] || null, n = base + (mine === 'up' ? 1 : mine === 'down' ? -1 : 0), dis = persona === 'anon' ? ' disabled' : '';
    function b(dir, ic) {
      return '<button type="button" data-vote="' + dir + '" data-key="' + key + '" aria-pressed="' + (mine === dir) + '" aria-label="' + esc(t('solve.vote.' + dir)) + '"' + dis + ' data-testid="' + tid + '-' + dir + '"><span class="icon i-' + ic + '" aria-hidden="true" data-testid="' + tid + '-' + dir + '-icon"></span></button>';
    }
    return '<div class="vote" role="group" aria-label="' + esc(t('solve.vote.group')) + '" data-testid="' + tid + '">' + b('up', 'arrow-up') +
      '<span class="vote-count" data-testid="' + tid + '-count">' + n + '</span><span class="visually-hidden" data-testid="' + tid + '-label">' + esc(t('solve.vote.votes')) + (dis ? '. ' + esc(t('solve.vote.signin')) : '') + '</span>' + b('down', 'arrow-down') + '</div>';
  }

  /* ---------- static-ish regions ---------- */
  function renderHead() {
    var h1 = $('#question-h1'); h1.setAttribute('lang', q.lang); h1.setAttribute('data-user-content', ''); h1.textContent = q.title;
    $('#bc-current').textContent = q.title; $('#bc-current').setAttribute('lang', q.lang);
    document.title = q.title + ' — 226os';
    var acc = isAccepted();
    $('#question-head-meta').innerHTML =
      (acc ? '<span class="badge badge-accepted meta-item" data-testid="question-status"><span class="icon i-check-circle" aria-hidden="true" data-testid="question-status-icon"></span>' + esc(t('solve.q.status.accepted')) + '</span>'
           : '<span class="badge badge-warn meta-item" data-testid="question-status"><span class="icon i-help" aria-hidden="true" data-testid="question-status-icon"></span>' + esc(t('solve.q.status.open')) + '</span>') +
      '<span class="meta-item">' + langBadge(q.lang, 'question-lang-badge') + '</span>' +
      '<span class="meta-item" data-testid="question-answer-count">' + esc(pl('solve.q.answers.n', answerCount())) + '</span>' +
      '<span class="meta-item" data-testid="question-asked"><span class="text-muted" data-testid="question-asked-label">' + esc(t('solve.q.asked')) + '</span>&nbsp;<time datetime="' + q.date + '" data-testid="question-date">' + M.fmtDate(q.date) + '</time></span>';
  }
  function renderPost() {
    $('#question-post').innerHTML = '<article class="solve-post" aria-labelledby="question-h1" data-testid="question-article">' +
      '<div class="solve-post-vote">' + voteCtl('q', q.votes, 'question-vote') + '</div>' +
      '<div class="solve-post-main stack">' +
      '<div class="prose" lang="' + q.lang + '" data-user-content data-testid="question-body">' + S.renderMd(th.body, mdOpts('question')) + '</div>' +
      '<ul class="tag-list list-reset" data-testid="question-tags">' + q.tags.map(function (x) { return '<li><span class="tag" data-testid="question-tag-' + x + '">' + esc(x) + '</span></li>'; }).join('') + '</ul>' +
      '<div class="meta solve-post-foot">' + personBox(q.author, 'question') + '</div></div></article>';
  }
  function renderRelated() {
    var r = M.byId(M.resources, q.resource), p = M.byId(M.projects, q.project), rq = q.resource ? S.all().filter(function (x) { return x.id !== q.id && x.resource === q.resource; })[0] : null, items = [];
    function item(kind, icon, href, name, sub, n) {
      return '<a class="linked-item" href="' + href + '" data-testid="related-' + kind + '"><span class="kind" data-testid="related-' + kind + '-kind"><span class="icon i-' + icon + '" aria-hidden="true" data-testid="related-' + kind + '-icon"></span>' + esc(t('solve.related.' + (kind === 'resource' ? 'api' : kind))) + '</span>' +
        '<strong data-testid="related-' + kind + '-title"' + (kind === 'question' ? ' lang="' + rq.lang + '" data-user-content' : '') + '>' + esc(name) + '</strong><span class="text-xs text-muted" data-testid="related-' + kind + '-sub">' + esc(sub) + '</span></a>';
    }
    if (r) items.push(item('resource', 'compass', 'resource.html?id=' + r.id, r.name, r.provider + ' · ' + r.version));
    if (p) items.push(item('project', 'hammer', 'project.html?id=' + p.id, p.name, p.summary));
    if (rq) items.push(item('question', 'help', 'question.html?id=' + rq.id, rq.title, pl('solve.q.answers.n', rq.answers)));
    $('#question-related').innerHTML = items.length ? '<div class="linked-items" data-testid="related-items">' + items.join('') + '</div>' : '<p class="text-muted text-sm" data-testid="related-none">' + esc(t('solve.related.none')) + '</p>';
  }
  function renderContext() {
    var p = M.person(q.author);
    $('#question-context').innerHTML = [
      ['author', t('solve.context.author'), p ? esc(p.name) : ''], ['env', t('solve.context.env'), '<span class="mono" data-testid="context-env-text">' + esc(q.env) + '</span>'],
      ['country', t('solve.context.country'), cc(q.country, 'context-country-flag') + ' <span data-testid="context-country-name">' + esc(M.countryName(q.country)) + '</span>'], ['lang', t('solve.context.lang'), esc(t('solve.lang.' + q.lang))],
      ['asked', t('solve.context.asked'), '<time datetime="' + q.date + '" data-testid="context-asked-date">' + M.fmtDate(q.date) + '</time>']
    ].map(function (r) { return '<dt data-testid="context-' + r[0] + '-term">' + esc(r[1]) + '</dt><dd data-testid="context-' + r[0] + '-value">' + r[2] + '</dd>'; }).join('');
  }

  /* ---------- viewer simulation ---------- */
  function renderViewAs() {
    $$('[data-persona]').forEach(function (b) { b.setAttribute('aria-pressed', b.getAttribute('data-persona') === persona); });
    var v = viewer();
    $('#viewas-desc').textContent = t('solve.viewas.desc.' + persona, { name: v ? v.name : '' });
    $('#viewas-api').hidden = persona === 'author';
    $('#answer-anon').hidden = persona !== 'anon'; $('#answer-form').hidden = persona === 'anon';
  }
  $$('[data-persona]').forEach(function (b) {
    b.addEventListener('click', function () {
      persona = b.getAttribute('data-persona'); $('#viewas-result').innerHTML = '';
      try { var u = new URL(location.href); u.searchParams.set('as', persona); history.replaceState(null, '', u); } catch (e) { /* file:// may refuse; harmless */ }
      renderViewAs(); rerender();
    });
  });
  $('#viewas-api-btn').addEventListener('click', function () {
    var k = persona === 'anon' ? '401' : '403';
    $('#viewas-result').innerHTML = '<div class="callout callout-danger" data-testid="viewas-result-' + k + '"><span class="icon icon-lead i-' + (k === '403' ? 'lock' : 'alert-circle') + '" aria-hidden="true" data-testid="viewas-result-icon"></span><div><strong data-testid="viewas-result-title">' + esc(t('solve.viewas.' + k + '.title')) + '</strong><span data-testid="viewas-result-body">' + esc(t('solve.viewas.' + k + '.body')) + '</span></div></div>';
  });

  /* ---------- answers (simulated async list so ?state=loading|empty|error works) ---------- */
  function sorted() {
    return answers.slice().sort(function (a, b) { return (b.accepted - a.accepted) || (b.votes - a.votes) || (a.date < b.date ? -1 : 1); });
  }
  function ansHtml(a, i) {
    var n = i + 1, tid = 'answer-' + a.id, mine = persona === 'author' && th.full;
    var accept = '';
    if (persona === 'author' && th.full) {
      accept = a.accepted ? '<button type="button" class="btn btn-sm btn-ghost" data-accept="' + a.id + '" aria-describedby="ans-h-' + n + '" data-testid="answer-unaccept-button-' + a.id + '"><span class="icon i-x" aria-hidden="true" data-testid="answer-unaccept-icon-' + a.id + '"></span>' + esc(t('solve.answer.unaccept')) + '</button>'
        : '<button type="button" class="btn btn-sm" data-accept="' + a.id + '" aria-describedby="ans-h-' + n + '" data-testid="answer-accept-button-' + a.id + '"><span class="icon i-check" aria-hidden="true" data-testid="answer-accept-icon-' + a.id + '"></span>' + esc(t('solve.answer.accept')) + '</button>';
    }
    var src = a.sources && a.sources.length ? '<div class="solve-sources" data-testid="' + tid + '-sources"><span class="text-sm text-muted" data-testid="' + tid + '-sources-label">' + esc(t('solve.answer.sources')) + '</span><ul class="list-reset">' +
      a.sources.map(function (s, j) { return '<li><a class="link-external text-sm" href="' + s.url + '" rel="noopener noreferrer" lang="fr" data-testid="' + tid + '-source-' + (j + 1) + '">' + esc(s.label) + '</a></li>'; }).join('') + '</ul></div>' : '';
    return '<li data-testid="' + tid + '-item"><article class="card solve-answer' + (a.accepted ? ' is-accepted' : '') + '" id="answer-' + a.id + '" tabindex="-1" aria-labelledby="ans-h-' + n + '" data-testid="' + tid + '">' +
      '<h3 class="visually-hidden" id="ans-h-' + n + '" data-testid="' + tid + '-heading">' + esc(t('solve.answer.n', { n: n })) + '</h3>' +
      '<div class="solve-post"><div class="solve-post-vote">' + voteCtl(a.id, a.votes, tid + '-vote') + '</div><div class="solve-post-main stack">' +
      (a.accepted ? '<p><span class="badge badge-accepted" data-testid="' + tid + '-accepted"><span class="icon i-check-circle" aria-hidden="true" data-testid="' + tid + '-accepted-icon"></span>' + esc(t('solve.answer.accepted')) + '</span></p>' : '') +
      '<div class="prose" lang="' + a.lang + '" data-user-content data-testid="' + tid + '-body">' + S.renderMd(a.md, mdOpts(tid)) + '</div>' + src +
      '<div class="meta solve-post-foot">' + personBox(a.author, tid) + '<time class="meta-item" datetime="' + a.date + '" data-testid="' + tid + '-date">' + M.fmtDate(a.date) + '</time>' + langBadge(a.lang, tid + '-lang-badge') +
      (a.mine ? '<span class="badge badge-sim" data-testid="' + tid + '-mine">' + esc(t('solve.answer.mine')) + '</span>' : '') + '</div>' +
      (accept ? '<div class="cluster solve-answer-actions" data-testid="' + tid + '-actions">' + accept + '</div>' : '') + '</div></div></article></li>';
  }
  var sim = App.Sim.load($('#answers-list'), {
    delay: 500, skeleton: 2, emptyKey: 'solve.answers.empty',
    load: function () { return answers.length ? sorted() : (!th.full && q.answers > 0 ? [{ missing: true }] : []); },
    render: function (items, box) {
      if (items[0] && items[0].missing) {
        box.innerHTML = '<div class="state" data-testid="answers-fixture"><span class="icon i-inbox" aria-hidden="true" data-testid="answers-fixture-icon"></span><h3 data-testid="answers-fixture-title">' + esc(t('solve.answers.fixture.title')) + '</h3><p data-testid="answers-fixture-body">' + esc(t('solve.answers.fixture.body', { n: q.answers })) + '</p></div>'; return;
      }
      box.innerHTML = '<ol class="list-reset stack" data-testid="answers-ol">' + items.map(ansHtml).join('') + '</ol>';
    }
  });
  function renderAnswersHead() { var n = answerCount(); $('#answers-h').textContent = n ? pl('solve.answers.h', n) : t('solve.answers.h.zero'); }

  function rerender(focusSel) {
    renderHead(); renderPost(); renderAnswersHead(); sim.refresh();
    if (focusSel) { var f = $(focusSel); if (f) f.focus(); }
  }

  /* ---------- actions (event delegation; regions are re-rendered, forms are not) ---------- */
  document.addEventListener('click', function (e) {
    var v = e.target.closest('[data-vote]'), a = e.target.closest('[data-accept]');
    if (v && !v.disabled) {
      var key = v.getAttribute('data-key'), dir = v.getAttribute('data-vote'); votes[key] = votes[key] === dir ? null : dir;   /* one vote per user and target: toggle */
      rerender('[data-vote="' + dir + '"][data-key="' + key + '"]');
    } else if (a && persona === 'author') {
      var aid = a.getAttribute('data-accept'), was = answers.filter(function (x) { return x.id === aid; })[0].accepted;
      answers.forEach(function (x) { x.accepted = !was && x.id === aid; });                                                   /* at most one accepted answer */
      App.Toast.show({ kind: 'success', msg: t(was ? 'solve.answer.unaccepted.toast' : 'solve.answer.accepted.toast') });
      rerender();
      var nf = $('[data-accept="' + aid + '"]'); if (nf) nf.focus();
    }
  });

  /* ---------- answer form ---------- */
  var af = $('#answer-form'); af.elements.lang.value = I18N.lang; var aerr = null;
  function setAErr() {
    var f = $('#answer-body-field'), inp = $('#answer-body'), msg = $('#answer-body-error'), sm = $('#answer-summary');
    f.classList.toggle('has-error', !!aerr); msg.hidden = !aerr; sm.hidden = !aerr;
    if (aerr) { inp.setAttribute('aria-invalid', 'true'); inp.setAttribute('aria-describedby', 'answer-body-hint answer-body-error');
      msg.innerHTML = '<span class="icon i-alert-circle" aria-hidden="true" data-testid="answer-body-error-icon"></span><span data-testid="answer-body-error-text">' + esc(t(aerr)) + '</span>';
      $('#answer-summary-list').innerHTML = '<li><a href="#answer-body" data-testid="answer-summary-link-body">' + esc(t(aerr)) + '</a></li>'; }
    else { inp.removeAttribute('aria-invalid'); inp.setAttribute('aria-describedby', 'answer-body-hint'); }
  }
  $('#answer-summary').addEventListener('click', function (e) { if (e.target.closest('a')) { e.preventDefault(); $('#answer-body').focus(); } });
  af.addEventListener('input', function () { if (aerr) { aerr = null; setAErr(); } });
  af.addEventListener('submit', function (e) {
    e.preventDefault(); var body = af.elements.body.value.trim();
    aerr = !body ? 'solve.answer.err.req' : body.length < 30 ? 'solve.answer.err.short' : S.unbalancedFence(body) ? 'solve.ask.err.body.fence' : null; setAErr();
    if (aerr) { $('#answer-summary').focus(); return; }
    var v = viewer(); newN++;
    answers.push({ id: 'a-new-' + newN, author: v.id, lang: af.elements.lang.value, votes: 0, date: '2026-11-11', accepted: false, md: body, sources: [], mine: true });
    th.full = true; af.reset(); af.elements.lang.value = I18N.lang; $('#answer-preview').innerHTML = '';
    App.Toast.show({ kind: 'success', msg: t('solve.answer.toast') });
    rerender(); var el = document.getElementById('answer-a-new-' + newN); if (el) el.focus();
  });
  function preview() {
    var v = af.elements.body.value, box = $('#answer-preview'); box.setAttribute('lang', af.elements.lang.value); box.setAttribute('data-user-content', '');
    box.innerHTML = v.trim() ? S.renderMd(v, mdOpts('answer-preview')) : '<p class="text-muted" data-testid="answer-preview-empty">' + esc(t('solve.preview.empty')) + '</p>';
  }
  var pt = $('#answer-tab-preview'); pt.addEventListener('click', preview);   /* arrow keys also fire click (App.Tabs) */ af.elements.body.addEventListener('input', preview);

  /* ---------- boot ---------- */
  function all() { renderHead(); renderPost(); renderRelated(); renderContext(); renderViewAs(); renderAnswersHead(); }
  all();
 
  document.addEventListener('i18n:change', function () { all(); setAErr(); var r = $('#viewas-result'); r.innerHTML = ''; });
})();
