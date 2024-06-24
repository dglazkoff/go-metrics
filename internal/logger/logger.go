package logger

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Log struct {
	*zap.SugaredLogger
}

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

func Initialize() (*Log, error) {
	logger, err := zap.NewDevelopment()
	defer logger.Sync()

	if err != nil {
		return nil, err
	}

	sugar := logger.Sugar()

	return &Log{sugar}, nil
}

func (log *Log) Request(h http.HandlerFunc) http.HandlerFunc {
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
