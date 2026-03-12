# Reconciliation Job Service

Scheduled batch job service for daily payment reconciliation. Runs at 02:00 WIB (configurable) to compare internal transaction records against payment provider data and detect discrepancies.

## Features

- **Scheduled Execution**: Cron-like scheduler for daily reconciliation runs
- **Distributed Locking**: Redis-based locking to prevent concurrent execution
- **Concurrent Processing**: Configurable batch processing with worker pools
- **Discrepancy Detection**: Identifies missing transactions, amount mismatches, and status discrepancies
- **Notifications**: Slack and email alerts for reconciliation summaries and critical issues
- **Mock Data**: POC-ready with mock provider data for testing

## Architecture

```
┌──────────────────────────────────────────────────┐
│           Reconciliation Job Service             │
├──────────────────────────────────────────────────┤
│                                                  │
│  ┌────────────┐    ┌──────────────────┐         │
│  │  Scheduler │───▶│ Reconciliation   │         │
│  │  (Cron)    │    │     Engine       │         │
│  └────────────┘    └──────────────────┘         │
│                            │                     │
│        ┌───────────────────┼─────────────┐       │
│        ▼                   ▼             ▼       │
│  ┌──────────┐      ┌────────────┐  ┌─────────┐  │
│  │ Database │      │   Redis    │  │ Notify  │  │
│  │ Service  │      │  Service   │  │ Service │  │
│  └──────────┘      └────────────┘  └─────────┘  │
│        │                   │             │       │
└────────┼───────────────────┼─────────────┼───────┘
         ▼                   ▼             ▼
    ┌─────────┐        ┌─────────┐   ┌──────────┐
    │  MySQL  │        │  Redis  │   │Slack/Mail│
    └─────────┘        └─────────┘   └──────────┘
```

## Configuration

Environment variables (see `.env.example`):

- **Database**: MySQL connection settings
- **Redis**: Redis connection for distributed locking and caching
- **Schedule**: `CRON_SCHEDULE=0 2 * * *` (02:00 daily), `TIMEZONE=Asia/Jakarta`
- **Processing**: `BATCH_SIZE=1000`, `CONCURRENCY=5`
- **Thresholds**: `DISCREPANCY_THRESHOLD=0.01` (minimum amount difference to flag)
- **Notifications**: Slack webhook, email recipients

## Reconciliation Process

### 1. Scheduled Trigger
- Runs daily at configured time (default: 02:00 WIB)
- Acquires distributed lock via Redis
- Prevents concurrent execution across multiple instances

### 2. Data Collection
- Fetches internal transactions from previous day (00:00-23:59)
- Retrieves provider data from payment gateway APIs (Xendit, Custody)
- Caches provider data in Redis for performance

### 3. Reconciliation
- Compares transactions in concurrent batches
- Detects discrepancies:
  - **Missing in Provider**: Transaction exists internally but not in provider data
  - **Amount Mismatch**: Transaction amounts differ beyond threshold
  - **Status Mismatch**: Transaction statuses don't match

### 4. Reporting
- Saves discrepancies to database
- Categorizes by severity: critical, major, minor
- Sends summary notification via Slack/email
- Sends immediate alert for critical issues

## Discrepancy Categories

| Category | Severity | Description |
|----------|----------|-------------|
| `missing_provider` | Critical | Transaction not found in provider data |
| `amount_mismatch` | Major/Minor | Amount difference exceeds threshold |
| `status_mismatch` | Minor | Transaction status differs |

## Building & Running

### Local Development

```bash
# Copy environment template
cp .env.example .env

# Edit configuration
vim .env

# Build
make build

# Run
make run
```

### Docker

```bash
# Build Docker image
make docker-build

# Run container
docker run -d \
  --name recon-job \
  --env-file .env \
  --network recon-network \
  recon-job:latest
```

### Docker Compose

Add to root `docker-compose.yml`:

```yaml
recon-job:
  build:
    context: ./recon-job
    dockerfile: Dockerfile
  container_name: recon-job
  environment:
    DB_HOST: mysql
    DB_PORT: 3306
    DB_USER: reconcile_user
    DB_PASSWORD: reconcile_password
    DB_NAME: reconcile_db
    REDIS_ADDR: redis:6379
    CRON_SCHEDULE: "0 2 * * *"
    TIMEZONE: "Asia/Jakarta"
    BATCH_SIZE: 1000
    CONCURRENCY: 5
    DISCREPANCY_THRESHOLD: 0.01
    NOTIFICATION_ENABLED: "true"
    LOG_LEVEL: info
  depends_on:
    mysql:
      condition: service_healthy
    redis:
      condition: service_healthy
  networks:
    - recon-network
  restart: unless-stopped
```

## Development Notes

### Mock Implementation

Current POC implementation uses mock data:

- **Database queries**: Log operations, return sample data
- **Provider API calls**: Return mock provider transactions
- **Notifications**: Log messages instead of sending

### Production Enhancement

To make production-ready, replace PSEUDOCODE blocks in:

1. **internal/services/database.go**: Implement real SQL queries
2. **internal/reconciliation/engine.go**: 
   - `fetchProviderData()`: Call actual provider APIs (Xendit, Custody)
   - Add provider-specific reconciliation logic
3. **internal/services/notification.go**: 
   - Implement Slack webhook POST requests
   - Add email sending via SMTP
4. **internal/scheduler/cron.go**: 
   - Add full cron expression parser (currently only supports "0 2 * * *")
   - Add advanced scheduling features (skip holidays, retry logic)

### Testing

```bash
# Run tests
make test

# Trigger immediate reconciliation (for testing)
# Add HTTP endpoint or CLI command to call scheduler.RunNow()
```

## Project Structure

```
recon-job/
├── cmd/
│   └── main.go                    # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go              # Configuration management
│   ├── models/
│   │   ├── reconciliation.go      # Data structures
│   │   └── provider.go            # Provider transaction model
│   ├── services/
│   │   ├── database.go            # Database operations
│   │   ├── redis.go               # Redis operations
│   │   └── notification.go        # Slack/email notifications
│   ├── reconciliation/
│   │   └── engine.go              # Core reconciliation logic
│   ├── scheduler/
│   │   └── cron.go                # Job scheduling
│   └── utils/
│       └── logger.go              # Logging utilities
├── Dockerfile                      # Container image definition
├── Makefile                        # Build automation
├── go.mod                          # Go module definition
├── .env.example                    # Environment template
└── README.md                       # This file
```

## Monitoring

### Logs

```bash
# View logs
docker logs -f recon-job

# Log levels: debug, info, warn, error
```

### Metrics

Monitor in Redis:
- `recon:job:lock` - Current job lock status
- `recon:stats:{date}` - Historical run statistics

Monitor in Database:
- `reconciliation_runs` - All reconciliation executions
- `discrepancies` - All detected issues

## Troubleshooting

### Job Not Running

1. Check scheduler logs for next scheduled time
2. Verify timezone configuration
3. Check Redis connectivity
4. Verify another instance isn't holding lock

### High Discrepancy Count

1. Check provider API connectivity
2. Verify provider data date range
3. Review threshold configuration
4. Check for system clock drift

### Performance Issues

1. Increase `CONCURRENCY` setting
2. Optimize `BATCH_SIZE` for your data volume
3. Enable Redis caching for provider data
4. Add database indexes on transaction queries

## License

Internal use only
