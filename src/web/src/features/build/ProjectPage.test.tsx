import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { callsTo, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';
import { selectSegment } from '../../test/segmented';
import {
  application,
  applicationId,
  member,
  need,
  needId,
  owner,
  page,
  project,
  projectId,
} from './testFixtures';

const base = `/projects/${projectId}`;

async function expectAlert(text: string) {
  await waitFor(() => {
    const alerts = screen.getAllByRole('alert').map((alert) => alert.textContent);
    expect(alerts.join(' ')).toContain(text);
  });
}

describe('ProjectPage read view', () => {
  it('shows the project, its repository link, owner and open needs, to anyone', async () => {
    setLanguages(['fr-FR']);
    mockApi({
      [`GET ${base}`]: () => jsonResponse(project()),
      [`GET ${base}/contribution-needs`]: () =>
        page([
          need(),
          need({ id: 'e1000000-0000-4000-8000-000000000002', title: 'Docs', status: 'closed' }),
        ]),
    });
    await renderApp(`/build/${projectId}`);

    expect(await screen.findByRole('heading', { level: 1, name: 'kora-sdk' })).toBeVisible();
    expect(screen.getByText('Un SDK pour les paiements mobiles.')).toBeVisible();
    expect(screen.getByRole('link', { name: /Voir le dépôt sur GitHub/ })).toHaveAttribute(
      'href',
      'https://github.com/kora-dev/kora-sdk',
    );
    expect(screen.getByRole('link', { name: 'Kofi Mensah' })).toHaveAttribute(
      'href',
      '/profils/kofi',
    );
    const needs = await screen.findByRole('region', { name: 'Besoins de contribution' });
    expect(within(needs).getByText('Tests d’intégration')).toBeVisible();
    expect(within(needs).getByText('Ouvert')).toBeVisible();
    expect(within(needs).getByText('Fermé')).toBeVisible();
  });

  it('shows the same not-found state for an unknown and a private project', async () => {
    setLanguages(['fr-FR']);
    mockApi({ [`GET ${base}`]: () => problemResponse(404, 'NOT_FOUND') });
    await renderApp(`/build/${projectId}`);
    expect(await screen.findByText('Projet introuvable')).toBeVisible();
    expect(screen.getByText('Ce projet n’existe pas ou n’est pas public.')).toBeVisible();
  });

  it('shows a retryable error when the project cannot be loaded', async () => {
    setLanguages(['fr-FR']);
    mockApi({ [`GET ${base}`]: () => problemResponse(500, 'INTERNAL_ERROR') });
    await renderApp(`/build/${projectId}`);
    await expectAlert('Une erreur est survenue');
    expect(screen.getByRole('button', { name: 'Réessayer' })).toBeVisible();
  });

  it('says so when there is no need', async () => {
    setLanguages(['fr-FR']);
    mockApi({
      [`GET ${base}`]: () => jsonResponse(project()),
      [`GET ${base}/contribution-needs`]: () => page([]),
    });
    await renderApp(`/build/${projectId}`);
    expect(await screen.findByText('Aucun besoin pour le moment.')).toBeVisible();
  });

  it('pages needs with the cursor', async () => {
    setLanguages(['fr-FR']);
    mockApi({
      [`GET ${base}`]: () => jsonResponse(project()),
      [`GET ${base}/contribution-needs`]: (request) =>
        new URL(request.url).searchParams.get('cursor') === 'c2'
          ? page([need({ id: 'e1000000-0000-4000-8000-000000000003', title: 'Traduction' })])
          : page([need()], 'c2'),
    });
    const user = userEvent.setup();
    await renderApp(`/build/${projectId}`);
    await user.click(await screen.findByRole('button', { name: 'Afficher plus' }));
    expect(await screen.findByText('Traduction')).toBeVisible();
  });
});

describe('ProjectPage repository link', () => {
  it('renders the repository as a link only for a github.com address', async () => {
    setLanguages(['fr-FR']);
    mockApi({
      [`GET ${base}`]: () => jsonResponse(project({ github_url: 'javascript:alert(1)' })),
      [`GET ${base}/contribution-needs`]: () => page([]),
    });
    await renderApp(`/build/${projectId}`);
    await screen.findByRole('heading', { level: 1, name: 'kora-sdk' });
    expect(
      screen.queryByRole('link', { name: /Voir le dépôt sur GitHub/ }),
    ).not.toBeInTheDocument();
  });
});

describe('ProjectPage visibility of controls', () => {
  const routes = {
    [`GET ${base}`]: () => jsonResponse(project()),
    [`GET ${base}/contribution-needs`]: () => page([need()]),
  };

  it('hides every owner-only control from a non-owner and never asks for applications', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member(routes);
    await renderApp(`/build/${projectId}`);
    await screen.findByText('Tests d’intégration');
    expect(screen.queryByRole('button', { name: 'Modifier le projet' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Ajouter un besoin' })).not.toBeInTheDocument();
    expect(screen.queryByRole('region', { name: 'Candidatures' })).not.toBeInTheDocument();
    expect(callsTo(fetchMock, `GET ${base}/applications`)).toHaveLength(0);
  });

  it('shows the owner controls and hides the apply form from the owner', async () => {
    setLanguages(['fr-FR']);
    owner({ ...routes, [`GET ${base}/applications`]: () => page([]) });
    await renderApp(`/build/${projectId}`);
    expect(await screen.findByRole('button', { name: 'Modifier le projet' })).toBeVisible();
    expect(screen.getByRole('button', { name: 'Ajouter un besoin' })).toBeVisible();
    expect(screen.queryByRole('button', { name: 'Postuler' })).not.toBeInTheDocument();
  });

  it('invites an anonymous visitor to sign in to apply, with a safe next', async () => {
    setLanguages(['fr-FR']);
    mockApi(routes);
    await renderApp(`/build/${projectId}`);
    expect(await screen.findByRole('link', { name: 'Se connecter pour postuler' })).toHaveAttribute(
      'href',
      `/connexion?next=${encodeURIComponent(`/build/${projectId}`)}`,
    );
  });

  it('does not offer to apply to a closed need', async () => {
    setLanguages(['fr-FR']);
    member({
      ...routes,
      [`GET ${base}/contribution-needs`]: () => page([need({ status: 'closed' })]),
    });
    await renderApp(`/build/${projectId}`);
    await screen.findByText('Fermé');
    expect(screen.queryByRole('button', { name: 'Postuler' })).not.toBeInTheDocument();
  });
});

describe('ProjectPage apply', () => {
  const routes = {
    [`GET ${base}`]: () => jsonResponse(project()),
    [`GET ${base}/contribution-needs`]: () => page([need()]),
  };
  const apply = async () => {
    const user = userEvent.setup();
    await renderApp(`/build/${projectId}`);
    await user.click(await screen.findByRole('button', { name: 'Postuler' }));
    return user;
  };

  it('requires a message and sends nothing when it is empty', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member(routes);
    const user = await apply();
    await user.click(screen.getByRole('button', { name: 'Envoyer ma candidature' }));
    expect(await screen.findByText(/Écrivez un message de 1 à 2/)).toBeVisible();
    expect(callsTo(fetchMock, `POST /contribution-needs/${needId}/applications`)).toHaveLength(0);
  });

  it('sends the message with CSRF, JSON and an Idempotency-Key, then lets the member withdraw', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({
      ...routes,
      [`POST /contribution-needs/${needId}/applications`]: () =>
        jsonResponse(application(), { status: 201 }),
      [`PATCH /applications/${applicationId}`]: () =>
        jsonResponse(application({ status: 'withdrawn' })),
    });
    const user = await apply();
    await user.type(
      screen.getByRole('textbox', { name: /Votre message au projet/ }),
      'Je connais ce domaine.',
    );
    await user.click(screen.getByRole('button', { name: 'Envoyer ma candidature' }));
    expect(await screen.findByText(/Candidature envoyée/)).toBeVisible();
    const [request] = callsTo(fetchMock, `POST /contribution-needs/${needId}/applications`);
    expect(request?.headers.get('X-CSRF-Token')).toBeTruthy();
    expect(request?.headers.get('Idempotency-Key')).toBeTruthy();
    expect(await request?.clone().json()).toEqual({ message: 'Je connais ce domaine.' });

    await user.click(screen.getByRole('button', { name: 'Retirer ma candidature' }));
    expect(await screen.findByText('Candidature retirée.')).toBeVisible();
    const [patch] = callsTo(fetchMock, `PATCH /applications/${applicationId}`);
    expect(patch?.headers.get('X-CSRF-Token')).toBeTruthy();
    expect(await patch?.clone().json()).toEqual({ status: 'withdrawn' });
  });

  it.each([
    ['DUPLICATE_APPLICATION', 'Vous avez déjà postulé à ce besoin.'],
    ['STATE_CONFLICT', 'Ce besoin n’accepte plus de candidatures'],
    ['EMAIL_NOT_VERIFIED', 'Confirmez d’abord votre adresse e-mail.'],
  ] as const)('explains a refused application (%s)', async (code, text) => {
    setLanguages(['fr-FR']);
    member({
      ...routes,
      [`POST /contribution-needs/${needId}/applications`]: () =>
        problemResponse(code === 'EMAIL_NOT_VERIFIED' ? 403 : 409, code),
    });
    const user = await apply();
    await user.type(screen.getByRole('textbox', { name: /Votre message/ }), 'Bonjour');
    await user.click(screen.getByRole('button', { name: 'Envoyer ma candidature' }));
    await expectAlert(text);
  });

  it('reuses the same Idempotency-Key when the same application is retried after a failure', async () => {
    setLanguages(['fr-FR']);
    let calls = 0;
    const fetchMock = member({
      ...routes,
      [`POST /contribution-needs/${needId}/applications`]: () =>
        ++calls === 1
          ? problemResponse(503, 'SERVICE_UNAVAILABLE')
          : jsonResponse(application(), { status: 201 }),
    });
    const user = await apply();
    await user.type(screen.getByRole('textbox', { name: /Votre message/ }), 'Bonjour');
    await user.click(screen.getByRole('button', { name: 'Envoyer ma candidature' }));
    await expectAlert('momentanément indisponible');
    await user.click(screen.getByRole('button', { name: 'Envoyer ma candidature' }));
    expect(await screen.findByText(/Candidature envoyée/)).toBeVisible();
    const [first, again] = callsTo(fetchMock, `POST /contribution-needs/${needId}/applications`);
    expect(first?.headers.get('Idempotency-Key')).toBeTruthy();
    expect(again?.headers.get('Idempotency-Key')).toBe(first?.headers.get('Idempotency-Key'));
  });

  it('withdraws through the id in Location when a replay omits the body', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({
      ...routes,
      [`POST /contribution-needs/${needId}/applications`]: () =>
        new Response(null, {
          status: 201,
          headers: {
            'Idempotency-Body-Omitted': 'true',
            Location: `/api/v1/applications/${applicationId}`,
          },
        }),
      [`PATCH /applications/${applicationId}`]: () =>
        jsonResponse(application({ status: 'withdrawn' })),
    });
    const user = await apply();
    await user.type(screen.getByRole('textbox', { name: /Votre message/ }), 'Bonjour');
    await user.click(screen.getByRole('button', { name: 'Envoyer ma candidature' }));
    await user.click(await screen.findByRole('button', { name: 'Retirer ma candidature' }));
    expect(await screen.findByText('Candidature retirée.')).toBeVisible();
    expect(callsTo(fetchMock, `PATCH /applications/${applicationId}`)).toHaveLength(1);
  });

  it('says the application was already handled when the owner decided before the withdrawal', async () => {
    setLanguages(['fr-FR']);
    member({
      ...routes,
      [`POST /contribution-needs/${needId}/applications`]: () =>
        jsonResponse(application(), { status: 201 }),
      [`PATCH /applications/${applicationId}`]: () => problemResponse(409, 'STATE_CONFLICT'),
    });
    const user = await apply();
    await user.type(screen.getByRole('textbox', { name: /Votre message/ }), 'Bonjour');
    await user.click(screen.getByRole('button', { name: 'Envoyer ma candidature' }));
    await user.click(await screen.findByRole('button', { name: 'Retirer ma candidature' }));
    await expectAlert('Cette candidature a déjà été traitée.');
  });

  it('keeps the sent state when the project page is left and opened again', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({
      ...routes,
      [`POST /contribution-needs/${needId}/applications`]: () =>
        jsonResponse(application(), { status: 201 }),
    });
    const user = userEvent.setup();
    const { router } = await renderApp(`/build/${projectId}`);
    await user.click(await screen.findByRole('button', { name: 'Postuler' }));
    await user.type(screen.getByRole('textbox', { name: /Votre message/ }), 'Bonjour');
    await user.click(screen.getByRole('button', { name: 'Envoyer ma candidature' }));
    await screen.findByText(/Candidature envoyée/);

    await router.navigate({ to: '/' });
    await router.navigate({ to: '/build/$projectId', params: { projectId } });
    expect(await screen.findByText(/Candidature envoyée/)).toBeVisible();
    expect(screen.queryByRole('button', { name: 'Postuler' })).not.toBeInTheDocument();
    expect(callsTo(fetchMock, `POST /contribution-needs/${needId}/applications`)).toHaveLength(1);
  });

  it('shows the pending application after a full reload and withdraws it by its id', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({
      ...routes,
      [`GET ${base}/contribution-needs`]: () =>
        page([need({ viewer_application: { id: applicationId, status: 'pending' } })]),
      [`PATCH /applications/${applicationId}`]: () =>
        jsonResponse(application({ status: 'withdrawn' })),
    });
    const user = userEvent.setup();
    await renderApp(`/build/${projectId}`);
    expect(await screen.findByText(/Candidature envoyée/)).toBeVisible();
    expect(screen.queryByRole('button', { name: 'Postuler' })).not.toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: 'Retirer ma candidature' }));
    expect(await screen.findByText('Candidature retirée.')).toBeVisible();
    expect(callsTo(fetchMock, `PATCH /applications/${applicationId}`)).toHaveLength(1);
    expect(callsTo(fetchMock, `POST /contribution-needs/${needId}/applications`)).toHaveLength(0);
  });

  it.each([
    ['accepted', 'Votre candidature a été acceptée.'],
    ['declined', 'Votre candidature n’a pas été retenue.'],
    ['withdrawn', 'Candidature retirée.'],
  ] as const)(
    'shows a %s application after a reload, without apply or withdraw',
    async (status, text) => {
      setLanguages(['fr-FR']);
      member({
        ...routes,
        [`GET ${base}/contribution-needs`]: () =>
          page([need({ viewer_application: { id: applicationId, status } })]),
      });
      await renderApp(`/build/${projectId}`);
      expect(await screen.findByText(text)).toBeVisible();
      expect(screen.queryByRole('button', { name: 'Postuler' })).not.toBeInTheDocument();
      expect(
        screen.queryByRole('button', { name: 'Retirer ma candidature' }),
      ).not.toBeInTheDocument();
    },
  );

  it.each([
    ['no Location', undefined],
    ['a cross-origin Location', `https://evil.test/api/v1/applications/${applicationId}`],
    ['a protocol-relative Location', `//evil.test/api/v1/applications/${applicationId}`],
    ['a Location that is not an application', `/api/v1/projects/${applicationId}`],
  ] as const)(
    'shows the sent state without withdraw when a replay has %s',
    async (_name, location) => {
      setLanguages(['fr-FR']);
      const fetchMock = member({
        ...routes,
        [`POST /contribution-needs/${needId}/applications`]: () =>
          new Response(null, {
            status: 201,
            headers: {
              'Idempotency-Body-Omitted': 'true',
              ...(location ? { Location: location } : {}),
            },
          }),
      });
      const user = await apply();
      await user.type(screen.getByRole('textbox', { name: /Votre message/ }), 'Bonjour');
      await user.click(screen.getByRole('button', { name: 'Envoyer ma candidature' }));
      expect(await screen.findByText(/Candidature envoyée/)).toBeVisible();
      expect(
        screen.queryByRole('button', { name: 'Retirer ma candidature' }),
      ).not.toBeInTheDocument();
      expect(
        fetchMock.mock.calls.some(([input]) =>
          String(input instanceof Request ? input.url : input).includes('/applications/undefined'),
        ),
      ).toBe(false);
    },
  );

  it('shows a server violation on the message field', async () => {
    setLanguages(['fr-FR']);
    member({
      ...routes,
      [`POST /contribution-needs/${needId}/applications`]: () =>
        problemResponse(422, 'VALIDATION_ERROR', {
          violations: [{ field: '/message', code: 'too_long' }],
        }),
    });
    const user = await apply();
    await user.type(screen.getByRole('textbox', { name: /Votre message/ }), 'Bonjour');
    await user.click(screen.getByRole('button', { name: 'Envoyer ma candidature' }));
    expect(await screen.findByText(/Écrivez un message de 1 à 2/)).toBeVisible();
  });
});

