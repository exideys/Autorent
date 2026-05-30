package handlers

import (
	"context"
	"errors"
	"net/http"

	"autorent-backend/internal/auth"
	"autorent-backend/internal/models"
	"autorent-backend/internal/repository"
	"autorent-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type RentalOrderStore interface {
	Create(ctx context.Context, userID int64, input models.RentalOrderInput) (*models.RentalOrder, error)
	ListByUserID(ctx context.Context, userID int64) ([]models.RentalOrder, error)
}

type RentalOrderHandler struct {
	store RentalOrderStore
}

func NewRentalOrderHandler(store RentalOrderStore) *RentalOrderHandler {
	return &RentalOrderHandler{store: store}
}

func RegisterRentalOrderRoutes(router gin.IRouter, store RentalOrderStore, tokens *auth.TokenManager) {
	handler := NewRentalOrderHandler(store)

	orders := router.Group("/rental-orders")
	orders.Use(RequireAuth(tokens))
	orders.POST("", handler.CreateRentalOrder)
	orders.GET("", handler.ListMyRentalOrders)
}

func RegisterAdminRentalOrderRoutes(router gin.IRouter, store RentalOrderStore) {
	handler := NewRentalOrderHandler(store)

	router.GET("/users/:id/rental-orders", handler.ListUserRentalOrders)
}

func (h *RentalOrderHandler) CreateRentalOrder(c *gin.Context) {
	claims, ok := authClaims(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "missing auth context")
		return
	}

	var input models.RentalOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "invalid rental order payload")
		return
	}

	order, err := h.store.Create(c.Request.Context(), claims.UserID, input)
	if err != nil {
		respondRentalOrderError(c, err, "car not found")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": order})
}

func (h *RentalOrderHandler) ListMyRentalOrders(c *gin.Context) {
	claims, ok := authClaims(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "missing auth context")
		return
	}

	orders, err := h.store.ListByUserID(c.Request.Context(), claims.UserID)
	if err != nil {
		respondRentalOrderError(c, err, "user not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": orders})
}

func (h *RentalOrderHandler) ListUserRentalOrders(c *gin.Context) {
	userID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	orders, err := h.store.ListByUserID(c.Request.Context(), userID)
	if err != nil {
		respondRentalOrderError(c, err, "user not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": orders})
}

func respondRentalOrderError(c *gin.Context, err error, notFoundMessage string) {
	switch {
	case errors.Is(err, services.ErrInvalidInput):
		respondError(c, http.StatusBadRequest, "invalid rental order payload")
	case errors.Is(err, repository.ErrNotFound):
		respondError(c, http.StatusNotFound, notFoundMessage)
	case errors.Is(err, repository.ErrUnavailable):
		respondError(c, http.StatusConflict, "car is not available")
	default:
		respondError(c, http.StatusInternalServerError, "failed to process rental order")
	}
}
