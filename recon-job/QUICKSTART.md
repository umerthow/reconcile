# recon-job Quick Start Guide

Get the reconciliation job service running in 5 minutes.

## What is recon-job?

Daily scheduled job that:
- Runs automatically at 02:00 WIB every day
- Compares yesterday's transactions against payment provider data
- Detects discrepancies (missing transactions, amount/status mismatches)
- Sends notifications to Slack/email for critical issues

## Quick Start (Docker Compose)

### 1. Start All Services

From the root `reconcile/` directory:

```bash
# Start entire system (includes recon-job)
docker-compose up -d

# Verify recon-job is running
docker ps | grep recon-job

# View logs
docker logs -f recon-job
```

Expected output:
```
INFO Starting Reconciliation Job Service
INFO Configuration loaded cron_schedule="0 2 * * *" timezone=Asia/Jakarta
INFO Database connection established
INFO Redis connection established
INFO Starting reconciliation scheduler
INFO Reconciliation Job Service started successfully
INFO Waiting for scheduled execution or shutdown signal...
```

### 2. Configuration

Edit `docker-compose.yml` recon-job section:

```yaml
recon-job:
  environment:
    CRON_SCHEDULE: "0 2 * * *"     # 02:00 daily (cron format)
    TIMEZONE: "Asia/Jakarta"        # Adjust for your timezone
    BATCH_SIZE: 1000                # Transactions per batch
    CONCURRENCY: 5                  # Worker goroutines
    DISCREPANCY_THRESHOLD: 0.01     # Min amount diff to flag (0.01 = 1 cent)
    NOTIFICATION_ENABLED: "true"    # Enable Slack/email alerts
    SLACK_WEBHOOK: "https://..."    # Your Slack webhook URL
    EMAIL_RECIPIENTS: "team@..."    # Comma-separated emails
```

### 3. Manual Trigger (Testing)

To test reconciliation without waiting for scheduled time:

```bash
# Connect to container
docker exec -it recon-job sh

# Trigger immediate run (TODO: add HTTP endpoint or CLI flag)
# For now, restart container to test startup and configuration
docker restart recon-job && docker logs -f recon-job
```

## Local Development

### 1. Setup

```bash
cd recon-job

# Copy environment template
cp .env.example .env

# Edit configuration
vim .env
```

### 2. Start Dependencies

```bash
# From root directory, start only infrastructure
docker-compose up -d mysql redis

# Wait for services to be healthy
docker-compose ps
```

### 3. Run Locally

```bash
# Install dependencies
go mod download

# Build
make build

# Run
make run
```

Expected log output:
```
5:22PM INF Starting Reconciliation Job Service
5:22PM INF Configuration loaded cron_schedule="0 2 * * *" timezone=Asia/Jakarta batch_size=1000 concurrency=5
5:22PM INF Database connection established
5:22PM INF Redis connection established
5:22PM INF Notification service enabled
5:22PM INF Starting reconciliation scheduler schedule="0 2 * * *" timezone=Asia/Jakarta
5:22PM INF Reconciliation Job Service started successfully
5:22PM INF Waiting for scheduled execution or shutdown signal...
```

## Understanding the Schedule

The `CRON_SCHEDULE` uses standard cron format:

```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of week (0 - 6, Sunday = 0)
│ │ │ │ │
│ │ │ │ │
* * * * *
```

Examples:
- `0 2 * * *` - Daily at 02:00
- `0 */6 * * *` - Every 6 hours (00:00, 06:00, 12:00, 18:00)
- `30 14 * * *` - Daily at 14:30

**Note:** Current POC implementation supports "0 2 * * *" format only. Full cron parser needed for production.

## How It Works

### Reconciliation Flow

```
1. Scheduled Trigger (02:00 WIB)
   ↓
2. Acquire Redis Lock (prevent duplicate runs)
   ↓
3. Fetch Internal Transactions (previous day 00:00-23:59)
   ↓
4. Fetch Provider Data (Xendit, Custody APIs)
   ↓
5. Compare in Batches (concurrent workers)
   ↓
6. Detect Discrepancies
   - Missing in provider
   - Amount mismatch
   - Status mismatch
   ↓
7. Save to Database (discrepancies table)
   ↓
8. Send Notifications
   - Summary to all recipients
   - Critical alerts for urgent issues
   ↓
9. Release Lock
```

