/* Developer profile page: view (per role), activity tabs (simulated load), edit dialog (client validation, nothing persisted). */
I18N.register({
  fr: {
    'profile.title': 'Profil développeur — 226os', 'profile.crumb': 'Profil développeur',
    'profile.lede': 'Profil public minimal : ce que les autres membres peuvent voir dépend des réglages de confidentialité.',
    'profile.edit.open': 'Modifier mon profil',
    'profile.notfound.title': 'Profil introuvable', 'profile.notfound.body': 'Ce profil n’existe pas ou n’est pas visible pour vous.', 'profile.notfound.cta': 'Voir mon profil',
    'profile.viewas.label': 'Voir ce profil en tant que', 'profile.viewas.self': 'Moi (propriétaire)', 'profile.viewas.member': 'Autre membre', 'profile.viewas.anon': 'Visiteur anonyme',
    'profile.viewas.status.self': 'Vue propriétaire : vous voyez tous vos champs et pouvez modifier le profil.',
    'profile.viewas.status.member': 'Vue autre membre connecté : champs publics et réservés aux membres.',
    'profile.viewas.status.anon': 'Vue visiteur anonyme : seulement les champs publics.',
    'profile.saved.title': 'Profil mis à jour. ', 'profile.saved.body': 'Simulation : rien n’est enregistré, les modifications disparaissent au rechargement.',
    'profile.anon.title': 'Vous naviguez sans compte. ', 'profile.anon.body': 'Certains champs sont réservés aux membres connectés.', 'profile.anon.signin': 'Se connecter',
    'profile.about.h': 'À propos', 'profile.bio.h': 'Bio', 'profile.skills.h': 'Compétences', 'profile.langs.h': 'Langues parlées', 'profile.open.h': 'Ouvert à', 'profile.open.mentorat': 'Mentorat', 'profile.open.collaboration': 'Collaboration', 'profile.open.preview': 'Aperçu Connect', 'profile.open.none': 'Aucune ouverture indiquée.',
    'profile.bio.empty': 'Aucune bio renseignée.', 'profile.skills.empty': 'Aucune compétence renseignée.', 'profile.bio.lang': 'Langue originale de la bio',
    'profile.lang.fr': 'Français', 'profile.lang.en': 'Anglais',
    'profile.details.h': 'Détails', 'profile.aside.label': 'Détails et confidentialité du profil',
    'profile.d.handle': 'Identifiant', 'profile.d.country': 'Pays', 'profile.d.city': 'Ville', 'profile.d.github': 'GitHub', 'profile.d.email': 'E-mail', 'profile.d.since': 'Membre depuis', 'profile.d.name': 'Nom public',
    'profile.d.hidden': 'Non visible pour ce rôle', 'profile.github.link': 'Dépôts GitHub (exemple fictif)',
    'profile.privacy.h': 'Confidentialité', 'profile.privacy.region': 'Tableau de visibilité des champs du profil',
    'profile.privacy.note': 'L’e-mail n’est jamais public. L’autorisation réelle est vérifiée côté serveur ; ici c’est une démonstration.',
    'profile.privacy.col.field': 'Champ', 'profile.privacy.col.who': 'Visible par', 'profile.privacy.col.now': 'Pour vous',
    'profile.vis.public': 'Tout le monde', 'profile.vis.members': 'Membres connectés', 'profile.vis.self': 'Vous uniquement',
    'profile.now.visible': 'Visible', 'profile.now.hidden': 'Masqué',
    'profile.activity.h': 'Activité', 'profile.tab.q': 'Questions', 'profile.tab.a': 'Réponses', 'profile.tab.p': 'Projets', 'profile.tab.c': 'Contributions',
    'profile.activity.q.empty.title': 'Aucune question publiée', 'profile.activity.q.empty.body': 'Ce membre n’a pas encore posé de question.',
    'profile.activity.a.empty.title': 'Aucune réponse publiée', 'profile.activity.a.empty.body': 'Ce membre n’a pas encore répondu à une question.',
    'profile.activity.p.empty.title': 'Aucun projet', 'profile.activity.p.empty.body': 'Ce membre ne possède aucun projet référencé.',
    'profile.activity.c.empty.title': 'Aucune contribution', 'profile.activity.c.empty.body': 'Ce membre n’a rejoint aucun projet.',
    'profile.row.question': 'Question', 'profile.row.answered': 'Réponse donnée', 'profile.row.project': 'Projet', 'profile.row.contrib': 'Contribution',
    'profile.status.active': 'Actif', 'profile.status.idea': 'Idée', 'profile.status.seeking': 'Cherche des contributeurs',
    'profile.reputation': 'Réputation et badges', 'profile.reputation.note': 'Prévus en P1 : aucun score n’est affiché dans ce prototype.',
    'profile.edit.title': 'Modifier mon profil', 'profile.edit.sim': 'Les modifications sont validées puis affichées localement, jamais enregistrées.',
    'profile.edit.summary': 'Corrigez les champs suivants',
    'profile.edit.name': 'Nom public', 'profile.edit.name.hint': '2 à 60 caractères. Visible par tout le monde.', 'profile.edit.name.err': 'Saisissez un nom public de 2 à 60 caractères.',
    'profile.edit.country': 'Pays', 'profile.edit.country.hint': 'Utilisé pour la découverte par pays.', 'profile.edit.country.err': 'Choisissez un pays dans la liste.', 'profile.edit.country.placeholder': 'Choisir…',
    'profile.edit.langs': 'Langues parlées', 'profile.edit.langs.hint': 'Au moins une langue.', 'profile.edit.langs.err': 'Sélectionnez au moins une langue.',
    'profile.edit.skills': 'Compétences', 'profile.edit.skills.hint': 'Séparées par des virgules, 8 maximum (ex. go, postgresql).', 'profile.edit.skills.err': 'Maximum 8 compétences, 24 caractères chacune (lettres, chiffres, + # . -).',
    'profile.edit.bio': 'Bio', 'profile.edit.bio.hint': '280 caractères maximum. Votre texte garde sa langue d’origine.', 'profile.edit.bio.err': 'La bio dépasse 280 caractères.',
    'profile.edit.biolang': 'Langue de la bio',
    'profile.edit.github': 'Lien GitHub', 'profile.edit.github.hint': 'Adresse https. Dans ce prototype, utilisez un exemple : https://example.org/gh/votre-id', 'profile.edit.github.err': 'Saisissez une adresse https valide ou laissez vide.',
    'profile.edit.citypub': 'Afficher ma ville aux visiteurs anonymes', 'profile.edit.citypub.hint': 'Par défaut, la ville n’est visible que des membres connectés.',
    'profile.edit.save': 'Enregistrer (simulé)', 'profile.edit.saving': 'Enregistrement…', 'profile.edit.toast': 'Profil mis à jour (simulation).'
  },
  en: {
    'profile.title': 'Developer profile — 226os', 'profile.crumb': 'Developer profile',
    'profile.lede': 'Minimal public profile: what other members can see depends on privacy settings.',
    'profile.edit.open': 'Edit my profile',
    'profile.notfound.title': 'Profile not found', 'profile.notfound.body': 'This profile does not exist or is not visible to you.', 'profile.notfound.cta': 'View my profile',
    'profile.viewas.label': 'View this profile as', 'profile.viewas.self': 'Me (owner)', 'profile.viewas.member': 'Other member', 'profile.viewas.anon': 'Anonymous visitor',
    'profile.viewas.status.self': 'Owner view: you see all your fields and can edit the profile.',
    'profile.viewas.status.member': 'Other signed-in member view: public and members-only fields.',
    'profile.viewas.status.anon': 'Anonymous visitor view: public fields only.',
    'profile.saved.title': 'Profile updated. ', 'profile.saved.body': 'Simulation: nothing is saved, changes disappear on reload.',
    'profile.anon.title': 'You are browsing without an account. ', 'profile.anon.body': 'Some fields are reserved for signed-in members.', 'profile.anon.signin': 'Sign in',
    'profile.about.h': 'About', 'profile.bio.h': 'Bio', 'profile.skills.h': 'Skills', 'profile.langs.h': 'Spoken languages', 'profile.open.h': 'Open to', 'profile.open.mentorat': 'Mentoring', 'profile.open.collaboration': 'Collaboration', 'profile.open.preview': 'Connect preview', 'profile.open.none': 'No openness stated.',
    'profile.bio.empty': 'No bio provided.', 'profile.skills.empty': 'No skills provided.', 'profile.bio.lang': 'Original language of the bio',
    'profile.lang.fr': 'French', 'profile.lang.en': 'English',
    'profile.details.h': 'Details', 'profile.aside.label': 'Profile details and privacy',
    'profile.d.handle': 'Handle', 'profile.d.country': 'Country', 'profile.d.city': 'City', 'profile.d.github': 'GitHub', 'profile.d.email': 'Email', 'profile.d.since': 'Member since', 'profile.d.name': 'Public name',
    'profile.d.hidden': 'Not visible for this role', 'profile.github.link': 'GitHub repositories (fictional example)',
    'profile.privacy.h': 'Privacy', 'profile.privacy.region': 'Profile field visibility table',
    'profile.privacy.note': 'Email is never public. Real authorization is enforced on the server; this is a demonstration.',
    'profile.privacy.col.field': 'Field', 'profile.privacy.col.who': 'Visible to', 'profile.privacy.col.now': 'For you',
    'profile.vis.public': 'Everyone', 'profile.vis.members': 'Signed-in members', 'profile.vis.self': 'You only',
    'profile.now.visible': 'Visible', 'profile.now.hidden': 'Hidden',
    'profile.activity.h': 'Activity', 'profile.tab.q': 'Questions', 'profile.tab.a': 'Answers', 'profile.tab.p': 'Projects', 'profile.tab.c': 'Contributions',
    'profile.activity.q.empty.title': 'No questions published', 'profile.activity.q.empty.body': 'This member has not asked a question yet.',
    'profile.activity.a.empty.title': 'No answers published', 'profile.activity.a.empty.body': 'This member has not answered a question yet.',
    'profile.activity.p.empty.title': 'No projects', 'profile.activity.p.empty.body': 'This member owns no listed project.',
    'profile.activity.c.empty.title': 'No contributions', 'profile.activity.c.empty.body': 'This member has not joined any project.',
    'profile.row.question': 'Question', 'profile.row.answered': 'Answer given', 'profile.row.project': 'Project', 'profile.row.contrib': 'Contribution',
    'profile.status.active': 'Active', 'profile.status.idea': 'Idea', 'profile.status.seeking': 'Seeking contributors',
    'profile.reputation': 'Reputation and badges', 'profile.reputation.note': 'Planned for P1: no score is shown in this prototype.',
    'profile.edit.title': 'Edit my profile', 'profile.edit.sim': 'Changes are validated and shown locally, never saved.',
    'profile.edit.summary': 'Please fix the following fields',
    'profile.edit.name': 'Public name', 'profile.edit.name.hint': '2 to 60 characters. Visible to everyone.', 'profile.edit.name.err': 'Enter a public name of 2 to 60 characters.',
    'profile.edit.country': 'Country', 'profile.edit.country.hint': 'Used for discovery by country.', 'profile.edit.country.err': 'Choose a country from the list.', 'profile.edit.country.placeholder': 'Choose…',
    'profile.edit.langs': 'Spoken languages', 'profile.edit.langs.hint': 'At least one language.', 'profile.edit.langs.err': 'Select at least one language.',
    'profile.edit.skills': 'Skills', 'profile.edit.skills.hint': 'Comma-separated, 8 at most (e.g. go, postgresql).', 'profile.edit.skills.err': 'At most 8 skills, 24 characters each (letters, digits, + # . -).',
    'profile.edit.bio': 'Bio', 'profile.edit.bio.hint': '280 characters maximum. Your text keeps its original language.', 'profile.edit.bio.err': 'The bio exceeds 280 characters.',
    'profile.edit.biolang': 'Bio language',
    'profile.edit.github': 'GitHub link', 'profile.edit.github.hint': 'https address. In this prototype use an example: https://example.org/gh/your-id', 'profile.edit.github.err': 'Enter a valid https address or leave empty.',
    'profile.edit.citypub': 'Show my city to anonymous visitors', 'profile.edit.citypub.hint': 'By default the city is visible to signed-in members only.',
    'profile.edit.save': 'Save (simulated)', 'profile.edit.saving': 'Saving…', 'profile.edit.toast': 'Profile updated (simulation).'
  }
});

