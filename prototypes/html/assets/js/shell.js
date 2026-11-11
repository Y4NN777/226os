/* Injects the shared app shell (header, primary nav, bottom bar, footer) around <main class="page">.
   Page contract: <body class="app" data-active="solve|discover|build|identity|connect|explore" data-root="../"> (data-root="" for index.html).
   Run synchronously at the end of <body> (no defer). Single source of truth for navigation: edit DOMAINS here only. */
(function () {
  'use strict';
  var b = document.body, R = b.getAttribute('data-root') || '', active = b.getAttribute('data-active') || '';
  /* Only real destinations live in the navigation. st: 'preview' = designed journey with a page but no real behaviour (Connect): small neutral "Soon" label. */
  var DOMAINS = [
    { id: 'explore', icon: 'home', href: 'index.html' },
    { id: 'solve', icon: 'help', href: 'pages/solve.html' },
    { id: 'discover', icon: 'compass', href: 'pages/discover.html' },
    { id: 'build', icon: 'hammer', href: 'pages/build.html' },
    { id: 'identity', icon: 'users', href: 'pages/profile.html' },
    { id: 'connect', icon: 'network', href: 'pages/connect.html', st: 'preview' }
  ];
  function t(k) { return 'data-i18n="' + k + '"'; }
  function item(x) {
    var lab = '<span class="icon i-' + x.icon + '" aria-hidden="true" data-testid="nav-' + x.id + '-icon"></span><span class="nav-label" ' + t('nav.' + x.id) + ' data-testid="nav-' + x.id + '-label"></span>';
    var cur = x.id === active ? ' aria-current="page"' : '';
    return '<li><a class="nav-item' + (x.st === 'preview' ? ' is-preview' : '') + '" href="' + R + x.href + '"' + cur + ' data-testid="nav-' + x.id + '">' + lab +
      (x.st === 'preview' ? '<span class="nav-soon" ' + t('nav.later') + ' data-testid="nav-' + x.id + '-soon"></span>' : '') + '</a></li>';
  }
  function mi(role, attrs, icon, key, id, check) {
    return '<' + (attrs.href ? 'a' : 'button type="button"') + ' class="menu-item" role="' + role + '" tabindex="-1" ' + attrs.raw + ' data-testid="' + id + '">' + (icon ? '<span class="icon i-' + icon + '" aria-hidden="true" data-testid="' + id + '-icon"></span>' : '') +
      '<span ' + (key ? t(key) : '') + (attrs.lang ? ' lang="' + attrs.lang + '"' : '') + ' data-testid="' + id + '-label">' + (attrs.text || '') + '</span>' + (check ? '<span class="icon i-check menu-check" aria-hidden="true" data-testid="' + id + '-check"></span>' : '') + '</' + (attrs.href ? 'a' : 'button') + '>';
  }
  var menu =
    '<div class="menu" data-menu data-testid="account-menu">' +
    '<button type="button" class="btn btn-icon tip-end" data-menu-button aria-haspopup="menu" aria-expanded="false" aria-controls="account-menu-panel" data-i18n-attr="aria-label:header.account" data-testid="account-menu-button"><span class="icon i-user" aria-hidden="true" data-testid="account-menu-button-icon"></span></button>' +
    '<div class="menu-panel" id="account-menu-panel" role="menu" data-menu-panel hidden data-i18n-attr="aria-label:header.account" data-testid="account-menu-panel">' +
    mi('menuitem', { href: 1, raw: 'href="' + R + 'pages/auth.html"' }, 'log-in', 'header.signin', 'account-menu-signin') +
    mi('menuitem', { href: 1, raw: 'href="' + R + 'pages/profile.html"' }, 'user', 'header.profile', 'account-menu-profile') +
    '<div class="menu-sep" role="separator" data-testid="account-menu-sep-1"></div>' +
    '<div role="group" aria-labelledby="account-menu-lang-label" data-testid="account-menu-lang"><span class="menu-label" id="account-menu-lang-label" ' + t('header.lang') + ' data-testid="account-menu-lang-label"></span>' +
    mi('menuitemradio', { raw: 'aria-checked="false" data-action="lang-set" data-lang="fr" lang="fr"', text: 'Français' }, '', '', 'account-menu-lang-fr', 1) +
    mi('menuitemradio', { raw: 'aria-checked="false" data-action="lang-set" data-lang="en" lang="en"', text: 'English' }, '', '', 'account-menu-lang-en', 1) + '</div>' +
    '<div class="menu-sep" role="separator" data-testid="account-menu-sep-2"></div>' +
    '<div role="group" aria-labelledby="account-menu-theme-label" data-testid="account-menu-theme"><span class="menu-label" id="account-menu-theme-label" ' + t('header.themeLabel') + ' data-testid="account-menu-theme-label"></span>' +
    mi('menuitemradio', { raw: 'aria-checked="false" data-action="theme-set" data-theme-value="light"' }, 'sun', 'header.theme.light', 'account-menu-theme-light', 1) +
    mi('menuitemradio', { raw: 'aria-checked="false" data-action="theme-set" data-theme-value="dark"' }, 'moon', 'header.theme.dark', 'account-menu-theme-dark', 1) +
    mi('menuitemradio', { raw: 'aria-checked="false" data-action="theme-set" data-theme-value="system"' }, 'settings', 'header.theme.system', 'account-menu-theme-system', 1) + '</div>' +
    '</div></div>';
  var main = DOMAINS.filter(function (x) { return x.id !== 'explore'; });
  var header =
    '<header class="app-header" data-testid="app-header">' +
    '<a class="brand" href="' + R + 'index.html" data-testid="brand" aria-label="226os"><span class="brand-mark" data-testid="brand-mark">226</span><span class="brand-os" data-testid="brand-os">os</span></a>' +
    '<span class="header-spacer"></span>' +
    '<form class="header-search" role="search" action="' + R + 'index.html" data-testid="header-search-form">' +
    '<label class="visually-hidden" for="global-q" ' + t('header.search.label') + ' data-testid="header-search-label"></label>' +
    '<div class="search-field"><span class="icon i-search" aria-hidden="true" data-testid="header-search-icon"></span><input id="global-q" class="input" type="search" name="q" autocomplete="off" data-testid="header-search-input" data-i18n-attr="placeholder:header.search.placeholder"></div></form>' +
    '<div class="header-actions">' +
    '<button type="button" class="btn btn-icon theme-toggle tip-end" data-action="theme-toggle" aria-pressed="false" data-i18n-attr="aria-label:header.theme" data-testid="theme-toggle"><span class="icon i-sun" aria-hidden="true" data-testid="theme-toggle-icon-light"></span><span class="icon i-moon" aria-hidden="true" data-testid="theme-toggle-icon-dark"></span></button>' +
    menu +
    '<a class="btn btn-primary header-create has-tip tip-end" href="' + R + 'pages/solve.html#ask" data-i18n-attr="aria-label:header.create" data-testid="header-create"><span class="icon i-plus" aria-hidden="true" data-testid="header-create-icon"></span><span class="header-create-label" ' + t('header.create') + ' data-testid="header-create-label"></span></a>' +
    '</div></header>';
  var aside =
    '<div class="nav-scrim" data-action="nav-close" data-testid="nav-scrim"></div>' +
    '<aside class="side-nav" id="side-nav" data-testid="side-nav"><div class="nav-drawer-head"><span class="eyebrow" ' + t('nav.menu') + ' data-testid="nav-menu-label"></span>' +
    '<button type="button" class="btn btn-icon" data-action="nav-close" data-i18n-attr="aria-label:common.close" data-testid="nav-close"><span class="icon i-x" aria-hidden="true" data-testid="nav-close-icon"></span></button></div>' +
    '<nav data-i18n-attr="aria-label:nav.label" data-testid="primary-nav"><ul class="nav-list list-reset">' + item(DOMAINS[0]) + main.map(item).join('') + '</ul></nav></aside>';
  function bb(x) { return '<a href="' + R + x.href + '"' + (x.id === active ? ' aria-current="page"' : '') + ' data-testid="bottom-' + x.id + '"><span class="icon i-' + x.icon + '" aria-hidden="true" data-testid="bottom-' + x.id + '-icon"></span><span ' + t('nav.' + x.id) + ' data-testid="bottom-' + x.id + '-label"></span></a>'; }
  var bottom = '<nav class="bottom-nav" data-i18n-attr="aria-label:nav.bottom" data-testid="bottom-nav">' + DOMAINS.slice(0, 4).map(bb).join('') +
    '<button type="button" data-action="nav-toggle" aria-expanded="false" aria-controls="side-nav" data-testid="bottom-more"><span class="icon i-menu" aria-hidden="true" data-testid="bottom-more-icon"></span><span ' + t('nav.more') + ' data-testid="bottom-more-label"></span></button></nav>';
  var footer = '<footer class="app-footer" data-testid="app-footer"><span data-testid="footer-brand">226os</span><span data-proto ' + t('footer.note') + ' data-testid="footer-note"></span></footer>';
  var main_ = document.querySelector('main');
  if (main_) { main_.classList.add('page'); main_.id = main_.id || 'main'; main_.tabIndex = -1; }
  b.insertAdjacentHTML('afterbegin', header + aside);
  var sk = document.querySelector('.skip-link'); if (sk) b.insertBefore(sk, b.firstChild);
  b.insertAdjacentHTML('beforeend', bottom + footer);
})();
