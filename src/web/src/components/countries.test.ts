import { describe, expect, it } from 'vitest';
import { WEST_AFRICA } from '../features/discover/region';
import { AFRICA_COUNTRIES, countryOptions } from './countries';

describe('AFRICA_COUNTRIES', () => {
  it('holds the 54 codes of the database, once each, West Africa first', () => {
    expect(AFRICA_COUNTRIES).toHaveLength(54);
    expect(new Set(AFRICA_COUNTRIES).size).toBe(54);
    expect(AFRICA_COUNTRIES.slice(0, WEST_AFRICA.length)).toEqual([...WEST_AFRICA]);
    expect(AFRICA_COUNTRIES).toEqual(expect.arrayContaining(['KE', 'ZA', 'EG', 'CD', 'ST']));
    expect(AFRICA_COUNTRIES).not.toContain('EH');
  });
});

describe('countryOptions', () => {
  it('lists West Africa by name, then the rest of the continent by name', () => {
    const options = countryOptions([], 'fr');
    expect(options).toHaveLength(54);
    const west = options.slice(0, WEST_AFRICA.length).map((option) => option.value);
    expect([...west].sort()).toEqual([...WEST_AFRICA].sort());
    const rest = options.slice(WEST_AFRICA.length);
    expect(rest[0]?.label.localeCompare(rest[1]?.label ?? '', 'fr')).toBeLessThan(0);
    expect(options.find((option) => option.value === 'KE')?.label).toBe('Kenya');
  });

  it('keeps a code outside the list at the end rather than dropping it', () => {
    expect(countryOptions(['FR'], 'en').at(-1)?.value).toBe('FR');
  });
});
