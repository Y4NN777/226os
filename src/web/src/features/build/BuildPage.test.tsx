import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { callsTo, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';
import { member, page, project, projectId } from './testFixtures';

const second = project({
  id: 'e0000000-0000-4000-8000-000000000002',
  name: 'sahel-docs',
  description: '',
  technologies: ['markdown'],
  open_needs_count: 0,
});

async function expectAlert(text: string) {
  await waitFor(() => {
    const alerts = screen.getAllByRole('alert').map((alert) => alert.textContent);
    expect(alerts.join(' ')).toContain(text);
  });
}

describe('BuildPage list', () => {
  it('lists projects as links with technologies, open needs and owner, in French', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /projects': () => page([project(), second]) });
    await renderApp('/build');

    expect(
      await screen.findByRole('heading', { level: 1, name: 'Construire' }),
    ).toBeInTheDocument();
    const link = await screen.findByRole('link', { name: 'kora-sdk' });
    expect(link).toHaveAttribute('href', `/build/${projectId}`);
    const results = screen.getByRole('region', { name: 'Projets' });
    expect(within(results).getByText('mobile-money')).toBeInTheDocument();
    expect(within(results).getByText('1 besoin ouvert')).toBeInTheDocument();
    expect(within(results).getByText('Aucun besoin ouvert')).toBeInTheDocument();
    expect(within(results).getAllByRole('link', { name: 'Kofi Mensah' })[0]).toHaveAttribute(
      'href',
      '/profils/kofi',
    );
  });

  it('loads the next page with the opaque cursor', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = mockApi({
      'GET /projects': (request) =>
        new URL(request.url).searchParams.get('cursor') === 'cur-2'
          ? page([second])
          : page([project()], 'cur-2'),
    });
    const user = userEvent.setup();
    await renderApp('/build');
    await screen.findByRole('link', { name: 'kora-sdk' });
    await user.click(screen.getByRole('button', { name: 'Afficher plus' }));
    expect(await screen.findByRole('link', { name: 'sahel-docs' })).toBeVisible();
    expect(callsTo(fetchMock, 'GET /projects')).toHaveLength(2);
    expect(screen.queryByRole('button', { name: 'Afficher plus' })).not.toBeInTheDocument();
  });

  it('says so when there is no project, and offers a retry when the load fails', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /projects': () => page([]) });
    await renderApp('/build');
    expect(await screen.findByText('Aucun projet pour le moment')).toBeInTheDocument();
  });

  it('shows a retryable error', async () => {
    setLanguages(['fr-FR']);
    let calls = 0;
    mockApi({
      'GET /projects': () =>
        ++calls === 1 ? problemResponse(500, 'INTERNAL_ERROR') : page([project()]),
    });
    const user = userEvent.setup();
    await renderApp('/build');
    await expectAlert('Une erreur est survenue');
    await user.click(screen.getByRole('button', { name: 'Réessayer' }));
    expect(await screen.findByRole('link', { name: 'kora-sdk' })).toBeVisible();
  });

  it('turns Construire into a live navigation link', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /projects': () => page([]) });
    await renderApp('/build');
    const nav = await screen.findByRole('navigation', { name: 'Navigation principale' });
    expect(within(nav).getByRole('link', { name: 'Construire' })).toHaveAttribute('href', '/build');
  });

  it('invites an anonymous visitor to sign in, with a safe next', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /projects': () => page([]) });
    await renderApp('/build');
    expect(
      await screen.findByRole('link', { name: 'Se connecter pour proposer un projet' }),
    ).toHaveAttribute('href', '/connexion?next=%2Fbuild');
    expect(screen.queryByRole('button', { name: 'Proposer un projet' })).not.toBeInTheDocument();
  });
});

const created = project({ id: 'e0000000-0000-4000-8000-0000000000bb' });

async function openForm() {
  const user = userEvent.setup();
  await renderApp('/build');
  await user.click(await screen.findByRole('button', { name: 'Proposer un projet' }));
  return user;
}

