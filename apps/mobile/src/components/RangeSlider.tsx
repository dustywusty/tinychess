import { useState } from "react";
import { Text, View } from "react-native";
import { useBoardTheme } from "@/lib/theme";
import { useUI } from "./UI";

export type RangeSliderProps = { label: string; valueLabel: string; valueText: string; min: number; max: number; step: number; value: number; onChange: (value: number) => void };

export function RangeSlider({ label, valueLabel, valueText, min, max, step, value, onChange }: RangeSliderProps) {
 const { colors } = useBoardTheme();
 const ui = useUI();
 const [width, setWidth] = useState(1);
 const update = (value: number) => onChange(Math.max(min, Math.min(max, min + Math.round((value - min) / step) * step)));
 const atX = (x: number) => update(min + (x - 12) / Math.max(1, width - 24) * (max - min));
 const percent = (value - min) / (max - min) * 100;
 // One dot per ~10px of track, drawn over both the fill and the remainder.
 const dots = Math.max(0, Math.floor((width - 24) / 10));
 return <View style={{ gap: 4 }}>
  <Text style={ui.eyebrow}>{label.toUpperCase()} · {valueLabel}</Text>
  <View accessible accessibilityRole="adjustable" accessibilityLabel={label}
   accessibilityValue={{ min, max, now: value, text: valueText }}
   accessibilityActions={[{ name: "increment" }, { name: "decrement" }]}
   onAccessibilityAction={event => update(value + (event.nativeEvent.actionName === "increment" ? step : -step))}
   onLayout={event => setWidth(event.nativeEvent.layout.width)}
   onStartShouldSetResponder={() => true} onMoveShouldSetResponder={() => true}
   onResponderGrant={event => atX(event.nativeEvent.locationX)}
   onResponderMove={event => atX(event.nativeEvent.locationX)}
   style={{ height: 48, justifyContent: "center", paddingHorizontal: 12 }}>
   <View pointerEvents="none" style={{ height: 14, borderRadius: 7, backgroundColor: colors.lilac, overflow: "hidden" }}>
    <View style={{ width: `${percent}%`, height: 14, backgroundColor: colors.mint }} />
    <View style={{ position: "absolute", left: 4, right: 4, top: 5, flexDirection: "row", justifyContent: "space-between" }}>
     {Array.from({ length: dots }, (_, index) => <View key={index} style={{ width: 4, height: 4, borderRadius: 2, backgroundColor: index / dots * 100 < percent ? "#252B2866" : colors.ink + "66" }} />)}
    </View>
   </View>
   <View pointerEvents="none" style={{ position: "absolute", top: 9, left: 12 + (width - 24) * percent / 100 - 3, width: 6, height: 30, borderRadius: 3, backgroundColor: colors.mint, borderWidth: 2, borderColor: colors.background }} />
  </View>
 </View>;
}
