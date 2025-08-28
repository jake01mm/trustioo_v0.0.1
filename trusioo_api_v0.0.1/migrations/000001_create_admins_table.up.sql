-- 创建管理员表
CREATE TABLE IF NOT EXISTS admins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'admin',
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_admins_email ON admins(email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_admins_active ON admins(active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_admins_created_at ON admins(created_at);

-- 创建默认超级管理员账户（密码: admin123）
INSERT INTO admins (email, name, password, role, active) 
VALUES (
    'admin@trusioo.com', 
    'Super Admin',
    '\$2a\$10\$5uYi4nTPak35f1.TVa7MKusFESFNnhHHEzUoEPOsTfrGyfCZ0rSTi', -- bcrypt hash of 'admin123'
    'super_admin',
    true
) ON CONFLICT (email) DO NOTHING;