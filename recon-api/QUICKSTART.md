# Recon API - Quick Start Guide

## 🚀 Getting Started (POC Mode)

This service is ready for POC with mock data. No real database queries needed!

### Option 1: Run Locally (Fastest)

```bash
cd recon-api

# Install dependencies
go mod download

# Copy environment variables
cp .env.example .env

# Run the service
go run cmd/main.go
```

The service will start on `http://localhost:8080` with mock data.

### Option 2: Docker Compose (Full Stack)

```bash
cd recon-api

# Start all services (MySQL, Redis, API)
docker-compose up -d

# View logs
docker-compose logs -f recon-api

# Stop all services
docker-compose down
```

## 🧪 Testing the API

### 1. Health Check (No Auth)

```bash
curl http://localhost:8080/health
```

### 2. Get Mock JWT Token

When running in development mode, check the logs for the mock JWT token:

```bash
# Look for: "Mock JWT token generated for testing"
# Copy the token value
```

Or generate one programmatically (see `internal/middleware/auth.go::GenerateMockToken`)

### 3. Test Webhook (No Auth)

```bash
curl -X POST http://localhost:8080/webhooks/xendit \
  -H "Content-Type: application/json" \
  -H "X-Xendit-Signature: mock-signature" \
  -d '{
    "provider": "xendit",
    "event_type": "payment.succeeded",
    "timestamp": "2024-01-15T10:30:00Z",
    "payload": {
      "id": "txn_123",
      "amount": 100000,
      "status": "SUCCEEDED"
    }
  }'
```

### 4. List Reconciliation Runs (Auth Required)

```bash
export TOKEN="<your-mock-jwt-token>"

curl http://localhost:8080/api/v1/reconciliations?page=1&limit=10 \
  -H "Authorization: Bearer $TOKEN"
```

### 5. Trigger Reconciliation (Auth Required)

```bash
curl -X POST http://localhost:8080/api/v1/reconciliations \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "window_start": "2024-01-15T00:00:00Z",
    "window_end": "2024-01-15T23:59:59Z",
    "dry_run": false
  }'
```

### 6. Get Discrepancies (Auth Required)

```bash
# Use run_id from mock data: "run_20240115_001" or "run_20240116_001"
curl http://localhost:8080/api/v1/reconciliations/run_20240115_001/discrepancies \
  -H "Authorization: Bearer $TOKEN"
```

### 7. Resolve Discrepancy (Auth Required)

```bash
# Use discrepancy_id from mock data: "disc_001", "disc_002", or "disc_003"
curl -X POST http://localhost:8080/api/v1/discrepancies/disc_001/resolve \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "resolved_by": "admin",
    "resolution_notes": "Manually verified and adjusted",
    "action": "resolve"
  }'
```

## 📊 Mock Data Available

### Reconciliation Runs

- **run_20240115_001**: Success, 52,347 transactions, 12 discrepancies
- **run_20240116_001**: Running, 48,923 transactions, 8 discrepancies

### Discrepancies

- **disc_001**: Amount mismatch (HIGH severity) - Xendit
- **disc_002**: Missing credit (HIGH severity) - Custody
- **disc_003**: Stale withdrawal (LOW severity) - Custody

## 🔧 Development Commands

```bash
# Format code
make fmt

# Run tests (when added)
make test

# Build binary
make build

# Run with hot reload (requires air)
make dev

# Docker build
make docker-build

# Docker run
make docker-run
```

## 📝 Key Files

- `cmd/main.go` - Application entry point
- `internal/handlers/` - HTTP request handlers
- `internal/services/database.go` - **Mock data location**
- `internal/middleware/auth.go` - JWT authentication & mock token generation
- `.env.example` - Environment configuration template

## 🚦 Next Steps for Production

1. **Database**: Replace mock queries in `internal/services/database.go` with real MySQL queries
2. **Migrations**: Add database schema migrations
3. **Signature Verification**: Implement real webhook signature verification
4. **Rate Limiting**: Add rate limiting middleware
5. **Monitoring**: Add Prometheus metrics
6. **Testing**: Add unit and integration tests
7. **CI/CD**: Set up GitHub Actions or similar

## 🐛 Troubleshooting

### Service won't start

- Check if ports 8080, 3306, 6379 are available
- Verify `.env` file exists and has correct values
- Check logs: `docker-compose logs recon-api`

### Can't connect to database

- This is expected in POC mode! Mock data doesn't need real DB
- If using Docker Compose, ensure MySQL is healthy: `docker-compose ps`

### JWT token invalid

- Get a fresh mock token from the startup logs
- Ensure token is passed in `Authorization: Bearer <token>` header
- Token expires after 24 hours (configurable in JWT_EXPIRY)

## 📚 Additional Resources

- [README.md](./README.md) - Full documentation
- [Architecture Planning](../planning.md) - System architecture overview
- [Case Study](../case_study.md) - Business requirements

## ❓ Questions?

This is a POC build with clean architecture. All mock data is in `internal/services/database.go`. 
Look for `TODO:` comments throughout the codebase for real implementation points.
