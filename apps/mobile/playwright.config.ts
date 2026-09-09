import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./tests/browser",
  timeout: 60000,
  workers: 1,
  use: {
    baseURL: process.env.PW_BASE_URL || "http://localhost:8081",
    viewport: { width: 393, height: 852 },
    browserName: "chromium",
    channel: "chrome",
    trace: "retain-on-failure",
  },
  outputDir: "./test-results",
});
