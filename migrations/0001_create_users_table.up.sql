CREATE TABLE IF NOT EXISTS users (
                       id SERIAL PRIMARY KEY,
                       passport_number VARCHAR(20) NOT NULL UNIQUE,
                       surname VARCHAR(100),
                       name VARCHAR(100),
                       patronymic VARCHAR(100),
                       address VARCHAR(255)
);