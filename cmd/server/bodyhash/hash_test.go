package bodyhash

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashHandle_Request(t *testing.T) {
	err := logger.Initialize()
	require.NoError(t, err)

	cfg := config.Config{
		RunAddr:         ":8080",
		FileStoragePath: "/tmp/metrics-db.json",
		StoreInterval:   300,
		IsRestore:       true,
		DatabaseDSN:     "",
		SecretKey:       "secret_123",
	}
	bh := Initialize(&cfg)

	tests := []struct {
		name         string
		hash         string
		resultStatus int
	}{
		{
			name:         "success",
			hash:         "5a7305380fe3259f1de01206f83366b58b52c9b9616c0555a155eef3927dc2ca",
			resultStatus: http.StatusOK,
		},
		{
			name:         "wring",
			hash:         "5a7305380fe3259f1de01206f83366b58b52c9b9616c0555a155eef3927dc2cb",
			resultStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			_, err = buf.Write([]byte("Hidden body"))
			require.NoError(t, err)

			handler := bh.BodyHash(func(w http.ResponseWriter, r *http.Request) {
			})

			req := httptest.NewRequest(http.MethodGet, "/", buf)

			req.Header.Set("HashSHA256", tt.hash)

			w := httptest.NewRecorder()
			handler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.resultStatus, resp.StatusCode)
		})
	}
}

func TestHashHandle_Response(t *testing.T) {
	err := logger.Initialize()
	require.NoError(t, err)

	cfg := config.Config{
		RunAddr:         ":8080",
		FileStoragePath: "/tmp/metrics-db.json",
		StoreInterval:   300,
		IsRestore:       true,
		DatabaseDSN:     "",
		SecretKey:       "secret_123",
	}
	bh := Initialize(&cfg)

	handler := bh.BodyHash(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hidden body"))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()

	decodedBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(t, "5a7305380fe3259f1de01206f83366b58b52c9b9616c0555a155eef3927dc2ca", resp.Header.Get("HashSHA256"))
	assert.Equal(t, []byte("Hidden body"), decodedBody)
}
