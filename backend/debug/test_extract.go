package main

import (
	"fmt"
	"strings"

	cs2log "github.com/janstuemmel/cs2-log"
)

func main() {
	// Test the extractActualContent function
	testLogs := []string{
		`[2025-08-19T16:03:52Z] 18a5c248-c891-42a6-b72e-af0b184937c1: 08/19/2025 - 19:04:10.830 - "NxS Sebo<6><[U:1:387734521]><TERRORIST>" disconnected (reason "NETWORK_DISCONNECT_DISCONNECT_BY_USER")`,
		`08/19/2025 - 19:03:31.480 - "alker007<8><[U:1:869707820]><CT>" [-1987 1958 0] killed "NxS Sebo<6><[U:1:387734521]><TERRORIST>" [-1946 1416 88] with "m4a1_silencer"`,
	}

	for i, log := range testLogs {
		fmt.Printf("\n=== Test %d ===\n", i+1)
		fmt.Printf("Original: %.100s...\n", log)
		
		// Apply extraction logic
		extracted := extractActualContent(log)
		fmt.Printf("Extracted: %s\n", extracted)
		
		// Try to parse
		parsed, err := cs2log.Parse(extracted)
		if err != nil {
			fmt.Printf("❌ Parse error: %v\n", err)
		} else {
			fmt.Printf("✅ Parsed successfully as %T\n", parsed)
		}
	}
}

func extractActualContent(content string) string {
	actualContent := content
	
	// Remove timestamp and UUID prefix
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