# Your Move rebuild plan

Date: 2026-08-12

## Executive assessment

Tiny Chess is a working proof of concept for the core product loop: one person
opens a generated URL, a second person takes the other seat, later visitors
spectate, moves are validated on the server, and emoji reactions are broadcast
in realtime. The current `main` branch is a small Go application with an inline
React web page. It is a useful migration baseline, not a suitable mobile or AI
architecture.

The unmerged `origin/dustywusty/react-coach` branch is the most recent product
work. It replaces the inline page with a typed React/Vite SPA, adds API-shaped
routes, a real storage interface, game/evaluation persistence, a Stockfish WASM
worker, and a Hashbrown-based post-game review experiment. Its web, Go, and
browser tests pass. We should adapt the SPA, storage boundary, and Stockfish
work. We should not ship its open-ended coach chat or make client-side engine
results authoritative.

Recommendation: retain Go for the authoritative API and game processes, add a
React web client and an Expo/React Native mobile client, and introduce shared
wire-protocol and constrained-coach packages. Phoenix is a sound alternative,
but replacing the tested Go domain and handlers now would add migration risk
without improving the first product milestone.

## Repository and branch audit

### Git state

- Remote: `origin` -> `https://github.com/dustywusty/tinychess.git`
- Default branch: `main` (`origin/HEAD` points to `origin/main`)
- Audited `main`: `4f7c294`, committed 2026-01-11
- Local `dustywusty/nouakchott`: identical to `main`; it is a separate worktree,
  not unmerged product work
- `origin/dustywusty/react-coach`: `7a0ba02`, committed 2026-05-02; 12 linear,
  unmerged commits ahead of `main`
- `origin/e2e-artifacts`: generated GIF artifacts only; not application source
- No unreachable commits were found by `git fsck`
- The worktree was clean before rebuild work began

The `react-coach` branch is the only meaningful branch newer than `main`. It was
inspected commit by commit and built in a temporary worktree; it should not be
blindly merged.

### Evolution

1. The project began as a single-process Go/link chess experiment.
2. Server-authoritative move validation, anonymous seat assignment,
   spectators, SSE, and reactions were added incrementally.
3. Postgres/GORM models were introduced, but on `main` the opened database is
   not connected to live game mutations.
4. The current page was converted to inline React loaded from a CDN and gained
   an unrestricted OpenAI/Hashbrown coach endpoint.
5. Browser E2E coverage was added for a complete two-player game at an iPhone
   viewport.
6. The unmerged `react-coach` branch extracted a Vite SPA, normalized `/api`
   routes, wired persistence, added Stockfish WASM and later added a richer
   Hashbrown/Anthropic review chat.

There is no Vue implementation or Vue experiment in any reachable ref. There
is no React Native, Expo, native deep-link, or mobile application work. Existing
"mobile" work is responsive web styling and a mobile-sized browser test only.

## Current architecture

### Backend

- Go `net/http` process
- `internal/game.Hub` holds every live game in memory
- `corentings/chess/v2` validates legal moves and produces FEN/PGN/outcomes
- POST endpoints mutate moves/reactions; SSE broadcasts state and reactions
- Anonymous browser UUIDs claim the first two seats; later clients spectate
- A process-local five-second reaction cooldown exists
- Games idle for 24 hours are removed by a cleanup goroutine

Strengths: extremely small, authoritative move validation, understandable
concurrency, and good fit for the initial traffic model.

Risks: mutable fields and locks are exposed publicly, creation is implicit in
several read/mutation paths, malformed short UCI input can panic, anonymous
session tokens are bearer identifiers with no signing, SSE state is not resumable,
and process memory is the source of truth. Horizontal replicas would diverge.

### Web frontend

`main` serves a 1,820-line HTML file containing React modules imported from
CDNs. It implements a responsive board, turn/status display, captured pieces,
reactions, seat release, themes, recent games, and a free-form coach panel.

The `react-coach` branch's Vite/TypeScript SPA is structurally preferable. It
has separated pages, components, API/SSE adapters, Zustand stores, FEN helpers,
session persistence, and a lazy Stockfish worker. Its React game flow passes the
existing two-browser checkmate test. It is worth migrating into `apps/web`.

The branch still duplicates protocol shapes in frontend code, is web-specific,
and makes Stockfish evaluation available as a browser-side LLM tool. That engine
work is useful for optional local review, but paid coach decisions and cached
artifacts must be produced or verified server-side.

