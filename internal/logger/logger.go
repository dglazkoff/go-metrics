package logger

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Logger struct {
	*zap.SugaredLogger
}

var Log *Logger

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.responseData.size += size
	return size, err
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.responseData.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func Initialize() error {
	if Log != nil {
		return nil
	}

	logger, err := zap.NewDevelopment()
	defer logger.Sync()

	if err != nil {
		return err
	}

	sugar := logger.Sugar()

	Log = &Logger{sugar}

	return nil
}

func (log *Logger) Request(h http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: writer,
			responseData:   responseData,
		}

		h(&lw, request)

		duration := time.Since(start)

		log.Infoln(
			"uri", request.URL.String(),
			"method", request.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	}
}
