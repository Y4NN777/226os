import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { callsTo, mockApi } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';
import { article, articleId, series, seriesId } from '../learn/testFixtures';
import { member, page, question } from './testFixtures';

// E4: « Poser une question sur cette leçon », from the lesson to the published question.
const lessonRoutes = {
  [`GET /learn/articles/${articleId}`]: () => jsonResponse(article()),
  [`GET /learn/series/${seriesId}`]: () => jsonResponse(series()),
  'GET /questions': () => page([]),
};

async function publish(user: ReturnType<typeof userEvent.setup>) {
  await user.type(await screen.findByRole('textbox', { name: 'Titre' }), 'Une question de cours');
  await user.type(screen.getByRole('textbox', { name: 'Votre question' }), 'Contexte.');
  await user.type(screen.getByRole('textbox', { name: /Étiquettes/ }), 'go');
  await user.click(screen.getByRole('button', { name: 'Publier la question' }));
}

describe('E4 ask about a lesson', () => {
  it('from an article: opens the form with the lesson linked, and createQuestion sends the link', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({
      ...lessonRoutes,
      'POST /questions': () => jsonResponse(question(), { status: 201 }),
    });
    const user = userEvent.setup();
    const { router } = await renderApp(`/apprendre/articles/${articleId}`);
    await user.click(
      await screen.findByRole('link', { name: 'Poser une question sur cette leçon' }),
    );

    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/solve');
    });
    expect(router.state.location.search).toMatchObject({ link: `article:${articleId}` });
    expect(await screen.findByRole('button', { name: 'Retirer le lien Article' })).toBeVisible();
    await publish(user);

    await waitFor(() => {
      expect(callsTo(fetchMock, 'POST /questions')).toHaveLength(1);
    });
    const [request] = callsTo(fetchMock, 'POST /questions');
    expect(request?.headers.get('X-CSRF-Token')).toBe('csrf-member');
    expect(request?.headers.get('Idempotency-Key')).toBeTruthy();
    const body = (await request?.clone().json()) as Record<string, unknown>;
    expect(body.links).toEqual([{ type: 'article', id: articleId }]);
  });

  it('from a series: links the first lesson, not the first one listed', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({
      ...lessonRoutes,
      'POST /questions': () => jsonResponse(question(), { status: 201 }),
    });
    const user = userEvent.setup();
    const { router } = await renderApp(`/apprendre/formations/${seriesId}`);
    await user.click(
      await screen.findByRole('link', { name: 'Poser une question sur la première leçon' }),
    );
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/solve');
    });
    await publish(user);
    await waitFor(() => {
      expect(callsTo(fetchMock, 'POST /questions')).toHaveLength(1);
    });
    const body = (await callsTo(fetchMock, 'POST /questions')[0]?.clone().json()) as Record<
      string,
      unknown
    >;
    expect(body.links).toEqual([{ type: 'article', id: articleId }]);
  });

  it('an anonymous visitor is taken to sign-in, and the link survives the round trip', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi(lessonRoutes);
    const user = userEvent.setup();
    const { router } = await renderApp(`/apprendre/articles/${articleId}`);
    await user.click(
      await screen.findByRole('link', { name: 'Poser une question sur cette leçon' }),
    );
    const invite = await screen.findByRole('link', {
      name: 'Se connecter pour poser une question',
    });
    expect(screen.queryByRole('textbox', { name: 'Titre' })).not.toBeInTheDocument();
    await user.click(invite);

    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/connexion');
    });
    expect(router.state.location.search).toMatchObject({
      next: `/solve?link=article:${articleId}`,
    });
    expect(callsTo(fetchMock, 'POST /questions')).toHaveLength(0);
  });
});
