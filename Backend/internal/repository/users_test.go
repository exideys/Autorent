package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"autorent-backend/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
)

func TestUserRepositoryCreate(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	now := time.Now()
	updatedAt := now.Add(time.Minute)
	mock.ExpectExec("INSERT INTO users").
		WithArgs("Test", "User", "user@example.com", "hashed-password", models.UserRoleUser).
		WillReturnResult(sqlmock.NewResult(7, 1))
	mock.ExpectQuery("SELECT id, first_name").
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"first_name",
			"last_name",
			"name",
			"email",
			"rating",
			"rating_count",
			"role",
			"status",
			"created_at",
			"updated_at",
		}).AddRow(7, "Test", "User", "Test User", "user@example.com", 5.0, 0, models.UserRoleUser, "active", now, updatedAt))

	repo := NewUserRepository(db)
	user, err := repo.Create(context.Background(), models.RegisterInput{
		FirstName: "Test",
		LastName:  "User",
		Email:     " USER@EXAMPLE.COM ",
	}, "hashed-password", models.UserRoleUser)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if user.ID != 7 || user.Email != "user@example.com" || user.FirstName != "Test" || user.Rating != 5.0 {
		t.Fatalf("unexpected user: %+v", user)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUserRepositoryCreateDuplicate(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectExec("INSERT INTO users").
		WillReturnError(&mysql.MySQLError{Number: 1062, Message: "duplicate entry"})

	repo := NewUserRepository(db)
	_, err := repo.Create(context.Background(), models.RegisterInput{
		FirstName: "Test",
		LastName:  "User",
		Email:     "user@example.com",
	}, "hash", models.UserRoleUser)
	if !errors.Is(err, ErrDuplicate) {
		t.Fatalf("expected ErrDuplicate, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUserRepositoryGetByEmail(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	now := time.Now()
	updatedAt := now.Add(time.Minute)
	mock.ExpectQuery("SELECT id, first_name").
		WithArgs("user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"first_name",
			"last_name",
			"name",
			"email",
			"password_hash",
			"rating",
			"rating_count",
			"role",
			"status",
			"created_at",
			"updated_at",
		}).AddRow(5, "Test", "User", "Test User", "user@example.com", "hash", 4.7, 3, models.UserRoleAdmin, "active", now, updatedAt))

	repo := NewUserRepository(db)
	user, err := repo.GetByEmail(context.Background(), " USER@EXAMPLE.COM ")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if user.ID != 5 || user.PasswordHash != "hash" || user.Role != models.UserRoleAdmin || user.RatingCount != 3 {
		t.Fatalf("unexpected user: %+v", user)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUserRepositoryGetByEmailNotFound(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectQuery("SELECT id, first_name").
		WithArgs("missing@example.com").
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"first_name",
			"last_name",
			"name",
			"email",
			"password_hash",
			"rating",
			"rating_count",
			"role",
			"status",
			"created_at",
			"updated_at",
		}))

	repo := NewUserRepository(db)
	_, err := repo.GetByEmail(context.Background(), "missing@example.com")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUserRepositoryGetByIDNotFound(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectQuery("SELECT id, first_name").
		WithArgs(int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"first_name",
			"last_name",
			"name",
			"email",
			"rating",
			"rating_count",
			"role",
			"status",
			"created_at",
			"updated_at",
		}))

	repo := NewUserRepository(db)
	_, err := repo.GetByID(context.Background(), 99)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUserRepositoryListCustomers(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	now := time.Now()
	updatedAt := now.Add(time.Minute)
	mock.ExpectQuery("SELECT id, first_name").
		WithArgs(models.UserRoleUser).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"first_name",
			"last_name",
			"name",
			"email",
			"rating",
			"rating_count",
			"role",
			"status",
			"created_at",
			"updated_at",
		}).AddRow(8, "Client", "One", "Client One", "client@example.com", 4.5, 2, models.UserRoleUser, "active", now, updatedAt))

	repo := NewUserRepository(db)
	users, err := repo.ListCustomers(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(users) != 1 || users[0].ID != 8 || users[0].Rating != 4.5 {
		t.Fatalf("unexpected users: %+v", users)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUserRepositoryRateCustomer(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	now := time.Now()
	updatedAt := now.Add(time.Minute)
	mock.ExpectExec("UPDATE users").
		WithArgs(4.0, int64(8), models.UserRoleUser).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT id, first_name").
		WithArgs(int64(8)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"first_name",
			"last_name",
			"name",
			"email",
			"rating",
			"rating_count",
			"role",
			"status",
			"created_at",
			"updated_at",
		}).AddRow(8, "Client", "One", "Client One", "client@example.com", 4.7, 3, models.UserRoleUser, "active", now, updatedAt))

	repo := NewUserRepository(db)
	user, err := repo.RateCustomer(context.Background(), 8, 4.0)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if user.ID != 8 || user.Rating != 4.7 || user.RatingCount != 3 {
		t.Fatalf("unexpected user: %+v", user)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sql mock: %v", err)
	}

	return db, mock, func() {
		_ = db.Close()
	}
}
