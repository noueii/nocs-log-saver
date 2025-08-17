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
		`CREATE TABLE IF NOT EXISTS servers (
			id VARCHAR(50) PRIMARY KEY,
			name VARCHAR(100),
			ip_address VARCHAR(45),
			last_seen TIMESTAMP,
			created_at TIMESTAMP DEFAULT NOW()
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
		
		`CREATE TABLE IF NOT EXISTS sessions (
			id VARCHAR(100) PRIMARY KEY,
			server_id VARCHAR(50) REFERENCES servers(id),
			map_name VARCHAR(100),
			started_at TIMESTAMP,
			ended_at TIMESTAMP,
			status VARCHAR(20) DEFAULT 'active',
			metadata JSONB
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
		
		`CREATE INDEX IF NOT EXISTS idx_raw_logs_server_id ON raw_logs(server_id)`,
		`CREATE INDEX IF NOT EXISTS idx_parsed_logs_session_id ON parsed_logs(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_parsed_logs_event_type ON parsed_logs(event_type)`,
		`CREATE INDEX IF NOT EXISTS idx_ip_whitelist_ip ON ip_whitelist(ip_address) WHERE enabled = true`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("run migration: %w", err)
		}
	}

	return nil
}