import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { callsTo, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';
import { member, page, question, questionId } from './testFixtures';
import { findVisibleText } from '../../test/visibleText';

const second = question({
  id: 'b0000000-0000-4000-8000-000000000002',
  title: 'Erreur CORS depuis une API Go',
  state: 'resolved',
  answer_count: 3,
  tags: ['go', 'cors'],
});

/** The error banner and its wrapper both expose role alert: look at all of them. */
async function expectAlert(text: string) {
  await waitFor(() => {
    const alerts = screen.getAllByRole('alert').map((alert) => alert.textContent);
    expect(alerts.join(' ')).toContain(text);
  });
}

function params(request: Request | undefined) {
  return new URL(request?.url ?? 'http://x').searchParams;
}

describe('SolvePage list', () => {
  it('lists questions as links with their state, tags and author, in French', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /questions': () => page([question(), second]) });
    await renderApp('/solve');

    expect(await screen.findByRole('heading', { level: 1, name: 'Résoudre' })).toBeInTheDocument();
    const link = await screen.findByRole('link', { name: 'Webhook non reçu après un paiement' });
    expect(link).toHaveAttribute('href', `/questions/${questionId}`);
    const results = screen.getByRole('region', { name: 'Questions' });
    expect(within(results).getByText('Sans réponse')).toBeInTheDocument();
    expect(within(results).getByText('Résolue')).toBeInTheDocument();
    expect(within(results).getByText('go')).toBeInTheDocument();
    expect(within(results).getAllByText('Par Kofi Mensah')).toHaveLength(2);
  });

  it('loads the next page with the opaque cursor', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({
      'GET /questions': (request) =>
        params(request).get('cursor') === 'cur-2' ? page([second]) : page([question()], 'cur-2'),
    });
    const user = userEvent.setup();
    await renderApp('/solve');
    await screen.findByRole('link', { name: 'Webhook non reçu après un paiement' });
    await user.click(screen.getByRole('button', { name: 'Afficher plus' }));

    expect(
      await screen.findByRole('link', { name: 'Erreur CORS depuis une API Go' }),
    ).toBeVisible();
    expect(callsTo(fetchMock, 'GET /questions')).toHaveLength(2);
    expect(screen.queryByRole('button', { name: 'Afficher plus' })).not.toBeInTheDocument();
  });

  it('sends the search text and the tag to the server, which ranks the results', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ 'GET /questions': () => page([question()]) });
    const user = userEvent.setup();
    await renderApp('/solve');
    await screen.findByRole('link', { name: 'Webhook non reçu après un paiement' });

    await user.type(screen.getByRole('searchbox', { name: 'Rechercher une question' }), 'cors');
    await user.type(screen.getByRole('textbox', { name: 'Filtrer par étiquette' }), 'go');
    await user.click(screen.getByRole('button', { name: 'Lancer la recherche' }));

    await waitFor(() => {
      const last = callsTo(fetchMock, 'GET /questions').at(-1);
      expect(params(last).get('q')).toBe('cors');
      expect(params(last).get('tag')).toBe('go');
    });
  });

  it('says so when nothing matches, and when the load fails', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /questions': () => page([]) });
    await renderApp('/solve?q=zzz');
    expect(await findVisibleText('Aucune question ne correspond')).toBeInTheDocument();
  });

  it('shows a retryable error when the list cannot be loaded', async () => {
    setLanguages(['fr-FR']);
    let calls = 0;
    mockApi({
      'GET /questions': () =>
        ++calls === 1 ? problemResponse(500, 'INTERNAL_ERROR') : page([question()]),
    });
    const user = userEvent.setup();
    await renderApp('/solve');
    await expectAlert('Une erreur est survenue');
    await user.click(screen.getByRole('button', { name: 'Réessayer' }));
    expect(
      await screen.findByRole('link', { name: 'Webhook non reçu après un paiement' }),
    ).toBeVisible();
  });
});

const linkedId = 'd0000000-0000-4000-8000-000000000009';

