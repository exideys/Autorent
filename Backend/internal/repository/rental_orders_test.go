package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"autorent-backend/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRentalOrderRepositoryCreateActiveCreatesOrderAndRentsCar(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	startDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC)
	createdAt := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT status, price_per_day, deposit").
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "price_per_day", "deposit"}).
			AddRow("available", 250.00, 1000.00))
	mock.ExpectExec("INSERT INTO rental_orders").
		WithArgs(
			int64(42),
			int64(7),
			startDate,
			endDate,
			"Airport",
			"09:30",
			"+1 555 0100",
			"Child seat",
			500.00,
			1000.00,
			models.RentalOrderStatusActive,
		).
		WillReturnResult(sqlmock.NewResult(99, 1))
	mock.ExpectExec("UPDATE cars SET status").
		WithArgs("rented", int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery("SELECT").
		WithArgs(int64(99)).
		WillReturnRows(rentalOrderRows().AddRow(
			99,
			42,
			7,
			"2026-06-01",
			"2026-06-03",
			"Airport",
			"09:30",
			"+1 555 0100",
			"Child seat",
			500.00,
			1000.00,
			models.RentalOrderStatusActive,
			createdAt,
			createdAt,
			7,
			"BMW",
			"M5",
			2024,
			250.00,
			1000.00,
			"rented",
			"https://example.com/bmw.jpg",
		))

	repo := NewRentalOrderRepository(db)
	order, err := repo.CreateActive(context.Background(), 42, models.RentalOrderInput{
		CarID:          7,
		PickupLocation: "Airport",
		Phone:          "+1 555 0100",
		Notes:          "Child seat",
	}, startDate, endDate, "09:30", 2)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if order.ID != 99 || order.TotalPrice != 500 || order.Car.Status != "rented" || order.Car.ImageURL == nil {
		t.Fatalf("unexpected order: %+v", order)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRentalOrderRepositoryCreateActiveRejectsUnavailableCar(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	startDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT status, price_per_day, deposit").
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "price_per_day", "deposit"}).
			AddRow("rented", 250.00, 1000.00))
	mock.ExpectRollback()

	repo := NewRentalOrderRepository(db)
	_, err := repo.CreateActive(context.Background(), 42, models.RentalOrderInput{CarID: 7}, startDate, endDate, "09:30", 1)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRentalOrderRepositoryListByUserID(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	createdAt := time.Now()
	mock.ExpectQuery("WHERE ro.user_id = \\?").
		WithArgs(int64(42)).
		WillReturnRows(rentalOrderRows().AddRow(
			100,
			42,
			8,
			"2026-07-01",
			"2026-07-02",
			"Office",
			"11:00",
			"+1",
			"",
			180.00,
			500.00,
			models.RentalOrderStatusActive,
			createdAt,
			createdAt,
			8,
			"Tesla",
			"Model 3",
			2024,
			180.00,
			500.00,
			"rented",
			nil,
		))

	repo := NewRentalOrderRepository(db)
	orders, err := repo.ListByUserID(context.Background(), 42)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(orders) != 1 || orders[0].UserID != 42 || orders[0].Car.ImageURL != nil {
		t.Fatalf("unexpected orders: %+v", orders)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRentalOrderRepositoryDeleteActiveByCarID(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectExec("DELETE FROM rental_orders").
		WithArgs(int64(7), models.RentalOrderStatusActive).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewRentalOrderRepository(db)
	if err := repo.DeleteActiveByCarID(context.Background(), 7); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func rentalOrderRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"user_id",
		"car_id",
		"start_date",
		"end_date",
		"pickup_location",
		"pickup_time",
		"phone",
		"notes",
		"total_price",
		"deposit",
		"status",
		"created_at",
		"updated_at",
		"car_id",
		"brand",
		"model",
		"year",
		"price_per_day",
		"deposit",
		"status",
		"image_url",
	})
}
