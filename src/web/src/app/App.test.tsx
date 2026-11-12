import { createMemoryHistory } from '@tanstack/react-router';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { jsonResponse, setLanguages } from '../test/render';
import { App } from './App';
import { createAppRouter } from './router';

function renderAt(path: string) {
  vi.stubGlobal(
    'fetch',
    vi.fn(() => Promise.resolve(jsonResponse({ status: 'ok' }))),
  );
  const router = createAppRouter({ history: createMemoryHistory({ initialEntries: [path] }) });
  return render(<App router={router} />);
}

describe('App routing', () => {
  it('renders the shell and the home route with the API status', async () => {
    setLanguages(['fr-FR']);
    renderAt('/');
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Bienvenue sur 226os' }),
    ).toBeInTheDocument();
    expect(screen.getByRole('navigation', { name: 'Navigation principale' })).toBeInTheDocument();
    expect(await screen.findByText('API opérationnelle')).toBeInTheDocument();
  });

  it('navigates to settings through the nav link without a full reload', async () => {
    setLanguages(['en-US']);
    renderAt('/');
    await userEvent.click(await screen.findByRole('link', { name: 'Settings' }));
    expect(await screen.findByRole('heading', { level: 1, name: 'Settings' })).toBeInTheDocument();
  });

  it('shows a localized not-found page with a way back', async () => {
    setLanguages(['fr-FR']);
    renderAt('/nope');
    expect(await screen.findByRole('heading', { name: 'Page introuvable' })).toBeInTheDocument();
    await userEvent.click(screen.getByRole('link', { name: 'Retour à l’accueil' }));
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Bienvenue sur 226os' }),
    ).toBeInTheDocument();
  });

  it('creates its own browser router when none is injected', async () => {
    setLanguages(['fr-FR']);
    vi.stubGlobal(
      'fetch',
      vi.fn(() => Promise.resolve(jsonResponse({ status: 'ok' }))),
    );
    render(<App />);
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Bienvenue sur 226os' }),
    ).toBeInTheDocument();
  });
});
