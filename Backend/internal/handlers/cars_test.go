package handlers

import (
	"context"
	"encoding/json"
	"errors"
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

type fakeCarStore struct {
	listFunc        func(ctx context.Context, filters models.CarFilters) ([]models.Car, error)
	getFunc         func(ctx context.Context, id int64) (*models.Car, error)
	createFunc      func(ctx context.Context, input models.CarInput) (*models.Car, error)
	updateFunc      func(ctx context.Context, id int64, input models.CarInput) (*models.Car, error)
	deleteFunc      func(ctx context.Context, id int64) error
	addImageFunc    func(ctx context.Context, carID int64, input models.CarImageInput) (*models.CarImage, error)
	deleteImageFunc func(ctx context.Context, imageID int64) error
}

func (f *fakeCarStore) List(ctx context.Context, filters models.CarFilters) ([]models.Car, error) {
	return f.listFunc(ctx, filters)
}

func (f *fakeCarStore) GetByID(ctx context.Context, id int64) (*models.Car, error) {
	return f.getFunc(ctx, id)
}

func (f *fakeCarStore) Create(ctx context.Context, input models.CarInput) (*models.Car, error) {
	return f.createFunc(ctx, input)
}

func (f *fakeCarStore) Update(ctx context.Context, id int64, input models.CarInput) (*models.Car, error) {
	return f.updateFunc(ctx, id, input)
}

func (f *fakeCarStore) Delete(ctx context.Context, id int64) error {
	return f.deleteFunc(ctx, id)
}

func (f *fakeCarStore) AddImage(ctx context.Context, carID int64, input models.CarImageInput) (*models.CarImage, error) {
	return f.addImageFunc(ctx, carID, input)
}

func (f *fakeCarStore) DeleteImage(ctx context.Context, imageID int64) error {
	return f.deleteImageFunc(ctx, imageID)
}

func TestListCarsAppliesFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var capturedFilters models.CarFilters
	store := newFakeCarStore()
	store.listFunc = func(_ context.Context, filters models.CarFilters) ([]models.Car, error) {
		capturedFilters = filters
		return []models.Car{
			{
				ID:           1,
				Brand:        "BMW",
				Model:        "M5",
				Year:         2024,
				CarClass:     "Business",
				BodyType:     "Sedan",
				Transmission: "Automatic",
				FuelType:     "Petrol",
				Seats:        5,
				Doors:        4,
				PricePerDay:  250,
				Deposit:      1000,
				Status:       "available",
				Images:       []models.CarImage{},
			},
		}, nil
	}

	router := gin.New()
	RegisterCarRoutes(router.Group("/api"), store)

	req, err := http.NewRequest(http.MethodGet, "/api/cars?available=true&car_class=Business&body_type=Sedan&transmission=Automatic&fuel_type=Petrol&sort=price_per_day&order=asc", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	if capturedFilters.Status != "available" {
		t.Fatalf("expected available status filter, got %q", capturedFilters.Status)
	}
	if capturedFilters.CarClass != "Business" || capturedFilters.BodyType != "Sedan" {
		t.Fatalf("unexpected filters: %+v", capturedFilters)
	}
	if capturedFilters.SortBy != "price_per_day" || capturedFilters.SortOrder != "asc" {
		t.Fatalf("unexpected sorting: %+v", capturedFilters)
	}

	var response struct {
		Data []models.Car `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(response.Data) != 1 || response.Data[0].Brand != "BMW" {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestPublicCarsDoNotAllowCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterCarRoutes(router.Group("/api"), newFakeCarStore())

	req, err := http.NewRequest(http.MethodPost, "/api/cars", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestListCarsStoreError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeCarStore()
	store.listFunc = func(context.Context, models.CarFilters) ([]models.Car, error) {
		return nil, errors.New("database failed")
	}

	router := gin.New()
	RegisterCarRoutes(router.Group("/api"), store)

	req, err := http.NewRequest(http.MethodGet, "/api/cars", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
}

func TestAdminCarsRequireAdminRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokens := auth.NewTokenManager("test-secret", time.Hour)
	token, err := tokens.Generate(models.User{
		ID:    1,
		Email: "user@example.com",
		Role:  models.UserRoleUser,
	})
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	router := gin.New()
	admin := router.Group("/api/admin")
	admin.Use(RequireAdmin(tokens))
	RegisterAdminCarRoutes(admin, newFakeCarStore())

	req, err := http.NewRequest(http.MethodGet, "/api/admin/cars", nil)
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

func TestGetCarReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeCarStore()
	store.getFunc = func(_ context.Context, _ int64) (*models.Car, error) {
		return nil, repository.ErrNotFound
	}

	router := gin.New()
	RegisterCarRoutes(router.Group("/api"), store)

	req, err := http.NewRequest(http.MethodGet, "/api/cars/42", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestGetCarRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterCarRoutes(router.Group("/api"), newFakeCarStore())

	req, err := http.NewRequest(http.MethodGet, "/api/cars/not-an-id", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestAdminCreateCar(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeCarStore()
	store.createFunc = func(_ context.Context, input models.CarInput) (*models.Car, error) {
		if input.Brand != "BMW" || input.Images[0].ImageURL != "https://example.com/bmw.jpg" {
			t.Fatalf("unexpected input: %+v", input)
		}

		return &models.Car{
			ID:           1,
			Brand:        input.Brand,
			Model:        input.Model,
			Year:         input.Year,
			CarClass:     input.CarClass,
			BodyType:     input.BodyType,
			Transmission: input.Transmission,
			FuelType:     input.FuelType,
			Seats:        input.Seats,
			Doors:        input.Doors,
			PricePerDay:  input.PricePerDay,
			Deposit:      input.Deposit,
			Status:       "available",
			Images:       []models.CarImage{},
		}, nil
	}

	router := gin.New()
	RegisterAdminCarRoutes(router.Group("/api/admin"), store)

	req, err := http.NewRequest(http.MethodPost, "/api/admin/cars", strings.NewReader(validCarPayload()))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, recorder.Code, recorder.Body.String())
	}
}

func TestAdminCreateCarRejectsInvalidPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterAdminCarRoutes(router.Group("/api/admin"), newFakeCarStore())

	req, err := http.NewRequest(http.MethodPost, "/api/admin/cars", strings.NewReader(`{"brand":"BMW"}`))
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

func TestAdminUpdateCar(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeCarStore()
	store.updateFunc = func(_ context.Context, id int64, input models.CarInput) (*models.Car, error) {
		if id != 7 {
			t.Fatalf("expected id 7, got %d", id)
		}

		return &models.Car{
			ID:           id,
			Brand:        input.Brand,
			Model:        input.Model,
			Year:         input.Year,
			CarClass:     input.CarClass,
			BodyType:     input.BodyType,
			Transmission: input.Transmission,
			FuelType:     input.FuelType,
			Seats:        input.Seats,
			Doors:        input.Doors,
			PricePerDay:  input.PricePerDay,
			Deposit:      input.Deposit,
			Status:       "available",
			Images:       []models.CarImage{},
		}, nil
	}

	router := gin.New()
	RegisterAdminCarRoutes(router.Group("/api/admin"), store)

	req, err := http.NewRequest(http.MethodPut, "/api/admin/cars/7", strings.NewReader(validCarPayload()))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
}

func TestAdminUpdateCarRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterAdminCarRoutes(router.Group("/api/admin"), newFakeCarStore())

	req, err := http.NewRequest(http.MethodPut, "/api/admin/cars/not-an-id", strings.NewReader(validCarPayload()))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestAdminUpdateCarReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeCarStore()
	store.updateFunc = func(context.Context, int64, models.CarInput) (*models.Car, error) {
		return nil, repository.ErrNotFound
	}

	router := gin.New()
	RegisterAdminCarRoutes(router.Group("/api/admin"), store)

	req, err := http.NewRequest(http.MethodPut, "/api/admin/cars/7", strings.NewReader(validCarPayload()))
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

func TestAdminDeleteCarAndImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var deletedCarID int64
	var deletedImageID int64
	store := newFakeCarStore()
	store.deleteFunc = func(_ context.Context, id int64) error {
		deletedCarID = id
		return nil
	}
	store.deleteImageFunc = func(_ context.Context, id int64) error {
		deletedImageID = id
		return nil
	}

	router := gin.New()
	RegisterAdminCarRoutes(router.Group("/api/admin"), store)

	req, err := http.NewRequest(http.MethodDelete, "/api/admin/cars/3", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected car delete status %d, got %d", http.StatusNoContent, recorder.Code)
	}

	req, err = http.NewRequest(http.MethodDelete, "/api/admin/car-images/9", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected image delete status %d, got %d", http.StatusNoContent, recorder.Code)
	}

	if deletedCarID != 3 || deletedImageID != 9 {
		t.Fatalf("unexpected deleted ids: car=%d image=%d", deletedCarID, deletedImageID)
	}
}

func TestAdminDeleteCarReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeCarStore()
	store.deleteFunc = func(context.Context, int64) error {
		return repository.ErrNotFound
	}

	router := gin.New()
	RegisterAdminCarRoutes(router.Group("/api/admin"), store)

	req, err := http.NewRequest(http.MethodDelete, "/api/admin/cars/3", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestAdminAddCarImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeCarStore()
	store.addImageFunc = func(_ context.Context, carID int64, input models.CarImageInput) (*models.CarImage, error) {
		if carID != 5 || input.ImageURL != "https://example.com/new.jpg" {
			t.Fatalf("unexpected input: carID=%d input=%+v", carID, input)
		}

		return &models.CarImage{
			ID:       20,
			CarID:    carID,
			ImageURL: input.ImageURL,
			IsMain:   input.IsMain,
		}, nil
	}

	router := gin.New()
	RegisterAdminCarRoutes(router.Group("/api/admin"), store)

	req, err := http.NewRequest(http.MethodPost, "/api/admin/cars/5/images", strings.NewReader(`{
		"image_url":"https://example.com/new.jpg",
		"is_main":true
	}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, recorder.Code, recorder.Body.String())
	}
}

func TestAdminAddCarImageReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeCarStore()
	store.addImageFunc = func(context.Context, int64, models.CarImageInput) (*models.CarImage, error) {
		return nil, repository.ErrNotFound
	}

	router := gin.New()
	RegisterAdminCarRoutes(router.Group("/api/admin"), store)

	req, err := http.NewRequest(http.MethodPost, "/api/admin/cars/5/images", strings.NewReader(`{
		"image_url":"https://example.com/new.jpg"
	}`))
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

func TestAdminDeleteCarImageReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeCarStore()
	store.deleteImageFunc = func(context.Context, int64) error {
		return repository.ErrNotFound
	}

	router := gin.New()
	RegisterAdminCarRoutes(router.Group("/api/admin"), store)

	req, err := http.NewRequest(http.MethodDelete, "/api/admin/car-images/9", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func validCarPayload() string {
	return `{
		"brand":"BMW",
		"model":"M5",
		"year":2024,
		"car_class":"Business",
		"body_type":"Sedan",
		"transmission":"Automatic",
		"fuel_type":"Petrol",
		"seats":5,
		"doors":4,
		"price_per_day":250,
		"deposit":1000,
		"images":[
			{
				"image_url":"https://example.com/bmw.jpg",
				"is_main":true
			}
		]
	}`
}

func newFakeCarStore() *fakeCarStore {
	return &fakeCarStore{
		listFunc: func(context.Context, models.CarFilters) ([]models.Car, error) {
			return []models.Car{}, nil
		},
		getFunc: func(context.Context, int64) (*models.Car, error) {
			return nil, repository.ErrNotFound
		},
		createFunc: func(context.Context, models.CarInput) (*models.Car, error) {
			return &models.Car{}, nil
		},
		updateFunc: func(context.Context, int64, models.CarInput) (*models.Car, error) {
			return &models.Car{}, nil
		},
		deleteFunc: func(context.Context, int64) error {
			return nil
		},
		addImageFunc: func(context.Context, int64, models.CarImageInput) (*models.CarImage, error) {
			return &models.CarImage{}, nil
		},
		deleteImageFunc: func(context.Context, int64) error {
			return nil
		},
	}
}
