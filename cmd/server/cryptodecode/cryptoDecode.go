package cryptodecode

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"
	"os"

	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/internal/logger"
)

type CryptoBody struct {
	cfg *config.Config
}

func Initialize(cfg *config.Config) *CryptoBody {
	return &CryptoBody{cfg}
}

func (cryptoBody *CryptoBody) CryptoDecode(handler http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		privateKeyPEM, err := os.ReadFile(cryptoBody.cfg.CryptoKey)
		if err != nil {
			logger.Log.Debug("Error reading private key: ", err)
			handler.ServeHTTP(writer, request)
			return
		}
		privateKeyBlock, _ := pem.Decode(privateKeyPEM)
		privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
		if err != nil {
			logger.Log.Debug("Error parsing private key: ", err)
			handler.ServeHTTP(writer, request)
			return
		}

		body, err := io.ReadAll(request.Body)

		if err != nil {
			logger.Log.Debug("Error reading request body: ", err)
			handler.ServeHTTP(writer, request)
			return
		}

		defer request.Body.Close()

		var decryptedBody bytes.Buffer
		segmentSize := privateKey.Size()

		for i := 0; i < len(body); i += segmentSize {
			j := i + segmentSize
			if j > len(body) {
				j = len(body)
			}
			encryptedSegment := body[i:j]

			decryptedSegment, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedSegment)
			if err != nil {
				logger.Log.Debug("Error while decrypting data: ", err)
				handler.ServeHTTP(writer, request)
				return
			}

			decryptedBody.Write(decryptedSegment)
		}

		logger.Log.Debug("Successful decryption body")
		request.Body = io.NopCloser(bytes.NewReader(decryptedBody.Bytes()))

		handler.ServeHTTP(writer, request)
	}
}
