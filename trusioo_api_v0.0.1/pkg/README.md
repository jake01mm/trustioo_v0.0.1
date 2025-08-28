# 企业级工具包 (pkg)

这个目录包含了项目的核心企业级工具和组件，这些组件是可复用的，遵循Go项目最佳实践。

## 📁 目录结构

```
pkg/
├── README.md              # 本文档
├── response/              # 统一响应格式
│   └── response.go        # Response包装器、错误码标准化、分页响应
├── errors/                # 错误处理机制
│   └── errors.go          # 自定义错误类型、错误码枚举
├── middleware/            # 中间件集合
│   └── error_handler.go   # 全局错误处理器
├── validator/             # 输入验证和数据绑定
│   └── validator.go       # Validator中间件、自定义验证规则
├── logger/                # 增强日志系统
│   └── logger.go          # 结构化日志、调用链追踪、性能监控日志
├── swagger/               # API文档
│   └── swagger.go         # Swagger/OpenAPI自动生成
├── metrics/               # 监控指标
│   └── metrics.go         # Prometheus指标、性能分析
├── cache/                 # 缓存策略
│   └── cache.go           # 缓存抽象层、缓存策略、失效机制
├── security/              # 安全增强
│   └── security.go        # API限流、SQL注入防护、XSS防护
├── config/                # 配置管理增强
│   └── manager.go         # 配置热重载、配置验证、多环境配置
└── examples/              # 使用示例
    ├── handler_example.go # 基础处理器示例
    └── enterprise_app.go  # 企业级应用集成示例
```

## 🚀 核心功能

### 1. 统一响应格式 (response)
- ✅ **Response包装器**: 标准化所有API响应格式
- ✅ **错误码标准化**: 统一的错误码管理
- ✅ **分页响应**: 支持分页数据的标准响应格式

### 2. 错误处理机制 (errors)
- ✅ **自定义错误类型**: AppError结构体
- ✅ **错误码枚举**: 预定义的错误码常量
- ✅ **全局错误处理器**: 统一的错误处理中间件

### 3. 输入验证和数据绑定 (validator)
- ✅ **Validator中间件**: 自动验证请求数据
- ✅ **自定义验证规则**: 扩展验证功能
- ✅ **多语言错误消息**: 支持中英文错误提示

### 4. 增强日志系统 (logger)
- ✅ **结构化日志**: JSON格式的结构化日志
- ✅ **调用链追踪**: 请求ID追踪
- ✅ **性能监控日志**: 请求耗时监控
- ✅ **日志轮转**: 文件大小和时间轮转

### 5. 密码加密和安全 (crypto)
- ✅ **双重加密**: HMAC-SHA256 + bcrypt双重加密
- ✅ **密码强度验证**: 复杂度验证规则
- ✅ **密码管理器**: 错误限制和账号锁定
- ✅ **随机密码生成**: 安全随机密码生成

### 6. 增强中间件 (middleware)
- ✅ **全局错误处理**: 统一错误处理机制
- ✅ **请求ID追踪**: 自动生成和传递请求ID
- ✅ **日志中间件**: 结构化HTTP请求日志
- ✅ **恢复中间件**: Panic恢复和堆栈追踪
- ✅ **超时中间件**: 请求超时控制

### 7. API文档 (swagger)
- ✅ **Swagger/OpenAPI**: 自动生成API文档
- ✅ **交互式文档**: 在线API测试界面
- ✅ **文档配置**: 可配置的文档信息

### 8. 监控指标 (metrics)
- ✅ **Prometheus指标**: 完整的指标收集
- ✅ **性能分析**: HTTP请求性能监控
- ✅ **业务指标**: 用户注册、登录等业务指标
- ✅ **系统资源**: 内存、CPU使用率监控

### 9. 缓存策略 (cache)
- ✅ **缓存抽象层**: 支持多种缓存后端
- ✅ **缓存策略**: 不同类型数据的TTL策略
- ✅ **失效机制**: 标签失效、模式删除
- ✅ **缓存预热**: 应用启动时预热热点数据

### 10. 安全增强 (security)
- ✅ **API限流**: 令牌桶算法限流
- ✅ **SQL注入防护**: 请求参数SQL注入检测
- ✅ **XSS防护**: 跨站脚本攻击防护
- ✅ **安全头**: 完整的HTTP安全头设置
- ✅ **CORS配置**: 跨域资源共享配置

