package main

import (
	"fmt"
	"os"
	"strings"
	cs2log "github.com/janstuemmel/cs2-log"
)

func main() {
	// Read first 100 lines to check what's being parsed as unknown
	content, err := os.ReadFile("match-test.txt")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	lines := strings.Split(string(content), "\n")
	
	unknownCount := 0
	parsedCount := 0
	categories := make(map[string]int)
	unknownExamples := make(map[string][]string)
	
	for i, line := range lines {
		if i >= 7210 || strings.TrimSpace(line) == "" {
			break
		}
		
		// Extract actual CS2 log content
		actualContent := extractActualContent(line)
		
		// Try to parse
		parsed, err := cs2log.Parse(actualContent)
		if err != nil {
			continue
		}
		
		parsedCount++
		
		// Check type
		eventType := fmt.Sprintf("%T", parsed)
		categories[eventType]++
		
		// If it's Unknown, collect more info
		if unknown, ok := parsed.(cs2log.Unknown); ok {
			unknownCount++
			
			// Categorize the unknown
			category := categorizeUnknown(unknown.Raw)
			
			if _, exists := unknownExamples[category]; !exists {
				unknownExamples[category] = []string{}
			}
			
			if len(unknownExamples[category]) < 3 {
				unknownExamples[category] = append(unknownExamples[category], unknown.Raw)
			}
		}
	}
	
	fmt.Printf("=== Parse Statistics ===\n")
	fmt.Printf("Total lines: %d\n", len(lines))
	fmt.Printf("Parsed: %d\n", parsedCount)
	fmt.Printf("Unknown type: %d\n", unknownCount)
	fmt.Printf("Unknown percentage: %.1f%%\n\n", float64(unknownCount)/float64(parsedCount)*100)
	
	fmt.Printf("=== Event Type Distribution ===\n")
	for eventType, count := range categories {
		percentage := float64(count) / float64(parsedCount) * 100
		fmt.Printf("%-40s: %5d (%.1f%%)\n", eventType, count, percentage)
	}
	
	fmt.Printf("\n=== Unknown Event Categories ===\n")
	for category, examples := range unknownExamples {
		fmt.Printf("\n%s:\n", category)
		for i, ex := range examples {
			if len(ex) > 120 {
				ex = ex[:120] + "..."
			}
			fmt.Printf("  %d. %s\n", i+1, ex)
		}
	}
}

func extractActualContent(content string) string {
	actualContent := content
	
	// Check if the line has our custom prefix format
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
	
	// Ensure the log starts with "L "
	if !strings.HasPrefix(actualContent, "L ") {
		if strings.Contains(actualContent, " - ") && strings.Contains(actualContent, ":") {
			actualContent = "L " + actualContent
		}
	}
	
	// Remove milliseconds from timestamp
	if strings.HasPrefix(actualContent, "L ") {
		parts := strings.SplitN(actualContent, " - ", 3)
		if len(parts) >= 3 {
			timePart := parts[1]
			if dotIdx := strings.Index(timePart, "."); dotIdx != -1 {
				timePart = timePart[:dotIdx] + ":"
				actualContent = parts[0] + " - " + timePart + " " + parts[2]
			} else if !strings.HasSuffix(timePart, ":") {
				actualContent = parts[0] + " - " + timePart + ": " + parts[2]
			}
		}
	}
	
	return actualContent
}

func categorizeUnknown(raw string) string {
	switch {
	// Chat events
	case strings.Contains(raw, "\" say_team \""):
		return "team_chat"
	case strings.Contains(raw, "\" say \""):
		if strings.Contains(raw, "Spectator") {
			return "spectator_chat"
		}
		return "player_chat"
		
	// JSON data
	case strings.Contains(raw, "JSON_START") || strings.Contains(raw, "JSON_END"):
		return "json_data"
	case strings.Contains(raw, "\"player_"):
		return "player_stats_json"
		
	// Blind events that weren't caught
	case strings.Contains(raw, "blinded"):
		return "blind_event_uncaught"
		
	// Other patterns
	case strings.Contains(raw, "left buyzone"):
		return "buyzone_event"
	case strings.Contains(raw, "ACCOLADE"):
		return "accolade_event"
	case strings.Contains(raw, "MatchStatus"):
		return "match_status"
	case strings.Contains(raw, "sv_throw"):
		return "throw_debug"
	case strings.Contains(raw, "Game Over"):
		return "game_over"
	case strings.Contains(raw, "server_cvar"):
		return "server_cvar"
	case strings.Contains(raw, "rcon from"):
		return "rcon"
	case strings.Contains(raw, "Loading map"):
		return "map_loading"
	case strings.Contains(raw, "triggered"):
		return "triggered_event"
		
	default:
		// Get first few words
		words := strings.Fields(raw)
		if len(words) > 2 {
			return fmt.Sprintf("starts_with_%s_%s", words[0], words[1])
		} else if len(words) > 0 {
			return fmt.Sprintf("starts_with_%s", words[0])
		}
		return "empty_or_unknown"
	}
}