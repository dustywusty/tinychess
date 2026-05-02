# Phase 1 parity checklist

The current playable UI is inline React in `internal/templates/game.html`
(1820 lines). This file enumerates every behavior the new SPA must preserve
before the cutover (PR 6).

## Storage keys (must NOT change)

- `sessionStorage["tinychess:clientId"]` — client UUID, generated on first
  load if absent. The e2e tests read this directly.
- `localStorage["accent"]` — board accent color hex.
- `localStorage["theme"]` — `"dark"` or `"light"`.
- `localStorage["tinychess:games:v1"]` — recent games map (`{ [gameId]: { id, createdAt, lastSeen, lastSeenLocal, status, result } }`).
- `localStorage["tinychess:recentEmojis:v1"]` — recent emoji array (max 15).

## Routes (after PR 1)

- `GET /` — home page (currently template; SPA after cutover).
- `GET /:gameId` — game page (currently template; SPA after cutover).
- `GET /new` — legacy redirect to `/:newGameId`.
- `POST /api/games` — create game, returns `{ id }`.
- `GET /api/sse/:gameId?clientId=…` — SSE stream.
- `POST /api/games/:gameId/move` — `{ uci, clientId }` → `{ ok, error?, state }`.
- `POST /api/games/:gameId/react` — `{ emoji, sender }` → `{ ok, error? }`.
- `POST /api/games/:gameId/release` — owner-only client kick.

## SSE event union

Discriminate on `kind`. Heartbeats are empty `{}`.

```ts
type ServerEvent =
  | {
      kind: "state";
      fen: string;
      turn: "white" | "black";
      status: string;
      pgn: string;
      uci: string[];
      lastSeen: number;
      watchers: number;
      color?: "white" | "black" | null;
      role?: "player" | "spectator";
      clientId?: string;
    }
  | { kind: "emoji"; emoji: string; at: number; sender: string }
  | {}; // heartbeat
```

## Behaviors to preserve

### Board
- 8x8 grid, rank 8 top for white / rank 1 top for black.
- `data-square="<algebraic>"` on each cell (e2e selector).
- Two-tap selection: first click sets selected, second click submits move.
- Drag/drop also supported.
- `.sel` outline on selected cell, `.last-move` shadow on prior move's from+to.
- Last move squares derived from the last entry in the UCI array.
- Pawn promotion auto-defaults to queen (server appends).
- Optimistic move: client applies locally, rolls back on `ok: false`.
- Captured pieces UI computed from FEN (`capturedFromFEN`).
- Piece glyphs are Unicode chess symbols (U+2659–U+265A).

### Game status
- `#turn` reads "Your turn" / "Their turn" or game-over text. (e2e selector)
- `#status` for errors and game-over reasons. (e2e selector)
- `#board` exists and is visible. (e2e selector)

### Emoji reactions
- `#reactbtn` button opens `<emoji-picker>` modal. (e2e selectors)
- 5s cooldown, enforced both client-side and server-side.
- Recent 15 emojis pinned as buttons under the picker (`#recent-emojis`).
- Animations: `.burst` (scale 0.4→1, 0.28s) on the sender icon and
  `.big-emoji` (scale 4→0.6, opacity 1→0, 1.2s) center-screen on every client.
- Emoji-picker-element web component fires `emoji-click` events.

### Spectator mode
- No color assigned → `role: "spectator"`.
- Cannot send moves; can send reactions.
- Board still renders live; "Release seat" hidden.

### Share
- `navigator.clipboard.writeText(location.href)` on copy button click.

### Theme
- 5 accent swatches + light/dark toggle in header. Header collapses to
  hamburger menu below 720px.
- CSS custom properties (`--accent`, `--bg`, `--panel`, `--text`, `--sq1/2/3`)
  with `color-mix(in oklab, …)`. Port verbatim into `index.css`.

### Home page
- "New game" button → `/new` (legacy redirect, server creates UUID and 302s).
- If active game exists in localStorage, "Open most recent" instead.
- Recent games list with Open / Copy Link / Forget actions.
- Theme picker mirrored from game page.

### Reconnect
- SSE EventSource reconnects with backoff on disconnect.
- On reconnect, server sends initial ClientState — client re-syncs.

## Removed in Phase 1

- The live `/coach` endpoint (OpenAI gpt-4o-mini via Hashbrown). Replaced in
  Phase 2 by `/api/coach/chat` (provider-configurable, Anthropic default) for
  post-game review only.
- `{{COACH_ENABLED}}` and `{{COACH_URL}}` template substitutions.
- Coach panel UI (`.coach-toggle`, `.coach-panel`, `.coach-messages`,
  `.coach-input`).
