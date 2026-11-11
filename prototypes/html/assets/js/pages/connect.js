/* Connect: designed journey (PRD §8: P0 = documented scope and designed journey; the rest = planned, P1/P2). Intended discovery pattern only; actions are non-functional. Filters work client-side on fictional samples. */
I18N.register({
  fr: {
    'connect.soon.title': 'Bientôt disponible', 'connect.soon.body': 'Mentorat, groupes d’étude et mise en relation entre pairs : bientôt.', 'connect.soon.back': 'Retour à l’accueil',
    'connect.title': 'Connect — 226os', 'connect.badge': 'Bientôt',
    'connect.lede': 'Mentorat, relecture de CV, préparation d’entretiens et groupes d’étude.',
    'connect.note.title': 'Non fonctionnel (P1/P2). ',
    'connect.note.body': 'Connect est prévu en P1/P2 (PRD §8). Les entrées ci-dessous sont des exemples fictifs qui montrent le modèle de découverte. Aucune mise en relation n’est possible.',
    'connect.disc.h': 'Découvrir mentors et collaborations', 'connect.disc.lede': 'Filtrez par compétence, pays et langue. Les actions de mise en relation arrivent plus tard.',
    'connect.filters.label': 'Filtrer les exemples Connect', 'connect.f.skill': 'Compétence', 'connect.f.lang': 'Langue parlée', 'connect.f.reset': 'Réinitialiser',
    'connect.f.skill.all': 'Toutes', 'connect.f.country.all': 'Tous', 'connect.f.lang.all': 'Toutes',
    'connect.lang.fr': 'Français', 'connect.lang.en': 'Anglais',
    'connect.count': '{n} exemple(s) affiché(s)',
    'connect.kind.mentor-offer': 'Offre de mentorat', 'connect.kind.mentor-request': 'Demande de mentorat', 'connect.kind.cv-review': 'Relecture de CV / portfolio',
    'connect.kind.interview': 'Préparation d’entretien', 'connect.kind.study-group': 'Groupe d’étude', 'connect.kind.mission': 'Mission à distance',
    'connect.row.by': 'Proposé par', 'connect.row.profile': 'Voir le profil de', 'connect.row.action': 'Prendre contact', 'connect.row.action.why': 'Bientôt : la mise en relation n’est pas disponible.',
    'connect.list.empty.title': 'Aucun exemple ne correspond', 'connect.list.empty.body': 'Retirez un filtre ou réinitialisez pour revoir tous les exemples.',
    'connect.spaces.h': 'Communautés et échanges', 'connect.spaces.badge': 'Plus tard — P2',
    'connect.spaces.lede': 'Espaces thématiques et fil d’activité sont décrits dans la feuille de route. Aucune discussion n’est affichée dans ce prototype.',
    'connect.space.by-country': 'Espaces par pays', 'connect.space.by-country.d': 'Retrouver les développeurs d’un même pays.',
    'connect.space.by-tech': 'Espaces par technologie', 'connect.space.by-tech.d': 'Échanger autour d’une même pile technique.',
    'connect.space.by-interest': 'Espaces par centre d’intérêt', 'connect.space.by-interest.d': 'Groupes thématiques, par exemple apprentissage ou open source.'
  },
  en: {
    'connect.soon.title': 'Coming soon', 'connect.soon.body': 'Finding peers, study groups and mentoring is on its way.', 'connect.soon.back': 'Back to home',
    'connect.title': 'Connect — 226os', 'connect.badge': 'Designed journey (P1/P2)',
    'connect.lede': 'Mentoring, CV reviews, interview preparation and study groups.',
    'connect.note.title': 'Designed journey (P1/P2), not functional. ',
    'connect.note.body': 'Per PRD §8, Connect’s scope is documented and its journey designed; the rest is planned for P1/P2. This page shows that designed journey, non-functional. The entries below are fictional examples showing the discovery pattern. No contact or matching is possible in this prototype.',
    'connect.disc.h': 'Discover mentors and collaborations', 'connect.disc.lede': 'Filter by skill, country and language. Contact actions come later.',
    'connect.filters.label': 'Filter Connect examples', 'connect.f.skill': 'Skill', 'connect.f.lang': 'Spoken language', 'connect.f.reset': 'Reset',
    'connect.f.skill.all': 'All', 'connect.f.country.all': 'All', 'connect.f.lang.all': 'All',
    'connect.lang.fr': 'French', 'connect.lang.en': 'English',
    'connect.count': '{n} example(s) shown',
    'connect.kind.mentor-offer': 'Mentoring offer', 'connect.kind.mentor-request': 'Mentoring request', 'connect.kind.cv-review': 'CV / portfolio review',
    'connect.kind.interview': 'Interview preparation', 'connect.kind.study-group': 'Study group', 'connect.kind.mission': 'Remote mission',
    'connect.row.by': 'Posted by', 'connect.row.profile': 'View profile of', 'connect.row.action': 'Get in touch', 'connect.row.action.why': 'Soon: contact is not available.',
    'connect.list.empty.title': 'No example matches', 'connect.list.empty.body': 'Remove a filter or reset to see all examples again.',
    'connect.spaces.h': 'Communities and discussions', 'connect.spaces.badge': 'Later — P2',
    'connect.spaces.lede': 'Topic spaces and an activity feed are described in the roadmap. No discussion is shown in this prototype.',
    'connect.space.by-country': 'Spaces by country', 'connect.space.by-country.d': 'Find developers from the same country.',
    'connect.space.by-tech': 'Spaces by technology', 'connect.space.by-tech.d': 'Exchange around a shared tech stack.',
    'connect.space.by-interest': 'Spaces by interest', 'connect.space.by-interest.d': 'Topic groups, for example learning or open source.'
  }
});

