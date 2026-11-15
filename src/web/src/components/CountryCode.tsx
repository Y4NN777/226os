import { VisuallyHidden } from '@astryxdesign/core/VisuallyHidden';
import { useI18n } from '../i18n/I18nProvider';

/** Localized region name from the platform, never a hand-kept list; the code itself as fallback. */
export function countryName(code: string, locale: string): string {
  try {
    return new Intl.DisplayNames([locale], { type: 'region' }).of(code) ?? code;
  } catch {
    return code;
  }
}

/**
 * A country as its ISO code in a small mono chip, never a flag (prototype `.country`). Screen
 * readers hear the full name instead of the letters; `withName` also writes the name next to the
 * chip (read once). Without `name`, the name comes from the platform in the UI language.
 */
export function CountryCode({
  code,
  name,
  withName = false,
}: {
  code: string;
  name?: string;
  withName?: boolean;
}) {
  const { locale } = useI18n();
  const label = name ?? countryName(code, locale);
  return (
    <span className="ui-country">
      <span className="ui-country-code" aria-hidden="true">
        {code}
      </span>
      {withName ? <span>{label}</span> : <VisuallyHidden>{label}</VisuallyHidden>}
    </span>
  );
}
