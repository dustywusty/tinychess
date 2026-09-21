# syntax=docker/dockerfile:1
FROM node:22-bookworm AS coach-engine
RUN apt-get update && apt-get install -y --no-install-recommends clang g++ git patch && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY engine/arasan/native.mjs engine/arasan/no-tablebases.patch ./engine/arasan/
RUN git clone --no-checkout --filter=blob:none https://github.com/jdart1/arasan-chess.git engine/arasan/vendor && \
    git -C engine/arasan/vendor checkout --detach 361840f407b588252a48b25bb3000c751a470c89
RUN CXX=clang++ node engine/arasan/native.mjs && \
    cp /usr/share/doc/libstdc++6/copyright engine/arasan/generated/native/LICENSE.libstdc++ && \
    cp /usr/share/doc/libc6/copyright engine/arasan/generated/native/LICENSE.glibc

FROM --platform=$BUILDPLATFORM golang:1.26-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
ARG TARGETOS
ARG TARGETARCH
ARG COMMIT=dev
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -buildvcs=false -ldflags="-s -w -X main.commit=${COMMIT}" -o /out/tinychess .

FROM gcr.io/distroless/cc-debian12:nonroot AS runtime
ARG COMMIT=dev
LABEL org.opencontainers.image.title="Tinychess" \
      org.opencontainers.image.source="https://github.com/dustywusty/tinychess" \
      org.opencontainers.image.revision=$COMMIT
WORKDIR /app
COPY --from=build /out/tinychess /app/tinychess
COPY --from=coach-engine /src/engine/arasan/generated/native/ /app/coach-engine/
ENV COACH_ENGINE_PATH=/app/coach-engine/arasan
ENV PORT=8080
EXPOSE 8080
USER 65532:65532
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/app/tinychess", "-healthcheck"]
ENTRYPOINT ["/app/tinychess"]
