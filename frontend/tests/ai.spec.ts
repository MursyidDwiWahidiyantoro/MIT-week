import { test, expect } from '@playwright/test';
import { fileURLToPath } from 'node:url';

/**
 * Jejak AI end to end. The bundled demo picture carries a Content
 * Credentials (C2PA) segment declaring AI generation: the demo shows the
 * stored verdict, and uploading the very same file to the real API (mock
 * model) must reach the same verdict from the file's own bytes.
 */
const LEAD = 'Ada tanda jelas gambar ini dibuat dengan AI. Jangan anggap sebagai foto asli.';
const C2PA_LINE = 'Content Credentials (C2PA) di file menyatakan gambar ini dibuat dengan AI.';

test('the AI picture demo explains what the file declares', async ({ page }) => {
  await page.goto('/deteksi');
  await page.getByRole('button', { name: /ikan raksasa/ }).click();
  await expect(page).toHaveURL(/\/hasil$/, { timeout: 10_000 });
  await expect(page.locator('[data-info-key="verdict.sentence"]')).toContainText(LEAD);
  await expect(page.locator('[data-info-key="pillar.ai"]')).toContainText(C2PA_LINE);
});

test('uploading the AI picture reaches the same verdict through the live pipeline', async ({ page }) => {
  await page.goto('/deteksi');
  await page.getByRole('tab', { name: 'Gambar' }).click();
  await page.setInputFiles('#amba-file', fileURLToPath(new URL('../static/demo/gambar-ai.jpg', import.meta.url)));
  await page.getByRole('button', { name: 'Mulai periksa' }).click();

  await expect(page).toHaveURL(/\/hasil$/, { timeout: 25_000 });
  await expect(page.locator('[data-info-key="verdict.sentence"]')).toContainText(LEAD);
  await expect(page.locator('[data-info-key="pillar.ai"]')).toContainText(C2PA_LINE);
  await expect(page.locator('[data-info-key="verdict.score"]')).toContainText('85');
});

test('an ordinary upload says missing origin data proves nothing', async ({ page }) => {
  await page.goto('/deteksi');
  await page.getByRole('tab', { name: 'Gambar' }).click();
  await page.setInputFiles('#amba-file', fileURLToPath(new URL('./fixtures/sample.png', import.meta.url)));
  await page.getByRole('button', { name: 'Mulai periksa' }).click();

  await expect(page).toHaveURL(/\/hasil$/, { timeout: 25_000 });
  await expect(page.locator('[data-info-key="pillar.ai"]')).toContainText('ketiadaannya bukan bukti apa pun');
});
