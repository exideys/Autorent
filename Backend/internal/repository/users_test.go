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
	mock.ExpectExec("INSERT INTO users").
		WithArgs("Test", "User", "user@example.com", "hashed-password", models.UserRoleUser).
		WillReturnResult(sqlmock.NewResult(7, 1))
	mock.ExpectQuery("SELECT id, TRIM").
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "role", "created_at"}).
			AddRow(7, "Test User", "user@example.com", models.UserRoleUser, now))

	repo := NewUserRepository(db)
	user, err := repo.Create(context.Background(), models.RegisterInput{
		FirstName: "Test",
		LastName:  "User",
		Email:     " USER@EXAMPLE.COM ",
	}, "hashed-password", models.UserRoleUser)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if user.ID != 7 || user.Email != "user@example.com" {
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
	mock.ExpectQuery("SELECT id, TRIM").
		WithArgs("user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password_hash", "role", "created_at"}).
			AddRow(5, "Test User", "user@example.com", "hash", models.UserRoleAdmin, now))

	repo := NewUserRepository(db)
	user, err := repo.GetByEmail(context.Background(), " USER@EXAMPLE.COM ")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if user.ID != 5 || user.PasswordHash != "hash" || user.Role != models.UserRoleAdmin {
		t.Fatalf("unexpected user: %+v", user)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUserRepositoryGetByEmailNotFound(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectQuery("SELECT id, TRIM").
		WithArgs("missing@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password_hash", "role", "created_at"}))

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

	mock.ExpectQuery("SELECT id, TRIM").
		WithArgs(int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "role", "created_at"}))

	repo := NewUserRepository(db)
	_, err := repo.GetByID(context.Background(), 99)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
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
