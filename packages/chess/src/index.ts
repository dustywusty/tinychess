import { Chess } from "chess.js";
export { gameResult, resultBanner, type GameResult, type ResultTone } from "./results.ts";

type Side = "white" | "black";
const order = "qrbnp";
const names: Record<string, string> = { q: "queen", r: "rook", b: "bishop", n: "knight", p: "pawn" };

// Replay captures rather than subtracting the current board from its starting
// material: promoted pieces and en passant need the actual move history.
export function capturedPieces(uci: readonly string[], startingFen?: string): Record<Side, string[]> {
  const captured: Record<Side, string[]> = { white: [], black: [] };
  const chess = new Chess(startingFen);
  for (const u of uci) {
    try {
      const move = chess.move({ from: u.slice(0, 2), to: u.slice(2, 4), promotion: u[4] });
      if (move.captured) {
        captured[move.color === "w" ? "white" : "black"].push(move.color === "w" ? move.captured : move.captured.toUpperCase());
      }
    } catch {
      // Keep only the valid prefix if an incomplete history arrives.
      break;
    }
  }
  for (const pieces of Object.values(captured)) pieces.sort((a, b) => order.indexOf(a.toLowerCase()) - order.indexOf(b.toLowerCase()));
  return captured;
}

export function capturedLabel(side: Side, pieces: readonly string[]): string {
  const counts = [...order].flatMap((piece) => {
    const count = pieces.filter((value) => value.toLowerCase() === piece).length;
    return count ? [`${count} ${side === "white" ? "black" : "white"} ${names[piece]}${count === 1 ? "" : "s"}`] : [];
  });
  return `Captured by ${side === "white" ? "White" : "Black"}: ${counts.join(", ") || "none"}`;
}
