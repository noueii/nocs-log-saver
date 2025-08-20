package main

import (
	"fmt"
	"strings"
	cs2log "github.com/janstuemmel/cs2-log"
)

func main() {
	// Test with actual content from the logs INCLUDING the prefix
	testLogs := []string{
		`[2025-08-19T16:02:32Z] 18a5c248-c891-42a6-b72e-af0b184937c1: 08/19/2025 - 19:02:50.839 - "SHESKY<7><[U:1:215888626]><TERRORIST>" blinded for 5.09 by "SHESKY<7><Ўs֬><TERRORIST>" from flashbang entindex 225`,
		`08/19/2025 - 19:02:50.839 - "SHESKY<7><[U:1:215888626]><TERRORIST>" blinded for 5.09 by "SHESKY<7><Ўs֬><TERRORIST>" from flashbang entindex 225`,
		`L 08/19/2025 - 19:02:50.839 - "SHESKY<7><[U:1:215888626]><TERRORIST>" blinded for 5.09 by "SHESKY<7><Ўs֬><TERRORIST>" from flashbang entindex 225`,
		`L 08/19/2025 - 19:02:50: "SHESKY<7><[U:1:215888626]><TERRORIST>" blinded for 5.09 by "SHESKY<7><Ўs֬><TERRORIST>" from flashbang entindex 225`,
	}

	fmt.Println("=== Testing Actual Blind Event Formats ===\n")

	for i, logLine := range testLogs {
		fmt.Printf("Test %d:\n", i+1)
		fmt.Printf("Input: %s\n", logLine)
		
		// Try extracting actual content first
		actualContent := extractActualContent(logLine)
		fmt.Printf("Extracted: %s\n", actualContent)
		
		// Parse the log
		parsed, err := cs2log.Parse(actualContent)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			// Check the type
			eventType := fmt.Sprintf("%T", parsed)
			fmt.Printf("Parsed Type: %s\n", eventType)
			
			// Check if it's Unknown
			if unknown, ok := parsed.(cs2log.Unknown); ok {
				fmt.Printf("Unknown Raw: %s\n", unknown.Raw)
			} else if _, ok := parsed.(cs2log.PlayerBlinded); ok {
				fmt.Printf("SUCCESS: PlayerBlinded parsed correctly\n")
			}
		}
		fmt.Println()
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