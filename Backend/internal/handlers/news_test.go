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

	"autorent-backend/internal/models"
	"autorent-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type fakeNewsStore struct {
	listFunc   func(ctx context.Context, filters models.NewsFilters) ([]models.NewsArticle, error)
	getFunc    func(ctx context.Context, id int64) (*models.NewsArticle, error)
	createFunc func(ctx context.Context, input models.NewsInput) (*models.NewsArticle, error)
	updateFunc func(ctx context.Context, id int64, input models.NewsInput) (*models.NewsArticle, error)
	deleteFunc func(ctx context.Context, id int64) error
}

func (f *fakeNewsStore) List(ctx context.Context, filters models.NewsFilters) ([]models.NewsArticle, error) {
	return f.listFunc(ctx, filters)
}

func (f *fakeNewsStore) GetByID(ctx context.Context, id int64) (*models.NewsArticle, error) {
	return f.getFunc(ctx, id)
}

func (f *fakeNewsStore) Create(ctx context.Context, input models.NewsInput) (*models.NewsArticle, error) {
	return f.createFunc(ctx, input)
}

func (f *fakeNewsStore) Update(ctx context.Context, id int64, input models.NewsInput) (*models.NewsArticle, error) {
	return f.updateFunc(ctx, id, input)
}

func (f *fakeNewsStore) Delete(ctx context.Context, id int64) error {
	return f.deleteFunc(ctx, id)
}

func TestListPublishedNewsAppliesPublicFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var capturedFilters models.NewsFilters
	store := newFakeNewsStore()
	store.listFunc = func(_ context.Context, filters models.NewsFilters) ([]models.NewsArticle, error) {
		capturedFilters = filters
		return []models.NewsArticle{*newsArticleResponse(1, models.NewsStatusPublished)}, nil
	}

	router := gin.New()
	RegisterNewsRoutes(router.Group("/api"), store)

	req, err := http.NewRequest(http.MethodGet, "/api/news?sort=title&order=asc", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if capturedFilters.Status != models.NewsStatusPublished || capturedFilters.SortBy != "title" || capturedFilters.SortOrder != "asc" {
		t.Fatalf("unexpected filters: %+v", capturedFilters)
	}

	var response struct {
		Data []models.NewsArticle `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(response.Data) != 1 || response.Data[0].ID != 1 {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestListPublishedNewsStoreError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeNewsStore()
	store.listFunc = func(context.Context, models.NewsFilters) ([]models.NewsArticle, error) {
		return nil, errors.New("database failed")
	}

	router := gin.New()
	RegisterNewsRoutes(router.Group("/api"), store)

	req, err := http.NewRequest(http.MethodGet, "/api/news", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
}

func TestGetPublishedNewsHidesDrafts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeNewsStore()
	store.getFunc = func(_ context.Context, id int64) (*models.NewsArticle, error) {
		if id != 2 {
			t.Fatalf("expected id 2, got %d", id)
		}
		return newsArticleResponse(id, models.NewsStatusDraft), nil
	}

	router := gin.New()
	RegisterNewsRoutes(router.Group("/api"), store)

	req, err := http.NewRequest(http.MethodGet, "/api/news/2", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestGetNewsRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterNewsRoutes(router.Group("/api"), newFakeNewsStore())

	req, err := http.NewRequest(http.MethodGet, "/api/news/not-an-id", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestListAdminNewsUsesAdminDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var capturedFilters models.NewsFilters
	store := newFakeNewsStore()
	store.listFunc = func(_ context.Context, filters models.NewsFilters) ([]models.NewsArticle, error) {
		capturedFilters = filters
		return []models.NewsArticle{*newsArticleResponse(3, models.NewsStatusDraft)}, nil
	}

	router := gin.New()
	RegisterAdminNewsRoutes(router.Group("/api/admin"), store)

	req, err := http.NewRequest(http.MethodGet, "/api/admin/news", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if capturedFilters.Status != "" || capturedFilters.SortBy != "created_at" || capturedFilters.SortOrder != "desc" {
		t.Fatalf("unexpected filters: %+v", capturedFilters)
	}
}

func TestGetAdminNews(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeNewsStore()
	store.getFunc = func(_ context.Context, id int64) (*models.NewsArticle, error) {
		if id != 5 {
			t.Fatalf("expected id 5, got %d", id)
		}
		return newsArticleResponse(id, models.NewsStatusDraft), nil
	}

	router := gin.New()
	RegisterAdminNewsRoutes(router.Group("/api/admin"), store)

	req, err := http.NewRequest(http.MethodGet, "/api/admin/news/5", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
}

func TestAdminCreateNews(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeNewsStore()
	store.createFunc = func(_ context.Context, input models.NewsInput) (*models.NewsArticle, error) {
		if input.Title != "Launch" || input.Status != models.NewsStatusPublished {
			t.Fatalf("unexpected input: %+v", input)
		}
		if input.ImageURL == nil || *input.ImageURL != "https://example.com/news.jpg" {
			t.Fatalf("unexpected image url: %+v", input.ImageURL)
		}
		return newsArticleResponse(4, input.Status), nil
	}

	router := gin.New()
	RegisterAdminNewsRoutes(router.Group("/api/admin"), store)

	req, err := http.NewRequest(http.MethodPost, "/api/admin/news", strings.NewReader(validNewsPayload()))
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

func TestAdminCreateNewsRejectsInvalidPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterAdminNewsRoutes(router.Group("/api/admin"), newFakeNewsStore())

	req, err := http.NewRequest(http.MethodPost, "/api/admin/news", strings.NewReader(`{"title":"Launch"}`))
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

func TestAdminUpdateNewsReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeNewsStore()
	store.updateFunc = func(context.Context, int64, models.NewsInput) (*models.NewsArticle, error) {
		return nil, repository.ErrNotFound
	}

	router := gin.New()
	RegisterAdminNewsRoutes(router.Group("/api/admin"), store)

	req, err := http.NewRequest(http.MethodPut, "/api/admin/news/9", strings.NewReader(validNewsPayload()))
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

func TestAdminUpdateNews(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeNewsStore()
	store.updateFunc = func(_ context.Context, id int64, input models.NewsInput) (*models.NewsArticle, error) {
		if id != 9 || input.Title != "Launch" {
			t.Fatalf("unexpected update input: id=%d input=%+v", id, input)
		}
		return newsArticleResponse(id, input.Status), nil
	}

	router := gin.New()
	RegisterAdminNewsRoutes(router.Group("/api/admin"), store)

	req, err := http.NewRequest(http.MethodPut, "/api/admin/news/9", strings.NewReader(validNewsPayload()))
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

func TestAdminDeleteNews(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var deletedID int64
	store := newFakeNewsStore()
	store.deleteFunc = func(_ context.Context, id int64) error {
		deletedID = id
		return nil
	}

	router := gin.New()
	RegisterAdminNewsRoutes(router.Group("/api/admin"), store)

	req, err := http.NewRequest(http.MethodDelete, "/api/admin/news/7", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
	if deletedID != 7 {
		t.Fatalf("expected deleted id 7, got %d", deletedID)
	}
}

func TestAdminDeleteNewsReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeNewsStore()
	store.deleteFunc = func(context.Context, int64) error {
		return repository.ErrNotFound
	}

	router := gin.New()
	RegisterAdminNewsRoutes(router.Group("/api/admin"), store)

	req, err := http.NewRequest(http.MethodDelete, "/api/admin/news/7", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func validNewsPayload() string {
	return `{
		"title":"Launch",
		"summary":"Short summary",
		"content":"Full content",
		"image_url":"https://example.com/news.jpg",
		"status":"published"
	}`
}

func newsArticleResponse(id int64, status string) *models.NewsArticle {
	now := time.Now()
	article := &models.NewsArticle{
		ID:        id,
		Title:     "Launch",
		Summary:   "Short summary",
		Content:   "Full content",
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if status == models.NewsStatusPublished {
		article.PublishedAt = &now
	}
	return article
}

func newFakeNewsStore() *fakeNewsStore {
	return &fakeNewsStore{
		listFunc: func(context.Context, models.NewsFilters) ([]models.NewsArticle, error) {
			return []models.NewsArticle{}, nil
		},
		getFunc: func(context.Context, int64) (*models.NewsArticle, error) {
			return nil, repository.ErrNotFound
		},
		createFunc: func(context.Context, models.NewsInput) (*models.NewsArticle, error) {
			return newsArticleResponse(1, models.NewsStatusDraft), nil
		},
		updateFunc: func(context.Context, int64, models.NewsInput) (*models.NewsArticle, error) {
			return newsArticleResponse(1, models.NewsStatusDraft), nil
		},
		deleteFunc: func(context.Context, int64) error {
			return nil
		},
	}
}
