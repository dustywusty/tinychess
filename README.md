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

- Go 1.24+
- Make
- Optional: `air` for live reload (`go install github.com/cosmtrek/air@latest`)
- Optional: Postgres for database schema initialization (game persistence is not implemented)

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

`DATABASE_URL` enables the Postgres connection and schema migrations. The app does not yet save or restore games through this connection.

For example:

```sh
export DATABASE_URL="postgres://user:pass@localhost:5432/tinychess?sslmode=disable"
```

## DigitalOcean App Platform

The app has two components on one public domain:

| Public path | Component | Purpose |
| --- | --- | --- |
| `/`, `/new/`, `/<game-id>` | Static frontend | Serves the files in `frontend/` without a build step. |
| `/api/*` | Go backend | Handles moves, player assignments, reactions, the coach, and live updates. |

The browser creates game IDs and reads the current game ID from the URL.
The static host serves `game.html` for shared game links through its catch-all configuration.
The frontend reads the public version and coach toggle from `/api/config`. API keys stay on the backend.
Some frontend libraries load from public CDNs.

The [App Platform spec](.do/app.yaml) defines both components, API routing, and a backend health check.
It targets the existing `hammerhead-app` in the `nyc` region.
The primary domain is `yourmove.fun`. The previous domain, `pawnd.dusty.wtf`, remains an alias.
The app ID is `e80ec2bc-7a78-4b44-b856-d2c26a1a1ca5`.
Both components use the `main` branch of `dustywusty/tinychess`.
The spec keeps the existing `tinychess` service name and its single 512 MB instance.
The backend image contains the Go binary and CA certificates. It contains no frontend files.
DigitalOcean builds this image from the Dockerfile, so a separate container registry is not required.

### Automatic deployment

The [CI workflow](.github/workflows/ci.yml) deploys after a push to `main`, including a pull request merge.
Deployment requires both the Go checks and Docker health check to pass.
Pull requests and other branches run the checks without deployment.
You can also run the workflow manually from the GitHub Actions page on `main`.

The workflow validates `.do/app.yaml`, applies it to the existing app, updates the source, and waits for deployment.
DigitalOcean builds the backend image and publishes the static frontend.
The workflow then checks `/api/config` through the app's starter domain.
Deployments run one at a time. A queued run skips deployment if `main` already has a newer commit.

GitHub Actions configuration:

| Kind | Name | Value |
| --- | --- | --- |
| Repository secret | `DIGITALOCEAN_ACCESS_TOKEN` | A DigitalOcean token with permission to read and update the app. |
| Repository variable | `DIGITALOCEAN_APP_ID` | `e80ec2bc-7a78-4b44-b856-d2c26a1a1ca5` |

DigitalOcean's direct GitHub deployment trigger is disabled. GitHub Actions controls deployments after the required checks.
See [DigitalOcean's GitHub Actions guide](https://docs.digitalocean.com/products/app-platform/how-to/deploy-from-github-actions/) for authentication details.

### Manual deployment

Push the code to the branch named in both components of `.do/app.yaml`.
Authenticate `doctl` with the DigitalOcean account that owns the app:

```sh
doctl auth init
```

Validate the proposed update without deployment:

```sh
doctl apps propose --app e80ec2bc-7a78-4b44-b856-d2c26a1a1ca5 --spec .do/app.yaml
```

Apply the spec to the existing app:

```sh
doctl apps update e80ec2bc-7a78-4b44-b856-d2c26a1a1ca5 --spec .do/app.yaml --update-sources --wait
```

This command starts a deployment and replaces the running app. Active games will be lost.

If you configure the components manually, preserve `/api` in the backend ingress rule.
Set the frontend index document to `index.html` and the catch-all document to `game.html`.
Keep both components on the same domain. This permits API requests without cross-origin configuration.

See the [DigitalOcean app spec reference](https://docs.digitalocean.com/products/app-platform/reference/app-spec/) for component and ingress fields.

### Backend configuration

| Variable | Default | Purpose |
| --- | --- | --- |
| `PORT` | `8080` | HTTP port inside the container. |
| `OPENAI_API_KEY` | Unset | Enables the optional chess coach. |
| `DATABASE_URL` | Unset | Connects to Postgres and creates tables. Does not persist gameplay. |

If you enable the coach, add `OPENAI_API_KEY` as an encrypted runtime variable on the backend component.
Keep secrets out of the frontend, Git, and Docker build arguments.
The backend health endpoint is `/healthz`. App Platform calls this endpoint directly on port 8080.

Keep the backend at one instance. All game state and event subscribers live in process memory.
A backend restart or deployment loses active games. Multiple replicas require shared game state and event delivery.
Session affinity per browser alone does not keep two players on the same instance.

Game updates use Server-Sent Events at `/api/sse/<game-id>`, with a heartbeat every 15 seconds.
If you add another proxy, disable buffering and caching for `/api/sse/` and `/api/coach`.
Set that proxy's stream idle timeout to at least 60 seconds.
The app sends `X-Accel-Buffering: no` for both streaming endpoints.
[NGINX supports this header](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffering).

After deployment, open the app in two browsers. Create a game, share its URL, and play a move in each browser.
Check that both boards update and that a refresh of the game URL still opens the board.

### Build the backend image locally

Build the image from the repository root:

```sh
docker build --build-arg COMMIT="$(git rev-parse --short HEAD)" -t tinychess-backend:local .
```

Run the backend:

```sh
docker run -d --name tinychess-backend --restart unless-stopped \
  --read-only --cap-drop ALL --security-opt no-new-privileges \
  -p 127.0.0.1:8080:8080 tinychess-backend:local
```

Check the backend:

```sh
curl --fail http://localhost:8080/healthz
curl --fail http://localhost:8080/api/config
```

The backend returns `404` for frontend pages. For local development with the frontend, run `make run` instead.
This command serves `frontend/` through Go with the same API paths and game-page fallback as the production routing.

App Platform requires [Linux AMD64 images](https://docs.digitalocean.com/products/app-platform/details/limits/#image-limits).
For an image that you will upload to App Platform, build for this architecture:

```sh
docker buildx build --platform linux/amd64 --load \
  --build-arg COMMIT="$(git rev-parse --short HEAD)" -t tinychess-backend:amd64 .
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
