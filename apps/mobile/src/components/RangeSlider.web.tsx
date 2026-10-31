import { useBoardTheme } from "@/lib/theme";
import type { RangeSliderProps } from "./RangeSlider";

export function RangeSlider({ label, valueLabel, valueText, min, max, step, value, onChange }: RangeSliderProps) {
 const { colors } = useBoardTheme();
 return <label style={{ color: colors.muted, fontSize: 10, fontWeight: 700, letterSpacing: 1.6 }}>
  {label.toUpperCase()} · {valueLabel}
  <input type="range" min={min} max={max} step={step} value={value}
   aria-label={label} aria-valuetext={valueText}
   onChange={event => onChange(Number(event.target.value))}
   style={{ display: "block", width: "100%", minHeight: 44, margin: 0, accentColor: colors.mint }} />
 </label>;
}
