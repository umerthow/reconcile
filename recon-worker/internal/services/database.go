package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/umerthow/reconcile/recon-worker/internal/utils"
)

type DatabaseService struct {
	db *sql.DB
}

func NewDatabaseService(connectionString string) (*DatabaseService, error) {
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Connection pool settings
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	utils.Info().Msg("Database connection established")

	return &DatabaseService{db: db}, nil
}

func (s *DatabaseService) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// UpdateTransactionFromWebhook updates transaction record from webhook data
// MOCK IMPLEMENTATION - Replace with real SQL queries
func (s *DatabaseService) UpdateTransactionFromWebhook(ctx context.Context, provider, providerRef, status string, incomingTimestamp time.Time) error {
	utils.Info().
		Str("provider", provider).
		Str("provider_ref", providerRef).
		Str("status", status).
		Time("incoming_timestamp", incomingTimestamp).
		Msg("[MOCK] Would update transaction in database")

	// PSEUDOCODE: Real implementation would be:
	// query := `
	//     UPDATE transactions
	//     SET status = ?,
	//         incoming_timestamp = ?,
	//         updated_at = NOW()
	//     WHERE provider = ?
	//       AND provider_ref = ?
	//       AND (incoming_timestamp IS NULL OR incoming_timestamp < ?)
	// `
	// result, err := s.db.ExecContext(ctx, query, status, incomingTimestamp, provider, providerRef, incomingTimestamp)
	// if err != nil {
	//     return fmt.Errorf("failed to update transaction: %w", err)
	// }
	// rowsAffected, _ := result.RowsAffected()
	// if rowsAffected == 0 {
	//     utils.Warn().Str("provider_ref", providerRef).Msg("No rows updated - possibly stale webhook")
	// }

	return nil
}

// InsertProviderCallback records webhook in audit trail
// MOCK IMPLEMENTATION - Replace with real SQL queries
func (s *DatabaseService) InsertProviderCallback(ctx context.Context, provider, callbackRef, rawPayload string, receivedAt time.Time) error {
	utils.Info().
		Str("provider", provider).
		Str("callback_ref", callbackRef).
		Time("received_at", receivedAt).
		Msg("[MOCK] Would insert provider callback to database")

	// PSEUDOCODE: Real implementation would be:
	// query := `
	//     INSERT INTO provider_callbacks (provider, callback_reference, raw_payload, received_at, processed_at)
	//     VALUES (?, ?, ?, ?, NOW())
	// `
	// _, err := s.db.ExecContext(ctx, query, provider, callbackRef, rawPayload, receivedAt)
	// if err != nil {
	//     return fmt.Errorf("failed to insert provider callback: %w", err)
	// }

	return nil
}

// GetPendingWithdrawals fetches pending crypto withdrawals for polling
// MOCK IMPLEMENTATION - Returns hardcoded data for POC
func (s *DatabaseService) GetPendingWithdrawals(ctx context.Context) ([]string, error) {
	utils.Info().Msg("[MOCK] Fetching pending withdrawals")

	// MOCK DATA: Return sample tx hashes for POC
	mockPendingTxs := []string{
		"0xabc123...",
		"0xdef456...",
		"0xghi789...",
	}

	// PSEUDOCODE: Real implementation would be:
	// query := `
	//     SELECT provider_ref
	//     FROM transactions
	//     WHERE provider = 'custody'
	//       AND type = 'withdrawal'
	//       AND status = 'pending'
	//       AND created_at > DATE_SUB(NOW(), INTERVAL 48 HOUR)
	// `
	// rows, err := s.db.QueryContext(ctx, query)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to query pending withdrawals: %w", err)
	// }
	// defer rows.Close()
	//
	// var txHashes []string
	// for rows.Next() {
	//     var txHash string
	//     if err := rows.Scan(&txHash); err != nil {
	//         return nil, err
	//     }
	//     txHashes = append(txHashes, txHash)
	// }

	return mockPendingTxs, nil
}

// Health check for database connection
func (s *DatabaseService) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
