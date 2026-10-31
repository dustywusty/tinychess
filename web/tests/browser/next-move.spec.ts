import { test, expect } from '@playwright/test';

test('recommendation previews the live root and marks an outdated position', async ({ page, request }) => {
 const owner = `next-${Date.now()}`;
 const { id } = await (await request.post('/api/games', { data: { botId: 'machine', clientId: owner, color: 'w' } })).json();
 await page.addInitScript(cid => localStorage.setItem('tinychess:clientId', cid), owner);
 await page.setViewportSize({ width: 1440, height: 1100 });
 await page.goto(`/g/${id}`);
 const snapshot = async () => (await (await request.get(`/api/games/${id}/snapshot?clientId=${owner}`)).json());
 await page.getByRole('button', { name: 'Best move now', exact: true }).click();
 await expect(page.getByText('Current position — White to move')).toBeVisible();
 await expect(page.locator('.coach-recommendation')).toBeVisible();
 await expect(page.locator('.move-comparison')).toHaveCount(0);
 const source = await page.locator('[data-coach-square="origin"]').getAttribute('data-square');
 await expect(page.locator(`[data-square="${source}"]`)).not.toHaveAttribute('aria-label', /empty/);
 await page.getByRole('button', { name: 'Preview suggested move' }).click();
 await expect(page.locator(`[data-square="${source}"]`)).toHaveAttribute('aria-label', /empty/);
 expect((await snapshot()).uci).toEqual([]);
 await page.getByRole('button', { name: 'Before', exact: true }).click();
 await expect(page.locator(`[data-square="${source}"]`)).not.toHaveAttribute('aria-label', /empty/);
 await page.screenshot({ path: '/tmp/yourmove-next-desktop.png', fullPage: true });
 await request.post(`/api/games/${id}/move`, { data: { clientId: owner, uci: 'e2e4' } });
 await expect(page.getByText(/The game has moved on/)).toBeVisible();
 await expect.poll(async () => (await snapshot()).uci.length).toBe(2);
 await page.getByRole('button', { name: 'Analyze latest position', exact: true }).click();
 await expect(page.getByText(/The game has moved on/)).toHaveCount(0);
 await expect(page.locator('[data-square="e4"]')).toHaveAttribute('aria-label', /white pawn/);
 await page.setViewportSize({ width: 393, height: 852 });
 await page.screenshot({ path: '/tmp/yourmove-next-mobile.png', fullPage: true });
 expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
 await page.getByRole('button', { name: 'Return to game', exact: true }).click();
 expect((await snapshot()).uci).toHaveLength(2);
});

test('natural language recommends now and keeps why attached to the suggestion', async ({ page, request }) => {
 test.skip(process.env.PW_COACH_LIVE !== '1', 'Live provider check');
 const owner = `next-language-${Date.now()}`;
 const { id } = await (await request.post('/api/games', { data: { botId: 'machine', clientId: owner, color: 'w' } })).json();
 await page.addInitScript(cid => localStorage.setItem('tinychess:clientId', cid), owner);
 const results: any[] = [];
 await page.route('**/coach/runtime/agent/coach/run', async route => {
  const response = await route.fetch(); const body = await response.text();
  const events = body.split('\n').filter(line => line.startsWith('data:')).map(line => JSON.parse(line.slice(5)));
  results.push(events.filter(event => event.type === 'STATE_SNAPSHOT').at(-1)?.snapshot.lesson);
  await route.fulfill({ response });
 });
 await page.goto(`/g/${id}`);
 const ask = async (question: string) => {
  const toggle = page.locator('.coach-more:not([open]) > summary'); if (await toggle.count()) await toggle.click();
  await page.getByLabel('Ask about this lesson').fill(question);
  const count = results.length;
  await page.getByRole('button', { name: 'Ask', exact: true }).click();
  await expect.poll(() => results.length, { timeout: 25000 }).toBe(count + 1);
  await expect(page.getByRole('alert')).toHaveCount(0);
  return results.at(-1);
 };
 const suggested = await ask('What is my next best move?');
 expect(suggested.currentPosition).toBe(true); expect(suggested.ply).toBe(0);
 expect(suggested.recommendation.move).toBeTruthy(); expect(suggested.assessment).toBeUndefined();
 const explained = await ask('Why is that a good move?');
 expect(explained.id).toBe(suggested.id); expect(explained.positionId).toBe(suggested.positionId);
 const snapshot = await (await request.get(`/api/games/${id}/snapshot?clientId=${owner}`)).json();
 expect(snapshot.uci).toEqual([]);
 await page.getByRole('button', { name: 'Return to game', exact: true }).click();
 await page.locator('[data-square="g1"]').click(); await page.locator('[data-square="f3"]').click();
 await expect.poll(async () => (await (await request.get(`/api/games/${id}/snapshot?clientId=${owner}`)).json()).uci.length).toBe(2);
 await page.getByRole('button', { name: 'Review move', exact: true }).click();
 await expect(page.getByText('Reviewing move 1 — White')).toBeVisible();
 const latest = await ask('What should I play now in the live game?');
 expect(latest.currentPosition).toBe(true); expect(latest.ply).toBe(2);
 expect(latest.positionId).not.toBe(suggested.positionId);
});

for (const currentPosition of [true, false]) test(`expired ${currentPosition ? 'recommendation' : 'review'} refreshes without reusing the lesson`, async ({ page, request }) => {
 const owner = `expired-${Date.now()}`;
 const { id } = await (await request.post('/api/games', { data: { botId: 'machine', clientId: owner, color: 'w' } })).json();
 if (!currentPosition) await request.post(`/api/games/${id}/move`, { data: { clientId: owner, uci: 'e2e4' } });
 await page.addInitScript(cid => localStorage.setItem('tinychess:clientId', cid), owner);
 await page.setViewportSize({ width: 1440, height: 1100 });
 await page.goto(`/g/${id}`);
 await page.getByRole('button', { name: currentPosition ? 'Best move now' : 'Review move', exact: true }).click();
 await expect(page.getByRole('button', { name: 'Return to game' })).toBeVisible();
 let expired: any; let refreshed: any;
 await page.route('**/coach/runtime/agent/coach/run', async route => {
  const input = route.request().postDataJSON(); const command = input.forwardedProps.coach;
  if (command.action === 'explain') {
   expired = command;
   await route.fulfill({ contentType: 'text/event-stream', body: `data: ${JSON.stringify({ type: 'RUN_STARTED', threadId: input.threadId, runId: input.runId })}\n\ndata: ${JSON.stringify({ type: 'RUN_ERROR', message: 'lesson expired; start a new review' })}\n\n` });
  } else { refreshed = command; await route.continue(); }
 });
 await page.locator('.coach-more > summary').click();
 await page.getByRole('button', { name: 'Why?', exact: true }).click();
 await expect(page.getByRole('alert')).toContainText('This analysis expired');
 await page.getByRole('button', { name: 'Refresh analysis', exact: true }).click();
 await expect(page.getByRole('alert')).toHaveCount(0);
 expect(refreshed.action).toBe(currentPosition ? 'best_now' : 'review');
 expect(refreshed.lessonId).toBeUndefined();
 expect(refreshed.requestId).not.toBe(expired.requestId);
 for (const selector of ['.coach-panel', '.reaction-card', '.moves-panel']) {
  expect((await page.locator(selector).boundingBox())!.width).toBe(400);
 }
 await page.setViewportSize({ width: 960, height: 1100 });
 for (const selector of ['.coach-panel', '.reaction-card', '.moves-panel']) {
  expect((await page.locator(selector).boundingBox())!.width).toBe(330);
 }
 expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
});
