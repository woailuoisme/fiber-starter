# syntax=docker/dockerfile:1.7

FROM golang:1.26.2-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    GOFLAGS=-mod=mod go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    GOFLAGS=-mod=mod CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/fiber-starter ./cmd/server

FROM alpine:3.23

RUN apk --no-cache add ca-certificates tzdata \
    && adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /out/fiber-starter /app/fiber-starter

ENV APP_ENV=production \
    APP_PORT=8080 \
    APP_HOST=0.0.0.0 \
    TZ=Asia/Shanghai

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/health >/dev/null || exit 1

USER appuser

CMD ["/app/fiber-starter"]
