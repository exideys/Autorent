package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"autorent-backend/internal/models"
	aiservice "autorent-backend/internal/services/ai"

	"github.com/gin-gonic/gin"
)

type fakeCarRecommendationService struct {
	response        *models.CarRecommendationResponse
	err             error
	receivedMessage string
}

func (f *fakeCarRecommendationService) RecommendCar(ctx context.Context, message string) (*models.CarRecommendationResponse, error) {
	f.receivedMessage = message
	return f.response, f.err
}

func setupAIHandlerRouter(service *fakeCarRecommendationService) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := NewAIHandler(service)

	router.POST("/api/ai/car-recommendation", handler.RecommendCar)

	return router
}

func TestAIHandlerRecommendCarSuccess(t *testing.T) {
	service := &fakeCarRecommendationService{
		response: &models.CarRecommendationResponse{
			Answer: "Recommended cars found.",
			Cars: []models.Car{
				{
					ID:          1,
					Brand:       "Mercedes-Benz",
					Model:       "S-Class",
					Seats:       5,
					PricePerDay: 220,
					Status:      "available",
				},
			},
		},
	}

	router := setupAIHandlerRouter(service)

	body := []byte(`{"message":"family car up to 240 dollars per day"}`)
	req, err := http.NewRequest(http.MethodPost, "/api/ai/car-recommendation", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}

	if service.receivedMessage != "family car up to 240 dollars per day" {
		t.Fatalf("expected service to receive trimmed message, got %q", service.receivedMessage)
	}

	var response models.CarRecommendationResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Answer != "Recommended cars found." {
		t.Fatalf("unexpected answer: %s", response.Answer)
	}

	if len(response.Cars) != 1 {
		t.Fatalf("expected 1 car, got %d", len(response.Cars))
	}
}

func TestAIHandlerRecommendCarEmptyMessage(t *testing.T) {
	service := &fakeCarRecommendationService{}
	router := setupAIHandlerRouter(service)

	body := []byte(`{"message":"   "}`)
	req, err := http.NewRequest(http.MethodPost, "/api/ai/car-recommendation", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusBadRequest, recorder.Code, recorder.Body.String())
	}

	var response models.ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error != "Message is required." {
		t.Fatalf("unexpected error: %s", response.Error)
	}
}

func TestAIHandlerRecommendCarAIUnavailable(t *testing.T) {
	service := &fakeCarRecommendationService{
		err: aiservice.ErrAIUnavailable,
	}

	router := setupAIHandlerRouter(service)

	body := []byte(`{"message":"family car up to 240 dollars per day"}`)
	req, err := http.NewRequest(http.MethodPost, "/api/ai/car-recommendation", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusServiceUnavailable, recorder.Code, recorder.Body.String())
	}

	var response models.ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expectedError := "AI car assistant is temporarily unavailable. Please try again later."
	if response.Error != expectedError {
		t.Fatalf("expected error %q, got %q", expectedError, response.Error)
	}
}

func TestAIHandlerRecommendCarUnexpectedError(t *testing.T) {
	service := &fakeCarRecommendationService{
		err: errors.New("unexpected failure"),
	}

	router := setupAIHandlerRouter(service)

	body := []byte(`{"message":"family car up to 240 dollars per day"}`)
	req, err := http.NewRequest(http.MethodPost, "/api/ai/car-recommendation", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusInternalServerError, recorder.Code, recorder.Body.String())
	}
}
