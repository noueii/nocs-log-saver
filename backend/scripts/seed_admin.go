package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Connect to database
	db, err := sqlx.Open("postgres", "postgresql://cs2admin:localpass123@localhost:5432/cs2logs?sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Check if admin user already exists
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM users WHERE username = 'admin'")
	if err == nil && count > 0 {
		log.Println("Admin user already exists")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Admin123!"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Create admin user
	adminID := uuid.New()
	query := `
		INSERT INTO users (id, email, username, password_hash, full_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	
	_, err = db.ExecContext(context.Background(), query,
		adminID,
		"admin@example.com",
		"admin",
		string(hashedPassword),
		"System Administrator",
		"super_admin",
		true,
		time.Now(),
		time.Now(),
	)
	
	if err != nil {
		log.Fatal("Failed to create admin user:", err)
	}

	fmt.Println("Admin user created successfully!")
	fmt.Println("Username: admin")
	fmt.Println("Password: Admin123!")
	fmt.Println("Role: super_admin")
}