package handlers

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgconn"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"net/http"
	"strconv"
	"strings"
	"time"
	_ "time-tracker/docs"
	"time-tracker/internal/API/apiDataUser"
	"time-tracker/internal/config"
	"time-tracker/internal/logger"
	"time-tracker/internal/models"
	"time-tracker/internal/usecase"
	"time-tracker/internal/validator"
)

func InitRoutes(useCase usecase.UseCaseStorage, conf *config.Config) chi.Router {
	r := chi.NewRouter()

	r.Use(logger.WithLogging)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://"+conf.SERVER_HOST+":"+conf.SERVER_PORT+"/swagger/doc.json"), //The url pointing to API definition
	))

	r.Post("/addUser", func(w http.ResponseWriter, r *http.Request) {
		HandlerAddUser(w, r, useCase, conf)
	})
	r.Delete("/delUser/{userID}", func(w http.ResponseWriter, r *http.Request) {
		HandlerDelete(w, r, useCase)
	})
	r.Put("/updUser/{userID}", func(w http.ResponseWriter, r *http.Request) {
		HandlerUpdate(w, r, useCase)
	})
	r.Get("/getUser/{userID}", func(w http.ResponseWriter, r *http.Request) {
		HandlerGetUser(w, r, useCase)
	})
	r.Post("/getUsers/{page}/{limit}", func(w http.ResponseWriter, r *http.Request) {
		HandlerGetUsers(w, r, useCase)
	})
	r.Post("/addTask/{userID}", func(w http.ResponseWriter, r *http.Request) {
		HandlerAddTask(w, r, useCase)
	})
	r.Put("/addTask/startTime/{taskID}", func(w http.ResponseWriter, r *http.Request) {
		HandlerStartTime(w, r, useCase)
	})
	r.Put("/addTask/endTime/{taskID}", func(w http.ResponseWriter, r *http.Request) {
		HandlerEndTime(w, r, useCase)
	})
	r.Post("/getTasks/{userID}", func(w http.ResponseWriter, r *http.Request) {
		HandlerGetTasks(w, r, useCase)
	})
	///тесты
	r.Post("/testAdd", func(w http.ResponseWriter, r *http.Request) {
		HandlerCreat(w, r, useCase)
	})

	return r
}

// @Summary Добавление нового пользователя
// @Description Добавляет нового пользователя на основе серии и номера паспорта, обогащает информацию через внешний API (если в .env не указан URL API - получим ответ 500)
// @Tags Users
// @Accept json
// @Produce json
// @Param body body models.PassportRequest true "Серия и номер пасспорта в формате `1234 123456` (4 цифры, пробел, 6 цифр)"
// @Success 200 {string} string "UserID"
// @Failure 400 {string} string "Ошибка декодирования тела запроса"
// @Failure 409 {string} string "Ошибка записи: Пользователь с таким номером паспорта уже существует"
// @Failure 422 {string} string "Ошибка валидации серии паспорта или номера паспорта"
// @Failure 500 {string} string "Ошибка сервера"
// @Failure 503 {string} string "Ошибка запроса к стороннему API"
// @Router /addUser [post]
func HandlerAddUser(w http.ResponseWriter, r *http.Request, useCase usecase.UseCaseStorage, conf *config.Config) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req models.PassportRequest
	dec := json.NewDecoder(r.Body)

	dec.DisallowUnknownFields()

	// Попытка декодировать JSON в структуру UserData
	if err := dec.Decode(&req); err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	parts := strings.Split(req.PassportNumber, " ")

	if err := validator.ValidateDigits(parts[0], 4); err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if err := validator.ValidateDigits(parts[1], 6); err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	userData, err := apiDataUser.GetPeopleInfoFromAPI(parts[0], parts[1], conf.API_URL)
	if err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	user_id, err := useCase.UseCaseCreate(*userData)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				logger.SugaredLogger().Debug(err)
				w.WriteHeader(http.StatusConflict)
				return
			}
		}
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Ошибка записи в БД"))
		return
	}

	response := map[string]int{"UserID": user_id}
	res, err := json.Marshal(response)
	if err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

