import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { ProjectThumb, repoPreviewUrl } from './ProjectThumb';

describe('repoPreviewUrl', () => {
  it('derives the GitHub social preview from a repository URL', () => {
    expect(repoPreviewUrl('https://github.com/kora-dev/kora-sdk')).toBe(
      'https://opengraph.githubassets.com/1/kora-dev/kora-sdk',
    );
  });

  it.each([
    ['another host', 'https://gitlab.com/kora-dev/kora-sdk'],
    ['plain http', 'http://github.com/kora-dev/kora-sdk'],
    ['an owner only', 'https://github.com/kora-dev'],
    ['a deeper path', 'https://github.com/kora-dev/kora-sdk/tree/main'],
    ['a query string', 'https://github.com/kora-dev/kora-sdk?tab=readme'],
    ['an encoded slash', 'https://github.com/kora-dev/kora%2Fsdk'],
  ])('returns null for %s', (_label, url) => {
    expect(repoPreviewUrl(url)).toBeNull();
  });
});

describe('ProjectThumb', () => {
  it('shows the repository preview as a decorative, lazy image', () => {
    const { container } = render(
      <ProjectThumb name="Kora SDK" githubUrl="https://github.com/kora-dev/kora-sdk" />,
    );
    const img = container.querySelector('img');
    expect(img).not.toBeNull();
    expect(img?.getAttribute('src')).toBe('https://opengraph.githubassets.com/1/kora-dev/kora-sdk');
    expect(img?.getAttribute('alt')).toBe('');
    expect(img?.getAttribute('loading')).toBe('lazy');
    expect(img?.getAttribute('decoding')).toBe('async');
    expect(img?.getAttribute('referrerpolicy')).toBe('no-referrer');
    // Decorative: the project name is written next to it, so it adds nothing for assistive technology.
    expect(screen.queryByRole('img')).toBeNull();
  });

  it('falls back to the monogram tile when the preview cannot load', () => {
    const { container } = render(
      <ProjectThumb name="Kora SDK" githubUrl="https://github.com/kora-dev/kora-sdk" />,
    );
    const img = container.querySelector('img');
    if (!img) throw new Error('no preview image');
    fireEvent.error(img);
    expect(container.querySelector('img')).toBeNull();
    expect(container.textContent).toBe('KS');
  });

  it('shows the monogram tile when the URL is not a GitHub repository', () => {
    const { container } = render(
      <ProjectThumb name="Kora SDK" githubUrl="https://gitlab.com/kora-dev/kora-sdk" />,
    );
    expect(container.querySelector('img')).toBeNull();
    expect(container.textContent).toBe('KS');
  });

  it('uses the large size on the project page', () => {
    const { container } = render(
      <ProjectThumb name="Kora SDK" githubUrl="https://github.com/kora-dev/kora-sdk" size="lg" />,
    );
    expect(container.querySelector('img')?.className).toContain('build-thumb-lg');
  });
});

describe('ProjectThumb, review fixes', () => {
  it('tries the preview again when the repository URL changes after a failure', () => {
    const { container, rerender } = render(
      <ProjectThumb name="Kora SDK" githubUrl="https://github.com/kora-dev/broken" />,
    );
    const first = container.querySelector('img');
    if (first) fireEvent.error(first);
    expect(container.querySelector('img')).toBeNull();
    rerender(<ProjectThumb name="Kora SDK" githubUrl="https://github.com/kora-dev/kora-sdk" />);
    expect(container.querySelector('img')?.getAttribute('src')).toBe(
      'https://opengraph.githubassets.com/1/kora-dev/kora-sdk',
    );
  });

  it('renders nothing instead of a second tile when the fallback is turned off', () => {
    const { container } = render(
      <ProjectThumb name="Kora SDK" githubUrl="https://gitlab.com/kora-dev/kora-sdk" fallback={false} />,
    );
    expect(container.firstChild).toBeNull();
  });
});
