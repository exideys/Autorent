package auth

import (
	"testing"
	"time"

	"autorent-backend/internal/models"
)

func TestTokenManagerGenerateAndVerify(t *testing.T) {
	manager := NewTokenManager("test-secret", time.Hour)
	user := models.User{
		ID:    42,
		Email: "admin@example.com",
		Role:  models.UserRoleAdmin,
	}

	token, err := manager.Generate(user)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := manager.Verify(token)
	if err != nil {
		t.Fatalf("failed to verify token: %v", err)
	}

	if claims.UserID != user.ID {
		t.Fatalf("expected user id %d, got %d", user.ID, claims.UserID)
	}
	if claims.Role != user.Role {
		t.Fatalf("expected role %q, got %q", user.Role, claims.Role)
	}
	if claims.Subject != user.Email {
		t.Fatalf("expected subject %q, got %q", user.Email, claims.Subject)
	}
}

func TestTokenManagerRejectsInvalidToken(t *testing.T) {
	manager := NewTokenManager("test-secret", time.Hour)

	if _, err := manager.Verify("not-a-token"); err == nil {
		t.Fatal("expected invalid token error")
	}
}

func TestTokenManagerRejectsExpiredToken(t *testing.T) {
	manager := NewTokenManager("test-secret", -time.Hour)

	token, err := manager.Generate(models.User{
		ID:    1,
		Email: "user@example.com",
		Role:  models.UserRoleUser,
	})
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	if _, err := manager.Verify(token); err == nil {
		t.Fatal("expected expired token error")
	}
}