describe('BuildPage create form', () => {
  it('sends the project with CSRF, JSON and an Idempotency-Key, then opens its page', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({
      'GET /projects': () => page([]),
      'POST /projects': () => jsonResponse(created, { status: 201 }),
      [`GET /projects/${created.id}`]: () => jsonResponse(created),
      [`GET /projects/${created.id}/contribution-needs`]: () => page([]),
    });
    const user = await openForm();
    await user.type(
      screen.getByRole('textbox', { name: /Adresse GitHub/ }),
      'https://github.com/kora-dev/kora-sdk',
    );
    await user.type(screen.getByRole('textbox', { name: /Technologies/ }), 'Go, mobile-money');
    await user.click(screen.getByRole('button', { name: 'Publier le projet' }));

    expect(await screen.findByRole('heading', { level: 1, name: 'kora-sdk' })).toBeVisible();
    const [request] = callsTo(fetchMock, 'POST /projects');
    expect(request?.headers.get('X-CSRF-Token')).toBeTruthy();
    expect(request?.headers.get('Content-Type')).toContain('application/json');
    expect(request?.headers.get('Idempotency-Key')).toBeTruthy();
    expect(await request?.clone().json()).toMatchObject({
      github_url: 'https://github.com/kora-dev/kora-sdk',
      technologies: ['go', 'mobile-money'],
      lang: 'fr',
    });
  });

  it('refuses an address that is not a GitHub repository without calling the API', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({ 'GET /projects': () => page([]) });
    const user = await openForm();
    await user.type(
      screen.getByRole('textbox', { name: /Adresse GitHub/ }),
      'https://example.com/x',
    );
    await user.click(screen.getByRole('button', { name: 'Publier le projet' }));
    expect(await screen.findByText(/Indiquez l’adresse d’un dépôt GitHub/)).toBeVisible();
    expect(callsTo(fetchMock, 'POST /projects')).toHaveLength(0);
  });

  it('shows server violations on the field', async () => {
    setLanguages(['fr-FR']);
    member({
      'GET /projects': () => page([]),
      'POST /projects': () =>
        problemResponse(422, 'VALIDATION_ERROR', {
          violations: [{ field: '/description', code: 'too_long' }],
        }),
    });
    const user = await openForm();
    await user.type(
      screen.getByRole('textbox', { name: /Adresse GitHub/ }),
      'https://github.com/kora-dev/kora-sdk',
    );
    await user.click(screen.getByRole('button', { name: 'Publier le projet' }));
    expect(await screen.findByText(/La description ne peut pas dépasser/)).toBeVisible();
  });

  it('says on the URL field when the repository is unknown or already listed', async () => {
    setLanguages(['fr-FR']);
    let status = 422;
    member({
      'GET /projects': () => page([]),
      'POST /projects': () =>
        status === 422
          ? problemResponse(422, 'INVALID_GITHUB_URL')
          : problemResponse(409, 'STATE_CONFLICT'),
    });
    const user = await openForm();
    await user.type(
      screen.getByRole('textbox', { name: /Adresse GitHub/ }),
      'https://github.com/kora-dev/kora-sdk',
    );
    await user.click(screen.getByRole('button', { name: 'Publier le projet' }));
    expect(await screen.findByText(/Dépôt public introuvable/)).toBeVisible();
    status = 409;
    await user.click(screen.getByRole('button', { name: 'Publier le projet' }));
    expect(await screen.findByText(/déjà présenté/)).toBeVisible();
    expect(screen.getByRole('textbox', { name: /Adresse GitHub/ })).toBeInvalid();
  });

  it('asks to retry later when GitHub is unavailable, reusing the same key', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member({
      'GET /projects': () => page([]),
      'POST /projects': () => problemResponse(503, 'UPSTREAM_UNAVAILABLE', { retryAfter: 30 }),
    });
    const user = await openForm();
    await user.type(
      screen.getByRole('textbox', { name: /Adresse GitHub/ }),
      'https://github.com/kora-dev/kora-sdk',
    );
    await user.click(screen.getByRole('button', { name: 'Publier le projet' }));
    await expectAlert('GitHub ne répond pas');
    await user.click(screen.getByRole('button', { name: 'Publier le projet' }));
    await waitFor(() => {
      expect(callsTo(fetchMock, 'POST /projects')).toHaveLength(2);
    });
    const [first, again] = callsTo(fetchMock, 'POST /projects');
    expect(again?.headers.get('Idempotency-Key')).toBe(first?.headers.get('Idempotency-Key'));
  });

  it('explains an unverified email', async () => {
    setLanguages(['fr-FR']);
    member({
      'GET /projects': () => page([]),
      'POST /projects': () => problemResponse(403, 'EMAIL_NOT_VERIFIED'),
    });
    const user = await openForm();
    await user.type(
      screen.getByRole('textbox', { name: /Adresse GitHub/ }),
      'https://github.com/kora-dev/kora-sdk',
    );
    await user.click(screen.getByRole('button', { name: 'Publier le projet' }));
    await expectAlert('Confirmez d’abord votre adresse e-mail.');
  });

  it('follows Location when the replay omits the body', async () => {
    setLanguages(['fr-FR']);
    member({
      'GET /projects': () => page([]),
      'POST /projects': () =>
        new Response(null, {
          status: 201,
          headers: {
            'Idempotency-Body-Omitted': 'true',
            Location: `/api/v1/projects/${created.id}`,
          },
        }),
      [`GET /projects/${created.id}`]: () => jsonResponse(created),
      [`GET /projects/${created.id}/contribution-needs`]: () => page([]),
    });
    const user = await openForm();
    await user.type(
      screen.getByRole('textbox', { name: /Adresse GitHub/ }),
      'https://github.com/kora-dev/kora-sdk',
    );
    await user.click(screen.getByRole('button', { name: 'Publier le projet' }));
    expect(await screen.findByRole('heading', { level: 1, name: 'kora-sdk' })).toBeVisible();
  });
});
