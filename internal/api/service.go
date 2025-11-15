package api

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/yourusername/reco-engine/internal/models"
	"github.com/yourusername/reco-engine/internal/store"
	"github.com/yourusername/reco-engine/internal/util/config"
	"github.com/yourusername/reco-engine/internal/util/logger"
	"github.com/yourusername/reco-engine/internal/util/metrics"
	"go.uber.org/zap"
)

// Service handles recommendation logic
type Service struct {
	redisStore *store.RedisStore
	pgStore    *store.PostgresStore
	cfg        *config.Config
}

// NewService creates a new recommendation service
func NewService(cfg *config.Config, redisStore *store.RedisStore, pgStore *store.PostgresStore) *Service {
	return &Service{
		redisStore: redisStore,
		pgStore:    pgStore,
		cfg:        cfg,
	}
}

// GetRecommendations generates personalized recommendations for a user
func (s *Service) GetRecommendations(ctx context.Context, userID int64, count int) (*models.RecommendationResponse, error) {
	start := time.Now()
	defer func() {
		metrics.RecommendationLatency.WithLabelValues("personalized").Observe(time.Since(start).Seconds())
	}()

	metrics.RecommendationRequests.Inc()

	// Try cache first
	if cached, err := s.getCachedRecommendations(ctx, userID); err == nil {
		metrics.RecommendationCacheHits.Inc()
		logger.Debug("Cache hit for recommendations", zap.Int64("user_id", userID))
		return cached, nil
	}
	metrics.RecommendationCacheMisses.Inc()

	// Generate recommendations
	recommendations, err := s.generateRecommendations(ctx, userID, count)
	if err != nil {
		return nil, err
	}

	response := &models.RecommendationResponse{
		UserID:          userID,
		Recommendations: recommendations,
	}

	// Cache the result
	go s.cacheRecommendations(context.Background(), userID, response)

	return response, nil
}

// GetPopularItems returns popular items
func (s *Service) GetPopularItems(ctx context.Context, category string, count int) ([]models.Recommendation, error) {
	start := time.Now()
	defer func() {
		metrics.RecommendationLatency.WithLabelValues("popular").Observe(time.Since(start).Seconds())
	}()

	// Get popular items from Redis
	popularItems, err := s.redisStore.GetPopularItems(ctx, count*2) // Get more for filtering
	if err != nil {
		return nil, fmt.Errorf("failed to get popular items: %w", err)
	}

	var recommendations []models.Recommendation
	for _, z := range popularItems {
		itemID, err := strconv.ParseInt(z.Member.(string), 10, 64)
		if err != nil {
			continue
		}

		// If category filter, check item category
		if category != "" {
			item, err := s.pgStore.GetItem(ctx, itemID)
			if err != nil || item.Category != category {
				continue
			}
		}

		recommendations = append(recommendations, models.Recommendation{
			ItemID: itemID,
			Score:  z.Score,
			Reason: "popular",
		})

		if len(recommendations) >= count {
			break
		}
	}

	return recommendations, nil
}

