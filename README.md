

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](docker-compose.yml)

A hybrid (real-time + offline) recommendation system that collects user interaction events, processes them in real-time for fast scoring (popularity, co-view, session-based), and uses offline models (embeddings/collaborative filtering) for high-quality recommendations.


This project demonstrates:
- ✅ **Production-grade Go microservices**
- ✅ **Real-time data streaming** with Kafka
- ✅ **Low-latency API** (<100ms P99)
- ✅ **Scalable architecture** (horizontal scaling)
- ✅ **Full observability** (Prometheus + Grafana)
- ✅ **Modern tech stack** (Kafka, Redis, PostgreSQL)
- ✅ **Comprehensive documentation**
- ✅ **Testing & monitoring**

## ⭐ Key Features

### Real-time Processing
- 🔥 **Event Ingestion**: High-throughput HTTP API with Kafka
- ⚡ **Stream Processing**: Real-time feature aggregation
- 🎯 **Sub-second Latency**: Fast event processing pipeline

### Hybrid Recommendations
- 🤝 **Co-view Matrix**: Session-based item affinity
- 📈 **Popularity Scoring**: Weighted by event importance
- 🧠 **Ready for ML**: Infrastructure for embeddings & ANN
- 🎨 **Multi-signal**: Combines multiple recommendation strategies

### Production-Ready
- 📊 **Full Observability**: Prometheus metrics + OpenTelemetry tracing
- 🐳 **Containerized**: Docker & Docker Compose ready
- 📈 **Scalable**: Horizontal scaling for all services
- 🛡️ **Resilient**: Health checks, graceful shutdown, error handling
- 📝 **Well-documented**: Comprehensive guides and examples

## 🏗️ Architecture

### High-Level Design

```
┌──────────┐
│  Client  │
└─────┬────┘
      │ HTTP
      ▼
┌─────────────────┐      ┌────────────┐
│  Ingest API     │─────▶│   Kafka    │
│  (Port 8080)    │      │  (events)  │
└─────────────────┘      └──────┬─────┘
                                │
        ┌───────────────────────┼────────────────────┐
        │                       │                    │
        ▼                       ▼                    ▼
┌──────────────┐        ┌────────────┐       ┌─────────────┐
│  PostgreSQL  │        │ Processor  │       │    Redis    │
│  (Metadata)  │        │  Service   │──────▶│  (Features) │
└──────────────┘        └────────────┘       └──────┬──────┘
        │                                            │
        │               ┌────────────────────────────┘
        │               │
        ▼               ▼
   ┌────────────────────────┐
   │  Recommendation API    │
   │     (Port 8081)        │
   └────────────────────────┘
```

### Components

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Ingest API** | Go + Gin | Accept events, publish to Kafka |
| **Processor** | Go + Kafka | Real-time aggregations to Redis |
| **Recommendation API** | Go + Gin | Serve recommendations |
| **Feature Store** | Redis | Online features & caching |
| **Metadata DB** | PostgreSQL | Items, users, events |
| **Message Queue** | Kafka | Event streaming |
| **Monitoring** | Prometheus + Grafana | Metrics & dashboards |

For detailed architecture, see **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)**

## Tech Stack

- **Go 1.21+** - Core services (ingest, processor, API)
- **Kafka** - Event streaming
- **Redis 7.x** - Online feature store and caching
- **PostgreSQL 15** - Metadata, user profiles, item catalog
- **Docker** - Containerization
- **Prometheus + Grafana** - Monitoring

## 🚀 Quick Start

### Prerequisites

- Docker Desktop (required)
- Go 1.21+ (optional, for local development)

### One-Command Setup

**Windows (PowerShell):**
```powershell
.\setup.ps1
```

**Linux/Mac:**
```bash
chmod +x setup.sh
./setup.sh
```

This will:
- ✅ Start all services (PostgreSQL, Redis, Kafka, APIs)
- ✅ Check health of all components
- ✅ Display service URLs and next steps

### Manual Setup

