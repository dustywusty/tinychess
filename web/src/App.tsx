import { useEffect, useState } from "react";
import { Game } from "./pages/Game";
import { Home } from "./pages/Home";
import { Review } from "./pages/Review";

interface Route {
  kind: "home" | "game" | "review";
  gameId?: string;
}

function parseRoute(): Route {
  const path = window.location.pathname;
  if (path.length <= 1) return { kind: "home" };
  const segments = path.slice(1).split("/").filter(Boolean);
  if (segments[0] === "review" && segments[1]) {
    return { kind: "review", gameId: segments[1] };
  }
  if (segments[0] === "index.html" || segments[0] === "new") {
    return { kind: "home" };
  }
  return { kind: "game", gameId: segments[0] };
}

export default function App() {
  const [route, setRoute] = useState<Route>(() => parseRoute());

  useEffect(() => {
    const onPop = () => setRoute(parseRoute());
    window.addEventListener("popstate", onPop);
    return () => window.removeEventListener("popstate", onPop);
  }, []);

  if (route.kind === "review" && route.gameId) {
    return <Review gameId={route.gameId} />;
  }
  if (route.kind === "game" && route.gameId) {
    return <Game gameId={route.gameId} />;
  }
  return <Home />;
}
