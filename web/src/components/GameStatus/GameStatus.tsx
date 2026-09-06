import { useGameStore } from "../../state/gameStore";
import { useUiStore } from "../../state/uiStore";

export function GameStatus({ connected }: { connected: boolean }) {
  const { turn, status, playerColor, isSpectator, lastSeen } = useGameStore();
  const { status: uiStatus, statusError } = useUiStore();
  const turnText = !lastSeen ? "Finding your board…" : !connected ? "Reconnecting…" : status || (isSpectator ? (turn === "white" ? "White to move" : "Black to move") : turn === playerColor ? "Your turn" : "Their turn");
  return <div className="game-heading">
    <div className="section-line"><span className="eyebrow">A FRIENDLY MATCH</span><span className="connection-label"><span className={"live-dot" + (!connected ? " offline-dot" : "")} />{connected ? "LIVE" : "CONNECTING"}</span></div>
    <h1 id="turn" data-testid="turn" aria-live="polite">{turnText}</h1>
    <p>{!lastSeen ? "We’re saving a seat for you." : isSpectator ? "The seats are full. You can watch and react." : status ? "Every game has something to teach us." : turn === playerColor ? "Take your time. Make it a good one." : "Good things come to those who wait."}</p>
    <div id="status" data-testid="status" role={statusError ? "alert" : "status"} className={statusError ? "error-message" : "game-message"}>{uiStatus || status}</div>
  </div>;
}
