CREATE TABLE IF NOT EXISTS users (
                       id SERIAL PRIMARY KEY,
                       passport_series VARCHAR(4) NOT NULL,
                       passport_number VARCHAR(6) NOT NULL,
                       surname VARCHAR(100),
                       name VARCHAR(100),
                       patronymic VARCHAR(100),
                       address VARCHAR(255)
);