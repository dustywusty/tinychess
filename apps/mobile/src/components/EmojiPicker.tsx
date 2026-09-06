import { useEffect, useMemo, useRef, useState } from "react";
import { FlatList, Keyboard, KeyboardAvoidingView, Modal, Platform, Pressable, ScrollView, StyleSheet, Text, TextInput, View, useWindowDimensions } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import AsyncStorage from "@react-native-async-storage/async-storage";
import emojiData from "emoji-picker-element-data/en/emojibase/data.json";
import { useBoardTheme, useThemedStyles, type AppColors } from "@/lib/theme";
import { commonEmojis, parseRecentEmojis, quickEmojis, withRecentEmoji } from "@/lib/emojis";

interface Emoji { emoji: string; annotation: string; group: number; tags?: string[]; shortcodes?: string[]; skins?: { emoji: string; tone: number | number[] }[] }
const catalog: Emoji[] = emojiData;
const searchable = catalog.map((item) => ({ ...item, search: [item.annotation, ...(item.tags ?? []), ...(item.shortcodes ?? []), item.emoji].join(" ").toLowerCase() }));
const categories = [{ id: 0, name: "Faces" }, { id: 1, name: "People" }, { id: 3, name: "Nature" }, { id: 4, name: "Food" }, { id: 5, name: "Places" }, { id: 6, name: "Play" }, { id: 7, name: "Things" }, { id: 8, name: "Symbols" }, { id: 9, name: "Flags" }];
const tones = ["Default", "Light", "Medium-light", "Medium", "Medium-dark", "Dark"];
const RECENT_KEY = "yourmove.recent-emojis";
const labelFor = (emoji: string) => commonEmojis.find((item) => item.emoji === emoji)?.label ?? catalog.find((item) => item.emoji === emoji || item.skins?.some((skin) => skin.emoji === emoji))?.annotation ?? emoji;

export function EmojiPicker({ disabled, onSend }: { disabled: boolean; onSend: (emoji: string) => Promise<boolean> }) {
  const { colors } = useBoardTheme();
  const styles = useThemedStyles(createStyles);
  const { width, height } = useWindowDimensions();
  const insets = useSafeAreaInsets();
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [category, setCategory] = useState(0);
  const [tone, setTone] = useState(0);
  const [recent, setRecent] = useState<string[]>([]);
  const changed = useRef(false);
  const [sending, setSending] = useState(false);
  const lock = useRef(false);
  const [cooldown, setCooldown] = useState(0);
  const trigger = useRef<View>(null);
  const list = useRef<FlatList<Emoji>>(null);
  const unavailable = disabled || sending || cooldown > Date.now();
  useEffect(() => {
    void AsyncStorage.getItem(RECENT_KEY).then((saved) => { if (!changed.current) setRecent(parseRecentEmojis(saved)); }).catch(() => {});
  }, []);
  useEffect(() => {
    if (!cooldown) return;
    const timer = setTimeout(() => setCooldown(0), Math.max(0, cooldown - Date.now()));
    return () => clearTimeout(timer);
  }, [cooldown]);
  useEffect(() => {
    if (!open || Platform.OS !== "web") return;
    const escape = (event: KeyboardEvent) => { if (event.key === "Escape") setOpen(false); };
    window.addEventListener("keydown", escape);
    return () => window.removeEventListener("keydown", escape);
  }, [open]);
  const results = useMemo(() => {
    const words = query.trim().toLowerCase().split(/\s+/);
    return searchable.filter((item) => query.trim() ? words.every((word) => item.search.includes(word)) : item.group === category);
  }, [query, category]);
  const send = async (emoji: string) => {
    if (unavailable || lock.current) return;
    lock.current = true; setSending(true); setOpen(false); Keyboard.dismiss();
    try {
      if (await onSend(emoji)) {
        changed.current = true;
        const next = withRecentEmoji(recent, emoji);
        setRecent(next);
        setCooldown(Date.now() + 5000);
        void AsyncStorage.setItem(RECENT_KEY, JSON.stringify(next)).catch(() => {});
      }
    } finally { lock.current = false; setSending(false); }
  };
  const resetScroll = () => list.current?.scrollToOffset({ offset: 0, animated: false });
  const panelWidth = Math.min(420, width - 32);
  const columns = Math.max(4, Math.floor((panelWidth - 32) / 48));
  const cellWidth = (panelWidth - 32) / columns;
  return <>
    <View style={styles.quick}>{quickEmojis(recent).map((emoji) => <Pressable key={emoji} accessibilityRole="button" accessibilityLabel={`Send ${labelFor(emoji)}`}
      disabled={unavailable} accessibilityState={{ disabled: unavailable }} onPress={() => void send(emoji)} style={[styles.reaction, unavailable && styles.disabled]}><Text style={styles.emoji}>{emoji}</Text></Pressable>)}</View>
    <View style={styles.footer}><Text style={styles.hint}>{recent.length ? "Your recent favorites" : "A few favorites"}</Text><Pressable ref={trigger} accessibilityRole="button" accessibilityLabel="More emoji" disabled={unavailable} accessibilityState={{ disabled: unavailable }}
      onPress={() => { setQuery(""); setOpen(true); }} style={[styles.more, unavailable && styles.disabled]}><Text style={styles.moreText}>{sending ? "Sending…" : cooldown ? "A little breather…" : "More emoji  +"}</Text></Pressable></View>
    <Modal visible={open} transparent animationType="fade" statusBarTranslucent onRequestClose={() => setOpen(false)} onDismiss={() => trigger.current?.focus()}>
      <KeyboardAvoidingView style={[styles.scrim, { paddingTop: insets.top + 16, paddingBottom: insets.bottom + 16 }]} behavior={Platform.OS === "ios" ? "padding" : undefined}>
        <Pressable accessibilityRole="button" accessibilityLabel="Dismiss emoji picker" style={StyleSheet.absoluteFill} onPress={() => setOpen(false)} />
        <View accessibilityViewIsModal accessibilityLabel="Choose an emoji" style={[styles.panel, { width: panelWidth, height: Math.min(600, height - insets.top - insets.bottom - 32) }]}>
          <View style={styles.heading}><Text accessibilityRole="header" style={styles.title}>Say it your way.</Text><Pressable accessibilityRole="button" accessibilityLabel="Close emoji picker" onPress={() => setOpen(false)} style={styles.close}><Text style={styles.closeText}>×</Text></Pressable></View>
          <TextInput accessibilityLabel="Search emoji" placeholder="Search emoji…" placeholderTextColor={colors.muted} value={query} onChangeText={(value) => { setQuery(value); resetScroll(); }} autoCorrect={false} autoCapitalize="none" style={styles.search} />
          <ScrollView horizontal showsHorizontalScrollIndicator={false} style={styles.categories} contentContainerStyle={{ gap: 6 }} keyboardShouldPersistTaps="handled">{categories.map((item) => <Pressable key={item.id} accessibilityRole="button" accessibilityLabel={`${item.name} emoji`} accessibilityState={{ selected: category === item.id && !query }} onPress={() => { setCategory(item.id); setQuery(""); resetScroll(); }} style={[styles.category, category === item.id && !query && styles.selected]}><Text style={styles.categoryText}>{item.name}</Text></Pressable>)}</ScrollView>
          <View style={styles.tones}>{tones.map((name, index) => <Pressable key={name} accessibilityRole="radio" accessibilityLabel={`${name} skin tone`} accessibilityState={{ checked: tone === index }} onPress={() => setTone(index)} style={[styles.tone, tone === index && styles.selected]}><Text style={{ fontSize: 22 }}>{["👋", "👋🏻", "👋🏼", "👋🏽", "👋🏾", "👋🏿"][index]}</Text></Pressable>)}</View>
          <Text style={styles.resultHint}>{query ? "Search results" : categories.find((item) => item.id === category)?.name} · {results.length}</Text>
          <FlatList ref={list} key={columns} data={results} numColumns={columns} extraData={[tone, unavailable]} keyExtractor={(item) => item.emoji} style={{ flex: 1 }} keyboardShouldPersistTaps="handled" initialNumToRender={48} windowSize={5}
            ListEmptyComponent={<Text style={styles.empty}>No emoji found. Try another word.</Text>}
            renderItem={({ item }) => { const emoji = item.skins?.find((skin) => Array.isArray(skin.tone) ? skin.tone.every((value) => value === tone) : skin.tone === tone)?.emoji ?? item.emoji;
              return <Pressable accessibilityRole="button" accessibilityLabel={`Send ${item.annotation}`} disabled={unavailable} accessibilityState={{ disabled: unavailable }} onPress={() => void send(emoji)} style={[styles.cell, { width: cellWidth }, unavailable && styles.disabled]}><Text style={styles.emoji}>{emoji}</Text></Pressable>;
            }} />
        </View>
      </KeyboardAvoidingView>
    </Modal>
  </>;
}

