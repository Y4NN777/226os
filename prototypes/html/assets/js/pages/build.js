/* Build — project discovery. Dictionary (domain: build) + filters + simulated list + "propose a project" dialog (simulated). */
I18N.register({
  fr: {
    'build.title': 'Construire · Projets open source — 226os', 'build.h1': 'Construire · Projets open source',
    'build.lede': 'Des projets de développeurs africains qui cherchent des contributeurs. Filtrez par compétence, pays, étape, langue ou licence.',
    'build.filters.title': 'Rechercher et filtrer les projets', 'build.search.label': 'Rechercher un projet, une technologie ou un besoin', 'build.search.placeholder': 'Ex. Go, webhook, documentation',
    'build.filter.country': 'Pays', 'build.filter.stage': 'Étape', 'build.filter.lang': 'Langue de travail', 'build.filter.license': 'Licence', 'build.filter.skills': 'Compétences recherchées',
    'build.filter.openOnly': 'Avec des besoins ouverts seulement', 'build.filter.reset': 'Réinitialiser les filtres', 'build.filter.any': 'Tous',
    'build.lang.fr': 'Français', 'build.lang.en': 'English',
    'build.sort.recent': 'Plus récents', 'build.sort.name': 'Nom (A à Z)', 'build.sort.open': 'Plus de besoins ouverts',
    'build.results.title': 'Projets', 'build.results.count': '{n} projet(s)',
    'build.row.looking': 'Recherche :', 'build.row.none': 'Aucun besoin ouvert', 'build.row.open.one': '1 besoin ouvert', 'build.row.open.other': '{n} besoins ouverts',
    'build.row.origin': 'Question d’origine :', 'build.row.updated': 'Mis à jour le', 'build.row.license': 'Licence', 'build.row.stack': 'Technologies',
    'build.empty.title': 'Aucun projet ne correspond', 'build.empty.body': 'Retirez un filtre ou changez de mot-clé. Vous pouvez aussi proposer votre propre projet.',
    'build.propose.open': 'Proposer un projet', 'build.propose.title': 'Proposer un projet', 'build.propose.sim': 'Simulation : rien n’est enregistré ni importé de GitHub.',
    'build.propose.url': 'URL du dépôt GitHub public', 'build.propose.url.hint': 'Dépôts publics https://github.com uniquement (vérifié aussi côté serveur).',
    'build.propose.name': 'Nom du projet', 'build.propose.summary': 'Description courte', 'build.propose.summary.hint': '20 à 280 caractères. Décrivez le problème résolu.',
    'build.propose.submit': 'Proposer', 'build.propose.ok.title': 'Projet publié', 'build.propose.ok.body': 'Vous pouvez maintenant publier un premier besoin.',
    'build.form.summary': 'Corrigez les champs suivants',
    'build.propose.prefill': 'Nom et description repris du dépôt. Modifiez-les si besoin.', 'build.propose.next': 'Continuer', 'build.propose.back': 'Retour', 'build.propose.publish': 'Publier le projet', 'build.propose.skip': 'Publier sans besoin',
    'build.propose.step': 'Étape {i} sur {n}', 'build.propose.reading': 'Lecture du dépôt…', 'build.propose.repo': 'Dépôt lu : {r}',
    'build.propose.needs.lede': 'De quelle aide avez-vous besoin en premier ? Un besoin précis reçoit plus de candidatures.', 'build.propose.need': 'Premier besoin (facultatif)', 'build.propose.need.ph': 'Ex. Relecture de la documentation en français',
    'build.propose.need.skills': 'Compétences utiles', 'build.propose.need.skills.hint': 'Choisissez-en une ou deux.', 'build.err.need.short': 'Décrivez le besoin en 8 caractères au moins, ou laissez le champ vide.',
    'build.propose.done.body': 'Les contributeurs peuvent maintenant le trouver. Partagez le lien pour recevoir les premières candidatures.', 'build.propose.share': 'Lien du projet', 'build.propose.share.copy': 'Copier', 'build.propose.copied': 'Lien copié.',
    'build.propose.similar': 'Des profils qui pourraient aider', 'build.propose.again': 'Voir les projets', 'build.propose.sample': 'Bibliothèque open source pour les développeurs africains : un premier jet à préciser avec la communauté.',
    'build.empty.reset': 'Effacer les filtres', 'build.row.state.none': 'Complet',
    'build.err.url.required': 'Saisissez l’URL du dépôt.', 'build.err.url.invalid': 'URL invalide : utilisez https://github.com/organisation/depot.',
    'build.err.name.required': 'Saisissez le nom du projet.', 'build.err.summary.required': 'Saisissez une description.', 'build.err.summary.short': 'La description doit contenir au moins 20 caractères.'
  },
  en: {
    'build.title': 'Build — Open source projects — 226os', 'build.h1': 'Build: open source projects',
    'build.lede': 'Projects by African developers looking for contributors. Filter by skill, country, stage, language or licence.',
    'build.filters.title': 'Search and filter projects', 'build.search.label': 'Search a project, technology or need', 'build.search.placeholder': 'e.g. Go, webhook, documentation',
    'build.filter.country': 'Country', 'build.filter.stage': 'Stage', 'build.filter.lang': 'Working language', 'build.filter.license': 'Licence', 'build.filter.skills': 'Skills needed',
    'build.filter.openOnly': 'Only projects with open needs', 'build.filter.reset': 'Reset filters', 'build.filter.any': 'All',
    'build.lang.fr': 'Français', 'build.lang.en': 'English',
    'build.sort.recent': 'Most recent', 'build.sort.name': 'Name (A to Z)', 'build.sort.open': 'Most open needs',
    'build.results.title': 'Projects', 'build.results.count': '{n} project(s)',
    'build.row.looking': 'Looking for:', 'build.row.none': 'No open needs', 'build.row.open.one': '1 open need', 'build.row.open.other': '{n} open needs',
    'build.row.origin': 'Born from the question:', 'build.row.updated': 'Updated', 'build.row.license': 'Licence', 'build.row.stack': 'Stack',
    'build.empty.title': 'No project matches', 'build.empty.body': 'Remove a filter or change the keyword. You can also propose your own project.',
    'build.propose.open': 'Propose a project', 'build.propose.title': 'Propose a project', 'build.propose.sim': 'Nothing is saved or imported from GitHub in this prototype.',
    'build.propose.url': 'Public GitHub repository URL', 'build.propose.url.hint': 'Only public repositories on https://github.com are accepted (also checked server-side).',
    'build.propose.name': 'Project name', 'build.propose.summary': 'Short description', 'build.propose.summary.hint': '20 to 280 characters. Describe the problem it solves.',
    'build.propose.submit': 'Propose', 'build.propose.ok.title': 'Project proposed', 'build.propose.ok.body': 'You can now publish a first contribution need.',
    'build.form.summary': 'Please fix the following fields',
    'build.propose.prefill': 'Name and description were taken from the repository. Edit them if needed.', 'build.propose.next': 'Continue', 'build.propose.back': 'Back', 'build.propose.publish': 'Publish the project', 'build.propose.skip': 'Publish without a need',
    'build.propose.step': 'Step {i} of {n}', 'build.propose.reading': 'Reading the repository…', 'build.propose.repo': 'Repository read: {r}',
    'build.propose.needs.lede': 'What help do you need first? A precise need gets more applications.', 'build.propose.need': 'First need (optional)', 'build.propose.need.ph': 'e.g. Review the French documentation',
    'build.propose.need.skills': 'Useful skills', 'build.propose.need.skills.hint': 'Pick one or two.', 'build.err.need.short': 'Describe the need in at least 8 characters, or leave it empty.',
    'build.propose.done.body': 'Contributors can now find it. Share the link to get your first applications.', 'build.propose.share': 'Project link', 'build.propose.share.copy': 'Copy', 'build.propose.copied': 'Link copied.',
    'build.propose.similar': 'Profiles that could help', 'build.propose.again': 'See the projects', 'build.propose.sample': 'Open-source library for African developers: a first draft to refine with the community.',
    'build.empty.reset': 'Clear filters', 'build.row.state.none': 'Complete',
    'build.err.url.required': 'Enter the repository URL.', 'build.err.url.invalid': 'Invalid URL: use https://github.com/organisation/repository.',
    'build.err.name.required': 'Enter the project name.', 'build.err.summary.required': 'Enter a description.', 'build.err.summary.short': 'The description must be at least 20 characters.'
  }
});

