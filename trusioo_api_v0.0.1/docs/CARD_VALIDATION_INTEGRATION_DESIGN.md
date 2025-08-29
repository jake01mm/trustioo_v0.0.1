# 卡片验证API集成开发说明文档

## 1. 项目概述

### 1.1 功能需求
- 集成第三方卡片验证API（支持iTunes、Razer、Nike等多种卡片类型）
- 实现基于TRU平台币的收费机制
- 支持用户付费使用和管理员免费使用
- 提供完整的卡片验证服务管理功能

### 1.2 技术架构
- 后端：Go语言，基于现有项目架构
- 数据库：PostgreSQL
- 加密：DES加密（根据第三方API文档要求）
- 签名：MD5签名算法

## 2. 数据库设计

### 2.1 卡片类型配置表 (card_types)
```sql
CREATE TABLE card_types (
    id BIGSERIAL PRIMARY KEY,
    product_mark VARCHAR(50) NOT NULL UNIQUE, -- 产品标识：iTunes, Razer, nike等
    name VARCHAR(100) NOT NULL,               -- 显示名称：苹果卡、雷蛇卡等
    description TEXT,                         -- 描述信息
    price_tru DECIMAL(10,2) NOT NULL,        -- TRU币价格
    is_active BOOLEAN DEFAULT true,           -- 是否启用
    requires_pin BOOLEAN DEFAULT false,      -- 是否需要PIN码
    requires_region BOOLEAN DEFAULT false,   -- 是否需要区域信息
    card_format VARCHAR(200),                 -- 卡号格式说明
    supported_regions JSONB,                 -- 支持的区域列表
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 初始化数据（严格按照API文档和用户需求）
INSERT INTO card_types (product_mark, name, price_tru, requires_pin, requires_region, card_format, supported_regions) VALUES
-- 用户指定收费：iTunes 1 TRU, Razer 2 TRU, Nike 5 TRU
('iTunes', '苹果卡', 1.00, false, true, '卡号格式不限', 
 '[{"id":1,"regionName":"英国"},{"id":2,"regionName":"美国"},{"id":3,"regionName":"德国"},{"id":4,"regionName":"澳大利亚"},{"id":5,"regionName":"加拿大"},{"id":6,"regionName":"日本"},{"id":8,"regionName":"西班牙"},{"id":9,"regionName":"意大利"},{"id":10,"regionName":"法国"},{"id":11,"regionName":"爱尔兰"},{"id":12,"regionName":"墨西哥"}]'),
('Razer', '雷蛇卡', 2.00, false, true, '卡号格式不限', 
 '[{"regionId":12,"chName":"美国"},{"regionId":6,"chName":"澳大利亚"},{"regionId":13,"chName":"巴西"},{"regionId":26,"chName":"柬埔寨"},{"regionId":20,"chName":"加拿大"},{"regionId":25,"chName":"智利"},{"regionId":22,"chName":"哥伦比亚"},{"regionId":17,"chName":"香港特别行政区"},{"regionId":4,"chName":"印度"},{"regionId":7,"chName":"印度尼西亚"},{"regionId":27,"chName":"日本"},{"regionId":1,"chName":"马来西亚"},{"regionId":19,"chName":"缅甸"},{"regionId":15,"chName":"新西兰"},{"regionId":29,"chName":"巴基斯坦"},{"regionId":8,"chName":"菲律宾"},{"regionId":5,"chName":"新加坡"},{"regionId":18,"chName":"土耳其"},{"regionId":33,"chName":"越南"},{"regionId":2,"chName":"其他"},{"regionId":28,"chName":"其他（中文）"},{"regionId":21,"chName":"墨西哥"}]'),
('nike', 'Nike卡', 5.00, true, false, '卡号固定19位，PIN码固定6位，格式：{codeNo}-{pinCode}', '[]'),
-- 其他卡片类型价格由管理员控制，设置默认值
('sephora', '丝芙兰卡', 3.00, true, false, '卡号16位数字，PIN码8位数字，格式：{codeNo}-{pinCode}', '[]'),
('amazon', '亚马逊卡', 2.50, false, true, '电子码14位卡号，实体卡15位卡号', 
 '[{"regionId":2,"regionName":"美亚/加亚"},{"regionId":1,"regionName":"欧盟区","supportedCountries":"英国、德国、荷兰、西班牙、法国、奥地利、丹麦、芬兰、希腊、意大利、波兰、葡萄牙、瑞典"}]'),
('xBox', 'Xbox卡', 3.50, false, true, '25个字符', 
 '["美国","加拿大","英国","澳大利亚","新西兰","新加坡","韩国","墨西哥","瑞典","哥伦比亚","阿根廷","尼日利亚","香港特别行政区","挪威","波兰","德国"]'),
('nd', 'ND卡', 4.00, true, false, '卡号固定16位，PIN码固定8位，格式：{codeNo}-{pinCode}', '[]');
```

### 2.2 卡片验证记录表 (card_validations)
```sql
CREATE TABLE card_validations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,                          -- 用户ID，NULL表示管理员操作
    admin_id BIGINT,                         -- 管理员ID，NULL表示用户操作
    card_type_id BIGINT NOT NULL REFERENCES card_types(id),
    card_number VARCHAR(500) NOT NULL,       -- 卡号（可能包含PIN）
    pin_code VARCHAR(100),                   -- PIN码（单独存储）
    region_id INTEGER,                       -- 区域ID
    region_name VARCHAR(100),                -- 区域名称
    auto_type INTEGER DEFAULT 0,             -- 自动识别类型（仅苹果卡）
    
    -- 验证结果
    status INTEGER DEFAULT 0,                -- 0:等待检测 1:测卡中 2:有效 3:无效 4:已兑换 5:检测失败 6:点数不足
    result_message TEXT,                     -- 检测结果信息
    check_time TIMESTAMP,                    -- 检测完成时间
    
    -- 查询控制（成本优化核心）
    last_query_time TIMESTAMP,              -- 上次查询第三方API时间
    query_count INTEGER DEFAULT 0,          -- 查询次数统计
    can_query_again BOOLEAN DEFAULT TRUE,   -- 是否允许再次查询
    is_final_status BOOLEAN DEFAULT FALSE,  -- 是否为最终状态（2/3/4/5/6）
    
    -- 费用相关
    cost_tru DECIMAL(10,2) DEFAULT 0,       -- 消耗的TRU币数量
    is_free BOOLEAN DEFAULT false,           -- 是否免费（管理员使用）
    
    -- 第三方API相关
    third_party_request_id VARCHAR(200),     -- 第三方请求ID（如果有）
    third_party_response JSONB,             -- 第三方完整响应
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_card_validations_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_card_validations_admin FOREIGN KEY (admin_id) REFERENCES admins(id)
);

-- 创建索引
CREATE INDEX idx_card_validations_user_id ON card_validations(user_id);
CREATE INDEX idx_card_validations_admin_id ON card_validations(admin_id);
CREATE INDEX idx_card_validations_status ON card_validations(status);
CREATE INDEX idx_card_validations_created_at ON card_validations(created_at);
CREATE INDEX idx_card_validations_last_query_time ON card_validations(last_query_time);
CREATE INDEX idx_card_validations_final_status ON card_validations(is_final_status);
```

