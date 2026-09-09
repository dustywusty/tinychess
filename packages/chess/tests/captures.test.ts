import assert from "node:assert/strict";
import { test } from "node:test";
import { capturedLabel, capturedPieces } from "../src/index.ts";

test("empty and non-capturing moves have no captured pieces", () => {
  for (const moves of [[], ["e2e4", "e7e5", "g1f3"]]) assert.deepEqual(capturedPieces(moves), { white: [], black: [] });
  assert.deepEqual(capturedPieces(["e1g1"], "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1"), { white: [], black: [] });
});
test("each player owns their captures with the opponent's piece color", () => {
  const moves = ["d2d4", "e7e5", "g1f3", "e5d4", "f3d4"];
  assert.deepEqual(capturedPieces(moves.slice(0, 4)), { white: [], black: ["P"] });
  assert.deepEqual(capturedPieces(moves), { white: ["p"], black: ["P"] });
  assert.deepEqual(capturedPieces([...moves, "bad", "d8d4"]), { white: ["p"], black: ["P"] });
});
test("en passant records the captured pawn", () => {
  assert.deepEqual(capturedPieces(["e2e4", "a7a6", "e4e5", "d7d5", "e5d6"]), { white: ["p"], black: [] });
});
test("promotion is not a lost pawn; a captured promoted queen is a queen", () => {
  const fen = "1r5k/P7/8/8/8/8/8/K7 w - - 0 1";
  assert.deepEqual(capturedPieces(["a7a8q"], fen), { white: [], black: [] });
  assert.deepEqual(capturedPieces(["a7a8q", "b8a8"], fen), { white: [], black: ["Q"] });
  assert.deepEqual(capturedPieces(["a7b8q"], fen), { white: ["r"], black: [] });
});
test("captures sort by piece and provide counted screen reader labels", () => {
  assert.deepEqual(capturedPieces(["a1a2", "h8g8", "a2b2"], "7k/8/8/8/8/8/pq6/R6K w - - 0 1").white, ["q", "p"]);
  assert.equal(capturedLabel("white", ["q", "p", "p"]), "Captured by White: 1 black queen, 2 black pawns");
  assert.equal(capturedLabel("black", ["N"]), "Captured by Black: 1 white knight");
});
