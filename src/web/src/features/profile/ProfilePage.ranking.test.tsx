import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { components } from '../../api/schema';
import { mockApi } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';

type PublicProfile = components['schemas']['PublicProfile'];

const profile: PublicProfile = {
  username: 'awa',
  display_name: 'Awa Traoré',
  github: 'awa-dev',
  profile_visibility: 'public',
  joined_at: '2026-11-01T10:00:00Z',
};

describe('ProfilePage GitHub rank', () => {
  it('shows the committers.top rank in the hero (FR)', async () => {
    setLanguages(['fr-FR']);
    mockApi({
      'GET /profiles/awa': () =>
        jsonResponse({
          ...profile,
          ranking: { country: 'SN', rank: 12, generated_at: '2026-11-14T06:00:00Z' },
        }),
    });
    await renderApp('/profils/awa');
    expect((await screen.findByText(/sur GitHub/)).textContent).toBe(
      'N°\u00A012 sur GitHub au Sénégal (committers.top)',
    );
  });

  it('shows it in English too', async () => {
    setLanguages(['en-US']);
    mockApi({
      'GET /profiles/awa': () =>
        jsonResponse({
          ...profile,
          ranking: { country: 'SN', rank: 12, generated_at: '2026-11-14T06:00:00Z' },
        }),
    });
    await renderApp('/profils/awa');
    expect(await screen.findByText('No. 12 on GitHub in Senegal (committers.top)')).toBeVisible();
  });

  it('shows nothing without a ranking', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /profiles/awa': () => jsonResponse(profile) });
    await renderApp('/profils/awa');
    await screen.findByRole('heading', { level: 1, name: 'Awa Traoré' });
    expect(screen.queryByText(/committers\.top/)).not.toBeInTheDocument();
  });
});
