/* Build — project discovery. Dictionary (domain: build) + filters + simulated list + "propose a project" dialog (simulated). */
I18N.register({
  fr: {
    'build.title': 'Build — Projets open source — 226os', 'build.h1': 'Build : projets open source',
    'build.lede': 'Des projets de développeurs africains qui cherchent des contributeurs. Filtrez par compétence, pays, étape, langue ou licence.',
    'build.filters.title': 'Rechercher et filtrer les projets', 'build.search.label': 'Rechercher un projet, une technologie ou un besoin', 'build.search.placeholder': 'Ex. Go, webhook, documentation',
    'build.filter.country': 'Pays', 'build.filter.stage': 'Étape', 'build.filter.lang': 'Langue de travail', 'build.filter.license': 'Licence', 'build.filter.skills': 'Compétences recherchées (au moins une)',
    'build.filter.openOnly': 'Uniquement les projets avec des besoins ouverts', 'build.filter.reset': 'Réinitialiser les filtres', 'build.filter.any': 'Tous',
    'build.lang.fr': 'Français', 'build.lang.en': 'English',
    'build.sort.recent': 'Plus récents', 'build.sort.name': 'Nom (A à Z)', 'build.sort.open': 'Plus de besoins ouverts',
    'build.results.title': 'Projets', 'build.results.count': '{n} projet(s)',
    'build.row.looking': 'Recherche :', 'build.row.none': 'Aucun besoin ouvert', 'build.row.open.one': '1 besoin ouvert', 'build.row.open.other': '{n} besoins ouverts',
    'build.row.origin': 'Né de la question :', 'build.row.updated': 'Mis à jour le', 'build.row.license': 'Licence', 'build.row.stack': 'Technologies',
    'build.empty.title': 'Aucun projet ne correspond', 'build.empty.body': 'Retirez un filtre ou changez de mot-clé. Vous pouvez aussi proposer votre propre projet.',
    'build.propose.open': 'Proposer un projet', 'build.propose.title': 'Proposer un projet', 'build.propose.sim': 'Rien n’est enregistré ni importé depuis GitHub dans ce prototype.',
    'build.propose.url': 'URL du dépôt GitHub public', 'build.propose.url.hint': 'Seuls les dépôts publics sur https://github.com sont acceptés (vérifié aussi côté serveur).',
    'build.propose.name': 'Nom du projet', 'build.propose.summary': 'Description courte', 'build.propose.summary.hint': '20 à 280 caractères. Décrivez le problème résolu.',
    'build.propose.submit': 'Proposer', 'build.propose.ok.title': 'Projet proposé (simulation)', 'build.propose.ok.body': 'Dans le produit, la fiche apparaîtrait ici et vous pourriez publier un premier besoin de contribution.',
    'build.form.summary': 'Corrigez les champs suivants',
    'build.err.url.required': 'Saisissez l’URL du dépôt.', 'build.err.url.invalid': 'URL invalide : utilisez https://github.com/organisation/depot.',
    'build.err.name.required': 'Saisissez le nom du projet.', 'build.err.summary.required': 'Saisissez une description.', 'build.err.summary.short': 'La description doit contenir au moins 20 caractères.'
  },
  en: {
    'build.title': 'Build — Open source projects — 226os', 'build.h1': 'Build: open source projects',
    'build.lede': 'Projects by African developers looking for contributors. Filter by skill, country, stage, language or licence.',
    'build.filters.title': 'Search and filter projects', 'build.search.label': 'Search a project, technology or need', 'build.search.placeholder': 'e.g. Go, webhook, documentation',
    'build.filter.country': 'Country', 'build.filter.stage': 'Stage', 'build.filter.lang': 'Working language', 'build.filter.license': 'Licence', 'build.filter.skills': 'Skills needed (at least one)',
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
    'build.propose.submit': 'Propose', 'build.propose.ok.title': 'Project proposed (simulation)', 'build.propose.ok.body': 'In the product the project page would appear here and you could publish a first contribution need.',
    'build.form.summary': 'Please fix the following fields',
    'build.err.url.required': 'Enter the repository URL.', 'build.err.url.invalid': 'Invalid URL: use https://github.com/organisation/repository.',
    'build.err.name.required': 'Enter the project name.', 'build.err.summary.required': 'Enter a description.', 'build.err.summary.short': 'The description must be at least 20 characters.'
  }
});

