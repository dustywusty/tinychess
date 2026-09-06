import { expect, test } from "@playwright/test";

test("six palettes fit on phones and new board colors persist in both modes", async ({ page }) => {
  await page.setViewportSize({ width: 320, height: 640 });
  await page.goto("/");
  const open = () => page.getByRole("button", { name: "Appearance", exact: true }).click();
  await open();
  await expect(page.getByRole("radio", { name: / board$/ })).toHaveCount(6);
  for (const [name, rgb] of [["Teal", "rgb(79, 156, 152)"], ["Sky", "rgb(127, 159, 195)"], ["Rose", "rgb(197, 143, 158)"]]) {
    await page.getByRole("radio", { name: `${name} board` }).click();
    await expect(page.getByRole("radio", { name: `${name} board` })).toHaveAttribute("aria-checked", "true");
    await expect(page.locator('[aria-label="b8, black knight"]')).toHaveCSS("background-color", rgb);
  }
  await page.getByRole("radio", { name: "Teal board" }).click();
  const matcha = await page.getByRole("radio", { name: "Matcha board" }).boundingBox();
  const teal = await page.getByRole("radio", { name: "Teal board" }).boundingBox();
  expect(teal!.y).toBeGreaterThan(matcha!.y + matcha!.height);
  expect(teal!.x).toBe(matcha!.x);
  await page.screenshot({ path: "test-results/palettes-light.png" });
  await page.getByRole("radio", { name: "Use dark theme" }).click();
  await page.reload();
  await open();
  await expect(page.getByRole("radio", { name: "Teal board" })).toHaveAttribute("aria-checked", "true");
  await expect(page.getByRole("radio", { name: "Use dark theme" })).toHaveAttribute("aria-checked", "true");
  await expect(page.locator('[aria-label="b8, black knight"]')).toHaveCSS("background-color", "rgb(79, 156, 152)");
  await page.screenshot({ path: "test-results/palettes-dark.png" });
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(320);
  await page.setViewportSize({ width: 640, height: 320 });
  await page.getByRole("radio", { name: "Use light theme" }).click();
  await expect(page.getByText("Good company.")).toHaveCSS("color", "rgb(37, 43, 40)");
});
