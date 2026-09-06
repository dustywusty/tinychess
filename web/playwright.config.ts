import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./tests/browser",
  timeout: 60000,
  workers: 1,
  use: {
    baseURL: "http://localhost:5173",
    viewport: { width: 393, height: 852 },
    browserName: "chromium",
    channel: "chrome",
    trace: "retain-on-failure",
  },
  outputDir: "./test-results",
});
