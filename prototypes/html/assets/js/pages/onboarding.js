/* Onboarding (bienvenue): ONE screen. Intent (4 cards) -> 1-3 skills -> immediate effect from mock data -> route by intent.
   Writes App.Session.update({intent, skills}) + setOnboarding({done:true}). No profile field is asked. Simulated. */
(function () {
  'use strict';
  var nb = ' ';
  I18N.register({
    fr: {
      'onb.title': 'Bienvenue — 226os', 'onb.h1': 'Bienvenue sur 226os', 'onb.lede': 'Deux questions, pour vous montrer tout de suite ce qui compte pour vous.',
      'onb.q1': 'Qu’est-ce qui vous amène' + nb + '?',
      'onb.ask': 'Poser une question', 'onb.askBody': 'Un blocage en cours',
      'onb.help': 'Aider les autres', 'onb.helpBody': 'Répondre à ce que vous connaissez',
      'onb.integrate': 'Intégrer une API', 'onb.integrateBody': 'Trouver une API fiable',
      'onb.contribute': 'Contribuer à un projet', 'onb.contributeBody': 'Rejoindre un projet ouvert',
      'onb.q2': 'Sur quoi pouvez-vous aider' + nb + '?', 'onb.q2sub': 'Choisissez 1 à 3 sujets. Vous pourrez les changer à tout moment.',
      'onb.search': 'Rechercher un sujet', 'onb.count': '{n} sur 3 choisis', 'onb.max': '3 sujets au maximum', 'onb.none': 'Aucun sujet ne correspond.',
      'onb.eff.q.one': '{n} question attend quelqu’un qui connaît {s}.', 'onb.eff.q.other': '{n} questions attendent quelqu’un qui connaît {s}.',
      'onb.eff.q.zero': 'Aucune question sans réponse sur {s} pour l’instant.',
      'onb.eff.r.one': '{n} question récente sur {s}.', 'onb.eff.r.other': '{n} questions récentes sur {s}.',
      'onb.eff.p.one': '{n} projet cherche ce profil.', 'onb.eff.p.other': '{n} projets cherchent ce profil.',
      'onb.eff.p.zero': 'Pas encore de projet qui cherche ce profil.',
      'onb.submit': 'Continuer', 'onb.skip': 'Passer'
    },
    en: {
      'onb.title': 'Welcome — 226os', 'onb.h1': 'Welcome to 226os', 'onb.lede': 'Two questions, so we can show you right away what matters to you.',
      'onb.q1': 'What brings you here?',
      'onb.ask': 'Ask a question', 'onb.askBody': 'Something is blocking you',
      'onb.help': 'Help others', 'onb.helpBody': 'Answer what you know',
      'onb.integrate': 'Integrate an API', 'onb.integrateBody': 'Find a reliable API',
      'onb.contribute': 'Contribute to a project', 'onb.contributeBody': 'Join an open project',
      'onb.q2': 'What can you help with?', 'onb.q2sub': 'Pick 1 to 3 topics. You can change them any time.',
      'onb.search': 'Search a topic', 'onb.count': '{n} of 3 selected', 'onb.max': '3 topics maximum', 'onb.none': 'No topic matches.',
      'onb.eff.q.one': '{n} question is waiting for someone who knows {s}.', 'onb.eff.q.other': '{n} questions are waiting for someone who knows {s}.',
      'onb.eff.q.zero': 'No unanswered question on {s} right now.',
      'onb.eff.r.one': '{n} recent question on {s}.', 'onb.eff.r.other': '{n} recent questions on {s}.',
      'onb.eff.p.one': '{n} project is looking for this profile.', 'onb.eff.p.other': '{n} projects are looking for this profile.',
      'onb.eff.p.zero': 'No project is looking for this profile yet.',
      'onb.submit': 'Continue', 'onb.skip': 'Skip'
    }
  });

  var $ = App.$, $$ = App.$$, t = function (k, p) { return I18N.t(k, p); }, esc = App.esc, S = App.Session, M = window.MOCK || {};
  if (!S.requireAuth()) return;
  /* Skills: id (stored in the profile), label, question tags, project-need skill ids, primary tag for routing. */
  var SK = [
    { id: 'go', l: 'Go', tags: ['go'], needs: ['go'], tag: 'go' },
    { id: 'postgresql', l: 'PostgreSQL', tags: ['postgresql', 'postgis'], needs: ['postgres'], tag: 'postgresql', need: 'postgres' },
    { id: 'node', l: 'JavaScript / Node.js', tags: ['node', 'javascript', 'typescript'], needs: ['node'], tag: 'node' },
    { id: 'react', l: 'React', tags: ['react', 'pwa'], needs: ['react'], tag: 'react' },
    { id: 'python', l: 'Python', tags: ['python', 'django'], needs: ['python'], tag: 'python' },
    { id: 'php', l: 'PHP', tags: ['php', 'laravel'], needs: ['php'], tag: 'php' },
    { id: 'docker', l: 'Docker et DevOps', lEn: 'Docker and DevOps', tags: ['docker', 'traefik', 'prometheus'], needs: ['devops'], tag: 'docker', need: 'devops' },
    { id: 'mobile-money', l: 'Mobile money', tags: ['mobile-money', 'payments'], needs: [], tag: 'mobile-money' },
    { id: 'sms', l: 'SMS et USSD', lEn: 'SMS and USSD', tags: ['sms', 'ussd'], needs: [], tag: 'sms' },
    { id: 'webhook', l: 'Webhooks', tags: ['webhook'], needs: [], tag: 'webhook' },
    { id: 'kyc', l: 'KYC', tags: ['kyc'], needs: [], tag: 'kyc' },
    { id: 'flutter', l: 'Flutter', tags: ['flutter'], needs: [], tag: 'flutter' },
    { id: 'docs', l: 'Rédaction technique', lEn: 'Technical writing', tags: ['docs'], needs: ['docs', 'review'], tag: 'docs' }
  ];
  function label(s) { return I18N.lang === 'en' && s.lEn ? s.lEn : s.l; }
  function byId(id) { return SK.filter(function (s) { return s.id === id; })[0]; }
  var st = { intent: '', skills: [], q: '' };

  /* ---------- mock data counts ---------- */
  function questions() { var S_ = M.solve && M.solve.all ? M.solve.all() : (M.questions || []); return S_; }
  function projects() { return M.build && M.build.projects ? M.build.projects() : []; }
  function counts(ids) {
    var tags = [], needs = []; ids.forEach(function (i) { var s = byId(i); if (s) { tags = tags.concat(s.tags); needs = needs.concat(s.needs); } });
    var q = questions().filter(function (x) { return x.answers === 0 && (x.tags || []).some(function (g) { return tags.indexOf(g) > -1; }); }).length;
    var p = projects().filter(function (x) { return (x.needs || []).some(function (n) { return n.status === 'open' && (n.skills || []).some(function (k) { return needs.indexOf(k) > -1; }); }); }).length;
    var r = questions().filter(function (x) { return (x.tags || []).some(function (g) { return tags.indexOf(g) > -1; }); }).length;
    return { q: q, p: p, r: r };
  }
  function list(a) { var l = a.map(function (i) { return label(byId(i)); }); return l.length > 1 ? l.slice(0, -1).join(', ') + (I18N.lang === 'fr' ? ' et ' : ' and ') + l[l.length - 1] : l[0]; }

  /* ---------- render ---------- */
  var chipsEl = $('#onb-chips'), countEl = $('#onb-count'), effEl = $('#onb-effect'), submit = $('#onb-submit');
  function renderChips() {
    var q = st.q.trim().toLowerCase(), full = st.skills.length >= 3;
    var vis = SK.filter(function (s) { return !q || label(s).toLowerCase().indexOf(q) > -1 || s.id.indexOf(q) > -1; });
    chipsEl.innerHTML = vis.length ? vis.map(function (s) {
      var on = st.skills.indexOf(s.id) > -1, dis = full && !on;
      return '<button type="button" class="chip" data-skill="' + s.id + '" aria-pressed="' + on + '"' + (dis ? ' aria-disabled="true"' : '') + ' data-testid="onb-chip-' + s.id + '">' + esc(label(s)) + '</button>';
    }).join('') : '<p class="onb-sub" data-testid="onb-chips-empty">' + esc(t('onb.none')) + '</p>';
    countEl.textContent = st.skills.length ? t('onb.count', { n: st.skills.length }) + (full ? ' · ' + t('onb.max') : '') : '';
  }
  function renderEffect() {
    if (!st.skills.length) { effEl.innerHTML = ''; return; }
    var c = counts(st.skills), s = list(st.skills);
    var qk = c.q > 0 ? (c.q === 1 ? 'onb.eff.q.one' : 'onb.eff.q.other') : c.r > 0 ? (c.r === 1 ? 'onb.eff.r.one' : 'onb.eff.r.other') : 'onb.eff.q.zero', qn = c.q > 0 ? c.q : c.r, pk = c.p === 0 ? 'onb.eff.p.zero' : c.p === 1 ? 'onb.eff.p.one' : 'onb.eff.p.other';
    effEl.innerHTML = '<p data-testid="onb-effect-questions"><span class="icon i-message" aria-hidden="true" data-testid="onb-effect-questions-icon"></span><span>' + esc(t(qk, { n: qn, s: s })) + '</span></p>' +
      '<p data-testid="onb-effect-projects"><span class="icon i-users" aria-hidden="true" data-testid="onb-effect-projects-icon"></span><span>' + esc(t(pk, { n: c.p })) + '</span></p>';
  }
  function render() { renderChips(); renderEffect(); submit.disabled = !st.intent; }

  /* ---------- events ---------- */
  $$('input[name="intent"]').forEach(function (r) {
    r.addEventListener('change', function () {
      st.intent = r.value; var s2 = $('#onb-step-skills'), first = s2.hidden; s2.hidden = false; render();
      if (first && window.matchMedia && !matchMedia('(prefers-reduced-motion: reduce)').matches) s2.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
    });
  });
  chipsEl.addEventListener('click', function (e) {
    var b = e.target.closest('[data-skill]'); if (!b || b.getAttribute('aria-disabled') === 'true') return;
    var id = b.getAttribute('data-skill'), i = st.skills.indexOf(id);
    if (i > -1) st.skills.splice(i, 1); else if (st.skills.length < 3) st.skills.push(id);
    render(); var again = $('[data-skill="' + id + '"]', chipsEl); if (again) again.focus();
  });
  $('#onb-search').addEventListener('input', function (e) { st.q = e.target.value; renderChips(); });
  document.addEventListener('i18n:change', function () { document.title = t('onb.title'); render(); });
  document.title = t('onb.title');

  function go(url) { location.href = url; }
  function home() { return '../' + (App.config.home || 'index.html'); }
  $('#onb-form').addEventListener('submit', function (e) {
    e.preventDefault(); if (!st.intent) return;
    S.update({ intent: st.intent, skills: st.skills.slice() }); S.setOnboarding({ done: true });
    var s = st.skills[0] ? byId(st.skills[0]) : null, enc = encodeURIComponent;
    if (st.intent === 'ask') go('solve.html#ask');
    else if (st.intent === 'help') go('solve.html?status=unanswered' + (s ? '&tag=' + enc(s.tag) : ''));
    else if (st.intent === 'integrate') go('discover.html');
    else go('build.html' + (s ? '?skill=' + enc(s.need || s.tag) : ''));
  });
  $('#onb-skip').addEventListener('click', function () { S.setOnboarding({ done: true }); go(home()); });
  render();
})();
