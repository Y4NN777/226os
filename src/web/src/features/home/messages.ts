// home catalogue, merged by src/i18n/messages.ts. Keys start with 'feed.'.
// French typography rules are enforced by src/i18n/translate.test.ts.
export const fr = {
  'feed.title': 'Accueil',
} as const;

export const en: Record<keyof typeof fr, string> = {
  'feed.title': 'Home',
};
