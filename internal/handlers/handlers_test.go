package handlers

import (
	"bytes"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time-tracker/internal/config"
	"time-tracker/internal/logger"
	"time-tracker/internal/models"
	"time-tracker/internal/usecase/mocks"
)

// мок сервер, имитация стороннего АПИ
func NewMockServer() *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем параметры запроса
		passportSerie := r.URL.Query().Get("passportSerie")
		passportNumber := r.URL.Query().Get("passportNumber")

		// Проверяем наличие параметров
		if passportSerie == "" || passportNumber == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Отправляем ответ
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
            "surname": "Иванов",
            "name": "Иван",
            "patronymic": "Иванович",
            "address": "г. Москва, ул. Ленина, д. 5, кв. 1"
        }`)
	})

	return httptest.NewServer(handler)
}

func TestHandlerAddUser(t *testing.T) {
	if err := logger.InitLogger(""); err != nil {
		panic("cannot initialize zap")
	}
	defer logger.SugaredLogger().Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := mocks.NewMockUseCaseStorage(ctrl)

	// Создание мок-сервера
	mockServer := NewMockServer()
	defer mockServer.Close()

	// Создаем конфигурацию и роутер с использованием моков
	conf := &config.Config{
		SERVER_HOST: "localhost",
		SERVER_PORT: "8080",
		API_URL:     mockServer.URL,
	}
	router := InitRoutes(mockUseCase, conf)

	type args struct {
		body io.Reader
	}
	tests := []struct {
		name       string
		method     string
		url        string
		body       args
		mockCreate func()
		wantStatus int
	}{
		{
			name:   "#1 Успешный запрос",
			method: http.MethodPost,
			url:    "/user",
			body:   args{bytes.NewBufferString(`{"passportNumber": "1234 567890"}`)},
			mockCreate: func() {
				mockUseCase.EXPECT().UseCaseCreate(gomock.Any()).Return(1, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "#2 Неверный метод запроса",
			method:     http.MethodGet,
			url:        "/user",
			body:       args{bytes.NewBufferString(`{"passportNumber": "1234 567890"}`)},
			mockCreate: func() {},
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "#3 Ошибка в теле запроса",
			method:     http.MethodPost,
			url:        "/user",
			body:       args{bytes.NewBufferString(`{"passportNumber": "1214 567890}`)},
			mockCreate: func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "#4 Ошибка валидации данных (серия)",
			method:     http.MethodPost,
			url:        "/user",
			body:       args{bytes.NewBufferString(`{"passportNumber": "12в4 567890"}`)},
			mockCreate: func() {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "#5 Ошибка валидации данных (номер)",
			method:     http.MethodPost,
			url:        "/user",
			body:       args{bytes.NewBufferString(`{"passportNumber": "1214 56в890"}`)},
			mockCreate: func() {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "#6 Неверный JSON",
			method:     http.MethodPost,
			url:        "/user",
			body:       args{bytes.NewBufferString(`{"par": "1214 561890"}`)},
			mockCreate: func() {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockCreate()

			req, err := http.NewRequest(tt.method, tt.url, tt.body.body)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}

func TestHandlerDelete(t *testing.T) {
	if err := logger.InitLogger(""); err != nil {
		panic("cannot initialize zap")
	}
	defer logger.SugaredLogger().Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := mocks.NewMockUseCaseStorage(ctrl)

	// Создаем конфигурацию и роутер с использованием моков
	conf := &config.Config{
		SERVER_HOST: "localhost",
		SERVER_PORT: "8080",
	}
	router := InitRoutes(mockUseCase, conf)

	tests := []struct {
		name       string
		method     string
		url        string
		mockCreate func()
		wantStatus int
	}{
		{
			name:   "#1 Успешный запрос",
			method: http.MethodDelete,
			url:    "/user/123",
			mockCreate: func() {
				mockUseCase.EXPECT().UseCaseDelete(gomock.Any()).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "#2 Неправильный метод",
			method:     http.MethodPost,
			url:        "/user/123",
			mockCreate: func() {},
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "#3 Некорректный ID",
			method:     http.MethodDelete,
			url:        "/user/fdg",
			mockCreate: func() {},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockCreate()

			req, err := http.NewRequest(tt.method, tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}

func TestHandlerUpdate(t *testing.T) {
	if err := logger.InitLogger(""); err != nil {
		panic("cannot initialize zap")
	}
	defer logger.SugaredLogger().Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := mocks.NewMockUseCaseStorage(ctrl)

	// Создаем конфигурацию и роутер с использованием моков
	conf := &config.Config{
		SERVER_HOST: "localhost",
		SERVER_PORT: "8080",
	}
	router := InitRoutes(mockUseCase, conf)

	type args struct {
		body io.Reader
	}
	tests := []struct {
		name       string
		method     string
		url        string
		body       args
		mockCreate func()
		wantStatus int
	}{
		{
			name:   "#1 Успешный запрос",
			method: http.MethodPut,
			url:    "/user/1",
			body:   args{bytes.NewBufferString(`{"id": "1", "surname": "dfd"}`)},
			mockCreate: func() {
				mockUseCase.EXPECT().UseCaseUpdate(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "#2 Ошибка ID",
			method:     http.MethodPut,
			url:        "/user/f",
			body:       args{bytes.NewBufferString(`{"UserID": "1", "Surname": "dfd"}`)},
			mockCreate: func() {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "#3 Ошибка декодирования тела запроса",
			method:     http.MethodPut,
			url:        "/user/1",
			body:       args{bytes.NewBufferString(`{"UserID": "1", "Surname": "dfd"}`)},
			mockCreate: func() {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockCreate()

			req, err := http.NewRequest(tt.method, tt.url, tt.body.body)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}

func TestHandlerGetUser(t *testing.T) {
	if err := logger.InitLogger(""); err != nil {
		panic("cannot initialize zap")
	}
	defer logger.SugaredLogger().Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := mocks.NewMockUseCaseStorage(ctrl)

	// Создаем конфигурацию и роутер с использованием моков
	conf := &config.Config{
		SERVER_HOST: "localhost",
		SERVER_PORT: "8080",
	}
	router := InitRoutes(mockUseCase, conf)

	tests := []struct {
		name       string
		method     string
		url        string
		mockCreate func()
		wantStatus int
	}{
		{
			name:   "#1 Успешный запрос",
			method: http.MethodGet,
			url:    "/user/1",
			mockCreate: func() {
				mockUseCase.EXPECT().UseCaseRead(gomock.Any()).Return(models.UserData{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "#2 Неправильный метод",
			method: http.MethodPost,
			url:    "/user/1",
			mockCreate: func() {
			},
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "#3 Неверный ID",
			method: http.MethodGet,
			url:    "/user/а",
			mockCreate: func() {
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockCreate()

			req, err := http.NewRequest(tt.method, tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}

func TestHandlerGetUsers(t *testing.T) {
	if err := logger.InitLogger(""); err != nil {
		panic("cannot initialize zap")
	}
	defer logger.SugaredLogger().Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := mocks.NewMockUseCaseStorage(ctrl)

	// Создаем конфигурацию и роутер с использованием моков
	conf := &config.Config{
		SERVER_HOST: "localhost",
		SERVER_PORT: "8080",
	}
	router := InitRoutes(mockUseCase, conf)

	type args struct {
		body io.Reader
	}
	tests := []struct {
		name       string
		method     string
		url        string
		body       args
		mockCreate func()
		wantStatus int
	}{
		{
			name:   "#1 Успешный запрос",
			method: http.MethodPost,
			url:    "/users/1/5",
			body:   args{bytes.NewBufferString(`{"name": "name"}`)},
			mockCreate: func() {
				mockUseCase.EXPECT().UseCaseGetUsers(gomock.Any(), 1, 5).Return([]models.UserData{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "#2 Неверный метод",
			method:     http.MethodGet,
			url:        "/users/1/5",
			body:       args{bytes.NewBufferString(`{"name": "name"}`)},
			mockCreate: func() {},
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "#3 Неправильное тело запроса",
			method:     http.MethodPost,
			url:        "/users/1/5",
			body:       args{bytes.NewBufferString(`{"name": "name}`)},
			mockCreate: func() {},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "#4 Неверные поля в теле запроса",
			method:     http.MethodPost,
			url:        "/users/1/5",
			body:       args{bytes.NewBufferString(`{"nae": "name"}`)},
			mockCreate: func() {},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockCreate()

			req, err := http.NewRequest(tt.method, tt.url, tt.body.body)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}

func TestHandlerAddTask(t *testing.T) {
	if err := logger.InitLogger(""); err != nil {
		panic("cannot initialize zap")
	}
	defer logger.SugaredLogger().Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := mocks.NewMockUseCaseStorage(ctrl)

	// Создаем конфигурацию и роутер с использованием моков
	conf := &config.Config{
		SERVER_HOST: "localhost",
		SERVER_PORT: "8080",
	}
	router := InitRoutes(mockUseCase, conf)

	type args struct {
		body io.Reader
	}
	tests := []struct {
		name       string
		method     string
		url        string
		body       args
		mockCreate func()
		wantStatus int
	}{
		{
			name:   "#1 Успешный запрос",
			method: http.MethodPost,
			url:    "/task/1",
			body:   args{bytes.NewBufferString(`{"task_name": "name"}`)},
			mockCreate: func() {
				mockUseCase.EXPECT().UseCaseCreateTask(gomock.Any(), gomock.Any()).Return(1, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "#2 Неверный метод",
			method:     http.MethodGet,
			url:        "/task/1",
			body:       args{bytes.NewBufferString(`{"task_name": "name"}`)},
			mockCreate: func() {},
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "#3 Неверный ID",
			method:     http.MethodPost,
			url:        "/task/d",
			body:       args{bytes.NewBufferString(`{"task_name": "name"}`)},
			mockCreate: func() {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "#4 Неверне тело запроса",
			method:     http.MethodPost,
			url:        "/task/1",
			body:       args{bytes.NewBufferString(`{"task_name": "name}`)},
			mockCreate: func() {},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "#5 Неверные поля запроса",
			method:     http.MethodPost,
			url:        "/task/1",
			body:       args{bytes.NewBufferString(`{"taskme": "name"}`)},
			mockCreate: func() {},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockCreate()

			req, err := http.NewRequest(tt.method, tt.url, tt.body.body)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}

func TestHandlerStartTime(t *testing.T) {
	if err := logger.InitLogger(""); err != nil {
		panic("cannot initialize zap")
	}
	defer logger.SugaredLogger().Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := mocks.NewMockUseCaseStorage(ctrl)

	// Создаем конфигурацию и роутер с использованием моков
	conf := &config.Config{
		SERVER_HOST: "localhost",
		SERVER_PORT: "8080",
	}
	router := InitRoutes(mockUseCase, conf)

	tests := []struct {
		name       string
		method     string
		url        string
		mockCreate func()
		wantStatus int
	}{
		{
			name:   "#1 Успешный запрос",
			method: http.MethodPut,
			url:    "/task/start/1",
			mockCreate: func() {
				mockUseCase.EXPECT().UseCaseAddStartTime(gomock.Any()).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "#2 Неверный метод",
			method:     http.MethodPost,
			url:        "/task/start/1",
			mockCreate: func() {},
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "#3 Некорректный TaskID",
			method:     http.MethodPut,
			url:        "/task/start/trt",
			mockCreate: func() {},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockCreate()

			req, err := http.NewRequest(tt.method, tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			// Проверка других аспектов ответа, если необходимо
		})
	}
}

// HandlerEndTime
func TestHandlerEndTime(t *testing.T) {
	if err := logger.InitLogger(""); err != nil {
		panic("cannot initialize zap")
	}
	defer logger.SugaredLogger().Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := mocks.NewMockUseCaseStorage(ctrl)

	// Создаем конфигурацию и роутер с использованием моков
	conf := &config.Config{
		SERVER_HOST: "localhost",
		SERVER_PORT: "8080",
	}
	router := InitRoutes(mockUseCase, conf)

	tests := []struct {
		name       string
		method     string
		url        string
		mockCreate func()
		wantStatus int
	}{
		{
			name:   "#1 Успешный запрос",
			method: http.MethodPut,
			url:    "/task/end/1",
			mockCreate: func() {
				mockUseCase.EXPECT().UseCaseAddEndTime(gomock.Any()).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "#2 Неверный метод",
			method:     http.MethodPost,
			url:        "/task/end/1",
			mockCreate: func() {},
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "#3 Некорректный TaskID",
			method:     http.MethodPut,
			url:        "/task/end/trt",
			mockCreate: func() {},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockCreate()

			req, err := http.NewRequest(tt.method, tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			// Проверка других аспектов ответа, если необходимо
		})
	}
}

func TestHandlerGetTasks(t *testing.T) {
	if err := logger.InitLogger(""); err != nil {
		panic("cannot initialize zap")
	}
	defer logger.SugaredLogger().Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := mocks.NewMockUseCaseStorage(ctrl)

	// Создаем конфигурацию и роутер с использованием моков
	conf := &config.Config{
		SERVER_HOST: "localhost",
		SERVER_PORT: "8080",
	}
	router := InitRoutes(mockUseCase, conf)

	type args struct {
		body io.Reader
	}
	tests := []struct {
		name       string
		method     string
		url        string
		body       args
		mockCreate func()
		wantStatus int
	}{
		{
			name:   "#1 Успешный запрос",
			method: http.MethodPost,
			url:    "/tasks/1",
			body:   args{bytes.NewBufferString(`{"start": "12.12.2024"}`)},
			mockCreate: func() {
				mockUseCase.EXPECT().UseCaseGetTasksUser(gomock.Any(), gomock.Any()).Return([]models.Tasks{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "#2 Неверный метод",
			method:     http.MethodGet,
			url:        "/tasks/1",
			body:       args{bytes.NewBufferString(`{"start": "12.12.2024"}`)},
			mockCreate: func() {},
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "#3 Неправильное тело запроса",
			method:     http.MethodPost,
			url:        "/tasks/1",
			body:       args{bytes.NewBufferString(`{"start": "12.12.2024}`)},
			mockCreate: func() {},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "#4 Неверные поля в теле запроса",
			method:     http.MethodPost,
			url:        "/tasks/1",
			body:       args{bytes.NewBufferString(`{"startdsf": "12.12.2024"}`)},
			mockCreate: func() {},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "#4 Невлидные поля в теле запроса",
			method:     http.MethodPost,
			url:        "/tasks/1",
			body:       args{bytes.NewBufferString(`{"start": "12.авы.2024"}`)},
			mockCreate: func() {},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockCreate()

			req, err := http.NewRequest(tt.method, tt.url, tt.body.body)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}
