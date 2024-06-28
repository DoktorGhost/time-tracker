package server

import (
	"log"
	"time-tracker/internal/logger"
	"time-tracker/internal/storage/postgres"
	"time-tracker/internal/usecase"
)

func StartServer() error {

	//инициализация логгера
	sugar, err := logger.InitLogger("../../log.txt")
	if err != nil {
		log.Println("Ошибка инициализации логгера: ", err)
		return err
	}

	//подключение к БД
	db, err := postgres.NewPostgresStorage(sugar)
	if err != nil {
		sugar.Fatalw("Ошибка при подключении к БД", "error", err)
		return err
	}
	_ = usecase.NewUseCaseStorage(db)
	sugar.Infow("Успешное подключение к БД")

	//создние сервера
	//err := http.ListenAndServe("localhost:8080", r)
	return nil
}
