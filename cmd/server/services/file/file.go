package file

import (
	"encoding/json"
	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/cmd/server/logger"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"github.com/dglazkoff/go-metrics/internal/models"
	"os"
	"path/filepath"
	"time"
)

type metricStorage interface {
	UpdateMetric(metric models.Metrics) error
}

type service struct {
	storage metricStorage
	cfg     *config.Config
}

// тут не надо по указателю передавать?
func New(s storage.MetricsStorage, cfg *config.Config) service {
	return service{storage: s, cfg: cfg}
}

func (s service) ReadMetrics() {
	if !s.cfg.IsRestore {
		return
	}

	dir, _ := os.Getwd()
	path := filepath.Join(dir, s.cfg.FileStoragePath)

	logger.Log.Debug("Opening file ", path)
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	defer file.Close()

	if err != nil {
		logger.Log.Debug("Error while open file ", err)
		return
	}

	value := models.Metrics{}
	err = json.NewDecoder(file).Decode(&value)

	if err != nil {
		logger.Log.Debug("Error while decode ", err)
		return
	}

	s.storage.UpdateMetric(value)
}

func (s service) WriteMetrics(isLoop bool) {
	if s.cfg.FileStoragePath == "" {
		return
	}

	dir, _ := os.Getwd()
	path := filepath.Join(dir, s.cfg.FileStoragePath)

	logger.Log.Debug("Creating dir ", filepath.Dir(path))
	err := os.MkdirAll(filepath.Dir(path), 0750)

	if err != nil {
		logger.Log.Debug("Error while create dir ", err)
		return
	}

	var isLoopTemp bool = true

	for isLoopTemp {
		time.Sleep(time.Duration(s.cfg.StoreInterval) * time.Second)

		logger.Log.Debug("Opening file ", path)
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		defer file.Close()
		if err != nil {
			logger.Log.Debug("Error while open file ", err)
			return
		}

		enc := json.NewEncoder(file)

		err = enc.Encode(s.storage)

		if err != nil {
			logger.Log.Debug("Error while write store to file ", err)
		}

		if !isLoop {
			isLoopTemp = false
		}
	}
}