(function () {
  'use strict';
  var B = MOCK.build, $ = App.$, esc = App.esc, t = function (k, p) { return I18N.t(k, p); };
  var all = B.projects();
  var st = { q: '', country: '', stage: '', lang: '', license: '', sort: 'recent', skills: [], openOnly: false, n: 0 };
  /* Deep links: ?q= from global search, ?skill= from onboarding (contribute intent). */
  (function () { var q0 = App.params.get('q'), k0 = App.params.get('skill'); if (q0) st.q = q0; if (k0) st.skills = [k0]; })();

  function hay(p) {
    var q = B.question(p.question);
    return [p.name, p.summary, p.stack, p.needs.map(function (n) { return n.title; }), p.skills.map(function (s) { return B.skills[s].fr + ' ' + B.skills[s].en; }), q ? q.title : '', p.country, MOCK.countryName(p.country)].join(' ');
  }
  all.forEach(function (p) { p._hay = hay(p); });

  function compute() {
    var where = { country: st.country, stage: st.stage, langs: st.lang, license: st.license, skills: st.skills };
    var sortKey = { recent: { key: 'updated', dir: 'desc' }, name: { key: 'name', dir: 'asc' }, open: { key: 'openCount', dir: 'desc' } }[st.sort];
    var out = App.Filter.run(all, { q: st.q, fields: ['_hay'], where: where, sort: sortKey });
    if (st.openOnly) out = out.filter(function (p) { return p.openCount > 0; });
    st.n = out.length; updateCount();
    return out;
  }
  function updateCount() { $('#build-count').textContent = t('build.results.count', { n: st.n }); }

  function countryHtml(code, tid) {
    return '<span class="country" data-testid="' + tid + '"><span class="country-code" aria-hidden="true" data-testid="' + tid + '-code">' + code + '</span><span class="visually-hidden" data-testid="' + tid + '-name">' + esc(MOCK.countryName(code)) + '</span></span>';
  }
  function row(p, i) {
    var n = p.id, shown = p.skills.slice(0, 3), q = B.question(p.question);
    return '<li class="row build-row" data-testid="build-row-' + n + '">' +
      '<div class="row-lead">' + App.thumb(p.name, { size: 'lg', testid: 'build-row-' + n + '-thumb' }) + '</div>' +
      '<div class="row-main">' +
        '<h3 class="row-title"><a href="project.html?id=' + p.id + '" lang="' + p.lang + '" data-user-content data-testid="build-row-' + n + '-link">' + esc(p.name) + '</a></h3>' +
        '<p class="row-excerpt" lang="' + p.lang + '" data-user-content data-testid="build-row-' + n + '-summary">' + esc(p.summary) + '</p>' +
        (shown.length ? '<ul class="tag-list build-skills" data-testid="build-row-' + n + '-skills">' + shown.map(function (k) { return '<li><span class="tag" data-testid="build-row-' + n + '-skill-' + k + '">' + esc(MOCK.label(B.skills[k])) + '</span></li>'; }).join('') + (p.skills.length > 3 ? '<li><span class="text-xs text-muted tabular-nums" data-testid="build-row-' + n + '-skills-more">+' + (p.skills.length - 3) + '</span></li>' : '') + '</ul>' : '') +
        '<div class="meta build-meta">' +
          '<span class="meta-item" data-testid="build-row-' + n + '-country">' + countryHtml(p.country, 'build-row-' + n + '-country-code') + ' ' + esc(MOCK.countryName(p.country)) + '</span>' +
          '<span class="meta-item" data-testid="build-row-' + n + '-stage">' + esc(MOCK.label(B.stages[p.stage])) + '</span>' +
          '<span class="meta-item tabular-nums" data-testid="build-row-' + n + '-updated"><time datetime="' + p.updated + '" data-testid="build-row-' + n + '-updated-date">' + esc(MOCK.fmtDate(p.updated)) + '</time></span>' +
          (q ? '<span class="meta-item row-origin" data-testid="build-row-' + n + '-origin"><span class="icon i-help" aria-hidden="true" data-testid="build-row-' + n + '-origin-icon"></span><span class="visually-hidden" data-testid="build-row-' + n + '-origin-label">' + esc(t('build.row.origin')) + '</span><a class="row-above" href="question.html?id=' + q.id + '" lang="' + q.lang + '" data-user-content data-testid="build-row-' + n + '-origin-link">' + esc(q.title) + '</a></span>' : '') +
        '</div>' +
      '</div>' +
      '<div class="row-aside">' + (p.openCount
        ? App.statePill('open', { testid: 'build-row-' + n + '-open', label: p.openCount === 1 ? t('build.row.open.one') : t('build.row.open.other', { n: p.openCount }) })
        : App.statePill('closed', { testid: 'build-row-' + n + '-none', label: t('build.row.none') })) +
      '</div></li>';
  }

  var sim = App.Sim.load($('#build-list'), {
    load: compute, skeleton: 4, empty: { icon: 'git-branch', key: 'build.empty', action: { labelKey: 'build.empty.reset', icon: 'refresh', testid: 'build-empty-reset', onClick: function () { $('#build-reset').click(); } } },
    render: function (items, box) { box.innerHTML = '<ul class="list list-reset" data-testid="build-list">' + items.map(row).join('') + '</ul>'; App.View.apply('build'); }
  });

  /* ---------- Controls (rebuilt on language change, values preserved) ---------- */
  function opts(sel, items, val) {
    sel.innerHTML = items.map(function (o) { return '<option value="' + esc(o[0]) + '"' + (o[0] === val ? ' selected' : '') + ' data-testid="' + sel.getAttribute('data-testid') + '-option-' + (o[0] || 'all') + '">' + esc(o[1]) + '</option>'; }).join('');
  }
  function fillControls() {
    var any = [['', t('build.filter.any')]];
    var countries = {}; all.forEach(function (p) { countries[p.country] = 1; });
    opts($('#build-country'), any.concat(Object.keys(countries).sort().map(function (c) { return [c, c + ' — ' + MOCK.countryName(c)]; })), st.country);
    opts($('#build-stage'), any.concat(Object.keys(B.stages).map(function (k) { return [k, MOCK.label(B.stages[k])]; })), st.stage);
    opts($('#build-lang'), any.concat([['fr', t('build.lang.fr')], ['en', t('build.lang.en')]]), st.lang);
    opts($('#build-license'), any.concat(B.licenses.map(function (l) { return [l, l]; })), st.license);
    opts($('#build-sort'), [['recent', t('build.sort.recent')], ['name', t('build.sort.name')], ['open', t('build.sort.open')]], st.sort);
    $('#build-skill-chips').innerHTML = Object.keys(B.skills).map(function (k) {
      return '<button type="button" class="chip" data-skill="' + k + '" aria-pressed="' + (st.skills.indexOf(k) > -1) + '" data-testid="build-skill-chip-' + k + '">' + esc(MOCK.label(B.skills[k])) + '</button>';
    }).join('');
    updateCount();
  }
  fillControls();
  document.addEventListener('i18n:change', fillControls);

  function sync() { sim.refresh(); }
  $('#build-filter-form').addEventListener('submit', function (e) { e.preventDefault(); });
  $('#build-q').value = st.q;
  $('#build-q').addEventListener('input', function (e) { st.q = e.target.value; sync(); });
  [['country', '#build-country'], ['stage', '#build-stage'], ['lang', '#build-lang'], ['license', '#build-license'], ['sort', '#build-sort']].forEach(function (x) {
    $(x[1]).addEventListener('change', function (e) { st[x[0]] = e.target.value; sync(); });
  });
  $('#build-open-only').addEventListener('change', function (e) { st.openOnly = e.target.checked; sync(); });
  $('#build-skill-chips').addEventListener('click', function (e) {
    var b = e.target.closest('[data-skill]'); if (!b) return; var k = b.getAttribute('data-skill'), i = st.skills.indexOf(k);
    if (i > -1) st.skills.splice(i, 1); else st.skills.push(k);
    b.setAttribute('aria-pressed', i < 0); sync();
  });
  $('#build-more-toggle').addEventListener('click', function () {
    var b = this, open = b.getAttribute('aria-expanded') !== 'true'; b.setAttribute('aria-expanded', open); b.setAttribute('aria-pressed', open); $('#build-more').hidden = !open;
  });
  $('#build-reset').addEventListener('click', function () {
    st.q = ''; st.country = st.stage = st.lang = st.license = ''; st.sort = 'recent'; st.skills = []; st.openOnly = false;
    $('#build-q').value = ''; $('#build-open-only').checked = false; fillControls(); sync(); $('#build-q').focus();
  });

  /* ---------- Propose a project: GitHub URL first, details prefilled, first need, publish, success (simulated) ---------- */
  var form = $('#propose-form'), dlg = $('#propose-dialog'), errs = {}, busy = false, step = 1, repo = null, chosen = [];
  var FIELDS = { url: '#propose-url', name: '#propose-name', summary: '#propose-summary-input', need: '#propose-need' };
  var PEOPLE_SKILL = { postgres: 'postgresql' };
  function pretty(slug) { return slug.replace(/[-_.]+/g, ' ').replace(/\b\w/, function (c) { return c.toUpperCase(); }); }
  function stepper() {
    $('#propose-stepper').textContent = step > 3 ? '' : t('build.propose.step', { i: step, n: 3 });
    [1, 2, 3].forEach(function (i) { $('#propose-step-' + i).hidden = i !== step; });
    $('#propose-success').hidden = step !== 4; $('#propose-back').hidden = step < 2 || step > 3;
    var sub = $('#propose-submit'), lab = sub.querySelector('span');
    sub.hidden = step === 4;
    lab.setAttribute('data-i18n', step === 3 ? 'build.propose.publish' : 'build.propose.next'); lab.textContent = t(step === 3 ? 'build.propose.publish' : 'build.propose.next');
    var c = $('#propose-cancel'); c.setAttribute('data-i18n', step === 4 ? 'common.close' : 'common.cancel'); c.textContent = t(step === 4 ? 'common.close' : 'common.cancel');
  }
  function fillChips() {
    $('#propose-need-chips').innerHTML = Object.keys(B.skills).map(function (k) {
      return '<button type="button" class="chip" data-nskill="' + k + '" aria-pressed="' + (chosen.indexOf(k) > -1) + '" data-testid="propose-need-chip-' + k + '">' + esc(MOCK.label(B.skills[k])) + '</button>'; }).join('');
  }
  $('#propose-need-chips').addEventListener('click', function (e) {
    var b = e.target.closest('[data-nskill]'); if (!b) return; var k = b.getAttribute('data-nskill'), i = chosen.indexOf(k);
    if (i > -1) chosen.splice(i, 1); else chosen.push(k); b.setAttribute('aria-pressed', i < 0);
  });
  function validate() {
    errs = {};
    if (step === 1) { var u = $('#propose-url').value.trim(); if (!u) errs.url = 'build.err.url.required'; else if (!/^https:\/\/github\.com\/[\w.-]+\/[\w.-]+\/?$/.test(u)) errs.url = 'build.err.url.invalid'; }
    if (step === 2) { var nm = $('#propose-name').value.trim(), sm = $('#propose-summary-input').value.trim();
      if (!nm) errs.name = 'build.err.name.required'; if (!sm) errs.summary = 'build.err.summary.required'; else if (sm.length < 20) errs.summary = 'build.err.summary.short'; }
    if (step === 3) { var nd = $('#propose-need').value.trim(); if (nd && nd.length < 8) errs.need = 'build.err.need.short'; }
    return !Object.keys(errs).length;
  }
  function showErrors(focus) {
    var keys = Object.keys(errs), sum = $('#propose-summary');
    Object.keys(FIELDS).forEach(function (f) {
      var inp = $(FIELDS[f]), msg = $('#propose-' + f + '-error'), box = inp.closest('.field'), has = !!errs[f];
      box.classList.toggle('has-error', has); if (has) inp.setAttribute('aria-invalid', 'true'); else inp.removeAttribute('aria-invalid');
      msg.hidden = !has; msg.innerHTML = has ? '<span class="icon i-alert-circle" aria-hidden="true" data-testid="propose-' + f + '-error-icon"></span><span data-testid="propose-' + f + '-error-text">' + esc(t(errs[f])) + '</span>' : '';
    });
    sum.hidden = !keys.length;
    $('#propose-summary-list').innerHTML = keys.map(function (f) { return '<li><a href="' + FIELDS[f] + '" data-testid="propose-summary-link-' + f + '">' + esc(t(errs[f])) + '</a></li>'; }).join('');
    if (focus && keys.length) sum.focus();
  }
  function similar() {
    var want = chosen.map(function (k) { return PEOPLE_SKILL[k] || k; }), ps = (MOCK.people || []).filter(function (x) { return x.skills.some(function (k) { return want.indexOf(k) > -1; }) && (x.open || []).indexOf('collaboration') > -1; }).slice(0, 3);
    if (!ps.length) return '';
    return '<h3 class="h4" data-testid="propose-similar-title">' + esc(t('build.propose.similar')) + '</h3><ul class="build-people list-reset" data-testid="propose-similar-list">' + ps.map(function (x) {
      return '<li><a class="build-person" href="profile.html?id=' + x.id + '" data-testid="propose-similar-' + x.id + '">' + App.avatar(x, { size: 'sm', testid: 'propose-similar-avatar-' + x.id }) + '<span class="build-person-body"><strong data-testid="propose-similar-name-' + x.id + '">' + esc(x.name) + '</strong><span class="text-sm text-muted" lang="' + x.lang + '" data-user-content data-testid="propose-similar-headline-' + x.id + '">' + esc(x.headline) + '</span></span></a></li>'; }).join('') + '</ul>';
  }
  function setBusy(on, key) { var b = $('#propose-submit'); busy = on; b.disabled = on; if (on) { b.setAttribute('aria-busy', 'true'); b.querySelector('span').textContent = t(key); } else b.removeAttribute('aria-busy'); }
  $('#propose-summary-list').addEventListener('click', function (e) { var a = e.target.closest('a'); if (a) { e.preventDefault(); $(a.getAttribute('href')).focus(); } });
  form.addEventListener('submit', function (e) {
    e.preventDefault(); if (busy) return;
    if (!validate()) { showErrors(true); return; }
    showErrors(false);
    if (step === 1) {
      repo = $('#propose-url').value.trim().replace(/\/$/, '').split('github.com/')[1];
      setBusy(true, 'build.propose.reading');
      setTimeout(function () {                              /* simulated GitHub read: nothing leaves the page */
        setBusy(false); step = 2; stepper();
        var slug = repo.split('/')[1]; $('#propose-repo').innerHTML = App.thumb(slug, { size: 'sm', testid: 'propose-repo-thumb' }) + '<span class="mono" data-testid="propose-repo-name">' + esc(repo) + '</span>';
        if (!$('#propose-name').value) $('#propose-name').value = pretty(slug); if (!$('#propose-summary-input').value) $('#propose-summary-input').value = t('build.propose.sample');
        $('#propose-name').focus();
      }, 600); return;
    }
    if (step === 2) { step = 3; stepper(); fillChips(); $('#propose-need').focus(); return; }
    setBusy(true, 'build.propose.publish');
    setTimeout(function () {
      setBusy(false); step = 4; stepper();
      $('#propose-success-body').textContent = t('build.propose.done.body');
      $('#propose-share-url').value = 'https://226os.example.org/projets/' + repo.split('/')[1];
      $('#propose-similar').innerHTML = similar();
      $('#propose-cancel').focus(); App.Toast.show({ kind: 'success', msg: t('build.propose.ok.title') });
    }, 700);
  });
  $('#propose-back').addEventListener('click', function () { step = Math.max(1, step - 1); errs = {}; showErrors(false); stepper(); var f = $('#propose-step-' + step).querySelector('input,textarea'); if (f) f.focus(); });
  $('#propose-share-copy').addEventListener('click', function () {
    var v = $('#propose-share-url').value; try { navigator.clipboard.writeText(v).then(function () {}, function () {}); } catch (x) { /* ignore */ }
    App.Toast.show({ kind: 'success', title: t('build.propose.copied'), ttl: 2500 });
  });
  dlg.addEventListener('close', function () { form.reset(); errs = {}; showErrors(false); step = 1; chosen = []; repo = null; stepper(); });
  document.addEventListener('i18n:change', function () { showErrors(false); stepper(); fillChips(); });
  stepper(); fillChips();

  /* gate: sign in first, then resume (?propose=1) */
  function openPropose(from) {
    var u = new URL(location.href); u.hash = ''; u.searchParams.set('propose', '1');
    if (App.Session.requireAuth(u.href)) App.Dialog.open('propose-dialog', from);
  }
  $('#build-propose-button').addEventListener('click', function () { openPropose(this); });
  if ((App.params.get('propose') === '1' || location.hash === '#propose') && App.Session.signedIn) {
    try { var cu = new URL(location.href); cu.searchParams.delete('propose'); cu.hash = ''; history.replaceState(null, '', cu); } catch (x) { /* file:// */ }
    setTimeout(function () { App.Dialog.open('propose-dialog', $('#build-propose-button')); }, 50);
  } else if (location.hash === '#propose') { openPropose($('#build-propose-button')); }
})();
