package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/noueii/nocs-log-saver/internal/domain/entities"
	"github.com/noueii/nocs-log-saver/internal/domain/repositories"
)

// PostgresWhitelistRepository implements WhitelistRepository using PostgreSQL
type PostgresWhitelistRepository struct {
	db    *sqlx.DB
	cache map[string]bool
	mu    sync.RWMutex
}

// NewPostgresWhitelistRepository creates a new PostgreSQL whitelist repository
func NewPostgresWhitelistRepository(db *sqlx.DB) repositories.WhitelistRepository {
	repo := &PostgresWhitelistRepository{
		db:    db,
		cache: make(map[string]bool),
	}
	// Load cache on startup
	go repo.refreshCache(context.Background())
	return repo
}

// Create adds a new IP to the whitelist
func (r *PostgresWhitelistRepository) Create(ctx context.Context, entry *entities.IPWhitelist) error {
	query := `
		INSERT INTO ip_whitelist (ip_address, server_id, description, enabled, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		entry.IPAddress,
		entry.ServerID,
		entry.Description,
		entry.Enabled,
		entry.CreatedBy,
	).Scan(&entry.ID, &entry.CreatedAt, &entry.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("create whitelist entry: %w", err)
	}
	
	// Update cache
	r.mu.Lock()
	if entry.Enabled {
		r.cache[entry.IPAddress] = true
	}
	r.mu.Unlock()
	
	return nil
}

// FindByIP retrieves a whitelist entry by IP
func (r *PostgresWhitelistRepository) FindByIP(ctx context.Context, ip string) (*entities.IPWhitelist, error) {
	query := `
		SELECT id, ip_address, server_id, description, enabled, created_at, updated_at, created_by
		FROM ip_whitelist
		WHERE ip_address = $1
	`
	
	entry := &entities.IPWhitelist{}
	err := r.db.QueryRowContext(ctx, query, ip).Scan(
		&entry.ID,
		&entry.IPAddress,
		&entry.ServerID,
		&entry.Description,
		&entry.Enabled,
		&entry.CreatedAt,
		&entry.UpdatedAt,
		&entry.CreatedBy,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find by IP: %w", err)
	}
	
	return entry, nil
}

// FindAll retrieves all whitelist entries
func (r *PostgresWhitelistRepository) FindAll(ctx context.Context) ([]*entities.IPWhitelist, error) {
	query := `
		SELECT id, ip_address, server_id, description, enabled, created_at, updated_at, created_by
		FROM ip_whitelist
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find all: %w", err)
	}
	defer rows.Close()
	
	var entries []*entities.IPWhitelist
	for rows.Next() {
		entry := &entities.IPWhitelist{}
		err := rows.Scan(
			&entry.ID,
			&entry.IPAddress,
			&entry.ServerID,
			&entry.Description,
			&entry.Enabled,
			&entry.CreatedAt,
			&entry.UpdatedAt,
			&entry.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		entries = append(entries, entry)
	}
	
	return entries, nil
}

// FindEnabled retrieves only enabled whitelist entries
func (r *PostgresWhitelistRepository) FindEnabled(ctx context.Context) ([]*entities.IPWhitelist, error) {
	query := `
		SELECT id, ip_address, server_id, description, enabled, created_at, updated_at, created_by
		FROM ip_whitelist
		WHERE enabled = true
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find enabled: %w", err)
	}
	defer rows.Close()
	
	var entries []*entities.IPWhitelist
	for rows.Next() {
		entry := &entities.IPWhitelist{}
		err := rows.Scan(
			&entry.ID,
			&entry.IPAddress,
			&entry.ServerID,
			&entry.Description,
			&entry.Enabled,
			&entry.CreatedAt,
			&entry.UpdatedAt,
			&entry.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		entries = append(entries, entry)
	}
	
	return entries, nil
}

// Update modifies an existing whitelist entry
func (r *PostgresWhitelistRepository) Update(ctx context.Context, entry *entities.IPWhitelist) error {
	query := `
		UPDATE ip_whitelist
		SET ip_address = $2, server_id = $3, description = $4, enabled = $5, updated_at = NOW()
		WHERE id = $1
	`
	
	_, err := r.db.ExecContext(ctx, query,
		entry.ID,
		entry.IPAddress,
		entry.ServerID,
		entry.Description,
		entry.Enabled,
	)
	
	if err != nil {
		return fmt.Errorf("update whitelist entry: %w", err)
	}
	
	// Refresh cache
	go r.refreshCache(context.Background())
	
	return nil
}

// Delete removes an IP from the whitelist
func (r *PostgresWhitelistRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM ip_whitelist WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete whitelist entry: %w", err)
	}
	
	// Refresh cache
	go r.refreshCache(context.Background())
	
	return nil
}

// IsAllowed checks if an IP is allowed (with caching)
func (r *PostgresWhitelistRepository) IsAllowed(ctx context.Context, ip string) (bool, error) {
	// Check cache first
	r.mu.RLock()
	allowed, exists := r.cache[ip]
	r.mu.RUnlock()
	
	if exists {
		return allowed, nil
	}
	
	// Fall back to database
	query := `
		SELECT EXISTS(
			SELECT 1 FROM ip_whitelist 
			WHERE ip_address = $1 AND enabled = true
		)
	`
	
	var dbAllowed bool
	err := r.db.QueryRowContext(ctx, query, ip).Scan(&dbAllowed)
	if err != nil {
		return false, fmt.Errorf("check IP allowed: %w", err)
	}
	
	// Update cache
	r.mu.Lock()
	r.cache[ip] = dbAllowed
	r.mu.Unlock()
	
	return dbAllowed, nil
}

// refreshCache reloads the cache from database
func (r *PostgresWhitelistRepository) refreshCache(ctx context.Context) {
	entries, err := r.FindEnabled(ctx)
	if err != nil {
		return
	}
	
	newCache := make(map[string]bool)
	for _, entry := range entries {
		newCache[entry.IPAddress] = true
	}
	
	r.mu.Lock()
	r.cache = newCache
	r.mu.Unlock()
}