import { useMemo, useState } from "react";
import {
  forgetGame,
  loadRecentGames,
  type RecentGameEntry,
} from "../../lib/recentGames";

export function RecentGames() {
  const [games, setGames] = useState(() => loadRecentGames());

  const sorted = useMemo<RecentGameEntry[]>(() => {
    return Object.values(games).sort((a, b) => b.lastSeen - a.lastSeen);
  }, [games]);

  if (sorted.length === 0) return null;

  const handleCopy = async (id: string) => {
    try {
      await navigator.clipboard.writeText(`${window.location.origin}/${id}`);
    } catch {
      /* ignore */
    }
  };

  const handleForget = (id: string) => {
    setGames(forgetGame(id));
  };

  return (
    <section className="w-full max-w-md text-left">
      <h2 className="text-sm font-medium opacity-70 mb-2">Recent games</h2>
      <ul className="space-y-2">
        {sorted.slice(0, 8).map((g) => (
          <li
            key={g.id}
            className="flex items-center justify-between gap-2 p-2 rounded-md bg-panel"
          >
            <a
              href={`/${g.id}`}
              className="font-mono text-xs truncate flex-1"
              title={g.id}
            >
              {g.id.slice(0, 8)}…
              {g.result && (
                <span className="ml-2 opacity-70">{g.result}</span>
              )}
            </a>
            <button
              type="button"
              onClick={() => void handleCopy(g.id)}
              className="text-xs px-2 py-1 rounded hover:bg-bg"
            >
              Copy link
            </button>
            <button
              type="button"
              onClick={() => handleForget(g.id)}
              className="text-xs px-2 py-1 rounded hover:bg-bg opacity-70"
            >
              Forget
            </button>
          </li>
        ))}
      </ul>
    </section>
  );
}
