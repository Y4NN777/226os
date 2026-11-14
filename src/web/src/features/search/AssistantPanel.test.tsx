import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { components } from '../../api/schema';
import { callsTo, memberSession, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';

type AssistantAnswer = components['schemas']['AssistantAnswer'];

const SEARCH = {
  query: 'momo',
  groups: { resources: { items: [], has_more: false } },
};

const answered: AssistantAnswer = {
  outcome: 'answered',
  generated_by: 'ai',
  lang: 'fr',
  answer:
    'Utilisez le bac à sable [S1].\nPuis testez <b>les webhooks</b> avec <img src=x onerror=alert(1)>.',
  how_to_verify: 'Relisez la documentation du fournisseur.',
  sources: [
    {
      type: 'resource',
      id: '7a1c2a34-0000-4000-8000-0000000000a1',
      title: 'API Mobile Money',
      url: '/discover/7a1c2a34-0000-4000-8000-0000000000a1',
      published_at: null,
    },
    {
      type: 'article',
      id: '7a1c2a34-0000-4000-8000-0000000000c1',
      title: 'Guide externe',
      url: 'https://docs.example.org/guide',
      published_at: null,
    },
  ],
};

const noSource: AssistantAnswer = {
  outcome: 'no_source',
  generated_by: 'ai',
  lang: 'fr',
  answer: 'Aucune source publiée ne correspond.',
  how_to_verify: 'Cherchez dans le catalogue.',
  sources: [],
  search_url: '/recherche?q=momo',
};

function member(extra: Record<string, () => Response> = {}) {
  return mockApi({
    'GET /auth/session': () => jsonResponse(memberSession),
    'GET /search': () => jsonResponse(SEARCH),
    ...extra,
  });
}

async function panel() {
  const heading = await screen.findByRole('heading', { level: 2, name: /^Assistant/ });
  return within(heading.closest('section') as HTMLElement);
}

describe('AssistantPanel', () => {
  it('shows the Google notice before the first question and sends nothing', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member();
    await renderApp('/recherche?q=momo');
    const inPanel = await panel();
    expect(await inPanel.findByText(/envoyée à Google \(AI Studio\)/)).toBeInTheDocument();
    expect(inPanel.getByText(/Seuls la question et des contenus publics/)).toBeInTheDocument();
    expect(callsTo(fetchMock, 'POST /assistant/answers')).toHaveLength(0);
  });

  it('posts the current query with the CSRF token as a JSON body', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({ 'POST /assistant/answers': () => jsonResponse(answered) });
    await renderApp('/recherche?q=momo');
    const inPanel = await panel();
    await userEvent
      .setup({ delay: null })
      .click(await inPanel.findByRole('button', { name: 'Demander à l’assistant' }));
    await inPanel.findByText('Comment vérifier');
    const [request] = callsTo(fetchMock, 'POST /assistant/answers');
    expect(request?.headers.get('X-CSRF-Token')).toBe('csrf-member');
    expect(request?.headers.get('Content-Type')).toContain('application/json');
    expect(await request?.json()).toEqual({ question: 'momo', lang: 'fr' });
  });

  it('disables the action with a hint when the query is shorter than 3 characters', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member();
    await renderApp('/recherche?q=mo');
    const inPanel = await panel();
    const ask = await inPanel.findByRole('button', { name: 'Demander à l’assistant' });
    expect(ask).toHaveAttribute('aria-disabled', 'true');
    expect(
      inPanel.getByText('Saisissez au moins 3 caractères pour interroger l’assistant.'),
    ).toBeVisible();
    await userEvent.setup({ delay: null }).click(ask);
    expect(callsTo(fetchMock, 'POST /assistant/answers')).toHaveLength(0);
    await userEvent
      .setup({ delay: null })
      .type(inPanel.getByRole('textbox', { name: /Votre question/ }), 'Comment ?');
    expect(inPanel.getByRole('button', { name: 'Demander à l’assistant' })).not.toHaveAttribute(
      'aria-disabled',
      'true',
    );
  });

  it('sends the free question when one is typed', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({ 'POST /assistant/answers': () => jsonResponse(answered) });
    await renderApp('/recherche?q=momo');
    const inPanel = await panel();
    await userEvent
      .setup({ delay: null })
      .type(
        await inPanel.findByRole('textbox', { name: /Votre question/ }),
        'Comment tester un webhook ?',
      );
    await userEvent
      .setup({ delay: null })
      .click(inPanel.getByRole('button', { name: 'Demander à l’assistant' }));
    await inPanel.findByText('Comment vérifier');
    const [request] = callsTo(fetchMock, 'POST /assistant/answers');
    expect(await request?.json()).toEqual({ question: 'Comment tester un webhook ?', lang: 'fr' });
  });

  it('renders an answer as plain text with the AI label, sources and how to verify, no vote', async () => {
    setLanguages(['fr-FR']);
    member({ 'POST /assistant/answers': () => jsonResponse(answered) });
    const { container } = await renderApp('/recherche?q=momo');
    const inPanel = await panel();
    await userEvent
      .setup({ delay: null })
      .click(await inPanel.findByRole('button', { name: 'Demander à l’assistant' }));
    expect(await inPanel.findByText(/Utilisez le bac à sable/)).toBeInTheDocument();
    expect(inPanel.getAllByText('IA').length).toBeGreaterThan(0);
    expect(
      inPanel.getByRole('heading', { level: 3, name: 'Comment vérifier' }),
    ).toBeInTheDocument();
    expect(inPanel.getByText('Relisez la documentation du fournisseur.')).toBeInTheDocument();
    const internal = inPanel.getByRole('link', { name: 'API Mobile Money' });
    expect(internal).toHaveAttribute('href', '/discover/7a1c2a34-0000-4000-8000-0000000000a1');
    expect(internal).not.toHaveAttribute('target');
    const external = inPanel.getByRole('link', { name: /Guide externe/ });
    expect(external).toHaveAttribute('href', 'https://docs.example.org/guide');
    expect(external).toHaveAttribute('target', '_blank');
    expect(external.getAttribute('rel')).toContain('noopener');
    expect(
      inPanel.queryByRole('button', { name: /voter|vote|accepter|accept|résolu|solved/i }),
    ).toBeNull();
    // raw HTML in the answer stays text
    expect(inPanel.getByText(/<b>les webhooks<\/b>/)).toBeInTheDocument();
    expect(container.querySelector('img[src="x"]')).toBeNull();
    expect(container.querySelector('b')).toBeNull();
  });

  const dated: AssistantAnswer = {
    ...answered,
    sources: [
      {
        type: 'project',
        id: '7a1c2a34-0000-4000-8000-0000000000b1',
        title: 'Projet Lumière',
        url: '/build/7a1c2a34-0000-4000-8000-0000000000b1',
        published_at: '2026-11-02T10:00:00Z',
      },
      {
        type: 'article',
        id: '7a1c2a34-0000-4000-8000-0000000000c2',
        title: 'Guide webhooks',
        url: '/apprendre/articles/7a1c2a34-0000-4000-8000-0000000000c2',
      },
    ],
  };

  async function askWith(answer: AssistantAnswer, label: string) {
    member({ 'POST /assistant/answers': () => jsonResponse(answer) });
    await renderApp('/recherche?q=momo');
    const inPanel = await panel();
    await userEvent.click(await inPanel.findByRole('button', { name: label }));
    return inPanel;
  }

  it('shows each source date in French inside a time element, and Non daté when absent', async () => {
    setLanguages(['fr-FR']);
    const inPanel = await askWith(dated, 'Demander à l’assistant');
    const date = await inPanel.findByText('2 novembre 2026');
    expect(date.closest('time')).toHaveAttribute('datetime', '2026-11-02T10:00:00Z');
    expect(date.parentElement?.textContent).toBe('Publié le 2 novembre 2026');
    expect(inPanel.getByText('Non daté')).toBeInTheDocument();
    expect(inPanel.queryByText(/Undated/)).toBeNull();
  });

  it('shows each source date in English, and Undated when absent', async () => {
    setLanguages(['en-US']);
    const inPanel = await askWith(dated, 'Ask the assistant');
    const date = await inPanel.findByText('2 November 2026');
    expect(date.closest('time')).toHaveAttribute('datetime', '2026-11-02T10:00:00Z');
    expect(date.parentElement?.textContent).toBe('Published 2 November 2026');
    expect(inPanel.getByText('Undated')).toBeInTheDocument();
  });

  it('treats a null published_at as undated', async () => {
    setLanguages(['fr-FR']);
    const inPanel = await askWith(answered, 'Demander à l’assistant');
    expect((await inPanel.findAllByText('Non daté')).length).toBe(2);
  });

  it('links project and article sources to their in-app pages', async () => {
    setLanguages(['fr-FR']);
    const inPanel = await askWith(dated, 'Demander à l’assistant');
    const project = await inPanel.findByRole('link', { name: 'Projet Lumière' });
    expect(project).toHaveAttribute('href', '/build/7a1c2a34-0000-4000-8000-0000000000b1');
    expect(project).not.toHaveAttribute('target');
    const article = inPanel.getByRole('link', { name: 'Guide webhooks' });
    expect(article).toHaveAttribute(
      'href',
      '/apprendre/articles/7a1c2a34-0000-4000-8000-0000000000c2',
    );
    expect(article).not.toHaveAttribute('target');
  });

  it('refuses clearly with a link to the search results on no_source', async () => {
    setLanguages(['fr-FR']);
    member({ 'POST /assistant/answers': () => jsonResponse(noSource) });
    await renderApp('/recherche?q=momo');
    const inPanel = await panel();
    await userEvent
      .setup({ delay: null })
      .click(await inPanel.findByRole('button', { name: 'Demander à l’assistant' }));
    expect(await inPanel.findByText(/Aucune source pour répondre/)).toBeInTheDocument();
    expect(inPanel.getByRole('link', { name: 'Voir les résultats de recherche' })).toHaveAttribute(
      'href',
      '/recherche?q=momo',
    );
    expect(inPanel.queryByText('Comment vérifier')).toBeNull();
  });

  it('shows the wait from Retry-After on 429 and keeps the form', async () => {
    setLanguages(['fr-FR']);
    member({
      'POST /assistant/answers': () => problemResponse(429, 'RATE_LIMITED', { retryAfter: 42 }),
    });
    await renderApp('/recherche?q=momo');
    const inPanel = await panel();
    await userEvent
      .setup({ delay: null })
      .click(await inPanel.findByRole('button', { name: 'Demander à l’assistant' }));
    expect(await inPanel.findByText(/Réessayez dans 42 s/)).toBeInTheDocument();
    expect(inPanel.getByRole('button', { name: 'Demander à l’assistant' })).toBeInTheDocument();
  });

  it('on 503 says the assistant is unavailable once, never retries, then hides the panel', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({
      'POST /assistant/answers': () =>
        problemResponse(503, 'SERVICE_UNAVAILABLE', { retryAfter: 30 }),
    });
    const { router } = await renderApp('/recherche?q=momo');
    const inPanel = await panel();
    await userEvent
      .setup({ delay: null })
      .click(await inPanel.findByRole('button', { name: 'Demander à l’assistant' }));
    expect(await inPanel.findByText(/L’assistant est indisponible/)).toBeInTheDocument();
    expect(inPanel.queryByRole('button', { name: 'Demander à l’assistant' })).toBeNull();
    await new Promise((resolve) => setTimeout(resolve, 50));
    expect(callsTo(fetchMock, 'POST /assistant/answers')).toHaveLength(1);

    await router.navigate({ to: '/recherche', search: { q: 'webhook' } });
    await screen.findByRole('searchbox', { name: 'Votre recherche' });
    await waitFor(() => {
      expect(screen.queryByRole('heading', { level: 2, name: /^Assistant/ })).toBeNull();
    });
    expect(screen.queryByText(/indisponible/)).toBeNull();
  });

  it('brings the panel back once the Retry-After delay of a 503 has passed', async () => {
    setLanguages(['fr-FR']);
    member({
      'POST /assistant/answers': () =>
        problemResponse(503, 'SERVICE_UNAVAILABLE', { retryAfter: 30 }),
    });
    const { router } = await renderApp('/recherche?q=momo');
    const inPanel = await panel();
    await userEvent
      .setup({ delay: null })
      .click(await inPanel.findByRole('button', { name: 'Demander à l’assistant' }));
    await inPanel.findByText(/L’assistant est indisponible/);

    const now = Date.now();
    vi.spyOn(Date, 'now').mockReturnValue(now + 31_000);
    await router.navigate({ to: '/recherche', search: { q: 'webhook' } });
    expect(
      await screen.findByRole('heading', { level: 2, name: /^Assistant/ }),
    ).toBeInTheDocument();
  });

  it('invites an anonymous visitor to sign in and returns to the search', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ 'GET /search': () => jsonResponse(SEARCH) });
    await renderApp('/recherche?q=momo');
    const inPanel = await panel();
    const link = await inPanel.findByRole('link', { name: 'Se connecter' });
    expect(link).toHaveAttribute('href', '/connexion?next=%2Frecherche%3Fq%3Dmomo');
    expect(inPanel.queryByRole('button', { name: 'Demander à l’assistant' })).toBeNull();
    expect(callsTo(fetchMock, 'POST /assistant/answers')).toHaveLength(0);
  });

  it('shows the server refusal when the email is not verified', async () => {
    setLanguages(['fr-FR']);
    member({ 'POST /assistant/answers': () => problemResponse(403, 'EMAIL_NOT_VERIFIED') });
    await renderApp('/recherche?q=momo');
    const inPanel = await panel();
    await userEvent
      .setup({ delay: null })
      .click(await inPanel.findByRole('button', { name: 'Demander à l’assistant' }));
    expect(await inPanel.findByText('Confirmez d’abord votre adresse e-mail.')).toBeInTheDocument();
  });

  it('uses French typography and speaks English', async () => {
    setLanguages(['fr-FR']);
    member();
    const first = await renderApp('/recherche?q=momo');
    const inPanel = await panel();
    expect(inPanel.getByText(new RegExp(`Google \\(AI Studio\\)`))).toBeInTheDocument();
    expect(first.container.textContent).not.toMatch(/[^\u00A0\u202F][?!:;](?=\s|$)/);
    first.unmount();

    setLanguages(['en-US']);
    member();
    await renderApp('/recherche?q=momo');
    const en = await panel();
    expect(await en.findByRole('button', { name: 'Ask the assistant' })).toBeInTheDocument();
    expect(en.getByText(/sent to Google \(AI Studio\)/)).toBeInTheDocument();
  });
});
