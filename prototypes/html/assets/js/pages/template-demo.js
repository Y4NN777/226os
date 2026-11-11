/* Demo logic for pages/_template.html. Real pages: copy to assets/js/pages/<name>.js (dictionary + render). */
I18N.register({
  fr: { 'tpl.title': 'Modèle de page — 226os', 'tpl.h1': 'Titre de la page', 'tpl.lede': 'Phrase d’introduction courte, jamais de métriques inventées.', 'tpl.action': 'Action principale' },
  en: { 'tpl.title': 'Page template — 226os', 'tpl.h1': 'Page title', 'tpl.lede': 'Short intro sentence; never invented metrics.', 'tpl.action': 'Primary action' }
});
App.Sim.load(App.$('#tpl-list'), {
  load: function () { return MOCK.questions.slice(0, 3); },
  render: function (items, box) {
    box.innerHTML = '<ul class="list list-reset" data-testid="tpl-list">' + items.map(function (q, i) {
      var n = i + 1;
      return '<li class="row" data-testid="tpl-row-' + n + '"><div class="row-lead"></div><div class="row-main"><h3 class="row-title"><a href="#" lang="' + q.lang +
        '" data-user-content data-testid="tpl-link-' + n + '">' + App.esc(q.title) + '</a></h3></div><div class="row-aside"></div></li>';
    }).join('') + '</ul>';
  }
});
