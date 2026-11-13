// identity catalogue, merged by src/i18n/messages.ts. Keys start with 'identity.'.
// French typography rules are enforced by src/i18n/translate.test.ts.
export const fr = {
  'identity.signIn.title': 'Connexion',
} as const;

export const en: Record<keyof typeof fr, string> = {
  'identity.signIn.title': 'Sign in',
};
