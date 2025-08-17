package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/noueii/nocs-log-saver/internal/domain/entities"
)

// PostgresUserRepository implements the UserRepository interface
type PostgresUserRepository struct {
	db *sqlx.DB
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(db *sqlx.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// Create creates a new user
func (r *PostgresUserRepository) Create(ctx context.Context, user *entities.User) error {
	query := `
		INSERT INTO users (id, email, username, password_hash, full_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.Username, user.PasswordHash,
		user.FullName, user.Role, user.IsActive, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

// FindByID finds a user by ID
func (r *PostgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	var user entities.User
	query := `SELECT * FROM users WHERE id = $1`
	err := r.db.GetContext(ctx, &user, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return &user, err
}

// FindByEmail finds a user by email
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	var user entities.User
	query := `SELECT * FROM users WHERE email = $1`
	err := r.db.GetContext(ctx, &user, query, email)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return &user, err
}

// FindByUsername finds a user by username
func (r *PostgresUserRepository) FindByUsername(ctx context.Context, username string) (*entities.User, error) {
	var user entities.User
	query := `SELECT * FROM users WHERE username = $1`
	err := r.db.GetContext(ctx, &user, query, username)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return &user, err
}

// Update updates a user
func (r *PostgresUserRepository) Update(ctx context.Context, user *entities.User) error {
	query := `
		UPDATE users 
		SET email = $2, username = $3, password_hash = $4, full_name = $5, 
			role = $6, is_active = $7, updated_at = $8
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.Username, user.PasswordHash,
		user.FullName, user.Role, user.IsActive, user.UpdatedAt,
	)
	return err
}

// UpdateLastLogin updates the last login timestamp
func (r *PostgresUserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET last_login = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// ListUsers lists all users with pagination
func (r *PostgresUserRepository) ListUsers(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	var users []*entities.User
	query := `SELECT * FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	err := r.db.SelectContext(ctx, &users, query, limit, offset)
	return users, err
}