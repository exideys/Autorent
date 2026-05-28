package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"autorent-backend/internal/models"
)

var ErrNotFound = errors.New("not found")

type CarRepository struct {
	db *sql.DB
}

func NewCarRepository(db *sql.DB) *CarRepository {
	return &CarRepository{db: db}
}

func (r *CarRepository) List(ctx context.Context, filters models.CarFilters) ([]models.Car, error) {
	query := baseCarQuery()
	args := make([]any, 0)
	conditions := make([]string, 0)

	if filters.Status != "" {
		conditions = append(conditions, "c.status = ?")
		args = append(args, filters.Status)
	}
	if filters.CarClass != "" {
		conditions = append(conditions, "c.car_class = ?")
		args = append(args, filters.CarClass)
	}
	if filters.BodyType != "" {
		conditions = append(conditions, "c.body_type = ?")
		args = append(args, filters.BodyType)
	}
	if filters.Transmission != "" {
		conditions = append(conditions, "c.transmission = ?")
		args = append(args, filters.Transmission)
	}
	if filters.FuelType != "" {
		conditions = append(conditions, "c.fuel_type = ?")
		args = append(args, filters.FuelType)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += carOrderBy(filters.SortBy, filters.SortOrder)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCars(rows)
}

func (r *CarRepository) GetByID(ctx context.Context, id int64) (*models.Car, error) {
	rows, err := r.db.QueryContext(ctx, baseCarQuery()+" WHERE c.id = ?"+carOrderBy("", ""), id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cars, err := scanCars(rows)
	if err != nil {
		return nil, err
	}
	if len(cars) == 0 {
		return nil, ErrNotFound
	}

	return &cars[0], nil
}

func (r *CarRepository) Create(ctx context.Context, input models.CarInput) (*models.Car, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer rollbackUnlessCommitted(tx)

	result, err := tx.ExecContext(ctx, `
		INSERT INTO cars (
			brand, model, year, car_class, body_type, transmission, fuel_type,
			seats, doors, engine_volume, horsepower, price_per_day, deposit, color, status
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		input.Brand,
		input.Model,
		input.Year,
		input.CarClass,
		input.BodyType,
		input.Transmission,
		input.FuelType,
		input.Seats,
		input.Doors,
		floatPtrValue(input.EngineVolume),
		intPtrValue(input.Horsepower),
		input.PricePerDay,
		input.Deposit,
		stringPtrValue(input.Color),
		statusOrDefault(input.Status),
	)
	if err != nil {
		return nil, err
	}

	carID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	if err := replaceImages(ctx, tx, carID, input.Images); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, carID)
}

func (r *CarRepository) Update(ctx context.Context, id int64, input models.CarInput) (*models.Car, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer rollbackUnlessCommitted(tx)

	result, err := tx.ExecContext(ctx, `
		UPDATE cars
		SET
			brand = ?,
			model = ?,
			year = ?,
			car_class = ?,
			body_type = ?,
			transmission = ?,
			fuel_type = ?,
			seats = ?,
			doors = ?,
			engine_volume = ?,
			horsepower = ?,
			price_per_day = ?,
			deposit = ?,
			color = ?,
			status = ?
		WHERE id = ?
	`,
		input.Brand,
		input.Model,
		input.Year,
		input.CarClass,
		input.BodyType,
		input.Transmission,
		input.FuelType,
		input.Seats,
		input.Doors,
		floatPtrValue(input.EngineVolume),
		intPtrValue(input.Horsepower),
		input.PricePerDay,
		input.Deposit,
		stringPtrValue(input.Color),
		statusOrDefault(input.Status),
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
		exists, err := carExists(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrNotFound
		}
	}

	if input.Images != nil {
		if err := replaceImages(ctx, tx, id, input.Images); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *CarRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM cars WHERE id = ?", id)
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

func (r *CarRepository) AddImage(ctx context.Context, carID int64, input models.CarImageInput) (*models.CarImage, error) {
	var exists int
	if err := r.db.QueryRowContext(ctx, "SELECT 1 FROM cars WHERE id = ?", carID).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO car_images (car_id, image_url, is_main, sort_order)
		VALUES (?, ?, ?, ?)
	`, carID, input.ImageURL, input.IsMain, input.SortOrder)
	if err != nil {
		return nil, err
	}

	imageID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetImageByID(ctx, imageID)
}

func (r *CarRepository) GetImageByID(ctx context.Context, id int64) (*models.CarImage, error) {
	var image models.CarImage
	err := r.db.QueryRowContext(ctx, `
		SELECT id, car_id, image_url, is_main, sort_order
		FROM car_images
		WHERE id = ?
	`, id).Scan(&image.ID, &image.CarID, &image.ImageURL, &image.IsMain, &image.SortOrder)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &image, nil
}

func (r *CarRepository) DeleteImage(ctx context.Context, imageID int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM car_images WHERE id = ?", imageID)
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

func baseCarQuery() string {
	return `
		SELECT
			c.id,
			c.brand,
			c.model,
			c.year,
			c.car_class,
			c.body_type,
			c.transmission,
			c.fuel_type,
			c.seats,
			c.doors,
			c.engine_volume,
			c.horsepower,
			c.price_per_day,
			c.deposit,
			c.color,
			c.status,
			c.created_at,
			ci.id,
			ci.car_id,
			ci.image_url,
			ci.is_main,
			ci.sort_order
		FROM cars c
		LEFT JOIN car_images ci ON ci.car_id = c.id
	`
}

func carOrderBy(sortBy string, sortOrder string) string {
	column := allowedCarSortColumns()[sortBy]
	if column == "" {
		column = "c.created_at"
	}

	direction := "DESC"
	if strings.EqualFold(sortOrder, "asc") {
		direction = "ASC"
	}

	return fmt.Sprintf(
		" ORDER BY %s %s, c.id DESC, ci.is_main DESC, ci.sort_order ASC, ci.id ASC",
		column,
		direction,
	)
}

func allowedCarSortColumns() map[string]string {
	return map[string]string{
		"brand":         "c.brand",
		"model":         "c.model",
		"year":          "c.year",
		"car_class":     "c.car_class",
		"body_type":     "c.body_type",
		"transmission":  "c.transmission",
		"fuel_type":     "c.fuel_type",
		"seats":         "c.seats",
		"doors":         "c.doors",
		"horsepower":    "c.horsepower",
		"price_per_day": "c.price_per_day",
		"deposit":       "c.deposit",
		"status":        "c.status",
		"created_at":    "c.created_at",
	}
}

func scanCars(rows *sql.Rows) ([]models.Car, error) {
	cars := make([]models.Car, 0)
	carIndexes := make(map[int64]int)

	for rows.Next() {
		var car models.Car
		var engineVolume sql.NullFloat64
		var horsepower sql.NullInt64
		var color sql.NullString
		var imageID sql.NullInt64
		var imageCarID sql.NullInt64
		var imageURL sql.NullString
		var imageIsMain sql.NullBool
		var imageSortOrder sql.NullInt64

		err := rows.Scan(
			&car.ID,
			&car.Brand,
			&car.Model,
			&car.Year,
			&car.CarClass,
			&car.BodyType,
			&car.Transmission,
			&car.FuelType,
			&car.Seats,
			&car.Doors,
			&engineVolume,
			&horsepower,
			&car.PricePerDay,
			&car.Deposit,
			&color,
			&car.Status,
			&car.CreatedAt,
			&imageID,
			&imageCarID,
			&imageURL,
			&imageIsMain,
			&imageSortOrder,
		)
		if err != nil {
			return nil, err
		}

		if engineVolume.Valid {
			car.EngineVolume = &engineVolume.Float64
		}
		if horsepower.Valid {
			value := int(horsepower.Int64)
			car.Horsepower = &value
		}
		if color.Valid {
			car.Color = &color.String
		}
		car.Images = []models.CarImage{}

		carIndex, ok := carIndexes[car.ID]
		if !ok {
			cars = append(cars, car)
			carIndex = len(cars) - 1
			carIndexes[car.ID] = carIndex
		}

		if imageID.Valid {
			cars[carIndex].Images = append(cars[carIndex].Images, models.CarImage{
				ID:        imageID.Int64,
				CarID:     imageCarID.Int64,
				ImageURL:  imageURL.String,
				IsMain:    imageIsMain.Valid && imageIsMain.Bool,
				SortOrder: int(imageSortOrder.Int64),
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cars, nil
}

func replaceImages(ctx context.Context, tx *sql.Tx, carID int64, images []models.CarImageInput) error {
	if _, err := tx.ExecContext(ctx, "DELETE FROM car_images WHERE car_id = ?", carID); err != nil {
		return err
	}

	for index, image := range images {
		sortOrder := image.SortOrder
		if sortOrder == 0 {
			sortOrder = index
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO car_images (car_id, image_url, is_main, sort_order)
			VALUES (?, ?, ?, ?)
		`, carID, image.ImageURL, image.IsMain, sortOrder); err != nil {
			return fmt.Errorf("insert car image: %w", err)
		}
	}

	return nil
}

func carExists(ctx context.Context, tx *sql.Tx, carID int64) (bool, error) {
	var exists int
	if err := tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM cars WHERE id = ?)", carID).Scan(&exists); err != nil {
		return false, err
	}

	return exists == 1, nil
}

func rollbackUnlessCommitted(tx *sql.Tx) {
	_ = tx.Rollback()
}

func statusOrDefault(status string) string {
	if status == "" {
		return "available"
	}
	return status
}

func floatPtrValue(value *float64) any {
	if value == nil {
		return nil
	}
	return *value
}

func intPtrValue(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func stringPtrValue(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
