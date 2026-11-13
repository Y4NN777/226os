// identity catalogue, merged by src/i18n/messages.ts. Keys start with 'identity.'.
// French typography rules are enforced by src/i18n/translate.test.ts.
export const fr = {
  'identity.field.email': 'Adresse e-mail',
  'identity.field.password': 'Mot de passe',
  'identity.field.newPassword': 'Nouveau mot de passe',
  'identity.field.code': 'Code de confirmation',
  'identity.field.code.description': 'Copiez le code reçu par e-mail et collez-le ici.',
  'identity.password.show': 'Afficher le mot de passe',
  'identity.err.email.required': 'Saisissez votre adresse e-mail.',
  'identity.err.email.format': 'Cette adresse e-mail semble incomplète.',
  'identity.err.password.required': 'Saisissez votre mot de passe.',
  'identity.err.password.tooShort': 'Le mot de passe doit contenir au moins 12 caractères.',
  'identity.err.password.tooLong': 'Le mot de passe ne peut pas dépasser 128 caractères.',
  'identity.err.code.required': 'Saisissez le code reçu par e-mail.',
  'identity.err.code.length': 'Le code doit contenir entre 16 et 256 caractères.',
  'identity.rule.met': '12 caractères minimum\u00A0: respecté',
  'identity.rule.unmet': '12 caractères minimum\u00A0: pas encore atteint',
  'identity.error.title': 'L’action a échoué',
  'identity.limit.title': 'Trop de tentatives',
  'identity.limit.body': 'Patientez un instant avant de réessayer.',
  'identity.limit.wait': 'Réessayer dans {seconds} s',
  'identity.back.signIn': 'Retour à la connexion',

  'identity.signIn.title': 'Connexion',
  'identity.signIn.submit': 'Se connecter',
  'identity.signIn.forgot': 'Mot de passe oublié\u00A0?',
  'identity.signIn.noAccount': 'Pas encore de compte\u00A0?',
  'identity.signIn.signUp': 'Créer un compte',
  'identity.signIn.bad.title': 'Connexion impossible',
  'identity.signIn.bad.body': 'E-mail ou mot de passe incorrect.',
  'identity.signIn.unverified.title': 'Adresse non confirmée',
  'identity.signIn.unverified.body':
    'Confirmez votre adresse e-mail avant de vous connecter, avec le code reçu à l’inscription.',
  'identity.signIn.unverified.link': 'Confirmer mon adresse',

  'identity.signUp.title': 'Créer votre compte',
  'identity.signUp.name.label': 'Nom d’affichage',
  'identity.signUp.username.label': 'Nom d’utilisateur',
  'identity.signUp.username.description':
    'Identifiant public. Lettres minuscules, chiffres, - et _.',
  'identity.signUp.err.username.required': 'Choisissez un nom d’utilisateur.',
  'identity.signUp.err.username.format':
    'De 3 à 30 caractères\u00A0: lettres minuscules, chiffres, - et _, en commençant par une lettre ou un chiffre.',
  'identity.signUp.err.name.tooLong': 'Le nom ne peut pas dépasser 80 caractères.',
  'identity.signUp.submit': 'Créer mon compte',
  'identity.signUp.hasAccount': 'Vous avez déjà un compte\u00A0?',
  'identity.signUp.signIn': 'Se connecter',

  'identity.verify.title': 'Vérifiez votre boîte mail',
  'identity.verify.intro':
    'Si votre inscription est valide, un e-mail contenant un code de confirmation vous attend. Pensez à regarder dans les courriers indésirables.',
  'identity.verify.submit': 'Confirmer',
  'identity.verify.checking': 'Vérification du code…',
  'identity.verify.done.title': 'Adresse confirmée',
  'identity.verify.done.body': 'Votre compte est prêt. Connectez-vous pour continuer.',
  'identity.verify.done.signIn': 'Se connecter',
  'identity.code.invalid.title': 'Ce code n’est pas valide ou a expiré',
  'identity.verify.invalid.body':
    'Il a peut-être déjà servi. Réessayez avec un autre code, créez de nouveau votre compte, ou contactez l’équipe si le problème persiste.',
  'identity.verify.invalid.retry': 'Saisir un autre code',
  'identity.verify.invalid.signUp': 'Créer un compte',

  'identity.forgot.title': 'Mot de passe oublié',
  'identity.forgot.intro':
    'Saisissez l’adresse de votre compte. Nous vous envoyons un code pour choisir un nouveau mot de passe.',
  'identity.forgot.submit': 'Envoyer le code',
  'identity.forgot.done.title': 'Consultez votre boîte mail',
  'identity.forgot.done.body':
    'Si un compte existe pour cette adresse, un code vient de partir. Il expire dans 30 minutes.',
  'identity.forgot.done.link': 'Saisir le code',

  'identity.reset.title': 'Choisir un nouveau mot de passe',
  'identity.reset.intro':
    'Saisissez le code reçu par e-mail, puis votre nouveau mot de passe. Il remplace l’ancien sur tous vos appareils.',
  'identity.reset.submit': 'Enregistrer',
  'identity.reset.done.title': 'Mot de passe modifié',
  'identity.reset.done.body':
    'Tous vos appareils ont été déconnectés. Connectez-vous avec votre nouveau mot de passe.',
  'identity.reset.done.signIn': 'Se connecter',
  'identity.reset.invalid.body': 'Il a peut-être déjà servi. Demandez-en un nouveau.',
  'identity.reset.invalid.retry': 'Demander un nouveau code',

  'identity.welcome.title': 'Bienvenue, {name}',
  'identity.welcome.anonymous.title': 'Bienvenue sur 226os',
  'identity.welcome.anonymous.body':
    'Connectez-vous pour poser vos questions, répondre et suivre vos projets.',
  'identity.welcome.intro': 'Voici par où commencer.',
  'identity.welcome.discover': 'Découvrir le catalogue',
  'identity.welcome.search': 'Rechercher',
  'identity.welcome.settings': 'Réglages',
} as const;

