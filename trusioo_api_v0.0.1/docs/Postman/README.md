# Trusioo API - Postman集合使用指南

## 概述

本文档详细说明了Trusioo API Postman集合的结构、使用方法以及与实际API代码的匹配验证。

## 📁 集合结构

### 1. 集合信息
- **名称**: Trusioo API - 完整认证集合
- **版本**: 1.0.0
- **文件**: `Trusioo_API_Complete_Collection.json`

### 2. 环境变量
```json
{
  "base_url": "http://localhost:8080",
  "access_token": "",
  "refresh_token": "",
  "user_id": "",
  "admin_id": "", 
  "buyer_id": ""
}
```

### 3. 模块结构

#### 🏥 01 健康检查 (Health Check)
- **整体健康检查**: `GET /health`
- **数据库健康检查**: `GET /health/database`
- **Redis健康检查**: `GET /health/redis`
- **API v1健康检查**: `GET /health/api/v1`
- **就绪状态检查**: `GET /health/readiness`
- **存活状态检查**: `GET /health/liveness`

#### 👤 02 用户认证 (User Auth)
- **用户注册（简化版）**: `POST /api/v1/auth/user/register`
- **用户登录（发送验证码）**: `POST /api/v1/auth/user/login`
- **验证登录并获取令牌**: `POST /api/v1/auth/user/verify-login`
- **获取用户资料**: `GET /api/v1/auth/user/profile` 🔒
- **用户登出**: `POST /api/v1/auth/user/logout` 🔒

#### 👨‍💼 03 管理员认证 (Admin Auth)
- **管理员登录（发送验证码）**: `POST /api/v1/auth/admin/login`
- **验证管理员登录**: `POST /api/v1/auth/admin/verify-login`
- **刷新管理员令牌**: `POST /api/v1/auth/admin/refresh` 🔒
- **获取管理员资料**: `GET /api/v1/auth/admin/profile` 🔒
- **修改管理员密码**: `PUT /api/v1/auth/admin/password` 🔒
- **管理员登出**: `POST /api/v1/auth/admin/logout` 🔒

#### 🏢 04 买家认证 (Buyer Auth)
- **买家注册**: `POST /api/v1/auth/buyer/register`
- **买家登录**: `POST /api/v1/auth/buyer/login`
- **获取买家资料**: `GET /api/v1/auth/buyer/profile` 🔒
- **买家登出**: `POST /api/v1/auth/buyer/logout` 🔒

> 🔒 表示需要Bearer Token认证

## 🚀 快速开始

### 1. 导入集合
1. 打开Postman
2. 点击 "Import"
3. 选择 `Trusioo_API_Complete_Collection.json` 文件
4. 导入完成

### 2. 设置环境
1. 创建新环境（例如：Trusioo Development）
2. 设置环境变量：
   ```
   base_url: http://localhost:8080
   ```
3. 其他变量会在API调用过程中自动设置

### 3. 测试流程

#### 用户认证流程
1. **注册用户** → 自动设置 `user_id`
2. **用户登录** → 获取验证码（开发环境会在响应中返回）
3. **验证登录** → 自动设置 `access_token` 和 `refresh_token`
4. **获取资料** → 使用自动设置的token
5. **登出** → 清除token

#### 管理员认证流程
1. **管理员登录** → 获取验证码
2. **验证登录** → 自动设置 `admin_access_token`
3. **使用其他API** → 自动使用管理员token
4. **刷新令牌** → 获取新token
5. **登出** → 清除token

#### 买家认证流程
1. **买家注册** → 自动设置token（如果直接注册成功）
2. **买家登录** → 自动设置 `buyer_access_token`
3. **使用其他API** → 自动使用买家token
4. **登出** → 清除token

## 📋 请求体示例

### 用户注册（简化版）
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

### 用户登录
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

### 验证登录
```json
{
  "email": "user@example.com",
  "password": "password123",
  "login_code": "123456",
  "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
}
```

### 买家注册
```json
{
  "email": "buyer@company.com",
  "password": "buyer123456",
  "company_name": "ABC Company Ltd",
  "contact_name": "John Smith",
  "phone": "+1234567890"
}
```

