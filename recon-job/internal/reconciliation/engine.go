package reconciliation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/umerthow/reconcile/recon-job/internal/models"
	"github.com/umerthow/reconcile/recon-job/internal/services"
	"github.com/umerthow/reconcile/recon-job/internal/utils"
)

// Engine handles the core reconciliation logic
type Engine struct {
	db              *services.DatabaseService
	redis           *services.RedisService
	notification    *services.NotificationService
	batchSize       int
	concurrency     int
	amountThreshold float64
}

func NewEngine(db *services.DatabaseService, redis *services.RedisService, notification *services.NotificationService, batchSize, concurrency int, amountThreshold float64) *Engine {
	return &Engine{
		db:              db,
		redis:           redis,
		notification:    notification,
		batchSize:       batchSize,
		concurrency:     concurrency,
		amountThreshold: amountThreshold,
	}
}

// RunReconciliation executes the daily reconciliation process
func (e *Engine) RunReconciliation(ctx context.Context, runDate string, windowStart, windowEnd time.Time) error {
	utils.Info().
		Str("run_date", runDate).
		Time("window_start", windowStart).
		Time("window_end", windowEnd).
		Msg("Starting reconciliation run")

	// Create reconciliation run record
	run := &models.ReconciliationRun{
		RunDate:     runDate,
		WindowStart: windowStart,
		WindowEnd:   windowEnd,
		Status:      "running",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	runID, err := e.db.CreateReconciliationRun(ctx, run)
	if err != nil {
		return fmt.Errorf("failed to create reconciliation run: %w", err)
	}
	run.ID = runID

	// Fetch internal transactions
	utils.Info().Msg("Fetching internal transactions")
	internalTxs, err := e.db.GetInternalTransactions(ctx, windowStart, windowEnd)
	if err != nil {
		return e.handleRunError(ctx, run, fmt.Errorf("failed to fetch internal transactions: %w", err))
	}

	utils.Info().Int("count", len(internalTxs)).Msg("Fetched internal transactions")

	// Fetch provider data
	utils.Info().Msg("Fetching provider data")
	providerData, err := e.fetchProviderData(ctx, runDate, internalTxs)
	if err != nil {
		return e.handleRunError(ctx, run, fmt.Errorf("failed to fetch provider data: %w", err))
	}

	// Process reconciliation in batches
	utils.Info().Msg("Processing reconciliation")
	discrepancies, err := e.processReconciliation(ctx, run.ID, internalTxs, providerData)
	if err != nil {
		return e.handleRunError(ctx, run, fmt.Errorf("failed to process reconciliation: %w", err))
	}

	// Save discrepancies
	if len(discrepancies) > 0 {
		utils.Info().Int("count", len(discrepancies)).Msg("Saving discrepancies")
		if err := e.db.InsertDiscrepancies(ctx, discrepancies); err != nil {
			return e.handleRunError(ctx, run, fmt.Errorf("failed to save discrepancies: %w", err))
		}
	}

	// Update run status
	run.Status = "completed"
	run.TotalProcessed = len(internalTxs)
	run.TotalDiscrepancies = len(discrepancies)
	run.UpdatedAt = time.Now()

	if err := e.db.UpdateReconciliationRun(ctx, run); err != nil {
		return fmt.Errorf("failed to update reconciliation run: %w", err)
	}

	// Send notifications
	utils.Info().Msg("Sending notifications")
	if err := e.notification.SendReconciliationSummary(run); err != nil {
		utils.Error().Err(err).Msg("Failed to send summary notification")
	}

	// Send critical alerts if needed
	criticalDiscrepancies := e.filterCriticalDiscrepancies(discrepancies)
	if len(criticalDiscrepancies) > 0 {
		if err := e.notification.SendDiscrepancyAlert(criticalDiscrepancies); err != nil {
			utils.Error().Err(err).Msg("Failed to send discrepancy alert")
		}
	}

	utils.Info().
		Int("total_processed", run.TotalProcessed).
		Int("total_discrepancies", run.TotalDiscrepancies).
		Dur("duration", run.UpdatedAt.Sub(run.CreatedAt)).
		Msg("Reconciliation run completed successfully")

	return nil
}

// fetchProviderData fetches transaction data from payment providers
// MOCK IMPLEMENTATION - Returns sample data for POC
func (e *Engine) fetchProviderData(ctx context.Context, runDate string, internalTxs []models.Transaction) (map[string]models.ProviderTransaction, error) {
	utils.Info().Msg("[MOCK] Fetching provider data")

	// MOCK DATA: Create provider transactions matching internal transactions
	providerData := make(map[string]models.ProviderTransaction)

	// Simulate Xendit transaction
	providerData["order-001"] = models.ProviderTransaction{
		Provider:        "xendit",
		ProviderTxID:    "xendit-tx-001",
		ProviderRef:     "order-001",
		Amount:          "500000",
		Currency:        "IDR",
		Status:          "completed",
		TransactionDate: time.Now().Add(-1 * time.Hour),
	}

	// Simulate Custody transaction with slight amount difference (for testing)
	providerData["0xabc123"] = models.ProviderTransaction{
		Provider:        "custody",
		ProviderTxID:    "custody-tx-001",
		ProviderRef:     "0xabc123",
		Amount:          "0.049", // Intentional difference: 0.049 vs 0.05
		Currency:        "BTC",
		Status:          "completed",
		TransactionDate: time.Now().Add(-2 * time.Hour),
	}

	// PSEUDOCODE: Real implementation
	// Group transactions by provider
	// providerGroups := make(map[string][]models.Transaction)
	// for _, tx := range internalTxs {
	//     providerGroups[tx.Provider] = append(providerGroups[tx.Provider], tx)
	// }
	//
	// providerData := make(map[string]models.ProviderTransaction)
	//
	// // Fetch from Xendit API
	// if xenditTxs, ok := providerGroups["xendit"]; ok {
	//     refs := extractRefs(xenditTxs)
	//     xenditData, err := e.fetchXenditData(ctx, runDate, refs)
	//     if err != nil {
	//         return nil, err
	//     }
	//     for k, v := range xenditData {
	//         providerData[k] = v
	//     }
	// }
	//
	// // Fetch from Custody API
	// if custodyTxs, ok := providerGroups["custody"]; ok {
	//     refs := extractRefs(custodyTxs)
	//     custodyData, err := e.fetchCustodyData(ctx, runDate, refs)
	//     if err != nil {
	//         return nil, err
	//     }
	//     for k, v := range custodyData {
	//         providerData[k] = v
	//     }
	// }

	return providerData, nil
}

// processReconciliation compares internal vs provider data and detects discrepancies
func (e *Engine) processReconciliation(ctx context.Context, runID int64, internalTxs []models.Transaction, providerData map[string]models.ProviderTransaction) ([]models.Discrepancy, error) {
	var discrepancies []models.Discrepancy
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Process transactions concurrently in batches
	txChan := make(chan models.Transaction, e.batchSize)

	// Start workers
	for i := 0; i < e.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for tx := range txChan {
				if d := e.reconcileTransaction(runID, tx, providerData); d != nil {
					mu.Lock()
					discrepancies = append(discrepancies, *d)
					mu.Unlock()
				}
			}
		}()
	}

	// Feed transactions to workers
	for _, tx := range internalTxs {
		select {
		case txChan <- tx:
		case <-ctx.Done():
			close(txChan)
			return nil, ctx.Err()
		}
	}
	close(txChan)

	// Wait for all workers to complete
	wg.Wait()

	return discrepancies, nil
}

