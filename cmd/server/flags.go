package main

import (
	"flag"
	"os"
)

var flagRunAddr string

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "address of the server")
	flag.Parse()

	if runAddr := os.Getenv("ADDRESS"); runAddr != "" {
		flagRunAddr = runAddr
	}
}
