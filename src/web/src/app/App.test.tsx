import { createMemoryHistory } from '@tanstack/react-router';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { mockApi } from '../test/api';
import { jsonResponse, setLanguages } from '../test/render';
import { App } from './App';
import { createAppRouter } from './router';

function stubApi() {
  mockApi({ 'GET /resources': () => jsonResponse({ items: [], next_cursor: null }) });
}

function renderAt(path: string) {
  stubApi();
  const router = createAppRouter({ history: createMemoryHistory({ initialEntries: [path] }) });
  return render(<App router={router} />);
}

describe('App routing', () => {
  it('sets a localized title for the current page', async () => {
    setLanguages(['fr-FR']);
    renderAt('/recherche');
    await screen.findByRole('heading', { level: 1, name: 'Recherche' });
    expect(document.title).toBe('Recherche · 226os');
  });

  it('renders the shell and the public landing page at /', async () => {
    setLanguages(['fr-FR']);
    renderAt('/');
    expect(
      await screen.findByRole('heading', {
        level: 1,
        name: 'Apprenez, échangez et construisez entre développeurs africains',
      }),
    ).toBeInTheDocument();
    expect(screen.getByRole('navigation', { name: 'Navigation principale' })).toBeInTheDocument();
    expect(await screen.findByRole('link', { name: 'Se connecter' })).toBeInTheDocument();
  });

  it('navigates to settings from the account menu without a full reload', async () => {
    setLanguages(['en-US']);
    renderAt('/');
    await userEvent.click(await screen.findByRole('button', { name: 'More options' }));
    await userEvent.click(await screen.findByRole('menuitem', { name: 'Settings' }));
    expect(await screen.findByRole('heading', { level: 1, name: 'Settings' })).toBeInTheDocument();
  });

  it('shows a localized not-found page with a way back', async () => {
    setLanguages(['fr-FR']);
    renderAt('/nope');
    const heading = await screen.findByRole('heading', { name: 'Page introuvable' });
    // No layout route matches an unknown address: the page frames itself in a main landmark.
    expect(screen.getByRole('main')).toContainElement(heading);
    await userEvent.click(screen.getByRole('link', { name: 'Retour à l’accueil' }));
    expect(
      await screen.findByRole('heading', {
        level: 1,
        name: 'Apprenez, échangez et construisez entre développeurs africains',
      }),
    ).toBeInTheDocument();
  });

  it('creates its own browser router when none is injected', async () => {
    setLanguages(['fr-FR']);
    stubApi();
    render(<App />);
    expect(
      await screen.findByRole('heading', {
        level: 1,
        name: 'Apprenez, échangez et construisez entre développeurs africains',
      }),
    ).toBeInTheDocument();
  });
});