```bash
# Start all services
docker-compose up -d

# Generate sample events (1000 events)
.\scripts\generate_events.ps1 -NumEvents 1000    # Windows
./scripts/generate_events.sh http://localhost:8080 1000  # Linux/Mac

# Test recommendations
curl "http://localhost:8081/recommendations?user_id=1&count=10"
```

### 📖 Complete Guide

For detailed setup instructions, see **[GETTING_STARTED.md](GETTING_STARTED.md)** (includes troubleshooting!)

## 🎮 Quick Test

### Ingest an Event
```bash
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "item_id": 5,
    "event_type": "VIEW",
    "session_id": "test_session"
  }'
```

### Get Recommendations
```bash
curl "http://localhost:8081/recommendations?user_id=1&count=10"
```

### Get Popular Items
```bash
curl "http://localhost:8081/popular?category=electronics&count=20"
```

## Services

### Event Ingest API (Port 8080)

Accepts user interaction events and publishes to Kafka.

```bash
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "item_id": 456,
    "event_type": "VIEW",
    "session_id": "abc-123"
  }'
```

### Recommendation API (Port 8081)

Serves personalized recommendations.

```bash
# Get personalized recommendations
curl http://localhost:8081/recommendations?user_id=123&count=10

# Get popular items
curl http://localhost:8081/popular?category=electronics&count=20
```

### Stream Processor

Consumes events from Kafka and updates Redis feature store in real-time.

## API Endpoints

### POST /events
Ingest user interaction event.

**Request:**
```json
{
  "user_id": 123,
  "item_id": 456,
  "event_type": "VIEW",
  "session_id": "abc-123",
  "timestamp": "2025-11-01T12:34:56Z"
}
```

### GET /recommendations
Get personalized recommendations for a user.

**Query Parameters:**
- `user_id` (required): User ID
- `count` (optional): Number of recommendations (default: 10)

**Response:**
```json
{
  "user_id": 123,
  "recommendations": [
    {"item_id": 111, "score": 0.92, "reason": "co_view"},
    {"item_id": 222, "score": 0.89, "reason": "embedding"}
  ]
}
```

### GET /popular
Get popular items.

**Query Parameters:**
- `category` (optional): Filter by category
- `count` (optional): Number of items (default: 20)

### POST /admin/retrain
Trigger offline model retraining (requires authentication).

## Development

### Project Structure

```
reco-engine/
├─ cmd/                 # Service entry points
│  ├─ ingest/          # Event ingest service
│  ├─ processor/       # Stream processor service
│  └─ api/             # Recommendation API service
├─ internal/           # Internal packages
│  ├─ ingest/         # Ingest handlers
│  ├─ processor/      # Event processing logic
│  ├─ api/            # API handlers
│  ├─ store/          # Database and Redis clients
│  ├─ models/         # Data models
│  └─ util/           # Utilities (config, logging, metrics)
├─ infra/             # Infrastructure configs
│  ├─ docker/         # Dockerfiles
│  └─ postgres/       # SQL schemas
├─ scripts/           # Utility scripts
└─ docker-compose.yml
```

### Build Services

```bash
# Build all services
make build

# Build specific service
go build -o bin/ingest ./cmd/ingest
go build -o bin/processor ./cmd/processor
go build -o bin/api ./cmd/api
```

### Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./...
```

### Load Testing

```bash
# Install k6
# Run load test
k6 run scripts/load_test_k6.js
```

## 📚 Documentation

| Document | Description |
|----------|-------------|
| **[GETTING_STARTED.md](GETTING_STARTED.md)** | 🇮🇩 **Mulai di sini!** Panduan lengkap untuk pemula (Bahasa Indonesia) |
| **[SUMMARY.md](SUMMARY.md)** | 📋 Project summary & highlights |
| **[COMMANDS.md](COMMANDS.md)** | ⚡ Quick reference for all commands |
| **[PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md)** | 🏗️ Detailed project structure |
| **[docs/QUICKSTART.md](docs/QUICKSTART.md)** | 🚀 Quick start guide |
| **[docs/API.md](docs/API.md)** | 📖 Complete API documentation |
| **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** | 🏛️ System architecture & algorithms |
| **[docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)** | 🚢 Production deployment guide |
| **[CONTRIBUTING.md](CONTRIBUTING.md)** | 🤝 How to contribute |
| **[CHANGELOG.md](CHANGELOG.md)** | 📝 Version history |

## 🛠️ Development

### Local Development Setup

```bash
# Install dependencies
go mod download

