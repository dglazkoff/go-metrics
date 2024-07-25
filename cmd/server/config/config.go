package config

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	RunAddr         string
	StoreInterval   int
	FileStoragePath string
	IsRestore       bool
	DatabaseDSN     string
}

func ParseConfig() Config {
	cfg := Config{}

	flag.StringVar(&cfg.RunAddr, "a", ":8080", "address of the server")
	flag.StringVar(&cfg.FileStoragePath, "f", "/tmp/metrics-db.json", "file path of metrics storage")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "interval to save metrics on disk")
	flag.BoolVar(&cfg.IsRestore, "r", true, "is restore saved metrics data")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "database dsn string")
	flag.Parse()

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

	return cfg
}
