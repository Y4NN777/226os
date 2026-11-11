/* Identity P0: sign-in + create account (PRD 5.4). 100% SIMULATED: nothing is sent, stored or authenticated.
   Security-faithful UX per G0-AUTH-001..004: generic sign-in error, generic sign-up/reset responses (no enumeration),
   rate-limit (429) state, session-expired state, email-verification pending. Force states with ?state=expired|ratelimit|badcreds|verify and ?tab=signup. */
(function () {
  'use strict';
  var nb = ' ';
  I18N.register({
    fr: {
      'auth.title': 'Connexion — 226os', 'auth.h1': 'Connexion et création de compte',
      'auth.lede': 'Un compte permet de poser des questions, voter, postuler à un projet et avoir un profil. Consulter le site ne demande pas de compte.',
      'auth.sim.text': 'Prototype : aucun compte n’est créé, aucun mot de passe n’est envoyé ni stocké.',
      'auth.demo.summary': 'Scénarios de test',
      'auth.demo.ok': 'Connexion réussie : adresse valide et mot de passe d’au moins 12 caractères.',
      'auth.demo.bad': 'Mot de passe plus court : erreur générique « E-mail ou mot de passe incorrect ».',
      'auth.demo.limit': 'limit@example.org : simule trop de tentatives (429).',
      'auth.demo.exists': 'Création : même réponse que l’adresse existe déjà (exists@example.org) ou non.',
      'auth.demo.states': 'États forcés : ?state=expired, ratelimit, badcreds ou verify ; ?tab=signup ouvre la création de compte.',
      'auth.card.label': 'Accès au compte', 'auth.tab.signin': 'Se connecter', 'auth.tab.signup': 'Créer un compte',
      'auth.email': 'Adresse e-mail', 'auth.password': 'Mot de passe', 'auth.pw.show': 'Afficher', 'auth.pw.hide': 'Masquer',
      'auth.pw.hint': '12 caractères minimum. Une phrase de passe fonctionne très bien.',
      'auth.name': 'Nom affiché', 'auth.name.hint': 'Public sur votre profil. Un pseudonyme convient.',
      'auth.more': 'Profil (facultatif)', 'auth.country': 'Pays', 'auth.country.choose': 'Choisir un pays', 'auth.langs': 'Langues', 'auth.langs.hint': 'Au moins une. Affichées sur votre profil ; vos contenus ne sont pas traduits.',
      'auth.lang.fr': 'Français', 'auth.lang.en': 'Anglais',
      'auth.privacy': 'Votre adresse e-mail reste privée et n’apparaît jamais sur votre profil public.',
      'auth.signin.submit': 'Se connecter', 'auth.signin.busy': 'Connexion…', 'auth.signup.submit': 'Créer mon compte', 'auth.signup.busy': 'Création…',
      'auth.signin.noaccount': 'Pas encore de compte' + nb + '?', 'auth.signup.hasaccount': 'Vous avez déjà un compte' + nb + '?',
      'auth.forgot.open': 'Mot de passe oublié' + nb + '?',
      'auth.summary.title': 'Le formulaire contient des erreurs',
      'auth.err.email.required': 'Saisissez votre adresse e-mail.', 'auth.err.email.format': 'Saisissez une adresse e-mail valide, par exemple nom@domaine.org.',
      'auth.err.password.required': 'Saisissez votre mot de passe.', 'auth.err.password.short': 'Le mot de passe doit contenir au moins 12 caractères.',
      'auth.err.name.required': 'Saisissez un nom affiché.', 'auth.err.name.short': 'Le nom affiché doit contenir au moins 2 caractères.',
      'auth.err.country.required': 'Choisissez un pays.', 'auth.err.langs.required': 'Choisissez au moins une langue.',
      'auth.banner.bad.title': 'Connexion impossible', 'auth.banner.bad.body': 'E-mail ou mot de passe incorrect.',
      'auth.banner.limit.title': 'Trop de tentatives (429)', 'auth.banner.limit.body': 'Connexion suspendue pour protéger les comptes. Réessayez dans {n} secondes.',
      'auth.banner.limit.body0': 'Vous pouvez réessayer maintenant.',
      'auth.banner.expired.title': 'Session expirée', 'auth.banner.expired.body': 'Votre session a expiré. Reconnectez-vous pour continuer.',
      'auth.banner.verify.title': 'E-mail à confirmer', 'auth.banner.verify.body': 'Ouvrez le lien reçu par e-mail pour activer votre compte. Pensez aux spams.',
      'auth.banner.verify.resend': 'Renvoyer l’e-mail',
      'auth.signin.limited': 'Réessayer dans {n} s',
      'auth.ok.title': 'Connexion réussie', 'auth.ok.body': 'Redirection vers votre profil…', 'auth.ok.link': 'Aller au profil',
      'auth.verify.title': 'Vérifiez votre boîte de réception', 'auth.verify.body': 'Si un compte peut être créé avec cette adresse, un e-mail de vérification vous est envoyé. Ce message est le même que l’adresse soit déjà utilisée ou non.',
      'auth.verify.sim': 'Aucun e-mail n’est réellement envoyé dans ce prototype.',
      'auth.verify.confirm': 'Simuler le clic sur le lien de vérification', 'auth.verify.resend': 'Renvoyer l’e-mail', 'auth.verify.back': 'Retour à la connexion',
      'auth.resend.toast.title': 'Demande envoyée', 'auth.resend.toast.msg': 'Si l’adresse est valide, un nouvel e-mail part.',
      'auth.forgot.title': 'Réinitialiser le mot de passe', 'auth.forgot.sim': 'Aucun e-mail n’est envoyé dans ce prototype.',
      'auth.forgot.body': 'Saisissez l’adresse de votre compte. La réponse est la même, que le compte existe ou non.', 'auth.forgot.submit': 'Envoyer le lien',
      'auth.forgot.done.title': 'Demande reçue. ', 'auth.forgot.done.body': 'Si un compte existe, vous recevrez un lien de réinitialisation à usage unique, valable peu de temps.',
      'auth.sec.h': 'Protections prévues côté serveur', 'auth.sec.1': 'Mot de passe haché (Argon2id) ; session serveur opaque, cookie HttpOnly, SameSite=Lax et jeton CSRF sur chaque action.',
      'auth.sec.2': 'Réponses génériques à la connexion, à l’inscription et à la réinitialisation : on ne peut pas deviner quelles adresses ont un compte.',
      'auth.sec.3': 'Limites de tentatives par adresse IP et par compte ; réinitialisation et vérification par lien à usage unique (exigences G0, à valider avant la production).'
    },
    en: {
      'auth.title': 'Sign in — 226os', 'auth.h1': 'Sign in and create an account',
      'auth.lede': 'An account lets you ask questions, vote, apply to a project and keep a minimal profile. Browsing stays possible without an account.',
      'auth.sim.text': 'Prototype: no account is created, no password is sent or stored.',
      'auth.demo.summary': 'Demo scenarios',
      'auth.demo.ok': 'Sign-in: a valid address and a password of 12 characters or more succeeds and leads to the profile.',
      'auth.demo.bad': 'Sign-in: a shorter password shows the generic error “email or password incorrect”.',
      'auth.demo.limit': 'Sign-in: the address limit@example.org simulates too many attempts (429).',
      'auth.demo.exists': 'Sign-up: the response is identical whether the address already exists (exists@example.org) or not.',
      'auth.demo.states': 'Forced states: ?state=expired, ratelimit, badcreds or verify; ?tab=signup opens account creation.',
      'auth.card.label': 'Account access', 'auth.tab.signin': 'Sign in', 'auth.tab.signup': 'Create account',
      'auth.email': 'Email address', 'auth.password': 'Password', 'auth.pw.show': 'Show', 'auth.pw.hide': 'Hide',
      'auth.pw.hint': 'At least 12 and at most 128 characters. A passphrase works well; spaces allowed, no forced upper-case or symbol rules.',
      'auth.name': 'Display name', 'auth.name.hint': 'Public on your profile. A nickname is fine.',
      'auth.more': 'More details (optional)', 'auth.country': 'Country', 'auth.country.choose': 'Choose a country', 'auth.langs': 'Languages', 'auth.langs.hint': 'At least one. Used for your profile, not to translate your content.',
      'auth.lang.fr': 'French', 'auth.lang.en': 'English',
      'auth.privacy': 'Your email address stays private and never appears on your public profile.',
      'auth.signin.submit': 'Sign in', 'auth.signin.busy': 'Signing in…', 'auth.signup.submit': 'Create my account', 'auth.signup.busy': 'Creating…',
      'auth.signin.noaccount': 'No account yet?', 'auth.signup.hasaccount': 'Already have an account?',
      'auth.forgot.open': 'Forgot password?',
      'auth.summary.title': 'The form has errors',
      'auth.err.email.required': 'Enter your email address.', 'auth.err.email.format': 'Enter a valid email address, for example name@domain.org.',
      'auth.err.password.required': 'Enter your password.', 'auth.err.password.short': 'The password must be at least 12 characters long.',
      'auth.err.name.required': 'Enter a display name.', 'auth.err.name.short': 'The display name must be at least 2 characters long.',
      'auth.err.country.required': 'Choose a country.', 'auth.err.langs.required': 'Choose at least one language.',
      'auth.banner.bad.title': 'Cannot sign in', 'auth.banner.bad.body': 'Email or password incorrect. Check your details and try again.',
      'auth.banner.limit.title': 'Too many attempts (429)', 'auth.banner.limit.body': 'To protect accounts, sign-in is temporarily paused. Try again in {n} seconds.',
      'auth.banner.limit.body0': 'You can try again now.',
      'auth.banner.expired.title': 'Session expired', 'auth.banner.expired.body': 'For your security you were signed out. Sign in again to continue; unsent input is not kept.',
      'auth.banner.verify.title': 'Email verification pending', 'auth.banner.verify.body': 'Open the link sent by email to activate your account. Check your spam folder too.',
      'auth.banner.verify.resend': 'Resend email',
      'auth.signin.limited': 'Try again in {n} s',
      'auth.ok.title': 'Signed in', 'auth.ok.body': 'Redirecting to your profile…', 'auth.ok.link': 'Go to profile',
      'auth.verify.title': 'Check your inbox', 'auth.verify.body': 'If an account can be created with this address, you will receive an email with a verification link. The message is the same whether the address is already in use or not.',
      'auth.verify.sim': 'No email is actually sent in this prototype.',
      'auth.verify.confirm': 'Simulate clicking the verification link', 'auth.verify.resend': 'Resend email', 'auth.verify.back': 'Back to sign in',
      'auth.resend.toast.title': 'Request received', 'auth.resend.toast.msg': 'If the address is eligible, a new message is sent .',
      'auth.forgot.title': 'Reset password', 'auth.forgot.sim': 'No email is sent in this prototype.',
      'auth.forgot.body': 'Enter your account address. We answer the same way whether an account exists or not.', 'auth.forgot.submit': 'Send link',
      'auth.forgot.done.title': 'Request received. ', 'auth.forgot.done.body': 'If an account exists for this address, a single-use, short-lived reset link is sent to you.',
      'auth.sec.h': 'Server-side protections planned', 'auth.sec.1': 'Hashed password (Argon2id); opaque server session, HttpOnly SameSite=Lax cookie and a CSRF token on every action.',
      'auth.sec.2': 'Generic responses on sign-in, sign-up and reset: nobody can guess which addresses have an account.',
      'auth.sec.3': 'Attempt limits per IP address and per account; reset and verification by single-use links (G0 requirements, to be validated before production).'
    }
  });

  var $ = App.$, $$ = App.$$, t = function (k, p) { return I18N.t(k, p); }, esc = App.esc, P = App.params;
  var M = window.MOCK || {};
  var PROFILE = 'profile.html';
  var EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/;
  var timers = [];

  /* ---------- tabs: initial selection from ?tab= (must precede App.Tabs.init at DOMContentLoaded) ---------- */
  var tabs = $$('#auth-tablist [role="tab"], [data-tabs] [role="tab"]');
  function preselect(i) { tabs.forEach(function (b, j) { b.setAttribute('aria-selected', j === i); b.tabIndex = j === i ? 0 : -1; var p = document.getElementById(b.getAttribute('aria-controls')); if (p) p.hidden = j !== i; }); }
  var wantSignup = P.get('tab') === 'signup';
  preselect(wantSignup ? 1 : 0);
  function goTab(i) { var b = tabs[i]; if (b) { b.focus(); b.click(); } }
  document.addEventListener('click', function (e) { var b = e.target.closest('[data-auth-tab]'); if (b) goTab(+b.getAttribute('data-auth-tab')); });

  /* ---------- country options (list rebuild only, never the form) ---------- */
  function renderCountries() {
    var sel = $('#auth-up-country'), cur = sel.value, L = I18N.lang;
    sel.innerHTML = '<option value="" data-testid="auth-signup-country-option-placeholder">' + esc(t('auth.country.choose')) + '</option>' + (M.countries || []).map(function (c) { return '<option value="' + c.code + '" data-testid="auth-signup-country-option-' + c.code.toLowerCase() + '">' + esc(c[L] || c.fr) + '</option>'; }).join('');
    sel.value = cur;
  }
  renderCountries();
  document.addEventListener('i18n:change', function () { renderCountries(); document.title = t('auth.title'); });
  document.title = t('auth.title');

  /* ---------- validation helpers ---------- */
  function setErr(pre, form, field, key) {
    var box = $('#' + pre + '-f-' + field), msg = $('#' + pre + '-' + field + '-error'), txt = $('span:not(.icon)', msg), ctl = $('#' + pre + '-' + (field === 'langs' ? 'langs' : field));
    if (key) { txt.setAttribute('data-i18n', key); txt.textContent = t(key); msg.hidden = false; box.classList.add('has-error'); if (field !== 'langs') ctl.setAttribute('aria-invalid', 'true'); }
    else { txt.removeAttribute('data-i18n'); txt.textContent = ''; msg.hidden = true; box.classList.remove('has-error'); if (field !== 'langs') ctl.removeAttribute('aria-invalid'); }
  }
  function summarise(pre, form, items) {
    var box = $('#' + pre + '-summary'), ul = $('#' + pre + '-summary-list');
    if (!items.length) { box.hidden = true; ul.innerHTML = ''; return; }
    ul.innerHTML = items.map(function (it, i) { return '<li data-testid="' + form + '-summary-item-' + (i + 1) + '"><a href="#' + it.target + '" data-field-link="' + it.target + '" data-i18n="' + it.key + '" data-testid="' + form + '-summary-link-' + (i + 1) + '"></a></li>'; }).join('');
    I18N.apply(ul); box.hidden = false; box.focus();
  }
  document.addEventListener('click', function (e) { var a = e.target.closest('[data-field-link]'); if (!a) return; e.preventDefault(); var c = document.getElementById(a.getAttribute('data-field-link')); if (c) c.focus(); });
  function checkEmail(v) { return !v ? 'auth.err.email.required' : !EMAIL_RE.test(v) ? 'auth.err.email.format' : null; }
  function run(pre, form, checks) {
    var items = [];
    checks.forEach(function (c) { var k = c.fn(); setErr(pre, form, c.field, k); if (k) items.push({ key: k, target: c.target }); });
    summarise(pre, form, items); return !items.length;
  }
  function busy(btn, on, labelEl, key) {
    btn.disabled = on; btn.setAttribute('aria-busy', on); var sp = $('.spinner', btn); if (sp) sp.hidden = !on;
    labelEl.setAttribute('data-i18n', key); labelEl.textContent = t(key);
  }
  function later(fn, ms) { var id = setTimeout(fn, ms); timers.push(id); return id; }

  /* ---------- sign-in banner (live region content) ---------- */
  var banner = $('#auth-signin-banner'), limitTimer = null, limitLeft = 0;
  function showBanner(kind) {
    var cfg = {
      bad: { cls: 'callout-danger', icon: 'alert-circle', role: 'alert' },
      limit: { cls: 'callout-warn', icon: 'clock', role: 'alert' },
      expired: { cls: '', icon: 'info', role: 'status' },
      verify: { cls: '', icon: 'inbox', role: 'status' }
    }[kind];
    var body = kind === 'limit' ? '<span id="auth-limit-body" data-testid="auth-banner-limit-body"></span>' : '<span data-i18n="auth.banner.' + kind + '.body" data-testid="auth-banner-' + kind + '-body"></span>';
    var action = kind === 'verify' ? '<p><button type="button" class="btn btn-sm" id="auth-banner-resend" data-testid="auth-banner-verify-resend"><span class="icon i-send" aria-hidden="true" data-testid="auth-banner-verify-resend-icon"></span><span data-i18n="auth.banner.verify.resend" data-testid="auth-banner-verify-resend-label"></span></button></p>' : '';
    banner.innerHTML = '<div class="callout ' + cfg.cls + '" role="' + cfg.role + '" data-testid="auth-banner-' + kind + '"><span class="icon i-' + cfg.icon + '" aria-hidden="true" data-testid="auth-banner-' + kind + '-icon"></span><div><strong data-i18n="auth.banner.' + kind + '.title" data-testid="auth-banner-' + kind + '-title"></strong>' + body + action + '</div></div>';
    I18N.apply(banner); if (kind === 'limit') paintLimit();
  }
  function clearBanner() { banner.innerHTML = ''; }
  function paintLimit() {
    var b = $('#auth-limit-body'); if (b) b.textContent = limitLeft > 0 ? t('auth.banner.limit.body', { n: limitLeft }) : t('auth.banner.limit.body0');
    var btn = $('#auth-signin-submit'), lab = $('#auth-signin-submit-label');
    if (limitLeft > 0) { btn.disabled = true; btn.setAttribute('aria-disabled', 'true'); lab.removeAttribute('data-i18n'); lab.textContent = t('auth.signin.limited', { n: limitLeft }); }
    else { btn.disabled = false; btn.removeAttribute('aria-disabled'); lab.setAttribute('data-i18n', 'auth.signin.submit'); lab.textContent = t('auth.signin.submit'); }
  }
  function startLimit(sec) {
    limitLeft = sec; showBanner('limit'); clearInterval(limitTimer);
    limitTimer = setInterval(function () { limitLeft--; if (limitLeft <= 0) { clearInterval(limitTimer); limitLeft = 0; } paintLimit(); }, 1000);
  }
  document.addEventListener('i18n:change', function () { if ($('#auth-limit-body')) paintLimit(); });
  function resend() { App.Toast.show({ kind: 'info', title: t('auth.resend.toast.title'), msg: t('auth.resend.toast.msg') }); }
  document.addEventListener('click', function (e) { if (e.target.closest('#auth-banner-resend, #auth-verify-resend')) resend(); });

  /* ---------- sign-in ---------- */
  var inForm = $('#auth-signin-form');
  function success(host) {
    App.Toast.show({ kind: 'success', title: t('auth.ok.title'), msg: t('auth.ok.body') });
    banner.innerHTML = '<div class="callout callout-success" role="status" data-testid="auth-banner-success"><span class="icon i-check-circle" aria-hidden="true" data-testid="auth-banner-success-icon"></span><div><strong data-i18n="auth.ok.title" data-testid="auth-banner-success-title"></strong><span data-i18n="auth.ok.body" data-testid="auth-banner-success-body"></span> <a href="' + PROFILE + '" data-i18n="auth.ok.link" data-testid="auth-banner-success-link"></a></div></div>';
    I18N.apply(banner); later(function () { location.href = PROFILE; }, 1400);
  }
  inForm.addEventListener('submit', function (e) {
    e.preventDefault(); clearBanner();
    var email = $('#auth-in-email').value.trim(), pw = $('#auth-in-password').value;
    var ok = run('auth-in', 'auth-signin', [
      { field: 'email', target: 'auth-in-email', fn: function () { return checkEmail(email); } },
      { field: 'password', target: 'auth-in-password', fn: function () { return pw ? null : 'auth.err.password.required'; } }
    ]);
    if (!ok) return;
    var btn = $('#auth-signin-submit'), lab = $('#auth-signin-submit-label');
    busy(btn, true, lab, 'auth.signin.busy');
    later(function () {
      busy(btn, false, lab, 'auth.signin.submit');
      if (email.toLowerCase() === 'limit@example.org') { startLimit(15); return; }
      if (pw.length < 12) { showBanner('bad'); $('#auth-in-password').value = ''; $('#auth-in-password').focus(); return; }
      success();
    }, 700);
  });

  /* ---------- sign-up ---------- */
  var upForm = $('#auth-signup-form');
  upForm.addEventListener('submit', function (e) {
    e.preventDefault();
    var name = $('#auth-up-name').value.trim(), email = $('#auth-up-email').value.trim(), pw = $('#auth-up-password').value, country = $('#auth-up-country').value, langs = $$('input[name="langs"]:checked', upForm).length;
    var ok = run('auth-up', 'auth-signup', [
      { field: 'name', target: 'auth-up-name', fn: function () { return !name ? 'auth.err.name.required' : name.length < 2 ? 'auth.err.name.short' : null; } },
      { field: 'email', target: 'auth-up-email', fn: function () { return checkEmail(email); } },
      { field: 'password', target: 'auth-up-password', fn: function () { return !pw ? 'auth.err.password.required' : pw.length < 12 ? 'auth.err.password.short' : null; } }
    ]);
    if (!ok) return;
    var btn = $('#auth-signup-submit'), lab = $('#auth-signup-submit-label');
    busy(btn, true, lab, 'auth.signup.busy');
    later(function () {   /* identical outcome for new and existing addresses (no enumeration) */
      busy(btn, false, lab, 'auth.signup.submit');
      $('#auth-signup-form-box').hidden = true; var v = $('#auth-verify-panel'); v.hidden = false; v.focus();
    }, 700);
  });
  $('#auth-verify-confirm').addEventListener('click', function () { App.Toast.show({ kind: 'success', title: t('auth.ok.title'), msg: t('auth.ok.body') }); later(function () { location.href = PROFILE; }, 900); });
  $('[data-testid="auth-verify-back"]').addEventListener('click', function () { $('#auth-verify-panel').hidden = true; $('#auth-signup-form-box').hidden = false; });

  /* ---------- forgot password (dialog): generic response ---------- */
  var fgForm = $('#auth-forgot-form');
  document.addEventListener('click', function (e) {
    if (!e.target.closest('[data-dialog-open="auth-forgot-dialog"]')) return;
    fgForm.reset(); $('#auth-forgot-fields').hidden = false; $('#auth-forgot-done').hidden = true; $('#auth-forgot-submit').hidden = false; setErr('auth-fg', 'auth-forgot', 'email', null);
    var v = $('#auth-in-email').value.trim(); if (v) $('#auth-fg-email').value = v;
  });
  fgForm.addEventListener('submit', function (e) {
    e.preventDefault(); var k = checkEmail($('#auth-fg-email').value.trim()); setErr('auth-fg', 'auth-forgot', 'email', k);
    if (k) { $('#auth-fg-email').focus(); return; }
    $('#auth-forgot-fields').hidden = true; $('#auth-forgot-submit').hidden = true; var d = $('#auth-forgot-done'); d.hidden = false; d.focus();
  });

  /* ---------- forced / arriving states ---------- */
  var st = P.get('state');
  if (st === 'expired') showBanner('expired');
  else if (st === 'badcreds') showBanner('bad');
  else if (st === 'ratelimit') startLimit(30);
  else if (st === 'verify') showBanner('verify');
})();
