package consumers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"github.com/umerthow/reconcile/recon-worker/internal/services"
	"github.com/umerthow/reconcile/recon-worker/internal/utils"
)

type DLQConsumer struct {
	db    *services.DatabaseService
	redis *services.RedisService
}

func NewDLQConsumer(db *services.DatabaseService, redis *services.RedisService) *DLQConsumer {
	return &DLQConsumer{
		db:    db,
		redis: redis,
	}
}

// Setup is called once when consumer starts
func (c *DLQConsumer) Setup(sarama.ConsumerGroupSession) error {
	utils.Info().Msg("DLQ consumer group started")
	return nil
}

// Cleanup is called once when consumer stops
func (c *DLQConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	utils.Info().Msg("DLQ consumer group stopped")
	return nil
}

// ConsumeClaim processes messages from the DLQ topic
func (c *DLQConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			c.processMessage(session.Context(), message)

			// Always mark as processed to avoid infinite loops
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

func (c *DLQConsumer) processMessage(ctx context.Context, message *sarama.ConsumerMessage) {
	startTime := time.Now()

	// Parse DLQ message
	var dlqMsg map[string]interface{}
	if err := json.Unmarshal(message.Value, &dlqMsg); err != nil {
		utils.Error().Err(err).Msg("Failed to parse DLQ message")
		return
	}

	utils.Info().
		Str("original_topic", dlqMsg["original_topic"].(string)).
		Int64("original_offset", int64(dlqMsg["original_offset"].(float64))).
		Str("error", dlqMsg["error"].(string)).
		Msg("Processing DLQ message")

	// PSEUDOCODE: Real implementation would:
	// 1. Log the failed message details
	// 2. Store in a separate DLQ database table for Finance review
	// 3. Send alert to Slack if it's a critical failure
	// 4. Provide manual replay mechanism via API
	//
	// For POC, we just log it:
	utils.Warn().
		Interface("dlq_message", dlqMsg).
		Dur("processing_time", time.Since(startTime)).
		Msg("[MOCK] DLQ message logged for manual review")

	// PSEUDOCODE: Insert into dlq_messages table
	// query := `
	//     INSERT INTO dlq_messages (original_topic, original_offset, error_message, payload, failed_at, processed_at)
	//     VALUES (?, ?, ?, ?, ?, NOW())
	// `
	// c.db.ExecContext(ctx, query, dlqMsg["original_topic"], dlqMsg["original_offset"],
	//                  dlqMsg["error"], dlqMsg["payload"], dlqMsg["failed_at"])
}
