import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { mockApi } from '../../test/api';
import { renderApp, setLanguages } from '../../test/render';
import { page, seriesSummary, summary } from './testFixtures';

const api = {
  'GET /learn/articles': () => page([summary()]),
  'GET /learn/series': () => page([seriesSummary()]),
};

async function listOf() {
  await screen.findByRole('link', { name: 'Vérifier un webhook signé' });
  return screen.getAllByRole('list', { name: 'Tous les contenus' })[0] as HTMLElement;
}

describe('LearnPage card and list views', () => {
  it('shows cards by default, each a link', async () => {
    setLanguages(['fr-FR']);
    mockApi(api);
    await renderApp('/apprendre');
    const list = await listOf();
    expect(list).toHaveClass('ui-card-grid');
    expect(screen.getByRole('button', { name: 'Cartes' })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('link', { name: 'Vérifier un webhook signé' })).toHaveAttribute(
      'href',
      '/apprendre/articles/d0000000-0000-4000-8000-000000000001',
    );
    expect(list.querySelector('img')).toBeNull();
  });

  it('switches to the list and back, keeping the filters in the address', async () => {
    setLanguages(['fr-FR']);
    mockApi(api);
    const user = userEvent.setup({ delay: null });
    const { router } = await renderApp('/apprendre?level=intermediate');
    await listOf();

    await user.click(screen.getByRole('button', { name: 'Liste' }));
    await waitFor(() => {
      expect(screen.getAllByRole('list', { name: 'Tous les contenus' })[0]).not.toHaveClass(
        'ui-card-grid',
      );
    });
    expect(screen.getByRole('button', { name: 'Liste' })).toHaveAttribute('aria-pressed', 'true');
    expect(router.state.location.search).toEqual({ level: 'intermediate', view: 'list' });

    await user.click(screen.getByRole('button', { name: 'Cartes' }));
    await waitFor(() => {
      expect(router.state.location.search).toEqual({ level: 'intermediate', view: 'cards' });
    });
  });

  it('honours the address, then the remembered choice, and the address wins', async () => {
    setLanguages(['fr-FR']);
    mockApi(api);
    localStorage.setItem('226os.view.learn', 'list');
    const first = await renderApp('/apprendre');
    expect(await listOf()).not.toHaveClass('ui-card-grid');
    first.unmount();
    await renderApp('/apprendre?view=cards');
    expect(await listOf()).toHaveClass('ui-card-grid');
  });

  it('does not count the view as a filter', async () => {
    setLanguages(['fr-FR']);
    mockApi(api);
    await renderApp('/apprendre?view=list');
    await listOf();
    expect(screen.queryByRole('button', { name: 'Effacer' })).toBeNull();
  });

  it('is labelled in English', async () => {
    setLanguages(['en-US']);
    mockApi(api);
    await renderApp('/apprendre');
    await screen.findByRole('link', { name: 'Vérifier un webhook signé' });
    expect(screen.getByRole('button', { name: 'Cards' })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: 'List' })).toHaveAttribute('aria-pressed', 'false');
  });
});
