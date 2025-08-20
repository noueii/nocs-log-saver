package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/noueii/nocs-log-saver/internal/application/services"
)

func main() {
	// Get database URL from environment or use default
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://cs2admin:localpass123@localhost:5432/cs2logs?sslmode=disable"
	}

	// Connect to database
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Create parser service
	parserService := services.NewParserService(db)

	// Query unparsed raw logs
	query := `
		SELECT r.id, r.server_id, r.content 
		FROM raw_logs r
		LEFT JOIN parsed_logs p ON r.id = p.raw_log_id
		LEFT JOIN failed_parses f ON r.id = f.raw_log_id
		WHERE p.id IS NULL AND f.id IS NULL
		LIMIT 1000
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Failed to query raw logs:", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, serverID, content string
		if err := rows.Scan(&id, &serverID, &content); err != nil {
			continue
		}
		
		// Parse and store
		if err := parserService.ParseAndStore(id, serverID, content); err != nil {
			fmt.Printf("Failed to parse log %s: %v\n", id, err)
		} else {
			count++
			if count%100 == 0 {
				fmt.Printf("Parsed %d logs...\n", count)
			}
		}
	}

	fmt.Printf("\n=== Completed ===\n")
	fmt.Printf("Total parsed: %d\n", count)
	
	// Now check for unknown event types
	fmt.Println("\n=== Checking for Unknown Event Types ===")
	
	unknownQuery := `
		SELECT event_type, COUNT(*) as count 
		FROM parsed_logs 
		WHERE event_type LIKE 'unknown%' OR event_type = 'unknown'
		GROUP BY event_type 
		ORDER BY count DESC
	`
	
	var unknownRows *sql.Rows
	unknownRows, err = db.Query(unknownQuery)
	if err != nil {
		log.Fatal("Failed to query unknown events:", err)
	}
	defer unknownRows.Close()
	
	for unknownRows.Next() {
		var eventType string
		var eventCount int
		if err := unknownRows.Scan(&eventType, &eventCount); err != nil {
			continue
		}
		fmt.Printf("%s: %d\n", eventType, eventCount)
	}
	
	// Get examples of unknown events
	fmt.Println("\n=== Unknown Event Examples ===")
	
	exampleQuery := `
		SELECT p.event_type, r.content, p.event_data
		FROM parsed_logs p
		JOIN raw_logs r ON p.raw_log_id = r.id
		WHERE p.event_type LIKE 'unknown%' OR p.event_type = 'unknown'
		LIMIT 10
	`
	
	exampleRows, err := db.Query(exampleQuery)
	if err != nil {
		log.Fatal("Failed to query examples:", err)
	}
	defer exampleRows.Close()
	
	for exampleRows.Next() {
		var eventType, content, eventData string
		if err := exampleRows.Scan(&eventType, &content, &eventData); err != nil {
			continue
		}
		fmt.Printf("\nType: %s\n", eventType)
		fmt.Printf("Raw: %s\n", content)
		fmt.Printf("Data: %.200s...\n", eventData)
		fmt.Println("---")
	}
}