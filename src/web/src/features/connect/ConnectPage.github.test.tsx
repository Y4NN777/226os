import { screen, waitFor, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { components } from '../../api/schema';
import { callsTo, mockApi } from '../../test/api';
import { jsonResponse, renderApp, setLanguages } from '../../test/render';
import { selectSegment } from '../../test/segmented';
import { validateConnectSearch } from './routeSearch';

type Summary = components['schemas']['ProfileSummary'];
type External = components['schemas']['ExternalDeveloper'];

const awa: Summary = { username: 'awa', display_name: 'Awa Traoré', country: 'BF' };
const ade: External = {
  login: 'ade-dev',
  name: 'Adé Ouédraogo',
  avatar_url: 'https://avatars.githubusercontent.com/u/1?v=4',
  country: 'BF',
  rank: 3,
  contributions: 4321,
  organizations: ['acme'],
  generated_at: '2026-11-14T06:00:00Z',
  source: 'committers.top',
};
const koffi: External = { ...ade, login: 'koffi', name: 'Koffi', country: 'CI', rank: 1 };

function page(items: Summary[], external?: External[], next: string | null = null) {
  return jsonResponse({ items, next_cursor: next, ...(external ? { external } : {}) });
}

function params(request: Request | undefined) {
  return new URL(request?.url ?? 'http://x').searchParams;
}

describe('ConnectPage GitHub developers', () => {
  it('lists external developers after the members, labelled with their rank and source (FR)', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /profiles': () => page([awa], [ade, koffi]) });
    await renderApp('/connect');
    const name = await screen.findByRole('link', { name: /Adé Ouédraogo/ });
    expect(name).toHaveAttribute('href', 'https://github.com/ade-dev');
    expect(name).toHaveAttribute('target', '_blank');
    const row = name.closest('li') as HTMLElement;
    expect(within(row).getByText(/^Profil GitHub public/).textContent).toBe(
      'Profil GitHub public · n°\u00A03 au Burkina Faso (committers.top)',
    );
    expect(
      within(row).getByRole('link', { name: /^C’est vous\s\? Rejoignez 226os$/ }),
    ).toHaveAttribute('href', '/inscription');
    // French typography: non-breaking spaces after n° and before ? and :.
    expect(row.textContent).toContain('n°\u00A03');
    expect(row.textContent).toContain('vous\u00A0?');
    const img = row.querySelector('img') as HTMLImageElement;
    expect(img.src).toBe('https://avatars.githubusercontent.com/u/1?v=4');
    expect(img.getAttribute('referrerpolicy')).toBe('no-referrer');
    const ci = (await screen.findByRole('link', { name: /Koffi/ })).closest('li') as HTMLElement;
    expect(within(ci).getByText(/n°\s1 en Côte d’Ivoire/)).toBeVisible();
    expect(
      screen.getByText(/^Source\s: committers\.top · généré le 14 novembre 2026$/),
    ).toBeVisible();
    // A private channel: a mail to support, never a public issue.
    const optout = screen.getByRole('link', { name: /^Demander le retrait/ });
    expect(optout).toHaveAttribute(
      'href',
      'mailto:support@226os.tech?subject=Retrait%20de%20l%E2%80%99annuaire%20Connect',
    );
    expect(optout).not.toHaveAttribute('target');
    // Members keep their own list, first.
    expect(screen.getByRole('list', { name: 'Membres' })).toContainElement(
      screen.getByRole('link', { name: 'Awa Traoré' }),
    );
    expect(screen.getByRole('list', { name: 'Développeurs GitHub' })).toContainElement(row);
  });

  it('speaks English too', async () => {
    setLanguages(['en-US']);
    mockApi({ 'GET /profiles': () => page([], [ade]) });
    await renderApp('/connect');
    const row = (await screen.findByRole('link', { name: /Adé Ouédraogo/ })).closest(
      'li',
    ) as HTMLElement;
    expect(
      within(row).getByText('Public GitHub profile · No. 3 in Burkina Faso (committers.top)'),
    ).toBeVisible();
    expect(within(row).getByRole('link', { name: 'Is this you? Join 226os' })).toBeVisible();
    expect(
      screen.getByText('Source: committers.top · generated on November 14, 2026'),
    ).toBeVisible();
  });

  it('draws no avatar from another host and no link for a malformed login', async () => {
    setLanguages(['fr-FR']);
    mockApi({
      'GET /profiles': () =>
        page([], [{ ...ade, login: '../x', avatar_url: 'https://evil.example/a.png' }]),
    });
    await renderApp('/connect');
    const row = (await screen.findByText(/Profil GitHub public/)).closest('li') as HTMLElement;
    expect(row.querySelector('img')).toBeNull();
    expect(within(row).queryByRole('link', { name: /Adé/ })).not.toBeInTheDocument();
  });

  it('sends no source by default and the GitHub segment once picked', async () => {
    setLanguages(['fr-FR']);
    const api = mockApi({ 'GET /profiles': () => page([awa], [ade]) });
    await renderApp('/connect');
    await screen.findByRole('link', { name: /Adé Ouédraogo/ });
    expect(params(callsTo(api, 'GET /profiles')[0]).has('source')).toBe(false);
    await selectSegment(screen.getByRole('radio', { name: 'GitHub' }));
    await waitFor(() => {
      expect(params(callsTo(api, 'GET /profiles').at(-1)).get('source')).toBe('github');
    });
  });

  it('sends the source from the address and checks its segment', async () => {
    setLanguages(['fr-FR']);
    const api = mockApi({ 'GET /profiles': () => page([awa]) });
    await renderApp('/connect?source=members');
    await screen.findByRole('link', { name: 'Awa Traoré' });
    expect(params(callsTo(api, 'GET /profiles')[0]).get('source')).toBe('members');
    expect(screen.getByRole('radio', { name: 'Membres' })).toHaveAttribute('aria-checked', 'true');
  });

  it('hides the members section when only GitHub developers are asked for', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /profiles': () => page([], [ade]) });
    await renderApp('/connect?source=github');
    await screen.findByRole('link', { name: /Adé Ouédraogo/ });
    expect(screen.queryByRole('heading', { name: 'Membres' })).not.toBeInTheDocument();
  });

  it('validates the source in the route search', () => {
    expect(validateConnectSearch({ source: 'github' })).toEqual({ source: 'github' });
    expect(validateConnectSearch({ source: 'members' })).toEqual({ source: 'members' });
    expect(validateConnectSearch({ source: 'all' })).toEqual({});
    expect(validateConnectSearch({ source: 'everyone' })).toEqual({});
  });
});
