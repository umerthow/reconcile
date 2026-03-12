package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"github.com/umerthow/reconcile/recon-api/internal/config"
	"github.com/umerthow/reconcile/recon-api/internal/utils"
)

var kafkaLogger = utils.GetLogger("kafka")

type KafkaService struct {
	producer sarama.SyncProducer
	config   *config.KafkaConfig
}

func NewKafkaService(cfg *config.KafkaConfig) (*KafkaService, error) {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.Producer.Return.Errors = true
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Retry.Max = 3
	kafkaConfig.Producer.Compression = sarama.CompressionSnappy

	// SASL authentication if credentials provided
	if cfg.Username != "" && cfg.Password != "" {
		kafkaConfig.Net.SASL.Enable = true
		kafkaConfig.Net.SASL.User = cfg.Username
		kafkaConfig.Net.SASL.Password = cfg.Password
		kafkaConfig.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	}

	// SSL/TLS configuration
	if cfg.SSLEnable {
		kafkaConfig.Net.TLS.Enable = true
	}

	producer, err := sarama.NewSyncProducer(cfg.Brokers, kafkaConfig)
	if err != nil {
		return nil, err
	}

	kafkaLogger.Info().Strs("brokers", cfg.Brokers).Msg("Kafka producer initialized")

	return &KafkaService{
		producer: producer,
		config:   cfg,
	}, nil
}

func (s *KafkaService) Close() error {
	if s.producer != nil {
		return s.producer.Close()
	}
	return nil
}

// PublishWebhook publishes a webhook to Kafka
func (s *KafkaService) PublishWebhook(ctx context.Context, topic string, webhookID string, payload interface{}) error {
	// Marshal payload to JSON
	data, err := json.Marshal(payload)
	if err != nil {
		kafkaLogger.Error().Err(err).Msg("Failed to marshal webhook payload")
		return err
	}

	// Create message
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(webhookID),
		Value: sarama.ByteEncoder(data),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("webhook_id"),
				Value: []byte(webhookID),
			},
			{
				Key:   []byte("timestamp"),
				Value: []byte(time.Now().Format(time.RFC3339)),
			},
		},
	}

	// Send message
	partition, offset, err := s.producer.SendMessage(msg)
	if err != nil {
		kafkaLogger.Error().
			Err(err).
			Str("topic", topic).
			Str("webhook_id", webhookID).
			Msg("Failed to publish webhook to Kafka")
		return err
	}

	kafkaLogger.Info().
		Str("topic", topic).
		Str("webhook_id", webhookID).
		Int32("partition", partition).
		Int64("offset", offset).
		Msg("Webhook published to Kafka")

	return nil
}

// PublishToXenditTopic publishes to xendit-webhooks topic
func (s *KafkaService) PublishToXenditTopic(ctx context.Context, webhookID string, payload interface{}) error {
	return s.PublishWebhook(ctx, s.config.Topics.XenditWebhooks, webhookID, payload)
}

// PublishToCustodyTopic publishes to custody-webhooks topic
func (s *KafkaService) PublishToCustodyTopic(ctx context.Context, webhookID string, payload interface{}) error {
	return s.PublishWebhook(ctx, s.config.Topics.CustodyWebhooks, webhookID, payload)
}

// PublishToDLQ publishes failed message to Dead Letter Queue
func (s *KafkaService) PublishToDLQ(ctx context.Context, webhookID string, payload interface{}, errorMsg string) error {
	dlqPayload := map[string]interface{}{
		"original_payload": payload,
		"error":            errorMsg,
		"timestamp":        time.Now(),
	}

	return s.PublishWebhook(ctx, s.config.Topics.DLQ, webhookID, dlqPayload)
}
