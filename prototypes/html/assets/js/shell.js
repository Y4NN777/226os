/* Injects the shared app shell (header, primary nav, bottom bar, footer) around <main class="page">.
   Page contract: <body class="app" data-active="solve|discover|build|identity|connect|explore" data-root="../"> (data-root="" for index.html).
   Run synchronously at the end of <body> (no defer). Single source of truth for navigation: edit DOMAINS here only. */
(function () {
  'use strict';
  var b = document.body, R = b.getAttribute('data-root') || '', active = b.getAttribute('data-active') || '';
  /* status: p0 = live in prototype with its own pages; preview = designed-journey page exists (Connect, P1/P2); builtin = cross-cutting domain whose P0 is delivered across all pages (Language, Access, Intelligence; links to the domain table, P1/P2 described in a visible note) */
  var DOMAINS = [
    { id: 'explore', icon: 'home', href: 'index.html', st: 'p0' },
    { id: 'solve', icon: 'help', href: 'pages/solve.html', st: 'p0' },
    { id: 'discover', icon: 'compass', href: 'pages/discover.html', st: 'p0' },
    { id: 'build', icon: 'hammer', href: 'pages/build.html', st: 'p0' },
    { id: 'identity', icon: 'users', href: 'pages/profile.html', st: 'p0' },
    { id: 'connect', icon: 'network', href: 'pages/connect.html', st: 'preview' },
    { id: 'language', icon: 'languages', href: 'index.html#domains', st: 'builtin' },
    { id: 'access', icon: 'accessibility', href: 'index.html#domains', st: 'builtin' },
    { id: 'intelligence', icon: 'sparkles', href: 'index.html#domains', st: 'builtin' }
  ];
  function t(k) { return 'data-i18n="' + k + '"'; }
  function item(x, i) {
    var lab = '<span class="icon i-' + x.icon + '" aria-hidden="true" data-testid="nav-' + x.id + '-icon"></span><span class="nav-label" ' + t('nav.' + x.id) + ' data-testid="nav-' + x.id + '-label"></span>';
    var tid = 'data-testid="nav-' + x.id + '"';
    if (x.st === 'builtin') return '<li class="nav-builtin"><a class="nav-item is-builtin" href="' + R + x.href + '" ' + tid + '>' + lab + '</a><div class="nav-note" data-testid="nav-' + x.id + '-note"><span class="badge badge-builtin" data-testid="nav-' + x.id + '-badge"><span class="icon i-check" aria-hidden="true" data-testid="nav-' + x.id + '-badge-icon"></span><span ' + t('nav.builtin') + ' data-testid="nav-' + x.id + '-badge-label"></span></span> <span ' + t('nav.' + x.id + '.why') + ' data-testid="nav-' + x.id + '-note-text"></span></div></li>';
    var cur = x.id === active ? ' aria-current="page"' : '';
    return '<li><a class="nav-item' + (x.st === 'preview' ? ' is-preview' : '') + '" href="' + R + x.href + '"' + cur + ' ' + tid + '>' + lab +
      (x.st === 'preview' ? '<span class="badge badge-preview" ' + t('nav.preview') + ' data-testid="nav-' + x.id + '-badge"></span>' : '') + '</a></li>';
  }
  var main = DOMAINS.filter(function (x) { return x.id !== 'explore'; });
  var header =
    '<header class="app-header" data-testid="app-header">' +
    '<a class="brand" href="' + R + 'index.html" data-testid="brand" aria-label="226os"><span class="brand-mark" data-testid="brand-mark">226</span><span class="brand-os" data-testid="brand-os">os</span></a>' +
    '<span class="header-spacer"></span>' +
    '<form class="header-search" role="search" action="' + R + 'index.html" data-testid="header-search-form">' +
    '<label class="visually-hidden" for="global-q" ' + t('header.search.label') + ' data-testid="header-search-label"></label>' +
    '<div class="search-field"><span class="icon i-search" aria-hidden="true" data-testid="header-search-icon"></span><input id="global-q" class="input" type="search" name="q" autocomplete="off" data-testid="header-search-input" data-i18n-attr="placeholder:header.search.placeholder"></div></form>' +
    '<div class="header-actions">' +
    '<div class="segmented" role="group" data-i18n-attr="aria-label:header.lang" data-testid="lang-toggle"><button type="button" data-action="lang-set" data-lang="fr" lang="fr" aria-pressed="true" data-testid="lang-fr" aria-label="Français">FR</button><button type="button" data-action="lang-set" data-lang="en" lang="en" aria-pressed="false" data-testid="lang-en" aria-label="English">EN</button></div>' +
    '<button type="button" class="btn btn-icon theme-toggle" data-action="theme-toggle" aria-pressed="false" data-i18n-attr="aria-label:header.theme" data-testid="theme-toggle"><span class="icon i-sun" aria-hidden="true" data-testid="theme-toggle-icon-light"></span><span class="icon i-moon" aria-hidden="true" data-testid="theme-toggle-icon-dark"></span></button>' +
    '<a class="btn header-signin" href="' + R + 'pages/auth.html" data-testid="header-signin"><span class="icon i-user" aria-hidden="true" data-testid="header-signin-icon"></span><span class="header-signin-label" ' + t('header.signin') + ' data-testid="header-signin-label"></span></a>' +
    '<a class="btn btn-primary header-create" href="' + R + 'pages/solve.html#ask" data-testid="header-create"><span class="icon i-plus" aria-hidden="true" data-testid="header-create-icon"></span><span class="header-create-label" ' + t('header.create') + ' data-testid="header-create-label"></span></a>' +
    '</div></header>';
  var aside =
    '<div class="nav-scrim" data-action="nav-close" data-testid="nav-scrim"></div>' +
    '<aside class="side-nav" id="side-nav" data-testid="side-nav"><div class="nav-drawer-head"><span class="eyebrow" ' + t('nav.menu') + ' data-testid="nav-menu-label"></span>' +
    '<button type="button" class="btn btn-ghost btn-icon" data-action="nav-close" data-i18n-attr="aria-label:common.close" data-testid="nav-close"><span class="icon i-x" aria-hidden="true" data-testid="nav-close-icon"></span></button></div>' +
    '<nav data-i18n-attr="aria-label:nav.label" data-testid="primary-nav"><ul class="nav-list list-reset">' + item(DOMAINS[0]) +
    '</ul><p class="nav-group-label" ' + t('nav.group') + ' data-testid="nav-group-label"></p><ul class="nav-list list-reset">' + main.map(item).join('') + '</ul></nav></aside>';
  function bb(x) { return '<a href="' + R + x.href + '"' + (x.id === active ? ' aria-current="page"' : '') + ' data-testid="bottom-' + x.id + '"><span class="icon i-' + x.icon + '" aria-hidden="true" data-testid="bottom-' + x.id + '-icon"></span><span ' + t('nav.' + x.id) + ' data-testid="bottom-' + x.id + '-label"></span></a>'; }
  var bottom = '<nav class="bottom-nav" data-i18n-attr="aria-label:nav.bottom" data-testid="bottom-nav">' + DOMAINS.slice(0, 4).map(bb).join('') +
    '<button type="button" data-action="nav-toggle" aria-expanded="false" aria-controls="side-nav" data-testid="bottom-more"><span class="icon i-menu" aria-hidden="true" data-testid="bottom-more-icon"></span><span ' + t('nav.more') + ' data-testid="bottom-more-label"></span></button></nav>';
  var footer = '<footer class="app-footer" data-testid="app-footer"><span class="badge badge-sim" ' + t('sim.badge') + ' data-testid="footer-sim"></span><span ' + t('footer.note') + ' data-testid="footer-note"></span></footer>';
  var main_ = document.querySelector('main');
  if (main_) { main_.classList.add('page'); main_.id = main_.id || 'main'; main_.tabIndex = -1; }
  b.insertAdjacentHTML('afterbegin', header + aside);
  var sk = document.querySelector('.skip-link'); if (sk) b.insertBefore(sk, b.firstChild);
  b.insertAdjacentHTML('beforeend', bottom + footer);
})();
