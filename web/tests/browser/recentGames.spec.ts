import { expect, test } from "@playwright/test";

test("a first visit shows the lonely recent-games card at phone and desktop sizes", async ({ page }) => {
  await page.goto("/");
  const recent = page.locator(".recent-games");
  await expect(recent.getByText("No games yet.")).toBeVisible();
  await expect(recent.getByText("A lonely knight, waiting for your first move.")).toBeVisible();
  await expect(recent.getByRole("link")).toHaveCount(0);
  await page.setViewportSize({ width: 320, height: 740 });
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(320);
  await recent.screenshot({ path: "test-results/recent-empty-phone.png" });
  await page.setViewportSize({ width: 1440, height: 1000 });
  await expect(recent).toBeVisible();
  await page.screenshot({ path: "test-results/recent-empty-desktop.png", fullPage: true });
});

test("recent games distinguish friend, bot and legacy games, then return to empty", async ({ page }) => {
  await page.addInitScript(() => {
    const entry = { createdAt: 1, lastSeen: 1, lastSeenLocal: 1, status: "White to move", result: "" };
    localStorage.setItem("tinychess:games:v1", JSON.stringify({
      friend: { ...entry, id: "friend", opponentType: "human" },
      computer: { ...entry, id: "computer", opponentType: "bot", botId: "pip" },
      legacy: { ...entry, id: "legacy" },
    }));
  });
  await page.goto("/");
  const recent = page.locator(".recent-games");
  await expect(recent.getByText("No games yet.")).toHaveCount(0);
  await expect(recent.locator('a[href="/g/friend"]')).toContainText("PvP");
  await expect(recent.locator('a[href="/g/computer"]')).toContainText("PvBotA match with Pip");
  await expect(recent.locator('a[href="/g/legacy"]')).toContainText("Saved game");
  await page.setViewportSize({ width: 320, height: 740 });
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(320);
  await recent.screenshot({ path: "test-results/recent-games-phone.png" });
  for (const id of ["friend", "computer", "legacy"]) {
    await recent.getByRole("button", { name: `Forget game ${id}` }).click();
  }
  await expect(recent.getByText("No games yet.")).toBeVisible();
});

test("finished cards have edge-to-edge win, loss, draw and spectator banners", async ({ page }) => {
  await page.addInitScript(() => {
    const entry = { createdAt: 1, lastSeen: 1, lastSeenLocal: 1, status: "", opponentType: "bot", botId: "ada" };
    localStorage.setItem("tinychess:games:v1", JSON.stringify({
      win: { ...entry, id: "win", result: "1-0", role: "player", playerColor: "w" },
      loss: { ...entry, id: "loss", result: "0-1", role: "player", playerColor: "w" },
      draw: { ...entry, id: "draw", result: "1/2-1/2", role: "player", playerColor: "b" },
      watched: { ...entry, id: "watched", result: "0-1", role: "spectator" },
      active: { ...entry, id: "active", result: "" },
    }));
  });
  await page.goto("/");
  const recent = page.locator(".recent-games");
  for (const label of ["Win", "Loss", "Draw", "Black wins"]) {
    await expect(recent.getByTestId("recent-result").getByText(label, { exact: true })).toBeVisible();
  }
  await expect(recent.getByTestId("recent-result")).toHaveCount(4);
  const card = await recent.locator("li").first().boundingBox();
  const banner = await recent.getByTestId("recent-result").first().boundingBox();
  expect(banner!.width).toBe(card!.width);
  expect(banner!.x).toBe(card!.x);
  await page.setViewportSize({ width: 320, height: 900 });
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(320);
  await recent.screenshot({ path: "test-results/recent-results-phone.png" });
  await page.setViewportSize({ width: 1440, height: 1100 });
  await recent.screenshot({ path: "test-results/recent-results-desktop.png" });
  await page.getByRole("button", { name: "Appearance", exact: true }).click();
  await page.getByRole("button", { name: "Use dark theme" }).click();
  await page.getByRole("button", { name: "Close appearance" }).click();
  await recent.screenshot({ path: "test-results/recent-results-dark.png" });
});
