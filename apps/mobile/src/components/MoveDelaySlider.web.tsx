import { useBoardTheme } from "@/lib/theme";

export function MoveDelaySlider({ value, onChange }: { value: number; onChange: (value: number) => void }) {
 const { colors } = useBoardTheme();
 return <label style={{ color: colors.ink, fontSize: 14 }}>
  Time between moves · {value === 0 ? "As fast as possible" : `${value / 1000} s`}
  <input type="range" min="0" max="5000" step="250" value={value}
   aria-label="Time between moves" aria-valuetext={`${value / 1000} seconds`}
   onChange={event => onChange(Number(event.target.value))}
   style={{ display: "block", width: "100%", minHeight: 44, margin: 0, accentColor: colors.ink }} />
 </label>;
}
