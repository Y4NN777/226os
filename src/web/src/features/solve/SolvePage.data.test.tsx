import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { callsTo, mockApi } from '../../test/api';
import { renderApp, setLanguages } from '../../test/render';
import { getVisibleText } from '../../test/visibleText';
import { page, question } from './testFixtures';

const second = question({ id: 'b0000000-0000-4000-8000-000000000002', title: 'Erreur CORS' });

function lastParams(fetchMock: ReturnType<typeof mockApi>) {
  const calls = callsTo(fetchMock, 'GET /questions');
  return new URL(calls[calls.length - 1]?.url ?? 'http://x').searchParams;
}

async function sortBy(option: string) {
  const user = userEvent.setup({ delay: null });
  await user.click(screen.getByRole('combobox', { name: /^Trier/ }));
  await user.click(await screen.findByRole('option', { name: option }));
}

describe('SolvePage count and sort', () => {
  it('counts every question from the total', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /questions': () => page([question(), second], 'c1', 41) });
    await renderApp('/solve');
    await screen.findByRole('link', { name: 'Webhook non reçu après un paiement' });
    expect(getVisibleText('41 questions')).toBeInTheDocument();
  });

  it('says « 1 question » in the singular', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /questions': () => page([question()]) });
    await renderApp('/solve');
    await screen.findByRole('link', { name: 'Webhook non reçu après un paiement' });
    expect(getVisibleText('1 question')).toBeInTheDocument();
  });

  it('counts in English', async () => {
    setLanguages(['en-US']);
    mockApi({ 'GET /questions': () => page([question(), second], null, 2) });
    await renderApp('/solve');
    await screen.findByRole('link', { name: 'Erreur CORS' });
    expect(getVisibleText('2 questions')).toBeInTheDocument();
  });

  it('offers the three sorts, Récentes first', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /questions': () => page([question()]) });
    await renderApp('/solve');
    await screen.findByRole('link', { name: 'Webhook non reçu après un paiement' });
    const user = userEvent.setup({ delay: null });
    const trigger = screen.getByRole('combobox', { name: /^Trier/ });
    expect(trigger).toHaveTextContent('Récentes');
    await user.click(trigger);
    expect(await screen.findByRole('option', { name: 'Votes' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'Sans réponse' })).toBeInTheDocument();
  });

  it('reads the sort from the address and sends it', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ 'GET /questions': () => page([question()]) });
    await renderApp('/solve?sort=votes');
    await screen.findByRole('link', { name: 'Webhook non reçu après un paiement' });
    expect(lastParams(fetchMock).get('sort')).toBe('votes');
    expect(screen.getByRole('combobox', { name: /^Trier/ })).toHaveTextContent('Votes');
  });

  it('sends no sort by default and ignores an unknown one', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ 'GET /questions': () => page([question()]) });
    await renderApp('/solve?sort=best');
    await screen.findByRole('link', { name: 'Webhook non reçu après un paiement' });
    expect(lastParams(fetchMock).has('sort')).toBe(false);
  });

  it('puts the sort in the address and restarts at the first page', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({
      'GET /questions': (request) =>
        new URL(request.url).searchParams.get('cursor') === 'c1'
          ? page([second], null, 3)
          : page([question()], 'c1', 3),
    });
    const { router } = await renderApp('/solve');
    await screen.findByRole('link', { name: 'Webhook non reçu après un paiement' });
    await userEvent
      .setup({ delay: null })
      .click(screen.getByRole('button', { name: 'Afficher plus' }));
    await screen.findByRole('link', { name: 'Erreur CORS' });
    await sortBy('Sans réponse');
    await waitFor(() => {
      expect(lastParams(fetchMock).get('sort')).toBe('unanswered');
    });
    expect(lastParams(fetchMock).has('cursor')).toBe(false);
    expect(router.state.location.search).toMatchObject({ sort: 'unanswered' });
  });

  it('keeps the sort when a filter changes and drops it for the default', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ 'GET /questions': () => page([question()]) });
    const { router } = await renderApp('/solve?sort=votes');
    await screen.findByRole('link', { name: 'Webhook non reçu après un paiement' });
    await userEvent
      .setup({ delay: null })
      .type(screen.getByRole('searchbox', { name: 'Rechercher une question' }), 'cors{Enter}');
    await waitFor(() => {
      expect(lastParams(fetchMock).get('q')).toBe('cors');
    });
    expect(lastParams(fetchMock).get('sort')).toBe('votes');
    await sortBy('Récentes');
    await waitFor(() => {
      expect(lastParams(fetchMock).has('sort')).toBe(false);
    });
    expect(router.state.location.search).not.toHaveProperty('sort');
  });
});
