# 构建阶段
FROM golang:1.26.0-alpine AS builder

WORKDIR /app

# 复制依赖并下载
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# 复制源码并构建
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o main .

# 运行阶段
FROM alpine:3.23

# 安装依赖并创建用户
RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -g '' appuser

WORKDIR /app

# 复制文件
COPY --from=builder /app/main .
COPY --from=builder /app/.env* ./

# 配置环境
ENV APP_ENV=production \
    TZ=Asia/Shanghai

USER appuser
EXPOSE 8080

CMD ["./main"]
