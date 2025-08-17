package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/noueii/nocs-log-saver/internal/domain/entities"
	"github.com/noueii/nocs-log-saver/internal/domain/repositories"
)

// WhitelistHandler handles IP whitelist management endpoints
type WhitelistHandler struct {
	repo repositories.WhitelistRepository
}

// NewWhitelistHandler creates a new whitelist handler
func NewWhitelistHandler(repo repositories.WhitelistRepository) *WhitelistHandler {
	return &WhitelistHandler{repo: repo}
}

// List returns all whitelist entries
func (h *WhitelistHandler) List(c *gin.Context) {
	entries, err := h.repo.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch whitelist",
		})
		return
	}

	c.JSON(http.StatusOK, entries)
}

// Create adds a new IP to the whitelist
func (h *WhitelistHandler) Create(c *gin.Context) {
	var req struct {
		IPAddress   string  `json:"ip_address" binding:"required"`
		ServerID    *string `json:"server_id"`
		Description string  `json:"description"`
		Enabled     bool    `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	entry := &entities.IPWhitelist{
		IPAddress:   req.IPAddress,
		ServerID:    req.ServerID,
		Description: req.Description,
		Enabled:     req.Enabled,
		CreatedBy:   c.GetString("user_id"), // Would be set by auth middleware
	}

	if err := h.repo.Create(c.Request.Context(), entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create whitelist entry",
		})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

// Update modifies an existing whitelist entry
func (h *WhitelistHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID",
		})
		return
	}

	var req struct {
		IPAddress   string  `json:"ip_address" binding:"required"`
		ServerID    *string `json:"server_id"`
		Description string  `json:"description"`
		Enabled     bool    `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	entry := &entities.IPWhitelist{
		ID:          id,
		IPAddress:   req.IPAddress,
		ServerID:    req.ServerID,
		Description: req.Description,
		Enabled:     req.Enabled,
	}

	if err := h.repo.Update(c.Request.Context(), entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update whitelist entry",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Updated successfully",
	})
}

// Delete removes an IP from the whitelist
func (h *WhitelistHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID",
		})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete whitelist entry",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Deleted successfully",
	})
}

// Check verifies if an IP is whitelisted
func (h *WhitelistHandler) Check(c *gin.Context) {
	ip := c.Param("ip")
	
	allowed, err := h.repo.IsAllowed(c.Request.Context(), ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check IP",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ip":      ip,
		"allowed": allowed,
	})
}