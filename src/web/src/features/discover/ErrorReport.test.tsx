import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { callsTo, memberSession, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';
import { resource, source } from './testFixtures';

const id = 'c0000000-0000-4000-8000-000000000001';
const sourceId = '22222222-2222-4222-8222-000000000001';

function routes(overrides: Record<string, () => Response> = {}) {
  return mockApi({
    [`GET /resources/${id}`]: () => jsonResponse(resource({ id, name: 'Wave Pay' })),
    [`GET /resources/${id}/sources`]: () =>
      jsonResponse({ items: [source({ id: sourceId, title: 'Documentation Wave' })] }),
    ...overrides,
  });
}

const member = { 'GET /auth/session': () => jsonResponse(memberSession) };
const created = () =>
  jsonResponse(
    {
      id: 'd0000000-0000-4000-8000-000000000001',
      resource_id: id,
      kind: 'broken_link',
      status: 'received',
      created_at: '2026-11-14T10:00:00Z',
    },
    { status: 201 },
  );

async function open() {
  setLanguages(['fr-FR']);
  await renderApp(`/discover/${id}`);
  await userEvent.click(await screen.findByRole('button', { name: 'Signaler une erreur' }));
}

async function chooseKind(name: string) {
  await userEvent.click(screen.getByRole('combobox', { name: /Type d’erreur/ }));
  await userEvent.click(await screen.findByRole('option', { name }));
}

async function describeError(text: string) {
  await userEvent.type(screen.getByRole('textbox', { name: /Que faut-il corriger/ }), text);
}

const submit = () =>
  userEvent.click(screen.getByRole('button', { name: 'Envoyer le signalement' }));

describe('Error report on a resource', () => {
  it('asks an anonymous visitor to sign in, without showing the form', async () => {
    routes();
    await open();
    expect(
      screen.getByText('Connectez-vous pour signaler une erreur sur cette fiche.'),
    ).toBeVisible();
    const hrefs = screen
      .getAllByRole('link', { name: 'Se connecter' })
      .map((link) => link.getAttribute('href'));
    expect(hrefs).toContain(`/connexion?next=${encodeURIComponent(`/discover/${id}`)}`);
    expect(screen.queryByRole('combobox', { name: /Type d’erreur/ })).not.toBeInTheDocument();
  });

  it('sends the kind, the message and the source with the CSRF header, then thanks the member', async () => {
    const fetchMock = routes({ ...member, [`POST /resources/${id}/error-reports`]: created });
    await open();
    await chooseKind('Lien cassé');
    await describeError('Le lien de la documentation renvoie une 404.');
    await userEvent.click(screen.getByRole('combobox', { name: /Source concernée/ }));
    await userEvent.click(await screen.findByRole('option', { name: 'Documentation Wave' }));
    await submit();

    expect(
      await screen.findByText('Merci, votre signalement a été envoyé. L’équipe va le vérifier.'),
    ).toBeVisible();
    const [request] = callsTo(fetchMock, `POST /resources/${id}/error-reports`);
    expect(request?.headers.get('X-CSRF-Token')).toBe('csrf-member');
    expect(await request?.json()).toEqual({
      kind: 'broken_link',
      body: 'Le lien de la documentation renvoie une 404.',
      source_id: sourceId,
    });
    expect(screen.queryByRole('textbox', { name: /Que faut-il corriger/ })).not.toBeInTheDocument();
  });

  it('omits the source when none is chosen', async () => {
    const fetchMock = routes({ ...member, [`POST /resources/${id}/error-reports`]: created });
    await open();
    await chooseKind('Autre');
    await describeError('  Tarif périmé  ');
    await submit();
    await screen.findByText(/Merci, votre signalement/);
    const [request] = callsTo(fetchMock, `POST /resources/${id}/error-reports`);
    expect(await request?.json()).toEqual({ kind: 'other', body: 'Tarif périmé' });
  });

  it('blocks an empty report and names what is missing, without calling the API', async () => {
    const fetchMock = routes(member);
    await open();
    await submit();
    expect(await screen.findByText('Choisissez un type d’erreur.')).toBeVisible();
    expect(screen.getByText('Décrivez l’erreur en 2 000 caractères au plus.')).toBeVisible();
    expect(callsTo(fetchMock, `POST /resources/${id}/error-reports`)).toHaveLength(0);
  });

  it('maps a 422 violation on the message to the field', async () => {
    routes({
      ...member,
      [`POST /resources/${id}/error-reports`]: () =>
        problemResponse(422, 'VALIDATION_ERROR', {
          violations: [{ field: '/body', code: 'too_long' }],
        }),
    });
    await open();
    await chooseKind('Information périmée');
    await describeError('x');
    await submit();
    expect(await screen.findByText('Décrivez l’erreur en 2 000 caractères au plus.')).toBeVisible();
  });

  it('says the listing is gone on a 404 and keeps the form', async () => {
    routes({
      ...member,
      [`POST /resources/${id}/error-reports`]: () => problemResponse(404, 'NOT_FOUND'),
    });
    await open();
    await chooseKind('Information incorrecte');
    await describeError('Faux');
    await submit();
    expect(
      await screen.findByText('Ce contenu n’existe pas ou n’est plus disponible.'),
    ).toBeVisible();
    expect(screen.getByRole('textbox', { name: /Que faut-il corriger/ })).toHaveValue('Faux');
  });

  it('is available in English', async () => {
    routes();
    setLanguages(['en-US']);
    await renderApp(`/discover/${id}`);
    await userEvent.click(await screen.findByRole('button', { name: 'Report an error' }));
    expect(screen.getByText('Sign in to report an error on this listing.')).toBeVisible();
  });
});