// @Summary		Тестовый хендлер: добавление пользователя в обход стороннего API
// @Description	Добавляет нового пользователя на основе серии и номера паспорта, остальная информация берется рандомно, для тестирования и отладки запросов
// @Tags		Users
// @Accept		json
// @Produce		plain
// @Param		body	body		models.PassportRequest	true	"Серия и номер пасспорта в формате `1234 123456` (4 цифры, пробел, 6 цифр)"
// @Success 200 {string} string "UserID"
// @Failure 400 {string} string "Ошибка декодирования тела запроса"
// @Failure 409 {string} string "Ошибка записи: Пользователь с таким номером паспорта уже существует"
// @Failure 422 {string} string "Ошибка валидации серии паспорта или номера паспорта"
// @Failure 500 {string} string "Ошибка сервера"
// @Router	/testAdd [post]
func HandlerCreat(w http.ResponseWriter, r *http.Request, useCase usecase.UseCaseStorage) {
	var req models.PassportRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// Попытка декодировать JSON в структуру UserData
	if err := dec.Decode(&req); err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	parts := strings.Split(req.PassportNumber, " ")
	passportSeries := parts[0]
	if err := validator.ValidateDigits(passportSeries, 4); err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	passportNumber := parts[1]
	if err := validator.ValidateDigits(passportNumber, 6); err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	userData := models.UserData{
		PassportNumber: req.PassportNumber,
		Surname:        validator.GenerateRandomString(7),
		Name:           validator.GenerateRandomString(5),
		Patronymic:     validator.GenerateRandomString(8),
		Address:        validator.GenerateRandomString(15),
	}

	userID, err := useCase.UseCaseCreate(userData)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				logger.SugaredLogger().Debug(err)
				w.WriteHeader(http.StatusConflict)
				return
			}
		}
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Ошибка записи в БД"))
		return
	}

	response := map[string]int{"UserID": userID}
	res, err := json.Marshal(response)
	if err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

