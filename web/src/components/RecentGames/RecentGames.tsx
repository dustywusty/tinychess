import { useMemo, useState } from "react";
import { forgetGame, loadRecentGames, type RecentGameEntry } from "../../lib/recentGames";
import { Piece } from "../Piece";
import { bots } from "@yourmove/chess/bots";
import { resultBanner } from "@yourmove/chess";

export function RecentGames() {
  const [games, setGames] = useState(() => loadRecentGames());
  const [notice, setNotice] = useState("");
  const sorted = useMemo<RecentGameEntry[]>(() => Object.values(games).sort((a, b) => b.lastSeen - a.lastSeen), [games]);
  return <section className="recent-games">
    <h2 className="eyebrow">PICK UP WHERE YOU LEFT OFF</h2>
    {!sorted.length ? <div className="recent-empty">
      <span className="recent-empty-icon" aria-hidden="true"><Piece piece="n" size={36} /><span>…</span></span>
      <div><strong>No games yet.</strong><p>A lonely knight, waiting for your first move.</p></div>
    </div> : <ul>{sorted.slice(0, 8).map((game) => {
      const bot = bots.find((bot) => bot.id === game.botId);
      const result = resultBanner(game);
      return <li key={game.id} className={result ? `recent-finished result-${result.tone}` : undefined}>
      <div className="recent-game-row">
      <a href={"/g/" + game.id} className="recent-link"><span className="recent-icon"><Piece piece="n" size={28} /></span><span>
        <span className="recent-mode">{game.opponentType === "bot" ? "PvBot" : game.opponentType === "human" ? "PvP" : "Saved game"}</span>
        <strong>{game.opponentType === "bot" ? `A match with ${bot?.name ?? "the computer"}` : game.opponentType === "human" ? "Your friendly match" : "Your saved match"}</strong>
        <small>{result ? "Completed" : "In progress"} · {game.id.slice(0, 8)}</small>
      </span></a>
      <button type="button" className="text-button" aria-label={"Copy link for game " + game.id.slice(0, 8)} onClick={() => {
        void navigator.clipboard.writeText(window.location.origin + "/g/" + game.id).then(() => setNotice("Game link copied.")).catch(() => setNotice("Couldn’t copy the link. Open the game and copy its address."));
      }}>Copy</button>
      <button type="button" className="text-button" aria-label={"Forget game " + game.id.slice(0, 8)} onClick={() => setGames(forgetGame(game.id))}>×</button>
      </div>
      {result && <div className="recent-result" data-testid="recent-result">
        <span><span className="result-symbol" aria-hidden="true">{result.symbol}</span><span>{result.label}</span></span>
        <span className="result-score">{result.score}</span>
      </div>}
    </li>; })}</ul>}
    {notice && <p role="status" className="muted">{notice}</p>}
  </section>;
}
