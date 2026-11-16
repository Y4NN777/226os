import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { callsTo, mockApi, problemResponse } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';
import { owner, page, project, projectId } from './testFixtures';

const base = `/projects/${projectId}`;

function routesFor(overrides: Parameters<typeof project>[0] = {}) {
  return {
    [`GET ${base}`]: () => jsonResponse(project(overrides)),
    [`GET ${base}/contribution-needs`]: () => page([]),
    [`GET ${base}/applications`]: () => page([]),
  };
}

async function edit() {
  const user = userEvent.setup({ delay: null });
  await renderApp(`/build/${projectId}`);
  await user.click(await screen.findByRole('button', { name: 'Modifier le projet' }));
  return user;
}

async function pick(user: ReturnType<typeof userEvent.setup>, label: RegExp, option: string) {
  await user.click(
    screen.getByRole(label.source.includes('Pays') ? 'button' : 'combobox', { name: label }),
  );
  await user.click(await screen.findByRole('option', { name: option }));
}

describe('ProjectPage stage, country and licence', () => {
  it('shows the stage and the country in the header meta', async () => {
    setLanguages(['fr-FR']);
    mockApi({
      ...routesFor({ stage: 'prototype', country_code: 'SN' }),
    });
    await renderApp(`/build/${projectId}`);
    await screen.findByRole('heading', { level: 1, name: 'kora-sdk' });
    const scope = within(screen.getByRole('main'));
    expect(scope.getByText('Prototype')).toBeVisible();
    expect(scope.getByText('SN')).toBeInTheDocument();
    expect(scope.getByText('Sénégal')).toBeInTheDocument();
  });

  it('leaves out the country when none is declared', async () => {
    setLanguages(['fr-FR']);
    mockApi(routesFor());
    await renderApp(`/build/${projectId}`);
    await screen.findByRole('heading', { level: 1, name: 'kora-sdk' });
    expect(screen.getByText('Idée')).toBeVisible();
    expect(screen.queryByText('Burkina Faso')).not.toBeInTheDocument();
  });

  it('lists the licence in the side facts, next to the repository link', async () => {
    setLanguages(['fr-FR']);
    mockApi(routesFor({ license_spdx: 'Apache-2.0' }));
    await renderApp(`/build/${projectId}`);
    const about = await screen.findByRole('region', { name: 'À propos' });
    expect(within(about).getByText('Licence')).toBeVisible();
    expect(within(about).getByText('Apache-2.0')).toBeVisible();
    expect(within(about).getByRole('link', { name: /Voir sur GitHub/ })).toHaveAttribute(
      'href',
      'https://github.com/kora-dev/kora-sdk',
    );
  });

  it('shows no licence line when none is declared', async () => {
    setLanguages(['fr-FR']);
    mockApi(routesFor());
    await renderApp(`/build/${projectId}`);
    const about = await screen.findByRole('region', { name: 'À propos' });
    expect(within(about).queryByText('Licence')).not.toBeInTheDocument();
  });

  it('writes the licence label in English', async () => {
    setLanguages(['en-US']);
    mockApi(routesFor({ license_spdx: 'MIT' }));
    await renderApp(`/build/${projectId}`);
    const about = await screen.findByRole('region', { name: 'About' });
    expect(within(about).getByText('Licence')).toBeVisible();
    expect(within(about).getByText('MIT')).toBeVisible();
  });
});

describe('ProjectPage owner edit of stage, licence and country', () => {
  it('starts from the current values', async () => {
    setLanguages(['fr-FR']);
    owner(routesFor({ stage: 'active', license_spdx: 'MIT', country_code: 'BF' }));
    await edit();
    expect(screen.getByRole('textbox', { name: /Licence/ })).toHaveValue('MIT');
    expect(screen.getByRole('combobox', { name: /^Étape/ })).toHaveTextContent('Actif');
    expect(screen.getByRole('button', { name: /^Pays/ })).toHaveTextContent('Burkina Faso');
  });

  it('patches the changed stage, licence and country only', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = owner({
      ...routesFor(),
      [`PATCH ${base}`]: () =>
        jsonResponse(project({ stage: 'prototype', license_spdx: 'MIT', country_code: 'SN' })),
    });
    const user = await edit();
    await pick(user, /^Étape/, 'Prototype');
    await user.type(screen.getByRole('textbox', { name: /Licence/ }), 'MIT');
    await pick(user, /^Pays/, 'Sénégal');
    await user.click(screen.getByRole('button', { name: 'Enregistrer' }));
    await waitFor(() => {
      expect(callsTo(fetchMock, `PATCH ${base}`)).toHaveLength(1);
    });
    const request = callsTo(fetchMock, `PATCH ${base}`)[0];
    expect(request?.headers.get('X-CSRF-Token')).toBeTruthy();
    expect(await request?.clone().json()).toEqual({
      stage: 'prototype',
      license_spdx: 'MIT',
      country_code: 'SN',
    });
    expect(await screen.findByText('Prototype')).toBeVisible();
  });

  it('clears the licence and the country with null', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = owner({
      ...routesFor({ license_spdx: 'MIT', country_code: 'BF' }),
      [`PATCH ${base}`]: () => jsonResponse(project()),
    });
    const user = await edit();
    await user.clear(screen.getByRole('textbox', { name: /Licence/ }));
    await user.click(screen.getByRole('button', { name: /Effacer.*Pays|Pays.*effacer/i }));
    await user.click(screen.getByRole('button', { name: 'Enregistrer' }));
    await waitFor(() => {
      expect(callsTo(fetchMock, `PATCH ${base}`)).toHaveLength(1);
    });
    expect(await callsTo(fetchMock, `PATCH ${base}`)[0]?.clone().json()).toEqual({
      license_spdx: null,
      country_code: null,
    });
  });

  it('refuses a licence that is not an SPDX identifier, without calling the server', async () => {
    setLanguages(['fr-FR']);
    const fetchMock = owner(routesFor());
    const user = await edit();
    await user.type(screen.getByRole('textbox', { name: /Licence/ }), 'GPL 3 only');
    await user.click(screen.getByRole('button', { name: 'Enregistrer' }));
    expect(await screen.findByText(/Identifiant SPDX non valide/)).toBeVisible();
    expect(callsTo(fetchMock, `PATCH ${base}`)).toHaveLength(0);
  });

  it('shows a 422 on /stage under the stage field', async () => {
    setLanguages(['fr-FR']);
    owner({
      ...routesFor(),
      [`PATCH ${base}`]: () =>
        problemResponse(422, 'VALIDATION_ERROR', {
          violations: [{ field: '/stage', code: 'invalid_value' }],
        }),
    });
    const user = await edit();
    await pick(user, /^Étape/, 'Actif');
    await user.click(screen.getByRole('button', { name: 'Enregistrer' }));
    expect(await screen.findByText('Cette étape demande un dépôt GitHub.')).toBeVisible();
  });

  it('shows a 422 on /country_code under the country field', async () => {
    setLanguages(['fr-FR']);
    owner({
      ...routesFor(),
      [`PATCH ${base}`]: () =>
        problemResponse(422, 'VALIDATION_ERROR', {
          violations: [{ field: '/country_code', code: 'invalid_value' }],
        }),
    });
    const user = await edit();
    await pick(user, /^Pays/, 'Sénégal');
    await user.click(screen.getByRole('button', { name: 'Enregistrer' }));
    expect(await screen.findByText('Pays inconnu.')).toBeVisible();
  });
});
