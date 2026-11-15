import { expect, test } from '@playwright/test';
import { fr, openFirstItem } from './support';

test.describe('Anonymous visitor', () => {
  test('the landing page loads with its title and its own pillar links', async ({ page }) => {
    // The landing carries no side navigation: only its own « Voir … » links are checked.
    await page.goto('/');
    await expect(page).toHaveTitle(/226os/);
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
    for (const name of [
      'Voir les articles',
      'Voir les questions',
      'Voir le catalogue',
      'Voir les projets',
    ]) {
      await expect(page.getByRole('main').getByRole('link', { name, exact: true })).toBeVisible();
    }
  });

  test('Découvrir lists catalogue entries and an entry opens', async ({ page }) => {
    await openFirstItem(page, '/discover', 'Résultats');
    await expect(page.getByRole('navigation', { name: 'Fil d’Ariane' })).toBeVisible();
  });

  test('Résoudre lists questions and a question opens', async ({ page }) => {
    await openFirstItem(page, '/solve', 'Questions');
    // The answers heading counts them (« 1 réponse », « 2 réponses ») or reads « Réponses ».
    await expect(
      page.getByRole('heading', { level: 2, name: /^(Réponses|\d+ réponses?)$/ }),
    ).toBeVisible();
  });

  test('Construire lists projects and a project opens', async ({ page }) => {
    await openFirstItem(page, '/build', 'Projets');
  });

  test('Apprendre lists articles and an article opens', async ({ page }) => {
    await page.goto('/apprendre');
    await expect(page.getByRole('heading', { level: 1, name: 'Apprendre' })).toBeVisible();
    // One list for every type since the filters pass (learn.browse.label).
    const all = page.getByRole('region', { name: 'Tous les contenus', exact: true });
    await expect(all.getByRole('listitem').first()).toBeVisible();
    await openFirstItem(page, '/apprendre', 'Tous les contenus');
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
    await page.goto('/discover');
    const nav = page.getByRole('navigation', { name: fr.nav.label });
    await expect(nav.getByRole('link', { name: fr.nav.solve, exact: true })).toBeVisible();

    await page.getByRole('button', { name: fr.account.prefs }).click();
    await page.getByRole('menuitemradio', { name: 'English' }).click();
    await expect(
      page
        .getByRole('navigation', { name: 'Main navigation' })
        .getByRole('link', { name: 'Solve', exact: true }),
    ).toBeVisible();

    await page.getByRole('button', { name: 'Preferences' }).click();
    await page.getByRole('menuitemradio', { name: 'Français' }).click();
    await expect(nav.getByRole('link', { name: fr.nav.solve, exact: true })).toBeVisible();
  });
});
