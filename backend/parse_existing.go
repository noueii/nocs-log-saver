package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	cs2log "github.com/janstuemmel/cs2-log"
	_ "github.com/lib/pq"
)

func main() {
	// Get database URL from environment or use default
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://cs2admin:localpass123@localhost:5432/cs2logs?sslmode=disable"
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Query unparsed raw logs
	query := `
		SELECT r.id, r.server_id, r.content 
		FROM raw_logs r
		LEFT JOIN parsed_logs p ON r.id = p.raw_log_id
		LEFT JOIN failed_parses f ON r.id = f.raw_log_id
		WHERE p.id IS NULL AND f.id IS NULL
		LIMIT 100
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Failed to query raw logs:", err)
	}
	defer rows.Close()

	eventTypes := make(map[string]int)
	unknownExamples := make(map[string][]string)
	failedExamples := []string{}
	
	totalCount := 0
	parsedCount := 0
	failedCount := 0

	for rows.Next() {
		var id, serverID, content string
		if err := rows.Scan(&id, &serverID, &content); err != nil {
			continue
		}
		
		totalCount++
		
		// Process the log
		actualContent := extractActualContent(content)
		
		// Try to parse
		parsed, err := cs2log.Parse(actualContent)
		if err != nil {
			failedCount++
			if len(failedExamples) < 10 {
				failedExamples = append(failedExamples, actualContent)
			}
			
			// Store as failed parse
			insertQuery := `
				INSERT INTO failed_parses (raw_log_id, error_message, created_at)
				VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
			`
			db.Exec(insertQuery, id, err.Error(), time.Now())
			continue
		}
		
		parsedCount++
		
		// Get event type
		eventType := getEventType(parsed)
		eventTypes[eventType]++
		
		// Collect unknown examples
		if strings.HasPrefix(eventType, "unknown") {
			if examples, ok := unknownExamples[eventType]; !ok || len(examples) < 3 {
				unknownExamples[eventType] = append(unknownExamples[eventType], actualContent)
			}
		}
		
		// Store parsed log
		eventData := cs2log.ToJSON(parsed)
		insertQuery := `
			INSERT INTO parsed_logs (raw_log_id, server_id, event_type, event_data, created_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT DO NOTHING
		`
		db.Exec(insertQuery, id, serverID, eventType, eventData, time.Now())
	}

	// Print results
	fmt.Println("=== Parsing Results ===")
	fmt.Printf("Total processed: %d\n", totalCount)
	fmt.Printf("Successfully parsed: %d\n", parsedCount)
	fmt.Printf("Failed to parse: %d\n\n", failedCount)

	fmt.Println("=== Event Type Distribution ===")
	for eventType, count := range eventTypes {
		fmt.Printf("%s: %d\n", eventType, count)
	}

	if len(unknownExamples) > 0 {
		fmt.Println("\n=== Unknown Event Examples ===")
		for eventType, examples := range unknownExamples {
			fmt.Printf("\n%s (%d examples):\n", eventType, len(examples))
			for i, example := range examples {
				fmt.Printf("  %d. %s\n", i+1, example)
			}
		}
	}

	if len(failedExamples) > 0 {
		fmt.Println("\n=== Failed Parse Examples ===")
		for i, example := range failedExamples {
			fmt.Printf("%d. %s\n", i+1, example)
		}
	}
}

func extractActualContent(content string) string {
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

func getEventType(parsedLog cs2log.Message) string {
	// The cs2-log library returns different types for different events
	// We need to type switch to determine the event
	switch msg := parsedLog.(type) {
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
		// For Unknown type, try to get more info from the Raw field
		if msg.Raw != "" {
			// Try to identify what kind of unknown event this is
			switch {
			case strings.Contains(msg.Raw, "killed other"):
				return "unknown_killed_other"
			case strings.Contains(msg.Raw, "assisted killing"):
				return "unknown_kill_assist"
			case strings.Contains(msg.Raw, "blinded"):
				return "unknown_blinded"
			case strings.Contains(msg.Raw, "money change"):
				return "unknown_money_change"
			case strings.Contains(msg.Raw, "left buyzone"):
				return "unknown_left_buyzone"
			case strings.Contains(msg.Raw, "entered the game"):
				return "unknown_entered_game"
			case strings.Contains(msg.Raw, "committed suicide"):
				return "unknown_suicide"
			case strings.Contains(msg.Raw, "server_cvar"):
				return "unknown_server_cvar"
			case strings.Contains(msg.Raw, "Log file"):
				return "unknown_log_file"
			case strings.Contains(msg.Raw, "STEAM USERID validated"):
				return "unknown_userid_validated"
			default:
				return "unknown_other"
			}
		}
		return "unknown"
	default:
		return fmt.Sprintf("unknown_%T", parsedLog)
	}
}