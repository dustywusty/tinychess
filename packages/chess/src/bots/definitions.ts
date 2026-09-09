import type { BotDefinition, BotId } from "./types.ts";

export const bots: readonly BotDefinition[] = [
 { id: "pip", name: "Pip", emoji: "🐣", description: "Just learning", maxLossCp: 350,
   profile: { excellent: 20, good: 25, inaccurate: 25, mistake: 20, blunder: 10 }, captureBias: .25, checkBias: .1, castleBias: 0,
   thinkTime: { minMs: 250, maxMs: 800 }, engine: { multiPv: 8, moveTimeMs: 220, depth: 5 } },
 { id: "max", name: "Max", emoji: "😎", description: "Casual player", maxLossCp: 220,
   profile: { excellent: 40, good: 30, inaccurate: 20, mistake: 8, blunder: 2 }, captureBias: .15, checkBias: .1, castleBias: .1,
   thinkTime: { minMs: 350, maxMs: 1000 }, engine: { multiPv: 7, moveTimeMs: 320, depth: 7 } },
 { id: "ada", name: "Ada", emoji: "🧠", description: "Club player", maxLossCp: 130,
   profile: { excellent: 65, good: 22, inaccurate: 10, mistake: 3, blunder: 0 }, captureBias: 0, checkBias: 0, castleBias: .2,
   thinkTime: { minMs: 500, maxMs: 1200 }, engine: { multiPv: 6, moveTimeMs: 500 } },
 { id: "viktor", name: "Viktor", emoji: "🦈", description: "Very strong", maxLossCp: 75,
   profile: { excellent: 85, good: 12, inaccurate: 3, mistake: 0, blunder: 0 }, captureBias: .2, checkBias: .35, castleBias: .05,
   thinkTime: { minMs: 600, maxMs: 1500 }, engine: { multiPv: 5, moveTimeMs: 700 } },
 { id: "machine", name: "The Machine", emoji: "☠️", description: "Good luck.", maxLossCp: 0,
   profile: { excellent: 100, good: 0, inaccurate: 0, mistake: 0, blunder: 0 }, captureBias: 0, checkBias: 0, castleBias: 0,
   thinkTime: { minMs: 250, maxMs: 900 }, engine: { multiPv: 1, moveTimeMs: 850 } },
];
export function botDefinition(id: BotId): BotDefinition {
 const result = bots.find(bot => bot.id === id);
 if (!result) throw new Error("Unknown computer opponent.");
 return result;
}
