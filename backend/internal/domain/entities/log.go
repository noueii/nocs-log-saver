package entities

import (
	"time"
)

// Log represents a raw log entry from a CS2 server
type Log struct {
	ID        string    `json:"id"`
	ServerID  string    `json:"server_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ParsedLog represents a successfully parsed log entry
type ParsedLog struct {
	ID        string                 `json:"id"`
	RawLogID  string                 `json:"raw_log_id"`
	ServerID  string                 `json:"server_id"`
	EventType string                 `json:"event_type"`
	EventData map[string]interface{} `json:"event_data"`
	GameTime  string                 `json:"game_time,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// FailedParse represents a log that couldn't be parsed
type FailedParse struct {
	ID           string    `json:"id"`
	RawLogID     string    `json:"raw_log_id"`
	ErrorMessage string    `json:"error_message"`
	RetryCount   int       `json:"retry_count"`
	LastRetry    time.Time `json:"last_retry,omitempty"`
	Resolved     bool      `json:"resolved"`
	CreatedAt    time.Time `json:"created_at"`
}