import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import type { components } from '../../api/schema';
import { callsTo, memberSession, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';

type Article = components['schemas']['Article'];

const questionId = 'c0000000-0000-4000-8000-0000000000a1';
const articleId = 'c0000000-0000-4000-8000-0000000000b2';

const created: Article = {
  id: articleId,
  type: 'article',
  title: 'Vérifier un webhook signé',
  summary: '',
  body_md: 'Corps',
  lang: 'fr',
  level: 'beginner',
  tags: ['go', 'webhook'],
  author: {
    id: '6f1c2a34-0000-4000-8000-000000000001',
    username: 'awa',
    display_name: 'Awa Traoré',
  },
  related: [],
  status: 'published',
  published_at: '2026-11-14T10:00:00Z',
  updated_at: '2026-11-14T10:00:00Z',
};

const member = { 'GET /auth/session': () => jsonResponse(memberSession) };

function ok() {
  return jsonResponse(created, {
    status: 201,
    headers: { Location: `/api/v1/learn/articles/${articleId}` },
  });
}

/** Pastes rather than types: one event per field keeps these form tests fast on a loaded machine. */
async function fillIn(field: HTMLElement, text: string) {
  await userEvent.click(field);
  await userEvent.paste(text);
}

async function fillRequired(title = 'Vérifier un webhook signé') {
  await fillIn(await screen.findByRole('textbox', { name: /^Titre/ }), title);
  await fillIn(
    screen.getByRole('textbox', { name: /^Contenu/ }),
    'Calculez le HMAC puis comparez.',
  );
  await fillIn(screen.getByRole('textbox', { name: /^Mots-clés/ }), 'go, webhook');
}

async function publish() {
  await userEvent.click(screen.getByRole('button', { name: 'Publier' }));
}

async function bodyOf(request: Request) {
  return (await request.clone().json()) as Record<string, unknown>;
}

describe('ecrire page', () => {
  it('publishes as published with a JSON body, the CSRF token and an Idempotency-Key', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ ...member, 'POST /learn/articles': ok });
    await renderApp('/ecrire');
    await fillRequired();
    await publish();

    expect(await screen.findByRole('heading', { name: 'Votre article est publié' })).toBeVisible();
    const [request] = callsTo(fetchMock, 'POST /learn/articles');
    expect(request).toBeDefined();
    expect(request?.headers.get('Content-Type')).toContain('application/json');
    expect(request?.headers.get('X-CSRF-Token')).toBe('csrf-member');
    expect(request?.headers.get('Idempotency-Key')).toMatch(/^[A-Za-z0-9._~-]{16,128}$/);
    expect(await bodyOf(request as Request)).toEqual({
      type: 'article',
      title: 'Vérifier un webhook signé',
      body_md: 'Calculez le HMAC puis comparez.',
      lang: 'fr',
      level: 'beginner',
      tags: ['go', 'webhook'],
      status: 'published',
    });
  });

  it('keeps the same Idempotency-Key across a retry, and takes a new one for new content', async () => {
    setLanguages(['fr-FR']);
    let attempts = 0;
    const fetchMock = mockApi({
      ...member,
      'POST /learn/articles': () =>
        ++attempts === 1 ? problemResponse(503, 'SERVICE_UNAVAILABLE') : ok(),
    });
    await renderApp('/ecrire');
    await fillRequired();
    await publish();
    expect(await screen.findByRole('alert')).toBeVisible();
    await publish();
    await screen.findByRole('heading', { name: 'Votre article est publié' });

    const keys = callsTo(fetchMock, 'POST /learn/articles').map((r) =>
      r.headers.get('Idempotency-Key'),
    );
    expect(keys).toHaveLength(2);
    expect(keys[0]).toBe(keys[1]);
  });

  it('uses a different key once the content changed, since the server refuses a reused key', async () => {
    setLanguages(['fr-FR']);
    let attempts = 0;
    const fetchMock = mockApi({
      ...member,
      'POST /learn/articles': () =>
        ++attempts === 1 ? problemResponse(503, 'SERVICE_UNAVAILABLE') : ok(),
    });
    await renderApp('/ecrire');
    await fillRequired();
    await publish();
    await screen.findByRole('alert');
    await fillIn(screen.getByRole('textbox', { name: /^Titre/ }), ' (suite)');
    await publish();
    await screen.findByRole('heading', { name: 'Votre article est publié' });

    const keys = callsTo(fetchMock, 'POST /learn/articles').map((r) =>
      r.headers.get('Idempotency-Key'),
    );
    expect(keys[0]).not.toBe(keys[1]);
  });

  it('follows Location when the replay carries no body (Idempotency-Body-Omitted)', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({
      ...member,
      'POST /learn/articles': () =>
        new Response(null, {
          status: 201,
          headers: {
            Location: `/api/v1/learn/articles/${articleId}`,
            'Idempotency-Replayed': 'true',
            'Idempotency-Body-Omitted': 'true',
            'Content-Length': '0',
          },
        }),
      [`GET /learn/articles/${articleId}`]: () => jsonResponse(created),
    });
    await renderApp('/ecrire');
    await fillRequired();
    await publish();

    expect(await screen.findByRole('heading', { name: 'Votre article est publié' })).toBeVisible();
    expect(screen.getByText(/Vérifier un webhook signé/)).toBeVisible();
    expect(callsTo(fetchMock, `GET /learn/articles/${articleId}`)).toHaveLength(1);
  });

  it('never fetches a Location that is not an article of this API', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({
      ...member,
      'POST /learn/articles': () =>
        new Response(null, {
          status: 201,
          headers: {
            Location: 'https://evil.example/api/v1/learn/articles/x',
            'Idempotency-Body-Omitted': 'true',
            'Content-Length': '0',
          },
        }),
    });
    await renderApp('/ecrire');
    await fillRequired();
    await publish();

    expect(await screen.findByRole('alert')).toBeVisible();
    expect(
      fetchMock.mock.calls.some(([input]) =>
        String(input instanceof Request ? input.url : input).includes('evil'),
      ),
    ).toBe(false);
  });

  it('puts a 422 violation next to its field, with no generic error', async () => {
    setLanguages(['fr-FR']);
    mockApi({
      ...member,
      'POST /learn/articles': () =>
        problemResponse(422, 'VALIDATION_ERROR', {
          violations: [
            { field: '/title', code: 'too_short' },
            { field: '/summary', code: 'too_long' },
            { field: '/tags', code: 'too_many_items' },
            { field: '/topic', code: 'invalid_value' },
          ],
        }),
    });
    await renderApp('/ecrire');
    await fillRequired();
    await publish();

    expect(await screen.findByText('Le titre doit faire entre 5 et 200 caractères.')).toBeVisible();
    expect(screen.getByText('Le résumé ne doit pas dépasser 500 caractères.')).toBeVisible();
    expect(screen.getByText(/^Ajoutez entre 1 et 5 mots-clés/)).toBeVisible();
    expect(screen.getByText('Choisissez un sujet de la liste.')).toBeVisible();
    expect(screen.queryByText('La publication a échoué')).not.toBeInTheDocument();
    expect(screen.queryByRole('heading', { name: 'Votre article est publié' })).toBeNull();
  });

  it('checks the contract bounds before sending anything', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ ...member, 'POST /learn/articles': ok });
    await renderApp('/ecrire');
    await fillIn(await screen.findByRole('textbox', { name: /^Titre/ }), 'Go');
    await publish();

    expect(await screen.findByText('Le titre doit faire entre 5 et 200 caractères.')).toBeVisible();
    expect(screen.getByText(/^Écrivez le contenu de l’article/)).toBeVisible();
    expect(screen.getByText(/^Ajoutez entre 1 et 5 mots-clés/)).toBeVisible();
    expect(callsTo(fetchMock, 'POST /learn/articles')).toHaveLength(0);
  });

  it('sends the optional summary and topic, and the chosen type and level', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({ ...member, 'POST /learn/articles': ok });
    await renderApp('/ecrire');
    await fillRequired();
    await fillIn(screen.getByRole('textbox', { name: /^Résumé/ }), 'Un résumé court.');
    await userEvent.click(screen.getByRole('radio', { name: 'Tutoriel' }));
    await userEvent.click(screen.getByRole('radio', { name: 'Avancé' }));
    await userEvent.click(screen.getByRole('combobox', { name: /Sujet/ }));
    await userEvent.click(
      await screen.findByRole('option', { name: 'Construire et évaluer des agents' }),
    );
    await publish();
    await screen.findByRole('heading', { name: 'Votre article est publié' });

    const [request] = callsTo(fetchMock, 'POST /learn/articles');
    expect(await bodyOf(request as Request)).toMatchObject({
      type: 'tutorial',
      level: 'advanced',
      summary: 'Un résumé court.',
      topic: 'ai-agents',
    });
  });

  it('offers the 14 learn topics', async () => {
    setLanguages(['fr-FR']);
    mockApi(member);
    await renderApp('/ecrire');
    await userEvent.click(await screen.findByRole('combobox', { name: /Sujet/ }));
    const options = await screen.findAllByRole('option');
    expect(options).toHaveLength(14);
  });

  it('writes in the interface language by default and lets the author change it', async () => {
    setLanguages(['en-US']);
    const fetchMock = mockApi({ ...member, 'POST /learn/articles': ok });
    await renderApp('/ecrire');
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Write an article' }),
    ).toBeVisible();
    await fillIn(screen.getByRole('textbox', { name: /^Title/ }), 'Verify a signed webhook');
    await fillIn(screen.getByRole('textbox', { name: /^Content/ }), 'Compute the HMAC.');
    await fillIn(screen.getByRole('textbox', { name: /^Keywords/ }), 'go');
    await userEvent.click(screen.getByRole('button', { name: 'Publish' }));
    await screen.findByRole('heading', { name: 'Your article is published' });

    const [request] = callsTo(fetchMock, 'POST /learn/articles');
    expect(await bodyOf(request as Request)).toMatchObject({ lang: 'en' });
  });

  it('offers no draft, and says what publishing means', async () => {
    setLanguages(['fr-FR']);
    mockApi(member);
    await renderApp('/ecrire');
    await screen.findByRole('textbox', { name: /^Titre/ });
    expect(screen.queryByText(/brouillon/i)).not.toBeInTheDocument();
    expect(screen.queryByRole('radio', { name: /brouillon|draft/i })).not.toBeInTheDocument();
    expect(screen.getByText(/publié immédiatement/)).toBeVisible();
  });

  describe('links', () => {
    it('shows no link field without a pre-fill, and sends no links', async () => {
      setLanguages(['fr-FR']);
      const fetchMock = mockApi({ ...member, 'POST /learn/articles': ok });
      await renderApp('/ecrire');
      await fillRequired();
      expect(screen.queryByText(/^Lié à/)).not.toBeInTheDocument();
      expect(screen.queryByRole('textbox', { name: /lien|link|id|uuid/i })).not.toBeInTheDocument();
      await publish();
      await screen.findByRole('heading', { name: 'Votre article est publié' });

      const [request] = callsTo(fetchMock, 'POST /learn/articles');
      expect(await bodyOf(request as Request)).not.toHaveProperty('links');
    });

    it('shows a pre-filled question link as a removable chip and sends it', async () => {
      setLanguages(['fr-FR']);
      const fetchMock = mockApi({ ...member, 'POST /learn/articles': ok });
      await renderApp(`/ecrire?link=question:${questionId}`);
      await fillRequired();
      expect(screen.getByText('Question')).toBeVisible();
      expect(screen.queryByText(questionId)).not.toBeInTheDocument();
      await publish();
      await screen.findByRole('heading', { name: 'Votre article est publié' });

      const [request] = callsTo(fetchMock, 'POST /learn/articles');
      expect(await bodyOf(request as Request)).toMatchObject({
        links: [{ type: 'question', id: questionId }],
      });
    });

    it('drops the link when the chip is removed', async () => {
      setLanguages(['fr-FR']);
      const fetchMock = mockApi({ ...member, 'POST /learn/articles': ok });
      await renderApp(`/ecrire?link=resource:${questionId}`);
      await fillRequired();
      await userEvent.click(
        screen.getByRole('button', { name: 'Retirer le lien « API ou outil »' }),
      );
      expect(screen.queryByText('API ou outil')).not.toBeInTheDocument();
      await publish();
      await screen.findByRole('heading', { name: 'Votre article est publié' });

      const [request] = callsTo(fetchMock, 'POST /learn/articles');
      expect(await bodyOf(request as Request)).not.toHaveProperty('links');
    });

    it.each(['article', 'series', 'answer', 'question:nope'])(
      'ignores a pre-fill of the unsupported kind or shape %s',
      async (link) => {
        setLanguages(['fr-FR']);
        const fetchMock = mockApi({ ...member, 'POST /learn/articles': ok });
        const value = link.includes(':') ? link : `${link}:${questionId}`;
        await renderApp(`/ecrire?link=${value}`);
        await fillRequired();
        expect(screen.queryByText(/^Lié à/)).not.toBeInTheDocument();
        await publish();
        await screen.findByRole('heading', { name: 'Votre article est publié' });

        const [request] = callsTo(fetchMock, 'POST /learn/articles');
        expect(await bodyOf(request as Request)).not.toHaveProperty('links');
      },
    );

    it('says so next to the chip when the server refuses the link', async () => {
      setLanguages(['fr-FR']);
      mockApi({
        ...member,
        'POST /learn/articles': () =>
          problemResponse(422, 'VALIDATION_ERROR', {
            violations: [{ field: '/links/0', code: 'invalid_value' }],
          }),
      });
      await renderApp(`/ecrire?link=project:${questionId}`);
      await fillRequired();
      await publish();

      expect(
        await screen.findByText('Ce lien n’est pas accepté. Retirez-le, puis publiez à nouveau.'),
      ).toBeVisible();
    });
  });

  describe('who can write', () => {
    it('invites a visitor to sign in, and brings them back here afterwards', async () => {
      setLanguages(['fr-FR']);
      mockApi();
      await renderApp(`/ecrire?link=question:${questionId}`);

      expect(
        await screen.findByRole('heading', { name: 'Connectez-vous pour écrire' }),
      ).toBeVisible();
      const signIn = screen
        .getAllByRole('link', { name: 'Se connecter' })
        .find((link) => link.getAttribute('href')?.includes('next='));
      const next = new URL(signIn?.getAttribute('href') ?? '', 'http://x').searchParams.get('next');
      // The router may escape the colon in the query; the destination is the same.
      const back = new URL(next ?? '', 'http://x');
      expect(back.pathname).toBe('/ecrire');
      expect(back.searchParams.get('link')).toBe(`question:${questionId}`);
      expect(screen.queryByRole('textbox', { name: /^Titre/ })).not.toBeInTheDocument();
    });

    it('tells a member without a confirmed email to confirm it', async () => {
      setLanguages(['fr-FR']);
      mockApi({
        ...member,
        'POST /learn/articles': () => problemResponse(403, 'EMAIL_NOT_VERIFIED'),
      });
      await renderApp('/ecrire');
      await fillRequired();
      await publish();

      expect(await screen.findByText('La publication a échoué')).toBeVisible();
      expect(screen.getByText('Confirmez d’abord votre adresse e-mail.')).toBeVisible();
      expect(screen.getByRole('textbox', { name: /^Titre/ })).toHaveValue(
        'Vérifier un webhook signé',
      );
    });
  });

  it('is reachable from the account menu of a signed-in member', async () => {
    setLanguages(['fr-FR']);
    mockApi(member);
    await renderApp('/');
    await userEvent.click(await screen.findByRole('button', { name: 'Awa Traoré' }));
    expect(await screen.findByRole('menuitem', { name: 'Écrire un article' })).toHaveAttribute(
      'href',
      '/ecrire',
    );
  });

  it('lets the author write another article after publishing', async () => {
    setLanguages(['fr-FR']);
    mockApi({ ...member, 'POST /learn/articles': ok });
    await renderApp('/ecrire');
    await fillRequired();
    await publish();
    await screen.findByRole('heading', { name: 'Votre article est publié' });
    expect(screen.getByRole('link', { name: 'Lire l’article' })).toHaveAttribute(
      'href',
      `/apprendre/articles/${articleId}`,
    );
    await userEvent.click(screen.getByRole('button', { name: 'Écrire un autre article' }));
    expect(await screen.findByRole('textbox', { name: /^Titre/ })).toHaveValue('');
  });
});
