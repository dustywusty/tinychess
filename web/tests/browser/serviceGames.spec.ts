import { expect, test } from "@playwright/test";

test("global games load without local history, paginate, refresh and fit on mobile", async ({ page }) => {
  let finished = false;
  await page.route("**/api/games?*", route => {
    const url = new URL(route.request().url());
    const active = url.searchParams.get("status") === "active";
    const next = url.searchParams.get("offset") === "8";
    return route.fulfill({ json: {
      games: [{ id: active ? (next ? "another-game" : "global-game") : "finished-game", result: active && !finished ? "" : "1-0", moveCount: 12, updatedAt: new Date().toISOString() }].filter(() => !active || !finished),
      hasMore: active && !next && !finished,
    } });
  });
  await page.goto("/");
  const section = page.getByRole("region", { name: "Around the boards" });
  const active = section.getByRole("region", { name: "Active games", exact: true });
  const completed = section.getByRole("region", { name: "Completed games", exact: true });
  await expect(active.locator('a[href="/g/global-game"]')).toBeVisible();
  await expect(completed).toContainText("White wins");
  await expect(page.locator(".recent-games")).toContainText("No games yet.");
  await active.getByRole("button", { name: "Next" }).click();
  await expect(active.locator('a[href="/g/another-game"]')).toBeVisible();
  await active.getByRole("button", { name: "Previous" }).click();
  await expect(active.locator('a[href="/g/global-game"]')).toBeVisible();
  await page.setViewportSize({ width: 320, height: 740 });
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(320);
  await section.screenshot({ path: "test-results/service-games-phone.png" });
  finished = true;
  await page.evaluate(() => window.dispatchEvent(new Event("focus")));
  await expect(active).toContainText("No active games yet.");
});

test("global list handles unavailable service and empty results", async ({ page }) => {
  await page.route("**/api/games?*", route => route.fulfill({ status: 503 }));
  await page.goto("/");
  const section = page.getByRole("region", { name: "Around the boards" });
  await expect(section.getByRole("status")).toHaveCount(2);
  await page.route("**/api/games?*", route => route.fulfill({ json: { games: [], hasMore: false } }));
  await page.evaluate(() => window.dispatchEvent(new Event("focus")));
  await expect(section).toContainText("No active games yet.");
  await expect(section).toContainText("No completed games yet.");
  await expect(section.getByRole("status")).toHaveCount(0);
});
