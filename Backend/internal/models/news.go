package models

import "time"

const (
	NewsStatusDraft     = "draft"
	NewsStatusPublished = "published"
)

type NewsArticle struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Summary     string     `json:"summary"`
	Content     string     `json:"content"`
	ImageURL    *string    `json:"image_url,omitempty"`
	Status      string     `json:"status"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type NewsInput struct {
	Title    string  `json:"title" binding:"required,max=120"`
	Summary  string  `json:"summary" binding:"required,max=240"`
	Content  string  `json:"content" binding:"required"`
	ImageURL *string `json:"image_url,omitempty" binding:"omitempty,max=2048"`
	Status   string  `json:"status,omitempty" binding:"omitempty,max=30"`
}

type NewsFilters struct {
	Status    string
	SortBy    string
	SortOrder string
}
