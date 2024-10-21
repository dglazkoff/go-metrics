package main

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"

	"github.com/dglazkoff/go-metrics/internal/logger"
)

type Config struct {
	RunAddr        string `json:"run_addr"`
	PollInterval   int    `json:"poll_interval"`
	ReportInterval int    `json:"report_interval"`
	SecretKey      string `json:"secret_key"`
	RateLimit      int    `json:"rate_limit"`
	CryptoKey      string `json:"crypto_key"`
}

/*
настоятельно не рекомендую использовать глобальные переменные,
самое потимально это создать ф-ю которая будет возращать структуру Config в нутри себя уже парсить флаги/файлы/переменные окружения
*/

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

	if config.PollInterval == 0 && fileConfig.PollInterval != 0 {
		config.PollInterval = fileConfig.PollInterval
	}

	if config.ReportInterval == 0 && fileConfig.ReportInterval != 0 {
		config.ReportInterval = fileConfig.ReportInterval
	}

	if config.SecretKey == "" && fileConfig.SecretKey != "" {
		config.SecretKey = fileConfig.SecretKey
	}

	if config.RateLimit == 0 && fileConfig.RateLimit != 0 {
		config.RateLimit = fileConfig.RateLimit
	}

	if config.CryptoKey == "" && fileConfig.CryptoKey != "" {
		config.CryptoKey = fileConfig.CryptoKey
	}
}

func parseConfig() Config {
	config := Config{}
	var configFile string

	flag.StringVar(&config.RunAddr, "a", "", "address of the server")
	flag.IntVar(&config.ReportInterval, "r", 0, "частота отправки метрик на сервер")
	flag.IntVar(&config.PollInterval, "p", 0, "частота опроса метрик из пакета runtime")
	flag.StringVar(&config.SecretKey, "k", "", "ключ для кодирования запроса")
	flag.StringVar(&config.CryptoKey, "crypto-key", "", "путь до файла с публичным ключом")
	flag.IntVar(&config.RateLimit, "l", 0, "количество одновременно исходящих запросов")
	flag.StringVar(&configFile, "c", "cmd/agent/config.json", "имя файла конфигурации")
	flag.Parse()

	readConfigFile(configFile, &config)

	if runAddr := os.Getenv("ADDRESS"); runAddr != "" {
		config.RunAddr = runAddr
	}

	if pollInterval := os.Getenv("POLL_INTERVAL"); pollInterval != "" {
		value, err := strconv.Atoi(pollInterval)

		if err == nil {
			config.PollInterval = value
		}
	}

	if reportInterval := os.Getenv("REPORT_INTERVAL"); reportInterval != "" {
		value, err := strconv.Atoi(reportInterval)

		if err == nil {
			config.ReportInterval = value
		}
	}

	if secretKey := os.Getenv("KEY"); secretKey != "" {
		config.SecretKey = secretKey
	}

	if rateLimit := os.Getenv("RATE_LIMIT"); rateLimit != "" {
		value, err := strconv.Atoi(rateLimit)

		if err == nil {
			config.RateLimit = value
		}
	}

	if configFileEnv := os.Getenv("CONFIG"); configFileEnv != "" {
		configFile = configFileEnv
	}

	if cryptoKey := os.Getenv("CRYPTO_KEY"); cryptoKey != "" {
		config.CryptoKey = cryptoKey
	}

	return config
}
