# 密码管理脚本

本目录包含用于管理管理员密码的脚本工具。

## 目录结构

```
scripts/
├── generate_password/          # 密码生成脚本
│   ├── generate_admin_password.go
│   └── password_utils.go
├── verify_password/           # 密码验证和更新脚本
│   ├── verify_and_update_admin_password.go
│   └── password_utils.go
└── README.md                  # 本文件
```

## 脚本说明

### 1. 生成管理员密码哈希

位置：`generate_password/generate_admin_password.go`

**功能**：生成符合系统安全要求的管理员密码哈希值，使用 HMAC-SHA256 + bcrypt 双重加密。

**使用方法**：
```bash
cd generate_password
go run *.go <password> <encryption_key>
```

**示例**：
```bash
cd generate_password
go run *.go admin123 your-ultra-secret-password-encryption-key-change-this
```

**输出**：
- 原始密码
- 加密密钥
- HMAC 签名
- 最终哈希密码
- 用于迁移文件的 SQL 格式

### 2. 验证和更新管理员密码

位置：`verify_password/verify_and_update_admin_password.go`

**功能**：直接连接数据库，更新管理员密码并验证更新结果。

**使用方法**：
```bash
cd verify_password
go run *.go <password> [encryption_key]
```

**示例**：
```bash
cd verify_password
go run *.go admin123 your-ultra-secret-password-encryption-key-change-this
```

**注意事项**：
- 需要确保数据库服务正在运行
- 默认连接到本地 PostgreSQL 数据库
- 如果不提供加密密钥，将使用默认值

## 安全说明

1. **加密密钥**：请使用强加密密钥，不要使用示例中的默认值
2. **密码强度**：建议使用包含大小写字母、数字和特殊字符的强密码
3. **数据库连接**：确保数据库连接字符串中的凭据安全存储

## 依赖项

这些脚本需要以下 Go 模块：
- `golang.org/x/crypto/bcrypt`
- `github.com/lib/pq`（仅验证脚本需要）

运行前请确保已安装相关依赖：
```bash
go mod tidy
```