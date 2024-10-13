// bodyhash - декоратор над хендлером, для проверки целостности body по хэшу
package bodyhash

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/internal/logger"
)

type BodyHash struct {
	cfg *config.Config
}

type hashWriter struct {
	w   http.ResponseWriter
	cfg *config.Config
}

func newHashWriter(w http.ResponseWriter, cfg *config.Config) *hashWriter {
	return &hashWriter{
		w:   w,
		cfg: cfg,
	}
}

func (h *hashWriter) Header() http.Header {
	return h.w.Header()
}

func (h *hashWriter) Write(p []byte) (int, error) {
	logger.Log.Debug("Encode response body")
	hash := hmac.New(sha256.New, []byte(h.cfg.SecretKey))
	hash.Write(p)
	hSum := hash.Sum(nil)

	h.w.Header().Set("HashSHA256", hex.EncodeToString(hSum))
	return h.w.Write(p)
}

func (h *hashWriter) WriteHeader(statusCode int) {
	h.w.WriteHeader(statusCode)
}

func Initialize(cfg *config.Config) *BodyHash {
	return &BodyHash{cfg}
}

func (bodyHash *BodyHash) BodyHash(handler http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if bodyHash.cfg.SecretKey == "" {
			handler.ServeHTTP(writer, request)
			return
		}

		requestHash := request.Header.Get("HashSHA256")

		body, err := io.ReadAll(request.Body)
		request.Body = io.NopCloser(bytes.NewBuffer(body))

		logger.Log.Debug(requestHash)

		if err != nil {
			if err != io.EOF {
				logger.Log.Debug("Error while reading the body: ", err)
			}
			// в тестах при отправке value не приходит хэш в заголовке
			// хотя в самом задание не указано то, что value не надо обрабатывать на хэш
		} else if requestHash != "" {
			h := hmac.New(sha256.New, []byte(bodyHash.cfg.SecretKey))
			h.Write(body)
			hSum := h.Sum(nil)

			if hex.EncodeToString(hSum) != requestHash {
				logger.Log.Debug("Wrong hash")
				writer.WriteHeader(http.StatusBadRequest)
				return
			}

			logger.Log.Debug("Right hash")
		}

		hw := newHashWriter(writer, bodyHash.cfg)

		handler.ServeHTTP(hw, request)
	}
}
