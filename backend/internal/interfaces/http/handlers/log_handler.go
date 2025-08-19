package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/noueii/nocs-log-saver/internal/application/services"
)

// HandleLogIngestion handles incoming CS2 server logs
func HandleLogIngestion(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		serverID := c.GetString("server_id") // Set by middleware
		clientIP := c.GetString("client_ip") // Set by middleware

		// Read request body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read request body",
			})
			return
		}

		// Convert body to string and split by lines
		content := string(body)
		lines := strings.Split(content, "\n")

		// Create parser service
		parserService := services.NewParserService(db)

		// Process each log line
		var savedCount int
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Save raw log and get ID
			logID := uuid.New().String()
			if err := saveRawLogWithID(db, logID, serverID, line); err != nil {
				// Log error but continue processing other lines
				continue
			}
			savedCount++
			
			// Parse log asynchronously
			go parserService.ParseAndStore(logID, serverID, line)
		}

		// Update server last seen
		updateServerLastSeen(db, serverID, clientIP)

		c.JSON(http.StatusOK, gin.H{
			"received":    true,
			"line_count":  savedCount,
			"server_id":   serverID,
			"timestamp":   time.Now().Unix(),
		})
	}
}

// saveRawLog saves a single log line to the database
func saveRawLog(db *sqlx.DB, serverID, content string) error {
	query := `
		INSERT INTO raw_logs (id, server_id, content, received_at)
		VALUES ($1, $2, $3, $4)
	`
	
	_, err := db.Exec(query,
		uuid.New().String(),
		serverID,
		content,
		time.Now(),
	)
	
	return err
}

// saveRawLogWithID saves a single log line with a specific ID
func saveRawLogWithID(db *sqlx.DB, id, serverID, content string) error {
	query := `
		INSERT INTO raw_logs (id, server_id, content, received_at)
		VALUES ($1, $2, $3, $4)
	`
	
	_, err := db.Exec(query, id, serverID, content, time.Now())
	return err
}

// updateServerLastSeen updates or creates server record
func updateServerLastSeen(db *sqlx.DB, serverID, ipAddress string) error {
	query := `
		INSERT INTO servers (id, ip_address, last_seen, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE
		SET ip_address = $2, last_seen = $3
	`
	
	_, err := db.Exec(query,
		serverID,
		ipAddress,
		time.Now(),
		time.Now(),
	)
	
	return err
}

// GetLogs handles fetching logs from the database
func GetLogs(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		serverID := c.Query("server_id")
		logType := c.Query("type") // raw, parsed, or failed
		limit := 100 // Default limit
		download := c.Query("download") == "true"
		
		var logs []gin.H
		
		switch logType {
		case "parsed":
			logs = getParsedLogs(db, serverID, limit)
		case "failed":
			logs = getFailedLogs(db, serverID, limit)
		default: // "raw" or empty
			logs = getRawLogs(db, serverID, limit)
		}
		
		// If download requested, return as text file
		if download {
			var content strings.Builder
			for _, log := range logs {
				content.WriteString(fmt.Sprintf("[%s] %s: %s\n", 
					log["created_at"], 
					log["server_id"], 
					log["content"]))
			}
			
			c.Header("Content-Type", "text/plain")
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=cs2-logs-%s.txt", time.Now().Format("20060102-150405")))
			c.String(http.StatusOK, content.String())
			return
		}
		
		c.JSON(http.StatusOK, logs)
	}
}

func getRawLogs(db *sqlx.DB, serverID string, limit int) []gin.H {
	var query string
	var args []interface{}
	
	if serverID != "" {
		query = `
			SELECT id, server_id, content, received_at as created_at 
			FROM raw_logs 
			WHERE server_id = $1 
			ORDER BY received_at DESC 
			LIMIT $2
		`
		args = []interface{}{serverID, limit}
	} else {
		query = `
			SELECT id, server_id, content, received_at as created_at 
			FROM raw_logs 
			ORDER BY received_at DESC 
			LIMIT $1
		`
		args = []interface{}{limit}
	}
	
	rows, err := db.Query(query, args...)
	if err != nil {
		return []gin.H{}
	}
	defer rows.Close()
	
	var logs []gin.H
	for rows.Next() {
		var log struct {
			ID        string    `json:"id"`
			ServerID  string    `json:"server_id"`
			Content   string    `json:"content"`
			CreatedAt time.Time `json:"created_at"`
		}
		
		if err := rows.Scan(&log.ID, &log.ServerID, &log.Content, &log.CreatedAt); err != nil {
			continue
		}
		
		logs = append(logs, gin.H{
			"id":         log.ID,
			"server_id":  log.ServerID,
			"content":    log.Content,
			"created_at": log.CreatedAt.Format(time.RFC3339),
			"type":       "raw",
		})
	}
	
	if logs == nil {
		return []gin.H{}
	}
	return logs
}

