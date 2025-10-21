#!/bin/bash

# ==========================================
# KyxApiQuotaBridge 快速部署脚本
# ==========================================

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 打印标题
print_header() {
    echo ""
    echo "=========================================="
    echo "  $1"
    echo "=========================================="
    echo ""
}

# 检查命令是否存在
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# 生成随机密码
generate_password() {
    local length=$1
    openssl rand -base64 $length | tr -d "=+/" | cut -c1-$length
}

# ==================== 主程序 ====================

print_header "KyxApiQuotaBridge 快速部署"

# 1. 检查系统依赖
print_info "检查系统依赖..."

if ! command_exists docker; then
    print_error "Docker 未安装，请先安装 Docker"
    echo "安装指南: https://docs.docker.com/engine/install/"
    exit 1
fi

if ! command_exists docker-compose; then
    print_error "Docker Compose 未安装，请先安装 Docker Compose"
    echo "安装指南: https://docs.docker.com/compose/install/"
    exit 1
fi

print_success "系统依赖检查通过"

# 2. 检查 .env 文件
print_info "检查配置文件..."

if [ ! -f .env ]; then
    print_warning ".env 文件不存在，正在创建..."

    if [ ! -f .env.example ]; then
        print_error ".env.example 文件不存在"
        exit 1
    fi

    cp .env.example .env
    print_success ".env 文件已创建"

    # 自动生成密码
    print_info "自动生成安全密码..."

    DB_PASSWORD=$(generate_password 32)
    REDIS_PASSWORD=$(generate_password 32)
    ADMIN_PASSWORD=$(generate_password 24)
    JWT_SECRET=$(generate_password 64)

    # 替换密码
    sed -i "s/DB_PASSWORD=.*/DB_PASSWORD=$DB_PASSWORD/" .env
    sed -i "s/REDIS_PASSWORD=.*/REDIS_PASSWORD=$REDIS_PASSWORD/" .env
    sed -i "s/ADMIN_PASSWORD=.*/ADMIN_PASSWORD=$ADMIN_PASSWORD/" .env
    sed -i "s/JWT_SECRET=.*/JWT_SECRET=$JWT_SECRET/" .env

    print_success "密码已自动生成并保存到 .env 文件"

    # 提示用户配置必需项
    print_warning "请编辑 .env 文件，配置以下必需项："
    echo "  1. GITHUB_USERNAME - 你的 GitHub 用户名"
    echo "  2. LINUX_DO_CLIENT_ID - Linux Do OAuth2 客户端 ID"
    echo "  3. LINUX_DO_CLIENT_SECRET - Linux Do OAuth2 客户端密钥"
    echo "  4. LINUX_DO_REDIRECT_URI - OAuth2 回调地址"
    echo ""
    read -p "配置完成后，按回车继续..."
else
    print_success ".env 文件已存在"
fi

# 3. 创建必要目录
print_info "创建必要目录..."
mkdir -p logs backups data migrations
print_success "目录创建完成"

# 4. 拉取 Docker 镜像
print_info "拉取 Docker 镜像（这可能需要几分钟）..."
if docker-compose pull; then
    print_success "镜像拉取完成"
else
    print_error "镜像拉取失败，请检查网络连接和 GITHUB_USERNAME 配置"
    exit 1
fi

# 5. 启动服务
print_info "启动服务..."
if docker-compose up -d; then
    print_success "服务启动成功"
else
    print_error "服务启动失败"
    exit 1
fi

# 6. 等待服务就绪
print_info "等待服务就绪（最多等待 60 秒）..."
max_wait=60
wait_time=0

while [ $wait_time -lt $max_wait ]; do
    if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
        print_success "服务已就绪！"
        break
    fi

    echo -n "."
    sleep 2
    wait_time=$((wait_time + 2))
done

echo ""

if [ $wait_time -ge $max_wait ]; then
    print_warning "服务启动超时，请检查日志"
    docker-compose logs --tail=50
    exit 1
fi

# 7. 显示服务状态
print_header "部署完成"

docker-compose ps

echo ""
print_success "KyxApiQuotaBridge 已成功部署！"
echo ""
echo "访问地址:"
echo "  - 应用前端: http://localhost:8080"
echo "  - 健康检查: http://localhost:8080/health"
echo ""
echo "管理员密码保存在 .env 文件中 (ADMIN_PASSWORD)"
echo ""
echo "常用命令:"
echo "  - 查看日志: docker-compose logs -f"
echo "  - 重启服务: docker-compose restart"
echo "  - 停止服务: docker-compose stop"
echo "  - 查看状态: docker-compose ps"
echo ""
print_info "更多信息请查看 DEPLOYMENT_GUIDE.md"
