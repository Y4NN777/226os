import * as discover from '../features/discover/messages';
import * as home from '../features/home/messages';
import * as identity from '../features/identity/messages';
import * as landing from '../features/landing/messages';
import * as search from '../features/search/messages';
import * as solve from '../features/solve/messages';

export type Locale = 'fr' | 'en';

export const locales = ['fr', 'en'] as const satisfies readonly Locale[];

export function isLocale(value: string): value is Locale {
  return (locales as readonly string[]).includes(value);
}

export const defaultLocale: Locale = 'fr';

// French typography: a non-breaking space (U+00A0) precedes : ; ? ! and
// typographic apostrophes/ellipsis are used. Enforced by translate.test.ts.
// Each feature owns its catalogue in features/<feature>/messages.ts (same rules);
// this file only merges them.
const core = {
  'app.name': '226os',
  'nav.label': 'Navigation principale',
  'nav.home': 'Accueil',
  'nav.settings': 'Réglages',
  'nav.learn': 'Apprendre',
  'nav.connect': 'Connect',
  'nav.solve': 'Résoudre',
  'nav.discover': 'Découvrir',
  'nav.build': 'Construire',
  'nav.search': 'Rechercher',
  'nav.soon': 'Bientôt',
  'nav.section.grow': 'Progresser',
  'nav.section.pillars': 'Piliers',
  'nav.open': 'Ouvrir la navigation',
  'nav.top': 'Barre du site',
  'account.signIn': 'Se connecter',
  'account.signUp': 'Créer un compte',
  'account.signOut': 'Se déconnecter',
  'account.settings': 'Réglages',
  'account.menu': 'Compte de {name}',
  'account.more': 'Plus d’options',
  'account.language': 'Langue',
  'account.signOutFailed': 'La déconnexion a échoué. Réessayez.',
  'home.title': 'Bienvenue sur 226os',
  'home.intro':
    'La plateforme d’entraide technique est en construction. Pour l’instant, cette page vérifie seulement que l’API répond.',
  'health.title': 'État de l’API',
  'health.loading': 'Vérification de l’API en cours…',
  'health.ok': 'API opérationnelle',
  'health.status': 'Statut : {status}',
  'health.error': 'L’API ne répond pas.',
  'health.errorHint': 'Vérifiez que le serveur est démarré, puis réessayez.',
  'health.retry': 'Vérifier à nouveau',
  'settings.title': 'Réglages',
  'settings.theme.label': 'Thème',
  'settings.theme.light': 'Clair',
  'settings.theme.dark': 'Sombre',
  'settings.theme.system': 'Système',
  'settings.language.label': 'Langue',
  'settings.language.fr': 'Français',
  'settings.language.en': 'English',
  'settings.visibility.title': 'Visibilité du profil',
  'settings.visibility.label': 'Qui peut voir votre profil ?',
  'settings.visibility.bylines':
    'Dans tous les cas, votre nom et votre identifiant restent affichés sur ce que vous publiez.',
  'settings.visibility.public': 'Public',
  'settings.visibility.public.hint': 'Tout le monde, même sans compte.',
  'settings.visibility.members': 'Membres',
  'settings.visibility.members.hint': 'Uniquement les membres connectés.',
  'settings.visibility.private': 'Privé',
  'settings.visibility.private.hint': 'Vous seul.',
  'settings.visibility.save': 'Enregistrer la visibilité',
  'settings.visibility.saved': 'Visibilité enregistrée.',
  'settings.visibility.failed': 'La visibilité n’a pas été enregistrée',
  'settings.visibility.soon':
    'Bientôt disponible : ce réglage s’activera dès que les profils seront en ligne.',
  'settings.visibility.signIn': 'Connectez-vous pour choisir qui voit votre profil',
  'form.name.label': 'Nom d’affichage',
  'form.name.description': 'Affiché dans l’en-tête et à côté de ce que vous publiez.',
  'form.name.required': 'Saisissez un nom d’affichage.',
  'form.name.tooShort': 'Le nom doit contenir au moins {min} caractères.',
  'form.name.tooLong': 'Le nom doit contenir {max} caractères au plus.',
  'form.name.failed': 'Le nom n’a pas été enregistré',
  'settings.name.signIn': 'Connectez-vous pour modifier votre nom',
  'form.save': 'Enregistrer',
  'form.saved': 'Nom enregistré.',
  'notFound.title': 'Page introuvable',
  'notFound.body': 'Cette adresse ne correspond à aucune page de 226os.',
  'notFound.back': 'Retour à l’accueil',
  'state.loading': 'Chargement…',
  'state.retry': 'Réessayer',
  'state.reference': 'Référence : {id}',
  'state.demo': 'Démo',
  'state.demo.hint': 'Donnée d’exemple, pas une activité réelle.',
  'error.title': 'Impossible de charger le contenu',
  'error.NETWORK': 'Le serveur ne répond pas. Vérifiez votre connexion, puis réessayez.',
  'error.RATE_LIMITED': 'Trop de demandes en peu de temps. Réessayez dans {seconds} s.',
  'error.RATE_LIMITED.now': 'Trop de demandes en peu de temps. Patientez un instant.',
  'error.NOT_FOUND': 'Ce contenu n’existe pas ou n’est plus disponible.',
  'error.UNAUTHENTICATED': 'Connectez-vous pour continuer.',
  'error.FORBIDDEN': 'Vous n’avez pas accès à cette action.',
  'error.EMAIL_NOT_VERIFIED': 'Confirmez d’abord votre adresse e-mail.',
  'error.CSRF_FAILED': 'Votre session a changé. Rechargez la page, puis réessayez.',
  'error.VALIDATION_ERROR': 'Certains champs sont à corriger.',
  'error.SERVICE_UNAVAILABLE': 'Ce service est momentanément indisponible. Réessayez plus tard.',
  'error.generic': 'Une erreur est survenue. Réessayez dans un instant.',
} as const;

