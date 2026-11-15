import { useI18n } from '../i18n/I18nProvider';

function languageName(code: string, locale: string): string {
  try {
    return new Intl.DisplayNames([locale], { type: 'language' }).of(code) ?? code;
  } catch {
    return code;
  }
}

/**
 * The original language of a piece of user content ("FR", "EN") as a small mono code, like the
 * country chip; the language name is its hint. The page also sets `lang` on the content itself.
 */
export function LangCode({ lang }: { lang: string }) {
  const { t, locale } = useI18n();
  return (
    <span className="ui-lang" title={t('lang.hint', { lang: languageName(lang, locale) })}>
      {lang.toUpperCase()}
    </span>
  );
}
