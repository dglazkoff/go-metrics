package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
)

type metricStorage interface {
	SaveMetrics(ctx context.Context, metrics []models.Metrics) error
	ReadMetrics(ctx context.Context) ([]models.Metrics, error)
}

type fileStorage struct {
	storage metricStorage
	cfg     *config.Config
}

func New(s metricStorage, cfg *config.Config) fileStorage {
	return fileStorage{storage: s, cfg: cfg}
}

func closeFile(f *os.File) {
	logger.Log.Debug("Closing file")
	err := f.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}

func (s fileStorage) ReadMetrics() {
	if !s.cfg.IsRestore {
		return
	}

	ctx := context.Background()
	dir, err := os.Getwd()

	if err != nil {
		logger.Log.Debug("Error while get current dir ", err)
		return
	}

	path := filepath.Join(dir, s.cfg.FileStoragePath)

	logger.Log.Debug("Opening file ", path)
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)

	/*
		тут есть потенциальная проблема, если os.OpenFile вернет ошибку, то file будет nil и defer closeFile(file) вызовет панику.
		По этой причине добавлять defer нужно после проверки на ошибку
	*/
	// defer closeFile(file)

	if err != nil {
		logger.Log.Debug("Error while open file ", err)
		return
	}

	defer closeFile(file)

	var metrics []models.Metrics
	err = json.NewDecoder(file).Decode(&metrics)

	if err != nil {
		logger.Log.Debug("Error while decode ", err)
		return
	}

	err = s.storage.SaveMetrics(ctx, metrics)

	if err != nil {
		logger.Log.Debug("Error while save metrics: ", err)
	}
}

func (s fileStorage) WriteMetrics(isLoop bool) {
	if s.cfg.FileStoragePath == "" {
		return
	}

	ctx := context.Background()
	dir, _ := os.Getwd()
	path := filepath.Join(dir, s.cfg.FileStoragePath)

	logger.Log.Debug("Creating dir ", filepath.Dir(path))
	err := os.MkdirAll(filepath.Dir(path), 0750)

	if err != nil {
		logger.Log.Debug("Error while create dir ", err)
		return
	}

	for {
		time.Sleep(time.Duration(s.cfg.StoreInterval) * time.Second)

		logger.Log.Debug("Opening file ", path)
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)

		if err != nil {
			logger.Log.Debug("Error while open file ", err)
			return
		}

		/*
			этот defer отработает только когда завершится ф-я, она может никогда не завершится и цикл будет копить defer, по этой причине defer в цикле - плохая идея
		*/
		defer closeFile(file)

		enc := json.NewEncoder(file)

		metrics, err := s.storage.ReadMetrics(ctx)

		if err != nil {
			logger.Log.Debug("Error while read metrics ", err)
			return

		}

		err = enc.Encode(metrics)

		if err != nil {
			logger.Log.Debug("Error while write store to file ", err)
		}

		if !isLoop {
			break
		}
	}
}
