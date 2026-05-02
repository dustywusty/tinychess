import { useMemo } from "react";
import { useGameStore } from "../../state/gameStore";
import { useUiStore } from "../../state/uiStore";

export function GameStatus() {
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
        {uiStatus}
      </div>
      {watchers > 0 && (
        <div className="text-xs opacity-50">
          {watchers === 1 ? "1 watcher" : `${watchers} watchers`}
        </div>
      )}
    </div>
  );
}
