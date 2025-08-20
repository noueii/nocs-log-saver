package config

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL            string
	MaxConnections int
	MaxIdleConns   int
	ConnMaxLifetime time.Duration
}

// NewDatabase creates a new database connection
func NewDatabase(config DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", config.URL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxConnections)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

// RunMigrations executes database migrations
func RunMigrations(db *sqlx.DB) error {
	migrations := []string{
		// Enable required extensions
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`,
		
		// Create users table FIRST (no dependencies)
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			full_name VARCHAR(255),
			role VARCHAR(20) DEFAULT 'viewer',
			is_active BOOLEAN DEFAULT true,
			last_login TIMESTAMP,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		
		// Create function to generate API keys
		`CREATE OR REPLACE FUNCTION generate_api_key() RETURNS TEXT AS $$
		BEGIN
			RETURN 'srv_' || encode(gen_random_bytes(32), 'hex');
		END;
		$$ LANGUAGE plpgsql`,
		
		`CREATE TABLE IF NOT EXISTS servers (
			id VARCHAR(50) PRIMARY KEY,
			name VARCHAR(100),
			ip_address VARCHAR(45),
			api_key VARCHAR(255) UNIQUE,
			description TEXT,
			is_active BOOLEAN DEFAULT true,
			last_seen TIMESTAMP,
			created_by UUID REFERENCES users(id),
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		
		`CREATE TABLE IF NOT EXISTS raw_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			server_id VARCHAR(50) REFERENCES servers(id),
			content TEXT NOT NULL,
			received_at TIMESTAMP DEFAULT NOW()
		)`,
		
		`CREATE TABLE IF NOT EXISTS parsed_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			raw_log_id UUID REFERENCES raw_logs(id),
			server_id VARCHAR(50) REFERENCES servers(id),
			event_type VARCHAR(50),
			event_data JSONB,
			game_time VARCHAR(20),
			session_id VARCHAR(100),
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		
		`CREATE TABLE IF NOT EXISTS failed_parses (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			raw_log_id UUID REFERENCES raw_logs(id),
			error_message TEXT,
			retry_count INTEGER DEFAULT 0,
			last_retry TIMESTAMP,
			resolved BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		
		// Rename to game_sessions to avoid conflict with user sessions
		`CREATE TABLE IF NOT EXISTS game_sessions (
			id VARCHAR(100) PRIMARY KEY,
			server_id VARCHAR(50) REFERENCES servers(id),
			map_name VARCHAR(100),
			started_at TIMESTAMP,
			ended_at TIMESTAMP,
			status VARCHAR(20) DEFAULT 'active',
			metadata JSONB
		)`,
		
		// Create sessions table for user authentication (JWT refresh tokens)
		`CREATE TABLE IF NOT EXISTS sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			refresh_token VARCHAR(500) UNIQUE NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			ip_address VARCHAR(45),
			user_agent TEXT,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		
		`CREATE TABLE IF NOT EXISTS ip_whitelist (
			id SERIAL PRIMARY KEY,
			ip_address VARCHAR(45) UNIQUE NOT NULL,
			server_id VARCHAR(50) REFERENCES servers(id),
			description VARCHAR(255),
			enabled BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),
			created_by VARCHAR(100)
		)`,
		
		// Create indexes
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_servers_api_key ON servers(api_key) WHERE is_active = true`,
		`CREATE INDEX IF NOT EXISTS idx_servers_active ON servers(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_raw_logs_server_id ON raw_logs(server_id)`,
		`CREATE INDEX IF NOT EXISTS idx_parsed_logs_session_id ON parsed_logs(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_parsed_logs_event_type ON parsed_logs(event_type)`,
		`CREATE INDEX IF NOT EXISTS idx_game_sessions_server_id ON game_sessions(server_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token ON sessions(refresh_token)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_ip_whitelist_ip ON ip_whitelist(ip_address) WHERE enabled = true`,
		
		// Insert default admin user (password: Admin123!)
		// This hash is bcrypt for "Admin123!" with cost 10
		`INSERT INTO users (username, email, password_hash, full_name, role, is_active) 
		 VALUES ('admin', 'admin@cs2logs.local', '$2a$10$aEfbkq9FjLKf08TDjViFQ.7f8i/Mwc2Z3boihMEgpMR39rIByH3A2', 'System Administrator', 'admin', true)
		 ON CONFLICT (username) DO NOTHING`,
		
		// Insert test server with generated API key
		`INSERT INTO servers (id, name, ip_address, api_key, is_active, description)
		 VALUES ('testserver', 'Test Server', '127.0.0.1', generate_api_key(), true, 'Default test server for development')
		 ON CONFLICT (id) DO UPDATE SET
		   api_key = COALESCE(servers.api_key, generate_api_key())`,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("run migration %d: %w", i, err)
		}
	}
	
	fmt.Println("✅ All database migrations completed successfully")
	return nil
}