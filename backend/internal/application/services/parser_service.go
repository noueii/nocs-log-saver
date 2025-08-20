package services

import (
	"fmt"
	"strings"
	"time"

	cs2log "github.com/noueii/cs2-log"
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
	// Extract the actual CS2 log content from our custom format
	actualContent := s.ExtractActualContent(content)
	
	// Try to parse the log using the enhanced Parse function with custom events
	parsedLog, err := cs2log.ParseEnhanced(actualContent)
	
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
	switch msg := parsedLog.(type) {
	// Kill and damage events
	case cs2log.PlayerKill:
		return "kill"
	case cs2log.PlayerKillAssist:
		return "kill_assist"
	case cs2log.PlayerAttack:
		return "attack"
	case cs2log.PlayerKilledBomb:
		return "killed_by_bomb"
	case cs2log.PlayerKilledSuicide:
		return "suicide"
	
	// Round events
	case cs2log.WorldRoundStart:
		return "round_start"
	case cs2log.WorldRoundEnd:
		return "round_end"
	case cs2log.WorldRoundRestart:
		return "round_restart"
	case cs2log.FreezTimeStart:
		return "freeze_period_start"
	case cs2log.FreezePeriod:
		if msg.Action == "start" {
			return "freeze_period_start"
		}
		return "freeze_period_end"
	
	// Match events
	case cs2log.WorldMatchStart:
		return "match_start"
	case cs2log.GameOver:
		return "match_end"
	case cs2log.GameOverDetailed:
		return "game_over_" + msg.Mode
	case cs2log.WorldGameCommencing:
		return "game_commencing"
	
	// Player connection events
	case cs2log.PlayerConnected:
		return "player_connect"
	case cs2log.PlayerDisconnected:
		return "player_disconnect"
	case cs2log.PlayerEntered:
		return "player_entered"
	case cs2log.PlayerValidated:
		return "userid_validated"
	
	// Communication events
	case cs2log.ChatCommand:
		// Return specific command types
		switch msg.Command {
		case "pause", "forcepause":
			return "chat_pause_command"
		case "restore", "resotre":
			return "chat_restore_command"
		case "ready", "rdy":
			return "chat_ready_command"
		case "unpause":
			return "chat_unpause_command"
		case "tech":
			return "chat_tech_command"
		case "tac":
			return "chat_tac_command"
		default:
			return "chat_command"
		}
	case cs2log.PlayerSay:
		// Check for gg messages
		if msg.Text == "gg" || msg.Text == "gg wp" {
			return "chat_gg"
		}
		return "chat_message"
	
	// Team events
	case cs2log.PlayerSwitched:
		return "team_switch"
	case cs2log.TeamScored:
		return "team_scored"
	case cs2log.TeamNotice:
		return "team_notice"
	case cs2log.TeamPlaying:
		return "team_playing"
	
	// Bomb events
	case cs2log.PlayerBombPlanted:
		return "bomb_planted"
	case cs2log.PlayerBombDefused:
		return "bomb_defused"
	case cs2log.PlayerBombBeginDefuse:
		return "bomb_begin_defuse"
	case cs2log.PlayerBombGot:
		return "bomb_got"
	case cs2log.PlayerBombDropped:
		return "bomb_dropped"
	case cs2log.BombEvent:
		return "bomb_" + msg.Action
	
	// Economy events
	case cs2log.PlayerPurchase:
		return "purchase"
	case cs2log.PlayerMoneyChange:
		return "money_change"
	case cs2log.PlayerLeftBuyzone:
		return "left_buyzone"
	
	// Item events
	case cs2log.PlayerPickedUp:
		return "picked_up"
	case cs2log.PlayerDropped:
		return "dropped"
	
	// Grenade events
	case cs2log.PlayerThrew:
		return "grenade_thrown"
	case cs2log.PlayerBlinded:
		return "blinded"
	case cs2log.ProjectileSpawned:
		return "projectile_spawned"
	case cs2log.GrenadeThrowDebug:
		return "throw_debug_" + strings.TrimPrefix(msg.GrenadeType, "grenade")
	
	// Achievement events
	case cs2log.PlayerAccolade:
		if msg.IsFinal {
			return "accolade_final_" + msg.Type
		}
		return "accolade_round_" + msg.Type
	
	// Match status events
	case cs2log.MatchStatus:
		return "match_status_score"
	case cs2log.MatchPause:
		return "match_pause_" + msg.Action
	
	// Server events
	case cs2log.ServerCvar:
		// Specific cvar types
		if strings.HasPrefix(msg.Name, "mp_") {
			if msg.Name == "mp_maxrounds" {
				return "cvar_maxrounds"
			}
			if strings.Contains(msg.Name, "overtime") {
				return "cvar_overtime"
			}
			if msg.Name == "mp_freezetime" {
				return "cvar_freezetime"
			}
			if msg.Name == "mp_tournament" {
				return "cvar_tournament"
			}
			return "cvar_mp_setting"
		}
		return "server_cvar"
	case cs2log.RconCommand:
		return "rcon_command"
	
	// Map events
	case cs2log.LoadingMap:
		return "loading_map"
	case cs2log.StartedMap:
		return "started_map"
	
	// Log file events
	case cs2log.LogFile:
		return "log_file_" + msg.Action
	
	// Stats events
	case cs2log.StatsJSON:
		return "stats_json_" + msg.Type
	
	// Triggered events
	case cs2log.TriggeredEvent:
		// Normalize event names
		event := strings.ToLower(msg.Event)
		event = strings.ReplaceAll(event, "_", "-")
		event = strings.ReplaceAll(event, " ", "-")
		return "trigger_" + event
	
	// Unknown events (should be rare now)
	case cs2log.Unknown:
		// Most unknowns should now be properly classified
		// This is just a fallback
		return "unknown"
	
	default:
		// This shouldn't happen if cs2-log is working correctly
		return fmt.Sprintf("unrecognized_%T", parsedLog)
	}
}

