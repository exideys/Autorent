package models

type CarRecommendationRequest struct {
	Message string `json:"message" binding:"required"`
}

type CarRecommendationResponse struct {
	Answer string `json:"answer"`
	Cars   []Car  `json:"cars"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type CarSearchFilters struct {
	MinSeats          int     `json:"min_seats"`
	MaxPricePerDay    float64 `json:"max_price_per_day"`
	PreferredBodyType string  `json:"preferred_body_type"`
	PreferredFuelType string  `json:"preferred_fuel_type"`
	PreferredClass    string  `json:"preferred_class"`
	Transmission      string  `json:"transmission"`
	Purpose           string  `json:"purpose"`
	OnlyAvailable     bool    `json:"only_available"`
}
