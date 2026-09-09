import assert from "node:assert/strict";
import { Worker } from "node:worker_threads";
import { ArasanEngine, botDefinition, chooseBotMove } from "../../packages/chess/src/bots/index.ts";
import { Chess } from "../../packages/chess/node_modules/chess.js/dist/esm/chess.js";
const engine = new ArasanEngine(() => {
 const worker = new Worker(new URL("./node-worker.cjs", import.meta.url));
 return { send: command => worker.postMessage(command), subscribe(line, error) { worker.on("message", line); worker.on("error", error); return () => { worker.off("message", line); worker.off("error", error); }; }, dispose: () => { void worker.terminate(); } };
});
try {
 const board = new Chess(); const moves: string[] = [];
 for (let ply = 0; ply < 6; ply++) {
  const start = Date.now();
  const selected = await chooseBotMove(engine, botDefinition(ply % 2 ? "pip" : "machine"), { fen: board.fen(), moves }, new AbortController().signal);
  board.move({from:selected.uci.slice(0,2), to:selected.uci.slice(2,4), promotion:selected.uci[4]}); moves.push(selected.uci);
  console.log({ ply, ...selected, ms:Date.now() - start });
 }
 const abort = new AbortController();
 const search = engine.analyzePosition({ fen:board.fen(), moves, multiPv:8, moveTimeMs:1500, signal:abort.signal });
 setTimeout(() => abort.abort(), 50);
 await assert.rejects(search, {name:"AbortError"});
 const recovered = await engine.analyzePosition({fen:board.fen(), moves, multiPv:2, moveTimeMs:100});
 assert.ok(recovered.candidates.length > 0);
 const mateHistory = ["f2f3", "e7e5", "g2g4"];
 const mateBoard = new Chess();
 for (const move of mateHistory) mateBoard.move({from:move.slice(0,2), to:move.slice(2,4)});
 const mate = await engine.analyzePosition({fen:mateBoard.fen(), moves:mateHistory, multiPv:1, moveTimeMs:200});
 assert.equal(mate.bestMove, "d8h4");
 assert.equal(mate.candidates[0].mateIn, 1);
 mateBoard.move({from:"d8", to:"h4"});
 assert.ok(mateBoard.isCheckmate());
 const terminal = await engine.analyzePosition({fen:mateBoard.fen(), moves:[...mateHistory, "d8h4"], multiPv:1, moveTimeMs:100});
 assert.equal(terminal.bestMove, null);
 console.log("Real Arasan WASM: legal moves, cancellation, and restart passed.");
} finally { await engine.dispose(); }
