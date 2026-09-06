# Your Move mobile

Expo SDK 56 / React Native client for iOS and Android, with an optional web
preview. One native UI codebase uses the existing Go API and shared wire types.

In development the app derives the Go API host from Metro's LAN address, so a
phone opened from the QR code reaches port 8080 on the development machine.
Set `EXPO_PUBLIC_API_URL` to override it. Set `EXPO_PUBLIC_WEB_URL` to the
deployed web origin used in shares.

```sh
corepack pnpm@9.15.0 install
corepack pnpm@9.15.0 --filter @yourmove/mobile start
```

Start the API with `make dev-api` after `make web-build`. For a browser preview,
run `corepack pnpm@9.15.0 --filter @yourmove/mobile web` and open port 8081.
Metro proxies `/api/` to the local Go server. This proxy is development-only;
a deployed web export needs an equivalent reverse proxy.

Routes mirror web links: `/g/[id]`. Set `EXPO_PUBLIC_WEB_URL` to share web links.
Without it, native shares use `yourmove://g/<id>`, which requires an installed
development or release build. The browser preview shares its own origin.
Universal/app links still require replacing `yourmove.example` in `app.json`
and hosting the platform association files. Set the deployed API origin before
building a release; the default API host is for local development.

## Included

- Create, join, share, and reopen recent games without an account.
- Live moves and emoji reactions through the existing SSE endpoint.
- Saved anonymous identity and board color on the same installation.
- Correct board orientation, legal destinations, last move, check, castling,
  en passant, and a choice of all four promotion pieces.
- Spectator mode, move notation, board flip, and reconnect recovery.
- A coach preview marked “Coming soon”; no generated advice or coach API calls.

The server validates every move. Local chess rules provide interaction hints.
Reactions show the last six events received during this visit. They do not
persist or replay after a reconnect. Device identity is a local seat identifier,
not an authenticated account or cross-device recovery token.

## Visual direction

Warm ivory and dark ink keep the board clear. Matcha, lilac, and peach board
themes add color. Original SVG pieces have consistent shapes across platforms.
Controls sit below the board, with player details above and below it.

References: [Deep Green](https://deepgreen.app/) for board focus,
[Chess UI Design](https://dribbble.com/shots/27245572-Chess-UI-Design) for color
variation, and [ChessRookie](https://doyeonyoo.com/chessrookie) for a friendly tone.
These informed the direction; no reference artwork is bundled.

## Verify

```sh
corepack pnpm@9.15.0 --filter @yourmove/mobile typecheck
corepack pnpm@9.15.0 --filter @yourmove/mobile test
corepack pnpm@9.15.0 --filter @yourmove/mobile exec expo export --platform all
```

Unit tests need Node 22.6+ for TypeScript stripping. With Chrome installed,
the API on port 8080, and the mobile web preview on port 8081:

```sh
corepack pnpm@9.15.0 --filter @yourmove/mobile test:e2e
```

Browser tests cover two players, bidirectional emoji, seat recovery, themes,
spectators, checkmate, recent games, and small phone layout. Screenshots are
saved under `test-results/`. Browser tests and native bundle exports do not
replace testing on iOS and Android devices before release.
