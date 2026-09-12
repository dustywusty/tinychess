import { expect, test } from "@playwright/test";

test("mobile keeps essential status and reactions while desktop retains its copy", async ({ page, browser }) => {
  await page.goto("/");
  await page.getByRole("button", { name: "Play a friend" }).click();
  await expect(page.locator("#turn")).toHaveText(/^(Your turn|Their turn)$/);
  const secondContext = await browser.newContext();
  const second = await secondContext.newPage();
  await second.goto(page.url());
  await expect(second.locator("#turn")).toHaveText(/^(Your turn|Their turn)$/);
  const spectatorContext = await browser.newContext({ viewport: { width: 393, height: 852 } });
  const spectator = await spectatorContext.newPage();
  await spectator.goto(page.url());
  await expect(spectator.locator("#turn")).toHaveText("White to move");
  await expect(spectator.getByText("Watching", { exact: true })).toBeVisible();
  await expect(spectator.getByText("LIVE", { exact: true })).toBeVisible();
  for (const copy of ["A FRIENDLY MATCH", "The seats are full. You can watch and react.", "A LITTLE BACK & FORTH", "Say it with an emoji.", "A few favorites"]) {
    await expect(spectator.getByText(copy, { exact: true })).not.toBeVisible();
  }
  const heading = await spectator.locator(".game-heading").boundingBox();
  expect(heading!.height).toBeLessThan(40);
  await expect(spectator.getByRole("button", { name: "More emoji" })).toBeVisible();
  await expect(spectator.locator(".quick-reactions button")).toHaveCount(6);
  await spectator.screenshot({ path: "test-results/game-compact-spectator.png", fullPage: true });

  await spectator.setViewportSize({ width: 320, height: 640 });
  expect(await spectator.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(320);
  await spectator.route("**/api/games/*/react", (route) => route.fulfill({ json: { ok: false, error: "Could not send reaction." } }));
  await spectator.getByRole("button", { name: "Send 🔥", exact: true }).click();
  await expect(spectator.getByRole("alert")).toHaveText("Could not send reaction.");
  await spectator.getByRole("button", { name: "More emoji" }).click();
  await expect(spectator.getByRole("dialog", { name: "Choose an emoji" })).toBeVisible();
  await spectator.keyboard.press("Escape");
  await spectator.setViewportSize({ width: 1440, height: 1000 });
  await expect(spectator.getByText("Watching", { exact: true })).not.toBeVisible();
  for (const copy of ["A FRIENDLY MATCH", "The seats are full. You can watch and react.", "A LITTLE BACK & FORTH", "Say it with an emoji.", "A few favorites"]) {
    await expect(spectator.getByText(copy, { exact: true })).toBeVisible();
  }
  await expect(spectator.locator("#turn")).toHaveCSS("font-size", "39px");
  await secondContext.close();
  await spectatorContext.close();
});
