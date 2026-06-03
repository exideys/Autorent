package models

import "time"

type Car struct {
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
	Horsepower   *int       `json:"horsepower,omitempty"`
	PricePerDay  float64    `json:"price_per_day"`
	Deposit      float64    `json:"deposit"`
	Color        *string    `json:"color,omitempty"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	Images       []CarImage `json:"images"`
}

type CarImage struct {
	ID        int64  `json:"id"`
	CarID     int64  `json:"car_id"`
	ImageURL  string `json:"image_url"`
	IsMain    bool   `json:"is_main"`
	SortOrder int    `json:"sort_order"`
}

type CarInput struct {
	Brand        string          `json:"brand" binding:"required,max=50"`
	Model        string          `json:"model" binding:"required,max=50"`
	Year         int             `json:"year" binding:"required"`
	CarClass     string          `json:"car_class" binding:"required,max=50"`
	BodyType     string          `json:"body_type" binding:"required,max=50"`
	Transmission string          `json:"transmission" binding:"required,max=30"`
	FuelType     string          `json:"fuel_type" binding:"required,max=30"`
	Seats        int             `json:"seats" binding:"required"`
	Doors        int             `json:"doors" binding:"required"`
	EngineVolume *float64        `json:"engine_volume,omitempty"`
	Horsepower   *int            `json:"horsepower,omitempty"`
	PricePerDay  float64         `json:"price_per_day" binding:"required"`
	Deposit      float64         `json:"deposit" binding:"required"`
	Color        *string         `json:"color,omitempty" binding:"omitempty,max=30"`
	Status       string          `json:"status,omitempty" binding:"omitempty,max=30"`
	Images       []CarImageInput `json:"images,omitempty" binding:"omitempty,dive"`
}

type CarImageInput struct {
	ImageURL  string `json:"image_url" binding:"required,max=2048"`
	IsMain    bool   `json:"is_main"`
	SortOrder int    `json:"sort_order"`
}

type CarFilters struct {
	Status       string
	CarClass     string
	BodyType     string
	Transmission string
	FuelType     string
	SortBy       string
	SortOrder    string
}
