package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/noueii/nocs-log-saver/internal/application/services"
	"github.com/noueii/nocs-log-saver/internal/domain/entities"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check Bearer format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := authService.ValidateAccessToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID.String())
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", string(claims.Role))
		c.Set("claims", claims)

		c.Next()
	}
}

// RequireRole checks if user has required role
func RequireRole(roles ...entities.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "No role found"})
			c.Abort()
			return
		}

		// Check if user has required role
		userRoleStr := userRole.(string)
		hasRole := false
		for _, role := range roles {
			if string(role) == userRoleStr {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RBACMiddleware is an alias for RequirePermission for backward compatibility
func RBACMiddleware(resource, action string) gin.HandlerFunc {
	return RequirePermission(resource, action)
}

// RequirePermission checks if user has specific permission
func RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get claims from context
		claimsInterface, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "No claims found"})
			c.Abort()
			return
		}

		claims := claimsInterface.(*services.JWTClaims)

		// Create temporary user object to check permissions
		user := &entities.User{
			Role: claims.Role,
		}

		if !user.HasPermission(resource, action) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
				"required": gin.H{
					"resource": resource,
					"action":   action,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth validates JWT if present but doesn't require it
func OptionalAuth(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Check Bearer format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		// Validate token
		claims, err := authService.ValidateAccessToken(parts[1])
		if err != nil {
			c.Next()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID.String())
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", string(claims.Role))
		c.Set("claims", claims)
		c.Set("authenticated", true)

		c.Next()
	}
}

// GetUserFromContext retrieves user information from context
func GetUserFromContext(c *gin.Context) (userID, username, email, role string, authenticated bool) {
	if val, exists := c.Get("authenticated"); exists {
		authenticated = val.(bool)
	}

	if val, exists := c.Get("user_id"); exists {
		userID = val.(string)
	}

	if val, exists := c.Get("username"); exists {
		username = val.(string)
	}

	if val, exists := c.Get("email"); exists {
		email = val.(string)
	}

	if val, exists := c.Get("role"); exists {
		role = val.(string)
	}

	return
}