(function () {
  'use strict';
  var $ = App.$, esc = App.esc, t = function (k, p) { return I18N.t(k, p); };
  var C = MOCK.connect, P = App.params;
  var id = P.get('id') || C.selfId, base = MOCK.person(id);
  var $content = $('#profile-content'), $nf = $('#profile-notfound');
  if (!base) { $content.hidden = true; $nf.hidden = false; $('#profile-name').textContent = '—'; $('#profile-edit-open').hidden = true; I18N.apply(); document.title = t('profile.title'); return; }

  var isOwn = id === C.selfId, extra = C.profiles[id] || { bioLang: base.lang, bio: '', github: '', email: '' };
  /* local mutable copy: edits are simulated and never persisted */
  var p = { id: id, name: base.name, handle: base.handle, country: base.country, city: base.city, langs: base.langs.slice(), skills: base.skills.slice(), open: base.open.slice(), since: base.since,
    bio: extra.bio, bioLang: extra.bioLang, github: extra.github, email: extra.email, cityPublic: false };
  var roles = isOwn ? ['self', 'member', 'anon'] : ['member', 'anon'];
  var role = P.get('as'); if (roles.indexOf(role) < 0) role = isOwn ? 'self' : 'member';

  var initials = App.initials;   /* shared helper (app.js) */
  function canSee(field) {
    var v = C.visibility[field];
    if (field === 'city' && p.cityPublic) v = 'public';
    return v === 'public' || (v === 'members' && role !== 'anon') || (v === 'self' && role === 'self');
  }
  function lbadge(lang) { return '<span class="badge" data-orig-badge><span aria-hidden="true">' + lang.toUpperCase() + '</span><span class="visually-hidden">' + esc(t('common.originalLang')) + ' : ' + esc(t('profile.lang.' + lang)) + '</span></span>'; }
  function countryHtml(code, tid) { return '<span class="country" data-testid="' + tid + '"><span class="country-code" aria-hidden="true">' + code + '</span><span class="visually-hidden">' + esc(MOCK.countryName(code)) + '</span></span> ' + esc(MOCK.countryName(code)); }

  /* ---------- overview ---------- */
  function renderOverview() {
    $('#profile-avatar').textContent = initials(p.name);
    var nm = $('#profile-name'); nm.textContent = p.name; nm.lang = base.lang;
    $('#profile-sub').innerHTML = '<span class="meta-item mono" data-testid="profile-handle">@' + esc(p.handle) + '</span><span class="meta-item" data-testid="profile-country">' + countryHtml(p.country, 'profile-country-code') + '</span>';
    $('#profile-edit-open').hidden = role !== 'self';
    $('#profile-anon-note').hidden = role !== 'anon';
    $('#profile-viewas-status').textContent = t('profile.viewas.status.' + role);
    App.$$('[data-as]').forEach(function (b) { var on = b.getAttribute('data-as') === role; b.setAttribute('aria-pressed', on); b.hidden = roles.indexOf(b.getAttribute('data-as')) < 0; });

    var openHtml = p.open.length ? '<ul class="tag-list list-reset" data-testid="profile-open-list">' + p.open.map(function (o, i) {
      return '<li><a class="tag" href="connect.html" data-testid="profile-open-' + (i + 1) + '">' + esc(t('profile.open.' + o)) + '</a></li>'; }).join('') + '</ul><p class="hint" data-testid="profile-open-note"><span class="badge badge-later" data-testid="profile-open-badge">' + esc(t('profile.open.preview')) + '</span></p>'
      : '<p class="text-muted" data-testid="profile-open-none">' + esc(t('profile.open.none')) + '</p>';
    $('#profile-about').innerHTML =
      '<div class="stack-sm"><h3 class="eyebrow" data-testid="profile-bio-h">' + esc(t('profile.bio.h')) + '</h3>' + (p.bio
        ? '<p class="prose" lang="' + p.bioLang + '" data-user-content data-testid="profile-bio">' + esc(p.bio) + '</p><p class="text-sm text-muted" data-testid="profile-bio-lang">' + lbadge(p.bioLang) + ' ' + esc(t('profile.bio.lang')) + '</p>'
        : '<p class="text-muted" data-testid="profile-bio-empty">' + esc(t('profile.bio.empty')) + '</p>') + '</div>' +
      '<div class="stack-sm"><h3 class="eyebrow" data-testid="profile-skills-h">' + esc(t('profile.skills.h')) + '</h3>' + (p.skills.length
        ? '<ul class="tag-list list-reset" data-testid="profile-skills">' + p.skills.map(function (s, i) { return '<li><span class="tag" data-testid="profile-skill-' + (i + 1) + '">' + esc(s) + '</span></li>'; }).join('') + '</ul>'
        : '<p class="text-muted" data-testid="profile-skills-empty">' + esc(t('profile.skills.empty')) + '</p>') + '</div>' +
      '<div class="stack-sm"><h3 class="eyebrow" data-testid="profile-langs-h">' + esc(t('profile.langs.h')) + '</h3><ul class="tag-list list-reset" data-testid="profile-langs">' +
        p.langs.map(function (l, i) { return '<li><span class="tag" data-testid="profile-lang-' + (i + 1) + '"><span aria-hidden="true">' + l.toUpperCase() + '</span><span class="visually-hidden">' + esc(t('profile.lang.' + l)) + '</span></span></li>'; }).join('') + '</ul></div>' +
      '<div class="stack-sm"><h3 class="eyebrow" data-testid="profile-open-h">' + esc(t('profile.open.h')) + '</h3>' + openHtml + '</div>' +
      '<p class="text-sm text-muted" data-testid="profile-reputation"><span class="badge badge-later" data-testid="profile-reputation-badge">' + esc(t('common.later')) + '</span> <strong>' + esc(t('profile.reputation')) + '</strong> — ' + esc(t('profile.reputation.note')) + '</p>';

    var rows = [];
    function dd(key, field, val, tid) { rows.push('<dt data-testid="profile-d-' + tid + '-label">' + esc(t('profile.d.' + key)) + '</dt><dd data-testid="profile-d-' + tid + '">' + (canSee(field) ? val : '<span class="text-muted"><span class="icon i-lock" aria-hidden="true"></span> ' + esc(t('profile.d.hidden')) + '</span>') + '</dd>'); }
    dd('name', 'name', '<span lang="' + base.lang + '" data-user-content>' + esc(p.name) + '</span>', 'name');
    dd('handle', 'handle', '<span class="mono">@' + esc(p.handle) + '</span>', 'handle');
    dd('country', 'country', countryHtml(p.country, 'profile-d-country-code'), 'country');
    dd('city', 'city', esc(p.city), 'city');
    dd('github', 'github', p.github ? '<a class="link-external" href="' + esc(p.github) + '" rel="noopener noreferrer" target="_blank" data-testid="profile-github-link">' + esc(t('profile.github.link')) + '</a>' : '—', 'github');
    dd('email', 'email', '<span class="mono">' + esc(p.email) + '</span>', 'email');
    dd('since', 'since', '<time datetime="' + p.since + '">' + esc(MOCK.fmtDate(p.since)) + '</time>', 'since');
    $('#profile-details').innerHTML = rows.join('');

    var order = ['name', 'handle', 'country', 'langs', 'skills', 'bio', 'github', 'openTo', 'since', 'city', 'email'];
    var lab = { name: 'name', handle: 'handle', country: 'country', langs: 'langs.h', skills: 'skills.h', bio: 'bio.h', github: 'github', openTo: 'open.h', since: 'since', city: 'city', email: 'email' };
    $('#profile-privacy-body').innerHTML = order.map(function (f) {
      var ok = canSee(f), v = C.visibility[f] === 'members' && f === 'city' && p.cityPublic ? 'public' : C.visibility[f];
      return '<tr data-testid="profile-privacy-row-' + f + '"><th scope="row" data-testid="profile-privacy-field-' + f + '">' + esc(t(f === 'langs' || f === 'skills' || f === 'bio' || f === 'openTo' ? 'profile.' + lab[f] : 'profile.d.' + lab[f])) + '</th>' +
        '<td data-testid="profile-privacy-who-' + f + '">' + esc(t('profile.vis.' + v)) + '</td>' +
        '<td data-testid="profile-privacy-now-' + f + '"><span class="profile-state"><span class="icon ' + (ok ? 'i-check-circle' : 'i-lock') + '" aria-hidden="true"></span>' + esc(t(ok ? 'profile.now.visible' : 'profile.now.hidden')) + '</span></td></tr>';
    }).join('');
    document.title = p.name + ' — ' + t('profile.title');
  }

  /* ---------- activity (simulated async, one marker in HTML) ---------- */
  function rowQ(q, n, pre, kindKey) {
    var a = MOCK.person(q.author);
    return '<li class="row" data-testid="profile-' + pre + '-row-' + n + '"><div class="row-lead"><span class="icon i-help" aria-hidden="true" data-testid="profile-' + pre + '-icon-' + n + '"></span></div>' +
      '<div class="row-main"><h3 class="row-title"><a href="question.html?id=' + q.id + '" lang="' + q.lang + '" data-user-content data-testid="profile-' + pre + '-link-' + n + '">' + esc(q.title) + '</a></h3>' +
      '<ul class="tag-list list-reset" data-testid="profile-' + pre + '-tags-' + n + '">' + q.tags.map(function (g, j) { return '<li><span class="tag" data-testid="profile-' + pre + '-tag-' + n + '-' + (j + 1) + '">' + esc(g) + '</span></li>'; }).join('') + '</ul>' +
      '<div class="meta"><span class="meta-item" data-testid="profile-' + pre + '-kind-' + n + '">' + esc(t(kindKey)) + '</span><span class="meta-item" data-testid="profile-' + pre + '-country-' + n + '">' + countryHtml(q.country, 'profile-' + pre + '-cc-' + n) + '</span><time datetime="' + q.date + '" data-testid="profile-' + pre + '-date-' + n + '">' + esc(MOCK.fmtDate(q.date)) + '</time>' + (pre === 'a' && a ? '<span class="meta-item" lang="' + a.lang + '" data-user-content data-testid="profile-a-author-' + n + '">' + esc(a.name) + '</span>' : '') + '</div></div>' +
      '<div class="row-aside">' + lbadge(q.lang) + (q.accepted ? '<span class="badge badge-accepted" data-testid="profile-' + pre + '-accepted-' + n + '"><span class="icon i-check" aria-hidden="true"></span>' + esc(t('common.accepted')) + '</span>' : '') + '</div></li>';
  }
  function rowP(pr, n, pre, kindKey) {
    return '<li class="row" data-testid="profile-' + pre + '-row-' + n + '"><div class="row-lead"><span class="icon i-hammer" aria-hidden="true" data-testid="profile-' + pre + '-icon-' + n + '"></span></div>' +
      '<div class="row-main"><h3 class="row-title"><a href="project.html?id=' + pr.id + '" data-testid="profile-' + pre + '-link-' + n + '"><span class="mono">' + esc(pr.name) + '</span></a></h3>' +
      '<p class="row-excerpt" lang="' + pr.lang + '" data-user-content data-testid="profile-' + pre + '-excerpt-' + n + '">' + esc(pr.summary) + '</p>' +
      '<ul class="tag-list list-reset" data-testid="profile-' + pre + '-tags-' + n + '">' + pr.stack.map(function (g, j) { return '<li><span class="tag" data-testid="profile-' + pre + '-tag-' + n + '-' + (j + 1) + '">' + esc(g) + '</span></li>'; }).join('') + '</ul>' +
      '<div class="meta"><span class="meta-item" data-testid="profile-' + pre + '-kind-' + n + '">' + esc(t(kindKey)) + '</span><span class="meta-item" data-testid="profile-' + pre + '-country-' + n + '">' + countryHtml(pr.country, 'profile-' + pre + '-cc-' + n) + '</span><time datetime="' + pr.updated + '" data-testid="profile-' + pre + '-date-' + n + '">' + esc(MOCK.fmtDate(pr.updated)) + '</time></div></div>' +
      '<div class="row-aside">' + lbadge(pr.lang) + '<span class="badge" data-testid="profile-' + pre + '-status-' + n + '">' + esc(t('profile.status.' + pr.status)) + '</span></div></li>';
  }
  function byIds(arr, ids) { return ids.map(function (i) { return MOCK.byId(arr, i); }).filter(Boolean); }
  function list(pre, builder, kindKey) {
    return function (items, box) { box.innerHTML = '<ul class="list list-reset" data-testid="profile-' + pre + '-list">' + items.map(function (x, i) { return builder(x, i + 1, pre, kindKey); }).join('') + '</ul>'; };
  }
  [['q', 'act-q', function () { return MOCK.questions.filter(function (q) { return q.author === id; }); }, list('q', rowQ, 'profile.row.question')],
   ['a', 'act-a', function () { return byIds(MOCK.questions, C.answers[id] || []); }, list('a', rowQ, 'profile.row.answered')],
   ['p', 'act-p', function () { return MOCK.projects.filter(function (x) { return x.owner === id; }); }, list('p', rowP, 'profile.row.project')],
   ['c', 'act-c', function () { return byIds(MOCK.projects, C.contributions[id] || []); }, list('c', rowP, 'profile.row.contrib')]
  ].forEach(function (a) { App.Sim.load($('#' + a[1]), { load: a[2], render: a[3], delay: 450, skeleton: 2, emptyKey: 'profile.activity.' + a[0] + '.empty' }); });

  /* ---------- view-as toggle ---------- */
  App.$$('[data-as]').forEach(function (b) { b.addEventListener('click', function () { role = b.getAttribute('data-as'); renderOverview(); }); });
  document.addEventListener('i18n:change', function () { populateSelects(); renderOverview(); });

  /* ---------- edit dialog ---------- */
  var dlg = $('#profile-edit-dialog'), form = $('#profile-edit-form'), sumBox = $('#pe-summary'), sumList = $('#pe-summary-list');
  var F = { name: $('#pe-name'), country: $('#pe-country'), skills: $('#pe-skills'), bio: $('#pe-bio'), biolang: $('#pe-biolang'), github: $('#pe-github'), citypub: $('#pe-citypub') };
  var langBoxes = App.$$('input[name="langs"]', form);
  function populateSelects() {
    var cur = F.country.value || p.country, h = '<option value="" data-testid="profile-edit-country-placeholder">' + esc(t('profile.edit.country.placeholder')) + '</option>';
    h += MOCK.countries.map(function (c) { return '<option value="' + c.code + '" data-testid="profile-edit-country-opt-' + c.code.toLowerCase() + '">' + esc(c[I18N.lang]) + '</option>'; }).join('');
    F.country.innerHTML = h; F.country.value = cur;
  }
  populateSelects();
  function fillForm() {
    F.name.value = p.name; F.country.value = p.country; F.skills.value = p.skills.join(', '); F.bio.value = p.bio; F.biolang.value = p.bioLang; F.github.value = p.github; F.citypub.checked = p.cityPublic;
    langBoxes.forEach(function (b) { b.checked = p.langs.indexOf(b.value) > -1; });
    clearErrors();
  }
  function clearErrors() {
    sumBox.hidden = true; sumList.innerHTML = '';
    App.$$('.field', form).forEach(function (f) {
      f.classList.remove('has-error'); var e = f.querySelector('.error-msg'); if (e) { e.hidden = true; e.innerHTML = ''; }
      App.$$('input,select,textarea', f).forEach(function (c) { c.removeAttribute('aria-invalid'); var d = (c.getAttribute('aria-describedby') || '').split(' ').filter(function (x) { return x && !/-error$/.test(x); }); c.setAttribute('aria-describedby', d.join(' ')); if (!d.length) c.removeAttribute('aria-describedby'); });
    });
  }
  App.$('#profile-edit-open').addEventListener('click', fillForm);
  form.addEventListener('submit', function (e) {
    e.preventDefault(); clearErrors();
    var errs = [], v = {};
    v.name = F.name.value.trim(); if (v.name.length < 2 || v.name.length > 60) errs.push(['name', 'pe-name', 'profile.edit.name.err', F.name]);
    v.country = F.country.value; if (!MOCK.byId(MOCK.countries.map(function (c) { return { id: c.code }; }), v.country)) errs.push(['country', 'pe-country', 'profile.edit.country.err', F.country]);
    v.langs = langBoxes.filter(function (b) { return b.checked; }).map(function (b) { return b.value; }); if (!v.langs.length) errs.push(['langs', 'pe-langs', 'profile.edit.langs.err', langBoxes[0]]);
    v.skills = F.skills.value.split(',').map(function (s) { return s.trim().toLowerCase(); }).filter(Boolean);
    if (v.skills.length > 8 || v.skills.some(function (s) { return !/^[a-z0-9.+#-]{1,24}$/.test(s); })) errs.push(['skills', 'pe-skills', 'profile.edit.skills.err', F.skills]);
    v.bio = F.bio.value.trim(); if (v.bio.length > 280) errs.push(['bio', 'pe-bio', 'profile.edit.bio.err', F.bio]);
    v.github = F.github.value.trim();
    if (v.github) { try { var u = new URL(v.github); if (u.protocol !== 'https:') throw 0; } catch (x) { errs.push(['github', 'pe-github', 'profile.edit.github.err', F.github]); } }
    if (errs.length) {
      errs.forEach(function (er, i) {
        var f = $('#pe-f-' + er[0]), msg = $('#pe-' + er[0] + '-error'); f.classList.add('has-error');
        msg.innerHTML = '<span class="icon i-alert-circle" aria-hidden="true"></span><span>' + esc(t(er[2])) + '</span>'; msg.hidden = false;
        var c = er[3]; c.setAttribute('aria-invalid', 'true'); c.setAttribute('aria-describedby', ((c.getAttribute('aria-describedby') || '') + ' pe-' + er[0] + '-error').trim());
        sumList.insertAdjacentHTML('beforeend', '<li><a href="#' + er[1] + '" data-testid="profile-edit-summary-link-' + (i + 1) + '">' + esc(t(er[2])) + '</a></li>');
      });
      sumBox.hidden = false; sumBox.focus(); return;
    }
    var btn = $('#pe-submit'); btn.setAttribute('aria-busy', 'true'); btn.disabled = true; btn.querySelector('span:last-child').textContent = t('profile.edit.saving');
    setTimeout(function () {
      p.name = v.name; p.country = v.country; p.langs = v.langs; p.skills = v.skills; p.bio = v.bio; p.bioLang = F.biolang.value; p.github = v.github; p.cityPublic = F.citypub.checked;
      btn.removeAttribute('aria-busy'); btn.disabled = false; btn.querySelector('span:last-child').textContent = t('profile.edit.save');
      App.Dialog.close(dlg); $('#profile-saved').hidden = false; renderOverview();
      App.Toast.show({ kind: 'success', title: t('state.success.title'), msg: t('profile.edit.toast') });
    }, 500);
  });
  sumList.addEventListener('click', function (e) { var a = e.target.closest('a'); if (a) { e.preventDefault(); var c = document.getElementById(a.getAttribute('href').slice(1)); if (c) c.focus(); } });

  renderOverview();
  if (P.get('edit') === '1' && role === 'self') { fillForm(); App.Dialog.open('profile-edit-dialog', $('#profile-edit-open')); if (P.get('bad') === '1') { F.name.value = ''; langBoxes.forEach(function (b) { b.checked = false; }); F.github.value = 'http://x'; form.requestSubmit(); } }
})();
