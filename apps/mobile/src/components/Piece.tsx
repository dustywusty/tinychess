import Svg, { Circle, G, Path } from "react-native-svg";
import { Platform } from "react-native";

// Original vector pieces keep the same silhouette on iOS and Android.
export function Piece({ piece, size = 40 }: { piece: string; size?: number }) {
  const white = piece === piece.toUpperCase();
  const fill = white ? "#FFFEF7" : "#283C35";
  const stroke = white ? "#43584E" : "#182C24";
  const type = piece.toLowerCase();
  return <Svg width={size} height={size} viewBox="0 0 48 48" {...Platform.select({ web: { "aria-hidden": true }, default: { accessible: false } })}>
    <G fill={fill} stroke={stroke} strokeWidth={1.6} strokeLinejoin="round" strokeLinecap="round">
      {type === "p" && <><Circle cx="24" cy="13" r="6" /><Path d="M20 19h8v6c0 5 2 7 5 10H15c3-3 5-5 5-10Z" /></>}
      {type === "r" && <Path d="M13 8h6v5h3V8h4v5h3V8h6v12l-5 4 2 11H16l2-11-5-4Z M16 20h16" />}
      {type === "b" && <><Path d="M24 6c-3 5-10 9-10 15 0 4 4 7 7 8l-5 6h16l-5-6c3-1 7-4 7-8 0-6-7-10-10-15Z" /><Path d="m26 13-5 8M19 29h10" /></>}
      {type === "n" && <><Path d="m15 35 3-11-6 2-4-5 10-10 2-6 6 5c12 3 12 12 9 25Z" /><Path d="m18 24 8-6M20 10l4 3" /><Circle cx="19" cy="17" r="1" fill={stroke} /></>}
      {type === "q" && <><Path d="m12 15 6 7 6-11 6 11 6-7-5 20H17Z M17 29h14" /><Circle cx="11" cy="12" r="2.5" /><Circle cx="24" cy="8" r="2.5" /><Circle cx="37" cy="12" r="2.5" /></>}
      {type === "k" && <><Path d="M24 5v10m-4-6h8" /><Path d="M24 18c-9-9-17 1-8 10l2 7h12l2-7c9-9 1-19-8-10Z M17 28h14" /></>}
      <Path d="M15 35h18l3 6H12Z" />
    </G>
  </Svg>;
}
