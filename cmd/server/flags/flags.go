package flags

import (
	"flag"
	"os"
	"strconv"
)

var FlagRunAddr string
var FlagStoreInterval int
var FlagFileStoragePath string
var FlagIsRestore bool

func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", ":8080", "address of the server")
	flag.StringVar(&FlagFileStoragePath, "f", "/tmp/metrics-db.json", "file path of metrics storage")
	flag.IntVar(&FlagStoreInterval, "i", 300, "interval to save metrics on disk")
	flag.BoolVar(&FlagIsRestore, "r", true, "is restore saved metrics data")
	flag.Parse()

	if runAddr := os.Getenv("ADDRESS"); runAddr != "" {
		FlagRunAddr = runAddr
	}

	if storeInterval := os.Getenv("STORE_INTERVAL"); storeInterval != "" {
		value, err := strconv.Atoi(storeInterval)

		if err == nil {
			FlagStoreInterval = value
		}
	}

	if storagePath := os.Getenv("FILE_STORAGE_PATH"); storagePath != "" {
		FlagFileStoragePath = storagePath
	}

	if isRestore := os.Getenv("RESTORE"); isRestore != "" {
		value, err := strconv.ParseBool(isRestore)

		if err == nil {
			FlagIsRestore = value
		}
	}
}
