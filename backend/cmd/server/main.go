package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/noueii/nocs-log-saver/internal/application/services"
	"github.com/noueii/nocs-log-saver/internal/infrastructure/config"
	"github.com/noueii/nocs-log-saver/internal/infrastructure/persistence"
	"github.com/noueii/nocs-log-saver/internal/interfaces/http/handlers"
	"github.com/noueii/nocs-log-saver/internal/interfaces/http/middleware"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found - using environment variables")
	}

	// Database configuration
	dbURL := getEnv("DATABASE_URL", "postgres://cs2admin:localpass123@localhost:5432/cs2logs?sslmode=disable")
	log.Printf("Connecting to database with URL: %s", dbURL[:20]+"...") // Log first part of URL for debugging
	
	dbConfig := config.DatabaseConfig{
		URL:             dbURL,
		MaxConnections:  25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	// Connect to database
	db, err := config.NewDatabase(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := config.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	
	// Run auth migrations
	if err := config.RunAuthMigrations(db.DB); err != nil {
		log.Fatalf("Failed to run auth migrations: %v", err)
	}

	// Initialize repositories
	userRepo := persistence.NewPostgresUserRepository(db)
	sessionRepo := persistence.NewPostgresSessionRepository(db)
	serverRepo := persistence.NewPostgresServerRepository(db)

	// Initialize services
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key-change-this-in-production")
	authService := services.NewAuthService(userRepo, sessionRepo, jwtSecret)

	// Initialize Gin router
	gin.SetMode(getEnv("GIN_MODE", gin.ReleaseMode))
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(middleware.CORSMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().Unix(),
		})
	})

	// API routes
	api := router.Group("/api")
	{
		// Authentication routes
		authHandler := handlers.NewAuthHandler(authService)
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", middleware.AuthMiddleware(authService), authHandler.Logout)
		}

		// Public API routes (no auth required for viewing logs)
		api.GET("/logs", handlers.GetLogs(db))
		api.GET("/servers", handlers.GetServers(db)) // List servers for dropdown
		
		// Admin routes for server management (protected)
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(authService))
		{
			// Server management routes
			serverHandler := handlers.NewServerHandler(serverRepo)
			servers := admin.Group("/servers")
			servers.Use(middleware.RBACMiddleware("servers", "read"))
			{
				servers.GET("", serverHandler.List)
				servers.GET("/:id", serverHandler.Get)
				servers.POST("", middleware.RBACMiddleware("servers", "create"), serverHandler.Create)
				servers.PUT("/:id", middleware.RBACMiddleware("servers", "update"), serverHandler.Update)
				servers.DELETE("/:id", middleware.RBACMiddleware("servers", "delete"), serverHandler.Delete)
				servers.POST("/:id/regenerate-key", middleware.RBACMiddleware("servers", "update"), serverHandler.RegenerateAPIKey)
			}
		}
	}

	// Log ingestion endpoint with server authentication middleware
	router.POST("/logs/:server_id", 
		middleware.ServerAuthMiddleware(serverRepo),
		handlers.HandleLogIngestion(db),
	)

	// Start server with graceful shutdown
	srv := &http.Server{
		Addr:    ":" + getEnv("PORT", "9090"),
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started on port %s", getEnv("PORT", "9090"))

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}