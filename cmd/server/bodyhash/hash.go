package bodyhash

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"hash"
	"io"
	"net/http"
)

type BodyHash struct {
	cfg *config.Config
}

type hashWriter struct {
	w    http.ResponseWriter
	hash hash.Hash
}

func newHashWriter(w http.ResponseWriter) *hashWriter {
	return &hashWriter{
		w:    w,
		hash: sha256.New(),
	}
}

func (h *hashWriter) Header() http.Header {
	return h.w.Header()
}

func (h *hashWriter) Write(p []byte) (int, error) {
	logger.Log.Debug("Encode response body")
	h.hash.Write(p)
	// можно ли несколько Write вызывать?
	h.w.Header().Set("HashSHA256", hex.EncodeToString(h.hash.Sum(nil)))
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
		// считываем все
		body, err := io.ReadAll(request.Body)
		request.Body = io.NopCloser(bytes.NewBuffer(body))

		//logger.Log.Debug(bodyHash.cfg.SecretKey)
		logger.Log.Debug(requestHash)
		//logger.Log.Debug(hex.EncodeToString(body))

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

		hw := newHashWriter(writer)

		handler.ServeHTTP(hw, request)
	}
}
