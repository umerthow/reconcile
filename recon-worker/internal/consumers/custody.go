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

type CustodyConsumer struct {
	db    *services.DatabaseService
	redis *services.RedisService
	kafka *services.KafkaService
}

func NewCustodyConsumer(db *services.DatabaseService, redis *services.RedisService, kafka *services.KafkaService) *CustodyConsumer {
	return &CustodyConsumer{
		db:    db,
		redis: redis,
		kafka: kafka,
	}
}

// Setup is called once when consumer starts
func (c *CustodyConsumer) Setup(sarama.ConsumerGroupSession) error {
	utils.Info().Msg("Custody consumer group started")
	return nil
}

// Cleanup is called once when consumer stops
func (c *CustodyConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	utils.Info().Msg("Custody consumer group stopped")
	return nil
}

// ConsumeClaim processes messages from the topic
func (c *CustodyConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
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
					Msg("Failed to process Custody message")

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

func (c *CustodyConsumer) processMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	startTime := time.Now()

	// Parse webhook message
	var webhookMsg models.WebhookMessage
	if err := json.Unmarshal(message.Value, &webhookMsg); err != nil {
		return err
	}

	utils.Info().
		Str("provider", webhookMsg.Provider).
		Str("timestamp", webhookMsg.Timestamp).
		Msg("Processing Custody webhook")

	// Extract txHash for deduplication
	txHash, ok := webhookMsg.Data["txHash"].(string)
	if !ok {
		utils.Error().Msg("Missing txHash in custody webhook")
		return nil // Skip invalid messages
	}

	// Check deduplication
	exists, err := c.redis.CheckDeduplication(ctx, "custody", txHash)
	if err != nil {
		utils.Warn().Err(err).Msg("Failed to check deduplication, continuing")
	} else if exists {
		utils.Info().Str("tx_hash", txHash).Msg("Duplicate webhook, skipping")
		return nil
	}

	// Extract relevant fields
	status, _ := webhookMsg.Data["status"].(string)

	// Parse incoming timestamp
	incomingTimestamp := time.Now()
	if ts, ok := webhookMsg.Data["timestamp"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
			incomingTimestamp = parsed
		}
	}

	// Update transaction in database
	if err := c.db.UpdateTransactionFromWebhook(ctx, "custody", txHash, status, incomingTimestamp); err != nil {
		return err
	}

	// Insert provider callback for audit trail
	rawPayload, _ := json.Marshal(webhookMsg.Data)
	if err := c.db.InsertProviderCallback(ctx, "custody", txHash, string(rawPayload), incomingTimestamp); err != nil {
		utils.Warn().Err(err).Msg("Failed to insert provider callback")
	}

	// Set deduplication cache
	if err := c.redis.SetDeduplication(ctx, "custody", txHash, 48*time.Hour); err != nil {
		utils.Warn().Err(err).Msg("Failed to set deduplication")
	}

	duration := time.Since(startTime)
	utils.Info().
		Str("tx_hash", txHash).
		Str("status", status).
		Dur("duration_ms", duration).
		Msg("Custody webhook processed successfully")

	return nil
}

func (c *CustodyConsumer) sendToDLQ(ctx context.Context, message *sarama.ConsumerMessage, originalErr error) {
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
