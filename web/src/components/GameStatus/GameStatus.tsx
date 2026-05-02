import { useMemo } from "react";
import { useGameStore } from "../../state/gameStore";
import { useUiStore } from "../../state/uiStore";

interface Props {
  gameId?: string;
}

export function GameStatus({ gameId }: Props = {}) {
  const turn = useGameStore((s) => s.turn);
  const status = useGameStore((s) => s.status);
  const playerColor = useGameStore((s) => s.playerColor);
  const isSpectator = useGameStore((s) => s.isSpectator);
  const watchers = useGameStore((s) => s.watchers);
  const uiStatus = useUiStore((s) => s.status);
  const statusError = useUiStore((s) => s.statusError);

  const turnText = useMemo(() => {
    if (status) return status;
    if (isSpectator) {
      return turn === "white" ? "White to move" : "Black to move";
    }
    return turn === playerColor ? "Your turn" : "Their turn";
  }, [turn, status, playerColor, isSpectator]);

  // Mirror the legacy behavior: when the server reports a game-over status
  // (e.g. "Checkmate"), surface it in both #turn (as the turn indicator) and
  // #status (as the message line). Local UI errors take precedence in #status.
  const statusText = uiStatus || status || "";

  return (
    <div className="flex flex-col gap-1 items-start">
      <div id="turn" data-testid="turn" className="text-base font-medium">
        {turnText}
      </div>
      <div
        id="status"
        data-testid="status"
        className={`text-sm min-h-[1.25rem] ${statusError ? "text-[color:var(--err)]" : "opacity-70"}`}
      >
        {statusText}
      </div>
      {watchers > 0 && (
        <div className="text-xs opacity-50">
          {watchers === 1 ? "1 watcher" : `${watchers} watchers`}
        </div>
      )}
      {status && gameId && !isSpectator && (
        <a
          href={`/review/${gameId}`}
          className="mt-1 inline-flex items-center px-3 py-1.5 rounded-md bg-[color:var(--accent)] text-bg text-sm font-medium"
        >
          Review this game →
        </a>
      )}
    </div>
  );
}
