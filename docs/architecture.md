# Architecture

Your Move optimizes for one path: create game -> share `/g/{id}` -> play or
watch -> optionally ask a bounded chess question.

## Components

- `internal/` is the current Go API and authoritative game domain. It owns
  seats, validates moves with `corentings/chess/v2`, produces FEN/PGN, and
  publishes accepted state. It remains at the repository root during the
  runnable migration and will move to `apps/api` after its package boundaries
  stop referencing legacy paths.
- `web/` is the React/Vite fallback and link-opening client. It was adapted from
  the tested `react-coach` branch. It keeps SSE during transport migration.
- `apps/mobile/` is an Expo SDK 56 Router application for iOS and Android. Its route
  `g/[id]` matches web links and the `yourmove://` scheme.
- `packages/protocol/` owns TypeScript command/event and coach-intent shapes used
  by both clients.
- `packages/coach/` owns deterministic server policy. It has no model SDK and no
  UI dependency.
- Postgres stores games, moves, results, anonymous seats, analysis, and usage.
  The current GORM persistence is an intermediate implementation.

## Authority and data flow

```text
Web / Mobile command
        |
        v
Go game process -> legal-move validation -> durable event/state
        |                                      |
        +---------- versioned event -----------+
        |
        v
coach policy -> Stockfish -> cached artifacts -> optional LLM explanation
```

Clients can preview board interaction but cannot declare a position or result.
The API publishes canonical snapshots only after validation.

## Key decisions

### Go over Phoenix

Go already supplies a tested authoritative chess loop and can model each active
game as one command-owning goroutine. Phoenix would be reasonable for a team
that prefers Elixir or immediately needs distributed Presence/PubSub, but it is
not necessary for the first production scale and would discard working domain
tests.

### Expo and React

Expo Router gives native iOS/Android builds and route-based deep linking while
React keeps protocol and interaction logic close to the web client. UI remains
platform-native; only pure models and wire types are shared.

### Transport migration

Web keeps SSE plus POST commands so the application remains deployable. Mobile
currently polls a snapshot endpoint every 1.5 seconds. The next transport step
is one versioned WebSocket subscription supported by both clients, with snapshot
recovery and monotonic sequence numbers. Snapshot polling is not the final
realtime design.

### CopilotKit

CopilotKit currently supports React Native, including Expo, through a provider,
headless hooks, and optional UI over AG-UI. It is viable as a later presentation
adapter. The product will not expose its general chat surface or let its runtime
bypass server coach policy. The constrained HTTP policy/analysis API lands first.

## Deferred production work

- Durable anonymous session tokens and seat recovery
- WebSocket replay/resume and multi-replica fanout
- Explicit database migrations and restart recovery
- Server-side Stockfish worker pool
- Provider-backed artifact generation and usage ledger
- Universal/app-link association files for the real domain
- Subscription receipt validation and entitlements
