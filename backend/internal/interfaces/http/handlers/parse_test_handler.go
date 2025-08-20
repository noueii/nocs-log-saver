package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/noueii/nocs-log-saver/internal/application/services"
)

// ParseTestRequest represents the request body for testing log parsing
type ParseTestRequest struct {
	Logs string `json:"logs" binding:"required"`
}

// ParseTestResponse represents a single parsed/failed log result
type ParseTestResult struct {
	LineNumber  int         `json:"line_number"`
	Content     string      `json:"content"`
	Success     bool        `json:"success"`
	EventType   string      `json:"event_type,omitempty"`
	EventData   interface{} `json:"event_data,omitempty"`
	Error       string      `json:"error,omitempty"`
}

// ParseTestResponse represents the full response
type ParseTestResponse struct {
	TotalLines   int                `json:"total_lines"`
	ParsedCount  int                `json:"parsed_count"`
	FailedCount  int                `json:"failed_count"`
	Results      []ParseTestResult  `json:"results"`
}

// HandleParseTest handles testing log parsing without saving to database
func HandleParseTest(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ParseTestRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		// Split logs by lines
		lines := strings.Split(req.Logs, "\n")
		
		// Create parser service
		parserService := services.NewParserService(db)
		
		// Process each line
		var results []ParseTestResult
		parsedCount := 0
		failedCount := 0
		
		for i, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			
			// Try to parse the log line
			parsedLog, err := parserService.ParseLogLine(line)
			
			result := ParseTestResult{
				LineNumber: i + 1,
				Content:    line,
			}
			
			if err != nil {
				// Failed to parse
				result.Success = false
				result.Error = err.Error()
				failedCount++
			} else {
				// Successfully parsed
				result.Success = true
				result.EventType = parsedLog.EventType
				result.EventData = parsedLog.EventData
				parsedCount++
			}
			
			results = append(results, result)
		}
		
		response := ParseTestResponse{
			TotalLines:  len(results),
			ParsedCount: parsedCount,
			FailedCount: failedCount,
			Results:     results,
		}
		
		c.JSON(http.StatusOK, response)
	}
}