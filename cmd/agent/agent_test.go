package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"runtime"
	"testing"

	"github.com/dglazkoff/go-metrics/cmd/agent/config"
	constants "github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMetrics(t *testing.T) {
	err := logger.Initialize()
	assert.NoError(t, err)

	floatValue := 42.5
	var intValue int64 = 5
	var uintValue uint64 = 10

	var uintToFloatValue float64 = 10

	tests := []struct {
		name     string
		gm       GaugeMetrics
		cm       CounterMetrics
		expected []models.Metrics
	}{
		{
			name: "basic parse",
			gm: GaugeMetrics{
				RandomValue: floatValue,
				MemStats: runtime.MemStats{
					TotalAlloc: uintValue,
				},
			},
			cm: CounterMetrics{
				PollCount: intValue,
			},
			expected: []models.Metrics{
				{MType: constants.MetricTypeGauge, ID: "RandomValue", Value: &floatValue},
				{MType: constants.MetricTypeGauge, ID: "TotalAlloc", Value: &uintToFloatValue},
				{MType: constants.MetricTypeCounter, ID: "PollCount", Delta: &intValue},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseMetrics(&tt.gm, &tt.cm)
			assert.Subset(t, result, tt.expected)
		})
	}
}

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
			encryptedBody, err := encryptBody(tt.body, cfg)

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