// @Summary Удаление пользователя по ID
// @Description Удаляет пользователя из системы по его идентификатору.
// @Tags Users
// @Accept json
// @Produce json
// @Param userID path int true "User ID" Format(int)
// @Success 200 {string} string "Пользователь успешно удален"
// @Failure 404 {string} string "Пользователь не найден"
// @Failure 422 {string} string "Ошибка конвертирования ID"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /delUser/{userID} [delete]
func HandlerDelete(w http.ResponseWriter, r *http.Request, useCase usecase.UseCaseStorage) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userIDStr := chi.URLParam(r, "userID")

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = useCase.UseCaseDelete(userID)
	if err != nil {
		if strings.Contains(err.Error(), "не найден") {
			logger.SugaredLogger().Debug(err)
			w.WriteHeader(http.StatusNotFound)
		} else {
			logger.SugaredLogger().Debug(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Обновление данных пользователя
// @Description Обновляет данные пользователя по его идентификатору.
// @Tags Users
// @Accept json
// @Produce json
// @Param userID path int true "User ID" Format(int64)
// @Param body body models.UserData true "Данные пользователя (неменяемые поля оставляем пустыми)"
// @Success 200 {string} string "Данные пользователя успешно обновлены"
// @Failure 400 {string} string "Ошибка декодирования тела запроса"
// @Failure 404 {string} string "Пользователь не найден"
// @Failure 409 {string} string "Ошибка записи: Пользователь с таким номером паспорта уже существует"
// @Failure 422 {string} string "Ошибка конвертирования ID"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /updUser/{userID} [put]
func HandlerUpdate(w http.ResponseWriter, r *http.Request, useCase usecase.UseCaseStorage) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userIDstr := chi.URLParam(r, "userID")

	userID, err := strconv.Atoi(userIDstr)
	if err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	var req models.UserData
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(req.PassportNumber) > 0 {
		parts := strings.Split(req.PassportNumber, " ")
		if err := validator.ValidateDigits(parts[0], 4); err != nil {
			logger.SugaredLogger().Debug(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := validator.ValidateDigits(parts[1], 6); err != nil {
			logger.SugaredLogger().Debug(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	err = useCase.UseCaseUpdate(userID, req)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				logger.SugaredLogger().Debug("Ошибка:", err)
				w.WriteHeader(http.StatusConflict)
				return
			}
		}
		if strings.Contains(err.Error(), "не найден") {
			logger.SugaredLogger().Debug(err)
			w.WriteHeader(http.StatusNotFound)
		} else {
			logger.SugaredLogger().Debug(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Получение информации о пользователе
// @Description Получает информацию о пользователе по его уникальному идентификатору.
// @Tags Users
// @Accept json
// @Produce json
// @Param userID path int true "User ID"
// @Success 200 {object} models.UserData "Успешный ответ с данными пользователя"
// @Failure 404 {string} string "Пользователь не найден"
// @Failure 422 {string} string "Ошибка конвертирования ID"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /getUser/{userID} [get]
func HandlerGetUser(w http.ResponseWriter, r *http.Request, useCase usecase.UseCaseStorage) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userIDStr := chi.URLParam(r, "userID")

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	userData, err := useCase.UseCaseRead(userID)
	if err != nil {
		if strings.Contains(err.Error(), "не найден") {
			logger.SugaredLogger().Debug(err)
			w.WriteHeader(http.StatusNotFound)
		} else {
			logger.SugaredLogger().Debug(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	response, err := json.Marshal(userData)
	if err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// @Summary Получение списка пользователей
// @Description Возвращает список пользователей с возможностью фильтрации и пагинации.
// @Tags Users
// @Accept json
// @Produce json
// @Param page path int true "Номер страницы"
// @Param limit path int true "Количество элементов на странице"
// @Param body body models.UserData false "Фильтр пользователей (выбираем по каким полям будет фильтрация, вписываем туда ключ фильтра. Ненужные делаем пусытими или удаляем)"
// @Success 200 {array} models.UserData "Успешный ответ с данными пользователей"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /getUsers/{page}/{limit} [post]
func HandlerGetUsers(w http.ResponseWriter, r *http.Request, useCase usecase.UseCaseStorage) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req models.UserData
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	// Получаем параметры пагинации из URL
	pageStr := chi.URLParam(r, "page")
	limitStr := chi.URLParam(r, "limit")

	// Устанавливаем значения по умолчанию для пагинации
	page := 1
	limit := 10

	// Конвертируем параметры пагинации в int
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	// Используем параметры фильтрации и пагинации в запросе к базе данных
	users, err := useCase.UseCaseGetUsers(req, page, limit)
	if err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(users)
	if err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)

}

// @Summary Добавление новой задачи
// @Description Добавляет новую задачу для указанного пользователя.
// @Tags Tasks
// @Accept json
// @Produce json
// @Param userID path int true "ID пользователя"
// @Param body body models.TaskName true "Название задачи"
// @Success 200 {string} string "TaskID: {taskID}"
// @Failure 404 {string} string "Пользователь не найден"
// @Failure 422 {string} string "Ошибка конвертирования UserID"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /addTask/{userID} [post]
func HandlerAddTask(w http.ResponseWriter, r *http.Request, useCase usecase.UseCaseStorage) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userIDstr := chi.URLParam(r, "userID")
	userID, err := strconv.Atoi(userIDstr)
	if err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	var taskName models.TaskName

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// Попытка декодировать JSON в структуру UserData
	if err := dec.Decode(&taskName); err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	taskID, err := useCase.UseCaseCreateTask(userID, taskName.Name)

	if err != nil {

		if strings.Contains(err.Error(), "не найден") {
			logger.SugaredLogger().Debug(err)
			w.WriteHeader(http.StatusNotFound)
		} else {
			logger.SugaredLogger().Debug(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	response := map[string]int{"TaskID": taskID}
	res, err := json.Marshal(response)
	if err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

// @Summary Начать отсчет времени по задаче для пользователя
// @Description Устанавливает время начала выполнения задачи по её ID.
// @Tags Tasks
// @Accept json
// @Produce json
// @Param taskID path int true "ID задачи"
// @Success 200 {string} string "TaskID: {taskID}"
// @Failure 404 {string} string "Задача не найдена"
// @Failure 409 {string} string "Время начала уже установлено"
// @Failure 422 {string} string "Ошибка Task ID"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /addTask/startTime/{taskID} [put]
func HandlerStartTime(w http.ResponseWriter, r *http.Request, useCase usecase.UseCaseStorage) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	taskIDStr := chi.URLParam(r, "taskID")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = useCase.UseCaseAddStartTime(taskID)
	if err != nil {
		if strings.Contains(err.Error(), "не найдена") {
			logger.SugaredLogger().Debug(err)
			w.WriteHeader(http.StatusNotFound)
		} else if strings.Contains(err.Error(), "уже заполнено") {
			logger.SugaredLogger().Debug(err)
			w.WriteHeader(http.StatusConflict)
		} else {
			logger.SugaredLogger().Debug(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Закончить отсчет времени по задаче для пользователя
// @Description Устанавливает время окончания выполнения задачи по её ID.
// @Tags Tasks
// @Accept json
// @Produce json
// @Param taskID path int true "ID задачи"
// @Success 200 {string} string "TaskID: {taskID}"
// @Failure 404 {string} string "Задача не найдена"
// @Failure 409 {string} string "Время старта уже задано"
// @Failure 422 {string} string "Ошибка Task ID"
// @Failure 428 {string} string "Не заполнено поле StartTime"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /addTask/endTime/{taskID} [put]
func HandlerEndTime(w http.ResponseWriter, r *http.Request, useCase usecase.UseCaseStorage) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	taskIDStr := chi.URLParam(r, "taskID")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = useCase.UseCaseAddEndTime(taskID)
	if err != nil {
		if strings.Contains(err.Error(), "не найдена") {
			logger.SugaredLogger().Debug(err)
			w.WriteHeader(http.StatusNotFound)
		} else if strings.Contains(err.Error(), "уже заполнено") {
			logger.SugaredLogger().Debug(err)
			w.WriteHeader(http.StatusConflict)
		} else if strings.Contains(err.Error(), "не заполнено") {
			logger.SugaredLogger().Debug(err)
			w.WriteHeader(http.StatusPreconditionRequired)
		} else {
			logger.SugaredLogger().Debug(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Получение задач пользователя
// @Description Возвращает список задач пользователя за указанный период времени.
// @Tags Tasks
// @Accept json
// @Produce json
// @Param userID path int true "ID пользователя"
// @Param body body models.TaskTime true "Фильтрация по периоду времени: start - начало периода, end - конец периода. Начало и конец прописывать в формате ДД.ММ.ГГГГ"
// @Success 200 {array} models.Tasks "Список задач пользователя"
// @Failure 404 {string} string "Данные не найдены"
// @Failure 422 {string} string "Неправильный ID пользователя"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /getTasks/{userID} [post]
func HandlerGetTasks(w http.ResponseWriter, r *http.Request, useCase usecase.UseCaseStorage) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userIDstr := chi.URLParam(r, "userID")

	userID, err := strconv.Atoi(userIDstr)
	if err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	var period models.TaskTime
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&period); err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	if period.Start == "" {
		period.Start = "01.01.1900"
	}
	if period.End == "" {
		period.End = time.Now().Format("02.01.2006")
	}

	// Преобразуем строки Start и End в тип time.Time
	sTime, err := time.Parse("02.01.2006", period.Start)
	if err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	eTime, err := time.Parse("02.01.2006", period.End)
	if err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	period.Start = sTime.Format("2006-01-02 15:04:05")
	period.End = eTime.Format("2006-01-02 15:04:05")

	userData, err := useCase.UseCaseGetTasksUser(userID, period)

	if err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(userData)
	if err != nil {
		logger.SugaredLogger().Debug(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)

}
