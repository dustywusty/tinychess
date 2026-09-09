import type { EngineCandidate } from "./types.ts";
export const validUci = (value: string) => /^[a-h][1-8][a-h][1-8][qrbn]?$/.test(value);

export function parseInfo(line: string): EngineCandidate | null {
 if (!line.startsWith("info ") || /\b(?:lowerbound|upperbound)\b/.test(line)) return null;
 const depth = line.match(/\bdepth (\d+)\b/);
 const rank = line.match(/\bmultipv (\d+)\b/);
 const score = line.match(/\bscore (cp|mate) (-?\d+)\b/);
 const pv = line.match(/\bpv (.+)$/)?.[1].trim().split(/\s+/);
 if (!depth || !score || !pv?.length || !pv.every(validUci)) return null;
 const value = Number(score[2]);
 if (!Number.isSafeInteger(value) || Number(rank?.[1] ?? 1) < 1) return null;
 return { uci: pv[0], pv, depth: Number(depth[1]), rank: Number(rank?.[1] ?? 1),
   ...(score[1] === "cp" ? { scoreCp: value } : { mateIn: value }) };
}
export function parseBestMove(line: string): string | null | undefined {
 const match = line.match(/^bestmove (\S+)(?:\s|$)/);
 if (!match) return undefined;
 if (match[1] === "0000" || match[1] === "(none)") return null;
 return validUci(match[1]) ? match[1] : undefined;
}
