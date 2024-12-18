package client

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/dglazkoff/go-metrics/cmd/agent/config"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_SendMetricsByHTTP_Success(t *testing.T) {
	err := logger.Initialize()
	assert.NoError(t, err)

	httpClient := NewClient([]time.Duration{1 * time.Millisecond, 2 * time.Millisecond, 3 * time.Millisecond})
	httpmock.ActivateNonDefault(httpClient.client.GetClient())
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "http://localhost:8080/updates/",
		httpmock.NewStringResponder(200, "OK"))

	cfg := &config.Config{
		RunAddr:   "localhost:8080",
		SecretKey: "testkey",
	}
	metrics := []models.Metrics{}

	httpClient.SendMetricsByHTTP(metrics, cfg)

	info := httpmock.GetCallCountInfo()
	assert.Equal(t, 1, info["POST http://localhost:8080/updates/"], "Expected /updates/ to be called once")
}

func TestClient_SendRequest_RetryLogic(t *testing.T) {
	err := logger.Initialize()
	assert.NoError(t, err)

	httpClient := NewClient([]time.Duration{1 * time.Millisecond, 2 * time.Millisecond, 3 * time.Millisecond})
	httpmock.ActivateNonDefault(httpClient.client.GetClient())
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "http://localhost:8080/updates/",
		func(req *http.Request) (*http.Response, error) {
			callCount := httpmock.GetTotalCallCount()
			if callCount < 3 {
				return nil, errors.New("simulated network error")
			}
			return httpmock.NewStringResponse(200, "OK"), nil
		},
	)

	cfg := &config.Config{
		RunAddr:   "localhost:8080",
		SecretKey: "testkey",
	}
	metrics := []models.Metrics{}
	httpClient.SendMetricsByHTTP(metrics, cfg)

	info := httpmock.GetCallCountInfo()
	assert.Equal(t, 3, info["POST http://localhost:8080/updates/"], "Expected /updates/ to be called three times")
}

//
//func TestClient_SendBody_ErrorHandling(t *testing.T) {
//	httpmock.ActivateNonDefault(resty.New().GetClient())
//	defer httpmock.DeactivateAndReset()
//
//	// Mock config
//	cfg := &config.Config{SecretKey: "testkey"}
//	client := NewClient([]time.Duration{1 * time.Millisecond, 2 * time.Millisecond, 3 * time.Millisecond})
//
//	// Mock response to simulate successful sending
//	httpmock.RegisterResponder("POST", "http://localhost:8080/updates/",
//		httpmock.NewStringResponder(200, "OK"))
//
//	client.client.SetBaseURL("http://localhost:8080")
//
//	body := []byte("test body")
//	client.sendBody(body, cfg)
//
//	// Assert that the /updates/ endpoint was called once
//	info := httpmock.GetCallCountInfo()
//	assert.Equal(t, 1, info["POST http://localhost:8080/updates/"], "Expected /updates/ to be called once")
//}

func TestEncryptBody(t *testing.T) {
	err := logger.Initialize()
	require.NoError(t, err)

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKeyBytes := x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: publicKeyBytes})
	keyFile, _ := os.CreateTemp("", "public_key.pem")
	defer os.Remove(keyFile.Name())
	os.WriteFile(keyFile.Name(), publicKeyPEM, 0644)

	cfg := &config.Config{CryptoKey: keyFile.Name()}

	tests := []struct {
		name        string
		body        []byte
		setup       func()
		expectError bool
	}{
		{
			name:        "successful encryption",
			body:        []byte("test body for encryption"),
			setup:       func() {},
			expectError: false,
		},
		{
			name: "error reading public key",
			body: []byte("test body for encryption"),
			setup: func() {
				cfg.CryptoKey = "invalid_path.pem"
			},
			expectError: true,
		},
		{
			name:        "error encrypting segment",
			body:        bytes.Repeat([]byte("A"), 256),
			setup:       func() {},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			encryptedBody, err := EncryptBody(tt.body, cfg)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, encryptedBody)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, encryptedBody)
				assert.NotEmpty(t, encryptedBody)
			}
		})
	}
}
