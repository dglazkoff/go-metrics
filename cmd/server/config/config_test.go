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
		"-f", "./storage",
		"-i", "88",
		"-k", "secret",
		"-r",
		"-d", "some_dsn",
		"-crypto-key", "crypto",
		"-c", "",
	}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cfg := ParseConfig()

	assert.Equal(t, ":9090", cfg.RunAddr)
	assert.Equal(t, 88, cfg.StoreInterval)
	assert.Equal(t, "./storage", cfg.FileStoragePath)
	assert.Equal(t, "secret", cfg.SecretKey)
	assert.Equal(t, "some_dsn", cfg.DatabaseDSN)
	assert.Equal(t, true, cfg.IsRestore)
	// assert.Equal(t, "crypto", cfg.CryptoKey)
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

	oldKey := os.Getenv("KEY")
	err = os.Setenv("KEY", "secret")
	defer os.Setenv("KEY", oldKey)
	require.NoError(t, err)

	oldCryptoKey := os.Getenv("CRYPTO_KEY")
	err = os.Setenv("CRYPTO_KEY", "crypto")
	defer os.Setenv("CRYPTO_KEY", oldCryptoKey)
	require.NoError(t, err)

	oldStoreInterval := os.Getenv("STORE_INTERVAL")
	err = os.Setenv("STORE_INTERVAL", "88")
	defer os.Setenv("STORE_INTERVAL", oldStoreInterval)
	require.NoError(t, err)

	oldStoragePath := os.Getenv("FILE_STORAGE_PATH")
	err = os.Setenv("FILE_STORAGE_PATH", "./storage")
	defer os.Setenv("FILE_STORAGE_PATH", oldStoragePath)
	require.NoError(t, err)

	oldIsRestore := os.Getenv("RESTORE")
	err = os.Setenv("RESTORE", "true")
	defer os.Setenv("RESTORE", oldIsRestore)
	require.NoError(t, err)

	oldDatabaseDSN := os.Getenv("DATABASE_DSN")
	err = os.Setenv("DATABASE_DSN", "some_dsn")
	defer os.Setenv("DATABASE_DSN", oldDatabaseDSN)
	require.NoError(t, err)

	cfg := ParseConfig()

	assert.Equal(t, ":9090", cfg.RunAddr)
	assert.Equal(t, 88, cfg.StoreInterval)
	assert.Equal(t, "./storage", cfg.FileStoragePath)
	assert.Equal(t, "secret", cfg.SecretKey)
	assert.Equal(t, "some_dsn", cfg.DatabaseDSN)
	assert.Equal(t, true, cfg.IsRestore)
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
		"-f", "./storage",
		"-i", "88",
		"-k", "secret",
		"-r", "true",
		"-d", "some_dsn",
		"-crypto-key", "crypto",
		"-c", "",
	}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	oldAddress := os.Getenv("ADDRESS")
	err = os.Setenv("ADDRESS", ":9091")
	defer os.Setenv("ADDRESS", oldAddress)
	require.NoError(t, err)

	oldKey := os.Getenv("KEY")
	err = os.Setenv("KEY", "secret1")
	defer os.Setenv("KEY", oldKey)
	require.NoError(t, err)

	oldCryptoKey := os.Getenv("CRYPTO_KEY")
	err = os.Setenv("CRYPTO_KEY", "crypto1")
	defer os.Setenv("CRYPTO_KEY", oldCryptoKey)
	require.NoError(t, err)

	oldStoreInterval := os.Getenv("STORE_INTERVAL")
	err = os.Setenv("STORE_INTERVAL", "100")
	defer os.Setenv("STORE_INTERVAL", oldStoreInterval)
	require.NoError(t, err)

	oldStoragePath := os.Getenv("FILE_STORAGE_PATH")
	err = os.Setenv("FILE_STORAGE_PATH", "./storage1")
	defer os.Setenv("FILE_STORAGE_PATH", oldStoragePath)
	require.NoError(t, err)

	oldIsRestore := os.Getenv("RESTORE")
	err = os.Setenv("RESTORE", "false")
	defer os.Setenv("RESTORE", oldIsRestore)
	require.NoError(t, err)

	oldDatabaseDSN := os.Getenv("DATABASE_DSN")
	err = os.Setenv("DATABASE_DSN", "some_dsn1")
	defer os.Setenv("DATABASE_DSN", oldDatabaseDSN)
	require.NoError(t, err)

	cfg := ParseConfig()

	assert.Equal(t, ":9091", cfg.RunAddr)
	assert.Equal(t, 100, cfg.StoreInterval)
	assert.Equal(t, "./storage1", cfg.FileStoragePath)
	assert.Equal(t, "secret1", cfg.SecretKey)
	assert.Equal(t, "some_dsn1", cfg.DatabaseDSN)
	assert.Equal(t, false, cfg.IsRestore)
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
		"secret_key": "secret",
		"crypto_key": "crypto",
		"store_interval": 88,
		"file_storage_path": "./storage",
		"database_dsn": "some_dsn",
		"is_restore": true
	}`
	_, err = tmpFile.Write([]byte(configContent))
	require.NoError(t, err)

	err = tmpFile.Close()
	require.NoError(t, err)

	cfg := ParseConfig()

	assert.Equal(t, ":9090", cfg.RunAddr)
	assert.Equal(t, 88, cfg.StoreInterval)
	assert.Equal(t, "./storage", cfg.FileStoragePath)
	assert.Equal(t, "secret", cfg.SecretKey)
	assert.Equal(t, "some_dsn", cfg.DatabaseDSN)
	assert.Equal(t, true, cfg.IsRestore)
	assert.Equal(t, "crypto", cfg.CryptoKey)
}
