/* Injects the shared app shell (header, primary nav, bottom bar, footer) around <main class="page">.
   Page contract: <body class="app" data-active="solve|discover|build|identity|connect|explore" data-root="../"> (data-root="" for index.html).
   Run synchronously at the end of <body> (no defer). Single source of truth for navigation: edit DOMAINS here only. */
(function () {
  'use strict';
  var b = document.body, R = b.getAttribute('data-root') || '', active = b.getAttribute('data-active') || '';
  /* status: p0 = live in prototype; preview = roadmap page exists (Connect); later = P1/P2 present-but-not-built (no page) */
  var DOMAINS = [
    { id: 'explore', icon: 'home', href: 'index.html', st: 'p0' },
    { id: 'solve', icon: 'help', href: 'pages/solve.html', st: 'p0' },
    { id: 'discover', icon: 'compass', href: 'pages/discover.html', st: 'p0' },
    { id: 'build', icon: 'hammer', href: 'pages/build.html', st: 'p0' },
    { id: 'identity', icon: 'users', href: 'pages/profile.html', st: 'p0' },
    { id: 'connect', icon: 'network', href: 'pages/connect.html', st: 'preview' },
    { id: 'language', icon: 'languages', st: 'later' },
    { id: 'access', icon: 'accessibility', st: 'later' },
    { id: 'intelligence', icon: 'sparkles', st: 'later' }
  ];
  function t(k) { return 'data-i18n="' + k + '"'; }
  function item(x, i) {
    var lab = '<span class="icon i-' + x.icon + '" aria-hidden="true"></span><span class="nav-label" ' + t('nav.' + x.id) + '></span>';
    var tid = 'data-testid="nav-' + x.id + '"';
    if (x.st === 'later') return '<li><span class="nav-item is-disabled" aria-disabled="true" ' + tid + '>' + lab + '<span class="badge badge-later" ' + t('nav.later') + ' data-testid="nav-' + x.id + '-badge"></span><span class="visually-hidden" ' + t('nav.' + x.id + '.why') + '></span></span></li>';
    var cur = x.id === active ? ' aria-current="page"' : '';
    return '<li><a class="nav-item' + (x.st === 'preview' ? ' is-preview' : '') + '" href="' + R + x.href + '"' + cur + ' ' + tid + '>' + lab +
      (x.st === 'preview' ? '<span class="badge badge-preview" ' + t('nav.preview') + ' data-testid="nav-' + x.id + '-badge"></span>' : '') + '</a></li>';
  }
  var main = DOMAINS.filter(function (x) { return x.id !== 'explore'; });
  var header =
    '<header class="app-header" data-testid="app-header">' +
    '<a class="brand" href="' + R + 'index.html" data-testid="brand" aria-label="226os"><span class="brand-mark">226</span><span class="brand-os">os</span></a>' +
    '<span class="header-spacer"></span>' +
    '<form class="header-search" role="search" action="' + R + 'index.html" data-testid="header-search-form">' +
    '<label class="visually-hidden" for="global-q" ' + t('header.search.label') + '></label>' +
    '<div class="search-field"><span class="icon i-search" aria-hidden="true"></span><input id="global-q" class="input" type="search" name="q" autocomplete="off" data-testid="header-search-input" data-i18n-attr="placeholder:header.search.placeholder"></div></form>' +
    '<div class="header-actions">' +
    '<div class="segmented" role="group" data-i18n-attr="aria-label:header.lang" data-testid="lang-toggle"><button type="button" data-action="lang-set" data-lang="fr" lang="fr" aria-pressed="true" data-testid="lang-fr" aria-label="Français">FR</button><button type="button" data-action="lang-set" data-lang="en" lang="en" aria-pressed="false" data-testid="lang-en" aria-label="English">EN</button></div>' +
    '<button type="button" class="btn btn-icon theme-toggle" data-action="theme-toggle" aria-pressed="false" data-i18n-attr="aria-label:header.theme" data-testid="theme-toggle"><span class="icon i-sun" aria-hidden="true"></span><span class="icon i-moon" aria-hidden="true"></span></button>' +
    '<a class="btn btn-primary header-create" href="' + R + 'pages/solve.html#ask" data-testid="header-create"><span class="icon i-plus" aria-hidden="true"></span><span class="header-create-label" ' + t('header.create') + '></span></a>' +
    '</div></header>';
  var aside =
    '<div class="nav-scrim" data-action="nav-close" data-testid="nav-scrim"></div>' +
    '<aside class="side-nav" id="side-nav" data-testid="side-nav"><div class="nav-drawer-head"><span class="eyebrow" ' + t('nav.menu') + '></span>' +
    '<button type="button" class="btn btn-ghost btn-icon" data-action="nav-close" data-i18n-attr="aria-label:common.close" data-testid="nav-close"><span class="icon i-x" aria-hidden="true"></span></button></div>' +
    '<nav data-i18n-attr="aria-label:nav.label" data-testid="primary-nav"><ul class="nav-list list-reset">' + item(DOMAINS[0]) +
    '</ul><p class="nav-group-label" ' + t('nav.group') + '></p><ul class="nav-list list-reset">' + main.map(item).join('') + '</ul></nav></aside>';
  function bb(x) { return '<a href="' + R + x.href + '"' + (x.id === active ? ' aria-current="page"' : '') + ' data-testid="bottom-' + x.id + '"><span class="icon i-' + x.icon + '" aria-hidden="true"></span><span ' + t('nav.' + x.id) + '></span></a>'; }
  var bottom = '<nav class="bottom-nav" data-i18n-attr="aria-label:nav.bottom" data-testid="bottom-nav">' + DOMAINS.slice(0, 4).map(bb).join('') +
    '<button type="button" data-action="nav-toggle" aria-expanded="false" aria-controls="side-nav" data-testid="bottom-more"><span class="icon i-menu" aria-hidden="true"></span><span ' + t('nav.more') + '></span></button></nav>';
  var footer = '<footer class="app-footer" data-testid="app-footer"><span class="badge badge-sim" ' + t('sim.badge') + ' data-testid="footer-sim"></span><span ' + t('footer.note') + ' data-testid="footer-note"></span></footer>';
  var main_ = document.querySelector('main');
  if (main_) { main_.classList.add('page'); main_.id = main_.id || 'main'; main_.tabIndex = -1; }
  b.insertAdjacentHTML('afterbegin', header + aside);
  var sk = document.querySelector('.skip-link'); if (sk) b.insertBefore(sk, b.firstChild);
  b.insertAdjacentHTML('beforeend', bottom + footer);
})();
