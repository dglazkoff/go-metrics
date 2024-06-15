package main

import (
	"flag"
	"os"
	"strconv"
)

var (
	flagRunAddr        string
	flagPollInterval   int
	flagReportInterval int
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "address of the server")
	flag.IntVar(&flagReportInterval, "r", 10, "частота отправки метрик на сервер")
	flag.IntVar(&flagPollInterval, "p", 2, "частота опроса метрик из пакета runtime")
	flag.Parse()

	if runAddr := os.Getenv("ADDRESS"); runAddr != "" {
		flagRunAddr = runAddr
	}

	if pollInterval := os.Getenv("POLL_INTERVAL"); pollInterval != "" {
		value, err := strconv.Atoi(pollInterval)

		if err == nil {
			flagPollInterval = value
		}
	}

	if reportInterval := os.Getenv("REPORT_INTERVAL"); reportInterval != "" {
		value, err := strconv.Atoi(reportInterval)

		if err == nil {
			flagReportInterval = value
		}
	}
}
