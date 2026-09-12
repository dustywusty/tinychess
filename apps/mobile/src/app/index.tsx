import { useCallback, useState } from "react";
import { useFocusEffect, useRouter } from "expo-router";
import { KeyboardAvoidingView, Platform, Pressable, ScrollView, StyleSheet, Text, TextInput, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { createGame } from "@/lib/api";
import { gameIDFromInput, initialFEN } from "@/lib/chess";
import { recentGames, type RecentGame } from "@/lib/recentGames";
import { useBoardTheme, useThemedStyles, type AppColors } from "@/lib/theme";
import { ChessBoard } from "@/components/ChessBoard";
import { Piece } from "@/components/Piece";
import { Button, CoachCard, ErrorMessage, useUI } from "@/components/UI";
import { AppearanceMenu } from "@/components/AppearanceMenu";
import { BotPicker } from "@/components/BotPicker";
import { clientID } from "@/lib/session";
import { bots, type BotId } from "@yourmove/chess/bots";
import { resultBanner } from "@yourmove/chess";

export default function HomeScreen() {
  const { colors } = useBoardTheme();
  const styles = useThemedStyles(createStyles);
  const ui = useUI();
  const router = useRouter();
  const [input, setInput] = useState("");
  const [busy, setBusy] = useState(false);
  const [computer, setComputer] = useState(false);
  const [error, setError] = useState("");
  const [games, setGames] = useState<RecentGame[]>([]);
  useFocusEffect(useCallback(() => {
    let active = true;
    void recentGames().then((value) => { if (active) setGames(value); });
    return () => { active = false; };
  }, []));
  const openGame = (id: string) => router.push({ pathname: "/g/[id]", params: { id } });
  const handleCreate = async (botId?: BotId, color: "w" | "b" = "w") => {
    setBusy(true);
    setError("");
    try { openGame((await createGame(botId ? { botId, color, clientId: await clientID() } : undefined)).id); setComputer(false); }
    catch { setComputer(false); setError("Couldn’t start your game. Check your connection and try again."); }
    finally { setBusy(false); }
  };
  const pastedID = gameIDFromInput(input);
  return <SafeAreaView style={styles.safe}>
    <KeyboardAvoidingView style={{ flex: 1 }} behavior={Platform.OS === "ios" ? "padding" : undefined}>
      <ScrollView contentContainerStyle={styles.container} keyboardShouldPersistTaps="handled" showsVerticalScrollIndicator={false}>
        <View style={ui.row}>
          <View style={styles.brand}><View style={styles.mark}><Piece piece="n" size={25} /></View><Text style={styles.wordmark}>your move<Text style={{ color: "#7B9E61" }}>.</Text></Text></View>
          <AppearanceMenu />
        </View>
        <View testID="home-play-card" style={styles.playCard}>
          <View style={ui.row}><Text style={ui.eyebrow}>YOUR NEXT GOOD GAME</Text><View style={styles.liveDot} /></View>
          <View style={styles.art} accessible accessibilityLabel="A colorful chessboard, ready for a game">
            <View style={styles.artBoard}><ChessBoard fen={initialFEN} preview /></View>
            <View style={[styles.bubble, styles.bubbleOne]}><Text style={styles.emoji}>👋</Text></View>
            <View style={[styles.bubble, styles.bubbleTwo]}><Text style={styles.emoji}>🤔</Text></View>
            <View style={styles.artCaption}><Text style={styles.artCaptionText}>you + a friend</Text></View>
          </View>
          <Button title="Play a friend    ↗" busy={busy} onPress={() => void handleCreate()} />
          <Pressable accessibilityRole="button" disabled={busy} onPress={() => setComputer(true)} style={{ padding: 18, marginTop: 8, alignItems: "center", borderWidth: 1, borderColor: colors.line, borderRadius: 18 }}><Text style={{ color: colors.ink, fontWeight: "600" }}>Play the computer    ✳</Text></Pressable>
        </View>
        {!!error && <ErrorMessage>{error}</ErrorMessage>}
        <View testID="invite-section" style={styles.join}>
          <Text accessibilityRole="header" style={styles.title}>A little chess with your favorite people.</Text>
          <Text style={ui.body}>Send a link. Make your move.</Text>
          <View style={styles.inputRow}>
            <TextInput accessibilityLabel="Game link or ID" value={input} onChangeText={setInput} autoCapitalize="none" autoCorrect={false}
              placeholder="Paste a game link or ID" placeholderTextColor={colors.muted} style={styles.input} returnKeyType="go"
              onSubmitEditing={() => { if (pastedID) openGame(pastedID); }} />
            <Pressable accessibilityRole="button" accessibilityLabel="Join game" disabled={!pastedID || busy}
              accessibilityState={{ disabled: !pastedID || busy }} onPress={() => openGame(pastedID)} style={[styles.joinButton, (!pastedID || busy) && { opacity: 0.4 }]}>
              <Text style={{ color: colors.background, fontSize: 20 }}>↗</Text>
            </Pressable>
          </View>
          {!!input.trim() && !pastedID && <Text style={styles.validation}>Use a game ID or a link ending in /g/your-game-id.</Text>}
        </View>
        <View testID="recent-games" style={{ gap: 10 }}>
          <Text style={ui.eyebrow}>PICK UP WHERE YOU LEFT OFF</Text>
          {games.length === 0 && <View style={styles.recentEmpty}>
            <View style={styles.emptyIcon} aria-hidden><Piece piece="n" size={36} /><Text style={styles.emptyBubble}>…</Text></View>
            <View style={{ flex: 1 }}><Text style={styles.recentTitle}>No games yet.</Text><Text style={styles.recentDetail}>A lonely knight, waiting for your first move.</Text></View>
          </View>}
          {games.slice(0, 3).map((game) => {
            const bot = bots.find((bot) => bot.id === game.botId);
            const result = resultBanner(game);
            const palette = result ? resultColors[result.tone] : undefined;
            return <Pressable key={game.id} accessibilityRole="button" onPress={() => openGame(game.id)} style={[styles.recent, palette && { borderWidth: 1, borderColor: palette.line }]}>
            <View style={styles.recentRow}>
            <View style={styles.recentIcon}><Piece piece="n" size={28} /></View>
            <View style={{ flex: 1 }}>
              <Text style={styles.recentMode}>{game.opponentType === "bot" ? "PvBot" : game.opponentType === "human" ? "PvP" : "Saved game"}</Text>
              <Text style={styles.recentTitle}>{game.opponentType === "bot" ? `A match with ${bot?.name ?? "the computer"}` : game.opponentType === "human" ? "Your friendly match" : "Your saved match"}</Text>
              <Text style={styles.recentDetail}>{result ? "Completed" : game.status || "In progress"} · {game.id.slice(0, 8)}</Text>
            </View>
            <Text style={{ fontSize: 20, color: colors.muted }}>↗</Text>
            </View>
            {result && palette && <View testID="recent-result" style={[styles.result, { backgroundColor: palette.bg, borderTopColor: palette.shine }]}>
              <View style={{ flexDirection: "row", alignItems: "center", gap: 8 }}><Text aria-hidden style={{ color: palette.ink, fontSize: 16 }}>{result.symbol}</Text><Text style={[styles.resultLabel, { color: palette.ink }]}>{result.label}</Text></View>
              <Text style={[styles.resultScore, { color: palette.ink }]}>{result.score}</Text>
            </View>}
          </Pressable>; })}
        </View>
        <CoachCard />
        <Text style={styles.footer}>64 squares. Endless possibilities.</Text>
      </ScrollView>
    </KeyboardAvoidingView>
    <BotPicker open={computer} busy={busy} onClose={() => setComputer(false)} onStart={(id, color) => void handleCreate(id, color)} />
  </SafeAreaView>;
}
const resultColors = {
  win: { bg: "#D9F38D", ink: "#29451E", line: "#96C34E", shine: "#F0FFC5" },
  loss: { bg: "#F7B9B2", ink: "#722D2B", line: "#DD857E", shine: "#FFE2D9" },
  draw: { bg: "#E6DFFA", ink: "#51406C", line: "#B6A0D6", shine: "#F6F0FF" },
  neutral: { bg: "#E2E8E3", ink: "#3E5045", line: "#A3B5A8", shine: "#F3F8F4" },
};
const createStyles = (colors: AppColors) => StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.background },
  container: { padding: 24, gap: 24, width: "100%", maxWidth: 480, alignSelf: "center", paddingBottom: 32 },
  brand: { flexDirection: "row", alignItems: "center", gap: 8 },
  mark: { width: 32, height: 32, backgroundColor: colors.mint, borderRadius: 10, alignItems: "center", justifyContent: "center" },
  wordmark: { fontSize: 22, fontWeight: "800", letterSpacing: -1.2, color: colors.ink },
  title: { fontSize: 26, lineHeight: 31, fontWeight: "600", letterSpacing: -0.8, color: colors.ink },
  playCard: { backgroundColor: colors.soft, borderRadius: 28, padding: 22, overflow: "hidden" },
  liveDot: { width: 7, height: 7, backgroundColor: "#7A9B62", borderRadius: 5 },
  art: { height: 205, alignItems: "center", justifyContent: "center", marginTop: 8, marginBottom: 16 },
  artBoard: { width: 176, transform: [{ rotate: "-9deg" }], borderWidth: 6, borderColor: "#fff", borderRadius: 15, boxShadow: "0 12px 20px #253D3518" },
  bubble: { position: "absolute", width: 53, height: 53, borderRadius: 18, alignItems: "center", justifyContent: "center", boxShadow: "0 6px 12px #253D3510" },
  bubbleOne: { left: 5, top: 24, backgroundColor: colors.coral, transform: [{ rotate: "-13deg" }] },
  bubbleTwo: { right: 3, bottom: 28, backgroundColor: colors.lilac, transform: [{ rotate: "12deg" }] },
  emoji: { fontSize: 29 },
  artCaption: { position: "absolute", bottom: 0, backgroundColor: colors.surface, paddingHorizontal: 13, paddingVertical: 7, borderRadius: 14, transform: [{ rotate: "-4deg" }] },
  artCaptionText: { fontSize: 11, color: colors.ink, fontWeight: "600" },
  join: { gap: 8 },
  inputRow: { flexDirection: "row", alignItems: "center", borderWidth: 1, borderColor: colors.line, backgroundColor: colors.surface, borderRadius: 18, padding: 5 },
  input: { flex: 1, minWidth: 0, minHeight: 44, color: colors.ink, paddingHorizontal: 10, fontSize: 14 },
  joinButton: { width: 44, height: 44, borderRadius: 13, alignItems: "center", justifyContent: "center", backgroundColor: colors.ink },
  validation: { color: colors.error, fontSize: 12, lineHeight: 18 },
  recent: { borderRadius: 18, backgroundColor: colors.surface, overflow: "hidden" },
  recentRow: { flexDirection: "row", alignItems: "center", padding: 14, gap: 12 },
  result: { flexDirection: "row", alignItems: "center", justifyContent: "space-between", gap: 12, paddingHorizontal: 16, paddingVertical: 8, borderTopWidth: 2 },
  resultLabel: { fontSize: 11, fontWeight: "800", letterSpacing: 1.4, textTransform: "uppercase" },
  resultScore: { fontSize: 11, fontWeight: "600", letterSpacing: 0.4, fontVariant: ["tabular-nums"] },
  recentIcon: { width: 44, height: 44, borderRadius: 14, backgroundColor: colors.soft, justifyContent: "center", alignItems: "center" },
  recentTitle: { color: colors.ink, fontSize: 14, fontWeight: "600", marginBottom: 3 },
  recentMode: { alignSelf: "flex-start", backgroundColor: colors.soft, color: colors.ink, fontSize: 10, fontWeight: "600", paddingHorizontal: 7, paddingVertical: 3, borderRadius: 6, marginBottom: 5 },
  recentDetail: { color: colors.muted, fontSize: 12, lineHeight: 18 },
  recentEmpty: { flexDirection: "row", alignItems: "center", gap: 18, padding: 20, borderWidth: 1, borderStyle: "dashed", borderColor: colors.line, borderRadius: 18, backgroundColor: colors.surface },
  emptyIcon: { width: 52, height: 52, borderRadius: 17, backgroundColor: colors.soft, alignItems: "center", justifyContent: "center", transform: [{ rotate: "-10deg" }] },
  emptyBubble: { position: "absolute", top: -10, right: -6, paddingHorizontal: 6, paddingBottom: 4, borderRadius: 8, backgroundColor: colors.surface, color: colors.muted, fontSize: 15, transform: [{ rotate: "10deg" }] },
  footer: { textAlign: "center", color: colors.muted, fontSize: 11, letterSpacing: 0.3 },
});
