import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { anonymousSession, callsTo, memberSession, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';
import { answer, member, memberId, page, question, questionId, withEtag } from './testFixtures';
import { findVisibleText } from '../../test/visibleText';

const base = `/questions/${questionId}`;
const key = (suffix = '') => `${base}${suffix}`;
const accepted = answer({
  id: 'c0000000-0000-4000-8000-000000000002',
  body_md: 'La réponse retenue.',
  is_accepted: true,
  author: {
    id: 'b0000000-0000-4000-8000-0000000000bb',
    username: 'ibrahim',
    display_name: 'Ibrahim Traoré',
  },
});
const plain = answer({ body_md: 'Une autre piste.' });

async function expectAlert(text: string) {
  await waitFor(() => {
    const alerts = screen.getAllByRole('alert').map((alert) => alert.textContent);
    expect(alerts.join(' ')).toContain(text);
  });
}

function readRoutes(overrides: Record<string, () => Response> = {}) {
  return {
    [`GET ${key()}`]: () => withEtag(question(), '"v1"'),
    [`GET ${key('/answers')}`]: () => page([plain, accepted]),
    ...overrides,
  };
}

describe('QuestionPage reading', () => {
  it('shows the question, its tags and snippet, and the answers with the accepted one first', async () => {
    setLanguages(['fr-FR']);
    mockApi(
      readRoutes({
        [`GET ${key()}`]: () =>
          withEtag(
            question({
              state: 'resolved',
              snippets: [{ language: 'go', code: 'http.ListenAndServe(":80", nil)' }],
              related: [
                { type: 'resource', id: 'd0000000-0000-4000-8000-000000000001', title: 'Wave Pay' },
              ],
            }),
            '"v1"',
          ),
      }),
    );
    await renderApp(base);

    expect(
      await screen.findByRole('heading', { level: 1, name: 'Webhook non reçu après un paiement' }),
    ).toBeInTheDocument();
    expect(screen.getAllByText('Par Kofi Mensah').length).toBeGreaterThan(0);
    expect(screen.getByText('webhook')).toBeInTheDocument();
    // The code block splits the code into highlighted spans: read the text as a whole.
    expect(document.body.textContent).toContain('http.ListenAndServe(":80", nil)');
    expect(screen.getByRole('link', { name: 'Wave Pay' })).toHaveAttribute(
      'href',
      '/discover/d0000000-0000-4000-8000-000000000001',
    );

    const list = await screen.findByRole('list', { name: 'Réponses' });
    const items = within(list).getAllByRole('listitem');
    expect(items[0]).toHaveTextContent('La réponse retenue.');
    expect(items[0]).toHaveTextContent('Réponse acceptée');
    expect(items[1]).toHaveTextContent('Une autre piste.');
    expect(items[1]).not.toHaveTextContent('Réponse acceptée');
  });

  it('never turns raw HTML in a body into elements', async () => {
    setLanguages(['fr-FR']);
    mockApi(
      readRoutes({
        [`GET ${key()}`]: () =>
          withEtag(
            question({
              body_md: 'Avant <img src=x onerror="alert(1)"> <script>alert(2)</script> après',
            }),
            '"v1"',
          ),
      }),
    );
    const { container } = await renderApp(base);
    await screen.findByRole('heading', { level: 1 });
    await screen.findByRole('list', { name: 'Réponses' });
    expect(container.querySelector('img, script, [onerror]')).toBeNull();
  });

  it('shows a not-found state for an unknown question', async () => {
    setLanguages(['fr-FR']);
    mockApi({ [`GET ${key()}`]: () => problemResponse(404, 'NOT_FOUND') });
    await renderApp(base);
    expect(await findVisibleText('Question introuvable')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Retour aux questions' })).toHaveAttribute(
      'href',
      '/solve',
    );
  });

  it('says when nobody has answered yet', async () => {
    setLanguages(['fr-FR']);
    mockApi(readRoutes({ [`GET ${key('/answers')}`]: () => page([]) }));
    await renderApp(base);
    expect(await findVisibleText(/Aucune réponse pour le moment/)).toBeInTheDocument();
  });

  it('is written in English when the browser asks for it', async () => {
    setLanguages(['en-GB']);
    mockApi(readRoutes());
    await renderApp(base);
    expect(await screen.findByRole('list', { name: 'Answers' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Sign in to answer' })).toBeInTheDocument();
  });
});

describe('QuestionPage answering', () => {
  it('invites an anonymous visitor to sign in, coming back to this question', async () => {
    setLanguages(['fr-FR']);
    mockApi(readRoutes());
    await renderApp(base);
    const link = await screen.findByRole('link', { name: 'Se connecter pour répondre' });
    expect(link).toHaveAttribute('href', `/connexion?next=${encodeURIComponent(base)}`);
    expect(
      screen.queryByRole('textbox', { name: 'Rédigez votre réponse' }),
    ).not.toBeInTheDocument();
  });

  it('posts an answer with an Idempotency-Key and shows it after the refresh', async () => {
    setLanguages(['fr-FR']);
    let posted = false;
    const fetchMock = member(
      readRoutes({
        [`GET ${key('/answers')}`]: () =>
          page(
            posted
              ? [
                  plain,
                  answer({
                    id: 'c0000000-0000-4000-8000-000000000009',
                    body_md: 'Ma contribution.',
                  }),
                ]
              : [plain],
          ),
        [`POST ${key('/answers')}`]: () => {
          posted = true;
          return jsonResponse(answer({ body_md: 'Ma contribution.' }), { status: 201 });
        },
      }),
    );
    const user = userEvent.setup();
    await renderApp(base);
    await user.type(
      await screen.findByRole('textbox', { name: 'Rédigez votre réponse' }),
      'Ma contribution.',
    );
    await user.click(screen.getByRole('button', { name: 'Publier la réponse' }));

    expect(await screen.findByText('Ma contribution.')).toBeInTheDocument();
    const post = callsTo(fetchMock, `POST ${key('/answers')}`)[0];
    expect(post?.headers.get('Idempotency-Key')).toBeTruthy();
    expect(post?.headers.get('X-CSRF-Token')).toBe('csrf-member');
    expect(await post?.clone().json()).toEqual({ body_md: 'Ma contribution.', lang: 'fr' });
    expect(screen.getByRole('textbox', { name: 'Rédigez votre réponse' })).toHaveValue('');
  });

  it('shows the 403 EMAIL_NOT_VERIFIED message and keeps the text', async () => {
    setLanguages(['fr-FR']);
    member(
      readRoutes({ [`POST ${key('/answers')}`]: () => problemResponse(403, 'EMAIL_NOT_VERIFIED') }),
    );
    const user = userEvent.setup();
    await renderApp(base);
    await user.type(
      await screen.findByRole('textbox', { name: 'Rédigez votre réponse' }),
      'Brouillon',
    );
    await user.click(screen.getByRole('button', { name: 'Publier la réponse' }));
    await expectAlert('Confirmez d’abord votre adresse e-mail.');
    expect(screen.getByRole('textbox', { name: 'Rédigez votre réponse' })).toHaveValue('Brouillon');
  });

  it('puts a 422 violation on the answer field', async () => {
    setLanguages(['fr-FR']);
    member(
      readRoutes({
        [`POST ${key('/answers')}`]: () =>
          problemResponse(422, 'VALIDATION_ERROR', {
            violations: [{ field: '/body_md', code: 'too_long' }],
          }),
      }),
    );
    const user = userEvent.setup();
    await renderApp(base);
    await user.type(await screen.findByRole('textbox', { name: 'Rédigez votre réponse' }), 'Trop');
    await user.click(screen.getByRole('button', { name: 'Publier la réponse' }));
    expect(await screen.findByText(/20 000 caractères au maximum/)).toBeVisible();
    expect(screen.getByRole('textbox', { name: 'Rédigez votre réponse' })).toBeInvalid();
  });
});

describe('QuestionPage editing', () => {
  const owned = () =>
    withEtag(
      question({ author: { id: memberId, username: 'awa', display_name: 'Awa Traoré' } }),
      '"v1"',
    );

  it('offers the edit to the author only', async () => {
    setLanguages(['fr-FR']);
    member(readRoutes());
    await renderApp(base);
    await screen.findByRole('heading', { level: 1 });
    expect(screen.queryByRole('button', { name: 'Modifier la question' })).not.toBeInTheDocument();
  });

  it('patches only what changed with the ETag in If-Match, then shows the new text', async () => {
    setLanguages(['fr-FR']);
    let title = 'Webhook non reçu après un paiement';
    const fetchMock = member(
      readRoutes({
        [`GET ${key()}`]: () =>
          withEtag(
            question({
              title,
              author: { id: memberId, username: 'awa', display_name: 'Awa Traoré' },
            }),
            '"v1"',
          ),
        [`PATCH ${key()}`]: () => {
          title = 'Webhook toujours non reçu';
          return withEtag(
            question({
              title,
              author: { id: memberId, username: 'awa', display_name: 'Awa Traoré' },
            }),
            '"v2"',
          );
        },
      }),
    );
    const user = userEvent.setup();
    await renderApp(base);
    await user.click(await screen.findByRole('button', { name: 'Modifier la question' }));
    const field = screen.getByRole('textbox', { name: 'Titre' });
    await user.clear(field);
    await user.type(field, 'Webhook toujours non reçu');
    await user.click(screen.getByRole('button', { name: 'Enregistrer' }));

    expect(
      await screen.findByRole('heading', { level: 1, name: 'Webhook toujours non reçu' }),
    ).toBeInTheDocument();
    const patch = callsTo(fetchMock, `PATCH ${key()}`)[0];
    expect(patch?.headers.get('If-Match')).toBe('"v1"');
    expect(patch?.headers.get('X-CSRF-Token')).toBe('csrf-member');
    expect(await patch?.clone().json()).toEqual({ title: 'Webhook toujours non reçu' });
  });

  it('edits with the signed-in caller’s ETag after loading the question anonymously', async () => {
    setLanguages(['fr-FR']);
    let signedIn = false;
    const mine = { id: memberId, username: 'awa', display_name: 'Awa Traoré' };
    const fetchMock = mockApi({
      'GET /auth/session': () => jsonResponse(signedIn ? memberSession : anonymousSession),
      'POST /auth/sign-in': () => {
        signedIn = true;
        return jsonResponse(memberSession);
      },
      [`GET ${key()}`]: () =>
        withEtag(question({ author: mine }), signedIn ? '"member-v1"' : '"anon-v1"'),
      [`GET ${key('/answers')}`]: () => page([]),
      [`PATCH ${key()}`]: (request) =>
        request.headers.get('If-Match') === '"member-v1"'
          ? withEtag(question({ author: mine, title: 'Webhook toujours non reçu' }), '"member-v2"')
          : problemResponse(412, 'PRECONDITION_FAILED'),
    });
    const user = userEvent.setup();
    const { router } = await renderApp(base);
    await screen.findByRole('heading', { level: 1, name: 'Webhook non reçu après un paiement' });

    await router.navigate({ to: '/connexion', search: { next: base } });
    await user.type(await screen.findByLabelText('Adresse e-mail'), 'awa@example.org');
    await user.type(screen.getByLabelText('Mot de passe'), 'a long passphrase');
    await user.click(screen.getByRole('button', { name: 'Se connecter' }));

    await user.click(await screen.findByRole('button', { name: 'Modifier la question' }));
    const field = screen.getByRole('textbox', { name: 'Titre' });
    await user.clear(field);
    await user.type(field, 'Webhook toujours non reçu');
    await user.click(screen.getByRole('button', { name: 'Enregistrer' }));

    expect(
      await screen.findByRole('heading', { level: 1, name: 'Webhook toujours non reçu' }),
    ).toBeInTheDocument();
    const patches = callsTo(fetchMock, `PATCH ${key()}`);
    expect(patches).toHaveLength(1);
    expect(patches[0]?.headers.get('If-Match')).toBe('"member-v1"');
  });

  it('shows a tag a moderator has hidden on the tags field, in plain French', async () => {
    setLanguages(['fr-FR']);
    const mine = { id: memberId, username: 'awa', display_name: 'Awa Traoré' };
    member(
      readRoutes({
        [`GET ${key()}`]: () => withEtag(question({ author: mine }), '"v1"'),
        [`PATCH ${key()}`]: () =>
          problemResponse(422, 'VALIDATION_ERROR', {
            violations: [{ field: '/tags', code: 'invalid_value' }],
          }),
      }),
    );
    const user = userEvent.setup();
    await renderApp(base);
    await user.click(await screen.findByRole('button', { name: 'Modifier la question' }));
    const tags = screen.getByRole('textbox', { name: /Étiquettes/ });
    await user.type(tags, ', ancienne');
    await user.click(screen.getByRole('button', { name: 'Enregistrer' }));

    expect(
      await screen.findByText(
        'Une des étiquettes n’est plus disponible. Retirez-la ou remplacez-la.',
      ),
    ).toBeVisible();
    expect(screen.getByRole('textbox', { name: /Étiquettes/ })).toBeInvalid();
  });

  it('offers a reload on 412 and resends the edit with the fresh ETag', async () => {
    setLanguages(['fr-FR']);
    let etag = '"v1"';
    const fetchMock = member(
      readRoutes({
        [`GET ${key()}`]: () =>
          withEtag(
            question({ author: { id: memberId, username: 'awa', display_name: 'Awa Traoré' } }),
            etag,
          ),
        [`PATCH ${key()}`]: (() => {
          let attempts = 0;
          return () => {
            attempts += 1;
            if (attempts === 1) {
              etag = '"v2"';
              return problemResponse(412, 'PRECONDITION_FAILED');
            }
            return withEtag(question({ title: 'Titre corrigé' }), '"v3"');
          };
        })(),
      }),
    );
    const user = userEvent.setup();
    await renderApp(base);
    await user.click(await screen.findByRole('button', { name: 'Modifier la question' }));
    const field = screen.getByRole('textbox', { name: 'Titre' });
    await user.clear(field);
    await user.type(field, 'Titre corrigé');
    await user.click(screen.getByRole('button', { name: 'Enregistrer' }));

    expect(await findVisibleText('Question modifiée entre-temps')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Recharger' }));
    await waitFor(() => {
      expect(callsTo(fetchMock, `GET ${key()}`)).toHaveLength(2);
    });
    // The edit in progress survives the reload.
    expect(screen.getByRole('textbox', { name: 'Titre' })).toHaveValue('Titre corrigé');
    await user.click(screen.getByRole('button', { name: 'Enregistrer' }));
    await waitFor(() => {
      expect(callsTo(fetchMock, `PATCH ${key()}`)).toHaveLength(2);
    });
    expect(callsTo(fetchMock, `PATCH ${key()}`)[1]?.headers.get('If-Match')).toBe('"v2"');
    expect(owned).toBeDefined();
  });

  it('keeps Save disabled until something changes, so an empty edit never closes silently', async () => {
    setLanguages(['fr-FR']);
    member(
      readRoutes({
        [`GET ${key()}`]: () =>
          withEtag(
            question({ author: { id: memberId, username: 'awa', display_name: 'Awa Traoré' } }),
            '"v1"',
          ),
      }),
    );
    const user = userEvent.setup();
    await renderApp(base);
    await user.click(await screen.findByRole('button', { name: 'Modifier la question' }));
    expect(screen.getByRole('button', { name: 'Enregistrer' })).toHaveAttribute(
      'aria-disabled',
      'true',
    );
    await user.type(screen.getByRole('textbox', { name: 'Titre' }), ' !');
    expect(screen.getByRole('button', { name: 'Enregistrer' })).not.toHaveAttribute(
      'aria-disabled',
    );
  });

  it('offers the reload prompt, and sends nothing, when a reload came back without an ETag', async () => {
    setLanguages(['fr-FR']);
    const author = { id: memberId, username: 'awa', display_name: 'Awa Traoré' };
    let reads = 0;
    const fetchMock = member(
      readRoutes({
        [`GET ${key()}`]: () => {
          reads += 1;
          return reads === 1
            ? withEtag(question({ author }), '"v1"')
            : jsonResponse(question({ author }));
        },
        [`PATCH ${key()}`]: () => problemResponse(412, 'PRECONDITION_FAILED'),
      }),
    );
    const user = userEvent.setup();
    await renderApp(base);
    await user.click(await screen.findByRole('button', { name: 'Modifier la question' }));
    await user.type(screen.getByRole('textbox', { name: 'Titre' }), ' !');
    await user.click(screen.getByRole('button', { name: 'Enregistrer' }));
    await user.click(await screen.findByRole('button', { name: 'Recharger' }));
    await waitFor(() => {
      expect(reads).toBe(2);
    });
    await user.click(screen.getByRole('button', { name: 'Enregistrer' }));
    expect(await findVisibleText('Question modifiée entre-temps')).toBeInTheDocument();
    expect(screen.getByRole('textbox', { name: 'Titre' })).toBeInTheDocument();
    expect(callsTo(fetchMock, `PATCH ${key()}`)).toHaveLength(1);
  });
});

describe('QuestionPage votes and acceptance', () => {
  const owner = { id: memberId, username: 'awa', display_name: 'Awa Traoré' };
  const ownedRoutes = (extra: Record<string, () => Response> = {}) =>
    readRoutes({ [`GET ${key()}`]: () => withEtag(question({ author: owner }), '"v1"'), ...extra });

  it('sends no probe: controls are shown to a member without any vote call', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member(readRoutes());
    await renderApp(base);
    expect(
      (await screen.findAllByRole('button', { name: 'Vote positif sur cette réponse' })).length,
    ).toBeGreaterThan(0);
    expect(callsTo(fetchMock, 'DELETE /votes')).toHaveLength(0);
    expect(callsTo(fetchMock, 'POST /votes')).toHaveLength(0);
  });

  it('hides vote and accept controls after a 501, with no error shown', async () => {
    setLanguages(['fr-FR']);
    member(ownedRoutes({ 'POST /votes': () => new Response(null, { status: 501 }) }));
    const user = userEvent.setup();
    await renderApp(base);
    const up = await screen.findAllByRole('button', { name: 'Vote positif sur cette réponse' });
    expect(
      screen.getAllByRole('button', { name: 'Accepter cette réponse' }).length,
    ).toBeGreaterThan(0);
    await user.click(up[0] as HTMLElement);

    await waitFor(() => {
      expect(screen.queryByRole('button', { name: /Vote positif/ })).not.toBeInTheDocument();
    });
    expect(
      screen.queryByRole('button', { name: /Accepter cette réponse/ }),
    ).not.toBeInTheDocument();
    // Empty live regions may exist; what matters is that no message is shown.
    expect(
      screen
        .queryAllByRole('alert')
        .map((alert) => alert.textContent.trim())
        .filter(Boolean),
    ).toEqual([]);
    expect(screen.getAllByText(/Score :/).length).toBeGreaterThan(0);
  });

  it('hides the controls when accepting answers 501', async () => {
    setLanguages(['fr-FR']);
    member(ownedRoutes({ [`POST ${key('/accept')}`]: () => new Response(null, { status: 501 }) }));
    const user = userEvent.setup();
    await renderApp(base);
    const buttons = await screen.findAllByRole('button', { name: 'Accepter cette réponse' });
    await user.click(buttons[0] as HTMLElement);
    await waitFor(() => {
      expect(screen.queryByRole('button', { name: /Vote positif/ })).not.toBeInTheDocument();
    });
    expect(
      screen.queryByRole('button', { name: /Accepter cette réponse/ }),
    ).not.toBeInTheDocument();
    // Empty live regions may exist; what matters is that no message is shown.
    expect(
      screen
        .queryAllByRole('alert')
        .map((alert) => alert.textContent.trim())
        .filter(Boolean),
    ).toEqual([]);
  });

  it('casts a vote and shows the score the server returned', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member(
      readRoutes({
        'POST /votes': () =>
          jsonResponse({ target_type: 'answer', target_id: plain.id, value: 1, score: 5 }),
      }),
    );
    const user = userEvent.setup();
    await renderApp(base);
    await screen.findByText('Une autre piste.');
    const up = await screen.findAllByRole('button', { name: 'Vote positif sur cette réponse' });
    await user.click(up[1] as HTMLElement);

    expect(await screen.findByText('Score : 5')).toBeInTheDocument();
    const vote = callsTo(fetchMock, 'POST /votes')[0];
    expect(await vote?.clone().json()).toEqual({
      target_type: 'answer',
      target_id: plain.id,
      value: 1,
    });
    expect(up[1]).toHaveAttribute('aria-pressed', 'true');
  });

  it('starts from the vote the server says the viewer already cast, and a press retracts it', async () => {
    setLanguages(['fr-FR']);
    const voted = answer({ body_md: 'Une autre piste.', viewer_vote: 1, vote_score: 4 });
    let answerReads = 0;
    const fetchMock = member(
      readRoutes({
        [`GET ${key('/answers')}`]: () => {
          answerReads += 1;
          return page([voted, accepted]);
        },
        'DELETE /votes': () =>
          jsonResponse({ target_type: 'answer', target_id: voted.id, value: 0, score: 3 }),
      }),
    );
    const user = userEvent.setup();
    await renderApp(base);
    await screen.findByText('Une autre piste.');
    const up = (
      await screen.findAllByRole('button', { name: 'Vote positif sur cette réponse' })
    )[1];
    expect(up).toHaveAttribute('aria-pressed', 'true');

    await user.click(up as HTMLElement);

    expect(await screen.findByText('Score : 3')).toBeInTheDocument();
    expect(callsTo(fetchMock, 'DELETE /votes')).toHaveLength(1);
    expect(callsTo(fetchMock, 'POST /votes')).toHaveLength(0);
    expect(up).toHaveAttribute('aria-pressed', 'false');
    await waitFor(() => {
      expect(answerReads).toBeGreaterThan(1);
    });
  });

  it('shows an existing downvote on the question', async () => {
    setLanguages(['fr-FR']);
    member(
      readRoutes({
        [`GET ${key()}`]: () => withEtag(question({ viewer_vote: -1 }), '"v1"'),
      }),
    );
    await renderApp(base);
    expect(
      await screen.findByRole('button', { name: 'Vote négatif sur la question' }),
    ).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: 'Vote positif sur la question' })).toHaveAttribute(
      'aria-pressed',
      'false',
    );
  });

  it('lets the question author accept an answer and refreshes', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member(
      ownedRoutes({
        [`POST ${key('/accept')}`]: () =>
          jsonResponse(question({ author: owner, accepted_answer_id: plain.id })),
      }),
    );
    const user = userEvent.setup();
    await renderApp(base);
    const buttons = await screen.findAllByRole('button', { name: 'Accepter cette réponse' });
    await user.click(buttons[0] as HTMLElement);
    await waitFor(() => {
      expect(callsTo(fetchMock, `POST ${key('/accept')}`)).toHaveLength(1);
    });
    expect(
      await callsTo(fetchMock, `POST ${key('/accept')}`)[0]
        ?.clone()
        .json(),
    ).toEqual({ answer_id: expect.any(String) as string });
    expect(screen.getByRole('button', { name: 'Retirer l’acceptation' })).toBeInTheDocument();
  });

  it('does not offer accept to someone who is not the author', async () => {
    setLanguages(['fr-FR']);
    member(readRoutes());
    await renderApp(base);
    await screen.findAllByRole('button', { name: 'Vote positif sur cette réponse' });
    expect(
      screen.queryByRole('button', { name: /Accepter cette réponse/ }),
    ).not.toBeInTheDocument();
  });

  it('explains a refused vote on one’s own content', async () => {
    setLanguages(['fr-FR']);
    member(readRoutes({ 'POST /votes': () => problemResponse(403, 'SELF_VOTE_FORBIDDEN') }));
    const user = userEvent.setup();
    await renderApp(base);
    const up = await screen.findAllByRole('button', { name: 'Vote positif sur cette réponse' });
    await user.click(up[0] as HTMLElement);
    await expectAlert('Vous ne pouvez pas voter pour votre propre contenu.');
  });
});
