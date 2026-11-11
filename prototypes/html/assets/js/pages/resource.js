/* Discover resource detail logic + UI dictionary (domain: resource). Data: MOCK.discover (fictional). */
I18N.register({
  fr: {
    'resource.title': 'Fiche API — 226os', 'resource.docs': 'Documentation officielle', 'resource.docs.ext': '(nouvel onglet)',
    'resource.tabs.label': 'Sections de la fiche', 'resource.tab.overview': 'Aperçu', 'resource.tab.integration': 'Intégration', 'resource.tab.discussions': 'Discussions',
    'resource.fictional': 'exemple fictif', 'resource.verifiedOn': 'Vérifié le',
    'resource.status.verified': 'Source vérifiée', 'resource.status.stale': 'Vérification ancienne', 'resource.status.unverified': 'Non vérifié (déclaratif)',
    'resource.status.verified.why': 'Au moins une source officielle consultée il y a moins de 60 jours.',
    'resource.status.stale.why': 'La dernière vérification date de plus de 60 jours : vérifiez la documentation avant de vous y fier.',
    'resource.status.unverified.why': 'Seules des affirmations du fournisseur sont disponibles : aucune source officielle ni confirmation indépendante.',
    'resource.ov.about': 'À propos', 'resource.ov.coverage': 'Couverture', 'resource.ov.countries': 'Pays pris en charge', 'resource.ov.category': 'Catégorie', 'resource.ov.type': 'Type',
    'resource.ov.version': 'Version', 'resource.ov.doclang': 'Langue de la documentation', 'resource.ov.auth': 'Modèle d’authentification', 'resource.ov.limits': 'Limites connues',
    'resource.ov.status': 'Statut de vérification', 'resource.ov.verified': 'Dernière vérification',
    'resource.src.h': 'Sources et preuves', 'resource.src.lede': 'Chaque information provient d’une source datée. Les confirmations et la date de vérification comptent plus que les votes.',
    'resource.src.caption': 'Sources de la fiche', 'resource.src.col.source': 'Source', 'resource.src.col.kind': 'Nature', 'resource.src.col.published': 'Publiée', 'resource.src.col.checked': 'Consultée',
    'resource.src.open': 'Ouvrir', 'resource.src.region': 'Tableau des sources, défilement horizontal possible',
    'resource.pit.h': 'Pièges connus', 'resource.pit.none': 'Aucun piège signalé pour le moment.',
    'resource.int.h': 'Notes d’intégration', 'resource.int.sdk': 'SDK et langages', 'resource.int.code': 'Exemple', 'resource.int.codeLabel': 'Extrait de code {l}', 'resource.int.snippetNote': 'Extrait fictif : il illustre le principe, pas une API réelle.',
    'resource.int.auth': 'Authentification',
    'resource.disc.h': 'Questions liées', 'resource.disc.none': 'Aucune question liée pour le moment.', 'resource.disc.ask': 'Poser une question sur cette API',
    'resource.disc.proj': 'Projets liés', 'resource.disc.projNone': 'Aucun projet lié pour le moment.',
    'resource.disc.votes': '{n} votes', 'resource.disc.answers': '{n} réponse(s)', 'resource.disc.accepted': 'Réponse acceptée',
    'resource.disc.feedback': 'Retours d’intégration de la communauté', 'resource.disc.feedback.body': 'Avis, retours d’expérience et comparaison de fournisseurs sont prévus ensuite ; ils n’existent pas encore dans ce prototype.',
    'resource.linked.h': 'Relié à cette fiche', 'resource.linked.question': 'Question', 'resource.linked.project': 'Projet', 'resource.linked.api': 'API / outil',
    'resource.linked.none': 'Aucun lien pour l’instant.',
    'resource.report.open': 'Signaler une erreur', 'resource.report.title': 'Signaler une information obsolète ou fausse',
    'resource.report.intro': 'Aidez à garder cette fiche à jour. Dans ce prototype, rien n’est envoyé ni enregistré.',
    'resource.report.summary': 'Corrigez les champs suivants', 'resource.report.kind': 'Nature du problème', 'resource.report.choose': 'Choisir…',
    'resource.report.k.outdated': 'Information obsolète', 'resource.report.k.broken': 'API ou lien qui ne fonctionne plus', 'resource.report.k.coverage': 'Couverture (pays) incorrecte', 'resource.report.k.other': 'Autre',
    'resource.report.detail': 'Ce qui est incorrect', 'resource.report.detail.hint': 'Au moins 20 caractères : indiquez la section, la valeur actuelle et la valeur attendue.',
    'resource.report.source': 'Lien vers une source (facultatif)', 'resource.report.source.hint': 'Une page de documentation ou un changelog qui confirme la correction.',
    'resource.report.submit': 'Envoyer le signalement',
    'resource.report.err.kind': 'Choisissez la nature du problème.', 'resource.report.err.detail': 'Décrivez le problème en au moins 20 caractères.', 'resource.report.err.source': 'Saisissez une adresse commençant par https://.',
    'resource.report.success.title': 'Signalement simulé', 'resource.report.success.msg': 'Merci. Dans un produit réel, la correction serait relue par la communauté ; ici, rien n’est enregistré.',
    'resource.notfound.title': 'Fiche introuvable', 'resource.notfound.body': 'Cette fiche n’existe pas dans les données fictives. Retournez au catalogue.', 'resource.back': 'Retour au catalogue',
    'resource.error.title': 'Impossible de charger la fiche', 'resource.error.body': 'La connexion simulée a échoué. Réessayez.',
    'resource.type.api': 'API', 'resource.type.sdk': 'SDK', 'resource.type.tool': 'Outil', 'resource.auth.hmac': 'Signature HMAC', 'resource.auth.apikey': 'Clé d’API', 'resource.auth.oauth2': 'OAuth 2.0', 'resource.auth.none': 'Aucune'
  },
  en: {
    'resource.title': 'API listing — 226os', 'resource.docs': 'Official documentation', 'resource.docs.ext': '(opens in a new tab)',
    'resource.tabs.label': 'Listing sections', 'resource.tab.overview': 'Overview', 'resource.tab.integration': 'Integration', 'resource.tab.discussions': 'Discussions',
    'resource.fictional': 'fictional sample', 'resource.verifiedOn': 'Verified on',
    'resource.status.verified': 'Source verified', 'resource.status.stale': 'Verification is dated', 'resource.status.unverified': 'Not verified (claims only)',
    'resource.status.verified.why': 'At least one official source was checked less than 60 days ago.',
    'resource.status.stale.why': 'The last verification is older than 60 days: check the documentation before relying on it.',
    'resource.status.unverified.why': 'Only provider claims are available: no official source and no independent confirmation.',
    'resource.ov.about': 'About', 'resource.ov.coverage': 'Coverage', 'resource.ov.countries': 'Supported countries', 'resource.ov.category': 'Category', 'resource.ov.type': 'Type',
    'resource.ov.version': 'Version', 'resource.ov.doclang': 'Documentation language', 'resource.ov.auth': 'Authentication model', 'resource.ov.limits': 'Known limits',
    'resource.ov.status': 'Verification status', 'resource.ov.verified': 'Last verified',
    'resource.src.h': 'Sources and evidence', 'resource.src.lede': 'Each piece of information comes from a dated source. Confirmations and the verification date count more than votes.',
    'resource.src.caption': 'Sources for this listing', 'resource.src.col.source': 'Source', 'resource.src.col.kind': 'Nature', 'resource.src.col.published': 'Published', 'resource.src.col.checked': 'Checked',
    'resource.src.open': 'Open', 'resource.src.region': 'Sources table, scrolls horizontally',
    'resource.pit.h': 'Known pitfalls', 'resource.pit.none': 'No pitfall reported yet.',
    'resource.int.h': 'Integration notes', 'resource.int.sdk': 'SDKs and languages', 'resource.int.code': 'Example', 'resource.int.codeLabel': '{l} code snippet', 'resource.int.snippetNote': 'Fictional snippet: it shows the idea, not a real API.',
    'resource.int.auth': 'Authentication',
    'resource.disc.h': 'Linked questions', 'resource.disc.none': 'No linked question yet.', 'resource.disc.ask': 'Ask a question about this API',
    'resource.disc.proj': 'Linked projects', 'resource.disc.projNone': 'No linked project yet.',
    'resource.disc.votes': '{n} votes', 'resource.disc.answers': '{n} answer(s)', 'resource.disc.accepted': 'Accepted answer',
    'resource.disc.feedback': 'Community integration feedback', 'resource.disc.feedback.body': 'Reviews, field reports and provider comparison are planned later; they do not exist yet in this prototype.',
    'resource.linked.h': 'Linked to this listing', 'resource.linked.question': 'Question', 'resource.linked.project': 'Project', 'resource.linked.api': 'API / tool',
    'resource.linked.none': 'No link yet.',
    'resource.report.open': 'Report an error', 'resource.report.title': 'Report outdated or wrong information',
    'resource.report.intro': 'Help keep this listing up to date. In this prototype, nothing is sent or saved.',
    'resource.report.summary': 'Please fix the following fields', 'resource.report.kind': 'Type of issue', 'resource.report.choose': 'Choose…',
    'resource.report.k.outdated': 'Outdated information', 'resource.report.k.broken': 'API or link no longer works', 'resource.report.k.coverage': 'Wrong country coverage', 'resource.report.k.other': 'Other',
    'resource.report.detail': 'What is incorrect', 'resource.report.detail.hint': 'At least 20 characters: name the section, the current value and the expected value.',
    'resource.report.source': 'Link to a source (optional)', 'resource.report.source.hint': 'A documentation page or changelog that confirms the correction.',
    'resource.report.submit': 'Send report',
    'resource.report.err.kind': 'Choose the type of issue.', 'resource.report.err.detail': 'Describe the problem in at least 20 characters.', 'resource.report.err.source': 'Enter an address starting with https://.',
    'resource.report.success.title': 'Simulated report', 'resource.report.success.msg': 'Thank you. In a real product the correction would be reviewed by the community; here nothing is saved.',
    'resource.notfound.title': 'Listing not found', 'resource.notfound.body': 'This listing does not exist in the fictional data. Go back to the catalogue.', 'resource.back': 'Back to the catalogue',
    'resource.error.title': 'Could not load the listing', 'resource.error.body': 'The simulated connection failed. Please retry.',
    'resource.type.api': 'API', 'resource.type.sdk': 'SDK', 'resource.type.tool': 'Tool', 'resource.auth.hmac': 'HMAC signature', 'resource.auth.apikey': 'API key', 'resource.auth.oauth2': 'OAuth 2.0', 'resource.auth.none': 'None'
  }
});

