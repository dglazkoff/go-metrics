package main

import "flag"

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
}
