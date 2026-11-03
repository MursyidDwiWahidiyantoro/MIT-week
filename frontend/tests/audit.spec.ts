import { test, expect } from '@playwright/test';

/**
 * The one flow that matters: a real audit against the real API must reach
 * /hasil with a verdict. This is the regression guard for the contract bug
 * where the server serialised empty slices as null and the Zod schema threw.
 */
test('a text audit reaches the result page with a verdict', async ({ page }) => {
  const consoleErrors: string[] = [];
  page.on('pageerror', (e) => consoleErrors.push(e.message));

  await page.goto('/deteksi');
  await page.getByRole('tab', { name: 'Teks' }).click();
  await page.getByLabel('Teks pesan').fill(
    'Bantuan Rp 500.000 dibagikan hari ini saja. Sebarkan ke 10 grup!'
  );
  await page.getByRole('button', { name: 'Mulai periksa' }).click();

  await expect(page).toHaveURL(/\/hasil$/, { timeout: 20_000 });
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Hasil pemeriksaan');
  // The proof card also prints the score; scope to the verdict heading.
  await expect(
    page.getByRole('heading', { level: 2 }).getByText(/Skor risiko \d+ dari 100/)
  ).toBeVisible();
  // The misleading network message must never appear on a successful audit.
  await expect(page.getByText('Pemeriksaan terhenti sebelum selesai')).toHaveCount(0);
  expect(consoleErrors).toEqual([]);
});

test('every pillar is accounted for on the result page', async ({ page }) => {
  await page.goto('/deteksi');
  await page.getByRole('tab', { name: 'Teks' }).click();
  await page.getByLabel('Teks pesan').fill('Akun Anda akan diblokir. Verifikasi sekarang.');
  await page.getByRole('button', { name: 'Mulai periksa' }).click();

  await expect(page).toHaveURL(/\/hasil$/, { timeout: 20_000 });

  // A pillar that could not run must say so, never show a 0 score.
  // (The mid-stream radar is covered by a unit test: in mock mode the stream
  // finishes in under a second, so e2e cannot reliably observe a frame.)
  const details = page.locator('details');
  await expect(details).toHaveCount(4);
  await expect(page.getByText('Tidak dapat diperiksa').first()).toBeVisible();
});

test('a bundled demo loads without touching the network', async ({ page }) => {
  const apiCalls: string[] = [];
  page.on('request', (r) => {
    if (r.url().includes('/api/v1/audit')) apiCalls.push(r.url());
  });

  await page.goto('/deteksi');
  await page.getByRole('button', { name: /Bantuan Rp 500\.000/ }).click();

  await expect(page).toHaveURL(/\/hasil$/, { timeout: 10_000 });
  // Scope to the demo banner: the proof card carries the same words as a stamp.
  await expect(page.getByRole('note').filter({ hasText: 'CONTOH / SIMULASI' })).toBeVisible();
  expect(apiCalls).toEqual([]);
});