const coreEn: Record<keyof typeof core, string> = {
  'app.name': '226os',
  'nav.label': 'Main navigation',
  'nav.home': 'Home',
  'nav.settings': 'Settings',
  'nav.learn': 'Learn',
  'nav.connect': 'Connect',
  'nav.solve': 'Solve',
  'nav.discover': 'Discover',
  'nav.build': 'Build',
  'nav.search': 'Search',
  'nav.soon': 'Soon',
  'nav.section.grow': 'Grow',
  'nav.section.pillars': 'Pillars',
  'nav.open': 'Open navigation',
  'nav.top': 'Site bar',
  'account.signIn': 'Sign in',
  'account.signUp': 'Create an account',
  'account.signOut': 'Sign out',
  'account.settings': 'Settings',
  'account.menu': '{name}’s account',
  'account.more': 'More options',
  'account.language': 'Language',
  'account.signOutFailed': 'Signing out failed. Try again.',
  'home.title': 'Welcome to 226os',
  'home.intro':
    'The technical mutual-help platform is under construction. For now, this page only checks that the API responds.',
  'health.title': 'API status',
  'health.loading': 'Checking the API…',
  'health.ok': 'API is up',
  'health.status': 'Status: {status}',
  'health.error': 'The API is not responding.',
  'health.errorHint': 'Check that the server is running, then try again.',
  'health.retry': 'Check again',
  'settings.title': 'Settings',
  'settings.theme.label': 'Theme',
  'settings.theme.light': 'Light',
  'settings.theme.dark': 'Dark',
  'settings.theme.system': 'System',
  'settings.language.label': 'Language',
  'settings.language.fr': 'Français',
  'settings.language.en': 'English',
  'settings.visibility.title': 'Profile visibility',
  'settings.visibility.label': 'Who can see your profile?',
  'settings.visibility.bylines':
    'Either way, your name and username still appear on what you publish.',
  'settings.visibility.public': 'Public',
  'settings.visibility.public.hint': 'Everyone, even without an account.',
  'settings.visibility.members': 'Members',
  'settings.visibility.members.hint': 'Signed-in members only.',
  'settings.visibility.private': 'Private',
  'settings.visibility.private.hint': 'Only you.',
  'settings.visibility.save': 'Save visibility',
  'settings.visibility.saved': 'Visibility saved.',
  'settings.visibility.failed': 'The visibility was not saved',
  'settings.visibility.soon': 'Coming soon: this setting turns on as soon as profiles are live.',
  'settings.visibility.signIn': 'Sign in to choose who sees your profile',
  'form.name.label': 'Display name',
  'form.name.description': 'Shown in the header and next to what you publish.',
  'form.name.required': 'Enter a display name.',
  'form.name.tooShort': 'The name must be at least {min} characters long.',
  'form.name.tooLong': 'The name must be {max} characters or fewer.',
  'form.name.failed': 'The name was not saved',
  'settings.name.signIn': 'Sign in to change your name',
  'form.save': 'Save',
  'form.saved': 'Name saved.',
  'notFound.title': 'Page not found',
  'notFound.body': 'This address does not match any 226os page.',
  'notFound.back': 'Back to home',
  'state.loading': 'Loading…',
  'state.retry': 'Try again',
  'state.reference': 'Reference: {id}',
  'state.demo': 'Demo',
  'state.demo.hint': 'Sample data, not real activity.',
  'error.title': 'The content could not be loaded',
  'error.NETWORK': 'The server is not responding. Check your connection, then try again.',
  'error.RATE_LIMITED': 'Too many requests in a short time. Try again in {seconds}s.',
  'error.RATE_LIMITED.now': 'Too many requests in a short time. Wait a moment.',
  'error.NOT_FOUND': 'This content does not exist or is no longer available.',
  'error.UNAUTHENTICATED': 'Sign in to continue.',
  'error.FORBIDDEN': 'You do not have access to this action.',
  'error.EMAIL_NOT_VERIFIED': 'Confirm your email address first.',
  'error.CSRF_FAILED': 'Your session changed. Reload the page, then try again.',
  'error.VALIDATION_ERROR': 'Some fields need fixing.',
  'error.SERVICE_UNAVAILABLE': 'This service is temporarily unavailable. Try again later.',
  'error.generic': 'Something went wrong. Try again in a moment.',
};

const fr = {
  ...core,
  ...identity.fr,
  ...discover.fr,
  ...search.fr,
  ...landing.fr,
  ...home.fr,
  ...solve.fr,
};

export type MessageKey = keyof typeof fr;

const en: Record<MessageKey, string> = {
  ...coreEn,
  ...identity.en,
  ...discover.en,
  ...search.en,
  ...landing.en,
  ...home.en,
  ...solve.en,
};

export type Catalogs = Record<Locale, Partial<Record<MessageKey, string>>>;

/** The French catalogues before merging, so a test can prove no key is overridden. */
export const frenchParts: readonly Record<string, string>[] = [
  core,
  identity.fr,
  discover.fr,
  search.fr,
  landing.fr,
  home.fr,
  solve.fr,
];

export const catalogs: Catalogs = { fr, en };