### 11. 配置管理增强 (config)
- ✅ **配置热重载**: 文件变化自动重载
- ✅ **配置验证**: 配置项验证机制
- ✅ **多环境配置**: 开发、测试、生产环境配置
- ✅ **多格式支持**: JSON、ENV等格式支持

## 🛠️ 使用方法

### 基础使用示例

```go
package main

import (
    "github.com/gin-gonic/gin"
    "trusioo_api_v0.0.1/pkg/response"
    "trusioo_api_v0.0.1/pkg/errors"
    "trusioo_api_v0.0.1/pkg/middleware"
)

func main() {
    router := gin.New()
    
    // 添加全局错误处理中间件
    router.Use(middleware.ErrorHandler(logger))
    
    // 示例API
    router.GET("/api/users", func(c *gin.Context) {
        users := []gin.H{
            {"id": 1, "name": "用户1"},
            {"id": 2, "name": "用户2"},
        }
        
        response.Success(c, users, "获取用户列表成功")
    })
    
    router.Run(":8080")
}
```

### 企业级集成示例

查看 `pkg/examples/enterprise_app.go` 文件，了解如何集成所有企业级组件。

### 快速启动

```bash
# 运行企业级示例应用
go run pkg/examples/enterprise_app.go

# 访问API文档
open http://localhost:8080/swagger/

# 查看监控指标
open http://localhost:8080/metrics

# 健康检查
curl http://localhost:8080/health
```

## 📊 监控和观测

### Prometheus指标

访问 `http://localhost:8080/metrics` 查看以下指标：

- `trusioo_api_http_requests_total` - HTTP请求总数
- `trusioo_api_http_request_duration_seconds` - 请求延迟分布
- `trusioo_api_user_registrations_total` - 用户注册总数
- `trusioo_api_user_logins_total` - 用户登录总数
- `trusioo_api_api_errors_total` - API错误总数

### API文档

访问 `http://localhost:8080/swagger/` 查看交互式API文档。

### 日志格式

```json
{
  "time": "2023-08-27T10:00:00Z",
  "level": "info",
  "msg": "HTTP request completed",
  "request_id": "req-123456",
  "method": "GET",
  "path": "/api/users",
  "status": 200,
  "duration": "15ms",
  "ip": "127.0.0.1"
}
```

## 🔧 配置说明

### 环境变量

```env
# 日志配置
LOG_LEVEL=info
LOG_FORMAT=json
LOG_FILE_PATH=logs/app.log

# Redis配置（缓存）
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# 限流配置
RATE_LIMIT_RPS=10
RATE_LIMIT_BURST=20

# 安全配置
ENABLE_RATE_LIMIT=true
ENABLE_SQL_PROTECTION=true
ENABLE_XSS_PROTECTION=true
```

## 🧪 测试

```bash
# 运行所有测试
go test ./pkg/...

# 测试特定包
go test ./pkg/response
go test ./pkg/validator

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./pkg/...
go tool cover -html=coverage.out
```

## 📝 最佳实践

1. **错误处理**: 始终使用统一的错误响应格式
2. **日志记录**: 在关键业务逻辑中添加结构化日志
3. **缓存使用**: 合理设置TTL，避免缓存穿透
4. **安全防护**: 在生产环境启用所有安全中间件
5. **监控指标**: 关注业务指标，及时发现问题
6. **配置管理**: 使用配置热重载，减少重启次数

## 🔄 升级指南

当前版本: v0.0.1

### 已完成功能
- ✅ 统一响应格式
- ✅ 错误处理机制
- ✅ 输入验证和数据绑定
- ✅ 增强日志系统
- ✅ API文档
- ✅ 监控指标
- ✅ 缓存策略
- ✅ 安全增强
- ✅ 配置管理增强

### 规划功能
- 🔄 测试框架
- 🔄 数据库连接池优化
- 🔄 部署和CI/CD

## 🤝 贡献指南

1. 遵循Go代码规范
2. 添加适当的单元测试
3. 更新相关文档
4. 提交前运行所有测试

## 📄 许可证

MIT License