export const en: Record<keyof typeof fr, string> = {
  'identity.field.email': 'Email address',
  'identity.field.password': 'Password',
  'identity.field.newPassword': 'New password',
  'identity.field.code': 'Confirmation code',
  'identity.field.code.description': 'Copy the code from the email and paste it here.',
  'identity.password.show': 'Show password',
  'identity.err.email.required': 'Enter your email address.',
  'identity.err.email.format': 'This email address looks incomplete.',
  'identity.err.password.required': 'Enter your password.',
  'identity.err.password.tooShort': 'The password must be at least 12 characters long.',
  'identity.err.password.tooLong': 'The password cannot be longer than 128 characters.',
  'identity.err.code.required': 'Enter the code from the email.',
  'identity.err.code.length': 'The code must be between 16 and 256 characters long.',
  'identity.rule.met': 'At least 12 characters: met',
  'identity.rule.unmet': 'At least 12 characters: not met yet',
  'identity.error.title': 'The action failed',
  'identity.limit.title': 'Too many attempts',
  'identity.limit.body': 'Wait a moment before trying again.',
  'identity.limit.wait': 'Try again in {seconds}s',
  'identity.back.signIn': 'Back to sign in',

  'identity.signIn.title': 'Sign in',
  'identity.signIn.submit': 'Sign in',
  'identity.signIn.forgot': 'Forgot your password?',
  'identity.signIn.noAccount': 'No account yet?',
  'identity.signIn.signUp': 'Create an account',
  'identity.signIn.bad.title': 'Cannot sign in',
  'identity.signIn.bad.body': 'Incorrect email or password.',
  'identity.signIn.unverified.title': 'Address not confirmed',
  'identity.signIn.unverified.body':
    'Confirm your email address before signing in, using the code you received when you signed up.',
  'identity.signIn.unverified.link': 'Confirm my address',

  'identity.signUp.title': 'Create your account',
  'identity.signUp.name.label': 'Display name',
  'identity.signUp.username.label': 'Username',
  'identity.signUp.username.description': 'Public identifier. Lowercase letters, digits, - and _.',
  'identity.signUp.err.username.required': 'Choose a username.',
  'identity.signUp.err.username.format':
    '3 to 30 characters: lowercase letters, digits, - and _, starting with a letter or a digit.',
  'identity.signUp.err.name.tooLong': 'The name cannot be longer than 80 characters.',
  'identity.signUp.submit': 'Create my account',
  'identity.signUp.hasAccount': 'Already have an account?',
  'identity.signUp.signIn': 'Sign in',

  'identity.verify.title': 'Check your inbox',
  'identity.verify.intro':
    'If your sign-up is valid, an email with a confirmation code is waiting for you. Look in your spam folder too.',
  'identity.verify.submit': 'Confirm',
  'identity.verify.checking': 'Checking the code…',
  'identity.verify.done.title': 'Address confirmed',
  'identity.verify.done.body': 'Your account is ready. Sign in to continue.',
  'identity.verify.done.signIn': 'Sign in',
  'identity.code.invalid.title': 'This code is not valid or has expired',
  'identity.verify.invalid.body':
    'It may already have been used. Try another code, create your account again, or contact the team if the problem persists.',
  'identity.verify.invalid.retry': 'Enter another code',
  'identity.verify.invalid.signUp': 'Create an account',

  'identity.forgot.title': 'Forgot your password?',
  'identity.forgot.intro':
    'Enter your account address. We send you a code to choose a new password.',
  'identity.forgot.submit': 'Send the code',
  'identity.forgot.done.title': 'Check your inbox',
  'identity.forgot.done.body':
    'If an account exists for this address, a code has just been sent. It expires in 30 minutes.',
  'identity.forgot.done.link': 'Enter the code',

  'identity.reset.title': 'Choose a new password',
  'identity.reset.intro':
    'Enter the code from the email, then your new password. It replaces the old one on all your devices.',
  'identity.reset.submit': 'Save',
  'identity.reset.done.title': 'Password changed',
  'identity.reset.done.body':
    'You were signed out of all your devices. Sign in with your new password.',
  'identity.reset.done.signIn': 'Sign in',
  'identity.reset.invalid.body': 'It may already have been used. Request a new one.',
  'identity.reset.invalid.retry': 'Request a new code',

  'identity.welcome.title': 'Welcome, {name}',
  'identity.welcome.anonymous.title': 'Welcome to 226os',
  'identity.welcome.anonymous.body': 'Sign in to ask questions, answer and follow your projects.',
  'identity.welcome.intro': 'Here is where to start.',
  'identity.welcome.discover': 'Browse the catalogue',
  'identity.welcome.search': 'Search',
  'identity.welcome.settings': 'Settings',
};
