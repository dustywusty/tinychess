import { useRef, useState } from "react";
import { createGame } from "../api/game";
import { Header } from "../components/Header/Header";
import { RecentGames } from "../components/RecentGames/RecentGames";
import { CoachCard } from "../components/CoachCard";
import { Board } from "../components/Board/Board";
import { ThemePicker } from "../components/ThemePicker/ThemePicker";
import { gameIDFromInput } from "../lib/invitation";
import { START_FEN } from "../types/chess";

export function Home() {
  const [busy, setBusy] = useState(false);
  const creating = useRef(false);
  const [error, setError] = useState("");
  const [input, setInput] = useState("");
  const id = gameIDFromInput(input);
  const handleNew = async () => {
    if (creating.current) return;
    creating.current = true; setBusy(true); setError("");
    try { window.location.assign("/g/" + (await createGame()).id); }
    catch { setError("Couldn’t start your game. Check your connection and try again."); creating.current = false; setBusy(false); }
  };
  return <main className="site-shell home-page">
    <Header />
    <div className="home-content">
      <section className="home-intro">
        <p className="eyebrow">A SMALL GAME. A GOOD TIME.</p>
        <h1>Good company.<br />Great moves.</h1>
        <p className="intro-copy">A little chess with your favorite people.<br />Send a link. Make your move.</p>
        <div className="intro-detail"><span aria-hidden="true">↗</span><p>Across the table or across the world.<br /><strong>There’s always room for a game.</strong></p></div>
      </section>
      <section className="play-card">
        <div className="section-line"><span className="eyebrow">YOUR NEXT GOOD GAME</span><span className="live-dot" /></div>
        <div className="board-art" aria-hidden="true">
          <div className="art-board"><Board fen={START_FEN} uci={[]} perspective="white" selected={null} disabled onSquareClick={() => {}} preview /></div>
          <span className="emoji-sticker sticker-wave">👋</span><span className="emoji-sticker sticker-think">🤔</span>
          <span className="art-caption">you + a friend</span>
        </div>
        <h2>Across the board.<br />Closer together.</h2>
        <p>No sign-up. No rush. Just chess.</p>
        <button id="newgame" type="button" className="primary-button" disabled={busy} onClick={() => void handleNew()}>{busy ? "Opening your board…" : <>Play a friend <span aria-hidden="true">↗</span></>}</button>
        {error && <p role="alert" className="error-message">{error}</p>}
      </section>
      <div className="home-secondary">
        <section className="join-card">
          <h2>Got an invite?</h2><p>A friendly match is a link away.</p>
          <form className="join-form" onSubmit={(event) => { event.preventDefault(); if (id && !busy) window.location.assign("/g/" + id); }}>
            <input aria-label="Game link or ID" placeholder="Paste a game link or ID" value={input} onChange={(event) => setInput(event.target.value)} autoCapitalize="none" autoCorrect="off" spellCheck={false} />
            <button type="submit" aria-label="Join game" disabled={!id || busy}>↗</button>
          </form>
          {!!input.trim() && !id && <p className="error-message" role="status">Use a game ID or a link ending in /g/your-game-id.</p>}
        </section>
        <RecentGames />
      </div>
      <div className="home-coach"><CoachCard /><div className="mood-row"><p>A board to match your mood.</p><ThemePicker /></div></div>
    </div>
    <footer className="site-footer"><span>64 squares. Endless possibilities.</span><span>Made for a little more together.</span></footer>
  </main>;
}
