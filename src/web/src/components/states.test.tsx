import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { ApiError } from '../api/http';
import { renderWithProviders, setLanguages } from '../test/render';
import { DemoBadge } from './DemoBadge';
import { errorMessage } from './errorMessage';
import { ErrorState } from './ErrorState';
import { LoadingState } from './LoadingState';

describe('errorMessage', () => {
  it.each([
    [new ApiError({ status: 0, code: 'NETWORK' }), 'error.NETWORK'],
    [new ApiError({ status: 404, code: 'NOT_FOUND' }), 'error.NOT_FOUND'],
    [new ApiError({ status: 403, code: 'CSRF_FAILED' }), 'error.CSRF_FAILED'],
    [new ApiError({ status: 409, code: 'STATE_CONFLICT' }), 'error.generic'],
    [new ApiError({ status: 502, code: 'UNKNOWN' }), 'error.generic'],
    [new Error('boom'), 'error.generic'],
  ])('maps %o to %s', (error, key) => {
    expect(errorMessage(error).key).toBe(key);
  });

  it('tells how long to wait on 429 when Retry-After is known', () => {
    expect(
      errorMessage(new ApiError({ status: 429, code: 'RATE_LIMITED', retryAfter: 30 })),
    ).toEqual({ key: 'error.RATE_LIMITED', params: { seconds: 30 } });
    expect(errorMessage(new ApiError({ status: 429, code: 'RATE_LIMITED' })).key).toBe(
      'error.RATE_LIMITED.now',
    );
  });
});

describe('ErrorState', () => {
  it('explains the failure, shows the support reference and retries', async () => {
    setLanguages(['fr-FR']);
    const onRetry = vi.fn();
    renderWithProviders(
      <ErrorState
        error={new ApiError({ status: 0, code: 'NETWORK', requestId: 'req-9' })}
        onRetry={onRetry}
      />,
    );
    const alert = screen.getByRole('alert');
    expect(alert).toHaveTextContent('Impossible de charger le contenu');
    expect(alert).toHaveTextContent('Le serveur ne répond pas.');
    expect(alert).toHaveTextContent(/Référence\s:\sreq-9/);
    await userEvent.setup({ delay: null }).click(screen.getByRole('button', { name: 'Réessayer' }));
    expect(onRetry).toHaveBeenCalledOnce();
  });

  it('offers no retry button when no retry is possible', () => {
    setLanguages(['en-US']);
    renderWithProviders(<ErrorState error={new Error('x')} />);
    expect(screen.getByRole('alert')).toHaveTextContent('Something went wrong.');
    expect(screen.queryByRole('button', { name: 'Try again' })).not.toBeInTheDocument();
  });
});

describe('LoadingState and DemoBadge', () => {
  it('announce loading and label demo data', () => {
    setLanguages(['fr-FR']);
    renderWithProviders(
      <>
        <LoadingState />
        <DemoBadge />
      </>,
    );
    expect(screen.getAllByRole('status')[0]).toHaveTextContent('Chargement…');
    expect(screen.getByText('Démo')).toBeInTheDocument();
    expect(screen.getByTitle('Donnée d’exemple, pas une activité réelle.')).toBeInTheDocument();
  });
});
