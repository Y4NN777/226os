import { describe, expect, it } from 'vitest';
import { render } from '@testing-library/react';
import { Theme } from '@astryxdesign/core/theme';
import { theme226os, tokens } from './theme226os';

function luminance(hex: string): number {
  const channels = [1, 3, 5].map((i) => {
    const c = parseInt(hex.slice(i, i + 2), 16) / 255;
    return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
  });
  return 0.2126 * (channels[0] ?? 0) + 0.7152 * (channels[1] ?? 0) + 0.0722 * (channels[2] ?? 0);
}

function contrast(a: string, b: string): number {
  const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return ((hi ?? 0) + 0.05) / ((lo ?? 0) + 0.05);
}

function pair(token: string): [string, string] {
  const value = tokens[token];
  if (!Array.isArray(value)) throw new Error(`${token} must be a [light, dark] tuple`);
  return value;
}

describe('Graphite & Lime theme tokens', () => {
  it('is a defined Astryx theme named 226os', () => {
    expect(theme226os).toMatchObject({ name: '226os' });
  });

  it('maps the DESIGN.md palette onto Astryx slots for both modes', () => {
    expect(pair('--color-accent')).toEqual(['#4D721D', '#BDE874']);
    expect(pair('--color-background-body')).toEqual(['#F6F7F4', '#111710']);
    expect(pair('--color-background-surface')).toEqual(['#FFFFFF', '#1B241A']);
    expect(pair('--color-text-primary')).toEqual(['#1D2420', '#F4F7EF']);
  });

  it('keeps the focus ring distinct from the brand green', () => {
    expect(pair('--focus-outline-color')).toEqual(['#275DCA', '#9CC4FF']);
    expect(pair('--focus-outline-color')).not.toEqual(pair('--color-accent'));
  });

  it.each([
    ['--color-on-accent', '--color-accent', 4.5],
    ['--color-text-primary', '--color-background-body', 4.5],
    ['--color-text-primary', '--color-background-surface', 4.5],
    ['--color-text-secondary', '--color-background-surface', 4.5],
    ['--color-text-accent', '--color-background-surface', 4.5],
    ['--color-text-error', '--color-background-surface', 4.5],
    ['--color-text-success', '--color-background-surface', 4.5],
    ['--color-text-warning', '--color-background-surface', 4.5],
    ['--focus-outline-color', '--color-background-surface', 3],
  ])('%s on %s meets WCAG contrast %s:1 in light and dark', (fg, bg, min) => {
    const [fgLight, fgDark] = pair(fg);
    const [bgLight, bgDark] = pair(bg);
    expect(contrast(fgLight, bgLight)).toBeGreaterThanOrEqual(min);
    expect(contrast(fgDark, bgDark)).toBeGreaterThanOrEqual(min);
  });

  it('sets the Astryx body size to 16px (core default is 14px)', () => {
    render(<Theme theme={theme226os}>content</Theme>);
    const css = [...document.querySelectorAll('style')].map((el) => el.textContent).join('\n');
    expect(css).toMatch(/--_scaled-font-size-base:\s*min\(calc\(1rem /);
  });

  it('uses system font stacks only (no remote font)', () => {
    for (const token of ['--font-family-body', '--font-family-heading', '--font-family-code']) {
      expect(tokens[token]).toEqual(expect.stringMatching(/^[^@]*$/));
      expect(String(tokens[token])).not.toMatch(/url\(|https?:/);
    }
  });
});
