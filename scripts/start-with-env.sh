#!/bin/bash

# AI Gateway 启动脚本（使用环境变量）
# 用法: ./scripts/start-with-env.sh

# 设置脚本在遇到错误时退出
set -e

echo "🚀 启动 AI Gateway..."

# 检查 .env 文件是否存在
if [ ! -f .env ]; then
    echo "⚠️  未找到 .env 文件"
    echo "📝 正在从 .env.example 创建 .env..."
    cp .env.example .env
    echo "✅ 已创建 .env 文件，请编辑填入真实值"
    echo ""
    exit 1
fi

# 加载环境变量
echo "📦 加载环境变量..."
export $(cat .env | grep -v '^#' | xargs)

# 验证必需的环境变量
if [ -z "$DB_PASSWORD" ]; then
    echo "❌ 错误: DB_PASSWORD 环境变量未设置"
    exit 1
fi

if [ -z "$JWT_SECRET" ]; then
    echo "❌ 错误: JWT_SECRET 环境变量未设置"
    exit 1
fi

echo "✅ 环境变量已加载"
echo "   DB_HOST: ${DB_HOST:-未设置}"
echo "   DB_USER: ${DB_USER:-未设置}"
echo "   DB_NAME: ${DB_NAME:-未设置}"
echo ""

# 启动服务
echo "🔧 启动服务..."
go run cmd/server/main.go --config=./config/config.yaml
