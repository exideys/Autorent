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
	"golang.org/x/crypto/bcrypt"
)

type fakeUserStore struct {
	createFunc     func(ctx context.Context, input models.RegisterInput, passwordHash string, role string) (*models.User, error)
	getByIDFunc    func(ctx context.Context, id int64) (*models.User, error)
	getByEmailFunc func(ctx context.Context, email string) (*models.UserWithPassword, error)
}

func (f *fakeUserStore) Create(ctx context.Context, input models.RegisterInput, passwordHash string, role string) (*models.User, error) {
	return f.createFunc(ctx, input, passwordHash, role)
}

func (f *fakeUserStore) GetByID(ctx context.Context, id int64) (*models.User, error) {
	return f.getByIDFunc(ctx, id)
}

func (f *fakeUserStore) GetByEmail(ctx context.Context, email string) (*models.UserWithPassword, error) {
	return f.getByEmailFunc(ctx, email)
}

func TestRegisterUserCreatesUserAndToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeUserStore()
	store.createFunc = func(_ context.Context, input models.RegisterInput, passwordHash string, role string) (*models.User, error) {
		if role != models.UserRoleUser {
			t.Fatalf("expected user role, got %q", role)
		}
		if input.Email != "user@example.com" {
			t.Fatalf("unexpected email %q", input.Email)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("password123")); err != nil {
			t.Fatalf("password was not hashed correctly: %v", err)
		}

		return &models.User{
			ID:    10,
			Name:  input.Name,
			Email: input.Email,
			Role:  role,
		}, nil
	}

	router := gin.New()
	RegisterAuthRoutes(router.Group("/api/auth"), store, auth.NewTokenManager("secret", time.Hour), "setup-token")

	req, err := http.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{
		"name":"Test User",
		"email":"user@example.com",
		"password":"password123"
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

	response := decodeAuthResponse(t, recorder)
	if response.Token == "" {
		t.Fatal("expected token in response")
	}
	if response.User.Role != models.UserRoleUser {
		t.Fatalf("expected user role, got %q", response.User.Role)
	}
}

func TestRegisterAdminRequiresSetupToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterAuthRoutes(router.Group("/api/auth"), newFakeUserStore(), auth.NewTokenManager("secret", time.Hour), "setup-token")

	req, err := http.NewRequest(http.MethodPost, "/api/auth/admin/register", strings.NewReader(`{
		"name":"Admin",
		"email":"admin@example.com",
		"password":"password123"
	}`))
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

func TestRegisterAdminDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterAuthRoutes(router.Group("/api/auth"), newFakeUserStore(), auth.NewTokenManager("secret", time.Hour), "")

	req, err := http.NewRequest(http.MethodPost, "/api/auth/admin/register", strings.NewReader(`{
		"name":"Admin",
		"email":"admin@example.com",
		"password":"password123"
	}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Setup-Token", "setup-token")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, recorder.Code)
	}
}

func TestRegisterAdminCreatesAdminWithSetupToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeUserStore()
	store.createFunc = func(_ context.Context, input models.RegisterInput, _ string, role string) (*models.User, error) {
		if role != models.UserRoleAdmin {
			t.Fatalf("expected admin role, got %q", role)
		}

		return &models.User{
			ID:    1,
			Name:  input.Name,
			Email: input.Email,
			Role:  role,
		}, nil
	}

	router := gin.New()
	RegisterAuthRoutes(router.Group("/api/auth"), store, auth.NewTokenManager("secret", time.Hour), "setup-token")

	req, err := http.NewRequest(http.MethodPost, "/api/auth/admin/register", strings.NewReader(`{
		"name":"Admin",
		"email":"admin@example.com",
		"password":"password123"
	}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Setup-Token", "setup-token")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, recorder.Code, recorder.Body.String())
	}

	response := decodeAuthResponse(t, recorder)
	if response.User.Role != models.UserRoleAdmin {
		t.Fatalf("expected admin role, got %q", response.User.Role)
	}
}

func TestRegisterUserReturnsConflictOnDuplicateEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeUserStore()
	store.createFunc = func(context.Context, models.RegisterInput, string, string) (*models.User, error) {
		return nil, repository.ErrDuplicate
	}

	router := gin.New()
	RegisterAuthRoutes(router.Group("/api/auth"), store, auth.NewTokenManager("secret", time.Hour), "setup-token")

	req, err := http.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{
		"name":"Test User",
		"email":"user@example.com",
		"password":"password123"
	}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, recorder.Code)
	}
}

func TestLoginRejectsInvalidPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	store := newFakeUserStore()
	store.getByEmailFunc = func(_ context.Context, email string) (*models.UserWithPassword, error) {
		return &models.UserWithPassword{
			User: models.User{
				ID:    1,
				Email: email,
				Role:  models.UserRoleUser,
			},
			PasswordHash: string(passwordHash),
		}, nil
	}

	router := gin.New()
	RegisterAuthRoutes(router.Group("/api/auth"), store, auth.NewTokenManager("secret", time.Hour), "setup-token")

	req, err := http.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{
		"email":"user@example.com",
		"password":"wrong-password"
	}`))
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

func TestLoginRejectsUnknownEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterAuthRoutes(router.Group("/api/auth"), newFakeUserStore(), auth.NewTokenManager("secret", time.Hour), "setup-token")

	req, err := http.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{
		"email":"missing@example.com",
		"password":"password123"
	}`))
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

func TestLoginReturnsToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	store := newFakeUserStore()
	store.getByEmailFunc = func(_ context.Context, email string) (*models.UserWithPassword, error) {
		return &models.UserWithPassword{
			User: models.User{
				ID:    2,
				Name:  "Test User",
				Email: email,
				Role:  models.UserRoleUser,
			},
			PasswordHash: string(passwordHash),
		}, nil
	}

	router := gin.New()
	RegisterAuthRoutes(router.Group("/api/auth"), store, auth.NewTokenManager("secret", time.Hour), "setup-token")

	req, err := http.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{
		"email":"user@example.com",
		"password":"password123"
	}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}

	response := decodeAuthResponse(t, recorder)
	if response.Token == "" {
		t.Fatal("expected token in response")
	}
}

func TestMeRequiresBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterAuthRoutes(router.Group("/api/auth"), newFakeUserStore(), auth.NewTokenManager("secret", time.Hour), "setup-token")

	req, err := http.NewRequest(http.MethodGet, "/api/auth/me", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestMeReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenManager := auth.NewTokenManager("secret", time.Hour)
	token, err := tokenManager.Generate(models.User{
		ID:    12,
		Email: "user@example.com",
		Role:  models.UserRoleUser,
	})
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	router := gin.New()
	RegisterAuthRoutes(router.Group("/api/auth"), newFakeUserStore(), tokenManager, "setup-token")

	req, err := http.NewRequest(http.MethodGet, "/api/auth/me", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestMeReturnsCurrentUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenManager := auth.NewTokenManager("secret", time.Hour)
	token, err := tokenManager.Generate(models.User{
		ID:    12,
		Email: "user@example.com",
		Role:  models.UserRoleUser,
	})
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	store := newFakeUserStore()
	store.getByIDFunc = func(_ context.Context, id int64) (*models.User, error) {
		if id != 12 {
			t.Fatalf("expected id 12, got %d", id)
		}

		return &models.User{
			ID:    id,
			Name:  "Test User",
			Email: "user@example.com",
			Role:  models.UserRoleUser,
		}, nil
	}

	router := gin.New()
	RegisterAuthRoutes(router.Group("/api/auth"), store, tokenManager, "setup-token")

	req, err := http.NewRequest(http.MethodGet, "/api/auth/me", nil)
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

func TestRequireAdminAllowsAdminRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenManager := auth.NewTokenManager("secret", time.Hour)
	token, err := tokenManager.Generate(models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.UserRoleAdmin,
	})
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	router := gin.New()
	router.GET("/admin", RequireAdmin(tokenManager), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req, err := http.NewRequest(http.MethodGet, "/admin", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
}

func TestRequireAuthRejectsInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/profile", RequireAuth(auth.NewTokenManager("secret", time.Hour)), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req, err := http.NewRequest(http.MethodGet, "/profile", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer invalid-token")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func decodeAuthResponse(t *testing.T, recorder *httptest.ResponseRecorder) models.AuthResponse {
	t.Helper()

	var response struct {
		Data models.AuthResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	return response.Data
}

func newFakeUserStore() *fakeUserStore {
	return &fakeUserStore{
		createFunc: func(context.Context, models.RegisterInput, string, string) (*models.User, error) {
			return &models.User{}, nil
		},
		getByIDFunc: func(context.Context, int64) (*models.User, error) {
			return nil, repository.ErrNotFound
		},
		getByEmailFunc: func(context.Context, string) (*models.UserWithPassword, error) {
			return nil, repository.ErrNotFound
		},
	}
}
