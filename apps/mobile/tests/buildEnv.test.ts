import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { test } from "node:test";

test("release builds require an HTTPS API origin without credentials or /api", () => {
  for (const value of ["", "http://localhost:8080", "https://localhost", "https://example.com/api", "https://user:secret@example.com", "https://example.com/?key=secret"]) {
    const result = spawnSync(process.execPath, ["scripts/check-build-env.cjs"], { env: { ...process.env, EXPO_PUBLIC_API_URL: value }, encoding: "utf8" });
    assert.equal(result.status, 1, value);
    assert.ok(!result.stderr.includes("user:secret"));
  }
  const result = spawnSync(process.execPath, ["scripts/check-build-env.cjs"], { env: { ...process.env, EXPO_PUBLIC_API_URL: "https://chess.example.com" } });
  assert.equal(result.status, 0);
});
