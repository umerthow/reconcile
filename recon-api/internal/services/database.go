package services

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/umerthow/reconcile/recon-api/internal/config"
	"github.com/umerthow/reconcile/recon-api/internal/utils"
)

var logger = utils.GetLogger("database")

type DatabaseService struct {
	db *sql.DB
}

func NewDatabaseService(cfg *config.DatabaseConfig) (*DatabaseService, error) {
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	logger.Info().Msg("Database connection established")

	return &DatabaseService{db: db}, nil
}

func (s *DatabaseService) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *DatabaseService) DB() *sql.DB {
	return s.db
}

// Health check
func (s *DatabaseService) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Mock data for POC - Get reconciliation runs
func (s *DatabaseService) GetReconciliationRuns(ctx context.Context, page, pageSize int) ([]map[string]interface{}, int, error) {
	// TODO: Implement actual database query
	// For POC, return mock data

	logger.Info().Int("page", page).Int("page_size", pageSize).Msg("Fetching reconciliation runs (mock)")

	mockRuns := []map[string]interface{}{
		{
			"id":                  1,
			"run_date":            "2026-03-12",
			"window_start":        "2026-03-11T17:00:00Z",
			"window_end":          "2026-03-12T16:59:59Z",
			"status":              "success",
			"total_processed":     50234,
			"total_discrepancies": 12,
			"created_at":          "2026-03-12T02:05:23Z",
			"updated_at":          "2026-03-12T02:15:45Z",
		},
		{
			"id":                  2,
			"run_date":            "2026-03-11",
			"window_start":        "2026-03-10T17:00:00Z",
			"window_end":          "2026-03-11T16:59:59Z",
			"status":              "success",
			"total_processed":     48921,
			"total_discrepancies": 8,
			"created_at":          "2026-03-11T02:04:15Z",
			"updated_at":          "2026-03-11T02:14:32Z",
		},
	}

	return mockRuns, len(mockRuns), nil
}

// Mock data for POC - Get discrepancies
func (s *DatabaseService) GetDiscrepancies(ctx context.Context, filters map[string]string, page, pageSize int) ([]map[string]interface{}, int, error) {
	// TODO: Implement actual database query with filters
	// For POC, return mock data

	logger.Info().
		Interface("filters", filters).
		Int("page", page).
		Int("page_size", pageSize).
		Msg("Fetching discrepancies (mock)")

	mockDiscrepancies := []map[string]interface{}{
		{
			"id":             1,
			"run_id":         1,
			"internal_tx_id": 12345,
			"provider":       "xendit",
			"provider_ref":   "XEN-INV-20260312-001",
			"category":       "amount_mismatch",
			"severity":       "high",
			"status":         "open",
			"details": map[string]interface{}{
				"internal": 500000,
				"provider": 497500,
				"delta":    2500,
			},
			"created_at": "2026-03-12T02:15:45Z",
		},
		{
			"id":           2,
			"run_id":       1,
			"provider":     "custody",
			"provider_ref": "0x1234567890abcdef",
			"category":     "missing_credit",
			"severity":     "high",
			"status":       "open",
			"details": map[string]interface{}{
				"provider_amount": "1.50000000",
				"currency":        "BTC",
			},
			"created_at": "2026-03-12T02:15:46Z",
		},
		{
			"id":             3,
			"run_id":         1,
			"internal_tx_id": 12347,
			"provider":       "xendit",
			"provider_ref":   "XEN-INV-20260312-003",
			"category":       "stale_withdrawal",
			"severity":       "low",
			"status":         "investigating",
			"resolved_by":    "finance@reku.id",
			"details": map[string]interface{}{
				"pending_hours": 52,
			},
			"created_at": "2026-03-12T02:15:47Z",
		},
	}

	return mockDiscrepancies, len(mockDiscrepancies), nil
}

// Mock data for POC - Resolve discrepancy
func (s *DatabaseService) ResolveDiscrepancy(ctx context.Context, id int64, resolvedBy, notes, action string) error {
	// TODO: Implement actual database update with audit logging

	logger.Info().
		Int64("discrepancy_id", id).
		Str("resolved_by", resolvedBy).
		Str("action", action).
		Msg("Resolving discrepancy (mock)")

	// Simulate database update
	return nil
}

// Mock data for POC - Trigger reconciliation run
func (s *DatabaseService) CreateReconciliationRun(ctx context.Context, windowStart, windowEnd time.Time) (int64, error) {
	// TODO: Implement actual database insert

	logger.Info().
		Time("window_start", windowStart).
		Time("window_end", windowEnd).
		Msg("Creating reconciliation run (mock)")

	// Return mock run ID
	return 3, nil
}
