package repositories

import (
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

	/*sort.SliceStable(matchedCars, func(i, j int) bool {
		scoreI := calculateCarScore(matchedCars[i], filters)
		scoreJ := calculateCarScore(matchedCars[j], filters)

		if scoreI == scoreJ {
			return matchedCars[i].PricePerDay < matchedCars[j].PricePerDay
		}

		return scoreI > scoreJ
	})*/
	sortCarsByFilters(matchedCars, filters)

	return matchedCars, nil
}
