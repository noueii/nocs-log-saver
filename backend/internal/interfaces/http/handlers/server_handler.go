package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/noueii/nocs-log-saver/internal/domain/entities"
	"github.com/noueii/nocs-log-saver/internal/infrastructure/persistence"
)

// ServerHandler handles server management endpoints
type ServerHandler struct {
	serverRepo *persistence.PostgresServerRepository
}

// NewServerHandler creates a new server handler
func NewServerHandler(serverRepo *persistence.PostgresServerRepository) *ServerHandler {
	return &ServerHandler{serverRepo: serverRepo}
}

// CreateServerRequest represents a request to create a server
type CreateServerRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description"`
}

// UpdateServerRequest represents a request to update a server
type UpdateServerRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

// List lists all active servers
func (h *ServerHandler) List(c *gin.Context) {
	servers, err := h.serverRepo.List(c.Request.Context(), 100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list servers"})
		return
	}

	// Hide API keys in list response
	for _, server := range servers {
		if len(server.APIKey) > 10 {
			server.APIKey = server.APIKey[:10] + "..."
		}
	}

	c.JSON(http.StatusOK, servers)
}

// Get gets a single server by ID
func (h *ServerHandler) Get(c *gin.Context) {
	serverID := c.Param("id")
	
	server, err := h.serverRepo.FindByID(c.Request.Context(), serverID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	c.JSON(http.StatusOK, server)
}

// Create creates a new server
func (h *ServerHandler) Create(c *gin.Context) {
	var req CreateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	server := &entities.Server{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    true,
		CreatedBy:   &userID,
		IPAddress:   c.ClientIP(), // Initial IP, will be updated when server connects
	}

	if err := h.serverRepo.Create(c.Request.Context(), server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create server"})
		return
	}

	c.JSON(http.StatusCreated, server)
}

// Update updates a server
func (h *ServerHandler) Update(c *gin.Context) {
	serverID := c.Param("id")
	
	var req UpdateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	server, err := h.serverRepo.FindByID(c.Request.Context(), serverID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	server.Name = req.Name
	server.Description = req.Description
	server.IsActive = req.IsActive

	if err := h.serverRepo.Update(c.Request.Context(), server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server"})
		return
	}

	c.JSON(http.StatusOK, server)
}

// Delete deactivates a server
func (h *ServerHandler) Delete(c *gin.Context) {
	serverID := c.Param("id")
	
	if err := h.serverRepo.Delete(c.Request.Context(), serverID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete server"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Server deactivated successfully"})
}

// RegenerateAPIKey generates a new API key for a server
func (h *ServerHandler) RegenerateAPIKey(c *gin.Context) {
	serverID := c.Param("id")
	
	apiKey, err := h.serverRepo.RegenerateAPIKey(c.Request.Context(), serverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to regenerate API key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"api_key": apiKey,
		"message": "API key regenerated successfully",
	})
}