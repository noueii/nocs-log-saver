package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/noueii/nocs-log-saver/internal/infrastructure/persistence"
)

// ServerAuthMiddleware validates that the server ID exists and is active
// It also validates the API key if provided in the query parameters
func ServerAuthMiddleware(serverRepo *persistence.PostgresServerRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		serverID := c.Param("server_id")
		if serverID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Server ID is required"})
			c.Abort()
			return
		}

		// Get the server details
		server, err := serverRepo.FindByID(c.Request.Context(), serverID)
		if err != nil {
			if err.Error() == "server not found" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or inactive server ID"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate server"})
			}
			c.Abort()
			return
		}

		// Check if server is active
		if !server.IsActive {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or inactive server ID"})
			c.Abort()
			return
		}

		// Check API key if provided
		apiKey := c.Query("key")
		if apiKey != "" {
			// If API key is provided, it must match
			if server.APIKey != apiKey {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
				c.Abort()
				return
			}
		}
		// Note: In the future, we can make API key required by checking:
		// else if server.RequireAPIKey { ... }

		// Store server ID and client IP in context for later use
		c.Set("server_id", serverID)
		c.Set("client_ip", c.ClientIP())

		c.Next()
	}
}