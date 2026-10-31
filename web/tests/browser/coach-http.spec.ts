import { test, expect } from '@playwright/test';

test('coach works without randomUUID and returns without runtime stop requests', async ({ page, request }) => {
  await page.addInitScript(() => {
    Object.defineProperty(crypto, 'randomUUID', { value: undefined, configurable: true });
  });
  const errors: string[] = [];
  const stops: string[] = [];
  page.on('pageerror', error => errors.push(error.message));
  page.on('request', req => { if (req.url().includes('/coach/stop/')) stops.push(req.url()); });
  const { id } = await (await request.post('/api/games', { data: {} })).json();
  await page.goto(`/g/${id}`);
  await expect.poll(() => page.evaluate(() => localStorage.getItem('tinychess:clientId'))).toBeTruthy();
  const owner = await page.evaluate(() => localStorage.getItem('tinychess:clientId'));
  expect(owner).toMatch(/^[0-9a-f-]{36}$/);
  const seat = await (await request.get(`/api/games/${id}/snapshot?clientId=${owner}`)).json();
  const other = `http-test-${id}`;
  await request.get(`/api/games/${id}/snapshot?clientId=${other}`);
  const white = seat.color === 'w' || seat.color === 'white' ? owner : other;
  const black = white === owner ? other : owner;
  for (const [i, uci] of ['f2f3', 'e7e5', 'g2g4', 'd8h4'].entries()) {
    expect((await (await request.post(`/api/games/${id}/move`, { data: { clientId: i % 2 ? black : white, uci } })).json()).ok).toBe(true);
  }
  await page.goto(`/g/${id}`);
  await page.locator('.coach-more > summary').click();
  await page.getByLabel('Move to review').selectOption('0');
  const run = page.waitForResponse(res => res.url().endsWith('/coach/run') && res.request().method() === 'POST');
  await page.getByRole('button', { name: 'Review move', exact: true }).click();
  expect((await run).ok()).toBe(true);
  await expect(page.getByText('Reviewing move 1 — White')).toBeVisible();
  await page.getByRole('button', { name: 'Return to game', exact: true }).click();
  await expect(page.getByText('Reviewing move 1 — White')).toHaveCount(0);
  expect(errors).toEqual([]);
  expect(stops).toEqual([]);
});
