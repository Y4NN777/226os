/* Profile (v2): public view of a member (?id=) and SELF view (signed in, no id or own id).
   Self view = data from the simulated session (so Settings edits show up) + the setup checklist at the top (App.setupChecklist: only while incomplete).
   Privacy: the e-mail is never shown; signed-out visitors do not get members-only fields (they are omitted, not locked). Namespace: profile.* */
I18N.register({
  fr: {
    'profile.title': 'Profil — 226os', 'profile.titleOf': '{name} — profil — 226os',
    'profile.edit': 'Modifier le profil', 'profile.since': 'Membre depuis {date}',
    'profile.open.collab': 'Ouvert aux collaborations', 'profile.open.mentorat': 'Ouvert au mentorat',
    'profile.github': 'GitHub', 'profile.languages': 'Langues', 'profile.skills': 'Compétences',
    'profile.privacy': 'Votre e-mail n’est jamais public. Vous choisissez ce que les autres voient.', 'profile.privacy.link': 'Confidentialité',
    'profile.tabs': 'Activité', 'profile.tab.questions': 'Questions', 'profile.tab.answers': 'Réponses', 'profile.tab.projects': 'Projets', 'profile.tab.contributions': 'Contributions',
    'profile.empty.self.questions': 'Vous n’avez pas encore posé de question.', 'profile.empty.self.answers': 'Vos réponses apparaîtront ici.',
    'profile.empty.self.projects': 'Vous n’avez pas encore publié de projet.', 'profile.empty.self.contributions': 'Vos contributions apparaîtront ici.',
    'profile.empty.questions': 'Aucune question publiée.', 'profile.empty.answers': 'Aucune réponse pour l’instant.',
    'profile.empty.projects': 'Aucun projet publié.', 'profile.empty.contributions': 'Aucune contribution pour l’instant.',
    'profile.empty.self.cta.questions': 'Poser une question', 'profile.empty.self.cta.answers': 'Voir les questions', 'profile.empty.self.cta.projects': 'Voir les projets', 'profile.empty.self.cta.contributions': 'Trouver un projet',
    'profile.answers': '{n} réponse(s)', 'profile.noanswer': 'Sans réponse', 'profile.needs': '{n} besoin(s) ouvert(s)', 'profile.noneeds': 'Aucun besoin ouvert',
    'profile.notfound.title': 'Profil introuvable', 'profile.notfound.body': 'Ce membre n’existe pas ou a retiré son profil.'
  },
  en: {
    'profile.title': 'Profile — 226os', 'profile.titleOf': '{name} — profile — 226os',
    'profile.edit': 'Edit profile', 'profile.since': 'Member since {date}',
    'profile.open.collab': 'Open to collaborate', 'profile.open.mentorat': 'Open to mentoring',
    'profile.github': 'GitHub', 'profile.languages': 'Languages', 'profile.skills': 'Skills',
    'profile.privacy': 'Your email is never public. You choose what others see.', 'profile.privacy.link': 'Privacy',
    'profile.tabs': 'Activity', 'profile.tab.questions': 'Questions', 'profile.tab.answers': 'Answers', 'profile.tab.projects': 'Projects', 'profile.tab.contributions': 'Contributions',
    'profile.empty.self.questions': 'You have not asked a question yet.', 'profile.empty.self.answers': 'Your answers will show up here.',
    'profile.empty.self.projects': 'You have not published a project yet.', 'profile.empty.self.contributions': 'Your contributions will show up here.',
    'profile.empty.questions': 'No question published.', 'profile.empty.answers': 'No answer yet.',
    'profile.empty.projects': 'No project published.', 'profile.empty.contributions': 'No contribution yet.',
    'profile.empty.self.cta.questions': 'Ask a question', 'profile.empty.self.cta.answers': 'See the questions', 'profile.empty.self.cta.projects': 'See the projects', 'profile.empty.self.cta.contributions': 'Find a project',
    'profile.answers': '{n} answer(s)', 'profile.noanswer': 'No answer', 'profile.needs': '{n} open need(s)', 'profile.noneeds': 'No open need',
    'profile.notfound.title': 'Profile not found', 'profile.notfound.body': 'This member does not exist or removed their profile.'
  }
});

