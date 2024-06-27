package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"sync"
)

type PostgresStorage struct {
	db *sql.DB
	mu sync.RWMutex
}

func NewPostgresStorage() (*PostgresStorage, error) {
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
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresStorage{db: db}, nil
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
