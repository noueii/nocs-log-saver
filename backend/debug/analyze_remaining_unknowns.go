package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
)

type ParseTestRequest struct {
	Logs string `json:"logs"`
}

type ParseTestResult struct {
	LineNumber int         `json:"line_number"`
	Content    string      `json:"content"`
	Success    bool        `json:"success"`
	EventType  string      `json:"event_type,omitempty"`
	EventData  interface{} `json:"event_data,omitempty"`
	Error      string      `json:"error,omitempty"`
}

type ParseTestResponse struct {
	TotalLines  int               `json:"total_lines"`
	ParsedCount int               `json:"parsed_count"`
	FailedCount int               `json:"failed_count"`
	Results     []ParseTestResult `json:"results"`
}

func main() {
	fmt.Println("=== Analyzing Remaining Unknown Events ===")
	
	// Read the entire match-test.txt file
	content, err := os.ReadFile("match-test.txt")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}
	
	lines := strings.Split(string(content), "\n")
	
	// Process in batches to avoid overwhelming the API
	batchSize := 500
	allUnknowns := make(map[string][]string) // event_type -> examples
	totalProcessed := 0
	
	for i := 0; i < len(lines); i += batchSize {
		end := i + batchSize
		if end > len(lines) {
			end = len(lines)
		}
		
		batch := lines[i:end]
		batchContent := strings.Join(batch, "\n")
		
		// Send to parse-test API
		req := ParseTestRequest{
			Logs: batchContent,
		}
		
		jsonData, err := json.Marshal(req)
		if err != nil {
			continue
		}
		
		resp, err := http.Post("http://localhost:9090/api/parse-test", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			continue
		}
		
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}
		
		var parseResp ParseTestResponse
		if err := json.Unmarshal(body, &parseResp); err != nil {
			continue
		}
		
		// Collect unknown events
		for _, result := range parseResp.Results {
			if result.Success && strings.Contains(result.EventType, "unknown") {
				// Extract the actual content without prefixes
				content := extractContent(result.Content)
				
				if _, exists := allUnknowns[result.EventType]; !exists {
					allUnknowns[result.EventType] = []string{}
				}
				
				// Store up to 5 examples per type
				if len(allUnknowns[result.EventType]) < 5 {
					allUnknowns[result.EventType] = append(allUnknowns[result.EventType], content)
				}
			}
		}
		
		totalProcessed += parseResp.TotalLines
		fmt.Printf("Processed %d/%d lines...\n", totalProcessed, len(lines))
	}
	
	// Analyze the unknown events
	fmt.Printf("\n=== Unknown Event Analysis ===\n")
	fmt.Printf("Total unique unknown event types: %d\n\n", len(allUnknowns))
	
	// Sort by frequency
	type unknownCount struct {
		EventType string
		Count     int
		Examples  []string
	}
	
	var unknownCounts []unknownCount
	for eventType, examples := range allUnknowns {
		unknownCounts = append(unknownCounts, unknownCount{
			EventType: eventType,
			Count:     len(examples),
			Examples:  examples,
		})
	}
	
	sort.Slice(unknownCounts, func(i, j int) bool {
		return unknownCounts[i].Count > unknownCounts[j].Count
	})
	
	// Now let's analyze the actual content of unknown_other events
	if examples, exists := allUnknowns["unknown_other"]; exists {
		fmt.Println("=== Analyzing unknown_other events ===")
		
		// Categorize by patterns
		patterns := make(map[string][]string)
		
		for _, example := range examples {
			category := categorizeUnknown(example)
			if _, exists := patterns[category]; !exists {
				patterns[category] = []string{}
			}
			if len(patterns[category]) < 3 {
				patterns[category] = append(patterns[category], example)
			}
		}
		
		// Display patterns
		fmt.Printf("\nFound %d different patterns in unknown_other:\n\n", len(patterns))
		
		for pattern, examples := range patterns {
			fmt.Printf("Pattern: %s\n", pattern)
			for i, ex := range examples {
				// Truncate long examples
				if len(ex) > 150 {
					ex = ex[:150] + "..."
				}
				fmt.Printf("  %d. %s\n", i+1, ex)
			}
			fmt.Println()
		}
	}
	
	// Also check chat_message events to see what kind of chat we're getting
	fmt.Println("\n=== Analyzing chat patterns ===")
	chatPatterns := analyzeChatPatterns()
	for pattern, count := range chatPatterns {
		fmt.Printf("%s: %d occurrences\n", pattern, count)
	}
}

func extractContent(line string) string {
	// Remove timestamp prefixes
	if idx := strings.Index(line, ": L "); idx != -1 {
		line = line[idx+2:]
	}
	
	// Remove date/time from CS2 log
	if strings.HasPrefix(line, "L ") {
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) > 1 {
			return parts[1]
		}
	}
	
	return line
}

func categorizeUnknown(content string) string {
	// Try to identify patterns in unknown content
	switch {
	case strings.Contains(content, "\" say_team \""):
		return "team_chat"
	case strings.Contains(content, "\" say \""):
		// Further categorize chat
		if strings.Contains(content, ".") && !strings.Contains(content, " ") {
			return "chat_command"
		}
		if len(content) < 50 {
			return "chat_short"
		}
		return "chat_long"
	case strings.Contains(content, "committed suicide"):
		return "suicide_event"
	case strings.Contains(content, "killed other"):
		return "killed_other"
	case strings.Contains(content, "killed"):
		return "kill_event"
	case strings.Contains(content, "attacked"):
		return "attack_event"
	case strings.Contains(content, "assisted killing"):
		return "assist_event"
	case strings.Contains(content, "purchased"):
		return "purchase_event"
	case strings.Contains(content, "threw"):
		return "throw_event"
	case strings.Contains(content, "blinded"):
		return "blind_event"
	case strings.Contains(content, "money change"):
		return "money_event"
	case strings.Contains(content, "picked up"):
		return "pickup_event"
	case strings.Contains(content, "dropped"):
		return "drop_event"
	case strings.Contains(content, "connected"):
		return "connect_event"
	case strings.Contains(content, "disconnected"):
		return "disconnect_event"
	case strings.Contains(content, "entered the game"):
		return "enter_event"
	case strings.Contains(content, "switched from team"):
		return "team_switch_event"
	case strings.Contains(content, "triggered"):
		return "triggered_event"
	case strings.Contains(content, "World triggered"):
		return "world_triggered_event"
	case strings.Contains(content, "Team "):
		return "team_event"
	default:
		// Try to get first word or pattern
		words := strings.Fields(content)
		if len(words) > 0 {
			return "starts_with_" + strings.ToLower(words[0])
		}
		return "unidentified"
	}
}

func analyzeChatPatterns() map[string]int {
	// This would analyze actual chat content from the API
	// For now, return a placeholder
	return map[string]int{
		"general_chat": 0,
		"commands": 0,
		"gg_messages": 0,
	}
}