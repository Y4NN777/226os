import { describe, expect, it } from 'vitest';
import { githubProfileUrl } from './github';

describe('githubProfileUrl', () => {
  it('builds the URL for a valid login', () => {
    expect(githubProfileUrl('awa-dev')).toBe('https://github.com/awa-dev');
    expect(githubProfileUrl('a'.repeat(39))).toBe(`https://github.com/${'a'.repeat(39)}`);
  });

  it.each([
    ['a slash', 'awa/dev'],
    ['a question mark', 'awa?x=1'],
    ['dot dot', '..'],
    ['a space', 'awa dev'],
    ['a javascript scheme', 'javascript:alert(1)'],
    ['an empty string', ''],
    ['40 characters', 'a'.repeat(40)],
    ['a trailing newline', 'awa\n'],
  ])('returns null for %s', (_name, login) => {
    expect(githubProfileUrl(login)).toBeNull();
  });
});
