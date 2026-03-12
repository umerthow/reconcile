# Recon API Service

Payment Reconciliation System - API Service for webhook ingestion and dashboard queries.

## Features

- **Webhook Ingestion**: Accepts webhooks from Xendit and Custody providers
- **Signature Verification**: HMAC-SHA256 validation for webhook authenticity
- **Idempotency**: Redis-based deduplication (48-hour window)
- **Event Publishing**: Kafka integration for async processing
- **Reconciliation Dashboard**: REST APIs for viewing reconciliation runs
- **Discrepancy Management**: Track and resolve payment discrepancies
- **JWT Authentication**: Secure API endpoints

## Tech Stack

- **Go 1.21+**: High-performance, statically typed
- **Gin**: Web framework for REST APIs
- **MySQL 5.7**: Relational database for transactions
- **Redis 7**: Cache for webhook deduplication
- **Kafka**: Message queue for event streaming
- **Docker**: Containerization

## Project Structure

```
recon-api/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── config/              # Configuration management
│   ├── handlers/            # HTTP request handlers
│   ├── middleware/          # Auth, validation, recovery
│   ├── models/              # Data structures
│   ├── services/            # Database, Redis, Kafka
│   └── utils/               # Logger, signature verification
├── .dockerignore
├── .gitignore
├── Dockerfile               # Multi-stage build
├── Makefile                 # Build automation
├── go.mod
└── README.md
```

## Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- MySQL 5.7
- Redis 7
- Kafka (running on localhost:29092)

### Environment Variables

Create a `.env` file:

```env
# Server
SERVER_PORT=8080
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s

# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=recon_user
DB_PASSWORD=recon_pass
DB_NAME=recon_db
DB_MAX_CONNS=20
DB_MAX_IDLE_CONNS=5

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Kafka
KAFKA_BROKERS=localhost:29092
KAFKA_XENDIT_TOPIC=xendit-webhooks
KAFKA_CUSTODY_TOPIC=custody-webhooks
KAFKA_DLQ_TOPIC=webhooks-dlq
KAFKA_SASL_ENABLED=false
KAFKA_SASL_USERNAME=
KAFKA_SASL_PASSWORD=

# JWT
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRY=24h

# Logger
LOG_LEVEL=info
ENVIRONMENT=development
```

### Running Locally

```bash
# Install dependencies
make deps

# Run the service
make run

# Or with hot reload (requires air)
make dev
```

### Running with Docker

```bash
# Build image
make docker-build

# Run container
make docker-run
```

## API Endpoints

### Health Check

```http
GET /health
```

### Webhooks (No Auth Required)

```http
POST /webhooks/xendit
Content-Type: application/json
X-Xendit-Signature: <signature>

{
  "provider": "xendit",
  "event_type": "payment.succeeded",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "id": "txn_123",
    "amount": 100000,
    "status": "SUCCEEDED"
  }
}
```

```http
POST /webhooks/custody
Content-Type: application/json
X-Custody-Signature: <signature>

{
  "provider": "custody",
  "event_type": "withdrawal.completed",
  "timestamp": "2024-03-01T09:00:00Z",
  "data": {
    "txHash": "0xabc123...",
    "asset": "USDT",
    "amount": "250.00",
    "status": "COMPLETED",
    "timestamp": "2024-03-01T09:00:00Z"
  }
}
```

### Reconciliation (JWT Auth Required)

```http
GET /api/v1/reconciliations?page=1&limit=20&status=success&provider=xendit
Authorization: Bearer <jwt-token>
```

```http
POST /api/v1/reconciliations
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "window_start": "2024-01-15T00:00:00Z",
  "window_end": "2024-01-15T23:59:59Z",
  "dry_run": false
}
```

### Discrepancies (JWT Auth Required)

```http
GET /api/v1/reconciliations/:run_id/discrepancies
Authorization: Bearer <jwt-token>
```

```http
POST /api/v1/discrepancies/:discrepancy_id/resolve
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "resolved_by": "admin",
  "resolution_notes": "Manual adjustment applied",
  "action": "resolve"
}
```

## Mock Data (POC)

For POC purposes, the service uses hardcoded mock data:

### Mock Reconciliation Runs
- 2 sample reconciliation runs
- 50K+ transactions processed
- Various statuses (success, running)
- Realistic timestamps and discrepancy counts

### Mock Discrepancies
- 3 sample discrepancies
- Categories: `amount_mismatch`, `missing_credit`, `stale_withdrawal`
- Severity levels: `high`, `low`
- Realistic provider references

### Mock JWT Token

In development mode, a mock JWT token is logged on startup for testing protected endpoints.

## Development

```bash
# Run tests
make test

# Run with coverage
make test-coverage

# Format code
make fmt

# Lint code
make lint

# Build binary
make build
```

## Docker Commands

```bash
# Build image
docker build -t recon-api:latest .

# Run container
docker run -p 8080:8080 --env-file .env recon-api:latest

# View logs
docker logs -f <container-id>
```

## Production Considerations

**Current State**: POC with mock data

**TODO for Production**:
1. Replace mock database queries with real MySQL queries
2. Implement actual webhook signature verification (provider-specific)
3. Add database migrations (e.g., golang-migrate)
4. Implement request rate limiting
5. Add metrics and monitoring (Prometheus)
6. Add distributed tracing (Jaeger/OpenTelemetry)
7. Implement proper error handling and retry logic
8. Add comprehensive test coverage
9. Set up CI/CD pipeline
10. Configure production-grade logging (JSON format)

## Architecture Notes

- **Clean Architecture**: Separation of concerns (handlers, services, models)
- **Mock Data**: Service layer returns hardcoded data for POC
- **Extensible**: TODO comments mark where real implementations go
- **Financial Precision**: Uses `shopspring/decimal` for money calculations
- **Idempotency**: 48-hour deduplication window via Redis
- **Async Processing**: Webhooks published to Kafka for worker consumption

## License

MIT
