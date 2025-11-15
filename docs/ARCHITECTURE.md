# Architecture Documentation

## System Overview

The Recommendation Engine is a hybrid real-time and offline system designed to provide low-latency, high-quality product recommendations for e-commerce platforms.

## Core Components

### 1. Event Ingest Service (Port 8080)

**Responsibility:** Receive user interaction events and publish to message queue.

**Technology:** Go, Gin, Kafka

**Key Features:**
- HTTP endpoint for event ingestion
- Event validation
- Asynchronous publishing to Kafka
- Async persistence to PostgreSQL
- Prometheus metrics

**Flow:**
```
Client → HTTP POST → Validate → Kafka Publish → Response
                         ↓
                   PostgreSQL (async)
```

### 2. Stream Processor Service

**Responsibility:** Real-time event processing and feature aggregation.

**Technology:** Go, Kafka Consumer, Redis

**Key Features:**
- Consumes events from Kafka
- Updates real-time features in Redis:
  - User recent items (LRU list)
  - Item popularity scores (sorted set with decay)
  - Co-view matrices (item-item affinity)
- Sliding window aggregations

**Processing Logic:**
```
Kafka Event → Deserialize → Update Redis Features
                              ├─ user:recent:{user_id}
                              ├─ item:popularity
                              └─ co_view:{item_id}
```

### 3. Recommendation API Service (Port 8081)

**Responsibility:** Serve personalized and popular recommendations.

**Technology:** Go, Gin, Redis, PostgreSQL

**Key Features:**
- Personalized recommendations endpoint
- Popular items endpoint
- Multi-signal candidate generation
- Scoring and ranking
- Response caching (5-minute TTL)

**Recommendation Flow:**
```
Request → Cache Check → Generate Candidates → Score & Rank → Response
                              ├─ Co-view expansion
                              ├─ Embedding neighbors (KNN)
                              └─ Popular items (fallback)
```

### 4. Feature Store (Redis)

**Data Structures:**

| Key Pattern | Type | Purpose | TTL |
|-------------|------|---------|-----|
| `user:recent:{user_id}` | List | User's recent items (LRU) | 24h |
| `item:popularity` | Sorted Set | Global popularity scores | None |
| `co_view:{item_id}` | Sorted Set | Co-viewed items | 7d |
| `item:knn:{item_id}` | List | Precomputed neighbors | 7d |
| `cache:reco:{user_id}` | String | Cached recommendations | 5m |

### 5. Metadata Store (PostgreSQL)

**Tables:**
- `items` - Product catalog
- `users` - User profiles
- `events` - Event log (append-only)
- `models` - ML model metadata

### 6. Message Queue (Kafka)

**Topics:**
- `events` - User interaction events (3 partitions)

**Consumer Groups:**
- `reco-processor` - Stream processor

## Data Flow

### Event Ingestion Flow

```
┌─────────┐
│ Client  │
└────┬────┘
     │ 1. POST /events
     ▼
┌─────────────┐
│Ingest API   │
└──┬──────┬───┘
   │      │ 3. Async write
   │      ▼
   │   ┌──────────┐
   │   │Postgres  │
   │   │(events)  │
   │   └──────────┘
   │ 2. Publish
   ▼
┌─────────┐
│  Kafka  │
│ (events)│
└────┬────┘
     │ 4. Consume
     ▼
┌────────────┐
│ Processor  │
└─────┬──────┘
      │ 5. Update features
      ▼
   ┌─────┐
   │Redis│
   └─────┘
```

### Recommendation Generation Flow

```
┌─────────┐
│ Client  │
└────┬────┘
     │ 1. GET /recommendations?user_id=X
     ▼
┌──────────────┐
│Reco API      │
└──┬───────────┘
   │ 2. Check cache
   ▼
┌─────────────┐    ┌──────────┐
│Redis (cache)│───▶│ Return   │
└─────────────┘    └──────────┘
   │ Cache miss
   │ 3. Generate
   ▼
┌─────────────────────┐
│Candidate Generation │
├─────────────────────┤
│• Get recent items   │
│• Expand via co-view │
│• Add KNN items      │
│• Add popular items  │
└──────┬──────────────┘
       │ 4. Score & rank
       ▼
┌─────────────────┐
│Score = W₁·CoView│
│      + W₂·Embed │
│      + W₃·Pop   │
│      + W₄·Rec   │
└──────┬──────────┘
       │ 5. Top-N
       ▼
    Response
```

