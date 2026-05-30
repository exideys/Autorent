package handlers

import (
	"context"
	"net/http"

	"autorent-backend/internal/models"

	"github.com/gin-gonic/gin"
)

type AdminUserStore interface {
	ListCustomers(ctx context.Context) ([]models.User, error)
	RateCustomer(ctx context.Context, id int64, rating float64) (*models.User, error)
}

type AdminUserHandler struct {
	store AdminUserStore
}

func NewAdminUserHandler(store AdminUserStore) *AdminUserHandler {
	return &AdminUserHandler{store: store}
}

func RegisterAdminUserRoutes(router gin.IRouter, store AdminUserStore) {
	handler := NewAdminUserHandler(store)

	users := router.Group("/users")
	users.GET("", handler.ListCustomers)
	users.PATCH("/:id/rating", handler.RateCustomer)
}

func (h *AdminUserHandler) ListCustomers(c *gin.Context) {
	users, err := h.store.ListCustomers(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load customers")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (h *AdminUserHandler) RateCustomer(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var input models.RateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "invalid rating payload")
		return
	}

	user, err := h.store.RateCustomer(c.Request.Context(), id, input.Rating)
	if err != nil {
		respondRepositoryError(c, err, "customer not found", "failed to rate customer")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}
