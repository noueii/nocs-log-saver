package main

import (
	"fmt"
	"strings"
	cs2log "github.com/janstuemmel/cs2-log"
)

func main() {
	// Test various blind event formats from the logs
	testLogs := []string{
		`L 08/19/2025 - 19:02:50: "SHESKY<7><[U:1:215888626]><TERRORIST>" blinded for 5.09 by "SHESKY<7><[U:1:215888626]><TERRORIST>" from flashbang entindex 225`,
		`L 08/19/2025 - 19:02:50: "NxS Sebo<6><[U:1:387734521]><TERRORIST>" blinded for 3.15 by "SHESKY<7><[U:1:215888626]><TERRORIST>" from flashbang entindex 225`,
		`L 08/19/2025 - 19:02:38: "SHESKY<7><[U:1:215888626]><TERRORIST>" blinded for 1.03 by "NxS Sebo<6><[U:1:387734521]><TERRORIST>" from flashbang entindex 465`,
		`L 08/19/2025 - 19:02:50: "SHESKY<7><[U:1:215888626]><TERRORIST>" threw flashbang [-393 1644 -42] flashbang entindex 225)`,
		`L 08/19/2025 - 19:02:38: "NxS Sebo<6><[U:1:387734521]><TERRORIST>" threw flashbang [-308 1032 52] flashbang entindex 465)`,
	}

	fmt.Println("=== Testing Blind/Flash Events ===\n")

	for i, logLine := range testLogs {
		fmt.Printf("Test %d:\n", i+1)
		fmt.Printf("Input: %s\n", logLine)
		
		// Parse the log
		parsed, err := cs2log.Parse(logLine)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			// Check the type
			eventType := fmt.Sprintf("%T", parsed)
			fmt.Printf("Parsed Type: %s\n", eventType)
			
			// Check if it's Unknown
			if unknown, ok := parsed.(cs2log.Unknown); ok {
				fmt.Printf("Unknown Raw: %s\n", unknown.Raw)
				
				// Try to classify it
				classification := classifyBlindEvent(unknown.Raw)
				fmt.Printf("Classification: %s\n", classification)
			} else if blinded, ok := parsed.(cs2log.PlayerBlinded); ok {
				fmt.Printf("PlayerBlinded: Victim=%s, Attacker=%s, Duration=%.2f\n", 
					blinded.Victim.Name, blinded.Attacker.Name, blinded.For)
			}
			
			// Convert to JSON
			jsonData := cs2log.ToJSON(parsed)
			fmt.Printf("JSON: %s\n", jsonData)
		}
		fmt.Println()
	}
}

func classifyBlindEvent(raw string) string {
	switch {
	case strings.Contains(raw, "blinded for"):
		if strings.Contains(raw, "by") {
			return "player_blinded"
		}
		return "blind_event"
	case strings.Contains(raw, "threw flashbang"):
		return "flashbang_thrown"
	case strings.Contains(raw, "sv_throw_flashgrenade"):
		return "debug_throw_flash"
	default:
		return "unknown"
	}
}