import { describe, expect, it } from 'vitest';
import { isSearchable, normalizeQuery } from './searchQuery';

describe('normalizeQuery', () => {
  it('trims and collapses nothing else', () => {
    expect(normalizeQuery('  orange  money ')).toBe('orange  money');
  });

  it('caps at 100 characters, counting code points like the server', () => {
    expect(Array.from(normalizeQuery('a'.repeat(150)))).toHaveLength(100);
    const emoji = normalizeQuery('😀'.repeat(150));
    expect(Array.from(emoji)).toHaveLength(100);
    expect(emoji.endsWith('😀')).toBe(true);
  });
});

describe('isSearchable', () => {
  it('needs at least two characters (contract minLength)', () => {
    expect(isSearchable('')).toBe(false);
    expect(isSearchable('a')).toBe(false);
    expect(isSearchable('😀')).toBe(false);
    expect(isSearchable('ab')).toBe(true);
  });
});
