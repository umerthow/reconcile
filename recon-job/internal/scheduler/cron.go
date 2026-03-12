package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/umerthow/reconcile/recon-job/internal/reconciliation"
	"github.com/umerthow/reconcile/recon-job/internal/services"
	"github.com/umerthow/reconcile/recon-job/internal/utils"
)

// Scheduler handles cron-like scheduling of reconciliation jobs
type Scheduler struct {
	engine   *reconciliation.Engine
	redis    *services.RedisService
	schedule string // Cron expression: "0 2 * * *" = 02:00 daily
	timezone *time.Location
	lockKey  string
	lockTTL  time.Duration
	stopChan chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

func NewScheduler(engine *reconciliation.Engine, redis *services.RedisService, schedule, timezoneName string) (*Scheduler, error) {
	location, err := time.LoadLocation(timezoneName)
	if err != nil {
		return nil, fmt.Errorf("failed to load timezone: %w", err)
	}

	return &Scheduler{
		engine:   engine,
		redis:    redis,
		schedule: schedule,
		timezone: location,
		lockKey:  "recon:job:lock",
		lockTTL:  2 * time.Hour, // Lock expires after 2 hours
		stopChan: make(chan struct{}),
	}, nil
}

// Start begins the scheduler loop
func (s *Scheduler) Start(ctx context.Context) error {
	utils.Info().
		Str("schedule", s.schedule).
		Str("timezone", s.timezone.String()).
		Msg("Starting reconciliation scheduler")

	s.wg.Add(1)
	go s.run(ctx)

	return nil
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop(ctx context.Context) error {
	utils.Info().Msg("Stopping reconciliation scheduler")

	s.stopOnce.Do(func() {
		close(s.stopChan)
	})

	// Wait for running jobs to complete (with timeout)
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		utils.Info().Msg("Scheduler stopped gracefully")
		return nil
	case <-ctx.Done():
		utils.Warn().Msg("Scheduler stop timed out")
		return ctx.Err()
	}
}

// run is the main scheduler loop
func (s *Scheduler) run(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	var lastRun time.Time

	for {
		select {
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			if s.shouldRun(now, lastRun) {
				utils.Info().Time("scheduled_time", now).Msg("Triggering scheduled reconciliation")

				// Run reconciliation in goroutine to avoid blocking
				s.wg.Add(1)
				go func(runTime time.Time) {
					defer s.wg.Done()

					if err := s.executeReconciliation(ctx, runTime); err != nil {
						utils.Error().Err(err).Msg("Scheduled reconciliation failed")
					}
				}(now)

				lastRun = now
			}
		}
	}
}

// shouldRun determines if reconciliation should run based on schedule
func (s *Scheduler) shouldRun(now, lastRun time.Time) bool {
	// Convert to configured timezone
	nowInTZ := now.In(s.timezone)

	// Parse cron schedule (simplified: only supports "0 2 * * *" format for now)
	// For POC: hard-coded to run at 02:00 daily
	targetHour := 2
	targetMinute := 0

	// Check if current time matches schedule
	if nowInTZ.Hour() != targetHour || nowInTZ.Minute() != targetMinute {
		return false
	}

	// Prevent duplicate runs within the same minute
	if !lastRun.IsZero() && now.Sub(lastRun) < 1*time.Minute {
		return false
	}

	return true
}

// executeReconciliation runs the reconciliation job
func (s *Scheduler) executeReconciliation(ctx context.Context, runTime time.Time) error {
	// Try to acquire distributed lock
	acquired, err := s.redis.AcquireLock(ctx, s.lockKey, s.lockTTL)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !acquired {
		utils.Warn().Msg("Another instance is already running reconciliation, skipping")
		return nil
	}

	// Release lock when done
	defer func() {
		if err := s.redis.ReleaseLock(ctx, s.lockKey); err != nil {
			utils.Error().Err(err).Msg("Failed to release lock")
		}
	}()

	// Calculate reconciliation window (previous day)
	runDate := runTime.AddDate(0, 0, -1).Format("2006-01-02")
	windowStart := runTime.AddDate(0, 0, -1).Truncate(24 * time.Hour)
	windowEnd := windowStart.Add(24 * time.Hour)

	utils.Info().
		Str("run_date", runDate).
		Time("window_start", windowStart).
		Time("window_end", windowEnd).
		Msg("Executing reconciliation")

	// Run reconciliation
	startTime := time.Now()
	err = s.engine.RunReconciliation(ctx, runDate, windowStart, windowEnd)
	duration := time.Since(startTime)

	if err != nil {
		utils.Error().
			Err(err).
			Dur("duration", duration).
			Msg("Reconciliation execution failed")
		return err
	}

	utils.Info().
		Dur("duration", duration).
		Msg("Reconciliation execution completed successfully")

	return nil
}

// RunNow triggers an immediate reconciliation (for testing/manual execution)
func (s *Scheduler) RunNow(ctx context.Context) error {
	utils.Info().Msg("Triggering immediate reconciliation")
	return s.executeReconciliation(ctx, time.Now())
}
