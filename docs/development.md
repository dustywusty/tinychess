# Development

## Requirements

- Go 1.24+
- Node 20+
- Corepack (the workspace pins pnpm 9.15.0)
- Chrome/Chromium for browser E2E
- Optional: Postgres
- Optional: Xcode/Android Studio or a physical device for native builds

The mobile app targets Expo SDK 56 because it matches the Expo Go runtime
currently distributed through the app stores. SDK 57 requires a matching
development build during Expo's transition; do not upgrade the project version
alone and expect the store Expo Go client to open it.

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

## Tests to add next

1. Mobile create/open/move component tests with mocked transport
2. Shared protocol fixtures decoded by both TypeScript and Go
3. Snapshot role and reconnect handler tests
4. WebSocket two-player/replay tests
5. Stockfish parser/classification fixtures
6. Coach provider failure and artifact-cache tests
