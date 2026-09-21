import { test, expect, type Page } from '@playwright/test';

async function expandCoach(page: Page) {
 const toggle = page.locator('.coach-more:not([open]) > summary');
 if (await toggle.count()) await toggle.click();
}

test('retry after a lost committed response reveals only one hint', async ({ page, request }) => {
 const owner = `retry-coach-${Date.now()}`;
 const { id } = await (await request.post('/api/games', { data: { botId: 'machine', clientId: owner, color: 'w' } })).json();
 await request.post(`/api/games/${id}/move`, { data: { clientId: owner, uci: 'e2e4' } });
 await page.addInitScript(cid => localStorage.setItem('tinychess:clientId', cid), owner);
 await page.goto(`/g/${id}`);
 await page.getByRole('button', { name: 'Review move', exact: true }).click();
 await expect(page.getByText('Reviewing move 1 — White')).toBeVisible();
 await expandCoach(page);
 await page.getByRole('button', { name: 'Try it', exact: true }).click();
 await expect(page.getByText('Look for checks, captures, and threats. What would you play?')).toBeVisible();
 const requests: Record<string, unknown>[] = [];
 const results: { revision: number; reveal: number }[] = [];
 await page.route('**/coach/runtime/agent/coach/run', async route => {
  const command = route.request().postDataJSON().forwardedProps.coach;
  if (command.action !== 'reveal') { await route.continue(); return; }
  requests.push(command);
  const response = await route.fetch();
  const body = await response.text();
  const snapshot = body.split('\n').filter(line => line.startsWith('data:')).map(line => JSON.parse(line.slice(5))).find(event => event.type === 'STATE_SNAPSHOT');
  results.push(snapshot.snapshot.lesson);
  // The server has committed this action; lose only its response to the browser.
  if (requests.length === 1) await route.abort('failed');
  else await route.fulfill({ response });
 });
 await page.getByRole('button', { name: 'Reveal next hint', exact: true }).click();
 await page.getByRole('button', { name: 'Retry request', exact: true }).click();
 await expect(page.getByText(/^Consider the piece on/)).toBeVisible();
 expect(requests).toHaveLength(2);
 expect(requests[0].requestId).toBeTruthy();
 expect(requests[1]).toEqual(requests[0]);
 expect(results[0].reveal).toBe(1);
 expect(results[1].reveal).toBe(1);
 expect(results[1].revision).toBe(results[0].revision);
 await expect(page.locator('.coach-arrows')).toHaveCount(0);
 await page.getByRole('button', { name: 'Return to game', exact: true }).click();
 await expect(page.getByText('Reviewing move 1 — White')).toHaveCount(0);
});

