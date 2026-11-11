/* Formation: series overview with ordered lessons, session-backed ticks, start/continue. Namespace: frm.* */
I18N.register({
  fr: {
    'frm.title': 'Formation · 226os', 'frm.crumb.learn': 'Apprendre',
    'frm.lessons.title': 'Les leçons', 'frm.next': 'Prochaine leçon',
    'frm.tick': 'Marquer la leçon {n} comme terminée', 'frm.untick': 'Leçon {n} terminée, annuler',
    'frm.prereq': 'Prérequis : {p}',
    'frm.signin': 'Connectez-vous pour suivre votre progression.', 'frm.signin.cta': 'Se connecter',
    'frm.complete': 'Formation terminée. Bravo, vous pouvez maintenant partager ce que vous avez appris.',
    'frm.outcomes': 'Ce que vous saurez faire', 'frm.prereqs': 'Avant de commencer',
    'frm.about': 'À propos', 'frm.about.body': 'Cette formation est offerte par l’équipe 226os. Elle reste gratuite, sans compte pour la lire et sans publicité.',
    'frm.invite.title': 'Vous aussi, partagez un tutoriel', 'frm.invite.body': 'Vous maîtrisez un sujet ? Écrivez le tutoriel que vous auriez aimé trouver.', 'frm.invite.cta': 'Écrire un tutoriel',
    'frm.notfound.title': 'Formation introuvable', 'frm.notfound.body': 'Cette formation n’existe pas ou a été retirée.', 'frm.notfound.cta': 'Retour à Apprendre'
  },
  en: {
    'frm.title': 'Training · 226os', 'frm.crumb.learn': 'Learn',
    'frm.lessons.title': 'Lessons', 'frm.next': 'Next lesson',
    'frm.tick': 'Mark lesson {n} as done', 'frm.untick': 'Lesson {n} done, undo',
    'frm.prereq': 'Prerequisites: {p}',
    'frm.signin': 'Sign in to track your progress.', 'frm.signin.cta': 'Sign in',
    'frm.complete': 'Training completed. Well done, you can now share what you learned.',
    'frm.outcomes': 'What you will be able to do', 'frm.prereqs': 'Before you start',
    'frm.about': 'About', 'frm.about.body': 'This training is offered by the 226os team. It stays free, readable without an account and without ads.',
    'frm.invite.title': 'You too, share a tutorial', 'frm.invite.body': 'You know a topic well? Write the tutorial you wish you had found.', 'frm.invite.cta': 'Write a tutorial',
    'frm.notfound.title': 'Training not found', 'frm.notfound.body': 'This training does not exist or was removed.', 'frm.notfound.cta': 'Back to Learn'
  }
});
(function () {
  'use strict';
  var root = document.getElementById('frm-root'); if (!root) return;
  var A = window.App, M = window.MOCK, L = M.learn, S = A.Session, t = function (k, p) { return I18N.t(k, p); }, esc = A.esc;
  var id = A.params.get('id') || 's-ia';
  function ic(n, i) { return '<span class="icon i-' + n + '" aria-hidden="true" data-testid="' + i + '"></span>'; }

  function crumbs(cur) {
    return '<nav class="breadcrumbs" aria-label="' + esc(t('common.breadcrumb')) + '" data-testid="frm-breadcrumbs"><ol><li><a href="apprendre.html" data-testid="frm-crumb-learn">' + esc(t('frm.crumb.learn')) + '</a></li><li><span aria-current="page" data-testid="frm-crumb-current">' + esc(cur) + '</span></li></ol></nav>';
  }
  function render() {
    var s = L.get(id), forced = A.Sim.forced();
    if (!s || s.type !== 'course' || forced === 'notfound' || forced === 'error' || forced === 'loading' || forced === 'empty') {
      root.innerHTML = crumbs(t('ln.type.course')) + '<div id="frm-state" data-state-id="frm-state" data-testid="frm-state"></div>';
      A.Sim.show(document.getElementById('frm-state'), forced && forced !== 'notfound' ? forced : 'notfound', { icon: 'book', titleKey: 'frm.notfound.title', bodyKey: 'frm.notfound.body', action: { labelKey: 'frm.notfound.cta', href: 'apprendre.html', primary: true, testid: 'frm-notfound-cta' } });
      return;
    }
    document.title = L.txt(s.title) + ' · 226os';
    var ls = L.lessons(s.id), n = ls.length, done = L.doneCount(s.id), next = L.nextLesson(s.id), inn = S.signedIn;
    var cta = inn && done && next ? t('ln.continue', { n: next.n }) : inn && done ? t('ln.review') : t('ln.start');
    var go = inn && done && next ? next : ls[0];
    var progress = inn && done ? '<div class="ln-progress" data-testid="frm-progress"><div class="ln-progress-text"><strong data-testid="frm-progress-text">' + esc(t('ln.progress', { done: done, total: n })) + '</strong><span class="tabular-nums" data-testid="frm-progress-pct">' + Math.round(100 * done / n) + ' %</span></div><div class="ln-bar" role="progressbar" aria-valuemin="0" aria-valuemax="' + n + '" aria-valuenow="' + done + '" aria-label="' + esc(t('ln.progress', { done: done, total: n })) + '" data-testid="frm-bar"><span style="width:' + Math.round(100 * done / n) + '%"></span></div></div>' : '';
    var hero = '<header class="ln-course-hero" data-testid="frm-hero"><p class="ln-kicker" data-testid="frm-kicker">' + ic('sparkles', 'frm-kicker-icon') + '<span data-testid="frm-team">' + esc(t('ln.team')) + '</span></p>' +
      '<h1 data-testid="frm-h1">' + esc(L.txt(s.title)) + '</h1><p class="ln-feature-sum" data-testid="frm-summary">' + esc(L.txt(s.summary)) + '</p>' +
      '<p class="ln-facts" data-testid="frm-facts"><span class="meta-item" data-testid="frm-lessons">' + ic('book', 'frm-lessons-icon') + esc(t('ln.lessons', { n: n })) + '</span><span class="meta-item" data-testid="frm-duration">' + ic('clock', 'frm-duration-icon') + esc(L.fmtMin(L.minutesOf(s.id))) + '</span>' +
      '<span class="meta-item" data-testid="frm-level">' + esc(t('ln.level.range', { a: t('ln.level.' + s.level).toLowerCase(), b: t('ln.level.' + s.levelTo).toLowerCase() })) + '</span><span class="meta-item" data-testid="frm-free">' + ic('check-circle', 'frm-free-icon') + esc(t('ln.free')) + '</span>' +
      '<span class="badge" data-testid="frm-lang">FR</span></p>' + progress +
      '<div class="ln-actions" data-testid="frm-actions"><a class="btn btn-primary" href="article.html?id=' + go.id + '" data-testid="frm-start"><span data-testid="frm-start-label">' + esc(cta) + '</span>' + ic('arrow-right', 'frm-start-icon') + '</a></div>' +
      (inn ? (done === n ? '<p class="ln-hint" data-testid="frm-complete">' + ic('check-circle', 'frm-complete-icon') + '<span>' + esc(t('frm.complete')) + '</span></p>' : '') : '<p class="ln-hint" data-testid="frm-signin">' + ic('lock', 'frm-signin-icon') + '<span>' + esc(t('frm.signin')) + ' <a href="' + 'connexion.html?next=' + encodeURIComponent(location.href) + '" data-testid="frm-signin-link">' + esc(t('frm.signin.cta')) + '</a></span></p>') + '</header>';
    var rows = ls.map(function (l) {
      var d = L.isDone(l.id), isNext = inn && next && next.id === l.id && done > 0, k = 'frm-lesson-' + l.n;
      return '<li class="ln-lesson' + (isNext ? ' is-next' : '') + '" data-testid="' + k + '"><button type="button" class="ln-tick" data-tick="' + l.id + '" aria-pressed="' + d + '" aria-label="' + esc(t(d ? 'frm.untick' : 'frm.tick', { n: l.n })) + '" data-testid="' + k + '-tick">' + ic('circle', k + '-tick-icon').replace('icon i-circle', 'icon icon-main i-circle') + ic('check-circle', k + '-tick-done').replace('icon i-check-circle', 'icon icon-alt i-check-circle') + '</button>' +
        '<div class="ln-lesson-body"><h3 data-testid="' + k + '-title"><a href="article.html?id=' + l.id + '" data-testid="' + k + '-link"><span class="tabular-nums" data-testid="' + k + '-n">' + l.n + '.</span> ' + esc(L.txt(l.title)) + '</a></h3><p data-testid="' + k + '-summary">' + esc(L.txt(l.summary)) + '</p>' +
        '<div class="meta"><span class="meta-item" data-testid="' + k + '-level">' + esc(t('ln.level.' + l.level)) + '</span><span class="meta-item" data-testid="' + k + '-minutes">' + ic('clock', k + '-minutes-icon') + esc(L.fmtMin(l.minutes)) + '</span></div>' +
        '<p class="hint" data-testid="' + k + '-prereq">' + esc(t('frm.prereq', { p: L.txt(l.prereq) })) + '</p></div>' +
        '<div class="ln-lesson-aside">' + (isNext ? A.statePill('open', { testid: k + '-next', label: t('frm.next') }) : '') + '</div></li>';
    }).join('');
    var aside = '<div class="ln-course-aside" data-testid="frm-aside"><section class="ln-aside-card" aria-labelledby="frm-out-h" data-testid="frm-outcomes"><h2 id="frm-out-h" data-i18n="frm.outcomes" data-testid="frm-outcomes-title"></h2><ul class="ln-checks" data-testid="frm-outcomes-list">' +
      s.outcomes.map(function (o, i) { return '<li data-testid="frm-outcome-' + (i + 1) + '">' + ic('check', 'frm-outcome-' + (i + 1) + '-icon') + '<span data-testid="frm-outcome-' + (i + 1) + '-text">' + esc(L.txt(o)) + '</span></li>'; }).join('') + '</ul></section>' +
      '<section class="ln-aside-card" aria-labelledby="frm-pre-h" data-testid="frm-prereqs"><h2 id="frm-pre-h" data-i18n="frm.prereqs" data-testid="frm-prereqs-title"></h2><ul class="ln-checks" data-testid="frm-prereqs-list">' +
      s.prereqs.map(function (o, i) { return '<li data-testid="frm-prereq-' + (i + 1) + '">' + ic('info', 'frm-prereq-' + (i + 1) + '-icon') + '<span data-testid="frm-prereq-' + (i + 1) + '-text">' + esc(L.txt(o)) + '</span></li>'; }).join('') + '</ul></section>' +
      '<section class="ln-aside-card" aria-labelledby="frm-abt-h" data-testid="frm-about"><h2 id="frm-abt-h" data-i18n="frm.about" data-testid="frm-about-title"></h2><p class="hint" data-i18n="frm.about.body" data-testid="frm-about-body"></p></section></div>';
    root.innerHTML = crumbs(L.txt(s.title)) + hero + '<div class="ln-course" style="margin-top:var(--space-8)" data-testid="frm-layout"><section aria-labelledby="frm-les-h" data-testid="frm-lessons-section"><h2 class="ln-sec-title" id="frm-les-h" data-testid="frm-lessons-title"><span data-i18n="frm.lessons.title"></span></h2><ol class="ln-lessons list-reset" data-testid="frm-lessons-list">' + rows + '</ol></section>' + aside + '</div>' +
      '<section class="ln-invite ln-invite-compact" aria-labelledby="frm-inv-h" data-testid="frm-invite"><span class="ln-icon-badge" aria-hidden="true" data-testid="frm-invite-badge">' + ic('edit', 'frm-invite-icon') + '</span><div><h2 id="frm-inv-h" data-i18n="frm.invite.title" data-testid="frm-invite-title"></h2><p data-i18n="frm.invite.body" data-testid="frm-invite-body"></p></div>' +
      '<a class="btn btn-primary" href="ecrire.html?type=tutorial" data-write data-testid="frm-invite-write">' + ic('edit', 'frm-invite-write-icon') + '<span data-i18n="frm.invite.cta" data-testid="frm-invite-write-label"></span></a></section>';
    I18N.apply(root);
  }
  root.addEventListener('click', function (e) {
    var b = e.target.closest('[data-tick]');
    if (b) { if (!S.requireAuth()) return; L.setDone(b.getAttribute('data-tick'), b.getAttribute('aria-pressed') !== 'true'); return; }
    var w = e.target.closest('[data-write]');
    if (w) { e.preventDefault(); var tg = new URL(w.getAttribute('href'), location.href).href; if (S.requireAuth(tg)) location.href = tg; }
  });
  document.addEventListener('learn:progress', render); document.addEventListener('i18n:change', render); document.addEventListener('session:change', render);
  render();
})();
