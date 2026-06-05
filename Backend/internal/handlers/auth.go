package handlers

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"autorent-backend/internal/auth"
	"autorent-backend/internal/models"
	"autorent-backend/internal/repository"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

type UserStore interface {
	Create(ctx context.Context, input models.RegisterInput, passwordHash string, role string) (*models.User, error)
	CreateGoogle(ctx context.Context, input models.GoogleUserInput) (*models.User, error)
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.UserWithPassword, error)
	GetByGoogleSub(ctx context.Context, googleSub string) (*models.User, error)
	LinkGoogleSub(ctx context.Context, id int64, googleSub string) (*models.User, error)
	UpdateProfile(ctx context.Context, id int64, input models.UpdateCurrentUserInput) (*models.User, error)
}

type GoogleIdentity struct {
	Subject       string
	Email         string
	EmailVerified bool
	FirstName     string
	LastName      string
}

type GoogleIdentityVerifier interface {
	Verify(ctx context.Context, credential string, audience string) (*GoogleIdentity, error)
}

type GoogleIDTokenVerifier struct{}

func (GoogleIDTokenVerifier) Verify(ctx context.Context, credential string, audience string) (*GoogleIdentity, error) {
	payload, err := idtoken.Validate(ctx, credential, audience)
	if err != nil {
		return nil, err
	}

	return &GoogleIdentity{
		Subject:       payload.Subject,
		Email:         claimString(payload.Claims, "email"),
		EmailVerified: claimBool(payload.Claims, "email_verified"),
		FirstName:     claimString(payload.Claims, "given_name"),
		LastName:      claimString(payload.Claims, "family_name"),
	}, nil
}

type AuthHandler struct {
	store           UserStore
	tokens          *auth.TokenManager
	adminSetupToken string
	googleClientID  string
	googleVerifier  GoogleIdentityVerifier
}

func NewAuthHandler(store UserStore, tokens *auth.TokenManager, adminSetupToken string, googleClientID string) *AuthHandler {
	return &AuthHandler{
		store:           store,
		tokens:          tokens,
		adminSetupToken: adminSetupToken,
		googleClientID:  googleClientID,
		googleVerifier:  GoogleIDTokenVerifier{},
	}
}

func RegisterAuthRoutes(router gin.IRouter, store UserStore, tokens *auth.TokenManager, adminSetupToken string, googleClientID string) {
	handler := NewAuthHandler(store, tokens, adminSetupToken, googleClientID)

	router.POST("/register", handler.RegisterUser)
	router.POST("/login", handler.Login)
	router.POST("/google", handler.GoogleLogin)
	router.POST("/admin/register", handler.RegisterAdmin)
	router.GET("/me", RequireAuth(tokens), handler.Me)
	router.PATCH("/me", RequireAuth(tokens), handler.UpdateMe)
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

	if user.PasswordHash == "" {
		respondError(c, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		respondError(c, http.StatusUnauthorized, "invalid email or password")
		return
	}

	h.respondWithToken(c, http.StatusOK, user.User)
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	if strings.TrimSpace(h.googleClientID) == "" {
		respondError(c, http.StatusServiceUnavailable, "Google login is not configured")
		return
	}

	var input models.GoogleLoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "invalid Google login payload")
		return
	}

	identity, err := h.googleVerifier.Verify(c.Request.Context(), input.Credential, h.googleClientID)
	if err != nil || !validGoogleIdentity(identity) {
		respondError(c, http.StatusUnauthorized, "invalid Google credential")
		return
	}

	user, err := h.store.GetByGoogleSub(c.Request.Context(), identity.Subject)
	if err == nil {
		h.respondWithToken(c, http.StatusOK, *user)
		return
	}
	if !errors.Is(err, repository.ErrNotFound) {
		respondError(c, http.StatusInternalServerError, "failed to login with Google")
		return
	}

	existingUser, err := h.store.GetByEmail(c.Request.Context(), identity.Email)
	if err == nil {
		linkedUser, linkErr := h.store.LinkGoogleSub(c.Request.Context(), existingUser.ID, identity.Subject)
		if linkErr != nil {
			if errors.Is(linkErr, repository.ErrDuplicate) {
				respondError(c, http.StatusConflict, "Google account is already linked")
				return
			}
			respondRepositoryError(c, linkErr, "user not found", "failed to login with Google")
			return
		}
		h.respondWithToken(c, http.StatusOK, *linkedUser)
		return
	}
	if !errors.Is(err, repository.ErrNotFound) {
		respondError(c, http.StatusInternalServerError, "failed to login with Google")
		return
	}

	firstName, lastName := googleNameParts(identity)
	createdUser, err := h.store.CreateGoogle(c.Request.Context(), models.GoogleUserInput{
		GoogleSub: identity.Subject,
		FirstName: firstName,
		LastName:  lastName,
		Email:     identity.Email,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			respondError(c, http.StatusConflict, "email already exists")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to login with Google")
		return
	}

	h.respondWithToken(c, http.StatusCreated, *createdUser)
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

func (h *AuthHandler) UpdateMe(c *gin.Context) {
	claims, ok := authClaims(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "missing auth context")
		return
	}

	var input models.UpdateCurrentUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "invalid profile payload")
		return
	}
	input.FirstName = strings.TrimSpace(input.FirstName)
	input.LastName = strings.TrimSpace(input.LastName)
	input.Email = strings.TrimSpace(input.Email)
	if input.FirstName == "" || input.LastName == "" || input.Email == "" {
		respondError(c, http.StatusBadRequest, "invalid profile payload")
		return
	}

	user, err := h.store.UpdateProfile(c.Request.Context(), claims.UserID, input)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			respondError(c, http.StatusConflict, "email already exists")
			return
		}
		respondRepositoryError(c, err, "user not found", "failed to update profile")
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
	if !input.HasName() {
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

func validGoogleIdentity(identity *GoogleIdentity) bool {
	return identity != nil &&
		strings.TrimSpace(identity.Subject) != "" &&
		strings.TrimSpace(identity.Email) != "" &&
		identity.EmailVerified
}

func googleNameParts(identity *GoogleIdentity) (string, string) {
	firstName := strings.TrimSpace(identity.FirstName)
	lastName := strings.TrimSpace(identity.LastName)
	if firstName != "" || lastName != "" {
		return firstName, lastName
	}

	email := strings.TrimSpace(identity.Email)
	localPart := email
	if atIndex := strings.Index(localPart, "@"); atIndex > 0 {
		localPart = localPart[:atIndex]
	}
	localPart = strings.TrimSpace(localPart)
	if localPart == "" {
		return "Google", "User"
	}

	return localPart, ""
}

func claimString(claims map[string]interface{}, key string) string {
	value, ok := claims[key].(string)
	if !ok {
		return ""
	}
	return value
}

func claimBool(claims map[string]interface{}, key string) bool {
	value, ok := claims[key].(bool)
	return ok && value
}
