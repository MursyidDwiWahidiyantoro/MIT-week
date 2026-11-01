import { defineConfig, devices } from '@playwright/test';

/**
 * Two servers, both real:
 *  - the Go API with LLM_MOCK=true, because the audit spec exists to catch
 *    contract drift between the server's JSON and the Zod schema. Fixtures
 *    would let that bug pass.
 *  - the built SPA via `vite preview`, so the tests exercise the production
 *    bundle and its CSP rather than the dev server.
 */
export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: process.env.CI ? [['list'], ['html', { open: 'never' }]] : 'list',
  use: {
    baseURL: 'http://localhost:4173',
    trace: 'on-first-retry'
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: [
    {
      command: 'go run ./cmd/server',
      cwd: '../backend',
      url: 'http://localhost:8080/healthz',
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      env: { SERVER_PORT: '8080', LLM_MOCK: 'true', ENVIRONMENT: 'development' }
    },
    {
      command: 'npm run preview -- --port 4173 --strictPort',
      url: 'http://localhost:4173',
      reuseExistingServer: !process.env.CI,
      timeout: 120_000
    }
  ]
});
