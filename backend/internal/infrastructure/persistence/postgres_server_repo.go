package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/noueii/nocs-log-saver/internal/domain/entities"
)

// PostgresServerRepository implements server repository using PostgreSQL
type PostgresServerRepository struct {
	db *sqlx.DB
}

// NewPostgresServerRepository creates a new PostgreSQL server repository
func NewPostgresServerRepository(db *sqlx.DB) *PostgresServerRepository {
	return &PostgresServerRepository{db: db}
}

// Create creates a new server
func (r *PostgresServerRepository) Create(ctx context.Context, server *entities.Server) error {
	// Generate API key if not provided
	if server.APIKey == "" {
		var apiKey string
		err := r.db.GetContext(ctx, &apiKey, "SELECT generate_api_key()")
		if err != nil {
			return fmt.Errorf("generate api key: %w", err)
		}
		server.APIKey = apiKey
	}

	// Generate ID if not provided
	if server.ID == "" {
		server.ID = uuid.New().String()
	}

	server.CreatedAt = time.Now()
	server.UpdatedAt = time.Now()

	query := `
		INSERT INTO servers (id, name, ip_address, api_key, description, is_active, created_by, created_at, updated_at, last_seen)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		server.ID, server.Name, server.IPAddress, server.APIKey,
		server.Description, server.IsActive, server.CreatedBy,
		server.CreatedAt, server.UpdatedAt, server.CreatedAt,
	)
	return err
}

// FindByID finds a server by ID
func (r *PostgresServerRepository) FindByID(ctx context.Context, id string) (*entities.Server, error) {
	var server entities.Server
	query := `SELECT * FROM servers WHERE id = $1`
	err := r.db.GetContext(ctx, &server, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("server not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query server: %w", err)
	}
	return &server, err
}

// FindByAPIKey finds a server by API key
func (r *PostgresServerRepository) FindByAPIKey(ctx context.Context, apiKey string) (*entities.Server, error) {
	var server entities.Server
	query := `SELECT * FROM servers WHERE api_key = $1 AND is_active = true`
	err := r.db.GetContext(ctx, &server, query, apiKey)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("server not found or inactive")
	}
	return &server, err
}

// Update updates a server
func (r *PostgresServerRepository) Update(ctx context.Context, server *entities.Server) error {
	server.UpdatedAt = time.Now()
	query := `
		UPDATE servers 
		SET name = $2, ip_address = $3, description = $4, is_active = $5, updated_at = $6
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		server.ID, server.Name, server.IPAddress,
		server.Description, server.IsActive, server.UpdatedAt,
	)
	return err
}

// UpdateLastSeen updates the last seen timestamp
func (r *PostgresServerRepository) UpdateLastSeen(ctx context.Context, serverID, ipAddress string) error {
	query := `
		UPDATE servers 
		SET last_seen = $2, ip_address = $3
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, serverID, time.Now(), ipAddress)
	return err
}

// Delete deletes a server (soft delete by deactivating)
func (r *PostgresServerRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE servers SET is_active = false, updated_at = $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, time.Now())
	return err
}

// List lists all active servers
func (r *PostgresServerRepository) List(ctx context.Context, limit, offset int) ([]*entities.Server, error) {
	var servers []*entities.Server
	query := `
		SELECT * FROM servers 
		WHERE is_active = true 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`
	err := r.db.SelectContext(ctx, &servers, query, limit, offset)
	return servers, err
}

// RegenerateAPIKey generates a new API key for a server
func (r *PostgresServerRepository) RegenerateAPIKey(ctx context.Context, serverID string) (string, error) {
	var apiKey string
	err := r.db.GetContext(ctx, &apiKey, "SELECT generate_api_key()")
	if err != nil {
		return "", fmt.Errorf("generate api key: %w", err)
	}

	query := `UPDATE servers SET api_key = $2, updated_at = $3 WHERE id = $1`
	_, err = r.db.ExecContext(ctx, query, serverID, apiKey, time.Now())
	if err != nil {
		return "", err
	}

	return apiKey, nil
}

// ValidateServerExists checks if a server exists and is active
func (r *PostgresServerRepository) ValidateServerExists(ctx context.Context, serverID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM servers WHERE id = $1 AND is_active = true)`
	err := r.db.GetContext(ctx, &exists, query, serverID)
	return exists, err
}