package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

// StatefulParserService handles multi-line log assembly for CS2
type StatefulParserService struct {
	db          *sqlx.DB
	parser      *ParserService
	buffers     map[string]*LogBuffer // serverID -> buffer
	bufferMutex sync.RWMutex
}

// LogBuffer holds multi-line log data being assembled
type LogBuffer struct {
	ServerID       string
	InJSONBlock    bool
	JSONLines      []string
	JSONStartTime  time.Time
	FirstRawLogID  string
	LastRawLogID   string
}

// RoundStats represents the assembled JSON statistics
type RoundStats struct {
	Name         string                 `json:"name"`
	RoundNumber  string                 `json:"round_number"`
	ScoreCT      string                 `json:"score_ct"`
	ScoreT       string                 `json:"score_t"`
	Map          string                 `json:"map"`
	Server       string                 `json:"server"`
	Fields       string                 `json:"fields"`
	Players      map[string]string      `json:"players"`
	Timestamp    time.Time              `json:"timestamp"`
	RawData      map[string]interface{} `json:"raw_data"` // For any additional fields
}

// NewStatefulParserService creates a new stateful parser service
func NewStatefulParserService(db *sqlx.DB) *StatefulParserService {
	return &StatefulParserService{
		db:      db,
		parser:  NewParserService(db),
		buffers: make(map[string]*LogBuffer),
	}
}

// ParseAndStore processes a log line with stateful awareness
func (s *StatefulParserService) ParseAndStore(rawLogID, serverID, content string) error {
	// Extract the actual CS2 log content
	actualContent := s.parser.ExtractActualContent(content)
	
	// Check if this is part of a JSON statistics block
	if s.isJSONStatsLine(actualContent) {
		return s.handleJSONStatsLine(rawLogID, serverID, actualContent)
	}
	
	// Otherwise, use the regular parser
	return s.parser.ParseAndStore(rawLogID, serverID, content)
}

// isJSONStatsLine checks if a line is part of JSON statistics
func (s *StatefulParserService) isJSONStatsLine(content string) bool {
	return strings.Contains(content, "JSON_BEGIN{") ||
		strings.Contains(content, "}}JSON_END") ||
		s.isInJSONBlock(content)
}

// isInJSONBlock checks if we're currently in a JSON block for any server
func (s *StatefulParserService) isInJSONBlock(content string) bool {
	// Check if the line looks like JSON data
	return (strings.Contains(content, "\"name\"") && strings.Contains(content, "round_stats")) ||
		(strings.Contains(content, "\"round_number\"") && strings.Contains(content, ":")) ||
		(strings.Contains(content, "\"score_ct\"") && strings.Contains(content, ":")) ||
		(strings.Contains(content, "\"score_t\"") && strings.Contains(content, ":")) ||
		(strings.Contains(content, "\"map\"") && strings.Contains(content, ":")) ||
		(strings.Contains(content, "\"server\"") && strings.Contains(content, ":")) ||
		(strings.Contains(content, "\"fields\"") && strings.Contains(content, ":")) ||
		(strings.Contains(content, "\"players\"") && strings.Contains(content, ":")) ||
		(strings.Contains(content, "\"player_") && strings.Contains(content, ":"))
}

// handleJSONStatsLine processes a line that's part of JSON statistics
func (s *StatefulParserService) handleJSONStatsLine(rawLogID, serverID, content string) error {
	s.bufferMutex.Lock()
	defer s.bufferMutex.Unlock()
	
	// Get or create buffer for this server
	buffer, exists := s.buffers[serverID]
	if !exists {
		buffer = &LogBuffer{
			ServerID:  serverID,
			JSONLines: []string{},
		}
		s.buffers[serverID] = buffer
	}
	
	// Handle JSON_BEGIN
	if strings.Contains(content, "JSON_BEGIN{") {
		// Start new JSON block
		buffer.InJSONBlock = true
		buffer.JSONLines = []string{"{"}
		buffer.JSONStartTime = time.Now()
		buffer.FirstRawLogID = rawLogID
		buffer.LastRawLogID = rawLogID
		return nil
	}
	
	// Handle JSON_END
	if strings.Contains(content, "}}JSON_END") {
		if buffer.InJSONBlock {
			// Close the JSON object
			buffer.JSONLines = append(buffer.JSONLines, "}")
			buffer.LastRawLogID = rawLogID
			
			// Assemble and store the complete JSON
			err := s.assembleAndStoreJSON(buffer)
			
			// Reset buffer
			buffer.InJSONBlock = false
			buffer.JSONLines = []string{}
			
			return err
		}
		// JSON_END without JSON_BEGIN - treat as regular log
		return s.parser.ParseAndStore(rawLogID, serverID, content)
	}
	
	// Handle JSON content lines
	if buffer.InJSONBlock {
		// Add this line to the buffer
		jsonLine := s.extractJSONContent(content)
		if jsonLine != "" {
			buffer.JSONLines = append(buffer.JSONLines, jsonLine)
			buffer.LastRawLogID = rawLogID
		}
		return nil
	}
	
	// Not in a JSON block but looks like JSON - might be out of order
	// Store as individual stat line using regular parser
	return s.parser.ParseAndStore(rawLogID, serverID, content)
}

