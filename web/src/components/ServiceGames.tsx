import { useEffect, useState } from "react";
import { resultBanner } from "@yourmove/chess";

type PublicGame = { id: string; result: string; moveCount: number; updatedAt: string };
type GamePage = { games: PublicGame[]; hasMore: boolean };

function GameList({ status }: { status: "active" | "completed" }) {
  const [offset, setOffset] = useState(0);
  const [page, setPage] = useState<GamePage>();
  const [error, setError] = useState(false);
  useEffect(() => {
    const controller = new AbortController();
    let pending = false;
    setPage(undefined);
    setError(false);
    const refresh = async () => {
      if (pending) return;
      pending = true;
      try {
        const response = await fetch(`/api/games?status=${status}&offset=${offset}`, { signal: controller.signal });
        if (!response.ok) throw new Error("Unable to load games");
        const data: GamePage = await response.json();
        if (!controller.signal.aborted) { setPage(data); setError(false); }
      } catch {
        if (!controller.signal.aborted) setError(true);
      } finally { pending = false; }
    };
    void refresh();
    const timer = window.setInterval(() => void refresh(), 30_000);
    window.addEventListener("focus", refresh);
    return () => {
      controller.abort();
      window.clearInterval(timer);
      window.removeEventListener("focus", refresh);
    };
  }, [status, offset]);
  const title = status === "active" ? "Active games" : "Completed games";
  return <section aria-label={title} className="service-game-list">
    <h3>{title}</h3>
    {error && <p role="status" className="muted">Couldn’t refresh games. Retrying shortly.</p>}
    {!page && !error && <p className="muted">Loading games…</p>}
    {page && !page.games.length && <p className="muted">{offset ? "No more games." : `No ${status} games yet.`}</p>}
    {!!page?.games.length && <ul>{page.games.map(game => <li key={game.id}>
      <a href={`/g/${game.id}`}>
        <span><strong>Game {game.id.slice(0, 8)}</strong><small>{Math.ceil(game.moveCount / 2)} moves · <time dateTime={game.updatedAt}>{new Date(game.updatedAt).toLocaleDateString()}</time></small></span>
        <span className="service-game-status">{status === "active" ? "In progress ↗" : resultBanner(game)?.label ?? "Completed"}</span>
      </a>
    </li>)}</ul>}
    {(offset > 0 || page?.hasMore) && <div className="service-pagination">
      <button type="button" className="text-button" disabled={offset === 0} onClick={() => setOffset(Math.max(0, offset - 8))}>Previous</button>
      <span>Page {offset / 8 + 1}</span>
      <button type="button" className="text-button" disabled={!page?.hasMore} onClick={() => setOffset(offset + 8)}>Next</button>
    </div>}
  </section>;
}

export function ServiceGames() {
  return <section className="service-games" aria-labelledby="service-games-heading">
    <h2 id="service-games-heading">Around the boards</h2>
    <p className="muted">Games across the service. Active games are unfinished with activity in the last 30 days. Updates every 30 seconds.</p>
    <div className="service-games-grid"><GameList status="active" /><GameList status="completed" /></div>
  </section>;
}
