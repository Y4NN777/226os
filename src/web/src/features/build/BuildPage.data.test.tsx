import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { callsTo, mockApi } from '../../test/api';
import { renderApp, setLanguages } from '../../test/render';
import { selectOption } from '../../test/select';
import { getVisibleText } from '../../test/visibleText';
import { page, project } from './testFixtures';

const second = project({
  id: 'e0000000-0000-4000-8000-000000000002',
  name: 'sahel-docs',
  technologies: [],
  open_needs_count: 0,
});

function lastParams(fetchMock: ReturnType<typeof mockApi>) {
  const calls = callsTo(fetchMock, 'GET /projects');
  return new URL(calls[calls.length - 1]?.url ?? 'http://x').searchParams;
}

async function pick(label: RegExp, option: string) {
  await selectOption(
    screen.getByRole(label.source.includes('Pays') ? 'button' : 'combobox', { name: label }),
    option,
  );
}

describe('BuildPage stage and country', () => {
  it('writes the stage and the country in each row meta', async () => {
    setLanguages(['fr-FR']);
    mockApi({
      'GET /projects': () => page([project({ stage: 'active', country_code: 'BF' }), second]),
    });
    await renderApp('/build');
    const row = (await screen.findByRole('link', { name: 'kora-sdk' })).closest('li');
    const scope = within(row as HTMLElement);
    expect(scope.getByText('Actif')).toBeVisible();
    expect(scope.getByText('BF')).toBeInTheDocument();
    expect(scope.getByText('Burkina Faso')).toBeInTheDocument();
    // A project with no country declares none: no empty placeholder.
    const bare = within(
      screen.getByRole('link', { name: 'sahel-docs' }).closest('li') as HTMLElement,
    );
    expect(bare.getByText('Idée')).toBeVisible();
    expect(bare.queryByText(/^[A-Z]{2}$/)).not.toBeInTheDocument();
  });

  it('names the stage in English', async () => {
    setLanguages(['en-US']);
    mockApi({ 'GET /projects': () => page([project({ stage: 'maintained' })]) });
    await renderApp('/build');
    const row = (await screen.findByRole('link', { name: 'kora-sdk' })).closest('li');
    expect(within(row as HTMLElement).getByText('Maintenance')).toBeVisible();
  });
});

describe('BuildPage count', () => {
  it('counts every project from the total, not only the loaded ones', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /projects': () => page([project(), second], 'cur-2', 12) });
    await renderApp('/build');
    await screen.findByRole('link', { name: 'kora-sdk' });
    expect(getVisibleText('12 projets')).toBeInTheDocument();
  });

  it('says « 1 projet » in the singular', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /projects': () => page([project()]) });
    await renderApp('/build');
    await screen.findByRole('link', { name: 'kora-sdk' });
    expect(getVisibleText('1 projet')).toBeInTheDocument();
  });

  it('counts in English', async () => {
    setLanguages(['en-US']);
    mockApi({ 'GET /projects': () => page([project(), second], null, 2) });
    await renderApp('/build');
    await screen.findByRole('link', { name: 'kora-sdk' });
    expect(getVisibleText('2 projects')).toBeInTheDocument();
  });
});

describe('BuildPage URL parameters', () => {
  it('reads every filter and the sort from the address', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ 'GET /projects': () => page([project()]) });
    await renderApp('/build?q=sms&open=true&stage=active&country=BF&sort=needs');
    await screen.findByRole('link', { name: 'kora-sdk' });
    const params = lastParams(fetchMock);
    expect(params.get('q')).toBe('sms');
    expect(params.get('open_needs')).toBe('true');
    expect(params.get('stage')).toBe('active');
    expect(params.get('country')).toBe('BF');
    expect(params.get('sort')).toBe('needs');
  });

  it('sends no sort by default', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ 'GET /projects': () => page([project()]) });
    await renderApp('/build');
    await screen.findByRole('link', { name: 'kora-sdk' });
    expect(lastParams(fetchMock).has('sort')).toBe(false);
  });

  it('puts the stage in the address and asks the server for it', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ 'GET /projects': () => page([project()]) });
    const { router } = await renderApp('/build');
    await screen.findByRole('link', { name: 'kora-sdk' });
    await pick(/^Étape/, 'Prototype');
    await waitFor(() => {
      expect(lastParams(fetchMock).get('stage')).toBe('prototype');
    });
    expect(router.state.location.search).toMatchObject({ stage: 'prototype' });
  });

  it('puts the country in the address and asks the server for it', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ 'GET /projects': () => page([project()]) });
    const { router } = await renderApp('/build');
    await screen.findByRole('link', { name: 'kora-sdk' });
    await pick(/^Pays/, 'Sénégal');
    await waitFor(() => {
      expect(lastParams(fetchMock).get('country')).toBe('SN');
    });
    expect(router.state.location.search).toMatchObject({ country: 'SN' });
  });

  it('puts the sort in the address, and drops it for the default one', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ 'GET /projects': () => page([project()]) });
    const { router } = await renderApp('/build');
    await screen.findByRole('link', { name: 'kora-sdk' });
    await pick(/^Trier/, 'Créés');
    await waitFor(() => {
      expect(lastParams(fetchMock).get('sort')).toBe('created');
    });
    expect(router.state.location.search).toMatchObject({ sort: 'created' });
    await pick(/^Trier/, 'Besoins');
    await waitFor(() => {
      expect(lastParams(fetchMock).get('sort')).toBe('needs');
    });
    await pick(/^Trier/, 'Récents');
    await waitFor(() => {
      expect(lastParams(fetchMock).has('sort')).toBe(false);
    });
    expect(router.state.location.search).not.toHaveProperty('sort');
  });

  it('keeps the text search and the open-needs box in the address', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ 'GET /projects': () => page([project()]) });
    const { router } = await renderApp('/build');
    await screen.findByRole('link', { name: 'kora-sdk' });
    const user = userEvent.setup({ delay: null });
    await user.type(screen.getByRole('searchbox', { name: 'Rechercher un projet' }), 'sms{Enter}');
    await waitFor(() => {
      expect(lastParams(fetchMock).get('q')).toBe('sms');
    });
    await user.click(screen.getByRole('checkbox', { name: 'Besoins ouverts uniquement' }));
    await waitFor(() => {
      expect(lastParams(fetchMock).get('open_needs')).toBe('true');
    });
    expect(router.state.location.search).toMatchObject({ q: 'sms', open: true });
  });

  it('restarts at the first page when a filter changes, never with the old cursor', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({
      'GET /projects': (request) =>
        new URL(request.url).searchParams.get('cursor') === 'cur-2'
          ? page([second], null, 3)
          : page([project()], 'cur-2', 3),
    });
    const user = userEvent.setup({ delay: null });
    await renderApp('/build');
    await screen.findByRole('link', { name: 'kora-sdk' });
    await user.click(screen.getByRole('button', { name: 'Afficher plus' }));
    await screen.findByRole('link', { name: 'sahel-docs' });
    await pick(/^Trier/, 'Créés');
    await waitFor(() => {
      expect(lastParams(fetchMock).get('sort')).toBe('created');
    });
    expect(lastParams(fetchMock).has('cursor')).toBe(false);
  });

  it('clears the stage, the country and the sort with the reset button', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /projects': () => page([project()]) });
    const { router } = await renderApp('/build?stage=active&country=BF&sort=needs');
    await screen.findByRole('link', { name: 'kora-sdk' });
    await userEvent
      .setup({ delay: null })
      .click(screen.getByRole('button', { name: 'Effacer les filtres' }));
    await waitFor(() => {
      expect(router.state.location.search).toEqual({});
    });
  });
});
