package server

import (
	"github.com/joho/godotenv"
	"net/http"
	"time-tracker/internal/config"
	"time-tracker/internal/handlers"
	"time-tracker/internal/logger"
	"time-tracker/internal/storage/postgres"
	"time-tracker/internal/usecase"
)

func StartServer() error {

	// Инициализация логгера
	if err := logger.InitLogger("log.txt"); err != nil {
		panic("cannot initialize zap")
	}
	defer logger.SugaredLogger().Sync()

	//считываем файл .env
	err := godotenv.Load(".env")
	if err != nil {
		logger.SugaredLogger().Errorw("Ошибка загрузки файла .env", "error", err)
	}
	//парсим переменные окружения
	conf, err := config.ParseConfigServer()
	if err != nil {
		logger.SugaredLogger().Errorw("Ошибка считывания переменных окружения", "error", err)
		return err
	}

	logger.SugaredLogger().Infow("Старт сервера", "addr", conf.SERVER_HOST+":"+conf.SERVER_PORT)

	//подключение к БД
	db, err := postgres.NewPostgresStorage(conf)
	if err != nil {
		logger.SugaredLogger().Fatalw("Ошибка при подключении к БД", "error", err)
		return err
	}
	useCase := usecase.NewUseCaseStorage(db)
	logger.SugaredLogger().Infow("Успешное подключение к БД")

	r := handlers.InitRoutes(useCase, conf)

	//создние сервера
	err = http.ListenAndServe(conf.SERVER_HOST+":"+conf.SERVER_PORT, r)
	if err != nil {
		return err
	}
	return nil
}
