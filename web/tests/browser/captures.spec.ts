import { expect, test } from "@playwright/test";

test("captures update live, survive reloads, and follow players when flipped", async ({ page, browser }) => {
  await page.goto("/");
  await page.getByRole("button", { name: "Play a friend" }).click();
  const turn = (client: typeof page) => client.getByText(/^(Your turn|Their turn|Your move\.|Over to them\.)$/);
  await expect(turn(page)).toBeVisible();
  const otherContext = await browser.newContext();
  const spectatorContext = await browser.newContext();
  try {
    const other = await otherContext.newPage();
    await other.goto(page.url());
    await expect(turn(other)).toBeVisible();
    const white = /^(Your turn|Your move\.)$/.test(await turn(page).innerText()) ? page : other;
    const black = white === page ? other : page;
    const spectator = await spectatorContext.newPage();
    await spectator.goto(page.url());
    await expect(spectator.getByText("Watching", { exact: true })).toBeVisible();
    await expect(page.getByTestId("captured-white")).toHaveCount(0);
    await expect(page.getByTestId("captured-black")).toHaveCount(0);
    const move = async (client: typeof page, from: string, to: string) => {
      await expect(turn(client)).toHaveText(/^(Your turn|Your move\.)$/);
      await client.getByRole("button", { name: new RegExp("^" + from + ",") }).click();
      await client.getByRole("button", { name: new RegExp("^" + to + ",.*legal move") }).click();
      await expect(turn(client)).toHaveText(/^(Their turn|Over to them\.)$/);
    };
    await move(white, "d2", "d4");
    await move(black, "e7", "e5");
    await move(white, "g1", "f3");
    await move(black, "e5", "d4");
    await expect(white.getByTestId("captured-black")).toHaveAttribute("aria-label", "Captured by Black: 1 white pawn");
    await expect(white.getByTestId("captured-white")).toHaveCount(0);
    await move(white, "f3", "d4");
    for (const client of [white, black, spectator]) {
      await expect(client.getByTestId("captured-white")).toHaveAttribute("aria-label", "Captured by White: 1 black pawn");
      await expect(client.getByTestId("captured-black")).toHaveAttribute("aria-label", "Captured by Black: 1 white pawn");
      await expect(client.getByTestId("captured-white").locator("svg")).toHaveCount(1);
      await expect(client.getByTestId("captured-black").locator("svg")).toHaveCount(1);
    }
    await white.reload();
    await expect(white.getByTestId("captured-white")).toBeVisible();
    const y = async (side: string) => (await white.getByTestId("captured-" + side).boundingBox())!.y;
    expect(await y("white")).toBeGreaterThan(await y("black"));
    await white.getByRole("button", { name: /Flip board/ }).click();
    expect(await y("white")).toBeLessThan(await y("black"));
    await white.screenshot({ path: "test-results/captures-phone.png", fullPage: true });
    await white.getByRole("button", { name: "Appearance", exact: true }).click();
    await white.getByRole("button", { name: "Use dark theme" }).click();
    await white.getByRole("button", { name: "Close appearance" }).click();
    await white.setViewportSize({ width: 320, height: 740 });
    expect(await white.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(320);
    await expect(white.getByTestId("captured-white")).toBeVisible();
    await expect(white.getByTestId("captured-black")).toBeVisible();
    await white.screenshot({ path: "test-results/captures-dark-small.png", fullPage: true });
    await white.setViewportSize({ width: 1440, height: 1000 });
    await white.screenshot({ path: "test-results/captures-wide.png", fullPage: true });
  } finally {
    await otherContext.close();
    await spectatorContext.close();
  }
});
