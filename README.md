# What is this?

I've been playing around w/ making something w just agents, purely out of curiousity, this is that!

# Tiny Chess

[![CI](https://github.com/dustywusty/tinychess/actions/workflows/ci.yml/badge.svg)](https://github.com/dustywusty/tinychess/actions/workflows/ci.yml)

A simple web-based chess game that you can play with just a link.

We give you a shareable URL and we'll randomly pick your side; the next person gets the other.

Anyone else who opens the link is a spectator.

Players can react with any emoji using the built-in emoji picker.

## Local development

### Bootstrap

- Go 1.22+
- Make
- Optional: `air` for live reload (`go install github.com/cosmtrek/air@latest`)
- Optional: Postgres if you want durable persistence (the app runs in-memory by default)

### Install + run

```sh
go mod download
make run
```

Open http://localhost:8080.

### Live reload (optional)

```sh
make dev
```

### Database (optional)

By default games and evals live in process memory and are lost on restart —
fine for local play and Review. Set `DATABASE_URL` for durable Postgres
persistence:

```sh
export DATABASE_URL="postgres://user:pass@localhost:5432/tinychess?sslmode=disable"
```

## Tests

### Unit tests

```sh
go test ./...
```

### Browser e2e test (full game)

Runs a headless browser test that opens two clients and plays a complete legal
game (Fool's Mate) to verify a game can finish.

```sh
go test -tags e2e ./internal/e2e -run TestPlayFullGame
```

Notes:
- Requires a local Chrome/Chromium install.
- If Chrome isn't on your PATH, set `CHROME_BIN` to the browser executable.
- If your environment needs it, set `CHROMEDP_NO_SANDBOX=1`.
- To watch the game being played, run headed with `CHROMEDP_HEADLESS=0`.
- To slow down moves, set `E2E_MOVE_DELAY_MS` (milliseconds between moves).
- To pause before the first move, set `E2E_START_DELAY_MS`.
- To wait for UI updates before capturing, set `E2E_CAPTURE_DELAY_MS`.
- Famous quick mates:
  - Fool's Mate: `go test -tags e2e ./internal/e2e -run TestPlayFullGame`
  - Scholar's Mate: `go test -tags e2e ./internal/e2e -run TestPlayScholarsMate`
- Optional recording:
  - Set `E2E_RECORD=1` to capture screenshots before each move (white client).
  - Output defaults to `e2e-artifacts/` (override with `E2E_RECORD_DIR`).
  - Use `E2E_RECORD_FORMAT=mp4` (default), `gif`, or `frames` (PNGs only).
  - Set `E2E_RECORD_FPS` to control playback (default `6`).
  - Set `E2E_RECORD_HOLD_MS` to hold the final frame (default `2000`).
  - Requires `ffmpeg` on PATH for mp4/gif.

### Make target (run all e2e tests + stitch gifs)

```sh
E2E_RECORD=1 E2E_RECORD_FORMAT=gif E2E_RECORD_FPS=6 make test-e2e
```

Defaults can be overridden:
- `CHROMEDP_HEADLESS=0` to watch the run.
- `CHROMEDP_VIEWPORT_WIDTH` and `CHROMEDP_VIEWPORT_HEIGHT` for mobile/aspect testing (defaults to iPhone 15 Pro: 393x852).
- `E2E_MOVE_DELAY_MS` and `E2E_START_DELAY_MS` for timing.
- `E2E_CAPTURE_DELAY_MS` to wait for SSE/UI updates before each capture.
- `E2E_RECORD_HOLD_MS` to extend the final frame.
- `E2E_RECORD_FORMAT=frames` to skip stitching (no ffmpeg needed).
- `E2E_SEND_EMOJI=0` to skip the emoji taunt (default on).
- `E2E_EMOJI` to override the emoji character.

## Links

- Production: https://tinychess.dusty.wtf/
- Sandbox: https://sandbox.tinychess.dusty.wtf
