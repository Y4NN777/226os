// landing catalogue, merged by src/i18n/messages.ts. Keys start with 'landing.'.
// French typography rules are enforced by src/i18n/translate.test.ts.
export const fr = {
  'landing.title': 'Bienvenue sur 226os',
} as const;

export const en: Record<keyof typeof fr, string> = {
  'landing.title': 'Welcome to 226os',
};
