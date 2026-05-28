package repositories

import (
	"testing"

	"autorent-backend/internal/models"
)

func TestMockCarRepositoryFindAvailableCarsBySeatsAndBudget(t *testing.T) {
	repo := NewMockCarRepository()

	cars, err := repo.FindAvailableCars(models.CarSearchFilters{
		MinSeats:       5,
		MaxPricePerDay: 240,
		OnlyAvailable:  true,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(cars) != 4 {
		t.Fatalf("expected 4 cars, got %d", len(cars))
	}

	expectedIDs := []int64{3, 2, 1, 5}
	for index, expectedID := range expectedIDs {
		if cars[index].ID != expectedID {
			t.Fatalf("expected car id %d at index %d, got %d", expectedID, index, cars[index].ID)
		}
	}
}

func TestMockCarRepositoryReturnsEmptySliceWhenNoCarsMatch(t *testing.T) {
	repo := NewMockCarRepository()

	cars, err := repo.FindAvailableCars(models.CarSearchFilters{
		MinSeats:       5,
		MaxPricePerDay: 100,
		OnlyAvailable:  true,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cars == nil {
		t.Fatal("expected empty slice, got nil")
	}

	if len(cars) != 0 {
		t.Fatalf("expected 0 cars, got %d", len(cars))
	}
}

func TestMockCarRepositoryUsesPreferencesAsSoftRanking(t *testing.T) {
	repo := NewMockCarRepository()

	cars, err := repo.FindAvailableCars(models.CarSearchFilters{
		MinSeats:          5,
		MaxPricePerDay:    240,
		PreferredBodyType: "SUV",
		OnlyAvailable:     true,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(cars) != 4 {
		t.Fatalf("expected preferences not to hard-filter cars, got %d cars", len(cars))
	}
}
