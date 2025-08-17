-- User roles enum
CREATE TYPE user_role AS ENUM ('super_admin', 'admin', 'viewer');

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    role user_role NOT NULL DEFAULT 'viewer',
    is_active BOOLEAN DEFAULT true,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for faster lookups
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_role ON users(role);

-- Audit log for tracking who made changes
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL, -- 'CREATE', 'UPDATE', 'DELETE'
    entity_type VARCHAR(50) NOT NULL, -- 'whitelist', 'user', 'server', etc.
    entity_id VARCHAR(100),
    old_values JSONB,
    new_values JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Add user_id to ip_whitelist for tracking who added/modified entries
ALTER TABLE ip_whitelist 
ADD COLUMN IF NOT EXISTS created_by_id UUID REFERENCES users(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS updated_by_id UUID REFERENCES users(id) ON DELETE SET NULL;

-- Session table for JWT refresh tokens
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(500) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Role permissions table (for future fine-grained permissions)
CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL, -- 'create', 'read', 'update', 'delete'
    description TEXT
);

-- Role to permissions mapping
CREATE TABLE IF NOT EXISTS role_permissions (
    role user_role NOT NULL,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role, permission_id)
);

-- Insert default permissions
INSERT INTO permissions (name, resource, action, description) VALUES
    ('whitelist.create', 'whitelist', 'create', 'Can add new IP addresses to whitelist'),
    ('whitelist.read', 'whitelist', 'read', 'Can view whitelist entries'),
    ('whitelist.update', 'whitelist', 'update', 'Can modify whitelist entries'),
    ('whitelist.delete', 'whitelist', 'delete', 'Can remove whitelist entries'),
    ('users.create', 'users', 'create', 'Can create new users'),
    ('users.read', 'users', 'read', 'Can view user information'),
    ('users.update', 'users', 'update', 'Can modify user information'),
    ('users.delete', 'users', 'delete', 'Can delete users'),
    ('logs.read', 'logs', 'read', 'Can view server logs'),
    ('logs.delete', 'logs', 'delete', 'Can delete logs'),
    ('servers.read', 'servers', 'read', 'Can view server information'),
    ('servers.update', 'servers', 'update', 'Can modify server configuration'),
    ('audit.read', 'audit', 'read', 'Can view audit logs')
ON CONFLICT (name) DO NOTHING;

-- Assign permissions to roles
INSERT INTO role_permissions (role, permission_id) 
SELECT 'super_admin', id FROM permissions
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role, permission_id) 
SELECT 'admin', id FROM permissions 
WHERE name IN (
    'whitelist.create', 'whitelist.read', 'whitelist.update', 'whitelist.delete',
    'logs.read', 'servers.read', 'servers.update'
)
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role, permission_id) 
SELECT 'viewer', id FROM permissions 
WHERE name IN ('whitelist.read', 'logs.read', 'servers.read')
ON CONFLICT DO NOTHING;

-- Create default super admin user (password: admin123 - CHANGE THIS!)
-- Password hash is for 'admin123' using bcrypt
INSERT INTO users (email, username, password_hash, full_name, role) VALUES
    ('admin@cs2logs.local', 'admin', '$2a$10$YKvXJ5ypN4n8kHvv35EBYuMzYFhCDwcQg1K9b3wMRbVK6Y9Xh1gHu', 'System Administrator', 'super_admin')
ON CONFLICT (email) DO NOTHING;