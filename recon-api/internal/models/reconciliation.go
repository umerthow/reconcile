package models

import "time"

// ReconciliationRun represents a reconciliation run
type ReconciliationRun struct {
	ID                 int64     `json:"id" db:"id"`
	RunDate            string    `json:"run_date" db:"run_date"`
	WindowStart        time.Time `json:"window_start" db:"window_start"`
	WindowEnd          time.Time `json:"window_end" db:"window_end"`
	Status             string    `json:"status" db:"status"` // running, success, failed
	TotalProcessed     int       `json:"total_processed" db:"total_processed"`
	TotalDiscrepancies int       `json:"total_discrepancies" db:"total_discrepancies"`
	ErrorMessage       *string   `json:"error_message,omitempty" db:"error_message"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// ReconciliationRunListResponse represents the list response
type ReconciliationRunListResponse struct {
	Runs     []ReconciliationRun `json:"runs"`
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

// TriggerReconciliationRequest represents a manual trigger request
type TriggerReconciliationRequest struct {
	WindowStart time.Time `json:"window_start" binding:"required"`
	WindowEnd   time.Time `json:"window_end" binding:"required"`
	DryRun      bool      `json:"dry_run"`
}

// TriggerReconciliationResponse represents the trigger response
type TriggerReconciliationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	RunID   int64  `json:"run_id,omitempty"`
}
