APP := tinychess
BIN := bin/$(APP)
PKG := .
PNPM := corepack pnpm@9.15.0
PORT ?= 8080
CHROMEDP_HEADLESS ?= 1
CHROMEDP_VIEWPORT_WIDTH ?= 393
CHROMEDP_VIEWPORT_HEIGHT ?= 852
E2E_RECORD ?= 1
E2E_RECORD_FORMAT ?= gif
E2E_RECORD_FPS ?= 6
E2E_MOVE_DELAY_MS ?= 0
E2E_START_DELAY_MS ?= 0
E2E_CAPTURE_DELAY_MS ?= 200
E2E_RECORD_HOLD_MS ?= 2000
E2E_SEND_EMOJI ?= 1

# ldflags embeds a build stamp and commit hash; feel free to remove
LDFLAGS := -s -w -X 'main.build=$$(date -u +%Y%m%d-%H%M%S)' -X 'main.commit=$$(git rev-parse --short HEAD)'

.PHONY: all bootstrap build run dev dev-api dev-web dev-mobile clean lint test race test-e2e typecheck web-install web-build web-dev web-typecheck mobile-typecheck

all: build

bootstrap:
	$(PNPM) install --frozen-lockfile

web-install: bootstrap

web-build:
	$(PNPM) --filter @yourmove/web build

web-typecheck:
	$(PNPM) --filter @yourmove/web typecheck

web-dev:
	$(PNPM) --filter @yourmove/web dev

mobile-typecheck:
	$(PNPM) --filter @yourmove/mobile typecheck

typecheck:
	$(PNPM) typecheck

build: web-build
	@mkdir -p bin
	go build -trimpath -ldflags="$(LDFLAGS)" -o $(BIN) $(PKG)

run: build
	./$(BIN)

race:
	@mkdir -p bin
	go build -race -o $(BIN) $(PKG)
	./$(BIN)

lint:
	@command -v golangci-lint >/dev/null || { echo "Install golangci-lint: https://golangci-lint.run/"; exit 1; }
	golangci-lint run

test: web-build
	go test ./...

test-e2e: web-build
	@if [ "$(E2E_RECORD)" = "1" ] && [ "$(E2E_RECORD_FORMAT)" != "frames" ]; then \
		command -v ffmpeg >/dev/null || { echo "ffmpeg not found (set E2E_RECORD_FORMAT=frames to skip stitching)"; exit 1; }; \
	fi
	@mkdir -p e2e-artifacts
	@rm -f e2e-artifacts/*.png e2e-artifacts/*.gif e2e-artifacts/*.mp4
	CHROMEDP_HEADLESS=$(CHROMEDP_HEADLESS) \
	CHROMEDP_VIEWPORT_WIDTH=$(CHROMEDP_VIEWPORT_WIDTH) \
	CHROMEDP_VIEWPORT_HEIGHT=$(CHROMEDP_VIEWPORT_HEIGHT) \
	E2E_RECORD=$(E2E_RECORD) \
	E2E_RECORD_FORMAT=$(E2E_RECORD_FORMAT) \
	E2E_RECORD_FPS=$(E2E_RECORD_FPS) \
	E2E_RECORD_HOLD_MS=$(E2E_RECORD_HOLD_MS) \
	E2E_MOVE_DELAY_MS=$(E2E_MOVE_DELAY_MS) \
	E2E_START_DELAY_MS=$(E2E_START_DELAY_MS) \
	E2E_CAPTURE_DELAY_MS=$(E2E_CAPTURE_DELAY_MS) \
	E2E_SEND_EMOJI=$(E2E_SEND_EMOJI) \
	go test -tags e2e ./internal/e2e -run TestPlay -v
	@echo "e2e artifacts: e2e-artifacts/"

dev-api:
	go run .

dev-web:
	$(PNPM) --filter @yourmove/web dev

dev-mobile:
	$(PNPM) --filter @yourmove/mobile start

dev:
	@set -eu; \
		$(MAKE) dev-api & api_pid=$$!; \
		$(MAKE) dev-web & web_pid=$$!; \
		cleanup() { kill $$api_pid $$web_pid 2>/dev/null || true; }; \
		trap cleanup INT TERM EXIT; \
		$(MAKE) dev-mobile
