import { useEffect, useRef, useState } from "react";
import { ArasanEngine, BOT_POLICY_VERSION, botDefinition, chooseBotMove, workerTransport } from "@yourmove/chess/bots";
import { useGameStore } from "../state/gameStore";
import { postMove } from "../api/game";

export function useBotGame(gameId: string, connected: boolean) {
 const { bot, fen, uci, status, playerColor, isSpectator, clientId } = useGameStore();
 const [thinking, setThinking] = useState(false);
 const [error, setError] = useState("");
 const [attempt, setAttempt] = useState(0);
 const [visible, setVisible] = useState(!document.hidden);
 const engine = useRef<ArasanEngine | null>(null);
 useEffect(() => { const change = () => setVisible(!document.hidden); document.addEventListener("visibilitychange", change); return () => document.removeEventListener("visibilitychange", change); }, []);
 useEffect(() => () => { void engine.current?.dispose(); engine.current = null; }, [gameId]);
 const historyKey = uci.join(" ");
 const activeBotId = bot?.color === fen.split(" ")[1] ? bot.id : bot?.playerBotId;
 useEffect(() => {
  setThinking(false); setError("");
  if (!bot || !activeBotId || !connected || !visible || isSpectator || !playerColor || !clientId || status) return;
  if (bot.policyVersion !== BOT_POLICY_VERSION) { setError("Update the app to play this computer opponent."); return; }
  const abort = new AbortController();
  const run = async () => {
   setThinking(true);
   try {
    engine.current ??= new ArasanEngine(() => workerTransport());
    let selected;
    for (let tries = 0; tries < 2; tries++) {
     try { selected = await chooseBotMove(engine.current, botDefinition(activeBotId), { fen, moves: historyKey ? historyKey.split(" ") : [] }, abort.signal, undefined, bot.moveDelayMs); break; }
     catch (error) { if (abort.signal.aborted || tries) throw error; await engine.current.dispose(); engine.current = new ArasanEngine(() => workerTransport()); }
    }
    const current = useGameStore.getState();
    if (!selected || abort.signal.aborted || current.fen !== fen || current.uci.length !== uci.length || current.status) return;
    const result = await postMove(gameId, selected.uci, clientId, { botMove: true, expectedPly: uci.length });
    if (abort.signal.aborted) return;
    if (result.state && result.state.uci.length >= useGameStore.getState().uci.length) current.applyServerState({ ...result.state, kind: "state" });
    if (!result.ok && result.error !== "position changed") throw new Error(result.error);
   } catch { if (!abort.signal.aborted) setError("The computer could not move. Try again."); }
   finally { if (!abort.signal.aborted) setThinking(false); }
  };
  void run();
  return () => { abort.abort(); };
 }, [gameId, activeBotId, bot?.color, bot?.policyVersion, bot?.moveDelayMs, connected, visible, isSpectator, playerColor, clientId, fen, historyKey, status, attempt]);
 return { thinking, activeBotId, error, retry: () => setAttempt(value => value + 1) };
}
