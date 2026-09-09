import assert from "node:assert/strict";
import { randomUUID } from "node:crypto";
import test from "node:test";
import { persistentIdentity } from "../src/lib/persistentIdentity.ts";

test("concurrent identity requests share one persisted UUID and survive app restart", async () => {
  let saved: string | null = null;
  let writes = 0;
  const storage = {
    async getItem() { return saved; },
    async setItem(_key: string, value: string) { saved = value; writes++; },
  };
  const identity = persistentIdentity(storage, "id", randomUUID);
  const results = await Promise.all(Array.from({ length: 20 }, () => identity()));
  assert.equal(new Set(results).size, 1);
  assert.match(results[0], /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/);
  assert.equal(writes, 1);
  assert.equal(await persistentIdentity(storage, "id", randomUUID)(), results[0]);
});

test("existing identities are preserved and failed persistence can retry", async () => {
  const old = persistentIdentity({ async getItem() { return "legacy-mobile-id"; }, async setItem() { assert.fail("replaced legacy identity"); } }, "id", randomUUID);
  assert.equal(await old(), "legacy-mobile-id");
  let fail = true;
  const identity = persistentIdentity({ async getItem() { return null; }, async setItem() { if (fail) throw new Error("storage unavailable"); } }, "id", randomUUID);
  await assert.rejects(identity(), /storage unavailable/);
  fail = false;
  assert.ok(await identity());
});
