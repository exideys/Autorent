package repositories

import (
	"sort"
	"strings"

	"autorent-backend/internal/models"
)

func calculateCarScore(car models.Car, filters models.CarSearchFilters) int {
	score := 0
	purpose := strings.ToLower(strings.TrimSpace(filters.Purpose))

	if filters.MinSeats > 0 && car.Seats >= filters.MinSeats {
		score += 3
	}

	if filters.MaxPricePerDay > 0 && car.PricePerDay <= filters.MaxPricePerDay {
		score += 3
	}

	if filters.MinSeats >= 5 {
		if car.Seats > filters.MinSeats {
			score += 1
		}

		if purpose == "family" || purpose == "comfort" || purpose == "travel" || purpose == "сім'я" || purpose == "сімʼя" || purpose == "родина" {
			if isLargeFamilyBodyType(car) {
				score += 3
			}
		}

		if purpose == "business" || purpose == "бізнес" {
			if isBusinessBodyType(car) {
				score += 3
			}

			if isLargeFamilyBodyType(car) && filters.MinSeats <= 5 {
				score -= 2
			}
		}

		if containsIgnoreCase(car.Transmission, "Automatic") {
			score += 1
		}
	}

	if filters.MaxPricePerDay >= 150 &&
		!containsIgnoreCase(filters.PreferredClass, "Economy") &&
		!containsIgnoreCase(filters.Purpose, "economy") &&
		!containsIgnoreCase(car.CarClass, "Economy") {
		score += 1
	}

	if filters.PreferredBodyType != "" && containsIgnoreCase(car.BodyType, filters.PreferredBodyType) {
		score += 2
	}

	if filters.PreferredFuelType != "" && containsIgnoreCase(car.FuelType, filters.PreferredFuelType) {
		score += 2
	}

	if filters.PreferredClass != "" && containsIgnoreCase(car.CarClass, filters.PreferredClass) {
		score += 4
	}

	if filters.Transmission != "" && containsIgnoreCase(car.Transmission, filters.Transmission) {
		score += 1
	}

	switch purpose {
	case "family", "сім'я", "сімʼя", "родина":
		if car.Seats >= 5 {
			score += 2
		}

		if isLargeFamilyBodyType(car) || containsIgnoreCase(car.BodyType, "Liftback") {
			score += 2
		}

	case "business", "бізнес":
		if isPremiumOrLuxury(car) {
			score += 5
		}

		if isBusinessBodyType(car) {
			score += 3
		}

		if isLargeFamilyBodyType(car) && filters.MinSeats <= 5 {
			score -= 2
		}

	case "electric", "електро":
		if containsIgnoreCase(car.FuelType, "Electric") {
			score += 3
		}
	}

	return score
}

func containsIgnoreCase(value string, search string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(search))
}

func isLargeFamilyBodyType(car models.Car) bool {
	return containsIgnoreCase(car.BodyType, "SUV") ||
		containsIgnoreCase(car.BodyType, "Van") ||
		containsIgnoreCase(car.BodyType, "Minivan")
}

func isBusinessBodyType(car models.Car) bool {
	return containsIgnoreCase(car.BodyType, "Sedan") ||
		containsIgnoreCase(car.BodyType, "Liftback")
}

func isPremiumOrLuxury(car models.Car) bool {
	return containsIgnoreCase(car.CarClass, "Premium") ||
		containsIgnoreCase(car.CarClass, "Luxury")
}

func sortCarsByFilters(cars []models.Car, filters models.CarSearchFilters) {
	sort.SliceStable(cars, func(i, j int) bool {
		switch strings.ToLower(strings.TrimSpace(filters.SortBy)) {
		case "price_asc":
			return cars[i].PricePerDay < cars[j].PricePerDay
		case "price_desc":
			return cars[i].PricePerDay > cars[j].PricePerDay
		case "year_desc":
			return cars[i].Year > cars[j].Year
		case "horsepower_desc":
			return cars[i].Horsepower > cars[j].Horsepower
		case "seats_desc":
			return cars[i].Seats > cars[j].Seats
		default:
			scoreI := calculateCarScore(cars[i], filters)
			scoreJ := calculateCarScore(cars[j], filters)

			if scoreI == scoreJ {
				return cars[i].PricePerDay < cars[j].PricePerDay
			}

			return scoreI > scoreJ
		}
	})
}
