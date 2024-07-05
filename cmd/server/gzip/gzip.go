package gzip

import (
	"compress/gzip"
	"github.com/dglazkoff/go-metrics/cmd/server/logger"
	"io"
	"net/http"
	"strings"
)

//type compressWriter struct {
//	http.ResponseWriter
//	Writer io.Writer
//}
//
////func (c *compressWriter) WriteHeader(statusCode int) {
////	if statusCode < 300 && strings.Contains(c.Header().Get("Accept-Encoding"), "gzip") &&
////		(c.Header().Get("Content-Type") == "application/json" || c.Header().Get("Content-Type") == "text/html") {
////		c.Header().Set("Content-Encoding", "gzip")
////	}
////	c.WriteHeader(statusCode)
////}
//
//type compressReader struct {
//	r  io.ReadCloser
//	zr *gzip.Reader
//}
//
//func (c compressReader) Read(p []byte) (n int, err error) {
//	return c.zr.Read(p)
//}
//
//func (c *compressReader) Close() error {
//	if err := c.r.Close(); err != nil {
//		return err
//	}
//	return c.zr.Close()
//}

// compressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func GzipHandle(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ow := writer

		if strings.Contains(request.Header.Get("Accept-Encoding"), "gzip") &&
			(request.Header.Get("Content-Type") == "application/json" || request.Header.Get("Content-Type") == "text/html") {

			logger.Log.Debug("Handler with gzip compression response ", request.URL)
			w, err := gzip.NewWriterLevel(writer, gzip.BestSpeed)

			if err == nil {
				ow = &compressWriter{ow, w}
				// ow.Header().Set("Content-Encoding", "gzip")
			}

			defer w.Close()
		}

		if request.Header.Get("Content-Encoding") == "gzip" {
			logger.Log.Debug("Request body with gzip compression ", request.URL)

			r, err := gzip.NewReader(request.Body)

			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}

			request.Body = &compressReader{r: request.Body, zr: r}
			defer r.Close()
		}

		next.ServeHTTP(ow, request)
	}
	//	return func(w http.ResponseWriter, r *http.Request) {
	//		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
	//		// который будем передавать следующей функции
	//		ow := w
	//
	//		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
	//		acceptEncoding := r.Header.Get("Accept-Encoding")
	//		supportsGzip := strings.Contains(acceptEncoding, "gzip")
	//		if supportsGzip {
	//			logger.Log.Debug("Handler with gzip compression response ", r.URL)
	//			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
	//			cw := newCompressWriter(w)
	//			// меняем оригинальный http.ResponseWriter на новый
	//			ow = cw
	//			// не забываем отправить клиенту все сжатые данные после завершения middleware
	//			defer cw.Close()
	//		}
	//
	//		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
	//		contentEncoding := r.Header.Get("Content-Encoding")
	//		sendsGzip := strings.Contains(contentEncoding, "gzip")
	//		if sendsGzip {
	//			logger.Log.Debug("Request body with gzip compression ", r.URL)
	//			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
	//			cr, err := newCompressReader(r.Body)
	//			if err != nil {
	//				w.WriteHeader(http.StatusInternalServerError)
	//				return
	//			}
	//			// меняем тело запроса на новое
	//			r.Body = cr
	//			defer cr.Close()
	//		}
	//
	//		// передаём управление хендлеру
	//		next.ServeHTTP(ow, r)
	//	}
}