### 2.3 卡片验证查询日志表 (card_validation_logs)
```sql
CREATE TABLE card_validation_logs (
    id BIGSERIAL PRIMARY KEY,
    validation_id BIGINT NOT NULL REFERENCES card_validations(id),
    api_type VARCHAR(20) NOT NULL,           -- 'checkCard' 或 'checkCardResult'
    request_data JSONB,                      -- 请求数据（加密前）
    response_data JSONB,                     -- 响应数据
    api_cost DECIMAL(10,4) DEFAULT 0,       -- 单次API调用成本
    response_time_ms INTEGER,                -- 响应时间（毫秒）
    is_success BOOLEAN DEFAULT false,       -- 是否成功
    error_message TEXT,                      -- 错误信息
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_validation_logs_validation_id ON card_validation_logs(validation_id);
CREATE INDEX idx_validation_logs_api_type ON card_validation_logs(api_type);
CREATE INDEX idx_validation_logs_created_at ON card_validation_logs(created_at);
```

### 2.4 卡片验证缓存表 (card_validation_cache)
```sql
CREATE TABLE card_validation_cache (
    id BIGSERIAL PRIMARY KEY,
    card_hash VARCHAR(64) NOT NULL UNIQUE,   -- 卡号+PIN的MD5哈希
    card_type_id BIGINT NOT NULL REFERENCES card_types(id),
    region_id INTEGER,
    status INTEGER NOT NULL,                 -- 最终状态
    result_message TEXT,                     -- 结果信息
    check_time TIMESTAMP NOT NULL,          -- 检测时间
    cache_expires_at TIMESTAMP,             -- 缓存过期时间（可选）
    hit_count INTEGER DEFAULT 0,            -- 缓存命中次数
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_validation_cache_hash ON card_validation_cache(card_hash);
CREATE INDEX idx_validation_cache_expires ON card_validation_cache(cache_expires_at);
```

### 2.5 第三方API配置表 (third_party_configs)
```sql
CREATE TABLE third_party_configs (
    id BIGSERIAL PRIMARY KEY,
    config_key VARCHAR(100) NOT NULL UNIQUE, -- 配置键
    config_value TEXT NOT NULL,              -- 配置值
    description TEXT,                        -- 描述
    is_encrypted BOOLEAN DEFAULT false,     -- 是否加密存储
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 初始化配置
INSERT INTO third_party_configs (config_key, config_value, description, is_encrypted) VALUES
('CARD_VALIDATOR_APP_ID', '2508042205539611639', '卡片验证器应用ID', false),
('CARD_VALIDATOR_APP_SECRET', '2caa437312d44edcaf3ab61910cf31b7', '卡片验证器应用密钥', true),
('CARD_VALIDATOR_BASE_URL', 'https://ckxiang.com', '第三方API基础URL', false),
('CARD_VALIDATOR_TIMEOUT', '30', 'API请求超时时间（秒）', false);
```

## 3. API接口设计

### 3.1 用户端接口 (User API)

**基础路径：** `/api/v1/user`  
**认证方式：** Bearer Token (用户JWT)  
**权限控制：** 用户身份验证 + TRU币余额检查

#### 3.1.1 获取可用卡片类型列表
```http
GET /api/v1/user/card-types
Authorization: Bearer {user_token}
```

**功能说明：** 返回用户可使用的卡片类型（仅显示启用的类型）

**响应示例：**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": [
    {
      "id": 1,
      "product_mark": "iTunes",
      "name": "苹果卡",
      "description": "苹果iTunes礼品卡验证",
      "price_tru": 1.00,
      "requires_pin": false,
      "requires_region": true,
      "card_format": "卡号格式不限",
      "supported_regions": [
        {"id": 1, "regionName": "英国"},
        {"id": 2, "regionName": "美国"}
      ]
    }
  ]
}
```

#### 3.1.2 提交卡片验证请求
```http
POST /api/v1/user/card-validations
Authorization: Bearer {user_token}
Content-Type: application/json
```

**请求参数：**
```json
{
  "card_type_id": 1,
  "card_number": "X123123123123123",
  "pin_code": "12345678",  // 可选，根据卡片类型要求
  "region_id": 2,          // 可选，根据卡片类型要求
  "region_name": "美国",    // 可选
  "auto_type": 0           // 可选，仅苹果卡使用
}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "验证请求已提交",
  "data": {
    "validation_id": 12345,
    "status": 1,
    "status_text": "测卡中",
    "cost_tru": 1.0,
    "remaining_balance": 98.5,
    "can_query_now": false,
    "next_query_time": "2024-01-01T10:05:00Z",
    "estimated_time": "请稍后点击查询按钮获取最新状态"
  }
}
```

#### 3.1.3 查询单个验证结果
```http
GET /api/v1/user/card-validations/{validation_id}
Authorization: Bearer {user_token}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "查询成功",
  "data": {
    "id": 12345,
    "card_type_name": "苹果卡",
    "card_number_masked": "X123****3456",
    "status": 2,
    "status_text": "有效",
    "result_message": "卡片验证成功，余额充足",
    "cost_tru": 1.0,
    "query_count": 2,
    "can_query_again": false,
    "is_final_status": true,
    "created_at": "2024-01-01T10:00:00Z",
    "check_time": "2024-01-01T10:03:30Z",
    "last_query_time": "2024-01-01T10:03:30Z"
  }
}
```

#### 3.1.4 主动查询验证结果（用户触发）
```http
POST /api/v1/user/card-validations/{validation_id}/query
Authorization: Bearer {user_token}
```

**功能说明：** 用户主动触发查询第三方API获取最新状态

**响应示例（查询成功）：**
```json
{
  "code": 200,
  "message": "查询成功",
  "data": {
    "validation_id": 12345,
    "status": 2,
    "status_text": "有效",
    "result_message": "卡片验证成功，余额充足",
    "check_time": "2024-01-01T10:03:30Z",
    "is_final": true,
    "query_count": 2,
    "last_query_time": "2024-01-01T10:03:30Z"
  }
}
```

**响应示例（查询限制）：**
```json
{
  "code": 429,
  "message": "查询过于频繁",
  "data": {
    "validation_id": 12345,
    "status": 1,
    "status_text": "测卡中",
    "can_query_again": false,
    "next_query_time": "2024-01-01T10:08:00Z",
    "last_query_time": "2024-01-01T10:03:00Z",
    "reason": "同一卡号5分钟内只能查询一次"
  }
}
```

#### 3.1.5 获取验证历史记录
```http
GET /api/v1/user/card-validations?page=1&limit=20&status=2&card_type_id=1
Authorization: Bearer {user_token}
```

**查询参数：**
- `page`: 页码（默认1）
- `limit`: 每页数量（默认20，最大100）
- `status`: 状态筛选（可选）
- `card_type_id`: 卡片类型筛选（可选）
- `start_date`: 开始日期（可选）
- `end_date`: 结束日期（可选）

**响应示例：**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "total": 50,
    "page": 1,
    "limit": 20,
    "total_pages": 3,
    "items": [
      {
        "id": 12345,
        "card_type_name": "苹果卡",
        "card_number_masked": "X123****3456",
        "status": 2,
        "status_text": "有效",
        "cost_tru": 1.0,
        "query_count": 2,
        "is_final_status": true,
        "created_at": "2024-01-01T10:00:00Z",
        "check_time": "2024-01-01T10:03:30Z"
      }
    ]
  }
}
```

