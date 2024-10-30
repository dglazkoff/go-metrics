package cryptodecode

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCryptoDecode(t *testing.T) {
	err := logger.Initialize()
	require.NoError(t, err)

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privateKeyBytes})
	keyFile, _ := os.CreateTemp("", "private_key.pem")
	defer os.Remove(keyFile.Name())
	os.WriteFile(keyFile.Name(), privateKeyPEM, 0644)

	cfg := &config.Config{CryptoKey: keyFile.Name()}
	cryptoBody := Initialize(cfg)

	encryptedData, _ := rsa.EncryptPKCS1v15(rand.Reader, &privateKey.PublicKey, []byte("test body"))

	tests := []struct {
		name     string
		body     []byte
		setup    func()
		expected []byte
	}{
		{
			name:     "successful decryption",
			body:     encryptedData,
			setup:    func() {},
			expected: []byte("test body"),
		},
		{
			name: "error reading private key",
			body: encryptedData,
			setup: func() {
				cryptoBody.cfg.CryptoKey = "invalid_path.pem"
			},
			expected: encryptedData,
		},
		//{
		//	name: "error parsing private key",
		//	body: encryptedData,
		//	setup: func() {
		//		os.WriteFile(keyFile.Name(), []byte("invalid_key_data"), 0644)
		//	},
		//	expectError: true,
		//},
		//{
		//	name:        "error reading request body",
		//	body:        nil,
		//	setup:       func() {},
		//	expectError: true,
		//},
		{
			name:     "error decrypting segment",
			body:     []byte("invalid_encrypted_data"),
			setup:    func() {},
			expected: []byte("invalid_encrypted_data"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(tt.body))
			rec := httptest.NewRecorder()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				w.Write(body)
			})

			cryptoBody.CryptoDecode(handler).ServeHTTP(rec, req)

			assert.Equal(t, tt.expected, rec.Body.Bytes())
		})
	}
}
