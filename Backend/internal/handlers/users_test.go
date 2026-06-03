package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"autorent-backend/internal/models"
	"autorent-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type fakeAdminUserStore struct {
	listFunc func(ctx context.Context) ([]models.User, error)
	rateFunc func(ctx context.Context, id int64, rating float64) (*models.User, error)
}

func (f *fakeAdminUserStore) ListCustomers(ctx context.Context) ([]models.User, error) {
	return f.listFunc(ctx)
}

func (f *fakeAdminUserStore) RateCustomer(ctx context.Context, id int64, rating float64) (*models.User, error) {
	return f.rateFunc(ctx, id, rating)
}

func TestAdminListCustomers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeAdminUserStore()
	store.listFunc = func(context.Context) ([]models.User, error) {
		return []models.User{*adminUserResponse(5, 4.5)}, nil
	}

	router := gin.New()
	RegisterAdminUserRoutes(router.Group("/api/admin"), store)

	req, err := http.NewRequest(http.MethodGet, "/api/admin/users", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "client@example.com") {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestAdminListCustomersStoreError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeAdminUserStore()
	store.listFunc = func(context.Context) ([]models.User, error) {
		return nil, errors.New("database failed")
	}

	router := gin.New()
	RegisterAdminUserRoutes(router.Group("/api/admin"), store)

	req, err := http.NewRequest(http.MethodGet, "/api/admin/users", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
}

func TestAdminRateCustomer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var capturedID int64
	var capturedRating float64
	store := newFakeAdminUserStore()
	store.rateFunc = func(_ context.Context, id int64, rating float64) (*models.User, error) {
		capturedID = id
		capturedRating = rating
		return adminUserResponse(id, 4.8), nil
	}

	router := gin.New()
	RegisterAdminUserRoutes(router.Group("/api/admin"), store)

	req, err := http.NewRequest(http.MethodPatch, "/api/admin/users/8/rating", strings.NewReader(`{"rating":4.5}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if capturedID != 8 || capturedRating != 4.5 {
		t.Fatalf("unexpected rating call: id=%d rating=%f", capturedID, capturedRating)
	}
}

func TestAdminRateCustomerRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterAdminUserRoutes(router.Group("/api/admin"), newFakeAdminUserStore())

	req, err := http.NewRequest(http.MethodPatch, "/api/admin/users/not-an-id/rating", strings.NewReader(`{"rating":4}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestAdminRateCustomerRejectsInvalidPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterAdminUserRoutes(router.Group("/api/admin"), newFakeAdminUserStore())

	req, err := http.NewRequest(http.MethodPatch, "/api/admin/users/8/rating", strings.NewReader(`{"rating":6}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestAdminRateCustomerReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeAdminUserStore()
	store.rateFunc = func(context.Context, int64, float64) (*models.User, error) {
		return nil, repository.ErrNotFound
	}

	router := gin.New()
	RegisterAdminUserRoutes(router.Group("/api/admin"), store)

	req, err := http.NewRequest(http.MethodPatch, "/api/admin/users/8/rating", strings.NewReader(`{"rating":4}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func adminUserResponse(id int64, rating float64) *models.User {
	now := time.Now()
	return &models.User{
		ID:          id,
		FirstName:   "Client",
		LastName:    "One",
		Name:        "Client One",
		Email:       "client@example.com",
		Rating:      rating,
		RatingCount: 3,
		Role:        models.UserRoleUser,
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func newFakeAdminUserStore() *fakeAdminUserStore {
	return &fakeAdminUserStore{
		listFunc: func(context.Context) ([]models.User, error) {
			return []models.User{}, nil
		},
		rateFunc: func(context.Context, int64, float64) (*models.User, error) {
			return adminUserResponse(1, 5), nil
		},
	}
}
