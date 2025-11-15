# API Documentation

## Base URLs

- **Event Ingest API**: `http://localhost:8080`
- **Recommendation API**: `http://localhost:8081`

## Authentication

Currently, the API is open for development. In production, implement:
- API keys for service-to-service authentication
- JWT tokens for user authentication
- Rate limiting per API key

---

## Event Ingest API

### POST /events

Ingest a user interaction event.

#### Request

**Headers:**
```
Content-Type: application/json
```

**Body:**
```json
{
  "user_id": 123,
  "item_id": 456,
  "event_type": "VIEW",
  "session_id": "abc-123",
  "timestamp": "2025-11-01T12:34:56Z",
  "metadata": {
    "source": "mobile_app",
    "version": "1.0.0"
  }
}
```

**Fields:**
- `user_id` (required, integer): User ID
- `item_id` (required, integer): Item/Product ID
- `event_type` (required, string): One of: `VIEW`, `CLICK`, `CART`, `PURCHASE`
- `session_id` (optional, string): Session identifier
- `timestamp` (optional, string): ISO 8601 timestamp (defaults to server time)
- `metadata` (optional, object): Additional event metadata

#### Response

**Success (200 OK):**
```json
{
  "status": "ok"
}
```

**Error (400 Bad Request):**
```json
{
  "error": "invalid request body"
}
```

**Error (500 Internal Server Error):**
```json
{
  "error": "failed to publish to Kafka: ..."
}
```

#### Examples

```bash
# View event
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "item_id": 456,
    "event_type": "VIEW",
    "session_id": "session_123"
  }'

# Purchase event
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "item_id": 456,
    "event_type": "PURCHASE",
    "session_id": "session_123",
    "metadata": {"amount": 150000}
  }'
```

---

### GET /health

Health check endpoint.

#### Response

```json
{
  "status": "healthy"
}
```

---

## Recommendation API

### GET /recommendations

Get personalized recommendations for a user.

#### Query Parameters

- `user_id` (required, integer): User ID
- `count` (optional, integer, default=10, max=100): Number of recommendations

#### Response

**Success (200 OK):**
```json
{
  "user_id": 123,
  "recommendations": [
    {
      "item_id": 111,
      "score": 0.92,
      "reason": "co_view"
    },
    {
      "item_id": 222,
      "score": 0.89,
      "reason": "embedding"
    },
    {
      "item_id": 333,
      "score": 0.85,
      "reason": "popular"
    }
  ]
}
```

**Fields:**
- `item_id`: Recommended item ID
- `score`: Recommendation score (0-1, higher is better)
- `reason`: Reason for recommendation (`co_view`, `embedding`, `popular`)

**Error (400 Bad Request):**
```json
{
  "error": "user_id is required"
}
```

#### Examples

```bash
# Get 10 recommendations
curl "http://localhost:8081/recommendations?user_id=123&count=10"

# Get 20 recommendations
curl "http://localhost:8081/recommendations?user_id=123&count=20"
```

---

### GET /popular

Get popular items.

#### Query Parameters

- `category` (optional, string): Filter by category
- `count` (optional, integer, default=20, max=100): Number of items

#### Response

**Success (200 OK):**
```json
{
  "category": "electronics",
  "recommendations": [
    {
      "item_id": 1,
      "score": 150.5,
      "reason": "popular"
    },
    {
      "item_id": 2,
      "score": 142.3,
      "reason": "popular"
    }
  ]
}
```

#### Examples

```bash
# Get top 20 popular items
curl "http://localhost:8081/popular?count=20"

# Get top 20 popular electronics
curl "http://localhost:8081/popular?category=electronics&count=20"
```

---

### GET /health

Health check endpoint.

#### Response

```json
{
  "status": "healthy"
}
```

---

## Metrics Endpoints

### GET /metrics

Prometheus metrics endpoint (available on both services).

#### Response

```
# HELP events_ingested_total Total number of events ingested
# TYPE events_ingested_total counter
events_ingested_total{event_type="VIEW"} 1234
events_ingested_total{event_type="CLICK"} 567
events_ingested_total{event_type="CART"} 89
events_ingested_total{event_type="PURCHASE"} 45

# HELP recommendation_latency_seconds Latency of recommendation requests
# TYPE recommendation_latency_seconds histogram
recommendation_latency_seconds_bucket{endpoint="personalized",le="0.005"} 100
recommendation_latency_seconds_bucket{endpoint="personalized",le="0.01"} 250
...
```

---

## Error Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 400 | Bad Request - Invalid input |
| 500 | Internal Server Error |

---

## Rate Limits

**Development:** No limits

**Production (recommended):**
- Event Ingest: 1000 requests/minute per IP
- Recommendations: 100 requests/minute per user
- Popular Items: 10 requests/minute per IP

Implement using middleware or API gateway (e.g., Kong, Nginx).

---

## Data Model

### Event Types

| Type | Weight | Description |
|------|--------|-------------|
| VIEW | 1.0 | User viewed item page |
| CLICK | 3.0 | User clicked on item |
| CART | 5.0 | User added to cart |
| PURCHASE | 10.0 | User purchased item |

### Recommendation Reasons

| Reason | Description |
|--------|-------------|
| co_view | Items frequently viewed together |
| embedding | Similar items based on ML model |
| popular | Trending/popular items |

---

## Best Practices

1. **Batching**: For bulk event ingestion, send events in batches
2. **Caching**: Recommendations are cached for 5 minutes by default
3. **Session IDs**: Use consistent session IDs for better co-view tracking
4. **Metadata**: Include useful metadata for analytics
5. **Error Handling**: Implement retry logic with exponential backoff

---

## Integration Examples

### JavaScript/Node.js

```javascript
// Ingest event
async function trackEvent(userId, itemId, eventType, sessionId) {
  const response = await fetch('http://localhost:8080/events', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      user_id: userId,
      item_id: itemId,
      event_type: eventType,
      session_id: sessionId
    })
  });
  return response.json();
}

// Get recommendations
async function getRecommendations(userId, count = 10) {
  const response = await fetch(
    `http://localhost:8081/recommendations?user_id=${userId}&count=${count}`
  );
  return response.json();
}
```

### Python

```python
import requests

# Ingest event
def track_event(user_id, item_id, event_type, session_id):
    response = requests.post('http://localhost:8080/events', json={
        'user_id': user_id,
        'item_id': item_id,
        'event_type': event_type,
        'session_id': session_id
    })
    return response.json()

# Get recommendations
def get_recommendations(user_id, count=10):
    response = requests.get(
        f'http://localhost:8081/recommendations',
        params={'user_id': user_id, 'count': count}
    )
    return response.json()
```

### Go

```go
// Ingest event
type Event struct {
    UserID    int64  `json:"user_id"`
    ItemID    int64  `json:"item_id"`
    EventType string `json:"event_type"`
    SessionID string `json:"session_id"`
}

func trackEvent(event Event) error {
    data, _ := json.Marshal(event)
    resp, err := http.Post(
        "http://localhost:8080/events",
        "application/json",
        bytes.NewBuffer(data),
    )
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    return nil
}

// Get recommendations
func getRecommendations(userID int64, count int) ([]Recommendation, error) {
    url := fmt.Sprintf(
        "http://localhost:8081/recommendations?user_id=%d&count=%d",
        userID, count,
    )
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result RecommendationResponse
    json.NewDecoder(resp.Body).Decode(&result)
    return result.Recommendations, nil
}
```
