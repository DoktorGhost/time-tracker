package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"strconv"
	"sync"
	"time-tracker/internal/config"
	"time-tracker/internal/models"
	"time-tracker/internal/validator"
)

type PostgresStorage struct {
	db *sql.DB
	mu sync.RWMutex
}

func NewPostgresStorage(conf *config.Config) (*PostgresStorage, error) {

	login := conf.DB_LOGIN
	password := conf.DB_PASS
	host := conf.DB_HOST
	port := conf.DB_PORT
	dbname := conf.DB_NAME

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", login, password, host, port, dbname)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		//sugar.Errorw("Не удалось создать объект миграции", "error", err)
		return nil, err
	}

	//применение миграций
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		//sugar.Errorw("Не удалось применить миграции", "error", err)
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

func (p *PostgresStorage) Create(userData models.UserData) (int, error) {
	query := `
INSERT INTO users (passport_number, surname, name, patronymic, address)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;
`
	var userID int
	err := p.db.QueryRow(query, userData.PassportNumber, userData.Surname, userData.Name, userData.Patronymic, userData.Address).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		errors.As(err, &pgErr)
		return 0, err
	}
	return userID, nil
}

func (p *PostgresStorage) Read(userID int) (models.UserData, error) {
	query := `
		SELECT id, passport_number, surname, name, patronymic, address FROM users WHERE id = $1;
	`

	data := models.UserData{}
	err := p.db.QueryRow(query, userID).Scan(
		&data.UserID,
		&data.PassportNumber,
		&data.Surname,
		&data.Name,
		&data.Patronymic,
		&data.Address,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.UserData{}, fmt.Errorf("пользователь с id %d не найден", userID)
		}
		return models.UserData{}, err
	}

	return data, nil
}

func (p *PostgresStorage) Update(userID int, userData models.UserData) error {
	data, err := p.Read(userID)
	if err != nil {
		return err
	}
	query := `
		UPDATE users
		SET passport_number = $2, surname = $3, name = $4, patronymic = $5, address = $6
		WHERE id = $1
	`
	if userData.PassportNumber != "" {
		data.PassportNumber = userData.PassportNumber
	}
	if userData.Surname != "" {
		data.Surname = userData.Surname
	}
	if userData.Name != "" {
		data.Name = userData.Name
	}
	if userData.Patronymic != "" {
		data.Patronymic = userData.Patronymic
	}
	if userData.Address != "" {
		data.Address = userData.Address
	}

	_, err = p.db.Exec(query,
		userID,
		data.PassportNumber,
		data.Surname,
		data.Name,
		data.Patronymic,
		data.Address,
	)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostgresStorage) Delete(userID int) error {
	query := `DELETE FROM users WHERE id = $1;`
	result, err := p.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("ошибка удаления записи: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("пользователь с id %d не найден", userID)
	}
	return nil
}
func (p *PostgresStorage) GetUsers(dataFilter models.UserData, page, limit int) ([]models.UserData, error) {
	query := `SELECT * FROM users WHERE 1=1`
	args := []interface{}{}
	argCounter := 1

	if dataFilter.UserID != "" {
		query += " AND id = $" + strconv.Itoa(argCounter)
		args = append(args, dataFilter.UserID)
		argCounter++
	}
	if dataFilter.PassportNumber != "" {
		query += " AND passport_number = $" + strconv.Itoa(argCounter)
		args = append(args, dataFilter.PassportNumber)
		argCounter++
	}
	if dataFilter.Surname != "" {
		query += " AND surname = $" + strconv.Itoa(argCounter)
		args = append(args, dataFilter.Surname)
		argCounter++
	}
	if dataFilter.Name != "" {
		query += " AND name = $" + strconv.Itoa(argCounter)
		args = append(args, dataFilter.Name)
		argCounter++
	}
	if dataFilter.Patronymic != "" {
		query += " AND patronymic = $" + strconv.Itoa(argCounter)
		args = append(args, dataFilter.Patronymic)
		argCounter++
	}
	if dataFilter.Address != "" {
		query += " AND address = $" + strconv.Itoa(argCounter)
		args = append(args, dataFilter.Address)
		argCounter++
	}

	offset := (page - 1) * limit
	query += " ORDER BY id ASC"
	query += " LIMIT $" + strconv.Itoa(argCounter) + " OFFSET $" + strconv.Itoa(argCounter+1)
	args = append(args, limit, offset)

	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []models.UserData{}
	for rows.Next() {
		var user models.UserData
		if err := rows.Scan(&user.UserID, &user.PassportNumber, &user.Surname, &user.Name, &user.Patronymic, &user.Address); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (p *PostgresStorage) CreateTask(userID int, nameTask string) (int, error) {
	_, err := p.Read(userID)
	if err != nil {
		return 0, err
	}

	query := `
INSERT INTO tasks (user_id, name_task)
VALUES ($1, $2)
RETURNING id;
`
	var taskID int
	err = p.db.QueryRow(query, userID, nameTask).Scan(&taskID)
	if err != nil {
		var pgErr *pgconn.PgError
		errors.As(err, &pgErr)
		return 0, err
	}
	return taskID, nil
}

func (p *PostgresStorage) ReadTask(taskID int) (models.TaskData, error) {
	query := `
		SELECT id, user_id, name_task, start_time, end_time, all_time FROM tasks WHERE id = $1;
	`

	data := models.TaskData{}
	err := p.db.QueryRow(query, taskID).Scan(
		&data.TaskID,
		&data.UserID,
		&data.NameTask,
		&data.StartTime,
		&data.EndTime,
		&data.AllTime,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.TaskData{}, fmt.Errorf("задача с id %d не найдена", taskID)
		}
		return models.TaskData{}, err
	}

	return data, nil
}

func (p *PostgresStorage) AddStartTime(taskID int) error {
	// Проверяем, что задача существует и получаем её данные
	task, err := p.ReadTask(taskID)
	if err != nil {
		return err
	}

	// Проверяем, что поле start_time равно NULL
	if !task.StartTime.Valid { // Проверяем, что start_time является NULL
		query := `
		UPDATE tasks
		SET start_time = NOW()
		WHERE id = $1 AND start_time IS NULL; -- Обновляем только если start_time равно NULL
		`
		_, err = p.db.Exec(query, taskID)

		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("поле start_time уже заполнено")
	}

	return nil
}

func (p *PostgresStorage) AddEndTime(taskID int) error {
	task, err := p.ReadTask(taskID)
	if err != nil {

		return err
	}
	if task.StartTime.Valid {
		if !task.EndTime.Valid {
			query := `
				UPDATE tasks
				SET end_time = NOW(),
    				all_time = EXTRACT(EPOCH FROM NOW() - start_time)
				WHERE id = $1;
				`
			_, err = p.db.Exec(query, taskID)

			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("поле end_time уже заполнено")
		}
	} else {
		return fmt.Errorf("поле start_time не заполнено")
	}

	return nil
}

func (p *PostgresStorage) GetTasksUser(userID int, timeTask models.TaskTime) ([]models.Tasks, error) {

	query := `
        SELECT name_task, all_time
        FROM tasks
        WHERE user_id = $1 AND start_time >= $2 AND end_time <= $3
        ORDER BY all_time DESC;
    `

	rows, err := p.db.Query(query, userID, timeTask.Start, timeTask.End)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Tasks
	var sec int
	for rows.Next() {
		var task models.Tasks
		if err := rows.Scan(&task.Name, &sec); err != nil {
			return nil, err
		}
		task.AllTime = validator.SecondToString(sec)
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}
