package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	cs2log "github.com/janstuemmel/cs2-log"
)

type UnknownEvent struct {
	Raw   string
	Count int
	Examples []string
}

func main() {
	// Read match-test.txt
	file, err := os.Open("match-test.txt")
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	unknownEvents := make(map[string]*UnknownEvent)
	totalLines := 0
	parsedCount := 0
	unknownCount := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		totalLines++

		// Extract actual CS2 log content
		actualContent := extractActualContent(line)
		
		// Try to parse
		parsed, err := cs2log.Parse(actualContent)
		if err != nil {
			continue
		}
		parsedCount++

		// Check if it's Unknown type
		if unknown, ok := parsed.(cs2log.Unknown); ok {
			unknownCount++
			
			// Categorize the unknown event
			category := categorizeUnknown(unknown.Raw)
			
			if _, exists := unknownEvents[category]; !exists {
				unknownEvents[category] = &UnknownEvent{
					Raw: category,
					Count: 0,
					Examples: []string{},
				}
			}
			
			unknownEvents[category].Count++
			if len(unknownEvents[category].Examples) < 3 {
				unknownEvents[category].Examples = append(unknownEvents[category].Examples, unknown.Raw)
			}
		}
	}

	fmt.Printf("=== Analysis Results ===\n")
	fmt.Printf("Total lines: %d\n", totalLines)
	fmt.Printf("Parsed: %d\n", parsedCount)
	fmt.Printf("Unknown events: %d\n\n", unknownCount)

	// Sort categories by count
	var categories []string
	for cat := range unknownEvents {
		categories = append(categories, cat)
	}
	sort.Slice(categories, func(i, j int) bool {
		return unknownEvents[categories[i]].Count > unknownEvents[categories[j]].Count
	})

	fmt.Printf("=== Unknown Event Categories ===\n")
	for _, cat := range categories {
		event := unknownEvents[cat]
		fmt.Printf("\n%s (Count: %d)\n", cat, event.Count)
		fmt.Printf("Description: %s\n", getDescription(cat))
		fmt.Printf("Examples:\n")
		for i, ex := range event.Examples {
			// Truncate long examples
			if len(ex) > 150 {
				ex = ex[:150] + "..."
			}
			fmt.Printf("  %d. %s\n", i+1, ex)
		}
	}

	// Generate code for improved classification
	fmt.Printf("\n\n=== Suggested Code for classifyUnknownEvent ===\n")
	generateClassificationCode(categories, unknownEvents)
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
	// Extract key patterns from unknown events
	switch {
	case strings.Contains(raw, "left buyzone"):
		return "left_buyzone"
	case strings.Contains(raw, "STEAM USERID validated"):
		return "userid_validated"
	case strings.Contains(raw, "ACCOLADE"):
		parts := strings.Split(raw, ",")
		if len(parts) > 1 {
			accoladeType := strings.TrimSpace(parts[1])
			return "accolade_" + strings.ToLower(accoladeType)
		}
		return "accolade"
	case strings.Contains(raw, "MatchStatus:"):
		return "match_status"
	case strings.Contains(raw, "Match pause"):
		return "match_pause"
	case strings.Contains(raw, "Match unpaused"):
		return "match_unpause"
	case strings.Contains(raw, "sv_throw"):
		return "throw_debug"
	case strings.Contains(raw, "Bomb_Begin_Plant"):
		return "bomb_begin_plant"
	case strings.Contains(raw, "Bomb_Planted"):
		return "bomb_planted_event"
	case strings.Contains(raw, "Bomb_Defused"):
		return "bomb_defused_event"
	case strings.Contains(raw, "server_cvar"):
		return "server_cvar"
	case strings.Contains(raw, "Log file"):
		if strings.Contains(raw, "started") {
			return "log_file_started"
		} else if strings.Contains(raw, "closed") {
			return "log_file_closed"
		}
		return "log_file"
	case strings.Contains(raw, "Starting Freeze period"):
		return "freeze_period_start"
	case strings.Contains(raw, "Game Over"):
		parts := strings.Split(raw, ":")
		if len(parts) > 1 {
			return "game_over_" + strings.ToLower(strings.TrimSpace(parts[1]))
		}
		return "game_over"
	case strings.Contains(raw, "triggered \""):
		start := strings.Index(raw, "triggered \"") + 11
		end := strings.Index(raw[start:], "\"")
		if end > 0 {
			event := strings.ToLower(raw[start:start+end])
			event = strings.ReplaceAll(event, "_", "-")
			return "trigger_" + event
		}
	case strings.Contains(raw, "Loading map"):
		return "loading_map"
	case strings.Contains(raw, "Started map"):
		return "started_map"
	case strings.Contains(raw, "rcon from"):
		return "rcon_command"
	case strings.Contains(raw, "mp_"):
		if strings.Contains(raw, "mp_maxrounds") {
			return "cvar_maxrounds"
		}
		if strings.Contains(raw, "mp_overtime") {
			return "cvar_overtime"
		}
		return "cvar_mp_setting"
	case strings.Contains(raw, "Team playing"):
		return "team_playing"
	case strings.Contains(raw, "scored"):
		if strings.Contains(raw, "CT scored") {
			return "team_ct_scored"
		}
		if strings.Contains(raw, "TERRORIST scored") {
			return "team_t_scored"
		}
		return "team_scored"
	}
	
	return "unknown_other"
}

