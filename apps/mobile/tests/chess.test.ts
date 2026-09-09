import assert from "node:assert/strict";
import { test } from "node:test";
import { Chess } from "chess.js";
import { gameIDFromInput, initialFEN, legalMoves, mergeState, moveLabels, squareAt } from "../src/lib/chess.ts";

test("game invitations accept IDs and supported links, stripping query strings", () => {
  for (const input of ["abc-123", " https://chess.test/g/abc-123/?invite=yes#board ", "yourmove://g/abc-123"]) {
    assert.equal(gameIDFromInput(input), "abc-123");
  }
  for (const input of ["", "hello there", "https://chess.test/profile/alice", "javascript:alert(1)", "https://chess.test/g/a/b"]) {
    assert.equal(gameIDFromInput(input), "");
  }
});
test("board coordinates rotate both axes for the black player", () => {
  assert.equal(squareAt(0, 0, "white"), "a8");
  assert.equal(squareAt(7, 7, "white"), "h1");
  assert.equal(squareAt(0, 0, "black"), "h1");
  assert.equal(squareAt(7, 7, "black"), "a8");
});
test("position broadcasts preserve seats but an explicit spectator snapshot clears them", () => {
  const seat = { kind: "state" as const, fen: initialFEN, turn: "w" as const, status: "", pgn: "", uci: [], lastSeen: 0, watchers: 1, role: "player" as const, color: "w" as const };
  const { role, color, ...broadcast } = seat;
  assert.equal(mergeState(seat, broadcast).role, "player");
  assert.equal(mergeState(seat, broadcast).color, "w");
  const spectator = mergeState(seat, { ...broadcast, role: "spectator", color: null });
  assert.equal(spectator.role, "spectator");
  assert.equal(spectator.color, null);
});
test("legal targets include en passant, castling and all four promotions", () => {
  assert.deepEqual(legalMoves(initialFEN, "e2").map((move) => move.to).sort(), ["e3", "e4"]);
  const chess = new Chess();
  for (const move of ["e4", "a6", "e5", "d5"]) chess.move(move);
  assert.ok(legalMoves(chess.fen(), "e5").some((move) => move.to === "d6" && move.isEnPassant()));
  assert.ok(legalMoves("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1", "e1").some((move) => move.to === "g1"));
  assert.deepEqual(legalMoves("7k/P7/8/8/8/8/8/7K w - - 0 1", "a7").map((move) => move.promotion).sort(), ["b", "n", "q", "r"]);
});
test("move history renders chess notation including checkmate", () => {
  assert.deepEqual(moveLabels(["f2f3", "e7e5", "g2g4", "d8h4"]), ["f3", "e5", "g4", "Qh4#"]);
});
