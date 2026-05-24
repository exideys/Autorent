package repositories

import (
	"sort"
	"strings"

	"autorent-backend/internal/data"
	"autorent-backend/internal/models"
)

type MockCarRepository struct {
	cars []models.Car
}

func NewMockCarRepository() *MockCarRepository {
	return &MockCarRepository{
		cars: data.MockCars(),
	}
}

func (r *MockCarRepository) FindAvailableCars(filters models.CarSearchFilters) ([]models.Car, error) {
	//var matchedCars []models.Car
	matchedCars := make([]models.Car, 0)

	for _, car := range r.cars {
		if filters.OnlyAvailable && !strings.EqualFold(car.Status, "available") {
			continue
		}

		if filters.MinSeats > 0 && car.Seats < filters.MinSeats {
			continue
		}

		if filters.MaxPricePerDay > 0 && car.PricePerDay > filters.MaxPricePerDay {
			continue
		}

		/*if filters.PreferredBodyType != "" && !containsIgnoreCase(car.BodyType, filters.PreferredBodyType) {
			continue
		}

		if filters.PreferredFuelType != "" && !containsIgnoreCase(car.FuelType, filters.PreferredFuelType) {
			continue
		}

		if filters.PreferredClass != "" && !containsIgnoreCase(car.CarClass, filters.PreferredClass) {
			continue
		}

		if filters.Transmission != "" && !containsIgnoreCase(car.Transmission, filters.Transmission) {
			continue
		}*/

		matchedCars = append(matchedCars, car)
	}

	sort.SliceStable(matchedCars, func(i, j int) bool {
		scoreI := calculateCarScore(matchedCars[i], filters)
		scoreJ := calculateCarScore(matchedCars[j], filters)

		if scoreI == scoreJ {
			return matchedCars[i].PricePerDay < matchedCars[j].PricePerDay
		}

		return scoreI > scoreJ
	})

	return matchedCars, nil
}

func calculateCarScore(car models.Car, filters models.CarSearchFilters) int {
	score := 0

	if filters.MinSeats > 0 && car.Seats >= filters.MinSeats {
		score += 3
	}

	if filters.MaxPricePerDay > 0 && car.PricePerDay <= filters.MaxPricePerDay {
		score += 3
	}

	if filters.PreferredBodyType != "" && containsIgnoreCase(car.BodyType, filters.PreferredBodyType) {
		score += 2
	}

	if filters.PreferredFuelType != "" && containsIgnoreCase(car.FuelType, filters.PreferredFuelType) {
		score += 2
	}

	if filters.PreferredClass != "" && containsIgnoreCase(car.CarClass, filters.PreferredClass) {
		score += 1
	}

	switch strings.ToLower(filters.Purpose) {
	case "family", "сім'я", "сімʼя", "семья":
		if car.Seats >= 5 {
			score += 2
		}

		if containsIgnoreCase(car.BodyType, "SUV") || containsIgnoreCase(car.BodyType, "Liftback") {
			score += 1
		}
	case "business", "бізнес", "бизнес":
		if containsIgnoreCase(car.CarClass, "Premium") || containsIgnoreCase(car.CarClass, "Luxury") {
			score += 2
		}
	case "electric", "електро", "электро":
		if containsIgnoreCase(car.FuelType, "Electric") {
			score += 3
		}
	}

	return score
}

func containsIgnoreCase(value string, search string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(search))
}
