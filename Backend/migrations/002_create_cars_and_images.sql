CREATE TABLE IF NOT EXISTS cars (
    id INT PRIMARY KEY AUTO_INCREMENT,
    brand VARCHAR(50) NOT NULL,
    model VARCHAR(50) NOT NULL,
    year INT NOT NULL,
    car_class VARCHAR(50) NOT NULL,
    body_type VARCHAR(50) NOT NULL,
    transmission VARCHAR(30) NOT NULL,
    fuel_type VARCHAR(30) NOT NULL,
    seats INT NOT NULL,
    doors INT NOT NULL,
    engine_volume DECIMAL(3,1),
    horsepower INT,
    price_per_day DECIMAL(10,2) NOT NULL,
    deposit DECIMAL(10,2) NOT NULL,
    color VARCHAR(30),
    status VARCHAR(30) DEFAULT 'available',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_cars_status (status),
    INDEX idx_cars_class (car_class),
    INDEX idx_cars_price_per_day (price_per_day),
    INDEX idx_cars_year (year),
    INDEX idx_cars_created_at (created_at)
);

CREATE TABLE IF NOT EXISTS car_images (
    id INT PRIMARY KEY AUTO_INCREMENT,
    car_id INT NOT NULL,
    image_url VARCHAR(255) NOT NULL,
    is_main BOOLEAN DEFAULT FALSE,
    sort_order INT DEFAULT 0,

    CONSTRAINT fk_car_images_car
        FOREIGN KEY (car_id)
        REFERENCES cars(id)
        ON DELETE CASCADE,

    INDEX idx_car_images_car_order (car_id, is_main, sort_order)
);
