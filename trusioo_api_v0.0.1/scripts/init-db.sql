-- 数据库初始化脚本

-- 创建数据库扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 创建函数用于自动更新 updated_at 字段
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 创建通用触发器函数，用于所有表的 updated_at 自动更新
-- 这个函数会在迁移文件中的每个表创建后被调用