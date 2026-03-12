package models

import "time"

// Discrepancy represents a reconciliation discrepancy
type Discrepancy struct {
	ID              int64                  `json:"id" db:"id"`
	RunID           int64                  `json:"run_id" db:"run_id"`
	InternalTxID    *int64                 `json:"internal_tx_id,omitempty" db:"internal_tx_id"`
	Provider        string                 `json:"provider" db:"provider"`
	ProviderRef     string                 `json:"provider_ref" db:"provider_ref"`
	Category        string                 `json:"category" db:"category"`
	Severity        string                 `json:"severity" db:"severity"` // low, high
	Status          string                 `json:"status" db:"status"`     // open, resolved, investigating
	Details         map[string]interface{} `json:"details,omitempty" db:"details"`
	ResolvedBy      *string                `json:"resolved_by,omitempty" db:"resolved_by"`
	ResolvedAt      *time.Time             `json:"resolved_at,omitempty" db:"resolved_at"`
	ResolutionNotes *string                `json:"resolution_notes,omitempty" db:"resolution_notes"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// DiscrepancyListResponse represents the list response
type DiscrepancyListResponse struct {
	Discrepancies []Discrepancy `json:"discrepancies"`
	Total         int           `json:"total"`
	Page          int           `json:"page"`
	PageSize      int           `json:"page_size"`
	Filters       FilterSummary `json:"filters"`
}

// FilterSummary provides summary of applied filters
type FilterSummary struct {
	Status   string `json:"status,omitempty"`
	Severity string `json:"severity,omitempty"`
	Provider string `json:"provider,omitempty"`
}

// ResolveDiscrepancyRequest represents a resolution request
type ResolveDiscrepancyRequest struct {
	ResolvedBy      string `json:"resolved_by" binding:"required"`
	ResolutionNotes string `json:"resolution_notes" binding:"required"`
	Action          string `json:"action" binding:"required,oneof=resolve investigate"`
}

// ResolveDiscrepancyResponse represents the resolution response
type ResolveDiscrepancyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