# Start infrastructure only
docker-compose up -d postgres redis kafka

# Run services locally
go run ./cmd/ingest      # Terminal 1
go run ./cmd/processor   # Terminal 2
go run ./cmd/api         # Terminal 3

# Run tests
go test ./...
go test -v -cover ./...
```

### Build Services

```bash
# Build all services
make build

# Build individually (Windows)
go build -o bin/ingest.exe ./cmd/ingest
go build -o bin/processor.exe ./cmd/processor
go build -o bin/api.exe ./cmd/api
```

### Code Quality

```bash
# Format code
go fmt ./...
make fmt

# Run linter (requires golangci-lint)
make lint

# Generate test coverage
make test-coverage
```

## 📊 Monitoring & Observability

### Service URLs

| Service | URL | Description |
|---------|-----|-------------|
| 📥 Event Ingest API | http://localhost:8080 | Ingest user events |
| 🎯 Recommendation API | http://localhost:8081 | Get recommendations |
| 📈 Prometheus | http://localhost:9090 | Metrics & monitoring |
| 📊 Grafana | http://localhost:3000 | Dashboards (admin/admin) |

### Key Metrics

- `events_ingested_total` - Total events ingested by type
- `events_processed_total` - Total events processed
- `recommendation_latency_seconds` - API latency histogram
- `recommendation_cache_hits_total` - Cache performance
- `kafka_messages_published_total` - Message throughput

### Example Prometheus Queries

```promql
# Event ingestion rate per second
rate(events_ingested_total[1m])

# P95 recommendation latency
histogram_quantile(0.95, rate(recommendation_latency_seconds_bucket[5m]))

# Cache hit ratio
recommendation_cache_hits_total / (recommendation_cache_hits_total + recommendation_cache_misses_total)
```

## 🧪 Testing

### Unit Tests
```bash
go test ./...
go test -cover ./...
```

### Load Testing (k6)
```bash
k6 run scripts/load_test_k6.js
```

### Integration Testing
```bash
# Generate 1000 sample events
.\scripts\generate_events.ps1 -NumEvents 1000

# Verify recommendations work
curl "http://localhost:8081/recommendations?user_id=1&count=10"
```


## 🚢 Production Deployment

### Scaling Guidelines

| Service | Replicas | CPU | Memory |
|---------|----------|-----|--------|
| Ingest API | 3-10 | 0.5-1 | 512MB-1GB |
| Processor | 3-5 | 0.5-1 | 512MB-1GB |
| Recommendation API | 3-10 | 0.5-1 | 512MB-1GB |

### Deployment Options

- **Kubernetes** - Full K8s manifests ready
- **Docker Swarm** - Swarm stack files
- **Cloud Services** - AWS ECS/EKS, GCP GKE, Azure AKS

See **[docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)** for complete guide.

## 🔐 Security

Production considerations:
- ✅ API authentication (API keys/JWT)
- ✅ Rate limiting per IP/user
- ✅ TLS/HTTPS for all connections
- ✅ Input validation & sanitization
- ✅ GDPR compliance (user opt-out, data anonymization)
- ✅ Network isolation (VPC/private subnets)


## 🎯 Use Cases

Perfect for:
- 🛍️ **E-commerce** - Product recommendations
- 📰 **Content platforms** - Article/video recommendations
- 🎵 **Music/Video streaming** - Personalized playlists
- 📱 **Mobile apps** - In-app recommendations
- 🏪 **Retail** - Cross-sell & upsell


### Quick Links

- 🇮🇩 [**Panduan Bahasa Indonesia**](GETTING_STARTED.md)
- 📖 [API Documentation](docs/API.md)
- 🏗️ [Architecture Guide](docs/ARCHITECTURE.md)
- 🚢 [Deployment Guide](docs/DEPLOYMENT.md)
- ⚡ [Command Reference](COMMANDS.md)
