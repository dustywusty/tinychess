import { Worker } from "node:worker_threads";
import { Chess } from "../../packages/chess/node_modules/chess.js/dist/esm/chess.js";
import { ArasanEngine, botDefinition, candidateFeatures, compareCandidates, selectMove, seededRandom, type BotId } from "../../packages/chess/src/bots/index.ts";

const bot = botDefinition((process.argv[2] || "pip") as BotId);
const fen = new Chess(process.argv[3]).fen();
const engine = new ArasanEngine(() => {
 const worker = new Worker(new URL("./node-worker.cjs", import.meta.url));
 return { send: command => worker.postMessage(command), subscribe(line, error) { worker.on("message", line); worker.on("error", error); return () => { worker.off("message", line); worker.off("error", error); }; }, dispose: () => { void worker.terminate(); } };
});
try {
 const { candidates } = await engine.analyzePosition({fen, initialFen:fen, moves:[], ...bot.engine});
 const ordered = candidates.sort(compareCandidates);
 console.table(ordered.map(candidate => ({uci:candidate.uci, depth:candidate.depth, cp:candidate.scoreCp, mate:candidate.mateIn,
  loss:candidate.scoreCp === undefined || ordered[0].scoreCp === undefined ? "mate" : ordered[0].scoreCp - candidate.scoreCp,
  ...candidateFeatures(fen, candidate.uci)})));
 if (candidates.length) console.log("Selected", selectMove(bot, fen, candidates, seededRandom(Number(process.argv[4] || 42))));
 else console.log("No candidates: the position is terminal or the engine returned no analysis.");
} finally { await engine.dispose(); }
