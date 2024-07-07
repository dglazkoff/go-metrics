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
}

/*
настоятельно не рекомендую использовать глобальные переменные,
самое потимально это создать ф-ю которая будет возращать структуру Config в нутри себя уже парсить флаги/файлы/переменные окружения
*/

func parseConfig() Config {
	config := Config{}

	flag.StringVar(&config.runAddr, "a", ":8080", "address of the server")
	flag.IntVar(&config.pollInterval, "r", 10, "частота отправки метрик на сервер")
	flag.IntVar(&config.reportInterval, "p", 2, "частота опроса метрик из пакета runtime")
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

	return config
}
