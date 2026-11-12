export type Locale = 'fr' | 'en';

export const locales = ['fr', 'en'] as const satisfies readonly Locale[];

export function isLocale(value: string): value is Locale {
  return (locales as readonly string[]).includes(value);
}

export const defaultLocale: Locale = 'fr';

// French typography: a non-breaking space (U+00A0) precedes : ; ? ! and
// typographic apostrophes/ellipsis are used. Enforced by translate.test.ts.
const fr = {
  'app.name': '226os',
  'nav.label': 'Navigation principale',
  'nav.home': 'Accueil',
  'nav.settings': 'Réglages',
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
  'form.name.label': 'Nom d’affichage',
  'form.name.description': 'Enregistré uniquement sur cet appareil, jamais envoyé au serveur.',
  'form.name.required': 'Saisissez un nom d’affichage.',
  'form.name.tooShort': 'Le nom doit contenir au moins {min} caractères.',
  'form.save': 'Enregistrer',
  'form.saved': 'Nom enregistré sur cet appareil.',
  'notFound.title': 'Page introuvable',
  'notFound.body': 'Cette adresse ne correspond à aucune page de 226os.',
  'notFound.back': 'Retour à l’accueil',
} as const;

export type MessageKey = keyof typeof fr;

const en: Record<MessageKey, string> = {
  'app.name': '226os',
  'nav.label': 'Main navigation',
  'nav.home': 'Home',
  'nav.settings': 'Settings',
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
  'form.name.label': 'Display name',
  'form.name.description': 'Saved on this device only, never sent to the server.',
  'form.name.required': 'Enter a display name.',
  'form.name.tooShort': 'The name must be at least {min} characters long.',
  'form.save': 'Save',
  'form.saved': 'Name saved on this device.',
  'notFound.title': 'Page not found',
  'notFound.body': 'This address does not match any 226os page.',
  'notFound.back': 'Back to home',
};

export type Catalogs = Record<Locale, Partial<Record<MessageKey, string>>>;

export const catalogs: Catalogs = { fr, en };
