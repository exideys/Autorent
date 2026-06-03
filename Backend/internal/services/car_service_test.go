package services

import (
	"context"
	"testing"

	"autorent-backend/internal/models"
)

type fakeCarServiceStore struct {
	listFunc        func(ctx context.Context, filters models.CarFilters) ([]models.Car, error)
	getFunc         func(ctx context.Context, id int64) (*models.Car, error)
	createFunc      func(ctx context.Context, input models.CarInput) (*models.Car, error)
	updateFunc      func(ctx context.Context, id int64, input models.CarInput) (*models.Car, error)
	deleteFunc      func(ctx context.Context, id int64) error
	addImageFunc    func(ctx context.Context, carID int64, input models.CarImageInput) (*models.CarImage, error)
	deleteImageFunc func(ctx context.Context, imageID int64) error
}

func (f fakeCarServiceStore) List(ctx context.Context, filters models.CarFilters) ([]models.Car, error) {
	if f.listFunc != nil {
		return f.listFunc(ctx, filters)
	}
	return nil, nil
}

func (f fakeCarServiceStore) GetByID(ctx context.Context, id int64) (*models.Car, error) {
	if f.getFunc != nil {
		return f.getFunc(ctx, id)
	}
	return nil, nil
}

func (f fakeCarServiceStore) Create(ctx context.Context, input models.CarInput) (*models.Car, error) {
	if f.createFunc != nil {
		return f.createFunc(ctx, input)
	}
	return nil, nil
}

func (f fakeCarServiceStore) Update(ctx context.Context, id int64, input models.CarInput) (*models.Car, error) {
	if f.updateFunc == nil {
		return nil, nil
	}
	return f.updateFunc(ctx, id, input)
}

func (f fakeCarServiceStore) Delete(ctx context.Context, id int64) error {
	if f.deleteFunc != nil {
		return f.deleteFunc(ctx, id)
	}
	return nil
}

func (f fakeCarServiceStore) AddImage(ctx context.Context, carID int64, input models.CarImageInput) (*models.CarImage, error) {
	if f.addImageFunc != nil {
		return f.addImageFunc(ctx, carID, input)
	}
	return nil, nil
}

func (f fakeCarServiceStore) DeleteImage(ctx context.Context, imageID int64) error {
	if f.deleteImageFunc != nil {
		return f.deleteImageFunc(ctx, imageID)
	}
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

func TestCarServiceDelegatesReadAndCreateMethods(t *testing.T) {
	service := NewCarService(fakeCarServiceStore{
		listFunc: func(_ context.Context, filters models.CarFilters) ([]models.Car, error) {
			if filters.CarClass != "Business" {
				t.Fatalf("unexpected filters: %+v", filters)
			}
			return []models.Car{{ID: 1}}, nil
		},
		getFunc: func(_ context.Context, id int64) (*models.Car, error) {
			if id != 2 {
				t.Fatalf("expected id 2, got %d", id)
			}
			return &models.Car{ID: id}, nil
		},
		createFunc: func(_ context.Context, input models.CarInput) (*models.Car, error) {
			if input.Brand != "BMW" {
				t.Fatalf("unexpected input: %+v", input)
			}
			return &models.Car{ID: 3, Brand: input.Brand}, nil
		},
	}, nil)

	cars, err := service.List(context.Background(), models.CarFilters{CarClass: "Business"})
	if err != nil || len(cars) != 1 || cars[0].ID != 1 {
		t.Fatalf("unexpected list result: cars=%+v err=%v", cars, err)
	}
	car, err := service.GetByID(context.Background(), 2)
	if err != nil || car.ID != 2 {
		t.Fatalf("unexpected get result: car=%+v err=%v", car, err)
	}
	car, err = service.Create(context.Background(), models.CarInput{Brand: "BMW"})
	if err != nil || car.ID != 3 {
		t.Fatalf("unexpected create result: car=%+v err=%v", car, err)
	}
}

func TestCarServiceDelegatesDeleteAndImageMethods(t *testing.T) {
	var deletedCarID int64
	var addedCarID int64
	var deletedImageID int64
	service := NewCarService(fakeCarServiceStore{
		deleteFunc: func(_ context.Context, id int64) error {
			deletedCarID = id
			return nil
		},
		addImageFunc: func(_ context.Context, carID int64, input models.CarImageInput) (*models.CarImage, error) {
			addedCarID = carID
			return &models.CarImage{ID: 9, CarID: carID, ImageURL: input.ImageURL}, nil
		},
		deleteImageFunc: func(_ context.Context, imageID int64) error {
			deletedImageID = imageID
			return nil
		},
	}, nil)

	if err := service.Delete(context.Background(), 4); err != nil {
		t.Fatalf("expected nil delete error, got %v", err)
	}
	image, err := service.AddImage(context.Background(), 5, models.CarImageInput{ImageURL: "https://example.com/car.jpg"})
	if err != nil {
		t.Fatalf("expected nil add image error, got %v", err)
	}
	if err := service.DeleteImage(context.Background(), 9); err != nil {
		t.Fatalf("expected nil delete image error, got %v", err)
	}
	if deletedCarID != 4 || addedCarID != 5 || image.ID != 9 || deletedImageID != 9 {
		t.Fatalf("unexpected calls: deletedCarID=%d addedCarID=%d image=%+v deletedImageID=%d", deletedCarID, addedCarID, image, deletedImageID)
	}
}
