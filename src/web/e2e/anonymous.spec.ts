import { expect, test } from '@playwright/test';
import { fr, openFirstItem } from './support';

test.describe('Anonymous visitor', () => {
  test('the landing page loads with its title and the pillars navigation', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveTitle(/226os/);
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
    const nav = page.getByRole('navigation', { name: fr.nav.label });
    for (const name of [fr.nav.learn, fr.nav.solve, fr.nav.discover, 'Construire']) {
      await expect(nav.getByRole('link', { name, exact: true })).toBeVisible();
    }
  });

  test('Découvrir lists catalogue entries and an entry opens', async ({ page }) => {
    await openFirstItem(page, '/discover', 'Résultats');
    await expect(page.getByRole('navigation', { name: 'Fil d’Ariane' })).toBeVisible();
  });

  test('Résoudre lists questions and a question opens', async ({ page }) => {
    await openFirstItem(page, '/solve', 'Questions');
    await expect(page.getByRole('heading', { level: 2, name: 'Réponses' })).toBeVisible();
  });

  test('Construire lists projects and a project opens', async ({ page }) => {
    await openFirstItem(page, '/build', 'Projets');
  });

  test('Apprendre lists articles and an article opens', async ({ page }) => {
    await page.goto('/apprendre');
    await expect(page.getByRole('heading', { level: 1, name: 'Apprendre' })).toBeVisible();
    const articles = page.getByRole('region', { name: 'Articles', exact: true });
    await expect(articles.getByRole('listitem').first()).toBeVisible();
    await openFirstItem(page, '/apprendre', 'Articles');
  });

  test('the site search returns results', async ({ page }) => {
    await page.goto('/recherche');
    await page.getByRole('searchbox', { name: 'Votre recherche' }).fill('webhook');
    await page.getByRole('button', { name: 'Rechercher', exact: true }).click();
    // At least one result group, and not the empty state.
    await expect(page.getByRole('main').getByRole('listitem').first()).toBeVisible();
    await expect(page.getByRole('heading', { name: /Aucun résultat/ })).toHaveCount(0);
  });

  test('the language switches to English and back to French', async ({ page }) => {
    await page.goto('/');
    const nav = page.getByRole('navigation', { name: fr.nav.label });
    await expect(nav.getByRole('link', { name: fr.nav.solve, exact: true })).toBeVisible();

    await page.getByRole('button', { name: fr.account.more }).click();
    await page.getByRole('menuitemradio', { name: 'English' }).click();
    await expect(
      page
        .getByRole('navigation', { name: 'Main navigation' })
        .getByRole('link', { name: 'Solve', exact: true }),
    ).toBeVisible();

    await page.getByRole('button', { name: 'More options' }).click();
    await page.getByRole('menuitemradio', { name: 'Français' }).click();
    await expect(nav.getByRole('link', { name: fr.nav.solve, exact: true })).toBeVisible();
  });
});
