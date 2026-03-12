package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/umerthow/reconcile/recon-api/internal/middleware"
	"github.com/umerthow/reconcile/recon-api/internal/models"
	"github.com/umerthow/reconcile/recon-api/internal/services"
)

// DiscrepancyHandler handles discrepancy operations
type DiscrepancyHandler struct {
	db *services.DatabaseService
}

func NewDiscrepancyHandler(db *services.DatabaseService) *DiscrepancyHandler {
	return &DiscrepancyHandler{db: db}
}

// ListDiscrepancies returns discrepancies for a reconciliation run
func (h *DiscrepancyHandler) ListDiscrepancies(c *gin.Context) {
	runID := c.Param("run_id")
	if runID == "" {
		c.JSON(400, gin.H{
			"success": false,
			"message": "run_id is required",
		})
		return
	}

	// Get discrepancies
	filters := map[string]string{
		"run_id": runID,
	}
	discrepancies, total, err := h.db.GetDiscrepancies(c.Request.Context(), filters, 1, 100)
	if err != nil {
		log.Error().Err(err).Str("run_id", runID).Msg("Failed to get discrepancies")
		c.JSON(500, gin.H{
			"success": false,
			"message": "Failed to fetch discrepancies",
		})
		return
	}
	_ = total

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"discrepancies": discrepancies,
			"total":         len(discrepancies),
		},
	})
}

// ResolveDiscrepancy marks a discrepancy as resolved
func (h *DiscrepancyHandler) ResolveDiscrepancy(c *gin.Context) {
	discrepancyID := c.Param("discrepancy_id")
	if discrepancyID == "" {
		c.JSON(400, gin.H{
			"success": false,
			"message": "discrepancy_id is required",
		})
		return
	}

	var req models.ResolveDiscrepancyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse request")
		c.JSON(400, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	// Validate request
	if err := middleware.ValidateStruct(&req); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"message": "Validation failed",
			"errors":  err.Error(),
		})
		return
	}

	// Get user from context
	userID, _ := c.Get("user_id")

	// Parse discrepancy ID
	// TODO: For real implementation, convert string to int64
	// For POC, we'll use a mock ID
	var discID int64 = 1

	// Resolve discrepancy
	err := h.db.ResolveDiscrepancy(c.Request.Context(), discID, req.ResolvedBy, req.ResolutionNotes, req.Action)
	if err != nil {
		log.Error().Err(err).Str("discrepancy_id", discrepancyID).Msg("Failed to resolve discrepancy")
		c.JSON(500, gin.H{
			"success": false,
			"message": "Failed to resolve discrepancy",
		})
		return
	}

	log.Info().
		Str("discrepancy_id", discrepancyID).
		Str("action", req.Action).
		Str("user_id", userID.(string)).
		Msg("Discrepancy resolved")

	c.JSON(200, gin.H{
		"success": true,
		"message": "Discrepancy resolved successfully",
	})
}
