package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Events metrics
	EventsIngested = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "events_ingested_total",
			Help: "Total number of events ingested",
		},
		[]string{"event_type"},
	)

	EventsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "events_processed_total",
			Help: "Total number of events processed",
		},
		[]string{"event_type"},
	)

	EventProcessingErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "event_processing_errors_total",
			Help: "Total number of event processing errors",
		},
	)

	// Recommendation metrics
	RecommendationRequests = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "recommendation_requests_total",
			Help: "Total number of recommendation requests",
		},
	)

	RecommendationLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "recommendation_latency_seconds",
			Help:    "Latency of recommendation requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)

	RecommendationCacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "recommendation_cache_hits_total",
			Help: "Total number of cache hits for recommendations",
		},
	)

	RecommendationCacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "recommendation_cache_misses_total",
			Help: "Total number of cache misses for recommendations",
		},
	)

	// Redis metrics
	RedisOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_operations_total",
			Help: "Total number of Redis operations",
		},
		[]string{"operation", "status"},
	)

	RedisLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_latency_seconds",
			Help:    "Latency of Redis operations",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation"},
	)

	// Kafka metrics
	KafkaMessagesPublished = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_messages_published_total",
			Help: "Total number of messages published to Kafka",
		},
		[]string{"topic"},
	)

	KafkaMessagesConsumed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_messages_consumed_total",
			Help: "Total number of messages consumed from Kafka",
		},
		[]string{"topic"},
	)

	KafkaPublishErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_publish_errors_total",
			Help: "Total number of Kafka publish errors",
		},
		[]string{"topic"},
	)
)
