package repositories

import (
	"context"
	"github.com/noueii/nocs-log-saver/internal/domain/entities"
)

// LogRepository defines the interface for log data access
type LogRepository interface {
	// Create saves a new raw log
	Create(ctx context.Context, log *entities.Log) error
	
	// FindByID retrieves a log by its ID
	FindByID(ctx context.Context, id string) (*entities.Log, error)
	
	// FindByServerID retrieves logs for a specific server
	FindByServerID(ctx context.Context, serverID string, limit int, offset int) ([]*entities.Log, error)
	
	// Count returns the total number of logs for a server
	Count(ctx context.Context, serverID string) (int64, error)
}

// ParsedLogRepository defines the interface for parsed log data access
type ParsedLogRepository interface {
	// Create saves a new parsed log
	Create(ctx context.Context, parsedLog *entities.ParsedLog) error
	
	// FindByID retrieves a parsed log by its ID
	FindByID(ctx context.Context, id string) (*entities.ParsedLog, error)
	
	// FindBySessionID retrieves parsed logs for a specific session
	FindBySessionID(ctx context.Context, sessionID string, limit int, offset int) ([]*entities.ParsedLog, error)
	
	// FindByEventType retrieves parsed logs by event type
	FindByEventType(ctx context.Context, eventType string, limit int, offset int) ([]*entities.ParsedLog, error)
}

// FailedParseRepository defines the interface for failed parse data access
type FailedParseRepository interface {
	// Create saves a new failed parse record
	Create(ctx context.Context, failedParse *entities.FailedParse) error
	
	// FindUnresolved retrieves unresolved failed parses
	FindUnresolved(ctx context.Context, limit int) ([]*entities.FailedParse, error)
	
	// MarkResolved marks a failed parse as resolved
	MarkResolved(ctx context.Context, id string) error
	
	// IncrementRetryCount increments the retry count for a failed parse
	IncrementRetryCount(ctx context.Context, id string) error
}