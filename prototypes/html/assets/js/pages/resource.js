/* Discover resource detail logic + UI dictionary (domain: resource). Data: MOCK.discover (fictional). */
I18N.register({
  fr: {
    'resource.title': 'Fiche API — 226os', 'resource.docs': 'Documentation officielle', 'resource.docs.ext': '(nouvel onglet)',
    'resource.tabs.label': 'Sections de la fiche', 'resource.tab.overview': 'Aperçu', 'resource.tab.integration': 'Intégration', 'resource.tab.discussions': 'Discussions',
    'resource.share': 'Copier le lien', 'resource.noReviews': 'Aucun avis',
    'resource.rev.h': 'Avis', 'resource.rev.write': 'Écrire un avis', 'resource.rev.sort': 'Trier les avis', 'resource.rev.sort.recent': 'Plus récents', 'resource.rev.sort.helpful': 'Plus utiles',
    'resource.rev.none': 'Aucun avis pour l’instant. Partagez votre retour d’intégration.',
    'resource.rev.helpful': 'Utile ({n})', 'resource.rev.helpfulLabel': 'Marquer cet avis comme utile ({n})', 'resource.rev.you': 'Vous', 'resource.rev.youInitial': 'V',
    'resource.rev.rating': 'Votre note', 'resource.rev.text': 'Votre retour', 'resource.rev.placeholder': 'Intégration, documentation, support, limites rencontrées…', 'resource.rev.submit': 'Publier l’avis',
    'resource.rev.err.rating': 'Choisissez une note de 1 à 5.', 'resource.rev.err.text': 'Écrivez au moins 20 caractères.',
    'resource.rev.success.title': 'Avis publié', 'resource.rev.success.msg': 'Merci pour votre retour.', 'resource.rev.proto': 'Simulation : l’avis reste sur cette page, rien n’est enregistré.',
    'resource.fictional': 'exemple fictif', 'resource.verifiedOn': 'Vérifié le',
    'resource.status.verified': 'Source vérifiée', 'resource.status.stale': 'À revérifier', 'resource.status.unverified': 'Non vérifié (déclaré par le fournisseur)',
    'resource.status.verified.why': 'Au moins une source officielle consultée il y a moins de 60 jours.',
    'resource.status.stale.why': 'Dernière vérification il y a plus de 60 jours. Relisez la documentation avant de vous y fier.',
    'resource.status.unverified.why': 'Seules des affirmations du fournisseur sont disponibles : aucune source officielle ni confirmation indépendante.',
    'resource.ov.about': 'À propos', 'resource.ov.coverage': 'Couverture', 'resource.ov.countries': 'Pays pris en charge', 'resource.ov.category': 'Catégorie', 'resource.ov.type': 'Type',
    'resource.ov.version': 'Version', 'resource.ov.doclang': 'Langue de la documentation', 'resource.ov.auth': 'Authentification', 'resource.ov.limits': 'Limites connues',
    'resource.ov.status': 'Statut de vérification', 'resource.ov.verified': 'Dernière vérification',
    'resource.src.h': 'Sources et preuves', 'resource.src.lede': 'Chaque information vient d’une source datée. La date de vérification compte plus que les votes.',
    'resource.src.caption': 'Sources de la fiche', 'resource.src.col.source': 'Source', 'resource.src.col.kind': 'Nature', 'resource.src.col.published': 'Publiée', 'resource.src.col.checked': 'Consultée',
    'resource.src.open': 'Ouvrir', 'resource.src.region': 'Tableau des sources, défilement horizontal possible',
    'resource.pit.h': 'Pièges à éviter', 'resource.pit.none': 'Aucun piège signalé.',
    'resource.int.h': 'Notes d’intégration', 'resource.int.sdk': 'SDK et langages', 'resource.int.code': 'Exemple', 'resource.int.codeLabel': 'Extrait de code {l}', 'resource.int.snippetNote': 'Extrait fictif : il illustre le principe, pas une API réelle.',
    'resource.int.auth': 'Authentification',
    'resource.disc.h': 'Questions liées', 'resource.disc.none': 'Aucune question liée.', 'resource.disc.ask': 'Poser une question sur cette API',
    'resource.disc.proj': 'Projets open source liés', 'resource.disc.projNone': 'Aucun projet open source lié.',
    'resource.disc.votes': '{n} votes', 'resource.disc.answers': '{n} réponse(s)', 'resource.disc.accepted': 'Réponse acceptée',
    'resource.disc.feedback': 'Retours d’intégration de la communauté', 'resource.disc.feedback.body': 'La comparaison de fournisseurs viendra plus tard.',
    'resource.linked.h': 'Liens de cette fiche', 'resource.linked.question': 'Question', 'resource.linked.project': 'Projet', 'resource.linked.api': 'API / outil',
    'resource.linked.none': 'Aucun lien pour l’instant.',
    'resource.report.open': 'Signaler une erreur', 'resource.report.title': 'Signaler une information obsolète ou fausse',
    'resource.report.intro': 'Aidez à garder la fiche à jour. Simulation : rien n’est envoyé.',
    'resource.report.summary': 'Corrigez les champs suivants', 'resource.report.kind': 'Nature du problème', 'resource.report.choose': 'Choisir…',
    'resource.report.k.outdated': 'Information obsolète', 'resource.report.k.broken': 'API ou lien qui ne fonctionne plus', 'resource.report.k.coverage': 'Couverture (pays) incorrecte', 'resource.report.k.other': 'Autre',
    'resource.report.detail': 'Ce qui est incorrect', 'resource.report.detail.hint': 'Au moins 20 caractères : indiquez la section, la valeur actuelle et la valeur attendue.',
    'resource.report.source': 'Lien vers une source (facultatif)', 'resource.report.source.hint': 'Une page de documentation ou un changelog qui confirme la correction.',
    'resource.report.submit': 'Envoyer le signalement',
    'resource.report.err.kind': 'Choisissez la nature du problème.', 'resource.report.err.detail': 'Décrivez le problème en au moins 20 caractères.', 'resource.report.err.source': 'Saisissez une adresse commençant par https://.',
    'resource.report.success.title': 'Signalement envoyé', 'resource.report.success.msg': 'Merci, la correction sera relue par la communauté.',
    'resource.notfound.title': 'Fiche introuvable', 'resource.notfound.body': 'Cette fiche n’existe pas ou n’est plus disponible. Retournez au catalogue.', 'resource.back': 'Retour au catalogue',
    'resource.error.title': 'Impossible de charger la fiche', 'resource.error.body': 'La connexion a échoué. Réessayez.',
    'resource.type.api': 'API', 'resource.type.sdk': 'SDK', 'resource.type.tool': 'Outil', 'resource.auth.hmac': 'Signature HMAC', 'resource.auth.apikey': 'Clé d’API', 'resource.auth.oauth2': 'OAuth 2.0', 'resource.auth.none': 'Aucune'
  },
  en: {
    'resource.title': 'API listing — 226os', 'resource.docs': 'Official documentation', 'resource.docs.ext': '(opens in a new tab)',
    'resource.tabs.label': 'Listing sections', 'resource.tab.overview': 'Overview', 'resource.tab.integration': 'Integration', 'resource.tab.discussions': 'Discussions',
    'resource.share': 'Copy link', 'resource.noReviews': 'No reviews',
    'resource.rev.h': 'Reviews', 'resource.rev.write': 'Write a review', 'resource.rev.sort': 'Sort reviews', 'resource.rev.sort.recent': 'Most recent', 'resource.rev.sort.helpful': 'Most helpful',
    'resource.rev.none': 'No reviews yet. Share your integration experience to help other developers.',
    'resource.rev.helpful': 'Helpful · {n}', 'resource.rev.helpfulLabel': 'Mark this review as helpful ({n})', 'resource.rev.you': 'You', 'resource.rev.youInitial': 'Y',
    'resource.rev.rating': 'Your rating', 'resource.rev.text': 'Your feedback', 'resource.rev.placeholder': 'Integration, documentation, support, limits you hit…', 'resource.rev.submit': 'Post review',
    'resource.rev.err.rating': 'Choose a rating from 1 to 5.', 'resource.rev.err.text': 'Write at least 20 characters.',
    'resource.rev.success.title': 'Review posted', 'resource.rev.success.msg': 'Thanks for your feedback.', 'resource.rev.proto': 'Prototype: the review is only added to this page and is not saved.',
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
    'resource.disc.proj': 'Related open-source projects', 'resource.disc.projNone': 'No related open-source project yet.',
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
    'resource.report.success.title': 'Report sent', 'resource.report.success.msg': 'Thanks, the correction will be reviewed by the community.',
    'resource.notfound.title': 'Listing not found', 'resource.notfound.body': 'This listing does not exist or is no longer available. Go back to the catalogue.', 'resource.back': 'Back to the catalogue',
    'resource.error.title': 'Could not load the listing', 'resource.error.body': 'The connection failed. Please retry.',
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
  var mine = [], sortMode = 'recent', helped = {};

  function lbl(o) { return o[I18N.lang]; }
  function cc(code, tid) { return '<span class="country" data-testid="' + tid + '"><span class="country-code" data-testid="' + tid + '-code">' + code + '</span><span class="visually-hidden" data-testid="' + tid + '-name">' + esc(M.countryName(code)) + '</span></span>'; }
  function langBadge(l, tid, doc) { var n = LN[l][I18N.lang]; return '<span class="badge" lang="' + I18N.lang + '" data-testid="' + tid + '"><span aria-hidden="true" data-testid="' + tid + '-code">' + l.toUpperCase() + '</span><span class="visually-hidden" data-testid="' + tid + '-name">' + esc((doc ? t('resource.ov.doclang') : t('common.originalLang')) + (I18N.lang === 'fr' ? '\u00a0: ' : ': ') + n) + '</span></span>'; }
  function date(iso, tid) { return '<time datetime="' + iso + '" data-testid="' + tid + '">' + M.fmtDate(iso) + '</time>'; }
  function provider(x) { return esc(x.provider.replace(/\s*\(fictif\)\s*/, '').replace(/\s*fictif\s*/, ' ').trim()) + '<span data-proto data-testid="resource-fictional-tag"> · ' + esc(t('resource.fictional')) + '</span>'; }

  /* ---- Markdown: shared renderer (assets/js/markdown.js); "##" -> h3 ---- */
  function md(src) { return window.Markdown.render(src, { id: 'resource-notes', headStart: 2, copyText: t('common.copy') }); }
  /* ---- tiny highlighter: comments, strings, keywords, calls ---- */
  var RE = /(\/\/.*|#.*)|("(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*'|`(?:[^`\\]|\\.)*`)|\b(func|return|const|let|var|await|async|if|else|for|import|from|def|new|true|false|nil|null|function|print)\b|\b([A-Za-z_]\w*)(?=\()/g;
  function hl(code) {
    var out = '', last = 0, m, tk = 0; RE.lastIndex = 0;
    while ((m = RE.exec(code))) {
      out += esc(code.slice(last, m.index));
      var c = m[1] ? 'com' : m[2] ? 'str' : m[3] ? 'kw' : 'fn';
      out += '<span class="tok-' + c + '" data-testid="resource-code-tok-' + (++tk) + '">' + esc(m[0]) + '</span>'; last = m.index + m[0].length;
    }
    return out + esc(code.slice(last));
  }
  var LANGNAME = { go: 'Go', js: 'JavaScript', py: 'Python', sh: 'Shell', php: 'PHP' };

  function kv(rows) { return '<dl class="dl-meta discover-dl">' + rows.map(function (x) { return '<dt data-testid="' + x[0] + '-term">' + esc(x[1]) + '</dt><dd data-testid="' + x[0] + '-value">' + x[2] + '</dd>'; }).join('') + '</dl>'; }

  function overview() {
    var s = D.status(r), c = D.cat(r.category);
    var h = '<div class="stack-lg">';
    h += '<section aria-labelledby="ov-cov-h" data-testid="resource-coverage-section"><h2 id="ov-cov-h" class="h3" data-testid="resource-coverage-h">' + esc(t('resource.ov.coverage')) + '</h2>' +
      kv([['resource-countries', t('resource.ov.countries'), '<span class="cluster">' + r.countries.map(function (x) { return cc(x, 'resource-country-' + x.toLowerCase()) + ' <span class="text-sm" data-testid="resource-country-name-' + x.toLowerCase() + '">' + esc(M.countryName(x)) + '</span>'; }).join('') + '</span>'],
        ['resource-category', t('resource.ov.category'), esc(lbl(c))],
        ['resource-type', t('resource.ov.type'), esc(t('resource.type.' + r.type))],
        ['resource-version', t('resource.ov.version'), '<span class="mono" data-testid="resource-version-text">' + esc(r.version) + '</span>'],
        ['resource-doclang', t('resource.ov.doclang'), '<span class="cluster">' + r.docLang.map(function (l) { return langBadge(l, 'resource-doclang-' + l, true); }).join('') + '</span>'],
        ['resource-auth', t('resource.ov.auth'), '<strong data-testid="resource-auth-model">' + esc(t('resource.auth.' + r.auth.model)) + '</strong><br><span lang="' + r.auth.lang + '" data-user-content data-testid="resource-auth-detail">' + esc(r.auth.text) + '</span> ' + langBadge(r.auth.lang, 'resource-auth-lang')],
        ['resource-limits', t('resource.ov.limits'), '<span lang="' + r.limits.lang + '" data-user-content data-testid="resource-limits-text">' + esc(r.limits.text) + '</span> ' + langBadge(r.limits.lang, 'resource-limits-lang')]]) + '</section>';

    h += '<section data-proto aria-labelledby="ov-status-h" data-testid="resource-status-section"><h2 id="ov-status-h" class="visually-hidden" data-testid="resource-status-h">' + esc(t('resource.ov.status')) + '</h2>' +
      '<div class="callout' + (s === 'verified' ? ' callout-success' : ' callout-warn') + '" data-testid="resource-status-callout"><span class="icon icon-lead i-' + STATUS[s].icon + '" aria-hidden="true" data-testid="resource-status-callout-icon"></span><div><strong data-testid="resource-status-title">' + esc(t('resource.status.' + s)) + '</strong><span data-testid="resource-status-why">' + esc(t('resource.status.' + s + '.why')) + '</span></div></div></section>';
    h += '<section aria-labelledby="ov-src-h" data-testid="resource-sources-section"><h2 id="ov-src-h" class="h3" data-testid="resource-sources-h">' + esc(t('resource.src.h')) + '</h2><p class="text-sm text-muted" data-proto data-testid="resource-sources-lede">' + esc(t('resource.src.lede')) + '</p>' +
      '<div class="table-wrap" tabindex="0" role="region" aria-label="' + esc(t('resource.src.region')) + '" data-testid="resource-sources-region"><table class="table" data-testid="resource-sources-table"><caption class="visually-hidden" data-testid="resource-sources-caption">' + esc(t('resource.src.caption')) + '</caption>' +
      '<thead><tr><th scope="col" data-testid="resource-sources-col-source">' + esc(t('resource.src.col.source')) + '</th><th scope="col" data-testid="resource-sources-col-kind">' + esc(t('resource.src.col.kind')) + '</th><th scope="col" data-testid="resource-sources-col-published">' + esc(t('resource.src.col.published')) + '</th><th scope="col" data-testid="resource-sources-col-checked">' + esc(t('resource.src.col.checked')) + '</th></tr></thead><tbody>' +
      r.sources.map(function (x, i) { var k = D.kinds[x.kind], n = x.id || (i + 1), href = x.internal ? '../' + x.url : x.url;
        return '<tr data-testid="resource-source-row-' + n + '"><th scope="row" data-testid="resource-source-cell-' + n + '"><a href="' + esc(href) + '"' + (x.internal ? '' : ' class="link-external" rel="noopener noreferrer"') + ' lang="' + x.lang + '" data-user-content data-testid="resource-source-link-' + n + '">' + esc(x.title) + '</a> ' + langBadge(x.lang, 'resource-source-lang-' + n) + '</th>' +
          '<td data-testid="resource-source-kind-cell-' + n + '"><span class="badge ' + k.cls + '" data-testid="resource-source-kind-' + n + '"><span class="icon i-' + k.icon + '" aria-hidden="true" data-testid="resource-source-kind-icon-' + n + '"></span>' + esc(lbl(k)) + '</span></td><td data-testid="resource-source-published-' + n + '">' + date(x.published, 'resource-source-published-date-' + n) + '</td><td data-testid="resource-source-checked-' + n + '">' + date(x.checked, 'resource-source-checked-date-' + n) + '</td></tr>'; }).join('') + '</tbody></table></div></section>';
    h += '<section aria-labelledby="ov-pit-h" data-testid="resource-pitfalls-section"><h2 id="ov-pit-h" class="h3" data-testid="resource-pitfalls-h">' + esc(t('resource.pit.h')) + '</h2>' +
      (r.pitfalls.length ? '<ul class="stack-sm list-reset" data-testid="resource-pitfalls">' + r.pitfalls.map(function (x, i) { return '<li data-testid="resource-pitfall-' + (i + 1) + '"><span class="icon i-alert" aria-hidden="true" data-testid="resource-pitfall-icon-' + (i + 1) + '"></span> <span lang="' + x.lang + '" data-user-content data-testid="resource-pitfall-text-' + (i + 1) + '">' + esc(x.text) + '</span> ' + langBadge(x.lang, 'resource-pitfall-lang-' + (i + 1)) + '</li>'; }).join('') + '</ul>' : '<p class="text-muted" data-testid="resource-pitfalls-none">' + esc(t('resource.pit.none')) + '</p>') + '</section>';
    return h + '</div>';
  }

  function integration() {
    var code = r.snippet, h = '<div class="stack-lg">';
    h += '<section aria-labelledby="int-notes-h" data-testid="resource-notes-section"><h2 id="int-notes-h" class="h3" data-testid="resource-notes-h">' + esc(t('resource.int.h')) + ' ' + langBadge(r.notesLang, 'resource-notes-lang') + '</h2>' +
      '<div class="prose" lang="' + r.notesLang + '" data-user-content data-testid="resource-notes">' + md(r.notes) + '</div></section>';
    h += '<section aria-labelledby="int-code-h" data-testid="resource-code-section"><h2 id="int-code-h" class="h3" data-testid="resource-code-h">' + esc(t('resource.int.code')) + '</h2>' +
      '<div class="code-block" data-testid="resource-code-block"><div class="code-head"><span data-testid="resource-code-lang">' + LANGNAME[code.lang] + '</span>' +
      '<button type="button" class="btn btn-icon btn-sm" data-copy="#resource-code" aria-label="' + esc(t('common.copy')) + '" data-testid="resource-code-copy"><span class="icon i-copy icon-main" aria-hidden="true" data-testid="resource-code-copy-icon"></span><span class="icon i-check icon-alt" aria-hidden="true" data-testid="resource-code-copy-icon-done"></span></button></div>' +
      '<pre tabindex="0" aria-label="' + esc(t('resource.int.codeLabel', { l: LANGNAME[code.lang] })) + '" data-testid="resource-code-pre"><code id="resource-code" data-testid="resource-code-text">' + hl(code.code) + '</code></pre></div>' +
      '<p class="text-sm text-muted" data-proto data-testid="resource-code-note">' + esc(t('resource.int.snippetNote')) + '</p></section>';
    return h + '</div>';
  }

  function discussions() {
    var qs = (r.questions || []).map(function (q) { return M.byId(M.questions, q); }).filter(Boolean), ps = [];
    var h = '<div class="stack-lg"><section aria-labelledby="disc-q-h" data-testid="resource-questions-section"><div class="section-head"><h2 id="disc-q-h" class="h3" data-testid="resource-questions-h">' + esc(t('resource.disc.h')) + '</h2>' +
      '<a class="btn btn-sm" href="solve.html#ask" data-testid="resource-ask-link"><span class="icon i-plus" aria-hidden="true" data-testid="resource-ask-link-icon"></span><span data-testid="resource-ask-link-label">' + esc(t('resource.disc.ask')) + '</span></a></div>';
    h += qs.length ? '<ul class="list list-reset" data-testid="resource-questions">' + qs.map(function (q) { var n = q.id;
      return '<li class="row" data-testid="resource-question-row-' + n + '"><div class="row-lead"><div class="vote" role="img" data-testid="resource-question-vote-' + n + '" aria-label="' + esc(t('resource.disc.votes', { n: q.votes })) + '"><span class="vote-count" data-testid="resource-question-votes-' + n + '">' + q.votes + '</span></div></div>' +
        '<div class="row-main"><h3 class="row-title"><a href="question.html?id=' + q.id + '" lang="' + q.lang + '" data-user-content data-testid="resource-question-link-' + n + '">' + esc(q.title) + '</a></h3>' +
        '<div class="meta"><span class="meta-item">' + langBadge(q.lang, 'resource-question-lang-' + n) + '</span><span class="meta-item" data-testid="resource-question-answers-' + n + '">' + esc(t('resource.disc.answers', { n: q.answers })) + '</span><span class="meta-item" data-testid="resource-question-date-' + n + '">' + date(q.date, 'resource-question-date-value-' + n) + '</span></div></div>' +
        '<div class="row-aside">' + (q.accepted ? '<span class="badge badge-accepted" data-testid="resource-question-accepted-' + n + '"><span class="icon i-check-circle" aria-hidden="true" data-testid="resource-question-accepted-icon-' + n + '"></span>' + esc(t('resource.disc.accepted')) + '</span>' : '') + '</div></li>'; }).join('') + '</ul>'
      : '<p class="text-muted" data-testid="resource-questions-none">' + esc(t('resource.disc.none')) + '</p>';
    h += '</section>';
    h += '</div>';
    return h;
  }

  function relatedProjects() {
    var ids = (r.projects || []).slice(); M.build.projects().forEach(function (p) { if (p.resource === r.id && ids.indexOf(p.id) < 0) ids.push(p.id); });
    var ps = ids.map(function (i) { return M.build.byId(i); }).filter(Boolean);
    if (!ps.length) return '<p class="text-muted" data-testid="resource-projects-none">' + esc(t('resource.disc.projNone')) + '</p>';
    return '<ul class="discover-oss list-reset" data-testid="resource-projects">' + ps.map(function (p) { var n = p.id;
      return '<li data-testid="resource-project-item-' + n + '"><a class="card discover-oss-card" href="project.html?id=' + n + '" data-testid="resource-project-link-' + n + '">' + A.thumb(p.name, { size: 'md', testid: 'resource-project-thumb-' + n }) +
        '<span class="discover-oss-body"><strong data-testid="resource-project-name-' + n + '">' + esc(p.name) + '</strong><span class="text-sm text-muted discover-oss-sum" lang="' + p.lang + '" data-user-content data-testid="resource-project-summary-' + n + '">' + esc(p.summary) + '</span>' +
        '<span class="mono text-xs text-muted" data-testid="resource-project-license-' + n + '">' + esc(p.license) + '</span></span></a></li>'; }).join('') + '</ul>';
  }

  function linked() {
    var a = [], n = 0;
    (r.questions || []).forEach(function (q) { var x = M.byId(M.questions, q); if (x) a.push('<a class="linked-item" href="question.html?id=' + x.id + '" data-testid="resource-linked-' + x.id + '"><span class="kind" data-testid="resource-linked-kind-' + x.id + '"><span class="icon i-help" aria-hidden="true" data-testid="resource-linked-icon-' + x.id + '"></span>' + esc(t('resource.linked.question')) + '</span><strong lang="' + x.lang + '" data-user-content data-testid="resource-linked-title-' + x.id + '">' + esc(x.title) + '</strong></a>'); });
    return a.length ? a.join('') : '<p class="text-muted" data-testid="resource-linked-none">' + esc(t('resource.linked.none')) + '</p>';
  }

  function head() {
    var s = D.status(r), o = STATUS[s];
    $('#resource-crumb').textContent = r.name;
    $('#resource-thumb').innerHTML = A.thumb(r.name, { size: 'lg', testid: 'resource-thumb' });
    $('#resource-eyebrow').innerHTML = '<span class="icon i-' + TYPE_ICON[r.type] + '" aria-hidden="true" data-testid="resource-eyebrow-icon"></span> ' + esc(t('resource.type.' + r.type)) + ' · ' + esc(lbl(D.cat(r.category)));
    $('#resource-name').innerHTML = esc(r.name) + ' <span class="mono text-muted text-sm" data-testid="resource-version-title">' + esc(r.version) + '</span>';
    $('#resource-provider').innerHTML = provider(r);
    var link = $('#resource-docs-link'); link.href = r.docs; link.setAttribute('aria-label', t('resource.docs') + ' ' + t('resource.docs.ext'));
    var st = D.stats(r.id, mine);
    $('#resource-badges').innerHTML = '<span class="badge ' + o.cls + '" data-testid="resource-status-badge"><span class="icon i-' + o.icon + '" aria-hidden="true" data-testid="resource-status-badge-icon"></span>' + esc(t('resource.status.' + s)) + '</span>' +
      '<span data-testid="resource-verified-date">' + esc(t('resource.verifiedOn')) + ' ' + date(r.verifiedAt, 'resource-verified-date-value') + '</span>' +
      '<span class="badge badge-sim" data-proto data-testid="resource-sim-badge">' + esc(t('resource.fictional')) + '</span>' +
      (st ? '<a href="#resource-reviews" data-testid="resource-rating-link">' + A.rating(st.avg, { count: st.n, testid: 'resource-rating' }) + '</a>' : '<a href="#resource-reviews" data-testid="resource-rating-none">' + esc(t('resource.noReviews')) + '</a>');
    var sm = $('#resource-summary'); sm.textContent = r.summary; sm.lang = r.summaryLang;
    document.title = r.name + ' — 226os';
  }
  function draw() {
    if (!r) return;
    head(); $('#panel-overview').innerHTML = overview(); $('#panel-integration').innerHTML = integration(); $('#panel-discussions').innerHTML = discussions(); $('#resource-linked').innerHTML = linked(); $('#resource-projects-body').innerHTML = relatedProjects(); reviews();
  }

  /* ---- states: forced ?state= / not found, via the shared App.Sim.show (tabs and layout are not rebuilt) ---- */
  function showState(kind) {
    var key = kind === 'error' ? 'resource.error' : 'resource.notfound';
    $('#resource-content').hidden = true; $('#resource-crumb').textContent = kind === 'loading' ? t('state.loading') : t(key + '.title');
    A.Sim.show($('#resource-state'), kind, { key: key, heading: 'h1', skeleton: 4,
      action: kind === 'error' ? null : { href: 'discover.html', icon: 'chevron-left', labelKey: 'resource.back', testid: 'state-back' } });
  }
  var fs = A.Sim.forced();
  if (fs) { r = null; showState(fs === 'empty' ? 'notfound' : fs); document.addEventListener('i18n:change', function () { if (fs !== 'loading') $('#resource-crumb').textContent = t((fs === 'error' ? 'resource.error' : 'resource.notfound') + '.title'); }); }
  else if (!r) { showState('notfound'); document.addEventListener('i18n:change', function () { $('#resource-crumb').textContent = t('resource.notfound.title'); }); }
  else { draw(); document.addEventListener('i18n:change', draw); }

  /* ---- reviews: fictional sample + reviews added in this session (not persisted) ---- */
  function allReviews() { return (D.reviews[r.id] || []).concat(mine); }
  function hn(x) { return x.helpful + (helped[x.id] ? 1 : 0); }
  function reviewList() {
    var l = allReviews().slice().sort(function (a, b) { return sortMode === 'helpful' ? (hn(b) - hn(a)) || (a.date < b.date ? 1 : -1) : (a.date < b.date ? 1 : a.date > b.date ? -1 : 0); });
    return l.map(function (x) {
      var tid = 'resource-review-' + x.id, rv = x.mine ? { person: { name: t('resource.rev.you'), initials: t('resource.rev.youInitial') }, rating: x.rating, date: x.date, lang: x.lang, text: x.text } : x;
      return '<li data-testid="' + tid + '-item">' + A.review(rv, { testid: tid }) +
        '<div class="review-actions" data-testid="' + tid + '-actions"><button type="button" class="btn btn-ghost btn-sm" data-helpful="' + x.id + '" aria-pressed="' + (helped[x.id] ? 'true' : 'false') + '" aria-label="' + esc(t('resource.rev.helpfulLabel', { n: hn(x) })) + '" data-testid="' + tid + '-helpful"><span class="icon i-arrow-up" aria-hidden="true" data-testid="' + tid + '-helpful-icon"></span><span aria-hidden="true" data-testid="' + tid + '-helpful-label">' + esc(t('resource.rev.helpful', { n: hn(x) })) + '</span></button></div></li>';
    }).join('');
  }
  function reviews() {
    var st = D.stats(r.id, mine), box = $('#resource-reviews-body');
    if (!st) { box.innerHTML = '<p class="text-muted" data-testid="resource-reviews-none">' + esc(t('resource.rev.none')) + '</p>'; return; }
    box.innerHTML = '<div class="review-summary" data-testid="resource-reviews-summary">' + A.rating(st.avg, { count: st.n, size: 'lg', testid: 'resource-reviews-rating' }) +
      '<span class="review-bar"><span class="icon i-sort" aria-hidden="true" data-testid="resource-reviews-sort-icon"></span><label class="visually-hidden" for="resource-reviews-sort" data-testid="resource-reviews-sort-label">' + esc(t('resource.rev.sort')) + '</label>' +
      '<select id="resource-reviews-sort" class="select select-quiet" data-testid="resource-reviews-sort"><option value="recent" data-testid="resource-reviews-sort-recent">' + esc(t('resource.rev.sort.recent')) + '</option><option value="helpful" data-testid="resource-reviews-sort-helpful">' + esc(t('resource.rev.sort.helpful')) + '</option></select></span></div>' +
      '<ul class="review-list" id="resource-review-list" data-testid="resource-reviews-list">' + reviewList() + '</ul>';
    $('#resource-reviews-sort').value = sortMode;
  }
  $('#resource-reviews-body').addEventListener('change', function (e) {
    if (e.target.id !== 'resource-reviews-sort') return; sortMode = e.target.value; $('#resource-review-list').innerHTML = reviewList();
  });
  $('#resource-reviews-body').addEventListener('click', function (e) {
    var b = e.target.closest('[data-helpful]'); if (!b) return; var k = b.getAttribute('data-helpful'); helped[k] = !helped[k];
    $('#resource-review-list').innerHTML = reviewList(); var nb = document.querySelector('[data-helpful="' + k + '"]'); if (nb) nb.focus();
  });
  $('#resource-share').addEventListener('click', function () {
    var b = this; function ok() { b.classList.add('is-done'); setTimeout(function () { b.classList.remove('is-done'); }, 1600); A.Toast.show({ kind: 'success', title: t('common.copied'), ttl: 2500 }); }
    function bad() { A.Toast.show({ kind: 'error', title: t('common.copyFailed') }); }
    try { navigator.clipboard.writeText(location.href.split('#')[0]).then(ok, bad); } catch (e) { bad(); }
  });

  /* ---- review dialog ---- */
  var rform = $('#resource-review-form'), rdlg = $('#resource-review-dialog');
  $('#review-rating-slot').innerHTML = A.ratingInput({ name: 'rating', legendKey: 'resource.rev.rating', required: true, testid: 'resource-review-rating-input' });
  var RF = { rating: { err: '#review-rating-error', key: 'resource.rev.err.rating', ok: function () { return !!rform.elements.rating.value; }, focus: function () { return $('#review-rating-slot input'); } },
    text: { err: '#review-text-error', key: 'resource.rev.err.text', ok: function () { return $('#review-text').value.trim().length >= 20; }, focus: function () { return $('#review-text'); } } };
  function rErr(k, bad) {
    var f = RF[k], msg = $(f.err), fld = msg.closest('.field');
    fld.classList.toggle('has-error', bad); f.focus().setAttribute('aria-invalid', bad ? 'true' : 'false'); msg.hidden = !bad; msg.setAttribute('data-err-key', bad ? f.key : '');
    msg.innerHTML = bad ? '<span class="icon i-alert-circle" aria-hidden="true" data-testid="resource-review-' + k + '-error-icon"></span><span data-testid="resource-review-' + k + '-error-text">' + esc(t(f.key)) + '</span>' : '';
  }
  function rSummary(bad) {
    $('#review-summary').hidden = !bad.length;
    $('#review-summary-list').innerHTML = bad.map(function (k) { return '<li><a href="#" data-field="' + k + '" data-testid="resource-review-summary-' + k + '">' + esc(t(RF[k].key)) + '</a></li>'; }).join('');
  }
  rform.addEventListener('submit', function (e) {
    e.preventDefault();
    var bad = Object.keys(RF).filter(function (k) { var b = !RF[k].ok(); rErr(k, b); return b; }); rSummary(bad);
    if (bad.length) { $('#review-summary').focus(); return; }
    mine.push({ id: 'mine-' + (mine.length + 1), mine: true, rating: +rform.elements.rating.value, date: D.today, lang: I18N.lang, text: $('#review-text').value.trim(), helpful: 0 });
    rform.reset(); rSummary([]); Object.keys(RF).forEach(function (k) { rErr(k, false); });
    if (rdlg.close) rdlg.close();
    sortMode = 'recent'; head(); reviews();
    A.Toast.show({ kind: 'success', title: t('resource.rev.success.title'), msg: t('resource.rev.success.msg'), ttl: 5000 });
  });
  $('#review-summary').addEventListener('click', function (e) { var a = e.target.closest('a[data-field]'); if (a) { e.preventDefault(); RF[a.getAttribute('data-field')].focus().focus(); } });
  rdlg.addEventListener('close', function () { rSummary([]); Object.keys(RF).forEach(function (k) { rErr(k, false); }); });
  document.addEventListener('i18n:change', function () {
    Object.keys(RF).forEach(function (k) { var m = $(RF[k].err); if (!m.hidden) m.lastElementChild.textContent = t(RF[k].key); });
    var bad = Object.keys(RF).filter(function (k) { return $(RF[k].err).hidden === false; }); if (!$('#review-summary').hidden) rSummary(bad);
  });
  if (!r) { var wb = document.querySelector('[data-testid="resource-review-open"]'); if (wb) wb.disabled = true; }

  /* ---- report dialog: validation + simulated success. Form is never rebuilt on language change. ---- */
  var form = $('#resource-report-form'), dlg = $('#resource-report-dialog');
  var F = { kind: { el: '#report-kind', err: '#report-kind-error', key: 'resource.report.err.kind', ok: function (v) { return !!v; } },
    detail: { el: '#report-detail', err: '#report-detail-error', key: 'resource.report.err.detail', ok: function (v) { return v.trim().length >= 20; } },
    source: { el: '#report-source', err: '#report-source-error', key: 'resource.report.err.source', ok: function (v) { return !v || /^https:\/\/\S+$/.test(v.trim()); } } };
  function setErr(k, bad) {
    var f = F[k], inp = $(f.el), msg = $(f.err), fld = inp.closest('.field');
    fld.classList.toggle('has-error', bad); inp.setAttribute('aria-invalid', bad ? 'true' : 'false'); msg.hidden = !bad; msg.setAttribute('data-err-key', bad ? f.key : '');
    msg.innerHTML = bad ? '<span class="icon i-alert-circle" aria-hidden="true" data-testid="resource-' + f.err.slice(1) + '-icon"></span><span data-testid="resource-' + f.err.slice(1) + '-text">' + esc(t(f.key)) + '</span>' : '';
  }
  function summary(bad) {
    var box = $('#report-summary'); box.hidden = !bad.length;
    $('#report-summary-list').innerHTML = bad.map(function (k) { return '<li><a href="#' + F[k].el.slice(1) + '" data-field="' + k + '" data-testid="resource-report-summary-' + k + '">' + esc(t(F[k].key)) + '</a></li>'; }).join('');
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