func (s *Service) generateRecommendations(ctx context.Context, userID int64, count int) ([]models.Recommendation, error) {
	candidates := make(map[int64]*candidateScore)

	// 1. Get user's recent items
	recentItems, err := s.redisStore.GetRecentItems(ctx, userID, 5)
	if err != nil {
		logger.Warn("Failed to get recent items", zap.Error(err))
	}

	// 2. Expand via co-view
	for _, itemStr := range recentItems {
		itemID, err := strconv.ParseInt(itemStr, 10, 64)
		if err != nil {
			continue
		}

		// Get co-viewed items
		coViewItems, err := s.redisStore.GetCoViewItems(ctx, itemID, 20)
		if err != nil {
			logger.Warn("Failed to get co-view items", zap.Error(err))
			continue
		}

		for _, z := range coViewItems {
			candItemID, err := strconv.ParseInt(z.Member.(string), 10, 64)
			if err != nil {
				continue
			}

			// Skip if in recent items
			if s.isRecentItem(candItemID, recentItems) {
				continue
			}

			if candidates[candItemID] == nil {
				candidates[candItemID] = &candidateScore{}
			}
			candidates[candItemID].coviewScore += z.Score
		}

		// Get KNN items (from offline model)
		knnItems, err := s.redisStore.GetItemKNN(ctx, itemID, 20)
		if err != nil {
			logger.Warn("Failed to get KNN items", zap.Error(err))
			continue
		}

		for idx, knnItemStr := range knnItems {
			knnItemID, err := strconv.ParseInt(knnItemStr, 10, 64)
			if err != nil {
				continue
			}

			// Skip if in recent items
			if s.isRecentItem(knnItemID, recentItems) {
				continue
			}

			if candidates[knnItemID] == nil {
				candidates[knnItemID] = &candidateScore{}
			}
			// Higher score for higher ranked items
			candidates[knnItemID].embeddingScore += float64(20-idx) / 20.0
		}
	}

	// 3. Add popular items as fallback
	if len(candidates) < count {
		popularItems, err := s.redisStore.GetPopularItems(ctx, count*2)
		if err == nil {
			for _, z := range popularItems {
				itemID, err := strconv.ParseInt(z.Member.(string), 10, 64)
				if err != nil {
					continue
				}

				if s.isRecentItem(itemID, recentItems) {
					continue
				}

				if candidates[itemID] == nil {
					candidates[itemID] = &candidateScore{}
				}
				candidates[itemID].popularityScore += z.Score
			}
		}
	}

	// 4. Score and rank candidates
	var recommendations []models.Recommendation
	for itemID, scores := range candidates {
		finalScore := s.calculateFinalScore(scores)
		reason := s.determineReason(scores)

		recommendations = append(recommendations, models.Recommendation{
			ItemID: itemID,
			Score:  finalScore,
			Reason: reason,
		})
	}

	// Sort by score descending
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	// Return top N
	if len(recommendations) > count {
		recommendations = recommendations[:count]
	}

	return recommendations, nil
}

type candidateScore struct {
	coviewScore     float64
	embeddingScore  float64
	popularityScore float64
	recencyScore    float64
}

func (s *Service) calculateFinalScore(scores *candidateScore) float64 {
	weights := s.cfg.Recommendation.Weights
	return scores.coviewScore*weights.Coview +
		scores.embeddingScore*weights.Embedding +
		scores.popularityScore*weights.Popularity +
		scores.recencyScore*weights.Recency
}

func (s *Service) determineReason(scores *candidateScore) string {
	if scores.coviewScore > scores.embeddingScore && scores.coviewScore > scores.popularityScore {
		return "co_view"
	}
	if scores.embeddingScore > scores.popularityScore {
		return "embedding"
	}
	return "popular"
}

func (s *Service) isRecentItem(itemID int64, recentItems []string) bool {
	itemStr := strconv.FormatInt(itemID, 10)
	for _, recent := range recentItems {
		if recent == itemStr {
			return true
		}
	}
	return false
}

func (s *Service) getCachedRecommendations(ctx context.Context, userID int64) (*models.RecommendationResponse, error) {
	data, err := s.redisStore.GetCachedRecommendations(ctx, userID)
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("cache miss")
		}
		return nil, err
	}

	var response models.RecommendationResponse
	if err := json.Unmarshal([]byte(data), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (s *Service) cacheRecommendations(ctx context.Context, userID int64, response *models.RecommendationResponse) {
	data, err := json.Marshal(response)
	if err != nil {
		logger.Error("Failed to marshal recommendations for cache", zap.Error(err))
		return
	}

	if err := s.redisStore.CacheRecommendations(ctx, userID, string(data), s.cfg.Recommendation.CacheTTL); err != nil {
		logger.Error("Failed to cache recommendations", zap.Error(err))
	}
}
