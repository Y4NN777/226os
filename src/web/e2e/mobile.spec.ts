import { expect, test } from '@playwright/test';

// Phone project (375 px). The bottom bar (« Navigation rapide ») arrives with the prototype shell;
// the drawer trigger is « Plus » in that bar, or « Ouvrir la navigation » in the top bar before it.
const drawerTrigger = { name: /^(Plus|Ouvrir la navigation)$/ };

test.describe('Mobile navigation', () => {
  test('the bottom bar reaches Résoudre and Découvrir', async ({ page }) => {
    await page.goto('/discover');
    const bar = page.getByRole('navigation', { name: 'Navigation rapide' });
    await expect(bar).toBeVisible();

    await bar.getByRole('link', { name: 'Résoudre' }).click();
    await expect(page).toHaveURL(/\/solve$/);
    await expect(page.getByRole('heading', { level: 1, name: 'Résoudre' })).toBeVisible();

    await bar.getByRole('link', { name: 'Découvrir' }).click();
    await expect(page).toHaveURL(/\/discover$/);
    await expect(page.getByRole('heading', { level: 1, name: 'Découvrir' })).toBeVisible();
  });

  test('the drawer opens with the pillars and closes with Escape', async ({ page }) => {
    await page.goto('/discover');
    const trigger = page.getByRole('button', drawerTrigger);
    await expect(trigger).toHaveAttribute('aria-expanded', 'false');

    await trigger.click();
    await expect(trigger).toHaveAttribute('aria-expanded', 'true');
    await expect(page.getByRole('link', { name: 'Construire', exact: true })).toBeVisible();

    await page.keyboard.press('Escape');
    await expect(trigger).toHaveAttribute('aria-expanded', 'false');
    await expect(page.getByRole('link', { name: 'Construire', exact: true })).toBeHidden();
  });

  test('no horizontal scroll on a catalogue page', async ({ page }) => {
    await page.goto('/discover');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
    const overflow = await page.evaluate(
      () => document.documentElement.scrollWidth - document.documentElement.clientWidth,
    );
    expect(overflow).toBeLessThanOrEqual(0);
  });
});
