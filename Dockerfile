# syntax=docker/dockerfile:1
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
