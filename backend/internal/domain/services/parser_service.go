package services

import (
	"context"
	"github.com/noueii/nocs-log-saver/internal/domain/entities"
)

// LogParser defines the interface for parsing CS2 logs
type LogParser interface {
	// Parse attempts to parse a raw log string
	Parse(ctx context.Context, raw string) (*entities.ParsedLog, error)
	
	// CanParse checks if a log line can be parsed
	CanParse(raw string) bool
	
	// GetEventType extracts the event type from a log line
	GetEventType(raw string) (string, error)
}

// SessionDetector defines the interface for detecting game sessions
type SessionDetector interface {
	// DetectSession determines the session ID for a parsed log
	DetectSession(ctx context.Context, parsedLog *entities.ParsedLog) (string, error)
	
	// IsSessionStart checks if a log indicates a session start
	IsSessionStart(parsedLog *entities.ParsedLog) bool
	
	// IsSessionEnd checks if a log indicates a session end
	IsSessionEnd(parsedLog *entities.ParsedLog) bool
}