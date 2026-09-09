import { ArasanEngine, workerTransport } from "@yourmove/chess/bots";
import { turnFromFEN } from "../board";

export interface EvalResult {
 fen: string;
 depth: number;
 /** Centipawn evaluation from White's perspective. */
 evalCp: number | null;
 /** Moves to mate from White's perspective, not plies. */
 mateIn: number | null;
 bestMove: string | null;
}
let engine: ArasanEngine | null = null;

// Compatibility entry point for position review. Bot games pass full history.
export async function evalPosition(fen: string, depth = 14): Promise<EvalResult> {
 engine ??= new ArasanEngine(() => workerTransport());
 const result = await engine.analyzePosition({ fen, initialFen: fen, moves: [], depth, multiPv: 1, moveTimeMs: 700 });
 const candidate = result.candidates[0];
 const sign = turnFromFEN(fen) === "white" ? 1 : -1;
 return { fen, depth: candidate?.depth ?? 0, evalCp: candidate?.scoreCp === undefined ? null : candidate.scoreCp * sign,
  mateIn: candidate?.mateIn === undefined ? null : candidate.mateIn * sign, bestMove: result.bestMove };
}
export function disposeEngine(): void { void engine?.dispose(); engine = null; }
