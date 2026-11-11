/* Project detail + application flow (simulated). Dictionary domain: project. Roles demo: owner / applicant / other member / anonymous. */
I18N.register({
  fr: {
    'project.title': 'Projet — 226os', 'project.notfound': 'Projet introuvable', 'project.jump': 'Voir les besoins', 'project.form.summary': 'Corrigez les champs suivants',
    'project.demo.title': 'Rôles (démo)', 'project.demo.note': 'En production, le rôle vient de la session et le serveur applique chaque règle. Ici, choisissez un rôle pour comparer.',
    'project.demo.viewas': 'Rôle simulé', 'project.demo.dup': 'Simuler une seconde candidature', 'project.demo.dup.none': 'Disponible quand le rôle choisi a déjà une candidature.',
    'project.demo.dup.ready': 'Rouvre le formulaire pour un besoin déjà postulé : le serveur doit refuser le doublon.',
    'project.role.owner': 'Mainteneur', 'project.role.applicant': 'Candidat', 'project.role.member': 'Autre membre', 'project.role.anonymous': 'Visiteur non connecté',
    'project.role.desc.owner': 'Mainteneur : lit les candidatures de son projet et les accepte ou les refuse. Ne peut pas postuler à son propre projet.',
    'project.role.desc.applicant': 'Membre connecté avec une candidature : voit uniquement ses propres candidatures. Une seconde candidature au même besoin est refusée.',
    'project.role.desc.member': 'Membre connecté sans candidature : peut candidater à un besoin ouvert. La liste privée des candidatures lui est refusée (403).',
    'project.role.desc.anonymous': 'Visiteur non connecté : lit la fiche publique, mais ne peut ni candidater ni lire les candidatures.',
    'project.badge.license': 'Licence {l}', 'project.badge.fake': 'Lien fictif',
    'project.ctx.title': 'Contexte lié', 'project.ctx.lede': 'La question d’origine et l’API concernée, à un clic.',
    'project.ctx.question': 'Question d’origine', 'project.ctx.resource': 'API / outil lié', 'project.ctx.maintainer': 'Mainteneur', 'project.ctx.none': 'Aucune question ni API n’est liée à ce projet pour l’instant.',
    'project.desc.title': 'Description', 'project.needs.title': 'Besoins de contribution', 'project.needs.skills': 'Compétences requises', 'project.needs.level': 'Niveau',
    'project.need.open': 'Ouvert', 'project.need.closed': 'Fermé', 'project.need.apply': 'Postuler', 'project.need.apply.aria': 'Postuler au besoin : {need}',
    'project.need.signin': 'Se connecter pour postuler', 'project.need.mine': 'Votre projet', 'project.need.sent': 'Candidature envoyée', 'project.need.apps': '{n} candidature(s)',
    'project.apps.title.owner': 'Candidatures reçues', 'project.apps.title.other': 'Candidatures', 'project.apps.mine': 'Mes candidatures',
    'project.apps.private': 'Liste privée, visible par le mainteneur uniquement.',
    'project.apps.denied.title': 'Accès refusé (403)', 'project.apps.denied.body': 'Seul le mainteneur peut lire la liste des candidatures. Le serveur refuse la demande et ne révèle ni données ni nombre.',
    'project.apps.anon.title': 'Connexion requise (401)', 'project.apps.anon.body': 'Connectez-vous pour postuler. La liste des candidatures reste réservée au responsable.',
    'project.apps.empty.title': 'Aucune candidature pour le moment', 'project.apps.empty.body': 'Les candidatures apparaîtront ici dès qu’un membre manifestera son intérêt.',
    'project.app.for': 'Besoin :', 'project.app.skills': 'Compétences :', 'project.app.avail': 'Disponibilité :', 'project.app.accept': 'Accepter', 'project.app.decline': 'Refuser',
    'project.app.accept.aria': 'Accepter la candidature de {name}', 'project.app.decline.aria': 'Refuser la candidature de {name}', 'project.app.decided': 'Décision enregistrée.',
    'project.status.pending': 'En attente', 'project.status.accepted': 'Acceptée', 'project.status.declined': 'Refusée', 'project.apps.toast.accepted': 'Candidature acceptée.', 'project.apps.toast.declined': 'Candidature refusée.',
    'project.avail.lt2': 'Moins de 2 h par semaine', 'project.avail.2to5': '2 à 5 h par semaine', 'project.avail.5to10': '5 à 10 h par semaine', 'project.avail.gt10': 'Plus de 10 h par semaine', 'project.avail.choose': 'Non précisée',
    'project.act.title': 'Activité', 'project.act.created': '{who} a créé le projet.', 'project.act.repo': '{who} a mis à jour la fiche du dépôt.', 'project.act.need': '{who} a publié un besoin :',
    'project.act.need-closed': '{who} a fermé le besoin :', 'project.act.link': '{who} a lié la question :',
    'project.side.repo': 'Dépôt', 'project.side.repo.link': 'Ouvrir le dépôt (lien fictif)', 'project.side.license': 'Licence', 'project.side.stage': 'Étape', 'project.side.country': 'Pays', 'project.side.langs': 'Langues de travail', 'project.side.updated': 'Mis à jour',
    'project.side.stack': 'Technologies', 'project.side.maint': 'Mainteneur',
    'project.lang.fr': 'Français', 'project.lang.en': 'English',
    'project.empty.title': 'Projet introuvable', 'project.empty.body': 'Aucun projet ne porte cet identifiant. Retournez à la liste.', 'project.empty.cta': 'Retour aux projets',
    'project.apply.title': 'Postuler pour contribuer', 'project.apply.need': 'Besoin :', 'project.apply.sim': 'Simulation : aucune candidature n’est envoyée.',
    'project.apply.motivation': 'Motivation', 'project.apply.motivation.hint': '30 à 600 caractères. Dites pourquoi ce besoin et ce que vous apporterez en premier.',
    'project.apply.skills': 'Compétences pertinentes', 'project.apply.skills.hint': 'Séparées par des virgules, ex. Go, PostgreSQL.', 'project.apply.avail': 'Disponibilité', 'project.apply.submit': 'Envoyer la candidature',
    'project.apply.ok.title': 'Candidature envoyée', 'project.apply.ok.body': 'Le mainteneur pourra l’accepter ou la refuser.',
    'project.apply.dup.title': 'Candidature déjà envoyée (409). ', 'project.apply.dup.body': 'Vous avez déjà postulé à ce besoin : aucune nouvelle demande n’a été créée.',
    'project.apply.toast': 'Candidature envoyée.',
    'project.err.motivation.required': 'Expliquez votre motivation.', 'project.err.motivation.short': 'La motivation doit contenir au moins 30 caractères.',
    'project.err.skills.required': 'Indiquez au moins une compétence.', 'project.err.avail.required': 'Choisissez une disponibilité.',
    'project.demo.sample.motivation': 'Je souhaite contribuer à ce besoin et peux commencer par une petite tâche.', 'project.demo.sample.skills': 'Go, PostgreSQL',
    'project.pub.btn': 'Publier un besoin', 'project.pub.title': 'Publier un besoin', 'project.pub.sim': 'Simulation : le besoin est ajouté à la liste, rien n’est enregistré.',
    'project.pub.t': 'Titre du besoin', 'project.pub.t.hint': 'Une phrase précise, 8 à 100 caractères.', 'project.pub.skills': 'Compétences requises', 'project.pub.skills.hint': 'Choisissez au moins une compétence.',
    'project.pub.level': 'Niveau attendu', 'project.pub.level.choose': 'Choisir un niveau', 'project.pub.d': 'Description', 'project.pub.d.hint': 'Ce que le contributeur fera, 30 caractères minimum.',
    'project.pub.submit': 'Publier le besoin', 'project.pub.toast': 'Besoin publié.', 'project.pub.summary': 'Corrigez les champs suivants',
    'project.pub.err.t': 'Saisissez un titre de 8 à 100 caractères.', 'project.pub.err.skills': 'Choisissez au moins une compétence.', 'project.pub.err.level': 'Choisissez un niveau.', 'project.pub.err.d': 'La description doit faire au moins 30 caractères.'
  },
  en: {
    'project.title': 'Project — 226os', 'project.notfound': 'Project not found', 'project.jump': 'See needs', 'project.form.summary': 'Please fix the following fields',
    'project.demo.title': 'Roles and access rights (prototype)', 'project.demo.note': 'In the product the role comes from the session and the server enforces every rule (authorization matrix). Here you pick the role to see the difference.',
    'project.demo.viewas': 'View the page as', 'project.demo.dup': 'Simulate a second submission', 'project.demo.dup.none': 'Available when the selected role already has an application.',
    'project.demo.dup.ready': 'Reopens the form for a need already applied to: the server must refuse the duplicate.',
    'project.role.owner': 'Project owner', 'project.role.applicant': 'Applicant', 'project.role.member': 'Other member', 'project.role.anonymous': 'Anonymous visitor',
    'project.role.desc.owner': 'Owner: reads the applications of their project and accepts or declines them. Cannot apply to their own project.',
    'project.role.desc.applicant': 'Signed-in member with an application: sees only their own applications. A second application to the same need is refused.',
    'project.role.desc.member': 'Signed-in member without an application: can apply to an open need. The private applications list is refused (403).',
    'project.role.desc.anonymous': 'Signed-out visitor: reads the public page but cannot apply or read applications.',
    'project.badge.license': 'Licence {l}', 'project.badge.fake': 'Fictional link',
    'project.ctx.title': 'Linked context', 'project.ctx.lede': 'Problem-to-Project: the originating question and the API involved stay one click away.',
    'project.ctx.question': 'Originating question', 'project.ctx.resource': 'Linked API / tool', 'project.ctx.maintainer': 'Maintainer', 'project.ctx.none': 'No question or API is linked to this project yet.',
    'project.desc.title': 'Description', 'project.needs.title': 'Contribution needs', 'project.needs.skills': 'Required skills', 'project.needs.level': 'Level',
    'project.need.open': 'Open', 'project.need.closed': 'Closed', 'project.need.apply': 'Apply', 'project.need.apply.aria': 'Apply to the need: {need}',
    'project.need.signin': 'Sign in to apply', 'project.need.mine': 'Your project', 'project.need.sent': 'Application sent', 'project.need.apps': '{n} application(s)',
    'project.apps.title.owner': 'Applications received', 'project.apps.title.other': 'Applications', 'project.apps.mine': 'My applications',
    'project.apps.private': 'Private list: reserved to the project owner.',
    'project.apps.denied.title': 'Access denied (403)', 'project.apps.denied.body': 'Only the project owner can read the applications list. The server refuses the request and reveals no data or count.',
    'project.apps.anon.title': 'Sign-in required (401)', 'project.apps.anon.body': 'Sign in to apply. The applications list stays reserved to the owner.',
    'project.apps.empty.title': 'No applications yet', 'project.apps.empty.body': 'Applications will appear here as soon as a member shows interest.',
    'project.app.for': 'Need:', 'project.app.skills': 'Skills:', 'project.app.avail': 'Availability:', 'project.app.accept': 'Accept', 'project.app.decline': 'Decline',
    'project.app.accept.aria': 'Accept the application from {name}', 'project.app.decline.aria': 'Decline the application from {name}', 'project.app.decided': 'Decision recorded.',
    'project.status.pending': 'Pending', 'project.status.accepted': 'Accepted', 'project.status.declined': 'Declined', 'project.apps.toast.accepted': 'Application accepted.', 'project.apps.toast.declined': 'Application declined.',
    'project.avail.lt2': 'Under 2 h per week', 'project.avail.2to5': '2 to 5 h per week', 'project.avail.5to10': '5 to 10 h per week', 'project.avail.gt10': 'Over 10 h per week', 'project.avail.choose': 'Not specified',
    'project.act.title': 'Activity', 'project.act.created': '{who} created the project.', 'project.act.repo': '{who} updated the repository details.', 'project.act.need': '{who} published a need:',
    'project.act.need-closed': '{who} closed the need:', 'project.act.link': '{who} linked the question:',
    'project.side.repo': 'Repository', 'project.side.repo.link': 'Open the repository (fictional link)', 'project.side.license': 'Licence', 'project.side.stage': 'Stage', 'project.side.country': 'Country', 'project.side.langs': 'Working languages', 'project.side.updated': 'Updated',
    'project.side.stack': 'Stack', 'project.side.maint': 'Project owner',
    'project.lang.fr': 'Français', 'project.lang.en': 'English',
    'project.empty.title': 'Project not found', 'project.empty.body': 'This identifier matches no project. Go back to the directory to pick one.', 'project.empty.cta': 'Back to projects',
    'project.apply.title': 'Apply to contribute', 'project.apply.need': 'Need:', 'project.apply.sim': 'No application is actually sent or saved.',
    'project.apply.motivation': 'Motivation', 'project.apply.motivation.hint': '30 to 600 characters. Say why this need and what you would do first.',
    'project.apply.skills': 'Relevant skills', 'project.apply.skills.hint': 'Comma-separated, e.g. Go, PostgreSQL.', 'project.apply.avail': 'Availability', 'project.apply.submit': 'Send application',
    'project.apply.ok.title': 'Application sent', 'project.apply.ok.body': 'The project owner can accept or decline it.',
    'project.apply.dup.title': 'Application already sent (409). ', 'project.apply.dup.body': 'You have already applied to this need: no new request was created.',
    'project.apply.toast': 'Application sent.',
    'project.err.motivation.required': 'Explain your motivation.', 'project.err.motivation.short': 'The motivation must be at least 30 characters.',
    'project.err.skills.required': 'Enter at least one skill.', 'project.err.avail.required': 'Choose an availability.',
    'project.demo.sample.motivation': 'I would like to contribute to this need and can start with a small task.', 'project.demo.sample.skills': 'Go, PostgreSQL',
    'project.pub.btn': 'Publish a need', 'project.pub.title': 'Publish a need', 'project.pub.sim': 'Simulation: the need is added to the list, nothing is saved.',
    'project.pub.t': 'Need title', 'project.pub.t.hint': 'One precise sentence, 8 to 100 characters.', 'project.pub.skills': 'Required skills', 'project.pub.skills.hint': 'Pick at least one skill.',
    'project.pub.level': 'Expected level', 'project.pub.level.choose': 'Choose a level', 'project.pub.d': 'Description', 'project.pub.d.hint': 'What the contributor will do, 30 characters minimum.',
    'project.pub.submit': 'Publish need', 'project.pub.toast': 'Need published.', 'project.pub.summary': 'Please fix the following fields',
    'project.pub.err.t': 'Enter a title of 8 to 100 characters.', 'project.pub.err.skills': 'Pick at least one skill.', 'project.pub.err.level': 'Choose a level.', 'project.pub.err.d': 'The description must be at least 30 characters.'
  }
});

