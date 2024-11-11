package config

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"

	"github.com/dglazkoff/go-metrics/internal/logger"
)

type Config struct {
	RunAddr         string `json:"run_addr"`
	StoreInterval   int    `json:"store_interval"`
	FileStoragePath string `json:"file_storage_path"`
	IsRestore       bool   `json:"is_restore"`
	DatabaseDSN     string `json:"database_dsn"`
	SecretKey       string `json:"secret_key"`
	CryptoKey       string `json:"crypto_key"`
	TrustedSubnet   string `json:"trusted_subnet"`
	IsGRPC          bool   `json:"is_grpc"`
}

func readConfigFile(configFile string, config *Config) {
	fileConfig := Config{}
	jsonFileConfig, err := os.ReadFile(configFile)

	if err != nil {
		logger.Log.Debug("Error reading config file: ", err)
		return
	}

	err = json.Unmarshal(jsonFileConfig, &fileConfig)

	if err != nil {
		logger.Log.Debug("Error unmarshaling config file: ", err)
		return
	}

	if config.RunAddr == "" && fileConfig.RunAddr != "" {
		config.RunAddr = fileConfig.RunAddr
	}

	if config.StoreInterval == 0 && fileConfig.StoreInterval != 0 {
		config.StoreInterval = fileConfig.StoreInterval
	}

	if config.FileStoragePath == "" && fileConfig.FileStoragePath != "" {
		config.FileStoragePath = fileConfig.FileStoragePath
	}

	if !config.IsRestore && fileConfig.IsRestore {
		config.IsRestore = fileConfig.IsRestore
	}

	if config.DatabaseDSN == "" && fileConfig.DatabaseDSN != "" {
		config.DatabaseDSN = fileConfig.DatabaseDSN
	}

	if config.SecretKey == "" && fileConfig.SecretKey != "" {
		config.SecretKey = fileConfig.SecretKey
	}

	if config.CryptoKey == "" && fileConfig.CryptoKey != "" {
		config.CryptoKey = fileConfig.CryptoKey
	}

	if config.TrustedSubnet == "" && fileConfig.TrustedSubnet != "" {
		config.TrustedSubnet = fileConfig.TrustedSubnet
	}

	if !config.IsGRPC && fileConfig.IsGRPC {
		config.IsGRPC = fileConfig.IsGRPC
	}
}

func ParseConfig() Config {
	cfg := Config{}
	var configFile string

	flag.StringVar(&cfg.RunAddr, "a", "", "address of the server")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "file path of metrics storage")
	flag.IntVar(&cfg.StoreInterval, "i", 0, "interval to save metrics on disk")
	flag.BoolVar(&cfg.IsRestore, "r", false, "is restore saved metrics data")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "database dsn string")
	flag.StringVar(&cfg.SecretKey, "k", "", "ключ для кодирования запроса")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "путь до файла с приватным ключом")
	flag.StringVar(&cfg.TrustedSubnet, "t", "", "строковое представление бесклассовой адресации (CIDR)")
	flag.BoolVar(&cfg.IsGRPC, "grpc", false, "отправка метрик через gRPC")
	flag.StringVar(&configFile, "c", "cmd/server/config/config.json", "имя файла конфигурации")
	flag.Parse()

	readConfigFile(configFile, &cfg)

	if runAddr := os.Getenv("ADDRESS"); runAddr != "" {
		cfg.RunAddr = runAddr
	}

	if storeInterval := os.Getenv("STORE_INTERVAL"); storeInterval != "" {
		value, err := strconv.Atoi(storeInterval)

		if err == nil {
			cfg.StoreInterval = value
		}
	}

	if storagePath := os.Getenv("FILE_STORAGE_PATH"); storagePath != "" {
		cfg.FileStoragePath = storagePath
	}

	if isRestore := os.Getenv("RESTORE"); isRestore != "" {
		value, err := strconv.ParseBool(isRestore)

		if err == nil {
			cfg.IsRestore = value
		}
	}

	if databaseDSN := os.Getenv("DATABASE_DSN"); databaseDSN != "" {
		cfg.DatabaseDSN = databaseDSN
	}

	if secretKey := os.Getenv("KEY"); secretKey != "" {
		cfg.SecretKey = secretKey
	}

	if cryptoKey := os.Getenv("CRYPTO_KEY"); cryptoKey != "" {
		cfg.CryptoKey = cryptoKey
	}

	if trustedSubnet := os.Getenv("TRUSTED_SUBNET"); trustedSubnet != "" {
		cfg.TrustedSubnet = trustedSubnet
	}

	return cfg
}