#### 3.1.6 获取用户TRU币余额
```http
GET /api/v1/user/balance
Authorization: Bearer {user_token}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "balance": 98.5,
    "currency": "TRU",
    "last_updated": "2024-01-01T10:00:00Z"
  }
}
```

### 3.2 管理员接口 (Admin API)

**基础路径：** `/api/v1/admin`  
**认证方式：** Bearer Token (管理员JWT)  
**权限控制：** 管理员身份验证 + 角色权限检查

#### 3.2.1 管理员卡片验证（免费）
```http
POST /api/v1/admin/card-validations
Authorization: Bearer {admin_token}
Content-Type: application/json
```

**请求参数：**
```json
{
  "card_type_id": 1,
  "card_number": "X123123123123123",
  "pin_code": "12345678",  // 可选
  "region_id": 2,          // 可选
  "region_name": "美国",    // 可选
  "auto_type": 0,          // 可选
  "note": "管理员测试验证"   // 可选，备注信息
}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "验证请求已提交",
  "data": {
    "validation_id": 12346,
    "status": 1,
    "status_text": "测卡中",
    "cost_tru": 0.0,
    "is_free": true,
    "admin_id": 1,
    "can_query_now": true
  }
}
```

#### 3.2.2 管理员查询验证结果
```http
POST /api/v1/admin/card-validations/{validation_id}/query
Authorization: Bearer {admin_token}
```

**功能说明：** 管理员可随时查询，无频率限制

#### 3.2.3 卡片类型管理

##### 获取所有卡片类型
```http
GET /api/v1/admin/card-types
Authorization: Bearer {admin_token}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": [
    {
      "id": 1,
      "product_mark": "iTunes",
      "name": "苹果卡",
      "description": "苹果iTunes礼品卡验证",
      "price_tru": 1.00,
      "is_active": true,
      "requires_pin": false,
      "requires_region": true,
      "card_format": "卡号格式不限",
      "supported_regions": [...],
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

##### 创建卡片类型
```http
POST /api/v1/admin/card-types
Authorization: Bearer {admin_token}
Content-Type: application/json
```

**请求参数：**
```json
{
  "product_mark": "steam",
  "name": "Steam卡",
  "description": "Steam游戏平台礼品卡验证",
  "price_tru": 2.50,
  "is_active": true,
  "requires_pin": true,
  "requires_region": false,
  "card_format": "卡号15位，PIN码4位",
  "supported_regions": []
}
```

##### 更新卡片类型
```http
PUT /api/v1/admin/card-types/{id}
Authorization: Bearer {admin_token}
Content-Type: application/json
```

##### 删除卡片类型
```http
DELETE /api/v1/admin/card-types/{id}
Authorization: Bearer {admin_token}
```

#### 3.2.4 验证记录管理

##### 获取所有验证记录
```http
GET /api/v1/admin/card-validations?page=1&limit=50&user_id=123&status=2
Authorization: Bearer {admin_token}
```

**查询参数：**
- `page`: 页码
- `limit`: 每页数量
- `user_id`: 用户ID筛选
- `admin_id`: 管理员ID筛选
- `status`: 状态筛选
- `card_type_id`: 卡片类型筛选
- `start_date`: 开始日期
- `end_date`: 结束日期
- `is_free`: 是否免费筛选

**响应示例：**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "total": 1000,
    "page": 1,
    "limit": 50,
    "total_pages": 20,
    "items": [
      {
        "id": 12345,
        "user_id": 123,
        "admin_id": null,
        "user_email": "user@example.com",
        "card_type_name": "苹果卡",
        "card_number_masked": "X123****3456",
        "status": 2,
        "status_text": "有效",
        "cost_tru": 1.0,
        "is_free": false,
        "query_count": 2,
        "created_at": "2024-01-01T10:00:00Z",
        "check_time": "2024-01-01T10:03:30Z"
      }
    ]
  }
}
```

##### 获取单个验证记录详情
```http
GET /api/v1/admin/card-validations/{id}
Authorization: Bearer {admin_token}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "id": 12345,
    "user_id": 123,
    "user_email": "user@example.com",
    "card_type_name": "苹果卡",
    "card_number": "X123123123123456",  // 管理员可看完整卡号
    "pin_code": "12345678",
    "region_name": "美国",
    "status": 2,
    "status_text": "有效",
    "result_message": "卡片验证成功，余额充足",
    "cost_tru": 1.0,
    "is_free": false,
    "query_count": 2,
    "third_party_response": {...},  // 第三方完整响应
    "created_at": "2024-01-01T10:00:00Z",
    "check_time": "2024-01-01T10:03:30Z",
    "last_query_time": "2024-01-01T10:03:30Z"
  }
}
```

#### 3.2.5 第三方API配置管理

