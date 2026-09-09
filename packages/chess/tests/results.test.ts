import assert from "node:assert/strict";
import test from "node:test";
import { gameResult, resultBanner } from "../src/results.ts";

test("results come from terminal server status or the final PGN result", () => {
  assert.equal(gameResult("0-1 by Checkmate"), "0-1");
  assert.equal(gameResult("1/2-1/2 by Stalemate"), "1/2-1/2");
  assert.equal(gameResult("", "1. e4 e5 1-0\n"), "1-0");
  assert.equal(gameResult("", '[Event "1-0"]\n1. e4 *'), undefined);
  assert.equal(gameResult("", "1. e4 { 0-1 } *"), undefined);
  assert.equal(gameResult("White to move"), undefined);
  assert.equal(resultBanner({ result: "*" }), undefined);
});

test("wins and losses use the saved seat, never the board orientation or bot side", () => {
  for (const playerColor of ["w", "white"]) {
    assert.equal(resultBanner({ result: "1-0", role: "player", playerColor })?.tone, "win");
    assert.equal(resultBanner({ result: "0-1", role: "player", playerColor })?.tone, "loss");
  }
  for (const playerColor of ["b", "black"]) {
    assert.equal(resultBanner({ result: "0-1", role: "player", playerColor })?.tone, "win");
    assert.equal(resultBanner({ result: "1-0", role: "player", playerColor })?.tone, "loss");
  }
});

test("draws, spectators and legacy records never claim a personal win", () => {
  assert.equal(resultBanner({ status: "1/2-1/2 by ThreefoldRepetition", role: "player", playerColor: "b" })?.label, "Draw");
  assert.equal(resultBanner({ result: "1-0", role: "spectator", playerColor: "w" })?.label, "White wins");
  assert.equal(resultBanner({ result: "0-1" })?.label, "Black wins");
  assert.equal(resultBanner({ result: "0-1", role: "player", playerColor: null })?.tone, "neutral");
});
