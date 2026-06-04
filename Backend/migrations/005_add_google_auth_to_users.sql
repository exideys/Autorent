ALTER TABLE users
    MODIFY password_hash VARCHAR(255) NULL,
    ADD COLUMN google_sub VARCHAR(255) NULL,
    ADD UNIQUE KEY users_google_sub_unique (google_sub);
