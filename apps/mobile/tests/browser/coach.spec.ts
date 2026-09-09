import { expect, test } from "@playwright/test";

test("mobile coach dismissal persists until local data is cleared", async ({ page }) => {
  await page.goto("/");
  await page.setViewportSize({ width: 320, height: 740 });
  const close = page.getByRole("button", { name: "Dismiss chess coach" });
  await expect(close).toBeVisible();
  const hitbox = await close.boundingBox();
  expect(hitbox!.width).toBeGreaterThanOrEqual(44);
  expect(hitbox!.height).toBeGreaterThanOrEqual(44);
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(320);
  await page.screenshot({ path: "test-results/coach-dismiss-phone.png", fullPage: true });
  await close.click();
  await expect(page.getByText("Meet your chess coach.", { exact: true })).toHaveCount(0);
  await expect.poll(() => page.evaluate(() => localStorage.getItem("yourmove.coach-dismissed"))).toBe("1");
  await page.reload();
  // Wait for stored preferences to load before checking that the card is absent.
  await expect(page.getByText("A little chess with your favorite people.")).toBeVisible();
  await expect(close).toHaveCount(0);
  await page.evaluate(() => localStorage.removeItem("yourmove.coach-dismissed"));
  await page.reload();
  await expect(close).toBeVisible();
});
