export const commonEmojis = [
  { emoji: "👋", label: "Wave" }, { emoji: "🤔", label: "Thinking" }, { emoji: "🔥", label: "Fire" },
  { emoji: "😂", label: "Laugh" }, { emoji: "👏", label: "Applause" }, { emoji: "🤝", label: "Good game" },
];
export function parseRecentEmojis(raw: string | null): string[] {
  try {
    const value: unknown = JSON.parse(raw ?? "[]");
    return Array.isArray(value) ? [...new Set(value.filter((item): item is string => typeof item === "string" && item.trim().length > 0 && item.length <= 64))].slice(0, 18) : [];
  } catch { return []; }
}
export function withRecentEmoji(recent: string[], emoji: string): string[] {
  return [emoji, ...recent.filter((item) => item !== emoji)].slice(0, 18);
}
export function quickEmojis(recent: string[]): string[] {
  return [...new Set([...recent, ...commonEmojis.map((item) => item.emoji)])].slice(0, 6);
}
