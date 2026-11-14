import { describe, expect, it } from 'vitest';
import { catalogs, frenchParts, locales, type MessageKey } from './messages';
import { translate } from './translate';

describe('translate', () => {
  it('returns the message of the requested locale', () => {
    expect(translate('en', 'nav.settings')).toBe('Settings');
    expect(translate('fr', 'nav.settings')).toBe('Réglages');
  });

  it('interpolates named parameters', () => {
    expect(translate('en', 'form.name.tooShort', { min: 2 })).toBe(
      'The name must be at least 2 characters long.',
    );
  });

  it('falls back to French when the locale lacks the key, then to the key itself', () => {
    const partial = { fr: { 'nav.home': 'Accueil' }, en: {} };
    expect(translate('en', 'nav.home', undefined, partial)).toBe('Accueil');
    expect(translate('en', 'unknown.key' as MessageKey, undefined, partial)).toBe('unknown.key');
  });
});

describe('catalogs', () => {
  it('expose every locale with exactly the same keys', () => {
    expect(locales).toEqual(['fr', 'en']);
    expect(Object.keys(catalogs.en).sort()).toEqual(Object.keys(catalogs.fr).sort());
  });

  it('merges feature catalogues without any key overriding another', () => {
    const total = frenchParts.reduce((sum, part) => sum + Object.keys(part).length, 0);
    expect(Object.keys(catalogs.fr)).toHaveLength(total);
  });

  it('has no empty message', () => {
    for (const locale of locales) {
      for (const [key, value] of Object.entries(catalogs[locale])) {
        expect(value, `${locale}:${key}`).not.toBe('');
      }
    }
  });

  it('puts a non-breaking space before : ; ? ! in French and none in English', () => {
    for (const [key, value] of Object.entries(catalogs.fr)) {
      expect(value, `fr:${key}`).not.toMatch(/ [:;?!]/); // plain space
      expect(value, `fr:${key}`).not.toMatch(/[^\u00A0\u202F][:;?!](?!\w)(?=\s|$)(?<![\d])/u);
    }
    for (const [key, value] of Object.entries(catalogs.en)) {
      expect(value, `en:${key}`).not.toMatch(/[\u00A0\u202F][:;?!]/);
    }
  });

  it('uses typographic apostrophes and ellipsis in French', () => {
    for (const [key, value] of Object.entries(catalogs.fr)) {
      expect(value, `fr:${key}`).not.toMatch(/'|\.\.\./);
    }
  });
});
