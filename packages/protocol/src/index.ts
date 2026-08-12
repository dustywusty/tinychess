export const protocolVersion = 1 as const;

export type WireColor = "w" | "b" | "white" | "black";
export type PlayerRole = "player" | "spectator";

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
  role?: PlayerRole;
  clientId?: string;
}

export interface EmojiEvent {
  kind: "emoji";
  emoji: string;
  at: number;
  sender: string;
}

export type ServerEvent = StateEvent | EmojiEvent | Record<string, never>;

export interface CreateGameResponse {
  id: string;
}

export interface CommandResponse {
  ok: boolean;
  error?: string;
}

export interface MoveResponse extends CommandResponse {
  state?: Omit<StateEvent, "kind">;
}

export const coachIntents = [
  "HINT",
  "ANOTHER_HINT",
  "WHY_BAD",
  "WHY_GOOD",
  "EXPLAIN_POSITION",
  "EXPLAIN_LAST_MOVE",
  "WHAT_SHOULD_I_NOTICE",
  "CHESS_RULE",
  "OFF_TOPIC",
  "REPEATED",
] as const;

export type CoachIntent = (typeof coachIntents)[number];

export function isStateEvent(value: unknown): value is StateEvent {
  return (
    typeof value === "object" &&
    value !== null &&
    (value as { kind?: unknown }).kind === "state"
  );
}

export function isEmojiEvent(value: unknown): value is EmojiEvent {
  return (
    typeof value === "object" &&
    value !== null &&
    (value as { kind?: unknown }).kind === "emoji"
  );
}
