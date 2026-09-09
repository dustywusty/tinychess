export type GameResult = "1-0" | "0-1" | "1/2-1/2";
export type ResultTone = "win" | "loss" | "draw" | "neutral";

export function gameResult(status = "", pgn = ""): GameResult | undefined {
  // The server's terminal status starts with the result. Otherwise use only
  // the final PGN token, never a score inside a tag or a comment.
  return (status.match(/^(1-0|0-1|1\/2-1\/2)(?:\s|$)/)?.[1]
    ?? pgn.trim().match(/(?:^|\s)(1-0|0-1|1\/2-1\/2)$/)?.[1]) as GameResult | undefined;
}

export function resultBanner(game: {
  result?: string;
  status?: string;
  playerColor?: string | null;
  role?: string;
}): { tone: ResultTone; label: string; symbol: string; score: GameResult } | undefined {
  const score = gameResult(game.result) ?? gameResult(game.status);
  if (!score) return;
  if (score === "1/2-1/2") return { tone: "draw", label: "Draw", symbol: "=", score };
  const winner = score === "1-0" ? "white" : "black";
  const side = game.playerColor === "w" || game.playerColor === "white" ? "white"
    : game.playerColor === "b" || game.playerColor === "black" ? "black" : null;
  if (game.role !== "player" || !side) {
    return { tone: "neutral", label: winner === "white" ? "White wins" : "Black wins", symbol: "⚑", score };
  }
  return side === winner ? { tone: "win", label: "Win", symbol: "✦", score }
    : { tone: "loss", label: "Loss", symbol: "↘", score };
}
