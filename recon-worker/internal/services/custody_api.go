package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
	"github.com/umerthow/reconcile/recon-worker/internal/config"
	"github.com/umerthow/reconcile/recon-worker/internal/models"
	"github.com/umerthow/reconcile/recon-worker/internal/utils"
)

type CustodyAPIService struct {
	config         *config.Config
	httpClient     *http.Client
	circuitBreaker *gobreaker.CircuitBreaker
}

func NewCustodyAPIService(cfg *config.Config) *CustodyAPIService {
	// Circuit breaker settings
	cbSettings := gobreaker.Settings{
		Name:        "CustodyAPI",
		MaxRequests: 3,
		Interval:    time.Minute,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			utils.Warn().
				Str("circuit", name).
				Str("from", from.String()).
				Str("to", to.String()).
				Msg("Circuit breaker state changed")
		},
	}

	return &CustodyAPIService{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		circuitBreaker: gobreaker.NewCircuitBreaker(cbSettings),
	}
}

// FetchPendingTransactionStatuses polls Custody API for pending withdrawals
// MOCK IMPLEMENTATION - Returns hardcoded data for POC
func (s *CustodyAPIService) FetchPendingTransactionStatuses(ctx context.Context, txHashes []string) ([]models.CustodyStatusUpdate, error) {
	if len(txHashes) == 0 {
		return []models.CustodyStatusUpdate{}, nil
	}

	utils.Info().
		Int("count", len(txHashes)).
		Msg("[MOCK] Fetching custody transaction statuses")

	// MOCK DATA: Return hardcoded updates for POC
	mockUpdates := []models.CustodyStatusUpdate{
		{
			TxHash:    txHashes[0],
			Status:    "completed",
			Asset:     "BTC",
			Amount:    "0.05",
			Timestamp: time.Now().Format(time.RFC3339),
		},
	}

	// Simulate API call delay
	time.Sleep(100 * time.Millisecond)

	// PSEUDOCODE: Real implementation would be:
	// result, err := s.circuitBreaker.Execute(func() (interface{}, error) {
	//     req, err := http.NewRequestWithContext(ctx, "POST", s.config.CustodyAPIURL+"/transactions/batch-status", nil)
	//     if err != nil {
	//         return nil, err
	//     }
	//
	//     req.Header.Set("Authorization", "Bearer "+s.config.CustodyAPIKey)
	//     req.Header.Set("Content-Type", "application/json")
	//
	//     // Add request body with txHashes
	//     body := map[string]interface{}{"tx_hashes": txHashes}
	//     jsonBody, _ := json.Marshal(body)
	//     req.Body = io.NopCloser(bytes.NewReader(jsonBody))
	//
	//     resp, err := s.httpClient.Do(req)
	//     if err != nil {
	//         return nil, fmt.Errorf("custody API request failed: %w", err)
	//     }
	//     defer resp.Body.Close()
	//
	//     if resp.StatusCode != http.StatusOK {
	//         return nil, fmt.Errorf("custody API returned status %d", resp.StatusCode)
	//     }
	//
	//     var response struct {
	//         Data []models.CustodyStatusUpdate `json:"data"`
	//     }
	//     if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
	//         return nil, err
	//     }
	//
	//     return response.Data, nil
	// })
	//
	// if err != nil {
	//     return nil, err
	// }
	//
	// return result.([]models.CustodyStatusUpdate), nil

	return mockUpdates, nil
}

// FetchTransactionStatus fetches single transaction status (unused in current flow, for future use)
func (s *CustodyAPIService) FetchTransactionStatus(ctx context.Context, txHash string) (*models.CustodyStatusUpdate, error) {
	utils.Info().
		Str("tx_hash", txHash).
		Msg("[MOCK] Fetching single custody transaction status")

	// MOCK DATA
	mockUpdate := &models.CustodyStatusUpdate{
		TxHash:    txHash,
		Status:    "completed",
		Asset:     "ETH",
		Amount:    "1.5",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Simulate circuit breaker and API call
	_, err := s.circuitBreaker.Execute(func() (interface{}, error) {
		time.Sleep(50 * time.Millisecond)
		return mockUpdate, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction status: %w", err)
	}

	return mockUpdate, nil
}

// makeRequest is a helper for actual HTTP requests (not used in mock)
func (s *CustodyAPIService) makeRequest(ctx context.Context, method, endpoint string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, s.config.CustodyAPIURL+endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.config.CustodyAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("custody API error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
