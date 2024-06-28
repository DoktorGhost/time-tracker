package postgres

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"sync"
)

type PostgresStorage struct {
	db    *sql.DB
	mu    sync.RWMutex
	sugar *zap.SugaredLogger
}

func NewPostgresStorage(sugar *zap.SugaredLogger) (*PostgresStorage, error) {
	/*
		login := os.Getenv("DB_LOGIN")
		password := os.Getenv("DB_PASSWORD")
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		dbname := os.Getenv("DB_NAME")
	*/
	login := "admin"
	password := "admin"
	host := "localhost"
	port := "5433"
	dbname := "postgres"

	//"postgres://admin:admin@localhost:5433/postgres?sslmode=disable"

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", login, password, host, port, dbname)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		sugar.Errorw("Ошибка подключения к БД", "error", err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		sugar.Errorw("Не удалось проверить связь с БД", "error", err)
		return nil, err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		sugar.Errorw("Не удалось создать драйвер миграции", "error", err)
		return nil, err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations",
		"postgres", driver)
	if err != nil {
		sugar.Errorw("Не удалось создать объект миграции", "error", err)
		return nil, err
	}

	//применение миграций
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		sugar.Errorw("Не удалось применить миграции", "error", err)
		return nil, err
	}

	sugar.Infow("Миграции успешно применены")

	return &PostgresStorage{db: db, sugar: sugar}, nil
}

func (p *PostgresStorage) Create(name string) error {
	return nil
}

func (p *PostgresStorage) Read(name string) (string, error) {
	return "", nil
}

func (p *PostgresStorage) Update(name string) error {
	return nil
}

func (p *PostgresStorage) Delete(name string) error {
	return nil
}
