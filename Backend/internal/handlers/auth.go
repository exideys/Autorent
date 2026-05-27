package handlers

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"

	"autorent-backend/internal/auth"
	"autorent-backend/internal/models"
	"autorent-backend/internal/repository"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserStore interface {
	Create(ctx context.Context, input models.RegisterInput, passwordHash string, role string) (*models.User, error)
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.UserWithPassword, error)
}

type AuthHandler struct {
	store           UserStore
	tokens          *auth.TokenManager
	adminSetupToken string
}

func NewAuthHandler(store UserStore, tokens *auth.TokenManager, adminSetupToken string) *AuthHandler {
	return &AuthHandler{
		store:           store,
		tokens:          tokens,
		adminSetupToken: adminSetupToken,
	}
}

func RegisterAuthRoutes(router gin.IRouter, store UserStore, tokens *auth.TokenManager, adminSetupToken string) {
	handler := NewAuthHandler(store, tokens, adminSetupToken)

	router.POST("/register", handler.RegisterUser)
	router.POST("/login", handler.Login)
	router.POST("/admin/register", handler.RegisterAdmin)
	router.GET("/me", RequireAuth(tokens), handler.Me)
}

func (h *AuthHandler) RegisterUser(c *gin.Context) {
	h.register(c, models.UserRoleUser)
}

func (h *AuthHandler) RegisterAdmin(c *gin.Context) {
	if h.adminSetupToken == "" {
		respondError(c, http.StatusForbidden, "admin registration is disabled")
		return
	}

	providedToken := c.GetHeader("X-Admin-Setup-Token")
	if subtle.ConstantTimeCompare([]byte(providedToken), []byte(h.adminSetupToken)) != 1 {
		respondError(c, http.StatusUnauthorized, "invalid admin setup token")
		return
	}

	h.register(c, models.UserRoleAdmin)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input models.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "invalid login payload")
		return
	}

	user, err := h.store.GetByEmail(c.Request.Context(), input.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(c, http.StatusUnauthorized, "invalid email or password")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to login")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		respondError(c, http.StatusUnauthorized, "invalid email or password")
		return
	}

	h.respondWithToken(c, http.StatusOK, user.User)
}

func (h *AuthHandler) Me(c *gin.Context) {
	claims, ok := authClaims(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "missing auth context")
		return
	}

	user, err := h.store.GetByID(c.Request.Context(), claims.UserID)
	if err != nil {
		respondRepositoryError(c, err, "user not found", "failed to load user")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (h *AuthHandler) register(c *gin.Context, role string) {
	var input models.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "invalid registration payload")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user, err := h.store.Create(c.Request.Context(), input, string(passwordHash), role)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			respondError(c, http.StatusConflict, "email already exists")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to register user")
		return
	}

	h.respondWithToken(c, http.StatusCreated, *user)
}

func (h *AuthHandler) respondWithToken(c *gin.Context, status int, user models.User) {
	token, err := h.tokens.Generate(user)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create token")
		return
	}

	c.JSON(status, gin.H{
		"data": models.AuthResponse{
			Token: token,
			User:  user,
		},
	})
}
