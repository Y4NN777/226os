import { fireEvent, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { renderWithProviders, setLanguages } from '../test/render';
import { ProfilePhoto } from './ProfilePhoto';

const base = 'https://avatars.githubusercontent.com';

function photo(container: HTMLElement) {
  return container.querySelector('img');
}

describe('ProfilePhoto', () => {
  it('draws the GitHub avatar at twice the pixel size, lazily and without referrer', () => {
    setLanguages(['fr-FR']);
    const { container } = renderWithProviders(
      <ProfilePhoto name="Awa Traoré" github="awa-dev" size="lg" />,
    );
    const img = photo(container);
    expect(img).not.toBeNull();
    expect(img).toHaveAttribute('src', `${base}/awa-dev?s=96`);
    expect(img).toHaveAttribute('loading', 'lazy');
    expect(img).toHaveAttribute('decoding', 'async');
    expect(img).toHaveAttribute('referrerpolicy', 'no-referrer');
    expect(img).toHaveAttribute('width', '48');
    expect(img).toHaveAttribute('height', '48');
    expect(img).toHaveStyle({ borderRadius: '50%', objectFit: 'cover' });
  });

  it.each([
    ['xs', 20],
    ['sm', 24],
    ['md', 36],
    ['lg', 48],
    ['xl', 72],
  ] as const)('maps size %s to %i px', (size, px) => {
    const { container } = renderWithProviders(
      <ProfilePhoto name="Awa Traoré" github="awa-dev" size={size} />,
    );
    const img = photo(container);
    expect(img).toHaveAttribute('width', String(px));
    expect(img).toHaveAttribute('height', String(px));
    expect(img).toHaveAttribute('src', `${base}/awa-dev?s=${String(px * 2)}`);
  });

  it('is hidden from assistive technology when decorative', () => {
    const { container } = renderWithProviders(<ProfilePhoto name="Awa Traoré" github="awa-dev" />);
    expect(photo(container)).toHaveAttribute('alt', '');
    expect(screen.queryByRole('img', { name: 'Awa Traoré' })).not.toBeInTheDocument();
  });

  it('is an image named by the person when not decorative', () => {
    renderWithProviders(<ProfilePhoto name="Awa Traoré" github="awa-dev" isDecorative={false} />);
    expect(screen.getByRole('img', { name: 'Awa Traoré' })).toHaveAttribute(
      'src',
      `${base}/awa-dev?s=72`,
    );
  });

  it('falls back to the initials when the picture fails to load', () => {
    const { container } = renderWithProviders(
      <ProfilePhoto name="Awa Traoré" github="awa-dev" isDecorative={false} />,
    );
    fireEvent.error(photo(container) as HTMLImageElement);
    expect(photo(container)).toBeNull();
    expect(screen.getByRole('img', { name: 'Awa Traoré' })).toHaveClass('astryx-avatar');
  });

  it.each([null, undefined, '', 'bad/../x', '-x', 'x-', 'a b', 'a'.repeat(40)])(
    'shows the initials and no picture for github=%s',
    (github) => {
      const { container } = renderWithProviders(
        <ProfilePhoto name="Awa Traoré" github={github} isDecorative={false} />,
      );
      expect(photo(container)).toBeNull();
      expect(screen.getByRole('img', { name: 'Awa Traoré' })).toHaveClass('astryx-avatar');
    },
  );
});
