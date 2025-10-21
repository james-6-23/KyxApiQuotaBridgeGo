# ==========================================
# Multi-Stage Dockerfile for KyxApiQuotaBridge
# 前后端一体化构建
# ==========================================

# ==================== 阶段 1: 前端构建 ====================
FROM node:18-alpine AS frontend-builder

# 设置工作目录
WORKDIR /frontend

# 复制前端依赖文件
COPY frontend/package*.json ./

# 安装依赖
RUN npm ci --only=production

# 复制前端源代码
COPY frontend/ ./

# 构建前端（生成静态文件到 dist 目录）
RUN npm run build

# ==================== 阶段 2: 后端构建 ====================
FROM golang:1.21-alpine AS backend-builder

# 设置工作目录
WORKDIR /build

# 安装构建依赖
RUN apk add --no-cache \
    git \
    gcc \
    musl-dev \
    ca-certificates \
    tzdata

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download && go mod verify

# 复制后端源代码
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/

# 构建参数
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

# 构建后端应用（静态编译，优化大小）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s \
    -X main.Version=${VERSION} \
    -X main.BuildTime=${BUILD_TIME} \
    -X main.GitCommit=${GIT_COMMIT}" \
    -a -installsuffix cgo \
    -o kyx-quota-bridge \
    ./cmd/server

# 验证二进制文件
RUN chmod +x kyx-quota-bridge

# ==================== 阶段 3: 最终运行镜像 ====================
FROM alpine:3.18

# 安装运行时依赖
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    wget \
    && rm -rf /var/cache/apk/*

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户和必要目录
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    mkdir -p /app/logs /app/web /app/data && \
    chown -R appuser:appuser /app

# 设置工作目录
WORKDIR /app

# 从后端构建阶段复制二进制文件
COPY --from=backend-builder /build/kyx-quota-bridge .
COPY --from=backend-builder /usr/share/zoneinfo /usr/share/zoneinfo

# 从前端构建阶段复制静态文件
COPY --from=frontend-builder --chown=appuser:appuser /frontend/dist ./web

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# 设置默认环境变量
ENV SERVER_PORT=8080 \
    SERVER_MODE=release \
    LOG_LEVEL=info \
    LOG_FORMAT=json

# 元数据标签
LABEL maintainer="kyx-quota-bridge" \
      org.opencontainers.image.title="KyxApiQuotaBridge" \
      org.opencontainers.image.description="公益站额度自助领取系统 - 前后端一体化部署" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.revision="${GIT_COMMIT}"

# 启动应用
ENTRYPOINT ["/app/kyx-quota-bridge"]
CMD []
