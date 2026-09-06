import { useEffect, useMemo, useRef, useState } from "react";
import { useLocalSearchParams, useRouter } from "expo-router";
import { Chess, type Square } from "chess.js";
import { ActivityIndicator, Modal, Platform, Pressable, ScrollView, Share, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { ChessBoard } from "@/components/ChessBoard";
import { Piece } from "@/components/Piece";
import { Button, CoachCard, ErrorMessage, ThemePicker, ui } from "@/components/UI";
import { makeMove, reactToGame } from "@/lib/api";
import { color, initialFEN, legalMoves, moveLabels } from "@/lib/chess";
import { colors } from "@/lib/theme";
import { useLiveGame } from "@/lib/useLiveGame";

const emojis = [
  { emoji: "👋", label: "Wave" }, { emoji: "🤔", label: "Thinking" }, { emoji: "🔥", label: "Fire" },
  { emoji: "😂", label: "Laugh" }, { emoji: "👏", label: "Applause" }, { emoji: "🤝", label: "Good game" },
];
export default function GameScreen() {
  const params = useLocalSearchParams<{ id: string | string[] }>();
  const gameID = Array.isArray(params.id) ? params.id[0] ?? "" : params.id ?? "";
  const router = useRouter();
  const game = useLiveGame(gameID);
  const { state, cid, connected } = game;
  const [selected, setSelected] = useState<string | null>(null);
  const [promotion, setPromotion] = useState<{ from: string; to: string } | null>(null);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [moving, setMoving] = useState(false);
  const moveLock = useRef(false);
  const [reacting, setReacting] = useState(false);
  const reactionLock = useRef(false);
  const [flipped, setFlipped] = useState(false);
  const [showMoves, setShowMoves] = useState(false);
  const playerColor = color(state?.color);
  const turn = color(state?.turn);
  const canMove = connected && !!state && state.role === "player" && !state.status && playerColor === turn && !moving;
  const perspective = flipped ? (playerColor === "black" ? "white" : "black") : playerColor ?? "white";
  const fen = state?.fen ?? initialFEN;
  const moves = useMemo(() => selected ? legalMoves(fen, selected) : [], [fen, selected]);
  const notation = useMemo(() => moveLabels(state?.uci ?? []), [state?.uci]);
  const check = useMemo(() => new Chess(fen).isCheck(), [fen]);
  useEffect(() => { setSelected(null); setPromotion(null); }, [fen, gameID, connected]);
  useEffect(() => {
    if (!notice) return;
    const timer = setTimeout(() => setNotice(""), 3500);
    return () => clearTimeout(timer);
  }, [notice]);

  const submit = async (from: string, to: string, promote?: string) => {
    if (!canMove || moveLock.current) return;
    moveLock.current = true;
    setMoving(true); setSelected(null); setPromotion(null); setError("");
    try {
      const response = await makeMove(gameID, from + to + (promote ?? ""), cid);
      if (!response.ok) setError(response.error ?? "That move wasn’t accepted. Try another square.");
      // A newer stream event may arrive before this command response.
      if (response.state && (response.state.uci?.length ?? 0) >= (game.stateRef.current?.uci.length ?? 0)) {
        game.accept({ ...response.state, kind: "state" });
      }
    } catch { setError("Your move couldn’t be confirmed. Wait for the board to reconnect before trying again."); game.retry(); }
    finally { moveLock.current = false; setMoving(false); }
  };
  const handleSquare = (square: string) => {
    if (!canMove) return;
    if (selected === square) { setSelected(null); return; }
    const piece = new Chess(fen).get(square as Square);
    if (piece && color(piece.color) === playerColor) { setSelected(square); return; }
    const move = moves.find((candidate) => candidate.to === square);
    if (!selected || !move) return;
    if (move.promotion) setPromotion({ from: selected, to: square });
    else void submit(selected, square);
  };
  const sendReaction = async (emoji: string) => {
    if (!connected || !cid || reactionLock.current) return;
    reactionLock.current = true; setReacting(true); setError("");
    try {
      const response = await reactToGame(gameID, emoji, cid);
      if (!response.ok) setError(response.error?.startsWith("cooldown") ? "Give that reaction a moment before sending another." : "That reaction didn’t send. Try again.");
      else game.addReaction({ kind: "emoji", emoji, sender: cid, at: Date.now() });
    } catch { setError("That reaction didn’t send. Check your connection and try again."); }
    finally { reactionLock.current = false; setReacting(false); }
  };
  const share = async () => {
    const root = process.env.EXPO_PUBLIC_WEB_URL?.replace(/\/$/, "");
    const url = root ? root + "/g/" + gameID : Platform.OS === "web" ? window.location.origin + "/g/" + gameID : "yourmove://g/" + gameID;
    try {
      if (Platform.OS === "web" && !navigator.share) {
        await navigator.clipboard.writeText(url);
        setNotice("Game link copied. Send it to a friend!");
      } else if (Platform.OS === "web") await navigator.share({ title: "Your Move", text: "A little chess?", url });
      else await Share.share({ message: "A little chess? " + url });
    } catch (cause) {
      if (!(cause instanceof Error && cause.name === "AbortError")) setError("Couldn’t share the link. Please try again.");
    }
  };
  const heading = !state ? "Finding your board…" : !connected ? "Reconnecting…" : state.status ? "That’s a game." : state.role === "spectator" ? "Enjoy the game." : canMove ? (check ? "You’re in check." : "Your move.") : "Over to them.";
  const bottomColor = perspective;
  const topColor = bottomColor === "white" ? "black" : "white";
  const player = (side: "white" | "black") => <View style={styles.player}>
    <View style={[styles.avatar, { backgroundColor: side === "white" ? "#EDF0E5" : "#DFE4D9" }]}><Piece piece={side === "white" ? "K" : "k"} size={29} /></View>
    <View style={{ flex: 1 }}><Text style={styles.playerName}>{state?.role === "player" ? side === playerColor ? "You" : "Your friend" : side === "white" ? "White" : "Black"}</Text><Text style={styles.playerMeta}>{side === "white" ? "White pieces" : "Black pieces"}</Text></View>
    {connected && !state?.status && turn === side && <View style={styles.turnBadge}><View style={styles.dot} /><Text style={styles.turnText}>TO MOVE</Text></View>}
  </View>;

  return <SafeAreaView style={styles.safe}>
    <ScrollView contentContainerStyle={styles.container} showsVerticalScrollIndicator={false}>
      <View style={ui.row}>
        <Pressable accessibilityRole="button" accessibilityLabel="Back to home" onPress={() => router.canGoBack() ? router.back() : router.replace("/")} style={styles.iconButton}><Text style={styles.icon}>←</Text></Pressable>
        <Text style={styles.wordmark}>your move.</Text>
        <Pressable accessibilityRole="button" accessibilityLabel="Share game" onPress={() => void share()} style={styles.iconButton}><Text style={styles.icon}>↗</Text></Pressable>
      </View>
      <View style={styles.headingBlock}>
        <View style={ui.row}><Text style={ui.eyebrow}>A FRIENDLY MATCH</Text><View style={styles.connection}><View style={[styles.dot, { backgroundColor: connected ? "#76915A" : "#CBA377" }]} /><Text style={styles.connectionText}>{connected ? "LIVE" : "CONNECTING"}</Text></View></View>
        <Text accessibilityLiveRegion="polite" style={styles.heading}>{heading}</Text>
        <Text style={ui.body}>{state?.status || (state?.role === "spectator" ? "The seats are full. You can watch and react." : canMove ? "Take your time. Make it a good one." : "Good things come to those who wait.")}</Text>
      </View>
      <View style={styles.boardArea}>
        {player(topColor)}
        <View>
          <ChessBoard fen={fen} perspective={perspective} selected={selected} disabled={!canMove}
            destinations={moves.map((move) => move.to)} lastMove={state?.uci.at(-1)} onSquare={handleSquare} />
          {!state && <View style={styles.loading}><ActivityIndicator size="large" color={colors.ink} /><Text style={styles.playerName}>Opening your game…</Text></View>}
        </View>
        {player(bottomColor)}
      </View>
      {(!!game.error || !!error) && <View style={{ gap: 8 }}>
        <ErrorMessage>{error || game.error}</ErrorMessage>
        {!!game.error && <Button title="Try reconnecting" onPress={game.retry} />}
      </View>}
      {!!notice && <Text accessibilityLiveRegion="polite" style={styles.notice}>{notice}</Text>}
      {!!state && !state.status && state.uci.length < 2 && <Pressable accessibilityRole="button" onPress={() => void share()} style={styles.invite}>
        <Text style={{ fontSize: 22 }}>👋</Text><View style={{ flex: 1 }}><Text style={styles.playerName}>Better with a friend.</Text><Text style={styles.playerMeta}>Share this game and meet at the board.</Text></View><Text style={styles.icon}>↗</Text>
      </Pressable>}
      <View style={styles.chat}>
        <View style={ui.row}><Text style={ui.eyebrow}>A LITTLE BACK & FORTH</Text><Text style={styles.playerMeta}>Say it with an emoji</Text></View>
        <View style={styles.reactions}>{emojis.map(({ emoji, label }) => <Pressable key={emoji}
          accessibilityRole="button" accessibilityLabel={"Send " + label} accessibilityState={{ disabled: !connected || reacting }} disabled={!connected || reacting}
          onPress={() => void sendReaction(emoji)} style={({ pressed }) => [styles.reaction, pressed && { backgroundColor: colors.mint }, (!connected || reacting) && { opacity: 0.4 }]}>
          <Text style={styles.emoji}>{emoji}</Text>
        </Pressable>)}</View>
        {game.reactions.length > 0 && <View accessibilityLiveRegion="polite" style={styles.chatHistory}>{game.reactions.map((reaction, index) => <View key={reaction.sender + reaction.at + index} style={[styles.chatBubble, { backgroundColor: reaction.sender === cid ? "#ECF3DD" : colors.lilac }]}>
          <Text style={{ fontSize: 19 }}>{reaction.emoji}</Text><Text style={styles.chatWho}>{reaction.sender === cid ? "You" : "Them"}</Text>
        </View>)}</View>}
      </View>
      <View style={styles.tools}>
        <Pressable accessibilityRole="button" accessibilityLabel="Flip board" onPress={() => setFlipped((value) => !value)} style={styles.toolButton}><Text style={styles.toolText}>↻  Flip</Text></Pressable>
        <Pressable accessibilityRole="button" aria-expanded={showMoves} accessibilityState={{ expanded: showMoves }} onPress={() => setShowMoves((value) => !value)} style={styles.toolButton}><Text style={styles.toolText}>≡  Moves{notation.length ? " · " + notation.length : ""}</Text></Pressable>
        <ThemePicker />
      </View>
      {showMoves && <View style={styles.movePanel}>
        <Text style={ui.eyebrow}>THE GAME SO FAR</Text>
        {notation.length === 0 ? <Text style={ui.body}>The first move is yours to make.</Text> : <View style={styles.moveList}>{notation.map((move, index) => <Text key={index} style={styles.moveText}>{index % 2 === 0 ? Math.floor(index / 2 + 1) + ". " : ""}{move}</Text>)}</View>}
      </View>}
      <CoachCard />
      <Text style={styles.footer}>A little less scrolling. A little more chess.</Text>
    </ScrollView>
    <Modal visible={!!promotion} transparent animationType="fade" onRequestClose={() => setPromotion(null)}>
      <View style={styles.scrim}><View accessibilityViewIsModal style={styles.promotion}>
        <Text style={ui.cardTitle}>A pawn with possibilities.</Text><Text style={ui.body}>Choose your promotion.</Text>
        <View style={styles.promotionOptions}>{["q", "r", "b", "n"].map((piece, index) => <Pressable key={piece} accessibilityRole="button"
          accessibilityLabel={"Promote to " + ["queen", "rook", "bishop", "knight"][index]} style={styles.promotionPiece}
          onPress={() => { if (promotion) void submit(promotion.from, promotion.to, piece); }}><Piece piece={playerColor === "white" ? piece.toUpperCase() : piece} size={46} /></Pressable>)}</View>
        <Button title="Cancel" onPress={() => setPromotion(null)} />
      </View></View>
    </Modal>
  </SafeAreaView>;
}
const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.background },
  container: { padding: 20, gap: 20, width: "100%", maxWidth: 480, alignSelf: "center", paddingBottom: 30 },
  iconButton: { width: 44, height: 44, borderRadius: 15, borderWidth: 1, borderColor: colors.line, alignItems: "center", justifyContent: "center" },
  icon: { fontSize: 22, color: colors.ink },
  wordmark: { fontSize: 20, fontWeight: "800", color: colors.ink, letterSpacing: -1 },
  headingBlock: { gap: 9, paddingTop: 7 },
  heading: { fontSize: 36, lineHeight: 42, fontWeight: "600", letterSpacing: -1.5, color: colors.ink },
  connection: { flexDirection: "row", gap: 5, alignItems: "center" },
  connectionText: { fontSize: 9, letterSpacing: 1, color: colors.muted, fontWeight: "600" },
  dot: { width: 6, height: 6, borderRadius: 4, backgroundColor: "#76915A" },
  boardArea: { gap: 12 },
  player: { flexDirection: "row", alignItems: "center", gap: 10 },
  avatar: { width: 41, height: 41, borderRadius: 14, alignItems: "center", justifyContent: "center" },
  playerName: { fontSize: 14, fontWeight: "600", color: colors.ink },
  playerMeta: { fontSize: 11, color: colors.muted, marginTop: 3 },
  turnBadge: { flexDirection: "row", alignItems: "center", gap: 6, backgroundColor: "#EAF0DE", paddingVertical: 8, paddingHorizontal: 10, borderRadius: 8 },
  turnText: { fontSize: 8, letterSpacing: 1, fontWeight: "700", color: "#536943" },
  loading: { position: "absolute", inset: 0, backgroundColor: "#F8F7F2DD", justifyContent: "center", alignItems: "center", gap: 12, borderRadius: 12 },
  notice: { padding: 14, backgroundColor: colors.mint, color: colors.ink, borderRadius: 14, fontSize: 13 },
  invite: { backgroundColor: "#EEF0E7", borderRadius: 18, padding: 14, flexDirection: "row", alignItems: "center", gap: 12 },
  chat: { gap: 13 },
  reactions: { flexDirection: "row", justifyContent: "space-between", gap: 5 },
  reaction: { flex: 1, minHeight: 48, borderRadius: 15, alignItems: "center", justifyContent: "center", backgroundColor: colors.surface, borderWidth: 1, borderColor: colors.line },
  emoji: { fontSize: 25 },
  chatHistory: { flexDirection: "row", flexWrap: "wrap", gap: 6 },
  chatBubble: { paddingVertical: 7, paddingHorizontal: 9, borderRadius: 12, flexDirection: "row", gap: 5, alignItems: "center" },
  chatWho: { fontSize: 10, color: colors.ink },
  tools: { flexDirection: "row", alignItems: "center", justifyContent: "space-between", borderTopWidth: 1, borderBottomWidth: 1, borderColor: colors.line, paddingVertical: 4 },
  toolButton: { minHeight: 44, paddingHorizontal: 5, justifyContent: "center" },
  toolText: { fontSize: 12, fontWeight: "500", color: colors.muted },
  movePanel: { padding: 18, gap: 12, backgroundColor: colors.surface, borderRadius: 18 },
  moveList: { flexDirection: "row", flexWrap: "wrap", gap: 12 },
  moveText: { color: colors.ink, fontSize: 14, fontVariant: ["tabular-nums"] },
  footer: { textAlign: "center", fontSize: 11, color: colors.muted },
  scrim: { flex: 1, backgroundColor: "#252B2866", justifyContent: "center", padding: 24 },
  promotion: { width: "100%", maxWidth: 400, alignSelf: "center", backgroundColor: colors.background, borderRadius: 26, padding: 24, gap: 12 },
  promotionOptions: { flexDirection: "row", gap: 8, marginVertical: 8 },
  promotionPiece: { flex: 1, height: 64, borderRadius: 14, alignItems: "center", justifyContent: "center", backgroundColor: "#E9EDDF" },
});
