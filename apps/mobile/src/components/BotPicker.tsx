import { useState } from "react";
import { Modal, Pressable, ScrollView, Switch, Text, View } from "react-native";
import { bots, MAX_BOT_ELO, MIN_BOT_ELO, type BotId } from "@yourmove/chess/bots";
import { useBoardTheme } from "@/lib/theme";
import { Button, useUI } from "./UI";
import type { BotSettings } from "@yourmove/protocol";
import { RangeSlider } from "./RangeSlider";

export function BotPicker({ open, busy, onClose, onStart }: { open: boolean; busy: boolean; onClose: () => void; onStart: (id: BotId, color: "w" | "b", settings: BotSettings) => void }) {
 const { colors } = useBoardTheme();
 const ui = useUI();
 const [selected, setSelected] = useState<BotId>("pip");
 const [color, setColor] = useState<"w" | "b">("w");
 const [custom, setCustom] = useState(false);
 const [elo, setElo] = useState(1500);
 const [advanced, setAdvanced] = useState(false);
 const [battle, setBattle] = useState(false);
 const [whiteBot, setWhiteBot] = useState<BotId>("pip");
 const [delay, setDelay] = useState<number>();
 const pace = delay ?? 1500;
 // Bottom sheet: tap the dimmed area above it to close.
 return <Modal visible={open} transparent animationType="slide" onRequestClose={onClose}>
  <View style={{ flex: 1, backgroundColor: "#252B2877", justifyContent: "flex-end" }}>
   <Pressable accessibilityRole="button" accessibilityLabel="Close opponent picker" onPress={onClose} style={{ flex: 1, minHeight: 40 }} />
   <ScrollView accessibilityViewIsModal style={{ flexGrow: 0, maxHeight: "92%", width: "100%", maxWidth: 560, alignSelf: "center", backgroundColor: colors.background, borderTopLeftRadius: 28, borderTopRightRadius: 28 }} contentContainerStyle={{ padding: 24, paddingTop: 12, gap: 12 }}>
    <View style={{ alignSelf: "center", width: 40, height: 5, borderRadius: 3, backgroundColor: colors.line, marginBottom: 8 }} />
    <Text accessibilityRole="header" style={{ color: colors.ink, fontSize: 30, fontWeight: "700", letterSpacing: -1 }}>{battle ? "Watch a bot battle." : "A worthy opponent."}</Text>
    {!battle && <Text style={ui.body}>Five personalities. Practice anywhere, even offline.</Text>}
    {!battle && bots.map(bot => <Pressable key={bot.id} accessibilityRole="radio" aria-checked={selected === bot.id} accessibilityState={{ checked: selected === bot.id }} accessibilityLabel={`${bot.name}, ${bot.description}`} onPress={() => setSelected(bot.id)}
     style={{ flexDirection: "row", alignItems: "center", gap: 14, padding: 14, borderRadius: 16, backgroundColor: selected === bot.id ? colors.soft : colors.surface, borderWidth: 1, borderColor: colors.line }}>
     <Text style={{ fontSize: 28 }}>{bot.emoji}</Text><View style={{ flex: 1 }}><Text style={{ color: colors.ink, fontWeight: "600", fontSize: 16 }}>{bot.name}</Text><Text style={ui.body}>{bot.description}</Text></View><Text style={{ color: colors.ink }}>{selected === bot.id ? "●" : "○"}</Text>
    </Pressable>)}
    {!battle && <><Text style={[ui.eyebrow, { marginTop: 8 }]}>YOUR PIECES</Text><View style={{ flexDirection: "row", gap: 12 }}>{(["w", "b"] as const).map(value => <Pressable key={value} accessibilityRole="radio" aria-checked={color === value} accessibilityState={{ checked: color === value }} onPress={() => setColor(value)}
     style={{ flexDirection: "row", alignItems: "center", gap: 8, minHeight: 48, paddingHorizontal: 20, borderRadius: 12, borderWidth: 1, borderColor: color === value ? colors.lilac : colors.line, backgroundColor: color === value ? colors.lilac : "transparent" }}>
     <Text style={{ color: colors.ink, fontSize: 18 }}>{value === "w" ? "♔" : "♚"}</Text><Text style={{ color: colors.ink, fontWeight: "600", fontSize: 16 }}>{value === "w" ? "White" : "Black"}</Text>
    </Pressable>)}</View></>}
    {!battle && <View style={[ui.row, { minHeight: 48 }]}>
     <Text style={{ color: colors.ink, fontSize: 16 }}>Custom difficulty</Text>
     <Switch accessibilityLabel="Custom difficulty" value={custom} onValueChange={setCustom} trackColor={{ false: colors.soft, true: colors.mint }} thumbColor={custom ? colors.buttonInk : colors.muted} />
    </View>}
    {!battle && custom && <>
     <RangeSlider label="Elo target" valueLabel={String(elo)} valueText={`Elo ${elo}`} min={MIN_BOT_ELO} max={MAX_BOT_ELO} step={50} value={elo} onChange={setElo} />
     <Text style={ui.body}>Approximate engine strength, not a measured player rating. Uses the chosen strength instead of personality difficulty.</Text>
    </>}
    <Pressable accessibilityRole="button" accessibilityState={{ expanded: advanced }} onPress={() => setAdvanced(value => !value)} style={{ paddingVertical: 12 }}><Text style={{ color: colors.ink, fontWeight: "600", fontSize: 16 }}>Advanced settings {advanced ? "⌃" : "⌄"}</Text></Pressable>
    {advanced && <View style={{ gap: 12 }}>
     <Pressable accessibilityRole="checkbox" accessibilityLabel="Bot vs. bot" aria-checked={battle} accessibilityState={{ checked: battle }} onPress={() => setBattle(value => !value)} style={{ flexDirection: "row", alignItems: "center", gap: 14, paddingVertical: 12 }}>
      <View style={{ width: 22, height: 22, borderRadius: 4, borderWidth: 2, borderColor: battle ? colors.mint : colors.muted, backgroundColor: battle ? colors.mint : "transparent", alignItems: "center", justifyContent: "center" }}>{battle && <Text style={{ color: colors.buttonInk, fontWeight: "800", fontSize: 14 }}>✓</Text>}</View>
      <Text style={{ color: colors.ink, fontSize: 16 }}>Bot vs. bot</Text>
     </Pressable>
     <Text style={ui.body}>Watch two computer opponents play a test game.</Text>
     {battle && [{ label: "White bot", value: whiteBot, set: setWhiteBot }, { label: "Black bot", value: selected, set: setSelected }].map(side => <View key={side.label} style={{ gap: 8 }}><Text style={ui.eyebrow}>{side.label.toUpperCase()}</Text><View style={{ flexDirection: "row", flexWrap: "wrap", gap: 8 }}>{bots.map(bot => <Pressable key={bot.id} accessibilityRole="radio" accessibilityLabel={`${side.label}: ${bot.name}`} accessibilityState={{ checked: side.value === bot.id }} onPress={() => side.set(bot.id)} style={{ padding: 12, borderRadius: 12, backgroundColor: side.value === bot.id ? colors.soft : colors.surface }}><Text style={{ color: colors.ink }}>{bot.emoji} {bot.name} {side.value === bot.id ? "●" : "○"}</Text></Pressable>)}</View></View>)}
     <RangeSlider label="Minimum move interval" valueLabel={pace === 0 ? "AS FAST AS POSSIBLE" : `${pace / 1000}S`} valueText={`${pace / 1000} seconds`} min={0} max={5000} step={250} value={pace} onChange={setDelay} />
     <Text style={ui.body}>Thinking may take longer.{!battle && delay === undefined ? " Regular games use each bot’s natural pace until you adjust this." : ""}</Text>
    </View>}
    <Button title={battle ? "Start bot battle ↗" : "Let’s play ↗"} busy={busy} onPress={() => onStart(selected, battle ? "w" : color, { playerBotId: battle ? whiteBot : undefined, moveDelayMs: delay ?? (battle ? 1500 : undefined), targetElo: !battle && custom ? elo : undefined })} />
   </ScrollView>
  </View>
 </Modal>;
}
