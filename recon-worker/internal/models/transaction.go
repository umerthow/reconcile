package models

import "time"

// Transaction represents internal transaction record
type Transaction struct {
	ID                int64      `json:"id"`
	Type              string     `json:"type"`     // deposit, withdrawal
	Provider          string     `json:"provider"` // xendit, custody
	ProviderRef       string     `json:"provider_ref"`
	Currency          string     `json:"currency"`
	Amount            string     `json:"amount"` // Use string for decimal precision
	Status            string     `json:"status"` // pending, completed, failed
	UserID            int64      `json:"user_id"`
	IncomingTimestamp *time.Time `json:"incoming_timestamp,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// ProviderCallback represents webhook callback audit trail
type ProviderCallback struct {
	ID                int64     `json:"id"`
	Provider          string    `json:"provider"`
	CallbackReference string    `json:"callback_reference"`
	RawPayload        string    `json:"raw_payload"` // JSON string
	ProcessedAt       time.Time `json:"processed_at"`
	ReceivedAt        time.Time `json:"received_at"`
}

// WebhookMessage represents Kafka message structure
type WebhookMessage struct {
	Provider  string                 `json:"provider"`
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// CustodyStatusUpdate represents polling result from Custody API
type CustodyStatusUpdate struct {
	TxHash    string `json:"txHash"`
	Status    string `json:"status"`
	Asset     string `json:"asset"`
	Amount    string `json:"amount"`
	Timestamp string `json:"timestamp"`
}
