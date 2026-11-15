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
    await userEvent
      .setup({ delay: null })
      .click((await screen.findAllByRole('link', { name: 'Se connecter' }))[0] as HTMLElement);
    expect(router.state.location.pathname).toBe('/connexion');
  });

  it('leaves modified clicks to the browser', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    const { router } = await renderApp('/');
    // The top bar link comes first; the landing page repeats the same link in its content.
    const link = (await screen.findAllByRole('link', { name: 'Se connecter' }))[0] as HTMLElement;
    const user = userEvent.setup({ delay: null });
    await user.keyboard('{Control>}');
    await user.click(link);
    await user.keyboard('{/Control}');
    expect(router.state.location.pathname).toBe('/');
  });
});
