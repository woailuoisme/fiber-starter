# syntax=docker/dockerfile:1.7

FROM golang:1.26.3-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates

ARG TARGETOS
ARG TARGETARCH

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    GOFLAGS=-mod=mod go mod download

COPY app ./app
COPY bootstrap ./bootstrap
COPY cmd ./cmd
COPY routes ./routes
COPY configs/config.go ./configs/config.go
COPY configs/internal ./configs/internal
COPY database/factories ./database/factories
COPY database/seeders ./database/seeders
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    GOFLAGS=-mod=mod CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/fiber-starter ./cmd/app

FROM alpine:3.23

RUN apk --no-cache add ca-certificates tzdata \
    && adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /out/fiber-starter /app/fiber-starter
COPY configs /app/configs
COPY docs /app/docs
COPY lang /app/lang
COPY public /app/public

RUN mkdir -p /app/storage/app /app/storage/logs /app/storage/framework/cache /app/storage/framework/sessions /app/storage/framework/views \
    && chown -R appuser:appuser /app/storage

ENV APP_ENV=production \
    CONFIG_WARN_MISSING_ENV_FILE=false \
    APP_PORT=8080 \
    APP_HOST=0.0.0.0 \
    TZ=Asia/Shanghai

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/health >/dev/null || exit 1

USER appuser

CMD ["/app/fiber-starter", "serve"]