### Persistence and identity

`main` creates Postgres tables through GORM but discards the returned database
handle, so live games and moves are not persisted. The newer branch introduces
a tested `Store` interface with memory and GORM implementations and persists
positions, moves, results, and evaluations. That interface should be adapted.

Identity is intentionally anonymous and stored in browser session storage. This
is correct for the first product experience. The next version should issue an
opaque, signed anonymous session token and persist seat claims so reconnects and
restarts do not change roles. Accounts remain optional.

### AI experiments

`main` forwards the whole submitted chat plus game context to `gpt-4o-mini` and
allows unrestricted chat. It has no engine grounding, deterministic intent
routing, rate/usage limits, caching, or relevant tests.

The newer branch is better isolated and provider-specific code sits behind a Go
interface. It uses Hashbrown structured UI, Anthropic streaming, browser-side
Stockfish, board annotations, and persisted evals for post-game review. It is
still an open chat system, is designed primarily for post-game review, and does
not implement per-move question/hint limits or cost controls.

Current CopilotKit documentation now lists official React Native support for
Expo or bare React Native, a headless provider, native hooks, and optional
prebuilt UI. It uses a Copilot Runtime endpoint and AG-UI. This makes CopilotKit
viable for a later constrained presentation/runtime adapter, but it should not
own coach policy. We should use its headless surface, if adopted, behind our own
intent and usage API rather than expose a general chat component.

Primary references:

- https://docs.copilotkit.ai/react-native
- https://docs.copilotkit.ai/reference/react-native
- https://docs.copilotkit.ai/a2a/backend/copilot-runtime

### Tests and delivery

- `go test ./...` passes on `main`
- The headless full-game Fool's Mate E2E test passes on `main`
- The newer branch's Vite production build, Go tests, and full-game E2E pass
- Tests cover game state, ownership/color assignment, spectators, moves,
  release behavior, middleware, storage (newer branch), coach frame decoding
  (newer branch), and the basic browser game loop
- CI runs Go build/tests; PR CI also runs Chrome E2E and publishes recordings
- There is no mobile test pipeline, protocol contract test, coach-policy test,
  subscription test, load test, or deployed database integration test
- There is no checked-in container, infrastructure-as-code, migration tool,
  native signing config, universal-link config, or app-link config

## Keep, adapt, discard

### Keep

- Server-authoritative legal move validation with `corentings/chess/v2`
- FEN, PGN, UCI history, outcome, checkmate, stalemate, and draw handling from
  the established chess library
- Anonymous create/share/join/spectate product behavior
- Reactions instead of general chat
- Existing Go domain and handler tests plus the browser full-game test
- UUID game identifiers and URL-addressed games

### Adapt

- React SPA from `react-coach` into `apps/web`
- Its API route normalization, storage interface, memory/Postgres stores, and
  position/evaluation records
- Its Stockfish WASM parser and queue as a non-authoritative web fallback and a
  reference for a server-side analysis adapter
- SSE during migration; define transport-neutral events and add WebSocket
  delivery for React Native before retiring SSE
- Anonymous client IDs into signed anonymous sessions with durable seat claims
- CI into separate API, web, shared-package, mobile, and contract jobs

### Throw away or replace

- Inline CDN-loaded React and server HTML templates once SPA parity lands
- The free-form `main` coach endpoint and unrestricted chat UI
- Hashbrown-specific binary framing as the product protocol
- Client-controlled evaluation uploads for paid coach decisions
- Implicit game creation from arbitrary game IDs
- Public lock/state fields and handler-owned chess reconstruction
- GORM auto-migration as the only production migration mechanism

### Useful work only on another branch

From `origin/dustywusty/react-coach`:

- Extracted React/Vite/TypeScript SPA and mobile-responsive components
- `/api` route design and embedded production assets
- Store interface plus memory and Postgres implementations
- Durable game/move/result/evaluation models and API reads
- Stockfish 18 lite WASM worker and UCI output parsing
- Board annotations and structured review UI concepts
- Provider abstraction and streamed-frame tests (concepts only; protocol and
  unrestricted chat should be replaced)

## Recommended target architecture

