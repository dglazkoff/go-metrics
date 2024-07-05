package storage

import (
	"encoding/json"
	"github.com/dglazkoff/go-metrics/cmd/server/flags"
	"github.com/dglazkoff/go-metrics/cmd/server/logger"
	"os"
	"path/filepath"
	"time"
)

func ReadMetrics(store *MemStorage) {
	if !flags.FlagIsRestore {
		return
	}

	dir, _ := os.Getwd()
	path := filepath.Join(dir, flags.FlagFileStoragePath)

	logger.Log.Debug("Opening file ", path)
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	defer file.Close()

	if err != nil {
		logger.Log.Debug("Error while open file ", err)
		return
	}

	value := MemStorage{}
	err = json.NewDecoder(file).Decode(&value)

	if err != nil {
		logger.Log.Debug("Error while decode ", err)
		return
	}

	*store = value
}

func WriteMetrics(store *MemStorage, isLoop bool) {
	if flags.FlagFileStoragePath == "" {
		return
	}

	dir, _ := os.Getwd()
	path := filepath.Join(dir, flags.FlagFileStoragePath)

	logger.Log.Debug("Creating dir ", filepath.Dir(path))
	err := os.MkdirAll(filepath.Dir(path), 0750)

	if err != nil {
		logger.Log.Debug("Error while create dir ", err)
		return
	}

	var isLoopTemp bool = true

	for isLoopTemp {
		time.Sleep(time.Duration(flags.FlagStoreInterval) * time.Second)

		logger.Log.Debug("Opening file ", path)
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		defer file.Close()
		if err != nil {
			logger.Log.Debug("Error while open file ", err)
			return
		}

		enc := json.NewEncoder(file)

		err = enc.Encode(*store)

		if err != nil {
			logger.Log.Debug("Error while write store to file ", err)
		}

		if !isLoop {
			isLoopTemp = false
		}
	}
}
