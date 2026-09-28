package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lebendig13/metrics/internal/config"
	"github.com/lebendig13/metrics/internal/logger"
	models "github.com/lebendig13/metrics/internal/model"
	"go.uber.org/zap"
)

func RestoreMetrics(memStorage *models.MemStorage, fileStoragePath string) error {
	data, err := os.ReadFile(fileStoragePath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Log.Info("File doesn't exist")
			return nil
		}
		return fmt.Errorf("cannot read backup file: %w", err)
	}

	var savedMetrics []models.Metrics
	err = json.Unmarshal(data, &savedMetrics)
	if err != nil {
		return fmt.Errorf("cannot unmarshal json metrics array: %w", err)
	}

	for _, m := range savedMetrics {
		err = memStorage.InsertOrUpdate(m)
		if err != nil {
			return fmt.Errorf("cannot restore metric %s to storage: %w", m.ID, err)
		}
	}

	return nil
}

func SaveMetrics(currentMetrics []models.Metrics, fpath string) error {
	dir := filepath.Dir(fpath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create directories for path %s: %w", dir, err)
	}

	var buf bytes.Buffer
	buf.WriteString("[\n")
	for i, metric := range currentMetrics {
		metricJSON, err := json.Marshal(metric)
		if err != nil {
			return fmt.Errorf("cannot marshal metric %s: %w", metric.ID, err)
		}

		buf.WriteString("  ")
		buf.Write(metricJSON)

		if i < len(currentMetrics)-1 {
			buf.WriteString(",\n")
		} else {
			buf.WriteString("\n")
		}
	}
	buf.WriteString("]")

	return os.WriteFile(fpath, buf.Bytes(), 0644)
}

func ProcessBackup(ctx context.Context, memStorage *models.MemStorage, cnf *config.ServerConfig) {
	if cnf.FileStoragePath == "" {
		logger.Log.Warn("Cannot backup metrics: empty file storage path")
		return
	}

	if cnf.Restore {
		err := RestoreMetrics(memStorage, cnf.FileStoragePath)
		if err != nil {
			logger.Log.Warn("Cannot restore metrics from file", zap.Error(err))
		}
	}

	if cnf.StoreInterval == 0 {
		logger.Log.Debug("Store interval == 0")
		return
	}
	if cnf.StoreInterval < 0 {
		logger.Log.Warn("Cannot backup metrics: invalid store interval")
		return
	}

	go func() {
		ticker := time.NewTicker(time.Duration(cnf.StoreInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				currentMetrics := memStorage.Snapshot()
				err := SaveMetrics(currentMetrics, cnf.FileStoragePath)
				if err != nil {
					logger.Log.Error("Cannot save current metrics", zap.Error(err))
				}
			case <-ctx.Done():
				logger.Log.Info("Saving current metrics before shutdown...")
				currentMetrics := memStorage.Snapshot()
				err := SaveMetrics(currentMetrics, cnf.FileStoragePath)
				if err != nil {
					logger.Log.Error("Cannot save current metrics", zap.Error(err))
				}
				return
			}
		}
	}()
}
