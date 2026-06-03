package models

import (
	"testing"
	"time"
)

func TestToRecommendedCarCopiesOptionalValues(t *testing.T) {
	engineVolume := 3.0
	horsepower := 400
	color := "White"
	createdAt := time.Now()

	recommended := ToRecommendedCar(Car{
		ID:           7,
		Brand:        "Audi",
		Model:        "A6",
		Year:         2024,
		CarClass:     "Business",
		BodyType:     "Sedan",
		Transmission: "Automatic",
		FuelType:     "Petrol",
		Seats:        5,
		Doors:        4,
		EngineVolume: &engineVolume,
		Horsepower:   &horsepower,
		PricePerDay:  180,
		Deposit:      700,
		Color:        &color,
		Status:       "available",
		CreatedAt:    createdAt,
		Images:       []CarImage{{ID: 1, ImageURL: "https://example.com/audi.jpg"}},
	})

	if recommended.ID != 7 || recommended.Brand != "Audi" || recommended.Horsepower != horsepower || recommended.Color != color {
		t.Fatalf("unexpected recommended car: %+v", recommended)
	}
	if recommended.EngineVolume == nil || *recommended.EngineVolume != engineVolume {
		t.Fatalf("unexpected engine volume: %+v", recommended.EngineVolume)
	}
	if len(recommended.Images) != 1 || recommended.Images[0].ID != 1 {
		t.Fatalf("unexpected images: %+v", recommended.Images)
	}
}

func TestToRecommendedCarDefaultsMissingOptionalValues(t *testing.T) {
	recommended := ToRecommendedCar(Car{ID: 1})
	if recommended.Horsepower != 0 || recommended.Color != "" || recommended.EngineVolume != nil {
		t.Fatalf("unexpected optional defaults: %+v", recommended)
	}
}