// ParsedLog represents a parsed log result
type ParsedLog struct {
	EventType string      `json:"event_type"`
	EventData interface{} `json:"event_data"`
}

// ExtractActualContent strips prefixes from log content
func (s *ParserService) ExtractActualContent(content string) string {
	actualContent := content
	
	// Check if the line has our custom prefix format
	// Format: [2025-08-19T15:12:44Z] 18a5c248-c891-42a6-b72e-af0b184937c1: actual_log_content
	if strings.HasPrefix(content, "[") {
		endIdx := strings.Index(content, "] ")
		if endIdx != -1 {
			remaining := content[endIdx+2:]
			colonIdx := strings.Index(remaining, ": ")
			if colonIdx != -1 {
				actualContent = remaining[colonIdx+2:]
			}
		}
	}
	
	// Also handle simpler format without brackets but with UUID prefix
	if !strings.HasPrefix(actualContent, "L ") && strings.Contains(actualContent, ": ") {
		colonIdx := strings.Index(actualContent, ": ")
		possibleUUID := actualContent[:colonIdx]
		if len(possibleUUID) == 36 && strings.Count(possibleUUID, "-") == 4 {
			actualContent = actualContent[colonIdx+2:]
		}
	}
	
	// Ensure the log starts with "L " for CS2 format
	if !strings.HasPrefix(actualContent, "L ") {
		if strings.Contains(actualContent, " - ") && strings.Contains(actualContent, ":") {
			actualContent = "L " + actualContent
		}
	}
	
	// Remove milliseconds from timestamp if present (CS2 parser expects HH:MM:SS not HH:MM:SS.mmm)
	// Format: L MM/DD/YYYY - HH:MM:SS.mmm - becomes L MM/DD/YYYY - HH:MM:SS:
	if strings.HasPrefix(actualContent, "L ") {
		parts := strings.SplitN(actualContent, " - ", 3)
		if len(parts) >= 3 {
			// Check if time part has milliseconds
			timePart := parts[1]
			if dotIdx := strings.Index(timePart, "."); dotIdx != -1 {
				// Remove milliseconds (.735) and ensure it ends with colon
				timePart = timePart[:dotIdx] + ":"
				actualContent = parts[0] + " - " + timePart + " " + parts[2]
			} else if !strings.HasSuffix(timePart, ":") {
				// Ensure time ends with colon
				actualContent = parts[0] + " - " + timePart + ": " + parts[2]
			}
		}
	}
	
	return actualContent
}

// ParseLogLine parses a single log line and returns the result
func (s *ParserService) ParseLogLine(content string) (*ParsedLog, error) {
	// Extract the actual CS2 log content from our custom format
	actualContent := s.ExtractActualContent(content)
	
	// Try to parse the log using the enhanced Parse function with custom events
	parsedLog, err := cs2log.ParseEnhanced(actualContent)
	
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	
	// Convert to JSON using the library's ToJSON function
	eventData := cs2log.ToJSON(parsedLog)
	
	// Determine event type
	eventType := s.getEventType(parsedLog)
	
	return &ParsedLog{
		EventType: eventType,
		EventData: eventData,
	}, nil
}

