# README screenshots

These PNG files show the website from this repository in Chrome.
Desktop captures use a 1280 × 940 viewport. Phone captures use a 393 × 852 viewport.
Captures use a device scale factor of 1.
The opponent picker capture includes only the viewport. The other captures include the full page.
Phone captures show the responsive website, not an iOS or Android build.

| File | State |
| --- | --- |
| `home.png` | Recent friend and computer games, with a completed game result |
| `friend-game.png` | Live game with captures, legal destinations, reactions, and move notation |
| `computer-picker.png` | All five opponents, with Ada selected |
| `computer-game.png` | Actual Arasan replies from Ada, with the Sky board color |
| `spectator-dark.png` | A third visitor watches the friend game in dark mode with the Peach board color |
| `checkmate.png` | Scholar's Mate with the Lilac board color |

## Refresh the screenshots

Requirements: workspace dependencies, generated Arasan assets, and Google Chrome.
The [root README](../../README.md#run-locally) describes the dependency and engine steps.

1. Start a local API with `make dev-api`.
2. In another terminal, start Vite with `make dev-web`.
3. From the repository root, run the capture script:

   ```sh
   corepack pnpm@9.15.0 --filter @yourmove/web exec node scripts/capture-readme.mjs
   ```

4. Review the PNG files before you commit them.

For a different local Vite port, set `PW_BASE_URL`:

```sh
PW_BASE_URL=http://localhost:5182 corepack pnpm@9.15.0 --filter @yourmove/web exec node scripts/capture-readme.mjs
```

The [capture script](../../web/scripts/capture-readme.mjs) creates fresh browser identities and three games through the app UI.
It plays legal moves, sends reactions, waits for engine replies, and checks each state before capture.
It does not change application code or supply mock game state.
Game IDs, seat colors, and engine replies can differ between runs.
The script replaces the six PNG files in this directory.
