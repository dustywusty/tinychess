import { mkdir } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import { chromium, expect } from "@playwright/test";

const baseURL = process.env.PW_BASE_URL || "http://localhost:5173";
const output = fileURLToPath(new URL("../../docs/screenshots/", import.meta.url));
const desktop = { width: 1280, height: 940 };
const phone = { width: 393, height: 852 };
const browser = await chromium.launch({ channel: "chrome" });
const errors = [];

async function newPage(viewport = desktop) {
  const context = await browser.newContext({ viewport, deviceScaleFactor: 1, colorScheme: "light", reducedMotion: "reduce" });
  const page = await context.newPage();
  page.on("pageerror", (error) => errors.push(error.message));
  page.setDefaultTimeout(20000);
  return page;
}

async function capture(page, name, fullPage = true) {
  await page.evaluate(() => document.fonts.ready);
  await expect(page.locator(".big-emoji")).toHaveCount(0);
  await page.evaluate(() => window.scrollTo(0, 0));
  await page.screenshot({ path: output + name + ".png", fullPage, animations: "disabled" });
  console.log(`Captured ${name}.png`);
}

async function friendGame(page) {
  await page.goto(baseURL);
  await page.getByRole("button", { name: "Play a friend" }).click();
  await expect(page.locator("#turn")).toHaveText(/^(Your turn|Their turn)$/);
  const other = await newPage();
  await other.goto(page.url());
  await expect(other.locator("#turn")).toHaveText(/^(Your turn|Their turn)$/);
  return (await page.locator("#turn").innerText()) === "Your turn" ? [page, other] : [other, page];
}

async function move(page, uci) {
  await expect(page.locator("#turn")).toHaveText("Your turn");
  await page.locator(`[data-square="${uci.slice(0, 2)}"]`).click();
  const target = page.locator(`[data-square="${uci.slice(2, 4)}"]`);
  await expect(target).toHaveAttribute("aria-label", /legal move/);
  await target.click();
  await expect(page.locator("#turn")).not.toHaveText("Your turn");
}

async function appearance(page, color, mode = "light") {
  await page.getByRole("button", { name: "Appearance", exact: true }).click();
  await page.getByRole("button", { name: `${color} board` }).click();
  await page.getByRole("button", { name: `Use ${mode} theme` }).click();
  await page.getByRole("button", { name: "Close appearance" }).click();
}

try {
  await mkdir(output, { recursive: true });
  const owner = await newPage();
  const [white, black] = await friendGame(owner);
  const opening = ["e2e4", "e7e5", "g1f3", "b8c6", "f1c4", "g8f6", "d2d4", "e5d4", "f3d4", "f8c5"];
  for (const [index, uci] of opening.entries()) await move(index % 2 ? black : white, uci);
  await expect(white.getByTestId("captured-white")).toBeVisible();
  await expect(white.getByTestId("captured-black")).toBeVisible();
  await white.getByRole("button", { name: "Send 🔥", exact: true }).click();
  await black.getByRole("button", { name: "Send 👏", exact: true }).click();
  await expect(white.locator(".reaction-history")).toContainText("👏Them");
  await expect(white.getByRole("button", { name: "More emoji", exact: true })).toBeEnabled({ timeout: 10000 });
  await white.locator('[data-square="b1"]').click();
  await expect(white.locator('[data-square="c3"]')).toHaveAttribute("aria-label", /legal move/);
  await capture(white, "friend-game");

  const spectator = await newPage(phone);
  await spectator.goto(white.url());
  await expect(spectator.getByText("Watching", { exact: true })).toBeVisible();
  await appearance(spectator, "Peach", "dark");
  await capture(spectator, "spectator-dark");

  const [winner, loser] = await friendGame(owner);
  const mate = ["e2e4", "e7e5", "f1c4", "b8c6", "d1h5", "g8f6", "h5f7"];
  for (const [index, uci] of mate.entries()) await move(index % 2 ? loser : winner, uci);
  await expect(winner.locator("#status")).toContainText("Checkmate");
  await expect(winner.locator("#pgn")).toContainText("Qxf7#");
  await appearance(winner, "Lilac");
  await capture(winner, "checkmate");

  await owner.goto(baseURL);
  await owner.setViewportSize(phone);
  await appearance(owner, "Matcha");
  await owner.getByRole("button", { name: "Play the computer" }).click();
  await expect(owner.getByRole("dialog")).toBeVisible();
  await owner.getByRole("radio", { name: /Ada/ }).check();
  await capture(owner, "computer-picker", false);
  await owner.getByRole("button", { name: "Let’s play" }).click();
  await expect(owner.locator("#turn")).toHaveText("Your turn");
  await move(owner, "e2e4");
  await expect(owner.locator("#turn")).toHaveText("Your turn", { timeout: 30000 });
  await move(owner, "g1f3");
  await expect(owner.locator("#turn")).toHaveText("Your turn", { timeout: 30000 });
  await expect(owner.locator(".player-info strong", { hasText: "Ada" })).toBeVisible();
  await appearance(owner, "Sky");
  await capture(owner, "computer-game");

  await owner.goto(baseURL);
  await owner.setViewportSize(desktop);
  await appearance(owner, "Matcha");
  await expect(owner.getByTestId("recent-result")).toBeVisible();
  await expect(owner.getByText("A match with Ada", { exact: true })).toBeVisible();
  await capture(owner, "home");
  expect(errors).toEqual([]);
} finally {
  await browser.close();
}
