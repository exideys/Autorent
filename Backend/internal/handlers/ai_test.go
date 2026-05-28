package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"autorent-backend/internal/ai"
	"autorent-backend/internal/models"

	"github.com/gin-gonic/gin"
)

type fakeCarFilterExtractor struct {
	filters models.CarRecommendationFilters
	err     error
}

func (f fakeCarFilterExtractor) ExtractCarFilters(context.Context, string) (models.CarRecommendationFilters, error) {
	return f.filters, f.err
}

type fakeRecommendationCarStore struct {
	searchFunc func(ctx context.Context, filters models.CarRecommendationFilters) ([]models.Car, error)
}

func (f fakeRecommendationCarStore) SearchRecommendations(ctx context.Context, filters models.CarRecommendationFilters) ([]models.Car, error) {
	return f.searchFunc(ctx, filters)
}

func TestRecommendCarsRejectsEmptyMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterAIRoutes(router.Group("/api"), fakeRecommendationCarStore{}, fakeCarFilterExtractor{})

	req, err := http.NewRequest(http.MethodPost, "/api/ai/car-recommendation", strings.NewReader(`{"message":"   "}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "Message is required.") {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestRecommendCarsReturnsUnavailableWhenExtractorFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterAIRoutes(
		router.Group("/api"),
		fakeRecommendationCarStore{},
		fakeCarFilterExtractor{err: ai.ErrUnavailable},
	)

	req, err := http.NewRequest(http.MethodPost, "/api/ai/car-recommendation", strings.NewReader(`{"message":"Tesla under 220 dollars"}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "AI car assistant is temporarily unavailable") {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestRecommendCarsReturnsCars(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var capturedFilters models.CarRecommendationFilters
	horsepower := 450
	color := "White"
	store := fakeRecommendationCarStore{
		searchFunc: func(_ context.Context, filters models.CarRecommendationFilters) ([]models.Car, error) {
			capturedFilters = filters
			return []models.Car{
				recommendationTestCar(1, "Tesla", "Model 3", 2021, "Sedan", "Electric", 180, 5, &horsepower, &color),
				recommendationTestCar(2, "Tesla", "Model Y", 2023, "SUV", "Electric", 210, 5, &horsepower, &color),
				recommendationTestCar(3, "Tesla", "Model S", 2022, "Liftback", "Electric", 220, 5, &horsepower, &color),
				recommendationTestCar(4, "Tesla", "Model X", 2020, "SUV", "Electric", 215, 7, &horsepower, &color),
				recommendationTestCar(5, "Tesla", "Model 3 Long Range", 2024, "Sedan", "Electric", 205, 5, &horsepower, &color),
				recommendationTestCar(6, "Tesla", "Model Y Long Range", 2024, "SUV", "Electric", 219, 5, &horsepower, &color),
			}, nil
		},
	}

	router := gin.New()
	RegisterAIRoutes(
		router.Group("/api"),
		store,
		fakeCarFilterExtractor{filters: models.CarRecommendationFilters{
			MaxPricePerDay: 220,
			PreferredBrand: "Tesla",
			SortBy:         "price_asc",
			OnlyAvailable:  true,
		}},
	)

	req, err := http.NewRequest(http.MethodPost, "/api/ai/car-recommendation", strings.NewReader(`{"message":"хочу орендувати Tesla до 220 доларів на день"}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if capturedFilters.PreferredBrand != "Tesla" || capturedFilters.MaxPricePerDay != 220 || !capturedFilters.OnlyAvailable {
		t.Fatalf("unexpected filters: %+v", capturedFilters)
	}

	var response models.CarRecommendationResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.TotalMatches != 6 {
		t.Fatalf("expected 6 total matches, got %d", response.TotalMatches)
	}
	if len(response.Cars) != 5 {
		t.Fatalf("expected 5 cars, got %d", len(response.Cars))
	}
	if response.Cars[0].PricePerDay != 180 || response.Cars[0].Brand != "Tesla" {
		t.Fatalf("unexpected first car: %+v", response.Cars[0])
	}
	if response.Cars[0].Horsepower != horsepower || response.Cars[0].Color != color {
		t.Fatalf("expected nullable fields to be mapped, got %+v", response.Cars[0])
	}
	if response.Answer == "" || !strings.Contains(response.Answer, "найдешев") {
		t.Fatalf("unexpected answer: %q", response.Answer)
	}
}

func TestRecommendCarsReturnsStoreError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterAIRoutes(
		router.Group("/api"),
		fakeRecommendationCarStore{
			searchFunc: func(context.Context, models.CarRecommendationFilters) ([]models.Car, error) {
				return nil, errors.New("database failed")
			},
		},
		fakeCarFilterExtractor{filters: models.CarRecommendationFilters{OnlyAvailable: true}},
	)

	req, err := http.NewRequest(http.MethodPost, "/api/ai/car-recommendation", strings.NewReader(`{"message":"Tesla"}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
}

func recommendationTestCar(id int64, brand string, model string, year int, bodyType string, fuelType string, price float64, seats int, horsepower *int, color *string) models.Car {
	return models.Car{
		ID:           id,
		Brand:        brand,
		Model:        model,
		Year:         year,
		CarClass:     "Electric Comfort",
		BodyType:     bodyType,
		Transmission: "Automatic",
		FuelType:     fuelType,
		Seats:        seats,
		Doors:        4,
		Horsepower:   horsepower,
		PricePerDay:  price,
		Deposit:      500,
		Color:        color,
		Status:       "available",
	}
}
