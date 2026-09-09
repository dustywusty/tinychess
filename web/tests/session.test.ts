import assert from "node:assert/strict";
import test from "node:test";

const key = "tinychess:clientId";
function storage() {
  const data = new Map<string, string>();
  return { getItem: (key: string) => data.get(key) ?? null, setItem: (key: string, value: string) => data.set(key, value), removeItem: (key: string) => data.delete(key) };
}

test("anonymous browser identity persists across tabs/reloads and migrates legacy sessions", async () => {
  const local = storage();
  const session = storage();
  Object.defineProperty(globalThis, "localStorage", { configurable: true, value: local });
  Object.defineProperty(globalThis, "sessionStorage", { configurable: true, value: session });
  const first = await import("../src/lib/session.ts?first");
  const id = first.getOrCreateClientId();
  assert.match(id, /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/);
  assert.equal(local.getItem(key), id);
  const nextTab = await import("../src/lib/session.ts?next-tab");
  assert.equal(nextTab.getOrCreateClientId(), id);
  first.clearClientId();
  session.setItem(key, "existing-seat-id");
  assert.equal(first.getOrCreateClientId(), "existing-seat-id");
  assert.equal(local.getItem(key), "existing-seat-id");
  assert.equal(session.getItem(key), null);
  session.setItem(key, "stale-tab-id");
  assert.equal(nextTab.getOrCreateClientId(), "existing-seat-id");
  first.clearClientId();
  assert.equal(local.getItem(key), null);
  assert.equal(session.getItem(key), null);
});

test("blocked storage keeps a stable in-memory identity without weak randomness", async () => {
  for (const name of ["localStorage", "sessionStorage"]) {
    Object.defineProperty(globalThis, name, { configurable: true, get() { throw new Error("blocked"); } });
  }
  const session = await import("../src/lib/session.ts?blocked");
  assert.equal(session.getOrCreateClientId(), session.getOrCreateClientId());
});
