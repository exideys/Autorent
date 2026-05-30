package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"autorent-backend/internal/models"
)

var ErrUnavailable = errors.New("unavailable")

type RentalOrderRepository struct {
	db *sql.DB
}

func NewRentalOrderRepository(db *sql.DB) *RentalOrderRepository {
	return &RentalOrderRepository{db: db}
}

func (r *RentalOrderRepository) CreateActive(
	ctx context.Context,
	userID int64,
	input models.RentalOrderInput,
	startDate time.Time,
	endDate time.Time,
	pickupTime string,
	rentalDays int,
) (*models.RentalOrder, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer rollbackUnlessCommitted(tx)

	var carStatus string
	var pricePerDay float64
	var deposit float64
	err = tx.QueryRowContext(ctx, `
		SELECT status, price_per_day, deposit
		FROM cars
		WHERE id = ?
		FOR UPDATE
	`, input.CarID).Scan(&carStatus, &pricePerDay, &deposit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if !strings.EqualFold(carStatus, "available") {
		return nil, ErrUnavailable
	}

	totalPrice := float64(rentalDays) * pricePerDay
	result, err := tx.ExecContext(ctx, `
		INSERT INTO rental_orders (
			user_id, car_id, start_date, end_date, pickup_location, pickup_time,
			phone, notes, total_price, deposit, status
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		userID,
		input.CarID,
		startDate,
		endDate,
		input.PickupLocation,
		pickupTime,
		input.Phone,
		input.Notes,
		totalPrice,
		deposit,
		models.RentalOrderStatusActive,
	)
	if err != nil {
		return nil, err
	}

	orderID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, "UPDATE cars SET status = ? WHERE id = ?", "rented", input.CarID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, orderID)
}

func (r *RentalOrderRepository) GetByID(ctx context.Context, id int64) (*models.RentalOrder, error) {
	rows, err := r.db.QueryContext(ctx, baseRentalOrderQuery()+" WHERE ro.id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders, err := scanRentalOrders(rows)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, ErrNotFound
	}

	return &orders[0], nil
}

func (r *RentalOrderRepository) ListByUserID(ctx context.Context, userID int64) ([]models.RentalOrder, error) {
	rows, err := r.db.QueryContext(ctx, baseRentalOrderQuery()+`
		WHERE ro.user_id = ?
		ORDER BY ro.created_at DESC, ro.id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRentalOrders(rows)
}

func (r *RentalOrderRepository) DeleteActiveByCarID(ctx context.Context, carID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM rental_orders
		WHERE car_id = ? AND status = ?
	`, carID, models.RentalOrderStatusActive)
	return err
}

func baseRentalOrderQuery() string {
	return `
		SELECT
			ro.id,
			ro.user_id,
			ro.car_id,
			DATE_FORMAT(ro.start_date, '%Y-%m-%d') AS start_date,
			DATE_FORMAT(ro.end_date, '%Y-%m-%d') AS end_date,
			ro.pickup_location,
			TIME_FORMAT(ro.pickup_time, '%H:%i') AS pickup_time,
			ro.phone,
			COALESCE(ro.notes, '') AS notes,
			ro.total_price,
			ro.deposit,
			ro.status,
			ro.created_at,
			ro.updated_at,
			c.id,
			c.brand,
			c.model,
			c.year,
			c.price_per_day,
			c.deposit,
			c.status,
			(
				SELECT ci.image_url
				FROM car_images ci
				WHERE ci.car_id = c.id
				ORDER BY ci.is_main DESC, ci.sort_order ASC, ci.id ASC
				LIMIT 1
			) AS image_url
		FROM rental_orders ro
		JOIN cars c ON c.id = ro.car_id
	`
}

func scanRentalOrders(rows *sql.Rows) ([]models.RentalOrder, error) {
	orders := make([]models.RentalOrder, 0)
	for rows.Next() {
		var order models.RentalOrder
		var pickupTime sql.NullString
		var imageURL sql.NullString

		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.CarID,
			&order.StartDate,
			&order.EndDate,
			&order.PickupLocation,
			&pickupTime,
			&order.Phone,
			&order.Notes,
			&order.TotalPrice,
			&order.Deposit,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.Car.ID,
			&order.Car.Brand,
			&order.Car.Model,
			&order.Car.Year,
			&order.Car.PricePerDay,
			&order.Car.Deposit,
			&order.Car.Status,
			&imageURL,
		)
		if err != nil {
			return nil, err
		}

		if pickupTime.Valid {
			order.PickupTime = pickupTime.String
		}
		if imageURL.Valid {
			order.Car.ImageURL = &imageURL.String
		}

		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
