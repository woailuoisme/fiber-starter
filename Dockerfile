# syntax=docker/dockerfile:1.7

# --- 阶段一：多阶段构建的 Go 编译器阶段 ---
FROM golang:1.26.3-alpine AS builder

WORKDIR /src

# 安装编译所需的系统工具与基础证书
RUN apk add --no-cache git ca-certificates

# 多架构编译目标系统与 CPU 架构参数
ARG TARGETOS
ARG TARGETARCH

# 缓存并下载 go modules 依赖包，加速二次构建
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    GOFLAGS=-mod=mod go mod download

# 复制核心业务源码与配置文件以进行编译（基于 .dockerignore 过滤无关文件）
COPY . .

# 挂载 Go 缓存、编译无符号与无调试信息的极简 Linux 静态二进制
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    GOFLAGS=-mod=mod CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/lfiber ./cmd/app

# --- 阶段二：生产运行的极简 Alpine 运行环境 ---
FROM alpine:3.23

# 安装 CA 证书、系统时区工具，并创建非 root 安全运行账号
RUN apk --no-cache add ca-certificates tzdata \
    && adduser -D -g '' appuser

WORKDIR /app

# 从构建器中复制已编译的二进制可执行文件与运行静态资源
COPY --from=builder /out/lfiber /app/lfiber
COPY configs /app/configs
COPY docs /app/docs
COPY lang /app/lang
COPY public /app/public

# 创建应用日志、持久化存储及框架缓存等运行必需目录，并变更宿主权限
RUN mkdir -p /app/storage/app /app/storage/logs /app/storage/framework/cache /app/storage/framework/sessions /app/storage/framework/views \
    && chown -R appuser:appuser /app/storage

# 设置容器运行默认环境变量（生产环境模式、端口、上海时区）
ENV APP_ENV=production \
    CONFIG_WARN_MISSING_ENV_FILE=false \
    APP_PORT=3100 \
    APP_HOST=0.0.0.0 \
    TZ=Asia/Shanghai

EXPOSE 3100

# 容器健康状态检测
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:3100/health >/dev/null || exit 1

# 以非 root 用户安全身份运行进程
USER appuser

CMD ["/app/lfiber", "serve"]