(function () {
  'use strict';
  var root = document.getElementById('profile-root'); if (!root) return;
  var A = window.App, M = window.MOCK, S = A.Session, C = M.connect, B = M.build, t = function (k, p) { return I18N.t(k, p); }, esc = A.esc, CFG = A.config;
  var id = A.params.get('id'), tab = 0, KEYS = ['questions', 'answers', 'projects', 'contributions'];
  var allQ = function () { return M.solve ? M.solve.all() : M.questions; };

  if (!id && !S.requireAuth()) return;

  function model() {
    var u = S.user, me = u ? C.selfPerson(u) : null, isSelf = !!u && (!id || (me && me.id === id));
    if (isSelf) {
      var gh = u.github ? { label: 'github.com/' + u.github, href: 'https://example.org/gh/' + encodeURIComponent(u.github) } : null;
      return { self: true, person: me, name: u.name, handle: u.username, country: u.country, city: '', since: me ? me.since : '', skills: u.skills || [], bio: u.bio || '', bioLang: I18N.lang, github: gh, collab: !!u.openToCollab, mentor: false, langs: u.languages || [], key: me ? me.id : null };
    }
    var p = M.person(id); if (!p) return null;
    var pr = C.profiles[p.id] || {}, vis = C.visibility, anon = !S.signedIn, show = function (f) { return vis[f] === 'public' || (!anon && vis[f] === 'members'); };
    return { self: false, person: p, name: p.name, handle: p.handle, country: show('country') ? p.country : '', city: show('city') ? p.city : '', since: show('since') ? p.since : '', skills: show('skills') ? p.skills : [], bio: show('bio') ? pr.bio || '' : '', bioLang: pr.bioLang || p.lang,
      github: show('github') && pr.github ? { label: 'github.com/' + p.handle, href: pr.github } : null, collab: show('openTo') && p.open.indexOf('collaboration') > -1, mentor: show('openTo') && p.open.indexOf('mentorat') > -1, langs: show('langs') ? p.langs : [], key: p.id };
  }

  function cc(code, tid) { return '<span class="country" data-testid="' + tid + '"><span class="country-code" data-testid="' + tid + '-code">' + code + '</span><span class="visually-hidden" data-testid="' + tid + '-name">' + esc(M.countryName(code)) + '</span></span>'; }
  function tagsHtml(a, tid, max) { return '<ul class="tag-list list-reset" data-testid="' + tid + '">' + a.slice(0, max || 4).map(function (x) { return '<li><span class="tag" data-testid="' + tid + '-' + String(x).replace(/[^a-z0-9]+/gi, '-').toLowerCase() + '">' + esc(x) + '</span></li>'; }).join('') + '</ul>'; }
  function monthYear(iso) { try { return new Date(iso + 'T12:00:00').toLocaleDateString(I18N.lang === 'fr' ? 'fr-FR' : 'en-GB', { month: 'long', year: 'numeric' }); } catch (e) { return iso; } }
  var LN = { fr: 'Français', en: 'English' };

  function rowQ(q) {
    var n = q.id, a = M.person(q.author);
    return '<li class="row" data-testid="profile-q-row-' + n + '"><div class="row-lead">' + A.avatar(a || '?', { testid: 'profile-q-avatar-' + n }) + '</div><div class="row-main"><h3 class="row-title"><a href="question.html?id=' + n + '" lang="' + q.lang + '" data-user-content data-testid="profile-q-link-' + n + '">' + esc(q.title) + '</a></h3>' +
      '<div class="meta">' + tagsHtml(q.tags, 'profile-q-tags-' + n, 3) + '<time class="meta-item" datetime="' + q.date + '" data-testid="profile-q-date-' + n + '">' + M.fmtDate(q.date) + '</time></div></div>' +
      '<div class="row-aside">' + (q.accepted ? A.statePill('accepted', { testid: 'profile-q-pill-' + n }) : '<span class="id-state-text" data-testid="profile-q-answers-' + n + '"><span class="icon i-message" aria-hidden="true" data-testid="profile-q-answers-icon-' + n + '"></span>' + esc(q.answers ? t('profile.answers', { n: q.answers }) : t('profile.noanswer')) + '</span>') + '</div></li>';
  }
  function rowP(p) {
    var n = p.id, open = p.openCount != null ? p.openCount : 0;
    return '<li class="row" data-testid="profile-p-row-' + n + '"><div class="row-lead">' + A.thumb(p.name, { icon: 'hammer', testid: 'profile-p-thumb-' + n }) + '</div><div class="row-main"><h3 class="row-title"><a href="project.html?id=' + n + '" class="id-mono" data-testid="profile-p-link-' + n + '">' + esc(p.name) + '</a></h3>' +
      '<p class="row-excerpt" lang="' + p.lang + '" data-user-content data-testid="profile-p-summary-' + n + '">' + esc(p.summary) + '</p>' + tagsHtml(p.stack, 'profile-p-tags-' + n, 3) + '</div>' +
      '<div class="row-aside">' + (open ? A.statePill('open', { testid: 'profile-p-pill-' + n }) + '<span class="tabular-nums" data-testid="profile-p-needs-' + n + '">' + esc(t('profile.needs', { n: open })) + '</span>' : '<span class="id-state-text" data-testid="profile-p-needs-' + n + '">' + esc(t('profile.noneeds')) + '</span>') + '</div></li>';
  }
  function collections(m) {
    var pid = m.key, Q = allQ(), P = B.projects();
    return {
      questions: pid ? Q.filter(function (q) { return q.author === pid; }) : [],
      answers: pid ? (C.answers[pid] || []).map(function (i) { return M.byId(Q, i); }).filter(Boolean) : [],
      projects: pid ? P.filter(function (p) { return p.owner === pid; }) : [],
      contributions: pid ? (C.contributions[pid] || []).map(function (i) { return M.byId(P, i); }).filter(Boolean) : []
    };
  }
  var ROWFN = { questions: rowQ, answers: rowQ, projects: rowP, contributions: rowP };
  var CTA = { questions: CFG.ask, answers: CFG.questions, projects: CFG.projects, contributions: CFG.projects };

  function panel(k, items, m) {
    if (items.length) return '<ul class="list list-reset" data-testid="profile-list-' + k + '">' + items.map(ROWFN[k]).join('') + '</ul>';
    var pre = m.self ? 'profile.empty.self.' : 'profile.empty.';
    return '<div class="id-empty" data-testid="profile-empty-' + k + '"><span class="icon i-' + (k === 'projects' || k === 'contributions' ? 'hammer' : 'message') + '" aria-hidden="true" data-testid="profile-empty-' + k + '-icon"></span><p data-testid="profile-empty-' + k + '-text">' + esc(t(pre + k)) + '</p>' +
      (m.self ? '<a class="btn btn-sm" href="' + CTA[k].replace(/^pages\//, '') + '" data-testid="profile-empty-' + k + '-cta">' + esc(t('profile.empty.self.cta.' + k)) + '</a>' : '') + '</div>';
  }

  function render() {
    var m = model();
    if (!m) { root.innerHTML = '<div id="profile-nf" data-testid="profile-notfound"></div>'; A.Sim.show(document.getElementById('profile-nf'), 'notfound', { titleKey: 'profile.notfound.title', bodyKey: 'profile.notfound.body', icon: 'user' }); document.title = t('profile.title'); return; }
    document.title = t('profile.titleOf', { name: m.name });
    var meta = [];
    if (m.country) meta.push('<span class="meta-item"><span class="icon i-map-pin" aria-hidden="true" data-testid="profile-country-icon"></span>' + cc(m.country, 'profile-country') + '<span data-testid="profile-country-label">' + esc(M.countryName(m.country)) + (m.city ? ', ' + esc(m.city) : '') + '</span></span>');
    if (m.langs.length) meta.push('<span class="meta-item"><span class="icon i-languages" aria-hidden="true" data-testid="profile-langs-icon"></span><span data-testid="profile-langs">' + m.langs.map(function (l) { return esc(LN[l] || l); }).join(', ') + '</span></span>');
    if (m.since) meta.push('<span class="meta-item"><span class="icon i-clock" aria-hidden="true" data-testid="profile-since-icon"></span><span data-testid="profile-since">' + esc(t('profile.since', { date: monthYear(m.since) })) + '</span></span>');
    var pills = (m.collab ? A.statePill('open', { testid: 'profile-pill-collab', labelKey: 'profile.open.collab' }) : '') + (m.mentor ? A.statePill('open', { testid: 'profile-pill-mentor', labelKey: 'profile.open.mentorat' }) : '');
    var col = collections(m), counts = KEYS.map(function (k) { return col[k].length; });
    var tabs = KEYS.map(function (k, i) {
      return '<button class="tab" role="tab" id="profile-tab-' + k + '" aria-controls="profile-panel-' + k + '" aria-selected="' + (i === tab) + '" data-testid="profile-tab-' + k + '"><span data-i18n="profile.tab.' + k + '" data-testid="profile-tab-' + k + '-label"></span><span class="chip-count tabular-nums" data-testid="profile-tab-' + k + '-count">' + counts[i] + '</span></button>';
    }).join('');
    var panels = KEYS.map(function (k, i) { return '<div class="tabpanel" role="tabpanel" id="profile-panel-' + k + '" aria-labelledby="profile-tab-' + k + '" tabindex="0" data-testid="profile-panel-' + k + '"' + (i === tab ? '' : ' hidden') + '>' + panel(k, col[k], m) + '</div>'; }).join('');
    root.innerHTML =
      '<header class="id-hero" data-testid="profile-hero">' + A.avatar(m.name, { size: 'lg', testid: 'profile-avatar' }).replace('avatar-lg', 'avatar-lg id-avatar') +
      '<div class="id-hero-main"><div class="id-hero-top"><div class="id-hero-id"><h1 lang="' + (m.person ? m.person.lang : I18N.lang) + '" data-user-content data-testid="profile-name">' + esc(m.name) + '</h1><p class="id-handle" data-testid="profile-handle">@' + esc(m.handle) + '</p></div>' +
      (m.self ? '<a class="btn btn-icon tip-start" href="parametres.html#profil" aria-label="' + esc(t('profile.edit')) + '" data-testid="profile-edit"><span class="icon i-edit" aria-hidden="true" data-testid="profile-edit-icon"></span></a>' : '') + '</div>' +
      (meta.length ? '<div class="id-hero-meta" data-testid="profile-meta">' + meta.join('') + '</div>' : '') +
      (m.bio ? '<p class="id-bio" lang="' + m.bioLang + '" data-user-content data-testid="profile-bio">' + esc(m.bio) + '</p>' : '') +
      (m.skills.length ? tagsHtml(m.skills, 'profile-skills', 12) : '') +
      ((pills || m.github) ? '<div class="id-links" data-testid="profile-links">' + (pills ? '<div class="id-pills" data-testid="profile-pills">' + pills + '</div>' : '') +
        (m.github ? '<a class="link-external" href="' + esc(m.github.href) + '" rel="noopener noreferrer" data-testid="profile-github"><span class="icon i-git-branch" aria-hidden="true" data-testid="profile-github-icon"></span><span data-testid="profile-github-label">' + esc(m.github.label) + '</span></a>' : '') + '</div>' : '') + '</div></header>' +
      (m.self ? '<p class="id-privacy" data-testid="profile-privacy"><span class="icon i-lock" aria-hidden="true" data-testid="profile-privacy-icon"></span><span><span data-testid="profile-privacy-text">' + esc(t('profile.privacy')) + '</span> <a href="parametres.html#confidentialite" data-testid="profile-privacy-link">' + esc(t('profile.privacy.link')) + '</a></span></p>' : '') +
      '<section class="id-tabs" aria-labelledby="profile-activity-h" data-testid="profile-activity"><h2 class="visually-hidden" id="profile-activity-h" data-i18n="profile.tabs" data-testid="profile-activity-title"></h2>' +
      '<div data-tabs data-testid="profile-tabs"><div class="id-tabs-bar"><div class="tablist" role="tablist" data-i18n-attr="aria-label:profile.tabs" data-testid="profile-tablist">' + tabs + '</div><div id="profile-view" data-testid="profile-view"></div></div>' + panels + '</div></section>';
    I18N.apply(root); A.Tabs.init(root);
    if (A.View) A.View.init(document.getElementById('profile-view'), '#profile-root .list', 'profile');
    root._m = m;
  }

  root.addEventListener('tabs:change', function (e) { tab = e.detail.index; });
  if (model() && model().self) A.setupChecklist('#profile-setup', { testid: 'setup' });
  document.addEventListener('i18n:change', render);
  document.addEventListener('session:change', function () { if (!id || (S.user && root._m && root._m.self)) render(); });
  render();
})();
