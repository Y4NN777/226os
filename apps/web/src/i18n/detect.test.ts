import { describe, expect, it } from 'vitest';
import { detectLocale } from './detect';

describe('detectLocale', () => {
  it.each([
    [['fr-FR', 'en-US'], 'fr'],
    [['en-GB'], 'en'],
    [['EN'], 'en'],
    [['de-DE', 'en-US'], 'en'],
    [['de-DE', 'es'], 'fr'],
    [[], 'fr'],
    [undefined, 'fr'],
  ] as const)('maps %j to %s (falls back to French)', (languages, expected) => {
    expect(detectLocale(languages)).toBe(expected);
  });
});
