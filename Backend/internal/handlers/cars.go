package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"autorent-backend/internal/models"
	"autorent-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type CarStore interface {
	List(ctx context.Context, filters models.CarFilters) ([]models.Car, error)
	GetByID(ctx context.Context, id int64) (*models.Car, error)
	Create(ctx context.Context, input models.CarInput) (*models.Car, error)
	Update(ctx context.Context, id int64, input models.CarInput) (*models.Car, error)
	Delete(ctx context.Context, id int64) error
	AddImage(ctx context.Context, carID int64, input models.CarImageInput) (*models.CarImage, error)
	DeleteImage(ctx context.Context, imageID int64) error
}

type CarHandler struct {
	store CarStore
}

func NewCarHandler(store CarStore) *CarHandler {
	return &CarHandler{store: store}
}

func RegisterCarRoutes(router gin.IRouter, store CarStore) {
	handler := NewCarHandler(store)

	cars := router.Group("/cars")
	cars.GET("", handler.ListCars)
	cars.GET("/:id", handler.GetCar)
}

func RegisterAdminCarRoutes(router gin.IRouter, store CarStore) {
	handler := NewCarHandler(store)

	cars := router.Group("/cars")
	cars.GET("", handler.ListCars)
	cars.POST("", handler.CreateCar)
	cars.GET("/:id", handler.GetCar)
	cars.PUT("/:id", handler.UpdateCar)
	cars.DELETE("/:id", handler.DeleteCar)
	cars.POST("/:id/images", handler.AddCarImage)

	router.DELETE("/car-images/:id", handler.DeleteCarImage)
}

func (h *CarHandler) ListCars(c *gin.Context) {
	filters := models.CarFilters{
		Status:       c.Query("status"),
		CarClass:     c.Query("car_class"),
		BodyType:     c.Query("body_type"),
		Transmission: c.Query("transmission"),
		FuelType:     c.Query("fuel_type"),
		SortBy:       c.DefaultQuery("sort", "created_at"),
		SortOrder:    c.DefaultQuery("order", "desc"),
	}

	if filters.Status == "" && c.Query("available") == "true" {
		filters.Status = "available"
	}

	cars, err := h.store.List(c.Request.Context(), filters)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load cars")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": cars})
}

func (h *CarHandler) GetCar(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	car, err := h.store.GetByID(c.Request.Context(), id)
	if err != nil {
		respondRepositoryError(c, err, "car not found", "failed to load car")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": car})
}

func (h *CarHandler) CreateCar(c *gin.Context) {
	var input models.CarInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "invalid car payload")
		return
	}

	car, err := h.store.Create(c.Request.Context(), input)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create car")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": car})
}

func (h *CarHandler) UpdateCar(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var input models.CarInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "invalid car payload")
		return
	}

	car, err := h.store.Update(c.Request.Context(), id, input)
	if err != nil {
		respondRepositoryError(c, err, "car not found", "failed to update car")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": car})
}

func (h *CarHandler) DeleteCar(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.store.Delete(c.Request.Context(), id); err != nil {
		respondRepositoryError(c, err, "car not found", "failed to delete car")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *CarHandler) AddCarImage(c *gin.Context) {
	carID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var input models.CarImageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "invalid car image payload")
		return
	}

	image, err := h.store.AddImage(c.Request.Context(), carID, input)
	if err != nil {
		respondRepositoryError(c, err, "car not found", "failed to add car image")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": image})
}

func (h *CarHandler) DeleteCarImage(c *gin.Context) {
	imageID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.store.DeleteImage(c.Request.Context(), imageID); err != nil {
		respondRepositoryError(c, err, "car image not found", "failed to delete car image")
		return
	}

	c.Status(http.StatusNoContent)
}

func parseIDParam(c *gin.Context, name string) (int64, bool) {
	value, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || value <= 0 {
		respondError(c, http.StatusBadRequest, "invalid id")
		return 0, false
	}

	return value, true
}

func respondRepositoryError(c *gin.Context, err error, notFoundMessage string, fallbackMessage string) {
	if errors.Is(err, repository.ErrNotFound) {
		respondError(c, http.StatusNotFound, notFoundMessage)
		return
	}

	respondError(c, http.StatusInternalServerError, fallbackMessage)
}

func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}
