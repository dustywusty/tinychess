# Development

## Requirements

- Go 1.24+
- Node 22.13+ (Node 24 recommended; used by CI)
- Corepack (the workspace pins pnpm 9.15.0)
- Chrome/Chromium for browser E2E
- Optional: Postgres
- Optional: Xcode/Android Studio or a physical device for native builds

The mobile app targets Expo SDK 57 and React Native 0.86. Use a matching SDK 57
Expo Go client for device previews. Android installation is available through
[Expo's download page](https://expo.dev/go?device=true&platform=android&sdkVersion=57).
An SDK 56 Expo Go client cannot open this project. Rebuild any custom development
client after upgrading the SDK.

## Commands

```sh
make bootstrap       # install JS workspace dependencies
make dev             # API + Vite + Expo
make dev-api         # Go API only on :8080
make dev-web         # Vite on :5173, proxying /api to :8080
make dev-mobile      # Expo development server
make typecheck       # protocol, web, and mobile
make test            # web production build and all Go tests
E2E_RECORD=0 make test-e2e
```

The production Go binary embeds `web/dist`, so `make build` builds web first.

## Mobile environment

Expo uses these public development variables:

- `EXPO_PUBLIC_API_URL`: API origin reachable by the simulator/device
- `EXPO_PUBLIC_WEB_URL`: origin used when sharing `https://.../g/{id}`

The default is derived from Metro's LAN host, with localhost/`10.0.2.2`
fallbacks for iOS/Android simulators. The phone and development machine must be
on a network that permits access to ports 8080 and 8081. Use development builds,
not Expo Go, when testing the `yourmove://` scheme or universal links.

## Database

Set `DATABASE_URL` for Postgres. Current startup uses GORM auto-migration as an
intermediate migration path. Do not rely on it for irreversible production
schema changes; checked migrations are planned before deployment.

Postgres transactions save moves and seats before publication. Recovery replays the full legal move history.
The database stores private seat identities in `games.live_state`, not public review metadata.
Without `DATABASE_URL`, the API uses temporary in-memory games.

### Backend recovery tests

Run the unit tests and race detector:

```sh
go test -race ./...
```

The PostgreSQL integration tests require a separate test database.
They never use `DATABASE_URL`. Each test creates and removes its own schema.

CAUTION: Use a disposable database for `TEST_DATABASE_URL`, never a production database.

Run the integration tests against that database:

```sh
TEST_DATABASE_URL='postgres://test_user:test_password@localhost:5432/tinychess_test?sslmode=disable' go test -race ./...
```

Without this variable, the integration tests skip. GitHub CI supplies PostgreSQL and runs them on each change.
Coverage includes restart recovery, seats, captures, special moves, repetition, concurrent moves, commit failure, unavailable storage, and legacy rows.
The commit-failure test checks that no board update or SSE event escapes a failed transaction.

## Tests to add next

Mobile unit tests now cover invitation parsing, board coordinates, special
moves, seat metadata, and stream retries. Mobile browser tests cover a complete
two-player game, emoji in both directions, seat recovery, spectators, and small
phone layout. See [mobile verification](../apps/mobile/README.md#verify).

Remaining coverage:

1. Native device lifecycle and interaction tests on iOS and Android
2. Shared protocol fixtures decoded by both TypeScript and Go
3. Native device tests for server restart and seat recovery
4. WebSocket two-player/replay tests
5. Stockfish parser/classification fixtures
6. Coach provider failure and artifact-cache tests
