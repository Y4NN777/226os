/* Standalone account pages: connexion, inscription, verifier-email, mot-de-passe-oublie, reinitialiser. 100% SIMULATED.
   Security-faithful UX (G0-AUTH): generic sign-in error, neutral sign-up / reset answers (no enumeration), 429 state,
   expired session, expired/invalid reset link. Passwords are never stored. Only {name,email} survives in sessionStorage for the verify step.
   Demo scenarios: password < 12 chars = generic error; limit@example.org = 429; exists@example.org = same answer as a new address.
   Forced states: connexion ?state=expired|ratelimit|badcreds ; reinitialiser ?state=invalid|expired. */
(function () {
  'use strict';
  var nb = ' ';
  I18N.register({
    fr: {
      'acc.email': 'Adresse e-mail', 'acc.password': 'Mot de passe', 'acc.name': 'Nom', 'acc.busy': 'Un instant…',
      'acc.login.title': 'Connexion — 226os', 'acc.login.h1': 'Content de vous revoir', 'acc.login.lede': 'Connectez-vous pour poser, répondre et suivre vos projets.',
      'acc.login.forgot': 'Mot de passe oublié' + nb + '?', 'acc.login.submit': 'Se connecter', 'acc.login.busy': 'Connexion…',
      'acc.login.noaccount': 'Pas encore de compte' + nb + '?', 'acc.login.signup': 'Créer un compte',
      'acc.proto.login': 'Prototype : un mot de passe de 12 caractères ou plus fonctionne ; plus court, c’est l’erreur générique ; limit@example.org simule le 429. Forcer : ?state=expired, ratelimit, badcreds.',
      'acc.err.email.required': 'Saisissez votre adresse e-mail.', 'acc.err.email.format': 'Cette adresse semble incomplète. Exemple : nom@domaine.org.',
      'acc.err.password.required': 'Saisissez votre mot de passe.', 'acc.err.password.short': 'Le mot de passe doit contenir au moins 12 caractères.',
      'acc.err.name.required': 'Indiquez votre nom.', 'acc.err.name.short': 'Votre nom doit contenir au moins 2 caractères.',
      'acc.banner.bad.title': 'Connexion impossible', 'acc.banner.bad.body': 'E-mail ou mot de passe incorrect.',
      'acc.banner.limit.title': 'Trop de tentatives', 'acc.banner.limit.body': 'Réessayez dans {n} secondes.', 'acc.banner.limit.body0': 'Vous pouvez réessayer maintenant.',
      'acc.banner.expired.title': 'Session expirée', 'acc.banner.expired.body': 'Votre session a expiré. Reconnectez-vous pour continuer là où vous en étiez.',
      'acc.login.limited': 'Réessayer dans {n} s',
      'acc.signup.title': 'Créer un compte — 226os', 'acc.signup.h1': 'Créer votre compte', 'acc.signup.lede': 'Trois champs, aucune fiche à remplir. Vous complétez le reste plus tard, quand cela vous sert.',
      'acc.signup.submit': 'Créer mon compte', 'acc.signup.busy': 'Création…', 'acc.signup.hasaccount': 'Vous avez déjà un compte' + nb + '?', 'acc.signup.login': 'Se connecter',
      'acc.signup.legal': 'En créant un compte, vous acceptez la <a href="{privacy}" rel="noopener noreferrer" data-testid="signup-legal-privacy">politique de confidentialité</a> et le <a href="{conduct}" rel="noopener noreferrer" data-testid="signup-legal-conduct">code de conduite</a>.',
      'acc.rule.len': '12 caractères minimum', 'acc.rule.tip': 'Une phrase de passe convient très bien. Espaces et accents autorisés.', 'acc.rule.met': '(respecté)', 'acc.rule.unmet': '(à respecter)',
      'acc.verify.title': 'Vérifiez votre boîte mail — 226os', 'acc.verify.h1': 'Vérifiez votre boîte mail', 'acc.verify.lede': 'Un lien de confirmation vous attend à',
      'acc.verify.neutral': 'Pensez à regarder dans les courriers indésirables. Le lien est valable une heure.',
      'acc.verify.resend': 'Renvoyer l’e-mail', 'acc.verify.resendWait': 'Renvoyer dans {n} s', 'acc.verify.sent': 'Si l’adresse est valide, un nouvel e-mail vient de partir.',
      'acc.verify.change': 'Modifier l’adresse', 'acc.verify.back': 'Retour à la connexion',
      'acc.proto.verify': 'Prototype : aucun e-mail n’est envoyé.', 'acc.verify.simulate': 'Simuler le clic sur le lien',
      'acc.forgot.title': 'Mot de passe oublié — 226os', 'acc.forgot.h1': 'Mot de passe oublié', 'acc.forgot.lede': 'Saisissez l’adresse de votre compte. Nous vous envoyons un lien pour en choisir un nouveau.',
      'acc.forgot.submit': 'Envoyer le lien', 'acc.forgot.back': 'Retour à la connexion',
      'acc.forgot.done.h1': 'Consultez votre boîte mail', 'acc.forgot.done.body': 'Si un compte existe pour cette adresse, un lien de réinitialisation vient de partir. Il est à usage unique et expire dans une heure.',
      'acc.proto.forgot': 'Prototype : la réponse est la même quelle que soit l’adresse.', 'acc.forgot.simulate': 'Simuler le clic sur le lien',
      'acc.reset.title': 'Nouveau mot de passe — 226os', 'acc.reset.h1': 'Choisir un nouveau mot de passe', 'acc.reset.lede': 'Ce mot de passe remplace l’ancien sur tous vos appareils.',
      'acc.reset.newpw': 'Nouveau mot de passe', 'acc.reset.submit': 'Enregistrer', 'acc.reset.busy': 'Enregistrement…',
      'acc.reset.invalid.h1': 'Ce lien n’est pas valide', 'acc.reset.invalid.body': 'Il a peut-être déjà servi, ou il est incomplet. Demandez-en un nouveau.',
      'acc.reset.expired.h1': 'Ce lien a expiré', 'acc.reset.expired.body': 'Par sécurité, un lien n’est valable qu’une heure. Demandez-en un nouveau.',
      'acc.reset.retry': 'Demander un nouveau lien', 'acc.reset.done.h1': 'Mot de passe modifié', 'acc.reset.done.body': 'Vous pouvez vous connecter avec votre nouveau mot de passe.'
    },
    en: {
      'acc.email': 'Email address', 'acc.password': 'Password', 'acc.name': 'Name', 'acc.busy': 'One moment…',
      'acc.login.title': 'Sign in — 226os', 'acc.login.h1': 'Welcome back', 'acc.login.lede': 'Sign in to ask, answer and follow your projects.',
      'acc.login.forgot': 'Forgot password?', 'acc.login.submit': 'Sign in', 'acc.login.busy': 'Signing in…',
      'acc.login.noaccount': 'No account yet?', 'acc.login.signup': 'Create an account',
      'acc.proto.login': 'Prototype: a password of 12+ characters works; shorter gives the generic error; limit@example.org simulates 429. Force: ?state=expired, ratelimit, badcreds.',
      'acc.err.email.required': 'Enter your email address.', 'acc.err.email.format': 'That address looks incomplete. Example: name@domain.org.',
      'acc.err.password.required': 'Enter your password.', 'acc.err.password.short': 'The password must be at least 12 characters long.',
      'acc.err.name.required': 'Enter your name.', 'acc.err.name.short': 'Your name must be at least 2 characters long.',
      'acc.banner.bad.title': 'Cannot sign in', 'acc.banner.bad.body': 'Email or password incorrect.',
      'acc.banner.limit.title': 'Too many attempts', 'acc.banner.limit.body': 'Try again in {n} seconds.', 'acc.banner.limit.body0': 'You can try again now.',
      'acc.banner.expired.title': 'Session expired', 'acc.banner.expired.body': 'Your session has expired. Sign in again to pick up where you left off.',
      'acc.login.limited': 'Try again in {n}s',
      'acc.signup.title': 'Create an account — 226os', 'acc.signup.h1': 'Create your account', 'acc.signup.lede': 'Three fields, no profile form. You fill in the rest later, when it helps you.',
      'acc.signup.submit': 'Create my account', 'acc.signup.busy': 'Creating…', 'acc.signup.hasaccount': 'Already have an account?', 'acc.signup.login': 'Sign in',
      'acc.signup.legal': 'By creating an account you accept the <a href="{privacy}" rel="noopener noreferrer" data-testid="signup-legal-privacy">privacy policy</a> and the <a href="{conduct}" rel="noopener noreferrer" data-testid="signup-legal-conduct">code of conduct</a>.',
      'acc.rule.len': 'At least 12 characters', 'acc.rule.tip': 'A passphrase works very well. Spaces and accents allowed.', 'acc.rule.met': '(met)', 'acc.rule.unmet': '(not met yet)',
      'acc.verify.title': 'Check your inbox — 226os', 'acc.verify.h1': 'Check your inbox', 'acc.verify.lede': 'A confirmation link is waiting at',
      'acc.verify.neutral': 'Look in your spam folder too. The link is valid for one hour.',
      'acc.verify.resend': 'Resend email', 'acc.verify.resendWait': 'Resend in {n}s', 'acc.verify.sent': 'If the address is valid, a new email has just been sent.',
      'acc.verify.change': 'Change address', 'acc.verify.back': 'Back to sign in',
      'acc.proto.verify': 'Prototype: no email is sent.', 'acc.verify.simulate': 'Simulate clicking the link',
      'acc.forgot.title': 'Forgot password — 226os', 'acc.forgot.h1': 'Forgot your password?', 'acc.forgot.lede': 'Enter your account address. We send you a link to choose a new one.',
      'acc.forgot.submit': 'Send link', 'acc.forgot.back': 'Back to sign in',
      'acc.forgot.done.h1': 'Check your inbox', 'acc.forgot.done.body': 'If an account exists for this address, a reset link has just been sent. It is single-use and expires in one hour.',
      'acc.proto.forgot': 'Prototype: the answer is identical whatever the address.', 'acc.forgot.simulate': 'Simulate clicking the link',
      'acc.reset.title': 'New password — 226os', 'acc.reset.h1': 'Choose a new password', 'acc.reset.lede': 'This password replaces the old one on all your devices.',
      'acc.reset.newpw': 'New password', 'acc.reset.submit': 'Save', 'acc.reset.busy': 'Saving…',
      'acc.reset.invalid.h1': 'This link is not valid', 'acc.reset.invalid.body': 'It may already have been used, or it is incomplete. Request a new one.',
      'acc.reset.expired.h1': 'This link has expired', 'acc.reset.expired.body': 'For your security a link is only valid for one hour. Request a new one.',
      'acc.reset.retry': 'Request a new link', 'acc.reset.done.h1': 'Password changed', 'acc.reset.done.body': 'You can now sign in with your new password.'
    }
  });

  var $ = App.$, $$ = App.$$, t = function (k, p) { return I18N.t(k, p); }, P = App.params, S = App.Session, CFG = App.config;
  var EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/, PKEY = '226os.pending', timers = [];
  function later(fn, ms) { var id = setTimeout(fn, ms); timers.push(id); return id; }
  function getPending() { try { return JSON.parse(sessionStorage.getItem(PKEY)) || {}; } catch (e) { return {}; } }
  function setPending(o) { try { sessionStorage.setItem(PKEY, JSON.stringify(o)); } catch (e) {} }
  function slug(s) { return String(s || '').normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '').slice(0, 24) || 'membre'; }
  function titleCase(s) { return s.replace(/[._-]+/g, ' ').replace(/\b\w/g, function (c) { return c.toUpperCase(); }).trim(); }
  function withNext(href) { var n = P.get('next'); return n ? href + '?next=' + encodeURIComponent(n) : href; }
  function setTitle(k) { document.title = t(k); }

  /* ---------- field helpers ---------- */
  function setErr(pre, field, key) {
    var box = $('#' + pre + '-f-' + field), msg = $('#' + pre + '-' + field + '-error'), txt = $('span:not(.icon)', msg), ctl = $('#' + pre + '-' + field);
    if (key) { txt.setAttribute('data-i18n', key); txt.textContent = t(key); msg.hidden = false; box.classList.add('has-error'); ctl.setAttribute('aria-invalid', 'true'); }
    else { txt.removeAttribute('data-i18n'); txt.textContent = ''; msg.hidden = true; box.classList.remove('has-error'); ctl.removeAttribute('aria-invalid'); }
  }
  function validate(pre, checks) {
    var first = null;
    checks.forEach(function (c) { var k = c.fn(); setErr(pre, c.field, k); if (k && !first) first = $('#' + pre + '-' + c.field); });
    if (first) first.focus(); return !first;
  }
  function checkEmail(v) { return !v ? 'acc.err.email.required' : !EMAIL_RE.test(v) ? 'acc.err.email.format' : null; }
  function checkNew(v) { return !v ? 'acc.err.password.required' : v.length < 12 ? 'acc.err.password.short' : null; }
  function busy(pre, on, key) {
    var b = $('#' + pre + '-submit'), l = $('#' + pre + '-submit-label'), sp = $('.spinner', b);
    b.disabled = on; b.setAttribute('aria-busy', on); sp.hidden = !on; l.setAttribute('data-i18n', key); l.textContent = t(key);
  }
  function clearOnInput(pre, field) { var c = $('#' + pre + '-' + field); if (c) c.addEventListener('input', function () { if (c.getAttribute('aria-invalid')) setErr(pre, field, null); }); }

  /* live password rule (shown before typing; length only, no composition rules: OWASP/NIST) */
  function bindRules(pre) {
    var pw = $('#' + pre + '-password'), li = $('#' + pre + '-rule-len'); if (!pw || !li) return;
    function paint() {
      var met = pw.value.length >= 12; li.setAttribute('data-met', met);
      var ic = $('.icon', li); ic.className = 'icon ' + (met ? 'i-check-circle' : 'i-circle');
      $('.visually-hidden', li).textContent = t(met ? 'acc.rule.met' : 'acc.rule.unmet');
    }
    pw.addEventListener('input', paint); document.addEventListener('i18n:change', paint); paint();
  }

  /* ================= connexion ================= */
  var loginForm = $('#login-form');
  if (loginForm) {
    var banner = $('#login-banner'), limitLeft = 0, limitTimer = null;
    var KINDS = { bad: { cls: 'callout-danger', icon: 'alert-circle', role: 'alert' }, limit: { cls: 'callout-warn', icon: 'clock', role: 'alert' }, expired: { cls: '', icon: 'info', role: 'status' } };
    var showBanner = function (kind) {
      var c = KINDS[kind], body = kind === 'limit' ? '<span id="login-limit-body" data-testid="login-banner-limit-body"></span>' : '<span data-i18n="acc.banner.' + kind + '.body" data-testid="login-banner-' + kind + '-body"></span>';
      banner.innerHTML = '<div class="callout ' + c.cls + '" role="' + c.role + '" data-testid="login-banner-' + kind + '"><span class="icon i-' + c.icon + '" aria-hidden="true" data-testid="login-banner-' + kind + '-icon"></span><div><strong data-i18n="acc.banner.' + kind + '.title" data-testid="login-banner-' + kind + '-title"></strong>' + body + '</div></div>';
      I18N.apply(banner); if (kind === 'limit') paintLimit();
    };
    var paintLimit = function () {
      var b = $('#login-limit-body'); if (b) b.textContent = limitLeft > 0 ? t('acc.banner.limit.body', { n: limitLeft }) : t('acc.banner.limit.body0');
      var btn = $('#login-submit'), lab = $('#login-submit-label');
      if (limitLeft > 0) { btn.disabled = true; lab.removeAttribute('data-i18n'); lab.textContent = t('acc.login.limited', { n: limitLeft }); }
      else { btn.disabled = false; lab.setAttribute('data-i18n', 'acc.login.submit'); lab.textContent = t('acc.login.submit'); }
    };
    var startLimit = function (sec) { limitLeft = sec; showBanner('limit'); clearInterval(limitTimer); limitTimer = setInterval(function () { limitLeft--; if (limitLeft <= 0) { clearInterval(limitTimer); limitLeft = 0; } paintLimit(); }, 1000); };
    document.addEventListener('i18n:change', function () { if ($('#login-limit-body')) paintLimit(); });
    $('#login-signup-link').href = withNext('inscription.html'); $('#login-forgot').href = 'mot-de-passe-oublie.html';
    clearOnInput('login', 'email'); clearOnInput('login', 'password');
    loginForm.addEventListener('submit', function (e) {
      e.preventDefault(); banner.innerHTML = '';
      var email = $('#login-email').value.trim(), pw = $('#login-password').value;
      if (!validate('login', [{ field: 'email', fn: function () { return checkEmail(email); } }, { field: 'password', fn: function () { return pw ? null : 'acc.err.password.required'; } }])) return;
      busy('login', true, 'acc.login.busy');
      later(function () {
        busy('login', false, 'acc.login.submit');
        if (email.toLowerCase() === 'limit@example.org') { startLimit(15); return; }
        if (pw.length < 12) { showBanner('bad'); var p = $('#login-password'); p.value = ''; p.focus(); return; }
        var known = S.state.user && S.state.user.name;
        S.signIn(known ? {} : { name: titleCase(email.split('@')[0]), username: slug(email.split('@')[0]) });
        S.returnTo(S.state.onboarding.done ? 'pages/accueil.html' : 'pages/bienvenue.html');
      }, 700);
    });
    var st = P.get('state');
    if (st === 'expired') showBanner('expired'); else if (st === 'badcreds') showBanner('bad'); else if (st === 'ratelimit') startLimit(30);
    setTitle('acc.login.title');
  }

  /* ================= inscription ================= */
  var signupForm = $('#signup-form');
  if (signupForm) {
    var legal = function () { $('#signup-legal').innerHTML = t('acc.signup.legal', { privacy: CFG.legal.privacy, conduct: CFG.legal.conduct }); };
    legal(); document.addEventListener('i18n:change', legal);
    $('#signup-login-link').href = withNext('connexion.html');
    ['name', 'email', 'password'].forEach(function (f) { clearOnInput('signup', f); }); bindRules('signup');
    signupForm.addEventListener('submit', function (e) {
      e.preventDefault();
      var name = $('#signup-name').value.trim(), email = $('#signup-email').value.trim(), pw = $('#signup-password').value;
      if (!validate('signup', [
        { field: 'name', fn: function () { return !name ? 'acc.err.name.required' : name.length < 2 ? 'acc.err.name.short' : null; } },
        { field: 'email', fn: function () { return checkEmail(email); } },
        { field: 'password', fn: function () { return checkNew(pw); } }])) return;
      busy('signup', true, 'acc.signup.busy');
      later(function () {   /* identical outcome for new and existing addresses (no enumeration) */
        setPending({ name: name, email: email, next: P.get('next') || '' });
        location.href = 'verifier-email.html';
      }, 700);
    });
    setTitle('acc.signup.title');
  }

  /* ================= verifier-email ================= */
  var resend = $('#verify-resend');
  if (resend) {
    var pend = getPending(), left = 0, vt = null;
    var mask = function (m) {
      if (!m || m.indexOf('@') < 1) return 'v••••@exemple.org';
      var a = m.split('@'), d = a[1], i = d.lastIndexOf('.');
      return a[0].charAt(0) + '•••@' + d.charAt(0) + '•••' + (i > 0 ? d.slice(i) : '');
    };
    $('#verify-mail').textContent = mask(pend.email);
    var paint = function () {
      var l = $('#verify-resend-label');
      if (left > 0) { resend.disabled = true; l.textContent = t('acc.verify.resendWait', { n: left }); }
      else { resend.disabled = false; l.textContent = t('acc.verify.resend'); }
    };
    var tick = function () { clearInterval(vt); left = 30; paint(); vt = setInterval(function () { left--; if (left <= 0) { clearInterval(vt); left = 0; } paint(); }, 1000); };
    document.addEventListener('i18n:change', paint); paint();
    resend.addEventListener('click', function () {
      $('#verify-status').textContent = ''; later(function () { $('#verify-status').textContent = t('acc.verify.sent'); }, 30);
      App.Toast.show({ kind: 'info', msg: t('acc.verify.sent') }); tick();
    });
    $('#verify-simulate').addEventListener('click', function () {
      var nm = pend.name || 'Nouveau membre';
      S.signIn({ name: nm, username: slug(nm) });
      var dest = 'bienvenue.html'; if (pend.next) dest += '?next=' + encodeURIComponent(pend.next);
      location.href = dest;
    });
    setTitle('acc.verify.title');
  }

  /* ================= mot-de-passe-oublie ================= */
  var forgotForm = $('#forgot-form');
  if (forgotForm) {
    clearOnInput('forgot', 'email');
    forgotForm.addEventListener('submit', function (e) {
      e.preventDefault(); var email = $('#forgot-email').value.trim();
      if (!validate('forgot', [{ field: 'email', fn: function () { return checkEmail(email); } }])) return;
      busy('forgot', true, 'acc.busy');
      later(function () { $('#forgot-ask').hidden = true; var d = $('#forgot-done'); d.hidden = false; d.focus(); }, 600);   /* same answer for every address */
    });
    setTitle('acc.forgot.title');
  }

  /* ================= reinitialiser ================= */
  var resetForm = $('#reset-form');
  if (resetForm) {
    var rs = P.get('state');
    if (rs === 'invalid' || rs === 'expired') {
      $('#reset-ask').hidden = true; var iv = $('#reset-invalid'); iv.hidden = false;
      $('#reset-invalid-h1').setAttribute('data-i18n', 'acc.reset.' + rs + '.h1'); $('#reset-invalid-body').setAttribute('data-i18n', 'acc.reset.' + rs + '.body');
      $('#reset-invalid-icon').className = 'icon ' + (rs === 'expired' ? 'i-clock' : 'i-alert-circle'); I18N.apply(iv);
    }
    clearOnInput('reset', 'password'); bindRules('reset');
    resetForm.addEventListener('submit', function (e) {
      e.preventDefault(); var pw = $('#reset-password').value;
      if (!validate('reset', [{ field: 'password', fn: function () { return checkNew(pw); } }])) return;
      busy('reset', true, 'acc.reset.busy');
      later(function () { $('#reset-ask').hidden = true; var d = $('#reset-done'); d.hidden = false; d.focus(); }, 700);
    });
    setTitle('acc.reset.title');
  }
  document.addEventListener('i18n:change', function () { var k = $('title').getAttribute('data-i18n'); if (k) document.title = t(k); });
})();
