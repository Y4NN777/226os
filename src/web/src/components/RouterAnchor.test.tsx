import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { mockApi } from '../test/api';
import { renderApp, setLanguages } from '../test/render';

describe('RouterAnchor (via LinkProvider)', () => {
  it('navigates in-app links without a reload', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    const { router } = await renderApp('/');
    await userEvent.click(await screen.findByRole('link', { name: 'Se connecter' }));
    expect(router.state.location.pathname).toBe('/connexion');
  });

  it('leaves modified clicks to the browser', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    const { router } = await renderApp('/');
    const link = await screen.findByRole('link', { name: 'Se connecter' });
    const user = userEvent.setup();
    await user.keyboard('{Control>}');
    await user.click(link);
    await user.keyboard('{/Control}');
    expect(router.state.location.pathname).toBe('/');
  });
});
