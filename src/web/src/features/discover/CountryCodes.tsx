import { VisuallyHidden } from '@astryxdesign/core/VisuallyHidden';
import { useI18n } from '../../i18n/I18nProvider';
import { countryName } from './present';

/**
 * Country codes as text, never flags (prototype `.country`). Each code is decorative and carries
 * its localized name for assistive technology; `withNames` also shows the name (detail page).
 */
export function CountryCodes({
  codes,
  withNames = false,
}: {
  codes: readonly string[];
  withNames?: boolean;
}) {
  const { locale } = useI18n();
  return (
    <span className="discover-countries">
      {codes.map((code) => (
        <span key={code} className="discover-country">
          <span className="discover-country-code" aria-hidden="true">
            {code}
          </span>
          {withNames ? (
            <span>{countryName(code, locale)}</span>
          ) : (
            <VisuallyHidden>{countryName(code, locale)}</VisuallyHidden>
          )}
        </span>
      ))}
    </span>
  );
}
