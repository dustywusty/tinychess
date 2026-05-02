import { useEffect, useState } from "react";
import { Game } from "./pages/Game";
import { Home } from "./pages/Home";

function currentGameId(): string {
  const path = window.location.pathname;
  if (path.length <= 1) return "";
  // Strip leading slash, ignore index.html, ignore /new (server-handled)
  const segment = path.slice(1).split("/")[0];
  if (!segment || segment === "index.html" || segment === "new") return "";
  return segment;
}

export default function App() {
  const [gameId, setGameId] = useState<string>(() => currentGameId());

  useEffect(() => {
    const onPop = () => setGameId(currentGameId());
    window.addEventListener("popstate", onPop);
    return () => window.removeEventListener("popstate", onPop);
  }, []);

  if (gameId) return <Game gameId={gameId} />;
  return <Home />;
}
