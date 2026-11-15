import { expect, test } from '@playwright/test';
import { credentials, fr, signIn } from './support';

// Demo seed accounts have no password (db/seeds/README.md, section 1.1), so a staging account that can
// publish (verified e-mail) comes from the environment. Without it the whole group is skipped, loudly.
const account = credentials();

test.describe('Signed-in member', () => {
  test.skip(
    account === undefined,
    'E2E_USER and E2E_PASSWORD are not set: signed-in flows are skipped (see docs/TEST.md, section 6 ter)',
  );

  test('signs in, asks a question, answers it and signs out', async ({ page }) => {
    if (account === undefined) return;
    const stamp = `${String(Date.now())}-${test.info().project.name}`;
    const title = `E2E question ${stamp}`;
    const answer = `E2E réponse ${stamp}`;

    await signIn(page, account.email, account.password);

    await page.goto('/solve');
    await page.getByRole('button', { name: 'Poser une question' }).click();
    await page.getByRole('textbox', { name: 'Titre', exact: true }).fill(title);
    await page
      .getByRole('textbox', { name: 'Votre question', exact: true })
      .fill(`Question créée par le test de bout en bout ${stamp}. Elle peut être ignorée.`);
    await page.getByRole('textbox', { name: 'Étiquettes', exact: true }).fill('api');
    await page.getByRole('button', { name: 'Publier la question' }).click();

    // The question has its own page with the title as heading.
    await expect(page.getByRole('heading', { level: 1, name: title })).toBeVisible();
    await expect(page).toHaveURL(/\/questions\/[0-9a-f-]+$/);

    await page.getByRole('textbox', { name: 'Rédigez votre réponse', exact: true }).fill(answer);
    await page.getByRole('button', { name: 'Publier la réponse' }).click();
    await expect(page.getByText(answer)).toBeVisible();

    // The question is listed.
    await page.goto('/solve');
    await expect(page.getByRole('link', { name: title })).toBeVisible();

    // Sign out: the account menu is the only button of the top bar once signed in.
    await page.getByRole('banner').getByRole('button').click();
    await page.getByRole('menuitem', { name: fr.account.signOut }).click();
    await expect(
      page.getByRole('banner').getByRole('link', { name: fr.account.signIn }),
    ).toBeVisible();
  });
});
