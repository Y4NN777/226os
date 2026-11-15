import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { mockApi } from '../../test/api';
import { renderApp, setLanguages } from '../../test/render';
import { page, project } from './testFixtures';

const second = project({
  id: 'e0000000-0000-4000-8000-000000000002',
  name: 'sahel-docs',
  github_url: 'https://gitlab.com/sahel/sahel-docs',
  technologies: ['markdown'],
});

async function listOf() {
  await screen.findByRole('link', { name: 'kora-sdk' });
  return screen.getByRole('list', { name: 'Projets' });
}

describe('BuildPage card and list views', () => {
  it('shows cards by default, each a link, with the preview image or the tile', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /projects': () => page([project(), second]) });
    const { container } = await renderApp('/build');
    const list = await listOf();
    expect(list).toHaveClass('ui-card-grid');
    expect(screen.getByRole('button', { name: 'Cartes' })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: 'Liste' })).toHaveAttribute('aria-pressed', 'false');
    expect(screen.getByRole('link', { name: 'kora-sdk' })).toHaveAttribute(
      'href',
      '/build/e0000000-0000-4000-8000-000000000001',
    );
    const [withPreview, withoutPreview] = Array.from(list.children);
    expect(withPreview?.querySelector('img')?.getAttribute('src')).toBe(
      'https://opengraph.githubassets.com/1/kora-dev/kora-sdk',
    );
    expect(withoutPreview?.querySelector('img')).toBeNull();
    expect(withoutPreview?.textContent).toContain('S');
    expect(container.querySelectorAll('.build-thumb-card')).toHaveLength(1);
  });

  it('switches to the list and back, keeping aria-pressed and the address in step', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /projects': () => page([project()]) });
    const user = userEvent.setup({ delay: null });
    const { router } = await renderApp('/build');
    await listOf();

    await user.click(screen.getByRole('button', { name: 'Liste' }));
    await waitFor(() => {
      expect(screen.getByRole('list', { name: 'Projets' })).not.toHaveClass('ui-card-grid');
    });
    expect(screen.getByRole('button', { name: 'Liste' })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: 'Cartes' })).toHaveAttribute('aria-pressed', 'false');
    expect(router.state.location.search).toEqual({ view: 'list' });

    await user.click(screen.getByRole('button', { name: 'Cartes' }));
    await waitFor(() => {
      expect(screen.getByRole('list', { name: 'Projets' })).toHaveClass('ui-card-grid');
    });
    expect(router.state.location.search).toEqual({ view: 'cards' });
  });

  it('honours the view in the address and keeps the filters when switching', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /projects': () => page([project()]) });
    const user = userEvent.setup({ delay: null });
    const { router } = await renderApp('/build?stage=idea&sort=needs&view=list');
    const list = await listOf();
    expect(list).not.toHaveClass('ui-card-grid');
    expect(screen.getByRole('button', { name: 'Liste' })).toHaveAttribute('aria-pressed', 'true');

    await user.click(screen.getByRole('button', { name: 'Cartes' }));
    await waitFor(() => {
      expect(router.state.location.search).toEqual({
        stage: 'idea',
        sort: 'needs',
        view: 'cards',
      });
    });
  });

  it('keeps the view in the address when a filter changes', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /projects': () => page([project()]) });
    const user = userEvent.setup({ delay: null });
    const { router } = await renderApp('/build?view=list');
    await listOf();
    await user.click(screen.getByRole('checkbox', { name: /besoin/i }));
    await waitFor(() => {
      expect(router.state.location.search).toEqual({ open: true, view: 'list' });
    });
  });

  it('uses the choice remembered in this browser when the address has none', async () => {
    setLanguages(['fr-FR']);
    localStorage.setItem('226os.view.build', 'list');
    mockApi({ 'GET /projects': () => page([project()]) });
    await renderApp('/build');
    expect(await listOf()).not.toHaveClass('ui-card-grid');
  });

  it('lets the address win over the remembered choice', async () => {
    setLanguages(['fr-FR']);
    localStorage.setItem('226os.view.build', 'list');
    mockApi({ 'GET /projects': () => page([project()]) });
    await renderApp('/build?view=cards');
    expect(await listOf()).toHaveClass('ui-card-grid');
  });

  it('remembers the choice made with the switch', async () => {
    setLanguages(['fr-FR']);
    mockApi({ 'GET /projects': () => page([project()]) });
    const user = userEvent.setup({ delay: null });
    await renderApp('/build');
    await listOf();
    await user.click(screen.getByRole('button', { name: 'Liste' }));
    expect(localStorage.getItem('226os.view.build')).toBe('list');
  });

  it('is labelled in English', async () => {
    setLanguages(['en-US']);
    mockApi({ 'GET /projects': () => page([project()]) });
    await renderApp('/build');
    await screen.findByRole('link', { name: 'kora-sdk' });
    expect(screen.getByRole('group', { name: 'View' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Cards' })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: 'List' })).toHaveAttribute('aria-pressed', 'false');
  });
});