##### 获取所有配置
```http
GET /api/v1/admin/third-party-configs
Authorization: Bearer {admin_token}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": [
    {
      "id": 1,
      "config_key": "CARD_VALIDATOR_APP_ID",
      "config_value": "2508042205539611639",
      "description": "卡片验证器应用ID",
      "is_encrypted": false,
      "updated_at": "2024-01-01T00:00:00Z"
    },
    {
      "id": 2,
      "config_key": "CARD_VALIDATOR_APP_SECRET",
      "config_value": "****",  // 加密字段显示为****
      "description": "卡片验证器应用密钥",
      "is_encrypted": true,
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

##### 更新配置
```http
PUT /api/v1/admin/third-party-configs/{config_key}
Authorization: Bearer {admin_token}
Content-Type: application/json
```

**请求参数：**
```json
{
  "config_value": "new_value",
  "description": "更新后的描述"
}
```

#### 3.2.6 统计数据接口

##### 获取验证统计
```http
GET /api/v1/admin/statistics/validations?period=7d
Authorization: Bearer {admin_token}
```

**查询参数：**
- `period`: 统计周期（1d, 7d, 30d, 90d）
- `start_date`: 自定义开始日期
- `end_date`: 自定义结束日期

**响应示例：**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "total_validations": 1000,
    "successful_validations": 800,
    "failed_validations": 200,
    "total_revenue_tru": 1500.0,
    "total_api_calls": 2500,
    "average_query_count": 2.5,
    "by_card_type": [
      {
        "card_type": "苹果卡",
        "count": 500,
        "revenue_tru": 500.0
      }
    ],
    "by_status": [
      {
        "status": 2,
        "status_text": "有效",
        "count": 600
      }
    ]
  }
}
```

## 4. 业务流程设计

### 4.1 用户验证流程（优化版 - 用户触发查询）

#### 4.1.1 获取卡片类型
1. 用户访问 `GET /api/v1/user/card-types`
2. 系统返回启用的卡片类型列表
3. 用户选择需要验证的卡片类型

#### 4.1.2 提交验证请求
1. 用户填写卡片信息（卡号、PIN码、区域等）
2. 前端调用 `POST /api/v1/user/card-validations`
3. 系统验证用户身份和TRU币余额
4. 系统扣除相应TRU币
5. 调用第三方API提交测卡请求（状态：0→1）
6. 返回验证ID和"测卡中"状态
7. 记录查询控制信息（last_query_time等）

#### 4.1.3 查询验证结果
1. **用户主动查询**：用户点击"查询最新状态"按钮
2. 前端调用 `POST /api/v1/user/card-validations/{id}/query`
3. 系统检查查询限制：
   - 是否为最终状态（直接返回缓存结果）
   - 是否在5分钟查询限制内
4. 如果可以查询，调用第三方API获取最新状态
5. 更新数据库状态和查询时间
6. 如果是最终状态，标记为final避免重复查询
7. 返回最新状态给用户

#### 4.1.4 查看历史记录
1. 用户访问 `GET /api/v1/user/card-validations`
2. 系统返回用户的验证历史记录
3. 支持按状态、卡片类型、日期等筛选

### 4.2 管理员验证流程

#### 4.2.1 管理员免费验证
1. 管理员登录后台系统
2. 访问 `POST /api/v1/admin/card-validations`
3. 填写卡片信息（无需扣费）
4. 系统直接调用第三方API（免费）
5. 管理员可随时查询结果（无频率限制）

#### 4.2.2 卡片类型管理
1. 管理员通过 `GET /api/v1/admin/card-types` 查看所有卡片类型
2. 可以创建、更新、删除卡片类型
3. 设置价格、区域支持、格式要求等

#### 4.2.3 验证记录管理
1. 管理员通过 `GET /api/v1/admin/card-validations` 查看所有验证记录
2. 可以查看用户完整卡号信息
3. 支持多维度筛选和统计

#### 4.2.4 系统配置管理
1. 管理员通过 `GET /api/v1/admin/third-party-configs` 管理第三方API配置
2. 可以更新API密钥、超时时间等配置
3. 敏感信息加密存储

### 4.3 第三方API调用流程

#### 4.3.1 提交测卡请求流程
1. **构建请求参数**：
   - cards数组（包含卡号、PIN码、区域等）
   - timestamp（当前时间戳）
   - 其他业务参数

2. **生成签名**：
   - 按参数名排序
   - 拼接参数值
   - 使用MD5生成签名

3. **加密请求**：
   - 使用DES加密整个请求数据
   - 设置正确的Content-Type

4. **调用API**：
   - POST到checkCard接口
   - 处理响应结果
   - 更新数据库状态为"测卡中"(1)

#### 4.3.2 查询测卡结果流程（用户/管理员触发）
1. **查询控制检查**：
   - 检查是否为最终状态（用户端）
   - 检查查询频率限制（用户端）
   - 管理员无限制

2. **构建查询参数**：
   - cardNo（卡号）
   - pinCode（PIN码，如果有）
   - timestamp（时间戳）

3. **生成签名和加密**：
   - 同提交请求流程

4. **调用API**：
   - POST到checkCardResult接口
   - 解密响应数据
   - 解析状态和结果信息

5. **更新数据库**：
   - 更新验证状态和结果
   - 更新查询时间和次数
   - 如果是最终状态，标记为final
   - 记录API调用日志

### 4.4 权限控制流程

#### 4.4.1 用户端权限验证
1. 验证JWT Token有效性
2. 检查用户状态（是否激活、是否被禁用）
3. 验证TRU币余额是否充足
4. 检查操作权限（只能操作自己的验证记录）

#### 4.4.2 管理员端权限验证
1. 验证管理员JWT Token有效性
2. 检查管理员角色和权限
3. 记录管理员操作日志
4. 允许查看所有用户数据（脱敏处理可选）

## 5. 成本优化策略

### 5.1 查询控制机制
1. **用户触发查询**：取消系统主动轮询，改为用户点击查询按钮时才调用API
2. **查询频率限制**：同一卡号5分钟内只能查询一次，防止频繁调用
3. **最终状态缓存**：状态为最终状态（2/3/4/5/6）的卡片不再查询API
4. **智能提示**：前端显示上次查询时间和下次可查询时间

### 5.2 成本节省效果
- **传统轮询模式**：每张卡平均调用10-20次API
- **优化后模式**：每张卡平均调用1-3次API
- **成本节省**：约80-90%的API调用成本

### 5.3 用户体验优化
1. **状态实时显示**：清晰显示当前状态和查询限制
2. **查询按钮状态**：根据查询限制动态启用/禁用
3. **批量查询**：支持多张卡片批量查询（高级功能）
4. **历史记录**：完整的查询历史和成本统计

## 6. 核心组件设计

### 6.1 第三方API客户端
- 签名生成器
- DES加密/解密器（严格按照第三方API文档实现）
- HTTP客户端封装
- 错误处理机制

### 6.2 用户端服务实现（Go语言）

#### 6.2.1 用户验证服务

