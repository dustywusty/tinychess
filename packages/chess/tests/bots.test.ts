import test from "node:test";
import assert from "node:assert/strict";
import { Chess } from "chess.js";
import { ArasanEngine, botDefinition, parseInfo, parseBestMove, selectMove, seededRandom, candidateFeatures, chooseBotMove } from "../src/bots/index.ts";
import type { EngineCandidate, EngineTransport } from "../src/bots/types.ts";
const fen = new Chess().fen();
const candidates: EngineCandidate[] = ["e2e4", "d2d4", "g1f3", "b1c3", "a2a3"].map((uci, i) => ({ uci, scoreCp: -[0, 20, 60, 150, 400][i], pv: [uci], depth: 5, rank: i + 1 }));

test("UCI parser accepts cp, mate, and MultiPV; rejects malformed/bound lines", () => {
 assert.deepEqual(parseInfo("info depth 7 multipv 2 score cp -32 nodes 100 pv e2e4 e7e5"), { uci: "e2e4", pv:["e2e4", "e7e5"], scoreCp:-32, depth:7, rank:2 });
 assert.equal(parseInfo("info depth 4 score mate -3 pv e2e4")?.mateIn, -3);
 for (const line of ["info string hi", "info depth x score cp 3 pv e2e4", "info depth 4 score cp 3 lowerbound pv e2e4", "info depth 4 score cp 3 pv nope"]) assert.equal(parseInfo(line), null);
 assert.equal(parseBestMove("bestmove e7e8n ponder a1a2"), "e7e8n");
 assert.equal(parseBestMove("bestmove (none)"), null);
 assert.equal(parseBestMove("bestmove invalid"), undefined);
});
test("policy is seeded and honors strength caps", () => {
 const a = seededRandom(42), b = seededRandom(42);
 for (let i = 0; i < 200; i++) assert.deepEqual(selectMove(botDefinition("pip"), fen, candidates, a), selectMove(botDefinition("pip"), fen, candidates, b));
 const rng = seededRandom(99); let weaker = 0;
 for (let i = 0; i < 500; i++) {
  const pip = selectMove(botDefinition("pip"), fen, candidates, rng); assert.ok(pip.loss <= 350); if (pip.loss >= 60) weaker++;
  assert.ok(selectMove(botDefinition("viktor"), fen, candidates, rng).loss <= 75);
  assert.equal(selectMove(botDefinition("machine"), fen, candidates, rng).uci, "e2e4");
 }
 assert.ok(weaker > 100);
});
test("mate outranks cp and cannot be discarded for personality", () => {
 const mate = {...candidates[1], scoreCp:undefined, mateIn:1};
 assert.equal(selectMove(botDefinition("pip"), fen, [...candidates, mate]).uci, mate.uci);
 assert.equal(selectMove(botDefinition("pip"), fen, [{...mate, mateIn:-1}, {...candidates[0], scoreCp:undefined, mateIn:-5}]).uci, "e2e4");
 assert.throws(() => selectMove(botDefinition("pip"), fen, [{...candidates[0], uci:"e2e5"}]));
});
test("candidate features use existing chess rules", () => {
 assert.equal(candidateFeatures(fen, "e2e5"), null);
 assert.equal(candidateFeatures("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1", "e1g1")?.castles, true);
 assert.equal(candidateFeatures("7k/P7/8/8/8/8/8/K7 w - - 0 1", "a7a8q")?.isPromotion, true);
 const capture = candidateFeatures("k7/8/8/8/8/8/r7/R6K w - - 0 1", "a1a2");
 assert.equal(capture?.isCapture, true); assert.equal(capture?.givesCheck, true);
});

