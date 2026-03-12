package models

import "time"

// ProviderTransaction represents transaction data from payment provider
type ProviderTransaction struct {
	Provider        string    `json:"provider"`
	ProviderTxID    string    `json:"provider_tx_id"`
	ProviderRef     string    `json:"provider_ref"`
	Amount          string    `json:"amount"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	TransactionDate time.Time `json:"transaction_date"`
}