(function () {
  'use strict';
  var B = MOCK.build, $ = App.$, esc = App.esc, t = function (k, p) { return I18N.t(k, p); };
  var all = B.projects();
  var st = { q: '', country: '', stage: '', lang: '', license: '', sort: 'recent', skills: [], openOnly: false, n: 0 };

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
    var n = p.id, q = B.question(p.question)   /* ids use the stable project id: build-row-<project-id>-... */, looking = p.skills.map(function (s) { return MOCK.label(B.skills[s]); }).join(', ');
    return '<li class="row" data-testid="build-row-' + n + '">' +
      '<div class="row-lead"><span class="icon i-hammer" aria-hidden="true" data-testid="build-row-' + n + '-icon"></span></div>' +
      '<div class="row-main">' +
        '<h3 class="row-title"><a href="project.html?id=' + p.id + '" lang="' + p.lang + '" data-user-content data-testid="build-row-' + n + '-link">' + esc(p.name) + '</a></h3>' +
        '<p class="row-excerpt" lang="' + p.lang + '" data-user-content data-testid="build-row-' + n + '-summary">' + esc(p.summary) + '</p>' +
        '<p class="build-looking" data-testid="build-row-' + n + '-needs">' + (p.openCount
          ? '<span class="text-muted" data-testid="build-row-' + n + '-needs-label">' + esc(t('build.row.looking')) + '</span> <strong data-testid="build-row-' + n + '-needs-list">' + esc(looking) + '</strong>'
          : '<span class="badge badge-later" data-testid="build-row-' + n + '-needs-none">' + esc(t('build.row.none')) + '</span>') + '</p>' +
        '<ul class="tag-list list-reset" aria-label="' + esc(t('build.row.stack')) + '" data-testid="build-row-' + n + '-stack">' + p.stack.map(function (s) { return '<li><span class="tag" data-testid="build-row-' + n + '-tag-' + s.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '') + '">' + esc(s) + '</span></li>'; }).join('') + '</ul>' +
        (q ? '<p class="text-sm build-origin" data-testid="build-row-' + n + '-origin"><span class="icon i-help" aria-hidden="true" data-testid="build-row-' + n + '-origin-icon"></span> <span class="text-muted" data-testid="build-row-' + n + '-origin-label">' + esc(t('build.row.origin')) + '</span> <a class="row-above" href="question.html?id=' + q.id + '" lang="' + q.lang + '" data-user-content data-testid="build-row-' + n + '-origin-link">' + esc(q.title) + '</a></p>' : '') +
        '<div class="meta">' +
          '<span class="meta-item" data-testid="build-row-' + n + '-license"><span class="icon i-shield-check" aria-hidden="true" data-testid="build-row-' + n + '-license-icon"></span>' + esc(p.license) + '</span>' +
          '<span class="meta-item" data-testid="build-row-' + n + '-updated"><span class="icon i-clock" aria-hidden="true" data-testid="build-row-' + n + '-updated-icon"></span>' + esc(t('build.row.updated')) + ' <time datetime="' + p.updated + '" data-testid="build-row-' + n + '-updated-date">' + esc(MOCK.fmtDate(p.updated)) + '</time></span>' +
        '</div>' +
      '</div>' +
      '<div class="row-aside">' + countryHtml(p.country, 'build-row-' + n + '-country') +
        '<span class="badge" data-testid="build-row-' + n + '-stage">' + esc(MOCK.label(B.stages[p.stage])) + '</span>' +
        '<span class="badge ' + (p.openCount ? 'badge-accepted' : 'badge-later') + '" data-testid="build-row-' + n + '-open"><span class="icon i-users" aria-hidden="true" data-testid="build-row-' + n + '-open-icon"></span>' + esc(p.openCount === 1 ? t('build.row.open.one') : t('build.row.open.other', { n: p.openCount })) + '</span>' +
        '<span class="badge" data-testid="build-row-' + n + '-lang" title="' + esc(t('common.originalLang')) + '"><span class="visually-hidden" data-testid="build-row-' + n + '-lang-label">' + esc(t('common.originalLang')) + ' </span>' + p.lang.toUpperCase() + '</span>' +
      '</div></li>';
  }

  var sim = App.Sim.load($('#build-list'), {
    load: compute, skeleton: 4, emptyKey: 'build.empty',
    render: function (items, box) { box.innerHTML = '<ul class="list list-reset" data-testid="build-list">' + items.map(row).join('') + '</ul>'; }
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
  $('#build-reset').addEventListener('click', function () {
    st.q = ''; st.country = st.stage = st.lang = st.license = ''; st.sort = 'recent'; st.skills = []; st.openOnly = false;
    $('#build-q').value = ''; $('#build-open-only').checked = false; fillControls(); sync(); $('#build-q').focus();
  });

  /* ---------- Propose a project (simulated; the form is never rebuilt on language change) ---------- */
  var form = $('#propose-form'), dlg = $('#propose-dialog'), errs = {}, busy = false;
  opts($('#propose-stage'), Object.keys(B.stages).map(function (k) { return [k, MOCK.label(B.stages[k])]; }), 'idea');
  opts($('#propose-license'), B.licenses.map(function (l) { return [l, l]; }), 'MIT');
  document.addEventListener('i18n:change', function () {
    var s = $('#propose-stage').value, l = $('#propose-license').value;
    opts($('#propose-stage'), Object.keys(B.stages).map(function (k) { return [k, MOCK.label(B.stages[k])]; }), s); opts($('#propose-license'), B.licenses.map(function (x) { return [x, x]; }), l); showErrors(false);
  });
  var FIELDS = { url: '#propose-url', name: '#propose-name', summary: '#propose-summary-input' };
  function validate() {
    errs = {};
    var u = $('#propose-url').value.trim(), nm = $('#propose-name').value.trim(), sm = $('#propose-summary-input').value.trim();
    if (!u) errs.url = 'build.err.url.required'; else if (!/^https:\/\/github\.com\/[\w.-]+\/[\w.-]+\/?$/.test(u)) errs.url = 'build.err.url.invalid';
    if (!nm) errs.name = 'build.err.name.required';
    if (!sm) errs.summary = 'build.err.summary.required'; else if (sm.length < 20) errs.summary = 'build.err.summary.short';
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
    $('#propose-summary-list').innerHTML = keys.map(function (f, i) { return '<li><a href="' + FIELDS[f] + '" data-testid="propose-summary-link-' + f + '">' + esc(t(errs[f])) + '</a></li>'; }).join('');
    if (focus && keys.length) sum.focus();
  }
  $('#propose-summary-list').addEventListener('click', function (e) { var a = e.target.closest('a'); if (a) { e.preventDefault(); $(a.getAttribute('href')).focus(); } });
  form.addEventListener('submit', function (e) {
    e.preventDefault(); if (busy) return;
    if (!validate()) { showErrors(true); return; }
    showErrors(false); busy = true; var b = $('#propose-submit'); b.setAttribute('aria-busy', 'true'); b.disabled = true;
    setTimeout(function () {
      busy = false; b.removeAttribute('aria-busy'); b.hidden = true; b.disabled = false;
      $('#propose-body').hidden = true; $('#propose-success').hidden = false; $('#propose-cancel').setAttribute('data-i18n', 'common.close'); $('#propose-cancel').textContent = t('common.close'); $('#propose-cancel').focus();
      App.Toast.show({ kind: 'success', title: t('sim.badge'), msg: t('build.propose.ok.title') });
    }, 700);
  });
  dlg.addEventListener('close', function () {
    form.reset(); errs = {}; showErrors(false); $('#propose-body').hidden = false; $('#propose-success').hidden = true; $('#propose-submit').hidden = false;
    $('#propose-cancel').setAttribute('data-i18n', 'common.cancel'); $('#propose-cancel').textContent = t('common.cancel');
  });
  if (location.hash === '#propose') App.Dialog.open('propose-dialog');
})();
