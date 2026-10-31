// IDs are stored with saved games. Rename bots through their definitions only.
export type BotId = "pip" | "max" | "ada" | "viktor" | "machine";
// Arasan accepts up to 3450, but targets above ~2200 play at nearly full
// strength, so the useful range stops at 2400.
export const MIN_BOT_ELO = 1000;
export const MAX_BOT_ELO = 2400;
export type Quality = "excellent" | "good" | "inaccurate" | "mistake" | "blunder";
export interface RandomSource { next(): number }
export interface BotDefinition {
 id: BotId; name: string; emoji: string; description: string;
 maxLossCp: number; profile: Record<Quality, number>;
 captureBias: number; checkBias: number; castleBias: number;
 thinkTime: { minMs: number; maxMs: number };
 engine: { multiPv: number; moveTimeMs: number; depth?: number };
}
/** Scores use the root side-to-move perspective. Mate is in moves, not plies. */
export interface EngineCandidate { uci: string; scoreCp?: number; mateIn?: number; pv: string[]; depth: number; rank: number }
export interface EngineAnalysis { candidates: EngineCandidate[]; bestMove: string | null }
export interface AnalyzePositionInput {
 initialFen?: string;
 fen: string; moves: readonly string[]; multiPv: number; moveTimeMs: number; depth?: number;
 /** Limits Arasan with UCI_Elo. Omit for full strength. */
 elo?: number;
 searchMoves?: string[]; signal?: AbortSignal;
}
export interface ChessEngine {
 initialize(): Promise<void>;
 analyzePosition(input: AnalyzePositionInput): Promise<EngineAnalysis>;
 dispose(): Promise<void>;
}
/** Transport is the only platform-specific part of the adapter. */
export interface EngineTransport {
 send(command: string): void;
 subscribe(line: (line: string) => void, error: (error: Error) => void): () => void;
 dispose(): void;
}
// Newest policy this client supports. Version 2 adds custom Elo targets.
export const BOT_POLICY_VERSION = 2;
