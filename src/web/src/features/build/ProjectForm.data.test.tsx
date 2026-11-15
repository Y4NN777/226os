import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { callsTo, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';
import { member, page, project } from './testFixtures';

const created = project({ id: 'e0000000-0000-4000-8000-0000000000bb' });
const url = 'https://github.com/kora-dev/kora-sdk';

async function openForm() {
  const user = userEvent.setup({ delay: null });
  await renderApp('/build');
  await user.click(await screen.findByRole('button', { name: 'Proposer un projet' }));
  // Paste, not type: typing a long URL one key at a time is slow under coverage in CI.
  await user.click(screen.getByRole('textbox', { name: /Dépôt GitHub/ }));
  await user.paste(url);
  return user;
}

/** The create form, apart from the page filters that share the « Étape » and « Pays » labels. */
function createForm() {
  const form = screen.getByRole('textbox', { name: /Dépôt GitHub/ }).closest('form');
  if (!form) throw new Error('create form not found');
  return within(form);
}

async function pick(user: ReturnType<typeof userEvent.setup>, label: RegExp, option: string) {
  await user.click(
    createForm().getByRole(label.source.includes('Pays') ? 'button' : 'combobox', { name: label }),
  );
  await user.click(await screen.findByRole('option', { name: option }));
}

function routes(post: () => Response) {
  return {
    'GET /projects': () => page([]),
    'POST /projects': post,
    [`GET /projects/${created.id}`]: () => jsonResponse(created),
    [`GET /projects/${created.id}/contribution-needs`]: () => page([]),
  };
}

describe('BuildPage create form: stage, licence and country', () => {
  it('sends the stage, the licence and a country from outside West Africa', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member(routes(() => jsonResponse(created, { status: 201 })));
    const user = await openForm();
    await pick(user, /^Étape/, 'Prototype');
    await user.click(screen.getByRole('textbox', { name: /Licence/ }));
    await user.paste('Apache-2.0');
    await pick(user, /^Pays/, 'Kenya');
    await user.click(screen.getByRole('button', { name: 'Publier le projet' }));
    await waitFor(() => {
      expect(callsTo(fetchMock, 'POST /projects')).toHaveLength(1);
    });
    expect(await callsTo(fetchMock, 'POST /projects')[0]?.clone().json()).toMatchObject({
      github_url: url,
      stage: 'prototype',
      license_spdx: 'Apache-2.0',
      country_code: 'KE',
    });
    // Two menus, one with the 54 countries: about 8 s locally, more under coverage in CI.
  }, 30_000);

  it('sends none of the three when they are left alone', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member(routes(() => jsonResponse(created, { status: 201 })));
    const user = await openForm();
    await user.click(screen.getByRole('button', { name: 'Publier le projet' }));
    await waitFor(() => {
      expect(callsTo(fetchMock, 'POST /projects')).toHaveLength(1);
    });
    const body = (await callsTo(fetchMock, 'POST /projects')[0]?.clone().json()) as object;
    expect(body).not.toHaveProperty('stage');
    expect(body).not.toHaveProperty('license_spdx');
    expect(body).not.toHaveProperty('country_code');
  });

  it('refuses a malformed licence without calling the API', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = member(routes(() => jsonResponse(created, { status: 201 })));
    const user = await openForm();
    await user.click(screen.getByRole('textbox', { name: /Licence/ }));
    await user.paste('not valid!');
    await user.click(screen.getByRole('button', { name: 'Publier le projet' }));
    expect(await screen.findByText(/Identifiant SPDX non valide/)).toBeVisible();
    expect(callsTo(fetchMock, 'POST /projects')).toHaveLength(0);
  });

  it('shows 422 on /stage, /license_spdx and /country_code on their fields', async () => {
    setLanguages(['fr-FR']);
    member(
      routes(() =>
        problemResponse(422, 'VALIDATION_ERROR', {
          violations: [
            { field: '/stage', code: 'invalid_value' },
            { field: '/license_spdx', code: 'invalid_value' },
            { field: '/country_code', code: 'invalid_value' },
          ],
        }),
      ),
    );
    const user = await openForm();
    await user.click(screen.getByRole('button', { name: 'Publier le projet' }));
    expect(await screen.findByText('Cette étape demande un dépôt GitHub.')).toBeVisible();
    expect(createForm().getByText(/Identifiant SPDX non valide/)).toBeVisible();
    expect(createForm().getByText('Pays inconnu.')).toBeVisible();
  });
});

describe('BuildPage country filter', () => {
  it('offers the whole continent, not only West Africa', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /projects': () => page([project()]) });
    const user = userEvent.setup({ delay: null });
    await renderApp('/build');
    await user.click(await screen.findByRole('button', { name: /^Pays/ }));
    expect(await screen.findByRole('option', { name: 'Kenya' })).toBeVisible();
    expect(screen.getByRole('option', { name: 'Afrique du Sud' })).toBeVisible();
  });
});