describe('SolvePage pre-filled link', () => {
  it('shows the link from the route as a removable chip and sends it', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({
      'GET /questions': () => page([]),
      'POST /questions': () => jsonResponse(question(), { status: 201 }),
    });
    const user = userEvent.setup();
    await renderApp(`/solve?link=article:${linkedId}`);
    // The form opens by itself: the visitor came from a lesson to ask about it.
    await user.type(await screen.findByRole('textbox', { name: 'Titre' }), 'Une question de cours');
    await user.type(screen.getByRole('textbox', { name: 'Votre question' }), 'Contexte.');
    await user.type(screen.getByRole('textbox', { name: /Étiquettes/ }), 'go');
    expect(screen.getByText('Article')).toBeInTheDocument();
    expect(screen.queryByRole('textbox', { name: /Identifiant/ })).not.toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Publier la question' }));

    await waitFor(() => {
      expect(callsTo(fetchMock, 'POST /questions')).toHaveLength(1);
    });
    const body = (await callsTo(fetchMock, 'POST /questions')[0]?.clone().json()) as Record<
      string,
      unknown
    >;
    expect(body.links).toEqual([{ type: 'article', id: linkedId }]);
  });

  it('drops the link when the chip is removed', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({
      'GET /questions': () => page([]),
      'POST /questions': () => jsonResponse(question(), { status: 201 }),
    });
    const user = userEvent.setup();
    await renderApp(`/solve?link=resource:${linkedId}`);
    await user.type(await screen.findByRole('textbox', { name: 'Titre' }), 'Une question de fiche');
    await user.type(screen.getByRole('textbox', { name: 'Votre question' }), 'Contexte.');
    await user.type(screen.getByRole('textbox', { name: /Étiquettes/ }), 'go');
    await user.click(screen.getByRole('button', { name: 'Retirer le lien Fiche' }));
    expect(screen.queryByText('Fiche')).not.toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Publier la question' }));
    await waitFor(() => {
      expect(callsTo(fetchMock, 'POST /questions')).toHaveLength(1);
    });
    const body = (await callsTo(fetchMock, 'POST /questions')[0]?.clone().json()) as Record<
      string,
      unknown
    >;
    expect(body.links).toBeUndefined();
  });

  it('ignores a malformed or unsupported link, and offers no free identifier field', async () => {
    setLanguages(['fr-FR']);
    member({ 'GET /questions': () => page([]) });
    const user = userEvent.setup();
    await renderApp('/solve?link=question:not-a-uuid');
    await user.click(await screen.findByRole('button', { name: 'Poser une question' }));
    expect(screen.queryByRole('textbox', { name: /Identifiant/ })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /Ajouter le lien/ })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /Retirer le lien/ })).not.toBeInTheDocument();
  });

  it('keeps the link through the sign-in invitation', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /questions': () => page([]) });
    await renderApp(`/solve?link=article:${linkedId}`);
    const invite = await screen.findByRole('link', {
      name: 'Se connecter pour poser une question',
    });
    expect(invite).toHaveAttribute(
      'href',
      `/connexion?next=${encodeURIComponent(`/solve?link=article:${linkedId}`)}`,
    );
  });
});

