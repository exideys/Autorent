package data

import "autorent-backend/internal/models"

func float64Ptr(value float64) *float64 {
	return &value
}

func intPtr(value int) *int {
	return &value
}

func stringPtr(value string) *string {
	return &value
}

func MockCars() []models.Car {
	return []models.Car{
		{
			ID:           1,
			Brand:        "Mercedes-Benz",
			Model:        "S-Class",
			Year:         2023,
			CarClass:     "Premium",
			BodyType:     "Sedan",
			Transmission: "Automatic",
			FuelType:     "Petrol",
			Seats:        5,
			Doors:        4,
			EngineVolume: float64Ptr(3.0),
			Horsepower:   intPtr(429),
			PricePerDay:  220.00,
			Deposit:      1000.00,
			Color:        stringPtr("Black"),
			Status:       "available",
		},
		{
			ID:           2,
			Brand:        "BMW",
			Model:        "7 Series",
			Year:         2023,
			CarClass:     "Premium",
			BodyType:     "Sedan",
			Transmission: "Automatic",
			FuelType:     "Petrol",
			Seats:        5,
			Doors:        4,
			EngineVolume: float64Ptr(3.0),
			Horsepower:   intPtr(375),
			PricePerDay:  210.00,
			Deposit:      950.00,
			Color:        stringPtr("White"),
			Status:       "available",
		},
		{
			ID:           3,
			Brand:        "Audi",
			Model:        "A8",
			Year:         2022,
			CarClass:     "Premium",
			BodyType:     "Sedan",
			Transmission: "Automatic",
			FuelType:     "Diesel",
			Seats:        5,
			Doors:        4,
			EngineVolume: float64Ptr(3.0),
			Horsepower:   intPtr(286),
			PricePerDay:  190.00,
			Deposit:      900.00,
			Color:        stringPtr("Gray"),
			Status:       "available",
		},
		{
			ID:           4,
			Brand:        "Porsche",
			Model:        "Cayenne",
			Year:         2023,
			CarClass:     "Luxury SUV",
			BodyType:     "SUV",
			Transmission: "Automatic",
			FuelType:     "Petrol",
			Seats:        5,
			Doors:        5,
			EngineVolume: float64Ptr(3.0),
			Horsepower:   intPtr(348),
			PricePerDay:  250.00,
			Deposit:      1200.00,
			Color:        stringPtr("Dark Blue"),
			Status:       "available",
		},
		{
			ID:           5,
			Brand:        "Tesla",
			Model:        "Model S",
			Year:         2024,
			CarClass:     "Electric Premium",
			BodyType:     "Liftback",
			Transmission: "Automatic",
			FuelType:     "Electric",
			Seats:        5,
			Doors:        5,
			EngineVolume: nil,
			Horsepower:   intPtr(670),
			PricePerDay:  230.00,
			Deposit:      1000.00,
			Color:        stringPtr("Red"),
			Status:       "available",
		},
	}
}
