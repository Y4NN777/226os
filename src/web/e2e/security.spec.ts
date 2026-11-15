import { expect, test } from '@playwright/test';

test.describe('Security smoke', () => {
  test('an unsafe call without the CSRF header is refused with 403', async ({ page }) => {
    // Same browser context as the app: cookies and Origin are the ones a real page would send.
    await page.goto('/');
    const response = await page.request.post('/api/v1/auth/sign-in', {
      headers: { 'content-type': 'application/json' },
      data: { email: 'nobody@example.invalid', password: 'not-a-real-password' },
    });
    expect(response.status()).toBe(403);
    expect(response.headers()['content-type']).toContain('application/problem+json');
    const problem = (await response.json()) as { code?: string; status?: number };
    expect(problem.code).toBe('CSRF_FAILED');
    expect(problem.status).toBe(403);
  });

  for (const path of ['/api/v1/questions', '/api/v1/resources', '/api/v1/projects']) {
    test(`anonymous GET ${path} sets no cookie`, async ({ request }) => {
      const response = await request.get(path);
      expect(response.status()).toBe(200);
      expect(
        response.headersArray().filter((header) => header.name.toLowerCase() === 'set-cookie'),
      ).toEqual([]);
    });
  }

  test('the session cookie, when the app sets one, is HttpOnly and SameSite', async ({ page }) => {
    // The app only issues the pre-session CSRF cookie to anonymous visitors; no session cookie exists yet.
    await page.goto('/connexion');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
    const cookies = await page.context().cookies();
    for (const cookie of cookies) {
      expect(cookie.httpOnly, `${cookie.name} must be HttpOnly`).toBe(true);
      expect(cookie.sameSite, `${cookie.name} must set SameSite`).not.toBe('None');
    }
  });
});