test('review, hints, exercise, and return preserve a completed game', async ({ page, request }) => {
 const owner = `coach-${Date.now()}`;
 const created = await request.post('/api/games', { data: {} }); const { id } = await created.json();
 const seat = await (await request.get(`/api/games/${id}/snapshot?clientId=${owner}`)).json();
 const white = seat.color === "w" || seat.color === "white" ? owner : `opponent-${owner}`;
 const black = white === owner ? `opponent-${owner}` : owner;
 await request.get(`/api/games/${id}/snapshot?clientId=opponent-${owner}`);
 for (const [i, uci] of ['f2f3', 'e7e5', 'g2g4', 'd8h4'].entries()) {
  const moved = await request.post(`/api/games/${id}/move`, { data: { clientId: i % 2 ? black : white, uci } });
  expect((await moved.json()).ok).toBe(true);
 }
 await page.addInitScript(cid => localStorage.setItem('tinychess:clientId', cid), white);
 const errors: string[] = []; page.on('pageerror', e => errors.push(e.message));
 await page.goto(`/g/${id}`);
 await expect(page.getByRole('button', { name: 'Review move', exact: true })).toBeVisible();
 await page.getByRole('button', { name: 'Review move', exact: true }).click();
 await expect(page.getByText('Reviewing move 2 — White')).toBeVisible();
 await expandCoach(page);
 await page.getByRole('button', { name: 'Why?', exact: true }).click();
 await expect(page.getByRole('status').filter({hasText: 'Preparing'})).toHaveCount(0);
 let navigationRequests = 0;
 page.on('request', r => { if (r.url().endsWith('/run')) navigationRequests++; });
 await page.getByRole('button', { name: 'Alternative', exact: true }).click();
 await expect(page.locator('.coach-arrows')).toHaveCount(0);
 await expect(page.locator('[data-coach-square="destination"]')).toHaveCount(1);
 await page.getByRole('button', { name: 'Before', exact: true }).click();
 await expect(page.locator('.coach-arrows')).toHaveAttribute('aria-label', /Alternative/);
 const boardBounds = await page.locator('.chess-board').boundingBox();
 const overlayBounds = await page.locator('.coach-arrows').boundingBox();
 expect(overlayBounds).toEqual(boardBounds);
 await page.getByRole('button', { name: 'Compare', exact: true }).click();
 await expect(page.locator('[data-square="g2"]')).toHaveAttribute('aria-label', /white pawn/);
 await expect(page.locator('[data-square="g4"]')).toHaveAttribute('aria-label', /empty/);
 await expect(page.locator('.coach-arrows line')).toHaveCount(2);
 await page.setViewportSize({ width: 1440, height: 1100 });
 await page.screenshot({ path: '/tmp/yourmove-review-compare.png', fullPage: true });
 await page.setViewportSize({ width: 393, height: 852 });
 await page.screenshot({ path: '/tmp/yourmove-review-compare-mobile.png', fullPage: true });
 expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
 await page.getByRole('button', { name: 'Alternative', exact: true }).click();
 await page.getByRole('button', { name: 'Before', exact: true }).click();
 const arrowX = Number(await page.locator('.coach-arrows line').first().getAttribute('x1'));
 await page.getByRole('button', { name: /Flip board/ }).click();
 await expect(page.locator('.coach-arrows line').first()).toHaveAttribute('x1', String(8 - arrowX));
 await page.getByRole('button', { name: 'Continue', exact: true }).click();
 await expect(page.getByRole('status').filter({hasText: 'Preparing'})).toHaveCount(0);
 expect(navigationRequests).toBe(0);
 await expandCoach(page);
 await page.getByRole('button', { name: 'Try it', exact: true }).click();
 await expect(page.getByText('Look for checks, captures, and threats. What would you play?')).toBeVisible();
 await expect(page.locator('.coach-arrows')).toHaveCount(0);
 let writes = 0; page.on('request', r => { if (r.url().endsWith('/move') && r.method() === 'POST') writes++; });
 await page.locator('[data-square="g2"]').click(); await page.locator('[data-square="g3"]').click();
 await expect(page.getByText(/sound alternative|prefers another|deeper comparable|matches the engine/)).toBeVisible();
 expect(writes).toBe(0);
 await page.getByRole('button', { name: 'Return to game', exact: true }).click();
 await expect(page.getByText('Reviewing move 2 — White')).toHaveCount(0);
 const snapshot = await (await request.get(`/api/games/${id}/snapshot?clientId=${owner}`)).json();
 expect(snapshot.uci).toEqual(['f2f3', 'e7e5', 'g2g4', 'd8h4']);
 expect(errors).toEqual([]);
});

test('a live computer response leaves review pinned and return shows the latest board', async ({ page, request }) => {
 const owner = `live-coach-${Date.now()}`;
 const created = await request.post('/api/games', { data: { botId: 'machine', clientId: owner, color: 'w' } });
 const { id } = await created.json();
 expect((await (await request.post(`/api/games/${id}/move`, { data: { clientId: owner, uci: 'e2e4' } })).json()).ok).toBe(true);
 await page.addInitScript(cid => localStorage.setItem('tinychess:clientId', cid), owner);
 await page.goto(`/g/${id}`);
 await page.getByRole('button', { name: 'Review move', exact: true }).click();
 await expect(page.getByText('Reviewing move 1 — White')).toBeVisible();
 const snapshot = async () => (await (await request.get(`/api/games/${id}/snapshot?clientId=${owner}`)).json());
 await expect.poll(async () => (await snapshot()).uci.length).toBe(2);
 const live = await snapshot();
 const from = live.uci[1].slice(0, 2);
 // The review is still immediately after White's move, before Black's reply.
 await expect(page.locator(`[data-square="${from}"]`)).not.toHaveAttribute('aria-label', /empty/);
 await page.getByRole('button', { name: 'Return to game', exact: true }).click();
 await expect(page.locator(`[data-square="${from}"]`)).toHaveAttribute('aria-label', /empty/);
 await page.getByRole('button', { name: 'Review move', exact: true }).click();
 await expect(page.getByText('Reviewing move 1 — White')).toBeVisible();
 await page.reload();
 await expect(page.getByText('Reviewing move 1 — White')).toHaveCount(0);
 await expect(page.locator(`[data-square="${from}"]`)).toHaveAttribute('aria-label', /empty/);
 expect((await snapshot()).uci).toEqual(live.uci);
});

