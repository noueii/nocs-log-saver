package services

import (
	"fmt"
	"time"

	cs2log "github.com/janstuemmel/cs2-log"
	"github.com/jmoiron/sqlx"
)

// ParserService handles CS2 log parsing
type ParserService struct {
	db *sqlx.DB
}

// NewParserService creates a new parser service
func NewParserService(db *sqlx.DB) *ParserService {
	return &ParserService{
		db: db,
	}
}

// ParseAndStore parses a raw log and stores the result
func (s *ParserService) ParseAndStore(rawLogID, serverID, content string) error {
	// Try to parse the log using the Parse function
	parsedLog, err := cs2log.Parse(content)
	
	if err != nil {
		// Store as failed parse
		return s.storeFailedParse(rawLogID, err.Error())
	}
	
	// Convert to JSON using the library's ToJSON function
	eventData := cs2log.ToJSON(parsedLog)
	
	// Determine event type
	eventType := s.getEventType(parsedLog)
	
	// Store parsed log
	query := `
		INSERT INTO parsed_logs (raw_log_id, server_id, event_type, event_data, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	
	_, err = s.db.Exec(query, rawLogID, serverID, eventType, eventData, time.Now())
	return err
}

// storeFailedParse stores a failed parse attempt
func (s *ParserService) storeFailedParse(rawLogID, errorMsg string) error {
	query := `
		INSERT INTO failed_parses (raw_log_id, error_message, created_at)
		VALUES ($1, $2, $3)
	`
	
	_, err := s.db.Exec(query, rawLogID, errorMsg, time.Now())
	return err
}

// getEventType determines the event type from parsed log
func (s *ParserService) getEventType(parsedLog cs2log.Message) string {
	// The cs2-log library returns different types for different events
	// We need to type switch to determine the event
	switch parsedLog.(type) {
	case *cs2log.PlayerKill:
		return "kill"
	case *cs2log.PlayerAttack:
		return "attack"
	case *cs2log.WorldRoundStart:
		return "round_start"
	case *cs2log.WorldRoundEnd:
		return "round_end"
	case *cs2log.PlayerConnected:
		return "player_connect"
	case *cs2log.PlayerDisconnected:
		return "player_disconnect"
	case *cs2log.PlayerSay:
		return "chat"
	case *cs2log.PlayerSwitched:
		return "team_switch"
	case *cs2log.WorldMatchStart:
		return "match_start"
	case *cs2log.GameOver:
		return "match_end"
	case *cs2log.PlayerBombPlanted:
		return "bomb_planted"
	case *cs2log.PlayerBombDefused:
		return "bomb_defused"
	case *cs2log.PlayerPurchase:
		return "purchase"
	case *cs2log.PlayerThrew:
		return "grenade_thrown"
	case *cs2log.Unknown:
		return "unknown"
	default:
		return fmt.Sprintf("unknown_%T", parsedLog)
	}
}

// ProcessUnparsedLogs processes all unparsed raw logs
func (s *ParserService) ProcessUnparsedLogs() error {
	query := `
		SELECT r.id, r.server_id, r.content 
		FROM raw_logs r
		LEFT JOIN parsed_logs p ON r.id = p.raw_log_id
		LEFT JOIN failed_parses f ON r.id = f.raw_log_id
		WHERE p.id IS NULL AND f.id IS NULL
		LIMIT 100
	`
	
	rows, err := s.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	for rows.Next() {
		var id, serverID, content string
		if err := rows.Scan(&id, &serverID, &content); err != nil {
			continue
		}
		
		// Parse in background
		go s.ParseAndStore(id, serverID, content)
	}
	
	return nil
}