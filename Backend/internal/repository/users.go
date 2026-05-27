package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"autorent-backend/internal/models"

	"github.com/go-sql-driver/mysql"
)

var ErrDuplicate = errors.New("duplicate")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, input models.RegisterInput, passwordHash string, role string) (*models.User, error) {
	email := normalizeEmail(input.Email)

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO users (name, email, password_hash, role)
		VALUES (?, ?, ?, ?)
	`, input.Name, email, passwordHash, role)
	if err != nil {
		if isDuplicateEntry(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, userID)
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, role, created_at
		FROM users
		WHERE id = ?
	`, id).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.UserWithPassword, error) {
	var user models.UserWithPassword
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, password_hash, role, created_at
		FROM users
		WHERE email = ?
	`, normalizeEmail(email)).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isDuplicateEntry(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
