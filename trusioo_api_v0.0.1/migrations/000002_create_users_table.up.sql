-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, inactive, suspended
    email_verified BOOLEAN NOT NULL DEFAULT false,
    email_verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_email_verified ON users(email_verified) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);