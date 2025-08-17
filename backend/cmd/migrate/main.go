package main

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/noueii/nocs-log-saver/internal/infrastructure/config"
)

func main() {
	var direction string
	flag.StringVar(&direction, "direction", "up", "Migration direction: up or down")
	flag.Parse()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get database URL
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://cs2admin:password@localhost:5432/cs2logs"
	}

	// Database configuration
	dbConfig := config.DatabaseConfig{
		URL: dbURL,
	}

	// Connect to database
	db, err := config.NewDatabase(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if direction == "up" {
		log.Println("Running migrations UP...")
		if err := config.RunMigrations(db); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		log.Println("✅ Migrations completed successfully")
	} else {
		log.Println("Migration rollback not implemented yet")
	}
}