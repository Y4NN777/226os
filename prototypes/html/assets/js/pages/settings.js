/* Paramètres (v2): Profil, Compte, Préférences, Confidentialité. One section at a time (hash driven), save per section with inline success.
   Profile fields write to App.Session so the own-profile checklist stays consistent (GitHub stored as bare username, languages / collaboration confirmed on save).
   Forms are built once and never rebuilt on i18n:change (texts use data-i18n). Namespace: set.* */
I18N.register({
  fr: {
    'set.title': 'Paramètres — 226os', 'set.h1': 'Paramètres', 'set.nav': 'Sections des paramètres',
    'set.nav.profil': 'Profil', 'set.nav.compte': 'Compte', 'set.nav.preferences': 'Préférences', 'set.nav.confidentialite': 'Confidentialité',
    'set.save': 'Enregistrer', 'set.saved': 'Modifications enregistrées.',
    'set.p.h': 'Profil', 'set.p.lede': 'Ce que les autres membres voient sur votre page.',
    'set.p.avatar': 'Votre avatar est généré à partir de vos initiales.',
    'set.p.name': 'Nom', 'set.p.name.hint': 'Votre nom complet ou celui sous lequel on vous connaît.',
    'set.p.username': 'Nom d’utilisateur', 'set.p.username.hint': 'Lettres minuscules, chiffres, tirets ou tirets bas (3 à 30 caractères).',
    'set.p.bio': 'Présentation', 'set.p.bio.hint': 'Deux phrases suffisent pour inspirer confiance.',
    'set.p.skills': 'Compétences', 'set.p.skills.hint': 'Jusqu’à 8. Elles composent votre accueil et vos recommandations.', 'set.p.skills.add': 'Ajouter une compétence', 'set.p.skills.addBtn': 'Ajouter', 'set.p.skills.suggest': 'Suggestions', 'set.p.skills.remove': 'Retirer {s}', 'set.p.skills.selected': 'Compétences choisies',
    'set.p.country': 'Pays', 'set.p.country.hint': 'Les API qui couvrent votre marché passent en premier.', 'set.p.country.none': 'Choisir un pays',
    'set.p.langs': 'Langues de travail', 'set.p.langs.hint': 'Pour lire d’abord les contenus qui vous conviennent.',
    'set.p.github': 'GitHub', 'set.p.github.hint': 'Nom d’utilisateur ou adresse. Les mainteneurs regardent votre GitHub en premier.', 'set.p.github.ph': 'github.com/votre-nom',
    'set.p.collab': 'Ouvert aux collaborations', 'set.p.collab.hint': 'Les mainteneurs qui cherchent de l’aide peuvent vous trouver.',
    'set.e.required': 'Ce champ est obligatoire.', 'set.e.username': 'Utilisez 3 à 30 caractères : minuscules, chiffres, - ou _.', 'set.e.skills': 'Choisissez au moins une compétence.', 'set.e.langs': 'Choisissez au moins une langue.',
    'set.e.github': 'Saisissez un nom d’utilisateur ou une adresse GitHub valide.', 'set.e.email': 'Saisissez une adresse e-mail valide.',
    'set.e.pwCurrent': 'Saisissez votre mot de passe actuel.', 'set.e.pwNew': 'Utilisez au moins 12 caractères.', 'set.e.pwSame': 'Les deux mots de passe ne correspondent pas.',
    'set.c.h': 'Compte', 'set.c.lede': 'Votre connexion et vos appareils.',
    'set.c.email': 'Adresse e-mail', 'set.c.email.hint': 'Utilisée pour vous connecter. Elle n’est jamais publique.', 'set.c.email.save': 'Mettre à jour l’e-mail', 'set.c.email.saved': 'Adresse e-mail mise à jour.',
    'set.c.pw': 'Mot de passe', 'set.c.pw.current': 'Mot de passe actuel', 'set.c.pw.new': 'Nouveau mot de passe', 'set.c.pw.new.hint': '12 caractères minimum. Une phrase longue est plus sûre qu’un mot compliqué.', 'set.c.pw.confirm': 'Confirmer le nouveau mot de passe', 'set.c.pw.save': 'Changer le mot de passe', 'set.c.pw.saved': 'Mot de passe modifié.',
    'set.c.sessions': 'Appareils connectés', 'set.c.sessions.lede': 'Déconnectez ceux que vous ne reconnaissez pas.', 'set.c.sessions.this': 'Cet appareil', 'set.c.sessions.now': 'Active maintenant', 'set.c.sessions.last': 'Dernière activité : {date}',
    'set.c.sessions.out': 'Déconnecter {d}', 'set.c.sessions.all': 'Déconnecter les autres appareils', 'set.c.sessions.done': 'Appareil déconnecté.', 'set.c.sessions.alldone': 'Les autres appareils sont déconnectés.', 'set.c.sessions.none': 'Aucun autre appareil connecté.',
    'set.f.h': 'Préférences', 'set.f.lede': 'S’appliquent immédiatement.',
    'set.f.lang': 'Langue de l’interface', 'set.f.lang.hint': 'Remplace la langue détectée automatiquement par votre navigateur. Vos contenus gardent leur langue d’origine.',
    'set.f.theme': 'Thème', 'set.f.theme.hint': 'Clair, sombre, ou celui de votre système.', 'set.f.theme.light': 'Clair', 'set.f.theme.dark': 'Sombre', 'set.f.theme.system': 'Système', 'set.f.saved': 'Préférence enregistrée.',
    'set.v.h': 'Confidentialité', 'set.v.lede': 'Choisissez qui voit chaque information de votre profil.',
    'set.v.public': 'Tout le monde', 'set.v.members': 'Membres connectés', 'set.v.self': 'Moi seul',
    'set.v.name': 'Nom', 'set.v.username': 'Nom d’utilisateur', 'set.v.email': 'Adresse e-mail', 'set.v.country': 'Pays', 'set.v.skills': 'Compétences', 'set.v.bio': 'Présentation', 'set.v.github': 'GitHub', 'set.v.collab': 'Ouvert aux collaborations', 'set.v.langs': 'Langues',
    'set.v.always': 'Toujours public', 'set.v.never': 'Jamais public', 'set.v.group': 'Visibilité de {f}'
  },
  en: {
    'set.title': 'Settings — 226os', 'set.h1': 'Settings', 'set.nav': 'Settings sections',
    'set.nav.profil': 'Profile', 'set.nav.compte': 'Account', 'set.nav.preferences': 'Preferences', 'set.nav.confidentialite': 'Privacy',
    'set.save': 'Save', 'set.saved': 'Changes saved.',
    'set.p.h': 'Profile', 'set.p.lede': 'What other members see on your page.',
    'set.p.avatar': 'Your avatar is generated from your initials.',
    'set.p.name': 'Name', 'set.p.name.hint': 'Your full name or the name you are known by.',
    'set.p.username': 'Username', 'set.p.username.hint': 'Lowercase letters, digits, hyphens or underscores (3 to 30 characters).',
    'set.p.bio': 'About you', 'set.p.bio.hint': 'Two sentences are enough to build trust.',
    'set.p.skills': 'Skills', 'set.p.skills.hint': 'Up to 8. They shape your home and recommendations.', 'set.p.skills.add': 'Add a skill', 'set.p.skills.addBtn': 'Add', 'set.p.skills.suggest': 'Suggestions', 'set.p.skills.remove': 'Remove {s}', 'set.p.skills.selected': 'Chosen skills',
    'set.p.country': 'Country', 'set.p.country.hint': 'APIs covering your market come first.', 'set.p.country.none': 'Choose a country',
    'set.p.langs': 'Working languages', 'set.p.langs.hint': 'So you read suitable content first.',
    'set.p.github': 'GitHub', 'set.p.github.hint': 'Username or URL. Maintainers look at your GitHub first.', 'set.p.github.ph': 'github.com/your-name',
    'set.p.collab': 'Open to collaborate', 'set.p.collab.hint': 'Maintainers looking for help can find you.',
    'set.e.required': 'This field is required.', 'set.e.username': 'Use 3 to 30 characters: lowercase letters, digits, - or _.', 'set.e.skills': 'Choose at least one skill.', 'set.e.langs': 'Choose at least one language.',
    'set.e.github': 'Enter a valid GitHub username or URL.', 'set.e.email': 'Enter a valid email address.',
    'set.e.pwCurrent': 'Enter your current password.', 'set.e.pwNew': 'Use at least 12 characters.', 'set.e.pwSame': 'The two passwords do not match.',
    'set.c.h': 'Account', 'set.c.lede': 'Your sign-in and your devices.',
    'set.c.email': 'Email address', 'set.c.email.hint': 'Used to sign in. It is never public.', 'set.c.email.save': 'Update email', 'set.c.email.saved': 'Email address updated.',
    'set.c.pw': 'Password', 'set.c.pw.current': 'Current password', 'set.c.pw.new': 'New password', 'set.c.pw.new.hint': 'At least 12 characters. A long phrase beats a complicated word.', 'set.c.pw.confirm': 'Confirm new password', 'set.c.pw.save': 'Change password', 'set.c.pw.saved': 'Password changed.',
    'set.c.sessions': 'Signed-in devices', 'set.c.sessions.lede': 'Sign out the ones you do not recognise.', 'set.c.sessions.this': 'This device', 'set.c.sessions.now': 'Active now', 'set.c.sessions.last': 'Last active: {date}',
    'set.c.sessions.out': 'Sign out {d}', 'set.c.sessions.all': 'Sign out other devices', 'set.c.sessions.done': 'Device signed out.', 'set.c.sessions.alldone': 'Other devices are signed out.', 'set.c.sessions.none': 'No other device signed in.',
    'set.f.h': 'Preferences', 'set.f.lede': 'Applied immediately.',
    'set.f.lang': 'Interface language', 'set.f.lang.hint': 'Overrides the language your browser reports. Your content keeps its original language.',
    'set.f.theme': 'Theme', 'set.f.theme.hint': 'Light, dark, or follow your system.', 'set.f.theme.light': 'Light', 'set.f.theme.dark': 'Dark', 'set.f.theme.system': 'System', 'set.f.saved': 'Preference saved.',
    'set.v.h': 'Privacy', 'set.v.lede': 'Choose who sees each piece of your profile.',
    'set.v.public': 'Everyone', 'set.v.members': 'Signed-in members', 'set.v.self': 'Only me',
    'set.v.name': 'Name', 'set.v.username': 'Username', 'set.v.email': 'Email address', 'set.v.country': 'Country', 'set.v.skills': 'Skills', 'set.v.bio': 'About you', 'set.v.github': 'GitHub', 'set.v.collab': 'Open to collaborate', 'set.v.langs': 'Languages',
    'set.v.always': 'Always public', 'set.v.never': 'Never public', 'set.v.group': 'Visibility of {f}'
  }
});

