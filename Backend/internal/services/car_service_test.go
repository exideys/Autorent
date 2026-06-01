package services

import (
	"context"
	"testing"

	"autorent-backend/internal/models"
)

type fakeCarServiceStore struct {
	updateFunc func(ctx context.Context, id int64, input models.CarInput) (*models.Car, error)
}

func (f fakeCarServiceStore) List(context.Context, models.CarFilters) ([]models.Car, error) {
	return nil, nil
}

func (f fakeCarServiceStore) GetByID(context.Context, int64) (*models.Car, error) {
	return nil, nil
}

func (f fakeCarServiceStore) Create(context.Context, models.CarInput) (*models.Car, error) {
	return nil, nil
}

func (f fakeCarServiceStore) Update(ctx context.Context, id int64, input models.CarInput) (*models.Car, error) {
	return f.updateFunc(ctx, id, input)
}

func (f fakeCarServiceStore) Delete(context.Context, int64) error {
	return nil
}

func (f fakeCarServiceStore) AddImage(context.Context, int64, models.CarImageInput) (*models.CarImage, error) {
	return nil, nil
}

func (f fakeCarServiceStore) DeleteImage(context.Context, int64) error {
	return nil
}

type fakeRentalOrderCleaner struct {
	deletedCarID int64
}

func (f *fakeRentalOrderCleaner) DeleteActiveByCarID(_ context.Context, carID int64) error {
	f.deletedCarID = carID
	return nil
}

func TestCarServiceUpdateAvailableDeletesActiveOrders(t *testing.T) {
	cleaner := &fakeRentalOrderCleaner{}
	service := NewCarService(fakeCarServiceStore{
		updateFunc: func(_ context.Context, id int64, input models.CarInput) (*models.Car, error) {
			if id != 7 || input.Status != "available" {
				t.Fatalf("unexpected update input: id=%d input=%+v", id, input)
			}
			return &models.Car{ID: id, Status: input.Status}, nil
		},
	}, cleaner)

	car, err := service.Update(context.Background(), 7, models.CarInput{Status: "available"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if car.ID != 7 || cleaner.deletedCarID != 7 {
		t.Fatalf("expected active orders to be deleted for car 7, car=%+v deleted=%d", car, cleaner.deletedCarID)
	}
}

func TestCarServiceUpdateRentedKeepsActiveOrders(t *testing.T) {
	cleaner := &fakeRentalOrderCleaner{}
	service := NewCarService(fakeCarServiceStore{
		updateFunc: func(_ context.Context, id int64, input models.CarInput) (*models.Car, error) {
			return &models.Car{ID: id, Status: input.Status}, nil
		},
	}, cleaner)

	_, err := service.Update(context.Background(), 7, models.CarInput{Status: "rented"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cleaner.deletedCarID != 0 {
		t.Fatalf("did not expect active orders to be deleted, deleted=%d", cleaner.deletedCarID)
	}
}
