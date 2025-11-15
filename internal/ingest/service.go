package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/yourusername/reco-engine/internal/models"
	"github.com/yourusername/reco-engine/internal/store"
	"github.com/yourusername/reco-engine/internal/util/config"
	"github.com/yourusername/reco-engine/internal/util/logger"
	"github.com/yourusername/reco-engine/internal/util/metrics"
	"go.uber.org/zap"
)

// Service handles event ingestion
type Service struct {
	kafkaWriter *kafka.Writer
	pgStore     *store.PostgresStore
	cfg         *config.Config
}

// NewService creates a new ingest service
func NewService(cfg *config.Config, pgStore *store.PostgresStore) *Service {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      cfg.Kafka.Brokers,
		Topic:        cfg.Kafka.Topics.Events,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    cfg.Processing.BatchSize,
		BatchTimeout: 10 * time.Millisecond,
		Async:        true,
		RequiredAcks: int(kafka.RequireOne),
	})

	return &Service{
		kafkaWriter: writer,
		pgStore:     pgStore,
		cfg:         cfg,
	}
}

// Close closes the service
func (s *Service) Close() error {
	return s.kafkaWriter.Close()
}

// IngestEvent ingests an event and publishes to Kafka
func (s *Service) IngestEvent(ctx context.Context, event *models.Event) error {
	// Validate event
	if err := s.validateEvent(event); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Serialize event
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish to Kafka
	msg := kafka.Message{
		Key:   []byte(strconv.FormatInt(event.UserID, 10)),
		Value: eventBytes,
		Time:  event.Timestamp,
	}

	if err := s.kafkaWriter.WriteMessages(ctx, msg); err != nil {
		metrics.KafkaPublishErrors.WithLabelValues(s.cfg.Kafka.Topics.Events).Inc()
		return fmt.Errorf("failed to publish to Kafka: %w", err)
	}

	// Update metrics
	metrics.EventsIngested.WithLabelValues(event.EventType).Inc()
	metrics.KafkaMessagesPublished.WithLabelValues(s.cfg.Kafka.Topics.Events).Inc()

	logger.Debug("Event ingested",
		zap.Int64("user_id", event.UserID),
		zap.Int64("item_id", event.ItemID),
		zap.String("event_type", event.EventType),
		zap.String("session_id", event.SessionID))

	// Optionally store in PostgreSQL (async)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.pgStore.InsertEvent(ctx, event); err != nil {
			logger.Error("Failed to store event in PostgreSQL", zap.Error(err))
		}
	}()

	return nil
}

func (s *Service) validateEvent(event *models.Event) error {
	if event.UserID <= 0 {
		return fmt.Errorf("user_id is required")
	}
	if event.ItemID <= 0 {
		return fmt.Errorf("item_id is required")
	}
	if event.EventType == "" {
		return fmt.Errorf("event_type is required")
	}

	// Validate event type
	validTypes := map[string]bool{
		models.EventTypeView:     true,
		models.EventTypeClick:    true,
		models.EventTypeCart:     true,
		models.EventTypePurchase: true,
	}
	if !validTypes[event.EventType] {
		return fmt.Errorf("invalid event_type: %s", event.EventType)
	}

	return nil
}
