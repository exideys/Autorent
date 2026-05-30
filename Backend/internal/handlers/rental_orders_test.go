package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"autorent-backend/internal/auth"
	"autorent-backend/internal/models"
	"autorent-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type fakeRentalOrderHandlerStore struct {
	createFunc func(ctx context.Context, userID int64, input models.RentalOrderInput) (*models.RentalOrder, error)
	listFunc   func(ctx context.Context, userID int64) ([]models.RentalOrder, error)
}

func (f fakeRentalOrderHandlerStore) Create(ctx context.Context, userID int64, input models.RentalOrderInput) (*models.RentalOrder, error) {
	return f.createFunc(ctx, userID, input)
}

func (f fakeRentalOrderHandlerStore) ListByUserID(ctx context.Context, userID int64) ([]models.RentalOrder, error) {
	return f.listFunc(ctx, userID)
}

func TestCreateRentalOrderRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterRentalOrderRoutes(router.Group("/api"), newFakeRentalOrderHandlerStore(), auth.NewTokenManager("secret", time.Hour))

	req, err := http.NewRequest(http.MethodPost, "/api/rental-orders", strings.NewReader(validRentalOrderPayload()))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestCreateRentalOrderUsesAuthenticatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenManager := auth.NewTokenManager("secret", time.Hour)
	token := rentalTestToken(t, tokenManager, 42, models.UserRoleUser)
	store := newFakeRentalOrderHandlerStore()
	store.createFunc = func(_ context.Context, userID int64, input models.RentalOrderInput) (*models.RentalOrder, error) {
		if userID != 42 || input.CarID != 7 || input.PickupLocation != "Airport" {
			t.Fatalf("unexpected create input: userID=%d input=%+v", userID, input)
		}
		return rentalOrderTestResponse(userID, input.CarID), nil
	}

	router := gin.New()
	RegisterRentalOrderRoutes(router.Group("/api"), store, tokenManager)

	req, err := http.NewRequest(http.MethodPost, "/api/rental-orders", strings.NewReader(validRentalOrderPayload()))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, recorder.Code, recorder.Body.String())
	}

	var response struct {
		Data models.RentalOrder `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.Data.UserID != 42 || response.Data.CarID != 7 {
		t.Fatalf("unexpected response: %+v", response.Data)
	}
}

func TestCreateRentalOrderReturnsConflictForUnavailableCar(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenManager := auth.NewTokenManager("secret", time.Hour)
	token := rentalTestToken(t, tokenManager, 42, models.UserRoleUser)
	store := newFakeRentalOrderHandlerStore()
	store.createFunc = func(context.Context, int64, models.RentalOrderInput) (*models.RentalOrder, error) {
		return nil, repository.ErrUnavailable
	}

	router := gin.New()
	RegisterRentalOrderRoutes(router.Group("/api"), store, tokenManager)

	req, err := http.NewRequest(http.MethodPost, "/api/rental-orders", strings.NewReader(validRentalOrderPayload()))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, recorder.Code)
	}
}

func TestListMyRentalOrdersUsesAuthenticatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenManager := auth.NewTokenManager("secret", time.Hour)
	token := rentalTestToken(t, tokenManager, 55, models.UserRoleUser)
	store := newFakeRentalOrderHandlerStore()
	store.listFunc = func(_ context.Context, userID int64) ([]models.RentalOrder, error) {
		if userID != 55 {
			t.Fatalf("expected user id 55, got %d", userID)
		}
		return []models.RentalOrder{*rentalOrderTestResponse(userID, 8)}, nil
	}

	router := gin.New()
	RegisterRentalOrderRoutes(router.Group("/api"), store, tokenManager)

	req, err := http.NewRequest(http.MethodGet, "/api/rental-orders", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
}

func TestAdminUserRentalOrdersRequireAdminRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenManager := auth.NewTokenManager("secret", time.Hour)
	token := rentalTestToken(t, tokenManager, 1, models.UserRoleUser)

	router := gin.New()
	admin := router.Group("/api/admin")
	admin.Use(RequireAdmin(tokenManager))
	RegisterAdminRentalOrderRoutes(admin, newFakeRentalOrderHandlerStore())

	req, err := http.NewRequest(http.MethodGet, "/api/admin/users/42/rental-orders", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, recorder.Code)
	}
}

func validRentalOrderPayload() string {
	return `{
		"car_id":7,
		"start_date":"2026-06-01",
		"end_date":"2026-06-03",
		"pickup_location":"Airport",
		"pickup_time":"09:30",
		"phone":"+1 555 0100",
		"notes":"Child seat"
	}`
}

func newFakeRentalOrderHandlerStore() *fakeRentalOrderHandlerStore {
	return &fakeRentalOrderHandlerStore{
		createFunc: func(context.Context, int64, models.RentalOrderInput) (*models.RentalOrder, error) {
			return &models.RentalOrder{}, nil
		},
		listFunc: func(context.Context, int64) ([]models.RentalOrder, error) {
			return []models.RentalOrder{}, nil
		},
	}
}

func rentalTestToken(t *testing.T, tokenManager *auth.TokenManager, userID int64, role string) string {
	t.Helper()

	token, err := tokenManager.Generate(models.User{
		ID:    userID,
		Email: "rental@example.com",
		Role:  role,
	})
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	return token
}

func rentalOrderTestResponse(userID int64, carID int64) *models.RentalOrder {
	return &models.RentalOrder{
		ID:             99,
		UserID:         userID,
		CarID:          carID,
		StartDate:      "2026-06-01",
		EndDate:        "2026-06-03",
		PickupLocation: "Airport",
		PickupTime:     "09:30",
		Phone:          "+1 555 0100",
		TotalPrice:     500,
		Deposit:        1000,
		Status:         models.RentalOrderStatusActive,
		Car: models.RentalOrderCarSummary{
			ID:          carID,
			Brand:       "BMW",
			Model:       "M5",
			Year:        2024,
			PricePerDay: 250,
			Deposit:     1000,
			Status:      "rented",
		},
	}
}
