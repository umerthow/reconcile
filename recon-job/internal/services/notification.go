package services

import (
	"fmt"
	"time"

	"github.com/umerthow/reconcile/recon-job/internal/models"
	"github.com/umerthow/reconcile/recon-job/internal/utils"
)

// NotificationService handles sending alerts for reconciliation results
type NotificationService struct {
	enabled         bool
	slackWebhook    string
	emailRecipients []string
}

func NewNotificationService(enabled bool, slackWebhook string, emailRecipients []string) *NotificationService {
	if enabled {
		utils.Info().Msg("Notification service enabled")
	} else {
		utils.Info().Msg("Notification service disabled")
	}

	return &NotificationService{
		enabled:         enabled,
		slackWebhook:    slackWebhook,
		emailRecipients: emailRecipients,
	}
}

// SendReconciliationSummary sends summary notification after reconciliation run
// MOCK IMPLEMENTATION - Logs for POC
func (s *NotificationService) SendReconciliationSummary(run *models.ReconciliationRun) error {
	if !s.enabled {
		utils.Debug().Msg("Notifications disabled, skipping")
		return nil
	}

	utils.Info().
		Int64("run_id", run.ID).
		Str("run_date", run.RunDate).
		Str("status", run.Status).
		Int("total_processed", run.TotalProcessed).
		Int("total_discrepancies", run.TotalDiscrepancies).
		Msg("[MOCK] Sending reconciliation summary notification")

	// PSEUDOCODE: Real Slack implementation
	// message := s.formatSlackMessage(run)
	// resp, err := http.Post(s.slackWebhook, "application/json", bytes.NewBuffer(message))
	// if err != nil {
	//     return fmt.Errorf("failed to send Slack notification: %w", err)
	// }
	// defer resp.Body.Close()

	// PSEUDOCODE: Real Email implementation
	// emailBody := s.formatEmailBody(run)
	// for _, recipient := range s.emailRecipients {
	//     err := s.sendEmail(recipient, "Reconciliation Summary - " + run.RunDate, emailBody)
	//     if err != nil {
	//         utils.Error().Err(err).Str("recipient", recipient).Msg("Failed to send email")
	//     }
	// }

	return nil
}

// SendDiscrepancyAlert sends immediate alert for critical discrepancies
// MOCK IMPLEMENTATION - Logs for POC
func (s *NotificationService) SendDiscrepancyAlert(discrepancies []models.Discrepancy) error {
	if !s.enabled {
		return nil
	}

	criticalCount := 0
	for _, d := range discrepancies {
		if d.Severity == "critical" {
			criticalCount++
		}
	}

	if criticalCount == 0 {
		return nil
	}

	utils.Info().
		Int("total_discrepancies", len(discrepancies)).
		Int("critical_count", criticalCount).
		Msg("[MOCK] Sending critical discrepancy alert")

	// PSEUDOCODE: Real implementation
	// Send immediate Slack alert for critical issues
	// message := s.formatCriticalAlert(discrepancies)
	// return sendSlackAlert(s.slackWebhook, message, "critical")

	return nil
}

// formatSlackMessage creates Slack-formatted message (for reference)
func (s *NotificationService) formatSlackMessage(run *models.ReconciliationRun) string {
	statusEmoji := "✅"
	if run.Status == "failed" {
		statusEmoji = "❌"
	} else if run.TotalDiscrepancies > 0 {
		statusEmoji = "⚠️"
	}

	message := fmt.Sprintf(`{
		"text": "Reconciliation Summary",
		"blocks": [
			{
				"type": "header",
				"text": {
					"type": "plain_text",
					"text": "%s Reconciliation Summary - %s"
				}
			},
			{
				"type": "section",
				"fields": [
					{"type": "mrkdwn", "text": "*Status:*\n%s"},
					{"type": "mrkdwn", "text": "*Run ID:*\n%d"},
					{"type": "mrkdwn", "text": "*Total Processed:*\n%d"},
					{"type": "mrkdwn", "text": "*Discrepancies:*\n%d"},
					{"type": "mrkdwn", "text": "*Duration:*\n%s"}
				]
			}
		]
	}`, statusEmoji, run.RunDate, run.Status, run.ID, run.TotalProcessed, run.TotalDiscrepancies, run.UpdatedAt.Sub(run.CreatedAt).Round(time.Second))

	return message
}

// formatEmailBody creates email body (for reference)
func (s *NotificationService) formatEmailBody(run *models.ReconciliationRun) string {
	body := fmt.Sprintf(`
Reconciliation Summary Report
================================

Run Date: %s
Run ID: %d
Status: %s

Summary:
--------
Total Transactions Processed: %d
Total Discrepancies Found: %d
Window: %s to %s

Duration: %s

`, run.RunDate, run.ID, run.Status, run.TotalProcessed, run.TotalDiscrepancies,
		run.WindowStart.Format(time.RFC3339),
		run.WindowEnd.Format(time.RFC3339),
		run.UpdatedAt.Sub(run.CreatedAt).Round(time.Second))

	if run.ErrorMessage != nil {
		body += fmt.Sprintf("\nError: %s\n", *run.ErrorMessage)
	}

	return body
}