(function () {
  'use strict';
  var B = MOCK.build, $ = App.$, esc = App.esc, t = function (k, p) { return I18N.t(k, p); };
  var id = App.params.get('id') || 'pr-webhook-kit', P = B.byId(id);
  var VIEWERS = { owner: P && P.owner, applicant: 'p-kofi', member: 'p-imane', anonymous: null };
  var ROLES = ['owner', 'applicant', 'member', 'anonymous'];
  App.Proto.register({ roles: ROLES.map(function (r) { return { id: r, labelKey: 'project.role.' + r }; }), 'default': 'member' });
  if (ROLES.indexOf(App.params.get('as')) > -1) App.Proto.setRole(App.params.get('as'));   /* ?as= kept as an alias of ?role= */
  var role = App.Proto.role;
  var apps = P ? P.applications.map(function (a) { return Object.assign({}, a); }) : [];
  var applyNeed = null, busy = false, errs = {}, dup = false;

  function name(pid) { var p = MOCK.person(pid); return p ? p.name : pid; }
  function viewer() { return VIEWERS[role]; }
  function mineOn(needId) { return apps.filter(function (a) { return a.person === viewer() && a.need === needId; })[0]; }
  function needOf(nid) { return P.needs.filter(function (n) { return n.id === nid; })[0]; }
  function badgeLang(l, tid) { return '<span class="badge" title="' + esc(t('common.originalLang')) + '" data-testid="' + tid + '"><span class="visually-hidden" data-testid="' + tid + '-label">' + esc(t('common.originalLang')) + ' </span>' + l.toUpperCase() + '</span>'; }
  function personHtml(pid, tid, withLink) {
    var p = MOCK.person(pid); if (!p) return '';
    var nm = withLink ? '<a href="profile.html?id=' + p.id + '" data-testid="' + tid + '-link">' + esc(p.name) + '</a>' : '<span data-testid="' + tid + '-name">' + esc(p.name) + '</span>';
    return '<span class="person" data-testid="' + tid + '">' + App.avatar(p, { size: 'sm', testid: tid + '-avatar' }) + nm + '</span>';
  }
  function countryHtml(code, tid) { return '<span class="country" data-testid="' + tid + '"><span class="country-code" aria-hidden="true" data-testid="' + tid + '-code">' + code + '</span><span data-testid="' + tid + '-name">' + esc(MOCK.countryName(code)) + '</span></span>'; }

  /* ---- Markdown: shared renderer (assets/js/markdown.js); "##" -> h3 ---- */
  function md(src) { return window.Markdown.render(src, { id: 'project-md', headStart: 2 }); }

  /* ---- header / role toolbar (outside the simulated area) ---- */
  function header() {
    if (!P) { $('#project-h1').textContent = t('project.notfound'); $('#project-h1').classList.remove('mono'); $('#project-crumb').textContent = t('project.notfound'); $('#project-jump').hidden = true; $('#project-summary').hidden = true; document.title = t('project.notfound') + ' — 226os'; return; }
    $('#project-h1').textContent = P.name; $('#project-h1').lang = P.lang; $('#project-h1').setAttribute('data-user-content', '');
    $('#project-thumb').innerHTML = App.thumb(P.name, { size: 'lg', testid: 'project-thumb-img' });
    $('#project-crumb').textContent = P.name; document.title = P.name + ' — 226os';
    $('#project-badges').innerHTML = '<span class="meta-item" data-testid="project-stage-badge">' + esc(MOCK.label(B.stages[P.stage])) + '</span>' +
      '<span class="meta-item" data-testid="project-country-item">' + countryHtml(P.country, 'project-country') + '</span>';
    $('#project-summary').innerHTML = '<span lang="' + P.lang + '" data-user-content data-testid="project-summary-text">' + esc(P.summary) + '</span>';
    $('#project-jump').hidden = false; $('#project-summary').hidden = false;
  }
  function toolbar() {
    $('[data-testid=project-demo]').hidden = !P; if (!P) return;
    $('#project-role-desc').textContent = t('project.role.desc.' + role);
    var can = P && !!dupTarget(); $('#project-dup-btn').disabled = !can;
    $('#project-dup-hint').textContent = can ? t('project.demo.dup.ready') : t('project.demo.dup.none');
  }
  function dupTarget() { var v = viewer(); if (!v || role === 'owner') return null; return apps.filter(function (a) { return a.person === v; })[0] || null; }
  document.addEventListener('proto:role', function (e) {
    role = e.detail.role; toolbar(); sim && sim.refresh();
    try { var u = new URL(location.href); u.searchParams.set('as', role); history.replaceState(null, '', u); } catch (x) { /* file:// may refuse */ }
  });

  /* ---- main view ---- */
  function needsHtml() {
    return P.needs.map(function (n, i) {
      var k = n.id, mine = mineOn(n.id), count = apps.filter(function (a) { return a.need === n.id; }).length, act;
      if (n.status === 'closed') act = '';
      else if (role === 'owner') act = '<span class="text-sm text-muted" data-testid="project-need-' + k + '-owner">' + esc(t('project.need.mine')) + ' · ' + esc(t('project.need.apps', { n: count })) + '</span>';
      else if (role === 'anonymous') act = '<a class="btn btn-sm" href="auth.html" data-testid="project-need-' + k + '-signin"><span class="icon i-lock" aria-hidden="true" data-testid="project-need-' + k + '-signin-icon"></span><span data-testid="project-need-' + k + '-signin-label">' + esc(t('project.need.signin')) + '</span></a>';
      else if (mine) act = '<span class="badge badge-verified" data-testid="project-need-' + k + '-sent"><span class="icon i-check" aria-hidden="true" data-testid="project-need-' + k + '-sent-icon"></span>' + esc(t('project.need.sent')) + '</span>';
      else act = '<button type="button" class="btn btn-primary btn-sm" data-apply="' + n.id + '" aria-label="' + esc(t('project.need.apply.aria', { need: n.title })) + '" data-testid="project-need-' + k + '-apply"><span class="icon i-send" aria-hidden="true" data-testid="project-need-' + k + '-apply-icon"></span>' + esc(t('project.need.apply')) + '</button>';
      return '<li class="row" data-testid="project-need-' + k + '"><div class="row-lead"><span class="icon i-' + (n.status === 'open' ? 'users' : 'lock') + '" aria-hidden="true" data-testid="project-need-' + k + '-icon"></span></div>' +
        '<div class="row-main"><h3 class="row-title" lang="' + P.lang + '" data-user-content data-testid="project-need-' + k + '-title">' + esc(n.title) + '</h3>' +
        '<p class="text-sm"><span class="text-muted" data-testid="project-need-' + k + '-skills-label">' + esc(t('project.needs.skills')) + (I18N.lang === 'fr' ? '\u00a0:' : ':') + '</span> <strong data-testid="project-need-' + k + '-skills">' + esc(n.skills.map(function (s) { return MOCK.label(B.skills[s]); }).join(', ')) + '</strong></p>' +
        '<div class="meta"><span class="meta-item" data-testid="project-need-' + k + '-level">' + esc(t('project.needs.level')) + (I18N.lang === 'fr' ? '\u00a0: ' : ': ') + esc(MOCK.label(B.levels[n.level])) + '</span></div></div>' +
        '<div class="row-aside"><span class="badge ' + (n.status === 'open' ? 'badge-accepted' : 'badge-later') + '" data-testid="project-need-' + k + '-status">' + esc(t(n.status === 'open' ? 'project.need.open' : 'project.need.closed')) + '</span>' + act + '</div></li>';
    }).join('');
  }
  function statusBadge(s, tid) { var c = { pending: 'badge-warn', accepted: 'badge-accepted', declined: 'badge-danger' }[s], ic = { pending: 'clock', accepted: 'check', declined: 'x' }[s];
    return '<span class="badge ' + c + '" data-testid="' + tid + '"><span class="icon i-' + ic + '" aria-hidden="true" data-testid="' + tid + '-icon"></span>' + esc(t('project.status.' + s)) + '</span>'; }
  function appRow(a, i, owner, pre) {
    var tid = pre + '-' + a.id, nd = needOf(a.need);
    return '<li class="row" tabindex="-1" data-app="' + a.id + '" data-testid="' + tid + '"><div class="row-lead">' + App.avatar(MOCK.person(a.person) || name(a.person), { testid: tid + '-avatar' }) + '</div>' +
      '<div class="row-main"><h' + (owner ? 3 : 4) + ' class="row-title">' + (owner ? '<a class="row-above" href="profile.html?id=' + a.person + '" data-testid="' + tid + '-person">' + esc(name(a.person)) + '</a>' : '<span data-testid="' + tid + '-person">' + esc(name(a.person)) + '</span>') + '</h' + (owner ? 3 : 4) + '>' +
      '<p class="text-sm"><span class="text-muted" data-testid="' + tid + '-need-label">' + esc(t('project.app.for')) + '</span> <span lang="' + P.lang + '" data-user-content data-testid="' + tid + '-need">' + esc(nd.title) + '</span></p>' +
      '<p lang="' + a.lang + '" data-user-content data-testid="' + tid + '-motivation">' + esc(a.motivation) + '</p>' +
      '<div class="meta"><span class="meta-item" data-testid="' + tid + '-skills">' + esc(t('project.app.skills')) + ' ' + esc(a.skills) + '</span><span class="meta-item" data-testid="' + tid + '-avail">' + esc(t('project.app.avail')) + ' ' + esc(t('project.avail.' + a.avail)) + '</span>' +
      '<time datetime="' + a.date + '" data-testid="' + tid + '-date">' + esc(MOCK.fmtDate(a.date)) + '</time>' + badgeLang(a.lang, tid + '-lang') + '</div></div>' +
      '<div class="row-aside">' + statusBadge(a.status, tid + '-status') +
      (owner ? (a.status === 'pending'
        ? '<div class="cluster row-above"><button type="button" class="btn btn-icon btn-sm" data-decide="accepted" data-app-id="' + a.id + '" aria-label="' + esc(t('project.app.accept.aria', { name: name(a.person) })) + '" data-testid="' + tid + '-accept"><span class="icon i-check" aria-hidden="true" data-testid="' + tid + '-accept-icon"></span></button>' +
          '<button type="button" class="btn btn-icon btn-sm" data-decide="declined" data-app-id="' + a.id + '" aria-label="' + esc(t('project.app.decline.aria', { name: name(a.person) })) + '" data-testid="' + tid + '-decline"><span class="icon i-x" aria-hidden="true" data-testid="' + tid + '-decline-icon"></span></button></div>'
        : '') : '') + '</div></li>';
  }
  function appsHtml() {
    var h = '<h2 id="project-apps-h" data-testid="project-apps-title">' + esc(t(role === 'owner' ? 'project.apps.title.owner' : 'project.apps.title.other')) + '</h2>';
    if (role === 'owner') {
      h += apps.length ? '<ul class="list list-reset" data-testid="project-apps-list">' + apps.map(function (a, i) { return appRow(a, i, true, 'project-app'); }).join('') + '</ul>'
        : '<div class="state" data-testid="project-apps-empty"><span class="icon i-inbox" aria-hidden="true" data-testid="project-apps-empty-icon"></span><h3 data-testid="project-apps-empty-title">' + esc(t('project.apps.empty.title')) + '</h3><p data-testid="project-apps-empty-body">' + esc(t('project.apps.empty.body')) + '</p></div>';
      return h;
    }
    var anon = role === 'anonymous', hasMine = !anon && apps.some(function (a) { return a.person === viewer(); });
    h = hasMine ? '' : '<h2 id="project-apps-h" data-testid="project-apps-title" data-proto>' + esc(t('project.apps.title.other')) + '</h2>';
    if (hasMine) h = '<h2 id="project-apps-h" data-testid="project-apps-title">' + esc(t('project.apps.mine')) + '</h2>';
    h += '<div class="state state-error" data-proto role="status" data-testid="project-apps-denied"><span class="icon i-' + (anon ? 'lock' : 'shield-check') + '" aria-hidden="true" data-testid="project-apps-denied-icon"></span><h3 data-testid="project-apps-denied-title">' + esc(t(anon ? 'project.apps.anon.title' : 'project.apps.denied.title')) + '</h3><p data-testid="project-apps-denied-body">' + esc(t(anon ? 'project.apps.anon.body' : 'project.apps.denied.body')) + '</p>' + (anon ? '<div class="state-actions"><a class="btn btn-sm" href="auth.html" data-testid="project-apps-signin"><span class="icon i-user" aria-hidden="true" data-testid="project-apps-signin-icon"></span><span data-testid="project-apps-signin-label">' + esc(t('project.need.signin')) + '</span></a></div>' : '') + '</div>';
    var mine = anon ? [] : apps.filter(function (a) { return a.person === viewer(); });
    if (mine.length) h += '<ul class="list list-reset" data-testid="project-mine-list">' + mine.map(function (a, i) { return appRow(a, i, false, 'project-mine'); }).join('') + '</ul>';
    return h;
  }
  function actHtml() {
    return P.activity.map(function (e, i) {
      var k = i + 1, tail = '';
      if (e.k === 'need' || e.k === 'need-closed') tail = ' <span lang="' + P.lang + '" data-user-content data-testid="project-activity-' + k + '-ref">« ' + esc(needOf(e.ref).title) + ' »</span>';
      if (e.k === 'link') { var q = B.question(e.ref); tail = ' <a href="question.html?id=' + q.id + '" lang="' + q.lang + '" data-user-content data-testid="project-activity-' + k + '-ref">' + esc(q.title) + '</a>'; }
      return '<li class="row" data-testid="project-activity-' + k + '"><div class="row-lead"><span class="icon i-' + { created: 'plus', repo: 'git-branch', need: 'users', 'need-closed': 'lock', link: 'link' }[e.k] + '" aria-hidden="true" data-testid="project-activity-' + k + '-icon"></span></div>' +
        '<div class="row-main"><p class="text-sm" data-testid="project-activity-' + k + '-text">' + esc(t('project.act.' + e.k, { who: name(e.who) })) + tail + '</p></div><div class="row-aside"><time datetime="' + e.d + '" data-testid="project-activity-' + k + '-date">' + esc(MOCK.fmtDate(e.d)) + '</time></div></li>';
    }).join('');
  }
  function ctxHtml() {
    var q = B.question(P.question), r = B.resource(P.resource), items = '';
    if (q) items += '<a class="linked-item" href="question.html?id=' + q.id + '" data-testid="project-link-question"><span class="kind" data-testid="project-link-question-kind"><span class="icon i-help" aria-hidden="true" data-testid="project-link-question-icon"></span>' + esc(t('project.ctx.question')) + '</span><strong lang="' + q.lang + '" data-user-content data-testid="project-link-question-title">' + esc(q.title) + '</strong></a>';
    if (r) items += '<a class="linked-item" href="resource.html?id=' + r.id + '" data-testid="project-link-resource"><span class="kind" data-testid="project-link-resource-kind"><span class="icon i-compass" aria-hidden="true" data-testid="project-link-resource-icon"></span>' + esc(t('project.ctx.resource')) + '</span><strong lang="fr" data-user-content data-testid="project-link-resource-title">' + esc(r.name) + '</strong></a>';
    items += '<a class="linked-item" href="profile.html?id=' + P.owner + '" data-testid="project-link-owner"><span class="kind" data-testid="project-link-owner-kind"><span class="icon i-user" aria-hidden="true" data-testid="project-link-owner-icon"></span>' + esc(t('project.ctx.maintainer')) + '</span><strong data-testid="project-link-owner-title">' + esc(name(P.owner)) + '</strong></a>';
    return '<h2 data-testid="project-ctx-title">' + esc(t('project.ctx.title')) + '</h2><p class="hint" data-proto data-testid="project-ctx-lede">' + esc(t('project.ctx.lede')) + '</p><div class="linked-items" data-testid="project-linked-items">' + items + '</div>' +
      (q || r ? '' : '<p class="text-muted text-sm" data-testid="project-ctx-none">' + esc(t('project.ctx.none')) + '</p>');
  }
  function asideHtml() {
    return '<aside class="stack" data-testid="project-aside"><section class="card" aria-labelledby="project-repo-h" data-testid="project-repo-card"><h2 id="project-repo-h" data-testid="project-repo-title">' + esc(t('project.side.repo')) + '</h2>' +
      '<p class="stack-sm"><a class="link-external" rel="noopener noreferrer" target="_blank" href="' + B.repoUrl(P) + '" data-testid="project-repo-link">' + esc(B.repoLabel(P)) + '</a> <span class="badge badge-sim" data-proto data-testid="project-repo-fake">' + esc(t('project.badge.fake')) + '</span></p>' +
      '<dl class="dl-meta" data-testid="project-meta">' +
        '<dt data-testid="project-meta-license-label">' + esc(t('project.side.license')) + '</dt><dd data-testid="project-meta-license">' + esc(P.license) + '</dd>' +
        '<dt data-testid="project-meta-stage-label">' + esc(t('project.side.stage')) + '</dt><dd data-testid="project-meta-stage">' + esc(MOCK.label(B.stages[P.stage])) + '</dd>' +
        '<dt data-testid="project-meta-country-label">' + esc(t('project.side.country')) + '</dt><dd data-testid="project-meta-country">' + esc(MOCK.countryName(P.country)) + ' (' + P.country + ')</dd>' +
        '<dt data-testid="project-meta-langs-label">' + esc(t('project.side.langs')) + '</dt><dd data-testid="project-meta-langs">' + esc(P.langs.map(function (l) { return t('project.lang.' + l); }).join(', ')) + '</dd>' +
        '<dt data-testid="project-meta-updated-label">' + esc(t('project.side.updated')) + '</dt><dd data-testid="project-meta-updated"><time datetime="' + P.updated + '" data-testid="project-meta-updated-date">' + esc(MOCK.fmtDate(P.updated)) + '</time></dd></dl></section>' +
      '<section class="card" aria-labelledby="project-stack-h" data-testid="project-stack-card"><h2 id="project-stack-h" data-testid="project-stack-title">' + esc(t('project.side.stack')) + '</h2><ul class="tag-list list-reset" data-testid="project-stack">' +
        P.stack.map(function (s) { return '<li><span class="tag" data-testid="project-stack-tag-' + s.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '') + '">' + esc(s) + '</span></li>'; }).join('') + '</ul></section>' +
      '<section class="card" aria-labelledby="project-maint-h" data-testid="project-maint-card"><h2 id="project-maint-h" data-testid="project-maint-title">' + esc(t('project.side.maint')) + '</h2>' + personHtml(P.owner, 'project-maint-1', true) +
        '<p class="text-sm text-muted" data-testid="project-maint-headline" lang="' + (MOCK.person(P.owner).lang) + '" data-user-content>' + esc(MOCK.person(P.owner).headline) + '</p></section></aside>';
  }
  function view() {
    return '<div class="with-aside"><div class="stack-lg">' +
      '<section aria-labelledby="project-ctx-h" data-testid="project-ctx">' + ctxHtml().replace('<h2 ', '<h2 id="project-ctx-h" ') + '</section>' +
      '<section aria-labelledby="project-desc-h" data-testid="project-desc"><div class="cluster"><h2 id="project-desc-h" data-testid="project-desc-title">' + esc(t('project.desc.title')) + '</h2>' + badgeLang(P.lang, 'project-desc-lang') + '</div>' +
        '<div class="prose" lang="' + P.lang + '" data-user-content data-testid="project-desc-body">' + md(P.description) + '</div></section>' +
      '<section id="needs" aria-labelledby="project-needs-h" data-testid="project-needs"><div class="section-head"><h2 id="project-needs-h" data-testid="project-needs-title">' + esc(t('project.needs.title')) + '</h2>' + (role === 'owner' ? '<button type="button" class="btn btn-sm" data-dialog-open="need-dialog" data-testid="project-publish-need"><span class="icon i-plus" aria-hidden="true" data-testid="project-publish-need-icon"></span>' + esc(t('project.pub.btn')) + '</button>' : '') + '</div><ul class="list list-reset project-list" data-testid="project-needs-list">' + needsHtml() + '</ul></section>' +
      '<section aria-labelledby="project-apps-h" data-testid="project-apps"' + (role !== 'owner' && !apps.some(function (a) { return a.person === viewer(); }) ? ' data-proto' : '') + '>' + appsHtml() + '</section>' +
      '<section aria-labelledby="project-act-h" data-testid="project-act"><h2 id="project-act-h" data-testid="project-act-title">' + esc(t('project.act.title')) + '</h2><ol class="list list-reset project-list" data-testid="project-activity-list">' + actHtml() + '</ol></section>' +
      '</div>' + asideHtml() + '</div>';
  }

  var sim = App.Sim.load($('#project-view'), {
    load: function () { return P ? [P] : []; }, skeleton: 4,
    empty: { key: 'project.empty', action: { href: 'build.html', icon: 'chevron-left', labelKey: 'project.empty.cta', testid: 'project-empty-back' } },
    render: function (items, box) { box.innerHTML = view(); }
  });
  document.addEventListener('i18n:change', function () { header(); toolbar(); fillAvail(); showErrors(false); });
  header(); toolbar();

  /* ---- owner decisions + apply buttons (delegated; survive re-render) ---- */
  $('#project-view').addEventListener('click', function (e) {
    var d = e.target.closest('[data-decide]'), a = e.target.closest('[data-apply]');
    if (d) {
      var rec = apps.filter(function (x) { return x.id === d.getAttribute('data-app-id'); })[0];
      if (!rec || rec.status !== 'pending' || role !== 'owner') return;   /* invalid transition / non-owner: refused */
      rec.status = d.getAttribute('data-decide'); sim.refresh(); toolbar();
      var li = $('[data-app="' + rec.id + '"]'); if (li) li.focus();
      App.Toast.show({ kind: 'success', msg: t('project.apps.toast.' + rec.status) });
    } else if (a) openApply(a.getAttribute('data-apply'), a);
  });

  /* ---- apply dialog (static markup; never rebuilt on language change) ---- */
  var form = $('#apply-form'), dlg = $('#apply-dialog');
  var FIELDS = { motivation: '#apply-motivation', skills: '#apply-skills', avail: '#apply-avail' };
  function fillAvail() {
    var sel = $('#apply-avail'), v = sel.value;
    sel.innerHTML = '<option value="" data-testid="apply-avail-option-none">' + esc(t('project.avail.choose')) + '</option>' + B.availability.map(function (k) { return '<option value="' + k + '"' + (k === v ? ' selected' : '') + ' data-testid="apply-avail-option-' + k + '">' + esc(t('project.avail.' + k)) + '</option>'; }).join('');
  }
  fillAvail();
  function openApply(needId, from, prefill) {
    applyNeed = needId; var n = needOf(needId); $('#apply-need').textContent = n.title; $('#apply-need').lang = P.lang;
    if (prefill) { $('#apply-motivation').value = t('project.demo.sample.motivation'); $('#apply-skills').value = t('project.demo.sample.skills'); $('#apply-avail').value = '2to5'; }
    App.Dialog.open('apply-dialog', from);
  }
  $('#project-dup-btn').addEventListener('click', function (e) { var a = dupTarget(); if (a) openApply(a.need, e.currentTarget, true); });
  function validate() {
    errs = {}; var m = $('#apply-motivation').value.trim(), s = $('#apply-skills').value.trim();
    if (!m) errs.motivation = 'project.err.motivation.required'; else if (m.length < 30) errs.motivation = 'project.err.motivation.short';
    if (!s) errs.skills = 'project.err.skills.required';
    return !Object.keys(errs).length;
  }
  function showErrors(focus) {
    var keys = Object.keys(errs), sum = $('#apply-summary');
    Object.keys(FIELDS).forEach(function (f) {
      var inp = $(FIELDS[f]), msg = $('#apply-' + f + '-error'), has = !!errs[f];
      inp.closest('.field').classList.toggle('has-error', has); if (has) inp.setAttribute('aria-invalid', 'true'); else inp.removeAttribute('aria-invalid');
      msg.hidden = !has; msg.innerHTML = has ? '<span class="icon i-alert-circle" aria-hidden="true" data-testid="apply-' + f + '-error-icon"></span><span data-testid="apply-' + f + '-error-text">' + esc(t(errs[f])) + '</span>' : '';
    });
    sum.hidden = !keys.length;
    $('#apply-summary-list').innerHTML = keys.map(function (f, i) { return '<li><a href="' + FIELDS[f] + '" data-testid="apply-summary-link-' + f + '">' + esc(t(errs[f])) + '</a></li>'; }).join('');
    if (focus && keys.length) sum.focus();
  }
  $('#apply-summary-list').addEventListener('click', function (e) { var a = e.target.closest('a'); if (a) { e.preventDefault(); $(a.getAttribute('href')).focus(); } });
  form.addEventListener('submit', function (e) {
    e.preventDefault(); if (busy) return;                                   /* a double click never creates two requests */
    $('#apply-dup').hidden = true;
    if (!validate()) { showErrors(true); return; }
    showErrors(false); busy = true; var b = $('#apply-submit'); b.setAttribute('aria-busy', 'true'); b.disabled = true;
    setTimeout(function () {
      busy = false; b.removeAttribute('aria-busy'); b.disabled = false;
      if (mineOn(applyNeed)) { var box = $('#apply-dup'); box.hidden = false; box.focus(); return; }   /* unique (user, need): 409 */
      apps.push({ id: 'a-new-' + apps.length, need: applyNeed, person: viewer(), status: 'pending', date: '2026-11-11', avail: $('#apply-avail').value, skills: $('#apply-skills').value.trim(), lang: I18N.lang, motivation: $('#apply-motivation').value.trim() });
      b.hidden = true; $('#apply-body').hidden = true; $('#apply-success').hidden = false; $('#apply-cancel').setAttribute('data-i18n', 'common.close'); $('#apply-cancel').textContent = t('common.close'); $('#apply-cancel').focus();
      sim.refresh(); toolbar(); App.Toast.show({ kind: 'success', msg: t('project.apply.toast') });
    }, 700);
  });
  dlg.addEventListener('close', function () {
    form.reset(); errs = {}; showErrors(false); $('#apply-dup').hidden = true; $('#apply-body').hidden = false; $('#apply-success').hidden = true; $('#apply-submit').hidden = false;
    $('#apply-cancel').setAttribute('data-i18n', 'common.cancel'); $('#apply-cancel').textContent = t('common.cancel');
  });

  /* ---- owner-only "Publish a need" dialog (static markup; simulated: the need is appended in memory, nothing persisted) ---- */
  var nform = $('#need-form'), ndlg = $('#need-dialog'), nerrs = {}, nbusy = false, nseq = 0;
  var NF = { t: '#need-title', skills: '#need-skills-group', level: '#need-level', d: '#need-desc' };
  function fillNeedForm() {
    var checked = [].slice.call(nform.querySelectorAll('[name=need-skill]:checked')).map(function (x) { return x.value; }), lv = $('#need-level').value;
    $('#need-skills-list').innerHTML = Object.keys(B.skills).map(function (k) {
      return '<label class="choice" data-testid="need-skill-' + k + '-choice"><input type="checkbox" name="need-skill" value="' + k + '"' + (checked.indexOf(k) > -1 ? ' checked' : '') + ' data-testid="need-skill-' + k + '"><span data-testid="need-skill-' + k + '-label">' + esc(MOCK.label(B.skills[k])) + '</span></label>';
    }).join('');
    $('#need-level').innerHTML = '<option value="" data-testid="need-level-option-none">' + esc(t('project.pub.level.choose')) + '</option>' + Object.keys(B.levels).map(function (k) {
      return '<option value="' + k + '"' + (k === lv ? ' selected' : '') + ' data-testid="need-level-option-' + k + '">' + esc(MOCK.label(B.levels[k])) + '</option>';
    }).join('');
  }
  function needValues() { return { t: $('#need-title').value.trim(), skills: [].slice.call(nform.querySelectorAll('[name=need-skill]:checked')).map(function (x) { return x.value; }), level: $('#need-level').value, d: $('#need-desc').value.trim() }; }
  function needValidate() {
    var v = needValues(); nerrs = {};
    if (v.t.length < 8 || v.t.length > 100) nerrs.t = 'project.pub.err.t';
    if (!v.skills.length) nerrs.skills = 'project.pub.err.skills';
    if (!v.level) nerrs.level = 'project.pub.err.level';
    if (v.d.length < 30) nerrs.d = 'project.pub.err.d';
    return !Object.keys(nerrs).length;
  }
  function needShowErrors(focus) {
    var keys = Object.keys(nerrs), sum = $('#need-summary');
    Object.keys(NF).forEach(function (f) {
      var inp = $(NF[f]), msg = $('#need-' + f + '-error'), has = !!nerrs[f];
      inp.closest('.field').classList.toggle('has-error', has);
      if (f !== 'skills') { if (has) inp.setAttribute('aria-invalid', 'true'); else inp.removeAttribute('aria-invalid'); }
      msg.hidden = !has; msg.innerHTML = has ? '<span class="icon i-alert-circle" aria-hidden="true" data-testid="need-' + f + '-error-icon"></span><span data-testid="need-' + f + '-error-text">' + esc(t(nerrs[f])) + '</span>' : '';
    });
    sum.hidden = !keys.length;
    $('#need-summary-list').innerHTML = keys.map(function (f, i) { return '<li><a href="' + NF[f] + '" data-need-field="' + f + '" data-testid="need-summary-link-' + f + '">' + esc(t(nerrs[f])) + '</a></li>'; }).join('');
    if (focus && keys.length) sum.focus();
  }
  fillNeedForm();
  $('#need-summary-list').addEventListener('click', function (e) {
    var a = e.target.closest('a[data-need-field]'); if (!a) return; e.preventDefault();
    var f = a.getAttribute('data-need-field'); var tgt = f === 'skills' ? nform.querySelector('[name=need-skill]') : $(NF[f]); if (tgt) tgt.focus();
  });
  nform.addEventListener('submit', function (e) {
    e.preventDefault(); if (nbusy || role !== 'owner') return;
    if (!needValidate()) { needShowErrors(true); return; }
    needShowErrors(false); nbusy = true; var btn = $('#need-submit'); btn.setAttribute('aria-busy', 'true'); btn.disabled = true;
    setTimeout(function () {
      var v = needValues(); nbusy = false; btn.removeAttribute('aria-busy'); btn.disabled = false;
      P.needs.push({ id: 'n-new-' + (++nseq), title: v.t, skills: v.skills, level: v.level, status: 'open', description: v.d });
      App.Dialog.close(ndlg); sim.refresh();
      App.Toast.show({ kind: 'success', msg: t('project.pub.toast') });
      var rows = document.querySelectorAll('[data-testid="project-needs-list"] > li'), last = rows[rows.length - 1];
      if (last) { last.tabIndex = -1; last.focus(); }
    }, 600);
  });
  ndlg.addEventListener('close', function () { nform.reset(); nerrs = {}; needShowErrors(false); });
  document.addEventListener('i18n:change', function () { fillNeedForm(); needShowErrors(false); });
})();