describe('SolvePage ask flow', () => {
  it('invites an anonymous visitor to sign in, coming back to /solve', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /questions': () => page([]) });
    await renderApp('/solve');
    const link = await screen.findByRole('link', { name: 'Se connecter pour poser une question' });
    expect(link).toHaveAttribute('href', '/connexion?next=%2Fsolve');
    expect(screen.queryByRole('button', { name: 'Poser une question' })).not.toBeInTheDocument();
  });

  async function fillAndOpen(user: ReturnType<typeof userEvent.setup>) {
    await user.click(await screen.findByRole('button', { name: 'Poser une question' }));
    await user.type(screen.getByRole('textbox', { name: 'Titre' }), 'Comment signer un webhook ?');
    await user.type(screen.getByRole('textbox', { name: 'Votre question' }), 'Voici le contexte.');
    await user.type(screen.getByRole('textbox', { name: /Étiquettes/ }), 'Go, webhook');
  }

  it('publishes with an Idempotency-Key and a JSON body, then opens the question', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({
      'GET /questions': () => page([]),
      'POST /questions': () => jsonResponse(question({ id: questionId }), { status: 201 }),
      'GET /questions/b0000000-0000-4000-8000-000000000001': () => jsonResponse(question()),
      'GET /questions/b0000000-0000-4000-8000-000000000001/answers': () => page([]),
    });
    const user = userEvent.setup();
    const { router } = await renderApp('/solve');
    await fillAndOpen(user);
    await user.click(screen.getByRole('button', { name: 'Publier la question' }));

    await waitFor(() => {
      expect(router.state.location.pathname).toBe(`/questions/${questionId}`);
    });
    const post = callsTo(fetchMock, 'POST /questions')[0];
    expect(post?.headers.get('Idempotency-Key')).toMatch(/^[0-9a-f-]{36}$/);
    expect(post?.headers.get('X-CSRF-Token')).toBe('csrf-member');
    expect(post?.headers.get('Content-Type')).toContain('application/json');
    expect(await post?.clone().json()).toEqual({
      title: 'Comment signer un webhook ?',
      body_md: 'Voici le contexte.',
      lang: 'fr',
      tags: ['go', 'webhook'],
    });
  });

  it('sends one snippet with its language', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({
      'GET /questions': () => page([]),
      'POST /questions': () => jsonResponse(question(), { status: 201 }),
    });
    const user = userEvent.setup();
    await renderApp('/solve');
    await fillAndOpen(user);
    await user.type(screen.getByRole('textbox', { name: /Langage de l’extrait/ }), 'go');
    await user.type(screen.getByRole('textbox', { name: /Code de l’extrait/ }), 'fmt.Println(1)');
    await user.click(screen.getByRole('button', { name: 'Publier la question' }));

    await waitFor(() => {
      expect(callsTo(fetchMock, 'POST /questions')).toHaveLength(1);
    });
    const body = (await callsTo(fetchMock, 'POST /questions')[0]?.clone().json()) as Record<
      string,
      unknown
    >;
    expect(body.snippets).toEqual([{ language: 'go', code: 'fmt.Println(1)' }]);
    expect(body.links).toBeUndefined();
  });

  it('keeps the same Idempotency-Key on a retry, and renews it once the content changes', async () => {
    setLanguages(['fr-FR']);
    let attempts = 0;
    const fetchMock = member({
      'GET /questions': () => page([]),
      'POST /questions': () =>
        ++attempts < 3
          ? problemResponse(503, 'SERVICE_UNAVAILABLE')
          : jsonResponse(question(), { status: 201 }),
    });
    const user = userEvent.setup();
    await renderApp('/solve');
    await fillAndOpen(user);
    const submit = screen.getByRole('button', { name: 'Publier la question' });
    await user.click(submit);
    await screen.findAllByRole('alert');
    await user.click(submit);
    await waitFor(() => {
      expect(callsTo(fetchMock, 'POST /questions')).toHaveLength(2);
    });
    await user.type(screen.getByRole('textbox', { name: 'Votre question' }), ' Précision.');
    await user.click(submit);
    await waitFor(() => {
      expect(callsTo(fetchMock, 'POST /questions')).toHaveLength(3);
    });
    const keys = callsTo(fetchMock, 'POST /questions').map((r) => r.headers.get('Idempotency-Key'));
    expect(keys[0]).toBe(keys[1]);
    expect(keys[2]).not.toBe(keys[0]);
  });

  it('puts server violations on the matching fields', async () => {
    setLanguages(['fr-FR']);
    member({
      'GET /questions': () => page([]),
      'POST /questions': () =>
        problemResponse(422, 'VALIDATION_ERROR', {
          violations: [
            { field: '/title', code: 'too_short' },
            { field: '/tags', code: 'too_many_items' },
          ],
        }),
    });
    const user = userEvent.setup();
    await renderApp('/solve');
    await fillAndOpen(user);
    await user.click(screen.getByRole('button', { name: 'Publier la question' }));

    expect(await screen.findByText('Le titre doit faire entre 5 et 200 caractères.')).toBeVisible();
    expect(screen.getByRole('textbox', { name: 'Titre' })).toBeInvalid();
    expect(screen.getByRole('textbox', { name: /Étiquettes/ })).toBeInvalid();
    expect(screen.getByRole('textbox', { name: 'Votre question' })).toBeValid();
  });

  it('refuses more than five tags before calling the server', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({ 'GET /questions': () => page([]) });
    const user = userEvent.setup();
    await renderApp('/solve');
    await user.click(await screen.findByRole('button', { name: 'Poser une question' }));
    await user.type(screen.getByRole('textbox', { name: 'Titre' }), 'Un titre valable');
    await user.type(screen.getByRole('textbox', { name: 'Votre question' }), 'Contexte.');
    await user.type(screen.getByRole('textbox', { name: /Étiquettes/ }), 'a, b, c, d, e, f');
    await user.click(screen.getByRole('button', { name: 'Publier la question' }));

    expect(await screen.findByText(/Indiquez de 1 à 5 étiquettes/)).toBeVisible();
    expect(callsTo(fetchMock, 'POST /questions')).toHaveLength(0);
  });

  it('shows the server message when the email is not verified', async () => {
    setLanguages(['fr-FR']);
    member({
      'GET /questions': () => page([]),
      'POST /questions': () => problemResponse(403, 'EMAIL_NOT_VERIFIED'),
    });
    const user = userEvent.setup();
    await renderApp('/solve');
    await fillAndOpen(user);
    await user.click(screen.getByRole('button', { name: 'Publier la question' }));
    await expectAlert('Confirmez d’abord votre adresse e-mail.');
  });

  it('is written in English when the browser asks for it', async () => {
    setLanguages(['en-GB']);
    member({ 'GET /questions': () => page([]) });
    await renderApp('/solve');
    expect(await screen.findByRole('heading', { level: 1, name: 'Solve' })).toBeInTheDocument();
    expect(await screen.findByRole('button', { name: 'Ask a question' })).toBeInTheDocument();
  });
});