func getParsedLogs(db *sqlx.DB, serverID string, limit int) []gin.H {
	var query string
	var args []interface{}
	
	if serverID != "" {
		query = `
			SELECT p.id, p.server_id, p.event_type, p.event_data, p.created_at, r.content
			FROM parsed_logs p
			JOIN raw_logs r ON p.raw_log_id = r.id
			WHERE p.server_id = $1
			ORDER BY p.created_at DESC
			LIMIT $2
		`
		args = []interface{}{serverID, limit}
	} else {
		query = `
			SELECT p.id, p.server_id, p.event_type, p.event_data, p.created_at, r.content
			FROM parsed_logs p
			JOIN raw_logs r ON p.raw_log_id = r.id
			ORDER BY p.created_at DESC
			LIMIT $1
		`
		args = []interface{}{limit}
	}
	
	rows, err := db.Query(query, args...)
	if err != nil {
		return []gin.H{}
	}
	defer rows.Close()
	
	var logs []gin.H
	for rows.Next() {
		var log struct {
			ID        string          `json:"id"`
			ServerID  string          `json:"server_id"`
			EventType string          `json:"event_type"`
			EventData json.RawMessage `json:"event_data"`
			CreatedAt time.Time       `json:"created_at"`
			Content   string          `json:"content"`
		}
		
		if err := rows.Scan(&log.ID, &log.ServerID, &log.EventType, &log.EventData, &log.CreatedAt, &log.Content); err != nil {
			continue
		}
		
		logs = append(logs, gin.H{
			"id":         log.ID,
			"server_id":  log.ServerID,
			"content":    log.Content,
			"event_type": log.EventType,
			"event_data": log.EventData,
			"created_at": log.CreatedAt.Format(time.RFC3339),
			"type":       "parsed",
		})
	}
	
	if logs == nil {
		return []gin.H{}
	}
	return logs
}

func getFailedLogs(db *sqlx.DB, serverID string, limit int) []gin.H {
	var query string
	var args []interface{}
	
	if serverID != "" {
		query = `
			SELECT f.id, r.server_id, r.content, f.error_message, f.created_at
			FROM failed_parses f
			JOIN raw_logs r ON f.raw_log_id = r.id
			WHERE r.server_id = $1
			ORDER BY f.created_at DESC
			LIMIT $2
		`
		args = []interface{}{serverID, limit}
	} else {
		query = `
			SELECT f.id, r.server_id, r.content, f.error_message, f.created_at
			FROM failed_parses f
			JOIN raw_logs r ON f.raw_log_id = r.id
			ORDER BY f.created_at DESC
			LIMIT $1
		`
		args = []interface{}{limit}
	}
	
	rows, err := db.Query(query, args...)
	if err != nil {
		return []gin.H{}
	}
	defer rows.Close()
	
	var logs []gin.H
	for rows.Next() {
		var log struct {
			ID           string    `json:"id"`
			ServerID     string    `json:"server_id"`
			Content      string    `json:"content"`
			ErrorMessage string    `json:"error_message"`
			CreatedAt    time.Time `json:"created_at"`
		}
		
		if err := rows.Scan(&log.ID, &log.ServerID, &log.Content, &log.ErrorMessage, &log.CreatedAt); err != nil {
			continue
		}
		
		logs = append(logs, gin.H{
			"id":           log.ID,
			"server_id":    log.ServerID,
			"content":      log.Content,
			"error_message": log.ErrorMessage,
			"created_at":   log.CreatedAt.Format(time.RFC3339),
			"type":         "failed",
		})
	}
	
	if logs == nil {
		return []gin.H{}
	}
	return logs
}

// GetServers returns list of servers for dropdown
func GetServers(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := `
			SELECT id, name, is_active 
			FROM servers 
			WHERE is_active = true
			ORDER BY name
		`
		
		rows, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch servers"})
			return
		}
		defer rows.Close()
		
		var servers []gin.H
		for rows.Next() {
			var server struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				IsActive bool   `json:"is_active"`
			}
			
			if err := rows.Scan(&server.ID, &server.Name, &server.IsActive); err != nil {
				continue
			}
			
			servers = append(servers, gin.H{
				"id":   server.ID,
				"name": server.Name,
			})
		}
		
		if servers == nil {
			servers = []gin.H{}
		}
		
		c.JSON(http.StatusOK, servers)
	}
}