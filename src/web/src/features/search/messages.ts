// search catalogue, merged by src/i18n/messages.ts. Keys start with 'search.'.
// French typography rules are enforced by src/i18n/translate.test.ts.
export const fr = {
  'search.title': 'Recherche',
} as const;

export const en: Record<keyof typeof fr, string> = {
  'search.title': 'Search',
};
