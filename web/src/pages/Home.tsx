import { useRef, useState } from "react";
import { createGame } from "../api/game";
import { Header } from "../components/Header/Header";
import { RecentGames } from "../components/RecentGames/RecentGames";
import { CoachCard } from "../components/CoachCard";
import { Board } from "../components/Board/Board";
import { gameIDFromInput } from "../lib/invitation";
import { START_FEN } from "../types/chess";
import { BotPicker } from "../components/BotPicker";
import { getOrCreateClientId } from "../lib/session";
import type { BotId } from "@yourmove/chess/bots";

export function Home() {
  const [busy, setBusy] = useState(false);
  const [computer, setComputer] = useState(false);
  const creating = useRef(false);
  const [error, setError] = useState("");
  const [input, setInput] = useState("");
  const id = gameIDFromInput(input);
  const handleNew = async (botId?: BotId, color: "w" | "b" = "w") => {
    if (creating.current) return;
    creating.current = true; setBusy(true); setError("");
    try { window.location.assign("/g/" + (await createGame(botId ? { botId, color, clientId: getOrCreateClientId() } : undefined)).id); }
    catch { setError("Couldn’t start your game. Check your connection and try again."); setComputer(false); creating.current = false; setBusy(false); }
  };
  return <main className="site-shell home-page">
    <Header />
    <div className="home-content">
      <section className="play-card">
        <div className="section-line"><span className="eyebrow">YOUR NEXT GOOD GAME</span><span className="live-dot" /></div>
        <div className="board-art" aria-hidden="true">
          <div className="art-board"><Board fen={START_FEN} uci={[]} perspective="white" selected={null} disabled onSquareClick={() => {}} preview /></div>
          <span className="emoji-sticker sticker-wave">👋</span><span className="emoji-sticker sticker-think">🤔</span>
          <span className="art-caption">you + a friend</span>
        </div>
        <button id="newgame" type="button" className="primary-button" disabled={busy} onClick={() => void handleNew()}>{busy ? "Opening your board…" : <>Play a friend <span aria-hidden="true">↗</span></>}</button>
        <button type="button" className="computer-button" disabled={busy} onClick={() => setComputer(true)}>Play the computer <span aria-hidden="true">✳</span></button>
        {error && <p role="alert" className="error-message">{error}</p>}
      </section>
      <div className="home-secondary">
        <section className="join-card">
          <h1>A little chess with your favorite people.</h1>
          <p>Send a link. Make your move.</p>
          <form className="join-form" onSubmit={(event) => { event.preventDefault(); if (id && !busy) window.location.assign("/g/" + id); }}>
            <input aria-label="Game link or ID" placeholder="Paste a game link or ID" value={input} onChange={(event) => setInput(event.target.value)} autoCapitalize="none" autoCorrect="off" spellCheck={false} />
            <button type="submit" aria-label="Join game" disabled={!id || busy}>↗</button>
          </form>
          {!!input.trim() && !id && <p className="error-message" role="status">Use a game ID or a link ending in /g/your-game-id.</p>}
        </section>
        <RecentGames />
      </div>
      <div className="home-coach"><CoachCard /></div>
    </div>
    <footer className="site-footer"><span>64 squares. Endless possibilities.</span><a href="/arasan/NOTICES.txt">Engine credits</a></footer>
    <BotPicker open={computer} busy={busy} onClose={() => setComputer(false)} onStart={(id, color) => void handleNew(id, color)} />
  </main>;
}
