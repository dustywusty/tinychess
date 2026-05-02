import type { Color } from "./chess";

export interface StateEvent {
  kind: "state";
  fen: string;
  turn: Color;
  status: string;
  pgn: string;
  uci: string[];
  lastSeen: number;
  watchers: number;
  color?: Color | null;
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
