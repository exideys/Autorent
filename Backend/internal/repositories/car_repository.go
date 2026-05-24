package repositories

import "autorent-backend/internal/models"

type CarRepository interface {
	FindAvailableCars(filters models.CarSearchFilters) ([]models.Car, error)
}
