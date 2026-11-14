import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { jsonResponse, renderWithProviders, setLanguages } from '../../test/render';
import { HealthStatus } from './HealthStatus';
import { findVisibleText } from '../../test/visibleText';

describe('HealthStatus', () => {
  it('announces loading, then success, with a text label and the raw status', async () => {
    setLanguages(['fr-FR']);
    vi.stubGlobal(
      'fetch',
      vi.fn(() => Promise.resolve(jsonResponse({ status: 'ok' }))),
    );
    renderWithProviders(<HealthStatus />);
    expect(
      screen.getByText('Vérification de l’API en cours…').closest('[role="status"]'),
    ).toBeInTheDocument();
    expect(await screen.findByText('API opérationnelle')).toBeInTheDocument();
    // Testing Library collapses U+00A0 into a plain space, so compare the raw text.
    expect(screen.getByText((_, el) => el?.textContent === 'Statut\u00A0: ok')).toBeInTheDocument();
    expect(fetch).toHaveBeenCalledWith('/healthz', expect.anything());
  });

  it('shows a localized error with guidance when the API is unreachable', async () => {
    setLanguages(['en-US']);
    vi.stubGlobal(
      'fetch',
      vi.fn(() => Promise.reject(new TypeError('offline'))),
    );
    renderWithProviders(<HealthStatus />);
    expect(await findVisibleText('The API is not responding.')).toBeInTheDocument();
    expect(screen.getByText(/check that the server is running/i)).toBeInTheDocument();
  });

  it('shows the error state for a 503 response', async () => {
    setLanguages(['fr-FR']);
    vi.stubGlobal(
      'fetch',
      vi.fn(() => Promise.resolve(jsonResponse({ status: 'ok' }, { status: 503 }))),
    );
    renderWithProviders(<HealthStatus />);
    expect(await findVisibleText('L’API ne répond pas.')).toBeInTheDocument();
  });

  it('recovers when the user retries after a failure', async () => {
    setLanguages(['en-US']);
    const fetchMock = vi
      .fn<typeof fetch>()
      .mockRejectedValueOnce(new TypeError('offline'))
      .mockResolvedValue(jsonResponse({ status: 'ok' }));
    vi.stubGlobal('fetch', fetchMock);
    renderWithProviders(<HealthStatus />);
    await findVisibleText('The API is not responding.');
    await userEvent
      .setup({ delay: null })
      .click(screen.getByRole('button', { name: 'Check again' }));
    expect(await screen.findByText('API is up')).toBeInTheDocument();
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
  });
});