```go
package service

import (
    "errors"
    "time"
    "gorm.io/gorm"
)

type UserCardValidationService struct {
    db                *gorm.DB
    thirdPartyService *ThirdPartyService
    queryControl      *QueryControlService
    truCoinService    *TruCoinService
}

func NewUserCardValidationService(db *gorm.DB, thirdParty *ThirdPartyService, queryControl *QueryControlService, truCoin *TruCoinService) *UserCardValidationService {
    return &UserCardValidationService{
        db:                db,
        thirdPartyService: thirdParty,
        queryControl:      queryControl,
        truCoinService:    truCoin,
    }
}

// SubmitValidation 用户提交验证请求
func (s *UserCardValidationService) SubmitValidation(userID uint, req *CardValidationRequest) (*CardValidation, error) {
    // 1. 检查用户TRU币余额
    balance, err := s.truCoinService.GetUserBalance(userID)
    if err != nil {
        return nil, err
    }
    
    cardType, err := s.getCardType(req.CardTypeID)
    if err != nil {
        return nil, err
    }
    
    if balance < cardType.Price {
        return nil, errors.New("TRU币余额不足")
    }
    
    // 2. 扣除TRU币
    if err := s.truCoinService.DeductBalance(userID, cardType.Price, "卡片验证"); err != nil {
        return nil, err
    }
    
    // 3. 创建验证记录
    validation := &CardValidation{
        UserID:       userID,
        CardTypeID:   req.CardTypeID,
        CardNumber:   req.CardNumber,
        PinCode:      req.PinCode,
        RegionID:     req.RegionID,
        RegionName:   req.RegionName,
        Status:       0, // 初始状态
        QueryCount:   0,
        IsFinalStatus: false,
        CostTru:      cardType.Price,
        CreatedAt:    time.Now(),
    }
    
    if err := s.db.Create(validation).Error; err != nil {
        // 回滚TRU币
        s.truCoinService.RefundBalance(userID, cardType.Price, "验证失败回滚")
        return nil, err
    }
    
    // 4. 调用第三方API提交测卡请求
    success, err := s.thirdPartyService.SubmitCardValidation(validation)
    if err != nil {
        // 更新状态为失败并回滚TRU币
        s.db.Model(validation).Updates(map[string]interface{}{
            "status": 5,
            "result_message": "提交失败: " + err.Error(),
            "is_final_status": true,
        })
        s.truCoinService.RefundBalance(userID, cardType.Price, "API调用失败回滚")
        return nil, err
    }
    
    if success {
        // 更新状态为测卡中
        now := time.Now()
        s.db.Model(validation).Updates(map[string]interface{}{
            "status": 1,
            "last_query_time": &now,
        })
    }
    
    return validation, nil
}

// QueryValidationResult 用户查询验证结果
func (s *UserCardValidationService) QueryValidationResult(userID, validationID uint) (*CardValidation, error) {
    // 1. 验证权限（用户只能查询自己的记录）
    var validation CardValidation
    if err := s.db.Where("id = ? AND user_id = ?", validationID, userID).First(&validation).Error; err != nil {
        return nil, errors.New("验证记录不存在或无权限访问")
    }
    
    // 2. 检查查询限制
    canQuery, message := s.queryControl.CanQuery(validationID)
    if !canQuery {
        return &validation, errors.New(message)
    }
    
    // 3. 调用第三方API查询结果
    result, err := s.thirdPartyService.QueryCardValidationResult(&validation)
    if err != nil {
        return &validation, err
    }
    
    // 4. 更新查询记录
    isFinal := result.Status >= 2 // 2:有效 3:无效 4:已兑换 5:检测失败 6:点数不足
    if err := s.queryControl.UpdateQueryRecord(validationID, result.Status, result.Message, isFinal); err != nil {
        return &validation, err
    }
    
    // 5. 重新获取更新后的记录
    s.db.First(&validation, validationID)
    return &validation, nil
}

// GetUserValidations 获取用户验证历史
func (s *UserCardValidationService) GetUserValidations(userID uint, page, pageSize int, filters map[string]interface{}) ([]*CardValidation, int64, error) {
    query := s.db.Where("user_id = ?", userID)
    
    // 应用筛选条件
    if status, ok := filters["status"]; ok {
        query = query.Where("status = ?", status)
    }
    if cardTypeID, ok := filters["card_type_id"]; ok {
        query = query.Where("card_type_id = ?", cardTypeID)
    }
    
    var total int64
    query.Model(&CardValidation{}).Count(&total)
    
    var validations []*CardValidation
    offset := (page - 1) * pageSize
    err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&validations).Error
    
    return validations, total, err
}
```

#### 6.2.2 查询控制服务

```go
package service

import (
    "fmt"
    "time"
    "gorm.io/gorm"
)

type QueryControlService struct {
    db *gorm.DB
}

func NewQueryControlService(db *gorm.DB) *QueryControlService {
    return &QueryControlService{db: db}
}

// CanQuery 检查卡片是否可以查询（用户端限制）
func (s *QueryControlService) CanQuery(validationID uint) (bool, string) {
    var validation CardValidation
    if err := s.db.First(&validation, validationID).Error; err != nil {
        return false, "验证记录不存在"
    }
    
    // 如果已经是最终状态，不需要再查询
    if validation.IsFinalStatus {
        return false, "已是最终状态，无需重复查询"
    }
    
    // 检查查询频率限制（5分钟内只能查询一次）
    if validation.LastQueryTime != nil {
        timeSinceLastQuery := time.Since(*validation.LastQueryTime)
        if timeSinceLastQuery < 5*time.Minute {
            remainingTime := 5*time.Minute - timeSinceLastQuery
            return false, fmt.Sprintf("请等待 %d 秒后再查询", int(remainingTime.Seconds()))
        }
    }
    
    return true, ""
}

// CanQueryAdmin 管理员查询检查（无限制）
func (s *QueryControlService) CanQueryAdmin(validationID uint) (bool, string) {
    var validation CardValidation
    if err := s.db.First(&validation, validationID).Error; err != nil {
        return false, "验证记录不存在"
    }
    return true, "" // 管理员无限制
}

// UpdateQueryRecord 更新查询记录
func (s *QueryControlService) UpdateQueryRecord(validationID uint, status int, resultMessage string, isFinal bool) error {
    now := time.Now()
    updates := map[string]interface{}{
        "last_query_time": &now,
        "query_count":     gorm.Expr("query_count + 1"),
        "status":          status,
        "result_message": resultMessage,
        "is_final_status": isFinal,
    }
    
    if isFinal {
        updates["check_time"] = &now
    }
    
    return s.db.Model(&CardValidation{}).Where("id = ?", validationID).Updates(updates).Error
}
```

### 6.3 管理员端服务实现（Go语言）

