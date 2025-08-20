package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

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

	// Query all unknown events with their raw content
	query := `
		SELECT 
			p.event_type,
			p.event_data,
			r.content,
			COUNT(*) OVER (PARTITION BY p.event_type) as type_count
		FROM parsed_logs p
		JOIN raw_logs r ON p.raw_log_id = r.id
		WHERE p.event_type LIKE 'unknown%'
		ORDER BY type_count DESC, p.created_at DESC
		LIMIT 100
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Failed to query unknown events:", err)
	}
	defer rows.Close()

	// Track unique patterns
	patterns := make(map[string][]string)
	
	fmt.Println("=== Analyzing Unknown Event Types ===\n")

	for rows.Next() {
		var eventType string
		var eventData json.RawMessage
		var content string
		var typeCount int

		if err := rows.Scan(&eventType, &eventData, &content, &typeCount); err != nil {
			continue
		}

		// Parse event data
		var data map[string]interface{}
		json.Unmarshal(eventData, &data)

		// Extract the actual log content (remove prefixes)
		actualContent := extractActualContent(content)
		
		// Try to identify the pattern
		pattern := identifyPattern(actualContent)
		
		if _, exists := patterns[pattern]; !exists {
			patterns[pattern] = []string{}
			fmt.Printf("Pattern: %s\n", pattern)
			fmt.Printf("Event Type: %s (Count: %d)\n", eventType, typeCount)
			fmt.Printf("Raw Log: %s\n", actualContent)
			fmt.Printf("Event Data: %s\n", string(eventData))
			fmt.Println(strings.Repeat("-", 80))
		}
		
		patterns[pattern] = append(patterns[pattern], actualContent)
	}

	// Summary
	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Found %d unique unknown event patterns\n", len(patterns))
	
	// Show pattern counts
	fmt.Printf("\n=== Pattern Counts ===\n")
	for pattern, examples := range patterns {
		fmt.Printf("%s: %d occurrences\n", pattern, len(examples))
		if len(examples) > 0 && len(examples) <= 3 {
			for _, ex := range examples {
				fmt.Printf("  Example: %s\n", ex)
			}
		}
	}
}

func extractActualContent(content string) string {
	// Remove timestamp and UUID prefix
	if strings.HasPrefix(content, "[") {
		if endIdx := strings.Index(content, "] "); endIdx != -1 {
			remaining := content[endIdx+2:]
			if colonIdx := strings.Index(remaining, ": "); colonIdx != -1 {
				content = remaining[colonIdx+2:]
			}
		}
	}
	
	// Remove UUID-only prefix
	if !strings.HasPrefix(content, "L ") && strings.Contains(content, ": ") {
		colonIdx := strings.Index(content, ": ")
		possibleUUID := content[:colonIdx]
		if len(possibleUUID) == 36 && strings.Count(possibleUUID, "-") == 4 {
			content = content[colonIdx+2:]
		}
	}
	
	// Remove "L " prefix and timestamp
	if strings.HasPrefix(content, "L ") {
		parts := strings.SplitN(content, " - ", 3)
		if len(parts) >= 3 {
			// Remove milliseconds if present
			timePart := parts[1]
			if dotIdx := strings.Index(timePart, "."); dotIdx != -1 {
				timePart = timePart[:dotIdx]
			}
			// Return just the actual log content
			return parts[2]
		}
	}
	
	return content
}

func identifyPattern(log string) string {
	// Common patterns in CS2 logs
	switch {
	case strings.Contains(log, "killed other"):
		return "killed_other"
	case strings.Contains(log, "assisted killing"):
		return "kill_assist"
	case strings.Contains(log, "blinded"):
		return "player_blinded"
	case strings.Contains(log, "money change"):
		return "money_change"
	case strings.Contains(log, "left buyzone"):
		return "left_buyzone"
	case strings.Contains(log, "entered the game"):
		return "entered_game"
	case strings.Contains(log, "switched from team"):
		return "team_switch"
	case strings.Contains(log, "STEAM USERID validated"):
		return "userid_validated"
	case strings.Contains(log, "committed suicide"):
		return "suicide"
	case strings.Contains(log, "was kicked"):
		return "player_kicked"
	case strings.Contains(log, "triggered"):
		return "world_trigger"
	case strings.Contains(log, "scored"):
		return "team_scored"
	case strings.Contains(log, "say_team"):
		return "team_chat"
	case strings.Contains(log, "changed name"):
		return "name_change"
	default:
		// Try to extract the action from the log
		if strings.Contains(log, "\" ") {
			parts := strings.Split(log, "\" ")
			if len(parts) > 1 {
				// Get the action part
				action := strings.TrimSpace(parts[1])
				if spaceIdx := strings.Index(action, " "); spaceIdx != -1 {
					return action[:spaceIdx]
				}
				return action
			}
		}
		return "unidentified"
	}
}