CREATE TABLE IF NOT EXISTS rental_orders (
    id INT PRIMARY KEY AUTO_INCREMENT,

    user_id INT NOT NULL,
    car_id INT NOT NULL,

    start_date DATE NOT NULL,
    end_date DATE NOT NULL,

    pickup_location VARCHAR(255) NOT NULL DEFAULT '',
    pickup_time TIME,
    phone VARCHAR(40) NOT NULL DEFAULT '',
    notes TEXT,

    total_price DECIMAL(10,2) NOT NULL,
    deposit DECIMAL(10,2) NOT NULL,

    status VARCHAR(30) NOT NULL DEFAULT 'active',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_rental_orders_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_rental_orders_car
        FOREIGN KEY (car_id)
        REFERENCES cars(id)
        ON DELETE CASCADE,

    INDEX idx_rental_orders_user (user_id),
    INDEX idx_rental_orders_car (car_id),
    INDEX idx_rental_orders_status (status),
    INDEX idx_rental_orders_created_at (created_at)
);