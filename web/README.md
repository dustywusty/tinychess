# Your Move web

The React and TypeScript website matches the Expo mobile client: warm ivory,
dark ink, three board colors, and original vector pieces. Desktop uses two
columns; phones use a single column. The saved dark theme remains available.

## Run locally

From the repository root:

```sh
corepack pnpm@9.15.0 install
make web-build
make dev-api
```

In a second terminal:

```sh
make dev-web
```

Open http://localhost:5173. Vite proxies API requests to port 8080.

## Verify

```sh
corepack pnpm@9.15.0 --filter @yourmove/web build
corepack pnpm@9.15.0 --filter @yourmove/web test:e2e
```

Browser tests require Chrome, the API, and the Vite server. They cover phone
and desktop layouts, saved themes, invitation validation, two-player moves,
Black's board orientation, emoji in both directions, seat recovery,
spectators, checkmate, and recent games. Screenshots go to `test-results/`.

The existing full-game suite also runs with `E2E_RECORD=0 make test-e2e`.

## Behavior

Create a game, share its link, and play. The server assigns seats and validates
moves. The board supports tap or click, drag and drop, legal-move hints, and
all four promotion choices. SSE supplies live positions and emoji reactions.

The coach card is marked “Coming soon.” It does not call a coach API.

`make build` embeds the website bundle in the Go binary. Restart the binary
after rebuilding to serve the updated website outside Vite.
