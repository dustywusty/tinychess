APP := tinychess
BIN := bin/$(APP)
PKG := .
PORT ?= 8080

# ldflags embeds a build stamp and commit hash; feel free to remove
LDFLAGS := -s -w -X 'main.build=$$(date -u +%Y%m%d-%H%M%S)' -X 'main.commit=$$(git rev-parse --short HEAD)'

.PHONY: all build run dev clean lint test race

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

dev:
	@command -v air >/dev/null || { echo "air not found"; exit 1; }
	air -c .air.toml