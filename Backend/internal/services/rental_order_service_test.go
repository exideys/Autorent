package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"autorent-backend/internal/models"
)

type fakeRentalOrderStore struct {
	createFunc func(ctx context.Context, userID int64, input models.RentalOrderInput, startDate time.Time, endDate time.Time, pickupTime string, rentalDays int) (*models.RentalOrder, error)
	listFunc   func(ctx context.Context, userID int64) ([]models.RentalOrder, error)
}

func (f fakeRentalOrderStore) CreateActive(ctx context.Context, userID int64, input models.RentalOrderInput, startDate time.Time, endDate time.Time, pickupTime string, rentalDays int) (*models.RentalOrder, error) {
	return f.createFunc(ctx, userID, input, startDate, endDate, pickupTime, rentalDays)
}

func (f fakeRentalOrderStore) ListByUserID(ctx context.Context, userID int64) ([]models.RentalOrder, error) {
	return f.listFunc(ctx, userID)
}

func TestRentalOrderServiceCreateNormalizesInputAndRentalDays(t *testing.T) {
	var capturedInput models.RentalOrderInput
	var capturedDays int
	store := fakeRentalOrderStore{
		createFunc: func(_ context.Context, userID int64, input models.RentalOrderInput, _ time.Time, _ time.Time, pickupTime string, rentalDays int) (*models.RentalOrder, error) {
			if userID != 42 || pickupTime != "09:30" {
				t.Fatalf("unexpected user or pickup time: userID=%d pickupTime=%s", userID, pickupTime)
			}
			capturedInput = input
			capturedDays = rentalDays
			return &models.RentalOrder{ID: 9}, nil
		},
		listFunc: func(context.Context, int64) ([]models.RentalOrder, error) {
			return nil, nil
		},
	}

	service := NewRentalOrderService(store)
	order, err := service.Create(context.Background(), 42, models.RentalOrderInput{
		CarID:          7,
		StartDate:      "2026-06-01",
		EndDate:        "2026-06-03",
		PickupLocation: " Airport ",
		PickupTime:     "09:30",
		Phone:          " +1 555 ",
		Notes:          " Child seat ",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if order.ID != 9 {
		t.Fatalf("unexpected order: %+v", order)
	}
	if capturedDays != 2 {
		t.Fatalf("expected 2 rental days, got %d", capturedDays)
	}
	if capturedInput.PickupLocation != "Airport" || capturedInput.Phone != "+1 555" || capturedInput.Notes != "Child seat" {
		t.Fatalf("input was not normalized: %+v", capturedInput)
	}
}

func TestRentalOrderServiceCreateSameDayCountsAsOneDay(t *testing.T) {
	store := fakeRentalOrderStore{
		createFunc: func(_ context.Context, _ int64, _ models.RentalOrderInput, _ time.Time, _ time.Time, _ string, rentalDays int) (*models.RentalOrder, error) {
			if rentalDays != 1 {
				t.Fatalf("expected same-day rental to count as 1 day, got %d", rentalDays)
			}
			return &models.RentalOrder{}, nil
		},
		listFunc: func(context.Context, int64) ([]models.RentalOrder, error) {
			return nil, nil
		},
	}

	service := NewRentalOrderService(store)
	_, err := service.Create(context.Background(), 1, models.RentalOrderInput{
		CarID:          2,
		StartDate:      "2026-06-01",
		EndDate:        "2026-06-01",
		PickupLocation: "Office",
		PickupTime:     "10:00",
		Phone:          "+1",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestRentalOrderServiceCreateRejectsInvalidDatesAndTime(t *testing.T) {
	service := NewRentalOrderService(fakeRentalOrderStore{
		createFunc: func(context.Context, int64, models.RentalOrderInput, time.Time, time.Time, string, int) (*models.RentalOrder, error) {
			t.Fatal("store should not be called")
			return nil, nil
		},
		listFunc: func(context.Context, int64) ([]models.RentalOrder, error) {
			return nil, nil
		},
	})

	tests := []models.RentalOrderInput{
		{CarID: 1, StartDate: "2026/06/01", EndDate: "2026-06-02", PickupLocation: "Office", PickupTime: "10:00", Phone: "+1"},
		{CarID: 1, StartDate: "2026-06-02", EndDate: "2026-06-01", PickupLocation: "Office", PickupTime: "10:00", Phone: "+1"},
		{CarID: 1, StartDate: "2026-06-01", EndDate: "2026-06-02", PickupLocation: "Office", PickupTime: "10", Phone: "+1"},
	}

	for _, input := range tests {
		_, err := service.Create(context.Background(), 1, input)
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput for %+v, got %v", input, err)
		}
	}
}
