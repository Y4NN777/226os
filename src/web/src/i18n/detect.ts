import { defaultLocale, isLocale, type Locale } from './messages';

/** First supported language in the browser preference order; French otherwise. */
export function detectLocale(languages: readonly string[] | undefined): Locale {
  for (const tag of languages ?? []) {
    const primary = tag.toLowerCase().split('-')[0] ?? '';
    if (isLocale(primary)) return primary;
  }
  return defaultLocale;
}