#### 6.3.1 管理员验证服务

```go
package service

import (
    "errors"
    "time"
    "gorm.io/gorm"
)

type AdminCardValidationService struct {
    db                *gorm.DB
    thirdPartyService *ThirdPartyService
    queryControl      *QueryControlService
}

func NewAdminCardValidationService(db *gorm.DB, thirdParty *ThirdPartyService, queryControl *QueryControlService) *AdminCardValidationService {
    return &AdminCardValidationService{
        db:                db,
        thirdPartyService: thirdParty,
        queryControl:      queryControl,
    }
}

// SubmitValidation 管理员提交验证请求（免费）
func (s *AdminCardValidationService) SubmitValidation(adminID uint, req *CardValidationRequest) (*CardValidation, error) {
    // 1. 创建验证记录（管理员免费）
    validation := &CardValidation{
        AdminID:      &adminID, // 管理员ID
        CardTypeID:   req.CardTypeID,
        CardNumber:   req.CardNumber,
        PinCode:      req.PinCode,
        RegionID:     req.RegionID,
        RegionName:   req.RegionName,
        Status:       0, // 初始状态
        QueryCount:   0,
        IsFinalStatus: false,
        CostTru:      0, // 管理员免费
        IsFree:       true,
        CreatedAt:    time.Now(),
    }
    
    if err := s.db.Create(validation).Error; err != nil {
        return nil, err
    }
    
    // 2. 调用第三方API提交测卡请求
    success, err := s.thirdPartyService.SubmitCardValidation(validation)
    if err != nil {
        // 更新状态为失败
        s.db.Model(validation).Updates(map[string]interface{}{
            "status": 5,
            "result_message": "提交失败: " + err.Error(),
            "is_final_status": true,
        })
        return nil, err
    }
    
    if success {
        // 更新状态为测卡中
        now := time.Now()
        s.db.Model(validation).Updates(map[string]interface{}{
            "status": 1,
            "last_query_time": &now,
        })
    }
    
    return validation, nil
}

// QueryValidationResult 管理员查询验证结果（无限制）
func (s *AdminCardValidationService) QueryValidationResult(validationID uint) (*CardValidation, error) {
    // 1. 获取验证记录
    var validation CardValidation
    if err := s.db.First(&validation, validationID).Error; err != nil {
        return nil, errors.New("验证记录不存在")
    }
    
    // 2. 管理员查询无限制
    canQuery, message := s.queryControl.CanQueryAdmin(validationID)
    if !canQuery {
        return &validation, errors.New(message)
    }
    
    // 3. 调用第三方API查询结果
    result, err := s.thirdPartyService.QueryCardValidationResult(&validation)
    if err != nil {
        return &validation, err
    }
    
    // 4. 更新查询记录
    isFinal := result.Status >= 2 // 2:有效 3:无效 4:已兑换 5:检测失败 6:点数不足
    if err := s.queryControl.UpdateQueryRecord(validationID, result.Status, result.Message, isFinal); err != nil {
        return &validation, err
    }
    
    // 5. 重新获取更新后的记录
    s.db.First(&validation, validationID)
    return &validation, nil
}

// GetAllValidations 获取所有验证记录（管理员）
func (s *AdminCardValidationService) GetAllValidations(page, pageSize int, filters map[string]interface{}) ([]*CardValidation, int64, error) {
    query := s.db.Model(&CardValidation{})
    
    // 应用筛选条件
    if userID, ok := filters["user_id"]; ok {
        query = query.Where("user_id = ?", userID)
    }
    if adminID, ok := filters["admin_id"]; ok {
        query = query.Where("admin_id = ?", adminID)
    }
    if status, ok := filters["status"]; ok {
        query = query.Where("status = ?", status)
    }
    if cardTypeID, ok := filters["card_type_id"]; ok {
        query = query.Where("card_type_id = ?", cardTypeID)
    }
    if startDate, ok := filters["start_date"]; ok {
        query = query.Where("created_at >= ?", startDate)
    }
    if endDate, ok := filters["end_date"]; ok {
        query = query.Where("created_at <= ?", endDate)
    }
    if isFree, ok := filters["is_free"]; ok {
        query = query.Where("is_free = ?", isFree)
    }
    
    var total int64
    query.Count(&total)
    
    var validations []*CardValidation
    offset := (page - 1) * pageSize
    err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&validations).Error
    
    return validations, total, err
}

// GetValidationStats 获取验证统计数据
func (s *AdminCardValidationService) GetValidationStats(startDate, endDate time.Time) (*ValidationStats, error) {
    var stats ValidationStats
    
    // 总验证次数
    s.db.Model(&CardValidation{}).Where("created_at BETWEEN ? AND ?", startDate, endDate).Count(&stats.TotalValidations)
    
    // 按状态统计
    var statusStats []struct {
        Status int
        Count  int64
    }
    s.db.Model(&CardValidation{}).Select("status, count(*) as count").Where("created_at BETWEEN ? AND ?", startDate, endDate).Group("status").Scan(&statusStats)
    
    stats.StatusBreakdown = make(map[int]int64)
    for _, stat := range statusStats {
        stats.StatusBreakdown[stat.Status] = stat.Count
    }
    
    // 总收入（TRU币）
    s.db.Model(&CardValidation{}).Select("COALESCE(SUM(cost_tru), 0)").Where("created_at BETWEEN ? AND ? AND user_id IS NOT NULL", startDate, endDate).Scan(&stats.TotalRevenue)
    
    // API调用次数和成本
    s.db.Model(&CardValidationLog{}).Where("created_at BETWEEN ? AND ?", startDate, endDate).Count(&stats.TotalAPICalls)
    s.db.Model(&CardValidationLog{}).Select("COALESCE(SUM(api_cost), 0)").Where("created_at BETWEEN ? AND ?", startDate, endDate).Scan(&stats.TotalAPICost)
    
    return &stats, nil
}

type ValidationStats struct {
    TotalValidations int64            `json:"total_validations"`
    StatusBreakdown  map[int]int64    `json:"status_breakdown"`
    TotalRevenue     float64          `json:"total_revenue"`
    TotalAPICalls    int64            `json:"total_api_calls"`
    TotalAPICost     float64          `json:"total_api_cost"`
}
```

#### 6.3.2 卡片类型管理服务

