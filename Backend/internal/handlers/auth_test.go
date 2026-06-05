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
	"golang.org/x/crypto/bcrypt"
)

type fakeUserStore struct {
	createFunc         func(ctx context.Context, input models.RegisterInput, passwordHash string, role string) (*models.User, error)
	createGoogleFunc   func(ctx context.Context, input models.GoogleUserInput) (*models.User, error)
	getByIDFunc        func(ctx context.Context, id int64) (*models.User, error)
	getByEmailFunc     func(ctx context.Context, email string) (*models.UserWithPassword, error)
	getByGoogleSubFunc func(ctx context.Context, googleSub string) (*models.User, error)
	linkGoogleSubFunc  func(ctx context.Context, id int64, googleSub string) (*models.User, error)
	updateProfileFunc  func(ctx context.Context, id int64, input models.UpdateCurrentUserInput) (*models.User, error)
}

func (f *fakeUserStore) Create(ctx context.Context, input models.RegisterInput, passwordHash string, role string) (*models.User, error) {
	return f.createFunc(ctx, input, passwordHash, role)
}

func (f *fakeUserStore) CreateGoogle(ctx context.Context, input models.GoogleUserInput) (*models.User, error) {
	return f.createGoogleFunc(ctx, input)
}

func (f *fakeUserStore) GetByID(ctx context.Context, id int64) (*models.User, error) {
	return f.getByIDFunc(ctx, id)
}

func (f *fakeUserStore) GetByEmail(ctx context.Context, email string) (*models.UserWithPassword, error) {
	return f.getByEmailFunc(ctx, email)
}

func (f *fakeUserStore) GetByGoogleSub(ctx context.Context, googleSub string) (*models.User, error) {
	return f.getByGoogleSubFunc(ctx, googleSub)
}

func (f *fakeUserStore) LinkGoogleSub(ctx context.Context, id int64, googleSub string) (*models.User, error) {
	return f.linkGoogleSubFunc(ctx, id, googleSub)
}

func (f *fakeUserStore) UpdateProfile(ctx context.Context, id int64, input models.UpdateCurrentUserInput) (*models.User, error) {
	return f.updateProfileFunc(ctx, id, input)
}

type fakeGoogleVerifier struct {
	identity *GoogleIdentity
	err      error
}

func (f fakeGoogleVerifier) Verify(context.Context, string, string) (*GoogleIdentity, error) {
	return f.identity, f.err
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
		if input.FirstName != "Test" || input.LastName != "User" {
			t.Fatalf("unexpected registration name parts: %+v", input)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("password123")); err != nil {
			t.Fatalf("password was not hashed correctly: %v", err)
		}

		return &models.User{
			ID:    10,
			Name:  input.DisplayName(),
			Email: input.Email,
			Role:  role,
		}, nil
	}

	router := gin.New()
	RegisterAuthRoutes(router.Group("/api/auth"), store, auth.NewTokenManager("secret", time.Hour), "setup-token", "")

	req, err := http.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{
		"name":"Test User",
		"first_name":"Test",
		"last_name":"User",
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
	RegisterAuthRoutes(router.Group("/api/auth"), newFakeUserStore(), auth.NewTokenManager("secret", time.Hour), "setup-token", "")

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
	RegisterAuthRoutes(router.Group("/api/auth"), newFakeUserStore(), auth.NewTokenManager("secret", time.Hour), "", "")

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
			Name:  input.DisplayName(),
			Email: input.Email,
			Role:  role,
		}, nil
	}

	router := gin.New()
	RegisterAuthRoutes(router.Group("/api/auth"), store, auth.NewTokenManager("secret", time.Hour), "setup-token", "")

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
	RegisterAuthRoutes(router.Group("/api/auth"), store, auth.NewTokenManager("secret", time.Hour), "setup-token", "")

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
	RegisterAuthRoutes(router.Group("/api/auth"), store, auth.NewTokenManager("secret", time.Hour), "setup-token", "")

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
	RegisterAuthRoutes(router.Group("/api/auth"), newFakeUserStore(), auth.NewTokenManager("secret", time.Hour), "setup-token", "")

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
	RegisterAuthRoutes(router.Group("/api/auth"), store, auth.NewTokenManager("secret", time.Hour), "setup-token", "")

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

