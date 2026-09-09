# Your Move

Create a chess game, share a link, and play. The first two anonymous visitors
take the seats; everyone else can watch and react. Chess is free. A constrained,
engine-grounded coaching add-on is being built behind a strict cost-control
policy.

## Repository

```text
apps/mobile/       Expo / React Native client
web/               React / Vite web client (moves to apps/web after migration)
packages/protocol/ Shared TypeScript wire types
packages/coach/    Server-side deterministic coach policy
internal/          Current Go API, game domain, storage, and handlers
docs/              Architecture, protocol, coach, and rebuild decisions
```

The authoritative server validates every move with
`github.com/corentings/chess/v2`. Web currently receives realtime events over
SSE. Mobile uses the same commands and SSE stream, including emoji reactions,
with reconnect and app foreground recovery.

Computer games offer Pip, Max, Ada, Viktor, and The Machine.
Arasan runs locally on the client. The backend still validates and saves every move.
See [computer opponents](docs/computer-opponents.md) for the engine build, architecture, tests, and licensing notes.

## Start locally

Requirements: Go 1.24+, Node 22.13+ (Node 24 recommended), Corepack, Make, and optionally Postgres.

```sh
make bootstrap
pnpm engine:build
make dev
```

`make dev` starts the Go API, Vite web client, and Expo development server.
The engine build requires Emscripten 4.0.14. Docker builds include this toolchain automatically.
Open `http://localhost:5173` for web. The API listens on `:8080`.

For a device, set `EXPO_PUBLIC_API_URL` to a host it can reach. Android emulators
default to `http://10.0.2.2:8080`; iOS simulators default to localhost.

Set `DATABASE_URL` to enable Postgres persistence:

```sh
export DATABASE_URL="postgres://user:pass@localhost:5432/yourmove?sslmode=disable"
```

Postgres saves games and anonymous player seats before the server acknowledges changes.
After a restart, web and mobile clients recover the full game with their existing client identity.
Without Postgres, games remain temporary and disappear on restart.
Older games without saved seat data retain review history but cannot resume safely.

## Verify

```sh
make typecheck
make test
E2E_RECORD=0 make test-e2e
```

The browser E2E suite opens two clients and plays complete legal games. Chrome
or Chromium is required; set `CHROME_BIN` if it is not discoverable.

See [development](docs/development.md), [architecture](docs/architecture.md),
[protocol](docs/protocol.md), [coach](docs/coach.md), and the full
[rebuild audit](docs/rebuild-plan.md).

## Deploy and share an APK

See [deployment and Android builds](docs/deployment.md) for Docker releases, DigitalOcean App Platform, and installable APKs.
The [.do/app.yaml](.do/app.yaml) spec defines the App Platform service.
The production domain is `yourmove.fun`. DNS and HTTPS activation require configuration in DigitalOcean and your DNS provider.
Run one server instance. Postgres provides game recovery, but live event delivery remains local to one server.
