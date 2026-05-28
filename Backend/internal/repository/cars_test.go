package repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"autorent-backend/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCarRepositoryListAppliesFiltersAndScansImages(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	createdAt := time.Now()
	engineVolume := 4.4
	horsepower := 617
	color := "Black"

	mock.ExpectQuery("SELECT").
		WithArgs("available", "Business").
		WillReturnRows(carRows().
			AddRow(
				1,
				"BMW",
				"M5",
				2024,
				"Business",
				"Sedan",
				"Automatic",
				"Petrol",
				5,
				4,
				engineVolume,
				horsepower,
				250.00,
				1000.00,
				color,
				"available",
				createdAt,
				10,
				1,
				"https://example.com/bmw.jpg",
				true,
				0,
			))

	repo := NewCarRepository(db)
	cars, err := repo.List(context.Background(), models.CarFilters{
		Status:    "available",
		CarClass:  "Business",
		SortBy:    "price_per_day",
		SortOrder: "asc",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(cars) != 1 {
		t.Fatalf("expected 1 car, got %d", len(cars))
	}
	car := cars[0]
	if car.ID != 1 || car.Brand != "BMW" || car.Model != "M5" {
		t.Fatalf("unexpected car: %+v", car)
	}
	if car.EngineVolume == nil || *car.EngineVolume != engineVolume {
		t.Fatalf("unexpected engine volume: %+v", car.EngineVolume)
	}
	if car.Horsepower == nil || *car.Horsepower != horsepower {
		t.Fatalf("unexpected horsepower: %+v", car.Horsepower)
	}
	if car.Color == nil || *car.Color != color {
		t.Fatalf("unexpected color: %+v", car.Color)
	}
	if len(car.Images) != 1 || car.Images[0].ImageURL != "https://example.com/bmw.jpg" {
		t.Fatalf("unexpected images: %+v", car.Images)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestCarRepositoryGetByIDNotFound(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectQuery("SELECT").WithArgs(int64(404)).WillReturnRows(carRows())

	repo := NewCarRepository(db)
	_, err := repo.GetByID(context.Background(), 404)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestCarRepositorySearchRecommendationsAppliesHardFilters(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	createdAt := time.Now()
	mock.ExpectQuery("c.status = \\?").
		WithArgs("available", 5, 220.0, "%Tesla%").
		WillReturnRows(carRows().
			AddRow(
				7,
				"Tesla",
				"Model Y",
				2024,
				"Electric Comfort",
				"SUV",
				"Automatic",
				"Electric",
				5,
				5,
				nil,
				nil,
				210.00,
				500.00,
				nil,
				nil,
				createdAt,
				nil,
				nil,
				nil,
				nil,
				nil,
			))

	repo := NewCarRepository(db)
	cars, err := repo.SearchRecommendations(context.Background(), models.CarRecommendationFilters{
		MinSeats:       5,
		MaxPricePerDay: 220,
		PreferredBrand: "Tesla",
		OnlyAvailable:  true,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(cars) != 1 {
		t.Fatalf("expected 1 car, got %d", len(cars))
	}
	car := cars[0]
	if car.Brand != "Tesla" || car.Horsepower != nil || car.Color != nil || car.Status != "" {
		t.Fatalf("unexpected scanned car: %+v", car)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestCarRepositoryCreateWithImages(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	engineVolume := 3.0
	horsepower := 400
	color := "White"
	input := models.CarInput{
		Brand:        "Audi",
		Model:        "A6",
		Year:         2024,
		CarClass:     "Business",
		BodyType:     "Sedan",
		Transmission: "Automatic",
		FuelType:     "Petrol",
		Seats:        5,
		Doors:        4,
		EngineVolume: &engineVolume,
		Horsepower:   &horsepower,
		PricePerDay:  180,
		Deposit:      700,
		Color:        &color,
		Images: []models.CarImageInput{
			{ImageURL: "https://example.com/audi.jpg", IsMain: true},
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO cars").
		WithArgs(
			input.Brand,
			input.Model,
			input.Year,
			input.CarClass,
			input.BodyType,
			input.Transmission,
			input.FuelType,
			input.Seats,
			input.Doors,
			engineVolume,
			horsepower,
			input.PricePerDay,
			input.Deposit,
			color,
			"available",
		).
		WillReturnResult(sqlmock.NewResult(11, 1))
	mock.ExpectExec("DELETE FROM car_images").
		WithArgs(int64(11)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO car_images").
		WithArgs(int64(11), "https://example.com/audi.jpg", true, 0).
		WillReturnResult(sqlmock.NewResult(30, 1))
	mock.ExpectCommit()
	mock.ExpectQuery("SELECT").
		WithArgs(int64(11)).
		WillReturnRows(carRows().AddRow(
			11,
			"Audi",
			"A6",
			2024,
			"Business",
			"Sedan",
			"Automatic",
			"Petrol",
			5,
			4,
			engineVolume,
			horsepower,
			180.00,
			700.00,
			color,
			"available",
			time.Now(),
			nil,
			nil,
			nil,
			nil,
			nil,
		))

	repo := NewCarRepository(db)
	car, err := repo.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if car.ID != 11 {
		t.Fatalf("expected created car id 11, got %d", car.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestCarRepositoryUpdateNotFound(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	input := models.CarInput{
		Brand:        "BMW",
		Model:        "M3",
		Year:         2020,
		CarClass:     "Sport",
		BodyType:     "Sedan",
		Transmission: "Automatic",
		FuelType:     "Petrol",
		Seats:        5,
		Doors:        4,
		PricePerDay:  200,
		Deposit:      800,
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE cars").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(int64(90)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(0))
	mock.ExpectRollback()

	repo := NewCarRepository(db)
	_, err := repo.Update(context.Background(), 90, input)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestCarRepositoryAddImageCarNotFound(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectQuery("SELECT 1 FROM cars").
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"1"}))

	repo := NewCarRepository(db)
	_, err := repo.AddImage(context.Background(), 9, models.CarImageInput{
		ImageURL: "https://example.com/car.jpg",
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestCarRepositoryAddImage(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectQuery("SELECT 1 FROM cars").
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectExec("INSERT INTO car_images").
		WithArgs(int64(9), "https://example.com/car.jpg", true, 3).
		WillReturnResult(sqlmock.NewResult(15, 1))
	mock.ExpectQuery("SELECT id, car_id, image_url, is_main, sort_order").
		WithArgs(int64(15)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "car_id", "image_url", "is_main", "sort_order"}).
			AddRow(15, 9, "https://example.com/car.jpg", true, 3))

	repo := NewCarRepository(db)
	image, err := repo.AddImage(context.Background(), 9, models.CarImageInput{
		ImageURL:  "https://example.com/car.jpg",
		IsMain:    true,
		SortOrder: 3,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if image.ID != 15 || image.CarID != 9 || !image.IsMain {
		t.Fatalf("unexpected image: %+v", image)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestCarRepositoryGetImageByIDNotFound(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectQuery("SELECT id, car_id, image_url, is_main, sort_order").
		WithArgs(int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "car_id", "image_url", "is_main", "sort_order"}))

	repo := NewCarRepository(db)
	_, err := repo.GetImageByID(context.Background(), 99)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestCarRepositoryDeleteAndDeleteImage(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectExec("DELETE FROM cars").
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM car_images").
		WithArgs(int64(2)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewCarRepository(db)
	if err := repo.Delete(context.Background(), 1); err != nil {
		t.Fatalf("expected nil delete error, got %v", err)
	}
	if err := repo.DeleteImage(context.Background(), 2); err != nil {
		t.Fatalf("expected nil delete image error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestCarOrderByUsesWhitelist(t *testing.T) {
	order := carOrderBy("price_per_day", "asc")
	if !strings.Contains(order, "ORDER BY c.price_per_day ASC") {
		t.Fatalf("expected price ascending order, got %q", order)
	}

	injected := carOrderBy("price_per_day; DROP TABLE cars", "asc; DROP TABLE users")
	if strings.Contains(injected, "DROP TABLE") {
		t.Fatalf("order by contains unsafe input: %q", injected)
	}
	if !strings.Contains(injected, "ORDER BY c.created_at DESC") {
		t.Fatalf("expected default ordering for invalid input, got %q", injected)
	}
}

func carRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"brand",
		"model",
		"year",
		"car_class",
		"body_type",
		"transmission",
		"fuel_type",
		"seats",
		"doors",
		"engine_volume",
		"horsepower",
		"price_per_day",
		"deposit",
		"color",
		"status",
		"created_at",
		"id",
		"car_id",
		"image_url",
		"is_main",
		"sort_order",
	})
}
