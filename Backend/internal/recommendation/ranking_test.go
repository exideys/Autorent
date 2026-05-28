package recommendation

import (
	"strings"
	"testing"

	"autorent-backend/internal/models"
)

func TestRankCarsBusinessPrefersPremiumSedans(t *testing.T) {
	vanPower := 260
	sedanPower := 340
	cars := []models.Car{
		testRankCar(1, "Mercedes-Benz", "V-Class", "Luxury", "Minivan", "Diesel", "Automatic", 5, 2023, 210, &vanPower),
		testRankCar(2, "BMW", "5 Series", "Premium", "Sedan", "Petrol", "Automatic", 5, 2022, 230, &sedanPower),
		testRankCar(3, "Audi", "A6", "Business", "Liftback", "Hybrid", "Automatic", 5, 2021, 200, &sedanPower),
	}

	ranked := RankCars(cars, models.CarRecommendationFilters{
		MinSeats:       5,
		PreferredClass: "Premium",
		Purpose:        "business",
		OnlyAvailable:  true,
	})

	if ranked[0].Brand != "BMW" {
		t.Fatalf("expected premium sedan first, got %+v", ranked[0])
	}
	if ranked[len(ranked)-1].BodyType != "Minivan" {
		t.Fatalf("expected minivan to be de-prioritized for 5-seat business request, got %+v", ranked)
	}
}

func TestRankCarsFamilyPrefersRoomyAutomaticSUVs(t *testing.T) {
	suvPower := 213
	sedanPower := 190
	cars := []models.Car{
		testRankCar(1, "Toyota", "Camry", "Comfort", "Sedan", "Hybrid", "Automatic", 5, 2023, 120, &sedanPower),
		testRankCar(2, "Nissan", "X-Trail", "Comfort", "SUV", "Hybrid", "Automatic", 7, 2022, 100, &suvPower),
	}

	ranked := RankCars(cars, models.CarRecommendationFilters{
		MinSeats:          5,
		PreferredBodyType: "SUV",
		Purpose:           "family comfort travel",
		OnlyAvailable:     true,
	})

	if ranked[0].Brand != "Nissan" {
		t.Fatalf("expected SUV first, got %+v", ranked[0])
	}
}

func TestSortOverridesUseRequestedOrder(t *testing.T) {
	cars := []models.Car{
		testRankCar(1, "BMW", "M5", "Premium", "Sedan", "Petrol", "Automatic", 5, 2024, 250, intPtr(617)),
		testRankCar(2, "Toyota", "Corolla", "Economy", "Sedan", "Petrol", "Manual", 5, 2022, 60, intPtr(140)),
		testRankCar(3, "Tesla", "Model Y", "Electric Comfort", "SUV", "Electric", "Automatic", 5, 2023, 180, intPtr(450)),
	}

	cheapest := RankCars(cars, models.CarRecommendationFilters{SortBy: "price_asc"})
	if cheapest[0].Brand != "Toyota" {
		t.Fatalf("expected cheapest car first, got %+v", cheapest[0])
	}

	mostPowerful := RankCars(cars, models.CarRecommendationFilters{SortBy: "horsepower_desc"})
	if mostPowerful[0].Brand != "BMW" {
		t.Fatalf("expected most powerful car first, got %+v", mostPowerful[0])
	}
}

func TestBuildAnswerUsesLanguageAndNoMatchText(t *testing.T) {
	ukrainian := BuildAnswer("мені потрібна найдешевша машина", 101, 5, "price_asc")
	if !strings.Contains(ukrainian, "101 доступний автомобіль") || !strings.Contains(ukrainian, "5 найдешевших варіантів") {
		t.Fatalf("unexpected Ukrainian answer: %q", ukrainian)
	}

	english := BuildAnswer("premium business car", 70, 5, "relevance")
	if english != "We found 70 options that match the basic criteria. Below are the top 5 recommendations." {
		t.Fatalf("unexpected English answer: %q", english)
	}

	noMatches := BuildAnswer("Tesla under 20 dollars", 0, 0, "relevance")
	if noMatches != "Unfortunately, we did not find cars that match your criteria." {
		t.Fatalf("unexpected no-match answer: %q", noMatches)
	}
}

func testRankCar(id int64, brand string, model string, class string, bodyType string, fuelType string, transmission string, seats int, year int, price float64, horsepower *int) models.Car {
	return models.Car{
		ID:           id,
		Brand:        brand,
		Model:        model,
		Year:         year,
		CarClass:     class,
		BodyType:     bodyType,
		Transmission: transmission,
		FuelType:     fuelType,
		Seats:        seats,
		Doors:        4,
		Horsepower:   horsepower,
		PricePerDay:  price,
		Deposit:      500,
		Status:       "available",
	}
}

func intPtr(value int) *int {
	return &value
}
