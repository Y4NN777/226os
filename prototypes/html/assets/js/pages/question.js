/* Solve question detail: public reading, gated actions (vote, answer) via App.Session.requireAuth with an anchor to resume,
   markdown body + code, answers (accepted first), accept loop (next unanswered question), openToCollab prompt for the answerer,
   related API/project. Viewer = App.Session (signed out = anonymous); the prototype role only picks author or member when signed in.
   Dictionary: assets/i18n/solve.js. Nothing is persisted. */
(function () {
  'use strict';
  var M = window.MOCK, S = M.solve, esc = App.esc, $ = App.$, $$ = App.$$, t = function (k, p) { return I18N.t(k, p); }, Sess = App.Session;
  function pl(base, n) { return t(base + (n === 1 ? '.one' : '.other'), { n: n }); }
  var id = App.params.get('id') || 'q-webhook-sig', q = S.get(id);
  if (!q) { $('#question-missing').hidden = false; document.title = t('solve.q.missing.title') + ' — 226os'; document.addEventListener('i18n:change', function () { document.title = t('solve.q.missing.title') + ' — 226os'; }); return; }
  $('#question-root').hidden = false;

  /* ---------- state ---------- */
  var th = S.thread(q), answers = th.answers, votes = {}, newN = 0, composerOpen = false, showNext = false, pending = location.hash;
  App.Proto.register({ roles: [{ id: 'author', label: { fr: 'Auteur', en: 'Author' } }, { id: 'member', label: { fr: 'Membre', en: 'Member' } }], 'default': 'author', legend: { fr: 'Connecté en tant que', en: 'Signed in as' } });
  var persona = 'anon';
  function syncPersona() { persona = !Sess.signedIn ? 'anon' : (App.Proto.role === 'member' ? 'member' : 'author'); }
  syncPersona();
  function me() { var u = Sess.user; return u ? { id: 'me', name: u.name || u.username || '?', country: u.country || '' } : null; }
  function person(pid) { return pid === 'me' ? me() : M.person(pid); }
  function mine(a) { var u = Sess.user, p = M.person(a.author); return !!(a.mine || (u && p && p.name === u.name)); }
  function isAccepted() { return th.full ? answers.some(function (a) { return a.accepted; }) : q.accepted; }
  function answerCount() { return th.full ? answers.length : q.answers; }
  function mdOpts(prefix) { return { id: prefix, headStart: 3, copyText: t('solve.code.copyshort'), copyLabel: function (l) { return t('solve.code.copy', { l: l }); }, codeLabel: function (l) { return t('solve.code.label', { l: l }); } }; }
  function resumeUrl(hash) { var u = new URL(location.href); u.searchParams.delete('session'); u.hash = hash || ''; return u.href; }   /* never carry the QA override into ?next= */
  function cc(code, tid) { return '<span class="country" data-testid="' + tid + '"><span class="country-code" data-testid="' + tid + '-code">' + code + '</span><span class="visually-hidden" data-testid="' + tid + '-name">' + esc(M.countryName(code)) + '</span></span>'; }
  function personBox(pid, prefix, withCountry) {
    var p = person(pid); if (!p) return '';
    var name = pid === 'me' ? esc(p.name) : '<a href="profile.html?id=' + p.id + '" data-testid="' + prefix + '-author">' + esc(p.name) + '</a>';
    return '<span class="person">' + App.avatar(p, { size: 'sm', testid: prefix + '-avatar' }) + (pid === 'me' ? '<span data-testid="' + prefix + '-author">' + name + '</span>' : name) + '</span>' + (withCountry && p.country ? cc(p.country, prefix + '-country') : '');
  }
  function pill(kind, label, tid) { return '<span class="state-pill state-pill-' + kind + '" data-testid="' + tid + '"><span class="icon i-' + (kind === 'accepted' ? 'check-circle' : 'circle-dot') + '" aria-hidden="true" data-testid="' + tid + '-icon"></span>' + esc(label) + '</span>'; }

  /* ---------- vote control (signed out: the click is gated, the buttons stay enabled) ---------- */
  function voteCtl(key, base, tid) {
    var mineV = votes[key] || null, n = base + (mineV === 'up' ? 1 : mineV === 'down' ? -1 : 0);
    function b(dir, ic) {
      return '<button type="button" id="vote-' + key + '-' + dir + '" data-vote="' + dir + '" data-key="' + key + '" aria-pressed="' + (mineV === dir) + '" aria-label="' + esc(t('solve.vote.' + dir)) + '" data-testid="' + tid + '-' + dir + '"><span class="icon i-' + ic + '" aria-hidden="true" data-testid="' + tid + '-' + dir + '-icon"></span></button>';
    }
    return '<div class="vote" role="group" aria-label="' + esc(t('solve.vote.group')) + '" data-testid="' + tid + '">' + b('up', 'arrow-up') +
      '<span class="vote-count tabular-nums" data-testid="' + tid + '-count">' + n + '</span><span class="visually-hidden" data-testid="' + tid + '-label">' + esc(t('solve.vote.votes')) + '</span>' + b('down', 'arrow-down') + '</div>';
  }

  /* ---------- head, post, side ---------- */
  function renderHead() {
    var h1 = $('#question-h1'); h1.setAttribute('lang', q.lang); h1.setAttribute('data-user-content', ''); h1.textContent = q.title;
    document.title = q.title + ' — 226os';
    var n = answerCount();
    $('#question-head-meta').innerHTML =
      '<span class="meta-item" data-testid="question-status">' + (isAccepted() ? pill('accepted', t('solve.q.status.accepted'), 'question-status-pill') : pill('open', t(n ? 'solve.q.status.open' : 'solve.q.status.none'), 'question-status-pill')) + '</span>' +
      '<span class="meta-item" data-testid="question-by">' + personBox(q.author, 'question', true) + '</span>' +
      '<span class="meta-item" data-testid="question-asked"><time datetime="' + q.date + '" data-testid="question-date">' + M.fmtDate(q.date) + '</time></span>' +
      '<span class="meta-item" data-testid="question-answer-count"><span class="icon i-message" aria-hidden="true" data-testid="question-answer-count-icon"></span><span class="tabular-nums" data-testid="question-answer-count-text">' + esc(pl('solve.q.answers.n', n)) + '</span></span>' +
      '<span class="meta-item solve-lang" data-testid="question-lang-badge"><span aria-hidden="true" data-testid="question-lang-badge-code">' + q.lang.toUpperCase() + '</span><span class="visually-hidden" data-testid="question-lang-badge-name">' + esc(t('solve.row.lang', { l: t('solve.lang.' + q.lang) })) + '</span></span>';
  }
  function renderPost() {
    $('#question-post').innerHTML = '<article class="solve-post solve-surface" aria-labelledby="question-h1" data-testid="question-article">' +
      '<div class="solve-post-vote">' + voteCtl('q', q.votes, 'question-vote') + '</div>' +
      '<div class="solve-post-main stack">' +
      '<div class="prose" lang="' + q.lang + '" data-user-content data-testid="question-body">' + S.renderMd(th.body, mdOpts('question')) + '</div>' +
      '<ul class="solve-tags list-reset" data-testid="question-tags">' + q.tags.map(function (x) { return '<li><a class="mono" href="solve.html?tag=' + esc(x) + '" data-testid="question-tag-' + x + '">' + esc(x) + '</a></li>'; }).join('') + '</ul></div></article>';
  }
  function renderRelated() {
    var r = M.byId(M.resources, q.resource), p = M.byId(M.projects, q.project), rq = q.resource ? S.all().filter(function (x) { return x.id !== q.id && x.resource === q.resource; })[0] : null, items = [];
    function item(kind, icon, href, name, sub) {
      return '<a class="linked-item solve-linked" href="' + href + '" data-testid="related-' + kind + '">' + App.thumb(name, { size: 'sm', icon: icon, testid: 'related-' + kind + '-thumb' }) + '<span class="solve-linked-text"><span class="kind" data-testid="related-' + kind + '-kind">' + esc(t('solve.related.' + (kind === 'resource' ? 'api' : kind))) + '</span>' +
        '<strong data-testid="related-' + kind + '-title"' + (kind === 'question' ? ' lang="' + rq.lang + '" data-user-content' : '') + '>' + esc(name) + '</strong><span class="text-xs text-muted" data-testid="related-' + kind + '-sub">' + esc(sub) + '</span></span></a>';
    }
    if (r) items.push(item('resource', 'compass', 'resource.html?id=' + r.id, r.name, r.provider + ' · ' + r.version));
    if (p) items.push(item('project', 'hammer', 'project.html?id=' + p.id, p.name, p.summary));
    if (rq) items.push(item('question', 'help', 'question.html?id=' + rq.id, rq.title, pl('solve.q.answers.n', rq.answers)));
    $('#question-related').innerHTML = items.length ? '<div class="solve-related" data-testid="related-items">' + items.join('') + '</div>' : '<p class="text-muted text-sm" data-testid="related-none">' + esc(t('solve.related.none')) + '</p>';
  }
  function renderContext() {
    $('#question-context').innerHTML = [
      ['env', t('solve.context.env'), '<span class="mono" data-testid="context-env-text">' + esc(q.env) + '</span>'],
      ['country', t('solve.context.country'), cc(q.country, 'context-country-flag') + ' <span data-testid="context-country-name">' + esc(M.countryName(q.country)) + '</span>']
    ].map(function (r) { return '<dt data-testid="context-' + r[0] + '-term">' + esc(r[1]) + '</dt><dd data-testid="context-' + r[0] + '-value">' + r[2] + '</dd>'; }).join('');
  }

  /* ---------- simulated viewer panel (prototype only) ---------- */
  function renderViewAs() {
    var v = me();
    $('#viewas-desc').textContent = t('solve.viewas.desc.' + persona, { name: v ? v.name : '' });
    $('#viewas-api').hidden = persona === 'author';
  }
  $('#viewas-api-btn').addEventListener('click', function () {
    var k = persona === 'anon' ? '401' : '403';
    $('#viewas-result').innerHTML = '<div class="callout callout-danger" data-testid="viewas-result-' + k + '"><span class="icon icon-lead i-' + (k === '403' ? 'lock' : 'alert-circle') + '" aria-hidden="true" data-testid="viewas-result-icon"></span><div><strong data-testid="viewas-result-title">' + esc(t('solve.viewas.' + k + '.title')) + '</strong><span data-testid="viewas-result-body">' + esc(t('solve.viewas.' + k + '.body')) + '</span></div></div>';
  });
  function onWho() { syncPersona(); $('#viewas-result').innerHTML = ''; showNext = false; renderComposer(); renderViewAs(); rerender(); }
  document.addEventListener('proto:role', onWho);
  document.addEventListener('session:change', onWho);

  /* ---------- answers ---------- */
  function sorted() { return answers.slice().sort(function (a, b) { return (b.accepted - a.accepted) || (b.votes - a.votes) || (a.date < b.date ? -1 : 1); }); }
  function ansHtml(a, i) {
    var n = i + 1, tid = 'answer-' + a.id, accept = '';
    if (persona === 'author' && th.full) {
      accept = '<button type="button" class="btn btn-icon btn-sm solve-accept tip-end" data-accept="' + a.id + '" aria-pressed="' + !!a.accepted + '" aria-label="' + esc(t(a.accepted ? 'solve.answer.unaccept' : 'solve.answer.accept')) + '" data-testid="answer-accept-button-' + a.id + '"><span class="icon i-check-circle" aria-hidden="true" data-testid="answer-accept-icon-' + a.id + '"></span></button>';
    }
    var src = a.sources && a.sources.length ? '<div class="solve-sources" data-testid="' + tid + '-sources"><p class="solve-sources-h" data-testid="' + tid + '-sources-label"><span class="icon i-link" aria-hidden="true" data-testid="' + tid + '-sources-icon"></span>' + esc(t('solve.answer.sources')) + '</p><ul class="list-reset">' +
      a.sources.map(function (s, j) { return '<li><a class="link-external text-sm" href="' + s.url + '" rel="noopener noreferrer" lang="fr" data-testid="' + tid + '-source-' + (j + 1) + '">' + esc(s.label) + '</a></li>'; }).join('') + '</ul></div>' : '';
    return '<li data-testid="' + tid + '-item"><article class="solve-answer solve-surface' + (a.accepted ? ' is-accepted' : '') + '" id="answer-' + a.id + '" tabindex="-1" aria-labelledby="ans-h-' + n + '" data-testid="' + tid + '">' +
      '<h3 class="visually-hidden" id="ans-h-' + n + '" data-testid="' + tid + '-heading">' + esc(t('solve.answer.n', { n: n })) + '</h3>' +
      (a.accepted ? '<p class="solve-answer-head">' + pill('accepted', t('solve.answer.accepted'), tid + '-accepted') + '</p>' : '') +
      '<div class="solve-post"><div class="solve-post-vote">' + voteCtl(a.id, a.votes, tid + '-vote') + accept + '</div><div class="solve-post-main stack">' +
      '<div class="prose" lang="' + a.lang + '" data-user-content data-testid="' + tid + '-body">' + S.renderMd(a.md, mdOpts(tid)) + '</div>' + src +
      '<div class="meta solve-post-foot">' + personBox(a.author, tid, false) + '<time class="meta-item" datetime="' + a.date + '" data-testid="' + tid + '-date">' + M.fmtDate(a.date) + '</time><span class="solve-lang" data-testid="' + tid + '-lang-badge"><span aria-hidden="true" data-testid="' + tid + '-lang-badge-code">' + a.lang.toUpperCase() + '</span><span class="visually-hidden" data-testid="' + tid + '-lang-badge-name">' + esc(t('solve.row.lang', { l: t('solve.lang.' + a.lang) })) + '</span></span>' +
      (a.mine ? '<span class="badge badge-sim" data-proto data-testid="' + tid + '-mine">' + esc(t('solve.answer.mine')) + '</span>' : '') + '</div>' +
      '</div></div></article></li>';
  }
  function emptyHtml() {
    var own = persona === 'author', sim = S.similar(q.title, q.tags, q.id, 1)[0];
    return '<div class="solve-empty" data-testid="answers-empty"><span class="icon i-message" aria-hidden="true" data-testid="answers-empty-icon"></span><div>' +
      '<h3 data-testid="answers-empty-title">' + esc(t(own ? 'solve.answers.empty.author.title' : 'solve.answers.empty.title')) + '</h3>' +
      '<p class="text-sm text-muted" data-testid="answers-empty-body">' + esc(t(own ? 'solve.answers.empty.author.body' : 'solve.answers.empty.body')) + '</p>' +
      (sim ? '<a class="solve-empty-sim" href="question.html?id=' + sim.id + '" lang="' + sim.lang + '" data-testid="answers-empty-similar"><span class="text-muted text-xs" data-testid="answers-empty-similar-label">' + esc(t('solve.answers.empty.similar')) + '</span><span data-user-content data-testid="answers-empty-similar-title">' + esc(sim.title) + '</span></a>' : '') +
      '</div></div>';
  }
  var sim = App.Sim.load($('#answers-list'), {
    delay: 400, skeleton: 2, emptyKey: 'solve.answers.empty',
    load: function () { return answers.length ? sorted() : (!th.full && q.answers > 0 ? [{ missing: true }] : [{ none: true }]); },
    render: function (items, box) {
      if (items[0] && items[0].missing) {
        box.innerHTML = '<p class="text-sm text-muted" data-testid="answers-fixture"><span class="icon i-inbox" aria-hidden="true" data-testid="answers-fixture-icon"></span> <span data-testid="answers-fixture-title">' + esc(t('solve.answers.fixture.title')) + '</span> <span data-proto data-testid="answers-fixture-body">' + esc(t('solve.answers.fixture.body', { n: q.answers })) + '</span></p>'; afterRender(); return;
      }
      if (items[0] && items[0].none) { box.innerHTML = emptyHtml(); afterRender(); return; }
      box.innerHTML = '<ol class="list-reset solve-answer-list" data-testid="answers-ol">' + items.map(ansHtml).join('') + '</ol>'; afterRender();
    }
  });
  function renderAnswersHead() { var n = answerCount(); $('#answers-h').textContent = n ? pl('solve.answers.h', n) : t('solve.answers.h.zero'); }

  /* ---------- accept loop + openToCollab prompt ---------- */
  function renderNext() {
    var box = $('#answers-next');
    if (!showNext || !isAccepted()) { box.innerHTML = ''; return; }
    var o = S.unansweredByTag(q.tags, q.id, 1)[0], any = false;
    if (!o) { o = S.all().filter(function (x) { return x.id !== q.id && x.answers === 0; })[0]; any = !!o; }
    box.innerHTML = '<div class="solve-next" role="status" data-testid="solve-next"><span class="icon i-check-circle" aria-hidden="true" data-testid="solve-next-icon"></span><div class="solve-next-body" data-testid="solve-next-body"><strong data-testid="solve-next-title">' + esc(t('solve.next.title')) + '</strong>' +
      (o ? '<a class="solve-next-link" href="question.html?id=' + o.id + '" data-testid="solve-next-link"><span class="text-sm" data-testid="solve-next-link-label">' + esc(any ? t('solve.next.any') : t('solve.next.other', { tag: o.tagMatch })) + '</span><span class="solve-next-q" lang="' + o.lang + '" data-user-content data-testid="solve-next-link-title">' + esc(o.title) + '</span></a>'
        : '<a class="solve-next-link" href="solve.html?status=unanswered" data-testid="solve-next-link"><span class="text-sm" data-testid="solve-next-link-label">' + esc(t('solve.next.none')) + '</span></a>') +
      '</div><button type="button" class="btn btn-icon btn-sm" id="solve-next-close" data-i18n-attr="aria-label:solve.next.close" data-testid="solve-next-close"><span class="icon i-x" aria-hidden="true" data-testid="solve-next-close-icon"></span></button></div>';
    I18N.apply(box);
  }
  /* the answerer (not the asker) gets the "be found by maintainers" prompt once one of their answers is accepted */
  function syncPrompt() {
    var box = $('#question-prompt');
    if (!Sess.signedIn || !answers.some(function (a) { return a.accepted && mine(a); })) { box.innerHTML = ''; return; }
    if (!box.children.length) Sess.prompt('openToCollab', { container: box, testid: 'solve-jit-openToCollab' });
  }
  $('#answers-next').addEventListener('click', function (e) { if (e.target.closest('#solve-next-close')) { showNext = false; renderNext(); } });

  /* ---------- composer (gated) ---------- */
  var af = $('#answer-form'); af.elements.lang.value = I18N.lang; var aerr = null;
  function renderComposer() {
    $('#answer-cta').hidden = composerOpen && Sess.signedIn; af.hidden = !(composerOpen && Sess.signedIn);
    $('#answer-cta-hint').textContent = t(Sess.signedIn ? 'solve.answer.cta.hint' : 'solve.answer.cta.hint.anon');
  }
  function openComposer(focus) { composerOpen = true; renderComposer(); if (focus) { af.elements.body.focus(); $('#answer').scrollIntoView({ block: 'center' }); } }
  $('#answer-open').addEventListener('click', function () {
    if (!Sess.requireAuth(resumeUrl('#answer'))) return;   /* signed out: sign-in, then back here with the editor open */
    openComposer(true);
  });
  $('#answer-cancel').addEventListener('click', function () { composerOpen = false; renderComposer(); $('#answer-open').focus(); });
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
    if (!Sess.requireAuth(resumeUrl('#answer'))) return;
    newN++;
    answers.push({ id: 'a-new-' + newN, author: 'me', lang: af.elements.lang.value, votes: 0, date: '2026-11-11', accepted: false, md: body, sources: [], mine: true });
    th.full = true; af.reset(); af.elements.lang.value = I18N.lang; $('#answer-preview').innerHTML = ''; composerOpen = false; renderComposer();
    App.Toast.show({ kind: 'success', msg: t('solve.answer.toast') });
    rerender(); var el = document.getElementById('answer-a-new-' + newN); if (el) el.focus();
  });
  function preview() {
    var v = af.elements.body.value, box = $('#answer-preview'); box.setAttribute('lang', af.elements.lang.value); box.setAttribute('data-user-content', '');
    box.innerHTML = v.trim() ? S.renderMd(v, mdOpts('answer-preview')) : '<p class="text-muted" data-testid="answer-preview-empty">' + esc(t('solve.preview.empty')) + '</p>';
  }
  $('#answer-tab-preview').addEventListener('click', preview);   /* arrow keys also fire click (App.Tabs) */ af.elements.body.addEventListener('input', preview);

  /* ---------- actions (event delegation; regions are re-rendered, forms are not) ---------- */
  document.addEventListener('click', function (e) {
    var v = e.target.closest('[data-vote]'), a = e.target.closest('[data-accept]');
    if (v) {
      var key = v.getAttribute('data-key'), dir = v.getAttribute('data-vote');
      if (!Sess.requireAuth(resumeUrl('#vote-' + key + '-' + dir))) return;   /* gated: sign in, then land back on this control */
      votes[key] = votes[key] === dir ? null : dir;   /* one vote per user and target: toggle */
      rerender('[data-vote="' + dir + '"][data-key="' + key + '"]');
    } else if (a && persona === 'author') {
      var aid = a.getAttribute('data-accept'), was = answers.filter(function (x) { return x.id === aid; })[0].accepted;
      answers.forEach(function (x) { x.accepted = !was && x.id === aid; });                                                   /* at most one accepted answer */
      showNext = !was;
      App.Toast.show({ kind: 'success', msg: t(was ? 'solve.answer.unaccepted.toast' : 'solve.answer.accepted.toast') });
      rerender();
      var nf = $('[data-accept="' + aid + '"]'); if (nf) nf.focus();
    }
  });

  function rerender(focusSel) {
    renderHead(); renderPost(); renderAnswersHead(); renderNext(); syncPrompt(); sim.refresh();
    if (focusSel) { var f = $(focusSel); if (f) f.focus(); }
  }

  /* ---------- resume after sign-in (anchor in the URL; only when signed in; never applies an action silently) ---------- */
  var resumed = false;
  function afterRender() {
    syncPrompt();
    if (resumed || !Sess.signedIn || !pending) return;
    if (pending === '#answer') { resumed = true; openComposer(true); }
    else if (pending.indexOf('#vote-') === 0) {
      var b = document.getElementById(pending.slice(1)); if (!b) return; resumed = true;
      b.scrollIntoView({ block: 'center' }); b.focus(); b.classList.add('is-resumed'); App.Toast.show({ kind: 'info', msg: t('solve.vote.resume') });
    }
  }

  /* ---------- boot ---------- */
  function all() { renderHead(); renderPost(); renderRelated(); renderContext(); renderViewAs(); renderAnswersHead(); renderNext(); renderComposer(); }
  all();
  document.addEventListener('i18n:change', function () { all(); setAErr(); $('#viewas-result').innerHTML = ''; });
})();
