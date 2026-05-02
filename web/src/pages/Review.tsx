import { useEffect, useMemo, useState } from "react";
import { HashbrownProvider } from "@hashbrownai/react";
import { Chess } from "chess.js";
import { Chat } from "../components/Chat/Chat";
import { Header } from "../components/Header/Header";
import { AnnotationLayer } from "../components/Board/AnnotationLayer";
import { Board } from "../components/Board/Board";
import { useAnnotationStore } from "../state/annotationStore";
import { START_FEN, type Color, type Square } from "../types/chess";

interface GameDTO {
  id: string;
  fen: string;
  pgn: string;
  result: string;
  startedAt: string;
  endedAt?: string;
  moveCount: number;
  whiteSession?: string;
  blackSession?: string;
}

export function Review({ gameId }: { gameId: string }) {
  const [game, setGame] = useState<GameDTO | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    fetch(`/api/games/${gameId}`)
      .then(async (res) => {
        if (!res.ok) throw new Error(`game fetch failed (${res.status})`);
        return (await res.json()) as GameDTO;
      })
      .then((dto) => {
        if (!cancelled) setGame(dto);
      })
      .catch((err) => {
        if (!cancelled) setLoadError(err.message);
      });
    return () => {
      cancelled = true;
    };
  }, [gameId]);

  // Derive per-ply FENs by replaying the PGN locally.
  const positions = useMemo<Record<number, string>>(() => {
    if (!game?.pgn) return { 0: START_FEN };
    const chess = new Chess();
    try {
      chess.loadPgn(game.pgn);
    } catch {
      return { 0: START_FEN };
    }
    const history = chess.history({ verbose: true });
    const out: Record<number, string> = { 0: START_FEN };
    const replay = new Chess();
    history.forEach((move, i) => {
      replay.move(move);
      out[i + 1] = replay.fen();
    });
    return out;
  }, [game?.pgn]);

  const activePly = useAnnotationStore((s) => s.activePly);
  const clear = useAnnotationStore((s) => s.clear);
  useEffect(() => () => clear(), [clear]);

  const fenForBoard =
    activePly !== null && positions[activePly]
      ? positions[activePly]
      : (game?.fen ?? START_FEN);

  const perspective: Color = "white"; // simple default; user-color customization deferred

  if (loadError) {
    return (
      <main className="min-h-screen bg-bg text-text">
        <Header />
        <section className="max-w-md mx-auto p-6 text-center text-sm opacity-80">
          Couldn't load this game for review: {loadError}
        </section>
      </main>
    );
  }

  if (!game) {
    return (
      <main className="min-h-screen bg-bg text-text">
        <Header />
        <section className="max-w-md mx-auto p-6 text-center text-sm opacity-60">
          Loading game…
        </section>
      </main>
    );
  }

  const initialPrompt =
    `I just finished playing this game. Walk me through the 2-4 most critical moments using InlineBoard components. Result: ${game.result || "unknown"}.`;

  return (
    <HashbrownProvider url="/api/coach/chat">
      <main className="min-h-screen bg-bg text-text">
        <Header />
        <section className="grid grid-cols-1 lg:grid-cols-2 gap-6 max-w-6xl mx-auto p-4">
          <div className="space-y-3">
            <h1 className="text-lg font-semibold">Review</h1>
            <div className="relative">
              <Board
                fen={fenForBoard}
                uci={[]}
                perspective={perspective}
                selected={null}
                disabled
                onSquareClick={() => {}}
              />
              <AnnotationLayer perspective={perspective} />
            </div>
            <p className="text-xs opacity-60">
              {activePly !== null
                ? `Showing position at ply ${activePly}`
                : "Final position"}
            </p>
          </div>
          <div className="lg:h-[80vh] flex flex-col rounded-md border border-[color:var(--btn-border,_rgba(255,255,255,0.1))] bg-panel">
            <Chat
              gameId={gameId}
              pgn={game.pgn}
              positions={positions}
              plyCount={game.moveCount}
              initialPrompt={initialPrompt}
            />
          </div>
        </section>
      </main>
    </HashbrownProvider>
  );
}

// Square unused in this file but kept exported via type imports above to
// keep AnnotationLayer's expectations explicit.
export type { Square };
