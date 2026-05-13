#!/bin/bash

# TinyWeb1 高效部署脚本
# 用法: ./deploy.sh [分支名]
# 示例: ./deploy.sh feature/html-pages

set -e

BRANCH=${1:-feature/html-pages}
PROJECT_DIR=~/tinyWeb1
SERVER_DIR="$PROJECT_DIR/server(数据库代码)"
PORT=8080

echo "========================================"
echo "  🚀 TinyWeb1 快速部署"
echo "========================================"
echo "分支: $BRANCH"

# 1. Git操作
cd "$PROJECT_DIR"
git fetch origin
git reset --hard "origin/$BRANCH"

# 2. 强制释放端口（关键）
fuser -k $PORT/tcp 2>/dev/null || true

# 3. 编译
cd "$SERVER_DIR"
go build -o server.exe main.go

# 4. 设置环境变量并启动
export STATIC_DIR="$PROJECT_DIR/fronted"
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASS=
export DB_NAME=tinyweb1
nohup bash -c "./server.exe" > server.log 2>&1 &

# 5. 健康检查
sleep 2
if curl -s http://localhost:$PORT/api/health > /dev/null; then
    echo "========================================"
    echo "  ✅ 部署成功"
    echo "========================================"
    echo "服务正在运行 (PID: $(pgrep -f server.exe))"
    echo "访问: http://1.15.224.88:$PORT"
    echo ""
else 
    echo "❌ 服务启动失败，查看日志:"
    tail -20 server.log
    exit 1
fi
