package entities

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents the role of a user
type UserRole string

const (
	RoleSuperAdmin UserRole = "super_admin"
	RoleAdmin      UserRole = "admin"
	RoleViewer     UserRole = "viewer"
)

// User represents a system user
type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Email        string     `json:"email" db:"email"`
	Username     string     `json:"username" db:"username"`
	PasswordHash string     `json:"-" db:"password_hash"`
	FullName     string     `json:"full_name" db:"full_name"`
	Role         UserRole   `json:"role" db:"role"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	LastLogin    *time.Time `json:"last_login" db:"last_login"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// UserSession represents a user authentication session
type UserSession struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	RefreshToken string    `json:"refresh_token" db:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Permission represents a system permission
type Permission struct {
	ID          int    `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Resource    string `json:"resource" db:"resource"`
	Action      string `json:"action" db:"action"`
	Description string `json:"description" db:"description"`
}

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID         int                    `json:"id" db:"id"`
	UserID     *uuid.UUID             `json:"user_id" db:"user_id"`
	Action     string                 `json:"action" db:"action"`
	EntityType string                 `json:"entity_type" db:"entity_type"`
	EntityID   string                 `json:"entity_id" db:"entity_id"`
	OldValues  map[string]interface{} `json:"old_values" db:"old_values"`
	NewValues  map[string]interface{} `json:"new_values" db:"new_values"`
	IPAddress  string                 `json:"ip_address" db:"ip_address"`
	UserAgent  string                 `json:"user_agent" db:"user_agent"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
}

// HasPermission checks if user has a specific permission
func (u *User) HasPermission(resource, action string) bool {
	// Super admin has all permissions
	if u.Role == RoleSuperAdmin {
		return true
	}

	// Define role-based permissions
	permissions := map[UserRole]map[string][]string{
		RoleAdmin: {
			"servers": {"create", "read", "update", "delete"},
			"logs":    {"read"},
			"users":   {"read"},
		},
		RoleViewer: {
			"servers": {"read"},
			"logs":    {"read"},
		},
	}

	if rolePerms, ok := permissions[u.Role]; ok {
		if actions, ok := rolePerms[resource]; ok {
			for _, a := range actions {
				if a == action {
					return true
				}
			}
		}
	}

	return false
}

// CanManageServers checks if user can manage servers
func (u *User) CanManageServers() bool {
	return u.Role == RoleSuperAdmin || u.Role == RoleAdmin
}

// CanManageUsers checks if user can manage other users
func (u *User) CanManageUsers() bool {
	return u.Role == RoleSuperAdmin
}

// CanViewAuditLogs checks if user can view audit logs
func (u *User) CanViewAuditLogs() bool {
	return u.Role == RoleSuperAdmin
}