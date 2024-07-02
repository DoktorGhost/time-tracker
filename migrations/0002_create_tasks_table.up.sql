CREATE TABLE IF NOT EXISTS tasks (
                       id SERIAL PRIMARY KEY,
                       user_id INT REFERENCES users(id) ON DELETE CASCADE,
                       name_task VARCHAR(100) NOT NULL UNIQUE,
                       start_time TIMESTAMP,
                       end_time TIMESTAMP,
                       all_time INT
);