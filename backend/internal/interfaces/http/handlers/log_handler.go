package handlers

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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

		// Process each log line
		var savedCount int
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Save raw log
			if err := saveRawLog(db, serverID, line); err != nil {
				// Log error but continue processing other lines
				continue
			}
			savedCount++
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