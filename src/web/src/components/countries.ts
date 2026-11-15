import { WEST_AFRICA } from '../features/discover/region';
import type { Locale } from '../i18n/messages';
import { countryName } from './CountryCode';

/** The other African states of the `countries` table (db/migrations/00002_reference_taxonomy.sql). */
const OTHER_AFRICA = [
  'AO',
  'BI',
  'BW',
  'CD',
  'CF',
  'CG',
  'CM',
  'DJ',
  'DZ',
  'EG',
  'ER',
  'ET',
  'GA',
  'GQ',
  'KE',
  'KM',
  'LS',
  'LY',
  'MA',
  'MG',
  'MU',
  'MW',
  'MZ',
  'NA',
  'RW',
  'SC',
  'SD',
  'SO',
  'SS',
  'ST',
  'SZ',
  'TD',
  'TN',
  'TZ',
  'UG',
  'ZA',
  'ZM',
  'ZW',
] as const;

/** The 54 codes the API accepts: West Africa first, then the rest of the continent. */
export const AFRICA_COUNTRIES: readonly string[] = [...WEST_AFRICA, ...OTHER_AFRICA];

/**
 * Select options for any country field or filter: West Africa by name, then the other states by
 * name, then any `extra` code outside the table (so a stored value is never silently dropped).
 */
export function countryOptions(extra: Iterable<string>, locale: Locale) {
  const byName = (codes: Iterable<string>) =>
    [...new Set(codes)]
      .map((code) => ({ value: code, label: countryName(code, locale) }))
      .sort((a, b) => a.label.localeCompare(b.label, locale));
  const west: readonly string[] = WEST_AFRICA;
  const rest = AFRICA_COUNTRIES.filter((code) => !west.includes(code));
  const known: readonly string[] = AFRICA_COUNTRIES;
  return [
    ...byName(west),
    ...byName(rest),
    ...byName([...extra].filter((code) => !known.includes(code))),
  ];
}
