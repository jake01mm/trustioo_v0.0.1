# 🔄 组件迁移指南

本文档说明了如何将`internal/`目录中的某些组件迁移到`pkg/`目录，以实现更好的代码复用和统一管理。

## 📋 迁移概览

根据Go项目最佳实践和企业级项目规范，我们已经将以下通用组件从`internal/`迁移到`pkg/`：

### ✅ 已迁移组件

#### 1. **Middleware组件** 
**从**: `internal/middleware/` → **到**: `pkg/middleware/error_handler.go`

**包含功能**:
- ✅ 全局错误处理中间件
- ✅ 请求ID中间件 (新增)
- ✅ 日志中间件 (增强版)
- ✅ Recovery中间件 (增强版)
- ✅ 超时中间件 (新增)

**优势**:
- 统一的中间件管理
- 更强的错误处理能力
- 完整的请求追踪
- 更好的代码复用

#### 2. **密码加密组件**
**从**: `internal/crypto/password.go` → **到**: `pkg/crypto/password.go`

**包含功能**:
- ✅ HMAC-SHA256 + bcrypt双重加密
- ✅ 密码强度验证
- ✅ 密码管理器(防暴力破解)
- ✅ 随机密码生成

**优势**:
- 更强的安全性
- 完整的密码管理策略
- 可配置的加密方法
- 通用的安全工具

### ❌ 保留在internal的组件

#### 1. **`internal/config/`** - 应用特定配置
**原因**: 
- 业务相关的配置结构
- 与具体应用紧密耦合
- pkg/config/已提供更强大的配置管理器

#### 2. **`internal/infrastructure/`** - 基础设施配置
**原因**:
- 包含具体的数据库连接配置
- 特定的路由设置
- 应用层基础设施，不适合通用化

#### 3. **`internal/modules/`** - 业务模块
**原因**:
- 包含具体业务逻辑
- 认证、健康检查等业务模块
- 属于应用核心，应保持私有

## 🛠️ 代码更新指导

### 更新导入路径

如果您在其他文件中使用了已迁移的组件，需要更新导入路径：

#### 原来的导入:
```go
import (
    "trusioo_api_v0.0.1/internal/middleware"
    "trusioo_api_v0.0.1/internal/crypto"
)
```

#### 更新为:
```go
import (
    "trusioo_api_v0.0.1/pkg/middleware"
    "trusioo_api_v0.0.1/pkg/crypto"
)
```

### 中间件使用更新

#### 原来的使用方式:
```go
// 使用单独的中间件
router.Use(middleware.Logger(logger))
router.Use(middleware.RequestID())
router.Use(middleware.Recovery())
```

#### 新的推荐方式:
```go
// 使用增强的pkg/middleware
router.Use(middleware.RequestID())
router.Use(middleware.Logger(logger))
router.Use(middleware.Recovery(logger))
router.Use(middleware.ErrorHandler(logger))
router.Use(middleware.Timeout(30*time.Second, logger))
```

### 密码加密使用更新

#### 原来的使用方式:
```go
encryptor := crypto.NewPasswordEncryptor(key, "hmac-bcrypt")
```

#### 新的增强方式:
```go
// 使用配置创建
config := &crypto.PasswordConfig{
    EncryptionKey: key,
    Method:        "hmac-bcrypt",
    SaltLength:    32,
    BcryptCost:    bcrypt.DefaultCost,
}
encryptor := crypto.NewPasswordEncryptorWithConfig(config)

// 或者使用密码管理器
manager := crypto.NewPasswordManager(encryptor)
err := manager.VerifyPasswordWithLockout(userID, password, hashedPassword)
```

## 🎯 迁移带来的优势

### 1. **更好的代码复用**
- 通用组件可在多个项目中使用
- 减少重复代码
- 提高开发效率

### 2. **统一的工具管理**
- 所有企业级工具集中在pkg目录
- 便于维护和升级
- 清晰的组件边界

### 3. **增强的功能**
- 中间件功能更加完善
- 密码安全性显著提升
- 更好的错误处理和日志记录

### 4. **符合最佳实践**
- 遵循Go项目标准结构
- 清晰的public/private API边界
- 便于团队协作

## 📝 注意事项

### 1. **导入路径更新**
- 检查所有使用已迁移组件的文件
- 更新导入路径
- 运行测试确保功能正常

### 2. **功能兼容性**
- 新版本的中间件功能更强但API兼容
- 密码加密组件向后兼容
- 建议逐步采用新功能

### 3. **配置更新**
- 某些组件可能需要额外配置
- 查看各组件的配置文档
- 环境变量可能需要调整

## 🔧 自动化迁移脚本

如果您有大量文件需要更新导入路径，可以使用以下脚本：

```bash
#!/bin/bash
# 更新导入路径的脚本

# 更新middleware导入
find . -name "*.go" -exec sed -i 's|trusioo_api_v0.0.1/internal/middleware|trusioo_api_v0.0.1/pkg/middleware|g' {} \;

# 更新crypto导入
find . -name "*.go" -exec sed -i 's|trusioo_api_v0.0.1/internal/crypto|trusioo_api_v0.0.1/pkg/crypto|g' {} \;

echo "导入路径更新完成"
```

## 🚀 后续计划

### 可能的未来迁移

在评估后，以下组件可能在未来版本中考虑迁移或重构：

1. **部分infrastructure组件** - 提取通用的数据库连接池管理
2. **通用的module工具** - 提取可复用的业务组件模式
3. **配置整合** - 统一internal/config和pkg/config的使用

### 建议

1. **逐步采用**: 不要一次性更新所有代码，逐步迁移
2. **测试验证**: 每次迁移后运行完整测试
3. **文档更新**: 更新相关文档和注释
4. **团队同步**: 确保团队成员了解变更

## 📞 支持

如果在迁移过程中遇到问题：

1. 查看pkg目录下各组件的README文档
2. 参考`pkg/examples/`中的使用示例
3. 运行项目测试确保功能正常

---

**迁移完成日期**: 2025年8月27日  
**影响范围**: 中间件和密码加密组件  
**兼容性**: 向后兼容，建议逐步升级到新API
**删除操作完成**: ✅ 已成功删除internal/crypto和internal/middleware目录  
**项目编译状态**: ✅ 主服务器编译通过，组件迁移成功