const createStyles = (colors: AppColors) => StyleSheet.create({
  quick: { flexDirection: "row", gap: 5 },
  reaction: { flex: 1, minHeight: 48, borderRadius: 15, alignItems: "center", justifyContent: "center", backgroundColor: colors.surface, borderWidth: 1, borderColor: colors.line },
  emoji: { fontSize: 26 },
  disabled: { opacity: 0.45 },
  footer: { flexDirection: "row", justifyContent: "space-between", alignItems: "center", marginTop: -8 },
  hint: { fontSize: 11, color: colors.muted },
  more: { minHeight: 44, justifyContent: "center", paddingHorizontal: 8 },
  moreText: { fontSize: 12, fontWeight: "600", color: colors.ink },
  scrim: { flex: 1, backgroundColor: "#15231B77", justifyContent: "center", alignItems: "center" },
  panel: { maxHeight: "100%", backgroundColor: colors.surface, padding: 16, borderRadius: 24 },
  heading: { flexDirection: "row", justifyContent: "space-between", alignItems: "center", marginBottom: 8 },
  title: { fontSize: 19, fontWeight: "600", letterSpacing: -0.5, color: colors.ink },
  close: { width: 44, height: 44, justifyContent: "center", alignItems: "center", margin: -6 },
  closeText: { color: colors.muted, fontSize: 23 },
  search: { minHeight: 44, borderWidth: 1, borderColor: colors.line, borderRadius: 12, paddingHorizontal: 12, color: colors.ink, backgroundColor: colors.background, fontSize: 14 },
  categories: { flexGrow: 0, flexShrink: 0, marginVertical: 10 },
  category: { minHeight: 44, justifyContent: "center", paddingHorizontal: 12, borderRadius: 12 },
  categoryText: { fontSize: 12, color: colors.ink },
  selected: { backgroundColor: colors.soft },
  tones: { flexDirection: "row", gap: 3, marginBottom: 8 },
  tone: { flex: 1, minHeight: 44, justifyContent: "center", alignItems: "center", borderRadius: 12 },
  resultHint: { color: colors.muted, fontSize: 11, marginBottom: 8 },
  cell: { height: 48, justifyContent: "center", alignItems: "center", borderRadius: 12 },
  empty: { color: colors.muted, fontSize: 14, textAlign: "center", paddingVertical: 32 },
});
