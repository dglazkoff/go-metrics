package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 && strings.Contains(c.Header().Get("Accept-Encoding"), "gzip") &&
		(c.Header().Get("Content-Type") == "application/json" || c.Header().Get("Content-Type") == "text/html") {
		c.Header().Set("Content-Encoding", "gzip")
	}
	c.WriteHeader(statusCode)
}

type compressReader struct {
	*http.Request
	Reader io.Reader
	Closer io.Closer
}

func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.Reader.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.Close(); err != nil {
		return err
	}
	return c.Close()
}

func GzipHandle(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ow := writer

		if strings.Contains(writer.Header().Get("Accept-Encoding"), "gzip") &&
			(writer.Header().Get("Content-Type") == "application/json" || writer.Header().Get("Content-Type") == "text/html") {

			w, err := gzip.NewWriterLevel(writer, gzip.BestSpeed)
			defer w.Close()

			if err == nil {
				ow = &compressWriter{ow, w}
				ow.Header().Set("Content-Encoding", "gzip")
			}
		}

		if request.Header.Get("Content-Encoding") == "gzip" {
			r, err := gzip.NewReader(request.Body)

			if err == nil {
				request.Body = &compressReader{request, r, r}
			}

			defer r.Close()
		}

		next.ServeHTTP(ow, request)
	}
}
