package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/umerthow/reconcile/recon-api/internal/middleware"
	"github.com/umerthow/reconcile/recon-api/internal/models"
	"github.com/umerthow/reconcile/recon-api/internal/services"
)

// ReconciliationHandler handles reconciliation operations
type ReconciliationHandler struct {
	db *services.DatabaseService
}

func NewReconciliationHandler(db *services.DatabaseService) *ReconciliationHandler {
	return &ReconciliationHandler{db: db}
}

// ListReconciliationRuns returns paginated reconciliation runs
func (h *ReconciliationHandler) ListReconciliationRuns(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	provider := c.Query("provider")

	// Get data from database
	// TODO: Pass filters (status, provider) when implementing real queries
	_ = status
	_ = provider
	runs, total, err := h.db.GetReconciliationRuns(c.Request.Context(), page, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get reconciliation runs")
		c.JSON(500, gin.H{
			"success": false,
			"message": "Failed to fetch data",
		})
		return
	}
	_ = runs // TODO: Use real data

	// Calculate pagination
	response := models.ReconciliationRunListResponse{
		Runs:     []models.ReconciliationRun{}, // TODO: Convert from database result
		Total:    total,
		Page:     page,
		PageSize: limit,
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    response,
	})
}

// TriggerReconciliation starts a new reconciliation run
func (h *ReconciliationHandler) TriggerReconciliation(c *gin.Context) {
	var req models.TriggerReconciliationRequest
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

	// Create reconciliation run
	runID, err := h.db.CreateReconciliationRun(c.Request.Context(), req.WindowStart, req.WindowEnd)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create reconciliation run")
		c.JSON(500, gin.H{
			"success": false,
			"message": "Failed to trigger reconciliation",
		})
		return
	}

	log.Info().
		Int64("run_id", runID).
		Time("window_start", req.WindowStart).
		Time("window_end", req.WindowEnd).
		Msg("Reconciliation triggered")

	c.JSON(200, gin.H{
		"success": true,
		"message": "Reconciliation triggered successfully",
		"data": gin.H{
			"run_id": runID,
		},
	})
}