## Algorithms

### Real-Time Signals

#### 1. Co-View Matrix

**Concept:** Items viewed together in the same session or time window.

**Update:**
```
For each event:
  recent_items = GET user:recent:{user_id}
  For each item in recent_items:
    ZINCRBY co_view:{item_id} 1 {event.item_id}
```

**Score:** Frequency-based affinity

#### 2. Popularity Score

**Concept:** Weighted event counts with recency decay.

**Update:**
```
weight = event_weights[event_type]
ZINCRBY item:popularity weight {item_id}
```

**Weights:**
- VIEW: 1.0
- CLICK: 3.0
- CART: 5.0
- PURCHASE: 10.0

#### 3. Session-Based

**Concept:** Items from user's recent interactions.

**Data:**
```
LPUSH user:recent:{user_id} {item_id}
LTRIM user:recent:{user_id} 0 49  // Keep last 50
```

### Offline Signals (Future)

#### 1. Item Embeddings

**Approach:** Item2Vec (skip-gram on user sessions)

**Training:**
- Daily batch job
- Input: User session sequences
- Output: 128-dim embeddings per item

**Serving:**
- Store in `item:knn:{item_id}` (top 100 neighbors)
- Or query ANN index online (Milvus/Faiss)

#### 2. Collaborative Filtering

**Approach:** Matrix Factorization (ALS on implicit feedback)

**Training:**
- Weekly batch job
- Input: User-item interaction matrix
- Output: User/item factors

### Scoring Formula

```
final_score = w_coview * coview_score
            + w_embed * embedding_score
            + w_pop * popularity_score
            + w_rec * recency_score
```

**Default weights:** (from config)
- Co-view: 0.4
- Embedding: 0.3
- Popularity: 0.2
- Recency: 0.1

## Scalability

### Horizontal Scaling

**Stateless Services (can scale horizontally):**
- Ingest API (load balanced)
- Recommendation API (load balanced)
- Stream Processor (via Kafka partitions)

**Stateful Services:**
- Redis (Redis Cluster for sharding)
- PostgreSQL (read replicas)
- Kafka (partitioning)

### Performance Targets

| Metric | Target |
|--------|--------|
| Event ingestion latency | < 50ms (p99) |
| Recommendation API latency | < 100ms (p99) |
| Stream processing lag | < 5 seconds |
| Cache hit ratio | > 80% |

## Monitoring

### Key Metrics

**Ingestion:**
- `events_ingested_total` (by event_type)
- `kafka_messages_published_total`

**Processing:**
- `events_processed_total` (by event_type)
- `event_processing_errors_total`

**Recommendations:**
- `recommendation_requests_total`
- `recommendation_latency_seconds` (histogram)
- `recommendation_cache_hits_total`
- `recommendation_cache_misses_total`

**Infrastructure:**
- Redis operations latency
- Kafka consumer lag
- PostgreSQL query time

### Alerting Rules

1. High error rate (> 5%)
2. High latency (p99 > 500ms)
3. Kafka consumer lag (> 10k messages)
4. Low cache hit ratio (< 60%)

## Security

### Current (Development)

- No authentication
- Open endpoints

### Production Requirements

1. **API Authentication:**
   - API keys for service-to-service
   - JWT for user endpoints

2. **Network Security:**
   - Private VPC for internal services
   - TLS for all connections
   - Firewall rules

3. **Data Security:**
   - Encrypt data at rest (PostgreSQL)
   - Encrypt data in transit (TLS)
   - PII anonymization

4. **Rate Limiting:**
   - Per-IP limits on ingest
   - Per-user limits on recommendations

## Future Enhancements

1. **Offline Training Pipeline:**
   - Airflow/Luigi for orchestration
   - Spark for large-scale processing
   - MLflow for model versioning

2. **Advanced Features:**
   - User embeddings
   - Contextual bandits
   - Multi-armed bandits for exploration
   - Deep learning models (Two-Tower, DLRM)

3. **ANN Integration:**
   - Milvus or Faiss for vector search
   - Online nearest neighbor queries

4. **A/B Testing:**
   - Experiment framework
   - Multi-variant testing
   - Statistical analysis

5. **Personalization:**
   - User profiles
   - Contextual signals (time, device, location)
   - Sequence modeling (RNN/Transformer)

6. **Business Rules:**
   - Stock filtering
   - Price range filtering
   - Category diversity
   - Freshness boost
