# Your Move

Free chess with friends or a computer opponent. Create a game, share its link, and play without an account.

**[Play at yourmove.fun](https://yourmove.fun)** · [Development guide](docs/development.md) · [Mobile app](apps/mobile/README.md)

![Your Move home with friend and computer games, recent games, and a completed game result](docs/screenshots/home.png)

## Play your way

Your Move includes a React website and an Expo app for iOS and Android. Both clients use the same Go API.
Available features include:

- **Play a friend.** The first two visitors take the seats. Everyone else can watch and send emoji reactions.
- **Play the computer.** Choose Pip, Max, Ada, Viktor, or The Machine, then choose White or Black. All five opponents start unlocked.
- **Make your move.** Tap or click to move. The website also supports drag and drop. Both boards show legal destinations, the last move, check, and captured pieces.
- **Keep the game close.** Share links, flip the board, read move notation, and reopen recent games with win, loss, or draw results.
- **Make it yours.** Choose light or dark mode and six board colors: Matcha, Lilac, Peach, Teal, Sky, and Rose.
- **Return to your seat.** The same browser or app installation retains your anonymous identity. With Postgres, games and seats survive server restarts.

Arasan runs on your device through WebAssembly. The Go server validates and saves moves, so computer games still require an API connection.
The [computer opponents guide](docs/computer-opponents.md) explains the engine, opponent behavior, and build requirements.

The chess coach is a **Coming soon** preview on the home screen. It does not provide advice or call a coach API.

## At the board

A live friend game, with legal destinations, captures, reactions, and move notation:

![Desktop friend game with a selected knight, legal moves, captured pawns, and emoji reactions](docs/screenshots/friend-game.png)

The website also adapts to phone screens:

<table>
  <tr>
    <th>Choose an opponent</th>
    <th>Play Ada</th>
    <th>Watch in dark mode</th>
  </tr>
  <tr>
    <td valign="top"><img src="docs/screenshots/computer-picker.png" alt="Phone layout with all five computer opponents and a choice of White or Black" width="280"></td>
    <td valign="top"><img src="docs/screenshots/computer-game.png" alt="Phone layout of a live game against Ada with the Sky board color" width="280"></td>
    <td valign="top"><img src="docs/screenshots/spectator-dark.png" alt="Phone spectator view with dark mode, the Peach board color, captures, and move notation" width="280"></td>
  </tr>
</table>

<details>
<summary>A completed game</summary>

![A completed game with the Lilac board color, checkmate status, and the full move notation](docs/screenshots/checkmate.png)

</details>

These screenshots show the local website with real games. Phone screenshots show the responsive website, not the native app.
The [screenshot guide](docs/screenshots/README.md) includes the capture script and refresh instructions.

## Run locally

Requirements: Go 1.24+, Node 22.13+ (Node 24 recommended), Corepack, Make, Git, `patch`, and Emscripten 4.0.14.
Postgres is optional. The workspace pins pnpm 9.15.0.

1. Install the workspace dependencies:

   ```sh
   make bootstrap
   ```

2. With Emscripten 4.0.14 active in your shell, build the engine assets:

   ```sh
   corepack pnpm@9.15.0 engine:build
   ```

3. Start the API:

   ```sh
   make dev-api
   ```

4. In another terminal, start the website:

   ```sh
   make dev-web
   ```

5. Open [localhost:5173](http://localhost:5173).

Vite forwards API requests to port 8080. The engine build downloads pinned Arasan sources and generates the web and mobile assets.
Production web builds require these assets. Docker and CI include the Emscripten toolchain.

For web and mobile development together, run `make dev` after the dependency and engine steps.
This command starts the API, Vite, and Expo.
The [mobile guide](apps/mobile/README.md) covers Expo Go, device previews, and installable Android builds.

For a phone that cannot reach the default API host, set `EXPO_PUBLIC_API_URL` to your development machine's LAN address.
`EXPO_PUBLIC_WEB_URL` controls the origin of shared web links.

### Save games across restarts

Before you start the API, set `DATABASE_URL` to a Postgres database:

```sh
export DATABASE_URL="postgres://user:pass@localhost:5432/yourmove?sslmode=disable"
```

Postgres saves games and seats before the server acknowledges changes. Without Postgres, games disappear after a server restart.
Anonymous seat recovery requires the original browser data or app installation. A shared game link does not transfer your seat.

## Verify changes

After the dependency and engine steps, run the checks from the repository root:

| Command | Coverage |
| --- | --- |
| `make typecheck` | Shared packages, web, and mobile TypeScript |
| `corepack pnpm@9.15.0 test` | Shared chess logic and client utilities |
| `make test` | Production web build and Go tests |
| `corepack pnpm@9.15.0 engine:test` | Actual Arasan WASM runtime |
| `E2E_RECORD=0 make test-e2e` | Complete legal games in two browser clients |

With the API and Vite servers active, run the web browser suite:

```sh
corepack pnpm@9.15.0 --filter @yourmove/web test:e2e
```

The Playwright suite requires Google Chrome. It covers computer opponents, friend games, spectators, themes, captures, reactions, recent games, and recovery.
The full-game Go suite accepts Chrome or Chromium through `CHROME_BIN`.
The [development guide](docs/development.md) covers Postgres integration tests and the race detector.

## Repository

| Path | Purpose |
| --- | --- |
| [web/](web/) | React, TypeScript, and Vite website |
| [apps/mobile/](apps/mobile/) | Expo and React Native app |
| [packages/chess/](packages/chess/) | Shared chess helpers and computer opponent policy |
| [packages/protocol/](packages/protocol/) | Shared TypeScript wire types |
| [packages/coach/](packages/coach/) | Deterministic policy for the planned coach |
| [engine/arasan/](engine/arasan/) | Pinned engine build, runtime checks, and tuning tools |
| [internal/](internal/) | Go game logic, API handlers, and storage |
| [docs/](docs/) | Development, deployment, protocol, and design documents |

The server validates moves with `github.com/corentings/chess/v2`. Both clients receive live game state and emoji reactions through Server-Sent Events (SSE).
The web engine uses a Web Worker. The native engine uses a bundled WebView.

## Deployment

The GitHub Actions workflow deploys successful `main` builds to DigitalOcean App Platform.
The [app spec](.do/app.yaml) serves the website as static files and the Go API as a separate Docker service.
Live event delivery requires one API instance, including deployments with Postgres.

The [deployment guide](docs/deployment.md) covers DigitalOcean, Docker images, database configuration, and Android APKs.
Further design details are in [architecture](docs/architecture.md), [protocol](docs/protocol.md), and the [coach plan](docs/coach.md).

## License

Your Move uses the [MIT license](LICENSE). Arasan and bundled engine dependencies have separate terms in [third-party notices](THIRD_PARTY_NOTICES.md).