```go
package service

import (
    "errors"
    "time"
    "gorm.io/gorm"
)

type CardTypeService struct {
    db *gorm.DB
}

func NewCardTypeService(db *gorm.DB) *CardTypeService {
    return &CardTypeService{db: db}
}

// CreateCardType 创建卡片类型
func (s *CardTypeService) CreateCardType(req *CreateCardTypeRequest) (*CardType, error) {
    cardType := &CardType{
        ProductMark:       req.ProductMark,
        Name:              req.Name,
        Description:       req.Description,
        PriceTru:          req.PriceTru,
        IsActive:          req.IsActive,
        RequiresPin:       req.RequiresPin,
        RequiresRegion:    req.RequiresRegion,
        CardFormat:        req.CardFormat,
        SupportedRegions:  req.SupportedRegions,
        CreatedAt:         time.Now(),
        UpdatedAt:         time.Now(),
    }
    
    if err := s.db.Create(cardType).Error; err != nil {
        return nil, err
    }
    
    return cardType, nil
}

// UpdateCardType 更新卡片类型
func (s *CardTypeService) UpdateCardType(id uint, req *UpdateCardTypeRequest) (*CardType, error) {
    var cardType CardType
    if err := s.db.First(&cardType, id).Error; err != nil {
        return nil, errors.New("卡片类型不存在")
    }
    
    updates := map[string]interface{}{
        "updated_at": time.Now(),
    }
    
    if req.Name != nil {
        updates["name"] = *req.Name
    }
    if req.Description != nil {
        updates["description"] = *req.Description
    }
    if req.PriceTru != nil {
        updates["price_tru"] = *req.PriceTru
    }
    if req.IsActive != nil {
        updates["is_active"] = *req.IsActive
    }
    if req.RequiresPin != nil {
        updates["requires_pin"] = *req.RequiresPin
    }
    if req.RequiresRegion != nil {
        updates["requires_region"] = *req.RequiresRegion
    }
    if req.CardFormat != nil {
        updates["card_format"] = *req.CardFormat
    }
    if req.SupportedRegions != nil {
        updates["supported_regions"] = *req.SupportedRegions
    }
    
    if err := s.db.Model(&cardType).Updates(updates).Error; err != nil {
        return nil, err
    }
    
    return &cardType, nil
}

// DeleteCardType 删除卡片类型
func (s *CardTypeService) DeleteCardType(id uint) error {
    // 检查是否有关联的验证记录
    var count int64
    s.db.Model(&CardValidation{}).Where("card_type_id = ?", id).Count(&count)
    if count > 0 {
        return errors.New("该卡片类型下存在验证记录，无法删除")
    }
    
    return s.db.Delete(&CardType{}, id).Error
}

// GetCardTypes 获取卡片类型列表
func (s *CardTypeService) GetCardTypes(enabledOnly bool) ([]*CardType, error) {
    query := s.db.Model(&CardType{})
    if enabledOnly {
        query = query.Where("is_active = ?", true)
    }
    
    var cardTypes []*CardType
    err := query.Order("created_at DESC").Find(&cardTypes).Error
    return cardTypes, err
}

// GetCardTypeByID 根据ID获取卡片类型
func (s *CardTypeService) GetCardTypeByID(id uint) (*CardType, error) {
    var cardType CardType
    if err := s.db.First(&cardType, id).Error; err != nil {
        return nil, errors.New("卡片类型不存在")
    }
    return &cardType, nil
}
```

### 6.4 MD5签名生成器
```go
package signature

import (
    "crypto/md5"
    "fmt"
    "sort"
    "strings"
)

type SignatureGenerator struct {
    secret string
}

func NewSignatureGenerator(secret string) *SignatureGenerator {
    return &SignatureGenerator{
        secret: secret,
    }
}

func (s *SignatureGenerator) GenerateSign(params map[string]interface{}) string {
    // 1. 参数排序
    keys := make([]string, 0, len(params))
    for k := range params {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    
    // 2. 拼接参数
    var parts []string
    for _, k := range keys {
        if v := params[k]; v != nil && v != "" {
            parts = append(parts, fmt.Sprintf("%s%v", k, v))
        }
    }
    
    // 3. 生成签名字符串
    signStr := s.secret + strings.Join(parts, "") + s.secret
    
    // 4. MD5加密
    hash := md5.Sum([]byte(signStr))
    return fmt.Sprintf("%x", hash)
}
```

### 6.4 收费系统集成
- TRU币余额检查
- 扣费操作
- 退费机制（可选）

### 6.5 权限控制
- 用户身份验证
- 管理员权限验证
- 接口访问控制

## 7. 详细API接口设计

### 7.1 提交验证请求
```http
POST /api/v1/card-validation
Content-Type: application/json
Authorization: Bearer {token}

{
    "card_type_id": 1,
    "card_number": "1234567890123456",
    "pin_code": "ABCD1234",
    "region_id": 1,
    "auto_type": 0
}
```

**响应示例：**
```json
{
    "code": 200,
    "message": "验证请求已提交",
    "data": {
        "validation_id": 12345,
        "status": 1,
        "status_text": "测卡中",
        "cost_tru": 0.5,
        "can_query_now": false,
        "next_query_time": "2024-01-01T10:05:00Z",
        "estimated_time": "请稍后点击查询按钮获取最新状态"
    }
}
```

### 7.2 查询验证结果（用户触发）
```http
GET /api/v1/card-validation/{validation_id}/query
Authorization: Bearer {token}
```

**响应示例（可以查询）：**
```json
{
    "code": 200,
    "message": "查询成功",
    "data": {
        "validation_id": 12345,
        "status": 2,
        "status_text": "有效",
        "result_message": "卡片验证成功，余额充足",
        "check_time": "2024-01-01T10:03:30Z",
        "is_final": true,
        "query_count": 2,
        "last_query_time": "2024-01-01T10:03:30Z"
    }
}
```

**响应示例（查询限制）：**
```json
{
    "code": 429,
    "message": "查询过于频繁",
    "data": {
        "validation_id": 12345,
        "status": 1,
        "status_text": "测卡中",
        "can_query_again": false,
        "next_query_time": "2024-01-01T10:08:00Z",
        "last_query_time": "2024-01-01T10:03:00Z",
        "reason": "同一卡号5分钟内只能查询一次"
    }
}
```

**响应示例（已缓存结果）：**
```json
{
    "code": 200,
    "message": "返回缓存结果",
    "data": {
        "validation_id": 12345,
        "status": 3,
        "status_text": "无效",
        "result_message": "卡片已失效或余额不足",
        "check_time": "2024-01-01T09:45:20Z",
        "is_final": true,
        "from_cache": true,
        "query_count": 3
    }
}
```

### 7.3 获取验证历史
```http
GET /api/v1/card-validations?page=1&limit=20&status=2
Authorization: Bearer {token}
```

