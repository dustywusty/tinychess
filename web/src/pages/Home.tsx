import { createGame } from "../api/game";

export function Home() {
  const handleNew = async () => {
    try {
      const { id } = await createGame();
      window.location.assign(`/${id}`);
    } catch {
      window.location.assign("/new");
    }
  };

  return (
    <main className="min-h-screen flex items-center justify-center bg-bg text-text">
      <div className="max-w-md w-full text-center space-y-4 p-6">
        <h1 className="text-3xl font-semibold">Tiny Chess</h1>
        <p className="opacity-80 text-sm">
          Play a quick game with a friend. Share the URL after you start.
        </p>
        <button
          id="newgame"
          type="button"
          onClick={handleNew}
          className="px-4 py-3 rounded-md bg-[color:var(--accent)] text-bg font-medium"
        >
          New game
        </button>
      </div>
    </main>
  );
}
