package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/umerthow/reconcile/recon-worker/internal/config"
	"github.com/umerthow/reconcile/recon-worker/internal/utils"
)

type KafkaService struct {
	client   sarama.Client
	producer sarama.SyncProducer
	config   *config.Config
}

func NewKafkaService(cfg *config.Config) (*KafkaService, error) {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Version = sarama.V2_6_0_0
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Retry.Max = 3
	kafkaConfig.Producer.Return.Successes = true

	// SSL configuration
	if cfg.KafkaSSLEnable {
		kafkaConfig.Net.TLS.Enable = true
		// Add TLS config here if needed
	}

	// SASL authentication
	if cfg.KafkaUsername != "" {
		kafkaConfig.Net.SASL.Enable = true
		kafkaConfig.Net.SASL.User = cfg.KafkaUsername
		kafkaConfig.Net.SASL.Password = cfg.KafkaPassword
		kafkaConfig.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	}

	client, err := sarama.NewClient(cfg.KafkaBrokers, kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %w", err)
	}

	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	utils.Info().
		Strs("brokers", cfg.KafkaBrokers).
		Msg("Kafka connection established")

	return &KafkaService{
		client:   client,
		producer: producer,
		config:   cfg,
	}, nil
}

func (s *KafkaService) Close() error {
	if s.producer != nil {
		if err := s.producer.Close(); err != nil {
			utils.Error().Err(err).Msg("Failed to close Kafka producer")
		}
	}
	if s.client != nil {
		if err := s.client.Close(); err != nil {
			utils.Error().Err(err).Msg("Failed to close Kafka client")
		}
	}
	return nil
}

// PublishMessage publishes a message to Kafka topic
func (s *KafkaService) PublishMessage(ctx context.Context, topic string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Value:     sarama.ByteEncoder(payload),
		Timestamp: time.Now(),
	}

	partition, offset, err := s.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	utils.Debug().
		Str("topic", topic).
		Int32("partition", partition).
		Int64("offset", offset).
		Msg("Message published to Kafka")

	return nil
}

// CreateConsumerGroup creates a new consumer group
func (s *KafkaService) CreateConsumerGroup(groupID string) (sarama.ConsumerGroup, error) {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Version = sarama.V2_6_0_0
	kafkaConfig.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	kafkaConfig.Consumer.Return.Errors = true

	// SSL configuration
	if s.config.KafkaSSLEnable {
		kafkaConfig.Net.TLS.Enable = true
	}

	// SASL authentication
	if s.config.KafkaUsername != "" {
		kafkaConfig.Net.SASL.Enable = true
		kafkaConfig.Net.SASL.User = s.config.KafkaUsername
		kafkaConfig.Net.SASL.Password = s.config.KafkaPassword
		kafkaConfig.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	}

	consumerGroup, err := sarama.NewConsumerGroup(s.config.KafkaBrokers, groupID, kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return consumerGroup, nil
}
