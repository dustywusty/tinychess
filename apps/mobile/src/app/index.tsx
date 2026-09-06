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

export default function HomeScreen() {
  const { colors } = useBoardTheme();
  const styles = useThemedStyles(createStyles);
  const ui = useUI();
  const router = useRouter();
  const [input, setInput] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [games, setGames] = useState<RecentGame[]>([]);
  useFocusEffect(useCallback(() => {
    let active = true;
    void recentGames().then((value) => { if (active) setGames(value); });
    return () => { active = false; };
  }, []));
  const openGame = (id: string) => router.push({ pathname: "/g/[id]", params: { id } });
  const handleCreate = async () => {
    setBusy(true);
    setError("");
    try { openGame((await createGame()).id); }
    catch { setError("Couldn’t start your game. Check your connection and try again."); }
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
        <View style={styles.hero}>
          <Text style={styles.kicker}>A SMALL GAME. A GOOD TIME.</Text>
          <Text style={styles.title}>Good company.{"\n"}Great moves.</Text>
          <Text style={styles.subtitle}>A little chess with your favorite people.{"\n"}Send a link. Make your move.</Text>
        </View>
        <View style={styles.playCard}>
          <View style={ui.row}><Text style={ui.eyebrow}>YOUR NEXT GOOD GAME</Text><View style={styles.liveDot} /></View>
          <View style={styles.art} accessible accessibilityLabel="A colorful chessboard, ready for a game">
            <View style={styles.artBoard}><ChessBoard fen={initialFEN} preview /></View>
            <View style={[styles.bubble, styles.bubbleOne]}><Text style={styles.emoji}>👋</Text></View>
            <View style={[styles.bubble, styles.bubbleTwo]}><Text style={styles.emoji}>🤔</Text></View>
            <View style={styles.artCaption}><Text style={styles.artCaptionText}>you + a friend</Text></View>
          </View>
          <Text style={styles.playTitle}>Across the board.{"\n"}Closer together.</Text>
          <Text style={[ui.body, { marginBottom: 16 }]}>No sign-up. No rush. Just chess.</Text>
          <Button title="Play a friend    ↗" busy={busy} onPress={() => void handleCreate()} />
        </View>
        {!!error && <ErrorMessage>{error}</ErrorMessage>}
        <View style={styles.join}>
          <Text style={ui.cardTitle}>Got an invite?</Text>
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
        {games.length > 0 && <View style={{ gap: 10 }}>
          <Text style={ui.eyebrow}>PICK UP WHERE YOU LEFT OFF</Text>
          {games.slice(0, 3).map((game) => <Pressable key={game.id} accessibilityRole="button" onPress={() => openGame(game.id)} style={styles.recent}>
            <View style={styles.recentIcon}><Piece piece="n" size={28} /></View>
            <View style={{ flex: 1 }}><Text style={styles.recentTitle}>{game.status || "Your friendly match"}</Text><Text style={ui.body}>{game.moves} moves · {game.id.slice(0, 8)}</Text></View>
            <Text style={{ fontSize: 20, color: colors.muted }}>↗</Text>
          </Pressable>)}
        </View>}
        <CoachCard />
        <Text style={styles.footer}>64 squares. Endless possibilities.</Text>
      </ScrollView>
    </KeyboardAvoidingView>
  </SafeAreaView>;
}
const createStyles = (colors: AppColors) => StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.background },
  container: { padding: 24, gap: 24, width: "100%", maxWidth: 480, alignSelf: "center", paddingBottom: 32 },
  brand: { flexDirection: "row", alignItems: "center", gap: 8 },
  mark: { width: 32, height: 32, backgroundColor: colors.mint, borderRadius: 10, alignItems: "center", justifyContent: "center" },
  wordmark: { fontSize: 22, fontWeight: "800", letterSpacing: -1.2, color: colors.ink },
  hero: { paddingTop: 14, gap: 14 },
  kicker: { fontSize: 10, letterSpacing: 2, fontWeight: "700", color: colors.kicker },
  title: { fontSize: 43, lineHeight: 47, fontWeight: "600", letterSpacing: -2.3, color: colors.ink },
  subtitle: { fontSize: 15, lineHeight: 23, color: colors.muted },
  playCard: { backgroundColor: colors.soft, borderRadius: 28, padding: 22, overflow: "hidden" },
  liveDot: { width: 7, height: 7, backgroundColor: "#7A9B62", borderRadius: 5 },
  art: { height: 205, alignItems: "center", justifyContent: "center", marginVertical: 8 },
  artBoard: { width: 176, transform: [{ rotate: "-9deg" }], borderWidth: 6, borderColor: "#fff", borderRadius: 15, boxShadow: "0 12px 20px #253D3518" },
  bubble: { position: "absolute", width: 53, height: 53, borderRadius: 18, alignItems: "center", justifyContent: "center", boxShadow: "0 6px 12px #253D3510" },
  bubbleOne: { left: 5, top: 24, backgroundColor: colors.coral, transform: [{ rotate: "-13deg" }] },
  bubbleTwo: { right: 3, bottom: 28, backgroundColor: colors.lilac, transform: [{ rotate: "12deg" }] },
  emoji: { fontSize: 29 },
  artCaption: { position: "absolute", bottom: 0, backgroundColor: colors.surface, paddingHorizontal: 13, paddingVertical: 7, borderRadius: 14, transform: [{ rotate: "-4deg" }] },
  artCaptionText: { fontSize: 11, color: colors.ink, fontWeight: "600" },
  playTitle: { fontSize: 26, lineHeight: 30, letterSpacing: -0.8, color: colors.ink, fontWeight: "600", marginTop: 5, marginBottom: 8 },
  join: { gap: 8 },
  inputRow: { flexDirection: "row", alignItems: "center", borderWidth: 1, borderColor: colors.line, backgroundColor: colors.surface, borderRadius: 18, padding: 5 },
  input: { flex: 1, minWidth: 0, minHeight: 44, color: colors.ink, paddingHorizontal: 10, fontSize: 14 },
  joinButton: { width: 44, height: 44, borderRadius: 13, alignItems: "center", justifyContent: "center", backgroundColor: colors.ink },
  validation: { color: colors.error, fontSize: 12, lineHeight: 18 },
  recent: { flexDirection: "row", alignItems: "center", padding: 14, gap: 12, borderRadius: 18, backgroundColor: colors.surface },
  recentIcon: { width: 44, height: 44, borderRadius: 14, backgroundColor: colors.soft, justifyContent: "center", alignItems: "center" },
  recentTitle: { color: colors.ink, fontSize: 14, fontWeight: "600", marginBottom: 3 },
  footer: { textAlign: "center", color: colors.muted, fontSize: 11, letterSpacing: 0.3 },
});
