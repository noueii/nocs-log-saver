package entities

import "time"

// IPWhitelist represents an allowed IP address
type IPWhitelist struct {
	ID          int       `json:"id"`
	IPAddress   string    `json:"ip_address"`
	ServerID    *string   `json:"server_id,omitempty"`
	Description string    `json:"description,omitempty"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   string    `json:"created_by,omitempty"`
}