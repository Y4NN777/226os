import { describe, expect, it } from 'vitest';
import { WEST_AFRICA, listQuery } from './region';

describe('WEST_AFRICA', () => {
  it('lists the 16 West African ISO codes', () => {
    expect(WEST_AFRICA).toHaveLength(16);
    expect(WEST_AFRICA).toContain('SN');
    expect(WEST_AFRICA).not.toContain('KE');
  });
});

describe('listQuery', () => {
  it('asks for 50 items of West Africa by default', () => {
    expect(listQuery({})).toEqual({ limit: 50, country: [...WEST_AFRICA] });
  });
  it('forwards q, type and category', () => {
    expect(listQuery({ q: 'momo', type: 'api', category: 'payments' })).toEqual({
      limit: 50,
      q: 'momo',
      type: 'api',
      category: 'payments',
      country: [...WEST_AFRICA],
    });
  });
  it('sends only the picked country', () => {
    expect(listQuery({ country: 'SN' })).toEqual({ limit: 50, country: ['SN'] });
  });
  it('sends no country for the whole continent', () => {
    expect(listQuery({ region: 'all' })).toEqual({ limit: 50 });
  });
});