// classifyUnknownEvent tries to classify unknown events based on their content
func (s *ParserService) classifyUnknownEvent(raw string) string {
	// Check for common patterns in unknown events
	switch {
	// Money change events (sometimes cs2-log can't parse when reaching max money)
	case strings.Contains(raw, "money change"):
		return "money_change"
	
	// Attack events (often with corrupted Steam IDs that cs2-log can't parse)
	case strings.Contains(raw, "attacked") && strings.Contains(raw, "with"):
		return "attack"
	
	// Blind events (often with corrupted Steam IDs that cs2-log can't parse)
	case strings.Contains(raw, "blinded for") && strings.Contains(raw, "by"):
		return "blinded"
	
	// Flashbang throw events (in case cs2-log fails to parse)
	case strings.Contains(raw, "threw flashbang"):
		return "grenade_thrown"
	
	// Buy zone events
	case strings.Contains(raw, "left buyzone"):
		return "left_buyzone"
	
	// Validation events
	case strings.Contains(raw, "STEAM USERID validated"):
		return "userid_validated"
	
	// Achievement/Award events
	case strings.Contains(raw, "ACCOLADE"):
		// Extract accolade type for more specific classification
		if strings.Contains(raw, ",") {
			parts := strings.Split(raw, ",")
			if len(parts) > 1 {
				accoladeType := strings.ToLower(strings.TrimSpace(parts[1]))
				// Clean up the type
				accoladeType = strings.ReplaceAll(accoladeType, " ", "_")
				accoladeType = strings.ReplaceAll(accoladeType, ":", "")
				accoladeType = strings.ReplaceAll(accoladeType, "{", "")
				accoladeType = strings.ReplaceAll(accoladeType, "}", "")
				return "accolade_" + accoladeType
			}
		}
		return "accolade"
	
	// Match status events
	case strings.Contains(raw, "MatchStatus:"):
		if strings.Contains(raw, "Score:") {
			return "match_status_score"
		}
		if strings.Contains(raw, "Team playing") {
			return "match_status_teams"
		}
		return "match_status"
	
	// Pause events
	case strings.Contains(raw, "Match pause"):
		if strings.Contains(raw, "enabled") {
			return "match_pause_enabled"
		}
		if strings.Contains(raw, "disabled") {
			return "match_pause_disabled"
		}
		return "match_pause"
	case strings.Contains(raw, "Match unpaused"):
		return "match_unpause"
	
	// Debug events
	case strings.Contains(raw, "sv_throw"):
		// Identify grenade type
		if strings.Contains(raw, "sv_throw_molotov") {
			return "throw_debug_molotov"
		}
		if strings.Contains(raw, "sv_throw_smokegrenade") {
			return "throw_debug_smoke"
		}
		if strings.Contains(raw, "sv_throw_flashgrenade") {
			return "throw_debug_flash"
		}
		if strings.Contains(raw, "sv_throw_hegrenade") {
			return "throw_debug_he"
		}
		return "throw_debug"
	
	// Bomb events not caught by main parser
	case strings.Contains(raw, "planted the bomb"):
		return "bomb_planted"
	case strings.Contains(raw, "defused the bomb"):
		return "bomb_defused"
	case strings.Contains(raw, "dropped the bomb"):
		return "bomb_dropped"
	case strings.Contains(raw, "Bomb_Begin_Plant"):
		return "bomb_begin_plant"
	case strings.Contains(raw, "Bomb_Planted"):
		return "bomb_planted_trigger"
	case strings.Contains(raw, "Bomb_Defused"):
		return "bomb_defused_trigger"
	
	// RCON commands - MUST check before mp_ settings
	case strings.Contains(raw, "rcon from"):
		return "rcon_command"
	
	// Server configuration events
	case strings.Contains(raw, "server_cvar"):
		return "server_cvar"
	case strings.Contains(raw, "\"mp_"):
		// Specific MP cvars
		if strings.Contains(raw, "mp_maxrounds") {
			return "cvar_maxrounds"
		}
		if strings.Contains(raw, "mp_overtime") {
			return "cvar_overtime"
		}
		if strings.Contains(raw, "mp_freezetime") {
			return "cvar_freezetime"
		}
		if strings.Contains(raw, "mp_tournament") {
			return "cvar_tournament"
		}
		return "cvar_mp_setting"
	
	// Log file events
	case strings.Contains(raw, "Log file"):
		if strings.Contains(raw, "started") {
			return "log_file_started"
		}
		if strings.Contains(raw, "closed") {
			return "log_file_closed"
		}
		return "log_file"
	
	// Map events
	case strings.Contains(raw, "Loading map"):
		return "loading_map"
	case strings.Contains(raw, "Started map"):
		return "started_map"
	
	// Team events
	case strings.Contains(raw, "Team playing"):
		return "team_playing"
	
	// Round events
	case strings.Contains(raw, "Starting Freeze period"):
		return "freeze_period_start"
	
	// Game end events
	case strings.Contains(raw, "Game Over"):
		// Extract game mode and score
		if strings.Contains(raw, "competitive") {
			return "game_over_competitive"
		}
		if strings.Contains(raw, "casual") {
			return "game_over_casual"
		}
		return "game_over"
	
	// Team notice events - Check before generic triggered events
	case strings.Contains(raw, "Team ") && strings.Contains(raw, "triggered"):
		if strings.Contains(raw, "SFUI_Notice") {
			return "team_notice"
		}
		return "team_triggered_event"
	
	// JSON Statistics dumps
	case strings.Contains(raw, "JSON_BEGIN{"):
		return "stats_json_start"
	case strings.Contains(raw, "}}JSON_END"):
		return "stats_json_end"
	case strings.Contains(raw, "\"player_") && strings.Contains(raw, ":"):
		// Player statistics line (part of JSON dump)
		return "stats_player_data"
	case strings.Contains(raw, "\"players\"") && strings.Contains(raw, ":"):
		return "stats_json_players"
	case strings.Contains(raw, "\"fields\"") && strings.Contains(raw, ":"):
		return "stats_json_fields"
	case strings.Contains(raw, "\"server\"") && strings.Contains(raw, ":"):
		return "stats_json_server"
	case strings.Contains(raw, "\"score_ct\"") && strings.Contains(raw, ":"):
		return "stats_json_score_ct"
	case strings.Contains(raw, "\"score_t\"") && strings.Contains(raw, ":"):
		return "stats_json_score_t"
	case strings.Contains(raw, "\"rounds_played\"") && strings.Contains(raw, ":"):
		return "stats_json_rounds"
	case strings.Contains(raw, "\"version\"") && strings.Contains(raw, ":"):
		return "stats_json_version"
	case strings.Contains(raw, "\"timestamp\"") && strings.Contains(raw, ":"):
		return "stats_json_timestamp"
	case strings.Contains(raw, "\"map\"") && strings.Contains(raw, ":"):
		return "stats_json_map"
	
	// Chat/Say events - very common
	case strings.Contains(raw, "\" say \""):
		// Extract the message
		msgStart := strings.Index(raw, "\" say \"") + 7
		msgEnd := strings.LastIndex(raw[msgStart:], "\"")
		if msgEnd > 0 {
			msg := raw[msgStart:msgStart+msgEnd]
			
			// Check for common chat commands
			switch {
			case strings.HasPrefix(msg, ".pause") || strings.HasPrefix(msg, ".forcepause"):
				return "chat_pause_command"
			case strings.HasPrefix(msg, ".restore") || strings.HasPrefix(msg, ".resotre"):
				return "chat_restore_command"  
			case strings.HasPrefix(msg, ".ready") || strings.HasPrefix(msg, ".rdy"):
				return "chat_ready_command"
			case strings.HasPrefix(msg, ".unpause"):
				return "chat_unpause_command"
			case strings.HasPrefix(msg, ".tech"):
				return "chat_tech_command"
			case strings.HasPrefix(msg, ".tac"):
				return "chat_tac_command"
			case strings.HasPrefix(msg, ".asay"):
				return "chat_admin_say"
			case strings.HasPrefix(msg, "."):
				// Generic command
				return "chat_command"
			case msg == "gg" || msg == "gg wp":
				return "chat_gg"
			default:
				return "chat_message"
			}
		}
		return "chat_message"
	
	// Triggered events (generic handler)
	case strings.Contains(raw, "triggered \""):
		start := strings.Index(raw, "triggered \"") + 11
		end := strings.Index(raw[start:], "\"")
		if end > 0 {
			event := raw[start:start+end]
			// Special cases for specific triggered events
			if event == "Round_Freeze_End" {
				return "freeze_time_start"  // Freeze time ends, action starts
			}
			// Normalize event names for generic triggers
			event = strings.ToLower(event)
			event = strings.ReplaceAll(event, "_", "-")
			event = strings.ReplaceAll(event, " ", "-")
			return "trigger_" + event
		}
		return "triggered_event"
	
	default:
		// Final fallback
		return "unknown_other"
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