import { useState } from "react";
import { Modal, Pressable, ScrollView, Text, View } from "react-native";
import { bots, type BotId } from "@yourmove/chess/bots";
import { useBoardTheme } from "@/lib/theme";
import { Button, useUI } from "./UI";
import type { BotSettings } from "@yourmove/protocol";
import { MoveDelaySlider } from "./MoveDelaySlider";

export function BotPicker({ open, busy, onClose, onStart }: { open: boolean; busy: boolean; onClose: () => void; onStart: (id: BotId, color: "w" | "b", settings: BotSettings) => void }) {
 const { colors } = useBoardTheme();
 const ui = useUI();
 const [selected, setSelected] = useState<BotId>("pip");
 const [color, setColor] = useState<"w" | "b">("w");
 const [advanced, setAdvanced] = useState(false);
 const [battle, setBattle] = useState(false);
 const [whiteBot, setWhiteBot] = useState<BotId>("pip");
 const [delay, setDelay] = useState<number>();
 return <Modal visible={open} transparent animationType="fade" onRequestClose={onClose}>
  <View style={{ flex: 1, backgroundColor: "#252B2877", justifyContent: "center", padding: 20 }}>
   <ScrollView accessibilityViewIsModal style={{ flexGrow: 0, width: "100%", maxWidth: 440, alignSelf: "center", backgroundColor: colors.background, borderRadius: 26 }} contentContainerStyle={{ padding: 24, gap: 12 }}>
    <View style={ui.row}><Text accessibilityRole="header" style={ui.cardTitle}>{battle ? "Watch a bot battle." : "Meet your next opponent."}</Text><Pressable accessibilityRole="button" accessibilityLabel="Close opponent picker" onPress={onClose} style={{ padding: 12 }}><Text style={{ color: colors.ink }}>✕</Text></Pressable></View>
    {!battle && bots.map(bot => <Pressable key={bot.id} accessibilityRole="radio" aria-checked={selected === bot.id} accessibilityState={{ checked: selected === bot.id }} accessibilityLabel={`${bot.name}, ${bot.description}`} onPress={() => setSelected(bot.id)}
     style={{ flexDirection: "row", alignItems: "center", gap: 14, padding: 14, borderRadius: 16, backgroundColor: selected === bot.id ? colors.soft : colors.surface, borderWidth: 1, borderColor: colors.line }}>
     <Text style={{ fontSize: 28 }}>{bot.emoji}</Text><View style={{ flex: 1 }}><Text style={{ color: colors.ink, fontWeight: "600", fontSize: 16 }}>{bot.name}</Text><Text style={ui.body}>{bot.description}</Text></View><Text style={{ color: colors.ink }}>{selected === bot.id ? "●" : "○"}</Text>
    </Pressable>)}
    {!battle && <><Text style={ui.eyebrow}>YOUR PIECES</Text><View style={{ flexDirection: "row", gap: 12 }}>{(["w", "b"] as const).map(value => <Pressable key={value} accessibilityRole="radio" aria-checked={color === value} accessibilityState={{ checked: color === value }} onPress={() => setColor(value)} style={{ padding: 12 }}><Text style={{ color: colors.ink }}>{color === value ? "●" : "○"} {value === "w" ? "White" : "Black"}</Text></Pressable>)}</View></>}
    <Pressable accessibilityRole="button" accessibilityState={{ expanded: advanced }} onPress={() => setAdvanced(value => !value)} style={{ paddingVertical: 12 }}><Text style={{ color: colors.ink, fontWeight: "600" }}>Advanced settings {advanced ? "⌃" : "⌄"}</Text></Pressable>
    {advanced && <View style={{ gap: 12 }}>
     <Pressable accessibilityRole="checkbox" accessibilityState={{ checked: battle }} onPress={() => setBattle(value => !value)} style={{ paddingVertical: 12 }}><Text style={{ color: colors.ink }}>{battle ? "☑" : "☐"} Bot vs. bot</Text></Pressable>
     <Text style={ui.body}>Watch two computer opponents play a test game.</Text>
     {battle && [{ label: "White bot", value: whiteBot, set: setWhiteBot }, { label: "Black bot", value: selected, set: setSelected }].map(side => <View key={side.label} style={{ gap: 8 }}><Text style={ui.eyebrow}>{side.label.toUpperCase()}</Text><View style={{ flexDirection: "row", flexWrap: "wrap", gap: 8 }}>{bots.map(bot => <Pressable key={bot.id} accessibilityRole="radio" accessibilityLabel={`${side.label}: ${bot.name}`} accessibilityState={{ checked: side.value === bot.id }} onPress={() => side.set(bot.id)} style={{ padding: 12, borderRadius: 12, backgroundColor: side.value === bot.id ? colors.soft : colors.surface }}><Text style={{ color: colors.ink }}>{bot.emoji} {bot.name} {side.value === bot.id ? "●" : "○"}</Text></Pressable>)}</View></View>)}
     <MoveDelaySlider value={delay ?? 1500} onChange={setDelay} />
     <Text style={ui.body}>Minimum interval; thinking may take longer.{!battle && delay === undefined ? " Regular games use each bot’s natural pace until you adjust this." : ""}</Text>
    </View>}
    <Button title={battle ? "Start bot battle ↗" : "Let’s play ↗"} busy={busy} onPress={() => onStart(selected, battle ? "w" : color, { playerBotId: battle ? whiteBot : undefined, moveDelayMs: delay ?? (battle ? 1500 : undefined) })} />
   </ScrollView>
  </View>
 </Modal>;
}
