import { test, expect } from '@playwright/test';

test('Jev chooses comparison, piece focus, reply, and practice views in conversation', async ({ page, request }) => {
 test.skip(process.env.PW_COACH_LIVE !== '1', 'Live Jev teaching evaluation');
 const owner = `teaching-${Date.now()}`;
 const { id } = await (await request.post('/api/games', { data: {} })).json();
 const seat = await (await request.get(`/api/games/${id}/snapshot?clientId=${owner}`)).json();
 const other = `opponent-${owner}`;
 await request.get(`/api/games/${id}/snapshot?clientId=${other}`);
 const white = seat.color === 'w' || seat.color === 'white' ? owner : other;
 const black = white === owner ? other : owner;
 for (const [i, uci] of ['f2f3', 'e7e5', 'g2g4', 'd8h4'].entries()) {
  expect((await (await request.post(`/api/games/${id}/move`, { data: { clientId: i % 2 ? black : white, uci } })).json()).ok).toBe(true);
 }
 await page.addInitScript(cid => localStorage.setItem('tinychess:clientId', cid), white);
 await page.setViewportSize({ width: 1440, height: 1100 });
 const results: any[] = [];
 let liveWrites = 0;
 page.on('request', r => { if (r.url().endsWith('/move') && r.method() === 'POST') liveWrites++; });
 await page.route('**/coach/runtime/agent/coach/run', async route => {
  const response = await route.fetch(); const body = await response.text();
  const events = body.split('\n').filter(line => line.startsWith('data:')).map(line => JSON.parse(line.slice(5)));
  results.push(events.filter(event => event.type === 'STATE_SNAPSHOT').at(-1)?.snapshot.lesson);
  await route.fulfill({ response });
 });
 await page.goto(`/g/${id}`);
 await page.getByRole('button', { name: 'Review move', exact: true }).click();
 await expect(page.getByText('Reviewing move 2 — White')).toBeVisible();
 const ask = async (text: string) => {
  await page.getByLabel('Ask about this lesson').fill(text);
  const count = results.length;
  await page.getByRole('button', { name: 'Ask', exact: true }).click();
  await expect.poll(() => results.length, { timeout: 25000 }).toBe(count + 1);
  await expect(page.getByRole('alert')).toHaveCount(0);
  return results.at(-1);
 };
 const comparison = await ask('Why was my move worse? Show both choices side by side on the board.');
 expect(comparison.teaching.view).toBe('compare'); expect(comparison.branch).toBe('compare');
 await expect(page.locator('.coach-arrows line')).toHaveCount(2);
 const focus = await ask(`I don't understand ${comparison.assessment.best}. Show me which piece moves and where it goes.`);
 expect(focus.teaching.view).toBe('move'); expect(focus.branch).toBe('alternative'); expect(focus.step).toBe(0);
 await expect(page.locator('[data-coach-focus="true"]')).toHaveCount(1);
 await expect(page.locator('.coach-next-question')).toBeVisible();
 await page.screenshot({ path: '/tmp/yourmove-teaching-desktop.png', fullPage: true });
 const reply = await ask('Show the opponent’s first reply after that suggested move.');
 expect(reply.teaching.view).toBe('reply'); expect(reply.step).toBe(2);
 expect(reply.conversation.length).toBe(3);
 if (reply.teaching.prompt) {
  const count = results.length;
  await page.locator('.coach-next-question').click();
  await expect.poll(() => results.length, { timeout: 25000 }).toBe(count + 1);
  await expect(page.getByRole('alert')).toHaveCount(0);
  expect(results.at(-1).conversation.at(-1).question).toBe(reply.teaching.prompt.question);
 }
 const practice = await ask('Let me solve this position myself. Don’t tell me the answer.');
 expect(practice.teaching.view).toBe('exercise'); expect(practice.reveal).toBe(0);
 expect(practice.assessment).toBeUndefined(); expect(practice.views).toBeUndefined();
 await expect(page.locator('.coach-arrows')).toHaveCount(0);
 await expect(page.locator('[data-coach-focus="true"]')).toHaveCount(0);
 await page.setViewportSize({ width: 393, height: 852 });
 await page.screenshot({ path: '/tmp/yourmove-teaching-mobile.png', fullPage: true });
 expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
 expect(liveWrites).toBe(0);
 const snapshot = await (await request.get(`/api/games/${id}/snapshot?clientId=${white}`)).json();
 expect(snapshot.uci).toEqual(['f2f3', 'e7e5', 'g2g4', 'd8h4']);
});
