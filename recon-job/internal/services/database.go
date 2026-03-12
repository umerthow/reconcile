package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/umerthow/reconcile/recon-job/internal/models"
	"github.com/umerthow/reconcile/recon-job/internal/utils"
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

// CreateReconciliationRun inserts a new reconciliation run
// MOCK IMPLEMENTATION - Returns success for POC
func (s *DatabaseService) CreateReconciliationRun(ctx context.Context, run *models.ReconciliationRun) (int64, error) {
	utils.Info().
		Str("run_date", run.RunDate).
		Time("window_start", run.WindowStart).
		Time("window_end", run.WindowEnd).
		Msg("[MOCK] Creating reconciliation run")

	// PSEUDOCODE: Real implementation
	// query := `
	//     INSERT INTO reconciliation_runs (run_date, window_start, window_end, status, created_at, updated_at)
	//     VALUES (?, ?, ?, ?, NOW(), NOW())
	// `
	// result, err := s.db.ExecContext(ctx, query, run.RunDate, run.WindowStart, run.WindowEnd, "running")
	// if err != nil {
	//     return 0, fmt.Errorf("failed to create reconciliation run: %w", err)
	// }
	// return result.LastInsertId()

	// Mock: Return fake ID
	return 1, nil
}

// UpdateReconciliationRun updates run status and counts
// MOCK IMPLEMENTATION - Logs for POC
func (s *DatabaseService) UpdateReconciliationRun(ctx context.Context, run *models.ReconciliationRun) error {
	utils.Info().
		Int64("run_id", run.ID).
		Str("status", run.Status).
		Int("total_processed", run.TotalProcessed).
		Int("total_discrepancies", run.TotalDiscrepancies).
		Msg("[MOCK] Updating reconciliation run")

	// PSEUDOCODE: Real implementation
	// query := `
	//     UPDATE reconciliation_runs
	//     SET status = ?, total_processed = ?, total_discrepancies = ?, error_message = ?, updated_at = NOW()
	//     WHERE id = ?
	// `
	// _, err := s.db.ExecContext(ctx, query, run.Status, run.TotalProcessed, run.TotalDiscrepancies, run.ErrorMessage, run.ID)
	// return err

	return nil
}

// InsertDiscrepancies bulk inserts discrepancies
// MOCK IMPLEMENTATION - Logs for POC
func (s *DatabaseService) InsertDiscrepancies(ctx context.Context, discrepancies []models.Discrepancy) error {
	if len(discrepancies) == 0 {
		return nil
	}

	utils.Info().
		Int("count", len(discrepancies)).
		Msg("[MOCK] Inserting discrepancies")

	// PSEUDOCODE: Real implementation
	// query := `
	//     INSERT INTO discrepancies (run_id, internal_tx_id, provider, provider_ref, category, severity, status, details, created_at)
	//     VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())
	// `
	// stmt, err := s.db.PrepareContext(ctx, query)
	// if err != nil {
	//     return err
	// }
	// defer stmt.Close()
	//
	// for _, d := range discrepancies {
	//     detailsJSON, _ := json.Marshal(d.Details)
	//     _, err := stmt.ExecContext(ctx, d.RunID, d.InternalTxID, d.Provider, d.ProviderRef,
	//         d.Category, d.Severity, d.Status, string(detailsJSON))
	//     if err != nil {
	//         return err
	//     }
	// }

	return nil
}

// GetInternalTransactions fetches transactions from internal ledger for time window
// MOCK IMPLEMENTATION - Returns sample data for POC
func (s *DatabaseService) GetInternalTransactions(ctx context.Context, windowStart, windowEnd time.Time) ([]models.Transaction, error) {
	utils.Info().
		Time("window_start", windowStart).
		Time("window_end", windowEnd).
		Msg("[MOCK] Fetching internal transactions")

	// MOCK DATA: Return sample transactions
	mockTransactions := []models.Transaction{
		{
			ID:          101,
			Type:        "deposit",
			Provider:    "xendit",
			ProviderRef: "order-001",
			Currency:    "IDR",
			Amount:      "500000",
			Status:      "completed",
			UserID:      1,
			CreatedAt:   windowStart.Add(2 * time.Hour),
			UpdatedAt:   windowStart.Add(2 * time.Hour),
		},
		{
			ID:          102,
			Type:        "withdrawal",
			Provider:    "custody",
			ProviderRef: "0xabc123",
			Currency:    "BTC",
			Amount:      "0.05",
			Status:      "completed",
			UserID:      2,
			CreatedAt:   windowStart.Add(5 * time.Hour),
			UpdatedAt:   windowStart.Add(6 * time.Hour),
		},
	}

	// PSEUDOCODE: Real implementation
	// query := `
	//     SELECT id, type, provider, provider_ref, currency, amount, status, user_id, created_at, updated_at
	//     FROM transactions
	//     WHERE created_at >= ? AND created_at < ?
	//     ORDER BY created_at ASC
	// `
	// rows, err := s.db.QueryContext(ctx, query, windowStart, windowEnd)
	// if err != nil {
	//     return nil, err
	// }
	// defer rows.Close()
	//
	// var transactions []models.Transaction
	// for rows.Next() {
	//     var tx models.Transaction
	//     err := rows.Scan(&tx.ID, &tx.Type, &tx.Provider, &tx.ProviderRef, &tx.Currency, &tx.Amount, &tx.Status, &tx.UserID, &tx.CreatedAt, &tx.UpdatedAt)
	//     if err != nil {
	//         return nil, err
	//     }
	//     transactions = append(transactions, tx)
	// }

	return mockTransactions, nil
}
