import { expect, test } from "@playwright/test";

test("mobile home keeps a lonely empty state on a fresh visit", async ({ page }) => {
  await page.goto("/");
  const recent = page.getByTestId("recent-games");
  await expect(recent.getByText("No games yet.")).toBeVisible();
  await expect(recent.getByText("A lonely knight, waiting for your first move.")).toBeVisible();
  await expect(recent.getByRole("button")).toHaveCount(0);
  await page.setViewportSize({ width: 320, height: 740 });
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(320);
  await recent.screenshot({ path: "test-results/recent-empty-phone.png" });
});

test("mobile recent games distinguish friend, bot and legacy records", async ({ page }) => {
  await page.addInitScript(() => {
    const entry = { fen: "", status: "White to move", moves: 2, updatedAt: 1 };
    localStorage.setItem("yourmove.recent-games", JSON.stringify([
      { ...entry, id: "friend", opponentType: "human" },
      { ...entry, id: "computer", opponentType: "bot", botId: "pip" },
      { ...entry, id: "legacy" },
    ]));
  });
  await page.goto("/");
  const recent = page.getByTestId("recent-games");
  await expect(recent.getByText("PvP", { exact: true })).toBeVisible();
  await expect(recent.getByText("PvBot", { exact: true })).toBeVisible();
  await expect(recent.getByText("A match with Pip")).toBeVisible();
  await expect(recent.getByText("Saved game", { exact: true })).toBeVisible();
  await expect(recent.getByText("No games yet.")).toHaveCount(0);
  await page.setViewportSize({ width: 320, height: 740 });
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(320);
  await recent.screenshot({ path: "test-results/recent-games-phone.png" });
});

test("mobile finished cards show personal results and highlighted full-width banners", async ({ page }) => {
  await page.addInitScript(() => {
    const entry = { fen: "", status: "", moves: 4, updatedAt: 1, opponentType: "bot", botId: "ada", role: "player", playerColor: "b" };
    localStorage.setItem("yourmove.recent-games", JSON.stringify([
      { ...entry, id: "win", result: "0-1" },
      { ...entry, id: "loss", result: "1-0" },
      { ...entry, id: "draw", result: "1/2-1/2" },
    ]));
  });
  await page.goto("/");
  const recent = page.getByTestId("recent-games");
  for (const label of ["Win", "Loss", "Draw"]) {
    await expect(recent.getByTestId("recent-result").getByText(label, { exact: true })).toBeVisible();
  }
  const card = await recent.getByRole("button").first().boundingBox();
  const banner = await recent.getByTestId("recent-result").first().boundingBox();
  expect(banner!.width).toBe(card!.width - 2);
  await page.setViewportSize({ width: 320, height: 900 });
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(320);
  await recent.screenshot({ path: "test-results/recent-results-phone.png" });
  await page.getByRole("button", { name: "Appearance", exact: true }).click();
  await page.getByRole("radio", { name: "Use dark theme" }).click();
  await page.getByRole("button", { name: "Close appearance" }).click();
  await recent.screenshot({ path: "test-results/recent-results-dark.png" });
});
