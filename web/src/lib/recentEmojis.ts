const KEY = "tinychess:recentEmojis:v1";
const MAX = 15;

export function loadRecentEmojis(): string[] {
  try {
    const raw = localStorage.getItem(KEY);
    if (!raw) return [];
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed.filter((s): s is string => typeof s === "string").slice(0, MAX);
  } catch {
    return [];
  }
}

export function rememberEmoji(emoji: string): string[] {
  const current = loadRecentEmojis();
  const next = [emoji, ...current.filter((e) => e !== emoji)].slice(0, MAX);
  try {
    localStorage.setItem(KEY, JSON.stringify(next));
  } catch {
    /* ignore */
  }
  return next;
}
