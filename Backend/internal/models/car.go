package models

type Car struct {
	ID           int      `json:"id"`
	Brand        string   `json:"brand"`
	Model        string   `json:"model"`
	Year         int      `json:"year"`
	CarClass     string   `json:"car_class"`
	BodyType     string   `json:"body_type"`
	Transmission string   `json:"transmission"`
	FuelType     string   `json:"fuel_type"`
	Seats        int      `json:"seats"`
	Doors        int      `json:"doors"`
	EngineVolume *float64 `json:"engine_volume"`
	Horsepower   int      `json:"horsepower"`
	PricePerDay  float64  `json:"price_per_day"`
	Deposit      float64  `json:"deposit"`
	Color        string   `json:"color"`
	Status       string   `json:"status"`
}
