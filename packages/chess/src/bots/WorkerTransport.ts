import type { EngineTransport } from "./types.ts";
export function workerTransport(url = "/arasan/worker.js"): EngineTransport {
 const worker = new Worker(url);
 return {
  send: command => worker.postMessage(command),
  subscribe(line, error) {
   worker.onmessage = event => { if (typeof event.data === "string") line(event.data); };
   worker.onerror = () => error(new Error("Computer engine failed to load."));
   return () => { worker.onmessage = null; worker.onerror = null; };
  },
  dispose: () => worker.terminate(),
 };
}
