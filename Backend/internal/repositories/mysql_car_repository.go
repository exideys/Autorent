package repositories

import (
	"database/sql"
	"strings"

	"autorent-backend/internal/models"
)

type MySQLCarRepository struct {
	db *sql.DB
}

func NewMySQLCarRepository(db *sql.DB) *MySQLCarRepository {
	return &MySQLCarRepository{
		db: db,
	}
}

func (r *MySQLCarRepository) FindAvailableCars(filters models.CarSearchFilters) ([]models.Car, error) {
	query := `
		SELECT
			id,
			brand,
			model,
			year,
			car_class,
			body_type,
			transmission,
			fuel_type,
			seats,
			doors,
			engine_volume,
			horsepower,
			price_per_day,
			deposit,
			color,
			status
		FROM cars
	`

	conditions := make([]string, 0)
	args := make([]any, 0)

	if filters.OnlyAvailable {
		conditions = append(conditions, "status = ?")
		args = append(args, "available")
	}

	if filters.MinSeats > 0 {
		conditions = append(conditions, "seats >= ?")
		args = append(args, filters.MinSeats)
	}

	if filters.MaxPricePerDay > 0 {
		conditions = append(conditions, "price_per_day <= ?")
		args = append(args, filters.MaxPricePerDay)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY price_per_day ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cars := make([]models.Car, 0)

	for rows.Next() {
		var car models.Car
		var engineVolume sql.NullFloat64
		var horsepower sql.NullInt64
		var color sql.NullString
		var status sql.NullString

		if err := rows.Scan(
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
			&status,
		); err != nil {
			return nil, err
		}

		if engineVolume.Valid {
			car.EngineVolume = &engineVolume.Float64
		} else {
			car.EngineVolume = nil
		}

		if horsepower.Valid {
			value := int(horsepower.Int64)
			car.Horsepower = &value
		}

		if color.Valid {
			value := color.String
			car.Color = &value
		}

		if status.Valid {
			car.Status = status.String
		}

		cars = append(cars, car)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	sortCarsByFilters(cars, filters)

	return cars, nil
}
