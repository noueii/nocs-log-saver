package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/noueii/nocs-log-saver/internal/domain/entities"
	"github.com/noueii/nocs-log-saver/internal/domain/repositories"
	"github.com/noueii/nocs-log-saver/internal/domain/services"
)

// IngestLogCommand represents the command to ingest a new log
type IngestLogCommand struct {
	ServerID  string
	Content   string
	ClientIP  string
	Timestamp time.Time
}

// IngestLogHandler handles the log ingestion use case
type IngestLogHandler struct {
	logRepo      repositories.LogRepository
	parsedRepo   repositories.ParsedLogRepository
	failedRepo   repositories.FailedParseRepository
	whitelistRepo repositories.WhitelistRepository
	parser       services.LogParser
	sessionDetector services.SessionDetector
}

// NewIngestLogHandler creates a new handler with dependencies injected
func NewIngestLogHandler(
	logRepo repositories.LogRepository,
	parsedRepo repositories.ParsedLogRepository,
	failedRepo repositories.FailedParseRepository,
	whitelistRepo repositories.WhitelistRepository,
	parser services.LogParser,
	sessionDetector services.SessionDetector,
) *IngestLogHandler {
	return &IngestLogHandler{
		logRepo:      logRepo,
		parsedRepo:   parsedRepo,
		failedRepo:   failedRepo,
		whitelistRepo: whitelistRepo,
		parser:       parser,
		sessionDetector: sessionDetector,
	}
}

// Handle processes the log ingestion command
func (h *IngestLogHandler) Handle(ctx context.Context, cmd IngestLogCommand) error {
	// Check IP whitelist
	allowed, err := h.whitelistRepo.IsAllowed(ctx, cmd.ClientIP)
	if err != nil {
		return fmt.Errorf("check whitelist: %w", err)
	}
	if !allowed {
		return fmt.Errorf("IP %s not whitelisted", cmd.ClientIP)
	}

	// Create and save raw log
	log := &entities.Log{
		ID:        uuid.New().String(),
		ServerID:  cmd.ServerID,
		Content:   cmd.Content,
		CreatedAt: cmd.Timestamp,
	}

	if err := h.logRepo.Create(ctx, log); err != nil {
		return fmt.Errorf("save raw log: %w", err)
	}

	// Parse asynchronously
	go h.parseAsync(context.Background(), log)

	return nil
}

// parseAsync handles asynchronous log parsing
func (h *IngestLogHandler) parseAsync(ctx context.Context, log *entities.Log) {
	// Attempt to parse
	parsed, err := h.parser.Parse(ctx, log.Content)
	if err != nil {
		// Save as failed parse
		failedParse := &entities.FailedParse{
			ID:           uuid.New().String(),
			RawLogID:     log.ID,
			ErrorMessage: err.Error(),
			RetryCount:   0,
			Resolved:     false,
			CreatedAt:    time.Now(),
		}
		_ = h.failedRepo.Create(ctx, failedParse)
		return
	}

	// Set additional fields
	parsed.ID = uuid.New().String()
	parsed.RawLogID = log.ID
	parsed.ServerID = log.ServerID
	parsed.CreatedAt = time.Now()

	// Detect session
	if sessionID, err := h.sessionDetector.DetectSession(ctx, parsed); err == nil {
		parsed.SessionID = sessionID
	}

	// Save parsed log
	_ = h.parsedRepo.Create(ctx, parsed)
}