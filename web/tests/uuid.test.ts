import assert from "node:assert/strict";
import test from "node:test";
import { randomUUID } from "../src/lib/uuid.ts";

test("UUIDs remain secure and valid when HTTP does not expose randomUUID", () => {
  const original = Object.getOwnPropertyDescriptor(globalThis, "crypto");
  const getRandomValues = crypto.getRandomValues.bind(crypto);
  Object.defineProperty(globalThis, "crypto", { configurable: true, value: { getRandomValues } });
  try {
    const ids = Array.from({ length: 100 }, randomUUID);
    assert.equal(new Set(ids).size, ids.length);
    for (const id of ids) assert.match(id, /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/);
  } finally {
    Object.defineProperty(globalThis, "crypto", original!);
  }
});
