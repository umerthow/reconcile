# recon-worker Quick Start Guide

## 1. Setup Environment

```bash
cd recon-worker

# Copy environment template
cp ../.env.example .env

# Edit .env with your settings (or use defaults for local dev)
```

## 2. Install Dependencies

```bash
# Download Go modules
go mod download

# Install development tools (optional)
make install-tools
```

## 3. Start Dependencies

Make sure MySQL, Redis, and Kafka are running:

```bash
# From root directory
cd ..
docker-compose up -d recon-mysql recon-redis

# Verify Kafka is running (should already be running from recon-api setup)
# If not, start your local Kafka instance
```

## 4. Run the Worker

### Option A: Direct Run
```bash
make run
```

### Option B: Development Mode (with hot reload)
```bash
make dev
```

### Option C: Docker
```bash
# Build image
make docker-build

# Run with docker-compose
cd ..
docker-compose up recon-worker
```

## 5. Verify It's Running

You should see logs like:

```
INF Starting recon-worker environment=development
INF Database connection established
INF Redis connection established
INF Kafka connection established brokers=[localhost:29092]
INF All services initialized successfully
INF Starting Xendit consumer group_id=recon-worker-group-xendit topics=[xendit-webhooks]
INF Starting Custody consumer group_id=recon-worker-group-custody topics=[custody-webhooks]
INF Starting Custody status updates consumer group_id=recon-worker-group-custody-status
INF Starting DLQ consumer group_id=recon-worker-group-dlq topics=[dlq-webhooks]
INF Starting Custody API polling loop interval=5m0s
INF All workers started successfully
```

## 6. Test the Worker

### Send Test Webhook via recon-api

From the recon-api directory:

```bash
# Send Xendit webhook
./test-xendit-webhook.sh

# Send Custody webhook
./test-custody-webhook.sh
```

Watch the recon-worker logs - you should see messages being processed.

## 7. Mock Data Behavior

The POC includes mock implementations:

- **Database updates**: Logged but not executed (see `[MOCK]` in logs)
- **Custody API polling**: Returns hardcoded transaction statuses
- **Pending withdrawals**: Returns 3 mock tx hashes

To add real implementations, search for `PSEUDOCODE` comments in the code.

## 8. Graceful Shutdown

```bash
# Press Ctrl+C or send SIGTERM
# Worker will:
# 1. Stop accepting new messages
# 2. Wait for in-flight processing (max 30s)
# 3. Close all connections gracefully
```

## Common Issues

### "Failed to connect to Kafka"
- Ensure Kafka is running on `localhost:29092`
- Check `KAFKA_BROKERS` in `.env`
- For Docker, use `host.docker.internal:29092`

### "Failed to connect to database"
- Ensure MySQL is running: `docker-compose ps`
- Check credentials in `.env`
- Test connection: `mysql -h localhost -u recon_user -p`

### "No pending withdrawals to poll"
- This is normal for POC - mock data returns empty list by default
- Real implementation would query `transactions` table

## Development Tips

1. **View Kafka Messages**
   ```bash
   # Console consumer
   kafka-console-consumer --bootstrap-server localhost:29092 \
     --topic xendit-webhooks --from-beginning
   ```

2. **Check Consumer Lag**
   ```bash
   kafka-consumer-groups --bootstrap-server localhost:29092 \
     --describe --group recon-worker-group-xendit
   ```

3. **Watch Logs**
   ```bash
   # Docker
   docker logs -f recon-worker
   
   # Local
   make run | jq .  # Pretty print JSON logs
   ```

4. **Run Tests**
   ```bash
   make test
   make test-coverage  # View coverage in browser
   ```

## Next Steps

- Review `internal/services/database.go` for SQL implementation examples
- Check `internal/services/custody_api.go` for API integration
- Add real webhook signature verification
- Implement comprehensive error handling
- Add monitoring and alerting

Ready to consume! 🚀
