import assert from "node:assert/strict";

const origin = process.argv[2] ?? "http://127.0.0.1:8080";
const json = async (path, body) => {
  const response = await fetch(origin + path, {
    signal: AbortSignal.timeout(5000),
    ...(body ? { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(body) } : {}),
  });
  assert.equal(response.status, 200, path);
  return response.json();
};
let ready = false;
for (let attempt = 0; attempt < 30; attempt++) {
  try { assert.deepEqual(await json("/healthz"), { ok: true }); ready = true; break; } catch { /* Allow startup. */ }
  await new Promise((resolve) => setTimeout(resolve, 500));
}
assert.ok(ready, "server must become healthy");
assert.equal(typeof (await json("/api/version")).commit, "string");
const html = await (await fetch(origin + "/g/smoke-spa", { signal: AbortSignal.timeout(5000) })).text();
const asset = html.match(/src="([^"]+\.js)"/)?.[1];
assert.ok(asset, "embedded SPA must include its JavaScript bundle");
const bundle = await fetch(origin + asset, { signal: AbortSignal.timeout(5000) });
assert.equal(bundle.status, 200);
assert.match(bundle.headers.get("content-type"), /javascript/);
await bundle.arrayBuffer();

const { id } = await json("/api/games", {});
const path = `/api/games/${id}`;
const a = await json(path + "/snapshot?clientId=smoke-a");
const b = await json(path + "/snapshot?clientId=smoke-b");
assert.equal(a.role, "player");
assert.equal(b.role, "player");
assert.notEqual(a.color, b.color);
const white = a.color === "w" ? "smoke-a" : "smoke-b";
const black = white === "smoke-a" ? "smoke-b" : "smoke-a";
const controller = new AbortController();
const timeout = setTimeout(() => controller.abort(), 20000);
try {
  const response = await fetch(origin + `/api/sse/${id}?clientId=smoke-watcher`, { signal: controller.signal });
  assert.equal(response.status, 200);
  assert.match(response.headers.get("content-type"), /text\/event-stream/);
  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  const event = async () => {
    while (!buffer.includes("\n\n")) {
      const { done, value } = await reader.read();
      assert.ok(!done, "SSE must remain open");
      buffer += decoder.decode(value, { stream: true });
    }
    const end = buffer.indexOf("\n\n");
    const block = buffer.slice(0, end);
    buffer = buffer.slice(end + 2);
    return JSON.parse(block.replace(/^data: /, ""));
  };
  assert.equal((await event()).role, "spectator");
  const moves = ["d2d4", "e7e5", "g1f3", "e5d4", "f3d4"];
  for (const [index, uci] of moves.entries()) {
    const result = await json(path + "/move", { uci, clientId: index % 2 ? black : white });
    assert.equal(result.ok, true);
    let update;
    do { update = await event(); } while (!update.uci);
    assert.deepEqual(update.uci, moves.slice(0, index + 1));
  }
  assert.equal((await json(path + "/react", { emoji: "👏", sender: white })).ok, true);
  let reaction;
  do { reaction = await event(); } while (reaction.kind !== "emoji");
  assert.equal(reaction.emoji, "👏");
  assert.notEqual(reaction.sender, white, "spectators must not receive seat credentials");
  assert.match(reaction.sender, /^public:[0-9a-f]{64}$/);
  assert.equal((await json(path + "/move", { uci: "b8c6", clientId: reaction.sender })).ok, false, "public alias cannot move");
  assert.deepEqual((await json(path + "/snapshot?clientId=smoke-watcher")).uci, moves);
} finally {
  clearTimeout(timeout);
  controller.abort();
}
console.log("Smoke test passed: health, version, embedded web, seats, captures, SSE, and reactions.");
