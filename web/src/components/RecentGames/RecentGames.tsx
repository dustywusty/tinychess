import { useMemo, useState } from "react";
import { forgetGame, loadRecentGames, type RecentGameEntry } from "../../lib/recentGames";
import { Piece } from "../Piece";

export function RecentGames() {
  const [games, setGames] = useState(() => loadRecentGames());
  const [notice, setNotice] = useState("");
  const sorted = useMemo<RecentGameEntry[]>(() => Object.values(games).sort((a, b) => b.lastSeen - a.lastSeen), [games]);
  if (!sorted.length) return null;
  return <section className="recent-games">
    <h2 className="eyebrow">PICK UP WHERE YOU LEFT OFF</h2>
    <ul>{sorted.slice(0, 8).map((game) => <li key={game.id}>
      <a href={"/g/" + game.id} className="recent-link"><span className="recent-icon"><Piece piece="n" size={28} /></span><span><strong>{game.result ? "A game well played" : "Your friendly match"}</strong><small>{game.result || "In progress"} · {game.id.slice(0, 8)}</small></span></a>
      <button type="button" className="text-button" aria-label={"Copy link for game " + game.id.slice(0, 8)} onClick={() => {
        void navigator.clipboard.writeText(window.location.origin + "/g/" + game.id).then(() => setNotice("Game link copied.")).catch(() => setNotice("Couldn’t copy the link. Open the game and copy its address."));
      }}>Copy</button>
      <button type="button" className="text-button" aria-label={"Forget game " + game.id.slice(0, 8)} onClick={() => setGames(forgetGame(game.id))}>×</button>
    </li>)}</ul>
    {notice && <p role="status" className="muted">{notice}</p>}
  </section>;
}
