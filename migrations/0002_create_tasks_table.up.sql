CREATE TYPE task_status AS ENUM ('выполнена', 'отменена', 'в процессе', 'ожидание', 'завершена');

CREATE TABLE IF NOT EXISTS tasks (
                       id SERIAL PRIMARY KEY,
                       user_id INT REFERENCES users(id),
                       name_task VARCHAR(100) NOT NULL,
                       start_time TIMESTAMP,
                       end_time TIMESTAMP,
                       all_time TIMESTAMP,
                       process_flag task_status
);