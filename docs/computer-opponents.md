# Computer opponents

## Architecture

Both clients use the same Arasan WASM build and shared TypeScript policy.
The web client runs the engine in a Web Worker.
The Android and iOS clients run it in a separate, hidden WebView with a bundled HTML asset.
The engine does not receive player credentials or call the backend.

The WebView approach avoids a custom native wrapper around Arasan's process-level input, output, and exit functions.
It also avoids separate engine behavior across platforms.
An engine transport isolates this choice from the bot policy.
A native transport can replace it after device measurements justify the extra integration work.

The controller passes the complete UCI history to the engine.
The adapter checks that this history matches the current FEN.
It compares candidates from one search depth and handles mate scores separately.
Pip and Max also analyze a small random sample of legal moves beside the best candidates.
This sample can contain mistakes that a top-N search excludes.

The policy selects a legal move within the opponent's loss limit.
Capture, check, and castling preferences only affect moves within the selected quality class.
The controller applies a short presentation delay without extra engine work.
The strongest opponent selects the best evaluated move.

The backend reserves a bot seat when it creates the game.
Only the human owner can submit a move for that seat.
Each bot submission includes `expectedPly`. The backend checks this value inside the move lock and database transaction.
Multiple tabs can calculate a move, but only one submission can commit for that position.

## Reused components

The implementation reuses Go's `Game` move validation, PGN history, persistence transactions, recovery snapshots, and SSE publication.
The shared client package reuses `chess.js` for legal candidates and move features.
Both clients reuse their existing board, capture display, anonymous identity, and move endpoint.
No new chess rules engine or backend analysis service is necessary.

Bot metadata contains the opponent ID, color, and policy version.
PostgreSQL stores this metadata with the existing recovery state. Existing friend games retain their behavior.
The current single-instance hosting requirement remains unchanged.

## Build the engine

Requirements: Git, Node 24, `patch`, and Emscripten 4.0.14 on the command path.

1. Activate Emscripten 4.0.14 in your shell.
2. Run `pnpm engine:build` from the repository root.
3. Run `pnpm engine:test`.
4. Run `pnpm build`.

The build downloads a pinned Arasan commit and checks that the source has no tracked changes.
It generates the web worker runtime, embedded network, version metadata, license notices, and offline mobile HTML.
Generated engine assets do not belong in Git. The web build fails if these assets are absent.

`ARASAN_SOURCE` selects an existing source checkout.
`EMXX` selects the `em++` executable.
`EMSCRIPTEN_ROOT` selects the Emscripten directory for license files when the compiler uses a wrapper.

Docker and CI compile these assets with the pinned Emscripten image.
This toolchain image only supports Linux AMD64. ARM hosts need Docker's AMD64 emulation for this build stage.
The generated WASM remains platform-independent. The final server image still supports AMD64 and ARM64.
The production server only serves static engine files. It never runs an engine process.

## Build a mobile test copy

1. Run `pnpm engine:build` before the EAS upload.
2. Run `pnpm engine:check`.
3. Follow the APK procedure in [deployment.md](deployment.md).

The EAS ignore file retains generated engine assets.
The post-install check stops a build with missing assets.
A native dependency change requires a new APK or development build, not only a JavaScript update.

The local Android build uses JDK 17. Android Studio's bundled JDK 25 failed in the native CMake tools during verification.
The ARM64 release variant compiles successfully with JDK 17 and includes the offline engine asset.
The local APK uses debug signing for tests. The EAS procedure remains the path for managed signing and shared builds.
If an installed copy uses a different signing key, Android rejects the update.
Do not uninstall that copy without a recovery plan for its stored player identity.

Engine execution needs no network after installation.
Game creation, move validation, and recovery still require the API connection.
This release does not add offline game synchronization.

## Tuning and tests

`packages/chess/src/bots/definitions.ts` contains all five opponents and their timing, loss limits, and preferences.
`BOT_POLICY_VERSION` identifies the policy in saved games.
The displayed descriptions are not measured Elo ratings.

Run `pnpm bot:lab pip` for an analysis of the starting position.
An optional quoted FEN and numeric seed follow the bot ID.
The lab shows candidate scores, loss, move features, and the selected quality class.
The seed makes selection reproducible for a fixed candidate set. Time-limited engine analysis can still vary between runs.
This command is development-only and does not run on the backend.

`pnpm test` covers parsing, deterministic selection, mate handling, cancellation, and existing client utilities.
`go test -race ./...` covers reserved seats, authorization, stale moves, concurrent submissions, history, and checkmate.
PostgreSQL tests require an explicit `TEST_DATABASE_URL` and use isolated schemas.
`pnpm engine:test` exercises the actual WASM engine, including cancellation and restart.
The browser test in `web/tests/browser/bots.spec.ts` exercises the picker, real bot replies, reload, and spectator restrictions.

## Release checks and limits

The engine uses one search thread, a 16 MB hash table, and bounded searches.
The WASM file is approximately 23 MB before compression. The offline mobile HTML is approximately 30 MB.
These sizes can change with the toolchain or network.

Device tests must cover Android System WebView and iOS WKWebView, including background, resume, and low-memory termination.
Desktop timing does not establish phone battery use or response time.
The runtime requires WebAssembly SIMD support.

Cancellation destroys an active transport when necessary, which guarantees that old output cannot enter a later search.
The app reuses the engine between ordinary moves.
The app retries engine errors once, then shows a retry control.

All opponents start unlocked. Reactions, progression, exact rating calibration, and new analytics services are outside this release.
Further tuning needs human play sessions, especially for Pip's mistakes and Ada's defensive style.

The full engine licensing details are in [THIRD_PARTY_NOTICES.md](../THIRD_PARTY_NOTICES.md).
