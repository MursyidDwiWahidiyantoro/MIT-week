import { test, expect, type Page } from '@playwright/test';

/**
 * PRD F14 / TECH_SPEC §8.2: Mode Jelajah and Mode Esensial are two
 * presentations of the same information. These tests render one Dossier in
 * both modes and compare every data-info-key, switch modes mid-flow, and
 * check calm motion and read-aloud.
 */

async function infoMap(page: Page): Promise<Record<string, string>> {
  return page.evaluate(() => {
    const out: Record<string, string> = {};
    for (const el of document.querySelectorAll<HTMLElement>('[data-info-key]')) {
      const key = el.dataset.infoKey!;
      out[key] = (el.innerText || el.textContent || '').replace(/\s+/g, ' ').trim();
    }
    return out;
  });
}

async function openDemo(page: Page, name: RegExp) {
  await page.goto('/deteksi');
  await page.getByRole('button', { name }).click();
  await expect(page).toHaveURL(/\/hasil$/, { timeout: 10_000 });
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Hasil pemeriksaan');
}

test('both modes show identical information keys and text', async ({ page }) => {
  await openDemo(page, /Bantuan Rp 500\.000/);

  await page.getByRole('button', { name: 'Jelajah', exact: true }).click();
  await expect(page.locator('html')).toHaveAttribute('data-mode', 'explore');
  const explore = await infoMap(page);

  await page.getByRole('button', { name: 'Esensial', exact: true }).click();
  await expect(page.locator('html')).toHaveAttribute('data-mode', 'essential');
  const essential = await infoMap(page);

  expect(Object.keys(explore).length).toBeGreaterThanOrEqual(8);
  expect(Object.keys(essential).sort()).toEqual(Object.keys(explore).sort());
  expect(essential).toEqual(explore);
  // Switching mode never resets the result.
  await expect(page).toHaveURL(/\/hasil$/);
});

test('switching mode mid-flow keeps the typed input', async ({ page }) => {
  await page.goto('/deteksi');
  const field = page.getByLabel('Tautan', { exact: true });
  await field.fill('https://contoh.example/belum-diperiksa');
  await page.getByRole('button', { name: 'Esensial', exact: true }).click();
  await expect(field).toHaveValue('https://contoh.example/belum-diperiksa');
  await page.getByRole('button', { name: 'Jelajah', exact: true }).click();
  await expect(field).toHaveValue('https://contoh.example/belum-diperiksa');
});

test('reduced motion is calm in both modes', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'reduce' });
  await page.goto('/');
  for (const name of ['Jelajah', 'Esensial']) {
    await page.getByRole('button', { name, exact: true }).click();
    await expect(page.locator('html')).toHaveAttribute('data-motion', 'calm');
  }
});

test('Essential mode uses base text of at least 18px', async ({ page }) => {
  await openDemo(page, /Bantuan Rp 500\.000/);
  await page.getByRole('button', { name: 'Esensial', exact: true }).click();
  const size = await page
    .locator('[data-info-key="verdict.explanation"]')
    .evaluate((el) => parseFloat(getComputedStyle(el).fontSize));
  expect(size).toBeGreaterThanOrEqual(18);
});

/** Replaces speechSynthesis with a recorder so the test can hear it. */
async function mockSpeech(page: Page) {
  await page.addInitScript(() => {
    const spoken: string[] = [];
    (window as unknown as { __spoken: string[] }).__spoken = spoken;
    class FakeUtterance {
      text: string;
      lang = '';
      rate = 1;
      constructor(text: string) {
        this.text = text;
      }
    }
    Object.defineProperty(window, 'SpeechSynthesisUtterance', { value: FakeUtterance, configurable: true });
    Object.defineProperty(window, 'speechSynthesis', {
      configurable: true,
      value: {
        speak: (u: FakeUtterance) => spoken.push(`${u.lang}|${u.text}`),
        cancel: () => {},
        getVoices: () => []
      }
    });
  });
}

test('read-aloud speaks the lead sentence first, in Indonesian', async ({ page }) => {
  await mockSpeech(page);
  await openDemo(page, /Bantuan Rp 500\.000/);
  const sentence = await page.locator('[data-info-key="verdict.sentence"]').innerText();

  await page.getByRole('button', { name: 'Bacakan hasil' }).click();
  const spoken = await page.evaluate(() => (window as unknown as { __spoken: string[] }).__spoken);
  expect(spoken).toHaveLength(1);
  expect(spoken[0].startsWith(`id-ID|${sentence}`)).toBe(true);
});

test('auto-read is off by default and speaks the next result once enabled', async ({ page }) => {
  await mockSpeech(page);
  await openDemo(page, /Bantuan Rp 500\.000/);
  expect(await page.evaluate(() => (window as unknown as { __spoken: string[] }).__spoken)).toEqual([]);

  await page.getByLabel('Bacakan otomatis hasil berikutnya').check();
  await openDemo(page, /situs resmi/);
  await expect
    .poll(() => page.evaluate(() => (window as unknown as { __spoken: string[] }).__spoken.length))
    .toBe(1);
});

for (const name of ['Jelajah', 'Esensial']) {
  test(`a link check takes at most 3 taps in ${name}`, async ({ page }) => {
    await page.goto('/');
    await page.getByRole('button', { name, exact: true }).click(); // choosing a mode is setup, not a tap of the flow
    await page.getByLabel('Tempel tautan yang ingin diperiksa').fill('https://bri-verifikasi.example/login');
    let taps = 0;
    await page.getByRole('button', { name: 'Periksa', exact: true }).click();
    taps++;
    await expect(page).toHaveURL(/\/hasil$/, { timeout: 20_000 });
    expect(taps).toBeLessThanOrEqual(3);
    await expect(page.getByRole('heading', { level: 1 })).toHaveText('Hasil pemeriksaan');
  });
}