describe('ProjectPage owner writes', () => {
  const routes = {
    [`GET ${base}`]: () => jsonResponse(project()),
    [`GET ${base}/contribution-needs`]: () => page([need()]),
    [`GET ${base}/applications`]: () => page([application()]),
  };

  it('edits the project with a merge patch of the changed fields only, and no link removal', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = owner({
      ...routes,
      [`PATCH ${base}`]: () => jsonResponse(project({ description: 'Nouvelle description' })),
    });
    const user = userEvent.setup();
    await renderApp(`/build/${projectId}`);
    await user.click(await screen.findByRole('button', { name: 'Modifier le projet' }));
    const field = screen.getByRole('textbox', { name: /Description/ });
    await user.clear(field);
    await user.type(field, 'Nouvelle description');
    expect(screen.queryByRole('textbox', { name: /Adresse GitHub/ })).not.toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Enregistrer' }));
    expect(await screen.findByText('Nouvelle description')).toBeVisible();
    const [request] = callsTo(fetchMock, `PATCH ${base}`);
    expect(request?.headers.get('X-CSRF-Token')).toBeTruthy();
    expect(await request?.clone().json()).toEqual({ description: 'Nouvelle description' });
  });

  it('saves nothing and calls nothing when the edit changes nothing', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = owner(routes);
    const user = userEvent.setup();
    await renderApp(`/build/${projectId}`);
    await user.click(await screen.findByRole('button', { name: 'Modifier le projet' }));
    await user.click(screen.getByRole('button', { name: 'Enregistrer' }));
    expect(await screen.findByRole('button', { name: 'Modifier le projet' })).toBeVisible();
    expect(callsTo(fetchMock, `PATCH ${base}`)).toHaveLength(0);
  });

  it('sends exactly the changed technologies and visibility, never links or github_url', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = owner({
      ...routes,
      [`PATCH ${base}`]: () =>
        jsonResponse(project({ technologies: ['go', 'rust'], visibility: 'private' })),
    });
    const user = userEvent.setup();
    await renderApp(`/build/${projectId}`);
    await user.click(await screen.findByRole('button', { name: 'Modifier le projet' }));
    const tech = screen.getByRole('textbox', { name: /Technologies/ });
    await user.clear(tech);
    await user.type(tech, 'Go, Rust');
    await selectSegment(screen.getByRole('radio', { name: 'Privé, visible de vous seul' }));
    await user.click(screen.getByRole('button', { name: 'Enregistrer' }));
    await waitFor(() => {
      expect(callsTo(fetchMock, `PATCH ${base}`)).toHaveLength(1);
    });
    const body: unknown = await callsTo(fetchMock, `PATCH ${base}`)[0]?.clone().json();
    expect(body).toEqual({ technologies: ['go', 'rust'], visibility: 'private' });
    expect(body).not.toHaveProperty('links');
    expect(body).not.toHaveProperty('github_url');
  });

  it('shows a calm error when a non-owner edit is refused', async () => {
    setLanguages(['fr-FR']);
    owner({ ...routes, [`PATCH ${base}`]: () => problemResponse(403, 'FORBIDDEN') });
    const user = userEvent.setup();
    await renderApp(`/build/${projectId}`);
    await user.click(await screen.findByRole('button', { name: 'Modifier le projet' }));
    const field = screen.getByRole('textbox', { name: /Description/ });
    await user.type(field, ' plus');
    await user.click(screen.getByRole('button', { name: 'Enregistrer' }));
    await expectAlert('Vous n’avez pas accès à cette action.');
  });

  it('adds a need with an Idempotency-Key and shows violations on fields', async () => {
    setLanguages(['fr-FR']);
    let refuse = true;
    const fetchMock = owner({
      ...routes,
      [`POST ${base}/contribution-needs`]: () =>
        refuse
          ? problemResponse(422, 'VALIDATION_ERROR', {
              violations: [{ field: '/skills', code: 'invalid_format' }],
            })
          : jsonResponse(need(), { status: 201 }),
    });
    const user = userEvent.setup();
    await renderApp(`/build/${projectId}`);
    await user.click(await screen.findByRole('button', { name: 'Ajouter un besoin' }));
    await user.type(screen.getByRole('textbox', { name: 'Titre du besoin' }), 'Traduction');
    await user.type(screen.getByRole('textbox', { name: /^Description/ }), 'Traduire la doc.');
    await user.type(screen.getByRole('textbox', { name: /Compétences/ }), 'go');
    await user.click(screen.getByRole('button', { name: 'Publier le besoin' }));
    expect(await screen.findByText(/Indiquez de 1 à 10 compétences/)).toBeVisible();
    refuse = false;
    await user.click(screen.getByRole('button', { name: 'Publier le besoin' }));
    await waitFor(() => {
      expect(callsTo(fetchMock, `POST ${base}/contribution-needs`)).toHaveLength(2);
    });
    const [request, retry] = callsTo(fetchMock, `POST ${base}/contribution-needs`);
    expect(request?.headers.get('X-CSRF-Token')).toBeTruthy();
    expect(request?.headers.get('Idempotency-Key')).toBeTruthy();
    // The retry after the 422 is the same submission: the key is reused, never regenerated.
    expect(retry?.headers.get('Idempotency-Key')).toBe(request?.headers.get('Idempotency-Key'));
    expect(await request?.clone().json()).toMatchObject({ title: 'Traduction', skills: ['go'] });
  });

  it('lists applications and accepts or declines with icon buttons', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = owner({
      ...routes,
      [`PATCH /applications/${applicationId}`]: () =>
        jsonResponse(application({ status: 'accepted' })),
    });
    const user = userEvent.setup();
    await renderApp(`/build/${projectId}`);
    const section = await screen.findByRole('region', { name: 'Candidatures' });
    expect(await within(section).findByText('Je connais bien ce domaine.')).toBeVisible();
    expect(
      within(section).getByRole('button', { name: 'Refuser la candidature de Awa Traoré' }),
    ).toBeVisible();
    await user.click(
      within(section).getByRole('button', { name: 'Accepter la candidature de Awa Traoré' }),
    );
    await waitFor(() => {
      expect(callsTo(fetchMock, `PATCH /applications/${applicationId}`)).toHaveLength(1);
    });
    const [request] = callsTo(fetchMock, `PATCH /applications/${applicationId}`);
    expect(request?.headers.get('X-CSRF-Token')).toBeTruthy();
    expect(await request?.clone().json()).toEqual({ status: 'accepted' });
  });

  it('says when an application was already handled', async () => {
    setLanguages(['fr-FR']);
    owner({
      ...routes,
      [`PATCH /applications/${applicationId}`]: () => problemResponse(409, 'STATE_CONFLICT'),
    });
    const user = userEvent.setup();
    await renderApp(`/build/${projectId}`);
    await user.click(
      await screen.findByRole('button', { name: 'Refuser la candidature de Awa Traoré' }),
    );
    await expectAlert('Cette candidature a déjà été traitée.');
  });

  it('shows no decision buttons on a decided application', async () => {
    setLanguages(['fr-FR']);
    owner({
      ...routes,
      [`GET ${base}/applications`]: () => page([application({ status: 'declined' })]),
    });
    await renderApp(`/build/${projectId}`);
    const section = await screen.findByRole('region', { name: 'Candidatures' });
    expect(await within(section).findByText('Refusée')).toBeVisible();
    expect(within(section).queryByRole('button')).not.toBeInTheDocument();
  });
});
