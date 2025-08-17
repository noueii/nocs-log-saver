package config

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// RunAuthMigrations runs authentication-related database migrations
func RunAuthMigrations(db *sql.DB) error {
	// Read migration file
	migrationPath := filepath.Join("migrations", "002_add_users_and_roles.sql")
	migrationSQL, err := ioutil.ReadFile(migrationPath)
	if err != nil {
		// If file doesn't exist, use inline migration
		return runInlineAuthMigrations(db)
	}

	// Execute migration from file
	if _, err := db.Exec(string(migrationSQL)); err != nil {
		return fmt.Errorf("run auth migration: %w", err)
	}

	return nil
}

// runInlineAuthMigrations runs auth migrations inline (fallback)
func runInlineAuthMigrations(db *sql.DB) error {
	migrations := []string{
		// Create user role enum
		`DO $$ BEGIN
			CREATE TYPE user_role AS ENUM ('super_admin', 'admin', 'viewer');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$`,

		// Users table
		`CREATE TABLE IF NOT EXISTS users (
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
		)`,

		// User indexes
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_users_role ON users(role)`,

		// Audit logs table
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id SERIAL PRIMARY KEY,
			user_id UUID REFERENCES users(id) ON DELETE SET NULL,
			action VARCHAR(50) NOT NULL,
			entity_type VARCHAR(50) NOT NULL,
			entity_id VARCHAR(100),
			old_values JSONB,
			new_values JSONB,
			ip_address VARCHAR(45),
			user_agent TEXT,
			created_at TIMESTAMP DEFAULT NOW()
		)`,

		// Sessions table for refresh tokens
		`CREATE TABLE IF NOT EXISTS user_sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			refresh_token VARCHAR(500) UNIQUE NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			ip_address VARCHAR(45),
			user_agent TEXT,
			created_at TIMESTAMP DEFAULT NOW()
		)`,

		// Session indexes
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON user_sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token ON user_sessions(refresh_token)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON user_sessions(expires_at)`,

		// Permissions table
		`CREATE TABLE IF NOT EXISTS permissions (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) UNIQUE NOT NULL,
			resource VARCHAR(100) NOT NULL,
			action VARCHAR(50) NOT NULL,
			description TEXT
		)`,

		// Role permissions mapping
		`CREATE TABLE IF NOT EXISTS role_permissions (
			role user_role NOT NULL,
			permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
			PRIMARY KEY (role, permission_id)
		)`,

		// Add user tracking to ip_whitelist
		`ALTER TABLE ip_whitelist 
		ADD COLUMN IF NOT EXISTS created_by_id UUID REFERENCES users(id) ON DELETE SET NULL`,
		
		`ALTER TABLE ip_whitelist 
		ADD COLUMN IF NOT EXISTS updated_by_id UUID REFERENCES users(id) ON DELETE SET NULL`,

		// Insert default permissions
		`INSERT INTO permissions (name, resource, action, description) VALUES
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
		ON CONFLICT (name) DO NOTHING`,

		// Assign all permissions to super_admin
		`INSERT INTO role_permissions (role, permission_id) 
		SELECT 'super_admin', id FROM permissions
		ON CONFLICT DO NOTHING`,

		// Assign limited permissions to admin
		`INSERT INTO role_permissions (role, permission_id) 
		SELECT 'admin', id FROM permissions 
		WHERE name IN (
			'whitelist.create', 'whitelist.read', 'whitelist.update', 'whitelist.delete',
			'logs.read', 'servers.read', 'servers.update'
		)
		ON CONFLICT DO NOTHING`,

		// Assign view-only permissions to viewer
		`INSERT INTO role_permissions (role, permission_id) 
		SELECT 'viewer', id FROM permissions 
		WHERE name IN ('whitelist.read', 'logs.read', 'servers.read')
		ON CONFLICT DO NOTHING`,

		// Create default admin user (password: Admin123!)
		// Password hash is bcrypt of 'Admin123!'
		`INSERT INTO users (email, username, password_hash, full_name, role) VALUES
			('admin@cs2logs.local', 'admin', 
			'$2a$10$xWHhVB5Tz7r0L3KlzQzJy.8vNjFZLXtmVvOv5yJ5PZvKHqGBqGOZa', 
			'System Administrator', 'super_admin')
		ON CONFLICT (email) DO NOTHING`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			// Log but don't fail on certain errors
			fmt.Printf("Warning during migration: %v\n", err)
		}
	}

	return nil
}