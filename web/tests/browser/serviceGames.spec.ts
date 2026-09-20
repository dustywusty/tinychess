import { expect, test } from "@playwright/test";

test("home shows only aggregate service counts and refreshes them", async ({ page }) => {
  let active = 12;
  await page.route("**/api/stats/games", route => route.fulfill({ json: { active, completed: 42 } }));
  await page.goto("/");
  const section = page.getByRole("region", { name: "Around the boards" });
  await expect(section.locator("dd")).toHaveText(["12", "42"]);
  await expect(section.locator("dt")).toHaveText(["Active games", "Completed games"]);
  await expect(section.getByRole("link")).toHaveCount(0);
  await expect(page.locator(".recent-games")).toContainText("No games yet.");
  await page.setViewportSize({ width: 320, height: 740 });
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(320);
  await section.screenshot({ path: "test-results/service-counts-phone.png" });
  active = 11;
  await page.evaluate(() => window.dispatchEvent(new Event("focus")));
  await expect(section.locator("dd")).toHaveText(["11", "42"]);
});

test("counts recover from unavailable service and show zero totals", async ({ page }) => {
  await page.route("**/api/stats/games", route => route.fulfill({ status: 503 }));
  await page.goto("/");
  const section = page.getByRole("region", { name: "Around the boards" });
  await expect(section.getByRole("status")).toBeVisible();
  await expect(section.locator("dd")).toHaveCount(0);
  await page.route("**/api/stats/games", route => route.fulfill({ json: { active: 0, completed: 0 } }));
  await page.evaluate(() => window.dispatchEvent(new Event("focus")));
  await expect(section.locator("dd")).toHaveText(["0", "0"]);
  await expect(section.getByRole("status")).toHaveCount(0);
});
