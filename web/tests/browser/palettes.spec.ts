import { expect, test } from "@playwright/test";

test("six palettes fit on phones and new board colors persist in both modes", async ({ page }) => {
  await page.setViewportSize({ width: 320, height: 640 });
  await page.goto("/");
  const open = () => page.getByRole("button", { name: "Appearance", exact: true }).click();
  await open();
  const options = page.getByRole("group", { name: "Board color", exact: true }).getByRole("button");
  await expect(options).toHaveCount(6);
  for (const [name, hex] of [["Teal", "#4f9c98"], ["Sky", "#7f9fc3"], ["Rose", "#c58f9e"]]) {
    await page.getByRole("button", { name: `${name} board` }).click();
    await expect(page.getByRole("button", { name: `${name} board` })).toHaveAttribute("aria-pressed", "true");
    expect(await page.evaluate(() => getComputedStyle(document.documentElement).getPropertyValue("--sq2").trim())).toBe(hex);
  }
  await page.getByRole("button", { name: "Teal board" }).click();
  const matcha = await page.getByRole("button", { name: "Matcha board" }).boundingBox();
  const teal = await page.getByRole("button", { name: "Teal board" }).boundingBox();
  expect(teal!.y).toBeGreaterThan(matcha!.y + matcha!.height);
  expect(teal!.x).toBe(matcha!.x);
  await page.screenshot({ path: "test-results/palettes-light.png", fullPage: true });
  await page.getByRole("button", { name: "Use dark theme" }).click();
  await page.reload();
  await open();
  await expect(page.getByRole("button", { name: "Teal board" })).toHaveAttribute("aria-pressed", "true");
  await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");
  await page.screenshot({ path: "test-results/palettes-dark.png", fullPage: true });
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(320);
  await page.setViewportSize({ width: 640, height: 320 });
  await page.getByRole("button", { name: "Use light theme" }).click();
  await expect(page.locator("html")).toHaveAttribute("data-theme", "light");
});
