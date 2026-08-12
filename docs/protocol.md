# Protocol

`packages/protocol` is the client-side source for version 1 wire types. The Go
API currently has matching structs; JSON Schema/OpenAPI generation will replace
that manual pairing as the API stabilizes.

## Routes in the migration API

- `POST /api/games` -> `{ "id": "uuid" }`
- `GET /api/games/{id}/snapshot?clientId=...` -> current state plus role/color
- `GET /api/sse/{id}?clientId=...` -> state, emoji, and heartbeat events
- `POST /api/games/{id}/move` with `{ uci, clientId }`
- `POST /api/games/{id}/react` with `{ emoji, sender }`
- `POST /api/games/{id}/release` with `{ clientId, targetId }`

Web routes use `/g/{id}`. Legacy `/{id}` routes remain readable during rollout.
The snapshot endpoint temporarily assigns anonymous seats for mobile parity.

## State event

```json
{
  "kind": "state",
  "fen": "...",
  "turn": "w",
  "status": "",
  "pgn": "...",
  "uci": ["e2e4"],
  "lastSeen": 0,
  "watchers": 2,
  "color": "w",
  "role": "player",
  "clientId": "opaque-session"
}
```

The API currently emits both chess-library colors (`w`/`b`) and tolerates long
colors in clients. Version 2 will normalize colors, add game lifecycle/result,
event sequence, server timestamp, and protocol version.

## Next WebSocket protocol

Commands and events will use explicit envelopes:

```json
{ "v": 1, "type": "move.submit", "requestId": "...", "gameId": "...", "payload": { "uci": "e2e4" } }
```

```json
{ "v": 1, "type": "game.state", "seq": 14, "gameId": "...", "payload": {} }
```

On reconnect, a client sends its last sequence. The server either replays newer
events or returns a full snapshot. Anonymous identity moves from a caller-chosen
ID to an opaque signed token before this protocol is exposed publicly.
