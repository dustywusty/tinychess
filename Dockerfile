# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM node:24-bookworm-slim AS web
WORKDIR /src
RUN corepack enable
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY web/package.json web/package.json
COPY apps/mobile/package.json apps/mobile/package.json
COPY packages/protocol/package.json packages/protocol/package.json
COPY packages/chess/package.json packages/chess/package.json
RUN --mount=type=cache,id=pnpm,target=/pnpm/store \
    corepack pnpm@9.15.0 --filter @yourmove/web... install --frozen-lockfile --store-dir=/pnpm/store
COPY web/ web/
COPY packages/protocol/ packages/protocol/
COPY packages/chess/ packages/chess/
RUN corepack pnpm@9.15.0 --filter @yourmove/web build

FROM --platform=$BUILDPLATFORM golang:1.26-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
COPY --from=web /src/web/dist ./web/dist
ARG TARGETOS
ARG TARGETARCH
ARG COMMIT=dev
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -buildvcs=false -ldflags="-s -w -X main.commit=${COMMIT}" -o /out/tinychess .

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
ARG COMMIT=dev
LABEL org.opencontainers.image.title="Tinychess" \
      org.opencontainers.image.source="https://github.com/dustywusty/tinychess" \
      org.opencontainers.image.revision=$COMMIT
WORKDIR /app
COPY --from=build /out/tinychess /app/tinychess
ENV PORT=8080
EXPOSE 8080
USER 65532:65532
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/app/tinychess", "-healthcheck"]
ENTRYPOINT ["/app/tinychess"]
