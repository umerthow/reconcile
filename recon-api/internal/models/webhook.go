package models

import "time"

// WebhookRequest represents an incoming webhook payload
type WebhookRequest struct {
	Provider  string                 `json:"provider" binding:"required"`
	EventType string                 `json:"event_type" binding:"required"`
	Data      map[string]interface{} `json:"data" binding:"required"`
	Signature string                 `json:"-"` // From header
	Timestamp time.Time              `json:"timestamp"`
}

// XenditWebhook represents Xendit webhook structure
type XenditWebhook struct {
	ID            string  `json:"id"`
	ExternalID    string  `json:"external_id"`
	Status        string  `json:"status"`
	Amount        float64 `json:"amount"`
	PaidAmount    float64 `json:"paid_amount"`
	FeesPaid      float64 `json:"fees_paid"`
	PaymentMethod string  `json:"payment_method"`
	Currency      string  `json:"currency"`
	Created       string  `json:"created"`
	Updated       string  `json:"updated"`
}

// CustodyWebhook represents Custody provider webhook structure
type CustodyWebhook struct {
	TxHash    string `json:"txHash"`
	Asset     string `json:"asset"`
	Amount    string `json:"amount"` // String to preserve precision
	Status    string `json:"status"` // COMPLETED, PENDING, FAILED
	Timestamp string `json:"timestamp"`
}

// WebhookResponse represents the webhook response
type WebhookResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ID      string `json:"id,omitempty"`
}