**响应示例：**
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "total": 50,
        "page": 1,
        "limit": 20,
        "items": [
            {
                "id": 12345,
                "card_type_name": "苹果卡",
                "card_number_masked": "X123****3456",
                "status": 2,
                "status_text": "有效",
                "cost_tru": 1.0,
                "query_count": 2,
                "created_at": "2024-01-01T10:00:00Z",
                "check_time": "2024-01-01T10:03:30Z"
            }
        ]
    }
}
```

## 8. 前端实现建议

### 8.1 查询按钮状态管理
```javascript
// 查询按钮状态控制
const QueryButton = ({ validation }) => {
  const [isQuerying, setIsQuerying] = useState(false);
  const [nextQueryTime, setNextQueryTime] = useState(null);
  
  const canQuery = () => {
    if (validation.is_final_status) return false;
    if (nextQueryTime && new Date() < nextQueryTime) return false;
    return true;
  };
  
  const handleQuery = async () => {
    setIsQuerying(true);
    try {
      const result = await queryValidationResult(validation.id);
      // 更新状态
      updateValidation(result.data);
      if (result.data.next_query_time) {
        setNextQueryTime(new Date(result.data.next_query_time));
      }
    } catch (error) {
      // 处理错误
    } finally {
      setIsQuerying(false);
    }
  };
  
  return (
    <button 
      disabled={!canQuery() || isQuerying}
      onClick={handleQuery}
      className={`query-btn ${canQuery() ? 'active' : 'disabled'}`}
    >
      {isQuerying ? '查询中...' : 
       validation.is_final_status ? '已完成' :
       canQuery() ? '查询最新状态' : 
       `请等待 ${getWaitTime(nextQueryTime)}`}
    </button>
  );
};
```

### 8.2 状态显示组件
```javascript
const StatusDisplay = ({ validation }) => {
  const getStatusInfo = (status) => {
    const statusMap = {
      0: { text: '等待检测', color: 'gray', icon: 'clock' },
      1: { text: '测卡中', color: 'blue', icon: 'loading' },
      2: { text: '有效', color: 'green', icon: 'check' },
      3: { text: '无效', color: 'red', icon: 'x' },
      4: { text: '已兑换', color: 'orange', icon: 'gift' },
      5: { text: '检测失败', color: 'red', icon: 'alert' },
      6: { text: '点数不足', color: 'yellow', icon: 'warning' }
    };
    return statusMap[status] || statusMap[0];
  };
  
  const statusInfo = getStatusInfo(validation.status);
  
  return (
    <div className={`status-display status-${statusInfo.color}`}>
      <Icon name={statusInfo.icon} />
      <span>{statusInfo.text}</span>
      {validation.query_count > 0 && (
        <small>已查询 {validation.query_count} 次</small>
      )}
    </div>
  );
};
```

## 9. 安全考虑

### 9.1 敏感信息保护
- APP_SECRET使用DES加密存储
- 卡号信息脱敏显示
- API请求日志脱敏
- 与第三方API通信使用DES加密保证数据安全
- 前端显示时卡号脱敏

### 9.2 防刷机制
- 用户请求频率限制
- 异常行为监控
- 余额不足保护
- 查询频率限制

### 9.3 接口安全
- API密钥管理
- 请求签名验证
- 防重放攻击
- 访问权限控制

## 10. 监控与日志

### 10.1 业务监控
- 验证成功率统计
- 收费统计
- 第三方API响应时间
- 查询频率监控
- 成本节省效果

### 10.2 日志记录
- 用户操作日志
- 第三方API调用日志
- 错误日志
- 查询控制日志

### 10.3 告警机制
- 系统异常告警
- 费用异常告警
- 第三方API异常告警
- 查询频率异常告警

## 11. 部署配置

### 11.1 环境变量
```bash
# 第三方API配置
CARD_VALIDATOR_APP_ID=xxx
CARD_VALIDATOR_APP_SECRET=xxx
CARD_VALIDATOR_BASE_URL=https://ckxiang.com

# 加密密钥
ENCRYPTION_KEY=xxx
```

### 11.2 数据库迁移
- 创建新表的迁移文件
- 初始化数据的迁移文件

## 12. 测试计划

### 12.1 单元测试
- 签名生成测试
- 加密解密测试
- 业务逻辑测试
- 查询控制逻辑测试

### 12.2 集成测试
- 第三方API调用测试
- 完整业务流程测试
- 查询频率限制测试

### 12.3 性能测试
- 并发验证测试
- 数据库性能测试
- 查询优化效果测试

## 13. 开发计划

### Phase 1: 基础架构（3-4天）
- 数据库表创建
- 基础模型定义
- 第三方API客户端开发

### Phase 2: 核心功能（4-5天）
- 用户验证接口
- 管理员验证接口
- 收费系统集成
- 查询控制机制

### Phase 3: 管理功能（2-3天）
- 卡片类型管理
- 配置管理
- 验证记录管理

### Phase 4: 前端优化（2-3天）
- 查询按钮状态管理
- 状态显示优化
- 用户体验提升

### Phase 5: 测试与优化（2-3天）
- 单元测试
- 集成测试
- 性能优化

## 14. 总结

### 14.1 优化效果
通过实施用户触发查询模式，预期可以实现：
- **成本节省**：API调用次数减少80-90%
- **用户体验**：清晰的状态显示和查询控制
- **系统稳定性**：减少不必要的API调用压力
- **可维护性**：简化的查询逻辑和状态管理

### 14.2 关键技术点
1. **查询控制机制**：通过数据库字段控制查询频率和缓存
2. **状态管理**：明确的状态流转和最终状态标记
3. **用户体验**：智能的按钮状态和友好的提示信息
4. **成本控制**：详细的查询日志和成本统计

### 14.3 实施建议
1. **分阶段实施**：先实现基础功能，再优化用户体验
2. **数据迁移**：现有数据需要添加新的控制字段
3. **用户教育**：向用户说明新的查询模式
4. **监控观察**：密切关注实施后的效果和用户反馈

---

**注意事项：**
1. **加密方式确认**：第三方API文档明确使用DES加密（非AES），Java示例代码也使用DES，必须严格按照此要求实现
2. **区域配置**：各卡片类型的支持区域已按原始文档完整配置，注意不同卡片类型的区域字段名称不同（regionId/regionName/chName）
3. **卡片格式验证**：严格按照文档要求验证卡号和PIN码格式，避免无效请求
4. **查询优化**：重点关注查询控制机制的实现，确保成本优化效果
5. 建议在开发环境先使用测试配置，避免产生实际费用
6. 考虑实现验证失败的退费机制，提升用户体验
7. 需要与现有钱包系统做好集成，确保TRU币扣费的准确性

请确认以上设计方案是否符合您的需求，我将根据您的反馈进行调整并开始具体的开发工作。