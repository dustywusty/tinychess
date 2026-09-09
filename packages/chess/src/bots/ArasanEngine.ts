import { Chess } from "chess.js";
import { parseBestMove, parseInfo, validUci } from "./parser.ts";
import type { AnalyzePositionInput, ChessEngine, EngineAnalysis, EngineCandidate, EngineTransport } from "./types.ts";

export function abortError(): Error { const error = new Error("Analysis canceled."); error.name = "AbortError"; return error; }

export class ArasanEngine implements ChessEngine {
 private transport?: EngineTransport;
 private unsubscribe?: () => void;
 private ready?: Promise<void>;
 private queue: Promise<unknown> = Promise.resolve();
 private listeners = new Set<(line: string) => void>();
 private failures = new Set<(error: Error) => void>();
 private disposed = false;
 private readonly createTransport: () => EngineTransport;
 private readonly timeoutMs: number;
 constructor(createTransport: () => EngineTransport, timeoutMs = 15000) { this.createTransport = createTransport; this.timeoutMs = timeoutMs; }

 initialize(): Promise<void> {
  if (this.disposed) return Promise.reject(abortError());
  if (this.ready) return this.ready;
  this.ready = (async () => {
   this.transport = this.createTransport();
   this.unsubscribe = this.transport.subscribe(line => {
    if (line === "engine-error") { this.fail(new Error("Computer engine stopped.")); return; }
    for (const listener of [...this.listeners]) listener(line);
   }, error => this.fail(error));
   await this.waitFor(line => line === "uciok" ? true : undefined, () => this.send("uci"));
   this.send("setoption name Threads value 1");
   this.send("setoption name Hash value 16");
   this.send("setoption name OwnBook value false");
   this.send("setoption name Ponder value false");
   this.send("setoption name UCI_LimitStrength value false");
   await this.waitFor(line => line === "readyok" ? true : undefined, () => this.send("isready"));
  })().catch(error => { this.reset(); throw error; });
  return this.ready;
 }

 private send(command: string) { if (!this.transport) throw new Error("Computer engine is unavailable."); this.transport.send(command); }
 private fail(error: Error) { for (const fail of [...this.failures]) fail(error); this.reset(); }
 private reset() { this.unsubscribe?.(); this.transport?.dispose(); this.transport = undefined; this.ready = undefined; }
 private waitFor<T>(parse: (line: string) => T | undefined, start: () => void, signal?: AbortSignal, timeout = this.timeoutMs): Promise<T> {
  return new Promise((resolve, reject) => {
   const finish = (error?: Error, value?: T) => {
    clearTimeout(timer); this.listeners.delete(line); this.failures.delete(fail); signal?.removeEventListener("abort", abort);
    if (error) reject(error); else resolve(value as T);
   };
   const line = (value: string) => { const result = parse(value); if (result !== undefined) finish(undefined, result); };
   const fail = (error: Error) => finish(error);
   // Destroying the transport guarantees prior output cannot reach a new search.
   const abort = () => { finish(abortError()); this.reset(); };
   const timer = setTimeout(() => { finish(new Error("Computer engine timed out.")); this.reset(); }, timeout);
   this.listeners.add(line); this.failures.add(fail); signal?.addEventListener("abort", abort, { once: true });
   if (signal?.aborted) { abort(); return; }
   try { start(); } catch (error) { fail(error instanceof Error ? error : new Error("Engine command failed.")); }
  });
 }

 analyzePosition(input: AnalyzePositionInput): Promise<EngineAnalysis> {
  const work = this.queue.then(() => this.analyze(input));
  this.queue = work.catch(() => {});
  return work;
 }
 private async analyze(input: AnalyzePositionInput): Promise<EngineAnalysis> {
  if (input.signal?.aborted || this.disposed) throw abortError();
  if (input.moves.length > 2000 || !input.moves.every(validUci)) throw new Error("Invalid position history.");
  const board = new Chess(input.initialFen);
  for (const move of input.moves) board.move({ from: move.slice(0, 2), to: move.slice(2, 4), promotion: move[4] });
  // FEN en-passant fields differ between rule libraries. Compare legal EP form.
  if (board.fen() !== new Chess(input.fen).fen()) throw new Error("Position history does not match the board.");
  if (input.searchMoves && (!input.searchMoves.length || !input.searchMoves.every(validUci))) throw new Error("Invalid candidate moves.");
  await this.initialize();
  if (input.signal?.aborted || this.disposed) throw abortError();
  const multiPv = Math.min(10, Math.max(1, Math.round(input.multiPv)));
  const moveTime = Math.min(1500, Math.max(20, Math.round(input.moveTimeMs)));
  const byDepth = new Map<number, Map<number, EngineCandidate>>();
  this.send(`setoption name MultiPV value ${multiPv}`);
  this.send(`position ${input.initialFen ? "fen " + new Chess(input.initialFen).fen() : "startpos"}${input.moves.length ? " moves " + input.moves.join(" ") : ""}`);
  let stopTimer: ReturnType<typeof setTimeout> | undefined;
  try { return await this.waitFor(line => {
   const candidate = parseInfo(line);
   if (candidate) {
    const row = byDepth.get(candidate.depth) ?? new Map(); row.set(candidate.rank, candidate); byDepth.set(candidate.depth, row);
   }
   const bestMove = parseBestMove(line);
   if (bestMove === undefined) return undefined;
   // Prefer a completed MultiPV iteration; never compare scores across depths.
   const rows = [...byDepth.entries()].sort((a, b) => b[0] - a[0]);
   const expected = Math.min(multiPv, input.searchMoves?.length ?? board.moves().length);
   const row = rows.find(([, values]) => values.size >= expected) ?? rows[0];
   const candidates = [...(row?.[1].values() ?? [])].sort((a, b) => a.rank - b.rank);
   return { candidates, bestMove };
  }, () => {
   // Arasan treats depth and movetime as alternatives. Enforce the time ceiling
   // with stop as well when a weak bot uses a shallow depth ceiling.
   this.send(`go ${input.depth ? "depth " + Math.min(20, Math.max(1, Math.round(input.depth))) : "movetime " + moveTime}${input.searchMoves ? " searchmoves " + input.searchMoves.join(" ") : ""}`);
   stopTimer = setTimeout(() => this.transport?.send("stop"), moveTime);
  }, input.signal, Math.min(this.timeoutMs, moveTime + 5000)); }
  finally { clearTimeout(stopTimer); }
 }
 async dispose(): Promise<void> { this.disposed = true; this.fail(abortError()); await this.queue; }
}
