import { countryName } from '../../components/CountryCode';

// Countries named with « en » in French (feminine or starting with a vowel); every other one takes « au ».
const FR_EN = new Set(['CI', 'GN']);

/** "au Burkina Faso", "en Côte d’Ivoire" in French; "in Ghana" in English. */
export function inCountry(code: string, locale: string): string {
  const name = countryName(code, locale);
  if (!locale.startsWith('fr')) return `in ${name}`;
  return `${FR_EN.has(code) ? 'en' : 'au'} ${name}`;
}
