#!/bin/bash

# 服务器部署脚本
# 用法: 
#   ./deploy.sh              # 默认拉取当前分支
#   ./deploy.sh main         # 拉取 main 分支
#   ./deploy.sh feature/xxx  # 拉取指定分支

# 获取分支名，默认为当前分支
BRANCH=${1:-$(git branch --show-current 2>/dev/null || echo "main")}

echo "========================================"
echo "  🚀 TinyWeb1 服务器部署脚本"
echo "========================================"
echo "  目标分支: $BRANCH"
echo ""

# 1. 进入项目目录
cd ~/tinyweb1 || exit 1

# 2. 拉取最新代码
echo "📥 拉取最新代码..."
git pull origin "$BRANCH"

# 3. 停止旧服务
echo "🛑 停止旧服务..."
pkill -f "server.exe" 2>/dev/null || echo "   没有运行中的服务"

# 4. 编译
echo "🔨 编译服务端..."
cd "server(数据库代码)" || exit 1
go build -o server.exe main.go

# 5. 设置环境变量
echo "⚙️  设置环境变量..."
export STATIC_DIR=/home/user/tinyweb1/fronted

# 6. 启动服务
echo "🚀 启动服务..."
nohup ./server.exe > server.log 2>&1 &

# 7. 等待服务启动
sleep 2

# 8. 检查状态
echo "========================================"
echo "  ✅ 部署完成！"
echo "========================================"
echo "服务状态:"
ps aux | grep server.exe | grep -v grep

echo ""
echo "日志文件: server(数据库代码)/server.log"
echo "访问地址: http://1.15.224.88:8080"
echo ""
