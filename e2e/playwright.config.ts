import { defineConfig, devices } from '@playwright/test'
import { env } from './support/env'

/**
 * Base E2E config.
 *
 * Principles:
 * - BASE_URL / API_URL come from env (see .env.example), so they work the same
 *   locally (Vite dev + uvicorn), in docker-compose, and in CI.
 * - Time zone is Europe/Moscow, same as the whole backend (see backend/docs/architecture/time.md).
 * - Traces/screenshots/video only on-failure, to avoid bloating artifacts.
 * - webServer is not started from Playwright yet: we assume the frontend and
 *   backend are already running (`pnpm dev` in frontend + backend locally or via
 *   docker-compose). When infra arrives, we will add a webServer block here.
 */
export default defineConfig({
  testDir: './tests',
  outputDir: './test-results',
  fullyParallel: true,
  forbidOnly: env.isCI,
  retries: env.isCI ? 2 : 0,
  workers: env.isCI ? 2 : undefined,
  timeout: 30_000,
  expect: {
    timeout: 5_000,
  },
  reporter: env.isCI
    ? [['github'], ['html', { open: 'never' }], ['list']]
    : [['list'], ['html', { open: 'never' }]],
  use: {
    baseURL: env.baseUrl,
    timezoneId: 'Europe/Moscow',
    locale: 'ru-RU',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    actionTimeout: 10_000,
    navigationTimeout: 15_000,
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
})
