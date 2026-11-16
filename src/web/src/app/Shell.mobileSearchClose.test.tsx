import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { mockApi } from '../test/api';
import { renderApp, setLanguages } from '../test/render';

/** Below the phone breakpoint: every media query matches. */
function stubPhone() {
  vi.stubGlobal(
    'matchMedia',
    vi.fn((query: string) => ({
      matches: true,
      media: query,
      onchange: null,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn(),
    })),
  );
  HTMLDialogElement.prototype.showModal = function showModal(this: HTMLDialogElement) {
    this.setAttribute('open', '');
  };
  HTMLDialogElement.prototype.close = function close(this: HTMLDialogElement) {
    this.removeAttribute('open');
  };
}

describe('Shell mobile search close button', () => {
  it('names the close button « Fermer la recherche » and closes the overlay', async () => {
    setLanguages(['fr-FR']);
    mockApi();
    stubPhone();
    await renderApp('/parametres');
    await screen.findByRole('navigation', { name: 'Barre du site' });
    const header = screen.getByRole('navigation', { name: 'Barre du site' });
    const user = userEvent.setup({ delay: null });
    await user.click(within(header).getByRole('button', { name: 'Rechercher sur le site' }));
    expect(within(header).getByRole('searchbox')).toBeInTheDocument();
    await user.click(within(header).getByRole('button', { name: /Fermer la recherche/ }));
    await waitFor(() => {
      expect(within(header).queryByRole('searchbox')).not.toBeInTheDocument();
    });
  });
});
