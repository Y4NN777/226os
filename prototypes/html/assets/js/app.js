/* 226os prototype runtime (vanilla, no deps). Exposes window.I18N and window.App.
   Load with <script defer src=".../app.js"> BEFORE dictionary/lane scripts. See CONVENTIONS.md. */
(function () {
  'use strict';
  var d = document, root = d.documentElement, P = new URLSearchParams(location.search);
  var $ = function (s, r) { return (r || d).querySelector(s); };
  var $$ = function (s, r) { return [].slice.call((r || d).querySelectorAll(s)); };
  function store(k, v) { try { if (v === undefined) return localStorage.getItem('226os.' + k); localStorage.setItem('226os.' + k, v); } catch (e) { return null; } }
  function el(tag, cls, html) { var e = d.createElement(tag); if (cls) e.className = cls; if (html != null) e.innerHTML = html; return e; }
  function esc(s) { return String(s).replace(/[&<>"']/g, function (c) { return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]; }); }

  /* ---------- Avatar with initials (no images in the prototype). App.avatar(personOrName, {size:'sm'|'lg', testid}) ---------- */
  function initials(n) { return String(n || '').split(/[\s-]+/).filter(Boolean).slice(0, 2).map(function (w) { return w[0].toUpperCase(); }).join(''); }
  function tone(name) { var h = 0, s = String(name || ''); for (var i = 0; i < s.length; i++) h = (h * 31 + s.charCodeAt(i)) >>> 0; return h % 5; }   /* deterministic 0..4 -> .tone-N (token palette only) */
  function avatar(p, o) {
    o = o || {}; var name = typeof p === 'string' ? p : (p && p.name), ini = (p && p.initials) || initials(name);
    return '<span class="avatar tone-' + tone(name) + (o.size ? ' avatar-' + o.size : '') + '" aria-hidden="true"' + (o.testid ? ' data-testid="' + esc(o.testid) + '"' : '') + '>' + esc(ini) + '</span>';
  }

  /* ---------- Thumbnail: App.thumb(name, {size:'sm|md|lg|wide', shape:'circle', icon:'code', img:'path', alt, tone:0-4, testid}) ----------
     Generated monogram + tone derived from the name (token palette only, AA text). `img` is an optional local image slot (no remote images). */
  function thumb(name, o) {
    o = o || {}; var t = o.tone != null ? o.tone : tone(name), tid = o.testid ? ' data-testid="' + esc(o.testid) + '"' : '', inner;
    if (o.img) inner = '<img src="' + esc(o.img) + '" alt="' + esc(o.alt || '') + '" loading="lazy"' + (o.testid ? ' data-testid="' + esc(o.testid) + '-img"' : '') + '>';
    else if (o.icon) inner = '<span class="icon i-' + o.icon + '" aria-hidden="true"' + (o.testid ? ' data-testid="' + esc(o.testid) + '-icon"' : '') + '></span>';
    else inner = '<span aria-hidden="true"' + (o.testid ? ' data-testid="' + esc(o.testid) + '-text"' : '') + '>' + esc(initials(name)) + '</span>';
    return '<span class="thumb tone-' + t + (o.size && o.size !== 'md' ? ' thumb-' + o.size : '') + (o.shape === 'circle' ? ' thumb-circle' : '') + '"' + (o.img && o.alt ? ' role="img" aria-label="' + esc(o.alt) + '"' : '') + tid + '>' + inner + '</span>';
  }

  /* ---------- Rating (read-only): App.rating(value, {count, max, size:'lg', testid}) -> role=img with a full text alternative.
     Half stars are rounded to the nearest 0.5; the number is always shown as text too (never colour/shape only). ---------- */
  function num1(v) { var x = (Math.round(v * 10) / 10).toFixed(1).replace(/\.0$/, ''); return I18N.lang === 'fr' ? x.replace('.', ',') : x; }
  function rating(v, o) {
    o = o || {}; var max = o.max || 5, r = Math.round(v * 2) / 2, tid = o.testid || 'rating', st = '';
    for (var i = 1; i <= max; i++) st += '<span class="icon i-' + (i <= r ? 'star-fill' : i - .5 === r ? 'star-half' : 'star') + (i <= r || i - .5 === r ? '' : ' is-empty') + '" aria-hidden="true" data-testid="' + esc(tid) + '-star-' + i + '"></span>';
    var label = I18N.t('common.rating.of', { v: num1(v), max: max }) + (o.count != null ? ', ' + I18N.t('common.reviews', { n: o.count }) : '');
    return '<span class="rating' + (o.size ? ' rating-' + o.size : '') + '" role="img" aria-label="' + esc(label) + '" data-testid="' + esc(tid) + '"><span class="rating-stars" data-testid="' + esc(tid) + '-stars">' + st + '</span><span class="rating-value" aria-hidden="true" data-testid="' + esc(tid) + '-value">' + num1(v) + '</span>' +
      (o.count != null ? '<span class="rating-count" aria-hidden="true" data-testid="' + esc(tid) + '-count">(' + o.count + ')</span>' : '') + '</span>';
  }
  /* Rating input (radio group): App.ratingInput({name, legend | legendKey, value, max, required, testid}) -> fieldset.rating-input.
     Native radios: arrow keys, label and form semantics come for free; the checked value is submitted as `name`. */
  function ratingInput(o) {
    o = o || {}; var max = o.max || 5, tid = o.testid || 'rating-input', name = o.name || 'rating', h = '';
    for (var i = 1; i <= max; i++) h += '<label data-testid="' + esc(tid) + '-option-' + i + '"><input type="radio" name="' + esc(name) + '" value="' + i + '"' + (+o.value === i ? ' checked' : '') + (o.required && i === 1 ? ' required' : '') + ' data-testid="' + esc(tid) + '-input-' + i + '"><span class="visually-hidden" data-testid="' + esc(tid) + '-label-' + i + '">' + i + ' / ' + max + '</span>' +
      '<span class="icon i-star" aria-hidden="true" data-testid="' + esc(tid) + '-star-' + i + '"></span><span class="icon i-star-fill" aria-hidden="true" data-testid="' + esc(tid) + '-star-fill-' + i + '"></span></label>';
    return '<fieldset class="rating-input" data-testid="' + esc(tid) + '"><legend' + (o.legendKey ? ' data-i18n="' + esc(o.legendKey) + '"' : '') + ' data-testid="' + esc(tid) + '-legend">' + (o.legendKey ? esc(I18N.t(o.legendKey)) : esc(o.legend || '')) + '</legend><div class="rating-input-stars" data-testid="' + esc(tid) + '-stars">' + h + '</div></fieldset>';
  }
  /* Review: App.review({name, person, country:'BF', rating, date:'2026-03-01', text, lang:'fr', verified:true}, {testid:'review-1'}) -> article.review.
     User text keeps its own language (lang + data-user-content) and carries an original-language badge; `verified` shows the "integration verified" marker. */
  var LANGN = { fr: { fr: 'Français', en: 'French' }, en: { fr: 'Anglais', en: 'English' } };
  function review(r, o) {
    o = o || {}; var tid = o.testid || 'review', M = window.MOCK || {}, who = r.person || { name: r.name }, ln = LANGN[r.lang] ? LANGN[r.lang][I18N.lang] : r.lang;
    var cn = r.country && M.countryName ? M.countryName(r.country) : '', date = r.date ? (M.fmtDate ? M.fmtDate(r.date) : new Date(r.date).toLocaleDateString(I18N.lang)) : '';
    return '<article class="review" data-testid="' + esc(tid) + '"><div class="review-head" data-testid="' + esc(tid) + '-head">' + avatar(who, { testid: tid + '-avatar' }) +
      '<div class="review-who" data-testid="' + esc(tid) + '-who"><p class="review-name" data-testid="' + esc(tid) + '-name">' + esc(who.name || r.name) + '</p><div class="review-sub">' +
      (r.country ? '<span class="country" data-testid="' + esc(tid) + '-country"><span class="country-code" data-testid="' + esc(tid) + '-country-code">' + esc(r.country) + '</span>' + (cn ? '<span class="visually-hidden" data-testid="' + esc(tid) + '-country-name">' + esc(cn) + '</span>' : '') + '</span>' : '') +
      (r.date ? '<time datetime="' + esc(r.date) + '" data-testid="' + esc(tid) + '-date">' + esc(date) + '</time>' : '') + '</div></div>' + rating(r.rating, { testid: tid + '-rating' }) + '</div>' +
      '<p class="review-text" lang="' + esc(r.lang || I18N.lang) + '" data-user-content data-testid="' + esc(tid) + '-text">' + esc(r.text) + '</p>' +
      '<div class="review-foot" data-testid="' + esc(tid) + '-foot"><span class="badge" lang="' + I18N.lang + '" title="' + esc(I18N.t('common.originalLang') + (I18N.lang === 'fr' ? '\u00a0: ' : ': ') + ln) + '" data-testid="' + esc(tid) + '-lang"><span aria-hidden="true" data-testid="' + esc(tid) + '-lang-code">' + esc(String(r.lang || '').toUpperCase()) + '</span><span class="visually-hidden" data-testid="' + esc(tid) + '-lang-name">' + esc(I18N.t('common.originalLang') + (I18N.lang === 'fr' ? '\u00a0: ' : ': ') + ln) + '</span></span>' +
      (r.verified ? '<span class="badge badge-verified" data-testid="' + esc(tid) + '-verified"><span class="icon i-shield-check" aria-hidden="true" data-testid="' + esc(tid) + '-verified-icon"></span>' + esc(I18N.t('common.integrationVerified')) + '</span>' : '') + '</div></article>';
  }

  /* ---------- i18n: UI strings only, never user content ---------- */
  /* Language: ?lang= (QA override, not persisted) > persisted user choice (account menu) > navigator.languages (fr* -> fr, otherwise en). No notice is shown. */
  function detectLang() { var l = (navigator.languages && navigator.languages[0]) || navigator.language || 'fr'; return /^fr/i.test(l) ? 'fr' : 'en'; }
  var dicts = { fr: {}, en: {} }, lang = P.get('lang');
  if (lang !== 'fr' && lang !== 'en') lang = store('lang');
  if (lang !== 'fr' && lang !== 'en') lang = detectLang();
  root.lang = lang;
  var warned = {};
  var FRP = /[ ]([:;?!»])/;
  var I18N = {
    get lang() { return lang; },
    register: function (dict) { Object.keys(dict).forEach(function (l) { dicts[l] = Object.assign(dicts[l] || {}, dict[l]); }); },
    t: function (k, p) {
      var s = dicts[lang][k]; if (s == null) s = dicts.fr[k];
      if (s == null) { if (!warned[k]) { warned[k] = 1; console.warn('[i18n] missing key', k); } return k; }
      if (lang === 'fr' && FRP.test(s)) s = s.replace(/ ([:;?!»])/g, '\u00a0$1').replace(/«\s/g, '«\u00a0'); /* French typography safety net: no-break space before : ; ? ! » and after « */
      return p ? s.replace(/\{(\w+)\}/g, function (m, n) { return p[n] != null ? p[n] : m; }) : s;
    },
    apply: function (r) {
      $$('[data-i18n]', r).forEach(function (n) { n.textContent = I18N.t(n.getAttribute('data-i18n')); });
      $$('[data-i18n-attr]', r).forEach(function (n) {
        n.getAttribute('data-i18n-attr').split(';').forEach(function (pair) { var a = pair.split(':'); if (a[1]) n.setAttribute(a[0].trim(), I18N.t(a[1].trim())); });
      });
      $$('[data-lang]').forEach(function (b) { b.setAttribute(b.getAttribute('role') === 'menuitemradio' ? 'aria-checked' : 'aria-pressed', b.getAttribute('data-lang') === lang); });
      syncPw();
    },
    set: function (l) {
      if (l !== 'fr' && l !== 'en') return; lang = l; root.lang = l; store('lang', l);
      I18N.apply(); d.dispatchEvent(new CustomEvent('i18n:change', { detail: { lang: l } }));
    }
  };
  window.I18N = I18N;

  /* ---------- Theme ---------- */
  function isDark() { var t = root.getAttribute('data-theme'); return t ? t === 'dark' : matchMedia('(prefers-color-scheme: dark)').matches; }
  function syncTheme() {
    $$('[data-action="theme-toggle"]').forEach(function (b) { b.setAttribute('aria-pressed', isDark()); });
    var cur = root.getAttribute('data-theme') || 'system'; if (cur !== 'dark' && cur !== 'light') cur = 'system';
    $$('[data-theme-value]').forEach(function (b) { b.setAttribute('aria-checked', b.getAttribute('data-theme-value') === cur); });
  }
  /* Theme.set('light'|'dark'|'system'); 'system' removes the explicit theme so the OS preference applies */
  var Theme = { set: function (t) { if (t === 'system') root.removeAttribute('data-theme'); else root.setAttribute('data-theme', t); store('theme', t); syncTheme(); }, toggle: function () { Theme.set(isDark() ? 'light' : 'dark'); } };
  function syncPw() { $$('[data-pw-toggle]').forEach(function (b) { b.setAttribute('aria-label', I18N.t(b.getAttribute('aria-pressed') === 'true' ? 'common.hidePassword' : 'common.showPassword')); }); }

  /* ---------- Toast ---------- */
  var Toast = {
    show: function (o) {
      var reg = $('.toast-region') || d.body.appendChild(el('div', 'toast-region'));
      reg.setAttribute('aria-live', 'polite');
      var k = o.kind || 'info', ic = { success: 'check-circle', error: 'alert-circle', info: 'info' }[k];
      var t = el('div', 'toast toast-' + k); t.setAttribute('role', k === 'error' ? 'alert' : 'status');
      var tid = 'toast-' + k, nth = 1; while ($('[data-testid="' + tid + '"]')) tid = 'toast-' + k + '-' + (++nth);   /* first toast of a kind: toast-info; then toast-info-2... */
      t.setAttribute('data-testid', tid);
      t.innerHTML = '<span class="icon icon-lead i-' + ic + '" aria-hidden="true" data-testid="' + tid + '-icon"></span><div data-testid="' + tid + '-body">' + (o.title ? '<strong data-testid="' + tid + '-title">' + esc(o.title) + '</strong>' : '') + '<span data-testid="' + tid + '-message">' + esc(o.msg || '') + '</span>' +
        '</div><button class="btn btn-ghost btn-sm btn-icon" type="button" data-testid="' + tid + '-close" aria-label="' + esc(I18N.t('common.close')) + '"><span class="icon i-x" aria-hidden="true" data-testid="' + tid + '-close-icon"></span></button>';
      reg.appendChild(t);
      var kill = function () { t.remove(); };
      t.querySelector('button').addEventListener('click', kill);
      if (o.ttl !== 0) setTimeout(kill, o.ttl || 5000);
    }
  };

  /* ---------- Dialog (native <dialog>: modal focus containment, Esc; we restore focus + handle backdrop) ---------- */
  var opener = null;
  var Dialog = {
    open: function (id, from) { var dl = d.getElementById(id); if (!dl) return; opener = from || d.activeElement; if (dl.showModal) dl.showModal(); else dl.setAttribute('open', ''); },
    close: function (dl) { if (dl.close) dl.close(); else dl.removeAttribute('open'); }
  };
  d.addEventListener('close', function (e) { if (e.target.classList && e.target.classList.contains('dialog') && opener) { opener.focus(); opener = null; } }, true);

  /* ---------- Tabs (ARIA tabs pattern, automatic activation) ----------
     App.Tabs.init(root)   wire every [data-tabs] inside root (or root itself); idempotent, safe for dynamically rendered tabs.
     Keyboard (arrows/Home/End) moves focus AND calls .click() on the target tab, so click, keyboard and Tabs.select share one
     activation path: listen to 'click' on a tab, or to the bubbling 'tabs:change' event ({detail:{index, tab, panel}}) on the [data-tabs] box. */
  function initTabs(box) {
    if (box._tabs) return box._tabs;
    var tabs = $$('[role="tab"]', box), panels = tabs.map(function (t) { return d.getElementById(t.getAttribute('aria-controls')); });
    function select(i, silent) {
      tabs.forEach(function (t, j) { var on = i === j; t.setAttribute('aria-selected', on); t.tabIndex = on ? 0 : -1; if (panels[j]) panels[j].hidden = !on; });
      if (!silent) box.dispatchEvent(new CustomEvent('tabs:change', { bubbles: true, detail: { index: i, tab: tabs[i], panel: panels[i] } }));
    }
    tabs.forEach(function (t, i) {
      t.addEventListener('click', function () { select(i); });
      t.addEventListener('keydown', function (e) {
        var n = { ArrowRight: i + 1, ArrowLeft: i - 1, Home: 0, End: tabs.length - 1 }[e.key];
        if (n == null) return; e.preventDefault(); var k = (n + tabs.length) % tabs.length; tabs[k].focus(); tabs[k].click();
      });
    });
    var cur = tabs.findIndex(function (t) { return t.getAttribute('aria-selected') === 'true'; }); select(cur < 0 ? 0 : cur, true);
    return (box._tabs = { select: function (i) { tabs[i].focus(); tabs[i].click(); } });
  }
  var Tabs = { init: function (r) { r = r || d; var boxes = $$('[data-tabs]', r); if (r.matches && r.matches('[data-tabs]')) boxes.unshift(r); boxes.forEach(initTabs); } };

  /* ---------- DEBUG-ONLY fallback data-testid assigner (enabled with ?autotestid=1; off by default, ids are hand-written) ----------
     Names every text / icon / interactive element that still lacks a data-testid, after its nearest testid ancestor
     (`<ancestor>-<icon-name|tag>-N`, unique per page). Runs at boot and on DOM changes (debounced). Explicit data-testid always wins. */
  var TestIds = {
    assign: function (rootEl) {
      var skip = { SCRIPT: 1, STYLE: 1, BR: 1, HEAD: 1, META: 1, LINK: 1, TITLE: 1 }, inter = /^(A|BUTTON|INPUT|SELECT|TEXTAREA|LABEL|IMG|PRE|TIME|OPTION|P|LI|H1|H2|H3|H4|DT|DD|CODE|STRONG|EM|LEGEND)$/;
      $$('*', rootEl || d.body).forEach(function (e) {
        if (e.hasAttribute('data-testid') || skip[e.tagName] || e.closest('svg')) return;
        var own = [].some.call(e.childNodes, function (n) { return n.nodeType === 3 && n.textContent.trim(); }), icon = e.classList.contains('icon');
        if (!(own || icon || inter.test(e.tagName))) return;
        var anc = e.parentElement && e.parentElement.closest('[data-testid]'), a = anc ? anc.getAttribute('data-testid') : 'page';
        var m = icon && /\bi-([a-z-]+)/.exec(e.className), kind = icon ? 'icon-' + (m ? m[1] : 'x') : /^tok-/.test(e.className) ? 'tok' : e.tagName.toLowerCase(), n = 1;
        while (d.querySelector('[data-testid="' + a + '-' + kind + '-' + n + '"]')) n++;
        e.setAttribute('data-testid', a + '-' + kind + '-' + n);
      });
    },
    auto: function () {
      if (TestIds._on) return; TestIds._on = true; var q = false;
      var run = function () { q = false; TestIds.assign(d.body); };
      new MutationObserver(function () { if (!q) { q = true; setTimeout(run, 0); } }).observe(d.body, { childList: true, subtree: true });
      run();
    }
  };

  /* ---------- Generic filter / sort / search over an array ---------- */
  var Filter = {
    norm: function (s) { return String(s == null ? '' : s).normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase(); },
    /* opts: {q, fields:[keys searched], where:{key: value|[values]} (array-valued item keys match any), sort:{key, dir}} */
    run: function (items, o) {
      o = o || {}; var words = Filter.norm(o.q).split(/\s+/).filter(Boolean), where = o.where || {};
      var out = items.filter(function (it) {
        for (var k in where) {
          var v = where[k]; if (v == null || v === '' || (Array.isArray(v) && !v.length)) continue;
          var xs = [].concat(it[k]), vs = [].concat(v);
          if (!vs.some(function (a) { return xs.indexOf(a) > -1; })) return false;
        }
        if (!words.length) return true;
        var hay = Filter.norm((o.fields || Object.keys(it)).map(function (f) { return [].concat(it[f]).join(' '); }).join(' '));
        return words.every(function (w) { return hay.indexOf(w) > -1; });
      });
      if (o.sort) { var k = o.sort.key, dir = o.sort.dir === 'desc' ? -1 : 1; out.sort(function (a, b) { return (a[k] > b[k] ? 1 : a[k] < b[k] ? -1 : 0) * dir; }); }
      return out;
    }
  };

  /* ---------- Simulated async state helper ----------
     Sim.load(container, {load: ()=>array, render:(items, container)=>void, delay, skeleton, emptyKey, empty, error})
       Returns {run, refresh}. Forced state via ?state=loading|empty|error|notfound. Shows one visible "Simulated" marker per page.
       o.empty / o.error customise the state: {icon, key:'prefix' (-> prefix.title / prefix.body) | titleKey, bodyKey, heading:'h1|h2|h3',
         action:{labelKey|label, href | onClick, icon, primary, testid}, retry:false (error only: drop the default Retry button)}.
     Sim.show(container, kind, cfg)  render a state (loading|empty|error|notfound) into ANY box, e.g. on a detail page, without rebuilding
       tabs or the surrounding layout; re-renders on language change. cfg has the same shape as o.empty / o.error. Retry reloads without ?state.
     Sim.forced()  -> 'loading'|'empty'|'error'|'notfound'|null from ?state=. */
  var sims = [];
  var KINDS = { loading: 1, empty: 1, error: 1, notfound: 1 };
  function ensureMarker(box) {
    if ($('[data-sim-marker]') || !box.parentNode) return;
    var m = el('p', 'sim-marker text-sm text-muted', '<span class="badge badge-sim" data-i18n="sim.badge" data-testid="sim-marker-badge"></span><span data-i18n="sim.note" data-testid="sim-marker-note"></span>');
    m.setAttribute('data-sim-marker', ''); m.setAttribute('data-proto', ''); m.setAttribute('data-testid', 'sim-marker'); box.parentNode.insertBefore(m, box); I18N.apply(m);
  }
  /* State ids: `<base>-<loading|empty|error|notfound>` + `-icon|-title|-body|-actions|-action|-retry`; base is `state` unless the box carries data-state-id
     (needed when a page has several async boxes) or the config passes stateId. */
  function sbase(box, o) { return (o && o.stateId) || (box && box.getAttribute('data-state-id')) || 'state'; }
  function actionHtml(a, acts, base) {
    if (!a) return '';
    var tidv = a.testid || (base + '-action'), tid = ' data-testid="' + esc(tidv) + '"';
    var ic = a.icon ? '<span class="icon i-' + a.icon + '" aria-hidden="true" data-testid="' + esc(tidv) + '-icon"></span>' : '', lab = esc(a.label != null ? a.label : I18N.t(a.labelKey)),
      cls = 'btn btn-sm' + (a.primary ? ' btn-primary' : '');
    if (a.href) return '<a class="' + cls + '" href="' + esc(a.href) + '"' + tid + '>' + ic + lab + '</a>';
    acts.push(a.onClick || function () {}); return '<button type="button" class="' + cls + '" data-sim-act="' + (acts.length - 1) + '"' + tid + '>' + ic + lab + '</button>';
  }
  function stateHtml(kind, o, acts, base) {
    var b = base + '-' + kind;
    var def = kind === 'error' ? { icon: 'alert-circle', key: 'state.error', cls: ' state-error' } : { icon: 'inbox', key: o.emptyKey || 'state.empty', cls: '' };
    var c = Object.assign({}, def, (kind === 'error' ? o.error : o.empty) || {}), hd = c.heading || 'h3';
    var act = c.action ? actionHtml(c.action, acts, b) : '', retry = '';
    if (kind === 'error' && c.retry !== false) retry = '<button type="button" class="btn btn-sm" data-sim-retry data-testid="' + base + '-retry"><span class="icon i-refresh" aria-hidden="true" data-testid="' + base + '-retry-icon"></span>' + esc(I18N.t('state.retry')) + '</button>';
    return '<div class="state' + (c.cls || '') + '" data-testid="' + b + '"><span class="icon i-' + c.icon + '" aria-hidden="true" data-testid="' + b + '-icon"></span><' + hd + (hd === 'h3' ? '' : ' class="h3"') + ' data-testid="' + b + '-title">' + esc(I18N.t(c.titleKey || c.key + '.title')) + '</' + hd + '><p data-testid="' + b + '-body">' +
      esc(I18N.t(c.bodyKey || c.key + '.body')) + '</p>' + (act || retry ? '<div class="state-actions" data-testid="' + b + '-actions">' + act + retry + '</div>' : '') + '</div>';
  }
  function skeletonHtml(n, base) {
    var h = ''; n = n || 3; while (n--) h += '<div class="skeleton-row"><span class="skeleton"></span><span class="skeleton"></span></div>';
    return '<div class="list" data-testid="' + base + '-loading" role="status" aria-label="' + esc(I18N.t('state.loading')) + '">' + h + '</div>';
  }
  function wire(box, retryFn) {
    box._simRetry = retryFn; if (box._simWired) return; box._simWired = true;
    box.addEventListener('click', function (e) {
      if (e.target.closest('[data-sim-retry]')) { box._simRetry && box._simRetry(); return; }
      var b = e.target.closest('[data-sim-act]'), f = b && box._simActs && box._simActs[+b.getAttribute('data-sim-act')]; if (f) f(e);
    });
  }
  var Sim = {
    forced: function () { var f = P.get('state'); return KINDS[f] ? f : null; },
    show: function (box, kind, o) {
      o = o || {}; ensureMarker(box); wire(box, function () { var u = new URL(location.href); u.searchParams.delete('state'); location.href = u.href; });
      function draw() {
        var acts = []; box._simActs = acts; box.setAttribute('aria-live', 'polite'); box.setAttribute('aria-busy', kind === 'loading'); box.setAttribute('data-state', kind);
        box.innerHTML = kind === 'loading' ? skeletonHtml(o.skeleton, sbase(box, o)) : stateHtml(kind, { error: o, empty: o }, acts, sbase(box, o)); I18N.apply(box);
      }
      box._simDraw = draw; draw();
      if (!box._simLang) { box._simLang = 1; d.addEventListener('i18n:change', function () { box._simDraw(); }); }
    },
    load: function (box, o) {
      var forced = P.get('state'), items = null, st = 'loading';
      ensureMarker(box); box.setAttribute('aria-live', 'polite');
      function draw() {
        var acts = []; box._simActs = acts; box.setAttribute('aria-busy', st === 'loading'); box.setAttribute('data-state', st);
        if (st === 'loading') box.innerHTML = skeletonHtml(o.skeleton, sbase(box, o));
        else if (st === 'error' || st === 'empty') box.innerHTML = stateHtml(st, o, acts, sbase(box, o));
        else o.render(items, box);
        I18N.apply(box);
      }
      function run() {
        st = 'loading'; draw();
        if (forced === 'loading') return;
        setTimeout(function () {
          try { items = forced === 'error' ? (function () { throw new Error('sim'); })() : (forced === 'empty' || forced === 'notfound') ? [] : o.load(); st = items.length ? 'success' : 'empty'; } catch (e) { st = 'error'; }
          draw();
        }, o.delay == null ? 600 : o.delay);
      }
      wire(box, function () { forced = null; run(); });
      sims.push(draw); run();
      return { run: run, refresh: function () { try { items = o.load(); st = items.length ? 'success' : 'empty'; } catch (e) { st = 'error'; } draw(); } };
    }
  };
  d.addEventListener('i18n:change', function () { sims.forEach(function (f) { f(); }); });

  /* ---------- Responsive nav drawer (mobile) ---------- */
  var mq = matchMedia('(max-width: 767px)');
  function navOpen(open) {
    d.body.classList.toggle('nav-open', open);
    $$('[data-action="nav-toggle"]').forEach(function (b) { b.setAttribute('aria-expanded', open); });
    syncNav(); var nav = $('#side-nav');
    if (open && nav) { var f = $('a, button', nav); f && f.focus(); } else if (!open) { var t = $('.bottom-nav [data-action="nav-toggle"]'); if (t && mq.matches && d.activeElement && d.activeElement.closest && d.activeElement.closest('#side-nav')) t.focus(); }
  }
  function syncNav() { var nav = $('#side-nav'); if (!nav) return; if (!mq.matches) d.body.classList.remove('nav-open'); nav.inert = mq.matches && !d.body.classList.contains('nav-open'); }
  mq.addEventListener('change', syncNav);

  /* ---------- Menu button (WAI-ARIA menu-button pattern). App.Menu.init(root) wires every [data-menu] ----------
     Markup: .menu[data-menu] > button[data-menu-button][aria-haspopup=menu][aria-expanded][aria-controls] + .menu-panel[role=menu][data-menu-panel][hidden]
     Items: [role=menuitem|menuitemradio|menuitemcheckbox] (a or button). Keys: Enter/Space/ArrowDown open on first, ArrowUp on last, arrows/Home/End/type-ahead move, Esc closes + refocuses, Tab closes. */
  function initMenu(box) {
    if (box._menu) return box._menu;
    var btn = $('[data-menu-button]', box), panel = $('[data-menu-panel]', box); if (!btn || !panel) return null;
    function items() { return $$('[role^="menuitem"]', panel).filter(function (i) { return !i.disabled && !i.hidden && i.offsetParent !== null; }); }
    function open(at) {
      if (!panel.hidden) return; panel.hidden = false; btn.setAttribute('aria-expanded', 'true');
      var it = items(); it.forEach(function (i) { i.tabIndex = -1; }); if (it.length) (at === 'last' ? it[it.length - 1] : it[0]).focus();
    }
    function close(ret) { if (panel.hidden) return; panel.hidden = true; btn.setAttribute('aria-expanded', 'false'); if (ret) btn.focus(); }
    btn.addEventListener('click', function () { panel.hidden ? open('first') : close(true); });
    btn.addEventListener('keydown', function (e) { if (e.key === 'ArrowDown') { e.preventDefault(); open('first'); } else if (e.key === 'ArrowUp') { e.preventDefault(); open('last'); } });
    panel.addEventListener('keydown', function (e) {
      var it = items(), i = it.indexOf(d.activeElement), n = null;
      if (e.key === 'ArrowDown') n = (i + 1) % it.length; else if (e.key === 'ArrowUp') n = (i - 1 + it.length) % it.length;
      else if (e.key === 'Home') n = 0; else if (e.key === 'End') n = it.length - 1;
      else if (e.key === 'Escape') { e.preventDefault(); e.stopPropagation(); close(true); return; }
      else if (e.key === 'Tab') { close(false); return; }
      else if (e.key.length === 1 && /\S/.test(e.key)) { var c = e.key.toLowerCase(); for (var k = 1; k <= it.length; k++) { var j = (i + k) % it.length; if ((it[j].textContent || '').trim().toLowerCase().indexOf(c) === 0) { n = j; break; } } }
      if (n != null) { e.preventDefault(); it[n].focus(); }
    });
    panel.addEventListener('click', function (e) { if (e.target.closest('[role^="menuitem"]')) close(!e.target.closest('a[href]')); });
    box.addEventListener('focusout', function (e) { if (!panel.hidden && e.relatedTarget && !box.contains(e.relatedTarget)) close(false); });
    d.addEventListener('click', function (e) { if (!box.contains(e.target)) close(false); });
    return (box._menu = { open: open, close: close });
  }
  var Menu = { init: function (r) { $$('[data-menu]', r || d).forEach(initMenu); } };

  /* ---------- Prototype controls: floating toggle + panel (role "view as", screen state, annotations) ----------
     App.Proto.register({roles:[{id, label:{fr,en} | labelKey}], default, legend?:{fr,en}})   one call per page; `?role=` overrides, last choice is remembered per page
     App.Proto.role / .setRole(id) / .is(id)    document 'proto:role' event: detail {role, previous, roles}  (not fired for the initial value: read Proto.role)
     App.Proto.state ('default'|'loading'|'empty'|'error'|'notfound')  setState(s) reloads with ?state=   App.Proto.annotations / .setAnnotations(bool)
     Everything marked [data-proto] is visible only when annotations are on (html.proto-annotations). */
  var PAGE = (location.pathname.split('/').pop() || 'index.html'), roleCfg = null, curRole = null;
  function plab(x) { return x.labelKey ? I18N.t(x.labelKey) : (x.label && typeof x.label === 'object' ? (x.label[I18N.lang] || x.label.fr) : x.label || x.id); }
  var Proto = {
    get role() { return curRole; }, get roles() { return roleCfg ? roleCfg.roles : []; },
    get state() { return Sim.forced() || P.get('state') || 'default'; },
    get annotations() { return root.classList.contains('proto-annotations'); },
    is: function (id) { return curRole === id; },
    register: function (cfg) {
      roleCfg = cfg; var ids = cfg.roles.map(function (r) { return r.id; }), r = P.get('role');
      if (ids.indexOf(r) < 0) r = store('proto.role.' + PAGE); if (ids.indexOf(r) < 0) r = cfg['default'] || ids[0];
      curRole = r; Proto.render();
    },
    setRole: function (id) {
      if (!roleCfg || id === curRole || roleCfg.roles.every(function (r) { return r.id !== id; })) return;
      var prev = curRole; curRole = id; store('proto.role.' + PAGE, id); Proto.render();
      d.dispatchEvent(new CustomEvent('proto:role', { detail: { role: id, previous: prev, roles: roleCfg.roles.map(function (r) { return r.id; }) } }));
    },
    setState: function (st) { var u = new URL(location.href); if (st === 'default') u.searchParams.delete('state'); else u.searchParams.set('state', st); location.href = u.href; },
    setAnnotations: function (on) { root.classList.toggle('proto-annotations', !!on); store('proto.annotations', on ? '1' : '0'); Proto.render(); },
    mount: function () {
      if ($('#proto-panel')) return;
      var fab = el('button', 'btn btn-icon proto-fab has-tip', '<span class="icon i-sliders" aria-hidden="true" data-testid="proto-toggle-icon"></span>');
      fab.type = 'button'; fab.id = 'proto-toggle'; fab.setAttribute('aria-expanded', 'false'); fab.setAttribute('aria-controls', 'proto-panel'); fab.setAttribute('data-testid', 'proto-toggle'); fab.setAttribute('data-i18n-attr', 'aria-label:proto.toggle'); fab.className += ' tip-top';
      var pan = el('div', 'proto-panel', '<h2 data-i18n="proto.title" data-testid="proto-title"></h2>' +
        '<fieldset class="proto-group" data-proto-roles hidden data-testid="proto-roles"><legend data-testid="proto-roles-legend"></legend><div class="proto-options" data-testid="proto-roles-options"></div></fieldset>' +
        '<fieldset class="proto-group" data-testid="proto-states"><legend data-i18n="proto.state" data-testid="proto-states-legend"></legend><div class="proto-options" data-testid="proto-states-options"></div></fieldset>' +
        '<button type="button" class="proto-switch" role="switch" aria-checked="false" data-proto-switch data-testid="proto-annotations"><span data-i18n="proto.annotations" data-testid="proto-annotations-label"></span><span class="proto-switch-track" aria-hidden="true" data-testid="proto-annotations-track"></span></button>');
      pan.id = 'proto-panel'; pan.hidden = true; pan.setAttribute('role', 'region'); pan.setAttribute('data-i18n-attr', 'aria-label:proto.title'); pan.setAttribute('data-testid', 'proto-panel');
      d.body.appendChild(pan); d.body.appendChild(fab);
      function toggle(open, ret) { pan.hidden = !open; fab.setAttribute('aria-expanded', open); if (!open && ret) fab.focus(); }
      fab.addEventListener('click', function () { toggle(pan.hidden); });
      pan.addEventListener('keydown', function (e) { if (e.key === 'Escape') { e.stopPropagation(); toggle(false, true); } });
      d.addEventListener('click', function (e) { if (!pan.hidden && !pan.contains(e.target) && !fab.contains(e.target)) toggle(false); });
      pan.addEventListener('change', function (e) { var r = e.target.closest('[data-proto-role]'), st = e.target.closest('[data-proto-state]'); if (r) Proto.setRole(r.value); else if (st) Proto.setState(st.value); });
      pan.addEventListener('click', function (e) { if (e.target.closest('[data-proto-switch]')) Proto.setAnnotations(!Proto.annotations); });
      I18N.apply(pan); I18N.apply(fab); Proto.render();
    },
    render: function () {
      var pan = $('#proto-panel'); if (!pan) return;
      function opts(name, list, cur, attr) { return list.map(function (o, i) { return '<label data-testid="proto-' + attr + '-option-' + o.id + '"><input type="radio" name="proto-' + attr + '" value="' + esc(o.id) + '" data-proto-' + attr + (o.id === cur ? ' checked' : '') + ' data-testid="proto-' + attr + '-input-' + o.id + '"><span data-testid="proto-' + attr + '-label-' + o.id + '">' + esc(o.text) + '</span></label>'; }).join(''); }
      var rg = $('[data-proto-roles]', pan);
      if (roleCfg) { rg.hidden = false; $('legend', rg).textContent = roleCfg.legend ? (roleCfg.legend[I18N.lang] || roleCfg.legend.fr) : I18N.t('proto.role'); $('.proto-options', rg).innerHTML = opts('role', roleCfg.roles.map(function (r) { return { id: r.id, text: plab(r) }; }), curRole, 'role'); }
      var cs = Proto.state === 'notfound' ? 'default' : Proto.state;
      $('[data-testid="proto-states-options"]', pan).innerHTML = opts('state', ['default', 'loading', 'empty', 'error'].map(function (k) { return { id: k, text: I18N.t('proto.state.' + k) }; }), cs, 'state');
      $('[data-proto-switch]', pan).setAttribute('aria-checked', Proto.annotations);
    }
  };
  d.addEventListener('i18n:change', function () { Proto.render(); });

  /* ---------- Delegated actions ---------- */
  d.addEventListener('click', function (e) {
    var a = e.target.closest('[data-action],[data-dialog-open],[data-dialog-close],[data-copy],[data-sim],[data-pw-toggle]');
    if (e.target.classList && e.target.classList.contains('dialog')) { Dialog.close(e.target); return; }   /* backdrop click */
    if (!a) return;
    var act = a.getAttribute('data-action');
    if (act === 'theme-toggle') Theme.toggle();
    else if (act === 'lang-set') I18N.set(a.getAttribute('data-lang'));
    else if (act === 'theme-set') Theme.set(a.getAttribute('data-theme-value'));
    else if (act === 'nav-toggle') navOpen(!d.body.classList.contains('nav-open'));
    else if (act === 'nav-close') navOpen(false);
    else if (a.hasAttribute('data-dialog-open')) Dialog.open(a.getAttribute('data-dialog-open'), a);
    else if (a.hasAttribute('data-dialog-close')) Dialog.close(a.closest('dialog'));
    else if (a.hasAttribute('data-pw-toggle')) {   /* <button class="btn btn-icon" data-pw-toggle="#pw" aria-pressed="false"> + .icon-main.i-eye + .icon-alt.i-eye-off */
      var inp = d.querySelector(a.getAttribute('data-pw-toggle')), show = a.getAttribute('aria-pressed') !== 'true';
      if (inp) inp.type = show ? 'text' : 'password'; a.setAttribute('aria-pressed', show); syncPw();
    } else if (a.hasAttribute('data-copy')) {
      var t = d.querySelector(a.getAttribute('data-copy')); var txt = t ? t.textContent : '';
      (navigator.clipboard ? navigator.clipboard.writeText(txt) : Promise.reject()).then(function () { a.classList.add('is-done'); setTimeout(function () { a.classList.remove('is-done'); }, 1600); Toast.show({ kind: 'success', msg: I18N.t('common.copied') }); }, function () { Toast.show({ kind: 'error', msg: I18N.t('common.copyFailed') }); });
    } else if (a.hasAttribute('data-sim')) { e.preventDefault(); Toast.show({ kind: 'info', title: I18N.t('sim.badge'), msg: I18N.t('sim.toast') }); }
  });
  /* Tooltip auto-flip: keep the CSS tooltip inside the viewport. Measures the ::after box (position:absolute, laid out even while hidden) and adds .tip-flip-end / .tip-flip-start. */
  function tipFit(e) {
    var el = e.target && e.target.closest ? e.target.closest('.btn-icon[aria-label], .has-tip[aria-label], [data-tip]') : null; if (el) tipOne(el);
  }
  function tipOne(el) {
    el.classList.remove('tip-flip-end', 'tip-flip-start');
    var w = parseFloat(getComputedStyle(el, '::after').width); if (!w) return;
    var r = el.getBoundingClientRect(), vw = d.documentElement.clientWidth, pad = 8, l, rt;
    if (el.classList.contains('tip-end')) { rt = r.right; l = rt - w; } else if (el.classList.contains('tip-start')) { l = r.left; rt = l + w; } else { l = r.left + r.width / 2 - w / 2; rt = l + w; }
    if (rt > vw - pad) el.classList.add('tip-flip-end'); else if (l < pad) el.classList.add('tip-flip-start');
  }
  d.addEventListener('mouseover', tipFit); d.addEventListener('focusin', tipFit);
  /* Hidden tooltips still count as scrollable overflow, so also fit them all once layout is stable. */
  function tipAll() { $$('.btn-icon[aria-label], .has-tip[aria-label], [data-tip]').forEach(tipOne); }
  window.addEventListener('load', function () { setTimeout(tipAll, 50); }); window.addEventListener('resize', tipAll); d.addEventListener('i18n:change', function () { setTimeout(tipAll, 0); });
  d.addEventListener('keydown', function (e) {
    if (e.key !== 'Escape') return;
    if (d.body.classList.contains('nav-open')) navOpen(false);
    root.classList.add('tips-off');   /* WCAG 1.4.13: Esc dismisses tooltips until the pointer or focus moves */
    var on = function () { root.classList.remove('tips-off'); d.removeEventListener('mousemove', on); d.removeEventListener('focusin', on); };
    d.addEventListener('mousemove', on); d.addEventListener('focusin', on);
  });


  /* ---------- State pill (Direction v2): App.statePill('accepted|verified|stale|unverified|open|closed|draft', {testid, label, labelKey}) ----------
     Icon + word on a neutral fill (.state-pill). Default labels come from the pill.* keys; pass label/labelKey for a gender-specific word. */
  var PILL = { accepted: 'check-circle', verified: 'shield-check', stale: 'clock', unverified: 'help', open: 'circle-dot', closed: 'lock', draft: 'edit' };
  function statePill(kind, o) {
    o = o || {}; var tid = o.testid || 'state-pill-' + kind, key = o.labelKey || 'pill.' + kind;
    return '<span class="state-pill state-pill-' + kind + '" data-testid="' + esc(tid) + '"><span class="icon i-' + PILL[kind] + '" aria-hidden="true" data-testid="' + esc(tid) + '-icon"></span><span' + (o.label ? '' : ' data-i18n="' + key + '"') + ' data-testid="' + esc(tid) + '-label">' + esc(o.label || I18N.t(key)) + '</span></span>';
  }

  /* ---------- Profile field editors, shared by the just-in-time prompts and the setup checklist ----------
     fieldEditor(key, {idp, onSaved}) -> <form class="jit-editor"> for country | github | bio | openToCollab | languages. Saving updates App.Session (and confirms languages / openToCollab). */
  var Session = window.App.Session;
  var PICON = { skills: 'tag', country: 'map-pin', github: 'git-branch', bio: 'file-text', openToCollab: 'users', languages: 'languages' };
  var GH = /^(?:https?:\/\/)?(?:www\.)?(?:github\.com\/)?@?([A-Za-z0-9](?:[A-Za-z0-9-]{0,38}))\/?$/;
  function fieldEditor(key, o) {
    o = o || {}; var idp = o.idp || ('jit-' + key), u = Session.user || {}, f = el('form', 'jit-editor'), M = window.MOCK || {}, lab = '<label class="label" for="' + idp + '-input" data-i18n="checklist.' + key + '.label" data-testid="' + idp + '-label"></label>', ctl = '';
    f.noValidate = true; f.setAttribute('data-testid', idp + '-form');
    if (key === 'country') ctl = '<select class="select" id="' + idp + '-input" data-testid="' + idp + '-input"><option value="" data-i18n="checklist.country.choose"></option>' + (M.countries || []).map(function (c) { return '<option value="' + c.code + '"' + (u.country === c.code ? ' selected' : '') + '>' + esc(c[I18N.lang] || c.fr) + '</option>'; }).join('') + '</select>';
    else if (key === 'github') ctl = '<input class="input" id="' + idp + '-input" type="text" inputmode="url" autocomplete="off" spellcheck="false" value="' + esc(u.github || '') + '" aria-describedby="' + idp + '-error" data-i18n-attr="placeholder:checklist.github.placeholder" data-testid="' + idp + '-input"><p class="error-msg" id="' + idp + '-error" role="alert" hidden data-i18n="checklist.github.error" data-testid="' + idp + '-error"></p>';
    else if (key === 'bio') ctl = '<textarea class="textarea" id="' + idp + '-input" rows="3" maxlength="280" lang="' + I18N.lang + '" data-user-content data-testid="' + idp + '-input">' + esc(u.bio || '') + '</textarea>';
    else if (key === 'openToCollab') { lab = ''; ctl = '<label class="choice"><input type="checkbox" id="' + idp + '-input"' + (u.openToCollab ? ' checked' : '') + ' data-testid="' + idp + '-input"><span data-i18n="checklist.openToCollab.choice" data-testid="' + idp + '-choice"></span></label>'; }
    else if (key === 'languages') { lab = ''; ctl = ['fr', 'en'].map(function (l) { return '<label class="choice"><input type="checkbox" value="' + l + '" id="' + idp + '-input-' + l + '"' + ((u.languages || [I18N.lang]).indexOf(l) > -1 ? ' checked' : '') + ' data-testid="' + idp + '-input-' + l + '"><span lang="' + l + '" data-testid="' + idp + '-choice-' + l + '">' + (l === 'fr' ? 'Français' : 'English') + '</span></label>'; }).join(''); }
    f.innerHTML = '<div class="field">' + lab + ctl + '</div><button type="submit" class="btn btn-primary btn-sm" data-testid="' + idp + '-save"><span data-i18n="common.save" data-testid="' + idp + '-save-label"></span></button>';
    f.addEventListener('submit', function (e) {
      e.preventDefault(); var inp = $('#' + idp + '-input', f), err = $('#' + idp + '-error', f);
      if (key === 'country') { if (!inp.value) { inp.focus(); return; } Session.update({ country: inp.value }); }
      else if (key === 'github') { var m = GH.exec(inp.value.trim()); if (!m) { inp.setAttribute('aria-invalid', 'true'); if (err) err.hidden = false; inp.focus(); return; } Session.update({ github: m[1] }); }
      else if (key === 'bio') { if (!inp.value.trim()) { inp.focus(); return; } Session.update({ bio: inp.value.trim() }); }
      else if (key === 'openToCollab') { Session.update({ openToCollab: inp.checked }); Session.confirm('openToCollab'); }
      else if (key === 'languages') { var ls = $$('input:checked', f).map(function (c) { return c.value; }); if (!ls.length) return; Session.update({ languages: ls }); Session.confirm('languages'); }
      if (o.onSaved) o.onSaved(key);
    });
    I18N.apply(f); return f;
  }

  /* ---------- Just-in-time prompt: App.Session.prompt(key, {container, textKey, actionKey, icon, href, onAction, testid}) ----------
     Returns the .jit-prompt element (appended to `container` when given) or null when signed out, already dismissed, or the item is done.
     Without href/onAction the action opens the matching inline editor (country, github, bio, openToCollab, languages) which ticks the same checklist item. Dismissal is remembered per key. */
  Session.prompt = function (key, o) {
    o = o || {}; if (!Session.signedIn || Session.promptDismissed(key)) return null;
    var it = Session.checklist().filter(function (i) { return i.key === key; })[0]; if (it && it.done) return null;
    var tid = o.testid || 'jit-' + key, box = el('div', 'jit-prompt'), tk = o.textKey || 'jit.' + key + '.text', ak = o.actionKey || 'jit.' + key + '.action';
    box.setAttribute('role', 'group'); box.setAttribute('aria-labelledby', tid + '-text'); box.setAttribute('data-testid', tid); box.setAttribute('data-jit', key);
    var act = o.href ? '<a class="btn btn-sm" href="' + esc(o.href) + '" data-testid="' + tid + '-action"><span data-i18n="' + ak + '" data-testid="' + tid + '-action-label"></span></a>' : '<button type="button" class="btn btn-sm" data-jit-action aria-expanded="false" data-testid="' + tid + '-action"><span data-i18n="' + ak + '" data-testid="' + tid + '-action-label"></span></button>';
    box.innerHTML = '<span class="icon i-' + (o.icon || PICON[key] || 'info') + '" aria-hidden="true" data-testid="' + tid + '-icon"></span><div class="jit-prompt-body" data-testid="' + tid + '-body"><p class="jit-prompt-text" id="' + tid + '-text" data-i18n="' + tk + '" data-testid="' + tid + '-text"></p>' + act + '</div>' +
      '<button type="button" class="btn btn-icon btn-sm" data-jit-dismiss data-i18n-attr="aria-label:common.dismiss" data-testid="' + tid + '-dismiss"><span class="icon i-x" aria-hidden="true" data-testid="' + tid + '-dismiss-icon"></span></button>';
    var gone = function () { d.removeEventListener('session:change', onChange); box.remove(); };
    var onChange = function () { var c = Session.checklist().filter(function (i) { return i.key === key; })[0]; if (!Session.signedIn || (c && c.done)) gone(); };
    d.addEventListener('session:change', onChange);
    box.querySelector('[data-jit-dismiss]').addEventListener('click', function () { Session.dismissPrompt(key); gone(); });
    var ab = box.querySelector('[data-jit-action]');
    if (ab) ab.addEventListener('click', function () {
      if (o.onAction) { o.onAction(box, key); return; }
      var body = $('.jit-prompt-body', box); ab.hidden = true; ab.setAttribute('aria-expanded', 'true');
      var f = fieldEditor(key, { idp: tid, onSaved: function () { if (o.onSaved) o.onSaved(key); } }); body.appendChild(f); var c = $('input, select, textarea', f); if (c) c.focus();
    });
    if (o.container) (typeof o.container === 'string' ? $(o.container) : o.container).appendChild(box);
    I18N.apply(box); return box;
  };

  /* ---------- Setup checklist: App.setupChecklist(container, {testid}) ----------
     ONLY for the user's own profile (self view), never the home page. Renders the .setup-checklist card (meter + rows + inline editors) from App.Session.checklist().
     Pending items open an editor in place; saving ticks the row. At 100% a short "Profil complet" confirmation shows, then the card is dismissed for good (onboarding.checklistDismissed). Collapsible (onboarding.checklistCollapsed). */
  function setupChecklist(box, o) {
    o = o || {}; box = typeof box === 'string' ? $(box) : box; if (!box) return null;
    var tid = o.testid || 'setup', openKey = null, prevDone = null, timer = null;
    function render(focusKey) {
      var s = Session.state, cl = Session.checklist(); clearTimeout(timer);
      if (!s.signedIn || s.onboarding.checklistDismissed) { box.innerHTML = ''; return; }
      var col = !!s.onboarding.checklistCollapsed, complete = cl.pct === 100, now = {}; cl.forEach(function (i) { now[i.key] = i.done; });
      var rows = cl.map(function (i) {
        var just = prevDone && !prevDone[i.key] && i.done, p = tid + '-item-' + i.key, op = openKey === i.key && !i.done;
        var action = i.done ? '' : (i.editor === 'dialog' ? '<a class="btn btn-sm" href="' + esc(i.href) + '" data-testid="' + p + '-action"><span data-i18n="checklist.' + i.key + '.action" data-testid="' + p + '-action-label"></span></a>' :
          '<button type="button" class="btn btn-sm" data-setup-open="' + i.key + '" aria-expanded="' + op + '" aria-controls="' + p + '-editor" data-testid="' + p + '-action"><span data-i18n="checklist.' + i.key + '.action" data-testid="' + p + '-action-label"></span></button>');
        return '<li class="setup-item' + (i.done ? ' is-done' : '') + (just ? ' just-done' : '') + '" data-testid="' + p + '"><div class="setup-row"><span class="icon icon-state i-' + (i.done ? 'check-circle' : 'circle') + '" aria-hidden="true" data-testid="' + p + '-icon"></span>' +
          '<div><p class="setup-label" data-i18n="' + i.labelKey + '" data-testid="' + p + '-label"></p>' + (i.done ? '<span class="visually-hidden" data-i18n="checklist.done" data-testid="' + p + '-done"></span>' : '<p class="setup-value" data-i18n="' + i.valueKey + '" data-testid="' + p + '-value"></p>') + '</div>' + action + '</div>' +
          '<div class="setup-editor" id="' + p + '-editor" data-editor="' + i.key + '"' + (op ? '' : ' hidden') + ' data-testid="' + p + '-editor"></div></li>';
      }).join('');
      box.innerHTML = '<section class="setup-checklist card" aria-labelledby="' + tid + '-title" data-testid="' + tid + '"><div class="setup-head"><h2 id="' + tid + '-title" data-i18n="checklist.title" data-testid="' + tid + '-title"></h2>' +
        '<div class="setup-meter" data-testid="' + tid + '-meter"><progress max="100" value="' + cl.pct + '" data-i18n-attr="aria-label:checklist.title" data-testid="' + tid + '-progress"></progress><span aria-hidden="true" data-testid="' + tid + '-count">' + cl.done + '/' + cl.total + '</span><span class="visually-hidden" data-testid="' + tid + '-pct">' + cl.pct + ' %</span></div>' +
        '<button type="button" class="btn btn-icon btn-sm setup-collapse" data-setup-collapse aria-expanded="' + !col + '" aria-controls="' + tid + '-list" data-i18n-attr="aria-label:checklist.toggle" data-testid="' + tid + '-collapse"><span class="icon i-chevron-down" aria-hidden="true" data-testid="' + tid + '-collapse-icon"></span></button></div>' +
        '<div id="' + tid + '-list"' + (col ? ' hidden' : '') + ' data-testid="' + tid + '-body">' + (complete ? '<p class="setup-done" role="status" data-testid="' + tid + '-complete"><span class="icon i-check-circle" aria-hidden="true" data-testid="' + tid + '-complete-icon"></span><span data-i18n="checklist.complete" data-testid="' + tid + '-complete-label"></span></p>' : '') +
        '<ul class="setup-list list-reset" data-testid="' + tid + '-list">' + rows + '</ul></div></section>';
      if (openKey && !now[openKey]) { var ed = $('[data-editor="' + openKey + '"]', box); if (ed) ed.appendChild(fieldEditor(openKey, { idp: tid + '-edit-' + openKey })); }
      I18N.apply(box); prevDone = now;
      if (focusKey) { var f = $('[data-editor="' + focusKey + '"] input, [data-editor="' + focusKey + '"] select, [data-editor="' + focusKey + '"] textarea', box); if (f) f.focus(); }
      if (complete) timer = setTimeout(function () { Session.setOnboarding({ checklistDismissed: true }); }, 3500);
    }
    if (!box._setupWired) {
      box._setupWired = true;
      box.addEventListener('click', function (e) {
        var op = e.target.closest('[data-setup-open]'), co = e.target.closest('[data-setup-collapse]');
        if (op) { var k = op.getAttribute('data-setup-open'); openKey = openKey === k ? null : k; render(openKey); }
        else if (co) { Session.setOnboarding({ checklistCollapsed: !Session.state.onboarding.checklistCollapsed }); }
      });
      box.addEventListener('keydown', function (e) { if (e.key === 'Escape' && openKey) { var k = openKey; openKey = null; render(); var b = $('[data-setup-open="' + k + '"]', box); if (b) b.focus(); } });
      d.addEventListener('session:change', function () { if (openKey && Session.checklist().filter(function (i) { return i.key === openKey; })[0].done) openKey = null; render(); });
    }
    render(); return { render: render };
  }

  function init() {
    I18N.apply(); syncTheme(); syncNav(); Tabs.init(d); Menu.init(d); Proto.mount();
    $$('.js-only').forEach(function (n) { n.hidden = false; });
    if (P.get('autotestid') === '1') TestIds.auto();   /* debug only: positional fallback ids; real ids are hand-written */
  }
  if (d.readyState === 'complete') init(); else d.addEventListener('DOMContentLoaded', init);   /* deferred scripts (dictionaries, pages) run before this */
  /* ---------- View mode: App.View.init(toggleEl, listSelector, key) ----------
     Renders a list/grid segmented control into toggleEl and toggles .is-grid on the target list(s).
     Priority: ?view=list|grid (QA, not persisted) > saved choice per key > 'list'. Re-applies after re-renders
     (call App.View.apply(key) or rely on the MutationObserver). Emits 'view:change' {key, mode} on document. */
  var View = (function () {
    var reg = {};
    function mode(key) { var q = P.get('view'); if (q === 'grid' || q === 'list') return q; var s = store('view.' + key); return s === 'grid' ? 'grid' : 'list'; }
    function apply(key) {
      var r = reg[key]; if (!r) return; var m = r.mode;
      $$(r.target).forEach(function (l) { l.classList.toggle('is-grid', m === 'grid'); });
      $$('[data-view]', r.el).forEach(function (b) { b.setAttribute('aria-pressed', String(b.getAttribute('data-view') === m)); });
    }
    function set(key, m) { var r = reg[key]; if (!r) return; r.mode = m; store('view.' + key, m); apply(key); d.dispatchEvent(new CustomEvent('view:change', { detail: { key: key, mode: m } })); }
    function render(r, key) {
      var t = window.I18N ? I18N.t.bind(I18N) : function (k) { return k; }, id = r.el.getAttribute('data-testid') || ('view-' + key);
      r.el.setAttribute('role', 'group'); r.el.setAttribute('aria-label', t('common.view'));
      r.el.innerHTML = ['list', 'grid'].map(function (m) {
        return '<button type="button" class="btn btn-icon btn-sm has-tip" data-view="' + m + '" aria-pressed="false" aria-label="' + esc(t('common.view.' + m)) + '" data-tip="' + esc(t('common.view.' + m)) + '" data-testid="' + id + '-' + m + '"><span class="icon i-view-' + m + '" aria-hidden="true" data-testid="' + id + '-' + m + '-icon"></span></button>';
      }).join('');
    }
    function init(el, target, key) {
      if (!el) return; key = key || location.pathname.split('/').pop().replace('.html', '');
      var r = reg[key] = { el: el, target: target, mode: mode(key) };
      el.classList.add('view-toggle'); render(r, key); apply(key);
      el.addEventListener('click', function (e) { var b = e.target.closest('[data-view]'); if (b) set(key, b.getAttribute('data-view')); });
      d.addEventListener('i18n:change', function () { render(r, key); apply(key); });
      $$(target).forEach(function (l) { new MutationObserver(function () { if (l.classList.contains('is-grid') !== (r.mode === 'grid')) apply(key); }).observe(l, { attributes: true, attributeFilter: ['class'] }); });
      return r;
    }
    function auto() { $$('[data-view-toggle]').forEach(function (el) { if (!el.hasAttribute('data-view-ready')) { el.setAttribute('data-view-ready', ''); init(el, el.getAttribute('data-view-target'), el.getAttribute('data-view-toggle')); } }); }
    if (d.readyState === 'loading') d.addEventListener('DOMContentLoaded', auto); else setTimeout(auto, 0);
    return { init: init, set: set, apply: apply, mode: function (k) { return reg[k] ? reg[k].mode : mode(k); }, auto: auto };
  })();

  window.App = Object.assign(window.App || {}, { View: View, initials: initials, statePill: statePill, fieldEditor: fieldEditor, setupChecklist: setupChecklist, avatar: avatar, thumb: thumb, rating: rating, ratingInput: ratingInput, review: review, Menu: Menu, Proto: Proto, store: store, $: $, $$: $$, el: el, esc: esc, Filter: Filter, Sim: Sim, Tabs: Tabs, TestIds: TestIds, Toast: Toast, Dialog: Dialog, Theme: Theme, params: P });
})();
