import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { mockApi } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';
import { resource } from './testFixtures';

const wave = resource({ id: 'a0000000-0000-4000-8000-000000000001', name: 'Wave Pay' });
function page(items: ReturnType<typeof resource>[]) {
  return jsonResponse({ items, next_cursor: null, total: items.length });
}

const api = { 'GET /resources': () => page([wave]) };

async function listOf() {
  await screen.findByRole('link', { name: 'Wave Pay' });
  return screen.getByRole('list', { name: 'Résultats' });
}

describe('DiscoverPage card and list views', () => {
  it('shows cards by default, each a link with the monogram tile', async () => {
    setLanguages(['fr-FR']);
    mockApi(api);
    await renderApp('/discover');
    const list = await listOf();
    expect(list).toHaveClass('ui-card-grid');
    expect(screen.getByRole('button', { name: 'Cartes' })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('link', { name: 'Wave Pay' })).toHaveAttribute(
      'href',
      '/discover/a0000000-0000-4000-8000-000000000001',
    );
    expect(list.querySelector('img')).toBeNull();
    expect(list.querySelector('.ui-tile')).not.toBeNull();
  });

  it('switches to the list and back, keeping the filters and the sort in the address', async () => {
    setLanguages(['fr-FR']);
    mockApi(api);
    const user = userEvent.setup({ delay: null });
    const { router } = await renderApp('/discover?region=all&sort=verified');
    await listOf();

    await user.click(screen.getByRole('button', { name: 'Liste' }));
    await waitFor(() => {
      expect(screen.getByRole('list', { name: 'Résultats' })).not.toHaveClass('ui-card-grid');
    });
    expect(screen.getByRole('button', { name: 'Liste' })).toHaveAttribute('aria-pressed', 'true');
    expect(router.state.location.search).toEqual({ region: 'all', sort: 'verified', view: 'list' });

    await user.click(screen.getByRole('button', { name: 'Cartes' }));
    await waitFor(() => {
      expect(router.state.location.search).toEqual({
        region: 'all',
        sort: 'verified',
        view: 'cards',
      });
    });
  });

  it('honours the address, then the remembered choice, and the address wins', async () => {
    setLanguages(['fr-FR']);
    mockApi(api);
    localStorage.setItem('226os.view.discover', 'list');
    const first = await renderApp('/discover');
    expect(await listOf()).not.toHaveClass('ui-card-grid');
    first.unmount();
    await renderApp('/discover?view=cards');
    expect(await listOf()).toHaveClass('ui-card-grid');
  });

  it('does not count the view as a filter', async () => {
    setLanguages(['fr-FR']);
    mockApi(api);
    await renderApp('/discover?view=list');
    await listOf();
    expect(screen.queryByRole('button', { name: 'Effacer les filtres' })).toBeNull();
  });

  it('is labelled in English', async () => {
    setLanguages(['en-US']);
    mockApi(api);
    await renderApp('/discover');
    await screen.findByRole('link', { name: 'Wave Pay' });
    expect(screen.getByRole('button', { name: 'Cards' })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: 'List' })).toHaveAttribute('aria-pressed', 'false');
  });
});
