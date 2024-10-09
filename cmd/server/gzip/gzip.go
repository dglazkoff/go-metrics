// gzip - декоратор над хендлером, для сжимания отправляемого запроса и распковки сжатого body
package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/dglazkoff/go-metrics/internal/logger"
)

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
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
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

// @tmvrus не нравится решение с isHTML, но не знаю другого, так как в request не приходит заголовок text/html при запросе страницы
func GzipHandle(next http.HandlerFunc, isHTML bool) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ow := writer

		if strings.Contains(request.Header.Get("Accept-Encoding"), "gzip") &&
			(request.Header.Get("Content-Type") == "application/json" || isHTML || request.Header.Get("Content-Type") == "text/html") {

			logger.Log.Debug("Handler with gzip compression response ", request.URL)
			cw := newCompressWriter(writer)

			ow = cw
			ow.Header().Set("Content-Encoding", "gzip")

			defer cw.Close()
		}

		if request.Header.Get("Content-Encoding") == "gzip" {
			logger.Log.Debug("Request body with gzip compression ", request.URL)

			r, err := gzip.NewReader(request.Body)

			if err != nil {
				logger.Log.Debug("Error while reading the body: ", err)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}

			request.Body = &compressReader{r: request.Body, zr: r}

			defer r.Close()
		}

		next.ServeHTTP(ow, request)
	}
}
