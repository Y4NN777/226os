import { describe, expect, it } from 'vitest';
import { WEST_AFRICA, inWestAfrica, listQuery } from './region';

describe('WEST_AFRICA', () => {
  it('lists the 16 West African ISO codes', () => {
    expect(WEST_AFRICA).toHaveLength(16);
    expect(WEST_AFRICA).toContain('SN');
    expect(WEST_AFRICA).not.toContain('KE');
  });
});

describe('inWestAfrica', () => {
  it('is true when at least one country is West African', () => {
    expect(inWestAfrica(['KE', 'GH'])).toBe(true);
  });
  it('is false for other regions and for an empty list', () => {
    expect(inWestAfrica(['KE', 'ZA'])).toBe(false);
    expect(inWestAfrica([])).toBe(false);
  });
});

describe('listQuery', () => {
  it('asks for 50 items by default and sends no country', () => {
    expect(listQuery({})).toEqual({ limit: 50 });
  });
  it('forwards q, type and category', () => {
    expect(listQuery({ q: 'momo', type: 'api', category: 'payments' })).toEqual({
      limit: 50,
      q: 'momo',
      type: 'api',
      category: 'payments',
    });
  });
  it('sends one country to the server when picked', () => {
    expect(listQuery({ country: 'SN' })).toEqual({ limit: 50, country: 'SN' });
  });
  it('sends no country for the whole continent', () => {
    expect(listQuery({ region: 'all' })).toEqual({ limit: 50 });
  });
});
