import assert from "node:assert/strict";
import { test } from "node:test";
import { parseRecentEmojis, quickEmojis, withRecentEmoji } from "../src/lib/emojis.ts";

test("recent emoji storage handles corrupt data and preserves complex emoji", () => {
  for (const raw of [null, "broken", "{}", "null"]) assert.deepEqual(parseRecentEmojis(raw), []);
  assert.deepEqual(parseRecentEmojis(JSON.stringify(["👍🏽", "👨‍👩‍👧‍👦", "👍🏽", null, 12, "", "x".repeat(65)])), ["👍🏽", "👨‍👩‍👧‍👦"]);
});
test("successful choices move to the front without duplicates and common emoji fill the quick row", () => {
  assert.deepEqual(withRecentEmoji(["🔥", "👏", "🤔"], "👏"), ["👏", "🔥", "🤔"]);
  assert.equal(withRecentEmoji(Array.from({ length: 20 }, (_, index) => String(index)), "🎉").length, 18);
  assert.deepEqual(quickEmojis(["🎉", "🔥"]), ["🎉", "🔥", "👋", "🤔", "😂", "👏"]);
  assert.equal(quickEmojis([]).length, 6);
});
