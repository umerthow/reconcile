package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/umerthow/reconcile/recon-api/internal/middleware"
	"github.com/umerthow/reconcile/recon-api/internal/models"
	"github.com/umerthow/reconcile/recon-api/internal/services"
	"github.com/umerthow/reconcile/recon-api/internal/utils"
)

// WebhookHandler handles webhook requests
type WebhookHandler struct {
	db    *services.DatabaseService
	redis *services.RedisService
	kafka *services.KafkaService
}

func NewWebhookHandler(db *services.DatabaseService, redis *services.RedisService, kafka *services.KafkaService) *WebhookHandler {
	return &WebhookHandler{
		db:    db,
		redis: redis,
		kafka: kafka,
	}
}

// HandleXenditWebhook processes Xendit webhooks
func (h *WebhookHandler) HandleXenditWebhook(c *gin.Context) {
	var req models.WebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse webhook")
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

	// Verify signature
	signature := c.GetHeader("X-Xendit-Signature")
	payload, _ := c.GetRawData()
	if !utils.VerifyXenditSignature(payload, signature, "xendit-secret") {
		log.Warn().Str("provider", "xendit").Msg("Invalid webhook signature (mocked)")
		// TODO: For POC, we'll skip signature verification failure
	}

	// Generate webhook ID from provider data
	externalID := ""
	if id, ok := req.Data["id"].(string); ok {
		externalID = id
	}
	webhookID := utils.GenerateWebhookID(req.Provider, externalID)

	// Check for duplicates
	isDuplicate, err := h.redis.CheckWebhookDuplicate(c.Request.Context(), webhookID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check duplicate")
		c.JSON(500, gin.H{
			"success": false,
			"message": "Internal error",
		})
		return
	}

	if isDuplicate {
		log.Info().Str("webhook_id", webhookID).Msg("Duplicate webhook received")
		c.JSON(200, gin.H{
			"success": true,
			"message": "Webhook already processed",
		})
		return
	}

	// Publish to Kafka
	if err := h.kafka.PublishToXenditTopic(c.Request.Context(), webhookID, req.Data); err != nil {
		log.Error().Err(err).Msg("Failed to publish to Kafka")
		c.JSON(500, gin.H{
			"success": false,
			"message": "Failed to process webhook",
		})
		return
	}

	// Mark as processed
	if err := h.redis.MarkWebhookProcessed(c.Request.Context(), webhookID); err != nil {
		log.Error().Err(err).Msg("Failed to mark webhook as processed")
	}

	log.Info().
		Str("webhook_id", webhookID).
		Str("provider", req.Provider).
		Str("event_type", req.EventType).
		Msg("Webhook processed successfully")

	c.JSON(200, gin.H{
		"success": true,
		"message": "Webhook received",
		"data": gin.H{
			"webhook_id": webhookID,
		},
	})
}

// HandleCustodyWebhook processes Custody webhooks
func (h *WebhookHandler) HandleCustodyWebhook(c *gin.Context) {
	var req models.WebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse webhook")
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

	// Verify signature
	signature := c.GetHeader("X-Custody-Signature")
	payload, _ := c.GetRawData()
	if !utils.VerifyCustodySignature(payload, signature, "custody-secret") {
		log.Warn().Str("provider", "custody").Msg("Invalid webhook signature (mocked)")
		// TODO: For POC, we'll skip signature verification failure
	}

	// Generate webhook ID from provider data
	externalID := ""
	if id, ok := req.Data["txHash"].(string); ok {
		externalID = id
	}
	webhookID := utils.GenerateWebhookID(req.Provider, externalID)

	// Check for duplicates
	isDuplicate, err := h.redis.CheckWebhookDuplicate(c.Request.Context(), webhookID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check duplicate")
		c.JSON(500, gin.H{
			"success": false,
			"message": "Internal error",
		})
		return
	}

	if isDuplicate {
		log.Info().Str("webhook_id", webhookID).Msg("Duplicate webhook received")
		c.JSON(200, gin.H{
			"success": true,
			"message": "Webhook already processed",
		})
		return
	}

	// Publish to Kafka
	if err := h.kafka.PublishToCustodyTopic(c.Request.Context(), webhookID, req.Data); err != nil {
		log.Error().Err(err).Msg("Failed to publish to Kafka")
		c.JSON(500, gin.H{
			"success": false,
			"message": "Failed to process webhook",
		})
		return
	}

	// Mark as processed
	if err := h.redis.MarkWebhookProcessed(c.Request.Context(), webhookID); err != nil {
		log.Error().Err(err).Msg("Failed to mark webhook as processed")
	}

	log.Info().
		Str("webhook_id", webhookID).
		Str("provider", req.Provider).
		Str("event_type", req.EventType).
		Msg("Webhook processed successfully")

	c.JSON(200, gin.H{
		"success": true,
		"message": "Webhook received",
		"data": gin.H{
			"webhook_id": webhookID,
		},
	})
}
