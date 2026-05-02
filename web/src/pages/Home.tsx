import { useMemo } from "react";
import { createGame } from "../api/game";
import { Header } from "../components/Header/Header";
import { RecentGames } from "../components/RecentGames/RecentGames";
import { loadRecentGames } from "../lib/recentGames";

export function Home() {
  const activeGames = useMemo(() => {
    const games = loadRecentGames();
    return Object.values(games).filter((g) => !g.result && !g.status);
  }, []);

  const oldestActive = useMemo(() => {
    if (activeGames.length === 0) return null;
    return [...activeGames].sort((a, b) => a.createdAt - b.createdAt)[0];
  }, [activeGames]);

  const handleNew = async () => {
    if (oldestActive) {
      window.location.assign(`/${oldestActive.id}`);
      return;
    }
    try {
      const { id } = await createGame();
      window.location.assign(`/${id}`);
    } catch {
      window.location.assign("/new");
    }
  };

  return (
    <main className="min-h-screen bg-bg text-text flex flex-col">
      <Header />
      <section className="flex-1 flex flex-col items-center justify-start p-6 gap-6">
        <div className="max-w-md w-full text-center space-y-3">
          <h1 className="text-3xl font-semibold">Tiny Chess</h1>
          <p className="opacity-80 text-sm">
            Play a quick game with a friend. Share the URL after you start.
          </p>
          <button
            id="newgame"
            type="button"
            onClick={() => void handleNew()}
            className="px-4 py-3 rounded-md bg-[color:var(--accent)] text-bg font-medium"
          >
            {oldestActive ? "Open most recent" : "New game"}
          </button>
        </div>
        <RecentGames />
      </section>
    </main>
  );
}
