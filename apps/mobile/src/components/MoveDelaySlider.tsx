import { useState } from "react";
import { Text, View } from "react-native";
import { useBoardTheme } from "@/lib/theme";
import { useUI } from "./UI";

export function MoveDelaySlider({ value, onChange }: { value: number; onChange: (value: number) => void }) {
 const { colors } = useBoardTheme();
 const ui = useUI();
 const [width, setWidth] = useState(1);
 const update = (value: number) => onChange(Math.max(0, Math.min(5000, Math.round(value / 250) * 250)));
 return <View>
  <Text style={ui.body}>Time between moves · {value === 0 ? "As fast as possible" : `${value / 1000} s`}</Text>
  <View accessible accessibilityRole="adjustable" accessibilityLabel="Time between moves"
   accessibilityValue={{ min: 0, max: 5000, now: value, text: `${value / 1000} seconds` }}
   accessibilityActions={[{ name: "increment" }, { name: "decrement" }]}
   onAccessibilityAction={event => update(value + (event.nativeEvent.actionName === "increment" ? 250 : -250))}
   onLayout={event => setWidth(event.nativeEvent.layout.width)}
   onStartShouldSetResponder={() => true} onMoveShouldSetResponder={() => true}
   onResponderGrant={event => update((event.nativeEvent.locationX - 12) / Math.max(1, width - 24) * 5000)}
   onResponderMove={event => update((event.nativeEvent.locationX - 12) / Math.max(1, width - 24) * 5000)}
   style={{ height: 44, justifyContent: "center", paddingHorizontal: 12 }}>
   <View pointerEvents="none" style={{ height: 4, backgroundColor: colors.line, borderRadius: 2 }}>
    <View style={{ width: `${value / 50}%`, height: 4, backgroundColor: colors.ink }} />
    <View style={{ position: "absolute", left: `${value / 50}%`, marginLeft: -12, top: -10, width: 24, height: 24, borderRadius: 12, backgroundColor: colors.ink }} />
   </View>
  </View>
 </View>;
}
