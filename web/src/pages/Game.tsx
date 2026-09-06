import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Chess, type Square as ChessSquare } from "chess.js";
import { Board } from "../components/Board/Board";
import { EmojiPicker } from "../components/EmojiPicker/EmojiPicker";
import { GameStatus } from "../components/GameStatus/GameStatus";
import { Header } from "../components/Header/Header";
import { ReactionsLayer } from "../components/Reactions/ReactionsLayer";
import { ShareLink } from "../components/ShareLink/ShareLink";
import { CoachCard } from "../components/CoachCard";
import { Piece } from "../components/Piece";
import { postMove, postReact } from "../api/game";
import { subscribeSSE } from "../api/sse";
import { useReactions } from "../hooks/useReactions";
import { getOrCreateClientId } from "../lib/session";
import { recordGameSeen } from "../lib/recentGames";
import { useGameStore } from "../state/gameStore";
import { useUiStore } from "../state/uiStore";
import type { Color, Square } from "../types/chess";

export function Game({ gameId }: { gameId: string }) {
  const { fen, uci, turn, status, playerColor, isSpectator, clientId, pgn, applyServerState, setClientId, reset } = useGameStore();
  const { selected, selectSquare, setStatus, clearStatus } = useUiStore();
  const { active: reactions, show: showReaction } = useReactions();
  const [connected, setConnected] = useState(false);
  const [busy, setBusy] = useState(false);
  const moveLock = useRef(false);
  const [flipped, setFlipped] = useState(false);
  const [history, setHistory] = useState<{ emoji: string; self: boolean; id: number }[]>([]);
  const [promotion, setPromotion] = useState<{ from: Square; to: Square } | null>(null);
  const dialog = useRef<HTMLDialogElement>(null);
  const chess = useMemo(() => new Chess(fen), [fen]);
  const recordReaction = useCallback((emoji: string, self: boolean) => {
    setHistory((previous) => [...previous, { emoji, self, id: Date.now() }].slice(-6));
    showReaction(emoji, self ? "self" : "remote");
  }, [showReaction]);
  useEffect(() => {
    reset(); clearStatus(); setConnected(false); setHistory([]);
    const cid = getOrCreateClientId();
    setClientId(cid);
    const sub = subscribeSSE(gameId, cid, {
      onState: (event) => {
        applyServerState(event); setConnected(true);
        recordGameSeen(gameId, { status: event.status, lastSeen: event.lastSeen, pgn: event.pgn });
      },
      onError: () => setConnected(false),
      onEmoji: (event) => { if (event.sender !== cid) recordReaction(event.emoji, false); },
    });
    return () => sub.close();
  }, [gameId, applyServerState, reset, setClientId, clearStatus, recordReaction]);
  useEffect(() => { selectSquare(null); setPromotion(null); }, [fen, connected, selectSquare]);
  useEffect(() => {
    if (promotion && !dialog.current?.open) dialog.current?.showModal();
    else if (!promotion && dialog.current?.open) dialog.current.close();
  }, [promotion]);
  const canMove = connected && !isSpectator && !status && playerColor === turn && !busy;
  const perspective: Color = flipped ? (playerColor === "black" ? "white" : "black") : playerColor ?? "white";
  const submitMove = async (from: Square, to: Square, promote = "") => {
    if (!canMove || moveLock.current) return;
    moveLock.current = true; setBusy(true); clearStatus(); selectSquare(null); setPromotion(null);
    try {
      const result = await postMove(gameId, from + to + promote, clientId);
      if (!result.ok) setStatus("Move not accepted: " + (result.error ?? "Try another square."), true);
    } catch { setStatus("Your move couldn’t be confirmed. Check your connection before trying again.", true); }
    finally { moveLock.current = false; setBusy(false); }
  };
  const attemptMove = (from: Square, to: Square) => {
    if (!canMove) return;
    const move = chess.moves({ square: from as ChessSquare, verbose: true }).find((candidate) => candidate.to === to);
    if (!move) { selectSquare(null); return; }
    if (move.promotion) setPromotion({ from, to });
    else void submitMove(from, to);
  };
  const handleSquare = (square: Square) => {
    if (!canMove) return;
    if (square === selected) { selectSquare(null); return; }
    const piece = chess.get(square as ChessSquare);
    if (piece && piece.color === (playerColor === "white" ? "w" : "b")) { selectSquare(square); return; }
    if (selected) attemptMove(selected, square);
  };
  const sendReaction = async (emoji: string) => {
    if (!connected || !clientId) return false;
    try {
      const result = await postReact(gameId, emoji, clientId);
      if (result.ok) { recordReaction(emoji, true); return true; }
      else setStatus(result.error ?? "That reaction didn’t send.", true);
    } catch { setStatus("That reaction didn’t send. Check your connection and try again.", true); }
    return false;
  };
  const player = (side: Color) => <div className="player-row">
    <span className="player-avatar"><Piece piece={side === "white" ? "K" : "k"} size={30} /></span>
    <div><strong>{isSpectator ? (side === "white" ? "White" : "Black") : playerColor === side ? "You" : "Your friend"}</strong><small>{side === "white" ? "White pieces" : "Black pieces"}</small></div>
    {connected && !status && turn === side && <span className="turn-badge"><span className="live-dot" />TO MOVE</span>}
  </div>;

  return <main className="site-shell game-page">
    <Header rightSlot={<ShareLink />} />
    <div className="game-layout">
      <section className="game-main">
        <GameStatus connected={connected} />
        {player(perspective === "white" ? "black" : "white")}
        <Board fen={fen} uci={uci} perspective={perspective} selected={selected} disabled={!canMove} onSquareClick={handleSquare} onMove={attemptMove} />
        {player(perspective)}
        <div className="board-tools"><button className="text-button" type="button" onClick={() => setFlipped((value) => !value)}>↻ Flip board</button><span className="muted">{uci.length} moves</span></div>
      </section>
      <aside className="game-sidebar">
        <section className="reaction-card">
          <div className="section-line"><h2 className="eyebrow">A LITTLE BACK & FORTH</h2><span aria-hidden="true">↗</span></div>
          <p>Say it with an emoji.</p>
          <EmojiPicker disabled={!connected} onSend={sendReaction} />
          {history.length > 0 && <div className="reaction-history" aria-live="polite">{history.map((item, index) => <span className={"reaction-bubble" + (item.self ? " from-self" : "")} key={item.id + "-" + index}><span>{item.emoji}</span>{item.self ? "You" : "Them"}</span>)}</div>}
        </section>
        <CoachCard />
        <details className="moves-panel" open><summary>The game so far <span>{uci.length}</span></summary>
          {uci.length ? <pre id="pgn" data-testid="pgn">{pgn}</pre> : <p>The first move is yours to make.</p>}
        </details>
        <p className="sidebar-note">{!status && uci.length < 2 ? "Send your friend the game link and meet at the board." : "A little less scrolling. A little more chess."}</p>
      </aside>
    </div>
    <footer className="site-footer"><a href="/">← Back to your games</a><span>64 squares. Endless possibilities.</span></footer>
    <ReactionsLayer reactions={reactions} />
    <dialog className="promotion-dialog" ref={dialog} onClose={() => setPromotion(null)}>
      <h2>A pawn with possibilities.</h2><p>Choose your promotion.</p>
      <div className="promotion-options">{["q", "r", "b", "n"].map((piece, index) => <button type="button" key={piece} aria-label={"Promote to " + ["queen", "rook", "bishop", "knight"][index]}
        onClick={() => { if (promotion) void submitMove(promotion.from, promotion.to, piece); }}><Piece piece={playerColor === "white" ? piece.toUpperCase() : piece} size={48} /></button>)}</div>
      <button type="button" className="primary-button" onClick={() => setPromotion(null)}>Cancel</button>
    </dialog>
  </main>;
}