(function () {
  'use strict';
  var A = window.App, M = window.MOCK, D = M.discover, t = I18N.t, esc = A.esc, $ = A.$, P = A.params;
  var LN = { fr: { fr: 'français', en: 'French' }, en: { fr: 'anglais', en: 'English' } };
  var STATUS = { verified: { cls: 'badge-verified', icon: 'shield-check' }, stale: { cls: 'badge-warn', icon: 'clock' }, unverified: { cls: 'badge-warn', icon: 'alert' } };
  var TYPE_ICON = { api: 'network', sdk: 'code', tool: 'hammer' };
  var id = P.get('id') || 'r-kalimba', r = D.get(id), forced = P.get('state');

  function lbl(o) { return o[I18N.lang]; }
  function cc(code, tid) { return '<span class="country" data-testid="' + tid + '"><span class="country-code">' + code + '</span><span class="visually-hidden">' + esc(M.countryName(code)) + '</span></span>'; }
  function langBadge(l, tid, doc) { var n = LN[l][I18N.lang]; return '<span class="badge" lang="' + I18N.lang + '" data-testid="' + tid + '"><span aria-hidden="true">' + l.toUpperCase() + '</span><span class="visually-hidden">' + esc((doc ? t('resource.ov.doclang') : t('common.originalLang')) + ' : ' + n) + '</span></span>'; }
  function date(iso) { return '<time datetime="' + iso + '">' + M.fmtDate(iso) + '</time>'; }
  function provider(x) { return esc(x.provider.replace(/\s*\(fictif\)\s*/, '').replace(/\s*fictif\s*/, ' ').trim()) + ' · ' + esc(t('resource.fictional')); }

  /* ---- minimal Markdown (headings, lists, **bold**, `code`) over escaped text ---- */
  function inline(s) { return esc(s).replace(/`([^`]+)`/g, '<code>$1</code>').replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>'); }
  function md(src) {
    var out = [], list = false;
    src.split('\n').forEach(function (ln) {
      var h = /^##\s+(.*)/.exec(ln), li = /^-\s+(.*)/.exec(ln);
      if (list && !li) { out.push('</ul>'); list = false; }
      if (h) out.push('<h3>' + inline(h[1]) + '</h3>');
      else if (li) { if (!list) { out.push('<ul>'); list = true; } out.push('<li>' + inline(li[1]) + '</li>'); }
      else if (ln.trim()) out.push('<p>' + inline(ln) + '</p>');
    });
    if (list) out.push('</ul>'); return out.join('');
  }
  /* ---- tiny highlighter: comments, strings, keywords, calls ---- */
  var RE = /(\/\/.*|#.*)|("(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*'|`(?:[^`\\]|\\.)*`)|\b(func|return|const|let|var|await|async|if|else|for|import|from|def|new|true|false|nil|null|function|print)\b|\b([A-Za-z_]\w*)(?=\()/g;
  function hl(code) {
    var out = '', last = 0, m; RE.lastIndex = 0;
    while ((m = RE.exec(code))) {
      out += esc(code.slice(last, m.index));
      var c = m[1] ? 'com' : m[2] ? 'str' : m[3] ? 'kw' : 'fn';
      out += '<span class="tok-' + c + '">' + esc(m[0]) + '</span>'; last = m.index + m[0].length;
    }
    return out + esc(code.slice(last));
  }
  var LANGNAME = { go: 'Go', js: 'JavaScript', py: 'Python', sh: 'Shell', php: 'PHP' };

  function kv(rows) { return '<dl class="dl-meta discover-dl">' + rows.map(function (x) { return '<dt data-testid="' + x[0] + '-term">' + esc(x[1]) + '</dt><dd data-testid="' + x[0] + '-value">' + x[2] + '</dd>'; }).join('') + '</dl>'; }

  function overview() {
    var s = D.status(r), c = D.cat(r.category);
    var h = '<div class="stack-lg">';
    h += '<section aria-labelledby="ov-cov-h" data-testid="resource-coverage-section"><h2 id="ov-cov-h" class="h3" data-testid="resource-coverage-h">' + esc(t('resource.ov.coverage')) + '</h2>' +
      kv([['resource-countries', t('resource.ov.countries'), '<span class="cluster">' + r.countries.map(function (x, i) { return cc(x, 'resource-country-' + (i + 1)) + ' <span class="text-sm" data-testid="resource-country-name-' + (i + 1) + '">' + esc(M.countryName(x)) + '</span>'; }).join('') + '</span>'],
        ['resource-category', t('resource.ov.category'), esc(lbl(c))],
        ['resource-type', t('resource.ov.type'), esc(t('resource.type.' + r.type))],
        ['resource-version', t('resource.ov.version'), '<span class="mono">' + esc(r.version) + '</span>'],
        ['resource-doclang', t('resource.ov.doclang'), '<span class="cluster">' + r.docLang.map(function (l, i) { return langBadge(l, 'resource-doclang-' + (i + 1), true); }).join('') + '</span>'],
        ['resource-auth', t('resource.ov.auth'), '<strong>' + esc(t('resource.auth.' + r.auth.model)) + '</strong><br><span lang="' + I18N.lang + '">' + esc(lbl(r.auth)) + '</span>'],
        ['resource-limits', t('resource.ov.limits'), esc(lbl(r.limits))],
        ['resource-lastverified', t('resource.ov.verified'), date(r.verifiedAt)]]) + '</section>';
    h += '<section aria-labelledby="ov-status-h" data-testid="resource-status-section"><h2 id="ov-status-h" class="visually-hidden">' + esc(t('resource.ov.status')) + '</h2>' +
      '<div class="callout' + (s === 'verified' ? ' callout-success' : ' callout-warn') + '" data-testid="resource-status-callout"><span class="icon icon-lead i-' + STATUS[s].icon + '" aria-hidden="true"></span><div><strong data-testid="resource-status-title">' + esc(t('resource.status.' + s)) + '</strong><span data-testid="resource-status-why">' + esc(t('resource.status.' + s + '.why')) + '</span></div></div></section>';
    h += '<section aria-labelledby="ov-src-h" data-testid="resource-sources-section"><h2 id="ov-src-h" class="h3" data-testid="resource-sources-h">' + esc(t('resource.src.h')) + '</h2><p class="text-sm text-muted" data-testid="resource-sources-lede">' + esc(t('resource.src.lede')) + '</p>' +
      '<div class="table-wrap" tabindex="0" role="region" aria-label="' + esc(t('resource.src.region')) + '" data-testid="resource-sources-region"><table class="table"><caption class="visually-hidden">' + esc(t('resource.src.caption')) + '</caption>' +
      '<thead><tr><th scope="col">' + esc(t('resource.src.col.source')) + '</th><th scope="col">' + esc(t('resource.src.col.kind')) + '</th><th scope="col">' + esc(t('resource.src.col.published')) + '</th><th scope="col">' + esc(t('resource.src.col.checked')) + '</th></tr></thead><tbody>' +
      r.sources.map(function (x, i) { var k = D.kinds[x.kind], n = i + 1, href = x.internal ? '../' + x.url : x.url;
        return '<tr data-testid="resource-source-row-' + n + '"><th scope="row"><a href="' + esc(href) + '"' + (x.internal ? '' : ' class="link-external" rel="noopener noreferrer"') + ' lang="' + x.lang + '" data-user-content data-testid="resource-source-link-' + n + '">' + esc(x.title) + '</a> ' + langBadge(x.lang, 'resource-source-lang-' + n) + '</th>' +
          '<td><span class="badge ' + k.cls + '" data-testid="resource-source-kind-' + n + '"><span class="icon i-' + k.icon + '" aria-hidden="true"></span>' + esc(lbl(k)) + '</span></td><td data-testid="resource-source-published-' + n + '">' + date(x.published) + '</td><td data-testid="resource-source-checked-' + n + '">' + date(x.checked) + '</td></tr>'; }).join('') + '</tbody></table></div></section>';
    h += '<section aria-labelledby="ov-pit-h" data-testid="resource-pitfalls-section"><h2 id="ov-pit-h" class="h3" data-testid="resource-pitfalls-h">' + esc(t('resource.pit.h')) + '</h2>' +
      (r.pitfalls.length ? '<ul class="stack-sm list-reset" data-testid="resource-pitfalls">' + r.pitfalls.map(function (x, i) { return '<li data-testid="resource-pitfall-' + (i + 1) + '"><span class="icon i-alert" aria-hidden="true" data-testid="resource-pitfall-icon-' + (i + 1) + '"></span> <span lang="' + x.lang + '" data-user-content>' + esc(x.text) + '</span> ' + langBadge(x.lang, 'resource-pitfall-lang-' + (i + 1)) + '</li>'; }).join('') + '</ul>' : '<p class="text-muted" data-testid="resource-pitfalls-none">' + esc(t('resource.pit.none')) + '</p>') + '</section>';
    return h + '</div>';
  }

  function integration() {
    var code = r.snippet, h = '<div class="stack-lg">';
    h += '<section aria-labelledby="int-notes-h" data-testid="resource-notes-section"><h2 id="int-notes-h" class="h3" data-testid="resource-notes-h">' + esc(t('resource.int.h')) + ' ' + langBadge(r.notesLang, 'resource-notes-lang') + '</h2>' +
      '<div class="prose" lang="' + r.notesLang + '" data-user-content data-testid="resource-notes">' + md(r.notes) + '</div></section>';
    h += '<section aria-labelledby="int-code-h" data-testid="resource-code-section"><h2 id="int-code-h" class="h3" data-testid="resource-code-h">' + esc(t('resource.int.code')) + '</h2>' +
      '<div class="code-block" data-testid="resource-code-block"><div class="code-head"><span data-testid="resource-code-lang">' + LANGNAME[code.lang] + '</span>' +
      '<button type="button" class="btn btn-sm" data-copy="#resource-code" data-testid="resource-code-copy"><span class="icon i-copy" aria-hidden="true"></span>' + esc(t('common.copy')) + '</button></div>' +
      '<pre tabindex="0" aria-label="' + esc(t('resource.int.codeLabel', { l: LANGNAME[code.lang] })) + '" data-testid="resource-code-pre"><code id="resource-code">' + hl(code.code) + '</code></pre></div>' +
      '<p class="text-sm text-muted" data-testid="resource-code-note">' + esc(t('resource.int.snippetNote')) + '</p></section>';
    return h + '</div>';
  }

  function discussions() {
    var qs = (r.questions || []).map(function (q) { return M.byId(M.questions, q); }).filter(Boolean), ps = (r.projects || []).map(function (p) { return M.byId(M.projects, p); }).filter(Boolean);
    var h = '<div class="stack-lg"><section aria-labelledby="disc-q-h" data-testid="resource-questions-section"><div class="section-head"><h2 id="disc-q-h" class="h3" data-testid="resource-questions-h">' + esc(t('resource.disc.h')) + '</h2>' +
      '<a class="btn btn-sm" href="solve.html#ask" data-testid="resource-ask-link"><span class="icon i-plus" aria-hidden="true"></span>' + esc(t('resource.disc.ask')) + '</a></div>';
    h += qs.length ? '<ul class="list list-reset" data-testid="resource-questions">' + qs.map(function (q, i) { var n = i + 1;
      return '<li class="row" data-testid="resource-question-row-' + n + '"><div class="row-lead"><div class="vote" role="img" aria-label="' + esc(t('resource.disc.votes', { n: q.votes })) + '"><span class="vote-count" data-testid="resource-question-votes-' + n + '">' + q.votes + '</span></div></div>' +
        '<div class="row-main"><h3 class="row-title"><a href="question.html?id=' + q.id + '" lang="' + q.lang + '" data-user-content data-testid="resource-question-link-' + n + '">' + esc(q.title) + '</a></h3>' +
        '<div class="meta"><span class="meta-item">' + langBadge(q.lang, 'resource-question-lang-' + n) + '</span><span class="meta-item" data-testid="resource-question-answers-' + n + '">' + esc(t('resource.disc.answers', { n: q.answers })) + '</span><span class="meta-item">' + date(q.date) + '</span></div></div>' +
        '<div class="row-aside">' + (q.accepted ? '<span class="badge badge-accepted" data-testid="resource-question-accepted-' + n + '"><span class="icon i-check-circle" aria-hidden="true"></span>' + esc(t('resource.disc.accepted')) + '</span>' : '') + '</div></li>'; }).join('') + '</ul>'
      : '<p class="text-muted" data-testid="resource-questions-none">' + esc(t('resource.disc.none')) + '</p>';
    h += '</section><section aria-labelledby="disc-p-h" data-testid="resource-projects-section"><h2 id="disc-p-h" class="h3" data-testid="resource-projects-h">' + esc(t('resource.disc.proj')) + '</h2>' +
      (ps.length ? '<div class="linked-items" data-testid="resource-projects">' + ps.map(function (p, i) { return '<a class="linked-item" href="project.html?id=' + p.id + '" data-testid="resource-project-link-' + (i + 1) + '"><span class="kind"><span class="icon i-hammer" aria-hidden="true"></span>' + esc(t('resource.linked.project')) + '</span><strong lang="' + p.lang + '" data-user-content>' + esc(p.name) + '</strong></a>'; }).join('') + '</div>'
        : '<p class="text-muted" data-testid="resource-projects-none">' + esc(t('resource.disc.projNone')) + '</p>') + '</section>';
    h += '<section aria-labelledby="disc-f-h" data-testid="resource-feedback-section"><h2 id="disc-f-h" class="h3" data-testid="resource-feedback-h">' + esc(t('resource.disc.feedback')) + ' <span class="badge badge-later" data-testid="resource-feedback-badge">' + esc(t('nav.later')) + '</span></h2><p class="text-muted text-sm" data-testid="resource-feedback-body">' + esc(t('resource.disc.feedback.body')) + '</p></section></div>';
    return h;
  }

  function linked() {
    var a = [], n = 0;
    (r.questions || []).forEach(function (q) { var x = M.byId(M.questions, q); if (x) a.push('<a class="linked-item" href="question.html?id=' + x.id + '" data-testid="resource-linked-' + (++n) + '"><span class="kind"><span class="icon i-help" aria-hidden="true"></span>' + esc(t('resource.linked.question')) + '</span><strong lang="' + x.lang + '" data-user-content>' + esc(x.title) + '</strong></a>'); });
    (r.projects || []).forEach(function (p) { var x = M.byId(M.projects, p); if (x) a.push('<a class="linked-item" href="project.html?id=' + x.id + '" data-testid="resource-linked-' + (++n) + '"><span class="kind"><span class="icon i-hammer" aria-hidden="true"></span>' + esc(t('resource.linked.project')) + '</span><strong lang="' + x.lang + '" data-user-content>' + esc(x.name) + '</strong></a>'); });
    return a.length ? a.join('') : '<p class="text-muted" data-testid="resource-linked-none">' + esc(t('resource.linked.none')) + '</p>';
  }

  function head() {
    var s = D.status(r), o = STATUS[s];
    $('#resource-crumb').textContent = r.name;
    $('#resource-eyebrow').innerHTML = '<span class="icon i-' + TYPE_ICON[r.type] + '" aria-hidden="true"></span> ' + esc(t('resource.type.' + r.type)) + ' · ' + esc(lbl(D.cat(r.category)));
    $('#resource-name').innerHTML = esc(r.name) + ' <span class="mono text-muted text-sm" data-testid="resource-version-title">' + esc(r.version) + '</span>';
    $('#resource-provider').innerHTML = provider(r);
    var link = $('#resource-docs-link'); link.href = r.docs; link.setAttribute('aria-label', t('resource.docs') + ' ' + t('resource.docs.ext'));
    $('#resource-badges').innerHTML = '<span class="badge ' + o.cls + '" data-testid="resource-status-badge"><span class="icon i-' + o.icon + '" aria-hidden="true"></span>' + esc(t('resource.status.' + s)) + '</span>' +
      '<span class="text-sm text-muted" data-testid="resource-verified-date">' + esc(t('resource.verifiedOn')) + ' ' + date(r.verifiedAt) + '</span>' +
      '<span class="badge badge-sim" data-testid="resource-sim-badge">' + esc(t('resource.fictional')) + '</span>';
    var sm = $('#resource-summary'); sm.textContent = r.summary; sm.lang = r.summaryLang;
    document.title = r.name + ' — 226os';
  }
  function draw() {
    if (!r) return;
    head(); $('#panel-overview').innerHTML = overview(); $('#panel-integration').innerHTML = integration(); $('#panel-discussions').innerHTML = discussions(); $('#resource-linked').innerHTML = linked();
  }

  /* ---- states: not found / forced error ---- */
  function showState(kind) {
    var box = $('#resource-state'), err = kind === 'error', icon = err ? 'alert-circle' : 'inbox', key = err ? 'resource.error' : 'resource.notfound';
    $('#resource-content').hidden = true; $('#resource-crumb').textContent = t(key + '.title');
    box.innerHTML = '<div class="state ' + (err ? 'state-error' : '') + '" data-testid="state-' + kind + '"><span class="icon i-' + icon + '" aria-hidden="true"></span><h1 class="h3">' + esc(t(key + '.title')) + '</h1><p>' + esc(t(key + '.body')) + '</p>' +
      (err ? '<a class="btn btn-sm" href="resource.html?id=' + esc(id) + '" data-testid="state-retry"><span class="icon i-refresh" aria-hidden="true"></span>' + esc(t('state.retry')) + '</a>' : '<a class="btn btn-sm" href="discover.html" data-testid="state-back"><span class="icon i-chevron-left" aria-hidden="true"></span>' + esc(t('resource.back')) + '</a>') + '</div>' +
      '<p class="sim-marker text-sm text-muted"><span class="badge badge-sim" data-testid="sim-marker-badge">' + esc(t('sim.badge')) + '</span> ' + esc(t('sim.note')) + '</p>';
  }
  if (forced === 'error') { r = null; showState('error'); document.addEventListener('i18n:change', function () { showState('error'); }); }
  else if (!r) { showState('notfound'); document.addEventListener('i18n:change', function () { showState('notfound'); }); }
  else { draw(); document.addEventListener('i18n:change', draw); }

  /* ---- report dialog: validation + simulated success. Form is never rebuilt on language change. ---- */
  var form = $('#resource-report-form'), dlg = $('#resource-report-dialog');
  var F = { kind: { el: '#report-kind', err: '#report-kind-error', key: 'resource.report.err.kind', ok: function (v) { return !!v; } },
    detail: { el: '#report-detail', err: '#report-detail-error', key: 'resource.report.err.detail', ok: function (v) { return v.trim().length >= 20; } },
    source: { el: '#report-source', err: '#report-source-error', key: 'resource.report.err.source', ok: function (v) { return !v || /^https:\/\/\S+$/.test(v.trim()); } } };
  function setErr(k, bad) {
    var f = F[k], inp = $(f.el), msg = $(f.err), fld = inp.closest('.field');
    fld.classList.toggle('has-error', bad); inp.setAttribute('aria-invalid', bad ? 'true' : 'false'); msg.hidden = !bad; msg.setAttribute('data-err-key', bad ? f.key : '');
    msg.innerHTML = bad ? '<span class="icon i-alert-circle" aria-hidden="true"></span><span>' + esc(t(f.key)) + '</span>' : '';
  }
  function summary(bad) {
    var box = $('#report-summary'); box.hidden = !bad.length;
    $('#report-summary-list').innerHTML = bad.map(function (k) { return '<li><a href="#' + F[k].el.slice(1) + '" data-field="' + k + '">' + esc(t(F[k].key)) + '</a></li>'; }).join('');
  }
  form.addEventListener('submit', function (e) {
    e.preventDefault();
    var bad = Object.keys(F).filter(function (k) { var v = $(F[k].el).value, b = !F[k].ok(v); setErr(k, b); return b; });
    summary(bad);
    if (bad.length) { $('#report-summary').focus(); return; }
    form.reset(); summary([]); Object.keys(F).forEach(function (k) { setErr(k, false); });
    if (dlg.close) dlg.close();
    A.Toast.show({ kind: 'success', title: t('resource.report.success.title'), msg: t('resource.report.success.msg'), ttl: 7000 });
  });
  $('#report-summary').addEventListener('click', function (e) { var a = e.target.closest('a[data-field]'); if (a) { e.preventDefault(); $(F[a.getAttribute('data-field')].el).focus(); } });
  Object.keys(F).forEach(function (k) { $(F[k].el).addEventListener('blur', function () { if (this.value || this.getAttribute('aria-invalid') === 'true') setErr(k, !F[k].ok(this.value)); }); });
  dlg.addEventListener('close', function () { summary([]); });
  document.addEventListener('i18n:change', function () {
    Object.keys(F).forEach(function (k) { var m = $(F[k].err); if (!m.hidden) m.lastElementChild.textContent = t(F[k].key); });
    var bad = Object.keys(F).filter(function (k) { return $(F[k].err).hidden === false; }); if (!$('#report-summary').hidden) summary(bad);
  });
  if (!r) { var rb = document.querySelector('[data-testid="resource-report-button"]'); if (rb) rb.disabled = true; }
})();
