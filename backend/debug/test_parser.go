package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	cs2log "github.com/janstuemmel/cs2-log"
)

type ParseResult struct {
	LineNumber int         `json:"line_number"`
	Content    string      `json:"content"`
	Success    bool        `json:"success"`
	EventType  string      `json:"event_type,omitempty"`
	EventData  interface{} `json:"event_data,omitempty"`
	Error      string      `json:"error,omitempty"`
}

type ParseSummary struct {
	TotalLines   int                    `json:"total_lines"`
	ParsedCount  int                    `json:"parsed_count"`
	FailedCount  int                    `json:"failed_count"`
	UnknownCount int                    `json:"unknown_count"`
	EventTypes   map[string]int         `json:"event_types"`
	UnknownTypes map[string][]string    `json:"unknown_types"`
	FailedLogs   []string               `json:"failed_logs"`
}

func main() {
	// Read the match test file
	filePath := filepath.Join("debug", "match-test.txt")
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var results []ParseResult
	summary := ParseSummary{
		EventTypes:   make(map[string]int),
		UnknownTypes: make(map[string][]string),
		FailedLogs:   []string{},
	}

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		summary.TotalLines++
		
		// Extract and process the log
		actualContent := extractActualContent(line)
		result := ParseResult{
			LineNumber: lineNumber,
			Content:    actualContent,
		}
		
		// Try to parse with cs2-log library
		parsed, err := cs2log.Parse(actualContent)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			summary.FailedCount++
			if len(summary.FailedLogs) < 10 {
				summary.FailedLogs = append(summary.FailedLogs, actualContent)
			}
		} else {
			result.Success = true
			result.EventType = getEventType(parsed)
			result.EventData = cs2log.ToJSON(parsed)
			summary.ParsedCount++
			summary.EventTypes[result.EventType]++
			
			// Track unknown events
			if strings.HasPrefix(result.EventType, "unknown") {
				summary.UnknownCount++
				if examples, ok := summary.UnknownTypes[result.EventType]; !ok || len(examples) < 3 {
					summary.UnknownTypes[result.EventType] = append(summary.UnknownTypes[result.EventType], actualContent)
				}
			}
		}
		
		results = append(results, result)
	}

	// Print detailed results
	fmt.Println("=== CS2 Log Parser Test Results ===")
	fmt.Printf("File: %s\n", filePath)
	fmt.Printf("Total Lines: %d\n", summary.TotalLines)
	fmt.Printf("Successfully Parsed: %d (%.1f%%)\n", summary.ParsedCount, float64(summary.ParsedCount)*100/float64(summary.TotalLines))
	fmt.Printf("Failed to Parse: %d (%.1f%%)\n", summary.FailedCount, float64(summary.FailedCount)*100/float64(summary.TotalLines))
	fmt.Printf("Unknown Events: %d\n\n", summary.UnknownCount)

	// Event type distribution
	fmt.Println("=== Event Type Distribution ===")
	for eventType, count := range summary.EventTypes {
		fmt.Printf("%-30s: %4d\n", eventType, count)
	}

	// Unknown event examples
	if len(summary.UnknownTypes) > 0 {
		fmt.Println("\n=== Unknown Event Types (Need Custom Parsing) ===")
		for eventType, examples := range summary.UnknownTypes {
			fmt.Printf("\n%s (%d examples):\n", eventType, len(examples))
			for i, example := range examples {
				fmt.Printf("  %d. %s\n", i+1, example)
				
				// Try to identify the pattern
				pattern := identifyPattern(example)
				if pattern != "" {
					fmt.Printf("     Pattern: %s\n", pattern)
				}
			}
		}
	}

	// Failed parse examples
	if len(summary.FailedLogs) > 0 {
		fmt.Println("\n=== Failed to Parse (First 10) ===")
		for i, log := range summary.FailedLogs {
			fmt.Printf("%d. %s\n", i+1, log)
		}
	}

	// Save results to JSON file
	outputPath := filepath.Join("debug", "parse_results.json")
	outputFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer outputFile.Close()

	encoder := json.NewEncoder(outputFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(map[string]interface{}{
		"summary": summary,
		"results": results,
		"timestamp": time.Now().Format(time.RFC3339),
	}); err != nil {
		fmt.Printf("Error writing JSON: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Results saved to %s\n", outputPath)
}

func extractActualContent(content string) string {
	actualContent := content
	
	// Remove timestamp and UUID prefix
	// Format: [2025-08-19T15:12:44Z] 18a5c248-c891-42a6-b72e-af0b184937c1: actual_log_content
	if strings.HasPrefix(content, "[") {
		if endIdx := strings.Index(content, "] "); endIdx != -1 {
			remaining := content[endIdx+2:]
			if colonIdx := strings.Index(remaining, ": "); colonIdx != -1 {
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
	
	// Remove milliseconds from timestamp if present
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
	case *cs2log.PlayerPickedUp:
		return "pickup"
	case *cs2log.PlayerDropped:
		return "dropped"
	case *cs2log.PlayerBlinded:
		return "blinded"
	case *cs2log.PlayerKilledBomb:
		return "killed_by_bomb"
	case *cs2log.PlayerKilledSuicide:
		return "suicide"
	case *cs2log.PlayerBombBeginDefuse:
		return "begin_defuse"
	case *cs2log.PlayerBombGot:
		return "got_bomb"
	case *cs2log.PlayerBombDropped:
		return "dropped_bomb"
	case *cs2log.PlayerMoneyChange:
		return "money_change"
	case *cs2log.TeamScored:
		return "team_scored"
	case *cs2log.Unknown:
		// For Unknown type, try to get more info
		if msg.Raw != "" {
			return classifyUnknownEvent(msg.Raw)
		}
		return "unknown"
	default:
		return fmt.Sprintf("unknown_%T", parsedLog)
	}
}

func classifyUnknownEvent(raw string) string {
	// Classify unknown events based on content
	switch {
	case strings.Contains(raw, "money change"):
		return "unknown_money_change"
	case strings.Contains(raw, "left buyzone"):
		return "unknown_left_buyzone"
	case strings.Contains(raw, "entered the game"):
		return "unknown_entered_game"
	case strings.Contains(raw, "STEAM USERID validated"):
		return "unknown_userid_validated"
	case strings.Contains(raw, "committed suicide"):
		return "unknown_suicide"
	case strings.Contains(raw, "was killed by the bomb"):
		return "unknown_bomb_kill"
	case strings.Contains(raw, "Begin_Bomb_Defuse"):
		return "unknown_begin_defuse"
	case strings.Contains(raw, "Bomb_Begin_Plant"):
		return "unknown_begin_plant"
	case strings.Contains(raw, "ACCOLADE"):
		return "unknown_accolade"
	case strings.Contains(raw, "MatchStatus"):
		return "unknown_match_status"
	case strings.Contains(raw, "Match pause"):
		return "unknown_match_pause"
	case strings.Contains(raw, "projectile spawned"):
		return "unknown_projectile"
	case strings.Contains(raw, "sv_throw"):
		return "unknown_throw_debug"
	default:
		return "unknown_other"
	}
}

func identifyPattern(log string) string {
	// Try to identify common patterns for custom parsing
	switch {
	case strings.Contains(log, "money change"):
		return "MONEY_CHANGE: player money change amount+change = $total"
	case strings.Contains(log, "left buyzone"):
		return "LEFT_BUYZONE: player left buyzone with [items]"
	case strings.Contains(log, "entered the game"):
		return "ENTERED_GAME: player entered the game"
	case strings.Contains(log, "STEAM USERID validated"):
		return "USERID_VALIDATED: player STEAM USERID validated"
	case strings.Contains(log, "Begin_Bomb_Defuse"):
		return "BEGIN_DEFUSE: player triggered Begin_Bomb_Defuse"
	case strings.Contains(log, "Bomb_Begin_Plant"):
		return "BEGIN_PLANT: player triggered Bomb_Begin_Plant at bombsite"
	case strings.Contains(log, "ACCOLADE"):
		return "ACCOLADE: achievement/award given to player"
	case strings.Contains(log, "MatchStatus"):
		return "MATCH_STATUS: match status update"
	case strings.Contains(log, "Match pause"):
		return "MATCH_PAUSE: match pause/unpause event"
	case strings.Contains(log, "projectile spawned"):
		return "PROJECTILE: grenade/projectile spawn event"
	case strings.Contains(log, "sv_throw"):
		return "THROW_DEBUG: debug info for thrown grenades"
	default:
		return ""
	}
}