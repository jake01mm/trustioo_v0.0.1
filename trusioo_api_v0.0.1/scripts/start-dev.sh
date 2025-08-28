#!/bin/bash

# Trusioo API 开发环境启动脚本

set -e

echo "🚀 Starting Trusioo API Development Environment..."

# 检查是否安装了Docker和Docker Compose
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# 检查环境变量文件
if [ ! -f .env ]; then
    echo "📋 Creating .env file from .env.example..."
    cp .env.example .env
    echo "✅ .env file created. Please review and update the configuration if needed."
fi

# 创建必要的目录
mkdir -p tmp logs

echo "🐳 Starting services with Docker Compose..."

# 启动基础服务（PostgreSQL 和 Redis）
echo "📦 Starting PostgreSQL and Redis..."
docker-compose up -d postgres redis

# 等待数据库启动
echo "⏳ Waiting for PostgreSQL to be ready..."
until docker-compose exec postgres pg_isready -U postgres; do
    sleep 1
done

echo "⏳ Waiting for Redis to be ready..."
until docker-compose exec redis redis-cli ping; do
    sleep 1
done

# 运行数据库迁移
echo "🔄 Running database migrations..."
docker-compose --profile migrate up migrate

# 启动应用程序
echo "🎯 Starting API server..."
docker-compose up -d app

echo "✅ All services are up and running!"
echo ""
echo "📍 Service URLs:"
echo "   🌐 API Server: http://localhost:8080"
echo "   🗄️ Database (Adminer): http://localhost:8081"
echo "   📊 Redis Commander: http://localhost:8082"
echo ""
echo "🔍 Health Check Endpoints:"
echo "   📊 Overall Health: http://localhost:8080/health"
echo "   🗄️ Database Health: http://localhost:8080/health/database"
echo "   🔗 Redis Health: http://localhost:8080/health/redis"
echo "   🚀 API v1 Health: http://localhost:8080/health/api/v1"
echo ""
echo "🔐 Authentication Endpoints:"
echo "   👤 Admin Login: POST http://localhost:8080/api/v1/auth/admin/login"
echo "   👥 User Register: POST http://localhost:8080/api/v1/auth/user/register"
echo "   👥 User Login: POST http://localhost:8080/api/v1/auth/user/login"
echo "   🏢 Buyer Register: POST http://localhost:8080/api/v1/auth/buyer/register"
echo "   🏢 Buyer Login: POST http://localhost:8080/api/v1/auth/buyer/login"
echo ""
echo "📝 Default Admin Account:"
echo "   Email: admin@trusioo.com"
echo "   Password: admin123"
echo ""
echo "📋 Useful Commands:"
echo "   📊 View logs: docker-compose logs -f app"
echo "   🛑 Stop services: docker-compose down"
echo "   🔄 Restart app: docker-compose restart app"
echo "   🧹 Clean up: docker-compose down -v"
echo ""
echo "🎉 Development environment is ready!"