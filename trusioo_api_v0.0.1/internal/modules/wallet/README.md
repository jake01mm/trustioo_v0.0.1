# 钱包模块 (Wallet Module)

## 概述
钱包模块是Trusioo平台的核心功能模块，负责处理用户钱包、货币汇率、银行账户管理和提现等业务逻辑。

## 功能特性

### 1. 钱包管理
- ✅ 获取用户钱包信息（余额、状态等）
- ✅ 设置交易密码
- ✅ 修改交易密码
- 📝 钱包余额调整（管理员功能）

### 2. 货币和汇率
- ✅ 获取支持的货币列表
- ✅ 查询货币汇率信息
- 📝 汇率管理（管理员功能）

### 3. 银行管理
- ✅ 获取支持的银行列表（按国家筛选）
- ✅ 用户银行账户管理
  - 添加银行账户
  - 更新银行账户信息
  - 删除银行账户
  - 查询用户所有银行账户

### 4. 提现功能
- 📝 提现申请
- 📝 提现审核（管理员功能）
- 📝 提现处理（管理员功能）
- 📝 提现记录查询

### 5. 交易记录
- 📝 交易记录查询
- 📝 交易统计分析

## API 端点

### 公开接口（无需认证）
- `GET /api/v1/wallet/currencies` - 获取货币列表
- `GET /api/v1/wallet/exchange-rate` - 获取汇率
- `GET /api/v1/wallet/banks` - 获取银行列表

### 用户接口（需要用户认证）
- `GET /api/v1/wallet` - 获取钱包信息
- `POST /api/v1/wallet/transaction-pin` - 设置交易密码
- `PUT /api/v1/wallet/transaction-pin` - 修改交易密码
- `GET /api/v1/wallet/bank-accounts` - 获取银行账户列表
- `POST /api/v1/wallet/bank-accounts` - 添加银行账户
- `PUT /api/v1/wallet/bank-accounts/:id` - 更新银行账户
- `DELETE /api/v1/wallet/bank-accounts/:id` - 删除银行账户

### 管理员接口（需要管理员认证）
- 📝 提现审核和管理相关接口
- 📝 汇率管理接口
- 📝 钱包统计和监控接口

## 数据库表结构

模块包含以下7个数据表：

1. **currencies** - 货币表
2. **exchange_rates** - 汇率表
3. **wallets** - 钱包表
4. **banks** - 银行表
5. **user_bank_accounts** - 用户银行账户表
6. **wallet_transactions** - 钱包交易记录表
7. **withdrawal_requests** - 提现申请表

## 文件结构

```
internal/modules/wallet/
├── README.md           # 本文档
├── model.go           # 数据模型和枚举定义
├── dto.go             # API请求/响应结构体
├── repository.go      # 数据访问层
├── service.go         # 业务逻辑层
├── handler.go         # HTTP处理器
└── routes.go          # 路由定义
```

## 状态说明

- ✅ 已完成
- 📝 待实现
- ⚠️ 需要优化

## 使用示例

### 获取货币列表
```bash
curl -X GET "http://localhost:8080/api/v1/wallet/currencies"
```

### 查询汇率
```bash
curl -X GET "http://localhost:8080/api/v1/wallet/exchange-rate?from=TRU&to=NGN"
```

### 获取银行列表
```bash
curl -X GET "http://localhost:8080/api/v1/wallet/banks?country_code=NGA"
```

## 注意事项

1. 所有涉及金额的操作都需要验证交易密码
2. 银行账户信息需要经过验证才能用于提现
3. 汇率信息具有时效性，需要定期更新
4. 提现操作需要管理员审核
5. 所有金额计算使用高精度数值，避免浮点数精度问题

## 开发规范

- 遵循三层架构：Repository -> Service -> Handler
- 使用依赖注入进行组件解耦
- 统一错误处理和日志记录
- API响应格式标准化
- 数据库操作使用事务确保一致性