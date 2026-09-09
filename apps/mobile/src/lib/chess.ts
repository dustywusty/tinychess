import { Chess, type Square } from "chess.js";
import type { StateEvent, WireColor } from "@yourmove/protocol";

export const initialFEN = new Chess().fen();
export type Color = "white" | "black";
export function color(value: WireColor | null | undefined): Color | null {
  return value === "w" || value === "white" ? "white" : value === "b" || value === "black" ? "black" : null;
}
export function squareAt(row: number, column: number, perspective: Color): Square {
  return `${String.fromCharCode(97 + (perspective === "white" ? column : 7 - column))}${perspective === "white" ? 8 - row : row + 1}` as Square;
}
// Broadcasts omit the recipient's seat metadata.
export function mergeState(previous: StateEvent | null, incoming: StateEvent): StateEvent {
  return { ...previous, ...incoming, uci: incoming.uci ?? [] };
}
export function gameIDFromInput(input: string): string {
  const value = input.trim();
  if (/^[a-zA-Z0-9_-]{1,100}$/.test(value)) return value;
  try {
    const url = new URL(value);
    if (!["https:", "http:", "yourmove:"].includes(url.protocol)) return "";
    const path = url.protocol === "yourmove:" ? `/${url.host}${url.pathname}` : url.pathname;
    return path.match(/^\/g\/([a-zA-Z0-9_-]{1,100})\/?$/)?.[1] ?? "";
  } catch { return ""; }
}
export function legalMoves(fen: string, square: string) {
  return new Chess(fen).moves({ square: square as Square, verbose: true });
}
export function moveLabels(uci: string[]): string[] {
  const chess = new Chess();
  try {
    return uci.map((move) => chess.move({ from: move.slice(0, 2), to: move.slice(2, 4), promotion: move[4] }).san);
  } catch { return uci; }
}
