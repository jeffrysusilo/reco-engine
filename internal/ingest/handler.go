package ingest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/reco-engine/internal/models"
)

// Handler handles HTTP requests for event ingestion
type Handler struct {
	service *Service
}

// NewHandler creates a new handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// HandleIngestEvent handles POST /events
func (h *Handler) HandleIngestEvent(c *gin.Context) {
	var event models.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.service.IngestEvent(c.Request.Context(), &event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// HandleHealth handles GET /health
func (h *Handler) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
