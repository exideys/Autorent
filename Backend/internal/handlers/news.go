package handlers

import (
	"context"
	"net/http"

	"autorent-backend/internal/models"

	"github.com/gin-gonic/gin"
)

type NewsStore interface {
	List(ctx context.Context, filters models.NewsFilters) ([]models.NewsArticle, error)
	GetByID(ctx context.Context, id int64) (*models.NewsArticle, error)
	Create(ctx context.Context, input models.NewsInput) (*models.NewsArticle, error)
	Update(ctx context.Context, id int64, input models.NewsInput) (*models.NewsArticle, error)
	Delete(ctx context.Context, id int64) error
}

type NewsHandler struct {
	store NewsStore
}

func NewNewsHandler(store NewsStore) *NewsHandler {
	return &NewsHandler{store: store}
}

func RegisterNewsRoutes(router gin.IRouter, store NewsStore) {
	handler := NewNewsHandler(store)

	news := router.Group("/news")
	news.GET("", handler.ListPublishedNews)
	news.GET("/:id", handler.GetPublishedNews)
}

func RegisterAdminNewsRoutes(router gin.IRouter, store NewsStore) {
	handler := NewNewsHandler(store)

	news := router.Group("/news")
	news.GET("", handler.ListAdminNews)
	news.POST("", handler.CreateNews)
	news.GET("/:id", handler.GetAdminNews)
	news.PUT("/:id", handler.UpdateNews)
	news.DELETE("/:id", handler.DeleteNews)
}

func (h *NewsHandler) ListPublishedNews(c *gin.Context) {
	articles, err := h.store.List(c.Request.Context(), models.NewsFilters{
		Status:    models.NewsStatusPublished,
		SortBy:    c.DefaultQuery("sort", "published_at"),
		SortOrder: c.DefaultQuery("order", "desc"),
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load news")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": articles})
}

func (h *NewsHandler) GetPublishedNews(c *gin.Context) {
	article, ok := h.getNews(c)
	if !ok {
		return
	}
	if article.Status != models.NewsStatusPublished {
		respondError(c, http.StatusNotFound, "news article not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": article})
}

func (h *NewsHandler) ListAdminNews(c *gin.Context) {
	articles, err := h.store.List(c.Request.Context(), models.NewsFilters{
		SortBy:    c.DefaultQuery("sort", "created_at"),
		SortOrder: c.DefaultQuery("order", "desc"),
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load news")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": articles})
}

func (h *NewsHandler) GetAdminNews(c *gin.Context) {
	article, ok := h.getNews(c)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": article})
}

func (h *NewsHandler) CreateNews(c *gin.Context) {
	var input models.NewsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "invalid news payload")
		return
	}

	article, err := h.store.Create(c.Request.Context(), input)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create news")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": article})
}

func (h *NewsHandler) UpdateNews(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var input models.NewsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "invalid news payload")
		return
	}

	article, err := h.store.Update(c.Request.Context(), id, input)
	if err != nil {
		respondRepositoryError(c, err, "news article not found", "failed to update news")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": article})
}

func (h *NewsHandler) DeleteNews(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.store.Delete(c.Request.Context(), id); err != nil {
		respondRepositoryError(c, err, "news article not found", "failed to delete news")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *NewsHandler) getNews(c *gin.Context) (*models.NewsArticle, bool) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return nil, false
	}

	article, err := h.store.GetByID(c.Request.Context(), id)
	if err != nil {
		respondRepositoryError(c, err, "news article not found", "failed to load news")
		return nil, false
	}

	return article, true
}
