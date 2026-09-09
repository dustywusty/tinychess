import type { AnalyzePositionInput, BotDefinition, ChessEngine, RandomSource } from "./types.ts";
import { random, selectMove } from "./policy.ts";
import { abortError } from "./ArasanEngine.ts";
import { Chess } from "chess.js";

export async function chooseBotMove(engine: ChessEngine, bot: BotDefinition, position: Pick<AnalyzePositionInput, "fen" | "moves">, signal: AbortSignal, rng: RandomSource = random) {
 const start = Date.now();
 const delay = bot.thinkTime.minMs + rng.next() * (bot.thinkTime.maxMs - bot.thinkTime.minMs);
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
 const selected = selectMove(bot, position.fen, analysis.candidates, rng);
 if (signal.aborted) throw abortError();
 await new Promise<void>((resolve, reject) => {
  const abort = () => { clearTimeout(timer); reject(abortError()); };
  const timer = setTimeout(() => { signal.removeEventListener("abort", abort); resolve(); }, Math.max(0, delay - (Date.now() - start)));
  signal.addEventListener("abort", abort, { once: true });
  if (signal.aborted) abort();
 });
 return selected;
}
