package handlers

import (
	"net/http"
	"strings"

	"autorent-backend/internal/auth"
	"autorent-backend/internal/models"

	"github.com/gin-gonic/gin"
)

const authClaimsContextKey = "auth_claims"

func RequireAuth(tokens *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			respondError(c, http.StatusUnauthorized, "missing bearer token")
			c.Abort()
			return
		}

		claims, err := tokens.Verify(tokenString)
		if err != nil {
			respondError(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		c.Set(authClaimsContextKey, claims)
		c.Next()
	}
}

func RequireAdmin(tokens *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			respondError(c, http.StatusUnauthorized, "missing bearer token")
			c.Abort()
			return
		}

		claims, err := tokens.Verify(tokenString)
		if err != nil {
			respondError(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		if claims.Role != models.UserRoleAdmin {
			respondError(c, http.StatusForbidden, "admin access required")
			c.Abort()
			return
		}

		c.Set(authClaimsContextKey, claims)
		c.Next()
	}
}

func authClaims(c *gin.Context) (*auth.Claims, bool) {
	value, ok := c.Get(authClaimsContextKey)
	if !ok {
		return nil, false
	}

	claims, ok := value.(*auth.Claims)
	return claims, ok
}

func bearerToken(header string) (string, bool) {
	prefix := "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return token, token != ""
}
