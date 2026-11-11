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
  function avatar(p, o) {
    o = o || {}; var name = typeof p === 'string' ? p : (p && p.name), ini = (p && p.initials) || initials(name);
    return '<span class="avatar' + (o.size ? ' avatar-' + o.size : '') + '" aria-hidden="true"' + (o.testid ? ' data-testid="' + esc(o.testid) + '"' : '') + '>' + esc(ini) + '</span>';
  }

  /* ---------- i18n: UI strings only, never user content ---------- */
  var dicts = { fr: {}, en: {} }, lang = P.get('lang') || store('lang');
  if (lang !== 'fr' && lang !== 'en') lang = 'fr';              /* French is the default UI language */
  root.lang = lang;
  var warned = {};
  var I18N = {
    get lang() { return lang; },
    register: function (dict) { Object.keys(dict).forEach(function (l) { dicts[l] = Object.assign(dicts[l] || {}, dict[l]); }); },
    t: function (k, p) {
      var s = dicts[lang][k]; if (s == null) s = dicts.fr[k];
      if (s == null) { if (!warned[k]) { warned[k] = 1; console.warn('[i18n] missing key', k); } return k; }
      return p ? s.replace(/\{(\w+)\}/g, function (m, n) { return p[n] != null ? p[n] : m; }) : s;
    },
    apply: function (r) {
      $$('[data-i18n]', r).forEach(function (n) { n.textContent = I18N.t(n.getAttribute('data-i18n')); });
      $$('[data-i18n-attr]', r).forEach(function (n) {
        n.getAttribute('data-i18n-attr').split(';').forEach(function (pair) { var a = pair.split(':'); if (a[1]) n.setAttribute(a[0].trim(), I18N.t(a[1].trim())); });
      });
      $$('[data-lang]').forEach(function (b) { b.setAttribute('aria-pressed', b.getAttribute('data-lang') === lang); });
    },
    set: function (l) {
      if (l !== 'fr' && l !== 'en') return; lang = l; root.lang = l; store('lang', l);
      I18N.apply(); d.dispatchEvent(new CustomEvent('i18n:change', { detail: { lang: l } }));
    }
  };
  window.I18N = I18N;

  /* ---------- Theme ---------- */
  function isDark() { var t = root.getAttribute('data-theme'); return t ? t === 'dark' : matchMedia('(prefers-color-scheme: dark)').matches; }
  function syncTheme() { $$('[data-action="theme-toggle"]').forEach(function (b) { b.setAttribute('aria-pressed', isDark()); }); }
  var Theme = { set: function (t) { root.setAttribute('data-theme', t); store('theme', t); syncTheme(); }, toggle: function () { Theme.set(isDark() ? 'light' : 'dark'); } };

  /* ---------- Toast ---------- */
  var Toast = {
    show: function (o) {
      var reg = $('.toast-region') || d.body.appendChild(el('div', 'toast-region'));
      reg.setAttribute('aria-live', 'polite');
      var k = o.kind || 'info', ic = { success: 'check-circle', error: 'alert-circle', info: 'info' }[k];
      var t = el('div', 'toast toast-' + k); t.setAttribute('role', k === 'error' ? 'alert' : 'status');
      t.setAttribute('data-testid', 'toast-' + k);
      t.innerHTML = '<span class="icon icon-lead i-' + ic + '" aria-hidden="true"></span><div>' + (o.title ? '<strong>' + esc(o.title) + '</strong>' : '') + esc(o.msg || '') +
        '</div><button class="btn btn-ghost btn-sm btn-icon" type="button" data-testid="toast-close" aria-label="' + esc(I18N.t('common.close')) + '"><span class="icon i-x" aria-hidden="true"></span></button>';
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

  /* ---------- Fallback data-testid assigner (CONVENTIONS.md section 5) ----------
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
    var m = el('p', 'sim-marker text-sm text-muted', '<span class="badge badge-sim" data-i18n="sim.badge"></span><span data-i18n="sim.note"></span>');
    m.setAttribute('data-sim-marker', ''); m.setAttribute('data-testid', 'sim-marker'); box.parentNode.insertBefore(m, box); I18N.apply(m);
  }
  function actionHtml(a, acts) {
    if (!a) return '';
    var ic = a.icon ? '<span class="icon i-' + a.icon + '" aria-hidden="true"></span>' : '', lab = esc(a.label != null ? a.label : I18N.t(a.labelKey)),
      cls = 'btn btn-sm' + (a.primary ? ' btn-primary' : ''), tid = a.testid ? ' data-testid="' + esc(a.testid) + '"' : '';
    if (a.href) return '<a class="' + cls + '" href="' + esc(a.href) + '"' + tid + '>' + ic + lab + '</a>';
    acts.push(a.onClick || function () {}); return '<button type="button" class="' + cls + '" data-sim-act="' + (acts.length - 1) + '"' + tid + '>' + ic + lab + '</button>';
  }
  function stateHtml(kind, o, acts) {
    var base = kind === 'error' ? { icon: 'alert-circle', key: 'state.error', cls: ' state-error' } : { icon: 'inbox', key: o.emptyKey || 'state.empty', cls: '' };
    var c = Object.assign({}, base, (kind === 'error' ? o.error : o.empty) || {}), hd = c.heading || 'h3';
    var act = c.action ? actionHtml(c.action, acts) : '', retry = '';
    if (kind === 'error' && c.retry !== false) retry = '<button type="button" class="btn btn-sm" data-sim-retry data-testid="state-retry"><span class="icon i-refresh" aria-hidden="true"></span>' + esc(I18N.t('state.retry')) + '</button>';
    return '<div class="state' + (c.cls || '') + '" data-testid="state-' + kind + '"><span class="icon i-' + c.icon + '" aria-hidden="true"></span><' + hd + (hd === 'h3' ? '' : ' class="h3"') + '>' + esc(I18N.t(c.titleKey || c.key + '.title')) + '</' + hd + '><p>' +
      esc(I18N.t(c.bodyKey || c.key + '.body')) + '</p>' + (act || retry ? '<div class="state-actions">' + act + retry + '</div>' : '') + '</div>';
  }
  function skeletonHtml(n) {
    var h = ''; n = n || 3; while (n--) h += '<div class="skeleton-row"><span class="skeleton"></span><span class="skeleton"></span></div>';
    return '<div class="list" data-testid="state-loading" role="status" aria-label="' + esc(I18N.t('state.loading')) + '">' + h + '</div>';
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
        box.innerHTML = kind === 'loading' ? skeletonHtml(o.skeleton) : stateHtml(kind, { error: o, empty: o }, acts); I18N.apply(box);
      }
      box._simDraw = draw; draw();
      if (!box._simLang) { box._simLang = 1; d.addEventListener('i18n:change', function () { box._simDraw(); }); }
    },
    load: function (box, o) {
      var forced = P.get('state'), items = null, st = 'loading';
      ensureMarker(box); box.setAttribute('aria-live', 'polite');
      function draw() {
        var acts = []; box._simActs = acts; box.setAttribute('aria-busy', st === 'loading'); box.setAttribute('data-state', st);
        if (st === 'loading') box.innerHTML = skeletonHtml(o.skeleton);
        else if (st === 'error' || st === 'empty') box.innerHTML = stateHtml(st, o, acts);
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

  /* ---------- Delegated actions ---------- */
  d.addEventListener('click', function (e) {
    var a = e.target.closest('[data-action],[data-dialog-open],[data-dialog-close],[data-copy],[data-sim]'); 
    if (e.target.classList && e.target.classList.contains('dialog')) { Dialog.close(e.target); return; }   /* backdrop click */
    if (!a) return;
    var act = a.getAttribute('data-action');
    if (act === 'theme-toggle') Theme.toggle();
    else if (act === 'lang-set') I18N.set(a.getAttribute('data-lang'));
    else if (act === 'nav-toggle') navOpen(!d.body.classList.contains('nav-open'));
    else if (act === 'nav-close') navOpen(false);
    else if (a.hasAttribute('data-dialog-open')) Dialog.open(a.getAttribute('data-dialog-open'), a);
    else if (a.hasAttribute('data-dialog-close')) Dialog.close(a.closest('dialog'));
    else if (a.hasAttribute('data-copy')) {
      var t = d.querySelector(a.getAttribute('data-copy')); var txt = t ? t.textContent : '';
      (navigator.clipboard ? navigator.clipboard.writeText(txt) : Promise.reject()).then(function () { Toast.show({ kind: 'success', msg: I18N.t('common.copied') }); }, function () { Toast.show({ kind: 'error', msg: I18N.t('common.copyFailed') }); });
    } else if (a.hasAttribute('data-sim')) { e.preventDefault(); Toast.show({ kind: 'info', title: I18N.t('sim.badge'), msg: I18N.t('sim.toast') }); }
  });
  d.addEventListener('keydown', function (e) { if (e.key === 'Escape' && d.body.classList.contains('nav-open')) navOpen(false); });

  function init() {
    I18N.apply(); syncTheme(); syncNav(); Tabs.init(d);
    $$('.js-only').forEach(function (n) { n.hidden = false; });
    TestIds.auto();
  }
  if (d.readyState === 'complete') init(); else d.addEventListener('DOMContentLoaded', init);   /* deferred scripts (dictionaries, pages) run before this */
  window.App = { initials: initials, avatar: avatar, $: $, $$: $$, el: el, esc: esc, Filter: Filter, Sim: Sim, Tabs: Tabs, TestIds: TestIds, Toast: Toast, Dialog: Dialog, Theme: Theme, params: P };
})();
