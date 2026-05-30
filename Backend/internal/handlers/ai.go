package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"autorent-backend/internal/ai"
	"autorent-backend/internal/models"
	"autorent-backend/internal/recommendation"

	"github.com/gin-gonic/gin"
)

type RecommendationCarStore interface {
	SearchRecommendations(ctx context.Context, filters models.CarRecommendationFilters) ([]models.Car, error)
}

type AIHandler struct {
	store     RecommendationCarStore
	extractor ai.CarFilterExtractor
}

func NewAIHandler(store RecommendationCarStore, extractor ai.CarFilterExtractor) *AIHandler {
	return &AIHandler{
		store:     store,
		extractor: extractor,
	}
}

func RegisterAIRoutes(router gin.IRouter, store RecommendationCarStore, extractor ai.CarFilterExtractor) {
	handler := NewAIHandler(store, extractor)

	aiRoutes := router.Group("/ai")
	aiRoutes.POST("/car-recommendation", handler.RecommendCars)
}

func (h *AIHandler) RecommendCars(c *gin.Context) {
	var input models.CarRecommendationRequest
	if err := c.ShouldBindJSON(&input); err != nil || strings.TrimSpace(input.Message) == "" {
		respondError(c, http.StatusBadRequest, "Message is required.")
		return
	}

	filters, err := h.extractor.ExtractCarFilters(c.Request.Context(), strings.TrimSpace(input.Message))
	if err != nil {
		if errors.Is(err, ai.ErrUnavailable) {
			log.Printf("AI extract failed: %v", err)
			respondError(c, http.StatusServiceUnavailable, "AI car assistant is temporarily unavailable. Please try again later.")
			return
		}

		log.Printf("AI extract unexpected error: %v", err)
		respondError(c, http.StatusInternalServerError, "failed to extract car filters")
		return
	}

	filters = recommendation.NormalizeFilters(filters)
	cars, err := h.store.SearchRecommendations(c.Request.Context(), filters)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to search cars")
		return
	}

	totalMatches := len(cars)
	topCars := recommendation.TopCars(cars, filters)
	responseCars := make([]models.RecommendedCar, 0, len(topCars))
	for _, car := range topCars {
		responseCars = append(responseCars, models.ToRecommendedCar(car))
	}

	c.JSON(http.StatusOK, models.CarRecommendationResponse{
		Answer:       recommendation.BuildAnswer(input.Message, totalMatches, len(responseCars), filters.SortBy),
		Cars:         responseCars,
		TotalMatches: totalMatches,
	})
}
