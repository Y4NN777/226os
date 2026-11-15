import { describe, expect, it } from 'vitest';
import { initials, tone } from './ResourceThumb';

describe('ResourceThumb', () => {
  it('takes the first letter of the first two words', () => {
    expect(initials('Kalimba Pay API')).toBe('KP');
    expect(initials('geojson-lint')).toBe('G');
    expect(initials('Éclair SMS')).toBe('ÉS');
  });

  it('ignores a trailing parenthesis such as « (démo) »', () => {
    expect(initials('Nimba Maps Studio (démo)')).toBe('NM');
    expect(tone('Nimba Maps Studio (démo)')).toBe(tone('Nimba Maps Studio'));
  });

  it('falls back to a placeholder when the name has no letter or digit', () => {
    expect(initials('—')).toBe('?');
  });

  it('keeps the tone within the five token pairs', () => {
    for (const name of ['A', 'Wave Pay', 'SDK SMS Baobab', '']) {
      expect(tone(name)).toBeGreaterThanOrEqual(0);
      expect(tone(name)).toBeLessThan(5);
    }
  });
});
