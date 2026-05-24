package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"autorent-backend/internal/models"
	aiservice "autorent-backend/internal/services/ai"

	"github.com/gin-gonic/gin"
)

type CarRecommendationService interface {
	RecommendCar(ctx context.Context, message string) (*models.CarRecommendationResponse, error)
}

type AIHandler struct {
	service CarRecommendationService
}

func NewAIHandler(service CarRecommendationService) *AIHandler {
	return &AIHandler{
		service: service,
	}
}

func (h *AIHandler) RecommendCar(c *gin.Context) {
	var request models.CarRecommendationRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Message is required.",
		})
		return
	}

	request.Message = strings.TrimSpace(request.Message)
	if request.Message == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Message is required.",
		})
		return
	}

	/*response, err := h.service.RecommendCar(c.Request.Context(), request.Message)
	if err != nil {
		if errors.Is(err, aiservice.ErrAIUnavailable) {
			c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
				Error: "AI car assistant is temporarily unavailable. Please try again later.",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to get car recommendation.",
		})
		return
	}*/

	response, err := h.service.RecommendCar(c.Request.Context(), request.Message)
	if err != nil {
		log.Printf("AI recommendation failed: %v", err)

		if errors.Is(err, aiservice.ErrAIUnavailable) {
			c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
				Error: "AI car assistant is temporarily unavailable. Please try again later.",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to get car recommendation.",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