### 修改管理员密码
```json
{
  "current_password": "admin123456",
  "new_password": "newadmin123456"
}
```

## 🔧 自动化功能

### 1. 自动令牌提取
集合包含自动化脚本，会在登录成功后自动提取并保存令牌：
- 用户令牌 → `access_token`, `refresh_token`
- 管理员令牌 → `admin_access_token`, `admin_refresh_token`
- 买家令牌 → `buyer_access_token`, `buyer_refresh_token`

### 2. 自动用户ID提取
注册和登录成功后自动提取用户ID：
- `user_id`
- `admin_id`
- `buyer_id`

### 3. 自动验证码提取
登录接口会自动提取验证码（仅开发环境）：
- `verification_code`
- `admin_verification_code`

## 🔍 API结构验证

### 验证脚本
运行验证脚本检查API结构：
```bash
chmod +x validate_api.sh
./validate_api.sh
```

### 验证结果
- ✅ **路由路径**: 与代码中路由定义完全匹配
- ✅ **HTTP方法**: 与handler方法定义匹配
- ✅ **请求体结构**: 基于DTO定义
- ✅ **响应结构**: 基于DTO定义
- ✅ **认证方式**: Bearer Token认证
- ✅ **环境变量**: 合理的变量配置

## 🎯 匹配验证详情

### 代码映射关系

#### 路由定义映射
```go
// internal/modules/auth/user/routes.go
user.POST("/register", r.handler.Register)        
user.POST("/login", r.handler.Login)              
user.POST("/verify-login", r.handler.VerifyLogin) 

// internal/modules/auth/admin/routes.go
admin.POST("/login", r.handler.Login)
admin.POST("/verify-login", r.handler.VerifyLogin)
admin.POST("/refresh", r.handler.RefreshToken)
admin.GET("/profile", r.handler.GetProfile)
admin.PUT("/password", r.handler.ChangePassword)

// internal/modules/auth/buyer/routes.go
buyer.POST("/register", r.handler.Register)
buyer.POST("/login", r.handler.Login)
buyer.GET("/profile", r.handler.GetProfile)
```

#### DTO结构映射
```go
// 用户注册请求
type SimpleRegisterRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

// 登录请求
type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

// 验证登录请求
type VerifyLoginRequest struct {
    Email     string `json:"email" binding:"required,email"`
    Password  string `json:"password" binding:"required,min=6"`
    LoginCode string `json:"login_code" binding:"required,len=6"`
    UserAgent string `json:"user_agent" binding:"omitempty"`
}
```

### JWT令牌结构
```go
// internal/modules/auth/jwt.go
type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    TokenType    string `json:"token_type"`
    ExpiresIn    int64  `json:"expires_in"`
}
```

## 🛠️ 故障排除

### 常见问题

1. **连接失败**
   - 检查API服务器是否运行
   - 确认 `base_url` 设置正确
   - 检查网络连接

2. **认证失败**
   - 确保令牌已正确设置
   - 检查令牌是否过期
   - 重新登录获取新令牌

3. **验证码问题**
   - 开发环境验证码在响应中返回
   - 生产环境需要查看邮件
   - 验证码有效期为5分钟

### 调试技巧
1. 启用Postman控制台查看详细日志
2. 检查环境变量是否正确设置
3. 查看响应头和响应体获取错误信息

## 📊 测试建议

### 测试顺序
1. 先测试健康检查端点
2. 测试用户注册和登录流程
3. 测试管理员认证流程
4. 测试买家认证流程
5. 测试需要认证的端点

### 数据准备
- 为不同角色准备测试账号
- 准备有效和无效的测试数据
- 测试各种错误场景

## 📈 扩展功能

### 添加新端点
1. 在对应模块下添加新请求
2. 设置正确的HTTP方法和URL
3. 配置请求体和认证
4. 添加响应处理脚本

### 环境管理
- Development: 本地开发环境
- Testing: 测试环境
- Staging: 预发布环境
- Production: 生产环境

---

## 📞 支持

如有问题，请检查：
1. API服务器日志
2. Postman控制台日志
3. 网络连接状态
4. 环境变量配置

**集合创建时间**: $(date)
**API版本**: v1.0.0
**文档版本**: 1.0.0