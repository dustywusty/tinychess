import { useCallback, useRef, useState } from "react";
import { useFocusEffect } from "expo-router";
import { AppState } from "react-native";
import EventSource from "react-native-sse";
import { isEmojiEvent, isStateEvent, type EmojiEvent, type StateEvent } from "@yourmove/protocol";
import { apiURL } from "./api";
import { mergeState } from "./chess";
import { liveConnection } from "./liveConnection";
import { rememberGame } from "./recentGames";
import { clientID } from "./session";

export function useLiveGame(gameID: string) {
  const [state, setState] = useState<StateEvent | null>(null);
  const [cid, setCID] = useState("");
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState("");
  const [attempt, setAttempt] = useState(0);
  const [reactions, setReactions] = useState<EmojiEvent[]>([]);
  const stateRef = useRef<StateEvent | null>(null);
  const addReaction = useCallback((event: EmojiEvent) => {
    setReactions((previous) => [...previous, event].slice(-6));
  }, []);
  const accept = useCallback((event: StateEvent) => {
    const next = mergeState(stateRef.current, event);
    stateRef.current = next;
    setState(next);
    rememberGame(gameID, next);
  }, [gameID]);

  useFocusEffect(useCallback(() => {
    let active = true;
    let stop: (() => void) | undefined;
    stateRef.current = null;
    setState(null);
    setCID("");
    setReactions([]);
    setConnected(false);
    setError("");
    const start = async () => {
      try {
        const id = await clientID();
        if (!active || (AppState.currentState && AppState.currentState !== "active")) return;
        setCID(id);
        stop?.();
        stop = liveConnection(
          () => new EventSource(`${apiURL}/api/sse/${encodeURIComponent(gameID)}?clientId=${encodeURIComponent(id)}`, { pollingInterval: 0, timeoutBeforeConnection: 0 }),
          (data) => {
            if (isStateEvent(data)) { accept(data); setConnected(true); setError(""); }
            else if (isEmojiEvent(data) && data.sender !== id) addReaction(data);
          },
          () => { setConnected(false); setError("Connection lost. Reconnecting…"); },
        );
      } catch { if (active) setError("Couldn’t restore your player identity. Please try again."); }
    };
    void start();
    const subscription = AppState.addEventListener("change", (status) => {
      stop?.(); stop = undefined; setConnected(false);
      if (status === "active") void start();
    });
    return () => { active = false; stop?.(); subscription.remove(); };
  }, [gameID, attempt, accept, addReaction]));
  return { state, stateRef, cid, connected, error, reactions, addReaction, accept, retry: () => setAttempt((value) => value + 1) };
}
