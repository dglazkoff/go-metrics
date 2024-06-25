package logger

import (
	"github.com/dglazkoff/go-metrics/internal/logger"
)

var Log *logger.Log

func Initialize() error {
	var err error
	Log, err = logger.Initialize()

	if err != nil {
		return err
	}

	return nil
}
