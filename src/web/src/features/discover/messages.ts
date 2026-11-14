// discover catalogue, merged by src/i18n/messages.ts. Keys start with 'discover.'.
// French typography rules are enforced by src/i18n/translate.test.ts.
export const fr = {
  'discover.title': 'Découvrir',
} as const;

export const en: Record<keyof typeof fr, string> = {
  'discover.title': 'Discover',
};
