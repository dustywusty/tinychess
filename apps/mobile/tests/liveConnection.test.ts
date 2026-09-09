import assert from "node:assert/strict";
import { test } from "node:test";
import { liveConnection, type EventStream } from "../src/lib/liveConnection.ts";

test("stream retries, ignores stale callbacks, and cancels timers on cleanup", (context) => {
  context.mock.timers.enable({ apis: ["setTimeout"] });
  const streams: { emit: (value: string) => void; fail: () => void; closed: boolean }[] = [];
  const received: unknown[] = [];
  let errors = 0;
  const create = () => {
    const item = { emit: (_: string) => {}, fail: () => {}, closed: false };
    streams.push(item);
    return {
      addEventListener(type: string, callback: (event?: { data: string }) => void) {
        if (type === "message") item.emit = (data: string) => callback({ data });
        else item.fail = () => callback();
      },
      close() { item.closed = true; },
    } as EventStream;
  };
  const stop = liveConnection(create, (event) => received.push(event), () => errors++, { retry: 10, maxRetry: 40, heartbeat: 100 });
  streams[0].emit('{"kind":"state"}');
  streams[0].emit("not json");
  assert.equal(received.length, 1);
  streams[0].fail();
  streams[0].emit('{"stale":true}');
  assert.equal(received.length, 1);
  context.mock.timers.tick(10);
  assert.equal(streams.length, 2);
  assert.equal(streams[0].closed, true);
  streams[1].emit("{}");
  context.mock.timers.tick(100);
  assert.equal(errors, 2);
  stop();
  context.mock.timers.tick(500);
  assert.equal(streams.length, 2);
  assert.equal(streams[1].closed, true);
});