func getDescription(category string) string {
	descriptions := map[string]string{
		"left_buyzone": "Player left the buy zone area",
		"userid_validated": "Steam user ID has been validated by the server",
		"accolade": "Player achievement or award during the match",
		"accolade_kills": "Kill-related achievement",
		"accolade_mvp": "MVP award",
		"accolade_utility": "Utility usage achievement",
		"match_status": "Current match status update with scores",
		"match_pause": "Match has been paused",
		"match_unpause": "Match has been unpaused",
		"throw_debug": "Debug information for grenade throws",
		"bomb_begin_plant": "Bomb planting has started",
		"bomb_planted_event": "Bomb has been successfully planted",
		"bomb_defused_event": "Bomb has been defused",
		"server_cvar": "Server console variable change",
		"log_file_started": "Log file recording has started",
		"log_file_closed": "Log file recording has stopped",
		"freeze_period_start": "Freeze time at round start",
		"game_over": "Match has ended",
		"trigger_": "Generic triggered event",
		"loading_map": "Server is loading a new map",
		"started_map": "Map has been loaded and started",
		"rcon_command": "Remote console command executed",
		"cvar_maxrounds": "Max rounds setting changed",
		"cvar_overtime": "Overtime settings changed",
		"cvar_mp_setting": "Multiplayer setting changed",
		"team_playing": "Team assignment for match",
		"team_ct_scored": "Counter-Terrorist team scored",
		"team_t_scored": "Terrorist team scored",
		"unknown_other": "Uncategorized event",
	}
	
	for key, desc := range descriptions {
		if strings.HasPrefix(category, key) {
			return desc
		}
	}
	
	return "Unknown event type"
}

func generateClassificationCode(categories []string, events map[string]*UnknownEvent) {
	fmt.Println("// Add these cases to classifyUnknownEvent function:")
	fmt.Println("func (s *ParserService) classifyUnknownEvent(raw string) string {")
	fmt.Println("\tswitch {")
	
	// Generate cases for each category
	for _, cat := range categories {
		if cat == "unknown_other" {
			continue
		}
		
		event := events[cat]
		if len(event.Examples) > 0 {
			// Find common pattern
			example := event.Examples[0]
			
			if strings.Contains(cat, "accolade") {
				fmt.Printf("\tcase strings.Contains(raw, \"ACCOLADE\"):\n")
				fmt.Printf("\t\t// %s\n", getDescription(cat))
				fmt.Printf("\t\tif strings.Contains(raw, \",\") {\n")
				fmt.Printf("\t\t\tparts := strings.Split(raw, \",\")\n")
				fmt.Printf("\t\t\tif len(parts) > 1 {\n")
				fmt.Printf("\t\t\t\taccoladeType := strings.ToLower(strings.TrimSpace(parts[1]))\n")
				fmt.Printf("\t\t\t\treturn \"accolade_\" + strings.ReplaceAll(accoladeType, \" \", \"_\")\n")
				fmt.Printf("\t\t\t}\n")
				fmt.Printf("\t\t}\n")
				fmt.Printf("\t\treturn \"accolade\"\n")
			} else if strings.Contains(cat, "trigger_") {
				// Already handled in existing code
				continue
			} else {
				// Find unique pattern
				pattern := findPattern(example, cat)
				if pattern != "" {
					fmt.Printf("\tcase strings.Contains(raw, \"%s\"):\n", pattern)
					fmt.Printf("\t\t// %s\n", getDescription(cat))
					fmt.Printf("\t\treturn \"%s\"\n", cat)
				}
			}
		}
	}
	
	fmt.Println("\tdefault:")
	fmt.Println("\t\treturn \"unknown_other\"")
	fmt.Println("\t}")
	fmt.Println("}")
}

func findPattern(example, category string) string {
	// Extract unique identifiers from examples
	patterns := map[string]string{
		"left_buyzone": "left buyzone",
		"userid_validated": "STEAM USERID validated",
		"match_status": "MatchStatus:",
		"match_pause": "Match pause",
		"match_unpause": "Match unpaused",
		"bomb_begin_plant": "Bomb_Begin_Plant",
		"server_cvar": "server_cvar",
		"log_file": "Log file",
		"freeze_period_start": "Starting Freeze period",
		"game_over": "Game Over",
		"loading_map": "Loading map",
		"started_map": "Started map",
		"rcon_command": "rcon from",
		"team_playing": "Team playing",
	}
	
	if pattern, exists := patterns[category]; exists {
		return pattern
	}
	
	return ""
}