(function () {
  'use strict';
  var $ = App.$, esc = App.esc, t = function (k, p) { return I18N.t(k, p); }, C = MOCK.connect;
  var skills = Object.keys(C.entries.reduce(function (m, e) { e.skills.forEach(function (s) { m[s] = 1; }); return m; }, {})).sort();
  var countries = Object.keys(C.entries.reduce(function (m, e) { m[e.country] = 1; return m; }, {})).sort();
  var F = { skill: $('#cf-skill'), country: $('#cf-country'), lang: $('#cf-lang') };

  function fillSelects() {   /* relabel without losing the selection */
    var cur = { skill: F.skill.value, country: F.country.value, lang: F.lang.value };
    F.skill.innerHTML = '<option value="" data-testid="connect-skill-opt-all">' + esc(t('connect.f.skill.all')) + '</option>' + skills.map(function (s) { return '<option value="' + s + '" data-testid="connect-skill-opt-' + s + '">' + esc(s) + '</option>'; }).join('');
    F.country.innerHTML = '<option value="" data-testid="connect-country-opt-all">' + esc(t('connect.f.country.all')) + '</option>' + countries.map(function (c) { return '<option value="' + c + '" data-testid="connect-country-opt-' + c.toLowerCase() + '">' + esc(MOCK.countryName(c)) + '</option>'; }).join('');
    F.lang.innerHTML = '<option value="" data-testid="connect-lang-opt-all">' + esc(t('connect.f.lang.all')) + '</option>' + ['fr', 'en'].map(function (l) { return '<option value="' + l + '" data-testid="connect-lang-opt-' + l + '">' + esc(t('connect.lang.' + l)) + '</option>'; }).join('');
    F.skill.value = cur.skill; F.country.value = cur.country; F.lang.value = cur.lang;
  }
  fillSelects();

  function current() { return App.Filter.run(C.entries, { where: { skills: F.skill.value, country: F.country.value, speaks: F.lang.value } }); }
  var shown = 0;
  function setCount(n) { shown = n; $('#cf-count').innerHTML = '<span data-testid="connect-count-text">' + esc(t('connect.count', { n: n })) + '</span>'; }

  var sim = App.Sim.load($('#connect-list'), {
    delay: 600, skeleton: 3, emptyKey: 'connect.list.empty',
    load: function () { var r = current(); setCount(r.length); return r; },
    render: function (items, box) {
      box.innerHTML = '<ul class="list list-reset" data-testid="connect-list">' + items.map(function (e, i) {
        var n = e.id, per = MOCK.person(e.person), why = 'cx-why-' + n;
        return '<li class="row" data-testid="connect-row-' + n + '"><div class="row-lead"><span class="icon i-' + (e.kind === 'study-group' ? 'users' : e.kind === 'mission' ? 'globe' : 'network') + '" aria-hidden="true" data-testid="connect-row-icon-' + n + '"></span></div>' +
          '<div class="row-main"><div class="cluster"><span class="badge badge-later" data-testid="connect-kind-' + n + '">' + esc(t('connect.kind.' + e.kind)) + '</span><span class="badge badge-sim" data-proto data-testid="connect-example-' + n + '">' + esc(t('sim.badge')) + '</span></div>' +
          '<h3 class="row-title" lang="' + e.lang + '" data-user-content data-testid="connect-title-' + n + '">' + esc(e.title) + '</h3>' +
          '<ul class="tag-list list-reset" data-testid="connect-tags-' + n + '">' + e.skills.map(function (s) { return '<li><span class="tag" data-testid="connect-tag-' + n + '-' + s.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '') + '">' + esc(s) + '</span></li>'; }).join('') + '</ul>' +
          '<div class="meta"><span class="meta-item" data-testid="connect-by-' + n + '">' + esc(t('connect.row.by')) + '&nbsp;<a href="profile.html?id=' + per.id + '" lang="' + per.lang + '" data-user-content data-testid="connect-person-' + n + '">' + esc(per.name) + '</a></span>' +
          '<span class="meta-item" data-testid="connect-country-' + n + '"><span class="country"><span class="country-code" aria-hidden="true" data-testid="connect-country-code-' + n + '">' + e.country + '</span><span class="visually-hidden" data-testid="connect-country-name-' + n + '">' + esc(MOCK.countryName(e.country)) + '</span></span> ' + esc(MOCK.countryName(e.country)) + '</span>' +
          '<span class="meta-item" data-testid="connect-speaks-' + n + '">' + e.speaks.map(function (l) { return esc(t('connect.lang.' + l)); }).join(', ') + '</span></div></div>' +
          '<div class="row-aside"><button type="button" class="btn btn-sm" disabled aria-describedby="' + why + '" data-testid="connect-action-' + n + '">' + esc(t('connect.row.action')) + '</button><span class="visually-hidden" id="' + why + '" data-testid="connect-action-why-' + n + '">' + esc(t('connect.row.action.why')) + '</span></div></li>';
      }).join('') + '</ul>';
    }
  });
  var form = $('#connect-filters');
  ['change'].forEach(function (ev) { form.addEventListener(ev, function () { sim.refresh(); }); });
  form.addEventListener('reset', function () { setTimeout(function () { sim.refresh(); }, 0); });
  form.addEventListener('submit', function (e) { e.preventDefault(); });
  document.addEventListener('i18n:change', function () { fillSelects(); setCount(shown); });

  function spaces() {
    $('#connect-spaces-list').innerHTML = C.spaces.map(function (s) {
      var n = s.id;
      return '<li class="row" data-testid="connect-space-' + n + '"><div class="row-lead"><span class="icon i-' + s.icon + '" aria-hidden="true" data-testid="connect-space-icon-' + n + '"></span></div><div class="row-main"><h3 class="row-title" data-testid="connect-space-title-' + n + '">' + esc(t('connect.space.' + s.id)) + '</h3>' +
        '<p class="row-excerpt" data-testid="connect-space-desc-' + n + '">' + esc(t('connect.space.' + s.id + '.d')) + '</p></div><div class="row-aside"><span class="badge badge-later" data-testid="connect-space-badge-' + n + '">' + esc(t('common.later')) + '</span></div></li>';
    }).join('');
  }
  spaces(); document.addEventListener('i18n:change', spaces);
})();
