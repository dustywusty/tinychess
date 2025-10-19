import { test, expect, Page } from '@playwright/test';

test.describe('TinyChess headless gameplay', () => {
  test('two players can complete opening moves with the mouse', async ({ browser }, testInfo) => {
    const contextA = await browser.newContext();
    const hostPage = await contextA.newPage();

    await hostPage.goto('/');
    await hostPage.waitForLoadState('networkidle');
    await hostPage.getByRole('link', { name: 'New game' }).first().click();
    await hostPage.waitForURL(/\/[^/]+$/);
    const gameUrl = hostPage.url();

    const contextB = await browser.newContext();
    const guestPage = await contextB.newPage();
    await guestPage.goto(gameUrl);

    await ensurePlayerSeat(hostPage);
    await ensurePlayerSeat(guestPage);

    const { whitePage, blackPage } = await identifyPlayers(hostPage, guestPage);

    await waitForYourTurn(whitePage);
    await performMove(whitePage, 'e2', 'e4');

    await waitForYourTurn(blackPage);
    await performMove(blackPage, 'e7', 'e5');

    await expect(whitePage.locator('#pgn')).toContainText('e4');
    await expect(whitePage.locator('#pgn')).toContainText('e5');

    const screenshotPath = testInfo.outputPath('final-board.png');
    await whitePage.screenshot({ path: screenshotPath, fullPage: true });
    await testInfo.attach('final-board', { path: screenshotPath, contentType: 'image/png' });

    await contextB.close();
    await contextA.close();
  });
});

async function ensurePlayerSeat(page: Page): Promise<void> {
  await expect(page.locator('#release')).toBeVisible({ timeout: 20_000 });
}

async function identifyPlayers(hostPage: Page, guestPage: Page): Promise<{ whitePage: Page; blackPage: Page }> {
  const [hostTurn, guestTurn] = await Promise.all([
    readTurnLabel(hostPage),
    readTurnLabel(guestPage),
  ]);

  if (/your turn/i.test(hostTurn) && /their turn/i.test(guestTurn)) {
    return { whitePage: hostPage, blackPage: guestPage };
  }
  if (/your turn/i.test(guestTurn) && /their turn/i.test(hostTurn)) {
    return { whitePage: guestPage, blackPage: hostPage };
  }

  throw new Error(`Unexpected turn indicators — host: "${hostTurn}" guest: "${guestTurn}"`);
}

async function readTurnLabel(page: Page): Promise<string> {
  const handle = await page.waitForFunction(() => {
    const el = document.getElementById('turn');
    const text = (el?.textContent || '').trim();
    return text || null;
  }, undefined, { timeout: 20_000 });

  const text = await handle.jsonValue<string | null>();
  if (!text) {
    throw new Error('Turn label not available');
  }
  return text;
}

async function waitForYourTurn(page: Page): Promise<void> {
  await page.waitForFunction(() => {
    const el = document.getElementById('turn');
    return !!el && /your turn/i.test(el.textContent || '');
  }, undefined, { timeout: 20_000 });
}

async function performMove(page: Page, from: string, to: string): Promise<void> {
  const fromSquare = page.locator(`[data-square="${from}"]`);
  await fromSquare.waitFor({ state: 'visible', timeout: 10_000 });
  await expect(fromSquare).not.toHaveText('', { timeout: 10_000 });
  await fromSquare.click();
  await page.waitForTimeout(200);
  const toSquare = page.locator(`[data-square="${to}"]`);
  await toSquare.waitFor({ state: 'visible', timeout: 10_000 });
  await toSquare.click();
  await page.waitForTimeout(400);
}
