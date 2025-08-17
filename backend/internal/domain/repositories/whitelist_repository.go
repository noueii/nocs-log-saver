package repositories

import (
	"context"
	"github.com/noueii/nocs-log-saver/internal/domain/entities"
)

// WhitelistRepository defines the interface for IP whitelist data access
type WhitelistRepository interface {
	// Create adds a new IP to the whitelist
	Create(ctx context.Context, entry *entities.IPWhitelist) error
	
	// FindByIP checks if an IP is whitelisted
	FindByIP(ctx context.Context, ip string) (*entities.IPWhitelist, error)
	
	// FindAll retrieves all whitelist entries
	FindAll(ctx context.Context) ([]*entities.IPWhitelist, error)
	
	// FindEnabled retrieves only enabled whitelist entries
	FindEnabled(ctx context.Context) ([]*entities.IPWhitelist, error)
	
	// Update modifies an existing whitelist entry
	Update(ctx context.Context, entry *entities.IPWhitelist) error
	
	// Delete removes an IP from the whitelist
	Delete(ctx context.Context, id int) error
	
	// IsAllowed checks if an IP is allowed (enabled in whitelist)
	IsAllowed(ctx context.Context, ip string) (bool, error)
}