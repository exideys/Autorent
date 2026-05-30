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
	firstName, lastName := input.NameParts()

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO users (first_name, last_name, email, password_hash, role)
		VALUES (?, ?, ?, ?, ?)
	`, firstName, lastName, email, passwordHash, role)
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
		SELECT id, first_name, last_name, TRIM(CONCAT_WS(' ', first_name, last_name)) AS name,
			email, rating, rating_count, role, status, created_at, updated_at
		FROM users
		WHERE id = ?
	`, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Name,
		&user.Email,
		&user.Rating,
		&user.RatingCount,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
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
		SELECT id, first_name, last_name, TRIM(CONCAT_WS(' ', first_name, last_name)) AS name,
			email, password_hash, rating, rating_count, role, status, created_at, updated_at
		FROM users
		WHERE email = ?
	`, normalizeEmail(email)).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Rating,
		&user.RatingCount,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) ListCustomers(ctx context.Context) ([]models.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, first_name, last_name, TRIM(CONCAT_WS(' ', first_name, last_name)) AS name,
			email, rating, rating_count, role, status, created_at, updated_at
		FROM users
		WHERE role = ?
		ORDER BY created_at DESC, id DESC
	`, models.UserRoleUser)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Name,
			&user.Email,
			&user.Rating,
			&user.RatingCount,
			&user.Role,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) RateCustomer(ctx context.Context, id int64, rating float64) (*models.User, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET
			rating = ROUND(((rating * rating_count) + ?) / (rating_count + 1), 1),
			rating_count = rating_count + 1
		WHERE id = ? AND role = ?
	`, rating, id, models.UserRoleUser)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, ErrNotFound
	}

	return r.GetByID(ctx, id)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isDuplicateEntry(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
