package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"autorent-backend/internal/models"
)

type NewsRepository struct {
	db *sql.DB
}

func NewNewsRepository(db *sql.DB) *NewsRepository {
	return &NewsRepository{db: db}
}

func (r *NewsRepository) List(ctx context.Context, filters models.NewsFilters) ([]models.NewsArticle, error) {
	query := baseNewsQuery()
	args := make([]any, 0)
	conditions := make([]string, 0)

	if filters.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filters.Status)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += newsOrderBy(filters.SortBy, filters.SortOrder)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanNewsArticles(rows)
}

func (r *NewsRepository) GetByID(ctx context.Context, id int64) (*models.NewsArticle, error) {
	rows, err := r.db.QueryContext(ctx, baseNewsQuery()+" WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles, err := scanNewsArticles(rows)
	if err != nil {
		return nil, err
	}
	if len(articles) == 0 {
		return nil, ErrNotFound
	}

	return &articles[0], nil
}

func (r *NewsRepository) Create(ctx context.Context, input models.NewsInput) (*models.NewsArticle, error) {
	status := newsStatusOrDefault(input.Status)
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO news (title, summary, content, image_url, status, published_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		strings.TrimSpace(input.Title),
		strings.TrimSpace(input.Summary),
		strings.TrimSpace(input.Content),
		trimmedStringPtrValue(input.ImageURL),
		status,
		newsPublishedAtValue(status),
	)
	if err != nil {
		return nil, err
	}

	articleID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, articleID)
}

func (r *NewsRepository) Update(ctx context.Context, id int64, input models.NewsInput) (*models.NewsArticle, error) {
	status := newsStatusOrDefault(input.Status)
	result, err := r.db.ExecContext(ctx, `
		UPDATE news
		SET
			title = ?,
			summary = ?,
			content = ?,
			image_url = ?,
			status = ?,
			published_at = CASE WHEN ? = 'published' THEN COALESCE(published_at, ?) ELSE NULL END
		WHERE id = ?
	`,
		strings.TrimSpace(input.Title),
		strings.TrimSpace(input.Summary),
		strings.TrimSpace(input.Content),
		trimmedStringPtrValue(input.ImageURL),
		status,
		status,
		time.Now().UTC(),
		id,
	)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		exists, err := newsExists(ctx, r.db, id)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrNotFound
		}
	}

	return r.GetByID(ctx, id)
}

func (r *NewsRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM news WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func baseNewsQuery() string {
	return `
		SELECT
			id,
			title,
			summary,
			content,
			image_url,
			status,
			published_at,
			created_at,
			updated_at
		FROM news
	`
}

func newsOrderBy(sortBy string, sortOrder string) string {
	column := allowedNewsSortColumns()[sortBy]
	if column == "" {
		column = "COALESCE(published_at, created_at)"
	}

	direction := "DESC"
	if strings.EqualFold(sortOrder, "asc") {
		direction = "ASC"
	}

	return fmt.Sprintf(" ORDER BY %s %s, id DESC", column, direction)
}

func allowedNewsSortColumns() map[string]string {
	return map[string]string{
		"title":        "title",
		"status":       "status",
		"published_at": "published_at",
		"created_at":   "created_at",
		"updated_at":   "updated_at",
	}
}

func scanNewsArticles(rows *sql.Rows) ([]models.NewsArticle, error) {
	articles := make([]models.NewsArticle, 0)

	for rows.Next() {
		var article models.NewsArticle
		var imageURL sql.NullString
		var publishedAt sql.NullTime

		if err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.Summary,
			&article.Content,
			&imageURL,
			&article.Status,
			&publishedAt,
			&article.CreatedAt,
			&article.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if imageURL.Valid {
			article.ImageURL = &imageURL.String
		}
		if publishedAt.Valid {
			article.PublishedAt = &publishedAt.Time
		}

		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

func newsStatusOrDefault(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case models.NewsStatusPublished:
		return models.NewsStatusPublished
	default:
		return models.NewsStatusDraft
	}
}

func newsPublishedAtValue(status string) any {
	if status != models.NewsStatusPublished {
		return nil
	}
	return time.Now().UTC()
}

func trimmedStringPtrValue(value *string) any {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func newsExists(ctx context.Context, db *sql.DB, id int64) (bool, error) {
	var exists int
	if err := db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM news WHERE id = ?)", id).Scan(&exists); err != nil {
		return false, err
	}

	return exists == 1, nil
}