// reconcileTransaction compares a single transaction against provider data
func (e *Engine) reconcileTransaction(runID int64, tx models.Transaction, providerData map[string]models.ProviderTransaction) *models.Discrepancy {
	// Look up provider transaction by reference
	providerTx, exists := providerData[tx.ProviderRef]
	if !exists {
		// Missing transaction in provider data
		utils.Warn().
			Int64("tx_id", tx.ID).
			Str("provider_ref", tx.ProviderRef).
			Msg("Transaction not found in provider data")

		txID := tx.ID
		return &models.Discrepancy{
			RunID:        runID,
			InternalTxID: &txID,
			Provider:     tx.Provider,
			ProviderRef:  tx.ProviderRef,
			Category:     "missing_provider",
			Severity:     "critical",
			Status:       "open",
			Details: map[string]interface{}{
				"message":         "Transaction not found in provider data",
				"internal_amount": tx.Amount,
				"currency":        tx.Currency,
			},
			CreatedAt: time.Now(),
		}
	}

	// Compare amounts
	internalAmount, err := decimal.NewFromString(tx.Amount)
	if err != nil {
		utils.Error().Err(err).Int64("tx_id", tx.ID).Msg("Failed to parse internal amount")
		return nil
	}

	providerAmount, err := decimal.NewFromString(providerTx.Amount)
	if err != nil {
		utils.Error().Err(err).Str("provider_tx_id", providerTx.ProviderTxID).Msg("Failed to parse provider amount")
		return nil
	}

	amountDiff := internalAmount.Sub(providerAmount).Abs()
	thresholdDecimal := decimal.NewFromFloat(e.amountThreshold)

	if amountDiff.GreaterThan(thresholdDecimal) {
		// Amount mismatch
		utils.Warn().
			Int64("tx_id", tx.ID).
			Str("internal_amount", internalAmount.String()).
			Str("provider_amount", providerAmount.String()).
			Str("difference", amountDiff.String()).
			Msg("Amount mismatch detected")

		severity := "minor"
		if amountDiff.GreaterThan(decimal.NewFromFloat(e.amountThreshold * 10)) {
			severity = "major"
		}

		txID := tx.ID
		return &models.Discrepancy{
			RunID:        runID,
			InternalTxID: &txID,
			Provider:     tx.Provider,
			ProviderRef:  tx.ProviderRef,
			Category:     "amount_mismatch",
			Severity:     severity,
			Status:       "open",
			Details: map[string]interface{}{
				"internal_amount": internalAmount.String(),
				"provider_amount": providerAmount.String(),
				"difference":      amountDiff.String(),
				"currency":        tx.Currency,
				"provider_tx_id":  providerTx.ProviderTxID,
			},
			CreatedAt: time.Now(),
		}
	}

	// Status mismatch
	if tx.Status != providerTx.Status {
		utils.Warn().
			Int64("tx_id", tx.ID).
			Str("internal_status", tx.Status).
			Str("provider_status", providerTx.Status).
			Msg("Status mismatch detected")

		txID := tx.ID
		return &models.Discrepancy{
			RunID:        runID,
			InternalTxID: &txID,
			Provider:     tx.Provider,
			ProviderRef:  tx.ProviderRef,
			Category:     "status_mismatch",
			Severity:     "minor",
			Status:       "open",
			Details: map[string]interface{}{
				"internal_status": tx.Status,
				"provider_status": providerTx.Status,
				"provider_tx_id":  providerTx.ProviderTxID,
			},
			CreatedAt: time.Now(),
		}
	}

	// No discrepancy
	return nil
}

// filterCriticalDiscrepancies returns only critical severity discrepancies
func (e *Engine) filterCriticalDiscrepancies(discrepancies []models.Discrepancy) []models.Discrepancy {
	var critical []models.Discrepancy
	for _, d := range discrepancies {
		if d.Severity == "critical" {
			critical = append(critical, d)
		}
	}
	return critical
}

// handleRunError updates run status to failed and returns error
func (e *Engine) handleRunError(ctx context.Context, run *models.ReconciliationRun, err error) error {
	utils.Error().Err(err).Msg("Reconciliation run failed")

	run.Status = "failed"
	errMsg := err.Error()
	run.ErrorMessage = &errMsg
	run.UpdatedAt = time.Now()

	if updateErr := e.db.UpdateReconciliationRun(ctx, run); updateErr != nil {
		utils.Error().Err(updateErr).Msg("Failed to update run status")
	}

	return err
}
