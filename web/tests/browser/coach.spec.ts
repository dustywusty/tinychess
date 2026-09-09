import { expect, test } from "@playwright/test";

test("coach dismissal persists across reloads and tabs until local data is cleared", async ({ page, context }) => {
  await page.goto("/");
  await page.setViewportSize({ width: 320, height: 740 });
  const close = page.getByRole("button", { name: "Dismiss chess coach" });
  await expect(close).toBeVisible();
  const hitbox = await close.boundingBox();
  expect(hitbox!.width).toBeGreaterThanOrEqual(44);
  expect(hitbox!.height).toBeGreaterThanOrEqual(44);
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(320);
  await page.locator(".coach-card").screenshot({ path: "test-results/coach-dismiss-phone.png" });
  await close.focus();
  await page.keyboard.press("Enter");
  await expect(page.getByText("Meet your chess coach.", { exact: true })).toHaveCount(0);
  await expect(page.locator(".home-coach")).toBeHidden();
  expect(await page.evaluate(() => localStorage.getItem("yourmove.coach-dismissed"))).toBe("1");
  await page.reload();
  await expect(close).toHaveCount(0);
  const nextTab = await context.newPage();
  await nextTab.goto(page.url());
  await expect(nextTab.getByText("Meet your chess coach.", { exact: true })).toHaveCount(0);
  await nextTab.close();
  await page.evaluate(() => localStorage.removeItem("yourmove.coach-dismissed"));
  await page.reload();
  await expect(close).toBeVisible();
});

test("coach still dismisses when storage writes are blocked", async ({ page }) => {
  await page.addInitScript(() => {
    const original = Storage.prototype.setItem;
    Storage.prototype.setItem = function(key, value) {
      if (key === "yourmove.coach-dismissed") throw new Error("Storage blocked");
      return original.call(this, key, value);
    };
  });
  await page.goto("/");
  await page.getByRole("button", { name: "Dismiss chess coach" }).click();
  await expect(page.getByText("Meet your chess coach.", { exact: true })).toHaveCount(0);
});
