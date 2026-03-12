package consumers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"github.com/umerthow/reconcile/recon-worker/internal/models"
	"github.com/umerthow/reconcile/recon-worker/internal/services"
	"github.com/umerthow/reconcile/recon-worker/internal/utils"
)

type XenditConsumer struct {
	db    *services.DatabaseService
	redis *services.RedisService
	kafka *services.KafkaService
}

func NewXenditConsumer(db *services.DatabaseService, redis *services.RedisService, kafka *services.KafkaService) *XenditConsumer {
	return &XenditConsumer{
		db:    db,
		redis: redis,
		kafka: kafka,
	}
}

// Setup is called once when consumer starts
func (c *XenditConsumer) Setup(sarama.ConsumerGroupSession) error {
	utils.Info().Msg("Xendit consumer group started")
	return nil
}

// Cleanup is called once when consumer stops
func (c *XenditConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	utils.Info().Msg("Xendit consumer group stopped")
	return nil
}

// ConsumeClaim processes messages from the topic
func (c *XenditConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			if err := c.processMessage(session.Context(), message); err != nil {
				utils.Error().
					Err(err).
					Str("topic", message.Topic).
					Int64("offset", message.Offset).
					Msg("Failed to process Xendit message")

				// Send to DLQ for manual retry
				c.sendToDLQ(session.Context(), message, err)
			}

			// Mark message as processed
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

func (c *XenditConsumer) processMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	startTime := time.Now()

	// Parse webhook message
	var webhookMsg models.WebhookMessage
	if err := json.Unmarshal(message.Value, &webhookMsg); err != nil {
		return err
	}

	utils.Info().
		Str("provider", webhookMsg.Provider).
		Str("timestamp", webhookMsg.Timestamp).
		Msg("Processing Xendit webhook")

	// Extract webhook ID for deduplication
	webhookID := ""
	if id, ok := webhookMsg.Data["id"].(string); ok {
		webhookID = id
	} else if externalID, ok := webhookMsg.Data["external_id"].(string); ok {
		webhookID = externalID
	} else {
		utils.Error().Msg("Missing webhook ID (id or external_id)")
		return nil // Skip invalid messages
	}

	// Check deduplication
	exists, err := c.redis.CheckDeduplication(ctx, "xendit", webhookID)
	if err != nil {
		utils.Warn().Err(err).Msg("Failed to check deduplication, continuing")
	} else if exists {
		utils.Info().Str("webhook_id", webhookID).Msg("Duplicate webhook, skipping")
		return nil
	}

	// Extract relevant fields
	externalID, _ := webhookMsg.Data["external_id"].(string)
	status, _ := webhookMsg.Data["status"].(string)

	// Parse incoming timestamp
	incomingTimestamp := time.Now()
	if ts, ok := webhookMsg.Data["updated"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
			incomingTimestamp = parsed
		}
	}

	// Update transaction in database
	if err := c.db.UpdateTransactionFromWebhook(ctx, "xendit", externalID, status, incomingTimestamp); err != nil {
		return err
	}

	// Insert provider callback for audit trail
	rawPayload, _ := json.Marshal(webhookMsg.Data)
	if err := c.db.InsertProviderCallback(ctx, "xendit", webhookID, string(rawPayload), incomingTimestamp); err != nil {
		utils.Warn().Err(err).Msg("Failed to insert provider callback")
	}

	// Set deduplication cache
	if err := c.redis.SetDeduplication(ctx, "xendit", webhookID, 48*time.Hour); err != nil {
		utils.Warn().Err(err).Msg("Failed to set deduplication")
	}

	duration := time.Since(startTime)
	utils.Info().
		Str("external_id", externalID).
		Str("status", status).
		Dur("duration_ms", duration).
		Msg("Xendit webhook processed successfully")

	return nil
}

func (c *XenditConsumer) sendToDLQ(ctx context.Context, message *sarama.ConsumerMessage, originalErr error) {
	dlqMessage := map[string]interface{}{
		"original_topic":     message.Topic,
		"original_partition": message.Partition,
		"original_offset":    message.Offset,
		"original_timestamp": message.Timestamp,
		"error":              originalErr.Error(),
		"payload":            string(message.Value),
		"failed_at":          time.Now().Format(time.RFC3339),
	}

	if err := c.kafka.PublishMessage(ctx, "dlq-webhooks", dlqMessage); err != nil {
		utils.Error().Err(err).Msg("Failed to send message to DLQ")
	} else {
		utils.Warn().
			Str("topic", message.Topic).
			Int64("offset", message.Offset).
			Msg("Message sent to DLQ")
	}
}
