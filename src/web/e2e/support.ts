import { expect, type Page } from '@playwright/test';

// French labels come from the catalogues in src/i18n/messages.ts and src/features/<feature>/messages.ts.
// They are repeated here on purpose: a renamed label must fail the suite, not follow it silently.
export const fr = {
  nav: {
    label: 'Navigation principale',
    learn: 'Apprendre',
    solve: 'Résoudre',
    discover: 'Découvrir',
  },
  account: { signIn: 'Se connecter', signOut: 'Se déconnecter', more: 'Plus d’options' },
} as const;

/** Open `path`, then open the first card of the list region named `region` and check the detail heading. */
export async function openFirstItem(page: Page, path: string, region: string): Promise<void> {
  await page.goto(path);
  const list = page.getByRole('region', { name: region, exact: true });
  const link = list.getByRole('listitem').first().getByRole('link').first();
  await expect(link).toBeVisible();
  const name = (await link.innerText()).trim();
  expect(name).not.toBe('');
  await link.click();
  await expect(page.getByRole('heading', { level: 1 })).toHaveText(name);
}

/** Credentials of a staging account that can publish (verified e-mail). Seeds have no password: see docs/TEST.md. */
export function credentials(): { email: string; password: string } | undefined {
  const email = process.env['E2E_USER'];
  const password = process.env['E2E_PASSWORD'];
  return email && password ? { email, password } : undefined;
}

export async function signIn(page: Page, email: string, password: string): Promise<void> {
  await page.goto('/connexion');
  await page.getByRole('textbox', { name: 'Adresse e-mail' }).fill(email);
  await page.getByRole('textbox', { name: 'Mot de passe', exact: true }).fill(password);
  await page.getByRole('button', { name: 'Se connecter' }).click();
  // The top bar leaves the « Se connecter » link once the session exists.
  await expect(page).not.toHaveURL(/\/connexion/);
  await expect(page.getByRole('banner').getByRole('link', { name: fr.account.signIn })).toHaveCount(
    0,
  );
}
