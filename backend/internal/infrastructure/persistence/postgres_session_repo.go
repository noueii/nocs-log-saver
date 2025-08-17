package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/noueii/nocs-log-saver/internal/domain/entities"
)

// PostgresSessionRepository implements the SessionRepository interface
type PostgresSessionRepository struct {
	db *sqlx.DB
}

// NewPostgresSessionRepository creates a new PostgreSQL session repository
func NewPostgresSessionRepository(db *sqlx.DB) *PostgresSessionRepository {
	return &PostgresSessionRepository{db: db}
}

// Create creates a new session
func (r *PostgresSessionRepository) Create(ctx context.Context, session *entities.UserSession) error {
	query := `
		INSERT INTO sessions (id, user_id, refresh_token, expires_at, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.RefreshToken, session.ExpiresAt,
		session.IPAddress, session.UserAgent, session.CreatedAt,
	)
	return err
}

// FindByToken finds a session by refresh token
func (r *PostgresSessionRepository) FindByToken(ctx context.Context, token string) (*entities.UserSession, error) {
	var session entities.UserSession
	query := `SELECT * FROM sessions WHERE refresh_token = $1`
	err := r.db.GetContext(ctx, &session, query, token)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	return &session, err
}

// DeleteByUserID deletes all sessions for a user
func (r *PostgresSessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// DeleteExpired deletes all expired sessions
func (r *PostgresSessionRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < NOW()`
	_, err := r.db.ExecContext(ctx, query)
	return err
}