class Fake implements EngineTransport {
 commands: string[] = []; stopped = false;
 line: (line: string) => void = () => {};
 send(command: string) { this.commands.push(command); queueMicrotask(() => {
  if (command === "uci") this.line("uciok");
  if (command === "isready") this.line("readyok");
 }); }
 subscribe(line: (line: string) => void) { this.line = line; return () => { this.line = () => {}; }; }
 dispose() { this.stopped = true; }
}
const tick = () => new Promise(resolve => setTimeout(resolve, 10));
test("adapter normalizes MultiPV ordering and queues searches", async () => {
 const fake = new Fake(); const engine = new ArasanEngine(() => fake);
 const first = engine.analyzePosition({ fen, moves:[], multiPv:2, moveTimeMs:100 });
 const second = engine.analyzePosition({ fen, moves:[], multiPv:1, moveTimeMs:100 });
 await tick(); assert.equal(fake.commands.filter(c => c.startsWith("go ")).length, 1);
 fake.line("info depth 3 multipv 2 score cp 10 pv d2d4");
 fake.line("info depth 3 multipv 1 score cp 20 pv e2e4");
 fake.line("info depth 4 multipv 1 score cp 25 pv e2e4");
 fake.line("bestmove e2e4");
 assert.deepEqual((await first).candidates.map(c => c.uci), ["e2e4", "d2d4"]);
 await tick(); fake.line("info depth 2 score cp 1 pv e2e4"); fake.line("bestmove e2e4"); await second;
 await engine.dispose(); assert.ok(fake.stopped);
});
test("cancellation destroys transport and rejects stale output", async () => {
 const transports: Fake[] = [];
 const engine = new ArasanEngine(() => { const fake = new Fake(); transports.push(fake); return fake; });
 const controller = new AbortController();
 const work = engine.analyzePosition({ fen, moves:[], multiPv:1, moveTimeMs:100, signal:controller.signal });
 const rejected = assert.rejects(work, {name:"AbortError"});
 await tick(); controller.abort(); await rejected; assert.ok(transports[0].stopped);
 const next = engine.analyzePosition({ fen, moves:[], multiPv:1, moveTimeMs:100 });
 await tick(); transports[0].line("bestmove a2a3");
 transports[1].line("info depth 1 score cp 1 pv e2e4"); transports[1].line("bestmove e2e4");
 assert.equal((await next).bestMove, "e2e4"); await engine.dispose();
});
test("readiness timeout and missing history fail closed", async () => {
 const fake = new Fake(); fake.send = () => {};
 const engine = new ArasanEngine(() => fake, 20);
 await assert.rejects(engine.initialize(), /timed out/); assert.ok(fake.stopped); await engine.dispose();
 const board = new Chess(); board.move("e4");
 await assert.rejects(new ArasanEngine(() => new Fake()).analyzePosition({ fen:board.fen(), moves:[], multiPv:1, moveTimeMs:100 }), /history/);
});
test("controller cancellation also cancels presentation delay", async () => {
 const abort = new AbortController();
 const promise = chooseBotMove({ initialize:async () => {}, dispose:async () => {}, analyzePosition:async () => ({candidates, bestMove:"e2e4"}) }, botDefinition("ada"), {fen, moves:[]}, abort.signal);
 const rejected = assert.rejects(promise, {name:"AbortError"}); await tick(); abort.abort(); await rejected;
});
test("search timeout rejects the result and destroys the stalled runtime", async () => {
 const fake = new Fake(); const engine = new ArasanEngine(() => fake, 30);
 await engine.initialize();
 await assert.rejects(engine.analyzePosition({fen, moves:[], multiPv:1, moveTimeMs:20}), /timed out/);
 assert.ok(fake.stopped); await engine.dispose();
});
test("engine errors reject a pending search instead of leaving it unresolved", async () => {
 const fake = new Fake(); const engine = new ArasanEngine(() => fake);
 const work = engine.analyzePosition({fen, moves:[], multiPv:1, moveTimeMs:100});
 const rejected = assert.rejects(work, /stopped/);
 await tick(); fake.line("engine-error"); await rejected;
 assert.ok(fake.stopped); await engine.dispose();
});
