# Quick Reference Commands

## First Time Setup

### Windows PowerShell
```powershell
# Setup everything automatically
.\setup.ps1

# Or manual steps:
docker-compose up -d
Start-Sleep -Seconds 10
.\scripts\generate_events.ps1 -NumEvents 1000
```

### Linux/Mac
```bash
# Setup everything automatically
./setup.sh

# Or manual steps:
docker-compose up -d
sleep 10
./scripts/generate_events.sh http://localhost:8080 1000
```

## Daily Usage

```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f
docker-compose logs -f ingest
docker-compose logs -f processor
docker-compose logs -f api

# Restart specific service
docker-compose restart ingest
docker-compose restart api
docker-compose restart processor

# Check status
docker-compose ps
```

## Testing APIs

### Event Ingestion
```bash
# Bash/Curl
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"item_id":5,"event_type":"VIEW","session_id":"test"}'

# PowerShell
$body = @{user_id=1;item_id=5;event_type="VIEW";session_id="test"} | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8080/events" -Method Post -Body $body -ContentType "application/json"
```

### Get Recommendations
```bash
curl "http://localhost:8081/recommendations?user_id=1&count=10"

# Or open in browser
start http://localhost:8081/recommendations?user_id=1&count=10
```

### Get Popular Items
```bash
curl "http://localhost:8081/popular?count=20"
curl "http://localhost:8081/popular?category=electronics&count=20"
```

## Development

```bash
# Download dependencies
go mod download
go mod tidy

# Build all services
make build

# Build individually (Windows)
go build -o bin/ingest.exe ./cmd/ingest
go build -o bin/processor.exe ./cmd/processor
go build -o bin/api.exe ./cmd/api

# Run tests
make test
go test ./...
go test -v ./...
go test -cover ./...

# Format code
go fmt ./...
make fmt

# Run locally (need infrastructure up)
go run ./cmd/ingest
go run ./cmd/processor
go run ./cmd/api
```

## Data Inspection

### Redis
```bash
# Enter Redis CLI
docker exec -it reco-redis redis-cli

# Commands inside Redis CLI:
KEYS *                                    # List all keys
LRANGE user:recent:1 0 -1                # User recent items
ZREVRANGE item:popularity 0 10 WITHSCORES # Top popular items
ZREVRANGE co_view:1 0 10 WITHSCORES      # Co-viewed with item 1
GET cache:reco:1                         # Cached recommendations
exit
```

### PostgreSQL
```bash
# Enter PostgreSQL
docker exec -it reco-postgres psql -U reco -d reco

# Commands inside psql:
\dt                                      # List tables
SELECT COUNT(*) FROM events;             # Count events
SELECT * FROM events LIMIT 10;           # View events
SELECT * FROM items;                     # View items
SELECT event_type, COUNT(*) FROM events GROUP BY event_type;  # Events by type
\q                                       # Exit
```

### Kafka
```bash
# List topics
docker exec -it reco-kafka kafka-topics.sh --list --bootstrap-server localhost:9092

# Consume messages
docker exec -it reco-kafka kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic events \
  --from-beginning
```

## Load Testing

```bash
# Install k6 (Windows - Chocolatey)
choco install k6

# Install k6 (Mac)
brew install k6

# Install k6 (Linux)
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Run load test
k6 run scripts/load_test_k6.js

# Custom load test
k6 run --vus 10 --duration 30s scripts/load_test_k6.js
```

## Monitoring

```bash
# Open Prometheus
start http://localhost:9090

# Open Grafana
start http://localhost:3000
# Login: admin/admin

# View metrics directly
curl http://localhost:8080/metrics
curl http://localhost:8081/metrics
```

## Troubleshooting

```bash
# Clean restart
docker-compose down
docker-compose up -d

# Full clean (removes data!)
docker-compose down -v
docker-compose up -d

# View service logs
docker-compose logs --tail=100 ingest
docker-compose logs --tail=100 processor
docker-compose logs --tail=100 api

# Check service health
curl http://localhost:8080/health
curl http://localhost:8081/health

# Restart specific service
docker-compose restart ingest
docker-compose restart processor
docker-compose restart api

# Rebuild after code changes
docker-compose build
docker-compose up -d

# Remove unused Docker resources
docker system prune
```

## Generate Sample Data

```bash
# Windows - 1000 events
.\scripts\generate_events.ps1 -NumEvents 1000

# Windows - Custom endpoint
.\scripts\generate_events.ps1 -BaseUrl "http://localhost:8080" -NumEvents 500

# Linux/Mac - 1000 events
./scripts/generate_events.sh http://localhost:8080 1000

# Multiple users, multiple sessions
.\scripts\generate_events.ps1 -NumEvents 5000
```

## Useful Queries

### Prometheus Queries
```promql
# Event ingestion rate
rate(events_ingested_total[1m])

# Events by type
sum by(event_type) (rate(events_ingested_total[5m]))

# P95 recommendation latency
histogram_quantile(0.95, rate(recommendation_latency_seconds_bucket[5m]))

# P99 recommendation latency
histogram_quantile(0.99, rate(recommendation_latency_seconds_bucket[5m]))

# Cache hit ratio
recommendation_cache_hits_total / (recommendation_cache_hits_total + recommendation_cache_misses_total)

# Error rate
rate(event_processing_errors_total[5m])
```

### PostgreSQL Queries
```sql
-- Events summary
SELECT 
    event_type, 
    COUNT(*) as count,
    COUNT(DISTINCT user_id) as unique_users,
    COUNT(DISTINCT item_id) as unique_items
FROM events 
GROUP BY event_type;

-- Top items by event count
SELECT 
    i.id, 
    i.title, 
    COUNT(e.id) as event_count
FROM items i
JOIN events e ON i.id = e.item_id
GROUP BY i.id, i.title
ORDER BY event_count DESC
LIMIT 10;

-- User activity
SELECT 
    user_id,
    COUNT(*) as total_events,
    COUNT(DISTINCT item_id) as unique_items_viewed
FROM events
GROUP BY user_id
ORDER BY total_events DESC
LIMIT 10;
```

## Port Reference

| Service | Port | URL |
|---------|------|-----|
| Ingest API | 8080 | http://localhost:8080 |
| Recommendation API | 8081 | http://localhost:8081 |
| PostgreSQL | 5432 | localhost:5432 |
| Redis | 6379 | localhost:6379 |
| Kafka | 9092 | localhost:9092 |
| Zookeeper | 2181 | localhost:2181 |
| Prometheus | 9090 | http://localhost:9090 |
| Grafana | 3000 | http://localhost:3000 |

## Environment Variables

Create `.env` file (copy from `.env.example`):

```bash
# Copy example
cp .env.example .env

# Or on Windows
copy .env.example .env

# Edit as needed
notepad .env
```

## Quick Health Check

```bash
# Check all services
curl http://localhost:8080/health  # Ingest
curl http://localhost:8081/health  # API

# Full system check
docker-compose ps
docker-compose logs --tail=10
```

## Backup & Restore

```bash
# Backup PostgreSQL
docker exec reco-postgres pg_dump -U reco reco > backup.sql

# Restore PostgreSQL
docker exec -i reco-postgres psql -U reco reco < backup.sql

# Backup Redis
docker exec reco-redis redis-cli SAVE
docker cp reco-redis:/data/dump.rdb ./redis-backup.rdb

# Restore Redis
docker cp ./redis-backup.rdb reco-redis:/data/dump.rdb
docker-compose restart redis
```
