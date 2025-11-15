package models

import "time"

// Event represents a user interaction event
type Event struct {
	ID        int64                  `json:"id,omitempty" db:"id"`
	UserID    int64                  `json:"user_id" db:"user_id"`
	ItemID    int64                  `json:"item_id" db:"item_id"`
	EventType string                 `json:"event_type" db:"event_type"`
	SessionID string                 `json:"session_id" db:"session_id"`
	Metadata  map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	Timestamp time.Time              `json:"timestamp" db:"timestamp"`
}

// EventType constants
const (
	EventTypeView     = "VIEW"
	EventTypeClick    = "CLICK"
	EventTypeCart     = "CART"
	EventTypePurchase = "PURCHASE"
)

// Item represents a product/item
type Item struct {
	ID        int64                  `json:"id" db:"id"`
	SKU       string                 `json:"sku" db:"sku"`
	Title     string                 `json:"title" db:"title"`
	Category  string                 `json:"category" db:"category"`
	Price     int64                  `json:"price" db:"price"`
	Stock     int                    `json:"stock" db:"stock"`
	Metadata  map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
}

// User represents a user
type User struct {
	ID         int64                  `json:"id" db:"id"`
	ExternalID string                 `json:"external_id" db:"external_id"`
	Metadata   map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at" db:"updated_at"`
}

// Recommendation represents a single recommendation
type Recommendation struct {
	ItemID int64   `json:"item_id"`
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

// RecommendationResponse is the API response
type RecommendationResponse struct {
	UserID          int64            `json:"user_id"`
	Recommendations []Recommendation `json:"recommendations"`
}

// Model represents an offline trained model
type Model struct {
	ID        int64                  `json:"id" db:"id"`
	ModelName string                 `json:"model_name" db:"model_name"`
	Version   string                 `json:"version" db:"version"`
	ModelType string                 `json:"model_type" db:"model_type"`
	Metrics   map[string]interface{} `json:"metrics" db:"metrics"`
	Config    map[string]interface{} `json:"config" db:"config"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}
