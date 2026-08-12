import { useCallback, useEffect, useMemo, useState } from "react";
import { useLocalSearchParams } from "expo-router";
import { Pressable, SafeAreaView, Share, StyleSheet, Text, View } from "react-native";
import type { StateEvent, WireColor } from "@yourmove/protocol";
import { ChessBoard } from "@/components/ChessBoard";
import { getSnapshot, makeMove, reactToGame } from "@/lib/api";
import { clientID } from "@/lib/session";

const reactions = ["👍", "😂", "😬", "🤔", "🔥", "👏"];

function color(value: WireColor | null | undefined): "white" | "black" | null {
  if (value === "w" || value === "white") return "white";
  if (value === "b" || value === "black") return "black";
  return null;
}

export default function GameScreen() {
  const params = useLocalSearchParams<{ id: string | string[] }>();
  const gameID = Array.isArray(params.id) ? params.id[0] ?? "" : params.id ?? "";
  const cid = useMemo(() => clientID(), []);
  const [state, setState] = useState<StateEvent | null>(null);
  const [selected, setSelected] = useState<string | null>(null);
  const [error, setError] = useState("");

  const refresh = useCallback(async () => {
    if (!gameID) return;
    try {
      setState(await getSnapshot(gameID, cid));
      setError("");
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Connection failed");
    }
  }, [cid, gameID]);

  useEffect(() => {
    void refresh();
    const timer = setInterval(() => void refresh(), 1500);
    return () => clearInterval(timer);
  }, [refresh]);

  const playerColor = color(state?.color);
  const turn = color(state?.turn);
  const canMove = !!state && state.role === "player" && !state.status && playerColor === turn;

  const handleSquare = async (square: string) => {
    if (!canMove) return;
    if (!selected) {
      setSelected(square);
      return;
    }
    if (selected === square) {
      setSelected(null);
      return;
    }
    const uci = `${selected}${square}`;
    setSelected(null);
    try {
      const response = await makeMove(gameID, uci, cid);
      if (!response.ok) setError(response.error ?? "Move rejected");
      await refresh();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Move failed");
    }
  };

  const share = () => {
    const root = process.env.EXPO_PUBLIC_WEB_URL ?? "https://yourmove.example";
    void Share.share({ message: `${root}/g/${gameID}` });
  };

  return (
    <SafeAreaView style={styles.safe}>
      <View style={styles.container}>
        <View style={styles.statusRow}>
          <View>
            <Text style={styles.turn}>
              {!state ? "Connecting…" : state.status || (state.role === "spectator" ? "Watching" : canMove ? "Your move" : "Their move")}
            </Text>
            <Text style={styles.role}>{state?.role === "player" ? `Playing ${playerColor}` : "Spectator"}</Text>
          </View>
          <Pressable style={styles.share} onPress={share}><Text style={styles.shareText}>Share</Text></Pressable>
        </View>
        <ChessBoard
          fen={state?.fen ?? "8/8/8/8/8/8/8/8 w - - 0 1"}
          perspective={playerColor ?? "white"}
          selected={selected}
          disabled={!canMove}
          onSquare={(square) => void handleSquare(square)}
        />
        <View style={styles.reactions}>
          {reactions.map((emoji) => (
            <Pressable key={emoji} style={styles.reaction} onPress={() => void reactToGame(gameID, emoji, cid)}>
              <Text style={styles.emoji}>{emoji}</Text>
            </Pressable>
          ))}
        </View>
        {!!error && <Text style={styles.error}>{error}</Text>}
        <Text style={styles.note}>Live polling is temporary; the versioned WebSocket transport is the next migration step.</Text>
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: "#0b1020" },
  container: { flex: 1, padding: 14, gap: 16 },
  statusRow: { flexDirection: "row", justifyContent: "space-between", alignItems: "center" },
  turn: { color: "#f8fafc", fontSize: 22, fontWeight: "700" },
  role: { color: "#94a3b8", marginTop: 3 },
  share: { borderWidth: 1, borderColor: "#67e8f9", paddingHorizontal: 15, paddingVertical: 9, borderRadius: 10 },
  shareText: { color: "#67e8f9", fontWeight: "600" },
  reactions: { flexDirection: "row", justifyContent: "space-between" },
  reaction: { width: 44, height: 44, alignItems: "center", justifyContent: "center", backgroundColor: "#172033", borderRadius: 12 },
  emoji: { fontSize: 24 },
  error: { color: "#fca5a5", textAlign: "center" },
  note: { color: "#64748b", fontSize: 11, textAlign: "center" },
});
