import { Text, View } from "react-native";
import Svg, { Ellipse, Path } from "react-native-svg";
import type { BotDefinition } from "@yourmove/chess/bots";

// Snobal is a white kitty, and Unicode has no white cat emoji, so draw one.
function WhiteCat({ size }: { size: number }) {
 return <Svg width={size} height={size} viewBox="0 0 36 36">
  <Path d="M5 3.5 L15 10.5 L5.5 19 Z M31 3.5 L21 10.5 L30.5 19 Z" fill="#FFFFFF" stroke="#BFC5CC" strokeWidth={1} strokeLinejoin="round" />
  <Path d="M7.2 7.5 L12.4 11.2 L7.5 15.6 Z M28.8 7.5 L23.6 11.2 L28.5 15.6 Z" fill="#F6B7C3" />
  <Ellipse cx={18} cy={21} rx={14} ry={12} fill="#FFFFFF" stroke="#BFC5CC" strokeWidth={1} />
  <Ellipse cx={12.5} cy={19.5} rx={1.7} ry={2.5} fill="#2F3337" />
  <Ellipse cx={23.5} cy={19.5} rx={1.7} ry={2.5} fill="#2F3337" />
  <Path d="M16.4 23.4 H19.6 L18 25.2 Z" fill="#F08FA4" />
  <Path d="M18 25.2 Q16.8 27.2 15 26.4 M18 25.2 Q19.2 27.2 21 26.4" fill="none" stroke="#7A8088" strokeWidth={0.9} strokeLinecap="round" />
  <Path d="M3 21.5 L9.5 23 M3.5 26 L9.5 24.6 M33 21.5 L26.5 23 M32.5 26 L26.5 24.6" stroke="#9AA1A9" strokeWidth={0.8} strokeLinecap="round" />
 </Svg>;
}

export function BotEmoji({ bot, size }: { bot: BotDefinition; size: number }) {
 if (bot.id === "max") return <View accessible accessibilityRole="image" accessibilityLabel="White cat" testID="snobal-white-cat"><WhiteCat size={size * 1.15} /></View>;
 return <Text style={{ fontSize: size }}>{bot.emoji}</Text>;
}
