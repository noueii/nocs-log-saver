package entities

import (
	"time"

	"github.com/google/uuid"
)

// Server represents a CS2 game server
type Server struct {
	ID          string     `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	IPAddress   string     `json:"ip_address" db:"ip_address"`
	APIKey      string     `json:"api_key" db:"api_key"`
	Description string     `json:"description" db:"description"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	LastSeen    time.Time  `json:"last_seen" db:"last_seen"`
	CreatedBy   *uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// GameSession represents a game session (server or match)
type GameSession struct {
	ID        string                 `json:"id"`
	ServerID  string                 `json:"server_id"`
	MapName   string                 `json:"map_name,omitempty"`
	StartedAt time.Time              `json:"started_at"`
	EndedAt   *time.Time             `json:"ended_at,omitempty"`
	Status    SessionStatus          `json:"status"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// SessionStatus represents the current state of a session
type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusTerminated SessionStatus = "terminated"
)