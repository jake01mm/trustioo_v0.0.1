# Makefile for Trusioo API

.PHONY: help build test clean dev docker up down logs migrate-up migrate-down

# Default target
.DEFAULT_GOAL := help

# 应用程序信息
APP_NAME := trusioo_api
VERSION := v0.0.1
BINARY_NAME := main

# Go相关变量
GO_CMD := go
GO_BUILD := $(GO_CMD) build
GO_CLEAN := $(GO_CMD) clean
GO_TEST := $(GO_CMD) test
GO_GET := $(GO_CMD) get
GO_MOD_TIDY := $(GO_CMD) mod tidy

# Docker相关变量
DOCKER_COMPOSE := docker-compose
MIGRATE_VERSION := v4.16.2

help: ## 显示此帮助信息
	@echo "Trusioo API - 可用命令:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 开发相关命令
dev: ## 启动开发环境
	./scripts/start-dev.sh

build: ## 编译应用程序
	$(GO_BUILD) -o $(BINARY_NAME) cmd/server/main.go

test: ## 运行测试
	$(GO_TEST) -v ./...

test-coverage: ## 运行测试并生成覆盖率报告
	$(GO_TEST) -v -coverprofile=coverage.out ./...
	$(GO_CMD) tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告生成: coverage.html"

clean: ## 清理构建文件
	$(GO_CLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

tidy: ## 整理依赖
	$(GO_MOD_TIDY)

# Docker相关命令
docker-build: ## 构建Docker镜像
	$(DOCKER_COMPOSE) build

up: ## 启动所有服务
	$(DOCKER_COMPOSE) up -d

down: ## 停止所有服务
	$(DOCKER_COMPOSE) down

logs: ## 查看应用程序日志
	$(DOCKER_COMPOSE) logs -f app

logs-all: ## 查看所有服务日志
	$(DOCKER_COMPOSE) logs -f

restart: ## 重启应用程序
	$(DOCKER_COMPOSE) restart app

# 数据库相关命令
migrate-up: ## 执行数据库迁移
	$(DOCKER_COMPOSE) --profile migrate up migrate

migrate-down: ## 回滚数据库迁移
	docker run --rm -v $(PWD)/migrations:/migrations --network $(shell docker-compose ps -q postgres | head -1 | xargs docker inspect -f '{{range .NetworkSettings.Networks}}{{.NetworkID}}{{end}}') \
		migrate/migrate:$(MIGRATE_VERSION) \
		-path=/migrations \
		-database="postgres://postgres:password@$$(docker-compose ps -q postgres | head -1 | xargs docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}'):5432/trusioo_api?sslmode=disable" \
		down 1

db-reset: ## 重置数据库
	$(DOCKER_COMPOSE) down postgres
	docker volume rm trusioo_api_v001_postgres_data || true
	$(DOCKER_COMPOSE) up -d postgres
	@echo "等待数据库启动..."
	@sleep 10
	$(MAKE) migrate-up

# 工具命令
tools: ## 启动管理工具 (Adminer + Redis Commander)
	$(DOCKER_COMPOSE) --profile tools up -d adminer redis-commander
	@echo "管理工具已启动:"
	@echo "  Adminer (数据库): http://localhost:8081"
	@echo "  Redis Commander: http://localhost:8082"

stop-tools: ## 停止管理工具
	$(DOCKER_COMPOSE) stop adminer redis-commander

# 代码质量
lint: ## 运行代码检查
	@command -v golangci-lint >/dev/null 2>&1 || { echo "请安装 golangci-lint"; exit 1; }
	golangci-lint run

fmt: ## 格式化代码
	$(GO_CMD) fmt ./...

vet: ## 运行go vet检查
	$(GO_CMD) vet ./...

# 生产环境命令
build-prod: ## 构建生产环境镜像
	docker build -f docker/Dockerfile -t $(APP_NAME):$(VERSION) .

deploy-prod: build-prod ## 部署到生产环境
	@echo "生产环境部署功能待实现"

# API测试命令
test-health: ## 测试健康检查端点
	@echo "测试健康检查端点..."
	curl -s http://localhost:8080/health | jq .
	curl -s http://localhost:8080/health/database | jq .
	curl -s http://localhost:8080/health/redis | jq .

test-api: ## 测试API端点
	@echo "测试API版本端点..."
	curl -s http://localhost:8080/api/v1 | jq .

# 清理命令
clean-docker: ## 清理Docker资源
	$(DOCKER_COMPOSE) down -v
	docker system prune -f
	docker volume prune -f

clean-all: clean clean-docker ## 清理所有资源

# 安装依赖
install-tools: ## 安装开发工具
	@echo "安装开发工具..."
	$(GO_CMD) install github.com/cosmtrek/air@latest
	@echo "工具安装完成"

# 状态检查
status: ## 检查服务状态
	$(DOCKER_COMPOSE) ps

# 数据库连接
db-shell: ## 连接到数据库
	$(DOCKER_COMPOSE) exec postgres psql -U postgres -d trusioo_api

redis-shell: ## 连接到Redis
	$(DOCKER_COMPOSE) exec redis redis-cli