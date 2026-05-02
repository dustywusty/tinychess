import { useCallback, useEffect, useState } from "react";
import { evalPosition, disposeEngine, type EvalResult } from "../lib/engine";

interface UseEngineState {
  ready: boolean;
  loading: boolean;
  error: Error | null;
  evalPosition: (fen: string, depth?: number) => Promise<EvalResult>;
}

/**
 * Lazy-loads the Stockfish-WASM worker on first call to evalPosition. Resolves
 * `ready` once the engine has handshook and is responsive. Disposes the worker
 * on unmount.
 */
export function useEngine(): UseEngineState {
  const [ready, setReady] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const wrappedEval = useCallback(async (fen: string, depth = 14) => {
    if (!ready) setLoading(true);
    try {
      const result = await evalPosition(fen, depth);
      if (!ready) {
        setReady(true);
        setLoading(false);
      }
      return result;
    } catch (err) {
      const e = err instanceof Error ? err : new Error("engine error");
      setError(e);
      setLoading(false);
      throw e;
    }
  }, [ready]);

  useEffect(() => {
    return () => {
      disposeEngine();
    };
  }, []);

  return { ready, loading, error, evalPosition: wrappedEval };
}
