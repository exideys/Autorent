package models

import "time"

const (
	RentalOrderStatusActive = "active"
)

type RentalOrderInput struct {
	CarID          int64  `json:"car_id" binding:"required"`
	StartDate      string `json:"start_date" binding:"required"`
	EndDate        string `json:"end_date" binding:"required"`
	PickupLocation string `json:"pickup_location" binding:"required,max=255"`
	PickupTime     string `json:"pickup_time" binding:"required"`
	Phone          string `json:"phone" binding:"required,max=40"`
	Notes          string `json:"notes,omitempty" binding:"omitempty,max=1000"`
}

type RentalOrderCarSummary struct {
	ID          int64   `json:"id"`
	Brand       string  `json:"brand"`
	Model       string  `json:"model"`
	Year        int     `json:"year"`
	PricePerDay float64 `json:"price_per_day"`
	Deposit     float64 `json:"deposit"`
	Status      string  `json:"status"`
	ImageURL    *string `json:"image_url,omitempty"`
}

type RentalOrder struct {
	ID             int64                 `json:"id"`
	UserID         int64                 `json:"user_id"`
	CarID          int64                 `json:"car_id"`
	StartDate      string                `json:"start_date"`
	EndDate        string                `json:"end_date"`
	PickupLocation string                `json:"pickup_location"`
	PickupTime     string                `json:"pickup_time"`
	Phone          string                `json:"phone"`
	Notes          string                `json:"notes"`
	TotalPrice     float64               `json:"total_price"`
	Deposit        float64               `json:"deposit"`
	Status         string                `json:"status"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
	Car            RentalOrderCarSummary `json:"car"`
}
