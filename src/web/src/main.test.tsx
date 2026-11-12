import { screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { jsonResponse } from './test/render';

describe('main bootstrap', () => {
  it('mounts the app into #root', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(() => Promise.resolve(jsonResponse({ status: 'ok' }))),
    );
    document.body.innerHTML = '<div id="root"></div>';
    await import('./main');
    expect(await screen.findByRole('heading', { level: 1 })).toBeInTheDocument();
  });
});
