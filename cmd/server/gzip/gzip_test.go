package gzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipHandle_Response(t *testing.T) {
	err := logger.Initialize()
	require.NoError(t, err)

	handler := GzipHandle(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, Gzip!"))
	}, false)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "text/html")

	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()

	gr, err := gzip.NewReader(resp.Body)
	require.NoError(t, err)
	defer resp.Body.Close()

	decodedBody, err := io.ReadAll(gr)
	require.NoError(t, err)

	assert.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))
	assert.Equal(t, "Hello, Gzip!", string(decodedBody))
}

func TestGzipHandle_Request(t *testing.T) {
	err := logger.Initialize()
	require.NoError(t, err)

	var resultBody []byte

	handler := GzipHandle(func(w http.ResponseWriter, r *http.Request) {
		resultBody, err = io.ReadAll(r.Body)
		defer r.Body.Close()
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}, false)

	buf := bytes.NewBuffer(nil)
	zw := gzip.NewWriter(buf)
	_, err = zw.Write([]byte("Hello, Gzip!"))
	zw.Close()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", buf)
	req.Header.Set("Content-Encoding", "gzip")

	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Hello, Gzip!", string(resultBody))
}
