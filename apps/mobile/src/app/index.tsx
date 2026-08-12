import { useState } from "react";
import { useRouter } from "expo-router";
import {
  ActivityIndicator,
  Pressable,
  SafeAreaView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";
import { createGame } from "@/lib/api";

function gameIDFromInput(input: string): string {
  const trimmed = input.trim().replace(/\/+$/, "");
  if (!trimmed) return "";
  const parts = trimmed.split("/").filter(Boolean);
  return parts.at(-1) ?? "";
}

export default function HomeScreen() {
  const router = useRouter();
  const [input, setInput] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const openGame = (id: string) => {
    router.push({ pathname: "/g/[id]", params: { id } });
  };

  const handleCreate = async () => {
    setBusy(true);
    setError("");
    try {
      const game = await createGame();
      openGame(game.id);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Could not create game");
    } finally {
      setBusy(false);
    }
  };

  const pastedID = gameIDFromInput(input);

  return (
    <SafeAreaView style={styles.safe}>
      <View style={styles.container}>
        <Text style={styles.mark}>♞</Text>
        <Text style={styles.title}>Your Move</Text>
        <Text style={styles.subtitle}>Start a game. Share the link. Play.</Text>
        <Pressable style={styles.primary} onPress={() => void handleCreate()} disabled={busy}>
          {busy ? <ActivityIndicator color="#07111f" /> : <Text style={styles.primaryText}>Create game</Text>}
        </Pressable>
        <View style={styles.divider} />
        <TextInput
          value={input}
          onChangeText={setInput}
          autoCapitalize="none"
          autoCorrect={false}
          placeholder="Paste a game link or ID"
          placeholderTextColor="#64748b"
          style={styles.input}
        />
        <Pressable
          style={[styles.secondary, !pastedID && styles.disabled]}
          onPress={() => pastedID && openGame(pastedID)}
          disabled={!pastedID}
        >
          <Text style={styles.secondaryText}>Open game</Text>
        </Pressable>
        {!!error && <Text style={styles.error}>{error}</Text>}
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: "#0b1020" },
  container: { flex: 1, justifyContent: "center", padding: 24, gap: 14 },
  mark: { color: "#67e8f9", fontSize: 42, textAlign: "center" },
  title: { color: "#f8fafc", fontSize: 32, fontWeight: "700", textAlign: "center" },
  subtitle: { color: "#94a3b8", fontSize: 16, textAlign: "center", marginBottom: 18 },
  primary: { minHeight: 52, borderRadius: 14, backgroundColor: "#67e8f9", alignItems: "center", justifyContent: "center" },
  primaryText: { color: "#07111f", fontSize: 17, fontWeight: "700" },
  divider: { height: 1, backgroundColor: "#1e293b", marginVertical: 8 },
  input: { minHeight: 50, borderWidth: 1, borderColor: "#334155", borderRadius: 12, color: "#f8fafc", paddingHorizontal: 14, fontSize: 15 },
  secondary: { minHeight: 48, borderRadius: 12, borderWidth: 1, borderColor: "#67e8f9", alignItems: "center", justifyContent: "center" },
  secondaryText: { color: "#67e8f9", fontSize: 16, fontWeight: "600" },
  disabled: { opacity: 0.4 },
  error: { color: "#fca5a5", textAlign: "center" },
});
