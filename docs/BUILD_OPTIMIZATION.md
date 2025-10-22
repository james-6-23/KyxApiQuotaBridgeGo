# Docker 构建性能优化指南

## 问题分析

多平台构建（AMD64 + ARM64）在 GitHub Actions 上非常慢，主要原因：

1. **ARM64 模拟慢**: GitHub Actions 运行在 AMD64 机器上，ARM64 通过 QEMU 模拟，速度慢 5-10 倍
2. **前端构建重**: 340 个 npm 包，TypeScript 类型检查 + Vite 构建耗时长
3. **重复构建**: 两个平台都要构建一遍前端和后端

## 优化方案

### 方案 1: 只构建 AMD64（推荐用于开发/测试）

**时间节省**: ~60-70%

修改 GitHub Actions workflow (`.github/workflows/docker-build.yml`):

```yaml
- name: Build and push
  uses: docker/build-push-action@v5
  with:
    platforms: linux/amd64  # 只构建 AMD64
    # ... 其他配置
```

**适用场景**:
- 开发环境
- 测试环境
- 大部分云服务器（AWS、GCP、Azure 主要是 AMD64）

### 方案 2: 跳过 TypeScript 类型检查（推荐用于生产构建）

**时间节省**: ~20-30% (前端构建部分)

在 GitHub Actions workflow 中添加构建参数:

```yaml
- name: Build and push
  uses: docker/build-push-action@v5
  with:
    build-args: |
      VERSION=${{ github.ref_name }}
      BUILD_TIME=${{ steps.prep.outputs.build_time }}
      GIT_COMMIT=${{ github.sha }}
      SKIP_TYPE_CHECK=true  # 跳过类型检查
```

**注意**:
- 类型检查应该在 CI 的单独步骤中完成
- 不影响运行时性能
- 生产环境可以安全使用

### 方案 3: 使用构建缓存（已启用）

当前已经在使用 GitHub Actions 缓存:

```yaml
cache-from: type=gha
cache-to: type=gha,mode=max
```

这会缓存：
- Go 模块依赖
- npm 依赖
- Docker 层

**效果**: 第二次构建速度提升 40-60%

### 方案 4: 组合优化（最快）

**时间节省**: ~80-85%

```yaml
- name: Build and push
  uses: docker/build-push-action@v5
  with:
    platforms: linux/amd64  # 只构建 AMD64
    build-args: |
      SKIP_TYPE_CHECK=true  # 跳过类型检查
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

## 性能对比

| 方案 | 构建时间 (首次) | 构建时间 (缓存) | 适用场景 |
|-----|----------------|----------------|---------|
| 默认 (双平台 + 类型检查) | ~15-20 分钟 | ~8-12 分钟 | 正式发布 |
| 只 AMD64 | ~8-10 分钟 | ~4-6 分钟 | 开发/测试 |
| 跳过类型检查 | ~12-15 分钟 | ~6-9 分钟 | 生产（CI 已检查） |
| 组合优化 | ~5-7 分钟 | ~2-4 分钟 | 快速迭代 |

## 本地构建优化

### 跳过类型检查

```bash
docker build \
  --build-arg SKIP_TYPE_CHECK=true \
  -t kyx-quota-bridge .
```

### 只构建 AMD64

```bash
docker build \
  --platform linux/amd64 \
  -t kyx-quota-bridge .
```

### 使用本地缓存

```bash
# 首次构建
docker build -t kyx-quota-bridge .

# 后续构建会自动使用缓存
docker build -t kyx-quota-bridge .
```

### 并行构建（如果需要双平台）

使用 BuildKit 的并行构建：

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --cache-from type=registry,ref=ghcr.io/username/kyx-quota-bridge:cache \
  --cache-to type=registry,ref=ghcr.io/username/kyx-quota-bridge:cache,mode=max \
  -t kyx-quota-bridge .
```

## 推荐配置

### 开发环境

```yaml
# .github/workflows/docker-build-dev.yml
platforms: linux/amd64
build-args: |
  SKIP_TYPE_CHECK=true
```

### 生产环境

```yaml
# .github/workflows/docker-build-prod.yml
platforms: linux/amd64,linux/arm64
build-args: |
  SKIP_TYPE_CHECK=true  # 前提：CI 已经单独运行了类型检查
```

### 正式发布

```yaml
# .github/workflows/docker-build-release.yml
platforms: linux/amd64,linux/arm64
# 不跳过类型检查，确保完整构建
```

## 额外优化建议

### 1. 拆分前端构建到单独的 Job

```yaml
jobs:
  build-frontend:
    runs-on: ubuntu-latest
    steps:
      - name: Build frontend
        run: npm run build
      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: frontend-dist
          path: frontend/dist

  build-docker:
    needs: build-frontend
    runs-on: ubuntu-latest
    steps:
      - name: Download frontend artifact
        uses: actions/download-artifact@v3
      # ... Docker 构建（跳过前端构建步骤）
```

### 2. 使用预构建的基础镜像

创建自定义基础镜像，包含所有依赖：

```dockerfile
# base.Dockerfile
FROM node:18-alpine AS frontend-base
WORKDIR /frontend
COPY frontend/package*.json ./
RUN npm ci
```

### 3. ARM64 使用交叉编译而非模拟

```dockerfile
# 在 AMD64 机器上交叉编译 ARM64
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS backend-builder
ARG TARGETPLATFORM
ARG BUILDPLATFORM
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$(echo $TARGETPLATFORM | cut -d/ -f2) go build ...
```

## 监控构建时间

在 GitHub Actions 中添加时间统计：

```yaml
- name: Build start
  run: echo "BUILD_START=$(date +%s)" >> $GITHUB_ENV

- name: Build and push
  # ... 构建步骤

- name: Build duration
  run: |
    BUILD_END=$(date +%s)
    DURATION=$((BUILD_END - BUILD_START))
    echo "Build took $DURATION seconds"
```
