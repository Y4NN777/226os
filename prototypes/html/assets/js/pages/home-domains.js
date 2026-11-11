/* Explore: domain priority table (PRD 5.x). Runs after home.js and re-renders #domains-body with a status column:
   own pages (P0) / built in across all pages (Language, Access, Intelligence) / designed journey (Connect, P1/P2). */
(function () {
  'use strict';
  var t = function (k) { return I18N.t(k); }, esc = App.esc;
  I18N.register({
    fr: {
      'home.dom.col.status': 'Statut',
      'home.dom.legend': 'Intégré : le P0 de ce domaine est livré dans toutes les pages, il n’a pas d’écran dédié. Parcours conçu : périmètre documenté et parcours conçu (PRD §8), contenu simulé, suite prévue (P1/P2).',
      'home.dom.status.pages': 'Pages dédiées', 'home.dom.status.builtin': 'Intégré à toutes les pages', 'home.dom.status.preview': 'Parcours conçu (P1/P2)',
      'home.dom.identity.now2': 'Création de compte et connexion (simulées), profils minimaux',
      'home.dom.language.now2': 'Interface FR/EN (sélecteur dans l’en-tête) ; le texte original des contenus n’est jamais traduit',
      'home.dom.language.later2': 'Traduction assistée des contenus (P1) ; recherche translingue (P2)',
      'home.dom.access.now2': 'Responsive mobile-first, états chargement/vide/erreur, navigation clavier, pages légères sans média lourd (optimisations réseau à mesurer)',
      'home.dom.access.later2': 'Mode léger explicite et cache amélioré (P1) ; brouillons hors ligne (P2)',
      'home.dom.intelligence.now2': 'Recherche textuelle (en-tête, Explorer) et liens contextuels déterministes question ↔ API ↔ projet',
      'home.dom.intelligence.later2': 'Suggestions simples (P1) ; recherche sémantique, résumés, assistant sourcé (P2)'
    },
    en: {
      'home.dom.col.status': 'Status',
      'home.dom.legend': 'Built in: this domain’s P0 is delivered across all pages and has no dedicated screen. Designed journey: scope documented and journey designed (PRD §8), content simulated, rest planned (P1/P2).',
      'home.dom.status.pages': 'Dedicated pages', 'home.dom.status.builtin': 'Built in across all pages', 'home.dom.status.preview': 'Designed journey (P1/P2)',
      'home.dom.identity.now2': 'Account creation and sign-in (simulated), minimal profiles',
      'home.dom.language.now2': 'FR/EN interface (switch in the header); original content text is never translated',
      'home.dom.language.later2': 'Assisted content translation (P1); cross-language search (P2)',
      'home.dom.access.now2': 'Mobile-first responsive layout, loading/empty/error states, keyboard navigation, light pages without heavy media (network optimisations to be measured)',
      'home.dom.access.later2': 'Explicit light mode and better cache (P1); offline drafts (P2)',
      'home.dom.intelligence.now2': 'Text search (header, Explore) and deterministic contextual links question ↔ API ↔ project',
      'home.dom.intelligence.later2': 'Simple suggestions (P1); semantic search, summaries, sourced assistant (P2)'
    }
  });
  var ST = { solve: 'pages', discover: 'pages', build: 'pages', identity: 'pages', connect: 'preview', language: 'builtin', access: 'builtin', intelligence: 'builtin' };
  var OV = { identity: ['now2'], language: ['now2', 'later2'], access: ['now2', 'later2'], intelligence: ['now2', 'later2'] };
  function key(d, w) { return 'home.dom.' + d + '.' + ((OV[d] || []).indexOf(w + '2') > -1 ? w + '2' : w); }
  function render() {
    var body = document.getElementById('domains-body'); if (!body) return;
    body.innerHTML = Object.keys(ST).map(function (d) {
      var st = ST[d], badge = st === 'pages' ? '<span class="badge" data-testid="domain-status-' + d + '">' + esc(t('home.dom.status.pages')) + '</span>' :
        st === 'builtin' ? '<span class="badge badge-builtin" data-testid="domain-status-' + d + '"><span class="icon i-check" aria-hidden="true" data-testid="domain-status-' + d + '-icon"></span>' + esc(t('home.dom.status.builtin')) + '</span>' :
        '<span class="badge badge-preview" data-testid="domain-status-' + d + '">' + esc(t('home.dom.status.preview')) + '</span>';
      return '<tr data-testid="domain-row-' + d + '"><th scope="row" data-testid="domain-name-' + d + '">' + esc(t('nav.' + d)) + '</th><td data-testid="domain-status-cell-' + d + '">' + badge + '</td><td data-testid="domain-now-' + d + '">' + esc(t(key(d, 'now'))) + '</td><td class="text-muted" data-testid="domain-later-' + d + '">' + esc(t(key(d, 'later'))) + '</td></tr>';
    }).join('');
  }
  render();
  document.addEventListener('i18n:change', render);
  if (document.readyState !== 'complete') document.addEventListener('DOMContentLoaded', render);
})();
