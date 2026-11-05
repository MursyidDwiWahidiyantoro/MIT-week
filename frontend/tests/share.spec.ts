import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

/**
 * The three verification outcomes must stay distinguishable. Conflating a
 * rate limit or a server error with "forged report" accuses the user of
 * tampering with their own evidence.
 */
test('a tampered token shows "Laporan tidak sah"', async ({ page }) => {
  await page.goto('/hasil#t=not-a-real-token');
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Laporan tidak sah', {
    timeout: 15_000
  });
});

test('a rate-limited verification never claims the report is invalid', async ({ page }) => {
  await page.route('**/api/v1/verify', (route) =>
    route.fulfill({
      status: 429,
      headers: { 'Retry-After': '12', 'Access-Control-Allow-Origin': '*' },
      body: JSON.stringify({ error: { code: 'RATE_LIMITED', message: 'slow down', retryable: true } })
    })
  );

  await page.goto('/hasil#t=any-token');
  await expect(page.getByRole('heading', { level: 1 })).toHaveText(
    'Laporan sedang tidak bisa diperiksa',
    { timeout: 15_000 }
  );
  await expect(page.getByText('Laporan tidak sah')).toHaveCount(0);
  // The exact countdown depends on the API exposing Retry-After to the browser
  // (see docs/DECISIONS.md); the state and the retry action must hold either way.
  await expect(page.getByRole('button', { name: 'Coba lagi' })).toBeVisible();
});

test('a server error is reported as unavailable, not forged', async ({ page }) => {
  await page.route('**/api/v1/verify', (route) =>
    route.fulfill({
      status: 500,
      headers: { 'Access-Control-Allow-Origin': '*' },
      body: '{}'
    })
  );

  await page.goto('/hasil#t=any-token');
  await expect(page.getByRole('heading', { level: 1 })).toHaveText(
    'Laporan sedang tidak bisa diperiksa',
    { timeout: 15_000 }
  );
});

test('Essential mode changes typography, not content', async ({ page }) => {
  await page.goto('/');
  const sizeOf = () =>
    page.getByRole('heading', { level: 1 }).evaluate((el) => getComputedStyle(el).fontSize);

  await page.evaluate(() => {
    document.documentElement.dataset.mode = 'explore';
  });
  const explore = await sizeOf();

  await page.evaluate(() => {
    document.documentElement.dataset.mode = 'essential';
  });
  const essential = await sizeOf();

  expect(parseFloat(essential)).toBeGreaterThan(parseFloat(explore));
});


/**
 * The README and PROGRESS claimed "axe: 0 violations on all routes" while the
 * dependency was never imported. This is the missing gate: every route must
 * have no serious or critical WCAG 2.1 A/AA violation.
 */
const audit = async (page: import('@playwright/test').Page) => {
  const results = await new AxeBuilder({ page })
    .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
    .analyze();
  const blocking = results.violations.filter(
    (v) => v.impact === 'serious' || v.impact === 'critical'
  );
  return blocking.map((v) => `${v.id} (${v.impact}): ${v.nodes.length} node(s)`);
};

for (const path of ['/', '/deteksi', '/edukasi', '/tentang', '/cek-cepat']) {
  test(`axe: no serious or critical violations on ${path}`, async ({ page }) => {
    await page.goto(path);
    expect(await audit(page), `dark theme: ${path}`).toEqual([]);
    // The theme toggle is a real control; light mode is a second palette and
    // must be gated, not assumed.
    await page.getByRole('button', { name: 'Terang' }).click();
    expect(await audit(page), `light theme: ${path}`).toEqual([]);
  });
}


// /hasil is the most important screen and the only one the route sweep cannot
// reach without a verdict. Load a bundled demo, then audit it.
test('axe: no serious or critical violations on /hasil', async ({ page }) => {
  await page.goto('/deteksi');
  await page.getByRole('button', { name: /Bantuan Rp 500\.000/ }).click();
  await expect(page).toHaveURL(/\/hasil$/, { timeout: 10_000 });
  expect(await audit(page), 'dark theme: /hasil').toEqual([]);
  await page.getByRole('button', { name: 'Terang' }).click();
  expect(await audit(page), 'light theme: /hasil').toEqual([]);
});
