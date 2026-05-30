package services

import (
	"context"
	"strings"

	"autorent-backend/internal/models"
)

type CarStore interface {
	List(ctx context.Context, filters models.CarFilters) ([]models.Car, error)
	GetByID(ctx context.Context, id int64) (*models.Car, error)
	Create(ctx context.Context, input models.CarInput) (*models.Car, error)
	Update(ctx context.Context, id int64, input models.CarInput) (*models.Car, error)
	Delete(ctx context.Context, id int64) error
	AddImage(ctx context.Context, carID int64, input models.CarImageInput) (*models.CarImage, error)
	DeleteImage(ctx context.Context, imageID int64) error
}

type RentalOrderCleaner interface {
	DeleteActiveByCarID(ctx context.Context, carID int64) error
}

type CarService struct {
	cars   CarStore
	orders RentalOrderCleaner
}

func NewCarService(cars CarStore, orders RentalOrderCleaner) *CarService {
	return &CarService{
		cars:   cars,
		orders: orders,
	}
}

func (s *CarService) List(ctx context.Context, filters models.CarFilters) ([]models.Car, error) {
	return s.cars.List(ctx, filters)
}

func (s *CarService) GetByID(ctx context.Context, id int64) (*models.Car, error) {
	return s.cars.GetByID(ctx, id)
}

func (s *CarService) Create(ctx context.Context, input models.CarInput) (*models.Car, error) {
	return s.cars.Create(ctx, input)
}

func (s *CarService) Update(ctx context.Context, id int64, input models.CarInput) (*models.Car, error) {
	car, err := s.cars.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}

	if s.orders != nil && strings.EqualFold(effectiveCarStatus(input.Status), "available") {
		if err := s.orders.DeleteActiveByCarID(ctx, id); err != nil {
			return nil, err
		}
	}

	return car, nil
}

func (s *CarService) Delete(ctx context.Context, id int64) error {
	return s.cars.Delete(ctx, id)
}

func (s *CarService) AddImage(ctx context.Context, carID int64, input models.CarImageInput) (*models.CarImage, error) {
	return s.cars.AddImage(ctx, carID, input)
}

func (s *CarService) DeleteImage(ctx context.Context, imageID int64) error {
	return s.cars.DeleteImage(ctx, imageID)
}

func effectiveCarStatus(status string) string {
	if strings.TrimSpace(status) == "" {
		return "available"
	}
	return strings.TrimSpace(status)
}
