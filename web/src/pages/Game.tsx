import { useCallback, useEffect } from "react";
import { Board } from "../components/Board/Board";
import { GameStatus } from "../components/GameStatus/GameStatus";
import { ShareLink } from "../components/ShareLink/ShareLink";
import { postMove } from "../api/game";
import { subscribeSSE } from "../api/sse";
import { getOrCreateClientId } from "../lib/session";
import { recordGameSeen } from "../lib/recentGames";
import { useGameStore } from "../state/gameStore";
import { useUiStore } from "../state/uiStore";
import type { Square } from "../types/chess";

export function Game({ gameId }: { gameId: string }) {
  const fen = useGameStore((s) => s.fen);
  const uci = useGameStore((s) => s.uci);
  const turn = useGameStore((s) => s.turn);
  const status = useGameStore((s) => s.status);
  const playerColor = useGameStore((s) => s.playerColor);
  const isSpectator = useGameStore((s) => s.isSpectator);
  const clientId = useGameStore((s) => s.clientId);
  const applyServerState = useGameStore((s) => s.applyServerState);
  const setClientId = useGameStore((s) => s.setClientId);
  const reset = useGameStore((s) => s.reset);

  const selected = useUiStore((s) => s.selected);
  const selectSquare = useUiStore((s) => s.selectSquare);
  const setStatus = useUiStore((s) => s.setStatus);
  const clearStatus = useUiStore((s) => s.clearStatus);

  // SSE subscription lifecycle
  useEffect(() => {
    if (!gameId) return;
    reset();
    const cid = getOrCreateClientId();
    setClientId(cid);
    const sub = subscribeSSE(gameId, cid, {
      onState: (event) => {
        applyServerState(event);
        recordGameSeen(gameId, {
          status: event.status,
          lastSeen: event.lastSeen,
          pgn: event.pgn,
        });
      },
      onEmoji: () => {
        // PR 4 wires emoji animations
      },
    });
    return () => sub.close();
  }, [gameId, applyServerState, reset, setClientId]);

  const submitMove = useCallback(
    async (from: Square, to: Square) => {
      if (!clientId) return;
      clearStatus();
      const move = `${from}${to}`;
      try {
        const res = await postMove(gameId, move, clientId);
        if (!res.ok) {
          setStatus(`Illegal move: ${res.error ?? "unknown"}`, true);
        }
      } catch {
        setStatus("Network error", true);
      }
    },
    [clientId, gameId, clearStatus, setStatus],
  );

  const handleSquareClick = useCallback(
    (sq: Square) => {
      if (isSpectator || status) return;
      if (turn !== playerColor) return;
      if (!selected) {
        selectSquare(sq);
        return;
      }
      if (selected === sq) {
        selectSquare(null);
        return;
      }
      const from = selected;
      selectSquare(null);
      void submitMove(from, sq);
    },
    [isSpectator, status, turn, playerColor, selected, selectSquare, submitMove],
  );

  const handleDragMove = useCallback(
    (from: Square, to: Square) => {
      if (isSpectator || status) return;
      if (turn !== playerColor) return;
      selectSquare(null);
      void submitMove(from, to);
    },
    [isSpectator, status, turn, playerColor, selectSquare, submitMove],
  );

  const isMyTurn =
    !isSpectator && !status && playerColor !== null && turn === playerColor;
  const boardDisabled = isSpectator || !!status || !isMyTurn;
  const perspective = playerColor ?? "white";

  return (
    <main className="min-h-screen bg-bg text-text">
      <header className="px-4 py-3 border-b border-[color:var(--btn-border,_rgba(255,255,255,0.1))] flex items-center justify-between">
        <a href="/" className="text-base font-semibold">
          Tiny Chess
        </a>
        <div className="flex items-center gap-2">
          <ShareLink />
        </div>
      </header>
      <section className="max-w-md mx-auto p-4 space-y-4">
        <GameStatus />
        <Board
          fen={fen}
          uci={uci}
          perspective={perspective}
          selected={selected}
          disabled={boardDisabled}
          onSquareClick={handleSquareClick}
          onMove={handleDragMove}
        />
        <PgnPanel />
      </section>
    </main>
  );
}

function PgnPanel() {
  const pgn = useGameStore((s) => s.pgn);
  if (!pgn) return null;
  return (
    <pre
      id="pgn"
      data-testid="pgn"
      className="text-xs whitespace-pre-wrap opacity-80 bg-panel rounded-md p-3"
    >
      {pgn}
    </pre>
  );
}
