package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"autorent-backend/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewsRepositoryListAppliesFiltersAndScansNullableFields(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	now := time.Now()
	imageURL := "https://example.com/news.jpg"
	publishedAt := now.Add(-time.Hour)
	mock.ExpectQuery("SELECT").
		WithArgs(models.NewsStatusPublished).
		WillReturnRows(newsRows().
			AddRow(1, "Launch", "Summary", "Content", imageURL, models.NewsStatusPublished, publishedAt, now, now))

	repo := NewNewsRepository(db)
	articles, err := repo.List(context.Background(), models.NewsFilters{
		Status:    models.NewsStatusPublished,
		SortBy:    "title",
		SortOrder: "asc",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(articles) != 1 {
		t.Fatalf("expected 1 article, got %d", len(articles))
	}
	article := articles[0]
	if article.ID != 1 || article.Title != "Launch" || article.Status != models.NewsStatusPublished {
		t.Fatalf("unexpected article: %+v", article)
	}
	if article.ImageURL == nil || *article.ImageURL != imageURL {
		t.Fatalf("unexpected image url: %+v", article.ImageURL)
	}
	if article.PublishedAt == nil || !article.PublishedAt.Equal(publishedAt) {
		t.Fatalf("unexpected published_at: %+v", article.PublishedAt)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestNewsRepositoryGetByIDNotFound(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectQuery("SELECT").WithArgs(int64(99)).WillReturnRows(newsRows())

	repo := NewNewsRepository(db)
	_, err := repo.GetByID(context.Background(), 99)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestNewsRepositoryCreateTrimsInputAndDefaultsDraft(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	now := time.Now()
	blankImage := "   "
	mock.ExpectExec("INSERT INTO news").
		WithArgs("Title", "Summary", "Content", nil, models.NewsStatusDraft, nil).
		WillReturnResult(sqlmock.NewResult(12, 1))
	mock.ExpectQuery("SELECT").
		WithArgs(int64(12)).
		WillReturnRows(newsRows().
			AddRow(12, "Title", "Summary", "Content", nil, models.NewsStatusDraft, nil, now, now))

	repo := NewNewsRepository(db)
	article, err := repo.Create(context.Background(), models.NewsInput{
		Title:    " Title ",
		Summary:  " Summary ",
		Content:  " Content ",
		ImageURL: &blankImage,
		Status:   "unknown",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if article.ID != 12 || article.Status != models.NewsStatusDraft || article.ImageURL != nil || article.PublishedAt != nil {
		t.Fatalf("unexpected article: %+v", article)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestNewsRepositoryCreatePublishedSetsPublishedAt(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	now := time.Now()
	imageURL := " https://example.com/news.jpg "
	mock.ExpectExec("INSERT INTO news").
		WithArgs("Title", "Summary", "Content", "https://example.com/news.jpg", models.NewsStatusPublished, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(13, 1))
	mock.ExpectQuery("SELECT").
		WithArgs(int64(13)).
		WillReturnRows(newsRows().
			AddRow(13, "Title", "Summary", "Content", "https://example.com/news.jpg", models.NewsStatusPublished, now, now, now))

	repo := NewNewsRepository(db)
	article, err := repo.Create(context.Background(), models.NewsInput{
		Title:    "Title",
		Summary:  "Summary",
		Content:  "Content",
		ImageURL: &imageURL,
		Status:   " published ",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if article.PublishedAt == nil {
		t.Fatalf("expected published_at to be scanned: %+v", article)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestNewsRepositoryUpdateNotFound(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectExec("UPDATE news").
		WithArgs("Title", "Summary", "Content", nil, models.NewsStatusDraft, models.NewsStatusDraft, sqlmock.AnyArg(), int64(88)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(int64(88)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(0))

	repo := NewNewsRepository(db)
	_, err := repo.Update(context.Background(), 88, models.NewsInput{
		Title:   "Title",
		Summary: "Summary",
		Content: "Content",
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestNewsRepositoryUpdateExistingUnchangedRow(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectExec("UPDATE news").
		WithArgs("Title", "Summary", "Content", nil, models.NewsStatusDraft, models.NewsStatusDraft, sqlmock.AnyArg(), int64(14)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(int64(14)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))
	mock.ExpectQuery("SELECT").
		WithArgs(int64(14)).
		WillReturnRows(newsRows().
			AddRow(14, "Title", "Summary", "Content", nil, models.NewsStatusDraft, nil, now, now))

	repo := NewNewsRepository(db)
	article, err := repo.Update(context.Background(), 14, models.NewsInput{
		Title:   "Title",
		Summary: "Summary",
		Content: "Content",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if article.ID != 14 {
		t.Fatalf("unexpected article: %+v", article)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestNewsRepositoryDeleteNotFound(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectExec("DELETE FROM news").
		WithArgs(int64(77)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	repo := NewNewsRepository(db)
	err := repo.Delete(context.Background(), 77)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestNewsRepositoryDelete(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectExec("DELETE FROM news").
		WithArgs(int64(77)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewNewsRepository(db)
	if err := repo.Delete(context.Background(), 77); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func newsRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"title",
		"summary",
		"content",
		"image_url",
		"status",
		"published_at",
		"created_at",
		"updated_at",
	})
}
