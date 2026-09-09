import { test, expect } from "@playwright/test";

test("computer picker starts a recoverable bot game and protects the reserved seat", async ({ page, browser, request }) => {
 await page.goto("/");
 await page.getByRole("button", { name: "Play the computer" }).click();
 await expect(page.getByRole("dialog")).toBeVisible();
 await expect(page.getByRole("radio", { name: /Pip/ })).toBeChecked();
 await page.getByLabel("Black", { exact:true }).check();
 await page.getByRole("button", { name: "Let’s play" }).click();
 await expect(page).toHaveURL(/\/g\//);
 await expect(page.locator(".player-info strong", { hasText: "Pip" })).toBeVisible();
 await expect(page.getByTestId("pgn")).toContainText("1.", {timeout:20000});
 await expect(page.getByTestId("turn")).toContainText("Your turn");
 const id = new URL(page.url()).pathname.split("/").at(-1)!;
 const snapshot = await request.get(`/api/games/${id}/snapshot?clientId=test-spectator`);
 const state = await snapshot.json();
 expect(state.role).toBe("spectator"); expect(state.bot.id).toBe("pip"); expect(state.uci).toHaveLength(1);
 const rejected = await request.post(`/api/games/${id}/move`, {data:{clientId:"test-spectator", uci:"e7e5", botMove:true, expectedPly:1}});
 expect((await rejected.json()).ok).toBe(false);
 // The real human move uses the same command endpoint as friend games.
 const cid = await page.evaluate(() => localStorage.getItem("tinychess:clientId"));
 const moved = await request.post(`/api/games/${id}/move`, { data:{clientId:cid, uci:"e7e5"} });
 expect((await moved.json()).ok).toBe(true);
 await expect.poll(async () => (await (await request.get(`/api/games/${id}/snapshot?clientId=test-spectator`)).json()).uci.length, {timeout:20000}).toBe(3);
 await page.reload();
 await expect(page.getByTestId("turn")).toContainText("Your turn");
 await expect(page.locator(".player-info strong", {hasText:"Pip"})).toBeVisible();
 const spectator = await browser.newPage();
 await spectator.goto(page.url());
 await expect(spectator.getByText("Watching", {exact:true})).toBeVisible();
 await spectator.close();
 await page.getByRole("link", { name: "Your Move home" }).click();
 await expect(page.locator(".recent-games").getByText("PvBot", { exact:true })).toBeVisible();
 await expect(page.locator(".recent-games").getByText("A match with Pip")).toBeVisible();
});
