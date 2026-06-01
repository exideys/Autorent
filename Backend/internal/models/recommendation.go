package models

import "time"

type CarRecommendationFilters struct {
	MinSeats          int     `json:"min_seats,omitempty"`
	MaxPricePerDay    float64 `json:"max_price_per_day,omitempty"`
	PreferredBrand    string  `json:"preferred_brand,omitempty"`
	PreferredBodyType string  `json:"preferred_body_type,omitempty"`
	PreferredFuelType string  `json:"preferred_fuel_type,omitempty"`
	PreferredClass    string  `json:"preferred_class,omitempty"`
	Transmission      string  `json:"transmission,omitempty"`
	Purpose           string  `json:"purpose,omitempty"`
	SortBy            string  `json:"sort_by,omitempty"`
	OnlyAvailable     bool    `json:"only_available"`
}

type CarRecommendationRequest struct {
	Message string `json:"message"`
}

type RecommendedCar struct {
	ID           int64      `json:"id"`
	Brand        string     `json:"brand"`
	Model        string     `json:"model"`
	Year         int        `json:"year"`
	CarClass     string     `json:"car_class"`
	BodyType     string     `json:"body_type"`
	Transmission string     `json:"transmission"`
	FuelType     string     `json:"fuel_type"`
	Seats        int        `json:"seats"`
	Doors        int        `json:"doors"`
	EngineVolume *float64   `json:"engine_volume,omitempty"`
	Horsepower   int        `json:"horsepower"`
	PricePerDay  float64    `json:"price_per_day"`
	Deposit      float64    `json:"deposit"`
	Color        string     `json:"color"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	Images       []CarImage `json:"images"`
}

type CarRecommendationResponse struct {
	Answer       string           `json:"answer"`
	Cars         []RecommendedCar `json:"cars"`
	TotalMatches int              `json:"total_matches"`
}

func ToRecommendedCar(car Car) RecommendedCar {
	recommended := RecommendedCar{
		ID:           car.ID,
		Brand:        car.Brand,
		Model:        car.Model,
		Year:         car.Year,
		CarClass:     car.CarClass,
		BodyType:     car.BodyType,
		Transmission: car.Transmission,
		FuelType:     car.FuelType,
		Seats:        car.Seats,
		Doors:        car.Doors,
		EngineVolume: car.EngineVolume,
		PricePerDay:  car.PricePerDay,
		Deposit:      car.Deposit,
		Status:       car.Status,
		CreatedAt:    car.CreatedAt,
		Images:       car.Images,
	}

	if car.Horsepower != nil {
		recommended.Horsepower = *car.Horsepower
	}
	if car.Color != nil {
		recommended.Color = *car.Color
	}

	return recommended
}
