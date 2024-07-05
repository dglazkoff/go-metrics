package main

import (
	"flag"
	"os"
	"strconv"
)

var flagRunAddr string
var flagStoreInterval int
var flagFileStoragePath string
var flagIsRestore bool

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "address of the server")
	flag.StringVar(&flagFileStoragePath, "f", "/tmp/metrics-db.json", "file path of metrics storage")
	flag.IntVar(&flagStoreInterval, "i", 300, "interval to save metrics on disk")
	flag.BoolVar(&flagIsRestore, "r", true, "is restore saved metrics data")
	flag.Parse()

	if runAddr := os.Getenv("ADDRESS"); runAddr != "" {
		flagRunAddr = runAddr
	}

	if storeInterval := os.Getenv("STORE_INTERVAL"); storeInterval != "" {
		value, err := strconv.Atoi(storeInterval)

		if err == nil {
			flagStoreInterval = value
		}
	}

	if storagePath := os.Getenv("FILE_STORAGE_PATH"); storagePath != "" {
		flagFileStoragePath = storagePath
	}

	if isRestore := os.Getenv("RESTORE"); isRestore != "" {
		value, err := strconv.ParseBool(isRestore)

		if err == nil {
			flagIsRestore = value
		}
	}
}
