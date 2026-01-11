APP := tinychess
BIN := bin/$(APP)
PKG := .
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

.PHONY: all build run dev clean lint test race test-e2e

all: build

build:
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

test:
	go test ./...

test-e2e:
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

dev:
	@command -v air >/dev/null || { echo "air not found"; exit 1; }
	air -c .air.toml
