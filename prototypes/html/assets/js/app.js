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

  /* ---------- Tabs (ARIA tabs pattern, automatic activation) ---------- */
  function initTabs(box) {
    var tabs = $$('[role="tab"]', box), panels = tabs.map(function (t) { return d.getElementById(t.getAttribute('aria-controls')); });
    function select(i, focus) {
      tabs.forEach(function (t, j) { var on = i === j; t.setAttribute('aria-selected', on); t.tabIndex = on ? 0 : -1; if (panels[j]) panels[j].hidden = !on; });
      if (focus) tabs[i].focus();
    }
    tabs.forEach(function (t, i) {
      t.addEventListener('click', function () { select(i); });
      t.addEventListener('keydown', function (e) {
        var n = { ArrowRight: i + 1, ArrowLeft: i - 1, Home: 0, End: tabs.length - 1 }[e.key];
        if (n == null) return; e.preventDefault(); select((n + tabs.length) % tabs.length, true);
      });
    });
    var cur = tabs.findIndex(function (t) { return t.getAttribute('aria-selected') === 'true'; }); select(cur < 0 ? 0 : cur);
  }

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
     Sim.load(container, {load: ()=>array, render:(items, container)=>void, delay, skeleton, emptyKey, rows})
     Returns {run, refresh} (refresh = recompute synchronously, e.g. after a filter change).
     Forced state via ?state=loading|empty|error. Always shows a visible "Simulated" badge. */
  var sims = [];
  var Sim = {
    load: function (box, o) {
      var forced = P.get('state'), items = null, st = 'loading';
      if (!$('[data-sim-marker]')) {   /* one marker per page is enough; add data-sim-marker yourself to place it */
        var m = el('p', 'sim-marker text-sm text-muted', '<span class="badge badge-sim" data-i18n="sim.badge"></span> <span data-i18n="sim.note"></span>');
        m.setAttribute('data-sim-marker', ''); m.setAttribute('data-testid', 'sim-marker'); box.parentNode.insertBefore(m, box); I18N.apply(m);
      }
      box.setAttribute('aria-live', 'polite');
      function state(cls, icon, key, retry) {
        box.innerHTML = '<div class="state ' + cls + '" data-testid="state-' + st + '"><span class="icon i-' + icon + '" aria-hidden="true"></span><h3>' + esc(I18N.t(key + '.title')) + '</h3><p>' + esc(I18N.t(key + '.body')) + '</p>' +
          (retry ? '<button type="button" class="btn btn-sm" data-sim-retry data-testid="state-retry"><span class="icon i-refresh" aria-hidden="true"></span>' + esc(I18N.t('state.retry')) + '</button>' : '') + '</div>';
      }
      function draw() {
        box.setAttribute('aria-busy', st === 'loading'); box.setAttribute('data-state', st);
        if (st === 'loading') { var n = o.skeleton || 3, h = ''; while (n--) h += '<div class="skeleton-row"><span class="skeleton"></span><span class="skeleton"></span></div>';
          box.innerHTML = '<div class="list" data-testid="state-loading" role="status" aria-label="' + esc(I18N.t('state.loading')) + '">' + h + '</div>'; }
        else if (st === 'error') state('state-error', 'alert-circle', 'state.error', true);
        else if (st === 'empty') state('', 'inbox', o.emptyKey || 'state.empty', false);
        else o.render(items, box);
        I18N.apply(box);
      }
      function run() {
        st = 'loading'; draw();
        if (forced === 'loading') return;
        setTimeout(function () {
          try { items = forced === 'error' ? (function () { throw new Error('sim'); })() : forced === 'empty' ? [] : o.load(); st = items.length ? 'success' : 'empty'; } catch (e) { st = 'error'; }
          draw();
        }, o.delay == null ? 600 : o.delay);
      }
      box.addEventListener('click', function (e) { if (e.target.closest('[data-sim-retry]')) { forced = null; run(); } });
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
    I18N.apply(); syncTheme(); syncNav(); $$('[data-tabs]').forEach(initTabs);
    $$('.js-only').forEach(function (n) { n.hidden = false; });
  }
  if (d.readyState === 'complete') init(); else d.addEventListener('DOMContentLoaded', init);   /* deferred scripts (dictionaries, pages) run before this */
  window.App = { $: $, $$: $$, el: el, esc: esc, Filter: Filter, Sim: Sim, Toast: Toast, Dialog: Dialog, Theme: Theme, params: P };
})();
