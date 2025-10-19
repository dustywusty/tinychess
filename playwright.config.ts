import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  fullyParallel: false,
  retries: process.env.CI ? 1 : 0,
  timeout: 90 * 1000,
  use: {
    baseURL: 'http://127.0.0.1:8080',
    headless: true,
    trace: 'on-first-retry',
  },
  reporter: [
    ['list'],
    ['html', { outputFolder: 'playwright-report', open: 'never' }],
  ],
  webServer: {
    command: 'go run .',
    url: 'http://127.0.0.1:8080',
    reuseExistingServer: !process.env.CI,
    stdout: 'pipe',
    stderr: 'pipe',
    timeout: 120 * 1000,
  },
});
