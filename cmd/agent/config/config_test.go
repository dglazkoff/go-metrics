package config

import (
	"flag"
	"os"
	"testing"

	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Flags(t *testing.T) {
	err := logger.Initialize()
	require.NoError(t, err)

	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{
		"cmd/agent",
		"-a", ":9090",
		"-p", "25",
		"-r", "88",
		"-k", "secret",
		"-l", "100",
		"-crypto-key", "crypto",
		"-c", "",
		"-grpc",
	}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cfg := ParseConfig()

	assert.Equal(t, ":9090", cfg.RunAddr)
	assert.Equal(t, 25, cfg.PollInterval)
	assert.Equal(t, 88, cfg.ReportInterval)
	assert.Equal(t, "secret", cfg.SecretKey)
	assert.Equal(t, 100, cfg.RateLimit)
	assert.Equal(t, "crypto", cfg.CryptoKey)
	assert.Equal(t, true, cfg.IsGRPC)
}

func TestConfig_SimpleEnv(t *testing.T) {
	err := logger.Initialize()
	require.NoError(t, err)

	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{
		"cmd/agent",
		"-c", "",
	}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	oldAddress := os.Getenv("ADDRESS")
	err = os.Setenv("ADDRESS", ":9090")
	defer os.Setenv("ADDRESS", oldAddress)
	require.NoError(t, err)

	oldPollInterval := os.Getenv("POLL_INTERVAL")
	err = os.Setenv("POLL_INTERVAL", "25")
	defer os.Setenv("POLL_INTERVAL", oldPollInterval)
	require.NoError(t, err)

	oldReportInterval := os.Getenv("REPORT_INTERVAL")
	err = os.Setenv("REPORT_INTERVAL", "88")
	defer os.Setenv("REPORT_INTERVAL", oldReportInterval)
	require.NoError(t, err)

	oldKey := os.Getenv("KEY")
	err = os.Setenv("KEY", "secret")
	defer os.Setenv("KEY", oldKey)
	require.NoError(t, err)

	oldRateLimit := os.Getenv("RATE_LIMIT")
	err = os.Setenv("RATE_LIMIT", "100")
	defer os.Setenv("RATE_LIMIT", oldRateLimit)
	require.NoError(t, err)

	oldCryptoKey := os.Getenv("CRYPTO_KEY")
	err = os.Setenv("CRYPTO_KEY", "crypto")
	defer os.Setenv("CRYPTO_KEY", oldCryptoKey)
	require.NoError(t, err)

	cfg := ParseConfig()

	assert.Equal(t, ":9090", cfg.RunAddr)
	assert.Equal(t, 25, cfg.PollInterval)
	assert.Equal(t, 88, cfg.ReportInterval)
	assert.Equal(t, "secret", cfg.SecretKey)
	assert.Equal(t, 100, cfg.RateLimit)
	assert.Equal(t, "crypto", cfg.CryptoKey)
}

func TestConfig_PriorityEnv(t *testing.T) {
	err := logger.Initialize()
	require.NoError(t, err)

	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{
		"cmd/agent",
		"-a", ":9090",
		"-p", "25",
		"-r", "88",
		"-k", "secret",
		"-l", "100",
		"-crypto-key", "crypto",
		"-c", "",
	}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	oldAddress := os.Getenv("ADDRESS")
	err = os.Setenv("ADDRESS", ":9091")
	defer os.Setenv("ADDRESS", oldAddress)
	require.NoError(t, err)

	oldPollInterval := os.Getenv("POLL_INTERVAL")
	err = os.Setenv("POLL_INTERVAL", "52")
	defer os.Setenv("POLL_INTERVAL", oldPollInterval)
	require.NoError(t, err)

	oldReportInterval := os.Getenv("REPORT_INTERVAL")
	err = os.Setenv("REPORT_INTERVAL", "100")
	defer os.Setenv("REPORT_INTERVAL", oldReportInterval)
	require.NoError(t, err)

	oldKey := os.Getenv("KEY")
	err = os.Setenv("KEY", "secret1")
	defer os.Setenv("KEY", oldKey)
	require.NoError(t, err)

	oldRateLimit := os.Getenv("RATE_LIMIT")
	err = os.Setenv("RATE_LIMIT", "101")
	defer os.Setenv("RATE_LIMIT", oldRateLimit)
	require.NoError(t, err)

	oldCryptoKey := os.Getenv("CRYPTO_KEY")
	err = os.Setenv("CRYPTO_KEY", "crypto1")
	defer os.Setenv("CRYPTO_KEY", oldCryptoKey)
	require.NoError(t, err)

	cfg := ParseConfig()

	assert.Equal(t, ":9091", cfg.RunAddr)
	assert.Equal(t, 52, cfg.PollInterval)
	assert.Equal(t, 100, cfg.ReportInterval)
	assert.Equal(t, "secret1", cfg.SecretKey)
	assert.Equal(t, 101, cfg.RateLimit)
	assert.Equal(t, "crypto1", cfg.CryptoKey)
}

func TestParseConfig_File(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	os.Args = []string{
		"cmd/agent",
		"-c", tmpFile.Name(),
	}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	configContent := `{
		"run_addr": ":9090",
		"poll_interval": 25,
		"report_interval": 88,
		"secret_key": "secret",
		"rate_limit": 100,
		"crypto_key": "crypto",
		"is_grpc": true
	}`
	_, err = tmpFile.Write([]byte(configContent))
	require.NoError(t, err)

	err = tmpFile.Close()
	require.NoError(t, err)

	cfg := ParseConfig()

	assert.Equal(t, ":9090", cfg.RunAddr)
	assert.Equal(t, 25, cfg.PollInterval)
	assert.Equal(t, 88, cfg.ReportInterval)
	assert.Equal(t, "secret", cfg.SecretKey)
	assert.Equal(t, 100, cfg.RateLimit)
	assert.Equal(t, "crypto", cfg.CryptoKey)
	assert.Equal(t, true, cfg.IsGRPC)
}