// extractJSONContent extracts the JSON content from a log line
func (s *StatefulParserService) extractJSONContent(content string) string {
	// The content should be like: "field_name" : "value"
	// or for players: "player_X" : "data"
	
	// Handle player data lines
	if strings.Contains(content, "\"player_") {
		return strings.TrimSpace(content)
	}
	
	// Handle other JSON fields
	jsonFields := []string{
		"\"name\"", "\"round_number\"", "\"score_ct\"", "\"score_t\"",
		"\"map\"", "\"server\"", "\"fields\"", "\"players\"",
		"\"timestamp\"", "\"version\"", "\"rounds_played\"",
	}
	
	for _, field := range jsonFields {
		if strings.Contains(content, field) {
			// Special handling for "players" : { line
			if field == "\"players\"" && strings.Contains(content, "{") {
				return strings.TrimSpace(content)
			}
			return strings.TrimSpace(content)
		}
	}
	
	// Handle closing brace for players object
	if strings.TrimSpace(content) == "}" {
		return "}"
	}
	
	return ""
}

// assembleAndStoreJSON assembles the buffered lines into a JSON object and stores it
func (s *StatefulParserService) assembleAndStoreJSON(buffer *LogBuffer) error {
	// Join all lines to create the JSON string
	jsonStr := strings.Join(buffer.JSONLines, "\n")
	
	// Try to parse it as JSON to validate
	var rawData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawData); err != nil {
		// If parsing fails, store individual lines using regular parser
		// This handles malformed JSON gracefully
		for _, line := range buffer.JSONLines {
			if line != "{" && line != "}" {
				// Create a synthetic log line for each JSON field
				s.parser.ParseAndStore(buffer.LastRawLogID, buffer.ServerID, line)
			}
		}
		return fmt.Errorf("failed to parse JSON stats: %w", err)
	}
	
	// Successfully parsed - store as a single round_stats event
	eventData, err := json.Marshal(rawData)
	if err != nil {
		return fmt.Errorf("failed to marshal round stats: %w", err)
	}
	
	// Store the assembled round statistics
	query := `
		INSERT INTO parsed_logs (raw_log_id, server_id, event_type, event_data, session_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	
	// Use the last raw_log_id as the reference
	// Event type is "round_stats" for the complete assembled statistics
	_, err = s.db.Exec(query, 
		buffer.LastRawLogID, 
		buffer.ServerID, 
		"round_stats",
		string(eventData),
		nil, // session_id can be set if available
		buffer.JSONStartTime,
	)
	
	if err != nil {
		return fmt.Errorf("failed to store round stats: %w", err)
	}
	
	// Also store a reference for the first raw_log_id if different
	if buffer.FirstRawLogID != buffer.LastRawLogID {
		// Store a reference entry pointing to the complete stats
		_, err = s.db.Exec(query,
			buffer.FirstRawLogID,
			buffer.ServerID,
			"round_stats_ref",
			fmt.Sprintf(`{"refers_to_raw_log_id":"%s"}`, buffer.LastRawLogID),
			nil,
			buffer.JSONStartTime,
		)
	}
	
	return nil
}

// CleanupOldBuffers removes stale buffers (for servers that disconnected mid-JSON)
func (s *StatefulParserService) CleanupOldBuffers(maxAge time.Duration) {
	s.bufferMutex.Lock()
	defer s.bufferMutex.Unlock()
	
	now := time.Now()
	for serverID, buffer := range s.buffers {
		if buffer.InJSONBlock && now.Sub(buffer.JSONStartTime) > maxAge {
			// Buffer is too old, probably incomplete
			delete(s.buffers, serverID)
		}
	}
}