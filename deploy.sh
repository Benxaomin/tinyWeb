#!/bin/bash

# TinyWeb1 部署脚本
# ====================

PROJECT_DIR="/home/user/tinyweb1"
FRONTEND_DIR="$PROJECT_DIR/fronted"
SERVER_CODE_DIR="$PROJECT_DIR/server(数据库代码)"

# 1. 进入项目目录
cd $PROJECT_DIR || exit 1

# 2. 拉取最新代码
echo "📥 拉取最新代码..."
git pull origin feature/html-pages

# 3. 编译 Go 程序
echo "🔨 编译 Go 程序..."
cd "$SERVER_CODE_DIR"
go build -o "$PROJECT_DIR/server.exe" main.go
if [ $? -ne 0 ]; then
    echo "❌ 编译失败"
    exit 1
fi
cd $PROJECT_DIR

# 4. 设置环境变量
echo "⚙️ 设置环境变量..."
export STATIC_DIR=$FRONTEND_DIR
export SERVER_PORT=8080
export APP_ENV=production

# 5. 停止旧服务
echo "🛑 停止旧服务..."
pkill -f "server.exe" 2>/dev/null || true
pkill -f "main" 2>/dev/null || true

# 等待进程完全退出
sleep 2

# 6. 启动新服务
echo "🚀 启动服务..."
nohup ./server.exe > server.log 2>&1 &

# 等待服务启动
sleep 3

# 6. 检查服务状态
echo "🔍 检查服务状态..."
if pgrep -f "server.exe" > /dev/null; then
    echo "✅ 服务启动成功！"
    echo "📍 访问地址: http://$(curl -s ifconfig.me):8080"
    echo "📂 静态文件目录: $STATIC_DIR"
    ps aux | grep -v grep | grep server.exe
else
    echo "❌ 服务启动失败，查看日志："
    tail -20 server.log
    exit 1
fi
