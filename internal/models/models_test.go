package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventSerialization(t *testing.T) {
	event := &Event{
		UserID:    123,
		ItemID:    456,
		EventType: EventTypeView,
		SessionID: "session_123",
		Timestamp: time.Now(),
	}

	// Marshal to JSON
	data, err := json.Marshal(event)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Unmarshal from JSON
	var decoded Event
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, event.UserID, decoded.UserID)
	assert.Equal(t, event.ItemID, decoded.ItemID)
	assert.Equal(t, event.EventType, decoded.EventType)
}

func TestRecommendationResponse(t *testing.T) {
	resp := &RecommendationResponse{
		UserID: 123,
		Recommendations: []Recommendation{
			{ItemID: 1, Score: 0.9, Reason: "co_view"},
			{ItemID: 2, Score: 0.8, Reason: "embedding"},
		},
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)

	var decoded RecommendationResponse
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, resp.UserID, decoded.UserID)
	assert.Len(t, decoded.Recommendations, 2)
}
