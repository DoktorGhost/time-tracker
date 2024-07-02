package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"time"
)

// структура для хранения данных о ответе
type responseData struct {
	status int
	size   int
}

// Обёртка для http.ResponseWriter, чтобы мы могли перехватывать данные об ответе
type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

var sugar *zap.SugaredLogger

func InitLogger(logFile string) error {
	// Настройка конфигурации логгера
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{logFile, "stdout"}

	// Настройка уровня логирования
	cfg.Level.SetLevel(zapcore.DebugLevel)

	cfg.DisableStacktrace = true

	// Создание логгера
	logger, err := cfg.Build()
	if err != nil {
		return err
	}

	// Инициализация SugaredLogger
	sugar = logger.Sugar()
	return nil
}

func SugaredLogger() *zap.SugaredLogger {
	return sugar
}

func WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		sugar.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	}
	return http.HandlerFunc(logFn)
}