(function () {
  'use strict';
  var root = document.getElementById('settings-root'); if (!root) return;
  var A = window.App, M = window.MOCK, S = A.Session, t = function (k, p) { return I18N.t(k, p); }, esc = A.esc, $ = A.$, $$ = A.$$;
  if (!S.requireAuth()) return;
  var GH = /^(?:https?:\/\/)?(?:www\.)?(?:github\.com\/)?@?([A-Za-z0-9](?:[A-Za-z0-9-]{0,38}))\/?$/;
  var SECTIONS = [['profil', 'user'], ['compte', 'lock'], ['preferences', 'sliders'], ['confidentialite', 'eye']];
  var FIELDKEYS = ['skills', 'country', 'github', 'bio', 'openToCollab', 'languages'];
  var SUGGEST = ['go', 'postgresql', 'javascript', 'react', 'php', 'python', 'webhook', 'mobile-money', 'ussd', 'flutter', 'docker', 'kyc'];
  var skills = [], sessions = null;

  /* ---------- markup helpers (all test ids derive from the field name) ---------- */
  function fld(name, labelKey, ctl, o) {
    o = o || {}; var d = []; if (o.hint) d.push('set-' + name + '-hint'); d.push('set-' + name + '-error');
    return '<div class="field" id="' + (o.anchor || name) + '" data-testid="set-' + name + '-field"><label class="label" for="set-' + name + '-input" data-i18n="' + labelKey + '" data-testid="set-' + name + '-label"></label>' +
      (o.hint ? '<p class="hint" id="set-' + name + '-hint" data-i18n="' + labelKey + '.hint" data-testid="set-' + name + '-hint"></p>' : '') + ctl +
      '<p class="error-msg" id="set-' + name + '-error" role="alert" hidden data-testid="set-' + name + '-error"></p></div>';
  }
  function input(name, attrs) { return '<input class="input" id="set-' + name + '-input" name="' + name + '" ' + (attrs || '') + ' aria-describedby="set-' + name + '-hint set-' + name + '-error" data-testid="set-' + name + '-input">'; }
  function pw(name, auto) { return '<div class="set-pw"><input class="input" id="set-' + name + '-input" name="' + name + '" type="password" autocomplete="' + auto + '" aria-describedby="set-' + name + '-hint set-' + name + '-error" data-testid="set-' + name + '-input">' +
    '<button type="button" class="btn btn-icon" data-pw-toggle="#set-' + name + '-input" aria-pressed="false" data-testid="set-' + name + '-toggle"><span class="icon i-eye icon-main" aria-hidden="true" data-testid="set-' + name + '-toggle-main"></span><span class="icon i-eye-off icon-alt" aria-hidden="true" data-testid="set-' + name + '-toggle-alt"></span></button></div>'; }
  function status(id, key) { return '<span class="set-status" id="' + id + '" role="status" hidden data-testid="' + id + '"><span class="icon i-check-circle" aria-hidden="true" data-testid="' + id + '-icon"></span><span data-i18n="' + key + '" data-testid="' + id + '-text"></span></span>'; }
  function actions(form, labelKey) { return '<div class="set-actions"><button type="submit" class="btn btn-primary" data-testid="set-' + form + '-save"><span data-i18n="' + labelKey + '" data-testid="set-' + form + '-save-label"></span></button>' + status('set-' + form + '-status', form === 'email' ? 'set.c.email.saved' : form === 'pw' ? 'set.c.pw.saved' : 'set.saved') + '</div>'; }
  function head(k, icon) { return '<header><h2 id="set-' + k + '-h" data-i18n="set.' + k + '.h" data-testid="set-' + k + '-title"></h2><p data-i18n="set.' + k + '.lede" data-testid="set-' + k + '-lede"></p></header>'; }
  function section(k, body) { return '<section class="set-section" id="' + k + '" aria-labelledby="set-' + k + '-h" hidden data-testid="set-section-' + k + '">' + body + '</section>'; }

  function build() {
    var U = S.user, cs = M.countries.map(function (c) { return '<option value="' + c.code + '" data-testid="set-country-opt-' + c.code.toLowerCase() + '">' + esc(c[I18N.lang] || c.fr) + '</option>'; }).join('');
    var nav = '<nav class="set-nav" data-i18n-attr="aria-label:set.nav" data-testid="set-nav"><ul>' + SECTIONS.map(function (s) { return '<li><a href="#' + s[0] + '" data-sec="' + s[0] + '" data-testid="set-nav-' + s[0] + '"><span class="icon i-' + s[1] + '" aria-hidden="true" data-testid="set-nav-' + s[0] + '-icon"></span><span data-i18n="set.nav.' + s[0] + '" data-testid="set-nav-' + s[0] + '-label"></span></a></li>'; }).join('') + '</ul></nav>';

    var profil = section('profil', head('p') + '<form class="set-form" id="set-profile-form" novalidate data-testid="set-profile-form">' +
      '<div class="set-avatar-row">' + A.avatar(U.name, { size: 'lg', testid: 'set-avatar' }) + '<p data-i18n="set.p.avatar" data-testid="set-avatar-hint"></p></div>' +
      '<div class="set-grid">' + fld('name', 'set.p.name', input('name', 'type="text" autocomplete="name" maxlength="60" required'), { hint: 1 }) +
      fld('username', 'set.p.username', input('username', 'type="text" autocomplete="username" autocapitalize="none" spellcheck="false" maxlength="30" required'), { hint: 1 }) + '</div>' +
      fld('bio', 'set.p.bio', '<textarea class="textarea" id="set-bio-input" name="bio" rows="4" maxlength="280" lang="' + I18N.lang + '" data-user-content aria-describedby="set-bio-hint set-bio-error" data-testid="set-bio-input"></textarea><span class="set-counter" id="set-bio-counter" data-testid="set-bio-counter"></span>', { hint: 1, anchor: 'bio' }) +
      '<div class="field" id="skills" data-testid="set-skills-field"><span class="label" id="set-skills-label" data-i18n="set.p.skills" data-testid="set-skills-label"></span><p class="hint" id="set-skills-hint" data-i18n="set.p.skills.hint" data-testid="set-skills-hint"></p>' +
      '<div class="set-skills" id="set-skills-chosen" role="group" aria-labelledby="set-skills-label" data-testid="set-skills-chosen"></div>' +
      '<div class="set-skills" id="set-skills-suggest" role="group" data-i18n-attr="aria-label:set.p.skills.suggest" data-testid="set-skills-suggest"></div>' +
      '<div class="set-skill-add"><input class="input" id="set-skills-input" type="text" autocomplete="off" spellcheck="false" maxlength="24" data-i18n-attr="aria-label:set.p.skills.add;placeholder:set.p.skills.add" aria-describedby="set-skills-hint set-skills-error" data-testid="set-skills-input"><button type="button" class="btn" id="set-skills-add" data-testid="set-skills-add"><span class="icon i-plus" aria-hidden="true" data-testid="set-skills-add-icon"></span><span data-i18n="set.p.skills.addBtn" data-testid="set-skills-add-label"></span></button></div>' +
      '<p class="error-msg" id="set-skills-error" role="alert" hidden data-testid="set-skills-error"></p></div>' +
      '<div class="set-grid">' + fld('country', 'set.p.country', '<select class="select" id="set-country-input" name="country" aria-describedby="set-country-hint set-country-error" data-testid="set-country-input"><option value="" data-i18n="set.p.country.none" data-testid="set-country-opt-none"></option>' + cs + '</select>', { hint: 1 }) +
      fld('github', 'set.p.github', input('github', 'type="text" inputmode="url" autocomplete="off" spellcheck="false" data-i18n-attr="placeholder:set.p.github.ph"'), { hint: 1 }) + '</div>' +
      '<fieldset class="fieldset-plain field" id="languages" data-testid="set-languages-field"><legend class="label" data-i18n="set.p.langs" data-testid="set-languages-label"></legend><p class="hint" id="set-languages-hint" data-i18n="set.p.langs.hint" data-testid="set-languages-hint"></p><div class="set-choices">' +
      ['fr', 'en'].map(function (l) { return '<label class="choice" data-testid="set-languages-choice-wrap-' + l + '"><input type="checkbox" name="languages" value="' + l + '" id="set-languages-input-' + l + '" aria-describedby="set-languages-error" data-testid="set-languages-input-' + l + '"><span lang="' + l + '" data-testid="set-languages-choice-' + l + '">' + (l === 'fr' ? 'Français' : 'English') + '</span></label>'; }).join('') + '</div><p class="error-msg" id="set-languages-error" role="alert" hidden data-testid="set-languages-error"></p></fieldset>' +
      '<div class="field" id="openToCollab" data-testid="set-collab-field"><label class="choice" data-testid="set-collab-choice-wrap"><input type="checkbox" name="openToCollab" id="set-collab-input" aria-describedby="set-collab-hint" data-testid="set-collab-input"><span data-i18n="set.p.collab" data-testid="set-collab-label"></span></label><p class="hint" id="set-collab-hint" data-i18n="set.p.collab.hint" data-testid="set-collab-hint"></p></div>' +
      actions('profile', 'set.save') + '</form>');

    var compte = section('compte', head('c') + '<form class="set-form" id="set-email-form" novalidate data-testid="set-email-form">' + fld('email', 'set.c.email', input('email', 'type="email" autocomplete="email" inputmode="email"'), { hint: 1 }) + actions('email', 'set.c.email.save') + '</form>' +
      '<hr class="set-sep" data-testid="set-sep-1">' +
      '<form class="set-form" id="set-pw-form" novalidate data-testid="set-pw-form"><h3 data-i18n="set.c.pw" data-testid="set-pw-title"></h3><input type="text" autocomplete="username" value="' + esc(U.username) + '" hidden aria-hidden="true" tabindex="-1" data-testid="set-pw-user">' +
      fld('pwcur', 'set.c.pw.current', pw('pwcur', 'current-password')) + '<div class="set-grid">' + fld('pwnew', 'set.c.pw.new', pw('pwnew', 'new-password'), { hint: 1 }) + fld('pwconf', 'set.c.pw.confirm', pw('pwconf', 'new-password')) + '</div>' + actions('pw', 'set.c.pw.save') + '</form>' +
      '<hr class="set-sep" data-testid="set-sep-2">' +
      '<div class="stack-sm" data-testid="set-sessions"><h3 data-i18n="set.c.sessions" data-testid="set-sessions-title"></h3><p class="hint" data-i18n="set.c.sessions.lede" data-testid="set-sessions-lede"></p><ul class="set-sessions" id="set-sessions-list" data-testid="set-sessions-list"></ul>' +
      '<div class="set-actions"><button type="button" class="btn" id="set-sessions-all" data-testid="set-sessions-all"><span class="icon i-log-out" aria-hidden="true" data-testid="set-sessions-all-icon"></span><span data-i18n="set.c.sessions.all" data-testid="set-sessions-all-label"></span></button><span class="set-status" id="set-sessions-status" role="status" hidden data-testid="set-sessions-status"><span class="icon i-check-circle" aria-hidden="true" data-testid="set-sessions-status-icon"></span><span id="set-sessions-status-text" data-testid="set-sessions-status-text"></span></span></div></div>');

    var prefs = section('preferences', head('f') + '<div class="set-form"><div class="set-pref" data-testid="set-lang-pref"><span class="label" id="set-lang-label" data-i18n="set.f.lang" data-testid="set-lang-label"></span><p class="hint" data-i18n="set.f.lang.hint" data-testid="set-lang-hint"></p>' +
      '<div class="segmented" role="group" aria-labelledby="set-lang-label" data-testid="set-lang"><button type="button" data-action="lang-set" data-lang="fr" lang="fr" aria-pressed="false" data-testid="set-lang-fr">Français</button><button type="button" data-action="lang-set" data-lang="en" lang="en" aria-pressed="false" data-testid="set-lang-en">English</button></div></div>' +
      '<div class="set-pref" data-testid="set-theme-pref"><span class="label" id="set-theme-label" data-i18n="set.f.theme" data-testid="set-theme-label"></span><p class="hint" data-i18n="set.f.theme.hint" data-testid="set-theme-hint"></p>' +
      '<div class="segmented" role="radiogroup" aria-labelledby="set-theme-label" data-testid="set-theme">' + [['light', 'sun'], ['dark', 'moon'], ['system', 'settings']].map(function (x) { return '<button type="button" role="radio" aria-checked="false" data-action="theme-set" data-theme-value="' + x[0] + '" data-testid="set-theme-' + x[0] + '"><span class="icon i-' + x[1] + '" aria-hidden="true" data-testid="set-theme-' + x[0] + '-icon"></span><span data-i18n="set.f.theme.' + x[0] + '" data-testid="set-theme-' + x[0] + '-label"></span></button>'; }).join('') + '</div></div>' +
      status('set-pref-status', 'set.f.saved') + '</div>');

    var rows = [['country', 'map-pin'], ['skills', 'tag'], ['bio', 'file-text'], ['github', 'git-branch'], ['collab', 'users'], ['langs', 'languages']];
    var conf = section('confidentialite', head('v') + '<form class="set-form" id="set-vis-form" novalidate data-testid="set-vis-form"><ul class="set-vis" data-testid="set-vis-list">' +
      [['name', 'user', 'always'], ['username', 'user', 'always']].map(function (r) { return '<li data-testid="set-vis-row-' + r[0] + '"><span class="vis-name"><span class="icon i-' + r[1] + '" aria-hidden="true" data-testid="set-vis-' + r[0] + '-icon"></span><span data-i18n="set.v.' + r[0] + '" data-testid="set-vis-' + r[0] + '-label"></span></span><span class="vis-fixed"><span class="icon i-lock" aria-hidden="true" data-testid="set-vis-' + r[0] + '-fixed-icon"></span><span data-i18n="set.v.always" data-testid="set-vis-' + r[0] + '-fixed"></span></span></li>'; }).join('') +
      '<li data-testid="set-vis-row-email"><span class="vis-name"><span class="icon i-lock" aria-hidden="true" data-testid="set-vis-email-icon"></span><span data-i18n="set.v.email" data-testid="set-vis-email-label"></span></span><span class="vis-fixed"><span class="icon i-eye-off" aria-hidden="true" data-testid="set-vis-email-fixed-icon"></span><span data-i18n="set.v.never" data-testid="set-vis-email-fixed"></span></span></li>' +
      rows.map(function (r) { return '<li data-testid="set-vis-row-' + r[0] + '"><label class="vis-name" for="set-vis-' + r[0] + '-input" data-testid="set-vis-' + r[0] + '-name"><span class="icon i-' + r[1] + '" aria-hidden="true" data-testid="set-vis-' + r[0] + '-icon"></span><span data-i18n="set.v.' + r[0] + '" data-testid="set-vis-' + r[0] + '-label"></span></label>' +
        '<select class="select" id="set-vis-' + r[0] + '-input" name="' + r[0] + '" data-testid="set-vis-' + r[0] + '-input">' + ['public', 'members', 'self'].map(function (v) { return '<option value="' + v + '" data-i18n="set.v.' + v + '" data-testid="set-vis-' + r[0] + '-opt-' + v + '"></option>'; }).join('') + '</select></li>'; }).join('') +
      '</ul>' + actions('vis', 'set.save') + '</form>');

    root.innerHTML = '<header class="page-header"><h1 data-i18n="set.h1" data-testid="set-h1"></h1></header><div class="set-layout" data-testid="set-layout">' + nav + '<div class="set-content" data-testid="set-content">' + profil + compte + prefs + conf + '</div></div>';
  }

  /* ---------- helpers ---------- */
  function err(name, key, ctl) {
    var e = $('#set-' + name + '-error'), host = e.closest('.field'), i = ctl || $('#set-' + name + '-input');
    if (key) { e.textContent = t(key); e.hidden = false; if (host) host.classList.add('has-error'); if (i) i.setAttribute('aria-invalid', 'true'); e.setAttribute('data-key', key); }
    else { e.hidden = true; e.textContent = ''; e.removeAttribute('data-key'); if (host) host.classList.remove('has-error'); if (i) i.removeAttribute('aria-invalid'); }
    return !key;
  }
  function flash(id) { var s = $('#' + id); if (!s) return; s.hidden = false; clearTimeout(s._t); s._t = setTimeout(function () { s.hidden = true; }, 4000); }
  function clearStatus(form) { $$('.set-status', form).forEach(function (s) { s.hidden = true; }); }
  function focusFirstError(form) { var e = $('.has-error input, .has-error select, .has-error textarea', form); if (e) e.focus(); }

  /* ---------- skills editor ---------- */
  function renderSkills() {
    var ch = $('#set-skills-chosen'), sg = $('#set-skills-suggest');
    ch.innerHTML = skills.map(function (s) { return '<button type="button" class="chip" aria-pressed="true" data-skill="' + esc(s) + '" aria-label="' + esc(t('set.p.skills.remove', { s: s })) + '" data-testid="set-skill-' + esc(s) + '">' + esc(s) + '</button>'; }).join('');
    sg.innerHTML = SUGGEST.filter(function (s) { return skills.indexOf(s) < 0; }).slice(0, 8).map(function (s) { return '<button type="button" class="chip" aria-pressed="false" data-skill-add="' + s + '" data-testid="set-skill-suggest-' + s + '">' + esc(s) + '</button>'; }).join('');
    $('#set-skills-add').disabled = skills.length >= 8;
  }
  function addSkill(v) {
    v = String(v || '').trim().toLowerCase().replace(/\s+/g, '-'); if (!v) return;
    if (skills.indexOf(v) < 0 && skills.length < 8) skills.push(v); renderSkills(); err('skills', null, $('#set-skills-input'));
  }

  /* ---------- load values from the session ---------- */
  function load() {
    var U = S.user, vis = U.visibility || {};
    $('#set-name-input').value = U.name || ''; $('#set-username-input').value = U.username || ''; $('#set-bio-input').value = U.bio || ''; counter();
    $('#set-country-input').value = U.country || ''; $('#set-github-input').value = U.github || ''; $('#set-collab-input').checked = !!U.openToCollab;
    ['fr', 'en'].forEach(function (l) { $('#set-languages-input-' + l).checked = (U.languages || []).indexOf(l) > -1; });
    skills = (U.skills || []).slice(); renderSkills();
    $('#set-email-input').value = U.email || ((U.username || 'membre') + '@example.org');
    ['country', 'skills', 'bio', 'github', 'collab', 'langs'].forEach(function (k) { $('#set-vis-' + k + '-input').value = vis[k] || 'public'; });
  }
  function counter() { var b = $('#set-bio-input'); $('#set-bio-counter').textContent = b.value.length + '/280'; }

  /* ---------- sessions (simulated devices) ---------- */
  function resetSessions() { sessions = [{ id: 'cur', dev: 'Chrome · Linux', place: 'Ouagadougou', cur: true, date: '' }, { id: 's2', dev: 'Firefox · Android', place: 'Bobo-Dioulasso', date: '2026-11-09' }, { id: 's3', dev: 'Safari · iPhone', place: 'Accra', date: '2026-11-02' }]; }
  function renderSessions() {
    $('#set-sessions-list').innerHTML = sessions.map(function (s) {
      return '<li class="set-session" data-testid="set-session-' + s.id + '"><span class="id-row-lead-icon" aria-hidden="true" data-testid="set-session-' + s.id + '-iconwrap"><span class="icon i-' + (s.id === 's2' ? 'send' : 'home') + '" data-testid="set-session-' + s.id + '-icon"></span></span><div><strong lang="fr" data-testid="set-session-' + s.id + '-device">' + esc(s.dev) + ' · ' + esc(s.place) + '</strong>' +
        '<span class="text" data-testid="set-session-' + s.id + '-when">' + (s.cur ? esc(t('set.c.sessions.now')) : esc(t('set.c.sessions.last', { date: M.fmtDate(s.date) }))) + '</span></div>' +
        (s.cur ? A.statePill('open', { testid: 'set-session-cur-pill', labelKey: 'set.c.sessions.this' }) : '<button type="button" class="btn btn-icon btn-sm has-tip" data-session-out="' + s.id + '" aria-label="' + esc(t('set.c.sessions.out', { d: s.dev })) + '" data-testid="set-session-' + s.id + '-out"><span class="icon i-log-out" aria-hidden="true" data-testid="set-session-' + s.id + '-out-icon"></span></button>') + '</li>';
    }).join('');
    $('#set-sessions-all').disabled = sessions.length < 2;
  }

  /* ---------- hash routing ---------- */
  function route(focusKeys) {
    var h = decodeURIComponent((location.hash || '').slice(1)), sec = SECTIONS.some(function (s) { return s[0] === h; }) ? h : (FIELDKEYS.indexOf(h) > -1 ? 'profil' : 'profil');
    $$('.set-section', root).forEach(function (s) { s.hidden = s.id !== sec; });
    $$('.set-nav a', root).forEach(function (a) { if (a.getAttribute('data-sec') === sec) a.setAttribute('aria-current', 'true'); else a.removeAttribute('aria-current'); });
    if (FIELDKEYS.indexOf(h) > -1) {
      var f = document.getElementById(h); if (f) { f.scrollIntoView({ block: 'center' }); f.classList.remove('set-field-focus'); void f.offsetWidth; f.classList.add('set-field-focus'); var c = $('input, select, textarea, button', f); if (c) c.focus({ preventScroll: true }); }
    }
  }

  /* ---------- forms ---------- */
  function saveProfile(e) {
    e.preventDefault(); var f = e.target, ok = true;
    var name = $('#set-name-input').value.trim(), un = $('#set-username-input').value.trim().replace(/^@/, ''), gh = $('#set-github-input').value.trim(), langs = ['fr', 'en'].filter(function (l) { return $('#set-languages-input-' + l).checked; });
    clearStatus(f);
    ok = err('name', name ? null : 'set.e.required') && ok;
    ok = err('username', /^[a-z0-9][a-z0-9_-]{2,29}$/.test(un) ? null : (un ? 'set.e.username' : 'set.e.required')) && ok;
    ok = err('skills', skills.length ? null : 'set.e.skills', $('#set-skills-input')) && ok;
    var gm = gh ? GH.exec(gh) : null; ok = err('github', !gh || gm ? null : 'set.e.github') && ok;
    ok = err('languages', langs.length ? null : 'set.e.langs', $('#set-languages-input-fr')) && ok;
    if (!ok) { focusFirstError(f); return; }
    S.update({ name: name, username: un, bio: $('#set-bio-input').value.trim(), skills: skills.slice(), country: $('#set-country-input').value, github: gm ? gm[1] : '', languages: langs, openToCollab: $('#set-collab-input').checked });
    S.confirm('languages'); S.confirm('openToCollab'); $('#set-github-input').value = gm ? gm[1] : ''; flash('set-profile-status');
  }
  function saveEmail(e) {
    e.preventDefault(); var v = $('#set-email-input').value.trim(); clearStatus(e.target);
    if (!err('email', /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/.test(v) ? null : 'set.e.email')) { $('#set-email-input').focus(); return; }
    S.update({ email: v }); flash('set-email-status');
  }
  function savePw(e) {
    e.preventDefault(); var f = e.target, c = $('#set-pwcur-input').value, n = $('#set-pwnew-input').value, k = $('#set-pwconf-input').value; clearStatus(f);
    var ok = err('pwcur', c ? null : 'set.e.pwCurrent'); ok = err('pwnew', n.length >= 12 ? null : 'set.e.pwNew') && ok; ok = err('pwconf', n && k === n ? null : 'set.e.pwSame') && ok;
    if (!ok) { focusFirstError(f); return; }
    ['pwcur', 'pwnew', 'pwconf'].forEach(function (x) { $('#set-' + x + '-input').value = ''; }); flash('set-pw-status');
  }
  function saveVis(e) {
    e.preventDefault(); var v = {}; ['country', 'skills', 'bio', 'github', 'collab', 'langs'].forEach(function (k) { v[k] = $('#set-vis-' + k + '-input').value; });
    S.update({ visibility: v }); flash('set-vis-status');
  }
  function sessStatus(key) { $('#set-sessions-status-text').textContent = t(key); $('#set-sessions-status-text').setAttribute('data-key', key); flash('set-sessions-status'); }

  function init() {
    build(); resetSessions(); I18N.apply(root); load(); renderSessions(); route();
    document.title = t('set.title');
    window.addEventListener('hashchange', function () { route(); });
    $('#set-profile-form').addEventListener('submit', saveProfile); $('#set-email-form').addEventListener('submit', saveEmail); $('#set-pw-form').addEventListener('submit', savePw); $('#set-vis-form').addEventListener('submit', saveVis);
    $('#set-bio-input').addEventListener('input', counter);
    $('#set-skills-input').addEventListener('keydown', function (e) { if (e.key === 'Enter') { e.preventDefault(); addSkill(e.target.value); e.target.value = ''; } });
    root.addEventListener('click', function (e) {
      var r = e.target.closest('[data-skill]'), a = e.target.closest('[data-skill-add]'), o = e.target.closest('[data-session-out]');
      if (r) { skills.splice(skills.indexOf(r.getAttribute('data-skill')), 1); renderSkills(); }
      else if (a) addSkill(a.getAttribute('data-skill-add'));
      else if (e.target.closest('#set-skills-add')) { var i = $('#set-skills-input'); addSkill(i.value); i.value = ''; i.focus(); }
      else if (o) { var id = o.getAttribute('data-session-out'); sessions = sessions.filter(function (s) { return s.id !== id; }); renderSessions(); sessStatus('set.c.sessions.done'); }
      else if (e.target.closest('#set-sessions-all')) { sessions = sessions.filter(function (s) { return s.cur; }); renderSessions(); sessStatus('set.c.sessions.alldone'); }
      else if (e.target.closest('[data-action="lang-set"],[data-action="theme-set"]')) { setTimeout(function () { flash('set-pref-status'); }, 0); }
    });
    document.addEventListener('i18n:change', function () {
      document.title = t('set.title'); renderSkills(); renderSessions();
      $$('.error-msg[data-key]', root).forEach(function (x) { x.textContent = t(x.getAttribute('data-key')); });
      var st = $('#set-sessions-status-text'); if (st.getAttribute('data-key')) st.textContent = t(st.getAttribute('data-key'));
      var sel = $('#set-country-input'), cur = sel.value; $$('option[value]', sel).forEach(function (o) { var c = M.countries.filter(function (x) { return x.code === o.value; })[0]; if (c) o.textContent = c[I18N.lang]; }); sel.value = cur;
    });
  }
  init();
})();
