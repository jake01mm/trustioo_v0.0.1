# Trusioo API v0.0.1

基于Go Gin框架的高性能API服务，使用PostgreSQL数据库和Redis缓存。

## 项目特性

- **框架**: Go Gin (高性能Web框架)
- **数据库**: PostgreSQL (使用原生SQL，无ORM)
- **缓存**: Redis
- **开发环境**: Docker + Air热重载
- **架构**: 清洁架构 + 模块化设计
- **认证**: JWT认证机制
- **健康检查**: 完整的健康监控API端点
- **安全**: CORS、限流、请求超时等安全机制

## 项目架构

本项目采用模块化的清洁架构设计：

- **cmd/**: 应用程序入口点
- **internal/**: 内部包，不对外暴露
  - **config/**: 配置管理
  - **infrastructure/**: 基础设施层（数据库、Redis、路由）
  - **middleware/**: 中间件
  - **modules/**: 业务模块
    - **auth/**: 认证模块（支持管理员、用户、买家三种角色）
    - **health/**: 健康检查模块
- **migrations/**: 数据库迁移文件
- **docker/**: Docker配置文件
- **scripts/**: 工具脚本

## 快速开始

### 前置要求

- Docker & Docker Compose
- Go 1.21+ (可选，用于本地开发)
- Make (可选，用于便捷命令)

### 1. 克隆项目

```bash
git clone <repository-url>
cd trusioo_api_v0.0.1
```

### 2. 启动开发环境

#### 使用Make命令（推荐）

```bash
# 查看可用命令
make help

# 启动开发环境
make dev
```

#### 手动启动

```bash
# 复制环境变量文件
cp .env.example .env

# 启动服务
./scripts/start-dev.sh
```

### 3. 验证安装

```bash
# 检查服务状态
make status

# 运行API测试
./scripts/test-api.sh

# 检查健康状态
curl http://localhost:8080/health
```

## 项目结构

```
trusioo_api_v0.0.1/
├── cmd/                    # 应用程序入口
│   └── server/
├── internal/               # 内部包
│   ├── modules/           # 业务模块
│   │   ├── auth/         # 认证模块
│   │   │   ├── admin/    # 管理员认证
│   │   │   ├── user/     # 用户认证
│   │   │   └── buyer/    # 买家认证
│   │   └── health/       # 健康检查模块
│   ├── infrastructure/   # 基础设施层
│   │   ├── database/    # 数据库连接
│   │   ├── redis/       # Redis连接
│   │   └── router/      # 路由配置
│   ├── config/          # 配置管理
│   └── middleware/      # 中间件
├── migrations/            # 数据库迁移文件
├── docker/               # Docker配置
├── scripts/              # 脚本文件
├── .env.example          # 环境变量示例
├── docker-compose.yml    # Docker编排
├── .air.toml             # Air热重载配置
└── go.mod               # Go模块文件
```

## 快速开始

### 1. 环境准备

```bash
# 复制环境变量文件
cp .env.example .env

# 编辑环境变量
vim .env
```

### 2. 使用Docker开发

```bash
# 启动开发环境
docker-compose up -d

# 查看日志
docker-compose logs -f app
```

### 3. 本地开发

```bash
# 安装Air热重载工具
go install github.com/cosmtrek/air@latest

# 启动开发服务器
air

# 或者直接运行
go run cmd/server/main.go
```

## API端点

### 健康检查

- `GET /health` - 整体健康状态
- `GET /health/database` - 数据库健康状态
- `GET /health/redis` - Redis健康状态
- `GET /health/api/v1` - API v1版本健康状态

### 认证模块

- `POST /api/v1/auth/admin/login` - 管理员登录
- `POST /api/v1/auth/user/login` - 用户登录
- `POST /api/v1/auth/buyer/login` - 买家登录

## 数据库迁移

```bash
# 创建迁移文件
migrate create -ext sql -dir migrations -seq create_users_table

# 执行迁移
migrate -path migrations -database "postgres://user:password@localhost:5432/database?sslmode=disable" up

# 回滚迁移
migrate -path migrations -database "postgres://user:password@localhost:5432/database?sslmode=disable" down
```

## 开发指南

### 添加新模块

1. 在 `internal/modules/` 下创建新的模块目录
2. 实现对应的 handler, service, repository 层
3. 在 `internal/infrastructure/router/` 中注册路由
4. 添加相应的测试文件

### 代码规范

- 使用 `gofmt` 格式化代码
- 遵循 Go 官方编码规范
- 每个公共函数和结构体都需要注释
- 单元测试覆盖率要求 > 80%

## 环境变量

详见 `.env.example` 文件中的配置说明。

## 贡献

请阅读 [CONTRIBUTING.md](CONTRIBUTING.md) 了解详细的贡献指南。

## 许可证

[MIT License](LICENSE)