### Discrepancy Categories

| Category | Severity | Example |
|----------|----------|---------|
| `missing_provider` | Critical | Transaction in DB but not in Xendit API |
| `amount_mismatch` | Major | Internal: Rp 500,000 vs Xendit: Rp 499,000 |
| `status_mismatch` | Minor | Internal: "completed" vs Custody: "pending" |

## Monitoring

### Check Job Status

```bash
# View logs
docker logs -f recon-job

# Check last run in Redis
docker exec -it recon-redis redis-cli
> GET recon:job:lock
> KEYS recon:stats:*
> GET recon:stats:2024-03-12
```

### Database Queries

```sql
-- View reconciliation runs
SELECT * FROM reconciliation_runs ORDER BY created_at DESC LIMIT 10;

-- View today's discrepancies
SELECT * FROM discrepancies 
WHERE created_at >= CURDATE() 
ORDER BY severity DESC, created_at DESC;

-- Count discrepancies by category
SELECT category, severity, COUNT(*) as count 
FROM discrepancies 
WHERE created_at >= CURDATE() 
GROUP BY category, severity;
```

## Troubleshooting

### Job Not Running

**Symptom:** No logs at scheduled time

**Solutions:**
1. Check timezone configuration:
   ```bash
   docker exec recon-job date
   # Should show time in configured timezone
   ```

2. Verify schedule format:
   ```bash
   docker logs recon-job | grep "cron_schedule"
   ```

3. Check Redis lock:
   ```bash
   docker exec -it recon-redis redis-cli
   > GET recon:job:lock
   > TTL recon:job:lock
   # If lock exists and TTL > 0, another instance may be running
   > DEL recon:job:lock  # Force release (use carefully!)
   ```

### High Discrepancy Count

**Symptom:** Many discrepancies detected

**Investigation:**
1. Check provider API connectivity
2. Review mock data vs production data
3. Verify date range calculation
4. Check amount threshold configuration

### Performance Issues

**Symptom:** Reconciliation takes too long

**Solutions:**
1. Increase `CONCURRENCY` (more workers)
2. Optimize `BATCH_SIZE` for your data volume
3. Enable Redis caching for provider data
4. Add database indexes:
   ```sql
   CREATE INDEX idx_tx_created ON transactions(created_at);
   CREATE INDEX idx_tx_provider_ref ON transactions(provider_ref);
   ```

## Mock Data (POC)

Current implementation uses mock data for:
- Provider API responses (Xendit, Custody)
- Database queries (sample transactions)
- Notifications (logs instead of sending)

### Production Readiness

To make production-ready, implement these PSEUDOCODE blocks:

**1. internal/services/database.go**
- Replace mock transactions with real SQL queries
- Add proper error handling and retries

**2. internal/reconciliation/engine.go**
- `fetchProviderData()`: Call actual provider APIs
- Add provider-specific client libraries
- Handle rate limiting and pagination

**3. internal/services/notification.go**
- Implement Slack webhook POST
- Add SMTP email sending
- Format messages with proper templates

**4. internal/scheduler/cron.go**
- Add full cron expression parser
- Implement skip-on-holiday logic
- Add retry mechanism for failed runs

## Next Steps

1. **Configure Notifications**
   - Set up Slack webhook URL
   - Add email recipients
   - Test notification delivery

2. **Customize Schedule**
   - Adjust `CRON_SCHEDULE` for your needs
   - Configure `TIMEZONE` to match your region

3. **Monitor & Tune**
   - Watch first few runs
   - Adjust `BATCH_SIZE` and `CONCURRENCY`
   - Fine-tune `DISCREPANCY_THRESHOLD`

4. **Production Deployment**
   - Replace mock implementations
   - Set up proper secrets management
   - Configure monitoring/alerting
   - Add health check endpoint

## Support

For issues or questions:
- Check logs: `docker logs recon-job`
- Review configuration: `docker exec recon-job env | grep -E 'CRON|TIMEZONE|BATCH'`
- Inspect database: Connect to MySQL and query `reconciliation_runs` table
