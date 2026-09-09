import assert from "node:assert/strict";
import test from "node:test";
import { forgetGame, loadRecentGames, recordGameSeen } from "../src/lib/recentGames.ts";

test("recent games persist friend/bot metadata and upgrade legacy entries on a new snapshot", () => {
  const data = new Map<string, string>();
  Object.defineProperty(globalThis, "localStorage", { configurable: true, value: {
    getItem: (key: string) => data.get(key) ?? null,
    setItem: (key: string, value: string) => data.set(key, value),
  } });
  assert.deepEqual(loadRecentGames(), {});
  const snapshot = { status: "White to move", lastSeen: 123, pgn: "" };
  recordGameSeen("friend", snapshot);
  recordGameSeen("computer", { ...snapshot, bot: { id: "pip", color: "b", policyVersion: 1 }, color: "w", role: "player" });
  assert.equal(loadRecentGames().friend.opponentType, "human");
  assert.equal(loadRecentGames().friend.botId, undefined);
  assert.equal(loadRecentGames().computer.opponentType, "bot");
  assert.equal(loadRecentGames().computer.botId, "pip");
  recordGameSeen("computer", { ...snapshot, status: "0-1 by Checkmate", bot: { id: "pip", color: "b", policyVersion: 1 } });
  assert.equal(loadRecentGames().computer.playerColor, "w");
  assert.equal(loadRecentGames().computer.role, "player");
  assert.equal(loadRecentGames().computer.result, "0-1");
  recordGameSeen("computer", { ...snapshot, color: null, role: "spectator" });
  assert.equal(loadRecentGames().computer.playerColor, null);
  assert.equal(loadRecentGames().computer.role, "spectator");
  const legacy = { id: "legacy", createdAt: 1, lastSeen: 1, lastSeenLocal: 1, status: "", result: "" };
  data.set("tinychess:games:v1", JSON.stringify({ legacy }));
  assert.equal(loadRecentGames().legacy.opponentType, undefined);
  recordGameSeen("legacy", { ...snapshot, bot: { id: "ada", color: "w", policyVersion: 1 } });
  assert.equal(loadRecentGames().legacy.createdAt, 1);
  assert.equal(loadRecentGames().legacy.botId, "ada");
  assert.deepEqual(forgetGame("legacy"), {});
  assert.deepEqual(loadRecentGames(), {});
});
