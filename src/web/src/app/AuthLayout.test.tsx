import { screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { mockApi } from '../test/api';
import { renderApp, setLanguages } from '../test/render';

describe('AuthLayout', () => {
  it('shows the 226os logo, linking home, above the account card', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/connexion');
    const main = await screen.findByRole('main');
    const home = within(main).getByRole('link', { name: '226os' });
    expect(home).toHaveAttribute('href', '/');
    // The logo image itself, as in prototypes/html/pages/connexion.html, not a text link.
    const logo = within(home).getByRole('img', { name: '226os' });
    expect(logo).toHaveAttribute('width', '92');
    expect(logo).toHaveAttribute('height', '32');
    expect(home.textContent).toBe('');
    // Logo first, then the card with the page's one h1.
    const heading = within(main).getByRole('heading', { level: 1 });
    expect(Boolean(home.compareDocumentPosition(heading) & Node.DOCUMENT_POSITION_FOLLOWING)).toBe(
      true,
    );
  });

  it('switches to the white logo on the dark theme', async () => {
    setLanguages(['fr-FR']);
    localStorage.setItem('226os.theme', 'dark');
    mockApi();
    await renderApp('/connexion');
    const logo = within(await screen.findByRole('main')).getByRole('img', { name: '226os' });
    expect(logo.getAttribute('src')).toMatch(/226os-logo-white/);
  });

  it('links to no placeholder page under the card', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    await renderApp('/connexion');
    const main = await screen.findByRole('main');
    await within(main).findByRole('heading', { level: 1 });
    for (const link of screen.getAllByRole('link')) {
      expect(link.getAttribute('href') ?? '').not.toMatch(/example\.org/);
    }
  });
});
