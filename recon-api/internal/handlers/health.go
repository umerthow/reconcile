package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/umerthow/reconcile/recon-api/internal/services"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db    *services.DatabaseService
	redis *services.RedisService
}

func NewHealthHandler(db *services.DatabaseService, redis *services.RedisService) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

// Health checks the health of all services
func (h *HealthHandler) Health(c *gin.Context) {
	status := gin.H{
		"status": "healthy",
		"services": gin.H{
			"database": "up",
			"redis":    "up",
		},
	}

	// Check database
	if err := h.db.Ping(c.Request.Context()); err != nil {
		status["status"] = "unhealthy"
		status["services"].(gin.H)["database"] = "down"
	}

	// Check Redis
	if err := h.redis.Ping(c.Request.Context()); err != nil {
		status["status"] = "unhealthy"
		status["services"].(gin.H)["redis"] = "down"
	}

	statusCode := 200
	if status["status"] == "unhealthy" {
		statusCode = 503
	}

	c.JSON(statusCode, status)
}

// Ready checks if the service is ready to accept traffic
func (h *HealthHandler) Ready(c *gin.Context) {
	c.JSON(200, gin.H{
		"ready": true,
	})
}