```text
apps/mobile (Expo)       apps/web (React/Vite)
          \                 /
           packages/protocol
                    |
            apps/api (Go)
                    |
      authoritative game actor + chess library
          /          |             \
   Postgres     realtime fanout   coach service
                                  /          \
                           Stockfish       LLM adapter
```

### Why Go instead of Phoenix now

The current Go server already models games as small concurrent in-memory units
and has passing domain and browser tests. A single owner goroutine per active
game can provide the same serialization property that motivates Phoenix
processes. Go also has straightforward WebSocket, Postgres, Stockfish process,
and Expo-compatible JSON integrations. Phoenix would become attractive if the
team prefers Elixir operationally or needs built-in distributed presence/pubsub
soon; neither is required for the first production milestone.

The API must remain authoritative. It validates moves, owns seats and lifecycle,
persists accepted events, and is the only component allowed to publish canonical
game state. Clients may use `chess.js` for interaction previews, never acceptance.

### Realtime and persistence

Use explicit `POST /v1/games`, `POST /v1/games/{id}/join`, and authenticated move,
reaction, resign, and rematch commands. Deliver versioned game events over a
WebSocket subscription; keep the existing SSE endpoint until the web client is
migrated. Include monotonically increasing event sequence numbers so reconnects
can request a snapshot and resume safely.

Persist games, seats, moves, results, and coach usage synchronously enough that
accepted state can be reconstructed. Reactions and connection presence can be
ephemeral. Persist selected reactions later only if product analytics requires
them. Postgres is sufficient; do not add Redis until multiple API replicas or
measured fanout needs justify it.

### Coach boundary

The coach API accepts a small supported intent, never an arbitrary chat history.
A deterministic router handles off-topic, repeated, too-long, and exhausted
requests without inference. Per-position state tracks two contextual questions
and three progressive hints. Stockfish produces normalized analysis; a policy
layer decides whether a move is interesting; an LLM adapter optionally turns a
small structured payload into cached coaching artifacts. Provider choice stays
behind an interface.

CopilotKit may later provide headless React/React Native transport and rendering,
but the Go coach policy remains the gate in front of any runtime/model. A plain
typed HTTP implementation should land first so the product does not depend on a
general agent runtime.

## Migration sequence

1. Preserve a green baseline and land this audit.
2. Adapt the tested React SPA and normalized API routes from `react-coach`.
3. Establish workspace roots for `apps/web`, `apps/mobile`, and shared protocol;
   keep the Go API runnable while it moves under `apps/api` in a later small step.
4. Add an Expo app that creates/opens `/g/{id}` links and renders the shared game
   snapshot; add native WebSocket transport next.
5. Version the protocol and introduce explicit create/join/snapshot commands.
6. Replace the lock-heavy hub with one command loop per live game and recover
   actors from Postgres.
7. Add durable anonymous sessions and seat claims without requiring accounts.
8. Run Stockfish server-side and cache normalized analysis by position hash and
   engine configuration.
9. Add deterministic coach routing, limits, progressive hints, usage records,
   provider abstraction, and failure fallbacks before any model UI.
10. Add headless CopilotKit only if it materially improves shared coach UI after
    the constrained HTTP coach works on both clients.
11. Add universal/app-link association files, subscriptions, and production
    observability after the core create/share/play loop is stable.

## First implementation slice

The first slice should remain deployable and reviewable:

1. Reuse the tested React SPA/API-route/Stockfish commits without merging the
   free-form Hashbrown coach commits.
2. Rename the product surface to Your Move.
3. Add the pnpm workspace, Expo application shell, and shared protocol package.
4. Add a server-side constrained coach policy package with deterministic intent
   routing and per-position limits, fully unit tested and with no LLM dependency.
5. Update local development and CI commands, then rerun Go, web build, and browser
   gameplay tests.

This slice deliberately does not claim production-ready mobile realtime,
payments, distributed game recovery, or LLM coaching. It establishes the
boundaries needed to implement those features without another rewrite.

### Expo compatibility note (2026-08-13)

The initial mobile scaffold used the SDK 57 template. At the time of the initial
compatibility change, store Expo Go clients embedded SDK 56, so the app was
aligned to the SDK 56 template for physical-device previews.

Update (2026-09-05): the app now uses SDK 57.0.20 with matching native packages.
Expo's SDK 57 Android download page now points to Google Play. Use a matching
SDK 57 Expo Go client, or rebuild a custom development client. The app requires
Node 22.13+. CI uses Node 24.
