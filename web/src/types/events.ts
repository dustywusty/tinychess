// The server emits "w"/"b" today (chess.Color.String()). This wire-level type
// is tolerant; gameStore normalizes via normalizeColor before storing.
export type WireColor = "w" | "b" | "white" | "black";

export interface StateEvent {
  kind: "state";
  fen: string;
  turn: WireColor;
  status: string;
  pgn: string;
  uci: string[];
  lastSeen: number;
  watchers: number;
  color?: WireColor | null;
  role?: "player" | "spectator";
  clientId?: string;
}

export interface EmojiEvent {
  kind: "emoji";
  emoji: string;
  at: number;
  sender: string;
}

export type ServerEvent = StateEvent | EmojiEvent | Record<string, never>;

export function isStateEvent(ev: unknown): ev is StateEvent {
  return (
    typeof ev === "object" &&
    ev !== null &&
    (ev as { kind?: unknown }).kind === "state"
  );
}

export function isEmojiEvent(ev: unknown): ev is EmojiEvent {
  return (
    typeof ev === "object" &&
    ev !== null &&
    (ev as { kind?: unknown }).kind === "emoji"
  );
}
