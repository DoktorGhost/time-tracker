package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	"strconv"
	"sync"
	"time"
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

	//sugar.Infow("Миграции успешно применены")

	return &PostgresStorage{db: db}, nil
}

func (p *PostgresStorage) Create(userData models.UserData) (int, error) {
	query := `
INSERT INTO users (passport_series, passport_number, surname, name, patronymic, address)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;
`
	var userID int
	err := p.db.QueryRow(query, userData.PassportSeries, userData.PassportNumber, userData.Surname, userData.Name, userData.Patronymic, userData.Address).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (p *PostgresStorage) Read(userID int) (models.UserData, error) {
	query := `
		SELECT id, passport_series, passport_number, surname, name, patronymic, address FROM users WHERE id = $1;
	`

	data := models.UserData{}
	err := p.db.QueryRow(query, userID).Scan(
		&data.UserID,
		&data.PassportSeries,
		&data.PassportNumber,
		&data.Surname,
		&data.Name,
		&data.Patronymic,
		&data.Address,
	)
	if err != nil {
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
		SET passport_number = $2, passport_series = $3, surname = $4, name = $5, patronymic = $6, address = $7
		WHERE id = $1
	`
	if userData.PassportNumber != "" {
		data.PassportNumber = userData.PassportNumber
	}
	if userData.PassportSeries != "" {
		data.PassportSeries = userData.PassportSeries
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
		data.PassportSeries,
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
	_, err := p.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("ошибка удаления записи: %v", err)
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
	if dataFilter.PassportSeries != "" {
		query += " AND passport_series = $" + strconv.Itoa(argCounter)
		args = append(args, dataFilter.PassportSeries)
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
		if err := rows.Scan(&user.UserID, &user.PassportSeries, &user.PassportNumber, &user.Surname, &user.Name, &user.Patronymic, &user.Address); err != nil {
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
		return models.TaskData{}, err
	}

	return data, nil
}

func (p *PostgresStorage) AddStartTime(taskID int) error {
	_, err := p.ReadTask(taskID)
	if err != nil {
		return err
	}

	query := `
UPDATE tasks
SET start_time = NOW()
WHERE id = $1;
`
	_, err = p.db.Exec(query, taskID)

	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresStorage) AddEndTime(taskID int) error {
	dataTask, err := p.ReadTask(taskID)
	if err != nil {
		return err
	}
	if !dataTask.StartTime.Valid {
		return errors.New("Start time is NULL")
	}

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
	return nil
}

func (p *PostgresStorage) GetTasksUser(userID int, timeTask models.TaskTime) ([]models.Tasks, error) {
	// Преобразуем строки Start и End в тип time.Time
	sTime, err := time.Parse("02.01.2006", timeTask.Start)
	if err != nil {
		return nil, err
	}
	eTime, err := time.Parse("02.01.2006", timeTask.End)
	if err != nil {
		return nil, err
	}
	startTime := sTime.Format("2006-01-02 15:04:05")
	endTime := eTime.Format("2006-01-02 15:04:05")
	query := `
        SELECT name_task, all_time
        FROM tasks
        WHERE user_id = $1 AND start_time >= $2 AND end_time <= $3
        ORDER BY all_time DESC;
    `

	rows, err := p.db.Query(query, userID, startTime, endTime)
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
