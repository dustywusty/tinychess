import { useEffect, useState } from "react";

type GameCounts = { active: number; completed: number };

export function ServiceGames() {
  const [counts, setCounts] = useState<GameCounts>();
  const [error, setError] = useState(false);
  useEffect(() => {
    const controller = new AbortController();
    let pending = false;
    const refresh = async () => {
      if (pending) return;
      pending = true;
      try {
        const response = await fetch("/api/stats/games", { signal: controller.signal });
        if (!response.ok) throw new Error("Unable to load game counts");
        const data: GameCounts = await response.json();
        if (!Number.isSafeInteger(data.active) || data.active < 0 || !Number.isSafeInteger(data.completed) || data.completed < 0) {
          throw new Error("Invalid game counts");
        }
        if (!controller.signal.aborted) { setCounts(data); setError(false); }
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
  }, []);
  return <section className="service-games" aria-labelledby="service-games-heading">
    <h2 id="service-games-heading">Around the boards</h2>
    <p className="muted">Across the service. Active games are unfinished with activity in the last 30 days.</p>
    {error && <p role="status" className="muted">Couldn’t refresh counts. Retrying shortly.</p>}
    {!counts && !error && <p className="muted">Loading game counts…</p>}
    {counts && <dl className="service-game-counts">
      <div><dt>Active games</dt><dd>{counts.active.toLocaleString()}</dd></div>
      <div><dt>Completed games</dt><dd>{counts.completed.toLocaleString()}</dd></div>
    </dl>}
  </section>;
}
