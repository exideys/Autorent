package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Car struct {
	ID           int64                `json:"id"`
	Name         string               `json:"name"`
	Category     string               `json:"category"`
	Location     string               `json:"location"`
	Seats        int                  `json:"seats"`
	Transmission string               `json:"transmission"`
	PricePerDay  float64              `json:"pricePerDay"`
	Image        string               `json:"image"`
	Availability []AvailabilityWindow `json:"availability"`
}

type AvailabilityWindow struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func CarsHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.QueryContext(c.Request.Context(), `
			SELECT
				id,
				name,
				category,
				location,
				seats,
				transmission,
				price_per_day,
				image,
				available_from,
				available_to
			FROM cars
			ORDER BY id
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load cars"})
			return
		}
		defer rows.Close()

		cars := make([]Car, 0)
		for rows.Next() {
			var car Car
			var availableFrom sql.NullTime
			var availableTo sql.NullTime

			if err := rows.Scan(
				&car.ID,
				&car.Name,
				&car.Category,
				&car.Location,
				&car.Seats,
				&car.Transmission,
				&car.PricePerDay,
				&car.Image,
				&availableFrom,
				&availableTo,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read cars"})
				return
			}

			car.Availability = make([]AvailabilityWindow, 0)
			if availableFrom.Valid && availableTo.Valid {
				car.Availability = append(car.Availability, AvailabilityWindow{
					From: availableFrom.Time.Format("2006-01-02"),
					To:   availableTo.Time.Format("2006-01-02"),
				})
			}

			cars = append(cars, car)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load cars"})
			return
		}

		c.JSON(http.StatusOK, cars)
	}
}
