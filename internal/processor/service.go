package processor

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

// Service handles stream processing
type Service struct {
	kafkaReader *kafka.Reader
	redisStore  *store.RedisStore
	cfg         *config.Config
}

// NewService creates a new processor service
func NewService(cfg *config.Config, redisStore *store.RedisStore) *Service {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Kafka.Brokers,
		Topic:          cfg.Kafka.Topics.Events,
		GroupID:        cfg.Kafka.ConsumerGroup,
		MinBytes:       1,
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})

	return &Service{
		kafkaReader: reader,
		redisStore:  redisStore,
		cfg:         cfg,
	}
}

// Close closes the service
func (s *Service) Close() error {
	return s.kafkaReader.Close()
}

// Start starts consuming and processing events
func (s *Service) Start(ctx context.Context) error {
	logger.Info("Starting event processor")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping event processor")
			return nil
		default:
			msg, err := s.kafkaReader.FetchMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					return nil
				}
				logger.Error("Failed to fetch message", zap.Error(err))
				continue
			}

			if err := s.processMessage(ctx, msg); err != nil {
				logger.Error("Failed to process message",
					zap.Error(err),
					zap.String("key", string(msg.Key)))
				metrics.EventProcessingErrors.Inc()
			} else {
				// Commit message
				if err := s.kafkaReader.CommitMessages(ctx, msg); err != nil {
					logger.Error("Failed to commit message", zap.Error(err))
				}
			}

			metrics.KafkaMessagesConsumed.WithLabelValues(s.cfg.Kafka.Topics.Events).Inc()
		}
	}
}

func (s *Service) processMessage(ctx context.Context, msg kafka.Message) error {
	var event models.Event
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// Process event based on type
	if err := s.processEvent(ctx, &event); err != nil {
		return fmt.Errorf("failed to process event: %w", err)
	}

	metrics.EventsProcessed.WithLabelValues(event.EventType).Inc()

	logger.Debug("Event processed",
		zap.Int64("user_id", event.UserID),
		zap.Int64("item_id", event.ItemID),
		zap.String("event_type", event.EventType))

	return nil
}

func (s *Service) processEvent(ctx context.Context, event *models.Event) error {
	// 1. Update user recent items
	if err := s.redisStore.AddRecentItem(ctx, event.UserID, event.ItemID, s.cfg.Processing.RecentItemsLimit); err != nil {
		logger.Error("Failed to add recent item", zap.Error(err))
	}

	// 2. Update item popularity with weighted score
	weight := s.getEventWeight(event.EventType)
	if err := s.redisStore.IncrPopularity(ctx, event.ItemID, weight); err != nil {
		logger.Error("Failed to increment popularity", zap.Error(err))
	}

	// 3. Update co-view counts
	if err := s.updateCoView(ctx, event); err != nil {
		logger.Error("Failed to update co-view", zap.Error(err))
	}

	return nil
}

func (s *Service) updateCoView(ctx context.Context, event *models.Event) error {
	// Get user's recent items
	recentItems, err := s.redisStore.GetRecentItems(ctx, event.UserID, s.cfg.Processing.CoviewWindow)
	if err != nil {
		return err
	}

	// Update co-view for each recent item
	for _, recentItemStr := range recentItems {
		recentItemID, err := strconv.ParseInt(recentItemStr, 10, 64)
		if err != nil {
			continue
		}

		// Skip if same item
		if recentItemID == event.ItemID {
			continue
		}

		// Increment co-view count (bidirectional)
		if err := s.redisStore.IncrCoView(ctx, event.ItemID, recentItemID); err != nil {
			logger.Error("Failed to increment co-view", zap.Error(err))
		}
		if err := s.redisStore.IncrCoView(ctx, recentItemID, event.ItemID); err != nil {
			logger.Error("Failed to increment co-view", zap.Error(err))
		}
	}

	return nil
}

func (s *Service) getEventWeight(eventType string) float64 {
	switch eventType {
	case models.EventTypeView:
		return s.cfg.EventWeights.View
	case models.EventTypeClick:
		return s.cfg.EventWeights.Click
	case models.EventTypeCart:
		return s.cfg.EventWeights.Cart
	case models.EventTypePurchase:
		return s.cfg.EventWeights.Purchase
	default:
		return 1.0
	}
}
