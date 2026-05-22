package database

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
)

var identifierPattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

func EnsureDatabase(serverDSN, databaseName string) error {
	if !identifierPattern.MatchString(databaseName) {
		return errors.New("database name can contain only letters, numbers, and underscores")
	}

	db, err := sql.Open("mysql", serverDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", databaseName))
	return err
}

func EnsureCarsTable(db *sql.DB) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS cars (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(120) NOT NULL,
			category VARCHAR(40) NOT NULL,
			location VARCHAR(120) NOT NULL,
			seats INT NOT NULL,
			transmission VARCHAR(30) NOT NULL,
			price_per_day DECIMAL(10, 2) NOT NULL,
			image VARCHAR(255) NOT NULL,
			available_from DATE NOT NULL,
			available_to DATE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return err
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM cars").Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	_, err := db.Exec(`
		INSERT INTO cars
			(name, category, location, seats, transmission, price_per_day, image, available_from, available_to)
		VALUES
			('Mercedes-Benz S-Class', 'Luxury', 'Downtown Showroom', 5, 'Automatic', 240.00, '/hero-main.png', '2026-05-01', '2026-12-31'),
			('Range Rover Vogue', 'SUV', 'Airport Terminal', 5, 'Automatic', 210.00, '/hero-main.png', '2026-05-01', '2026-12-31'),
			('Porsche 911 Carrera', 'Sport', 'Premium Garage', 2, 'Automatic', 320.00, '/hero-main.png', '2026-05-01', '2026-12-31'),
			('BMW 7 Series', 'Business', 'Business District', 5, 'Automatic', 190.00, '/hero-main.png', '2026-05-01', '2026-12-31'),
			('Audi Q8', 'SUV', 'City Center', 5, 'Automatic', 180.00, '/hero-main.png', '2026-05-01', '2026-12-31'),
			('Lamborghini Huracan', 'Sport', 'Luxury Hub', 2, 'Automatic', 520.00, '/hero-main.png', '2026-05-01', '2026-12-31')
	`)
	return err
}
