# Quick Start Guide

## Prerequisites

- Docker and Docker Compose installed
- (Optional) Go 1.21+ for local development
- (Optional) k6 for load testing

## Running the System

### 1. Start All Services

```bash
docker-compose up -d
```

This will start:
- PostgreSQL (port 5432)
- Redis (port 6379)
- Kafka + Zookeeper (port 9092)
- Event Ingest Service (port 8080)
- Stream Processor Service
- Recommendation API (port 8081)
- Prometheus (port 9090)
- Grafana (port 3000)

### 2. Check Service Health

```bash
# Check if all services are running
docker-compose ps

# Check logs
docker-compose logs -f ingest
docker-compose logs -f processor
docker-compose logs -f api
```

### 3. Test the APIs

#### Ingest an Event

```bash
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "item_id": 5,
    "event_type": "VIEW",
    "session_id": "test_session_1"
  }'
```

#### Get Recommendations

```bash
curl "http://localhost:8081/recommendations?user_id=1&count=10"
```

#### Get Popular Items

```bash
curl "http://localhost:8081/popular?count=20"

# With category filter
curl "http://localhost:8081/popular?category=electronics&count=20"
```

### 4. Generate Sample Data

Using PowerShell (Windows):

```powershell
.\scripts\generate_events.ps1 -BaseUrl "http://localhost:8080" -NumEvents 1000
```

Using Bash (Linux/Mac):

```bash
chmod +x scripts/generate_events.sh
./scripts/generate_events.sh http://localhost:8080 1000
```

### 5. Monitor the System

#### Prometheus Metrics

Open http://localhost:9090

Example queries:
- `rate(events_ingested_total[1m])` - Event ingestion rate
- `histogram_quantile(0.95, rate(recommendation_latency_seconds_bucket[5m]))` - P95 latency
- `recommendation_cache_hits_total / (recommendation_cache_hits_total + recommendation_cache_misses_total)` - Cache hit ratio

#### Grafana Dashboards

1. Open http://localhost:3000
2. Login: admin / admin
3. Add Prometheus as data source (http://prometheus:9090)
4. Import or create dashboards

### 6. Load Testing

Install k6:

```bash
# Windows (using chocolatey)
choco install k6

# Mac
brew install k6

# Linux
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

Run load test:

```bash
k6 run scripts/load_test_k6.js
```

## Development

### Build Services Locally

```bash
# Install dependencies
go mod download

# Build all services
make build

# Or build individually
go build -o bin/ingest.exe ./cmd/ingest
go build -o bin/processor.exe ./cmd/processor
go build -o bin/api.exe ./cmd/api
```

### Run Services Locally

Make sure you have PostgreSQL, Redis, and Kafka running (via Docker Compose), then:

```bash
# Terminal 1: Run ingest service
go run ./cmd/ingest

# Terminal 2: Run processor service
go run ./cmd/processor

# Terminal 3: Run API service
go run ./cmd/api
```

### Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Troubleshooting

### Services won't start

```bash
# Clean up and restart
docker-compose down -v
docker-compose up -d
```

### Check Kafka topics

```bash
docker exec -it reco-kafka kafka-topics.sh --list --bootstrap-server localhost:9092
```

### Check Redis data

```bash
docker exec -it reco-redis redis-cli

# In Redis CLI:
> KEYS *
> LRANGE user:recent:1 0 -1
> ZREVRANGE item:popularity 0 10 WITHSCORES
```

### Check PostgreSQL data

```bash
docker exec -it reco-postgres psql -U reco -d reco

# In psql:
\dt
SELECT COUNT(*) FROM events;
SELECT * FROM items LIMIT 10;
```

## Production Considerations

1. **Scaling**: Use Kubernetes or Docker Swarm for orchestration
2. **High Availability**: Run multiple replicas of each service
3. **Security**: 
   - Add authentication to API endpoints
   - Use TLS for all connections
   - Implement rate limiting
4. **Monitoring**: Set up alerts for key metrics
5. **Backup**: Regular backups of PostgreSQL and Redis
6. **Configuration**: Use environment-specific configs
7. **Logging**: Centralized logging with ELK stack or similar

## Architecture Diagram

```
┌─────────┐
│ Client  │
└────┬────┘
     │
     ▼
┌─────────────────┐      ┌────────┐
│  Ingest API     │─────▶│ Kafka  │
│  (Port 8080)    │      └───┬────┘
└─────────────────┘          │
                             │
     ┌───────────────────────┼───────────────────┐
     │                       │                   │
     ▼                       ▼                   ▼
┌─────────┐          ┌────────────┐      ┌──────────┐
│Postgres │          │ Processor  │      │  Redis   │
│  (DB)   │          │  Service   │─────▶│ (Cache)  │
└─────────┘          └────────────┘      └────┬─────┘
     │                                         │
     │                 ┌───────────────────────┘
     │                 │
     ▼                 ▼
┌─────────────────────────┐
│   Recommendation API    │
│     (Port 8081)         │
└─────────────────────────┘
```

## Next Steps

1. Implement offline training pipeline (Python/Go)
2. Add ANN index (Milvus/Faiss) for vector similarity
3. Implement A/B testing framework
4. Add more sophisticated ranking algorithms
5. Implement real-time feature updates
6. Add user authentication and authorization
7. Implement data privacy controls (GDPR compliance)
