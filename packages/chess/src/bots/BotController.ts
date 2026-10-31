import type { AnalyzePositionInput, BotDefinition, ChessEngine, RandomSource } from "./types.ts";
import { candidateFeatures, random, selectMove } from "./policy.ts";
import { abortError } from "./ArasanEngine.ts";
import { Chess } from "chess.js";

export async function chooseBotMove(engine: ChessEngine, bot: BotDefinition, position: Pick<AnalyzePositionInput, "fen" | "moves">, signal: AbortSignal, rng: RandomSource = random, moveDelayMs?: number, targetElo?: number) {
 const start = Date.now();
 const delay = moveDelayMs ?? bot.thinkTime.minMs + rng.next() * (bot.thinkTime.maxMs - bot.thinkTime.minMs);
 if (targetElo !== undefined) return pace(await eloMove(engine, position, signal, targetElo), signal, delay, start);
 let analysis = await engine.analyzePosition({ ...position, ...bot.engine, signal });
 // Top-N lines often contain no beginner mistakes. Re-score a small legal
 // sample beside the best lines in one search, with a common evaluation depth.
 if ((bot.id === "pip" || bot.id === "max") && !analysis.candidates.some(candidate => candidate.mateIn !== undefined)) {
  const anchors = analysis.candidates.slice(0, 3).map(candidate => candidate.uci);
  const rest = new Chess(position.fen).moves({ verbose: true }).map(move => move.from + move.to + (move.promotion ?? "")).filter(move => !anchors.includes(move));
  for (let index = rest.length - 1; index > 0; index--) {
   const other = Math.floor(rng.next() * (index + 1)); [rest[index], rest[other]] = [rest[other], rest[index]];
  }
  const searchMoves = [...anchors, ...rest.slice(0, 4)];
  if (searchMoves.length > 1) analysis = await engine.analyzePosition({ ...position, ...bot.engine, multiPv: searchMoves.length, searchMoves, signal });
 }
 return pace(selectMove(bot, position.fen, analysis.candidates, rng), signal, delay, start);
}

// Custom difficulty replaces the personality policy with Arasan's own
// strength limit. Arasan picks its weaker move itself, so trust bestmove.
async function eloMove(engine: ChessEngine, position: Pick<AnalyzePositionInput, "fen" | "moves">, signal: AbortSignal, elo: number) {
 const analysis = await engine.analyzePosition({ ...position, multiPv: 1, moveTimeMs: 500, elo, signal });
 const uci = analysis.bestMove && candidateFeatures(position.fen, analysis.bestMove) ? analysis.bestMove : analysis.candidates.find(candidate => candidateFeatures(position.fen, candidate.uci))?.uci;
 if (!uci) throw new Error("Computer engine returned no legal candidates.");
 return { uci };
}

async function pace<T>(selected: T, signal: AbortSignal, delay: number, start: number) {
 if (signal.aborted) throw abortError();
 await new Promise<void>((resolve, reject) => {
  const abort = () => { clearTimeout(timer); reject(abortError()); };
  const timer = setTimeout(() => { signal.removeEventListener("abort", abort); resolve(); }, Math.max(0, delay - (Date.now() - start)));
  signal.addEventListener("abort", abort, { once: true });
  if (signal.aborted) abort();
 });
 return selected;
}
