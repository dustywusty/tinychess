import { test, expect } from "@playwright/test";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

test("bundled mobile engine runs without network access and accepts stop", async ({ page }) => {
 await page.route("**/*", route => route.abort());
 await page.addInitScript(() => {
  Object.assign(window, { engineLines: [], ReactNativeWebView: { postMessage: (line: string) => (window as any).engineLines.push(line) } });
 });
 await page.goto("about:blank");
 await page.setContent(readFileSync(resolve("assets/engine/arasan.html"), "utf8"));
 await expect.poll(() => page.evaluate(() => !!(window as any).arasan), { timeout:15000 }).toBe(true);
 await page.evaluate(() => (window as any).arasan.commands.push("uci", "isready"));
 await expect.poll(() => page.evaluate(() => (window as any).engineLines.includes("readyok"))).toBe(true);
 await page.evaluate(() => (window as any).arasan.commands.push("setoption name OwnBook value false", "setoption name MultiPV value 3", "position startpos", "go infinite"));
 await expect.poll(() => page.evaluate(() => (window as any).engineLines.some((line: string) => line.startsWith("info depth")))).toBe(true);
 await page.evaluate(() => (window as any).arasan.commands.push("stop"));
 await expect.poll(() => page.evaluate(() => (window as any).engineLines.some((line: string) => line.startsWith("bestmove ")))).toBe(true);
});

test("mobile opponent picker and controller play both sides and recover on reload", async ({ page, request }) => {
 const errors: string[] = [];
 page.on("pageerror", error => errors.push(error.message));
 await page.goto("/");
 await page.getByRole("button", {name:"Play the computer"}).click();
 await expect(page.getByRole("radio", {name:"Pip, Just learning"})).toHaveAttribute("aria-checked", "true");
 await page.getByRole("radio", {name:/Black/}).click();
 await page.getByRole("button", {name:"Let’s play"}).click();
 await expect(page).toHaveURL(/\/g\//);
 await expect(page.getByText("🐣 Pip", {exact:true})).toBeVisible();
 await expect(page.getByTestId("game-status")).toContainText("Your move.", {timeout:25000});
 await page.getByRole("button", {name:/^e7, black pawn/}).click();
 await page.getByRole("button", {name:/^e5,.*legal move/}).click();
 await expect(page.getByTestId("game-status")).toContainText("Your move.", {timeout:25000});
 const id = new URL(page.url()).pathname.split("/").at(-1)!;
 await expect.poll(async () => (await (await request.get(`/api/games/${id}/snapshot?clientId=visitor`)).json()).uci.length).toBe(3);
 await page.reload();
 await expect(page.getByTestId("game-status")).toContainText("Your move.");
 await page.getByRole("button", { name: "Back to home" }).click();
 await expect(page.getByTestId("recent-games").getByText("PvBot", { exact:true })).toBeVisible();
 await expect(page.getByTestId("recent-games").getByText("A match with Pip")).toBeVisible();
 expect(errors).toEqual([]);
});

test("mobile bot battle uses both selections and resumes with its saved delay", async ({ page, request }) => {
 const errors: string[] = [];
 page.on("pageerror", error => errors.push(error.message));
 await page.goto("/");
 await page.getByRole("button", { name: "Play the computer" }).click();
 await page.getByRole("button", { name: /Advanced settings/ }).click();
 await page.getByRole("checkbox", { name: /Bot vs. bot/ }).click();
 await page.getByRole("radio", { name: "White bot: Max", exact: true }).click();
 await page.getByRole("radio", { name: "Black bot: Ada", exact: true }).click();
 await page.getByRole("slider", { name: "Time between moves" }).fill("1000");
 await page.getByRole("button", { name: "Start bot battle" }).click();
 await expect(page).toHaveURL(/\/g\//);
 await expect(page.getByText("😎 Max", { exact: true })).toBeVisible();
 await expect(page.getByText("🧠 Ada", { exact: true })).toBeVisible();
 await expect(page.getByText("Watching", { exact: true })).toBeVisible();
 const id = new URL(page.url()).pathname.split("/").at(-1)!;
 const snapshot = async () => (await (await request.get(`/api/games/${id}/snapshot?clientId=mobile-battle-spectator`)).json());
 await expect.poll(async () => (await snapshot()).uci.length, { timeout: 30000 }).toBeGreaterThanOrEqual(4);
 const before = await snapshot();
 expect(before.bot).toMatchObject({ id: "ada", playerBotId: "max", moveDelayMs: 1000 });
 await page.reload();
 await expect.poll(async () => (await snapshot()).uci.length, { timeout: 20000 }).toBeGreaterThan(before.uci.length);
 expect(errors).toEqual([]);
});
