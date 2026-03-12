package polling

import (
	"context"
	"time"

	"github.com/umerthow/reconcile/recon-worker/internal/models"
	"github.com/umerthow/reconcile/recon-worker/internal/services"
	"github.com/umerthow/reconcile/recon-worker/internal/utils"
)

type CustodyPoller struct {
	db         *services.DatabaseService
	custodyAPI *services.CustodyAPIService
	kafka      *services.KafkaService
	interval   time.Duration
	stopChan   chan struct{}
}

func NewCustodyPoller(
	db *services.DatabaseService,
	custodyAPI *services.CustodyAPIService,
	kafka *services.KafkaService,
	intervalSec int,
) *CustodyPoller {
	return &CustodyPoller{
		db:         db,
		custodyAPI: custodyAPI,
		kafka:      kafka,
		interval:   time.Duration(intervalSec) * time.Second,
		stopChan:   make(chan struct{}),
	}
}

// Start begins the polling loop
func (p *CustodyPoller) Start(ctx context.Context) {
	utils.Info().
		Dur("interval", p.interval).
		Msg("Starting Custody API polling loop")

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	// Run immediately on start
	p.poll(ctx)

	for {
		select {
		case <-ticker.C:
			p.poll(ctx)
		case <-p.stopChan:
			utils.Info().Msg("Custody poller stopped")
			return
		case <-ctx.Done():
			utils.Info().Msg("Custody poller context cancelled")
			return
		}
	}
}

// Stop gracefully stops the poller
func (p *CustodyPoller) Stop() {
	close(p.stopChan)
}

func (p *CustodyPoller) poll(ctx context.Context) {
	startTime := time.Now()

	utils.Info().Msg("Starting Custody API poll cycle")

	// Step 1: Get pending withdrawals from database
	pendingTxHashes, err := p.db.GetPendingWithdrawals(ctx)
	if err != nil {
		utils.Error().Err(err).Msg("Failed to get pending withdrawals")
		return
	}

	if len(pendingTxHashes) == 0 {
		utils.Info().Msg("No pending withdrawals to poll")
		return
	}

	utils.Info().
		Int("count", len(pendingTxHashes)).
		Msg("Found pending withdrawals")

	// Step 2: Fetch statuses from Custody API (with circuit breaker)
	updates, err := p.custodyAPI.FetchPendingTransactionStatuses(ctx, pendingTxHashes)
	if err != nil {
		utils.Error().Err(err).Msg("Failed to fetch custody statuses")
		return
	}

	utils.Info().
		Int("updates_count", len(updates)).
		Msg("Received custody status updates")

	// Step 3: Publish updates to Kafka (custody-status-updates topic)
	// The custody consumer will then process these updates
	for _, update := range updates {
		if err := p.publishStatusUpdate(ctx, update); err != nil {
			utils.Error().
				Err(err).
				Str("tx_hash", update.TxHash).
				Msg("Failed to publish status update")
			continue
		}

		utils.Debug().
			Str("tx_hash", update.TxHash).
			Str("status", update.Status).
			Msg("Published status update to Kafka")
	}

	duration := time.Since(startTime)
	utils.Info().
		Int("pending_count", len(pendingTxHashes)).
		Int("updates_count", len(updates)).
		Dur("duration", duration).
		Msg("Custody poll cycle completed")
}

func (p *CustodyPoller) publishStatusUpdate(ctx context.Context, update models.CustodyStatusUpdate) error {
	// Format as webhook-compatible message for consistency
	message := models.WebhookMessage{
		Provider:  "custody",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: map[string]interface{}{
			"txHash":    update.TxHash,
			"status":    update.Status,
			"asset":     update.Asset,
			"amount":    update.Amount,
			"timestamp": update.Timestamp,
			"source":    "polling", // Distinguish from webhook
		},
	}

	return p.kafka.PublishMessage(ctx, "custody-status-updates", message)
}
