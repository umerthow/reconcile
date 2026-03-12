package models

import "time"

// ReconciliationRun represents a single reconciliation execution
type ReconciliationRun struct {
	ID                 int64     `json:"id"`
	RunDate            string    `json:"run_date"` // YYYY-MM-DD
	WindowStart        time.Time `json:"window_start"`
	WindowEnd          time.Time `json:"window_end"`
	Status             string    `json:"status"` // running, success, failed
	TotalProcessed     int       `json:"total_processed"`
	TotalDiscrepancies int       `json:"total_discrepancies"`
	ErrorMessage       *string   `json:"error_message,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// Discrepancy represents a mismatch found during reconciliation
type Discrepancy struct {
	ID             int64                  `json:"id"`
	RunID          int64                  `json:"run_id"`
	InternalTxID   *int64                 `json:"internal_tx_id,omitempty"`
	Provider       string                 `json:"provider"`
	ProviderRef    string                 `json:"provider_ref"`
	Category       string                 `json:"category"` // missing_credit, stale_withdrawal, amount_mismatch, duplicate, missing_internal
	Severity       string                 `json:"severity"` // low, high
	Status         string                 `json:"status"`   // open, resolved, investigating
	Details        map[string]interface{} `json:"details"`
	ResolvedBy     string                 `json:"resolved_by,omitempty"`
	ResolvedAt     *time.Time             `json:"resolved_at,omitempty"`
	ResolutionNote string                 `json:"resolution_note,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// Transaction represents internal transaction record
type Transaction struct {
	ID          int64     `json:"id"`
	Type        string    `json:"type"` // deposit, withdrawal
	Provider    string    `json:"provider"`
	ProviderRef string    `json:"provider_ref"`
	Currency    string    `json:"currency"`
	Amount      string    `json:"amount"` // Decimal string
	Status      string    `json:"status"`
	UserID      int64     `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ExternalTransaction represents transaction from external source (CSV or API)
type ExternalTransaction struct {
	Provider    string                 `json:"provider"`
	ProviderRef string                 `json:"provider_ref"`
	Currency    string                 `json:"currency"`
	Amount      string                 `json:"amount"` // Decimal string
	Status      string                 `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	RawData     map[string]interface{} `json:"raw_data,omitempty"`
}

// MatchResult represents the result of matching internal vs external transaction
type MatchResult struct {
	Internal        *Transaction
	External        *ExternalTransaction
	Matched         bool
	DiscrepancyType string // "", "amount_mismatch", "missing_credit", "missing_internal", "duplicate"
	AmountDelta     string // For amount mismatch
}
