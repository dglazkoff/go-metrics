package subnetvalidate

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubnetValidateHandle(t *testing.T) {
	err := logger.Initialize()
	require.NoError(t, err)

	cfg := config.Config{
		TrustedSubnet: "192.168.31.206/32",
	}
	ts := Initialize(&cfg)

	tests := []struct {
		name         string
		ip           string
		resultStatus int
	}{
		{
			name:         "success",
			ip:           "192.168.31.206",
			resultStatus: http.StatusOK,
		},
		{
			name:         "no header",
			ip:           "",
			resultStatus: http.StatusForbidden,
		},
		{
			name:         "wrong ip format",
			ip:           "192.168.31,206",
			resultStatus: http.StatusForbidden,
		},
		{
			name:         "wrong subnet",
			ip:           "192.168.31.205",
			resultStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := ts.Validate(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)

			if tt.ip != "" {
				req.Header.Set("X-Real-IP", tt.ip)
			}

			w := httptest.NewRecorder()
			handler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.resultStatus, resp.StatusCode)
		})
	}

	t.Run("success if no trusted subnet was set", func(t *testing.T) {
		cfg = config.Config{
			TrustedSubnet: "",
		}
		ts = Initialize(&cfg)

		handler := ts.Validate(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Real-IP", "192.168.31,206")

		w := httptest.NewRecorder()
		handler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
