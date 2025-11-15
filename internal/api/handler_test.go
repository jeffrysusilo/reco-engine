package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/yourusername/reco-engine/internal/util/config"
)

func TestHandleGetRecommendations_MissingUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{}
	svc := &Service{cfg: cfg}
	handler := NewHandler(svc)

	// Create request without user_id
	req, _ := http.NewRequest("GET", "/recommendations", nil)
	w := httptest.NewRecorder()

	router := gin.New()
	router.GET("/recommendations", handler.HandleGetRecommendations)
	router.ServeHTTP(w, req)

	// Should return 400
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{}
	svc := &Service{cfg: cfg}
	handler := NewHandler(svc)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router := gin.New()
	router.GET("/health", handler.HandleHealth)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
