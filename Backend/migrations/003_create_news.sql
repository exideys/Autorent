CREATE TABLE IF NOT EXISTS news (
    id INT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(120) NOT NULL,
    summary VARCHAR(240) NOT NULL,
    content TEXT NOT NULL,
    image_url VARCHAR(255),
    status VARCHAR(30) NOT NULL DEFAULT 'draft',
    published_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_news_status (status),
    INDEX idx_news_published_at (published_at),
    INDEX idx_news_created_at (created_at)
);
