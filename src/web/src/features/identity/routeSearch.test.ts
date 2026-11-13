import { describe, expect, it } from 'vitest';
import { safeNext, validateSignInSearch } from './routeSearch';

describe('safeNext', () => {
  it.each([
    ['/discover', '/discover'],
    ['/discover?q=momo', '/discover?q=momo'],
    ['https://evil.example', undefined],
    ['//evil.example', undefined],
    ['/\\evil.example', undefined],
    ['javascript:alert(1)', undefined],
    [42, undefined],
  ])('%s -> %s', (input, expected) => {
    expect(safeNext(input)).toBe(expected);
  });
});

describe('validateSignInSearch', () => {
  it('drops an unsafe next and keeps a safe one', () => {
    expect(validateSignInSearch({ next: '//x' })).toEqual({});
    expect(validateSignInSearch({ next: '/recherche' })).toEqual({ next: '/recherche' });
  });
});
