import { test, expect } from '@playwright/test';

/**
 * PRD F1 against the real API (LLM_MOCK=true). The test server cannot reach
 * the internet, which is exactly the "page could not be read" case: the
 * result must say so and offer the screenshot route instead of pretending
 * the content was checked.
 */
test('an unreadable link explains itself and offers a screenshot upload', async ({ page }) => {
  await page.goto('/deteksi');
  await page.getByRole('tab', { name: 'Tautan' }).click();
  await page.getByLabel(/tautan/i).first().fill('https://bri-verifikasi.example/login');
  await page.getByRole('button', { name: 'Mulai periksa' }).click();

  await expect(page).toHaveURL(/\/hasil$/, { timeout: 20_000 });
  const preview = page.locator('[data-info-key="source.metadata"]');
  await expect(preview).toContainText('Isi halaman tautan tidak dapat dibaca');

  // The link pillar still ran and found the lookalike; the verdict is never
  // "belum cukup data" for a link this risky.
  await expect(page.getByText('Domain meniru nama layanan populer').first()).toBeAttached();
  await expect(page.locator('[data-info-key="verdict.score"]')).not.toContainText('Belum cukup data');

  await preview.getByRole('button', { name: 'Unggah tangkapan layar' }).click();
  await expect(page).toHaveURL(/\/deteksi\?tab=file/);
  await expect(page.getByRole('tab', { name: 'Gambar' })).toHaveAttribute('aria-selected', 'true');
});

test('the screenshot demo opens the ELA viewer and OCR tactics', async ({ page }) => {
  const apiCalls: string[] = [];
  page.on('request', (r) => {
    if (r.url().includes('/api/v1/')) apiCalls.push(r.url());
  });

  await page.goto('/deteksi');
  await page.getByRole('button', { name: /Tangkapan layar berita/ }).click();
  await expect(page).toHaveURL(/\/hasil$/, { timeout: 10_000 });

  await expect(page.getByLabel('Geser untuk membandingkan gambar asli dan hasil ELA')).toBeVisible({ timeout: 15_000 });
  await expect(page.getByText('Dibaca dari gambar')).toBeVisible();
  await expect(page.locator('mark.tactic')).toContainText('SEBARKAN SEBELUM DIHAPUS');
  expect(apiCalls).toEqual([]);
});
