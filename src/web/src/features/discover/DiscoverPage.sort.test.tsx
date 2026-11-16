import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { mockApi } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';
import { selectOption } from '../../test/select';
import { getVisibleText } from '../../test/visibleText';
import { listParams, resource } from './testFixtures';

const wave = resource({ id: 'a0000000-0000-4000-8000-000000000001', name: 'Wave Pay' });
const hub = resource({ id: 'a0000000-0000-4000-8000-000000000002', name: 'Safari Hub' });

function page(
  items: ReturnType<typeof resource>[],
  next: string | null = null,
  total = items.length,
) {
  return jsonResponse({ items, next_cursor: next, total });
}

describe('DiscoverPage sort', () => {
  it('offers Nom first and Vérifiées récemment', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /resources': () => page([wave]) });
    await renderApp('/discover');
    await screen.findByRole('link', { name: 'Wave Pay' });
    const user = userEvent.setup({ delay: null });
    const trigger = screen.getByRole('combobox', { name: /^Trier/ });
    expect(trigger).toHaveTextContent('Nom');
    // Open by keyboard and look inside the listbox: a page-wide role query costs about 600 ms here.
    trigger.focus();
    await user.keyboard('{Enter}');
    expect(
      within(screen.getByRole('listbox')).getByText('Vérifiées récemment', {
        selector: '[role="option"], [role="option"] *',
      }),
    ).toBeInTheDocument();
  });

  it('sends no sort by default', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ 'GET /resources': () => page([wave]) });
    await renderApp('/discover');
    await screen.findByRole('link', { name: 'Wave Pay' });
    expect(listParams(fetchMock, 'last').has('sort')).toBe(false);
  });

  it('reads the sort from the address and ignores an unknown one', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ 'GET /resources': () => page([wave]) });
    await renderApp('/discover?sort=verified');
    await screen.findByRole('link', { name: 'Wave Pay' });
    expect(listParams(fetchMock, 'last').get('sort')).toBe('verified');
    expect(screen.getByRole('combobox', { name: /^Trier/ })).toHaveTextContent(
      'Vérifiées récemment',
    );
  });

  it('drops an unknown sort', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ 'GET /resources': () => page([wave]) });
    await renderApp('/discover?sort=newest');
    await screen.findByRole('link', { name: 'Wave Pay' });
    expect(listParams(fetchMock, 'last').has('sort')).toBe(false);
  });

  it('puts the sort in the address, restarts at the first page and keeps it with a filter', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({
      'GET /resources': (request) =>
        new URL(request.url).searchParams.get('cursor') === 'c1'
          ? page([hub], null, 3)
          : page([wave], 'c1', 3),
    });
    const user = userEvent.setup({ delay: null });
    const { router } = await renderApp('/discover');
    await screen.findByRole('link', { name: 'Wave Pay' });
    await user.click(screen.getByRole('button', { name: 'Afficher plus' }));
    await screen.findByRole('link', { name: 'Safari Hub' });
    await selectOption(screen.getByRole('combobox', { name: /^Trier/ }), 'Vérifiées récemment');
    await waitFor(() => {
      expect(listParams(fetchMock, 'last').get('sort')).toBe('verified');
    });
    expect(listParams(fetchMock, 'last').has('cursor')).toBe(false);
    expect(router.state.location.search).toMatchObject({ sort: 'verified' });
    await user.type(
      screen.getByRole('searchbox', { name: /Rechercher dans le catalogue/ }),
      'momo{Enter}',
    );
    await waitFor(() => {
      expect(listParams(fetchMock, 'last').get('q')).toBe('momo');
    });
    expect(listParams(fetchMock, 'last').get('sort')).toBe('verified');
  });

  it('counts from the total, in English too', async () => {
    setLanguages(['en-US']);
    mockApi({ 'GET /resources': () => page([wave], 'c1', 57) });
    await renderApp('/discover');
    await screen.findByRole('link', { name: 'Wave Pay' });
    expect(getVisibleText('57 listings')).toBeInTheDocument();
  });
});
