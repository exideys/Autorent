package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"autorent-backend/internal/models"
)

var ErrInvalidInput = errors.New("invalid input")

type RentalOrderStore interface {
	CreateActive(
		ctx context.Context,
		userID int64,
		input models.RentalOrderInput,
		startDate time.Time,
		endDate time.Time,
		pickupTime string,
		rentalDays int,
	) (*models.RentalOrder, error)
	ListByUserID(ctx context.Context, userID int64) ([]models.RentalOrder, error)
}

type RentalOrderService struct {
	store RentalOrderStore
}

func NewRentalOrderService(store RentalOrderStore) *RentalOrderService {
	return &RentalOrderService{store: store}
}

func (s *RentalOrderService) Create(ctx context.Context, userID int64, input models.RentalOrderInput) (*models.RentalOrder, error) {
	if userID <= 0 || input.CarID <= 0 {
		return nil, ErrInvalidInput
	}

	input.PickupLocation = strings.TrimSpace(input.PickupLocation)
	input.PickupTime = strings.TrimSpace(input.PickupTime)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Notes = strings.TrimSpace(input.Notes)
	if input.PickupLocation == "" || input.Phone == "" {
		return nil, ErrInvalidInput
	}

	startDate, err := time.Parse("2006-01-02", strings.TrimSpace(input.StartDate))
	if err != nil {
		return nil, ErrInvalidInput
	}
	endDate, err := time.Parse("2006-01-02", strings.TrimSpace(input.EndDate))
	if err != nil {
		return nil, ErrInvalidInput
	}
	if endDate.Before(startDate) {
		return nil, ErrInvalidInput
	}

	parsedPickupTime, err := time.Parse("15:04", input.PickupTime)
	if err != nil {
		return nil, ErrInvalidInput
	}
	input.StartDate = startDate.Format("2006-01-02")
	input.EndDate = endDate.Format("2006-01-02")
	input.PickupTime = parsedPickupTime.Format("15:04")

	rentalDays := int(endDate.Sub(startDate).Hours() / 24)
	if rentalDays < 1 {
		rentalDays = 1
	}

	return s.store.CreateActive(ctx, userID, input, startDate, endDate, input.PickupTime, rentalDays)
}

func (s *RentalOrderService) ListByUserID(ctx context.Context, userID int64) ([]models.RentalOrder, error) {
	if userID <= 0 {
		return nil, ErrInvalidInput
	}

	return s.store.ListByUserID(ctx, userID)
}
