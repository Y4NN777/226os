/* Injects the shared shell around <main class="page"> and owns the simulated session (App.Session). Runs synchronously at the end of <body> (no defer).
   Page contract: <body data-layout="app|public|account" data-active="solve|discover|build|identity|connect|explore" data-root="../">
     app (default; legacy `class="app"` still works): header + sidebar/rail + bottom bar on phones + footer.
     public: landing and public content: header (logo, Questions, Catalogue, Projets, sign in / create account), no sidebar, full footer.
     account: standalone card page: logo only, no navigation; the page puts its content in <div class="account-card"> inside <main>.
   Single source of truth: routes (window.AppConfig) and navigation (DOMAINS). See CONVENTIONS.md sections 10-12. */
(function () {
  'use strict';
  var d = document, b = d.body, R = b.getAttribute('data-root') || '', active = b.getAttribute('data-active') || '';
  var layout = b.getAttribute('data-layout'); if (layout !== 'public' && layout !== 'account') layout = 'app';
  var P = new URLSearchParams(location.search);

  /* ---------- Routes (relative to data-root). Integration: when the app home becomes pages/accueil.html change `home` only. ---------- */
  var CFG = window.AppConfig = {
    landing: 'index.html', home: 'index.html',
    signin: 'pages/connexion.html', signup: 'pages/inscription.html',
    profile: 'pages/profile.html', settings: 'pages/parametres.html', ask: 'pages/solve.html#ask',
    questions: 'pages/solve.html', catalogue: 'pages/discover.html', projects: 'pages/build.html',
    legal: { licence: 'https://example.org/226os/licence', conduct: 'https://example.org/226os/code-de-conduite', contribute: 'https://example.org/226os/contribuer', privacy: 'https://example.org/226os/confidentialite' }
  };

  /* ---------- Simulated session (localStorage 226os.session). Nothing here is real authentication. ----------
     ?session=in|out is a QA override (not persisted; `in` without stored data uses the demo user). */
  var SKEY = '226os.session', OV = P.get('session'); if (OV !== 'in' && OV !== 'out') OV = null;
  var DEMO = { name: 'Aminata Ouédraogo', username: 'aminata', intent: 'help', skills: ['go', 'postgresql'], country: 'BF', languages: ['fr'], github: '', bio: '', openToCollab: false };
  function clone(o) { return JSON.parse(JSON.stringify(o)); }
  function blankUser() { return { name: '', username: '', intent: '', skills: [], country: '', languages: [], github: '', bio: '', openToCollab: false }; }
  function blank() { return { signedIn: false, user: null, onboarding: { done: false, checklistDismissed: false, checklistCollapsed: false, confirmed: {} }, firstVisitChooserSeen: false, prompts: {} }; }
  function raw() {
    var s = blank();
    try { var o = JSON.parse(localStorage.getItem(SKEY)); if (o && typeof o === 'object') { Object.assign(s, o); s.onboarding = Object.assign(blank().onboarding, o.onboarding || {}); s.onboarding.confirmed = Object.assign({}, s.onboarding.confirmed); s.prompts = Object.assign({}, o.prompts); } } catch (e) {}
    return s;
  }
  function view() { var s = raw(); if (OV === 'in') { s.signedIn = true; if (!s.user) s.user = clone(DEMO); } else if (OV === 'out') s.signedIn = false; return s; }
  function save(v) {
    if (OV) v.signedIn = raw().signedIn;   /* a QA override never persists */
    try { localStorage.setItem(SKEY, JSON.stringify(v)); } catch (e) {}
    d.dispatchEvent(new CustomEvent('session:change', { detail: Session.state }));
  }
  function mutate(fn) { var v = view(); fn(v); save(v); return Session.state; }
  function safeNext(n) {   /* same-origin http(s)/file URL only; no javascript:, no protocol-relative tricks */
    if (!n || typeof n !== 'string' || /^\s*(javascript|data|vbscript):/i.test(n) || /^\s*\/\//.test(n) || n.indexOf('\\') > -1) return null;
    try { var u = new URL(n, location.href); if (!/^(https?|file):$/.test(u.protocol) || u.protocol !== location.protocol || u.host !== location.host) return null; return u.href; } catch (e) { return null; }
  }
  var ITEMS = [   /* checklist: key, editor kind, settings anchor */
    { key: 'skills', editor: 'dialog' }, { key: 'country', editor: 'inline' }, { key: 'github', editor: 'inline' },
    { key: 'bio', editor: 'inline' }, { key: 'openToCollab', editor: 'inline' }, { key: 'languages', editor: 'inline' }
  ];
  function isDone(key, v) {
    var u = v.user || blankUser(), c = v.onboarding.confirmed;
    switch (key) {
      case 'skills': return !!(u.skills && u.skills.length);
      case 'country': return !!u.country;
      case 'github': return !!u.github;
      case 'bio': return !!(u.bio && u.bio.trim());
      case 'openToCollab': return !!u.openToCollab || !!c.openToCollab;
      case 'languages': return !!c.languages;
    }
    return false;
  }
  var Session = {
    get state() { return view(); },
    get signedIn() { return view().signedIn; },
    get user() { var v = view(); return v.signedIn ? v.user : null; },
    get chooserSeen() { return !!view().firstVisitChooserSeen; },
    /* signIn(user): marks the session signed in; user fields are merged over any stored profile. Languages are inferred from the UI language, never asked. */
    signIn: function (user) {
      return mutate(function (v) {
        v.user = Object.assign(blankUser(), v.user || {}, user || {});
        if (!v.user.languages || !v.user.languages.length) v.user.languages = [d.documentElement.lang === 'en' ? 'en' : 'fr'];
        if (!v.user.name) v.user.name = v.user.username || '';
        v.signedIn = true;
      });
    },
    signOut: function () { return mutate(function (v) { v.signedIn = false; }); },   /* profile data is kept so the demo can sign in again */
    reset: function () { try { localStorage.removeItem(SKEY); } catch (e) {} d.dispatchEvent(new CustomEvent('session:change', { detail: Session.state })); },
    update: function (patch) { return mutate(function (v) { v.user = Object.assign(blankUser(), v.user || {}, patch || {}); }); },
    confirm: function (key) { return mutate(function (v) { v.onboarding.confirmed[key] = true; }); },
    setOnboarding: function (patch) { return mutate(function (v) { Object.assign(v.onboarding, patch || {}); }); },
    markChooserSeen: function () { return mutate(function (v) { v.firstVisitChooserSeen = true; }); },
    /* requireAuth(nextUrl): true when signed in; otherwise redirects to the sign-in page with ?next=<encoded absolute URL> (default: the current page) and returns false. */
    requireAuth: function (nextUrl) {
      if (Session.signedIn) return true;
      var n = safeNext(nextUrl || location.href) || location.href;
      location.href = R + CFG.signin + '?next=' + encodeURIComponent(n); return false;
    },
    /* returnTo(fallback): after sign-in, go back to ?next= (validated) or to fallback (default: the app home). Returns the target URL. */
    returnTo: function (fallback) {
      var t = safeNext(P.get('next')) || R + (fallback || CFG.home); location.href = t; return t;
    },
    /* checklist(): setup items for the OWN profile page only. Array of {key, done, labelKey, valueKey, editor:'inline'|'dialog', href}; also .done .total .pct and .progress {done,total,pct}. */
    checklist: function () {
      var v = view(), items = ITEMS.map(function (x) {
        return { key: x.key, done: isDone(x.key, v), labelKey: 'checklist.' + x.key + '.label', valueKey: 'checklist.' + x.key + '.value', editor: x.editor, href: R + CFG.settings + '#' + x.key };
      }), n = items.filter(function (i) { return i.done; }).length;
      items.done = n; items.total = items.length; items.pct = Math.round(100 * n / items.length); items.progress = { done: n, total: items.length, pct: items.pct };
      return items;
    },
    promptDismissed: function (key) { return !!view().prompts[key]; },
    dismissPrompt: function (key) { return mutate(function (v) { v.prompts[key] = true; }); },
    safeNext: safeNext
  };
  var App = window.App = window.App || {}; App.Session = Session; App.config = CFG;

  /* ---------- Markup helpers ---------- */
  var DOMAINS = [
    { id: 'explore', icon: 'home', href: CFG.home },
    { id: 'solve', icon: 'help', href: 'pages/solve.html' },
    { id: 'discover', icon: 'compass', href: 'pages/discover.html' },
    { id: 'build', icon: 'hammer', href: 'pages/build.html' },
    { id: 'identity', icon: 'users', href: 'pages/profile.html' },
    { id: 'connect', icon: 'network', href: 'pages/connect.html', st: 'preview' }
  ];
  function t(k) { return 'data-i18n="' + k + '"'; }
  function esc(s) { return String(s == null ? '' : s).replace(/[&<>"']/g, function (c) { return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]; }); }
  function ini(n) { return String(n || '').split(/[\s-]+/).filter(Boolean).slice(0, 2).map(function (w) { return w[0].toUpperCase(); }).join('') || '?'; }
  function tone(n) { var h = 0, s = String(n || ''); for (var i = 0; i < s.length; i++) h = (h * 31 + s.charCodeAt(i)) >>> 0; return h % 5; }
  function icon(name, id) { return '<span class="icon i-' + name + '" aria-hidden="true" data-testid="' + id + '"></span>'; }
  function item(x) {
    var lab = icon(x.icon, 'nav-' + x.id + '-icon') + '<span class="nav-label" ' + t('nav.' + x.id) + ' data-testid="nav-' + x.id + '-label"></span>';
    var cur = x.id === active ? ' aria-current="page"' : '';
    return '<li><a class="nav-item' + (x.st === 'preview' ? ' is-preview' : '') + '" href="' + R + x.href + '"' + cur + ' data-testid="nav-' + x.id + '">' + lab +
      (x.st === 'preview' ? '<span class="nav-soon" ' + t('nav.later') + ' data-testid="nav-' + x.id + '-soon"></span>' : '') + '</a></li>';
  }
  function mi(role, attrs, ic, key, id, check) {
    return '<' + (attrs.href ? 'a' : 'button type="button"') + ' class="menu-item" role="' + role + '" tabindex="-1" ' + attrs.raw + ' data-testid="' + id + '">' + (ic ? icon(ic, id + '-icon') : '') +
      '<span ' + (key ? t(key) : '') + (attrs.lang ? ' lang="' + attrs.lang + '"' : '') + ' data-testid="' + id + '-label">' + (attrs.text || '') + '</span>' + (check ? '<span class="icon i-check menu-check" aria-hidden="true" data-testid="' + id + '-check"></span>' : '') + '</' + (attrs.href ? 'a' : 'button') + '>';
  }
  function prefsGroups() {
    return '<div class="menu-sep" role="separator" data-testid="account-menu-sep-1"></div>' +
      '<div role="group" aria-labelledby="account-menu-lang-label" data-testid="account-menu-lang"><span class="menu-label" id="account-menu-lang-label" ' + t('header.lang') + ' data-testid="account-menu-lang-label"></span>' +
      mi('menuitemradio', { raw: 'aria-checked="false" data-action="lang-set" data-lang="fr" lang="fr"', text: 'Français' }, '', '', 'account-menu-lang-fr', 1) +
      mi('menuitemradio', { raw: 'aria-checked="false" data-action="lang-set" data-lang="en" lang="en"', text: 'English' }, '', '', 'account-menu-lang-en', 1) + '</div>' +
      '<div class="menu-sep" role="separator" data-testid="account-menu-sep-2"></div>' +
      '<div role="group" aria-labelledby="account-menu-theme-label" data-testid="account-menu-theme"><span class="menu-label" id="account-menu-theme-label" ' + t('header.themeLabel') + ' data-testid="account-menu-theme-label"></span>' +
      mi('menuitemradio', { raw: 'aria-checked="false" data-action="theme-set" data-theme-value="light"' }, 'sun', 'header.theme.light', 'account-menu-theme-light', 1) +
      mi('menuitemradio', { raw: 'aria-checked="false" data-action="theme-set" data-theme-value="dark"' }, 'moon', 'header.theme.dark', 'account-menu-theme-dark', 1) +
      mi('menuitemradio', { raw: 'aria-checked="false" data-action="theme-set" data-theme-value="system"' }, 'settings', 'header.theme.system', 'account-menu-theme-system', 1) + '</div>';
  }
  /* Account menu: signed in = avatar button + profile, settings, language, theme, sign out. Signed out = preferences (globe) button + language, theme. */
  function accountMenu(sv) {
    var u = sv.signedIn && sv.user, btn, head = '', body;
    if (u) {
      btn = '<span class="avatar header-avatar tone-' + tone(u.name) + '" aria-hidden="true" data-testid="account-menu-avatar">' + esc(ini(u.name)) + '</span>';
      head = '<div class="menu-user" role="none" data-testid="account-menu-user"><strong data-testid="account-menu-user-name">' + esc(u.name) + '</strong><span data-testid="account-menu-user-handle">@' + esc(u.username || '') + '</span></div>' +
        mi('menuitem', { href: 1, raw: 'href="' + R + CFG.profile + '"' }, 'user', 'header.profile', 'account-menu-profile') +
        mi('menuitem', { href: 1, raw: 'href="' + R + CFG.settings + '"' }, 'settings', 'header.settings', 'account-menu-settings');
      body = head + prefsGroups() + '<div class="menu-sep" role="separator" data-testid="account-menu-sep-3"></div>' + mi('menuitem', { raw: 'data-action="signout"' }, 'log-out', 'header.signout', 'account-menu-signout');
    } else {
      btn = icon('globe', 'account-menu-button-icon');
      body = prefsGroups().replace(/^<div class="menu-sep"[^>]*><\/div>/, '');
    }
    return '<div class="menu" data-menu data-testid="account-menu">' +
      '<button type="button" class="btn btn-icon tip-end" data-menu-button aria-haspopup="menu" aria-expanded="false" aria-controls="account-menu-panel" data-i18n-attr="aria-label:' + (u ? 'header.account' : 'header.prefs') + '" data-testid="account-menu-button">' + btn + '</button>' +
      '<div class="menu-panel" id="account-menu-panel" role="menu" data-menu-panel hidden data-i18n-attr="aria-label:' + (u ? 'header.account' : 'header.prefs') + '" data-testid="account-menu-panel">' + body + '</div></div>';
  }
  function authButtons(sm) {
    return '<a class="btn btn-ghost' + (sm ? ' btn-sm' : '') + '" href="' + R + CFG.signin + '" data-testid="header-signin"><span ' + t('header.signin') + ' data-testid="header-signin-label"></span></a>' +
      '<a class="btn btn-primary' + (sm ? ' btn-sm' : '') + '" href="' + R + CFG.signup + '" data-testid="header-signup"><span ' + t('header.signup') + ' data-testid="header-signup-label"></span></a>';
  }
  /* Brand: concept D logo (raster, AI-generated, see AI_USAGE.md). Light and dark variants swap with the theme. */
  function logo(id) { return '<img class="brand-logo brand-logo-light" src="' + R + 'assets/brand/226os-logo.png" alt="" width="132" height="48" data-testid="' + id + '-logo">' +
      '<img class="brand-logo brand-logo-dark" src="' + R + 'assets/brand/226os-logo-white.png" alt="" width="132" height="48" data-testid="' + id + '-logo-dark">'; }
  function brand(id) { id = id || 'brand'; return '<a class="brand" href="' + R + CFG.landing + '" data-testid="' + id + '" aria-label="226os">' + logo(id) + '</a>'; }
  (function icons() {   /* favicons for every page, without editing each <head> */
    [['icon', 'assets/brand/favicon-32.png', '32x32'], ['icon', 'assets/brand/icon-192.png', '192x192'], ['apple-touch-icon', 'assets/brand/icon-192.png', '192x192']].forEach(function (i) {
      var l = document.createElement('link'); l.rel = i[0]; l.type = 'image/png'; l.sizes = i[2]; l.href = R + i[1]; document.head.appendChild(l); });
  })();
  function createBtn() { return '<a class="btn btn-primary header-create has-tip tip-end" href="' + R + CFG.ask + '" data-i18n-attr="aria-label:header.create" data-testid="header-create">' + icon('plus', 'header-create-icon') + '<span class="header-create-label" ' + t('header.create') + ' data-testid="header-create-label"></span></a>'; }
  var sv = view(), html = '', tail = '';

  if (layout === 'app') {
    var main = DOMAINS.filter(function (x) { return x.id !== 'explore'; });
    html =
      '<header class="app-header" data-testid="app-header">' + brand() + '<span class="header-spacer"></span>' +
      '<form class="header-search" role="search" action="' + R + 'index.html" data-testid="header-search-form">' +
      '<label class="visually-hidden" for="global-q" ' + t('header.search.label') + ' data-testid="header-search-label"></label>' +
      '<div class="search-field">' + icon('search', 'header-search-icon') + '<input id="global-q" class="input" type="search" name="q" autocomplete="off" data-testid="header-search-input" data-i18n-attr="placeholder:header.search.placeholder"></div></form>' +
      '<div class="header-actions">' +
      '<button type="button" class="btn btn-icon theme-toggle tip-end" data-action="theme-toggle" aria-pressed="false" data-i18n-attr="aria-label:header.theme" data-testid="theme-toggle">' + icon('sun', 'theme-toggle-icon-light') + icon('moon', 'theme-toggle-icon-dark') + '</button>' +
      accountMenu(sv) + (sv.signedIn ? createBtn() : '<span class="header-auth">' + authButtons(true) + '</span>') +
      '</div></header>' +
      '<div class="nav-scrim" data-action="nav-close" data-testid="nav-scrim"></div>' +
      '<aside class="side-nav" id="side-nav" data-testid="side-nav"><div class="nav-drawer-head"><span class="eyebrow" ' + t('nav.menu') + ' data-testid="nav-menu-label"></span>' +
      '<button type="button" class="btn btn-icon" data-action="nav-close" data-i18n-attr="aria-label:common.close" data-testid="nav-close">' + icon('x', 'nav-close-icon') + '</button></div>' +
      '<nav data-i18n-attr="aria-label:nav.label" data-testid="primary-nav"><ul class="nav-list list-reset">' + item(DOMAINS[0]) + main.map(item).join('') + '</ul></nav></aside>';
    var bb = function (x) { return '<a href="' + R + x.href + '"' + (x.id === active ? ' aria-current="page"' : '') + ' data-testid="bottom-' + x.id + '">' + icon(x.icon, 'bottom-' + x.id + '-icon') + '<span ' + t('nav.' + x.id) + ' data-testid="bottom-' + x.id + '-label"></span></a>'; };
    tail = '<nav class="bottom-nav" data-i18n-attr="aria-label:nav.bottom" data-testid="bottom-nav">' + DOMAINS.slice(0, 4).map(bb).join('') +
      '<button type="button" data-action="nav-toggle" aria-expanded="false" aria-controls="side-nav" data-testid="bottom-more">' + icon('menu', 'bottom-more-icon') + '<span ' + t('nav.more') + ' data-testid="bottom-more-label"></span></button></nav>' +
      '<footer class="app-footer" data-testid="app-footer"><span data-testid="footer-brand">226os</span><span data-proto ' + t('footer.note') + ' data-testid="footer-note"></span></footer>';
  } else if (layout === 'public') {
    var pl = function (id, key, href, act) { return '<a href="' + R + href + '"' + (act === active ? ' aria-current="page"' : '') + ' ' + t(key) + ' data-testid="pub-nav-' + id + '"></a>'; };
    html =
      '<header class="pub-header" data-testid="pub-header"><div class="pub-header-inner">' + brand() +
      '<nav class="pub-nav" data-i18n-attr="aria-label:layout.pub.nav" data-testid="pub-nav">' + pl('questions', 'nav.questions', CFG.questions, 'solve') + pl('catalogue', 'nav.catalogue', CFG.catalogue, 'discover') + pl('projects', 'nav.projects', CFG.projects, 'build') + '</nav>' +
      '<div class="pub-actions">' + (sv.signedIn ? accountMenu(sv) + createBtn() : authButtons(true)) + '</div></div></header>';   /* signed out: language lives in the footer */
    var lk = function (id, key, href) { return '<li><a href="' + href + '" rel="noopener noreferrer" ' + t(key) + ' data-testid="footer-' + id + '"></a></li>'; };
    tail = '<footer class="pub-footer" data-testid="pub-footer"><div class="pub-footer-inner">' +
      '<div class="pub-footer-about"><span class="brand" aria-hidden="true" data-testid="footer-brand">' + logo('footer-brand') + '</span><p ' + t('footer.about') + ' data-testid="footer-about"></p></div>' +
      '<ul class="pub-footer-links" data-testid="footer-links">' + lk('licence', 'footer.licence', CFG.legal.licence) + lk('conduct', 'footer.conduct', CFG.legal.conduct) + lk('contribute', 'footer.contribute', CFG.legal.contribute) + lk('privacy', 'footer.privacy', CFG.legal.privacy) + '</ul>' +
      '<div class="pub-footer-lang" role="group" data-i18n-attr="aria-label:header.lang" data-testid="footer-lang"><span ' + t('footer.language') + ' data-testid="footer-lang-label"></span><div class="segmented">' +
      '<button type="button" data-action="lang-set" data-lang="fr" lang="fr" aria-pressed="false" data-testid="footer-lang-fr">FR</button><button type="button" data-action="lang-set" data-lang="en" lang="en" aria-pressed="false" data-testid="footer-lang-en">EN</button></div></div>' +
      '<p class="text-xs" data-proto ' + t('footer.note') + ' data-testid="footer-note"></p></div></footer>';
  } else {
    html = '<div class="account-top" data-testid="account-top">' + brand() + '</div>';
    tail = '<footer class="account-foot" data-testid="account-foot"><a href="' + CFG.legal.privacy + '" rel="noopener noreferrer" ' + t('footer.privacy') + ' data-testid="account-foot-privacy"></a><a href="' + CFG.legal.conduct + '" rel="noopener noreferrer" ' + t('footer.conduct') + ' data-testid="account-foot-conduct"></a></footer>';
  }

  b.classList.add('layout-' + layout); if (layout !== 'app') b.classList.remove('app'); else b.classList.add('app');
  var main_ = d.querySelector('main');
  if (main_) { main_.classList.add('page'); main_.id = main_.id || 'main'; main_.tabIndex = -1; }
  b.insertAdjacentHTML('afterbegin', html);
  var sk = d.querySelector('.skip-link'); if (sk) b.insertBefore(sk, b.firstChild);
  b.insertAdjacentHTML('beforeend', tail);

  d.addEventListener('click', function (e) {
    if (e.target.closest && e.target.closest('[data-action="signout"]')) { Session.signOut(); location.href = R + CFG.landing + (OV ? '?session=out' : ''); }
  });
})();
