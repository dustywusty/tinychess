import { useCallback, useEffect, useRef, useState } from "react";
import { AppState, View } from "react-native";
import { useFocusEffect } from "expo-router";
import type { StateEvent } from "@yourmove/protocol";
import { ArasanEngine, BOT_POLICY_VERSION, botDefinition, chooseBotMove, type EngineTransport } from "@yourmove/chess/bots";
import { makeMove } from "@/lib/api";
import { Button, ErrorMessage } from "./UI";
import { EngineHost } from "./EngineHost";

export function BotTurn({ gameID, state, cid, connected, accept, onThinkingChange }: { gameID: string; state: StateEvent; cid: string; connected: boolean; accept: (state: StateEvent) => void; onThinkingChange: (thinking: boolean) => void }) {
 const [focused, setFocused] = useState(false);
 useFocusEffect(useCallback(() => { setFocused(true); return () => setFocused(false); }, []));
 const [active, setActive] = useState(AppState.currentState === "active");
 const [generation, setGeneration] = useState(0);
 const [transport, setTransport] = useState<EngineTransport | null>(null);
 const [thinking, setThinking] = useState(false);
 useEffect(() => { onThinkingChange(thinking); return () => onThinkingChange(false); }, [thinking, onThinkingChange]);
 const [error, setError] = useState("");
 const engine = useRef<ArasanEngine | null>(null);
 const retries = useRef(0);
 const current = useRef(state); current.current = state;
 const onReady = useCallback((value: EngineTransport) => setTransport(value), []);
 const historyKey = state.uci.join(" ");
 useEffect(() => { retries.current = 0; }, [historyKey, gameID]);
 const eligible = focused && active && connected && state.role === "player" && !!cid && !state.status && state.bot?.color === state.fen.split(" ")[1];
 useEffect(() => { const listener = AppState.addEventListener("change", value => setActive(value === "active")); return () => listener.remove(); }, []);
 useEffect(() => {
  if (!transport) return;
  const value = new ArasanEngine(() => transport);
  engine.current = value;
  return () => { void value.dispose(); if (engine.current === value) engine.current = null; };
 }, [transport]);
 useEffect(() => {
  setThinking(false);
  if (!eligible || !transport || !engine.current || !state.bot) return;
  if (state.bot.policyVersion !== BOT_POLICY_VERSION) { setError("Update the app to play this computer opponent."); return; }
  const abort = new AbortController();
  const run = async () => {
   setThinking(true); setError("");
   try {
    const selected = await chooseBotMove(engine.current!, botDefinition(state.bot!.id), { fen: state.fen, moves: state.uci }, abort.signal);
    if (abort.signal.aborted || current.current.fen !== state.fen || current.current.uci.length !== state.uci.length) return;
    const response = await makeMove(gameID, selected.uci, cid, { botMove: true, expectedPly: state.uci.length });
    if (abort.signal.aborted) return;
    if (response.state && response.state.uci.length >= current.current.uci.length) accept({ ...response.state, kind: "state" });
    if (!response.ok && response.error !== "position changed") throw new Error(response.error);
   } catch {
    if (!abort.signal.aborted) {
     if (retries.current++ === 0) { setTransport(null); setGeneration(value => value + 1); }
     else setError("The computer could not move. Try again.");
    }
   }
   finally { if (!abort.signal.aborted) setThinking(false); }
  };
  void run();
  return () => abort.abort();
 }, [eligible, transport, gameID, cid, state.fen, historyKey, state.bot?.id, accept]);
 // Remount the isolated runtime after backgrounding or cancellation.
 useEffect(() => { if (!active || !focused || !connected) { setTransport(null); setGeneration(value => value + 1); } }, [active, focused, connected]);
 return <View style={{ gap: 8 }}>
  {focused && active && connected && <EngineHost key={generation} onReady={onReady} />}
  {!!error && <><ErrorMessage>{error}</ErrorMessage><Button title="Try again" onPress={() => { retries.current = 0; setTransport(null); setGeneration(value => value + 1); setError(""); }} /></>}
 </View>;
}
