# Payment Reconciliation System

Multi-service payment reconciliation system for processing ~50K transactions/day (fiat + crypto).

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   recon-api     │────▶│  Kafka/RedPanda  │────▶│  recon-worker   │
│  (Web Server)   │     │  (Event Queue)   │     │ (Async Consumer)│
└─────────────────┘     └──────────────────┘     └─────────────────┘
        │                        │                         │
        ▼                        ▼                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                         MySQL 5.7 Database                       │
└─────────────────────────────────────────────────────────────────┘
```

## Services

### 1. recon-api
- HTTP-facing webhook receiver
- REST APIs for reconciliation dashboard
- Kafka producer for async processing
- Port: 8080

### 2. recon-worker
- Kafka consumer (4 consumer groups)
- Custody API poller (5-minute intervals)
- Transaction ledger updater
- DLQ handler for failed messages

### 3. recon-job (Coming Soon)
- Daily batch reconciliation at 02:00 WIB
- Matching engine
- Discrepancy detection and alerting

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.23+ (for local development)
- Kafka running on `localhost:29092` (or update KAFKA_BROKERS in .env)

### Start All Services

```bash
# Clone and navigate to project
cd reconcile

# Copy environment template
cp .env.example .env

# Start services with Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f

# Check health
curl http://localhost:8080/health
```

### Services Started

- MySQL 5.7 → `localhost:3306`
- Redis 7 → `localhost:6379`
- recon-api → `localhost:8080`
- recon-worker → (background consumer, no exposed port)

### Stop Services

```bash
docker-compose down

# Remove volumes (clean slate)
docker-compose down -v
```

## Testing

### Test Webhooks

```bash
# Get JWT token from recon-api logs
docker logs recon-api | grep "Mock JWT"

# Test Xendit webhook
curl -X POST http://localhost:8080/webhooks/xendit \
  -H "Content-Type: application/json" \
  -H "x-callback-token: xendit-secret-key" \
  -d '{
    "event": "payment.paid",
    "data": {
      "id": "payment-123",
      "external_id": "order-001",
      "status": "PAID",
      "amount": 500000,
      "currency": "IDR",
      "updated": "2026-03-12T17:01:19Z"
    }
  }'

# Test Custody webhook
curl -X POST http://localhost:8080/webhooks/custody \
  -H "Content-Type: application/json" \
  -H "x-api-key: custody-secret-key" \
  -d '{
    "event": "deposit.confirmed",
    "data": {
      "txHash": "0xabc123...",
      "asset": "BTC",
      "amount": "0.05",
      "status": "completed",
      "timestamp": "2026-03-12T17:01:19Z"
    }
  }'

# Watch worker process the webhooks
docker logs -f recon-worker
```

### Test REST APIs

```bash
# Get reconciliation runs
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/v1/reconciliations

# Get discrepancies
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/v1/discrepancies
```

## Development

### Local Development (without Docker)

```bash
# Start dependencies only
docker-compose up -d mysql redis

# recon-api
cd recon-api
make dev  # Hot reload

# recon-worker (in another terminal)
cd recon-worker
make dev  # Hot reload
```

### Build from Source

```bash
# recon-api
cd recon-api
make build
./bin/recon-api

# recon-worker
cd recon-worker
make build
./bin/recon-worker
```

### Run Tests

```bash
# recon-api
cd recon-api
make test

# recon-worker
cd recon-worker
make test
```

## Configuration

See `.env.example` for all available configuration options.

Key environment variables:

- `DB_*`: Database connection settings
- `REDIS_*`: Redis connection settings
- `KAFKA_BROKERS`: Kafka broker addresses
- `CUSTODY_API_*`: External custody provider API
- `JWT_SECRET`: Secret for JWT token signing
- `LOG_LEVEL`: info, debug, warn, error

## Project Structure

```
reconcile/
├── docker-compose.yml       # Orchestrates all services
├── .env.example             # Environment template
├── planning.md              # Detailed architecture plan
├── recon-api/               # Web API service
│   ├── cmd/main.go
│   ├── internal/
│   ├── Dockerfile
│   └── README.md
├── recon-worker/            # Async worker service
│   ├── cmd/main.go
│   ├── internal/
│   ├── Dockerfile
│   └── README.md
└── recon-job/               # Batch job (coming soon)
```

## Mock Implementation (POC)

This is a proof-of-concept build with mock data for rapid development:

- ✅ Complete service architecture implemented
- ✅ All HTTP endpoints functional
- ✅ Kafka consumers running
- ✅ Circuit breaker patterns
- ✅ Graceful shutdown handling
- 🔄 Database queries are mocked (logged, not executed)
- 🔄 Custody API returns hardcoded responses
- 🔄 All critical flows marked with `[MOCK]` and `PSEUDOCODE` comments

**Ready for production enhancement**: Replace mock implementations with real SQL queries and API calls.

## Monitoring

### View Service Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker logs -f recon-api
docker logs -f recon-worker

# Follow with grep
docker logs -f recon-worker | grep MOCK
```

### Check Kafka Consumer Lag

```bash
kafka-consumer-groups --bootstrap-server localhost:29092 \
  --describe --group recon-worker-group-xendit
```

### Database Access

```bash
docker exec -it recon-mysql mysql -u recon_user -p recon_db
# Password: recon_pass
```

### Redis Access

```bash
docker exec -it recon-redis redis-cli
> KEYS webhook:*
> TTL webhook:xendit:payment-123
```

## Documentation

- `planning.md` - Complete architecture and implementation plan
- `recon-api/README.md` - API service documentation
- `recon-api/QUICKSTART.md` - API quick start guide
- `recon-worker/README.md` - Worker service documentation
- `recon-worker/QUICKSTART.md` - Worker quick start guide

## Next Steps

1. ✅ recon-api - Complete
2. ✅ recon-worker - Complete
3. 🔄 recon-job - Batch reconciliation (next phase)
4. 🔄 Replace mock implementations with real SQL
5. 🔄 Implement real webhook signature verification
6. 🔄 Add comprehensive tests
7. 🔄 Add Prometheus metrics
8. 🔄 Production deployment (K8s manifests)

## License

Proprietary - Reku Payment Reconciliation System

## Support

For questions or issues, contact the engineering team.
