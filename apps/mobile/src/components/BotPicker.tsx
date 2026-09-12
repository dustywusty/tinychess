import { useState } from "react";
import { Modal, Pressable, ScrollView, Text, View } from "react-native";
import { bots, type BotId } from "@yourmove/chess/bots";
import { useBoardTheme } from "@/lib/theme";
import { Button, useUI } from "./UI";

export function BotPicker({ open, busy, onClose, onStart }: { open: boolean; busy: boolean; onClose: () => void; onStart: (id: BotId, color: "w" | "b") => void }) {
 const { colors } = useBoardTheme();
 const ui = useUI();
 const [selected, setSelected] = useState<BotId>("pip");
 const [color, setColor] = useState<"w" | "b">("w");
 return <Modal visible={open} transparent animationType="fade" onRequestClose={onClose}>
  <View style={{ flex: 1, backgroundColor: "#252B2877", justifyContent: "center", padding: 20 }}>
   <ScrollView accessibilityViewIsModal style={{ flexGrow: 0, width: "100%", maxWidth: 440, alignSelf: "center", backgroundColor: colors.background, borderRadius: 26 }} contentContainerStyle={{ padding: 24, gap: 12 }}>
    <View style={ui.row}><Text accessibilityRole="header" style={ui.cardTitle}>Meet your next opponent.</Text><Pressable accessibilityRole="button" accessibilityLabel="Close opponent picker" onPress={onClose} style={{ padding: 12 }}><Text style={{ color: colors.ink }}>✕</Text></Pressable></View>
    {bots.map(bot => <Pressable key={bot.id} accessibilityRole="radio" aria-checked={selected === bot.id} accessibilityState={{ checked: selected === bot.id }} accessibilityLabel={`${bot.name}, ${bot.description}`} onPress={() => setSelected(bot.id)}
     style={{ flexDirection: "row", alignItems: "center", gap: 14, padding: 14, borderRadius: 16, backgroundColor: selected === bot.id ? colors.soft : colors.surface, borderWidth: 1, borderColor: colors.line }}>
     <Text style={{ fontSize: 28 }}>{bot.emoji}</Text><View style={{ flex: 1 }}><Text style={{ color: colors.ink, fontWeight: "600", fontSize: 16 }}>{bot.name}</Text><Text style={ui.body}>{bot.description}</Text></View><Text style={{ color: colors.ink }}>{selected === bot.id ? "●" : "○"}</Text>
    </Pressable>)}
    <Text style={ui.eyebrow}>YOUR PIECES</Text><View style={{ flexDirection: "row", gap: 12 }}>{(["w", "b"] as const).map(value => <Pressable key={value} accessibilityRole="radio" aria-checked={color === value} accessibilityState={{ checked: color === value }} onPress={() => setColor(value)} style={{ padding: 12 }}><Text style={{ color: colors.ink }}>{color === value ? "●" : "○"} {value === "w" ? "White" : "Black"}</Text></Pressable>)}</View>
    <Button title="Let’s play ↗" busy={busy} onPress={() => onStart(selected, color)} />
   </ScrollView>
  </View>
 </Modal>;
}
