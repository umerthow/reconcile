# recon-worker

Asynchronous worker service for the Payment Reconciliation System. Consumes webhook messages from Kafka, polls custody APIs for pending withdrawals, and updates the internal transaction ledger.

## Architecture

The recon-worker consists of:

1. **Kafka Consumers** (4 consumer groups)
   - `xendit-webhooks` consumer - Processes Xendit fiat transaction webhooks
   - `custody-webhooks` consumer - Processes Custody crypto deposit webhooks
   - `custody-status-updates` consumer - Processes polling results
   - `dlq-webhooks` consumer - Handles failed messages for manual replay

2. **Custody API Poller**
   - Runs every 5 minutes (configurable)
   - Fetches status of pending crypto withdrawals
   - Publishes updates to `custody-status-updates` topic

3. **Service Layer**
   - Database service (MySQL connection pool)
   - Redis service (deduplication cache)
   - Kafka service (consumer groups)
   - Custody API service (with circuit breaker)

## Quick Start

### Prerequisites

- Go 1.23+
- Docker & Docker Compose (for dependencies)
- Running MySQL, Redis, and Kafka (from recon-api setup)

### Development

```bash
# Install dependencies
make deps

# Run locally (requires .env file)
make run

# Run with hot reload
make dev

# Run tests
make test

# Build binary
make build
```

### Docker

```bash
# Build image
make docker-build

# Run in Docker Compose (from root directory)
docker-compose up recon-worker
```

## Configuration

Environment variables (see `.env.example`):

```bash
# Database
DB_HOST=localhost
DB_PORT=3306
DB_NAME=reconciliation
DB_USER=recon_user
DB_PASSWORD=recon_pass

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Kafka
KAFKA_BROKERS=localhost:29092
KAFKA_GROUP_ID=recon-worker-group
KAFKA_SSL_ENABLE=false

# Custody API
CUSTODY_API_URL=https://api.custody-provider.com
CUSTODY_API_KEY=your-api-key

# Polling
POLLING_INTERVAL_SEC=300  # 5 minutes

# Logging
GO_ENV=development
LOG_LEVEL=info
```

## Mock Implementation

**This is a POC build with mock data:**

- Database queries return mock success responses
- Custody API returns hardcoded transaction statuses
- All critical business logic marked with `[MOCK]` and `PSEUDOCODE` comments
- Ready for enhancement with real implementations

## Consumer Flow

### Xendit Webhook Flow
```
Kafka Topic (xendit-webhooks)
    ↓
XenditConsumer
    ↓
1. Parse webhook message
2. Check Redis deduplication
3. Update transactions table (status, incoming_timestamp)
4. Insert provider_callbacks (audit trail)
5. Set Redis cache (48h TTL)
    ↓
Success / DLQ on failure
```

### Custody Webhook Flow
```
Kafka Topic (custody-webhooks)
    ↓
CustodyConsumer
    ↓
1. Parse webhook message
2. Check Redis deduplication (by txHash)
3. Update transactions table (status, incoming_timestamp)
4. Insert provider_callbacks (audit trail)
5. Set Redis cache (48h TTL)
    ↓
Success / DLQ on failure
```

### Custody Polling Flow
```
Every 5 minutes
    ↓
CustodyPoller
    ↓
1. Query pending withdrawals from DB
2. Fetch statuses from Custody API (with circuit breaker)
3. Publish updates to custody-status-updates topic
    ↓
CustodyConsumer processes updates
```

## Idempotency Strategy

1. **Redis Deduplication Cache**
   - Key: `webhook:{provider}:{webhook_id}`
   - TTL: 48 hours
   - Prevents duplicate webhook processing

2. **Database Timestamp Check**
   - Only update if `incoming_timestamp > updated_at`
   - Prevents stale webhooks from overwriting newer data

3. **Kafka Offset Management**
   - Manual commit after successful processing
   - On failure, message sent to DLQ instead of retry

## Error Handling

1. **Consumer Level**
   - Parse errors → Send to DLQ
   - Processing errors → Send to DLQ
   - Mark message as processed to avoid blocking

2. **Custody API Level**
   - Circuit breaker (3 failures → open for 30s)
   - Exponential backoff (1s, 2s, 4s, 8s, 16s)
   - Graceful degradation (skip poll cycle on failure)

3. **Database Level**
   - Connection pool with retry
   - Timeout on queries (5s)
   - Transaction rollback on error

## Graceful Shutdown

Worker listens for `SIGINT` and `SIGTERM`:

1. Stop accepting new messages
2. Stop poller loop
3. Wait for in-flight messages to complete (max 30s)
4. Commit Kafka offsets
5. Close all connections

## Monitoring

Key metrics to monitor:

- Kafka consumer lag (per topic)
- Message processing time (p50, p95, p99)
- DLQ message count
- Circuit breaker state (open/closed)
- Custody API response time
- Redis cache hit rate

## Next Steps (Production)

1. Replace mock database queries with real SQL
2. Implement actual Custody API integration
3. Add Prometheus metrics
4. Add distributed tracing (OpenTelemetry)
5. Implement retry logic with backoff
6. Add alerting for high DLQ count
7. Scale horizontally (multiple worker replicas)

## Project Structure

```
recon-worker/
├── cmd/
│   └── main.go              # Entry point
├── internal/
│   ├── config/              # Configuration
│   ├── consumers/           # Kafka consumers
│   │   ├── xendit.go
│   │   ├── custody.go
│   │   └── dlq.go
│   ├── polling/             # Custody API poller
│   │   └── custody_poller.go
│   ├── services/            # Service layer
│   │   ├── database.go
│   │   ├── redis.go
│   │   ├── kafka.go
│   │   └── custody_api.go
│   ├── models/              # Data models
│   │   └── transaction.go
│   └── utils/               # Utilities
│       └── logger.go
├── Dockerfile
├── Makefile
└── README.md
```

## License

Proprietary - Reku Payment Reconciliation System
