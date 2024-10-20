package main

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	runAddr        string
	pollInterval   int
	reportInterval int
	secretKey      string
	rateLimit      int
	cryptoKey      string
}

/*
настоятельно не рекомендую использовать глобальные переменные,
самое потимально это создать ф-ю которая будет возращать структуру Config в нутри себя уже парсить флаги/файлы/переменные окружения
*/

func parseConfig() Config {
	config := Config{}

	flag.StringVar(&config.runAddr, "a", ":8080", "address of the server")
	flag.IntVar(&config.reportInterval, "r", 10, "частота отправки метрик на сервер")
	flag.IntVar(&config.pollInterval, "p", 2, "частота опроса метрик из пакета runtime")
	flag.StringVar(&config.secretKey, "k", "", "ключ для кодирования запроса")
	flag.StringVar(&config.cryptoKey, "crypto-key", "", "путь до файла с публичным ключом")
	flag.IntVar(&config.rateLimit, "l", 1, "количество одновременно исходящих запросов")
	flag.Parse()

	if runAddr := os.Getenv("ADDRESS"); runAddr != "" {
		config.runAddr = runAddr
	}

	if pollInterval := os.Getenv("POLL_INTERVAL"); pollInterval != "" {
		value, err := strconv.Atoi(pollInterval)

		if err == nil {
			config.pollInterval = value
		}
	}

	if reportInterval := os.Getenv("REPORT_INTERVAL"); reportInterval != "" {
		value, err := strconv.Atoi(reportInterval)

		if err == nil {
			config.reportInterval = value
		}
	}

	if secretKey := os.Getenv("KEY"); secretKey != "" {
		config.secretKey = secretKey
	}

	if rateLimit := os.Getenv("RATE_LIMIT"); rateLimit != "" {
		value, err := strconv.Atoi(rateLimit)

		if err == nil {
			config.rateLimit = value
		}
	}

	if cryptoKey := os.Getenv("CRYPTO_KEY"); cryptoKey != "" {
		config.cryptoKey = cryptoKey
	}

	return config
}
