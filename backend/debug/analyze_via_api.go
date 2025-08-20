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
	fmt.Println("=== Analyzing Unknown Events via Parse-Test API ===")
	
	// Read the entire match-test.txt file
	content, err := os.ReadFile("match-test.txt")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}
	
	// Send entire file to parse-test endpoint
	req := ParseTestRequest{
		Logs: string(content),
	}
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		return
	}
	
	fmt.Println("Sending request to parse-test API...")
	resp, err := http.Post("http://localhost:9090/api/parse-test", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}
	
	var parseResp ParseTestResponse
	if err := json.Unmarshal(body, &parseResp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}
	
	fmt.Printf("\n=== Parse Results ===\n")
	fmt.Printf("Total Lines: %d\n", parseResp.TotalLines)
	fmt.Printf("Parsed: %d\n", parseResp.ParsedCount)
	fmt.Printf("Failed: %d\n\n", parseResp.FailedCount)
	
	// Count event types
	eventCounts := make(map[string]int)
	unknownExamples := make(map[string][]string)
	
	for _, result := range parseResp.Results {
		if result.Success {
			eventCounts[result.EventType]++
			
			// Collect examples of unknown events
			if strings.Contains(result.EventType, "unknown") || 
			   strings.Contains(result.EventType, "chat") ||
			   strings.Contains(result.EventType, "trigger") {
				if _, exists := unknownExamples[result.EventType]; !exists {
					unknownExamples[result.EventType] = []string{}
				}
				if len(unknownExamples[result.EventType]) < 5 {
					// Extract the meaningful part
					content := result.Content
					if idx := strings.LastIndex(content, " - "); idx != -1 {
						content = content[idx+3:]
					}
					unknownExamples[result.EventType] = append(unknownExamples[result.EventType], content)
				}
			}
		}
	}
	
	// Sort event types by count
	type eventCount struct {
		EventType string
		Count     int
	}
	
	var sortedEvents []eventCount
	for eventType, count := range eventCounts {
		sortedEvents = append(sortedEvents, eventCount{EventType: eventType, Count: count})
	}
	
	sort.Slice(sortedEvents, func(i, j int) bool {
		return sortedEvents[i].Count > sortedEvents[j].Count
	})
	
	// Display top event types
	fmt.Println("=== Top Event Types ===")
	for i, event := range sortedEvents {
		if i < 30 { // Show top 30
			fmt.Printf("%-40s: %5d (%.1f%%)\n", 
				event.EventType, 
				event.Count, 
				float64(event.Count)/float64(parseResp.ParsedCount)*100)
		}
	}
	
	// Show unknown/chat/trigger events with examples
	fmt.Println("\n=== Events Needing Further Classification ===")
	
	// Focus on specific categories
	focusTypes := []string{"unknown_other", "chat_message", "trigger"}
	
	for _, prefix := range focusTypes {
		fmt.Printf("\n--- %s Events ---\n", prefix)
		for eventType, examples := range unknownExamples {
			if strings.HasPrefix(eventType, prefix) {
				fmt.Printf("\n%s (%d total):\n", eventType, eventCounts[eventType])
				
				// Analyze patterns in these examples
				patterns := make(map[string]int)
				for _, ex := range examples {
					pattern := identifyPattern(ex)
					patterns[pattern]++
				}
				
				// Show pattern distribution
				fmt.Println("Patterns found:")
				for pattern, count := range patterns {
					fmt.Printf("  - %s: %d\n", pattern, count)
				}
				
				// Show examples
				fmt.Println("Examples:")
				for i, ex := range examples {
					if len(ex) > 150 {
						ex = ex[:150] + "..."
					}
					fmt.Printf("  %d. %s\n", i+1, ex)
				}
			}
		}
	}
	
	// Count total unknowns
	totalUnknown := 0
	for eventType, count := range eventCounts {
		if strings.Contains(eventType, "unknown") {
			totalUnknown += count
		}
	}
	
	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Total unknown events: %d (%.1f%% of all parsed)\n", 
		totalUnknown, 
		float64(totalUnknown)/float64(parseResp.ParsedCount)*100)
	
	// Show chat message breakdown if significant
	if chatCount, exists := eventCounts["chat_message"]; exists && chatCount > 100 {
		fmt.Printf("Chat messages: %d (%.1f%% of all parsed)\n", 
			chatCount, 
			float64(chatCount)/float64(parseResp.ParsedCount)*100)
	}
}

func identifyPattern(content string) string {
	// Remove quotes to analyze the actual content
	content = strings.Trim(content, "\"")
	
	switch {
	// Player data JSON
	case strings.Contains(content, "}}JSON_END"):
		return "json_end_marker"
	case strings.Contains(content, "JSON_START{{"):
		return "json_start_marker"
	case strings.Contains(content, "\"player_"):
		return "player_stats_json"
		
	// Chat patterns
	case strings.Contains(content, "\" say \""):
		msg := extractChatMessage(content)
		switch {
		case strings.HasPrefix(msg, "."):
			return fmt.Sprintf("command: %s", strings.Fields(msg)[0])
		case msg == "gg" || msg == "gg wp":
			return "gg_message"
		case len(msg) < 10:
			return "short_chat"
		default:
			return "general_chat"
		}
		
	// Player events
	case strings.Contains(content, "blinded"):
		return "blind_event"
	case strings.Contains(content, "left buyzone"):
		return "buyzone_event"
	case strings.Contains(content, "ACCOLADE"):
		return "accolade_event"
	case strings.Contains(content, "MatchStatus"):
		return "match_status"
	case strings.Contains(content, "sv_throw"):
		return "throw_debug"
		
	default:
		// Get first word
		words := strings.Fields(content)
		if len(words) > 0 {
			return fmt.Sprintf("starts_with: %s", words[0])
		}
		return "empty"
	}
}

func extractChatMessage(content string) string {
	// Extract message from say command
	if idx := strings.Index(content, "\" say \""); idx != -1 {
		start := idx + 7
		end := strings.LastIndex(content[start:], "\"")
		if end > 0 {
			return content[start:start+end]
		}
	}
	return ""
}