package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/noueii/nocs-log-saver/internal/domain/repositories"
)

// IPWhitelistMiddleware creates a middleware that checks IP whitelist
func IPWhitelistMiddleware(whitelistRepo repositories.WhitelistRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := getClientIP(c)
		
		// Check if IP is whitelisted
		allowed, err := whitelistRepo.IsAllowed(c.Request.Context(), clientIP)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check whitelist",
			})
			c.Abort()
			return
		}
		
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "IP not whitelisted",
				"ip":    clientIP,
			})
			c.Abort()
			return
		}
		
		// Store client IP in context for later use
		c.Set("client_ip", clientIP)
		c.Next()
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header (for proxies)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	
	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to remote address
	return c.ClientIP()
}

// CORSMiddleware handles CORS for the admin UI
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}