func TestGoogleLoginReturnsUnavailableWhenNotConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterAuthRoutes(router.Group("/api/auth"), newFakeUserStore(), auth.NewTokenManager("secret", time.Hour), "setup-token", "")

	req, err := http.NewRequest(http.MethodPost, "/api/auth/google", strings.NewReader(`{"credential":"token"}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}
}

func TestGoogleLoginRejectsInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	store := newFakeUserStore()
	handler := NewAuthHandler(store, auth.NewTokenManager("secret", time.Hour), "setup-token", "google-client-id")
	handler.googleVerifier = fakeGoogleVerifier{err: errors.New("invalid token")}
	router.POST("/api/auth/google", handler.GoogleLogin)

	req, err := http.NewRequest(http.MethodPost, "/api/auth/google", strings.NewReader(`{"credential":"bad-token"}`))
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

func TestGoogleLoginCreatesNewUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeUserStore()
	store.createGoogleFunc = func(_ context.Context, input models.GoogleUserInput) (*models.User, error) {
		if input.GoogleSub != "google-sub" || input.Email != "new@example.com" {
			t.Fatalf("unexpected Google input: %+v", input)
		}
		if input.FirstName != "New" || input.LastName != "Client" {
			t.Fatalf("unexpected Google name: %+v", input)
		}
		return &models.User{
			ID:        22,
			FirstName: input.FirstName,
			LastName:  input.LastName,
			Name:      input.FirstName + " " + input.LastName,
			Email:     input.Email,
			Role:      models.UserRoleUser,
		}, nil
	}

	router := gin.New()
	handler := NewAuthHandler(store, auth.NewTokenManager("secret", time.Hour), "setup-token", "google-client-id")
	handler.googleVerifier = fakeGoogleVerifier{identity: &GoogleIdentity{
		Subject:       "google-sub",
		Email:         "new@example.com",
		EmailVerified: true,
		FirstName:     "New",
		LastName:      "Client",
	}}
	router.POST("/api/auth/google", handler.GoogleLogin)

	req, err := http.NewRequest(http.MethodPost, "/api/auth/google", strings.NewReader(`{"credential":"good-token"}`))
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
	if response.Token == "" || response.User.ID != 22 {
		t.Fatalf("unexpected auth response: %+v", response)
	}
}

func TestGoogleLoginUsesExistingGoogleSub(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeUserStore()
	store.getByGoogleSubFunc = func(_ context.Context, googleSub string) (*models.User, error) {
		if googleSub != "google-sub" {
			t.Fatalf("unexpected google sub %q", googleSub)
		}
		return &models.User{ID: 33, Name: "Returning User", Email: "returning@example.com", Role: models.UserRoleUser}, nil
	}

	router := gin.New()
	handler := NewAuthHandler(store, auth.NewTokenManager("secret", time.Hour), "setup-token", "google-client-id")
	handler.googleVerifier = fakeGoogleVerifier{identity: &GoogleIdentity{
		Subject:       "google-sub",
		Email:         "returning@example.com",
		EmailVerified: true,
	}}
	router.POST("/api/auth/google", handler.GoogleLogin)

	req, err := http.NewRequest(http.MethodPost, "/api/auth/google", strings.NewReader(`{"credential":"good-token"}`))
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
	if response.User.ID != 33 {
		t.Fatalf("unexpected auth response: %+v", response)
	}
}

func TestGoogleLoginAutoLinksExistingEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeUserStore()
	store.getByEmailFunc = func(_ context.Context, email string) (*models.UserWithPassword, error) {
		if email != "client@example.com" {
			t.Fatalf("unexpected email %q", email)
		}
		return &models.UserWithPassword{
			User: models.User{ID: 44, Name: "Existing Client", Email: email, Role: models.UserRoleUser},
		}, nil
	}
	store.linkGoogleSubFunc = func(_ context.Context, id int64, googleSub string) (*models.User, error) {
		if id != 44 || googleSub != "google-sub" {
			t.Fatalf("unexpected link request: id=%d sub=%q", id, googleSub)
		}
		return &models.User{ID: id, Name: "Existing Client", Email: "client@example.com", Role: models.UserRoleUser}, nil
	}

	router := gin.New()
	handler := NewAuthHandler(store, auth.NewTokenManager("secret", time.Hour), "setup-token", "google-client-id")
	handler.googleVerifier = fakeGoogleVerifier{identity: &GoogleIdentity{
		Subject:       "google-sub",
		Email:         "client@example.com",
		EmailVerified: true,
	}}
	router.POST("/api/auth/google", handler.GoogleLogin)

	req, err := http.NewRequest(http.MethodPost, "/api/auth/google", strings.NewReader(`{"credential":"good-token"}`))
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
	if response.User.ID != 44 || response.Token == "" {
		t.Fatalf("unexpected auth response: %+v", response)
	}
}

func TestMeRequiresBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterAuthRoutes(router.Group("/api/auth"), newFakeUserStore(), auth.NewTokenManager("secret", time.Hour), "setup-token", "")

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
	RegisterAuthRoutes(router.Group("/api/auth"), newFakeUserStore(), tokenManager, "setup-token", "")

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
	RegisterAuthRoutes(router.Group("/api/auth"), store, tokenManager, "setup-token", "")

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

func TestUpdateMeReturnsUpdatedProfile(t *testing.T) {
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
	store.updateProfileFunc = func(_ context.Context, id int64, input models.UpdateCurrentUserInput) (*models.User, error) {
		if id != 12 {
			t.Fatalf("expected id 12, got %d", id)
		}
		if input.FirstName != "Updated" || input.LastName != "User" || input.Email != "updated@example.com" {
			t.Fatalf("unexpected profile input: %+v", input)
		}
		return &models.User{
			ID:        id,
			FirstName: input.FirstName,
			LastName:  input.LastName,
			Name:      input.FirstName + " " + input.LastName,
			Email:     input.Email,
			Role:      models.UserRoleUser,
		}, nil
	}

	router := gin.New()
	RegisterAuthRoutes(router.Group("/api/auth"), store, tokenManager, "setup-token", "")

	req, err := http.NewRequest(http.MethodPatch, "/api/auth/me", strings.NewReader(`{
		"first_name":" Updated ",
		"last_name":" User ",
		"email":"updated@example.com"
	}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
}

func TestUpdateMeReturnsConflictOnDuplicateEmail(t *testing.T) {
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
	store.updateProfileFunc = func(context.Context, int64, models.UpdateCurrentUserInput) (*models.User, error) {
		return nil, repository.ErrDuplicate
	}

	router := gin.New()
	RegisterAuthRoutes(router.Group("/api/auth"), store, tokenManager, "setup-token", "")

	req, err := http.NewRequest(http.MethodPatch, "/api/auth/me", strings.NewReader(`{
		"first_name":"Updated",
		"last_name":"User",
		"email":"taken@example.com"
	}`))
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
		createGoogleFunc: func(context.Context, models.GoogleUserInput) (*models.User, error) {
			return &models.User{}, nil
		},
		getByIDFunc: func(context.Context, int64) (*models.User, error) {
			return nil, repository.ErrNotFound
		},
		getByEmailFunc: func(context.Context, string) (*models.UserWithPassword, error) {
			return nil, repository.ErrNotFound
		},
		getByGoogleSubFunc: func(context.Context, string) (*models.User, error) {
			return nil, repository.ErrNotFound
		},
		linkGoogleSubFunc: func(context.Context, int64, string) (*models.User, error) {
			return nil, repository.ErrNotFound
		},
		updateProfileFunc: func(context.Context, int64, models.UpdateCurrentUserInput) (*models.User, error) {
			return nil, repository.ErrNotFound
		},
	}
}
