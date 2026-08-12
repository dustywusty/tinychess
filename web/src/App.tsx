import { useEffect, useState } from "react";
import { Game } from "./pages/Game";
import { Home } from "./pages/Home";

function currentGameId(): string {
  const path = window.location.pathname;
  if (path.length <= 1) return "";
  const segments = path.slice(1).split("/").filter(Boolean);
  if (segments[0] === "g" && segments[1]) return segments[1];
  // Keep legacy /:gameId links working while shared links move to /g/:gameId.
  const segment = segments[0];
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
