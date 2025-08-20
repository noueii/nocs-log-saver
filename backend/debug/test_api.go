package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
	// Read first 10 lines from match-test.txt
	content, err := os.ReadFile("debug/match-test.txt")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	lines := strings.Split(string(content), "\n")
	testLines := []string{}
	for i, line := range lines {
		if i >= 10 {
			break
		}
		if strings.TrimSpace(line) != "" {
			testLines = append(testLines, line)
		}
	}

	// Create request
	req := ParseTestRequest{
		Logs: strings.Join(testLines, "\n"),
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		return
	}

	// Send request to parse-test endpoint
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

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error response (status %d): %s\n", resp.StatusCode, string(body))
		return
	}

	// Parse response
	var parseResp ParseTestResponse
	if err := json.Unmarshal(body, &parseResp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		fmt.Printf("Raw response: %s\n", string(body))
		return
	}

	// Display results
	fmt.Println("=== Parse Test API Results ===")
	fmt.Printf("Total Lines: %d\n", parseResp.TotalLines)
	fmt.Printf("Parsed: %d\n", parseResp.ParsedCount)
	fmt.Printf("Failed: %d\n\n", parseResp.FailedCount)

	// Count event types
	eventTypes := make(map[string]int)
	for _, result := range parseResp.Results {
		if result.Success {
			eventTypes[result.EventType]++
		}
	}

	fmt.Println("=== Event Types ===")
	for eventType, count := range eventTypes {
		fmt.Printf("%-30s: %d\n", eventType, count)
		if strings.Contains(eventType, "unrecognized") {
			// Show example
			for _, result := range parseResp.Results {
				if result.EventType == eventType {
					fmt.Printf("  Example: %.100s...\n", result.Content)
					break
				}
			}
		}
	}

	// Show failed examples
	if parseResp.FailedCount > 0 {
		fmt.Println("\n=== Failed Parses ===")
		failCount := 0
		for _, result := range parseResp.Results {
			if !result.Success {
				fmt.Printf("%d. %.100s...\n", failCount+1, result.Content)
				fmt.Printf("   Error: %s\n", result.Error)
				failCount++
				if failCount >= 3 {
					break
				}
			}
		}
	}
}