describe('SolvePage assistant panel', () => {
  const answered = {
    outcome: 'answered',
    generated_by: 'ai',
    lang: 'fr',
    answer: 'Vérifiez la signature <b>HMAC</b>.',
    how_to_verify: 'Relisez la documentation du fournisseur.',
    sources: [],
  } as const;

  async function panel() {
    const heading = await screen.findByRole('heading', { level: 2, name: /^Assistant/ });
    return within(heading.closest('section') as HTMLElement);
  }

  async function open(user: ReturnType<typeof userEvent.setup>) {
    await user.click(await screen.findByRole('button', { name: 'Poser une question' }));
  }

  it('shows the Google notice, sends nothing, and stays disabled under 3 characters', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({ 'GET /questions': () => page([]) });
    const user = userEvent.setup();
    await renderApp('/solve');
    await open(user);
    const inPanel = await panel();
    expect(inPanel.getByText(/envoyée à Google \(AI Studio\)/)).toBeInTheDocument();
    const ask = inPanel.getByRole('button', { name: 'Demander à l’assistant' });
    expect(ask).toHaveAttribute('aria-disabled', 'true');
    expect(
      inPanel.getByText('Saisissez au moins 3 caractères pour interroger l’assistant.'),
    ).toBeVisible();
    await user.type(screen.getByRole('textbox', { name: 'Titre' }), 'ab');
    expect(inPanel.getByRole('button', { name: 'Demander à l’assistant' })).toHaveAttribute(
      'aria-disabled',
      'true',
    );
    expect(callsTo(fetchMock, 'POST /assistant/answers')).toHaveLength(0);
  });

  it('asks with the draft title and body, CSRF header and JSON body, no vote control', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({
      'GET /questions': () => page([]),
      'POST /assistant/answers': () => jsonResponse(answered),
    });
    const user = userEvent.setup();
    await renderApp('/solve');
    await open(user);
    await user.type(screen.getByRole('textbox', { name: 'Titre' }), 'Comment signer un webhook ?');
    await user.type(screen.getByRole('textbox', { name: 'Votre question' }), 'Voici le contexte.');
    const inPanel = await panel();
    await user.click(inPanel.getByRole('button', { name: 'Demander à l’assistant' }));
    expect(await inPanel.findByText(/Vérifiez la signature/)).toBeInTheDocument();
    const [request] = callsTo(fetchMock, 'POST /assistant/answers');
    expect(request?.headers.get('X-CSRF-Token')).toBe('csrf-member');
    expect(request?.headers.get('Content-Type')).toContain('application/json');
    expect(await request?.json()).toEqual({
      question: 'Comment signer un webhook ?\n\nVoici le contexte.',
      lang: 'fr',
    });
    expect(inPanel.getAllByText('IA').length).toBeGreaterThan(0);
    expect(
      inPanel.getByRole('heading', { level: 3, name: 'Comment vérifier' }),
    ).toBeInTheDocument();
    expect(inPanel.queryByRole('button', { name: /voter|accepter|résolu/i })).toBeNull();
    expect(callsTo(fetchMock, 'POST /questions')).toHaveLength(0);
  });

  it('refuses with no_source', async () => {
    setLanguages(['fr-FR']);
    member({
      'GET /questions': () => page([]),
      'POST /assistant/answers': () =>
        jsonResponse({ ...answered, outcome: 'no_source', search_url: '/recherche?q=abc' }),
    });
    const user = userEvent.setup();
    await renderApp('/solve');
    await open(user);
    await user.type(screen.getByRole('textbox', { name: 'Titre' }), 'Un titre valable');
    const inPanel = await panel();
    await user.click(inPanel.getByRole('button', { name: 'Demander à l’assistant' }));
    expect(await inPanel.findByText(/Aucune source pour répondre/)).toBeInTheDocument();
  });

  it('on 503 hides the panel while the question can still be posted', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({
      'GET /questions': () => page([]),
      'POST /assistant/answers': () => problemResponse(503, 'SERVICE_UNAVAILABLE'),
      'POST /questions': () => jsonResponse(question({ id: questionId }), { status: 201 }),
      'GET /questions/b0000000-0000-4000-8000-000000000001': () => jsonResponse(question()),
      'GET /questions/b0000000-0000-4000-8000-000000000001/answers': () => page([]),
    });
    const user = userEvent.setup();
    const { router } = await renderApp('/solve');
    await open(user);
    await user.type(screen.getByRole('textbox', { name: 'Titre' }), 'Comment signer un webhook ?');
    await user.type(screen.getByRole('textbox', { name: 'Votre question' }), 'Contexte.');
    await user.type(screen.getByRole('textbox', { name: /Étiquettes/ }), 'go');
    const inPanel = await panel();
    await user.click(inPanel.getByRole('button', { name: 'Demander à l’assistant' }));
    expect(await inPanel.findByText(/L’assistant est indisponible/)).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Publier la question' }));
    await waitFor(() => {
      expect(router.state.location.pathname).toBe(`/questions/${questionId}`);
    });
    expect(callsTo(fetchMock, 'POST /assistant/answers')).toHaveLength(1);
    expect(callsTo(fetchMock, 'POST /questions')).toHaveLength(1);
  });

  it('posts the question independently of the assistant, which is never asked', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({
      'GET /questions': () => page([]),
      'POST /questions': () => jsonResponse(question({ id: questionId }), { status: 201 }),
      'GET /questions/b0000000-0000-4000-8000-000000000001': () => jsonResponse(question()),
      'GET /questions/b0000000-0000-4000-8000-000000000001/answers': () => page([]),
    });
    const user = userEvent.setup();
    const { router } = await renderApp('/solve');
    await open(user);
    await user.type(screen.getByRole('textbox', { name: 'Titre' }), 'Comment signer un webhook ?');
    await user.type(screen.getByRole('textbox', { name: 'Votre question' }), 'Contexte.');
    await user.type(screen.getByRole('textbox', { name: /Étiquettes/ }), 'go');
    await user.click(screen.getByRole('button', { name: 'Publier la question' }));
    await waitFor(() => {
      expect(router.state.location.pathname).toBe(`/questions/${questionId}`);
    });
    expect(callsTo(fetchMock, 'POST /assistant/answers')).toHaveLength(0);
  });

  it('speaks English', async () => {
    setLanguages(['en-US']);
    member({ 'GET /questions': () => page([]) });
    const user = userEvent.setup();
    await renderApp('/solve');
    await user.click(await screen.findByRole('button', { name: 'Ask a question' }));
    const inPanel = await panel();
    expect(inPanel.getByRole('button', { name: 'Ask the assistant' })).toHaveAttribute(
      'aria-disabled',
      'true',
    );
  });
});
