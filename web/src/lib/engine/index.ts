import { turnFromFEN } from "../board";

export interface EvalResult {
  fen: string;
  depth: number;
  /** Centipawn evaluation from white's POV. null when the position is mate. */
  evalCp: number | null;
  /** Plies-to-mate from white's POV (positive = white mates, negative = black). null when not mate. */
  mateIn: number | null;
  /** Best move in UCI notation, e.g. "e2e4". null if engine never reported one. */
  bestMove: string | null;
}

interface PendingEval {
  fen: string;
  resolve: (result: EvalResult) => void;
  reject: (err: Error) => void;
  latestDepth: number;
  latestCp: number | null;
  latestMate: number | null;
  bestMove: string | null;
}

const STOCKFISH_URL = "/stockfish/stockfish-18-lite-single.js";

let workerPromise: Promise<Worker> | null = null;
const queue: Array<{ fen: string; depth: number; resolve: (r: EvalResult) => void; reject: (e: Error) => void }> = [];
let active: PendingEval | null = null;

function sideToMove(fen: string): "white" | "black" {
  return turnFromFEN(fen);
}

function loadWorker(): Promise<Worker> {
  if (workerPromise) return workerPromise;
  workerPromise = new Promise((resolve, reject) => {
    let worker: Worker;
    try {
      worker = new Worker(STOCKFISH_URL);
    } catch (err) {
      workerPromise = null;
      reject(err instanceof Error ? err : new Error("worker construction failed"));
      return;
    }
    let ready = false;
    worker.onmessage = (ev: MessageEvent<string>) => {
      const line = typeof ev.data === "string" ? ev.data : "";
      if (!ready) {
        if (line === "uciok") {
          worker.postMessage("isready");
          return;
        }
        if (line === "readyok") {
          ready = true;
          resolve(worker);
          return;
        }
        return;
      }
      handleEngineLine(worker, line);
    };
    worker.onerror = (err) => {
      if (!ready) {
        reject(new Error(`stockfish worker error: ${err.message}`));
        workerPromise = null;
      }
    };
    worker.postMessage("uci");
  });
  return workerPromise;
}

function handleEngineLine(worker: Worker, line: string) {
  if (!active) return;
  if (line.startsWith("info")) {
    const parts = line.split(/\s+/);
    let depth: number | null = null;
    let cp: number | null = null;
    let mate: number | null = null;
    for (let i = 0; i < parts.length; i++) {
      if (parts[i] === "depth") {
        depth = Number(parts[i + 1]);
      } else if (parts[i] === "score" && parts[i + 1] === "cp") {
        cp = Number(parts[i + 2]);
      } else if (parts[i] === "score" && parts[i + 1] === "mate") {
        mate = Number(parts[i + 2]);
      }
    }
    if (depth !== null) active.latestDepth = depth;
    if (cp !== null) active.latestCp = cp;
    if (mate !== null) active.latestMate = mate;
    return;
  }
  if (line.startsWith("bestmove")) {
    const parts = line.split(/\s+/);
    active.bestMove = parts[1] && parts[1] !== "(none)" ? parts[1] : null;
    finalize(worker);
  }
}

function finalize(worker: Worker) {
  if (!active) return;
  const stm = sideToMove(active.fen);
  const sign = stm === "white" ? 1 : -1;
  const result: EvalResult = {
    fen: active.fen,
    depth: active.latestDepth,
    evalCp: active.latestCp !== null ? active.latestCp * sign : null,
    mateIn: active.latestMate !== null ? active.latestMate * sign : null,
    bestMove: active.bestMove,
  };
  active.resolve(result);
  active = null;
  pump(worker);
}

function pump(worker: Worker) {
  if (active) return;
  const next = queue.shift();
  if (!next) return;
  active = {
    fen: next.fen,
    resolve: next.resolve,
    reject: next.reject,
    latestDepth: 0,
    latestCp: null,
    latestMate: null,
    bestMove: null,
  };
  worker.postMessage(`position fen ${next.fen}`);
  worker.postMessage(`go depth ${next.depth}`);
}

export async function evalPosition(fen: string, depth = 14): Promise<EvalResult> {
  const worker = await loadWorker();
  return new Promise<EvalResult>((resolve, reject) => {
    queue.push({ fen, depth, resolve, reject });
    pump(worker);
  });
}

/** Tear down the engine worker. Useful for tests and route changes. */
export function disposeEngine(): void {
  if (workerPromise) {
    void workerPromise.then((w) => w.terminate()).catch(() => {});
    workerPromise = null;
  }
  queue.length = 0;
  active = null;
}
