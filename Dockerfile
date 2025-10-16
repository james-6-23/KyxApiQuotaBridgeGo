# ================================
# 第一阶段：构建前端
# ================================
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

# 复制前端依赖文件
COPY frontend/package*.json ./

# 安装依赖
RUN npm ci --only=production

# 复制前端源代码
COPY frontend/ .

# 构建前端（输出到 ../web）
RUN npm run build

# ================================
# 第二阶段：构建后端
# ================================
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app

# 安装必要的构建工具
RUN apk add --no-cache git gcc musl-dev

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制后端源代码
COPY . .

# 从前端构建阶段复制静态文件
COPY --from=frontend-builder /app/web ./web

# 构建后端应用（启用 CGO 以支持 SQLite）
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o main .

# ================================
# 第三阶段：运行环境
# ================================
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为亚洲/上海
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制文件和静态文件
COPY --from=backend-builder /app/main .
COPY --from=backend-builder /app/web ./web

# 创建数据目录
RUN mkdir -p /app/data

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# 运行应用
CMD ["./main"]

