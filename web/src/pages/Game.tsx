import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Chess, type Square as ChessSquare } from "chess.js";
import { capturedLabel, capturedPieces } from "@yourmove/chess";
import { Board } from "../components/Board/Board";
import { EmojiPicker } from "../components/EmojiPicker/EmojiPicker";
import { GameStatus } from "../components/GameStatus/GameStatus";
import { Header } from "../components/Header/Header";
import { ReactionsLayer } from "../components/Reactions/ReactionsLayer";
import { ShareLink } from "../components/ShareLink/ShareLink";
import { Piece } from "../components/Piece";
import { postMove, postReact } from "../api/game";
import { subscribeSSE } from "../api/sse";
import { useReactions } from "../hooks/useReactions";
import { getOrCreateClientId } from "../lib/session";
import { recordGameSeen } from "../lib/recentGames";
import { useGameStore } from "../state/gameStore";
import { useUiStore } from "../state/uiStore";
import type { Color, Square } from "../types/chess";
import { botDefinition } from "@yourmove/chess/bots";
import { useBotGame } from "../hooks/useBotGame";

import { useCoach } from "../coach/useCoach";
import { useMoveFeedback } from "../coach/useMoveFeedback";
import { CoachPanel } from "../coach/CoachPanel";

export function Game({ gameId }: { gameId: string }) {
  const { fen, uci, turn, status, playerColor, isSpectator, clientId, pgn, bot, applyServerState, setClientId, reset } = useGameStore();
  const { selected, selectSquare, setStatus, clearStatus } = useUiStore();
  const { active: reactions, show: showReaction } = useReactions();
  const [connected, setConnected] = useState(false);
  const computer = useBotGame(gameId, connected);
  const coach = useCoach(gameId, clientId);
  const coachEnabled = import.meta.env.VITE_COACH_ENABLED === "1" && !isSpectator && (!!bot || !!status);
  const [hideAnswers, setHideAnswers] = useState(false);
  const feedback = useMoveFeedback(gameId, clientId, uci, playerColor, coachEnabled && connected && !hideAnswers && !bot?.playerBotId);
  const lesson = coachEnabled ? coach.lesson : null;
  const reviewActive = !!lesson || coach.pending;
  const [exerciseSelected, setExerciseSelected] = useState<Square | null>(null);
  const viewFen = lesson?.fen ?? fen;
  const viewChess = useMemo(() => new Chess(viewFen), [viewFen]);
  useEffect(() => { setExerciseSelected(null); setPromotion(null); }, [viewFen, lesson?.id]);
  const opponent = bot ? botDefinition(bot.id) : null;
  const playerBot = bot?.playerBotId ? botDefinition(bot.playerBotId) : null;
  const [busy, setBusy] = useState(false);
  const moveLock = useRef(false);
  const [flipped, setFlipped] = useState(false);
  const [history, setHistory] = useState<{ emoji: string; self: boolean; id: number }[]>([]);
  const [promotion, setPromotion] = useState<{ from: Square; to: Square } | null>(null);
  const dialog = useRef<HTMLDialogElement>(null);
  const chess = useMemo(() => new Chess(fen), [fen]);
  const captured = useMemo(() => capturedPieces(uci), [uci]);
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
        recordGameSeen(gameId, event);
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
  const canMove = connected && !reviewActive && !playerBot && !isSpectator && !status && playerColor === turn && !busy;
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
    if (lesson?.mode === "exercising") {
      const move = viewChess.moves({ verbose: true }).find(m => m.from === from && m.to === to);
      if (move) { if (move.promotion) setPromotion({ from, to }); else void coach.command("attempt", { candidate: from + to }); }
      setExerciseSelected(null); return;
    }
    if (!canMove) return;
    const move = chess.moves({ square: from as ChessSquare, verbose: true }).find((candidate) => candidate.to === to);
    if (!move) { selectSquare(null); return; }
    if (move.promotion) setPromotion({ from, to });
    else void submitMove(from, to);
  };
  const handleSquare = (square: Square) => {
    if (lesson?.mode === "exercising" && !coach.pending) {
      if (viewChess.get(square as ChessSquare)?.color === viewChess.turn()) setExerciseSelected(square);
      else if (exerciseSelected) attemptMove(exerciseSelected, square);
      return;
    }
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
    <div className="player-info"><strong>{bot && bot.color === (side === "white" ? "w" : "b") ? `${opponent?.emoji} ${opponent?.name}` : playerBot ? `${playerBot.emoji} ${playerBot.name}` : isSpectator ? (side === "white" ? "White" : "Black") : playerColor === side ? "You" : "Your friend"}</strong><small>{side === "white" ? "White pieces" : "Black pieces"}</small>
      {captured[side].length > 0 && <div className="captured-pieces" role="img" aria-label={capturedLabel(side, captured[side])} data-testid={"captured-" + side}>
        {captured[side].map((piece, index) => <span className="captured-piece" key={index}><Piece piece={piece} size={20} /></span>)}
      </div>}
    </div>
    {!lesson && connected && !status && turn === side && <span className="turn-badge"><span className="live-dot" />TO MOVE</span>}
  </div>;

  return <main className="site-shell game-page">
    <Header rightSlot={<ShareLink />} />
    <div className="game-layout">
      <section className="game-main">
        {player(perspective === "white" ? "black" : "white")}
        {lesson && <div className="board-review-context" aria-live="polite">
          <div><span className="eyebrow">{lesson.currentPosition ? `ANALYSIS · ${lesson.ply % 2 ? "BLACK" : "WHITE"} TO MOVE` : `REVIEW · MOVE ${Math.floor(lesson.ply / 2) + 1}`}</span><strong>{lesson.mode === "exercising" ? "Your turn to try" : lesson.step === 0 ? lesson.currentPosition ? "Position being analyzed" : "Before the move" : lesson.step === 1 ? lesson.branch === "alternative" ? `After ${lesson.recommendation?.move ?? lesson.assessment?.best ?? "the alternative"}` : `After ${lesson.assessment?.played ?? "your move"}` : `Possible continuation · ${lesson.step - 1} ${lesson.step === 2 ? "reply" : "replies"}`}</strong></div>
          {!hideAnswers && lesson.assessment && lesson.mode !== "exercising" && <div className="board-review-legend">{lesson.branch === "compare" ? <><span data-tone="played">Your move · {lesson.assessment.played}</span>{lesson.assessment.played !== lesson.assessment.best ? <span data-tone="alternative">Alternative · {lesson.assessment.best}</span> : <small>Also the best move</small>}</> : <span data-tone={lesson.branch === "alternative" ? "alternative" : "played"}>{lesson.branch === "alternative" ? "Alternative" : lesson.branch === "explored" ? "Your exploration" : "Your move"}</span>}</div>}
        </div>}
        <Board showRoutes={!lesson || lesson.step === 0} fen={viewFen} uci={lesson ? [] : uci} arrows={hideAnswers ? [] : lesson?.step === 0 && lesson.branch !== "compare" && lesson.reveal === 3 && lesson.mode !== "exercising" ? lesson.views?.[lesson.branch]?.[1]?.arrows : lesson?.arrows} perspective={perspective} selected={lesson ? exerciseSelected : selected} disabled={lesson?.mode === "exercising" ? coach.pending : !canMove} onSquareClick={handleSquare} onMove={attemptMove} />
        {player(perspective)}
        <div className="board-tools"><button className="text-button" type="button" onClick={() => setFlipped((value) => !value)}>↻ Flip board</button><span className="muted">{lesson && !hideAnswers && lesson.arrows.length > 0 ? "Dashed: from · Solid: to" : `${uci.length} moves`}</span></div>
      </section>
      <aside className="game-sidebar">

        <GameStatus connected={connected} thinkingLabel={computer.thinking && computer.activeBotId ? `${botDefinition(computer.activeBotId).name} is thinking…` : ""} />
        {computer.error && <div role="alert" className="error-message">{computer.error} <button type="button" className="text-button" onClick={computer.retry}>Try again</button></div>}
        {coachEnabled && <CoachPanel lesson={lesson} pending={coach.pending} error={coach.error} moves={uci} playerColor={playerColor} command={coach.command} feedback={feedback} hidden={hideAnswers} setHidden={setHideAnswers} openLesson={coach.openLesson} />}
        <details open className="reaction-card sidebar-widget" aria-label="Reactions">
          <summary className="section-line"><h2 className="eyebrow">A LITTLE BACK & FORTH</h2><span className="widget-chevron" aria-hidden="true">⌄</span></summary>
          <p>Say it with an emoji.</p>
          <EmojiPicker disabled={!connected} onSend={sendReaction} />
          {history.length > 0 && <div className="reaction-history" aria-live="polite">{history.map((item, index) => <span className={"reaction-bubble" + (item.self ? " from-self" : "")} key={item.id + "-" + index}><span>{item.emoji}</span>{item.self ? "You" : "Them"}</span>)}</div>}
        </details>
        <details className="moves-panel sidebar-widget" open><summary>The game so far <span>{uci.length} <span className="widget-chevron" aria-hidden="true">⌄</span></span></summary>
          {uci.length ? <pre id="pgn" data-testid="pgn">{pgn}</pre> : <p>{playerBot ? "Waiting for the opening move." : "The first move is yours to make."}</p>}
        </details>
        <p className="sidebar-note">{playerBot ? `Bot battle · ${(bot?.moveDelayMs ?? 1500) / 1000} s minimum between moves. Keep the host’s game open to continue. Friends can watch with your game link.` : bot ? "Your opponent plays on your device. Friends can watch with your game link." : !status && uci.length < 2 ? "Send your friend the game link and meet at the board." : "A little less scrolling. A little more chess."}</p>
      </aside>
    </div>
    <footer className="site-footer"><a href="/">← Back to your games</a><span>64 squares. Endless possibilities.</span></footer>
    <ReactionsLayer reactions={reactions} />
    <dialog className="promotion-dialog" ref={dialog} onClose={() => setPromotion(null)}>
      <h2>A pawn with possibilities.</h2><p>Choose your promotion.</p>
      <div className="promotion-options">{["q", "r", "b", "n"].map((piece, index) => <button type="button" key={piece} aria-label={"Promote to " + ["queen", "rook", "bishop", "knight"][index]}
        onClick={() => { if (promotion) { if (lesson?.mode === "exercising") { void coach.command("attempt", { candidate: promotion.from + promotion.to + piece }); setPromotion(null); } else void submitMove(promotion.from, promotion.to, piece); } }}><Piece piece={(lesson ? viewChess.turn() === "w" : playerColor === "white") ? piece.toUpperCase() : piece} size={48} /></button>)}</div>
      <button type="button" className="primary-button" onClick={() => setPromotion(null)}>Cancel</button>
    </dialog>
  </main>;
}
