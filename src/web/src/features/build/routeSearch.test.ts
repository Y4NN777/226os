import { describe, expect, it } from 'vitest';
import { validateBuildSearch } from './routeSearch';

describe('validateBuildSearch', () => {
  it('keeps the values the API accepts', () => {
    expect(
      validateBuildSearch({
        q: ' sms ',
        open: true,
        stage: 'prototype',
        country: 'BF',
        sort: 'needs',
      }),
    ).toEqual({ q: 'sms', open: true, stage: 'prototype', country: 'BF', sort: 'needs' });
  });

  it('reads the flag as text too, as a hand-typed address gives it', () => {
    expect(validateBuildSearch({ open: 'true' })).toEqual({ open: true });
  });

  it('drops unknown stages, sorts and country codes', () => {
    expect(
      validateBuildSearch({ stage: 'done', country: 'bf', sort: 'name', open: false, q: '  ' }),
    ).toEqual({});
  });

  it('leaves the default sort out of the address', () => {
    expect(validateBuildSearch({ sort: 'updated' })).toEqual({});
  });

  it('keeps the other sorts', () => {
    expect(validateBuildSearch({ sort: 'created' })).toEqual({ sort: 'created' });
  });
});