test('engine and language errors preserve retry, local navigation, and return', async ({ page, request }) => {
 const owner = `failure-coach-${Date.now()}`;
 const { id } = await (await request.post('/api/games', { data: { botId: 'machine', clientId: owner, color: 'w' } })).json();
 await request.post(`/api/games/${id}/move`, { data: { clientId: owner, uci: 'e2e4' } });
 await page.addInitScript(cid => localStorage.setItem('tinychess:clientId', cid), owner);
 let failReview = true;
 await page.route('**/coach/runtime/agent/coach/run', async route => {
  const input = route.request().postDataJSON();
  const action = input.forwardedProps.coach.action;
  let message = '';
  if (action === 'review' && failReview) { failReview = false; message = 'engine timed out or stopped'; }
  if (action === 'language') message = 'language controls are unavailable; use the buttons';
  if (!message) { await route.continue(); return; }
  const events = [
   { type: 'RUN_STARTED', threadId: input.threadId, runId: input.runId },
   { type: 'RUN_ERROR', message },
  ];
  await route.fulfill({ contentType: 'text/event-stream', body: events.map(event => `data: ${JSON.stringify(event)}\n\n`).join('') });
 });
 await page.goto(`/g/${id}`);
 await page.getByRole('button', { name: 'Review move', exact: true }).click();
 await expect(page.getByRole('alert')).toContainText('engine timed out');
 await expect(page.getByText('Reviewing move 1 — White')).toHaveCount(0);
 await page.getByRole('button', { name: 'Retry request', exact: true }).click();
 await expect(page.getByText('Reviewing move 1 — White')).toBeVisible();
 await expandCoach(page);
 await page.getByLabel('Ask about this lesson').fill('Show their response');
 await page.getByRole('button', { name: 'Ask', exact: true }).click();
 await expect(page.getByRole('alert')).toContainText('use the buttons');
 await page.getByRole('button', { name: 'Alternative', exact: true }).click();
 await expect(page.locator('.coach-arrows')).toHaveCount(0);
 await expect(page.locator('[data-coach-square="destination"]')).toHaveCount(1);
 await page.getByRole('button', { name: 'Before', exact: true }).click();
 await expect(page.locator('.coach-arrows')).toHaveAttribute('aria-label', /Alternative/);
 const boardBounds = await page.locator('.chess-board').boundingBox();
 const overlayBounds = await page.locator('.coach-arrows').boundingBox();
 expect(overlayBounds).toEqual(boardBounds);
 await expect(page.getByRole('alert')).toHaveCount(0);
 await page.getByRole('button', { name: 'Return to game', exact: true }).click();
 await expect(page.getByText('Reviewing move 1 — White')).toHaveCount(0);
});

test('live Jev opens a lesson and controls its board through CopilotKit', async ({ page, request }) => {
 test.skip(process.env.PW_COACH_LIVE !== '1', 'Opt-in paid provider validation; server credentials required');
 const owner = `provider-coach-${Date.now()}`;
 const { id } = await (await request.post('/api/games', { data: { botId: 'machine', clientId: owner, color: 'w' } })).json();
 await request.post(`/api/games/${id}/move`, { data: { clientId: owner, uci: 'e2e4' } });
 await page.addInitScript(cid => localStorage.setItem('tinychess:clientId', cid), owner);
 const bodies: string[] = [];
 await page.route('**/coach/runtime/agent/coach/run', async route => {
  const response = await route.fetch();
  bodies.push(await response.text());
  await route.fulfill({ response });
 });
 await page.goto(`/g/${id}`);
 const ask = async (text: string) => {
  await expandCoach(page);
 await page.getByLabel('Ask about this lesson').fill(text);
  const count = bodies.length;
  await page.getByRole('button', { name: 'Ask', exact: true }).click();
  await expect.poll(() => bodies.length, { timeout: 20000 }).toBe(count + 1);
  const body = bodies[count];
  await expect(page.getByRole('status').filter({ hasText: 'Preparing' })).toHaveCount(0, { timeout: 20000 });
  await expect(page.getByRole('alert')).toHaveCount(0);
  const events = body.split('\n').filter(line => line.startsWith('data:')).map(line => JSON.parse(line.slice(5)));
  const snapshots = events.filter(event => event.type === 'STATE_SNAPSHOT');
  return { lesson: snapshots.at(-1)?.snapshot.lesson, narrated: events.some(event => event.type === 'TEXT_MESSAGE_START') };
 };
 await ask('Review the selected move.');
 await expect(page.getByText('Reviewing move 1 — White')).toBeVisible();
 const alternative = await ask('Show me what I should have done.');
 expect(alternative.lesson.branch).toBe('alternative');
 expect(alternative.narrated).toBe(false);
 const next = await ask('Show their response.');
 expect(next.lesson.step).toBe(alternative.lesson.step + 1);
 expect(next.narrated).toBe(false);
 const back = await ask('Go back one move.');
 expect(back.lesson.step).toBe(alternative.lesson.step);
 expect(back.narrated).toBe(false);
 await ask('Let me find the answer myself.');
 await expect(page.getByText('Try a move on the board. It stays in this exercise.')).toBeVisible();
 await expect(page.locator('.coach-arrows')).toHaveCount(0);
 await ask('Go back to my game.');
 await expect(page.getByText('Reviewing move 1 — White')).toHaveCount(0);
});
