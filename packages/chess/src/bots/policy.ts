import { Chess } from "chess.js";
import type { BotDefinition, EngineCandidate, Quality, RandomSource } from "./types.ts";
import { validUci } from "./parser.ts";

export const qualities: Quality[] = ["excellent", "good", "inaccurate", "mistake", "blunder"];
export function qualityFor(loss: number): Quality { return loss <= 20 ? "excellent" : loss <= 50 ? "good" : loss <= 100 ? "inaccurate" : loss <= 200 ? "mistake" : "blunder"; }
export const random: RandomSource = { next: () => Math.random() };
export function seededRandom(seed: number): RandomSource {
 return { next: () => { seed = (Math.imul(seed, 1664525) + 1013904223) >>> 0; return seed / 4294967296; } };
}
// Mate is ordered explicitly: win sooner, lose later. It is never treated as cp.
export function compareCandidates(a: EngineCandidate, b: EngineCandidate): number {
 const tier = (c: EngineCandidate) => c.mateIn === undefined ? 1 : c.mateIn > 0 ? 2 : 0;
 if (tier(a) !== tier(b)) return tier(b) - tier(a);
 if (a.mateIn !== undefined && b.mateIn !== undefined) return a.mateIn - b.mateIn;
 return (b.scoreCp ?? -Infinity) - (a.scoreCp ?? -Infinity);
}
export function candidateFeatures(fen: string, uci: string) {
 if (!validUci(uci)) return null;
 try {
  const board = new Chess(fen);
  const move = board.move({ from: uci.slice(0, 2), to: uci.slice(2, 4), promotion: uci[4] });
  return { isCapture: !!move.captured, givesCheck: board.isCheck(), castles: move.isKingsideCastle() || move.isQueensideCastle(), isPromotion: !!move.promotion };
 } catch { return null; }
}
export function selectMove(bot: BotDefinition, fen: string, candidates: EngineCandidate[], rng: RandomSource = random) {
 const legal = candidates.filter(candidate => (candidate.scoreCp !== undefined || candidate.mateIn !== undefined) && candidateFeatures(fen, candidate.uci)).sort(compareCandidates);
 if (!legal.length) throw new Error("Computer engine returned no legal candidates.");
 const best = legal[0];
 if (bot.id === "machine" || best.mateIn !== undefined) return { uci: best.uci, bucket: "excellent" as Quality, loss: 0 };
 const scored = legal.filter(candidate => candidate.mateIn === undefined).map(candidate => ({ candidate, loss: Math.max(0, best.scoreCp! - candidate.scoreCp!) }))
  .filter(entry => entry.loss <= bot.maxLossCp);
 let draw = rng.next() * Object.values(bot.profile).reduce((a, b) => a + b, 0);
 let bucket: Quality = "excellent";
 for (const quality of qualities) { draw -= bot.profile[quality]; if (draw < 0) { bucket = quality; break; } }
 let pool = scored.filter(entry => qualityFor(entry.loss) === bucket);
 // Missing bucket: prefer the nearest better class, never a bigger mistake.
 if (!pool.length) {
  const permitted = scored.filter(entry => qualities.indexOf(qualityFor(entry.loss)) < qualities.indexOf(bucket));
  const fallback = permitted.at(-1) ?? scored[0];
  pool = scored.filter(entry => qualityFor(entry.loss) === qualityFor(fallback.loss));
 }
 const weights = pool.map(entry => {
  const feature = candidateFeatures(fen, entry.candidate.uci)!;
  return 1 + Number(feature.isCapture) * bot.captureBias + Number(feature.givesCheck) * bot.checkBias + Number(feature.castles) * bot.castleBias;
 });
 let choice = rng.next() * weights.reduce((a, b) => a + b, 0);
 let selected = pool[pool.length - 1];
 for (let i = 0; i < pool.length; i++) { choice -= weights[i]; if (choice < 0) { selected = pool[i]; break; } }
 return { uci: selected.candidate.uci, bucket: qualityFor(selected.loss), loss: selected.loss };
}
